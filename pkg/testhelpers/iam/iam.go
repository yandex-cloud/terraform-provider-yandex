package iam

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"sort"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"google.golang.org/grpc"
)

type SDKGetter func() BindingsGetter

type BindingsGetter interface {
	ListAccessBindings(ctx context.Context, in *access.ListAccessBindingsRequest, opts ...grpc.CallOption) (*access.ListAccessBindingsResponse, error)
}

type idGetter interface {
	GetId() string
}

func TestAccCheckIamBindingEqualsMembers(ctx context.Context, sdkGetter SDKGetter, idGetter idGetter, role string, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(expected) == 0 {
			return fmt.Errorf("expected members list is empty for role %q; use TestAccCheckIamBindingEmpty instead", role)
		}

		id := idGetter.GetId()

		actual, err := getRoleMembers(ctx, sdkGetter, id, role)
		if err != nil {
			return err
		}

		sort.Strings(expected)
		sort.Strings(actual)

		if reflect.DeepEqual(expected, actual) {
			return nil
		}

		return fmt.Errorf("members mismatch for role %q on resource %q: expected %v, got %v", role, id, expected, actual)
	}
}

func TestAccCheckIamBindingContainsMembers(ctx context.Context, sdkGetter SDKGetter, idGetter idGetter, role string, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(expected) == 0 {
			return fmt.Errorf("expected members list is empty for role %q; use TestAccCheckIamBindingEmpty instead", role)
		}

		id := idGetter.GetId()

		actual, err := getRoleMembers(ctx, sdkGetter, id, role)
		if err != nil {
			return err
		}

		for _, member := range expected {
			if !slices.Contains(actual, member) {
				return fmt.Errorf("member %q not found for role %q on resource %q; actual members: %v", member, role, id, actual)
			}
		}

		return nil
	}
}

func TestAccCheckIamBindingEmpty(ctx context.Context, sdkGetter SDKGetter, idGetter idGetter, role string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := idGetter.GetId()

		members, err := getRoleMembers(ctx, sdkGetter, id, role)
		if err != nil {
			return err
		}

		if len(members) == 0 {
			return nil
		}

		return fmt.Errorf("expected no access bindings for role %q on resource %q, but found: %v", role, id, members)
	}
}

func getRoleMembers(ctx context.Context, sdkGetter SDKGetter, resourceID string, role string) ([]string, error) {
	bindings, err := sdkGetter().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
		ResourceId: resourceID,
	})
	if err != nil {
		return nil, fmt.Errorf("get access bindings for resource id(%s): %w", resourceID, err)
	}

	var members []string
	for _, binding := range bindings.AccessBindings {
		if binding.RoleId == role {
			members = append(members, binding.Subject.Type+":"+binding.Subject.Id)
		}
	}

	return members, nil
}

func IAMBindingImportTestStep(resourceName string, idGetter idGetter, role, identifierAttribute string) resource.TestStep {
	return resource.TestStep{
		ResourceName:                         resourceName,
		ImportStateIdFunc:                    importIAMBindingIDFunc(idGetter, role),
		ImportState:                          true,
		ImportStateVerifyIdentifierAttribute: identifierAttribute,
	}
}

func importIAMBindingIDFunc(idGetter idGetter, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return idGetter.GetId() + "," + role, nil
	}
}
