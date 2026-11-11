package datasphere_project_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	fixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datasphere/testfixtures"
)

func TestAccDatasphereProjectResource_basic(t *testing.T) {
	var (
		communityName = test.ResourceName(63)

		projectName = test.ResourceName(63)
		projectDesc = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey    = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue  = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             fixtures.AccCheckProjectDestroy(projectClient),
		Steps: []resource.TestStep{
			basicDatasphereProjectTestStep(communityName, projectName, projectDesc, labelKey, labelValue),
			datasphereProjectImportTestStep(),
		},
	})
}

func TestAccDatasphereProjectResource_fullData(t *testing.T) {
	var (
		communityName = test.ResourceName(63)

		projectName = test.ResourceName(63)
		projectDesc = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		saName      = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             fixtures.AccCheckProjectDestroy(projectClient),
		Steps: []resource.TestStep{
			{
				Config: testDatasphereProjectFullConfig(communityName, projectName, projectDesc, saName),
				Check: resource.ComposeTestCheckFunc(
					fixtures.ProjectExists(projectClient, fixtures.ProjectResourceName),
					resource.TestCheckResourceAttr(fixtures.ProjectResourceName, "name", projectName),
					resource.TestCheckResourceAttr(fixtures.ProjectResourceName, "description", projectDesc),
					resource.TestCheckResourceAttrSet(fixtures.ProjectResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(fixtures.ProjectResourceName, "created_by"),
					resource.TestCheckResourceAttr(fixtures.ProjectResourceName, "limits.max_units_per_hour", "10"),
					resource.TestCheckResourceAttr(fixtures.ProjectResourceName, "limits.max_units_per_execution", "10"),
					resource.TestCheckResourceAttr(fixtures.ProjectResourceName, "limits.balance", "10"),
					resource.TestCheckResourceAttr(fixtures.ProjectResourceName, "settings.stale_exec_timeout_mode", "ONE_HOUR"),
					test.AccCheckCreatedAtAttr(fixtures.ProjectResourceName),
				),
			},
			datasphereProjectImportTestStep(),
		},
	})
}

func TestAccDatasphereProjectResource_update(t *testing.T) {
	var (
		communityName = test.ResourceName(63)

		projectName = test.ResourceName(63)
		projectDesc = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey    = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue  = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)

		projectNameUpdated = test.ResourceName(63)
		projectDescUpdated = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKeyUpdated    = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValueUpdated  = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             fixtures.AccCheckProjectDestroy(projectClient),
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
		ResourceName:      fixtures.ProjectResourceName,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func basicDatasphereProjectTestStep(communityName, projectName, projectDesc, labelKey, labelValue string) resource.TestStep {
	return resource.TestStep{
		Config: testDatasphereProjectBasicConfig(communityName, projectName, projectDesc, labelKey, labelValue),
		Check: resource.ComposeTestCheckFunc(
			fixtures.ProjectExists(projectClient, fixtures.ProjectResourceName),
			resource.TestCheckResourceAttr(fixtures.ProjectResourceName, "name", projectName),
			resource.TestCheckResourceAttr(fixtures.ProjectResourceName, "description", projectDesc),
			resource.TestCheckResourceAttr(fixtures.ProjectResourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
			resource.TestCheckResourceAttrSet(fixtures.ProjectResourceName, "created_at"),
			resource.TestCheckResourceAttrSet(fixtures.ProjectResourceName, "created_by"),
			test.AccCheckCreatedAtAttr(fixtures.ProjectResourceName),
		),
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
`, communityName, test.GetExampleOrganizationID(), test.GetBillingAccountId(), projectName, desc, labelKey, labelValue)
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
	security_group_ids = [yandex_vpc_security_group.test-security-group.id]
	default_folder_id = "%s"
	stale_exec_timeout_mode = "ONE_HOUR"
  }
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "%s"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
}
`, communityName, test.GetExampleOrganizationID(), test.GetBillingAccountId(), saName, projectName, desc, test.GetExampleFolderID(), test.GetExampleFolderID())
}
