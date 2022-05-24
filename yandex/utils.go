package yandex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	sdkoperation "github.com/yandex-cloud/go-sdk/operation"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

type instanceAction int

const (
	instanceActionStop instanceAction = iota
	instanceActionStart
	instanceActionRestart
)

const defaultTimeFormat = time.RFC3339
const defaultListSize = 1000

type Policy struct {
	Bindings []*access.AccessBinding
}

func getZone(d *schema.ResourceData, config *Config) (string, error) {
	res, ok := d.GetOk("zone")
	if !ok {
		if config.Zone != "" {
			return config.Zone, nil
		}
		return "", fmt.Errorf("cannot determine zone: please set 'zone' key in this resource or at provider level")
	}
	return res.(string), nil
}

func getCloudID(d *schema.ResourceData, config *Config) (string, error) {
	res, ok := d.GetOk("cloud_id")
	if !ok {
		if config.CloudID != "" {
			return config.CloudID, nil
		}
		return "", fmt.Errorf("cannot determine cloud_id: please set 'cloud_id' key in this resource or at provider level")
	}
	return res.(string), nil
}

func getOrganizationID(d *schema.ResourceData, config *Config) (string, error) {
	res, ok := d.GetOk("organization_id")
	if !ok {
		if config.OrganizationID != "" {
			return config.OrganizationID, nil
		}
		return "", fmt.Errorf("cannot determine organization_id: please set 'organization_id' key in this resource or at provider level")
	}
	return res.(string), nil
}

func getFolderID(d *schema.ResourceData, config *Config) (string, error) {
	res, ok := d.GetOk("folder_id")
	if !ok {
		if config.FolderID != "" {
			return config.FolderID, nil
		}
		return "", fmt.Errorf("cannot determine folder_id: please set 'folder_id' key in this resource or at provider level")
	}
	return res.(string), nil
}

func cloudIDOfFolderID(config *Config, folderID string) (string, error) {
	folder, err := config.sdk.ResourceManager().Folder().Get(config.Context(), &resourcemanager.GetFolderRequest{
		FolderId: folderID,
	})
	if err != nil {
		return "", err
	}
	return folder.CloudId, nil
}

func lockCloudByFolderID(config *Config, folderID string) (func(), error) {
	cloudID, err := cloudIDOfFolderID(config, folderID)
	if err != nil {
		return nil, fmt.Errorf("error getting cloud ID of `folder_id` %s: %s", folderID, err)
	}
	c := CloudIamUpdater{cloudID: cloudID}
	mutexKey := c.GetMutexKey()
	mutexKV.Lock(mutexKey)
	return func() {
		mutexKV.Unlock(mutexKey)
	}, nil
}

func createTemporaryStaticAccessKey(roleID string, config *Config) (accessKey, secretKey string, cleanup func(), err error) {
	op, err := config.sdk.WrapOperation(config.sdk.IAM().ServiceAccount().Create(context.Background(), &iam.CreateServiceAccountRequest{
		FolderId: config.FolderID,
		Name:     acctest.RandomWithPrefix("tmp-sa-"),
	}))
	if err != nil {
		return
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return
	}

	md, ok := protoMetadata.(*iam.CreateServiceAccountMetadata)
	if !ok {
		err = fmt.Errorf("could not get temporary service account ID from create operation metadata")
		return
	}

	saID := md.ServiceAccountId

	err = op.Wait(context.Background())
	if err != nil {
		return
	}

	deleteSa := func() {
		// Use a folder updater in a similar way iam resources do to prevent modifying folder access binding while
		// service account is being deleted.
		updater, err := newFolderIamUpdaterFromFolderID(config.FolderID, config)
		if err != nil {
			log.Printf("[WARN] error deleting temporary service account: %s", err)
			return
		}

		mutexKey := updater.GetMutexKey()
		mutexKV.Lock(mutexKey)
		defer mutexKV.Unlock(mutexKey)

		op, err := config.sdk.WrapOperation(config.sdk.IAM().ServiceAccount().Delete(context.Background(), &iam.DeleteServiceAccountRequest{
			ServiceAccountId: saID,
		}))
		if err != nil {
			log.Printf("[WARN] error deleting temporary service account: %s", err)
			return
		}

		err = op.Wait(context.Background())
		if err != nil {
			log.Printf("[WARN] error deleting temporary service account: %s", err)
		}
	}

	createKey := func() (*awscompatibility.CreateAccessKeyResponse, error) {
		op, err = config.sdk.WrapOperation(config.sdk.ResourceManager().Folder().UpdateAccessBindings(context.Background(), &access.UpdateAccessBindingsRequest{
			ResourceId: config.FolderID,
			AccessBindingDeltas: []*access.AccessBindingDelta{
				{
					Action: access.AccessBindingAction_ADD,
					AccessBinding: &access.AccessBinding{
						RoleId: roleID,
						Subject: &access.Subject{
							Id:   saID,
							Type: "serviceAccount",
						},
					},
				},
			},
		}))
		if err != nil {
			return nil, err
		}

		err = op.Wait(context.Background())
		if err != nil {
			return nil, err
		}

		sak, err := config.sdk.IAM().AWSCompatibility().AccessKey().Create(context.Background(), &awscompatibility.CreateAccessKeyRequest{
			ServiceAccountId: saID,
		})
		if err != nil {
			return nil, err
		}

		return sak, err
	}

	sak, err := createKey()
	if err != nil {
		deleteSa()
		return
	}

	accessKey = sak.AccessKey.KeyId
	secretKey = sak.Secret
	cleanup = func() {
		_, err := config.sdk.IAM().AWSCompatibility().AccessKey().Delete(context.Background(), &awscompatibility.DeleteAccessKeyRequest{
			AccessKeyId: sak.AccessKey.Id,
		})
		if err != nil {
			log.Printf("[WARN] error deleting temporary access key: %s", err)
		}

		deleteSa()
	}
	return
}

