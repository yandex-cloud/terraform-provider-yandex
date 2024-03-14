package yandex

import (
	"context"
	"fmt"
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func init() {
	resource.AddTestSweepers("yandex_iam_service_account", &resource.Sweeper{
		Name: "yandex_iam_service_account",
		F:    testSweepIAMServiceAccounts,
		Dependencies: []string{
			"yandex_dataproc_cluster",
			"yandex_kubernetes_cluster",
			"yandex_compute_instance_group",
			"yandex_audit_trails_trail",
		},
	})
}

func testSweepIAMServiceAccounts(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &iam.ListServiceAccountsRequest{FolderId: conf.FolderID}
	it := conf.sdk.IAM().ServiceAccount().ServiceAccountIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepIAMServiceAccount(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep IAM service account %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepIAMServiceAccount(conf *Config, id string) bool {
	return sweepWithRetry(sweepIAMServiceAccountOnce, conf, "IAM Service Account", id)
}

func sweepIAMServiceAccountOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexIAMServiceAccountDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.IAM().ServiceAccount().Delete(ctx, &iam.DeleteServiceAccountRequest{
		ServiceAccountId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

// NOTE(dxan): function may return non-empty string and non-nil error. Example:
// Resource is successfully created, but wait fails: the function returns id and wait error
func createIAMServiceAccountForSweeper(conf *Config) (string, error) {
	ctx, cancel := conf.ContextWithTimeout(yandexIAMServiceAccountDefaultTimeout)
	defer cancel()
	op, err := conf.sdk.WrapOperation(conf.sdk.IAM().ServiceAccount().Create(ctx, &iam.CreateServiceAccountRequest{
		FolderId:    conf.FolderID,
		Name:        acctest.RandomWithPrefix("sweeper"),
		Description: "created by sweeper",
	}))
	if err != nil {
		return "", fmt.Errorf("failed to create service account: %v", err)
	}

	protoMD, err := op.Metadata()
	if err != nil {
		return "", fmt.Errorf("failed to get metadata from create service account operation: %v", err)
	}

	md, ok := protoMD.(*iam.CreateServiceAccountMetadata)
	if !ok {
		return "", fmt.Errorf("failed to get Service Account ID from create operation metadata")
	}
	id := md.ServiceAccountId

	err = op.Wait(ctx)
	if err != nil {
		return id, fmt.Errorf("error while waiting for create service account operation: %v", err)
	}

	err = assignEditorRoleToSweeperServiceAccount(conf, id)
	if err != nil {
		return id, err
	}

	return md.ServiceAccountId, nil
}

func assignEditorRoleToSweeperServiceAccount(conf *Config, saID string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexResourceManagerFolderDefaultTimeout)
	defer cancel()
	const role_EDITOR = "editor"
	op, err := conf.sdk.WrapOperation(conf.sdk.ResourceManager().Folder().UpdateAccessBindings(ctx, &access.UpdateAccessBindingsRequest{
		ResourceId: conf.FolderID,
		AccessBindingDeltas: []*access.AccessBindingDelta{
			{
				Action: access.AccessBindingAction_ADD,
				AccessBinding: &access.AccessBinding{
					RoleId: role_EDITOR,
					Subject: &access.Subject{
						Id:   saID,
						Type: "serviceAccount",
					},
				},
			},
		},
	}))
	if err != nil {
		return fmt.Errorf("failed to assign '%s' role to the service account %q: %v", role_EDITOR, saID, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for grant access bindings operation '%q': %v", op.Id(), err)
	}

	debugLog("Service account '%s' was granted role '%s' to folder ID '%s'", saID, role_EDITOR, conf.FolderID)

	return nil
}

// Test that a service account resource can be created, updated, and destroyed
func TestAccServiceAccount_basic(t *testing.T) {
	t.Parallel()

	accountName := "a" + acctest.RandString(10)
	accountDesc := "Terraform Test"
	accountDesc2 := "Terraform Test Update"
	folderID := getExampleFolderID()
	uniqueID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// The first step creates a basic service account
			{
				Config: testAccServiceAccountBasic(accountName, accountDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexIAMServiceAccountExists("yandex_iam_service_account.acceptance"),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "folder_id", folderID),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "name", accountName),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "description", accountDesc),
					testAccCheckCreatedAtAttr("yandex_iam_service_account.acceptance"),
				),
			},
			// The second step updates the service account
			{
				Config: testAccServiceAccountBasic(accountName, accountDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexIAMServiceAccountNameModified("yandex_iam_service_account.acceptance", accountDesc2),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "folder_id", folderID),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "name", accountName),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "description", accountDesc2),
					testAccStoreServiceAccountUniqueID(&uniqueID),
				),
			},
			// The third step explicitly adds the same default folderID to the service account configuration
			// and ensure the service account is not recreated by comparing the value of its ID with the one from the previous step
			{
				Config: testAccServiceAccountWithFolderID(folderID, accountName, accountDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexIAMServiceAccountNameModified("yandex_iam_service_account.acceptance", accountDesc2),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "folder_id", folderID),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "name", accountName),
					resource.TestCheckResourceAttr(
						"yandex_iam_service_account.acceptance", "description", accountDesc2),
					resource.TestCheckResourceAttrPtr(
						"yandex_iam_service_account.acceptance", "id", &uniqueID),
				),
			},
			{
				ResourceName:      "yandex_iam_service_account.acceptance",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccStoreServiceAccountUniqueID(uniqueID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		*uniqueID = s.RootModule().Resources["yandex_iam_service_account.acceptance"].Primary.ID
		return nil
	}
}

func testAccCheckYandexIAMServiceAccountExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		return nil
	}
}

func testAccCheckYandexIAMServiceAccountExistsWithID(n string, sa *iam.ServiceAccount) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.IAM().ServiceAccount().Get(context.Background(), &iam.GetServiceAccountRequest{
			ServiceAccountId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Service account not found")
		}

		*sa = *found

		return nil
	}
}

func testAccCheckYandexIAMServiceAccountNameModified(r, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.Attributes["description"] != n {
			return fmt.Errorf("description is %q expected %q", rs.Primary.Attributes["description"], n)
		}

		return nil
	}
}

func testAccCheckIAMServiceAccountDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_service_account" {
			continue
		}

		_, err := config.sdk.IAM().ServiceAccount().Get(context.Background(), &iam.GetServiceAccountRequest{
			ServiceAccountId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Service account still exists")
		}
	}

	return nil
}

func testAccServiceAccountBasic(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  name        = "%v"
  description = "%v"
}
`, name, desc)
}

func testAccServiceAccountWithFolderID(folderID, name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "acceptance" {
  folder_id   = "%v"
  name        = "%v"
  description = "%v"
}
`, folderID, name, desc)
}
