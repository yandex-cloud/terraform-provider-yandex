package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func init() {
	resource.AddTestSweepers("yandex_compute_filesystem", &resource.Sweeper{
		Name: "yandex_compute_filesystem",
		F:    testSweepComputeFilesystem,
		Dependencies: []string{
			"yandex_compute_instance",
		},
	})
}

func testSweepComputeFilesystem(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListFilesystemsRequest{FolderId: conf.FolderID}
	it := conf.sdk.Compute().Filesystem().FilesystemIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepComputeFilesystem(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Compute Filesystem %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputeFilesystem(conf *Config, id string) bool {
	return sweepWithRetry(sweepComputeFilesystemOnce, conf, "Compute Filesystem", id)
}

func sweepComputeFilesystemOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexComputeFilesystemDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Compute().Filesystem().Delete(ctx, &compute.DeleteFilesystemRequest{
		FilesystemId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccComputeFilesystem_basic(t *testing.T) {
	t.Parallel()

	fsName := acctest.RandomWithPrefix("tf-test")
	var fs compute.Filesystem

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckComputeFilesystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeFilesystem_basic(fsName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeFilesystemExists("yandex_compute_filesystem.foobar", &fs),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar", "name", fsName),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar", "size", "10"),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar",
						"type", "network-hdd"),
				),
			},
		},
	})
}

func TestAccComputeFilesystem_update(t *testing.T) {
	t.Parallel()

	fsName := acctest.RandomWithPrefix("tf-test")
	var fs compute.Filesystem

	newFsName := acctest.RandomWithPrefix("tf-test")
	newFsDesc := "new description"
	newFsSize := "20"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckComputeFilesystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeFilesystem_basic(fsName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeFilesystemExists("yandex_compute_filesystem.foobar", &fs),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar", "name", fsName),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar", "size", "10"),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar",
						"type", "network-hdd"),
				),
			},
			{
				Config: testAccComputeFilesystem_updated(newFsName, newFsDesc, newFsSize),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeFilesystemExists("yandex_compute_filesystem.foobar", &fs),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar", "name", newFsName),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar", "description", newFsDesc),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar", "size", newFsSize),
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

//revive:disable:var-naming
func testAccComputeFilesystem_basic(name string) string {
	return fmt.Sprintf(`
resource "yandex_compute_filesystem" "foobar" {
  name     = "%s"
  size     = 10
  type     = "network-hdd"

  labels = {
    my-label = "my-label-value"
  }
}
`, name)
}

//revive:disable:var-naming
func testAccComputeFilesystem_updated(name, desc, size string) string {
	return fmt.Sprintf(`
resource "yandex_compute_filesystem" "foobar" {
  name        = "%s"
  description = "%s"
  size        = %s
  type        = "network-hdd"

  labels = {
    my-label = "my-label-value"
  }
}
`, name, desc, size)
}
