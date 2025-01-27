package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NOTE:
//	* there is no sweeper because Lifecycle Policy is a child of Container Repository which
//	  in turn is a child of Container Registry for what sweeper is already exist.

func TestAccContainerRepositoryLifecyclePolicy(t *testing.T) {
	t.Parallel()

	t.Run("test everything step by step", func(t *testing.T) {
		var (
			registryName        = acctest.RandomWithPrefix("tf-registry")
			repositoryName      = acctest.RandomWithPrefix("tf-repository")
			lifecyclePolicyName = acctest.RandomWithPrefix("tf-lifecycle-policy")
		)

		const lifecyclePolicyResourceName = "yandex_container_repository_lifecycle_policy.my_lifecycle_policy"

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckContainerRegistryDestroy,
				testAccCheckContainerRepositoryDestroy,
				testAccCheckContainerRepositoryLifecyclePolicyDestroy,
			),
			Steps: []resource.TestStep{
				// create registry, repository, lifecycle policy
				{
					Config: getAccResourceContainerRepositoryLifecyclePolicyConfig(registryName, repositoryName, lifecyclePolicyName),
					Check: resource.ComposeTestCheckFunc(
						/*
							yandex_container_repository_lifecycle_policy.my_lifecycle_policy

							resource "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
								name          = "%v"
								status        = "disabled"
								repository_id = yandex_container_repository.my_repository.id
							}
						*/

						resource.TestCheckResourceAttrSet(lifecyclePolicyResourceName, "id"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "status", "disabled"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "name", lifecyclePolicyName),
						testAccCheckCreatedAtAttr(lifecyclePolicyResourceName),
					),
				},

				// update lifecycle policy fields
				{
					Config: getAccResourceContainerRepositoryLifecyclePolicyConfigUpdated(registryName, repositoryName, lifecyclePolicyName),
					Check: resource.ComposeTestCheckFunc(
						/*
							yandex_container_repository_lifecycle_policy.my_lifecycle_policy

							resource "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
								name          = "%v"
								status        = "active"
								repository_id = yandex_container_repository.my_repository.id
								description   = "my description"
							}
						*/

						resource.TestCheckResourceAttrSet(lifecyclePolicyResourceName, "id"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "status", "active"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "name", lifecyclePolicyName),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "description", "my description"),
						testAccCheckCreatedAtAttr(lifecyclePolicyResourceName),
					),
				},

				// add one rule
				{
					Config: getAccResourceContainerRepositoryLifecyclePolicyConfigWithOneRule(registryName, repositoryName, lifecyclePolicyName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(lifecyclePolicyResourceName, "id"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "status", "active"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "name", lifecyclePolicyName),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "description", "my description"),

						/*
							yandex_container_repository_lifecycle_policy.my_lifecycle_policy.rule.0

							rule {
								description  = "untagged and retained top 1"
								untagged     = true
								retained_top = 1
							}
						*/

						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.description", "untagged and retained top 1"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.untagged", "true"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.retained_top", "1"),
						testAccCheckCreatedAtAttr(lifecyclePolicyResourceName),
					),
				},

				// update rule
				{
					Config: getAccResourceContainerRepositoryLifecyclePolicyConfigWithOneRuleUpdated(registryName, repositoryName, lifecyclePolicyName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(lifecyclePolicyResourceName, "id"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "status", "active"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "name", lifecyclePolicyName),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "description", "my description"),
						testAccCheckCreatedAtAttr(lifecyclePolicyResourceName),

						/*
							yandex_container_repository_lifecycle_policy.my_lifecycle_policy.rule.0

							rule {
								description   = "everything is complicated"
								untagged      = true
								tag_regexp    = ".name*"
								expire_period = "24h"
								retained_top  = 10
							}
						*/

						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.description", "everything is complicated"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.untagged", "true"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.tag_regexp", ".name*"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.expire_period", "24h"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.retained_top", "10"),
					),
				},

				// add two more rules
				{
					Config: getAccResourceContainerRepositoryLifecyclePolicyConfigWithThreeRules(registryName, repositoryName, lifecyclePolicyName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(lifecyclePolicyResourceName, "id"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "status", "active"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "name", lifecyclePolicyName),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "description", "my description"),
						testAccCheckCreatedAtAttr(lifecyclePolicyResourceName),

						/*
							yandex_container_repository_lifecycle_policy.my_lifecycle_policy.rule.*

							rule {
								description   = "everything is complicated"
								untagged      = true
								tag_regexp    = ".name*"
								expire_period = "24h"
								retained_top  = 10
							}
							rule {
								description   = "untagged and retained top 5"
								untagged      = true
								retained_top  = 5
							}
							rule {
								description   = "regexped and retained top 1"
								untagged      = false
								tag_regexp    = ".name*"
								retained_top  = 1
							}
						*/

						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.description", "everything is complicated"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.untagged", "true"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.tag_regexp", ".name*"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.expire_period", "24h0m0s"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.0.retained_top", "10"),

						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.1.description", "untagged and retained top 5"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.1.untagged", "true"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.1.retained_top", "5"),

						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.2.description", "regexped and retained top 1"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.2.untagged", "false"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.2.tag_regexp", ".name*"),
						resource.TestCheckResourceAttr(lifecyclePolicyResourceName, "rule.2.retained_top", "1"),
					),
				},

				// shuffle rules (no changes expected)
				{
					Config:             getAccResourceContainerRepositoryLifecyclePolicyConfigWithThreeRulesShuffled(registryName, repositoryName, lifecyclePolicyName),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},

				// taint lifecycle policy (causes recreation of lifecycle policy)
				{
					Config: getAccResourceContainerRepositoryLifecyclePolicyConfigWithThreeRules(registryName, repositoryName, lifecyclePolicyName),
					Taint:  []string{"yandex_container_repository_lifecycle_policy.my_lifecycle_policy"},
				},

				// taint repository (causes recreation of repository, lifecycle policy)
				{
					Config: getAccResourceContainerRepositoryLifecyclePolicyConfigWithThreeRules(registryName, repositoryName, lifecyclePolicyName),
					Taint:  []string{"yandex_container_repository.my_repository"},
				},

				// taint registry (causes recreation of registry, repository, lifecycle policy)
				{
					Config: getAccResourceContainerRepositoryLifecyclePolicyConfigWithThreeRules(registryName, repositoryName, lifecyclePolicyName),
					Taint:  []string{"yandex_container_registry.my_registry"},
				},

				// import
				{
					ResourceName:      lifecyclePolicyResourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func getAccResourceContainerRepositoryLifecyclePolicyConfig(registryName, repositoryName, lifecyclePolicyName string) string {
	return fmt.Sprintf(`
		resource "yandex_container_registry" "my_registry" {
			name = "%v"
		}
		
		resource "yandex_container_repository" "my_repository" {
			name = "${yandex_container_registry.my_registry.id}/%v"
		}
		
		resource "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
			name          = "%v"
			status        = "disabled"
			repository_id = yandex_container_repository.my_repository.id
		}`, repositoryName, registryName, lifecyclePolicyName)
}

func getAccResourceContainerRepositoryLifecyclePolicyConfigUpdated(registryName, repositoryName, lifecyclePolicyName string) string {
	return fmt.Sprintf(`
		resource "yandex_container_registry" "my_registry" {
			name = "%v"
		}
		
		resource "yandex_container_repository" "my_repository" {
			name = "${yandex_container_registry.my_registry.id}/%v"
		}
		
		resource "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
			name          = "%v"
			status        = "active"
			repository_id = yandex_container_repository.my_repository.id
			description   = "my description"
		}`, repositoryName, registryName, lifecyclePolicyName)
}

func getAccResourceContainerRepositoryLifecyclePolicyConfigWithOneRule(registryName, repositoryName, lifecyclePolicyName string) string {
	return fmt.Sprintf(`
		resource "yandex_container_registry" "my_registry" {
			name = "%v"
		}
		
		resource "yandex_container_repository" "my_repository" {
			name = "${yandex_container_registry.my_registry.id}/%v"
		}
		
		resource "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
			name          = "%v"
			status        = "active"
			repository_id = yandex_container_repository.my_repository.id
			description   = "my description"

			rule {
				description  = "untagged and retained top 1"
				untagged     = true
				retained_top = 1
			}
		}`, repositoryName, registryName, lifecyclePolicyName)
}

func getAccResourceContainerRepositoryLifecyclePolicyConfigWithOneRuleUpdated(registryName, repositoryName, lifecyclePolicyName string) string {
	return fmt.Sprintf(`
		resource "yandex_container_registry" "my_registry" {
			name = "%v"
		}
		
		resource "yandex_container_repository" "my_repository" {
			name = "${yandex_container_registry.my_registry.id}/%v"
		}
		
		resource "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
			name          = "%v"
			status        = "active"
			repository_id = yandex_container_repository.my_repository.id
			description   = "my description"

			rule {
				description   = "everything is complicated"
				untagged      = true
				tag_regexp    = ".name*"
				expire_period = "24h"
				retained_top  = 10
			}
		}`, repositoryName, registryName, lifecyclePolicyName)
}

func getAccResourceContainerRepositoryLifecyclePolicyConfigWithThreeRules(registryName, repositoryName, lifecyclePolicyName string) string {
	return fmt.Sprintf(`
		resource "yandex_container_registry" "my_registry" {
			name = "%v"
		}
		
		resource "yandex_container_repository" "my_repository" {
			name = "${yandex_container_registry.my_registry.id}/%v"
		}
		
		resource "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
			name          = "%v"
			status        = "active"
			repository_id = yandex_container_repository.my_repository.id
			description   = "my description"

			rule {
				description   = "everything is complicated"
				untagged      = true
				tag_regexp    = ".name*"
				expire_period = "24h"
				retained_top  = 10
			}
			rule {
				description   = "untagged and retained top 5"
				untagged      = true
				retained_top  = 5
			}
			rule {
				description   = "regexped and retained top 1"
				untagged      = false
				tag_regexp    = ".name*"
				retained_top  = 1
			}
		}`, repositoryName, registryName, lifecyclePolicyName)
}

func getAccResourceContainerRepositoryLifecyclePolicyConfigWithThreeRulesShuffled(registryName, repositoryName, lifecyclePolicyName string) string {
	return fmt.Sprintf(`
		resource "yandex_container_registry" "my_registry" {
			name = "%v"
		}
		
		resource "yandex_container_repository" "my_repository" {
			name = "${yandex_container_registry.my_registry.id}/%v"
		}
		
		resource "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
			name          = "%v"
			status        = "active"
			repository_id = yandex_container_repository.my_repository.id
			description   = "my description"

			rule {
				description   = "regexped and retained top 1"
				untagged      = false
				tag_regexp    = ".name*"
				retained_top  = 1
			}
			rule {
				description   = "everything is complicated"
				untagged      = true
				tag_regexp    = ".name*"
				expire_period = "24h"
				retained_top  = 10
			}
			rule {
				description   = "untagged and retained top 5"
				untagged      = true
				retained_top  = 5
			}
		}`, repositoryName, registryName, lifecyclePolicyName)
}

func testAccCheckContainerRepositoryLifecyclePolicyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_container_repository_lifecycle_policy" {
			continue
		}

		lifecyclePolicyService := config.sdk.ContainerRegistry().LifecyclePolicy()
		getLifecyclePolicyRequest := containerregistry.GetLifecyclePolicyRequest{LifecyclePolicyId: rs.Primary.ID}
		_, err := lifecyclePolicyService.Get(context.Background(), &getLifecyclePolicyRequest)
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}

			return fmt.Errorf("Container Registry still exists")
		}
	}

	return nil
}
