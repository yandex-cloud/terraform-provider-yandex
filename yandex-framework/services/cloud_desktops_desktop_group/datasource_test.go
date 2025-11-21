package cloud_desktops_desktop_group_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/clouddesktop/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	desktop_group "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/cloud_desktops_desktop_group"
	"google.golang.org/grpc/codes"
)

func TestAccDataSourceCloudDesktopsDesktopGroup_basic(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-desktop-group-datasource")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudDesktopsDesktopGroupCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCloudDesktopsDesktopGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceCloudDesktopsDesktopGroupCheck(
						"data.yandex_cloud_desktops_desktop_group.data_desktop_group_name_and_folder", "yandex_cloud_desktops_desktop_group.desktop_group",
					),
					testAccDataSourceCloudDesktopsDesktopGroupCheck(
						"data.yandex_cloud_desktops_desktop_group.data_desktop_group_only_name", "yandex_cloud_desktops_desktop_group.desktop_group",
					),
					testAccDataSourceCloudDesktopsDesktopGroupCheck(
						"data.yandex_cloud_desktops_desktop_group.data_desktop_group_id", "yandex_cloud_desktops_desktop_group.desktop_group",
					),
				),
			},
		},
	})
}

func testAccDataSourceCloudDesktopsDesktopGroupCheck(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		dsName, dsFolder, _, err := desktop_group.DeconstructID(ds.Primary.ID)
		if err != nil {
			return fmt.Errorf("datasource ID is not in correct form: %w", err)
		}

		rsName, rsFolder, _, err := desktop_group.DeconstructID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("resource ID is not in correct form: %w", err)
		}

		if dsName != rsName {
			return fmt.Errorf("data source name from ID does not match resource name from ID: %s and %s", dsName, rsName)
		}
		if dsFolder != rsFolder {
			return fmt.Errorf("data source folderID from ID does not match resource folderID from ID: %s and %s", dsFolder, rsFolder)
		}

		datasourceAttributes := ds.Primary.Attributes
		resourceAttributes := rs.Primary.Attributes

		toCheck := []string{
			"desktop_group_id", "folder_id", "name", "description", "labels.%", "labels.label1", "labels.label2",
			"desktop_template.resources.memory", "desktop_template.resources.cores", "desktop_template.resources.core_fraction",
			"desktop_template.network_interface.network_id", "desktop_template.network_interface.subnet_ids.#", "desktop_template.network_interface.subnet_ids.0",
			"desktop_template.boot_disk.initialize_params.size", "desktop_template.boot_disk.initialize_params.type",
			"desktop_template.data_disk.initialize_params.size", "desktop_template.data_disk.initialize_params.type",
			"group_config.min_ready_desktops", "group_config.max_desktops_amount", "group_config.desktop_type",
			"group_config.members.#", "group_config.members.0.id", "group_config.members.0.type",
		}
		for _, attrKey := range toCheck {
			var dsVal, rsVal string
			if dsVal, ok = datasourceAttributes[attrKey]; !ok {
				return fmt.Errorf("data source has no attribute %s", attrKey)
			}
			if rsVal, ok = resourceAttributes[attrKey]; !ok {
				return fmt.Errorf("resource has no attribute %s", attrKey)
			}

			if dsVal != rsVal {
				return fmt.Errorf("data source attribute %s doesn't match one in resource: %s != %s", attrKey, dsVal, rsVal)
			}
		}

		return nil
	}
}

func testAccDataSourceCloudDesktopsDesktopGroupConfig(groupName string) string {
	return fmt.Sprintf(VPCDependencies+`
data "yandex_cloud_desktops_image" "image" {
	name 		= "Ubuntu 20.04 LTS (2024-12-03)"
	folder_id 	= "%s"
}

resource "yandex_cloud_desktops_desktop_group" "desktop_group" {
	name 		= "%s"
	folder_id 	= "%s"
	image_id 	= data.yandex_cloud_desktops_image.image.id
	description = "Sample Description"
	
	desktop_template = {
		resources = {
			cores 			= 4
			memory 			= 8
			core_fraction 	= 100
		}
		boot_disk = {
			initialize_params = {
				size = 24
				type = "SSD"
			}
		}
		data_disk = {
			initialize_params = {
				size = 16
				type = "HDD"
			}
		}
		network_interface = {
			network_id = yandex_vpc_network.network.id
			subnet_ids = ["${yandex_vpc_subnet.subnet.id}"]
		}
	}
	group_config = {
		min_ready_desktops 	= 1
		max_desktops_amount = 5
		desktop_type 		= "PERSISTENT"
		members				= [
			{
				id 		= "%s"
				type 	= "userAccount"
			}
		]
	}
		
	labels = {
    	label1 = "label1-value"
    	label2 = "label2-value"
  	}
}

data "yandex_cloud_desktops_desktop_group" "data_desktop_group_name_and_folder" {
	name 		= yandex_cloud_desktops_desktop_group.desktop_group.name
	folder_id 	= yandex_cloud_desktops_desktop_group.desktop_group.folder_id
}

data "yandex_cloud_desktops_desktop_group" "data_desktop_group_only_name" {
	name 		= yandex_cloud_desktops_desktop_group.desktop_group.name
}

data "yandex_cloud_desktops_desktop_group" "data_desktop_group_id" {
	desktop_group_id = yandex_cloud_desktops_desktop_group.desktop_group.desktop_group_id
}
`, test.GetExampleFolderID(), groupName, test.GetExampleFolderID(), test.GetExampleUserID1())
}

func testAccCheckCloudDesktopsDesktopGroupCheckDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_cloud_desktops_desktop_group" {
			continue
		}

		name, folderID, _, err := desktop_group.DeconstructID(rs.Primary.ID)
		if err != nil {
			return err
		}

		dGroups, err := config.SDK.CloudDesktop().DesktopGroup().List(context.Background(), &clouddesktop.ListDesktopGroupsRequest{
			FolderId: folderID,
		})
		if err != nil {
			if validate.IsStatusWithCode(err, codes.NotFound) {
				return fmt.Errorf("Can't find groups in the folder: %w", err)
			}
			return fmt.Errorf("Error checking for Desktop Group existence: %w", err)
		}

		for len(dGroups.DesktopGroups) != 0 {
			for _, dg := range dGroups.DesktopGroups {
				if dg.Name == name {
					return fmt.Errorf("Desktop Group %s still exists", dg.Name)
				}
			}

			if dGroups.NextPageToken == "" {
				break
			}

			dGroups, err = config.SDK.CloudDesktop().DesktopGroup().List(context.Background(), &clouddesktop.ListDesktopGroupsRequest{
				FolderId:  folderID,
				PageToken: dGroups.NextPageToken,
			})
			if err != nil {
				return fmt.Errorf("Error checking for Desktop Group existence: %w", err)
			}
		}
	}

	return nil
}
