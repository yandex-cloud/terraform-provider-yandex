package cloudregistry_ip_permission_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceCloudRegistryIPPermission(t *testing.T) {

	const dataCloudRegistryIPPermissionName = "data.yandex_cloudregistry_registry_ip_permission.my_ip_permission"

	var (
		registryName = acctest.RandomWithPrefix("tf-registry")
		push         = []string{"10.0.0.0/16"}
		pull         = []string{"10.1.0.0/16"}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckCloudRegistryDestroy,
			testAccCheckCloudRegistryIPPermissionDestroy,
		),
		Steps: []resource.TestStep{
			// by Name
			{
				Config: getAccDataCloudRegistryIPPermissionConfigByName(registryName, "DOCKER", "LOCAL", push, pull),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataCloudRegistryIPPermissionName, "id"),
					resource.TestCheckTypeSetElemAttr(dataCloudRegistryIPPermissionName, "push.*", push[0]),
					resource.TestCheckResourceAttr(dataCloudRegistryIPPermissionName, "push.#", fmt.Sprint(len(push))),
					resource.TestCheckTypeSetElemAttr(dataCloudRegistryIPPermissionName, "pull.*", pull[0]),
					resource.TestCheckResourceAttr(dataCloudRegistryIPPermissionName, "pull.#", fmt.Sprint(len(pull))),
				),
			},

			// by ID
			{
				Config: getAccDataCloudRegistryIPPermissionConfigByID(registryName, "DOCKER", "LOCAL", push, pull),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataCloudRegistryIPPermissionName, "id"),
					resource.TestCheckTypeSetElemAttr(dataCloudRegistryIPPermissionName, "push.*", push[0]),
					resource.TestCheckResourceAttr(dataCloudRegistryIPPermissionName, "push.#", fmt.Sprint(len(push))),
					resource.TestCheckTypeSetElemAttr(dataCloudRegistryIPPermissionName, "pull.*", pull[0]),
					resource.TestCheckResourceAttr(dataCloudRegistryIPPermissionName, "pull.#", fmt.Sprint(len(pull))),
				),
			},
		},
	})
}

func getAccDataCloudRegistryIPPermissionConfigByName(registryName, kind, typeName string, push, pull []string) string {
	return getAccDataCloudRegistryIPPermissionConfig(registryName, kind, typeName, push, pull) + `
	data "yandex_cloudregistry_registry_ip_permission" "my_ip_permission" {
		registry_name = yandex_cloudregistry_registry.my_registry.name
		# registry_name adds dependency only to registry, not ip_permission
		depends_on = [
			yandex_cloudregistry_registry_ip_permission.my_ip_permission
		]
	}`
}

func getAccDataCloudRegistryIPPermissionConfigByID(registryName, kind, typeName string, push, pull []string) string {
	return getAccDataCloudRegistryIPPermissionConfig(registryName, kind, typeName, push, pull) + `
	data "yandex_cloudregistry_registry_ip_permission" "my_ip_permission" {
		registry_id = yandex_cloudregistry_registry.my_registry.id
		# registry_id adds dependency only to registry, not ip_permission
		depends_on = [
			yandex_cloudregistry_registry_ip_permission.my_ip_permission
		]
	}`
}

func getAccDataCloudRegistryIPPermissionConfig(registryName, kind, typeName string, push, pull []string) string {
	return getAccDataCloudRegistryIPPermissionRegistryConfig(registryName, kind, typeName) + fmt.Sprintf(`
		resource "yandex_cloudregistry_registry_ip_permission" "my_ip_permission" {
			registry_id = yandex_cloudregistry_registry.my_registry.id
			push        = [ %v ]
			pull        = [ %v ]
		}`,
		newCloudRegistryIPPermissionCIDRSJoin(push),
		newCloudRegistryIPPermissionCIDRSJoin(pull))
}

func getAccDataCloudRegistryIPPermissionRegistryConfig(registryName, kind, typeName string) string {
	return fmt.Sprintf(`
		resource "yandex_cloudregistry_registry" "my_registry" {
			name = "%v"
			kind = "%s"
            type = "%s"
		}`, registryName, kind, typeName)
}

func newCloudRegistryIPPermissionCIDRSJoin(cidrs []string) string {
	if len(cidrs) > 0 {
		return stringifyCloudRegistryIPPermissionSlice(cidrs)
	}

	return ""
}
