package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceComputeFilesystem_byID(t *testing.T) {
	t.Parallel()

	fsName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckComputeFilesystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomFilesystemConfig(fsName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_compute_filesystem.source", "filesystem_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_filesystem.source",
						"name", fsName),
					resource.TestCheckResourceAttrSet("data.yandex_compute_filesystem.source",
						"description"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_filesystem.source",
						"id"),
					resource.TestCheckResourceAttr("data.yandex_compute_filesystem.source",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("data.yandex_compute_filesystem.source",
						"type", "network-hdd"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_filesystem.source",
						"zone"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_filesystem.source",
						"folder_id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_filesystem.source",
						"status"),
					resource.TestCheckResourceAttr("data.yandex_compute_filesystem.source",
						"size", "8"),
					resource.TestCheckResourceAttr("data.yandex_compute_filesystem.source",
						"block_size", "4096"),
					testAccCheckCreatedAtAttr("data.yandex_compute_filesystem.source"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeFilesystem_byName(t *testing.T) {
	t.Parallel()

	fsName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckComputeFilesystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomFilesystemConfig(fsName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_compute_filesystem.source", "filesystem_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_filesystem.source",
						"name", fsName),
					resource.TestCheckResourceAttrSet("data.yandex_compute_filesystem.source",
						"description"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_filesystem.source",
						"id"),
					resource.TestCheckResourceAttr("data.yandex_compute_filesystem.source",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("data.yandex_compute_filesystem.source",
						"type", "network-hdd"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_filesystem.source",
						"zone"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_filesystem.source",
						"folder_id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_filesystem.source",
						"status"),
					resource.TestCheckResourceAttr("data.yandex_compute_filesystem.source",
						"size", "8"),
					resource.TestCheckResourceAttr("data.yandex_compute_filesystem.source",
						"block_size", "4096"),
					testAccCheckCreatedAtAttr("data.yandex_compute_filesystem.source"),
				),
			},
		},
	})
}

func testAccCheckComputeFilesystemDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_filesystem" {
			continue
		}

		_, err := config.sdk.Compute().Filesystem().Get(context.Background(), &compute.GetFilesystemRequest{
			FilesystemId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("filesystem still exists")
		}
	}

	return nil
}

func testAccCheckComputeFilesystemExists(n string, fs *compute.Filesystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.Compute().Filesystem().Get(context.Background(), &compute.GetFilesystemRequest{
			FilesystemId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Filesystem not found")
		}

		*fs = *found

		return nil
	}
}

func testAccDataSourceCustomFilesystemResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "yandex_compute_filesystem" "foo" {
  name        = "%s"
  description = "fs description"
  zone        = "ru-central1-a"
  size        = 8
  block_size  = 4096

  labels = {
    my-label = "my-label-value"
  }
}
`, name)
}

func testAccDataSourceCustomFilesystemConfig(name string, useID bool) string {
	if useID {
		return testAccDataSourceCustomFilesystemResourceConfig(name) + computeFilesystemDataByIDConfig
	}

	return testAccDataSourceCustomFilesystemResourceConfig(name) + computeFilesystemDataByNameConfig
}

const computeFilesystemDataByIDConfig = `
data "yandex_compute_filesystem" "source" {
  filesystem_id = "${yandex_compute_filesystem.foo.id}"
}
`

const computeFilesystemDataByNameConfig = `
data "yandex_compute_filesystem" "source" {
  name = "${yandex_compute_filesystem.foo.name}"
}
`
