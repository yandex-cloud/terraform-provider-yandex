package sweep

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func registerDatalensConnectionCommonAcc(add func(string, *resource.Sweeper)) {
	add("yandex_datalens_connection", &resource.Sweeper{
		Name: "yandex_datalens_connection",
		F:    sweepAllDatalensConnection,
	})
}

func sweepAllDatalensConnection(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	orgID := test.GetExampleOrganizationID()
	if orgID == "" {
		return nil // skip sweep when org ID is not configured
	}

	dlClient, err := datalens.NewClient(datalens.Config{
		Endpoint: conf.ProviderState.DatalensEndpoint.ValueString(),
		TokenProvider: func(ctx context.Context) (string, error) {
			resp, err := conf.SDKv2.CreateIAMToken(ctx)
			if err != nil {
				return "", err
			}
			return resp.GetIamToken(), nil
		},
	})
	if err != nil {
		return fmt.Errorf("error creating DataLens client for sweeper: %w", err)
	}

	_ = dlClient // TODO: list and delete test connections when a list API is available
	return nil
}
