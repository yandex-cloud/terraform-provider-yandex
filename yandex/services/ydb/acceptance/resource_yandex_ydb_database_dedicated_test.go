package acceptance

import (
	"context"
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
	ydbsdk "github.com/yandex-cloud/go-sdk/services/ydb/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

const (
	ydbDatabaseDedicatedResource = "yandex_ydb_database_dedicated.test-ydb-database-dedicated"
	ydbLocationId                = "ru-central1"
)

func TestAccYandexYDBDatabaseDedicated_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")
	labelKey := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label")
	labelValue := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value")
	deletionProtection := "false"
	ydbLocationId := ydbLocationId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			suite.basicYandexYDBDatabaseDedicatedTestStep(databaseName, databaseDesc, deletionProtection, labelKey, labelValue, ydbLocationId, &database),
		},
	})
}

func TestAccYandexYDBDatabaseDedicated_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")
	labelKey := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label")
	labelValue := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value")
	deletionProtection := "true"
	ydbLocationId := ydbLocationId

	databaseNameUpdated := acctest.RandomWithPrefix("tf-ydb-database-dedicated-updated")
	databaseDescUpdated := acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc-updated")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-updated")
	labelValueUpdated := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value-updated")
	deletionProtectionUpdated := "false"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			suite.basicYandexYDBDatabaseDedicatedTestStep(databaseName, databaseDesc, deletionProtection, labelKey, labelValue, ydbLocationId, &database),
			suite.basicYandexYDBDatabaseDedicatedTestStep(databaseNameUpdated, databaseDescUpdated, deletionProtectionUpdated, labelKeyUpdated, labelValueUpdated, ydbLocationId, &database),
		},
	})
}

func TestAccYandexYDBDatabaseDedicated_full(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var database ydb.Database
	params := testYandexYDBDatabaseDedicatedParameters{}
	params.name = acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	params.desc = acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label")
	params.labelValue = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value")
	params.deletionProtection = "true"
	params.ydbLocationId = ydbLocationId

	paramsUpdated := testYandexYDBDatabaseDedicatedParameters{}
	paramsUpdated.name = acctest.RandomWithPrefix("tf-ydb-database-dedicated-updated")
	paramsUpdated.desc = acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc-updated")
	paramsUpdated.labelKey = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-updated")
	paramsUpdated.labelValue = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value-updated")
	paramsUpdated.deletionProtection = "false"
	paramsUpdated.ydbLocationId = ydbLocationId

	testConfigFunc := func(params testYandexYDBDatabaseDedicatedParameters) resource.TestStep {
		return resource.TestStep{
			Config: testYandexYDBDatabaseDedicatedFull(params),
			Check: resource.ComposeTestCheckFunc(
				suite.testYandexYDBDatabaseDedicatedExists(ydbDatabaseDedicatedResource, &database),
				resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "name", params.name),
				resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "description", params.desc),
				resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "deletion_protection", params.deletionProtection),
				resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedResource, "folder_id"),
				testYandexYDBDatabaseDedicatedContainsLabel(&database, params.labelKey, params.labelValue),
				statecheck.CheckCreatedAtAttr(ydbDatabaseDedicatedResource),
			),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			testConfigFunc(params),
			testConfigFunc(paramsUpdated),
		},
	})
}

func (suite testSuite) basicYandexYDBDatabaseDedicatedTestStep(databaseName, databaseDesc, deletionProtection, labelKey, labelValue, ydbLocationId string, database *ydb.Database) resource.TestStep {
	return resource.TestStep{
		Config: testYandexYDBDatabaseDedicatedBasic(databaseName, databaseDesc, deletionProtection, labelKey, labelValue, ydbLocationId),
		Check: resource.ComposeTestCheckFunc(
			suite.testYandexYDBDatabaseDedicatedExists(ydbDatabaseDedicatedResource, database),
			resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "name", databaseName),
			resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "description", databaseDesc),
			resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "deletion_protection", deletionProtection),
			resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedResource, "folder_id"),
			testYandexYDBDatabaseDedicatedContainsLabel(database, labelKey, labelValue),
			statecheck.CheckCreatedAtAttr(ydbDatabaseDedicatedResource),
		),
	}
}

func (suite testSuite) testYandexYDBDatabaseDedicatedDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_ydb_database_dedicated" {
			continue
		}

		_, err := testGetYDBDatabaseDedicatedByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("YDB dedicated database still exists")
		}
	}

	return nil
}

func (suite testSuite) testYandexYDBDatabaseDedicatedExists(name string, database *ydb.Database) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		found, err := testGetYDBDatabaseDedicatedByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("YDB dedicated database not found")
		}

		*database = *found
		return nil
	}
}

func testGetYDBDatabaseDedicatedByID(config *providerconfig.Config, ID string) (*ydb.Database, error) {
	req := ydb.GetDatabaseRequest{
		DatabaseId: ID,
	}

	client := ydbsdk.NewDatabaseClient(config.SDK)

	return client.Get(context.Background(), &req)
}

func testYandexYDBDatabaseDedicatedContainsLabel(database *ydb.Database, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := database.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

const ydbDatabaseDedicatedDependencies = `
resource "yandex_vpc_network" "ydb-db-dedicated-test-net" {}

resource "yandex_vpc_subnet" "ydb-db-dedicated-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "ydb-db-dedicated-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "ydb-db-dedicated-test-subnet-d" {
  zone           = "ru-central1-d"
  network_id     = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  v4_cidr_blocks = ["10.4.0.0/24"]
}

resource "yandex_vpc_security_group" "ydb-db-dedicated-test-security-group" {
  network_id     = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  ingress {
    protocol = "TCP"
        v6_cidr_blocks = ["::/0"]
        v4_cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    protocol = "TCP"
        v6_cidr_blocks = ["::/0"]
        v4_cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "yandex_vpc_security_group" "ydb-db-dedicated-test-security-group-second" {
  network_id     = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  ingress {
    protocol = "TCP"
        v6_cidr_blocks = ["::/0"]
        v4_cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    protocol = "TCP"
        v6_cidr_blocks = ["::/0"]
        v4_cidr_blocks = ["0.0.0.0/0"]
  }
}
`

func testYandexYDBDatabaseDedicatedBasic(name, desc, deletionProtection, labelKey, labelValue, ydbLocationId string) string {
	return fmt.Sprintf(ydbDatabaseDedicatedDependencies+`
resource "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name        = "%s"
  description = "%s"

  deletion_protection = %s

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
`, name, desc, deletionProtection, labelKey, labelValue, ydbLocationId)
}

type testYandexYDBDatabaseDedicatedParameters struct {
	name               string
	desc               string
	labelKey           string
	labelValue         string
	deletionProtection string
	ydbLocationId      string
}

func testYandexYDBDatabaseDedicatedFull(params testYandexYDBDatabaseDedicatedParameters) string {
	return fmt.Sprintf(ydbDatabaseDedicatedDependencies+`
resource "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name        = "%s"
  description = "%s"

  deletion_protection = %s

  labels = {
    %s          = "%s"
    empty-label = ""
  }

  location_id = "%s"

  resource_preset_id  = "medium"

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
		params.deletionProtection,
		params.labelKey,
		params.labelValue,
		params.ydbLocationId)
}
