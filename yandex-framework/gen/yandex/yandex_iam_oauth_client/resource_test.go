package yandex_iam_oauth_client_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	iam "github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const oauthClientResourceName = "yandex_iam_oauth_client.test-oauth-client"

func init() {
	resource.AddTestSweepers("yandex_iam_oauth_client", &resource.Sweeper{
		Name:         "yandex_iam_oauth_client",
		F:            testSweepOauthClient,
		Dependencies: []string{},
	})
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepOauthClient(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := iamsdk.NewOAuthClientClient(conf.SDKv2).List(context.Background(), &iam.ListOAuthClientsRequest{
		FolderId: test.GetExampleFolderID(),
	})
	if err != nil {
		return fmt.Errorf("error getting OAuthClients: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.OauthClients {
		if !sweepOauthClient(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep OAuth client %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepOauthClient(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepOauthClientOnce, conf, "OAuthClient", id)
}

func sweepOauthClientOnce(conf *provider_config.Config, id string) error {
	op, err := iamsdk.NewOAuthClientClient(conf.SDKv2).Delete(context.Background(), &iam.DeleteOAuthClientRequest{
		OauthClientId: id,
	})

	if err != nil {
		return err
	}

	_, err = op.Wait(context.Background())
	return err
}

func TestAccIAMOauthClient_full(t *testing.T) {
	var (
		clientName        = test.ResourceName(63)
		clientNameUpdated = test.ResourceName(63)

		folderID = test.GetExampleFolderID()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckProjectDestroy,
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
			resource.TestCheckTypeSetElemAttr(oauthClientResourceName, "scopes.*", "iam"),
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
  scopes        = ["iam"]
}`, clientName, folderID)
}
