package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceContainerRegistryIPPermission(t *testing.T) {
	t.Parallel()

	const dataContainerRegistryIPPermissionName = "data.yandex_container_registry_ip_permission.my_ip_permission"

	var (
		registryName = acctest.RandomWithPrefix("tf-registry")
		push         = []string{"10.0.0.0/16"}
		pull         = []string{"10.1.0.0/16"}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckContainerRegistryDestroy,
			testAccCheckContainerRegistryIPPermissionDestroy,
		),
		Steps: []resource.TestStep{
			// by Name
			{
				Config: getAccDataContainerRegistryIPPermissionConfigByName(registryName, push, pull),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataContainerRegistryIPPermissionName, "id"),
					resource.TestCheckTypeSetElemAttr(dataContainerRegistryIPPermissionName, "push.*", push[0]),
					resource.TestCheckResourceAttr(dataContainerRegistryIPPermissionName, "push.#", fmt.Sprint(len(push))),
					resource.TestCheckTypeSetElemAttr(dataContainerRegistryIPPermissionName, "pull.*", pull[0]),
					resource.TestCheckResourceAttr(dataContainerRegistryIPPermissionName, "pull.#", fmt.Sprint(len(pull))),
				),
			},

			// by ID
			{
				Config: getAccDataContainerRegistryIPPermissionConfigByID(registryName, push, pull),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataContainerRegistryIPPermissionName, "id"),
					resource.TestCheckTypeSetElemAttr(dataContainerRegistryIPPermissionName, "push.*", push[0]),
					resource.TestCheckResourceAttr(dataContainerRegistryIPPermissionName, "push.#", fmt.Sprint(len(push))),
					resource.TestCheckTypeSetElemAttr(dataContainerRegistryIPPermissionName, "pull.*", pull[0]),
					resource.TestCheckResourceAttr(dataContainerRegistryIPPermissionName, "pull.#", fmt.Sprint(len(pull))),
				),
			},
		},
	})
}

func getAccDataContainerRegistryIPPermissionConfigByName(registryName string, push, pull []string) string {
	return getAccDataContainerRegistryIPPermissionConfig(registryName, push, pull) + `
	data "yandex_container_registry_ip_permission" "my_ip_permission" {
		registry_name = yandex_container_registry.my_registry.name

		# registry_name adds dependency only to registry, not ip_permission
		depends_on = [
			yandex_container_registry_ip_permission.my_ip_permission
		]
	}`
}

func getAccDataContainerRegistryIPPermissionConfigByID(registryName string, push, pull []string) string {
	return getAccDataContainerRegistryIPPermissionConfig(registryName, push, pull) + `
	data "yandex_container_registry_ip_permission" "my_ip_permission" {
		registry_id = yandex_container_registry.my_registry.id

		# registry_id adds dependency only to registry, not ip_permission
		depends_on = [
			yandex_container_registry_ip_permission.my_ip_permission
		]
	}`
}

func getAccDataContainerRegistryIPPermissionConfig(registryName string, push, pull []string) string {
	return getAccDataContainerRegistryIPPermissionRegistryConfig(registryName) + fmt.Sprintf(`
		resource "yandex_container_registry_ip_permission" "my_ip_permission" {
			registry_id = yandex_container_registry.my_registry.id
			push        = [ %v ]
			pull        = [ %v ]
		}`,
		containerRegistryIPPermissionCIDRSJoin(push),
		containerRegistryIPPermissionCIDRSJoin(pull))
}

func getAccDataContainerRegistryIPPermissionRegistryConfig(registryName string) string {
	return fmt.Sprintf(`
		resource "yandex_container_registry" "my_registry" {
			name = "%v"
		}`, registryName)
}
