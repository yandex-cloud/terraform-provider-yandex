package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

const functionDataSource = "data.yandex_function.test-function"

func TestAccDataSourceYandexFunction_byID(t *testing.T) {
	t.Parallel()

	var function functions.Function
	functionName := acctest.RandomWithPrefix("tf-function")
	functionDesc := acctest.RandomWithPrefix("tf-function-desc")
	zipFilename := "test-fixtures/serverless/main.zip"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionByID(functionName, functionDesc, zipFilename),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(functionDataSource, &function),
					resource.TestCheckResourceAttrSet(functionDataSource, "function_id"),
					resource.TestCheckResourceAttr(functionDataSource, "name", functionName),
					resource.TestCheckResourceAttr(functionDataSource, "description", functionDesc),
					resource.TestCheckResourceAttrSet(functionDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(functionDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexFunction_byName(t *testing.T) {
	t.Parallel()

	var function functions.Function
	functionName := acctest.RandomWithPrefix("tf-function")
	functionDesc := acctest.RandomWithPrefix("tf-function-desc")
	zipFilename := "test-fixtures/serverless/main.zip"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionByName(functionName, functionDesc, zipFilename),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(functionDataSource, &function),
					resource.TestCheckResourceAttrSet(functionDataSource, "function_id"),
					resource.TestCheckResourceAttr(functionDataSource, "name", functionName),
					resource.TestCheckResourceAttr(functionDataSource, "description", functionDesc),
					resource.TestCheckResourceAttrSet(functionDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(functionDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexFunction_full(t *testing.T) {
	t.Parallel()

	var function functions.Function
	params := testYandexFunctionParameters{}
	params.name = acctest.RandomWithPrefix("tf-function")
	params.desc = acctest.RandomWithPrefix("tf-function-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-function-label")
	params.labelValue = acctest.RandomWithPrefix("tf-function-label-value")
	params.userHash = acctest.RandomWithPrefix("tf-function-hash")
	params.runtime = "python37"
	params.memory = "128"
	params.executionTimeout = "10"
	params.serviceAccount = acctest.RandomWithPrefix("tf-service-account")
	params.envKey = "tf_function_env"
	params.envValue = "tf_function_env_value"
	params.tags = acctest.RandomWithPrefix("tf-function-tag")
	params.secret = testSecretParameters{
		secretName:   "tf-function-secret-name",
		secretKey:    "tf-function-secret-key",
		secretEnvVar: "TF_FUNCTION_ENV_KEY",
		secretValue:  "tf-function-secret-value",
	}
	params.zipFilename = "test-fixtures/serverless/main.zip"
	params.maxAsyncRetries = "2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionDataSource(params),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(functionDataSource, &function),
					resource.TestCheckResourceAttr(functionDataSource, "name", params.name),
					resource.TestCheckResourceAttr(functionDataSource, "description", params.desc),
					resource.TestCheckResourceAttrSet(functionDataSource, "folder_id"),
					testYandexFunctionContainsLabel(&function, params.labelKey, params.labelValue),
					resource.TestCheckResourceAttr(functionDataSource, "runtime", params.runtime),
					resource.TestCheckResourceAttr(functionDataSource, "memory", params.memory),
					resource.TestCheckResourceAttr(functionDataSource, "execution_timeout", params.executionTimeout),
					resource.TestCheckResourceAttrSet(functionDataSource, "service_account_id"),
					testYandexFunctionContainsEnv(functionResource, params.envKey, params.envValue),
					testYandexFunctionContainsTag(functionDataSource, params.tags),
					resource.TestCheckResourceAttrSet(functionDataSource, "version"),
					resource.TestCheckResourceAttrSet(functionDataSource, "image_size"),
					resource.TestCheckResourceAttrSet(functionDataSource, "secrets.0.id"),
					resource.TestCheckResourceAttrSet(functionDataSource, "secrets.0.version_id"),
					resource.TestCheckResourceAttr(functionDataSource, "secrets.0.key", params.secret.secretKey),
					resource.TestCheckResourceAttr(functionDataSource, "secrets.0.environment_variable", params.secret.secretEnvVar),
					resource.TestCheckResourceAttr(functionDataSource, "async_invocation.0.retries_count", params.maxAsyncRetries),
					testAccCheckCreatedAtAttr(functionDataSource),
				),
			},
		},
	})
}

func testYandexFunctionByID(name string, desc string, zipFilename string) string {
	return fmt.Sprintf(`
data "yandex_function" "test-function" {
  function_id = "${yandex_function.test-function.id}"
}

resource "yandex_function" "test-function" {
  name        = "%s"
  description = "%s"
  user_hash   = "user_hash"
  runtime     = "python37"
  entrypoint  = "main"
  memory      = "128"
  content {
    zip_filename = "%s"
  }
}
	`, name, desc, zipFilename)
}

func testYandexFunctionByName(name string, desc string, zipFilename string) string {
	return fmt.Sprintf(`
data "yandex_function" "test-function" {
  name = "${yandex_function.test-function.name}"
}

resource "yandex_function" "test-function" {
  name        = "%s"
  description = "%s"
  user_hash   = "user_hash"
  runtime     = "python37"
  entrypoint  = "main"
  memory      = "128"
  content {
    zip_filename = "%s"
  }
}
	`, name, desc, zipFilename)
}

func testYandexFunctionDataSource(params testYandexFunctionParameters) string {
	return fmt.Sprintf(`
data "yandex_function" "test-function" {
  function_id = "${yandex_function.test-function.id}"
}

resource "yandex_function" "test-function" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  user_hash          = "%s"
  runtime            = "%s"
  entrypoint         = "main"
  memory             = "%s"
  execution_timeout  = "%s"
  service_account_id = "${yandex_iam_service_account.test-account.id}"
  depends_on = [
	yandex_resourcemanager_folder_iam_binding.payload-viewer
  ]
  environment = {
    %s = "%s"
  }
  tags = ["%s"]
  secrets {
    id = yandex_lockbox_secret.secret.id
    version_id = yandex_lockbox_secret_version.secret_version.id
    key = "%s"
    environment_variable = "%s"
  }
  content {
    zip_filename = "%s"
  }
  async_invocation {
    retries_count = "%s"
  }
}

resource "yandex_iam_service_account" "test-account" {
  name = "%s"
}

resource "yandex_resourcemanager_folder_iam_binding" "payload-viewer" {
  folder_id   = yandex_lockbox_secret.secret.folder_id
  role        = "lockbox.payloadViewer"
  members     = [
    "serviceAccount:${yandex_iam_service_account.test-account.id}",
  ]
}

resource "yandex_lockbox_secret" "secret" {
  name        = "%s"
}

resource "yandex_lockbox_secret_version" "secret_version" {
  secret_id = yandex_lockbox_secret.secret.id
  entries {
    key        = "%s"
    text_value = "%s"
  }
}
	`,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue,
		params.userHash,
		params.runtime,
		params.memory,
		params.executionTimeout,
		params.envKey,
		params.envValue,
		params.tags,
		params.secret.secretKey,
		params.secret.secretEnvVar,
		params.zipFilename,
		params.maxAsyncRetries,
		params.serviceAccount,
		params.secret.secretName,
		params.secret.secretKey,
		params.secret.secretValue)
}