func retryConflictingOperation(ctx context.Context, config *Config, action func() (*operation.Operation, error)) (*sdkoperation.Operation, error) {
	for {
		op, err := config.sdk.WrapOperation(action())
		if err == nil {
			return op, nil
		}

		operationID := ""
		message := status.Convert(err).Message()
		submatchGoApi := regexp.MustCompile(`conflicting operation "(.+)" detected`).FindStringSubmatch(message)
		submatchPyApi := regexp.MustCompile(`Conflicting operation (.+) detected`).FindStringSubmatch(message)
		if len(submatchGoApi) > 0 {
			operationID = submatchGoApi[1]
		} else if len(submatchPyApi) > 0 {
			operationID = submatchPyApi[1]
		} else {
			return op, err
		}

		log.Printf("[DEBUG] Waiting for conflicting operation %q to complete", operationID)
		req := &operation.GetOperationRequest{OperationId: operationID}
		op, err = config.sdk.WrapOperation(config.sdk.Operation().Get(ctx, req))
		if err != nil {
			return nil, err
		}

		_ = op.Wait(ctx)
		log.Printf("[DEBUG] Conflicting operation %q has completed. Going to retry initial action.", operationID)
	}
}

func handleNotFoundError(err error, d *schema.ResourceData, resourceName string) error {
	if isStatusWithCode(err, codes.NotFound) {
		log.Printf("[WARN] Removing %s because resource doesn't exist anymore", resourceName)
		d.SetId("")
		return nil
	}
	return fmt.Errorf("error reading %s: %s", resourceName, err)
}

func isStatusWithCode(err error, code codes.Code) bool {
	grpcStatus, ok := status.FromError(err)
	return ok && grpcStatus.Code() == code
}

func isRequestIDPresent(err error) (string, bool) {
	st, ok := status.FromError(err)
	if ok {
		for _, d := range st.Details() {
			if reqInfo, ok := d.(*errdetails.RequestInfo); ok {
				return reqInfo.RequestId, true
			}
		}
	}
	return "", false
}

func errorMessage(err error) string {
	grpcStatus, _ := status.FromError(err)
	return grpcStatus.Message()
}

func convertStringArrToInterface(sslice []string) []interface{} {
	islice := make([]interface{}, len(sslice))
	for i, str := range sslice {
		islice[i] = str
	}
	return islice
}

func mergeSchemas(a, b map[string]*schema.Schema) map[string]*schema.Schema {
	merged := make(map[string]*schema.Schema, len(a)+len(b))

	for k, v := range a {
		merged[k] = v
	}

	for k, v := range b {
		merged[k] = v
	}

	return merged
}

func roleMemberToAccessBinding(role, member string) *access.AccessBinding {
	chunks := strings.SplitN(member, ":", 2)
	return &access.AccessBinding{
		RoleId: role,
		Subject: &access.Subject{
			Type: chunks[0],
			Id:   chunks[1],
		},
	}
}

