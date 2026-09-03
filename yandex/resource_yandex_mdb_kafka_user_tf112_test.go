//go:build tf1_12

package yandex

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const kafkaUserWoResourceName = "yandex_mdb_kafka_user.password_wo"

func TestAccMDBKafkaUserPasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-kafka-user-pw-wo")
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBKafkaUserConfigPasswordWo(clusterName, "initial-password", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(kafkaUserWoResourceName, "name", "password-wo-user"),
					resource.TestCheckResourceAttr(kafkaUserWoResourceName, "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr(kafkaUserWoResourceName, "password_wo"),
				),
			},
			{
				Config:             testAccMDBKafkaUserConfigPasswordWo(clusterName, "initial-password", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:      testAccMDBKafkaUserConfigPasswordConflict(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`only one of .password. or .password_wo. can be specified`),
			},
			{
				Config:      testAccMDBKafkaUserConfigPasswordWoWithoutVersion(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`all of .password_wo,password_wo_version. must\s+be\s+specified`),
			},
			{
				Config:      testAccMDBKafkaUserConfigPasswordWoVersionWithoutPassword(clusterName),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`all of .password_wo,password_wo_version. must\s+be\s+specified`),
			},
			{
				Config: testAccMDBKafkaUserConfigPasswordWo(clusterName, "rotated-password", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(kafkaUserWoResourceName, "name", "password-wo-user"),
					resource.TestCheckResourceAttr(kafkaUserWoResourceName, "password_wo_version", "2"),
					resource.TestCheckNoResourceAttr(kafkaUserWoResourceName, "password_wo"),
				),
			},
		},
	})
}

func testAccMDBKafkaUserConfigPasswordWo(name, passwordWo string, passwordWoVersion int) string {
	return testAccMDBKafkaUserConfigStep0(name) + fmt.Sprintf(`
resource "yandex_mdb_kafka_user" "password_wo" {
  cluster_id          = yandex_mdb_kafka_cluster.foo.id
  name                = "password-wo-user"
  password_wo         = %q
  password_wo_version = %d
}
`, passwordWo, passwordWoVersion)
}

func testAccMDBKafkaUserConfigPasswordConflict(name string) string {
	return testAccMDBKafkaUserConfigStep0(name) + `
resource "yandex_mdb_kafka_user" "password_wo" {
  cluster_id          = yandex_mdb_kafka_cluster.foo.id
  name                = "password-wo-user"
  password            = "legacy-password"
  password_wo         = "write-only-password"
  password_wo_version = 1
}
`
}

func testAccMDBKafkaUserConfigPasswordWoWithoutVersion(name string) string {
	return testAccMDBKafkaUserConfigStep0(name) + `
resource "yandex_mdb_kafka_user" "password_wo" {
  cluster_id  = yandex_mdb_kafka_cluster.foo.id
  name        = "password-wo-user"
  password_wo = "write-only-password"
}
`
}

func testAccMDBKafkaUserConfigPasswordWoVersionWithoutPassword(name string) string {
	return testAccMDBKafkaUserConfigStep0(name) + `
resource "yandex_mdb_kafka_user" "password_wo" {
  cluster_id          = yandex_mdb_kafka_cluster.foo.id
  name                = "password-wo-user"
  password_wo_version = 1
}
`
}
