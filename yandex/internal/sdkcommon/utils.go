package sdkcommon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/davecgh/go-spew/spew"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	sdkoperationv2 "github.com/yandex-cloud/go-sdk/v2/pkg/operation"
	sdkresolversv2 "github.com/yandex-cloud/go-sdk/v2/pkg/sdkresolvers"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const defaultTimeFormat = time.RFC3339
const DefaultListSize = 1000

func ProtoDump(msg proto.Message) string {
	data, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(msg)
	if err != nil {
		return fmt.Sprintf("**ERROR json mashal failed: %+v** message dump: %s", err, spewConfig.Sdump(msg))
	}
	return string(data)
}

var spewConfig = spew.ConfigState{
	Indent:                  " ",
	DisablePointerAddresses: true,
	DisableCapacities:       true,
	SortKeys:                true,
}

type Policy struct {
	Bindings []*access.AccessBinding
}

func GetZone(d *schema.ResourceData, config *providerconfig.Config) (string, error) {
	res, ok := d.GetOk("zone")
	if !ok {
		if config.Zone != "" {
			return config.Zone, nil
		}
		return "", fmt.Errorf("cannot determine zone: please set 'zone' key in this resource or at provider level")
	}
	return res.(string), nil
}

// doesn't show change disk_size on decreasing if enabled autoscaling disk limit
// Example for `autoscalingPath` = disk_size_autoscaling.0.disk_size_limit
func SuppressDiskSizeChangeOnAutoscaling(autoscalingPath string) schema.SchemaDiffSuppressFunc {
	return func(_, old, new string, d *schema.ResourceData) bool {
		autoscaling := d.Get(autoscalingPath).(int)

		newInt, err := strconv.Atoi(new)
		if err != nil {
			return false
		}

		oldInt, err := strconv.Atoi(old)
		if err != nil {
			return false
		}

		if oldInt > newInt && autoscaling > 0 {
			return true
		}

		return false
	}
}

func GetFolderID(d *schema.ResourceData, config *providerconfig.Config) (string, error) {
	res, ok := d.GetOk("folder_id")
	if !ok {
		if config.FolderID != "" {
			return config.FolderID, nil
		}
		return "", fmt.Errorf("cannot determine folder_id: please set 'folder_id' key in this resource or at provider level")
	}
	return res.(string), nil
}

type SdkV2Operation interface {
	Abstract() sdkoperationv2.AbstractOperation
}

func RetryConflictingOperationV2(ctx context.Context, config *providerconfig.Config, action func() (SdkV2Operation, error)) (sdkoperationv2.AbstractOperation, error) {
	for {
		typedOperation, err := action()
		if err == nil {
			return typedOperation.Abstract(), nil
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
			return nil, err
		}

		log.Printf("[DEBUG] Waiting for conflicting operation %q to complete", operationID)
		if err := waitOperationByIDV2(ctx, config, operationID); err != nil {
			return nil, err
		}
		log.Printf("[DEBUG] Conflicting operation %q has completed. Going to retry initial action.", operationID)
	}
}

func waitOperationByIDV2(ctx context.Context, config *providerconfig.Config, operationID string) error {
	connection, err := config.SDK.GetConnection(ctx, protoreflect.FullName("yandex.cloud.operation.OperationService.Get"))
	if err != nil {
		return err
	}
	client := operation.NewOperationServiceClient(connection)
	return WaitConflictingOperationV2(ctx, operationID, func(ctx context.Context, operationID string, opts ...grpc.CallOption) (sdkoperationv2.YCOperation, error) {
		return client.Get(ctx, &operation.GetOperationRequest{OperationId: operationID}, opts...)
	}, func(int) time.Duration { return time.Second })
}

func WaitConflictingOperationV2(ctx context.Context, operationID string, poll sdkoperationv2.PollFunc, pollInterval sdkoperationv2.PollIntervalFunc) error {
	initial, err := poll(ctx, operationID)
	if err != nil {
		return err
	}
	if initial.GetDone() {
		return nil
	}

	// SDK v1 propagated the initial Get error, but ignored the result of Wait.
	// Preserve both parts of that contract while using SDK v2 polling semantics.
	_, _ = sdkoperationv2.PollUntilDone(ctx, operationID, poll, pollInterval)
	return nil
}

func HandleNotFoundError(err error, d *schema.ResourceData, resourceName string) error {
	if IsStatusWithCode(err, codes.NotFound) {
		log.Printf("[WARN] Removing %s because resource doesn't exist anymore", resourceName)
		d.SetId("")
		return nil
	}
	return fmt.Errorf("error reading %s: %s", resourceName, err)
}

// isStatusWithCode checks if any nested error matches provided code
func IsStatusWithCode(err error, code codes.Code) bool {
	grpcStatus, ok := status.FromError(err)
	check := ok && grpcStatus.Code() == code

	nestedErr := errors.Unwrap(err)
	if nestedErr != nil {
		return IsStatusWithCode(nestedErr, code) || check
	}
	return check
}

