package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const serverlessContainerDataSource = "data.yandex_serverless_container.test-container"

func TestAccDataSourceYandexServerlessContainer_byID(t *testing.T) {
	t.Parallel()

	var container containers.Container
	containerName := acctest.RandomWithPrefix("tf-container")
	containerDesc := acctest.RandomWithPrefix("tf-container-desc")
	memory := (1 + acctest.RandIntRange(1, 3)) * 128

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexServerlessContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexServerlessContainerByID(containerName, containerDesc, memory, serverlessContainerTestImage1),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessContainerExists(serverlessContainerDataSource, &container),
					resource.TestCheckResourceAttrSet(serverlessContainerDataSource, "container_id"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "name", containerName),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "description", containerDesc),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "memory", strconv.Itoa(memory)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.url", serverlessContainerTestImage1),
					resource.TestCheckResourceAttrSet(serverlessContainerDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(serverlessContainerDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexServerlessContainer_byName(t *testing.T) {
	t.Parallel()

	var container containers.Container
	containerName := acctest.RandomWithPrefix("tf-container")
	containerDesc := acctest.RandomWithPrefix("tf-container-desc")
	memory := (1 + acctest.RandIntRange(1, 3)) * 128

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexServerlessContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexServerlessContainerByName(containerName, containerDesc, memory, serverlessContainerTestImage1),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessContainerExists(serverlessContainerDataSource, &container),
					resource.TestCheckResourceAttrSet(serverlessContainerDataSource, "container_id"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "name", containerName),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "description", containerDesc),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "memory", strconv.Itoa(memory)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.url", serverlessContainerTestImage1),
					resource.TestCheckResourceAttrSet(serverlessContainerDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(serverlessContainerDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexServerlessContainer_full(t *testing.T) {
	t.Parallel()

	var container containers.Container
	params := testYandexServerlessContainerParameters{}
	params.name = acctest.RandomWithPrefix("tf-container")
	params.desc = acctest.RandomWithPrefix("tf-container-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-container-label")
	params.labelValue = acctest.RandomWithPrefix("tf-container-label-value")
	params.memory = (1 + acctest.RandIntRange(1, 3)) * 128
	params.cores = 1
	params.coreFraction = 100
	params.executionTimeout = strconv.FormatInt(int64(1+acctest.RandIntRange(1, 10)), 10) + "s"
	params.concurrency = acctest.RandIntRange(0, 3)
	params.imageURL = serverlessContainerTestImage1
	params.workDir = acctest.RandomWithPrefix("tf-container-work-dir")
	params.command = acctest.RandomWithPrefix("tf-container-command")
	params.argument = acctest.RandomWithPrefix("tf-container-argument")
	params.envVarKey = "env_var_key"
	params.envVarValue = acctest.RandomWithPrefix("tf-container-env-value")
	params.serviceAccount = acctest.RandomWithPrefix("tf-container-sa")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexServerlessContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexServerlessContainerDataSource(params),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessContainerExists(serverlessContainerDataSource, &container),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "name", params.name),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "description", params.desc),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "labels.%", "1"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "labels."+params.labelKey, params.labelValue),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "memory", strconv.Itoa(params.memory)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "cores", strconv.Itoa(params.cores)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "core_fraction", strconv.Itoa(params.coreFraction)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "execution_timeout", params.executionTimeout),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "concurrency", strconv.Itoa(params.concurrency)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.url", params.imageURL),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.digest", serverlessContainerTestDigest1),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.work_dir", params.workDir),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.command.#", "1"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.command.0", params.command),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.args.#", "1"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.args.0", params.argument),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.environment.%", "1"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.environment."+params.envVarKey, params.envVarValue),
					testYandexServerlessContainerServiceAccountAttr(serverlessContainerDataSource, serverlessContainerServiceAccountResource),
					resource.TestCheckResourceAttrSet(serverlessContainerResource, "revision_id"),
					resource.TestCheckResourceAttrSet(serverlessContainerResource, "folder_id"),
					resource.TestCheckResourceAttrSet(serverlessContainerResource, "url"),
					testAccCheckCreatedAtAttr(serverlessContainerResource),
				),
			},
		},
	})
}

func testYandexServerlessContainerServiceAccountAttr(name string, serviceAccountResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found serverless container: %s", name)
		}
		sa, ok := s.RootModule().Resources[serviceAccountResource]
		if !ok {
			return fmt.Errorf("Not found service account: %s", serviceAccountResource)
		}
		serviceAccountID := rs.Primary.Attributes["service_account_id"]
		if serviceAccountID != sa.Primary.ID {
			return fmt.Errorf("Incorrect service account id: expected '%s' but found '%s'", sa.Primary.ID, serviceAccountID)
		}
		return nil
	}
}

func testYandexServerlessContainerByID(name string, desc string, memory int, image string) string {
	return fmt.Sprintf(`
data "yandex_serverless_container" "test-container" {
  container_id = "${yandex_serverless_container.test-container.id}"
}

resource "yandex_serverless_container" "test-container" {
  name        = "%s"
  description = "%s"
  memory      = %d
  image {
    url = "%s"
  }
}
	`, name, desc, memory, image)
}

func testYandexServerlessContainerByName(name string, desc string, memory int, image string) string {
	return fmt.Sprintf(`
data "yandex_serverless_container" "test-container" {
  name = "${yandex_serverless_container.test-container.name}"
}

resource "yandex_serverless_container" "test-container" {
  name        = "%s"
  description = "%s"
  memory      = %d
  image {
    url = "%s"
  }
}
	`, name, desc, memory, image)
}

func testYandexServerlessContainerDataSource(params testYandexServerlessContainerParameters) string {
	return fmt.Sprintf(`
data "yandex_serverless_container" "test-container" {
  container_id = "${yandex_serverless_container.test-container.id}"
}

resource "yandex_serverless_container" "test-container" {
  name        = "%s"
  description = "%s"
  labels = {
    %s   = "%s"
  }
  memory             = %d
  cores              = %d
  core_fraction      = %d
  execution_timeout  = "%s"
  concurrency        = %d
  service_account_id = "${yandex_iam_service_account.test-account.id}"
  image {
    url         = "%s"
    work_dir    = "%s"
    command     = ["%s"]
    args        = ["%s"]
    environment = {
      %s = "%s"
    }
  }
}

resource "yandex_iam_service_account" "test-account" {
  name = "%s"
}
	`,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue,
		params.memory,
		params.cores,
		params.coreFraction,
		params.executionTimeout,
		params.concurrency,
		params.imageURL,
		params.workDir,
		params.command,
		params.argument,
		params.envVarKey,
		params.envVarValue,
		params.serviceAccount)
}
