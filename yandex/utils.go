package yandex

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

type instanceAction int

const (
	instanceActionStop instanceAction = iota
	instanceActionStart
	instanceActionRestart
)

const defaultTimeFormat = time.RFC3339

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

func expandLabels(v interface{}) (map[string]string, error) {
	m := make(map[string]string)
	if v == nil {
		return m, nil
	}
	for k, val := range v.(map[string]interface{}) {
		m[k] = val.(string)
	}
	return m, nil
}

func expandProductIds(v interface{}) ([]string, error) {
	m := []string{}
	if v == nil {
		return m, nil
	}
	tagsSet := v.(*schema.Set)
	for _, val := range tagsSet.List() {
		m = append(m, val.(string))
	}
	return m, nil
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

// Primary use to store value from API in state file as Gigabytes
func toGigabytes(bytesCount int64) int {
	return int((datasize.ByteSize(bytesCount) * datasize.B).GBytes())
}

// Primary use to send byte value to API
func toBytes(gigabytesCount int) int64 {
	return int64((datasize.ByteSize(gigabytesCount) * datasize.GB).Bytes())
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
