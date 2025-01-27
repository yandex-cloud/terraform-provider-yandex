package compute_filesystem_iam_binding_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const timeout = 15 * time.Minute

var fsName = test.GenerateNameForResource(10)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccComputeFilesystem_basicIamMember(t *testing.T) {
	var (
		fs     compute.Filesystem
		userID = "allUsers"
		role   = "editor"

		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	)

	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeFilesystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeFilesystemWithIAMMember_basic(fsName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeFilesystemExists("yandex_compute_filesystem.foobar", &fs),
					test.TestAccCheckIamBindingExists(ctx, func() test.BindingsGetter {
						cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
						return cfg.SDK.Compute().Filesystem()
					}, &fs, role, []string{"system:" + userID}),
				),
			},
		},
	})
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

		found, err := config.SDK.Compute().Filesystem().Get(context.Background(), &compute.GetFilesystemRequest{
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
func testAccComputeFilesystemWithIAMMember_basic(name, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_compute_filesystem" "foobar" {
  name     = "%s"
  size     = 10
  type     = "network-hdd"

  labels = {
    my-label = "my-label-value"
  }
}

resource "yandex_compute_filesystem_iam_binding" "test-filesystem-binding" {
  role = "%s"
  members = ["system:%s"]
  filesystem_id = yandex_compute_filesystem.foobar.id
}
`, name, role, userID)
}

func testAccCheckComputeFilesystemDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_filesystem" {
			continue
		}

		_, err := config.SDK.Compute().Filesystem().Get(context.Background(), &compute.GetFilesystemRequest{
			FilesystemId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("filesystem still exists")
		}
	}

	return nil
}
