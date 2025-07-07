package gitlab_instance_test

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/gitlab/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

func infraResources(t *testing.T, randSuffix string) string {
	type params struct {
		RandSuffix string
		FolderID   string
	}
	p := params{
		RandSuffix: randSuffix,
		FolderID:   os.Getenv("YC_FOLDER_ID"),
	}
	tpl, err := template.New("gitlab").Parse(`
resource "yandex_vpc_network" "gitlab-net" {}

resource "yandex_vpc_subnet" "gitlab-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.gitlab-net.id
  v4_cidr_blocks = ["10.128.0.0/24"]
}
`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, p))
	return b.String()
}

type gitlabInstanceConfigParams struct {
	RandSuffix                string
	Description               string
	IncludeBlockLabels        bool
	Labels                    map[string]string
	ResourcePresetID          string
	DiskSize                  int64
	AdminLogin                string
	AdminEmail                string
	BackupRetainPeriodDays    int64
	MaintenanceDeleteUntagged bool
	ApprovalRulesId           string
	IncludeBlockTimeouts      bool
	DeletionProtection        bool
}

func gitlabInstanceConfig(t *testing.T, params gitlabInstanceConfigParams) string {
	tpl, err := template.New("gitlab").Parse(`
resource "yandex_gitlab_instance" "gitlab_instance" {
  name = "gitlab-{{ .RandSuffix }}"

  {{ if .Description }}
  description = "{{ .Description }}"
  {{ end }}

  {{ if .IncludeBlockLabels }}
  labels = {
    {{ range $key, $val := .Labels }}
    {{ $key }} = "{{ $val }}"
    {{ end }}
  }
  {{ end }}

  resource_preset_id = "{{ .ResourcePresetID }}"

  disk_size = {{ .DiskSize }}

  admin_login = "{{ .AdminLogin }}"

  admin_email = "{{ .AdminEmail }}"

  domain = "gitlab-{{ .RandSuffix }}.gitlab.yandexcloud.net"

  subnet_id = yandex_vpc_subnet.gitlab-a.id

  backup_retain_period_days = {{ .BackupRetainPeriodDays }}

  {{ if .MaintenanceDeleteUntagged }}
  maintenance_delete_untagged = {{ .MaintenanceDeleteUntagged }}
  {{ end }}

  {{ if .DeletionProtection }}
  deletion_protection = {{ .DeletionProtection }}
  {{ end }}

  approval_rules_id = "{{ .ApprovalRulesId }}"

  {{ if .IncludeBlockTimeouts }}
  timeouts {
    create = "60m"
    update = "60m"
    delete = "60m"
  }
  {{ end }}
}`)
	require.NoError(t, err)
	b := new(bytes.Buffer)
	require.NoError(t, tpl.Execute(b, params))

	return fmt.Sprintf("%s\n%s", infraResources(t, params.RandSuffix), b.String())
}

func testAccCheckGitlabInstanceDestroy(s *terraform.State) error {
	sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_gitlab_instance" {
			continue
		}

		_, err := sdk.Gitlab().Instance().Get(context.Background(), &gitlab.GetInstanceRequest{
			InstanceId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Gitlab instance still exists")
		}
	}

	return nil
}

func testAccCheckGitlabInstanceExists(name string, instance *gitlab.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		sdk := testhelpers.AccProvider.(*provider.Provider).GetConfig().SDK
		found, err := sdk.Gitlab().Instance().Get(context.Background(), &gitlab.GetInstanceRequest{
			InstanceId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Gitlab instance not found")
		}

		if instance != nil {
			*instance = *found
		}

		return nil
	}
}