func IsRequestIDPresent(err error) (string, bool) {
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

func ErrorMessage(err error) string {
	grpcStatus, _ := status.FromError(err)
	return grpcStatus.Message()
}

func ConvertStringArrToInterface(sslice []string) []interface{} {
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

func RoleMemberToAccessBinding(role, member string) *access.AccessBinding {
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
	bm := RolesToMembersMap(bindings)
	var rb []*access.AccessBinding

	for role, members := range bm {
		for member := range members {
			rb = append(rb, RoleMemberToAccessBinding(role, member))
		}
	}

	return rb
}

// Map a role to a map of members, allowing easy merging of multiple bindings.
func RolesToMembersMap(bindings []*access.AccessBinding) map[string]map[string]bool {
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
	bm := RolesToMembersMap(bindings)
	var rb []*access.AccessBinding

	for role, members := range bm {
		if role == roleForRemove {
			continue
		}
		for member := range members {
			rb = append(rb, RoleMemberToAccessBinding(role, member))
		}
	}

	return rb
}

func CountBatches(size, batchSize int) int {
	iterations := size / batchSize
	if size%batchSize > 0 {
		iterations++
	}
	return iterations
}

func (p Policy) String() string {
	result := ""
	for i, binding := range p.Bindings {
		result = result + fmt.Sprintf("\n#:%d role:%-27s\taccount:%-24s\ttype:%s",
			i, binding.RoleId, binding.Subject.Id, binding.Subject.Type)
	}
	return result + "\n"
}

func ConvertStringSet(set *schema.Set) []string {
	s := make([]string, set.Len())
	for i, v := range set.List() {
		s[i] = v.(string)
	}
	return s
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
	sort.Sort(SortableBindings(newPolicy.Bindings))
	sort.Sort(SortableBindings(oldPolicy.Bindings))
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

type SortableBindings []*access.AccessBinding

func (b SortableBindings) Len() int {
	return len(b)
}
func (b SortableBindings) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
func (b SortableBindings) Less(i, j int) bool {
	return b.String(i) < b.String(j)
}

func (b SortableBindings) String(i int) string {
	return fmt.Sprintf("%s\x00%s\x00%s", b[i].RoleId, b[i].Subject.Type, b[i].Subject.Id)
}

func ValidateCidrBlocks(v interface{}, k string) (warnings []string, errors []error) {
	_, _, err := net.ParseCIDR(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is not a valid IP CIDR range: %s", k, err))
	}
	return
}

// Primary use to store value from API in state file as Gigabytes
func ToGigabytes(bytesCount int64) int {
	return int((datasize.ByteSize(bytesCount) * datasize.B).GBytes())
}

func ToGigabytesInFloat(bytesCount int64) float64 {
	return (datasize.ByteSize(bytesCount) * datasize.B).GBytes()
}

// Primary use to send byte value to API
func ToBytes(gigabytesCount int) int64 {
	return int64((datasize.ByteSize(gigabytesCount) * datasize.GB).Bytes())
}

func KilobytesToBytes(kilobytes int) int64 {
	return int64((datasize.ByteSize(kilobytes) * datasize.KB).Bytes())
}

func BytesToKilobytes(bytes int64) int {
	return int(datasize.ByteSize(bytes).KBytes())
}

func ToBytesFromFloat(gigabytesCountF float64) int64 {
	return int64(gigabytesCountF * float64(datasize.GB))
}

func ParseDuration(s string) (*durationpb.Duration, error) {
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

func FormatDuration(d *durationpb.Duration) string {
	if d == nil {
		return ""
	}
	return d.AsDuration().String()
}

func GetTimestamp(timestamp *timestamppb.Timestamp) string {
	if timestamp == nil {
		return ""
	}
	return timestamp.AsTime().Format(defaultTimeFormat)
}

func ParseTimestamp(ts string) (*timestamppb.Timestamp, error) {
	t, err := time.Parse(defaultTimeFormat, ts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %s", err)
	}

	return timestamppb.New(t), nil
}

func GetEnumValueMapKeys(m map[string]int32) []string {
	return GetEnumValueMapKeysExt(m, false)
}

func GetEnumValueMapKeysExt(m map[string]int32, skipDefault bool) []string {
	keys := make([]string, 0, len(m))
	for k, v := range m {
		if v == 0 && skipDefault {
			continue
		}

		keys = append(keys, k)
	}
	return keys
}

func GetJoinedKeys(keys []string) string {
	return "`" + strings.Join(keys, "`, `") + "`"
}

func CheckOneOf(d *schema.ResourceData, keys ...string) error {
	var gotKey bool
	for _, key := range keys {
		_, ok := d.GetOk(key)

		if ok {
			if gotKey {
				return fmt.Errorf("only one of %s can be provided", GetJoinedKeys(keys))
			}

			gotKey = true
		}
	}

	if !gotKey {
		return fmt.Errorf("one of %s should be provided", GetJoinedKeys(keys))
	}

	return nil
}

type objectResolverFuncV2 func(name string, opts ...sdkresolversv2.ResolveOption) sdkresolversv2.Resolver

func ResolverWithClient[C any](client C, resolver func(string, C, ...sdkresolversv2.ResolveOption) sdkresolversv2.Resolver) objectResolverFuncV2 {
	return func(name string, opts ...sdkresolversv2.ResolveOption) sdkresolversv2.Resolver {
		return resolver(name, client, opts...)
	}
}

// this function can be only used to resolve objects that belong to some folder (have folder_id attribute)
// do not use this function to resolve cloud (or similar objects) ID by name.
func ResolveObjectIDByNameAndFolderIDV2(ctx context.Context, name, folderID string, resolverFunc objectResolverFuncV2) (string, error) {
	if name == "" {
		return "", fmt.Errorf("non empty name should be provided")
	}
	resolver := resolverFunc(name, sdkresolversv2.FolderID(folderID))
	if err := resolver.Run(ctx); err != nil {
		return "", err
	}
	if err := resolver.Err(); err != nil {
		return "", err
	}
	return resolver.ID(), nil
}

func ResolveObjectIDV2(ctx context.Context, config *providerconfig.Config, d *schema.ResourceData, resolverFunc objectResolverFuncV2) (string, error) {
	name, ok := d.GetOk("name")
	if !ok {
		return "", fmt.Errorf("non empty name should be provided")
	}

	folderID, err := GetFolderID(d, config)
	if err != nil {
		return "", err
	}

	return ResolveObjectIDByNameAndFolderIDV2(ctx, name.(string), folderID, resolverFunc)
}

func TemplateConfig(tmpl string, ctx ...map[string]interface{}) string {
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

func GetResourceID(n string, s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[n]
	if !ok {
		return "", fmt.Errorf("terraform resource '%s' not found", n)
	}

	if rs.Primary.ID == "" {
		return "", fmt.Errorf("no ID is set for terraform resource '%s'", n)
	}

	return rs.Primary.ID, nil
}

func ConvertToOptional(originalSchema map[string]*schema.Schema) map[string]*schema.Schema {
	optionalSchema := map[string]*schema.Schema{}
	for key, value := range originalSchema {
		newItem := *value
		newItem.Required = false
		newItem.ExactlyOneOf = []string{}
		newItem.Optional = true
		newItem.ForceNew = false

		switch newItem.Type {
		case schema.TypeList, schema.TypeSet:
			switch newItem.Elem.(type) {
			case *schema.Resource:
				elem := *newItem.Elem.(*schema.Resource)
				elem.Schema = ConvertToOptional(elem.Schema)
				newItem.Elem = &elem
			}
		}

		optionalSchema[key] = &newItem
	}
	return optionalSchema
}

func RemoveWriteOnlyFields(resourceSchema map[string]*schema.Schema) {
	for fieldName, fieldSchema := range resourceSchema {
		if nestedResource, ok := fieldSchema.Elem.(*schema.Resource); ok {
			RemoveWriteOnlyFields(nestedResource.Schema)
		}

		if !fieldSchema.WriteOnly {
			continue
		}

		delete(resourceSchema, fieldName)
		delete(resourceSchema, fieldName+"_version")
		for _, requiredFieldPath := range fieldSchema.RequiredWith {
			requiredFieldName := requiredFieldPath[strings.LastIndex(requiredFieldPath, ".")+1:]
			delete(resourceSchema, requiredFieldName)
		}
	}
}

func ConvertResourceToDataSource(resource *schema.Resource) *schema.Resource {
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

func FieldDeprecatedForAnother(deprecatedFieldName string, newFieldName string) string {
	return fmt.Sprintf("The '%s' field has been deprecated. Please use '%s' instead.", deprecatedFieldName, newFieldName)
}

func UseResourceInstead(deprecatedFieldName string, newResource string) string {
	return fmt.Sprintf("to manage %ss, please switch to using a separate resource type %s", deprecatedFieldName, newResource)
}

func GenerateFieldMasks(d *schema.ResourceData, fieldsMap map[string]string) []string {
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

func ConstructResourceId(clusterID string, resourceName string) string {
	return fmt.Sprintf("%s:%s", clusterID, resourceName)
}

func DeconstructResourceId(resourceID string) (string, string, error) {
	parts := strings.SplitN(resourceID, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource id format: %q", resourceID)
	}

	clusterID := parts[0]
	resourceName := parts[1]
	return clusterID, resourceName, nil
}

func ExpandEnum(keyName string, value string, enumValues map[string]int32) (*int32, error) {
	if val, ok := enumValues[value]; ok {
		return &val, nil
	} else {
		return nil, fmt.Errorf("value for '%s' must be one of %s, not `%s`",
			keyName, GetJoinedKeys(GetEnumValueMapKeys(enumValues)), value)
	}
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
