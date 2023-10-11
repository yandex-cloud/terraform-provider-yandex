package test

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider-config"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testProjectResourceName = "yandex_datasphere_project.test-project"

func init() {
	resource.AddTestSweepers("yandex_datasphere_project", &resource.Sweeper{
		Name:         "yandex_datasphere_project",
		F:            testSweepProject,
		Dependencies: []string{},
	})
}
func testSweepProject(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	it := conf.SDK.Datasphere().Project().ProjectIterator(
		context.Background(),
		&datasphere.ListProjectsRequest{},
	)
	result := &multierror.Error{}

	for it.Next() {
		projectId := it.Value().GetId()
		if !sweepProject(conf, projectId) {
			result = multierror.Append(
				result,
				fmt.Errorf("failed to sweep project id %q", projectId),
			)
		}
	}

	return result.ErrorOrNil()
}

func sweepProject(conf *provider_config.Config, cloudId string) bool {
	return sweepWithRetry(sweepProjectOnce, conf, "yandex_datasphere_project", cloudId)
}

func sweepProjectOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

	defer cancel()

	op, err := conf.SDK.Datasphere().Project().Delete(ctx, &datasphere.DeleteProjectRequest{
		ProjectId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccDatasphereProjectResource_basic(t *testing.T) {
	communityName := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)

	projectName := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
	projectDesc := acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
	labelKey := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
	labelValue := acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			basicDatasphereProjectTestStep(communityName, projectName, projectDesc, labelKey, labelValue),
			datasphereProjectImportTestStep(),
		},
	})
}

func TestAccDatasphereProjectResource_fullData(t *testing.T) {
	communityName := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)

	projectName := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
	projectDesc := acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
	saName := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasphereProjectFullConfig(communityName, projectName, projectDesc, saName),
				Check: resource.ComposeTestCheckFunc(
					testDatasphereProjectExists(testProjectResourceName),
					resource.TestCheckResourceAttr(testProjectResourceName, "name", projectName),
					resource.TestCheckResourceAttr(testProjectResourceName, "description", projectDesc),
					resource.TestCheckResourceAttrSet(testProjectResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testProjectResourceName, "created_by"),
					resource.TestCheckResourceAttr(testProjectResourceName, "limits.max_units_per_hour", "10"),
					resource.TestCheckResourceAttr(testProjectResourceName, "limits.max_units_per_execution", "10"),
					resource.TestCheckResourceAttr(testProjectResourceName, "limits.balance", "10"),
					resource.TestCheckResourceAttr(testProjectResourceName, "settings.commit_mode", "AUTO"),
					resource.TestCheckResourceAttr(testProjectResourceName, "settings.ide", "JUPYTER_LAB"),
					resource.TestCheckResourceAttr(testProjectResourceName, "settings.stale_exec_timeout_mode", "ONE_HOUR"),
					testAccCheckCreatedAtAttr(testProjectResourceName),
				),
			},
			datasphereProjectImportTestStep(),
		},
	})
}

func TestAccDatasphereProjectResource_update(t *testing.T) {
	communityName := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)

	projectName := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
	projectDesc := acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
	labelKey := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
	labelValue := acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)

	projectNameUpdated := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
	projectDescUpdated := acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
	labelKeyUpdated := acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
	labelValueUpdated := acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			basicDatasphereProjectTestStep(communityName, projectName, projectDesc, labelKey, labelValue),
			datasphereProjectImportTestStep(),
			basicDatasphereProjectTestStep(communityName, projectNameUpdated, projectDescUpdated, labelKeyUpdated, labelValueUpdated),
			datasphereProjectImportTestStep(),
		},
	})
}

func datasphereProjectImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      testProjectResourceName,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccCheckProjectDestroy(s *terraform.State) error {
	config := testAccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_project_project" {
			continue
		}
		id := rs.Primary.ID

		_, err := config.SDK.Datasphere().Project().Get(context.Background(), &datasphere.GetProjectRequest{
			ProjectId: id,
		})
		if err == nil {
			return fmt.Errorf("project still exists")
		}
	}

	return nil
}

func basicDatasphereProjectTestStep(communityName, projectName, projectDesc, labelKey, labelValue string) resource.TestStep {
	return resource.TestStep{
		Config: testDatasphereProjectBasicConfig(communityName, projectName, projectDesc, labelKey, labelValue),
		Check: resource.ComposeTestCheckFunc(
			testDatasphereProjectExists(testProjectResourceName),
			resource.TestCheckResourceAttr(testProjectResourceName, "name", projectName),
			resource.TestCheckResourceAttr(testProjectResourceName, "description", projectDesc),
			resource.TestCheckResourceAttr(testProjectResourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
			resource.TestCheckResourceAttrSet(testProjectResourceName, "created_at"),
			resource.TestCheckResourceAttrSet(testProjectResourceName, "created_by"),
			testAccCheckCreatedAtAttr(testProjectResourceName),
		),
	}
}

func testDatasphereProjectExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.(*yandex_framework.Provider).GetConfig()

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		id := rs.Primary.ID

		found, err := config.SDK.Datasphere().Project().Get(context.Background(), &datasphere.GetProjectRequest{
			ProjectId: id,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("project not found")
		}

		return nil
	}
}

func testDatasphereProjectBasicConfig(communityName, projectName, desc, labelKey, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_datasphere_community" "test-community" {
  name = "%s"
  organization_id = "%s"
  billing_account_id = "%s"
}

resource "yandex_datasphere_project" "test-project" {
  name = "%s"
  description = "%s"
  labels = {
    "%s": "%s"
  }
  community_id = yandex_datasphere_community.test-community.id
}
`, communityName, getExampleOrganizationID(), getBillingAccountId(), projectName, desc, labelKey, labelValue)
}

func testDatasphereProjectFullConfig(communityName, projectName, desc, saName string) string {
	return fmt.Sprintf(`
resource "yandex_datasphere_community" "test-community" {
  name = "%s"
  organization_id = "%s"
  billing_account_id = "%s"
}

resource "yandex_vpc_network" "test-network" {}

resource "yandex_vpc_subnet" "test-subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.test-network.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_security_group" "test-security-group" {
  network_id = yandex_vpc_network.test-network.id

  ingress {
    protocol       = "TCP"
    description    = "healthchecks"
    port           = 30080
    v4_cidr_blocks = ["198.18.235.0/24", "198.18.248.0/24"]
  }
}

resource "yandex_iam_service_account" "test-account" {
  name        = "%s"
  description = "tf-test"
}

resource "yandex_datasphere_project" "test-project" {
  name = "%s"
  description = "%s"

  labels = {
    test-label: "test-label-value"
  }

  community_id = yandex_datasphere_community.test-community.id
  
  limits = {
	max_units_per_hour = 10
    max_units_per_execution = 10
	balance = 10
  }

  settings = {
	service_account_id = yandex_iam_service_account.test-account.id
 	subnet_id = yandex_vpc_subnet.test-subnet.id
	commit_mode = "AUTO"
	security_group_ids = [yandex_vpc_security_group.test-security-group.id]
	ide = "JUPYTER_LAB"
	default_folder_id = "%s"
	stale_exec_timeout_mode = "ONE_HOUR"
  }
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "%s"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
}
`, communityName, getExampleOrganizationID(), getBillingAccountId(), saName, projectName, desc, getExampleFolderID(), getExampleFolderID())
}
