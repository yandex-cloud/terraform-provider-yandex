package iam

import (
	"context"
	"fmt"
	"reflect"
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

func TestAccCheckIamBindingExists(ctx context.Context, sdkGetter SDKGetter, idGetter idGetter, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := idGetter.GetId()

		bindings, err := sdkGetter().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: id,
		})
		if err != nil {
			return fmt.Errorf("get access bindings for resource id(%s):  %w", id, err)
		}

		var roleMembers []string
		for _, binding := range bindings.AccessBindings {
			if binding.RoleId == role {
				member := binding.Subject.Type + ":" + binding.Subject.Id
				roleMembers = append(roleMembers, member)
			}
		}
		sort.Strings(members)
		sort.Strings(roleMembers)

		if reflect.DeepEqual(members, roleMembers) {
			return nil
		}

		return fmt.Errorf("binding found but expected members is %v, got %v", members, roleMembers)
	}
}
