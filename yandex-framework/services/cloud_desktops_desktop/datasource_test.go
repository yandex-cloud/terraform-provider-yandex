package cloud_desktops_desktop_test

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
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/cloud_desktops_desktop"
	"google.golang.org/grpc/codes"
)

func TestAccDataSourceCloudDesktopsDesktop_basic(t *testing.T) {
	t.Parallel()

	desktopGroupName := acctest.RandomWithPrefix("tf-desktop-group")
	desktopName := acctest.RandomWithPrefix("tf-desktop-desktop")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudDesktopsDesktopCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCloudDesktopsDesktopConfigStep1(desktopGroupName, desktopName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceCloudDesktopsDesktopCheck(
						"data.yandex_cloud_desktops_desktop.data_desktop_name_and_folder", "yandex_cloud_desktops_desktop.desktop"),
					testAccDataSourceCloudDesktopsDesktopCheck(
						"data.yandex_cloud_desktops_desktop.data_desktop_name", "yandex_cloud_desktops_desktop.desktop"),
					testAccDataSourceCloudDesktopsDesktopCheck(
						"data.yandex_cloud_desktops_desktop.data_desktop_id", "yandex_cloud_desktops_desktop.desktop"),
				),
			},
			{
				Config: testAccDataSourceCloudDesktopsDesktopConfigStep2(desktopGroupName, desktopName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceCloudDesktopsDesktopCheck(
						"data.yandex_cloud_desktops_desktop.data_desktop_name_and_folder", "yandex_cloud_desktops_desktop.desktop"),
					testAccDataSourceCloudDesktopsDesktopCheck(
						"data.yandex_cloud_desktops_desktop.data_desktop_name", "yandex_cloud_desktops_desktop.desktop"),
					testAccDataSourceCloudDesktopsDesktopCheck(
						"data.yandex_cloud_desktops_desktop.data_desktop_id", "yandex_cloud_desktops_desktop.desktop"),
				),
			},
		},
	})
}

func testAccDataSourceCloudDesktopsDesktopCheck(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		toCheck := []string{
			"desktop_id",
			"name",
			"desktop_group_id",
			"members.#",
			"members.0.subject_id",
			"members.0.subject_type",
			"labels.%",
			"labels.label1",
			"labels.label2",
		}
		for _, attrKey := range toCheck {
			var dsVal, rsVal string
			if dsVal, ok = ds.Primary.Attributes[attrKey]; !ok {
				return fmt.Errorf("data source has no attribute %s", attrKey)
			}
			if rsVal, ok = rs.Primary.Attributes[attrKey]; !ok {
				return fmt.Errorf("resource has no attribute %s", attrKey)
			}

			if dsVal != rsVal {
				return fmt.Errorf("data source attribute %s doesn't match one in resource: %s != %s", attrKey, dsVal, rsVal)
			}
		}

		return nil
	}
}

func testAccCheckCloudDesktopsDesktopCheckDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		// need to do this, since sometimes it finds datasource instead of resource and datasource doesn't have an ID
		if rs.Type != "yandex_cloud_desktops_desktop" || rs.Primary.ID == "" || rs.Primary.ID == "id-attribute-not-set" {
			continue
		}

		desktopId, _, err := cloud_desktops_desktop.DeconstructID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = config.SDK.CloudDesktop().Desktop().Get(context.Background(), &clouddesktop.GetDesktopRequest{
			DesktopId: desktopId,
		})
		if err == nil {
			return fmt.Errorf("Cloud Desktop still exists")
		}
		if !validate.IsStatusWithCode(err, codes.NotFound) {
			return fmt.Errorf("Got error different from Not Found: %s", err.Error())
		}
	}
	return nil
}

func testAccDataSourceCloudDesktopsDesktopConfigStep0(groupName string) string {
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
		min_ready_desktops 	= 0
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
`, test.GetExampleFolderID(), groupName, test.GetExampleFolderID(), test.GetExampleUserID1())
}

func testAccDataSourceCloudDesktopsDesktopConfigStep1(groupName, desktopName string) string {
	return fmt.Sprintf(testAccDataSourceCloudDesktopsDesktopConfigStep0(groupName)+`
resource "yandex_cloud_desktops_desktop" "desktop" {
	name 				= "%s"
	desktop_group_id 	= yandex_cloud_desktops_desktop_group.desktop_group.desktop_group_id
	
	network_interface = {
		subnet_id = yandex_vpc_subnet.subnet.id
	}

	members = [
		{
			subject_id 		= "%s"
			subject_type 	= "userAccount"
		},
	]

	labels = {
		label1 = "label1-value"
		label2 = "label2-value"
	}
}

data "yandex_cloud_desktops_desktop" "data_desktop_name_and_folder" {
	name = yandex_cloud_desktops_desktop.desktop.name
	folder_id = "%s"
}

data "yandex_cloud_desktops_desktop" "data_desktop_name" {
	name = yandex_cloud_desktops_desktop.desktop.name
}

data "yandex_cloud_desktops_desktop" "data_desktop_id" {
	desktop_id = yandex_cloud_desktops_desktop.desktop.desktop_id
}
`, desktopName, test.GetExampleUserID1(), test.GetExampleFolderID())
}

func testAccDataSourceCloudDesktopsDesktopConfigStep2(groupName, desktopName string) string {
	return fmt.Sprintf(testAccDataSourceCloudDesktopsDesktopConfigStep0(groupName)+`
resource "yandex_cloud_desktops_desktop" "desktop" {
	name 				= "%s"
	desktop_group_id 	= yandex_cloud_desktops_desktop_group.desktop_group.desktop_group_id
	
	network_interface = {
		subnet_id = yandex_vpc_subnet.subnet.id
	}

	members = [
		{
			subject_id 		= "%s"
			subject_type 	= "userAccount"
		},
	]

	labels = {
		label1 = "label1-value"
		label2 = "label2-value-new"
	}
}

data "yandex_cloud_desktops_desktop" "data_desktop_name_and_folder" {
	name = yandex_cloud_desktops_desktop.desktop.name
	folder_id = "%s"
}

data "yandex_cloud_desktops_desktop" "data_desktop_name" {
	name = yandex_cloud_desktops_desktop.desktop.name
}

data "yandex_cloud_desktops_desktop" "data_desktop_id" {
	desktop_id = yandex_cloud_desktops_desktop.desktop.desktop_id
}
`, desktopName, test.GetExampleUserID1(), test.GetExampleFolderID())
}