func mergeBindings(bindings []*access.AccessBinding) []*access.AccessBinding {
	bm := rolesToMembersMap(bindings)
	var rb []*access.AccessBinding

	for role, members := range bm {
		for member := range members {
			rb = append(rb, roleMemberToAccessBinding(role, member))
		}
	}

	return rb
}

// Map a role to a map of members, allowing easy merging of multiple bindings.
func rolesToMembersMap(bindings []*access.AccessBinding) map[string]map[string]bool {
	bm := make(map[string]map[string]bool)
	// Get each binding
	for _, b := range bindings {
		// Initialize members map
		if _, ok := bm[b.RoleId]; !ok {
			bm[b.RoleId] = make(map[string]bool)
		}
		// Get each member (user/principal) for the binding
		m := canonicalMember(b)
		bm[b.RoleId][m] = true
	}
	return bm
}

func roleToMembersList(role string, bindings []*access.AccessBinding) []string {
	var members []string

	for _, b := range bindings {
		if b.RoleId != role {
			continue
		}
		m := canonicalMember(b)
		members = append(members, m)
	}
	return members
}

func removeRoleFromBindings(roleForRemove string, bindings []*access.AccessBinding) []*access.AccessBinding {
	bm := rolesToMembersMap(bindings)
	var rb []*access.AccessBinding

	for role, members := range bm {
		if role == roleForRemove {
			continue
		}
		for member := range members {
			rb = append(rb, roleMemberToAccessBinding(role, member))
		}
	}

	return rb
}

func (p Policy) String() string {
	result := ""
	for i, binding := range p.Bindings {
		result = result + fmt.Sprintf("\n#:%d role:%-27s\taccount:%-24s\ttype:%s",
			i, binding.RoleId, binding.Subject.Id, binding.Subject.Type)
	}
	return result + "\n"
}

func convertStringSet(set *schema.Set) []string {
	s := make([]string, set.Len())
	for i, v := range set.List() {
		s[i] = v.(string)
	}
	return s
}

func convertStringMap(v map[string]interface{}) map[string]string {
	m := make(map[string]string)
	if v == nil {
		return m
	}
	for k, val := range v {
		m[k] = val.(string)
	}
	return m
}

func shouldSuppressDiffForPolicies(k, old, new string, d *schema.ResourceData) bool {
	var oldPolicy, newPolicy Policy
	if err := json.Unmarshal([]byte(old), &oldPolicy); err != nil {
		log.Printf("[ERROR] Could not unmarshal old policy %s: %v", old, err)
		return false
	}
	if err := json.Unmarshal([]byte(new), &newPolicy); err != nil {
		log.Printf("[ERROR] Could not unmarshal new policy %s: %v", new, err)
		return false
	}
	oldPolicy.Bindings = mergeBindings(oldPolicy.Bindings)
	newPolicy.Bindings = mergeBindings(newPolicy.Bindings)
	if len(newPolicy.Bindings) != len(oldPolicy.Bindings) {
		return false
	}
	sort.Sort(sortableBindings(newPolicy.Bindings))
	sort.Sort(sortableBindings(oldPolicy.Bindings))
	for pos, newBinding := range newPolicy.Bindings {
		oldBinding := oldPolicy.Bindings[pos]
		if oldBinding.RoleId != newBinding.RoleId {
			return false
		}
		if oldBinding.Subject.Type != newBinding.Subject.Type {
			return false
		}
		if oldBinding.Subject.Id != newBinding.Subject.Id {
			return false
		}
	}
	return true
}

type sortableBindings []*access.AccessBinding

func (b sortableBindings) Len() int {
	return len(b)
}
func (b sortableBindings) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
func (b sortableBindings) Less(i, j int) bool {
	return b.String(i) < b.String(j)
}

func (b sortableBindings) String(i int) string {
	return fmt.Sprintf("%s\x00%s\x00%s", b[i].RoleId, b[i].Subject.Type, b[i].Subject.Id)
}

func validateCidrBlocks(v interface{}, k string) (warnings []string, errors []error) {
	_, _, err := net.ParseCIDR(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is not a valid IP CIDR range: %s", k, err))
	}
	return
}

// Primary use to store value from API in state file as Gigabytes
func toGigabytes(bytesCount int64) int {
	return int((datasize.ByteSize(bytesCount) * datasize.B).GBytes())
}

func toGigabytesInFloat(bytesCount int64) float64 {
	return (datasize.ByteSize(bytesCount) * datasize.B).GBytes()
}

