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

type ClientFactory func() (BindingsGetter, error)

type BindingsGetter interface {
	ListAccessBindings(ctx context.Context, in *access.ListAccessBindingsRequest, opts ...grpc.CallOption) (*access.ListAccessBindingsResponse, error)
}

type idGetter interface {
	GetId() string
}

func TestAccCheckIamBindingEqualsMembers(ctx context.Context, clientForTest ClientFactory, idGetter idGetter, role string, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(expected) == 0 {
			return fmt.Errorf("expected members list is empty for role %q; use TestAccCheckIamBindingEmpty instead", role)
		}

		id := idGetter.GetId()

		actual, err := getRoleMembers(ctx, clientForTest, id, role)
		if err != nil {
			return err
		}

		expected := slices.Clone(expected)
		sort.Strings(expected)
		sort.Strings(actual)

		if reflect.DeepEqual(expected, actual) {
			return nil
		}

		return fmt.Errorf("members mismatch for role %q on resource %q: expected %v, got %v", role, id, expected, actual)
	}
}

func TestAccCheckIamBindingContainsMembers(ctx context.Context, clientForTest ClientFactory, idGetter idGetter, role string, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(expected) == 0 {
			return fmt.Errorf("expected members list is empty for role %q; use TestAccCheckIamBindingEmpty instead", role)
		}

		id := idGetter.GetId()

		actual, err := getRoleMembers(ctx, clientForTest, id, role)
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

func TestAccCheckIamBindingEmpty(ctx context.Context, clientForTest ClientFactory, idGetter idGetter, role string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := idGetter.GetId()

		members, err := getRoleMembers(ctx, clientForTest, id, role)
		if err != nil {
			return err
		}

		if len(members) == 0 {
			return nil
		}

		return fmt.Errorf("expected no access bindings for role %q on resource %q, but found: %v", role, id, members)
	}
}

func getRoleMembers(ctx context.Context, clientForTest ClientFactory, resourceID string, role string) ([]string, error) {
	client, err := clientForTest()
	if err != nil {
		return nil, fmt.Errorf("get IAM client: %w", err)
	}
	var members []string
	var pageToken string
	for {
		bindings, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: resourceID,
			PageToken:  pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("get access bindings for resource id(%s): %w", resourceID, err)
		}

		for _, binding := range bindings.AccessBindings {
			if binding.RoleId == role {
				members = append(members, binding.Subject.Type+":"+binding.Subject.Id)
			}
		}
		pageToken = bindings.NextPageToken
		if pageToken == "" {
			break
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
