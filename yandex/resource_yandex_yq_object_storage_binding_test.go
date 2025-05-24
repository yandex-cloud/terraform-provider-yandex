package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccYandexYQObjectStorageBindingBasic(t *testing.T) {
	connectionName := fmt.Sprintf("my-conn-%s", acctest.RandString(5))
	connectionResourceName := "my-connection"
	bindingName := fmt.Sprintf("my-bnd-%s", acctest.RandString(5))
	bindingResourceName := "my-binding"

	existingBindingResourceName := fmt.Sprintf("yandex_yq_object_storage_binding.%s", bindingResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return testYandexYQAllBindingsDestroyed(s, "yandex_yq_object_storage_binding")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccYQObjectStorageBindingConfig(connectionName, connectionResourceName, bindingName, bindingResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccYQBindingExists(bindingName, existingBindingResourceName),
				),
			},
			{
				ResourceName:      existingBindingResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccYQObjectStorageBindingConfig(connectionName string, connectionResourceName string, bindingName string, bindingResourceName string) string {
	return templateConfig(`
	resource "yandex_yq_object_storage_connection" "{{.ConnectionResourceName}}" {
        name = "{{.ConnectionName}}"
		description = "my_desc"
        bucket = "my_bucket"
    }
	
	resource "yandex_yq_object_storage_binding" "{{.BindingResourceName}}" {
    	name = "{{.BindingName}}"
    	connection_id = yandex_yq_object_storage_connection.{{.ConnectionResourceName}}.id
    	format = "csv_with_names"
    	path_pattern = "x/"
		format_setting = {
			"file_pattern" = "abc/*.csv"
		}

		column {
			name="zzzz"
			type="string"
		}
		column {
			name = "year"
			type = "int32"
			not_null = true
		}
		column {
			name = "z2"
			type = "UTF8"
		}

		partitioned_by = ["year"]
		projection = {
			"projection.enabled" : "true",
			"projection.year.type" : "integer",
			"projection.year.min" : "2020",
			"projection.year.max" : "2027",
			"projection.year.interval" : "1",
			"storage.location.template" : "/$${year}",
		}
	}`, map[string]interface{}{
		"ConnectionName":         connectionName,
		"ConnectionResourceName": connectionResourceName,
		"BindingName":            bindingName,
		"BindingResourceName":    bindingResourceName,
	})
}
