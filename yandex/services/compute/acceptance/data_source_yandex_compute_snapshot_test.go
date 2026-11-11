package acceptance

import (
	"fmt"
	"testing"

	kmssdk "github.com/yandex-cloud/go-sdk/services/kms/v1"

	kmsfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/kms/testfixtures"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccDataSourceComputeSnapshot_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	diskName := acctest.RandomWithPrefix("tf-disk")
	snapshotName := acctest.RandomWithPrefix("tf-snap")
	label := acctest.RandomWithPrefix("label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			suite.testAccCheckComputeDiskDestroy,
			suite.testAccCheckComputeSnapshotDestroy,
			kmsfixtures.CheckSymmetricKeysDestroy(func() (kmsfixtures.SymmetricKeyGetter, error) {
				config, err := suite.Config()
				if err != nil {
					return nil, err
				}
				return kmssdk.NewSymmetricKeyClient(config.SDK), nil
			}),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSnapshotConfig(diskName, snapshotName, label, true),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_compute_snapshot.source", "snapshot_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_snapshot.source",
						"name", snapshotName),
					resource.TestCheckResourceAttrSet("data.yandex_compute_snapshot.source",
						"id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_snapshot.source",
						"source_disk_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_snapshot.source",
						"labels.test_label", label),
					statecheck.CheckCreatedAtAttr("data.yandex_compute_snapshot.source"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_snapshot.source",
						"kms_key_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_snapshot.source", "hardware_generation.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeSnapshot_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	diskName := acctest.RandomWithPrefix("tf-disk")
	snapshotName := acctest.RandomWithPrefix("tf-snap")
	label := acctest.RandomWithPrefix("label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			suite.testAccCheckComputeDiskDestroy,
			suite.testAccCheckComputeSnapshotDestroy,
			kmsfixtures.CheckSymmetricKeysDestroy(func() (kmsfixtures.SymmetricKeyGetter, error) {
				config, err := suite.Config()
				if err != nil {
					return nil, err
				}
				return kmssdk.NewSymmetricKeyClient(config.SDK), nil
			}),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSnapshotConfig(diskName, snapshotName, label, false),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_compute_snapshot.source", "snapshot_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_snapshot.source",
						"name", snapshotName),
					resource.TestCheckResourceAttrSet("data.yandex_compute_snapshot.source",
						"id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_snapshot.source",
						"source_disk_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_snapshot.source",
						"labels.test_label", label),
					statecheck.CheckCreatedAtAttr("data.yandex_compute_snapshot.source"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_snapshot.source",
						"kms_key_id"),
				),
			},
		},
	})
}

func testAccDataSourceSnapshotResourceConfig(diskName, snapshotName, labelValue string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_kms_symmetric_key" "disk-encrypt" {}

resource "yandex_compute_disk" "foobar" {
  name       = "%s"
  image_id   = "${data.yandex_compute_image.ubuntu.id}"
  size       = 4
  kms_key_id = "${yandex_kms_symmetric_key.disk-encrypt.id}"
}

resource "yandex_compute_snapshot" "foobar" {
  name           = "%s"
  source_disk_id = "${yandex_compute_disk.foobar.id}"

  labels = {
    test_label = "%s"
  }
}
`, diskName, snapshotName, labelValue)
}

func testAccDataSourceSnapshotConfig(diskName, snapshotName, labelValue string, useID bool) string {
	if useID {
		return testAccDataSourceSnapshotResourceConfig(diskName, snapshotName, labelValue) + computeSnapshotDataByIDConfig
	}

	return testAccDataSourceSnapshotResourceConfig(diskName, snapshotName, labelValue) + computeSnapshotDataByNameConfig
}

const computeSnapshotDataByIDConfig = `
data "yandex_compute_snapshot" "source" {
  snapshot_id = "${yandex_compute_snapshot.foobar.id}"
}
`

const computeSnapshotDataByNameConfig = `
data "yandex_compute_snapshot" "source" {
  name = "${yandex_compute_snapshot.foobar.name}"
}
`
