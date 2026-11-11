package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	kmssdk "github.com/yandex-cloud/go-sdk/services/kms/v1"

	computefixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/compute/testfixtures"

	kmsfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/kms/testfixtures"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccDataSourceComputeImage_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	family := "ubuntu-1804-lts"
	name := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			func(s *terraform.State) error {
				providerConfig, configErr := suite.Config()
				if configErr != nil {
					return configErr
				}
				return computefixtures.CheckImageDestroy(providerConfig, s)
			},
			suite.testAccCheckComputeDiskDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomImageConfig(family, name, true),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_compute_image.source", "image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"name", name),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"family", family),
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.source",
						"id"),
					statecheck.CheckCreatedAtAttr("data.yandex_compute_image.source"),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source", "hardware_generation.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeImage_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	family := "ubuntu-1804-lts"
	name := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			func(s *terraform.State) error {
				providerConfig, configErr := suite.Config()
				if configErr != nil {
					return configErr
				}
				return computefixtures.CheckImageDestroy(providerConfig, s)
			},
			suite.testAccCheckComputeDiskDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomImageConfig(family, name, false),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_compute_image.source", "image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"name", name),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"family", family),
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.source",
						"id"),
					statecheck.CheckCreatedAtAttr("data.yandex_compute_image.source"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeImage_byIDAndFolder(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	family := "ubuntu-1804-lts"
	name := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			func(s *terraform.State) error {
				providerConfig, configErr := suite.Config()
				if configErr != nil {
					return configErr
				}
				return computefixtures.CheckImageDestroy(providerConfig, s)
			},
			suite.testAccCheckComputeDiskDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomImageWithFolderConfig(family, name, true),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_compute_image.source", "image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"name", name),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"family", family),
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.source",
						"id"),
					statecheck.CheckCreatedAtAttr("data.yandex_compute_image.source"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeImage_byNameAndFolder(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	family := "ubuntu-1804-lts"
	name := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			func(s *terraform.State) error {
				providerConfig, configErr := suite.Config()
				if configErr != nil {
					return configErr
				}
				return computefixtures.CheckImageDestroy(providerConfig, s)
			},
			suite.testAccCheckComputeDiskDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomImageWithFolderConfig(family, name, false),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_compute_image.source", "image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"name", name),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"family", family),
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.source",
						"id"),
					statecheck.CheckCreatedAtAttr("data.yandex_compute_image.source"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeImage_StandardByFamily(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	family := "ubuntu-1804-lts"
	re := regexp.MustCompile("ubuntu")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceStandardImageByFamily(family),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_compute_image.by_family", "image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_image.by_family",
						"family", family),
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.by_family",
						"id"),
					resource.TestMatchResourceAttr("data.yandex_compute_image.by_family",
						"name", re),
					statecheck.CheckCreatedAtAttr("data.yandex_compute_image.by_family"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeImage_encrypted(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	family := "ubuntu-1804-lts"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			func(s *terraform.State) error {
				providerConfig, configErr := suite.Config()
				if configErr != nil {
					return configErr
				}
				return computefixtures.CheckImageDestroy(providerConfig, s)
			},
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
				Config: testAccDataSourceEncryptedImageConfig(family),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.source", "kms_key_id"),
				),
			},
		},
	})
}

func testAccDataSourceCustomImageResourceConfig(family, name string) string {
	return fmt.Sprintf(`
resource "yandex_compute_image" "image" {
  family        = "%s"
  name          = "%s"
  source_disk   = "${yandex_compute_disk.disk.id}"
  min_disk_size = 10
  os_type       = "LINUX"
}

resource "yandex_compute_disk" "disk" {
  name = "%s-disk"
  zone = "ru-central1-a"
  size = 4
}
`, family, name, name)
}

const computeImageDataByIDConfig = `
data "yandex_compute_image" "source" {
  image_id = "${yandex_compute_image.image.id}"
}
`

const computeImageDataByNameConfig = `
data "yandex_compute_image" "source" {
  name = "${yandex_compute_image.image.name}"
}
`

const computeImageDataByIDAndFolderConfig = `
data "yandex_compute_image" "source" {
  image_id = "${yandex_compute_image.image.id}"
  folder_id = "${yandex_compute_image.image.folder_id}"
}
`

const computeImageDataByNameAndFolderConfig = `
data "yandex_compute_image" "source" {
  name = "${yandex_compute_image.image.name}"
  folder_id = "${yandex_compute_image.image.folder_id}"
}
`

func testAccDataSourceCustomImageConfig(family, name string, useID bool) string {
	if useID {
		return testAccDataSourceCustomImageResourceConfig(family, name) + computeImageDataByIDConfig
	}

	return testAccDataSourceCustomImageResourceConfig(family, name) + computeImageDataByNameConfig
}

func testAccDataSourceCustomImageWithFolderConfig(family, name string, useID bool) string {
	if useID {
		return testAccDataSourceCustomImageResourceConfig(family, name) + computeImageDataByIDAndFolderConfig
	}

	return testAccDataSourceCustomImageResourceConfig(family, name) + computeImageDataByNameAndFolderConfig
}

func testAccDataSourceStandardImageByFamily(family string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "by_family" {
  family = "%s"
}
`, family)
}

func testAccDataSourceEncryptedImageConfig(family string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "%s"
}
resource "yandex_kms_symmetric_key" "disk-encrypt" {}

resource "yandex_compute_disk" "disk" {
  zone       = "ru-central1-a"
  image_id   = "${data.yandex_compute_image.ubuntu.id}"
  size       = 8
  block_size = 4096
  kms_key_id = "${yandex_kms_symmetric_key.disk-encrypt.id}"
}

resource "yandex_compute_image" "image" {
  source_disk   = "${yandex_compute_disk.disk.id}"
}

data "yandex_compute_image" "source" {
  image_id = "${yandex_compute_image.image.id}"
}
`, family)
}
