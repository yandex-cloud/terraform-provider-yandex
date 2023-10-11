package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const lifecyclePolicyDatasourceName = "data.yandex_container_repository_lifecycle_policy.my_lifecycle_policy"

func TestAccDataSourceContainerRepositoryLifecyclePolicy(t *testing.T) {
	t.Parallel()

	t.Run("test data_yandex_container_repository_lifecycle_policy by ID", func(t *testing.T) {
		var (
			registryName        = acctest.RandomWithPrefix("tf-registry")
			repositoryName      = acctest.RandomWithPrefix("tf-repository")
			lifecyclePolicyName = acctest.RandomWithPrefix("tf-lifecycle-policy")
		)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckContainerRegistryDestroy,
				testAccCheckContainerRepositoryDestroy,
				testAccCheckContainerRepositoryLifecyclePolicyDestroy,
			),
			Steps: []resource.TestStep{
				getAccDataSourceContainerRepositoryLifecyclePolicyTestStep(repositoryName, registryName, lifecyclePolicyName, true /* isByID */),
			},
		})
	})

	t.Run("test data_yandex_container_repository_lifecycle_policy by name", func(t *testing.T) {
		var (
			registryName        = acctest.RandomWithPrefix("tf-registry")
			repositoryName      = acctest.RandomWithPrefix("tf-repository")
			lifecyclePolicyName = acctest.RandomWithPrefix("tf-lifecycle-policy")
		)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckContainerRegistryDestroy,
				testAccCheckContainerRepositoryDestroy,
				testAccCheckContainerRepositoryLifecyclePolicyDestroy,
			),
			Steps: []resource.TestStep{
				getAccDataSourceContainerRepositoryLifecyclePolicyTestStep(repositoryName, registryName, lifecyclePolicyName, false /* isByID */),
			},
		})
	})
}

func getAccDataSourceContainerRepositoryLifecyclePolicyTestStep(repositoryName, registryName, lifecyclePolicyName string, isByID bool) resource.TestStep {
	return resource.TestStep{
		Config: getAccDataSourceContainerRepositoryLifecyclePolicyConfig(repositoryName, registryName, lifecyclePolicyName, isByID),
		Check: resource.ComposeTestCheckFunc(
			// IDs
			testAccCheckResourceIDField(lifecyclePolicyDatasourceName, "lifecycle_policy_id"),
			resource.TestCheckResourceAttrSet(lifecyclePolicyDatasourceName, "id"),
			resource.TestCheckResourceAttrSet(lifecyclePolicyDatasourceName, "repository_id"),

			// attributes
			resource.TestCheckResourceAttr(lifecyclePolicyDatasourceName, "name", lifecyclePolicyName),
			resource.TestCheckResourceAttr(lifecyclePolicyDatasourceName, "status", "active"),
			resource.TestCheckResourceAttr(lifecyclePolicyDatasourceName, "description", "my description"),
			resource.TestCheckResourceAttr(lifecyclePolicyDatasourceName, "rule.0.description", "my description"),
			resource.TestCheckResourceAttr(lifecyclePolicyDatasourceName, "rule.0.expire_period", "24h0m0s"),
			resource.TestCheckResourceAttr(lifecyclePolicyDatasourceName, "rule.0.tag_regexp", ".*"),
			resource.TestCheckResourceAttr(lifecyclePolicyDatasourceName, "rule.0.untagged", "true"),
			resource.TestCheckResourceAttr(lifecyclePolicyDatasourceName, "rule.0.retained_top", "2"),
			testAccCheckCreatedAtAttr(lifecyclePolicyDatasourceName),
		),
	}
}

func getAccDataSourceContainerRepositoryLifecyclePolicyConfig(repositoryName, registryName, lifecyclePolicyName string, isByID bool) string {
	prereqResources := fmt.Sprintf(`
		resource "yandex_container_registry" "my_registry" {
			name = "%v"
		}
		
		resource "yandex_container_repository" "my_repository" {
			name = "${yandex_container_registry.my_registry.id}/%v"
		}
		
		resource "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
			name          = "%v"
			status        = "active"
			description   = "my description"
			repository_id = yandex_container_repository.my_repository.id
		
			rule {
			description   = "my description"
			expire_period = "24h"
			tag_regexp    = ".*"
			untagged      = true
			retained_top  = 2
			}
		}`, repositoryName, registryName, lifecyclePolicyName)

	const data_yandex_container_repository_lifecycle_policy_byID = `

		data "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
			lifecycle_policy_id = yandex_container_repository_lifecycle_policy.my_lifecycle_policy.id
		}`

	const data_yandex_container_repository_lifecycle_policy_byName = `

		data "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
			name          = yandex_container_repository_lifecycle_policy.my_lifecycle_policy.name
			repository_id = yandex_container_repository.my_repository.id
		}`

	if isByID {
		return prereqResources + data_yandex_container_repository_lifecycle_policy_byID
	}

	return prereqResources + data_yandex_container_repository_lifecycle_policy_byName
}
