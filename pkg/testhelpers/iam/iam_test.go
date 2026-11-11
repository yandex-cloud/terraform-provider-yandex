package iam

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"google.golang.org/grpc"
)

func TestGetRoleMembersClientError(t *testing.T) {
	expected := errors.New("provider is not configured")
	members, err := getRoleMembers(context.Background(), func() (BindingsGetter, error) { return nil, expected }, "resource-id", "viewer")
	if members != nil || !errors.Is(err, expected) {
		t.Fatalf("getRoleMembers() = %v, %v; want provider configuration error", members, err)
	}
}

type bindingPages struct {
	t        *testing.T
	calls    []string
	failPage string
	err      error
}

func (c *bindingPages) ListAccessBindings(_ context.Context, in *access.ListAccessBindingsRequest, _ ...grpc.CallOption) (*access.ListAccessBindingsResponse, error) {
	if in.ResourceId != "resource-id" {
		c.t.Fatalf("resource id = %q; want resource-id", in.ResourceId)
	}
	c.calls = append(c.calls, in.PageToken)
	if in.PageToken == c.failPage && c.err != nil {
		return nil, c.err
	}
	switch in.PageToken {
	case "":
		return &access.ListAccessBindingsResponse{
			AccessBindings: []*access.AccessBinding{
				{RoleId: "editor", Subject: &access.Subject{Type: "userAccount", Id: "ignored"}},
				{RoleId: "viewer", Subject: &access.Subject{Type: "userAccount", Id: "first"}},
			},
			NextPageToken: "second-page",
		}, nil
	case "second-page":
		return &access.ListAccessBindingsResponse{
			AccessBindings: []*access.AccessBinding{
				{RoleId: "viewer", Subject: &access.Subject{Type: "userAccount", Id: "second"}},
				{RoleId: "later-viewer", Subject: &access.Subject{Type: "userAccount", Id: "second"}},
			},
		}, nil
	default:
		c.t.Fatalf("unexpected page token %q", in.PageToken)
		return nil, nil
	}
}

func TestGetRoleMembersPagination(t *testing.T) {
	client := &bindingPages{t: t}
	factory := func() (BindingsGetter, error) { return client, nil }
	got, err := getRoleMembers(context.Background(), factory, "resource-id", "viewer")
	want := []string{"userAccount:first", "userAccount:second"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("getRoleMembers() = %v, %v; want %v", got, err, want)
	}
	if !reflect.DeepEqual(client.calls, []string{"", "second-page"}) {
		t.Fatalf("requested pages = %v", client.calls)
	}
}

func TestGetRoleMembersLaterPageError(t *testing.T) {
	expected := errors.New("access denied on second page")
	client := &bindingPages{t: t, failPage: "second-page", err: expected}
	members, err := getRoleMembers(context.Background(), func() (BindingsGetter, error) { return client, nil }, "resource-id", "viewer")
	if members != nil || !errors.Is(err, expected) {
		t.Fatalf("getRoleMembers() = %v, %v; want second-page error", members, err)
	}
}

type bindingResourceID string

func (id bindingResourceID) GetId() string { return string(id) }

func TestIAMChecksUseAllPages(t *testing.T) {
	ctx := context.Background()
	id := bindingResourceID("resource-id")
	tests := []struct {
		name      string
		run       func(ClientFactory) error
		wantError bool
	}{
		{"equals includes second page", func(factory ClientFactory) error {
			return TestAccCheckIamBindingEqualsMembers(ctx, factory, id, "viewer", []string{"userAccount:first", "userAccount:second"})(nil)
		}, false},
		{"equals detects extra member on second page", func(factory ClientFactory) error {
			return TestAccCheckIamBindingEqualsMembers(ctx, factory, id, "viewer", []string{"userAccount:first"})(nil)
		}, true},
		{"contains finds second page", func(factory ClientFactory) error {
			return TestAccCheckIamBindingContainsMembers(ctx, factory, id, "viewer", []string{"userAccount:second"})(nil)
		}, false},
		{"empty detects member on later page", func(factory ClientFactory) error {
			return TestAccCheckIamBindingEmpty(ctx, factory, id, "later-viewer")(nil)
		}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &bindingPages{t: t}
			err := tc.run(func() (BindingsGetter, error) { return client, nil })
			if (err != nil) != tc.wantError {
				t.Fatalf("check error = %v; wantError %v", err, tc.wantError)
			}
		})
	}
}

func TestIAMEqualsDoesNotModifyExpected(t *testing.T) {
	expected := []string{"userAccount:second", "userAccount:first"}
	original := append([]string(nil), expected...)
	client := &bindingPages{t: t}
	check := TestAccCheckIamBindingEqualsMembers(context.Background(), func() (BindingsGetter, error) { return client, nil }, bindingResourceID("resource-id"), "viewer", expected)
	if err := check(nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, original) {
		t.Fatalf("expected was mutated: %v; want %v", expected, original)
	}
}
