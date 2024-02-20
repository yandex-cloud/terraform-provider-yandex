package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
)

func init() {
	resource.AddTestSweepers("yandex_ydb_database_serverless", &resource.Sweeper{
		Name: "yandex_ydb_database_serverless",
		F:    testSweepYDBDatabaseServerless,
	})
}

func testSweepYDBDatabaseServerless(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.YDB().Database().List(conf.Context(), &ydb.ListDatabasesRequest{
		FolderId: conf.FolderID,
		PageSize: 1000,
	})
	if err != nil {
		return fmt.Errorf("error getting YDB databases: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Databases {
		// only serverless, other types are swept in a separate sweeper
		if _, ok := c.DatabaseType.(*ydb.Database_ServerlessDatabase); ok {
			if !sweepYDBServerlessDatabase(conf, c.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep YDB serverless database %q", c.Id))
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepYDBServerlessDatabase(conf *Config, id string) bool {
	return sweepWithRetry(sweepYDBServerlessDatabaseOnce, conf, "YDB serverless database", id)
}

func sweepYDBServerlessDatabaseOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexYDBServerlessDefaultTimeout)
	defer cancel()

	err := checkAndUnsetYDBDeletionProtection(conf, ctx, id)
	if err != nil {
		return err
	}

	op, err := conf.sdk.YDB().Database().Delete(ctx, &ydb.DeleteDatabaseRequest{
		DatabaseId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccYandexYDBDatabaseServerless_basic(t *testing.T) {
	t.Parallel()
	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-serverless")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-serverless-desc")
	labelKey := acctest.RandomWithPrefix("tf-ydb-database-serverless-label")
	labelValue := acctest.RandomWithPrefix("tf-ydb-database-serverless-label-value")
	ydbResourceName := fmt.Sprintf("test-ydb-database-serverless-%s", acctest.RandString(5))
	deletionProtection := "false"
	ydbLocationId := ydbLocationId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			basicYandexYDBDatabaseServerlessTestStep(databaseName, databaseDesc, deletionProtection, labelKey, labelValue, &database, ydbResourceName, ydbLocationId),
		},
	})
}

func TestAccYandexYDBDatabaseServerless_update(t *testing.T) {
	t.Parallel()
	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-serverless")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-serverless-desc")
	ydbResourceName := fmt.Sprintf("test-ydb-database-serverless-%s", acctest.RandString(5))
	labelKey := acctest.RandomWithPrefix("tf-ydb-database-serverless-label")
	labelValue := acctest.RandomWithPrefix("tf-ydb-database-serverless-label-value")
	deletionProtection := "true"
	ydbLocationId := ydbLocationId

	databaseNameUpdated := acctest.RandomWithPrefix("tf-ydb-database-serverless-updated")
	databaseDescUpdated := acctest.RandomWithPrefix("tf-ydb-database-serverless-desc-updated")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-ydb-database-serverless-label-updated")
	labelValueUpdated := acctest.RandomWithPrefix("tf-ydb-database-serverless-label-value-updated")
	deletionProtectionUpdated := "false"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			basicYandexYDBDatabaseServerlessTestStep(databaseName, databaseDesc, deletionProtection, labelKey, labelValue, &database, ydbResourceName, ydbLocationId),
			basicYandexYDBDatabaseServerlessTestStep(databaseNameUpdated, databaseDescUpdated, deletionProtectionUpdated, labelKeyUpdated, labelValueUpdated, &database, ydbResourceName, ydbLocationId),
		},
	})
}

func TestAccYandexYDBDatabaseServerless_full(t *testing.T) {
	t.Parallel()

	var database ydb.Database
	ydbResourceName := fmt.Sprintf("test-ydb-database-serverless-%s", acctest.RandString(5))
	params := testYandexYDBDatabaseServerlessParameters{}
	params.name = acctest.RandomWithPrefix("tf-ydb-database-serverless")
	params.desc = acctest.RandomWithPrefix("tf-ydb-database-serverless-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-ydb-database-serverless-label")
	params.labelValue = acctest.RandomWithPrefix("tf-ydb-database-serverless-label-value")
	params.deletionProtection = "true"
	params.ydbLocationId = ydbLocationId

	paramsUpdated := testYandexYDBDatabaseServerlessParameters{}
	paramsUpdated.name = acctest.RandomWithPrefix("tf-ydb-database-serverless-updated")
	paramsUpdated.desc = acctest.RandomWithPrefix("tf-ydb-database-serverless-desc-updated")
	paramsUpdated.labelKey = acctest.RandomWithPrefix("tf-ydb-database-serverless-label-updated")
	paramsUpdated.labelValue = acctest.RandomWithPrefix("tf-ydb-database-serverless-label-value-updated")
	paramsUpdated.deletionProtection = "false"
	paramsUpdated.ydbLocationId = ydbLocationId
	key := fmt.Sprintf("yandex_ydb_database_serverless.%s", ydbResourceName)

	testConfigFunc := func(params testYandexYDBDatabaseServerlessParameters) resource.TestStep {
		return resource.TestStep{
			Config: testYandexYDBDatabaseServerlessFull(params, ydbResourceName),
			Check: resource.ComposeTestCheckFunc(
				testYandexYDBDatabaseServerlessExists(key, &database),
				resource.TestCheckResourceAttr(key, "name", params.name),
				resource.TestCheckResourceAttr(key, "description", params.desc),
				resource.TestCheckResourceAttr(key, "deletion_protection", params.deletionProtection),
				resource.TestCheckResourceAttrSet(key, "folder_id"),
				resource.TestCheckResourceAttr(key, "serverless_database.#", "1"),
				resource.TestCheckResourceAttr(key, "serverless_database.0.throttling_rcu_limit", "30"),
				resource.TestCheckResourceAttr(key, "serverless_database.0.storage_size_limit", "90"),
				resource.TestCheckResourceAttr(key, "serverless_database.0.enable_throttling_rcu_limit", "true"),
				resource.TestCheckResourceAttr(key, "serverless_database.0.provisioned_rcu_limit", "50"),
				testYandexYDBDatabaseServerlessContainsLabel(&database, params.labelKey, params.labelValue),
				testAccCheckCreatedAtAttr(key),
			),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			testConfigFunc(params),
			testConfigFunc(paramsUpdated),
		},
	})
}

