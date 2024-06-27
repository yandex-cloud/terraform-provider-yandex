package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
)

const apiGatewayDataSource = "data.yandex_api_gateway.test-api-gateway"

func TestAccDataSourceYandexAPIGateway_byID(t *testing.T) {
	t.Parallel()

	var apiGateway apigateway.ApiGateway
	apiGatewayName := acctest.RandomWithPrefix("tf-api-gateway")
	apiGatewayDesc := acctest.RandomWithPrefix("tf-api-gateway-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexAPIGatewayByID(apiGatewayName, apiGatewayDesc),
				Check: resource.ComposeTestCheckFunc(
					testYandexAPIGatewayExists(apiGatewayDataSource, &apiGateway),
					resource.TestCheckResourceAttrSet(apiGatewayDataSource, "api_gateway_id"),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "name", apiGatewayName),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "description", apiGatewayDesc),
					resource.TestCheckResourceAttrSet(apiGatewayDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(apiGatewayDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexAPIGateway_byName(t *testing.T) {
	t.Parallel()

	var apiGateway apigateway.ApiGateway
	apiGatewayName := acctest.RandomWithPrefix("tf-api-gateway")
	apiGatewayDesc := acctest.RandomWithPrefix("tf-api-gateway-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexAPIGatewayByName(apiGatewayName, apiGatewayDesc),
				Check: resource.ComposeTestCheckFunc(
					testYandexAPIGatewayExists(apiGatewayDataSource, &apiGateway),
					resource.TestCheckResourceAttrSet(apiGatewayDataSource, "api_gateway_id"),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "name", apiGatewayName),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "description", apiGatewayDesc),
					resource.TestCheckResourceAttrSet(apiGatewayDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(apiGatewayDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexAPIGateway_full(t *testing.T) {
	t.Parallel()

	var apiGateway apigateway.ApiGateway
	params := testYandexAPIGatewayParameters{}
	params.name = acctest.RandomWithPrefix("tf-api-gateway")
	params.desc = acctest.RandomWithPrefix("tf-api-gateway-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-api-gateway-label")
	params.labelValue = acctest.RandomWithPrefix("tf-api-gateway-label-value")
	params.certificateId = getTestCertificateId(t)
	params.domain = fmt.Sprintf("%s.tf-acc-tests.prod.apigwtest.ru", acctest.RandomWithPrefix("test"))
	params.logOptions = testLogOptions{
		disabled: false,
		minLevel: "WARN",
	}
	params.executionTimeoutSeconds = "5"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexAPIGatewayDataSource(params),
				Check: resource.ComposeTestCheckFunc(
					testYandexAPIGatewayExists(apiGatewayDataSource, &apiGateway),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "name", params.name),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "description", params.desc),
					resource.TestCheckResourceAttrSet(apiGatewayDataSource, "folder_id"),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "custom_domains.0.certificate_id", params.certificateId),
					resource.TestCheckResourceAttrSet(apiGatewayDataSource, "custom_domains.0.domain_id"),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "custom_domains.0.fqdn", params.domain),
					resource.TestCheckNoResourceAttr(apiGatewayDataSource, "custom_domains.1"),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "log_options.0.disabled", fmt.Sprint(params.logOptions.disabled)),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "log_options.0.min_level", params.logOptions.minLevel),
					resource.TestCheckResourceAttrSet(apiGatewayDataSource, "log_options.0.log_group_id"),
					resource.TestCheckResourceAttr(apiGatewayDataSource, executionTimeoutKey, "5"),

					testYandexAPIGatewayContainsLabel(&apiGateway, params.labelKey, params.labelValue),
					testAccCheckCreatedAtAttr(apiGatewayDataSource),
				),
			},
		},
	})
}

func testYandexAPIGatewayByID(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_api_gateway" "test-api-gateway" {
  api_gateway_id = "${yandex_api_gateway.test-api-gateway.id}"
}

resource "yandex_api_gateway" "test-api-gateway" {
  name        = "%s"
  description = "%s"
  spec = <<EOF
%sEOF
}
	`, name, desc, spec)
}

func testYandexAPIGatewayByName(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_api_gateway" "test-api-gateway" {
  name = "${yandex_api_gateway.test-api-gateway.name}"
}

resource "yandex_api_gateway" "test-api-gateway" {
  name        = "%s"
  description = "%s"
  spec = <<EOF
%sEOF
}
	`, name, desc, spec)
}

func testYandexAPIGatewayDataSource(params testYandexAPIGatewayParameters) string {
	return fmt.Sprintf(`
data "yandex_api_gateway" "test-api-gateway" {
  api_gateway_id = "${yandex_api_gateway.test-api-gateway.id}"
}

resource "yandex_api_gateway" "test-api-gateway" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  custom_domains {
    certificate_id = "%s"
    fqdn = "%s"
  }
  log_options {
	disabled = "%t"
	log_group_id = yandex_logging_group.logging-group.id
    min_level = "%s"
  }
  execution_timeout = "%s"
 spec = <<EOF
%sEOF
}

resource "yandex_logging_group" "logging-group" {
}
	`,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue,
		params.certificateId,
		params.domain,
		params.logOptions.disabled,
		params.logOptions.minLevel,
		params.executionTimeoutSeconds,
		spec)
}
