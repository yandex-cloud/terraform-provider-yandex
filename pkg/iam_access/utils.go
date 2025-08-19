package iam_access

import (
	"errors"
	"strings"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func MergeBindings(bindings []*access.AccessBinding) []*access.AccessBinding {
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
		m := CanonicalMember(b)
		bm[b.RoleId][m] = true
	}
	return bm
}

func CanonicalMember(b *access.AccessBinding) string {
	return b.Subject.Type + ":" + b.Subject.Id
}

func RoleToMembersList(role string, bindings []*access.AccessBinding) []string {
	var members []string

	for _, b := range bindings {
		if b.RoleId != role {
			continue
		}
		m := CanonicalMember(b)
		members = append(members, m)
	}
	return members
}

func RemoveRoleFromBindings(roleForRemove string, bindings []*access.AccessBinding) []*access.AccessBinding {
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

func CountBatches(size, batchSize int) int {
	iterations := size / batchSize
	if size%batchSize > 0 {
		iterations++
	}
	return iterations
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

func IsStatusWithCode(err error, code codes.Code) bool {
	grpcStatus, ok := status.FromError(err)
	check := ok && grpcStatus.Code() == code

	nestedErr := errors.Unwrap(err)
	if nestedErr != nil {
		return IsStatusWithCode(nestedErr, code) || check
	}
	return check
}
