package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccYandexYQYDSBindingBasic(t *testing.T) {
	connectionName := fmt.Sprintf("my-conn-%s", acctest.RandString(5))
	connectionResourceName := "my-connection"
	bindingName := fmt.Sprintf("my-bnd-%s", acctest.RandString(5))
	bindingResourceName := "my-binding"

	existingBindingResourceName := fmt.Sprintf("yandex_yq_yds_binding.%s", bindingResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return testYandexYQAllBindingsDestroyed(s, "yandex_yq_yds_binding")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccYQYDSBindingConfig(connectionName, connectionResourceName, bindingName, bindingResourceName),
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

func testAccYQYDSBindingConfig(connectionName string, connectionResourceName string, bindingName string, bindingResourceName string) string {
	return templateConfig(`
	resource "yandex_yq_yds_connection" "{{.ConnectionResourceName}}" {
        name = "{{.ConnectionName}}"
		description = "my_desc"
		database_id = "123abc"
    }
	
	resource "yandex_yq_yds_binding" "{{.BindingResourceName}}" {
    	name = "{{.BindingName}}"
    	connection_id = yandex_yq_yds_connection.{{.ConnectionResourceName}}.id
    	format = "csv_with_names"
    	stream = "my_stream"
		format_setting = {
			"data.datetime.format_name" = "POSIX"
		}
	    column {
			name = "z"
			type = "int8"
			not_null = true
		}
		column {
			name = "z2"
			type = "UTF8"
		}
	}`, map[string]interface{}{
		"ConnectionName":         connectionName,
		"ConnectionResourceName": connectionResourceName,
		"BindingName":            bindingName,
		"BindingResourceName":    bindingResourceName,
	})
}
