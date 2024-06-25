package diskcheck

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test"
)

func TestAccCheckComputeDiskExists(n string, disk *compute.Disk, timeout time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.Compute().Disk().Get(ctx, &compute.GetDiskRequest{
			DiskId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Disk not found")
		}

		*disk = *found

		return nil
	}
}

func TestAccCheckComputeDiskPlacementGroupExists(n string, placement *compute.DiskPlacementGroup, timeout time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.Compute().DiskPlacementGroup().Get(ctx, &compute.GetDiskPlacementGroupRequest{
			DiskPlacementGroupId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Disk not found")
		}

		*placement = *found

		return nil
	}
}
