package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
)

const ydbDatabaseServerlessResource = "yandex_ydb_database_serverless.test-ydb-database-serverless"

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
	deletionProtection := "false"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			basicYandexYDBDatabaseServerlessTestStep(databaseName, databaseDesc, deletionProtection, labelKey, labelValue, &database),
		},
	})
}

func TestAccYandexYDBDatabaseServerless_update(t *testing.T) {
	t.Parallel()

	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-serverless")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-serverless-desc")
	labelKey := acctest.RandomWithPrefix("tf-ydb-database-serverless-label")
	labelValue := acctest.RandomWithPrefix("tf-ydb-database-serverless-label-value")
	deletionProtection := "true"

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
			basicYandexYDBDatabaseServerlessTestStep(databaseName, databaseDesc, deletionProtection, labelKey, labelValue, &database),
			basicYandexYDBDatabaseServerlessTestStep(databaseNameUpdated, databaseDescUpdated, deletionProtectionUpdated, labelKeyUpdated, labelValueUpdated, &database),
		},
	})
}

func TestAccYandexYDBDatabaseServerless_full(t *testing.T) {
	t.Parallel()

	var database ydb.Database
	params := testYandexYDBDatabaseServerlessParameters{}
	params.name = acctest.RandomWithPrefix("tf-ydb-database-serverless")
	params.desc = acctest.RandomWithPrefix("tf-ydb-database-serverless-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-ydb-database-serverless-label")
	params.labelValue = acctest.RandomWithPrefix("tf-ydb-database-serverless-label-value")
	params.deletionProtection = "true"

	paramsUpdated := testYandexYDBDatabaseServerlessParameters{}
	paramsUpdated.name = acctest.RandomWithPrefix("tf-ydb-database-serverless-updated")
	paramsUpdated.desc = acctest.RandomWithPrefix("tf-ydb-database-serverless-desc-updated")
	paramsUpdated.labelKey = acctest.RandomWithPrefix("tf-ydb-database-serverless-label-updated")
	paramsUpdated.labelValue = acctest.RandomWithPrefix("tf-ydb-database-serverless-label-value-updated")
	paramsUpdated.deletionProtection = "false"

	testConfigFunc := func(params testYandexYDBDatabaseServerlessParameters) resource.TestStep {
		return resource.TestStep{
			Config: testYandexYDBDatabaseServerlessFull(params),
			Check: resource.ComposeTestCheckFunc(
				testYandexYDBDatabaseServerlessExists(ydbDatabaseServerlessResource, &database),
				resource.TestCheckResourceAttr(ydbDatabaseServerlessResource, "name", params.name),
				resource.TestCheckResourceAttr(ydbDatabaseServerlessResource, "description", params.desc),
				resource.TestCheckResourceAttr(ydbDatabaseServerlessResource, "deletion_protection", params.deletionProtection),
				resource.TestCheckResourceAttrSet(ydbDatabaseServerlessResource, "folder_id"),
				testYandexYDBDatabaseServerlessContainsLabel(&database, params.labelKey, params.labelValue),
				testAccCheckCreatedAtAttr(ydbDatabaseServerlessResource),
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

func basicYandexYDBDatabaseServerlessTestStep(databaseName, databaseDesc, deletionProtection, labelKey, labelValue string, database *ydb.Database) resource.TestStep {
	return resource.TestStep{
		Config: testYandexYDBDatabaseServerlessBasic(databaseName, databaseDesc, deletionProtection, labelKey, labelValue),
		Check: resource.ComposeTestCheckFunc(
			testYandexYDBDatabaseServerlessExists(ydbDatabaseServerlessResource, database),
			resource.TestCheckResourceAttr(ydbDatabaseServerlessResource, "name", databaseName),
			resource.TestCheckResourceAttr(ydbDatabaseServerlessResource, "description", databaseDesc),
			resource.TestCheckResourceAttr(ydbDatabaseServerlessResource, "deletion_protection", deletionProtection),
			resource.TestCheckResourceAttrSet(ydbDatabaseServerlessResource, "folder_id"),
			testYandexYDBDatabaseServerlessContainsLabel(database, labelKey, labelValue),
			testAccCheckCreatedAtAttr(ydbDatabaseServerlessResource),
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

func testYandexYDBDatabaseServerlessBasic(name string, desc string, deletionProtection string, labelKey string, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  name        = "%s"
  description = "%s"

  deletion_protection = %s

  labels = {
    %s          = "%s"
    empty-label = ""
  }
}
`, name, desc, deletionProtection, labelKey, labelValue)
}

type testYandexYDBDatabaseServerlessParameters struct {
	name               string
	desc               string
	labelKey           string
	labelValue         string
	deletionProtection string
}

func testYandexYDBDatabaseServerlessFull(params testYandexYDBDatabaseServerlessParameters) string {
	return fmt.Sprintf(`
resource "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  name        = "%s"
  description = "%s"

  deletion_protection = %s

  labels = {
    %s          = "%s"
    empty-label = ""
  }
}
`,
		params.name,
		params.desc,
		params.deletionProtection,
		params.labelKey,
		params.labelValue)
}
