package datalens_dataset_test

import (
	"fmt"
	"os"
)

// datalensTestYDB returns the HCL resource block for a fresh YDB serverless
// database used as a backing source in DataLens acceptance tests, plus the
// expressions to reference its host and database path.
func datalensTestYDB(name string) (resourceBlock, hostExpr, dbPathExpr string) {
	resourceBlock = fmt.Sprintf(`
resource "yandex_ydb_database_serverless" "test" {
  name        = %q
  folder_id   = %q
  location_id = "ru-central1"
}
`, name, os.Getenv("YC_FOLDER_ID"))
	return resourceBlock,
		`split(":", yandex_ydb_database_serverless.test.ydb_api_endpoint)[0]`,
		`yandex_ydb_database_serverless.test.database_path`
}
