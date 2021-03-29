package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
)

const apiGatewayDataSource = "data.yandex_api_gateway.test-api-gateway"

func TestAccDataSourceYandexAPIGateway_byID(t *testing.T) {
	t.Parallel()

	var apiGateway apigateway.ApiGateway
	apiGatewayName := acctest.RandomWithPrefix("tf-api-gateway")
	apiGatewayDesc := acctest.RandomWithPrefix("tf-api-gateway-desc")
	yamlFilename := specFile

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexAPIGatewayByID(apiGatewayName, apiGatewayDesc, yamlFilename),
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
	yamlFilename := specFile

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexAPIGatewayByName(apiGatewayName, apiGatewayDesc, yamlFilename),
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
	params.yamlFilename = specFile

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
					testYandexAPIGatewayContainsLabel(&apiGateway, params.labelKey, params.labelValue),
					testAccCheckCreatedAtAttr(apiGatewayDataSource),
				),
			},
		},
	})
}

func testYandexAPIGatewayByID(name string, desc string, yamlFilename string) string {
	return fmt.Sprintf(`
data "yandex_api_gateway" "test-api-gateway" {
  api_gateway_id = "${yandex_api_gateway.test-api-gateway.id}"
}

resource "yandex_api_gateway" "test-api-gateway" {
  name        = "%s"
  description = "%s"
  spec = "%s"
  spec_content_hash = %d
}
	`, name, desc, yamlFilename, specHash)
}

func testYandexAPIGatewayByName(name string, desc string, yamlFilename string) string {
	return fmt.Sprintf(`
data "yandex_api_gateway" "test-api-gateway" {
  name = "${yandex_api_gateway.test-api-gateway.name}"
}

resource "yandex_api_gateway" "test-api-gateway" {
  name        = "%s"
  description = "%s"
  spec = "%s"
  spec_content_hash = %d
}
	`, name, desc, yamlFilename, specHash)
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
  spec = "%s"
  spec_content_hash = %d
}
	`,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue,
		params.yamlFilename,
		specHash)
}
