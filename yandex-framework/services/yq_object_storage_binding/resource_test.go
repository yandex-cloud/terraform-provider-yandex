package yq_object_storage_binding_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccYQObjectStorageBindingBasic(t *testing.T) {
	connectionName := fmt.Sprintf("my-conn-%s", acctest.RandString(5))
	connectionResourceName := "my-connection"
	bindingName := fmt.Sprintf("my-bnd-%s", acctest.RandString(5))
	bindingResourceName := "my-binding"

	existingBindingResourceName := fmt.Sprintf("yandex_yq_object_storage_binding.%s", bindingResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return test.TestYandexYQAllBindingsDestroyed(s, "yandex_yq_object_storage_binding")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccYQObjectStorageBindingConfig(connectionName, connectionResourceName, bindingName, bindingResourceName),
				Check: resource.ComposeTestCheckFunc(
					test.TestAccYQBindingExists(bindingName, existingBindingResourceName),
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
	return test.TemplateConfig(`
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
			type="String"
		}
		column {
			name = "year"
			type = "Int32"
			not_null = true
		}
		column {
			name = "z2"
			type = "Utf8"
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
