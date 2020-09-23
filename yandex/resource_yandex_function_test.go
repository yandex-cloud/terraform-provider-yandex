package yandex

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

const functionResource = "yandex_function.test-function"

func init() {
	resource.AddTestSweepers("yandex_function", &resource.Sweeper{
		Name: "yandex_function",
		F:    testSweepFunction,
		Dependencies: []string{
			"yandex_function_trigger",
		},
	})
}

func testSweepFunction(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	it := conf.sdk.Serverless().Functions().Function().FunctionIterator(conf.Context(), conf.FolderID)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepFunction(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Function %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepFunction(conf *Config, id string) bool {
	return sweepWithRetry(sweepFunctionOnce, conf, "Function", id)
}

func sweepFunctionOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexFunctionDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Serverless().Functions().Function().Delete(ctx, &functions.DeleteFunctionRequest{
		FunctionId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccYandexFunction_basic(t *testing.T) {
	t.Parallel()

	var function functions.Function
	functionName := acctest.RandomWithPrefix("tf-function")
	functionDesc := acctest.RandomWithPrefix("tf-function-desc")
	labelKey := acctest.RandomWithPrefix("tf-function-label")
	labelValue := acctest.RandomWithPrefix("tf-function-label-value")

	zipFilename := "test-fixtures/serverless/main.zip"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps:        []resource.TestStep{basicYandexFunctionTestStep(functionName, functionDesc, labelKey, labelValue, zipFilename, &function)},
	})
}

func TestAccYandexFunction_update(t *testing.T) {
	t.Parallel()

	var function functions.Function
	functionName := acctest.RandomWithPrefix("tf-function")
	functionDesc := acctest.RandomWithPrefix("tf-function-desc")
	labelKey := acctest.RandomWithPrefix("tf-function-label")
	labelValue := acctest.RandomWithPrefix("tf-function-label-value")

	functionNameUpdated := acctest.RandomWithPrefix("tf-function-updated")
	functionDescUpdated := acctest.RandomWithPrefix("tf-function-desc-updated")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-function-label-updated")
	labelValueUpdated := acctest.RandomWithPrefix("tf-function-label-value-updated")

	zipFilename := "test-fixtures/serverless/main.zip"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			basicYandexFunctionTestStep(functionName, functionDesc, labelKey, labelValue, zipFilename, &function),
			basicYandexFunctionTestStep(functionNameUpdated, functionDescUpdated, labelKeyUpdated, labelValueUpdated, zipFilename, &function),
		},
	})
}

func TestAccYandexFunction_full(t *testing.T) {
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

	paramsUpdated := testYandexFunctionParameters{}
	paramsUpdated.name = acctest.RandomWithPrefix("tf-function-updated")
	paramsUpdated.desc = acctest.RandomWithPrefix("tf-function-desc-updated")
	paramsUpdated.labelKey = acctest.RandomWithPrefix("tf-function-label-updated")
	paramsUpdated.labelValue = acctest.RandomWithPrefix("tf-function-label-value-updated")
	paramsUpdated.userHash = acctest.RandomWithPrefix("tf-function-hash-updated")
	paramsUpdated.runtime = "python27"
	paramsUpdated.memory = "256"
	paramsUpdated.executionTimeout = "11"
	paramsUpdated.serviceAccount = acctest.RandomWithPrefix("tf-service-account")
	paramsUpdated.envKey = "tf_function_env_updated"
	paramsUpdated.envValue = "tf_function_env_value_updated"
	paramsUpdated.tags = acctest.RandomWithPrefix("tf-function-tag-updated")
	paramsUpdated.zipFilename = "test-fixtures/serverless/main.zip"

	testConfigFunc := func(params testYandexFunctionParameters) resource.TestStep {
		return resource.TestStep{
			Config: testYandexFunctionFull(params),
			Check: resource.ComposeTestCheckFunc(
				testYandexFunctionExists(functionResource, &function),
				resource.TestCheckResourceAttr(functionResource, "name", params.name),
				resource.TestCheckResourceAttr(functionResource, "description", params.desc),
				resource.TestCheckResourceAttrSet(functionResource, "folder_id"),
				testYandexFunctionContainsLabel(&function, params.labelKey, params.labelValue),
				resource.TestCheckResourceAttr(functionResource, "user_hash", params.userHash),
				resource.TestCheckResourceAttr(functionResource, "runtime", params.runtime),
				resource.TestCheckResourceAttr(functionResource, "memory", params.memory),
				resource.TestCheckResourceAttr(functionResource, "execution_timeout", params.executionTimeout),
				resource.TestCheckResourceAttrSet(functionResource, "service_account_id"),
				testYandexFunctionContainsEnv(functionResource, params.envKey, params.envValue),
				testYandexFunctionContainsTag(functionResource, params.tags),
				resource.TestCheckResourceAttrSet(functionResource, "version"),
				resource.TestCheckResourceAttrSet(functionResource, "image_size"),
				resource.TestCheckResourceAttrSet(functionResource, "loggroup_id"),
				testAccCheckCreatedAtAttr(functionResource),
			),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps:        []resource.TestStep{testConfigFunc(params), testConfigFunc(paramsUpdated)},
	})
}

