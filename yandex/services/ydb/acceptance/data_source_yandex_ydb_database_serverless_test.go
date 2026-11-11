package acceptance

import (
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

const ydbDatabaseServerlessDataSource = "data.yandex_ydb_database_serverless.test-ydb-database-serverless"

func TestAccDataSourceYandexYDBDatabaseServerless_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-serverless")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-serverless-desc")
	ydbLocationId := ydbLocationId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseServerlessByID(databaseName, databaseDesc, ydbLocationId),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexYDBDatabaseServerlessExists(ydbDatabaseServerlessDataSource, &database),
					resource.TestCheckResourceAttrSet(ydbDatabaseServerlessDataSource, "database_id"),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "name", databaseName),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "description", databaseDesc),
					resource.TestCheckResourceAttrSet(ydbDatabaseServerlessDataSource, "folder_id"),
					statecheck.CheckCreatedAtAttr(ydbDatabaseServerlessDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexYDBDatabaseServerless_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-serverless")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-serverless-desc")
	ydbLocationId := ydbLocationId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseServerlessByName(databaseName, databaseDesc, ydbLocationId),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexYDBDatabaseServerlessExists(ydbDatabaseServerlessDataSource, &database),
					resource.TestCheckResourceAttrSet(ydbDatabaseServerlessDataSource, "database_id"),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "name", databaseName),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "description", databaseDesc),
					resource.TestCheckResourceAttrSet(ydbDatabaseServerlessDataSource, "folder_id"),
					statecheck.CheckCreatedAtAttr(ydbDatabaseServerlessDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexYDBDatabaseServerless_full(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var database ydb.Database
	params := testYandexYDBDatabaseServerlessParameters{}
	params.name = acctest.RandomWithPrefix("tf-ydb-database-serverless")
	params.desc = acctest.RandomWithPrefix("tf-ydb-database-serverless-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-ydb-database-serverless-label")
	params.labelValue = acctest.RandomWithPrefix("tf-ydb-database-serverless-label-value")
	params.ydbLocationId = ydbLocationId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseServerlessDataSource(params),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexYDBDatabaseServerlessExists(ydbDatabaseServerlessDataSource, &database),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "name", params.name),
					resource.TestCheckResourceAttr(ydbDatabaseServerlessDataSource, "description", params.desc),
					resource.TestCheckResourceAttrSet(ydbDatabaseServerlessDataSource, "folder_id"),
					testYandexYDBDatabaseServerlessContainsLabel(&database, params.labelKey, params.labelValue),
					statecheck.CheckCreatedAtAttr(ydbDatabaseServerlessDataSource),
				),
			},
		},
	})
}

func testYandexYDBDatabaseServerlessByID(name, desc, ydbLocationId string) string {
	return fmt.Sprintf(`
data "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  database_id = "${yandex_ydb_database_serverless.test-ydb-database-serverless.id}"
}

resource "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  name        = "%s"
  description = "%s"
  location_id = "%s"
}`, name, desc, ydbLocationId)
}

func testYandexYDBDatabaseServerlessByName(name, desc, ydbLocationId string) string {
	return fmt.Sprintf(`
data "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  name = "${yandex_ydb_database_serverless.test-ydb-database-serverless.name}"
}

resource "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  name        = "%s"
  description = "%s"
  location_id = "%s"
}
`, name, desc, ydbLocationId)
}

func testYandexYDBDatabaseServerlessDataSource(params testYandexYDBDatabaseServerlessParameters) string {
	return fmt.Sprintf(`
data "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  database_id = "${yandex_ydb_database_serverless.test-ydb-database-serverless.id}"
}

resource "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }

  location_id = "%s"
}
`,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue,
		params.ydbLocationId)
}
