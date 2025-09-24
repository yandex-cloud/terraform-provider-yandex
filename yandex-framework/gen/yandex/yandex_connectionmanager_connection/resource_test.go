package yandex_connectionmanager_connection_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	connectionmanager "github.com/yandex-cloud/go-genproto/yandex/cloud/connectionmanager/v1"
	connectionmanagerv1sdk "github.com/yandex-cloud/go-sdk/services/connectionmanager/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"

	"google.golang.org/grpc/metadata"
)

var testConnectionResourceName = "yandex_connectionmanager_connection.test-connection"
var PostgreSQLClusterId = "c9q9bkjusdgluum7q2su"

func init() {
	resource.AddTestSweepers("yandex_connectionmanager_connection", &resource.Sweeper{
		Name:         "yandex_connectionmanager_connection",
		F:            testSweepConnection,
		Dependencies: []string{},
	})
	os.Setenv("TF_VAR_postgres_password", "password")
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepConnection(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	resp, err := connectionmanagerv1sdk.NewConnectionClient(conf.SDKv2).List(context.Background(), &connectionmanager.ListConnectionRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error getting Connection Manager connection: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Connection {
		if !sweepConnection(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Connection Manager connection %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepConnection(conf *provider_config.Config, id string) bool {
	return testhelpers.SweepWithRetry(sweepConnectionOnce, conf, "yandex_connectionmanager_connection", id)
}

func sweepConnectionOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	op, err := connectionmanagerv1sdk.NewConnectionClient(conf.SDKv2).Delete(ctx, &connectionmanager.DeleteConnectionRequest{
		ConnectionId: id,
	})
	_, err = op.Wait(context.Background())
	return err
}

func TestAccConnectionManagerConnectionResource_basic(t *testing.T) {
	var (
		folderId              = testhelpers.GetExampleFolderID()
		connectionName        = testhelpers.ResourceName(63)
		connectionDescription = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey              = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue            = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
		params                = testConnectionPostgresParams(
			"user",
			testPasswordParams("password"),
		)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             AccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			basicConnectionTestStep(folderId, connectionName, connectionDescription, labelKey, labelValue, params),
			connectionImportTestStep(),
		},
	})
}

func TestAccConnectionManagerConnectionResource_update(t *testing.T) {
	var (
		folderId              = testhelpers.GetExampleFolderID()
		connectionName        = testhelpers.ResourceName(63)
		connectionDescription = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey              = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue            = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
		params                = testConnectionPostgresParams(
			"user",
			testPasswordParams("password"),
		)
		updatedParams = testConnectionPostgresParams(
			"user",
			testPasswordParams("new-password"),
		)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             AccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			basicConnectionTestStep(folderId, connectionName, connectionDescription, labelKey, labelValue, params),
			connectionImportTestStep(),
			updateConnectionTestStep(folderId, connectionName, connectionDescription, labelKey, labelKey, updatedParams),
			connectionImportTestStep(),
		},
	})
}

func AccCheckConnectionDestroy(s *terraform.State) error {
	config := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_connectionmanager_connection" {
			continue
		}
		id := rs.Primary.ID

		reqApi := &connectionmanager.GetConnectionRequest{
			ConnectionId: id}
		md := new(metadata.MD)

		_, err := connectionmanagerv1sdk.NewConnectionClient(config.SDKv2).Get(context.Background(), reqApi, grpc.Header(md))

		if err == nil {
			return fmt.Errorf("connection still exists")
		}
	}

	return nil
}

func ConnectionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()
		a := s.RootModule().Resources
		fmt.Printf("%s", a)
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		id := rs.Primary.ID

		reqApi := &connectionmanager.GetConnectionRequest{
			ConnectionId: id,
		}
		md := new(metadata.MD)

		found, err := connectionmanagerv1sdk.NewConnectionClient(config.SDKv2).Get(context.Background(), reqApi, grpc.Header(md))
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("connection not found")
		}

		return nil
	}
}

func testConnectionPostgresParams(username string, password string) string {
	return fmt.Sprintf(` {
	postgresql = {
		managed_cluster_id = "%s"
		auth = {
			user_password = {
				user = "%s"
				password = %s
			}
		}
	}
}`, PostgreSQLClusterId, username, password)
}

func testPasswordParams(raw string) string {
	return ` {
					raw = var.postgres_password
				}
`
}

func testConnectionBasic(folder_id, name, description, labelKey, labelValue, params string) string {
	return fmt.Sprintf(`
variable "postgres_password" {
	type        = string
	sensitive   = true
}
resource "yandex_connectionmanager_connection" "test-connection" {
  folder_id = "%s"
  name = "%s"
  description = "%s"
  labels = {
	"%s" = "%s"
  }
  params = %s
}
`, folder_id, name, description, labelKey, labelValue, params)
}

func basicConnectionTestStep(folderId, name, description, labelKey, labelValue, params string) resource.TestStep {
	return resource.TestStep{
		Config: testConnectionBasic(folderId, name, description, labelKey, labelValue, params),
		Check: resource.ComposeTestCheckFunc(
			ConnectionExists(testConnectionResourceName),
			resource.TestCheckResourceAttr(testConnectionResourceName, "folder_id", folderId),
			resource.TestCheckResourceAttr(testConnectionResourceName, "name", name),
			resource.TestCheckResourceAttr(testConnectionResourceName, "description", description),
			resource.TestCheckResourceAttr(testConnectionResourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "created_at"),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "updated_at"),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "created_by"),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "lockbox_secret.version"),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "lockbox_secret.id"),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "lockbox_secret.newest_version"),

			resource.TestCheckResourceAttrSet(testConnectionResourceName, "params.postgresql.managed_cluster_id"),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "params.postgresql.auth.user_password.user"),
		),
	}
}

func updateConnectionTestStep(folderId, name, description, labelKey, labelValue, params string) resource.TestStep {
	return resource.TestStep{
		Config: testConnectionBasic(folderId, name, description, labelKey, labelValue, params),
		Check: resource.ComposeTestCheckFunc(
			ConnectionExists(testConnectionResourceName),
			resource.TestCheckResourceAttr(testConnectionResourceName, "folder_id", folderId),
			resource.TestCheckResourceAttr(testConnectionResourceName, "name", name),
			resource.TestCheckResourceAttr(testConnectionResourceName, "description", description),
			resource.TestCheckResourceAttr(testConnectionResourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "updated_at"),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "lockbox_secret.version"),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "lockbox_secret.id"),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "lockbox_secret.newest_version"),

			resource.TestCheckResourceAttrSet(testConnectionResourceName, "params.postgresql.managed_cluster_id"),
			resource.TestCheckResourceAttrSet(testConnectionResourceName, "params.postgresql.auth.user_password.user"),
		),
	}
}

func connectionImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      testConnectionResourceName,
		ImportState:       true,
		ImportStateVerify: false,
	}
}
