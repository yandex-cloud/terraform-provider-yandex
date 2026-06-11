package datalens_chart_test

import (
	"fmt"
	"os"
)

// datalensTestYDB — see datalens_dataset/ydb_fixture_test.go for full doc.
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
