package datalens_connection_test

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
	dlConnectionResource   = "yandex_datalens_connection.test"
	dlConnectionDataSource = "data.yandex_datalens_connection.test"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("yandex_datalens_connection", &resource.Sweeper{
		Name: "yandex_datalens_connection",
		F:    testSweepDatalensConnection,
	})
}

func testSweepDatalensConnection(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	orgID := test.GetExampleOrganizationID()
	if orgID == "" {
		return fmt.Errorf("YC_ORGANIZATION_ID must be set for sweepers")
	}

	dlClient, err := datalens.NewClient(datalens.Config{
		Endpoint: conf.ProviderState.DatalensEndpoint.ValueString(),
		TokenProvider: func(ctx context.Context) (string, error) {
			resp, err := conf.SDK.CreateIAMToken(ctx)
			if err != nil {
				return "", err
			}
			return resp.IamToken, nil
		},
	})
	if err != nil {
		return fmt.Errorf("error creating DataLens client for sweeper: %w", err)
	}

	_ = dlClient // TODO: list and delete test connections when a list API is available
	return nil
}

func testAccCheckDatalensConnectionDestroy(s *terraform.State) error {
	cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	orgID := test.GetExampleOrganizationID()

	dlClient, err := newTestDatalensClient(&cfg)
	if err != nil {
		return fmt.Errorf("error creating DataLens client: %w", err)
	}

	var result *multierror.Error
	for _, rs := range s.RootModule().Resources {
		func() {
			if rs.Type != "yandex_datalens_connection" {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			req := map[string]string{"connectionId": rs.Primary.ID}
			err := dlClient.Do(ctx, "/rpc/getConnection", orgID, req, nil)
			if err == nil {
				result = multierror.Append(result, fmt.Errorf("DataLens connection %s still exists", rs.Primary.ID))
			} else if !datalens.IsNotFound(err) {
				result = multierror.Append(result, fmt.Errorf("error checking connection %s: %w", rs.Primary.ID, err))
			}
		}()
	}

	return result.ErrorOrNil()
}

func testAccCheckDatalensConnectionExists(resourceName string) resource.TestCheckFunc {
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

		dlClient, err := newTestDatalensClient(&cfg)
		if err != nil {
			return fmt.Errorf("error creating DataLens client: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req := map[string]string{"connectionId": rs.Primary.ID}
		err = dlClient.Do(ctx, "/rpc/getConnection", orgID, req, nil)
		if err != nil {
			return fmt.Errorf("DataLens connection %s does not exist: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckDatalensConnectionResourceID(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		orgID := rs.Primary.Attributes["organization_id"]
		expectedImportID := resourceid.Construct(orgID, rs.Primary.ID)

		deconstructedOrg, deconstructedID, err := resourceid.Deconstruct(expectedImportID)
		if err != nil {
			return fmt.Errorf("failed to deconstruct import ID %q: %w", expectedImportID, err)
		}
		if deconstructedOrg != orgID {
			return fmt.Errorf("expected org_id %q, got %q", orgID, deconstructedOrg)
		}
		if deconstructedID != rs.Primary.ID {
			return fmt.Errorf("expected connection_id %q, got %q", rs.Primary.ID, deconstructedID)
		}

		return nil
	}
}

func datalensConnectionImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      dlConnectionResource,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateIdFunc: datalensConnectionImportIDFunc(dlConnectionResource),
		ImportStateVerifyIgnore: []string{
			"ydb.token",    // write-only, never returned by API
			"ydb.ssl_ca",   // write-only, never returned by API
			"ydb.dir_path", // not returned by getConnection
		},
	}
}

func datalensConnectionImportIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		orgID := rs.Primary.Attributes["organization_id"]
		if orgID == "" {
			return "", fmt.Errorf("organization_id is not set on %s", resourceName)
		}

		return resourceid.Construct(orgID, rs.Primary.ID), nil
	}
}

// testAccDatalensConnectionInfraConfig returns HCL that creates all infrastructure
// needed for DataLens connection tests:
//   - a service account with ydb.editor role
//   - a serverless YDB database
//
// Names must be generated once per test to stay deterministic across test steps.
func testAccDatalensConnectionInfraConfig(saName, dbName string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "datalens_test_sa" {
  name        = "%s"
  description = "SA for DataLens connection acceptance tests"
  folder_id   = "%s"
}

resource "yandex_resourcemanager_folder_iam_member" "datalens_test_sa_ydb_editor" {
  folder_id = "%s"
  role      = "ydb.editor"
  member    = "serviceAccount:${yandex_iam_service_account.datalens_test_sa.id}"
}

resource "yandex_ydb_database_serverless" "test" {
  name      = "%s"
  folder_id = "%s"
  location_id = "global"
}

locals {
  ydb_host = split(":", yandex_ydb_database_serverless.test.ydb_api_endpoint)[0]
}
`, saName, test.GetExampleFolderID(), test.GetExampleFolderID(), dbName, test.GetExampleFolderID())
}

func newTestDatalensClient(config *provider_config.Config) (*datalens.Client, error) {
	return datalens.NewClient(datalens.Config{
		Endpoint: config.ProviderState.DatalensEndpoint.ValueString(),
		TokenProvider: func(ctx context.Context) (string, error) {
			resp, err := config.SDK.CreateIAMToken(ctx)
			if err != nil {
				return "", err
			}
			return resp.IamToken, nil
		},
	})
}
