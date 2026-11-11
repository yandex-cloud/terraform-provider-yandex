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

const ydbDatabaseDedicatedDataSource = "data.yandex_ydb_database_dedicated.test-ydb-database-dedicated"

func TestAccDataSourceYandexYDBDatabaseDedicated_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")
	ydbLocationId := ydbLocationId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseDedicatedByID(databaseName, databaseDesc, ydbLocationId),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexYDBDatabaseDedicatedExists(ydbDatabaseDedicatedDataSource, &database),
					resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedDataSource, "database_id"),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "name", databaseName),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "description", databaseDesc),
					resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedDataSource, "folder_id"),
					statecheck.CheckCreatedAtAttr(ydbDatabaseDedicatedDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexYDBDatabaseDedicated_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")
	ydbLocationId := ydbLocationId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseDedicatedByName(databaseName, databaseDesc, ydbLocationId),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexYDBDatabaseDedicatedExists(ydbDatabaseDedicatedDataSource, &database),
					resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedDataSource, "database_id"),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "name", databaseName),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "description", databaseDesc),
					resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedDataSource, "folder_id"),
					statecheck.CheckCreatedAtAttr(ydbDatabaseDedicatedDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexYDBDatabaseDedicated_full(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var database ydb.Database
	params := testYandexYDBDatabaseDedicatedParameters{}
	params.name = acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	params.desc = acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label")
	params.labelValue = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value")
	params.ydbLocationId = ydbLocationId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexYDBDatabaseDedicatedDataSource(params),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexYDBDatabaseDedicatedExists(ydbDatabaseDedicatedDataSource, &database),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "name", params.name),
					resource.TestCheckResourceAttr(ydbDatabaseDedicatedDataSource, "description", params.desc),
					resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedDataSource, "folder_id"),
					testYandexYDBDatabaseDedicatedContainsLabel(&database, params.labelKey, params.labelValue),
					statecheck.CheckCreatedAtAttr(ydbDatabaseDedicatedDataSource),
				),
			},
		},
	})
}

func testYandexYDBDatabaseDedicatedByID(name, desc, ydbLocationId string) string {
	return fmt.Sprintf(ydbDatabaseDedicatedDependencies+`
data "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  database_id = "${yandex_ydb_database_dedicated.test-ydb-database-dedicated.id}"
}

resource "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name               = "%s"
  description        = "%s"
  resource_preset_id = "medium"

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  location_id = "%s"

  storage_config {
    group_count     = 1
    storage_type_id = "ssd"
  }

  network_id = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  subnet_ids = [
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-a.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-b.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-d.id}",
  ]
  security_group_ids = [
    "${yandex_vpc_security_group.ydb-db-dedicated-test-security-group.id}",
    "${yandex_vpc_security_group.ydb-db-dedicated-test-security-group-second.id}",
  ]
}`, name, desc, ydbLocationId)
}

func testYandexYDBDatabaseDedicatedByName(name, desc, ydbLocationId string) string {
	return fmt.Sprintf(ydbDatabaseDedicatedDependencies+`
data "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name = "${yandex_ydb_database_dedicated.test-ydb-database-dedicated.name}"
}

resource "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name               = "%s"
  description        = "%s"
  resource_preset_id = "medium"

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  location_id = "%s"

  storage_config {
    group_count     = 1
    storage_type_id = "ssd"
  }

  network_id = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  subnet_ids = [
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-a.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-b.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-d.id}",
  ]
  security_group_ids = [
    "${yandex_vpc_security_group.ydb-db-dedicated-test-security-group.id}",
    "${yandex_vpc_security_group.ydb-db-dedicated-test-security-group-second.id}",
  ]
}
`, name, desc, ydbLocationId)
}

func testYandexYDBDatabaseDedicatedDataSource(params testYandexYDBDatabaseDedicatedParameters) string {
	return fmt.Sprintf(ydbDatabaseDedicatedDependencies+`
data "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  database_id = "${yandex_ydb_database_dedicated.test-ydb-database-dedicated.id}"
}

resource "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name        = "%s"
  description = "%s"

  labels = {
    %s          = "%s"
    empty-label = ""
  }

  location_id = "%s"

  resource_preset_id = "medium"

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  storage_config {
    group_count     = 1
    storage_type_id = "ssd"
  }

  network_id = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  subnet_ids = [
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-a.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-b.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-d.id}",
  ]
  security_group_ids = [
    "${yandex_vpc_security_group.ydb-db-dedicated-test-security-group.id}",
    "${yandex_vpc_security_group.ydb-db-dedicated-test-security-group-second.id}",
  ]
}
`,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue,
		params.ydbLocationId)
}
