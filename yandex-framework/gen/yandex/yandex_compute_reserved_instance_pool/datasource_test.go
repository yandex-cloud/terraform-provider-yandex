package yandex_compute_reserved_instance_pool_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

func TestAccComputeReservedInstancePool_byID(t *testing.T) {
	instanceReservedPoolName := acctest.RandomWithPrefix("tf-instance-reserved-pool")
	cfg := testAccDataSourceReservedInstancePool(instanceReservedPoolName)

	dataPath := "data.yandex_compute_reserved_instance_pool.pool_data"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckReservedInstancePoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataPath, "id"),
					testAccDataReservedInstancePoolEqApi(
						"yandex_compute_reserved_instance_pool.pool",
						"data.yandex_compute_reserved_instance_pool.pool_data",
					),
					test.AccCheckCreatedAtAttr(dataPath),
				),
			},
		},
	})
}

func testAccDataReservedInstancePoolEqApi(resourcePath string, dataSourcePath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("Not found: %s", resourcePath)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := computesdk.NewReservedInstancePoolClient(config.SDKv2).Get(context.Background(), &compute.GetReservedInstancePoolRequest{
			ReservedInstancePoolId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("ReservedInstancePool %s not found", rs.Primary.ID)
		}

		ds, ok := s.RootModule().Resources[dataSourcePath]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourcePath)
		}

		if ds.Primary.ID != rs.Primary.ID {
			return fmt.Errorf("ReservedInstancePool %s not found", ds.Primary.ID)
		}

		if err := resource.TestCheckResourceAttr(dataSourcePath, "id", found.Id)(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "zone", found.ZoneId)(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "cloud_id", found.CloudId)(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "folder_id", found.FolderId)(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "name", found.Name)(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "description", found.Description)(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "platform_id", found.PlatformId)(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "size", strconv.FormatInt(found.Size, 10))(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "resources_spec.memory", strconv.FormatInt(found.ResourcesSpec.Memory, 10))(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "resources_spec.cores", strconv.FormatInt(found.ResourcesSpec.Cores, 10))(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "resources_spec.gpus", strconv.FormatInt(found.ResourcesSpec.Gpus, 10))(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "resources_spec.core_fraction", strconv.FormatInt(found.ResourcesSpec.CoreFraction, 10))(s); err != nil {
			return err
		}
		if err := resource.TestCheckResourceAttr(dataSourcePath, "network_settings.type", found.NetworkSettings.Type.String())(s); err != nil {
			return err
		}

		return nil
	}
}

func testAccDataSourceReservedInstancePool(poolName string) string {
	cfg, _ := testAccReservedInstancePool_basic(poolName)
	data := `
data "yandex_compute_reserved_instance_pool" "pool_data" {
  reserved_instance_pool_id = "${yandex_compute_reserved_instance_pool.pool.id}"
}
`

	return cfg + data
}
