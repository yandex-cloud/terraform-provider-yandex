package yandex_connectionmanager_connection_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const testConnectionDataSourceName = "data.yandex_connectionmanager_connection.test-connection-data"

func TestConnectionManagerConnectionDataSource(t *testing.T) {
	var (
		folderId              = testhelpers.GetExampleFolderID()
		connectionName        = testhelpers.ResourceName(63)
		connectionDescription = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey              = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue            = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
		params                = testConnectionPostgresParams(
			"user",
			testPasswordParams("password"),
		)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             AccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConnectionDataConfig(folderId, connectionName, connectionDescription, labelKey, labelValue, params),
				Check: resource.ComposeTestCheckFunc(
					ConnectionExists(testConnectionDataSourceName),
					resource.TestCheckResourceAttr(testConnectionDataSourceName, "folder_id", folderId),
					resource.TestCheckResourceAttr(testConnectionDataSourceName, "name", connectionName),
					resource.TestCheckResourceAttr(testConnectionDataSourceName, "description", connectionDescription),
					resource.TestCheckResourceAttr(testConnectionDataSourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
					resource.TestCheckResourceAttrSet(testConnectionDataSourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testConnectionDataSourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(testConnectionDataSourceName, "created_by"),

					resource.TestCheckResourceAttrSet(testConnectionDataSourceName, "params.postgresql.managed_cluster_id"),
					resource.TestCheckResourceAttrSet(testConnectionDataSourceName, "params.postgresql.auth.user_password.user"),
				),
			},
		},
	})
}

func testConnectionDataConfig(folder_id, name, description, labelKey, labelValue, params string) string {
	return fmt.Sprintf(`
data "yandex_connectionmanager_connection" "test-connection-data" {
	connection_id = yandex_connectionmanager_connection.test-connection.id
}

variable "postgres_password" {
	type        = string
	sensitive   = true
}

resource "yandex_connectionmanager_connection" "test-connection" {
  folder_id = "%s"
  name = "%s"
  description = "%s"
  labels = {
    "%s" = "%s"
  }
  params = %s
}
`, folder_id, name, description, labelKey, labelValue, params)
}
