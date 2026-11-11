package acceptance

import (
	"fmt"
	"testing"

	apifixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/apigateway/testfixtures"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

const apiGatewayDataSource = "data.yandex_api_gateway.test-api-gateway"

func TestAccDataSourceYandexAPIGateway_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var apiGateway apigateway.ApiGateway
	apiGatewayName := acctest.RandomWithPrefix("tf-api-gateway")
	apiGatewayDesc := acctest.RandomWithPrefix("tf-api-gateway-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexAPIGatewayByID(apiGatewayName, apiGatewayDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexAPIGatewayExists(apiGatewayDataSource, &apiGateway),
					resource.TestCheckResourceAttrSet(apiGatewayDataSource, "api_gateway_id"),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "name", apiGatewayName),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "description", apiGatewayDesc),
					resource.TestCheckResourceAttrSet(apiGatewayDataSource, "folder_id"),
					statecheck.CheckCreatedAtAttr(apiGatewayDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexAPIGateway_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var apiGateway apigateway.ApiGateway
	apiGatewayName := acctest.RandomWithPrefix("tf-api-gateway")
	apiGatewayDesc := acctest.RandomWithPrefix("tf-api-gateway-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexAPIGatewayByName(apiGatewayName, apiGatewayDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexAPIGatewayExists(apiGatewayDataSource, &apiGateway),
					resource.TestCheckResourceAttrSet(apiGatewayDataSource, "api_gateway_id"),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "name", apiGatewayName),
					resource.TestCheckResourceAttr(apiGatewayDataSource, "description", apiGatewayDesc),
					resource.TestCheckResourceAttrSet(apiGatewayDataSource, "folder_id"),
					statecheck.CheckCreatedAtAttr(apiGatewayDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexAPIGateway_full(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
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
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexAPIGatewayDataSource(params),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexAPIGatewayExists(apiGatewayDataSource, &apiGateway),
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
					resource.TestCheckResourceAttr(apiGatewayDataSource, "execution_timeout", "5"),

					testYandexAPIGatewayContainsLabel(&apiGateway, params.labelKey, params.labelValue),
					statecheck.CheckCreatedAtAttr(apiGatewayDataSource),
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
	`, name, desc, apifixtures.Spec)
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
	`, name, desc, apifixtures.Spec)
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
		apifixtures.Spec)
}
