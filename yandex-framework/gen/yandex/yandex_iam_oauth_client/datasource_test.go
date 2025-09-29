package yandex_iam_oauth_client_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const oauthClientDataSourceName = "data.yandex_iam_oauth_client.test-oauth-client-data"

func TestAccDataSourceIAMOauthClient(t *testing.T) {
	var (
		clientName = test.ResourceName(63)

		folderID = test.GetExampleFolderID()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testOAuthClientDataSourceFullConfig(clientName, folderID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(oauthClientDataSourceName, "name", clientName),
					resource.TestCheckTypeSetElemAttr(oauthClientDataSourceName, "redirect_uris.*", "https://localhost"),
					resource.TestCheckTypeSetElemAttr(oauthClientDataSourceName, "scopes.*", "iam"),
					resource.TestCheckResourceAttr(oauthClientDataSourceName, "folder_id", folderID),
				),
			},
			oauthClientImportTestStep(),
		},
	})
}

func testOAuthClientDataSourceFullConfig(clientName, folderID string) string {
	return fmt.Sprintf(`
data "yandex_iam_oauth_client" "test-oauth-client-data" {
  oauth_client_id = yandex_iam_oauth_client.test-oauth-client.id
}

resource "yandex_iam_oauth_client" "test-oauth-client" {
  name          = "%s"
  folder_id 	= "%s"
  redirect_uris = ["https://localhost"]
  scopes        = ["iam"]
}`, clientName, folderID)
}
