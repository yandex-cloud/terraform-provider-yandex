package mdb_greenplum_resource_group_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	mgResourceGroupResourceName1 = "yandex_mdb_greenplum_resource_group.resource_group1"
	mgResourceGroupResourceName2 = "yandex_mdb_greenplum_resource_group.resource_group2"

	VPCDependencies = `
	resource "yandex_vpc_network" "foo" {}
	
	resource "yandex_vpc_subnet" "foo" {
	  zone           = "ru-central1-b"
	  network_id     = yandex_vpc_network.foo.id
	  v4_cidr_blocks = ["10.1.0.0/24"]
	}
	`
)

type Permission struct {
	DatabaseName string
	Roles        []string
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// Test that a Greenplum Resource Group can be created, updated and destroyed
func TestAccMDBGreenplumResourceGroup_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-greenplum-resource-group")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBGreenplumResourceGroupConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "name", "resource_group1"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "is_user_defined", "true"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "concurrency", "10"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "cpu_rate_limit", "10"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "memory_limit", "10"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "memory_shared_quota", "10"),
				),
			},
			mdbGreenplumResourceGroupImportStep(mgResourceGroupResourceName1),
			{
				Config: testAccMDBGreenplumResourceGroupConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgResourceGroupResourceName2, "name", "resource_group2"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName2, "is_user_defined", "true"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName2, "concurrency", "15"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName2, "cpu_rate_limit", "25"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName2, "memory_limit", "35"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName2, "memory_shared_quota", "45"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName2, "memory_spill_ratio", "55"),
				),
			},
			mdbGreenplumResourceGroupImportStep(mgResourceGroupResourceName1),
			mdbGreenplumResourceGroupImportStep(mgResourceGroupResourceName2),
			{
				Config: testAccMDBGreenplumResourceGroupConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "name", "resource_group1"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "is_user_defined", "true"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "concurrency", "15"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "cpu_rate_limit", "25"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "memory_limit", "35"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "memory_shared_quota", "45"),
					resource.TestCheckResourceAttr(mgResourceGroupResourceName1, "memory_spill_ratio", "55"),
				),
			},
			mdbGreenplumResourceGroupImportStep(mgResourceGroupResourceName1),
		},
	})
}

func mdbGreenplumResourceGroupImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password", // password is not returned
		},
	}
}

func testAccMDBGreenplumResourceGroupConfigStep0(name string) string {
	return fmt.Sprintf(VPCDependencies+`
resource "yandex_mdb_greenplum_cluster" "foo" {
	name        = "%s"
	description = "greenplum ResourceGroup Terraform Test"
	environment = "PRESTABLE"
	network_id  = yandex_vpc_network.foo.id

	zone = "ru-central1-b"
	subnet_id = yandex_vpc_subnet.foo.id
	assign_public_ip = false
	version = "6.28"
	
	labels = { test_key_create : "test_value_create" }
	
	master_host_count  = 2
	
	master_subcluster {
		resources {
			resource_preset_id = "s2.micro"
			disk_size          = 24
			disk_type_id       = "network-ssd"
		}
	}
	segment_subcluster {
		resources {
			resource_preset_id = "s2.small"
			disk_size          = 24
			disk_type_id       = "network-ssd"
		}
	}

	segment_host_count = 2
	segment_in_host = 2
	
	user_name     = "user1"
	user_password = "mysecurepassword"
}
`, name)
}

// Create cluster, resource group and database
func testAccMDBGreenplumResourceGroupConfigStep1(name string) string {
	return testAccMDBGreenplumResourceGroupConfigStep0(name) + `
resource "yandex_mdb_greenplum_resource_group" "resource_group1" {
	cluster_id          = yandex_mdb_greenplum_cluster.foo.id
	name                = "resource_group1"
	concurrency         = 10
	cpu_rate_limit      = 10
	memory_limit        = 10
	memory_shared_quota = 10
}`
}

// Create another resource group
func testAccMDBGreenplumResourceGroupConfigStep2(name string) string {
	return testAccMDBGreenplumResourceGroupConfigStep1(name) + `
resource "yandex_mdb_greenplum_resource_group" "resource_group2" {
	cluster_id          = yandex_mdb_greenplum_cluster.foo.id
	name                = "resource_group2"
	concurrency         = 15
	cpu_rate_limit      = 25
	memory_limit        = 35
	memory_shared_quota = 45
	memory_spill_ratio  = 55
}`
}

// Change resource_group1 attrs
func testAccMDBGreenplumResourceGroupConfigStep3(name string) string {
	return testAccMDBGreenplumResourceGroupConfigStep0(name) + `
resource "yandex_mdb_greenplum_resource_group" "resource_group1" {
	cluster_id = yandex_mdb_greenplum_cluster.foo.id
	name       = "resource_group1"
	concurrency         = 15
	cpu_rate_limit      = 25
	memory_limit        = 35
	memory_shared_quota = 45
	memory_spill_ratio  = 55
}`
}
