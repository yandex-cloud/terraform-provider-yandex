package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

func TestYandexFunctionVersion_createVersion(t *testing.T) {
	t.Parallel()
	const (
		functionResourceID        = "create-version-function"
		functionVersionResourceID = "create-version-function-version"
		zipFilename               = "test-fixtures/serverless/main.zip"
	)
	functionName := acctest.RandomWithPrefix("tf-function")

	var function functions.Function
	var version functions.Version

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionVersionCreateVersion(functionResourceID, functionName, functionVersionResourceID, zipFilename),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(functionName, &function),
					testYandexFunctionVersionExists("testing-1", &function, &version),
					testYandexFunctionVersionExists("testing-2", &function, &version),

					resource.TestCheckResourceAttr(functionVersionResourceID, "runtime", "python37"),
					resource.TestCheckResourceAttr(functionVersionResourceID, "memory", "128"),
				),
			},
		},
	})
}

func TestYandexFunctionVersion_updateVersion(t *testing.T) {
	t.Parallel()
	const (
		functionResourceID        = "update-version-function"
		functionVersionResourceID = "update-version-function-version"
		zipFilename               = "test-fixtures/serverless/main.zip"
	)

	functionName := acctest.RandomWithPrefix("tf-function")

	var function functions.Function
	var version functions.Version

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionVersionCreateVersion(functionResourceID, functionName, functionVersionResourceID, zipFilename),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(functionResourceID, &function),
					testYandexFunctionVersionExists("testing-1", &function, &version),

					resource.TestCheckResourceAttr(functionVersionResourceID, "runtime", "python37"),
					resource.TestCheckResourceAttr(functionVersionResourceID, "memory", "128"),
				),
			},
			{
				Config: testYandexFunctionVersionUpdateVersion(functionResourceID, functionName, functionVersionResourceID, zipFilename),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionVersionDoesNotExist("testing-2", &function),
					testYandexFunctionVersionExists("testing-3", &function, &version),

					resource.TestCheckResourceAttr(functionVersionResourceID, "runtime", "python38"),
					resource.TestCheckResourceAttr(functionVersionResourceID, "memory", "256"),
					resource.TestCheckResourceAttr(functionResourceID, "id", version.Id),
				),
			},
		},
	})
}

func TestYandexFunctionVersion_deleteVersion(t *testing.T) {
	t.Parallel()
	const (
		functionResourceID        = "delete-version-function"
		functionVersionResourceID = "delete-version-function-version"
		zipFilename               = "test-fixtures/serverless/main.zip"
	)
	functionName := acctest.RandomWithPrefix("tf-function")

	var function functions.Function
	var version functions.Version

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionVersionCreateVersion(functionResourceID, functionName, functionVersionResourceID, zipFilename),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(functionResourceID, &function),
					testYandexFunctionVersionExists("$latest", &function, &version),
					resource.TestCheckResourceAttr(functionVersionResourceID, "runtime", "python37"),
					resource.TestCheckResourceAttr(functionVersionResourceID, "memory", "128"),
				),
			},
			{
				Config: testYandexFunctionVersionWithoutVersions(functionResourceID, functionName),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionVersionDoesNotExist("testing-1", &function),
					testYandexFunctionVersionDoesNotExist("testing-2", &function),
				),
			},
		},
	})
}

func testYandexFunctionVersionExists(tag string, function *functions.Function, version *functions.Version) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		req := &functions.GetFunctionVersionByTagRequest{
			FunctionId: function.Id,
			Tag:        tag,
		}
		v, err := config.sdk.Serverless().Functions().Function().GetVersionByTag(context.Background(), req)
		if err != nil {
			return err
		}

		*version = *v
		return nil
	}
}

func testYandexFunctionVersionDoesNotExist(tag string, function *functions.Function) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var version *functions.Version
		err := testYandexFunctionVersionExists(tag, function, version)(s)
		if isStatusWithCode(err, codes.NotFound) {
			return nil
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Unexpected version for function %q with tag %q", function.Id, tag)
	}
}

func testYandexFunctionVersionWithoutVersions(functionResourceID, functionName string) string {
	return fmt.Sprintf(`
resource "yandex_function" "%s" {
	name = "%s"
}`, functionResourceID, functionName)
}

func testYandexFunctionVersionCreateVersion(functionResourceID, functionName, functionVersionResourceID, zipFilepath string) string {
	return fmt.Sprintf(`
resource "yandex_function" "%s" {
  name = "%s"
}

resource "yandex_function_version" "%s" {
  function_id = yandex_function.%s.id
  
  entrypoint         = "x.FunctionExample"
  memory             = "128m"
  execution_timeout  = "10s"
  runtime            = "python37"
  content {
    zip_filename     = "%s"
  }
  
  tags = ["testing-1", "testing-2"]
}
`, functionResourceID, functionName, functionVersionResourceID, functionResourceID, zipFilepath)
}

func testYandexFunctionVersionUpdateVersion(functionResourceID, functionName, functionVersionResourceID, zipFilepath string) string {
	return fmt.Sprintf(`
resource "yandex_function" "%s" {
  name = "%s"
}

resource "yandex_function_version" "%s" {
  function_id = yandex_function.%s.id
  
  entrypoint         = "x.NewFunctionExample"
  memory             = "256m"
  execution_timeout  = "10s"
  runtime            = "python38"
  content {
    zip_filename     = "%s"
  }

  tags = ["testing-1", "testing-3"]
}
`, functionResourceID, functionName, functionVersionResourceID, functionResourceID, zipFilepath)
}
