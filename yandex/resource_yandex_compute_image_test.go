package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

//revive:disable:var-naming
func TestAccComputeImage_basic(t *testing.T) {
	t.Parallel()

	var image compute.Image

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeImage_basic("image-test-" + acctest.RandString(8)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeImageExists(
						"yandex_compute_image.foobar", &image),
					testAccCheckComputeImageDescription(&image, "description-test"),
					testAccCheckComputeImageFamily(&image, "ubuntu-1804-lts"),
					testAccCheckComputeImageContainsLabel(&image, "tf-label", "tf-label-value"),
					testAccCheckComputeImageContainsLabel(&image, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_compute_image.foobar"),
				),
			},
		},
	})
}

func TestAccComputeImage_productID(t *testing.T) {
	t.Parallel()
	t.Skip("broken test")
	var image compute.Image

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeImage_productID("image-test-" + acctest.RandString(8)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeImageExists(
						"yandex_compute_image.foobar", &image),
					testAccCheckComputeImageDescription(&image, "description-test"),
					testAccCheckComputeImageFamily(&image, "kube-master"),
					testAccCheckComputeImageContainsLabel(&image, "tf-label", "tf-label-value"),
					testAccCheckComputeImageContainsLabel(&image, "empty-label", ""),
					testAccCheckComputeImageContainsProductId(&image, "super-product"),
					testAccCheckComputeImageContainsProductId(&image, "very-good"),
					testAccCheckCreatedAtAttr("yandex_compute_image.foobar"),
				),
			},
		},
	})
}

func TestAccComputeImage_update(t *testing.T) {
	t.Parallel()

	var image compute.Image

	name := "image-test-" + acctest.RandString(8)
	// Only labels supports an update
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeImage_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeImageExists(
						"yandex_compute_image.foobar", &image),
					testAccCheckComputeImageContainsLabel(&image, "tf-label", "tf-label-value"),
					testAccCheckComputeImageContainsLabel(&image, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_compute_image.foobar"),
				),
			},
			{
				Config: testAccComputeImage_update(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeImageExists(
						"yandex_compute_image.foobar", &image),
					testAccCheckComputeImageDoesNotContainLabel(&image, "tf-label"),
					testAccCheckComputeImageContainsLabel(&image, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckComputeImageContainsLabel(&image, "new-field", "only-shows-up-when-updated"),
					testAccCheckCreatedAtAttr("yandex_compute_image.foobar"),
				),
			},
			{
				ResourceName:            "yandex_compute_image.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_url", "os_type", "source_family"},
			},
		},
	})
}

func TestAccComputeImage_basedondisk(t *testing.T) {
	t.Parallel()

	var image compute.Image

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeImage_basedondisk(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeImageExists(
						"yandex_compute_image.foobar", &image),
					testAccCheckComputeImageHasSourceDisk("yandex_compute_image.foobar"),
					testAccCheckCreatedAtAttr("yandex_compute_image.foobar"),
					resource.TestCheckResourceAttr("yandex_compute_image.foobar", "hardware_generation.#", "1"),
					resource.TestCheckResourceAttr("yandex_compute_image.foobar", "hardware_generation.0.generation2_features.#", "1"),
				),
			},
			{
				ResourceName:            "yandex_compute_image.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_snapshot", "source_disk", "source_url", "source_image", "os_type"},
			},
		},
	})
}

func testAccCheckComputeImageDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_image" {
			continue
		}

		r, err := config.sdk.Compute().Image().Get(context.Background(), &compute.GetImageRequest{
			ImageId: rs.Primary.ID,
		})

		// Do not trigger error on images from "standard-images" folder
		if err == nil && r.FolderId != StandardImagesFolderID {
			return fmt.Errorf("Image still exists: %q", r)
		}
	}

	return nil
}

func testAccCheckComputeImageExists(n string, image *compute.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.Compute().Image().Get(context.Background(), &compute.GetImageRequest{
			ImageId: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Image not found")
		}

		*image = *found

		return nil
	}
}

func testAccCheckComputeImageDescription(image *compute.Image, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if image.Description != description {
			return fmt.Errorf("Wrong image description: expected '%s' got '%s'", description, image.Description)
		}
		return nil
	}
}

func testAccCheckComputeImageFamily(image *compute.Image, family string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if image.Family != family {
			return fmt.Errorf("Wrong image family: expected '%s' got '%s'", family, image.Family)
		}
		return nil
	}
}

func testAccCheckComputeImageContainsLabel(image *compute.Image, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := image.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccCheckComputeImageContainsProductId(image *compute.Image, expectedProductID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		for _, productID := range image.ProductIds {
			if productID == expectedProductID {
				return nil
			}
		}

		return fmt.Errorf("Expected product_id '%s' was not found", expectedProductID)
	}
}

func testAccCheckComputeImageDoesNotContainLabel(image *compute.Image, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v, ok := image.Labels[key]; ok {
			return fmt.Errorf("Expected no label for key '%s' but found one with value '%s'", key, v)
		}

		return nil
	}
}

func testAccCheckComputeImageHasSourceDisk(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		rs_attr := rs.Primary.Attributes
		image_source_attrs_to_test := []string{"source_snapshot", "source_disk", "source_url", "source_image"}

		for _, attr_to_check := range image_source_attrs_to_test {
			if _, ok := rs_attr[attr_to_check]; ok {
				return nil
			}
		}

		return fmt.Errorf("No one source attribure found for image %s", n)
	}
}

func testAccComputeImage_basic(name string) string {
	return fmt.Sprintf(`
resource "yandex_compute_image" "foobar" {
  name          = "%s"
  description   = "description-test"
  family        = "ubuntu-1804-lts"
  source_family = "ubuntu-1804-lts"
  min_disk_size = 10
  os_type       = "linux"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}
`, name)
}

func testAccComputeImage_productID(name string) string {
	return fmt.Sprintf(`
resource "yandex_compute_image" "foobar" {
  name          = "%s"
  description   = "description-test"
  family        = "kube-master"
  source_url    = "https://storage.yandexcloud.net/image4tests/kube-master-bios.img"
  min_disk_size = 10
  os_type       = "linux"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }

  product_ids = [
    "super-product",
    "very-good",
  ]
}
`, name)
}

func testAccComputeImage_update(name string) string {
	return fmt.Sprintf(`
resource "yandex_compute_image" "foobar" {
  name          = "%s"
  description   = "description-test"
  source_family = "ubuntu-1804-lts"
  min_disk_size = 10
  os_type       = "linux"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}
`, name)
}

func testAccComputeImage_basedondisk() string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "disk-test-%s"
  zone     = "ru-central1-a"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
  size     = 4
  hardware_generation {
    legacy_features {
	  pci_topology = "PCI_TOPOLOGY_V2"
	}
  }
}

resource "yandex_compute_image" "foobar" {
  name          = "image-test-%s"
  source_disk   = "${yandex_compute_disk.foobar.id}"
  min_disk_size = 8
  os_type       = "linux"
  hardware_generation {
    generation2_features {}
  }
}
`, acctest.RandString(8), acctest.RandString(8))
}
