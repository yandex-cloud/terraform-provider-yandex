package cloud_desktops_desktop_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	VPCDependencies = `
resource "yandex_vpc_network" "network" {}

resource "yandex_vpc_subnet" "subnet" {
	zone 			= "ru-central1-a"
	network_id 		= yandex_vpc_network.network.id
	v4_cidr_blocks 	= ["10.1.0.0/24"]
}
`
	desktopName      = "yandex_cloud_desktops_desktop.desktop"
	desktopGroupName = "yandex_cloud_desktops_desktop_group.desktop_group"
	vpcSubnetName    = "yandex_vpc_subnet.subnet"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccResourceCloudDesktopsDesktop_full(t *testing.T) {
	t.Parallel()

	desktopGroupName := acctest.RandomWithPrefix("tf-desktop-group")
	desktopName := acctest.RandomWithPrefix("tf-desktop-desktop")
	desktopName1 := desktopName + "-1"
	desktopName2 := desktopName + "-2"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCloudDesktopsDesktopConfigStep2(desktopName1, desktopGroupName),
				Check:  testAccResourceCloudDesktopsDesktopEqualityCheck(desktopName1),
			},
			testsImportStep(),
			{
				Config: testAccResourceCloudDesktopsDesktopConfigStep2(desktopName2, desktopGroupName),
				Check:  testAccResourceCloudDesktopsDesktopEqualityCheck(desktopName2),
			},
			testsImportStep(),
		},
	})
}

func testAccResourceCloudDesktopsDesktopEqualityCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[desktopName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", desktopName)
		}

		curName := rs.Primary.Attributes["name"]
		if curName != name {
			return fmt.Errorf("resource name is not equal to the expected: %s != %s", curName, name)
		}

		group, ok := s.RootModule().Resources[desktopGroupName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", desktopGroupName)
		}

		desktopGroupId := group.Primary.Attributes["desktop_group_id"]
		rsDesktopGroupId := rs.Primary.Attributes["desktop_group_id"]
		if desktopGroupId != rsDesktopGroupId {
			return fmt.Errorf("resource desktop_group_id is not the expected: %s != %s", rsDesktopGroupId, desktopGroupId)
		}

		vpc, ok := s.RootModule().Resources[vpcSubnetName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", vpcSubnetName)
		}

		networkId := vpc.Primary.Attributes["id"]
		rsNetworkId := rs.Primary.Attributes["network_interface.subnet_id"]
		if networkId != rsNetworkId {
			return fmt.Errorf("resource network_id is not the expected: %s != %s", rsNetworkId, networkId)
		}

		labelsNum := rs.Primary.Attributes["labels.%"]
		expectedLabelsNum := "2"
		if labelsNum != expectedLabelsNum {
			return fmt.Errorf("resource labels size is not the expected: %s != %s", labelsNum, expectedLabelsNum)
		}

		membersNum := rs.Primary.Attributes["members.#"]
		expectedMembersNum := "1"
		if membersNum != expectedMembersNum {
			return fmt.Errorf("resource members size is not the expected: %s != %s", membersNum, expectedMembersNum)
		}

		membersId := rs.Primary.Attributes["members.0.subject_id"]
		expectedMembersId := test.GetExampleUserID1()
		if membersId != expectedMembersId {
			return fmt.Errorf("resource member is is not the expected: %s != %s", membersId, expectedMembersId)
		}

		return nil
	}
}

func testsImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      desktopName,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccResourceCloudDesktopsDesktopConfigStep0() string {
	return fmt.Sprintf(VPCDependencies+`
data "yandex_cloud_desktops_image" "image" {
	name 		= "Ubuntu 20.04 LTS (2024-12-03)"
	folder_id 	= "%s"
}
`, test.GetExampleFolderID())
}

func testAccResourceCloudDesktopsDesktopConfigStep1(name string) string {
	return fmt.Sprintf(testAccResourceCloudDesktopsDesktopConfigStep0()+`
resource "yandex_cloud_desktops_desktop_group" "desktop_group" {
	name 		= "%s"
	folder_id 	= "%s"
	image_id 	= data.yandex_cloud_desktops_image.image.id
	description = "sample description"
	
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
`, name, test.GetExampleFolderID(), test.GetExampleUserID1())
}

func testAccResourceCloudDesktopsDesktopConfigStep2(name, groupName string) string {
	return fmt.Sprintf(testAccResourceCloudDesktopsDesktopConfigStep1(groupName)+`
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
`, name, test.GetExampleUserID1())
}