func TestAccGitlabInstance1_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	var instance gitlab.Instance

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckGitlabInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: gitlabInstanceConfig(t, gitlabInstanceConfigParams{
					RandSuffix:             randSuffix,
					ResourcePresetID:       "s2.micro",
					DiskSize:               30,
					AdminLogin:             "robot-gitlab-ci",
					AdminEmail:             "robot-gitlab-ci@yandex-team.ru",
					BackupRetainPeriodDays: 7,
					ApprovalRulesId:        "NONE",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabInstanceExists("yandex_gitlab_instance.gitlab_instance", &instance),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "name", fmt.Sprintf("gitlab-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "resource_preset_id", "s2.micro"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "disk_size", fmt.Sprintf("%d", 30)),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "admin_login", "robot-gitlab-ci"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "admin_email", "robot-gitlab-ci@yandex-team.ru"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "domain", fmt.Sprintf("gitlab-%s.gitlab.yandexcloud.net", randSuffix)),
					resource.TestCheckResourceAttrSet("yandex_gitlab_instance.gitlab_instance", "subnet_id"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "backup_retain_period_days", "7"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "maintenance_delete_untagged", "false"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "approval_rules_id", "NONE"),
					resource.TestCheckResourceAttrSet("yandex_gitlab_instance.gitlab_instance", "gitlab_version"),
				),
			},
			{
				ResourceName:      "yandex_gitlab_instance.gitlab_instance",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGitlabInstance2_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	var instance gitlab.Instance

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckGitlabInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: gitlabInstanceConfig(t, gitlabInstanceConfigParams{
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
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabInstanceExists("yandex_gitlab_instance.gitlab_instance", &instance),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "name", fmt.Sprintf("gitlab-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "description", "description"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "labels.key", "value"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "resource_preset_id", "s2.micro"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "disk_size", fmt.Sprintf("%d", 30)),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "admin_login", "robot-gitlab-ci"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "admin_email", "robot-gitlab-ci@yandex-team.ru"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "domain", fmt.Sprintf("gitlab-%s.gitlab.yandexcloud.net", randSuffix)),
					resource.TestCheckResourceAttrSet("yandex_gitlab_instance.gitlab_instance", "subnet_id"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "backup_retain_period_days", "14"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "maintenance_delete_untagged", "true"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "approval_rules_id", "BASIC"),
					resource.TestCheckResourceAttrSet("yandex_gitlab_instance.gitlab_instance", "gitlab_version"),
				),
			},
			{
				ResourceName:      "yandex_gitlab_instance.gitlab_instance",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: gitlabInstanceConfig(t, gitlabInstanceConfigParams{
					RandSuffix:       randSuffix,
					ResourcePresetID: "s2.micro",
					DiskSize:         30,
					AdminLogin:       "robot-gitlab-ci-updated",
					AdminEmail:       "robot-gitlab-ci@yandex-team.ru",
					Description:      "description",
					Labels: map[string]string{
						"key": "value",
					},
					IncludeBlockLabels:        true,
					BackupRetainPeriodDays:    14,
					MaintenanceDeleteUntagged: true,
					ApprovalRulesId:           "BASIC",
				}),
				ExpectError: regexp.MustCompile(".*Attribute admin_login can't be changed.*"),
			},
			{
				ResourceName:      "yandex_gitlab_instance.gitlab_instance",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: gitlabInstanceConfig(t, gitlabInstanceConfigParams{
					RandSuffix:       randSuffix,
					ResourcePresetID: "s2.micro",
					DiskSize:         30,
					AdminLogin:       "robot-gitlab-ci",
					AdminEmail:       "robot-gitlab-ci-updated@yandex-team.ru",
					Description:      "description",
					Labels: map[string]string{
						"key": "value",
					},
					IncludeBlockLabels:        true,
					BackupRetainPeriodDays:    14,
					MaintenanceDeleteUntagged: true,
					ApprovalRulesId:           "BASIC",
				}),
				ExpectError: regexp.MustCompile(".*Attribute admin_email can't be changed.*"),
			},
			{
				ResourceName:      "yandex_gitlab_instance.gitlab_instance",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: gitlabInstanceConfig(t, gitlabInstanceConfigParams{
					RandSuffix:       randSuffix + "-updated",
					ResourcePresetID: "s2.micro",
					DiskSize:         30,
					AdminLogin:       "robot-gitlab-ci",
					AdminEmail:       "robot-gitlab-ci-updated@yandex-team.ru",
					Description:      "description",
					Labels: map[string]string{
						"key": "value",
					},
					IncludeBlockLabels:        true,
					BackupRetainPeriodDays:    14,
					MaintenanceDeleteUntagged: true,
					ApprovalRulesId:           "BASIC",
				}),
				ExpectError: regexp.MustCompile(".*Attribute domain can't be changed.*"),
			},
			{
				ResourceName:      "yandex_gitlab_instance.gitlab_instance",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: gitlabInstanceConfig(t, gitlabInstanceConfigParams{
					RandSuffix:       randSuffix,
					ResourcePresetID: "s2.micro",
					DiskSize:         32,
					AdminLogin:       "robot-gitlab-ci",
					AdminEmail:       "robot-gitlab-ci@yandex-team.ru",
					Description:      "description",
					Labels: map[string]string{
						"key": "value",
					},
					IncludeBlockLabels:        true,
					BackupRetainPeriodDays:    7,
					MaintenanceDeleteUntagged: false,
					ApprovalRulesId:           "NONE",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabInstanceExists("yandex_gitlab_instance.gitlab_instance", &instance),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "name", fmt.Sprintf("gitlab-%s", randSuffix)),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "description", "description"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "labels.key", "value"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "resource_preset_id", "s2.micro"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "disk_size", fmt.Sprintf("%d", 32)),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "admin_login", "robot-gitlab-ci"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "admin_email", "robot-gitlab-ci@yandex-team.ru"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "domain", fmt.Sprintf("gitlab-%s.gitlab.yandexcloud.net", randSuffix)),
					resource.TestCheckResourceAttrSet("yandex_gitlab_instance.gitlab_instance", "subnet_id"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "backup_retain_period_days", "7"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "maintenance_delete_untagged", "false"),
					resource.TestCheckResourceAttr("yandex_gitlab_instance.gitlab_instance", "approval_rules_id", "NONE"),
					resource.TestCheckResourceAttrSet("yandex_gitlab_instance.gitlab_instance", "gitlab_version"),
				),
			},
		},
	})
}
