package yandex_iam_oauth_client_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const oauthClientResourceName = "yandex_iam_oauth_client.test-oauth-client"

func TestAccIAMOauthClient_full(t *testing.T) {
	var (
		clientName        = test.ResourceName(63)
		clientNameUpdated = test.ResourceName(63)

		folderID = test.GetExampleFolderID()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			oauthClientBaseTestStep(clientName, folderID),
			oauthClientImportTestStep(),
			oauthClientBaseTestStep(clientNameUpdated, folderID),
			oauthClientImportTestStep(),
		},
	})
}

func oauthClientBaseTestStep(clientName, folderID string) resource.TestStep {
	return resource.TestStep{
		Config: testOAuthClientResourceFullConfig(clientName, folderID),
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(oauthClientResourceName, "name", clientName),
			resource.TestCheckTypeSetElemAttr(oauthClientResourceName, "redirect_uris.*", "https://localhost"),
			resource.TestCheckTypeSetElemAttr(oauthClientResourceName, "scopes.*", "openid"),
			resource.TestCheckResourceAttr(oauthClientResourceName, "folder_id", folderID),
		),
	}
}

func oauthClientImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      oauthClientResourceName,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testOAuthClientResourceFullConfig(clientName, folderID string) string {
	return fmt.Sprintf(`
resource "yandex_iam_oauth_client" "test-oauth-client" {
  name          = "%s"
  folder_id 	= "%s"
  redirect_uris = ["https://localhost"]
  scopes        = ["openid"]
}`, clientName, folderID)
}
