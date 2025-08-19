package testhelpers

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
	cm1sdk "github.com/yandex-cloud/go-sdk/services/certificatemanager/v1"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

func TestAccCheckCMCertificateExists(name string, certificate *certificatemanager.Certificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.Certificates().Certificate().Get(ctx, &certificatemanager.GetCertificateRequest{
			CertificateId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("certificate not found")
		}

		*certificate = *found

		return nil
	}
}

func TestAccCheckCMCertificateIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		bindings, err := getCMResourceAccessBindings(ctx, s, resourceName)
		if err != nil {
			return err
		}

		var roleMembers []string
		for _, binding := range bindings {
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

func TestAccCheckCMEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		bindings, err := getCMResourceAccessBindings(ctx, s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("binding found but expected empty for %s", resourceName)
	}
}

func getCMResourceAccessBindings(ctx context.Context, s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := AccProvider.(*yandex_framework.Provider).GetConfig()

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	var bindings []*access.AccessBinding
	pageToken := ""

	for {
		resp, err := cm1sdk.NewCertificateClient(config.SDKv2).ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: rs.Primary.ID,
			PageSize:   1000,
			PageToken:  pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error listing access bindings for certificate %s: %w", rs.Primary.ID, err)
		}
		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return bindings, nil
}
