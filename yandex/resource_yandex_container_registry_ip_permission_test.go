package yandex

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccContainerRegistryIPPermission(t *testing.T) {
	t.Parallel()

	var (
		registryName             = acctest.RandomWithPrefix("tf-registry")
		ipPermissionResourceName = "yandex_container_registry_ip_permission.my_ip_permission"
	)

	t.Run("test update from only push to only pull", func(t *testing.T) {
		t.Parallel()

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckContainerRegistryDestroy,
				testAccCheckContainerRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, push, nil),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckNoResourceAttr(ipPermissionResourceName, "pull"),
					),
				},
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, nil, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckNoResourceAttr(ipPermissionResourceName, "push"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
			},
		})
	})

	t.Run("test update from only pull to only push", func(t *testing.T) {
		t.Parallel()

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckContainerRegistryDestroy,
				testAccCheckContainerRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, nil, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckNoResourceAttr(ipPermissionResourceName, "push"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, push, nil),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckNoResourceAttr(ipPermissionResourceName, "pull"),
					),
				},
			},
		})
	})

	t.Run("test update from only push to push + pull", func(t *testing.T) {
		t.Parallel()

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckContainerRegistryDestroy,
				testAccCheckContainerRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, push, nil),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckNoResourceAttr(ipPermissionResourceName, "pull"),
					),
				},
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, push, pull),
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
		t.Parallel()

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckContainerRegistryDestroy,
				testAccCheckContainerRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, nil, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckNoResourceAttr(ipPermissionResourceName, "push"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, push, pull),
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
		t.Parallel()

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckContainerRegistryDestroy,
				testAccCheckContainerRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, push, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, push, nil),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckNoResourceAttr(ipPermissionResourceName, "pull"),
					),
				},
			},
		})
	})

	t.Run("test update from push + pull to only pull", func(t *testing.T) {
		t.Parallel()

		var (
			push = []string{"10.0.0.0/16"}
			pull = []string{"10.1.0.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckContainerRegistryDestroy,
				testAccCheckContainerRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, push, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "push.*", push[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "push.#", fmt.Sprint(len(push))),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, nil, pull),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(ipPermissionResourceName, "id"),
						resource.TestCheckNoResourceAttr(ipPermissionResourceName, "push"),
						resource.TestCheckTypeSetElemAttr(ipPermissionResourceName, "pull.*", pull[0]),
						resource.TestCheckResourceAttr(ipPermissionResourceName, "pull.#", fmt.Sprint(len(pull))),
					),
				},
			},
		})
	})

	t.Run("taints and import", func(t *testing.T) {
		t.Parallel()

		var (
			push = []string{"10.0.0.0/16", "10.0.1.0/16", "10.0.2.0/16"}
			pull = []string{"10.1.0.0/16", "10.1.2.0/16", "10.1.2.0/16"}
		)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckContainerRegistryDestroy,
				testAccCheckContainerRegistryIPPermissionDestroy,
			),
			Steps: []resource.TestStep{
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, push, pull),
				},

				// taint ip_permission (causes recreation of ip_permission)
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, push, pull),
					Taint:  []string{"yandex_container_registry_ip_permission.my_ip_permission"},
				},

				// taint registry (causes recreation of registry, ip_permission)
				{
					Config: getAccResourceContainerRegistryIPPermissionConfig(registryName, push, pull),
					Taint:  []string{"yandex_container_registry.my_registry"},
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

// TODO: deadcode
//func importContainerRegistryIPPermissionID(registry *containerregistry.Registry) func(*terraform.State) (string, error) {
//	return func(s *terraform.State) (string, error) {
//		return registry.Id + containerRegistryIPPermissionIDSuffix, nil
//	}
//}

func testAccCheckContainerRegistryIPPermissionDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_container_registry_ip_permission" {
			continue
		}

		containerRegistryService := config.sdk.ContainerRegistry().Registry()
		listIPPermissionRequest := &containerregistry.ListIpPermissionRequest{
			// TODO: SA1024: cutset contains duplicate characters (staticcheck)
			RegistryId: strings.TrimRight(rs.Primary.ID, containerRegistryIPPermissionIDSuffix),
		}
		_, err := containerRegistryService.ListIpPermission(context.Background(), listIPPermissionRequest)
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}

			return fmt.Errorf("container Registry still exists")
		}
	}

	return nil
}

// createAccResourceContainerRegistryIPPermission(ipPermissionName string, ...permissions permission)

func getAccResourceContainerRegistryIPPermissionConfig(registryName string, push, pull []string) string {
	if len(push) == 0 {
		return getAccResourceContainerRegistryIPPermissionRegistryConfig(registryName) + fmt.Sprintf(`
		resource "yandex_container_registry_ip_permission" "my_ip_permission" {
			registry_id = yandex_container_registry.my_registry.id
			pull        = [ %v ]
		}`, containerRegistryIPPermissionCIDRSJoin(pull))
	}

	if len(pull) == 0 {
		return getAccResourceContainerRegistryIPPermissionRegistryConfig(registryName) + fmt.Sprintf(`
		resource "yandex_container_registry_ip_permission" "my_ip_permission" {
			registry_id = yandex_container_registry.my_registry.id
			push        = [ %v ]
		}`, containerRegistryIPPermissionCIDRSJoin(push))
	}

	return getAccResourceContainerRegistryIPPermissionRegistryConfig(registryName) + fmt.Sprintf(`
		resource "yandex_container_registry_ip_permission" "my_ip_permission" {
			registry_id = yandex_container_registry.my_registry.id
			push        = [ %v ]
			pull        = [ %v ]
		}`,
		containerRegistryIPPermissionCIDRSJoin(push),
		containerRegistryIPPermissionCIDRSJoin(pull))
}

func getAccResourceContainerRegistryIPPermissionRegistryConfig(registryName string) string {
	return fmt.Sprintf(`
		resource "yandex_container_registry" "my_registry" {
			name = "%v"
		}`, registryName)
}

func containerRegistryIPPermissionCIDRSJoin(cidrs []string) string {
	if len(cidrs) > 0 {
		return stringifyContainerRegistryIPPermissionSlice(cidrs)
	}

	return ""
}
