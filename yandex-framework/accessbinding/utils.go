package accessbinding

import (
	"fmt"
	"strings"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

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

func canonicalMember(ab *access.AccessBinding) string {
	return ab.Subject.Type + ":" + ab.Subject.Id
}

func CountBatches(size, batchSize int) int {
	iterations := size / batchSize
	if size%batchSize > 0 {
		iterations++
	}
	return iterations
}
