package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

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
	params.zipFilename = "test-fixtures/serverless/main.zip"

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
					resource.TestCheckResourceAttrSet(functionDataSource, "loggroup_id"),
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
  environment = {
    %s = "%s"
  }
  tags = ["%s"]
  content {
    zip_filename = "%s"
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
		params.userHash,
		params.runtime,
		params.memory,
		params.executionTimeout,
		params.envKey,
		params.envValue,
		params.tags,
		params.zipFilename,
		params.serviceAccount)
}
