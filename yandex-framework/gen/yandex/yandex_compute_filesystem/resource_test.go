package yandex_compute_filesystem_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccComputeFilesystem_UpgradeFromSDKv2(t *testing.T) {
	t.Parallel()

	fsName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testAccCheckComputeFilesystemDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccComputeFilesystem_basic(fsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar", "name", fsName),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar", "size", "10"),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("yandex_compute_filesystem.foobar",
						"type", "network-hdd"),
				),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccComputeFilesystem_basic(fsName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccComputeFilesystem_basic(t *testing.T) {
	t.Parallel()

	fsName := acctest.RandomWithPrefix("tf-test")
	var fs compute.Filesystem

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeFilesystemDestroy,
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
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeFilesystemDestroy,
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
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_filesystem" {
			continue
		}

		_, err := computesdk.NewFilesystemClient(config.SDKv2).Get(context.Background(), &compute.GetFilesystemRequest{
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

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := computesdk.NewFilesystemClient(config.SDKv2).Get(context.Background(), &compute.GetFilesystemRequest{
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
