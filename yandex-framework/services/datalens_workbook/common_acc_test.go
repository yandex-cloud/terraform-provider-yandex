package datalens_workbook_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	dlWorkbookResource   = "yandex_datalens_workbook.test"
	dlWorkbookDataSource = "data.yandex_datalens_workbook.test"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("yandex_datalens_workbook", &resource.Sweeper{
		Name: "yandex_datalens_workbook",
		F:    testSweepDatalensWorkbook,
	})
}

// testSweepDatalensWorkbook is a no-op sweeper. The DataLens API does expose
// getWorkbooksList, but enumerating and force-deleting workbooks across an
// entire org is too dangerous for a sweeper that runs unattended; users may
// have hand-created important workbooks in the test org.
func testSweepDatalensWorkbook(_ string) error {
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
	return nil
}

func testAccCheckDatalensWorkbookDestroy(s *terraform.State) error {
	cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	orgID := test.GetExampleOrganizationID()

	dl, err := newTestDatalensClient(&cfg)
	if err != nil {
		return err
	}

	var result *multierror.Error
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_datalens_workbook" {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := dl.Do(ctx, "/rpc/getWorkbook", orgID, map[string]string{"workbookId": rs.Primary.ID}, nil)
		cancel()
		if err == nil {
			result = multierror.Append(result, fmt.Errorf("DataLens workbook %s still exists", rs.Primary.ID))
		} else if !datalens.IsNotFound(err) {
			result = multierror.Append(result, fmt.Errorf("error checking workbook %s: %w", rs.Primary.ID, err))
		}
	}
	return result.ErrorOrNil()
}

func testAccCheckDatalensWorkbookExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}
		cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		orgID := test.GetExampleOrganizationID()
		dl, err := newTestDatalensClient(&cfg)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return dl.Do(ctx, "/rpc/getWorkbook", orgID, map[string]string{"workbookId": rs.Primary.ID}, nil)
	}
}

func datalensWorkbookImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      dlWorkbookResource,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateIdFunc: datalensWorkbookImportIDFunc(dlWorkbookResource),
	}
}

func datalensWorkbookImportIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		orgID := rs.Primary.Attributes["organization_id"]
		if orgID == "" {
			return "", fmt.Errorf("organization_id not set")
		}
		return resourceid.Construct(orgID, rs.Primary.ID), nil
	}
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
