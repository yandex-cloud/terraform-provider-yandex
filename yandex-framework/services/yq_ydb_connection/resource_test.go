package yq_ydb_connection_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccYQYDBConnectionBasic(t *testing.T) {
	connectionName := fmt.Sprintf("my-conn-%s", acctest.RandString(5))
	connectionResourceName := "my-connection"
	existingConnectionResourceName := fmt.Sprintf("yandex_yq_ydb_connection.%s", connectionResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return test.TestYandexYQAllConnectionsDestroyed(s, "yandex_yq_ydb_connection")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccYQYDBConnectionConfig(connectionName, connectionResourceName),
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

func testAccYQYDBConnectionConfig(connectionName string, connectionResourceName string) string {
	return fmt.Sprintf(`
	resource "yandex_iam_service_account" "foo" {
  		name        = "sa-%s"
	}

	resource "yandex_yq_ydb_connection" "%s" {
        name = "%s"
		description = "my_desc"
        database_id = "abc123"
		service_account_id = yandex_iam_service_account.foo.id
    }`,
		connectionName,
		connectionResourceName,
		connectionName,
	)
}
