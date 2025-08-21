package cloudregistry_ip_permission_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const cloudRegistryIPPermissionIDSuffix = "/ip_permission"

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccCloudRegistryIPPermission(t *testing.T) {

	var (
		registryName             = acctest.RandomWithPrefix("tf-registry")
		ipPermissionResourceName = "yandex_cloudregistry_registry_ip_permission.my_ip_permission"
	)

	t.Run("test update from only push to only pull", func(t *testing.T) {

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { test.AccPreCheck(t) },
			ProtoV6ProviderFactories: test.AccProviderFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckCloudRegistryDestroy,
				testAccCheckCloudRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, push, nil),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", "0"),
					),
				},
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, nil, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", "0"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
			},
		})
	})

	t.Run("test update from only pull to only push", func(t *testing.T) {

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { test.AccPreCheck(t) },
			ProtoV6ProviderFactories: test.AccProviderFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckCloudRegistryDestroy,
				testAccCheckCloudRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, nil, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckNoResourceAttr(ipPermissionResourceName, "push"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, push, nil),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", "0"),
					),
				},
			},
		})
	})

	t.Run("test update from only push to push + pull", func(t *testing.T) {

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { test.AccPreCheck(t) },
			ProtoV6ProviderFactories: test.AccProviderFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckCloudRegistryDestroy,
				testAccCheckCloudRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, push, nil),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckNoResourceAttr(ipPermissionResourceName, "pull"),
					),
				},
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, push, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
			},
		})
	})

	t.Run("test update from only pull to push + pull", func(t *testing.T) {

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { test.AccPreCheck(t) },
			ProtoV6ProviderFactories: test.AccProviderFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckCloudRegistryDestroy,
				testAccCheckCloudRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, nil, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckNoResourceAttr(ipPermissionResourceName, "push"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, push, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
			},
		})
	})

	t.Run("test update from push + pull to only push", func(t *testing.T) {

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { test.AccPreCheck(t) },
			ProtoV6ProviderFactories: test.AccProviderFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckCloudRegistryDestroy,
				testAccCheckCloudRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, push, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, push, nil),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", "0"),
					),
				},
			},
		})
	})

	t.Run("test update from push + pull to only pull", func(t *testing.T) {

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { test.AccPreCheck(t) },
			ProtoV6ProviderFactories: test.AccProviderFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckCloudRegistryDestroy,
				testAccCheckCloudRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, push, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, nil, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", "0"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
			},
		})
	})

	t.Run("taints and import", func(t *testing.T) {

		var (
			push = []string{"10.0.0.0/16", "10.0.1.0/16", "10.0.2.0/16"}
			pull = []string{"10.1.0.0/16", "10.1.2.0/16", "10.1.2.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { test.AccPreCheck(t) },
			ProtoV6ProviderFactories: test.AccProviderFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckCloudRegistryDestroy,
				testAccCheckCloudRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, push, pull),
				},

				// taint ip_permission (causes recreation of ip_permission)
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, push, pull),
					Taint:  []string{"yandex_cloudregistry_registry_ip_permission.my_ip_permission"},
				},

				// taint registry (causes recreation of registry, ip_permission)
				{
					Config: getAccResourceCloudRegistryIPPermissionConfig(registryName, push, pull),
					Taint:  []string{"yandex_cloudregistry_registry.my_registry"},
				},

				// import
				{
					ResourceName:      ipPermissionResourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func testAccCheckCloudRegistryIPPermissionDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_сloud_registry_ip_permission" {
			continue
		}

		CloudRegistryService := config.SDK.CloudRegistry().Registry()
		listIPPermissionRequest := &cloudregistry.ListIpPermissionsRequest{
			RegistryId: strings.TrimRight(rs.Primary.ID, cloudRegistryIPPermissionIDSuffix),
		}
		_, err := CloudRegistryService.ListIpPermissions(context.Background(), listIPPermissionRequest)
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}

			return fmt.Errorf("Cloud Registry still exists")
		}
	}

	return nil
}

func testAccCheckCloudRegistryDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_cloudregistry_registry" {
			continue
		}

		_, err := config.SDK.CloudRegistry().Registry().Get(context.Background(), &cloudregistry.GetRegistryRequest{
			RegistryId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("Cloud Registry still exists")
		}
	}

	return nil
}

func getAccResourceCloudRegistryIPPermissionConfig(registryName string, push, pull []string) string {
	if len(push) == 0 {
		return getAccResourceCloudRegistryIPPermissionRegistryConfig(registryName, "DOCKER", "LOCAL") + fmt.Sprintf(`
		resource "yandex_cloudregistry_registry_ip_permission" "my_ip_permission" {
			registry_id = yandex_cloudregistry_registry.my_registry.id
			pull        = [ %v ]
		}`, cloudRegistryIPPermissionCIDRSJoin(pull))
	}

	if len(pull) == 0 {
		return getAccResourceCloudRegistryIPPermissionRegistryConfig(registryName, "DOCKER", "LOCAL") + fmt.Sprintf(`
		resource "yandex_cloudregistry_registry_ip_permission" "my_ip_permission" {
			registry_id = yandex_cloudregistry_registry.my_registry.id
			push        = [ %v ]
		}`, cloudRegistryIPPermissionCIDRSJoin(push))
	}

	return getAccResourceCloudRegistryIPPermissionRegistryConfig(registryName, "DOCKER", "LOCAL") + fmt.Sprintf(`
		resource "yandex_cloudregistry_registry_ip_permission" "my_ip_permission" {
			registry_id = yandex_cloudregistry_registry.my_registry.id
			push        = [ %v ]
			pull        = [ %v ]
		}`,
		cloudRegistryIPPermissionCIDRSJoin(push),
		cloudRegistryIPPermissionCIDRSJoin(pull))
}

func getAccResourceCloudRegistryIPPermissionRegistryConfig(registryName, kind, typeName string) string {
	return fmt.Sprintf(`
		resource "yandex_cloudregistry_registry" "my_registry" {
			name = "%v"
			kind = "%s"
  			type = "%s"
		}`, registryName, kind, typeName)
}

func cloudRegistryIPPermissionCIDRSJoin(cidrs []string) string {
	if len(cidrs) > 0 {
		return stringifyCloudRegistryIPPermissionSlice(cidrs)
	}

	return ""
}

func stringifyCloudRegistryIPPermissionSlice(ipPermissions []string) string {
	return `"` + strings.Join(ipPermissions, `", "`) + `"`
}