func basicYandexYDBDatabaseServerlessTestStep(
	databaseName,
	databaseDesc,
	deletionProtection,
	labelKey,
	labelValue string,
	database *ydb.Database,
	ydbResourceName,
	ydbLocationId string,
) resource.TestStep {
	key := fmt.Sprintf("yandex_ydb_database_serverless.%s", ydbResourceName)
	return resource.TestStep{
		Config: testYandexYDBDatabaseServerlessBasic(databaseName, databaseDesc, deletionProtection, labelKey, labelValue, ydbResourceName, ydbLocationId),
		Check: resource.ComposeTestCheckFunc(
			testYandexYDBDatabaseServerlessExists(key, database),
			resource.TestCheckResourceAttr(key, "name", databaseName),
			resource.TestCheckResourceAttr(key, "description", databaseDesc),
			resource.TestCheckResourceAttr(key, "deletion_protection", deletionProtection),
			resource.TestCheckResourceAttrSet(key, "folder_id"),
			resource.TestCheckResourceAttr(key, "serverless_database.#", "1"),
			testYandexYDBDatabaseServerlessContainsLabel(database, labelKey, labelValue),
			testAccCheckCreatedAtAttr(key),
		),
	}
}

func testYandexYDBDatabaseServerlessDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_ydb_database_serverless" {
			continue
		}

		_, err := testGetYDBDatabaseServerlessByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("YDB serverless database still exists")
		}
	}

	return nil
}

func testYandexYDBDatabaseServerlessExists(name string, database *ydb.Database) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetYDBDatabaseServerlessByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("YDB serverless database not found")
		}

		*database = *found
		return nil
	}
}

func testGetYDBDatabaseServerlessByID(config *Config, ID string) (*ydb.Database, error) {
	req := ydb.GetDatabaseRequest{
		DatabaseId: ID,
	}

	return config.sdk.YDB().Database().Get(context.Background(), &req)
}

func testYandexYDBDatabaseServerlessContainsLabel(database *ydb.Database, key string, value string) resource.TestCheckFunc {
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

func testYandexYDBDatabaseServerlessBasic(name, desc, deletionProtection, labelKey, labelValue, ydbResourceName, ydbLocationId string) string {
	return fmt.Sprintf(`
resource "yandex_ydb_database_serverless" "%s" {
  name        = "%s"
  description = "%s"

  deletion_protection = %s

  labels = {
    %s          = "%s"
    empty-label = ""
  }

  location_id = "%s"
}
`, ydbResourceName, name, desc, deletionProtection, labelKey, labelValue, ydbLocationId)
}

type testYandexYDBDatabaseServerlessParameters struct {
	name               string
	desc               string
	labelKey           string
	labelValue         string
	deletionProtection string
	ydbLocationId      string
}

func testYandexYDBDatabaseServerlessFull(params testYandexYDBDatabaseServerlessParameters, ydbResourceName string) string {
	return fmt.Sprintf(`
resource "yandex_ydb_database_serverless" "%s" {
  name        = "%s"
  description = "%s"

  deletion_protection = %s

  labels = {
    %s          = "%s"
    empty-label = ""
  }

  location_id = "%s"

  serverless_database {
    throttling_rcu_limit        = 30
    storage_size_limit          = 90
    enable_throttling_rcu_limit = true
    provisioned_rcu_limit       = 50
  }
}
`,
		ydbResourceName,
		params.name,
		params.desc,
		params.deletionProtection,
		params.labelKey,
		params.labelValue,
		params.ydbLocationId)
}
