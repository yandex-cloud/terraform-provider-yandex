package datalens_chart_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	dlChartResource   = "yandex_datalens_chart.test"
	dlChartDataSource = "data.yandex_datalens_chart.test"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("yandex_datalens_chart", &resource.Sweeper{
		Name: "yandex_datalens_chart",
		F:    testSweepDatalensChart,
	})
}

func testSweepDatalensChart(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	if test.GetExampleOrganizationID() == "" {
		return nil
	}
	if _, err := newTestDatalensClient(conf); err != nil {
		return err
	}
	// DataLens does not expose a list-by-org RPC; nothing to enumerate.
	return nil
}

func testAccCheckDatalensChartDestroy(s *terraform.State) error {
	cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	orgID := test.GetExampleOrganizationID()

	dl, err := newTestDatalensClient(&cfg)
	if err != nil {
		return err
	}

	var result *multierror.Error
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_datalens_chart" {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		// We don't know the chart type in raw state without re-parsing;
		// try wizard first, then ql, then editor.
		err := chartGetAny(ctx, dl, orgID, rs.Primary.ID, rs.Primary.Attributes["type"])
		cancel()
		if err == nil {
			result = multierror.Append(result, fmt.Errorf("DataLens chart %s still exists", rs.Primary.ID))
		} else if !datalens.IsNotFound(err) {
			result = multierror.Append(result, fmt.Errorf("error checking chart %s: %w", rs.Primary.ID, err))
		}
	}
	return result.ErrorOrNil()
}

func chartGetAny(ctx context.Context, dl *datalens.Client, orgID, chartID, chartType string) error {
	suffix := "Wizard"
	switch chartType {
	case "ql":
		suffix = "QL"
	case "editor":
		suffix = "Editor"
	}
	return dl.Do(ctx, "/rpc/get"+suffix+"Chart", orgID, map[string]string{"chartId": chartID}, nil)
}

func newTestDatalensClient(config *provider_config.Config) (*datalens.Client, error) {
	return datalens.NewClient(datalens.Config{
		Endpoint: config.ProviderState.DatalensEndpoint.ValueString(),
		TokenProvider: func(ctx context.Context) (string, error) {
			t, err := config.SDK.CreateIAMToken(ctx)
			if err != nil {
				return "", err
			}
			return t.IamToken, nil
		},
	})
}
