package cloud_desktops_desktop_group_test

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
	desktopGroupName   = "desktop_group"
	desktopGroupName1  = desktopGroupName + "_1"
	desktopGroupName2  = desktopGroupName + "_2"
	desktopGroupPrefix = "yandex_cloud_desktops_desktop_group."
	description1       = "Cloud Desktops Desktop Group Test"
	description2       = description1 + " Updated"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccResourceCloudDesktopsDesktopGroup_full(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-desktop-group-resource")
	groupName1 := groupName + "-1"
	groupName2 := groupName + "-2"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCloudDesktopsDesktopGroupConfigStep1(groupName1, description1, desktopGroupName1, false),
				Check:  testAccResourceCloudDesktopsDesktopGroupEqualityCheck(groupName1, description1, desktopGroupName1),
			},
			testsImportStep(desktopGroupName1),
			{
				Config: testAccResourceCloudDesktopsDesktopGroupConfigStep1(groupName1, description2, desktopGroupName1, false),
				Check:  testAccResourceCloudDesktopsDesktopGroupEqualityCheck(groupName1, description2, desktopGroupName1),
			},
			testsImportStep(desktopGroupName1),
			{
				Config: testAccResourceCloudDesktopsDesktopGroupConfigStep1(groupName2, description1, desktopGroupName2, true),
				Check:  testAccResourceCloudDesktopsDesktopGroupEqualityCheck(groupName2, description1, desktopGroupName2),
			},
			testsImportStep(desktopGroupName2),
			{
				Config: testAccResourceCloudDesktopsDesktopGroupConfigStep1(groupName2, description2, desktopGroupName2, true),
				Check:  testAccResourceCloudDesktopsDesktopGroupEqualityCheck(groupName2, description2, desktopGroupName2),
			},
			testsImportStep(desktopGroupName2),
		},
	})
}

func testAccResourceCloudDesktopsDesktopGroupEqualityCheck(name, description, terraformName string) resource.TestCheckFunc {
	checkArray := [][2]string{
		{"name", name},
		{"folder_id", test.GetExampleFolderID()},
		{"image_id", "fdvvheamqk751hr09co9"},
		{"description", description},
		{"desktop_template.resources.cores", "4"},
		{"desktop_template.resources.memory", "8"},
		{"desktop_template.resources.core_fraction", "100"},
		{"desktop_template.boot_disk.initialize_params.size", "24"},
		{"desktop_template.boot_disk.initialize_params.type", "SSD"},
		{"desktop_template.data_disk.initialize_params.size", "16"},
		{"desktop_template.data_disk.initialize_params.type", "HDD"},
		{"group_config.min_ready_desktops", "1"},
		{"group_config.max_desktops_amount", "5"},
		{"group_config.desktop_type", "PERSISTENT"},
		{"group_config.members.#", "1"},
		{"group_config.members.0.id", test.GetExampleUserID1()},
		{"group_config.members.0.type", "userAccount"},
		{"labels.%", "2"},
		{"labels.label1", "label1-value"},
		{"labels.label2", "label2-value"},
	}

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[desktopGroupPrefix+terraformName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", terraformName)
		}

		for _, entry := range checkArray {
			val, ok := rs.Primary.Attributes[entry[0]]
			if !ok {
				return fmt.Errorf("resource has no resource named %s", entry[0])
			}

			if val != entry[1] {
				return fmt.Errorf("resource attribute %s value isn't the expected: %s != %s", entry[0], val, entry[1])
			}
		}

		return nil
	}
}

func testsImportStep(terraformName string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      desktopGroupPrefix + terraformName,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccResourceCloudDesktopsDesktopGroupConfigStep0() string {
	return fmt.Sprintf(VPCDependencies+`
data "yandex_cloud_desktops_image" "image" {
	name 		= "Ubuntu 20.04 LTS (2024-12-03)"
	folder_id 	= "%s"
}
`, test.GetExampleFolderID())
}

func testAccResourceCloudDesktopsDesktopGroupConfigStep1(name, description, terraformName string, withoutFolder bool) string {
	folderIdLine := ""
	if !withoutFolder {
		folderIdLine = fmt.Sprintf("folder_id = \"%s\"", test.GetExampleFolderID())
	}

	return fmt.Sprintf(testAccResourceCloudDesktopsDesktopGroupConfigStep0()+`
resource "yandex_cloud_desktops_desktop_group" "%s" {
	name 		= "%s"
	%s
	image_id 	= data.yandex_cloud_desktops_image.image.id
	description = "%s"
	
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
`, terraformName, name, folderIdLine, description, test.GetExampleUserID1())
}
