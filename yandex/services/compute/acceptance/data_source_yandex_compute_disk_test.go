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

func TestAccDataSourceComputeDisk_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	family := "ubuntu-1804-lts"
	diskName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			suite.testAccCheckComputeDiskDestroy,
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
				Config: testAccDataSourceCustomDiskConfig(family, diskName, true),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_compute_disk.source", "disk_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"name", diskName),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"instance_ids.#", "0"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"type", "network-hdd"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"zone"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"block_size", "4096"),
					statecheck.CheckCreatedAtAttr("data.yandex_compute_disk.source"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"kms_key_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source", "hardware_generation.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeDisk_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	family := "ubuntu-1804-lts"
	diskName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			suite.testAccCheckComputeDiskDestroy,
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
				Config: testAccDataSourceCustomDiskConfig(family, diskName, false),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_compute_disk.source", "disk_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"name", diskName),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"instance_ids.#", "0"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"type", "network-hdd"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"zone"),
					resource.TestCheckResourceAttr("data.yandex_compute_disk.source",
						"block_size", "4096"),
					statecheck.CheckCreatedAtAttr("data.yandex_compute_disk.source"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_disk.source",
						"kms_key_id"),
				),
			},
		},
	})
}

func testAccDataSourceCustomDiskResourceConfig(family, name string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "%s"
}
resource "yandex_kms_symmetric_key" "disk-encrypt" {}

resource "yandex_compute_disk" "foo" {
  name       = "%s"
  zone       = "ru-central1-a"
  image_id   = "${data.yandex_compute_image.ubuntu.id}"
  size       = 8
  block_size = 4096
  kms_key_id = "${yandex_kms_symmetric_key.disk-encrypt.id}"

  labels = {
    my-label = "my-label-value"
  }
}
`, family, name)
}

func testAccDataSourceCustomDiskConfig(family, name string, useID bool) string {
	if useID {
		return testAccDataSourceCustomDiskResourceConfig(family, name) + computeDiskDataByIDConfig
	}

	return testAccDataSourceCustomDiskResourceConfig(family, name) + computeDiskDataByNameConfig
}

const computeDiskDataByIDConfig = `
data "yandex_compute_disk" "source" {
  disk_id = "${yandex_compute_disk.foo.id}"
}
`

const computeDiskDataByNameConfig = `
data "yandex_compute_disk" "source" {
  name = "${yandex_compute_disk.foo.name}"
}
`
