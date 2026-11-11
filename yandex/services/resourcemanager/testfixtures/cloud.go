package testfixtures

import (
	"fmt"
)

func CloudDataSourceByID(cloudID string) string {
	return fmt.Sprintf(`
data "yandex_resourcemanager_cloud" "acceptance" {
  cloud_id = "%s"
}
`, cloudID)
}
