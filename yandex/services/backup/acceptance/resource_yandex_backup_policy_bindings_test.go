package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	backuppb "github.com/yandex-cloud/go-genproto/yandex/cloud/backup/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_backup "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/backup/pluginsdk"
)

func TestAccResourceBackupPolicyBindingsBasic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	config, resourceName := testAccBackupPolicyBindingsBasicConfig()

	var application backuppb.PolicyApplication
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckBackupPolicyBindingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					suite.testBackupPolicyBindingsExists(resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
					resource.TestCheckResourceAttrSet(resourceName, "enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "processing"),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					id := svc_backup.MakeBackupPolicyBindingsID(application.PolicyId, application.ComputeInstanceId)
					return id, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"created_at",
					"processing",
					"enabled",
				},
			},
		},
	})
}

func testAccBackupPolicyBindingsBasicConfig() (config, outResourceName string) {
	const (
		resourceType = "yandex_backup_policy_bindings"
		resourceName = "test_backup_binding"
		template     = `
resource "yandex_iam_service_account" "test_sa" {
  name = "sa-backup-editor"
}

resource "yandex_resourcemanager_folder_iam_member" "test_binding" {
  folder_id = yandex_iam_service_account.test_sa.folder_id
  role      = "backup.editor"
  member    = "serviceAccount:${yandex_iam_service_account.test_sa.id}"
}

resource "yandex_vpc_network" "test_backup_network" {}

resource "yandex_vpc_subnet" "test_backup_subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.test_backup_network.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

# https://yandex.cloud/docs/backup/concepts/vm-connection#vm-network-access
resource "yandex_vpc_security_group" "test_backup_security_group" {
  name       = "cloud-backup"
  network_id = yandex_vpc_network.test_backup_network.id
  egress {
    protocol       = "TCP"
    from_port      = 7770
    to_port        = 7800
    v4_cidr_blocks = ["84.47.172.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 443
    v4_cidr_blocks = ["213.180.204.0/24", "213.180.193.0/24", "178.176.128.0/24", "84.201.181.0/24", "84.47.172.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 80
    v4_cidr_blocks = ["213.180.204.0/24", "213.180.193.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 8443
    v4_cidr_blocks = ["84.47.172.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 44445
    v4_cidr_blocks = ["51.250.1.0/24"]
  } 
}

data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-2004-lts"
}

resource "yandex_compute_instance" "test_backup_compute" {
  name        = "test-backup-compute"
  platform_id = "standard-v1"
  zone        = "ru-central1-a"

  service_account_id = yandex_iam_service_account.test_sa.id

  network_interface {
    subnet_id          = yandex_vpc_subnet.test_backup_subnet.id
    security_group_ids = [yandex_vpc_security_group.test_backup_security_group.id]
    nat = true
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu.id
    }
  }

  resources {
    cores  = 2
    memory = 4
  }

  metadata = {
    user-data = "#cloud-config\npackages:\n  - curl\n  - perl\n  - jq\nruncmd:\n  - curl https://storage.yandexcloud.net/backup-distributions/agent_installer.sh | sudo bash\n"
  }
}

resource "yandex_backup_policy" "test_backup_policy" {
  name            = "test_backup_policy_name"
  splitting_bytes = "9223372036854775807"
  scheduling {
    enabled             = true
	backup_sets {
      execute_by_interval = 86400
	}
  }

  retention {
    after_backup = true
  }

  reattempts {
    enabled      = true
    interval     = "5m"
    max_attempts = 5
  }

  vm_snapshot_reattempts {
    enabled      = true
    interval     = "5m"
    max_attempts = 5
  }
}

resource %q %q {
  instance_id = yandex_compute_instance.test_backup_compute.id
  policy_id = yandex_backup_policy.test_backup_policy.id
}
`
	)
	outResourceName = resourceType + "." + resourceName
	return fmt.Sprintf(template, resourceType, resourceName), outResourceName
}

func (suite testSuite) testAccCheckBackupPolicyBindingsDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_backup_policy_bindings" {
			continue
		}

		policyID, instanceID, err := svc_backup.ParseBackupPolicyBindingsID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = svc_backup.GetBackupPolicyApplication(context.Background(), config, policyID, instanceID)
		if err == nil {
			return fmt.Errorf("backup policy application %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func (suite testSuite) testBackupPolicyBindingsExists(resourceName string, application *backuppb.PolicyApplication) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("backup policy bindings not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		policyID, instanceID, err := svc_backup.ParseBackupPolicyBindingsID(rs.Primary.ID)
		if err != nil {
			return err
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		found, err := svc_backup.GetBackupPolicyApplication(context.Background(), config, policyID, instanceID)
		if err != nil {
			return err
		}

		//goland:noinspection GoVetCopyLock
		*application = *found

		return nil
	}
}
