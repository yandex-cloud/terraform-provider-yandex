package sweep

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	iam "github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func registerYandexIamOauthClientResource(add func(string, *resource.Sweeper)) {
	add("yandex_iam_oauth_client", &resource.Sweeper{
		Name:         "yandex_iam_oauth_client",
		F:            sweepAllOauthClient,
		Dependencies: []string{},
	})
}

func sweepAllOauthClient(_ string) error {
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
