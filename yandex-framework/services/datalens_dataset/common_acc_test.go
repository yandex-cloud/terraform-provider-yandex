package datalens_dataset_test

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
	dlDatasetResource   = "yandex_datalens_dataset.test"
	dlDatasetDataSource = "data.yandex_datalens_dataset.test"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("yandex_datalens_dataset", &resource.Sweeper{
		Name: "yandex_datalens_dataset",
		F:    testSweepDatalensDataset,
	})
}

func testSweepDatalensDataset(_ string) error {
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

func testAccCheckDatalensDatasetDestroy(s *terraform.State) error {
	cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	orgID := test.GetExampleOrganizationID()

	dl, err := newTestDatalensClient(&cfg)
	if err != nil {
		return fmt.Errorf("error creating DataLens client: %w", err)
	}

	var result *multierror.Error
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_datalens_dataset" {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := dl.Do(ctx, "/rpc/getDataset", orgID, map[string]string{"datasetId": rs.Primary.ID}, nil)
		cancel()
		if err == nil {
			result = multierror.Append(result, fmt.Errorf("DataLens dataset %s still exists", rs.Primary.ID))
		} else if !datalens.IsNotFound(err) {
			result = multierror.Append(result, fmt.Errorf("error checking dataset %s: %w", rs.Primary.ID, err))
		}
	}
	return result.ErrorOrNil()
}

func testAccCheckDatalensDatasetExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID set for %s", resourceName)
		}
		cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		orgID := test.GetExampleOrganizationID()
		dl, err := newTestDatalensClient(&cfg)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return dl.Do(ctx, "/rpc/getDataset", orgID, map[string]string{"datasetId": rs.Primary.ID}, nil)
	}
}

func datalensDatasetImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      dlDatasetResource,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateIdFunc: datalensDatasetImportIDFunc(dlDatasetResource),
		ImportStateVerifyIgnore: []string{
			"created_via",
			"preview",
			// DataLens getDataset response does not echo workbook_id back, so
			// import cannot restore it. The user has to re-attach it on next
			// apply (or rely on `dir_path`).
			"workbook_id",
			// schema_update_enabled has Default=true in schema but DataLens
			// does not return it; import cannot recover the original value.
			"dataset.schema_update_enabled",
		},
	}
}

func datalensDatasetImportIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		orgID := rs.Primary.Attributes["organization_id"]
		if orgID == "" {
			return "", fmt.Errorf("organization_id is not set")
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
