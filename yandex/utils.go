package yandex

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/pkg/requestid"
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
	folder, err := config.sdk.ResourceManager().Folder().Get(config.ContextWithClientTraceID(), &resourcemanager.GetFolderRequest{
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

func handleNotFoundError(err error, d *schema.ResourceData, resourceName string) error {
	if isStatusWithCode(err, codes.NotFound) {
		log.Printf("[WARN] Removing %s because resource doesn't exist anymore", resourceName)
		d.SetId("")
		return nil
	}
	return fmt.Errorf("Error reading %s: %s", resourceName, err)
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

func validateIPV4CidrBlocks(v interface{}, k string) (warnings []string, errors []error) {
	_, _, err := net.ParseCIDR(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is not a valid IP CIDR range: %s", k, err))
	}
	return
}

// parseFunc should take exactly one argument of the type specified in the schema
// and return an error as its last return value
func validateParsableValue(parseFunc interface{}) schema.SchemaValidateFunc {
	return func(value interface{}, key string) (warnings []string, errors []error) {
		tryCall := func() (vs []reflect.Value, err error) {
			defer func() {
				if p := recover(); p != nil {
					err = fmt.Errorf("could not call parse function: %v", p)
				}
			}()

			vs = reflect.ValueOf(parseFunc).Call([]reflect.Value{reflect.ValueOf(value)})
			return
		}

		vs, err := tryCall()
		if err != nil {
			errors = append(errors, err)
			return
		}

		if len(vs) == 0 {
			errors = append(errors, fmt.Errorf("expected parse function to return at least one value"))
			return
		}

		last := vs[len(vs)-1]
		if last.Kind() == reflect.Interface {
			err, ok := last.Interface().(error)
			if ok || last.IsNil() {
				if err != nil {
					errors = append(errors, err)
				}
				return
			}
		}
		errors = append(errors, fmt.Errorf("expected parse function's last return value to be an error"))
		return
	}
}

// FloatAtLeast returns a SchemaValidateFunc which tests if the provided value
// is of type float64 and is at least min (inclusive)
func FloatAtLeast(min float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (_ []string, errors []error) {
		v, ok := i.(float64)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be float64", k))
			return nil, errors
		}

		if v < min {
			errors = append(errors, fmt.Errorf("expected %s to be at least (%f), got %f", k, min, v))
			return nil, errors
		}

		return nil, errors
	}
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

func getTimestamp(protots *timestamp.Timestamp) (string, error) {
	ts, err := ptypes.Timestamp(protots)
	if err != nil {
		return "", fmt.Errorf("failed to convert protobuf timestamp: %s", err)
	}

	return ts.Format(defaultTimeFormat), nil
}

func getEnumValueMapKeys(m map[string]int32) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
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

func resolveObjectID(ctx context.Context, config *Config, name string, resolverFunc objectResolverFunc) (string, error) {
	if name == "" {
		return "", fmt.Errorf("non empty name should be provided")
	}

	var objectID string
	resolver := resolverFunc(name, sdkresolvers.Out(&objectID), sdkresolvers.FolderID(config.FolderID))

	err := config.sdk.Resolve(ctx, resolver)

	if err != nil {
		return "", err
	}

	return objectID, nil
}

func getSnapshotMinStorageSize(snapshotID string, config *Config) (size int64, err error) {
	ctx := config.ContextWithClientTraceID()

	snapshot, err := config.sdk.Compute().Snapshot().Get(ctx, &compute.GetSnapshotRequest{
		SnapshotId: snapshotID,
	})

	if err != nil {
		return 0, fmt.Errorf("Error on retrieve snapshot properties: %s", err)
	}

	return snapshot.DiskSize, nil
}

func getImageMinStorageSize(imageID string, config *Config) (size int64, err error) {
	ctx := config.ContextWithClientTraceID()

	image, err := config.sdk.Compute().Image().Get(ctx, &compute.GetImageRequest{
		ImageId: imageID,
	})

	if err != nil {
		return 0, fmt.Errorf("Error on retrieve image properties: %s", err)
	}

	return image.MinDiskSize, nil
}

func contextWithClientTraceID(parent context.Context) context.Context {
	return requestid.ContextWithClientTraceID(parent, uuid.New().String())
}