// Primary use to send byte value to API
func toBytes(gigabytesCount int) int64 {
	return int64((datasize.ByteSize(gigabytesCount) * datasize.GB).Bytes())
}

func toBytesFromFloat(gigabytesCountF float64) int64 {
	return int64(gigabytesCountF * float64(datasize.GB))
}

func (action instanceAction) String() string {
	switch action {
	case instanceActionStop:
		return "Stop"
	case instanceActionStart:
		return "Start"
	case instanceActionRestart:
		return "Restart"
	default:
		return "Unknown"
	}
}

func parseDuration(s string) (*durationpb.Duration, error) {
	if s == "" {
		return nil, nil
	}

	v, err := time.ParseDuration(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %v", err)
	}

	if v < 0 {
		return nil, fmt.Errorf("can not use negative duration")
	}

	return durationpb.New(v), nil
}

func formatDuration(d *durationpb.Duration) string {
	if d == nil {
		return ""
	}
	return d.AsDuration().String()
}

func getTimestamp(timestamp *timestamppb.Timestamp) string {
	if timestamp == nil {
		return ""
	}
	return timestamp.AsTime().Format(defaultTimeFormat)
}

func stringSliceToLower(s []string) []string {
	var ret []string
	for _, v := range s {
		ret = append(ret, strings.ToLower(v))
	}
	return ret
}

func getEnumValueMapKeys(m map[string]int32) []string {
	return getEnumValueMapKeysExt(m, false)
}

func getEnumValueMapKeysExt(m map[string]int32, skipDefault bool) []string {
	keys := make([]string, 0, len(m))
	for k, v := range m {
		if v == 0 && skipDefault {
			continue
		}

		keys = append(keys, k)
	}
	return keys
}

func getJoinedKeys(keys []string) string {
	return "`" + strings.Join(keys, "`, `") + "`"
}

func checkOneOf(d *schema.ResourceData, keys ...string) error {
	var gotKey bool
	for _, key := range keys {
		_, ok := d.GetOk(key)

		if ok {
			if gotKey {
				return fmt.Errorf("only one of %s can be provided", getJoinedKeys(keys))
			}

			gotKey = true
		}
	}

	if !gotKey {
		return fmt.Errorf("one of %s should be provided", getJoinedKeys(keys))
	}

	return nil
}

type objectResolverFunc func(name string, opts ...sdkresolvers.ResolveOption) ycsdk.Resolver

// this function can be only used to resolve objects that belong to some folder (have folder_id attribute)
// do not use this function to resolve cloud (or similar objects) ID by name.
func resolveObjectID(ctx context.Context, config *Config, d *schema.ResourceData, resolverFunc objectResolverFunc) (string, error) {
	name, ok := d.GetOk("name")

	if !ok {
		return "", fmt.Errorf("non empty name should be provided")
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return "", err
	}

	return resolveObjectIDByNameAndFolderID(ctx, config, name.(string), folderID, resolverFunc)
}

func resolveObjectIDByNameAndFolderID(ctx context.Context, config *Config, name, folderID string, resolverFunc objectResolverFunc) (string, error) {
	if name == "" {
		return "", fmt.Errorf("non empty name should be provided")
	}

	var objectID string
	resolver := resolverFunc(name, sdkresolvers.Out(&objectID), sdkresolvers.FolderID(folderID))

	err := config.sdk.Resolve(ctx, resolver)

	if err != nil {
		return "", err
	}

	return objectID, nil
}

func getSnapshotMinStorageSize(snapshotID string, config *Config) (size int64, err error) {
	ctx := config.Context()

	snapshot, err := config.sdk.Compute().Snapshot().Get(ctx, &compute.GetSnapshotRequest{
		SnapshotId: snapshotID,
	})

	if err != nil {
		return 0, fmt.Errorf("Error on retrieve snapshot properties: %s", err)
	}

	return snapshot.DiskSize, nil
}

func getImageMinStorageSize(imageID string, config *Config) (size int64, err error) {
	ctx := config.Context()

	image, err := config.sdk.Compute().Image().Get(ctx, &compute.GetImageRequest{
		ImageId: imageID,
	})

	if err != nil {
		return 0, fmt.Errorf("Error on retrieve image properties: %s", err)
	}

	return image.MinDiskSize, nil
}

func templateConfig(tmpl string, ctx ...map[string]interface{}) string {
	p := make(map[string]interface{})
	for _, c := range ctx {
		for k, v := range c {
			p[k] = v
		}
	}
	b := &bytes.Buffer{}
	err := template.Must(template.New("").Parse(tmpl)).Execute(b, p)
	if err != nil {
		panic(fmt.Errorf("failed to execute config template: %v", err))
	}
	return b.String()
}

