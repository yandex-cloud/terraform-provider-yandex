package testfixtures

import (
	"fmt"
	"os"
)

func ServiceAccountWithEditor(randInt int) string {
	return ServiceAccountWithRole(randInt, "editor")
}

func ServiceAccountWithStorageAdmin(randInt int) string {
	return ServiceAccountWithRole(randInt, "storage.admin")
}

func ServiceAccountWithRole(randInt int, role string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "sa" {
	name = "test-sa-for-tf-test-%[1]d"
}

resource "yandex_resourcemanager_folder_iam_member" "binding" {
	folder_id   = "%[3]s"
	member      = "serviceAccount:${yandex_iam_service_account.sa.id}"
	role        = "%[2]s"
	sleep_after = 30
}

resource "yandex_iam_service_account_static_access_key" "sa-key" {
	service_account_id = "${yandex_iam_service_account.sa.id}"

	depends_on = [
		yandex_resourcemanager_folder_iam_member.binding
	]
}
`, randInt, role, os.Getenv("YC_FOLDER_ID"))
}
