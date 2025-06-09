package gitlab_instance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceGitlabInstance_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckGitlabInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: gitlabInstanceDatasourceClusterConfig(t, randSuffix, true),
				Check:  datasourceTestCheckComposeFunc(randSuffix),
			},
			{
				Config: gitlabInstanceDatasourceClusterConfig(t, randSuffix, false),
				Check:  datasourceTestCheckComposeFunc(randSuffix),
			},
		},
	})
}

func gitlabInstanceDatasourceClusterConfig(t *testing.T, randSuffix string, byID bool) string {
	resource := gitlabInstanceConfig(t, gitlabInstanceConfigParams{
		RandSuffix:       randSuffix,
		ResourcePresetID: "s2.micro",
		DiskSize:         30,
		AdminLogin:       "robot-gitlab-ci",
		AdminEmail:       "robot-gitlab-ci@yandex-team.ru",
		Description:      "description",
		Labels: map[string]string{
			"key": "value",
		},
		IncludeBlockLabels:        true,
		BackupRetainPeriodDays:    14,
		MaintenanceDeleteUntagged: true,
		ApprovalRulesId:           "BASIC",
		DeletionProtection:        true,
	})

	var datasource string

	if byID {
		datasource = `
data "yandex_gitlab_instance" "gitlab_instance" {
  id = yandex_gitlab_instance.gitlab_instance.id
}`
	} else {
		datasource = `
data "yandex_gitlab_instance" "gitlab_instance" {
  name = yandex_gitlab_instance.gitlab_instance.name
}`
	}

	return fmt.Sprintf("%s\n%s", resource, datasource)
}

func datasourceTestCheckComposeFunc(randSuffix string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckGitlabInstanceExists("yandex_gitlab_instance.gitlab_instance", nil),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "name", fmt.Sprintf("gitlab-%s", randSuffix)),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "description", "description"),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "labels.key", "value"),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "resource_preset_id", "s2.micro"),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "disk_size", fmt.Sprintf("%d", 30<<30)),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "admin_login", "robot-gitlab-ci"),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "admin_email", "robot-gitlab-ci@yandex-team.ru"),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "domain", fmt.Sprintf("gitlab-%s.yandexcloud.net", randSuffix)),
		resource.TestCheckResourceAttrSet("yandex_gitlab_instance.gitlab_instance", "subnet_id"),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "backup_retain_period_days", "14"),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "maintenance_delete_untagged", "true"),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "deletion_protection", "true"),
		resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "approval_rules_id", "BASIC"),
		resource.TestCheckResourceAttrSet("yandex_gitlab_instance.gitlab_instance", "gitlab_version"),
	)
}