func getResourceID(n string, s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[n]
	if !ok {
		return "", fmt.Errorf("terraform resource '%s' not found", n)
	}

	if rs.Primary.ID == "" {
		return "", fmt.Errorf("no ID is set for terraform resource '%s'", n)
	}

	return rs.Primary.ID, nil
}

type schemaGetHelper struct {
	pathPrefix string
	d          *schema.ResourceData
}

func schemaHelper(d *schema.ResourceData, path string) *schemaGetHelper {
	return &schemaGetHelper{
		pathPrefix: path,
		d:          d,
	}
}

func (h *schemaGetHelper) AppendPath(path string) *schemaGetHelper {
	return &schemaGetHelper{
		pathPrefix: h.pathPrefix + path,
		d:          h.d,
	}
}

func (h *schemaGetHelper) Get(key string) interface{} {
	return h.d.Get(h.pathPrefix + key)
}

func (h *schemaGetHelper) GetOk(key string) (interface{}, bool) {
	return h.d.GetOk(h.pathPrefix + key)
}

func (h *schemaGetHelper) GetString(key string) string {
	return h.d.Get(h.pathPrefix + key).(string)
}

func (h *schemaGetHelper) GetInt(key string) int {
	return h.d.Get(h.pathPrefix + key).(int)
}

func convertResourceToDataSource(resource *schema.Resource) *schema.Resource {
	return recursivelyUpdateResource(resource, func(schema *schema.Schema) {
		schema.Computed = true
		schema.Required = false
		schema.Optional = false
		schema.ForceNew = false
		schema.Default = nil
		schema.ValidateFunc = nil
		schema.MaxItems = 0
		schema.MinItems = 0
	})
}

func recursivelyUpdateResource(resource *schema.Resource, callback func(*schema.Schema)) *schema.Resource {
	attributes := make(map[string]*schema.Schema)
	for key, attributeSchema := range resource.Schema {
		copyOfAttributeSchema := *attributeSchema
		callback(&copyOfAttributeSchema)
		if copyOfAttributeSchema.Elem != nil {
			switch elem := copyOfAttributeSchema.Elem.(type) {
			case *schema.Schema:
				elementCopy := *elem
				copyOfAttributeSchema.Elem = &elementCopy
			case *schema.Resource:
				copyOfAttributeSchema.Elem = recursivelyUpdateResource(elem, callback)
			default:
				log.Printf("[ERROR] Unexpected Elem type %T for attribute %s!\n", elem, key)
			}
		}

		attributes[key] = &copyOfAttributeSchema
	}

	return &schema.Resource{Schema: attributes}
}

func sortInterfaceListByResourceData(listToSort []interface{}, d *schema.ResourceData, entityName string, cmpFieldName string) {
	templateList, ok := d.GetOk(entityName)
	if !ok || templateList == nil {
		return
	}
	sortInterfaceListByTemplate(listToSort, templateList.([]interface{}), cmpFieldName)
}

func sortInterfaceListByTemplate(listToSort []interface{}, templateList []interface{}, cmpFieldName string) {
	if len(templateList) == 0 || len(listToSort) == 0 {
		return
	}

	sortRule := map[string]int{}

	for i := range templateList {
		sortRule[getField(templateList[i], cmpFieldName)] = i
	}

	sort.Slice(listToSort, func(i int, j int) bool {
		return lessInterfaceList(listToSort, cmpFieldName, i, j, sortRule)
	})
}

func lessInterfaceList(list []interface{}, name string, i int, j int, sortRule map[string]int) bool {
	nameI := getField(list[i], name)
	nameJ := getField(list[j], name)

	posI, okI := sortRule[nameI]
	posJ, okJ := sortRule[nameJ]

	if okI && okJ {
		return posI < posJ
	}

	if okI {
		return true
	}

	if okJ {
		return false
	}

	return nameI < nameJ
}

func getField(value interface{}, field string) string {
	return (value.(map[string]interface{}))[field].(string)
}

func fieldDeprecatedForAnother(deprecatedFieldName string, newFieldName string) string {
	return fmt.Sprintf("The '%s' field has been deprecated. Please use '%s' instead.", deprecatedFieldName, newFieldName)
}

func useResourceInstead(deprecatedFieldName string, newResource string) string {
	return fmt.Sprintf("to manage %ss, please switch to using a separate resource type %s", deprecatedFieldName, newResource)
}

