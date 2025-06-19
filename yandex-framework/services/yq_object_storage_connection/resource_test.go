package yq_object_storage_connection_test

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

func TestAccYQObjectStorageConnectionBasic(t *testing.T) {
	connectionName := fmt.Sprintf("my-conn-%s", acctest.RandString(5))
	connectionResourceName := "my-connection"
	existingConnectionResourceName := fmt.Sprintf("yandex_yq_object_storage_connection.%s", connectionResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return test.TestYandexYQAllConnectionsDestroyed(s, "yandex_yq_object_storage_connection")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccYQObjectStorageConnectionConfig(connectionName, connectionResourceName),
				Check: resource.ComposeTestCheckFunc(
					test.TestAccYQConnectionExists(connectionName, existingConnectionResourceName),
				),
			},
			{
				ResourceName:      existingConnectionResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccYQObjectStorageConnectionConfig(connectionName string, connectionResourceName string) string {
	return fmt.Sprintf(`
	resource "yandex_yq_object_storage_connection" "%s" {
        name = "%s"
		description = "my_desc"
        bucket = "my_bucket"
    }`,
		connectionResourceName,
		connectionName,
	)
}