func basicYandexFunctionTestStep(functionName, functionDesc, labelKey, labelValue, zipFilename string, function *functions.Function) resource.TestStep {
	return resource.TestStep{
		Config: testYandexFunctionBasic(functionName, functionDesc, labelKey, labelValue, zipFilename),
		Check: resource.ComposeTestCheckFunc(
			testYandexFunctionExists(functionResource, function),
			resource.TestCheckResourceAttr(functionResource, "name", functionName),
			resource.TestCheckResourceAttr(functionResource, "description", functionDesc),
			resource.TestCheckResourceAttrSet(functionResource, "folder_id"),
			testYandexFunctionContainsLabel(function, labelKey, labelValue),
			testAccCheckCreatedAtAttr(functionResource),
		),
	}
}

func testYandexFunctionDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_function" {
			continue
		}

		_, err := testGetFunctionByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Function still exists")
		}
	}

	return nil
}

func testYandexFunctionExists(name string, function *functions.Function) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetFunctionByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Function not found")
		}

		*function = *found
		return nil
	}
}

func testGetFunctionByID(config *Config, ID string) (*functions.Function, error) {
	req := functions.GetFunctionRequest{
		FunctionId: ID,
	}

	return config.sdk.Serverless().Functions().Function().Get(context.Background(), &req)
}

func testYandexFunctionContainsLabel(function *functions.Function, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := function.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testYandexFunctionContainsEnv(name string, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resources, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found environment: %s in %s", value, s.RootModule().Path)
		}

		for k, v := range resources.Primary.Attributes {
			if strings.HasPrefix(k, "environment") && strings.Contains(k, key) && v == value {
				return nil
			}
		}

		return fmt.Errorf("Not found environment: %s, value: %s in %s", key, value, s.RootModule().Path)
	}
}

func testYandexFunctionContainsTag(name, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resources, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found tags: %s in %s", value, s.RootModule().Path)
		}

		for k, v := range resources.Primary.Attributes {
			if strings.HasPrefix(k, "tags") && v == value {
				return nil
			}
		}

		return fmt.Errorf("Not found tags: %s in %s", value, s.RootModule().Path)
	}
}

func testYandexFunctionBasic(name string, desc string, labelKey string, labelValue string, zipFileName string) string {
	return fmt.Sprintf(`
resource "yandex_function" "test-function" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "%s"
  }
}
	`, name, desc, labelKey, labelValue, zipFileName)
}

type testYandexFunctionParameters struct {
	name             string
	desc             string
	labelKey         string
	labelValue       string
	userHash         string
	runtime          string
	memory           string
	executionTimeout string
	serviceAccount   string
	envKey           string
	envValue         string
	tags             string
	zipFilename      string
}

func testYandexFunctionFull(params testYandexFunctionParameters) string {
	return fmt.Sprintf(`
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