func getSDK(config *Config) *ycsdk.SDK {
	return config.sdk
}

func generateFieldMasks(d *schema.ResourceData, fieldsMap map[string]string) []string {
	changedPaths := make(map[string]bool)

	for longField, longPath := range fieldsMap {
		if !d.HasChange(longField) {
			continue
		}

		fields := splitFieldPath(longField)
		paths := strings.Split(longPath, ".")

		if len(paths) != len(fields) {
			panic(fmt.Sprintf("different length: %s and %s", longField, longPath))
		}

		var found bool

		for i, field := range fields {
			path := strings.Join(paths[0:i+1], ".")

			if _, ok := changedPaths[path]; ok {
				found = true
				break
			}

			if !d.HasChange(field) {
				continue
			}

			if strings.HasSuffix(field, ".0") {
				size := d.Get(field[:len(field)-2] + ".#").(int)
				if size == 0 {
					changedPaths[path] = true
					found = true
					break
				}
			}

			if _, ok := d.GetOk(field); !ok {
				changedPaths[path] = true
				found = true
				break
			}
		}

		if !found {
			changedPaths[longPath] = true
		}
	}

	res := make([]string, 0, len(changedPaths))
	for k := range changedPaths {
		res = append(res, k)
	}

	return res
}

type fieldTreeNode struct {
	protobufFieldName      string
	terraformAttributeName string
	children               []*fieldTreeNode
}

func generateEndpointFieldMasks(d *schema.ResourceData, fieldTreeRoot *fieldTreeNode) []string {
	var fieldMasks []string
	for _, node := range fieldTreeRoot.children {
		fieldMasks = append(fieldMasks, generateEndpointFieldMasksForPath(d, node, "", "")...)
	}
	return fieldMasks
}

func generateEndpointFieldMasksForPath(d *schema.ResourceData, node *fieldTreeNode, terraformPathPrefix, protobufPathPrefix string) []string {
	terraformAttributePath := terraformPathPrefix + node.terraformAttributeName
	protobufFieldPath := protobufPathPrefix + node.protobufFieldName

	if !d.HasChange(terraformAttributePath) {
		return nil // No changes => empty field mask
	}
	// There's a change at terraformAttributePath. Try to refine it by recursing into the attribute

	if node.children == nil {
		// It's either a repeated field of any kind or it's a singular primitive (i.e. non-message) field.
		// Either way, there's nothing to recurse into. Just return the node path
		return []string{protobufFieldPath}
	}

	// The node is a singular message field.
	// Check if the count has changed (0->1 or 1->0). If so, then the entire field has changed, and there is no need to recurse
	countChanged := d.HasChange(terraformAttributePath + ".#")
	if countChanged {
		return []string{protobufFieldPath}
	}

	// The count has not changed, but the field has. Try recursing into the message to find out precisely which attributes that have changed
	var nestedChanges []string
	for _, nestedNode := range node.children {
		nestedChanges = append(nestedChanges, generateEndpointFieldMasksForPath(d, nestedNode, terraformAttributePath+".0.", protobufFieldPath+".")...)
	}
	if len(nestedChanges) != 0 {
		return nestedChanges
	}

	// No nested changes found, but the current node somehow has changes. Return the current path in that case
	return []string{protobufFieldPath}
}

func splitFieldPath(path string) []string {
	newPath := strings.ReplaceAll(path, ".0", "__0")
	var sb strings.Builder
	var paths []string

	for i, token := range strings.Split(newPath, ".") {
		if i != 0 {
			sb.WriteString(".")
		}
		sb.WriteString(strings.ReplaceAll(token, "__0", ".0"))
		paths = append(paths, sb.String())
	}
	return paths
}

func constructResourceId(clusterID string, resourceName string) string {
	return fmt.Sprintf("%s:%s", clusterID, resourceName)
}

func deconstructResourceId(resourceID string) (string, string, error) {
	parts := strings.SplitN(resourceID, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource id format: %q", resourceID)
	}

	clusterID := parts[0]
	resourceName := parts[1]
	return clusterID, resourceName, nil
}

func expandEnum(keyName string, value string, enumValues map[string]int32) (*int32, error) {
	if val, ok := enumValues[value]; ok {
		return &val, nil
	} else {
		return nil, fmt.Errorf("value for '%s' must be one of %s, not `%s`",
			keyName, getJoinedKeys(getEnumValueMapKeys(enumValues)), value)
	}
}
