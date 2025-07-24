package mdb_greenplum_user_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	mgUserResourceNameAlice = "yandex_mdb_greenplum_user.alice"
	mgUserResourceNameBob   = "yandex_mdb_greenplum_user.bob"

	VPCDependencies = `
	resource "yandex_vpc_network" "foo" {}
	
	resource "yandex_vpc_subnet" "foo" {
	  zone           = "ru-central1-b"
	  network_id     = yandex_vpc_network.foo.id
	  v4_cidr_blocks = ["10.1.0.0/24"]
	}
	`
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// Test that a Greenplum User can be created, updated and destroyed
func TestAccMDBGreenplumUser_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-greenplum-user")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBGreenplumUserConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(mgUserResourceNameAlice, "resource_group", "some_group2"),
				),
			},
			mdbGreenplumUserImportStep(mgUserResourceNameAlice),
			{
				Config: testAccMDBGreenplumUserConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgUserResourceNameBob, "name", "bob"),
					resource.TestCheckResourceAttr(mgUserResourceNameBob, "resource_group", "some_group2"),
				),
			},
			mdbGreenplumUserImportStep(mgUserResourceNameAlice),
			mdbGreenplumUserImportStep(mgUserResourceNameBob),
			{
				Config: testAccMDBGreenplumUserConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(mgUserResourceNameAlice, "resource_group", "some_group1"),
				),
			},
			mdbGreenplumUserImportStep(mgUserResourceNameAlice),
		},
	})
}

func mdbGreenplumUserImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password", // password is not returned
		},
	}
}

func testAccMDBGreenplumUserConfigStep0(name string) string {
	return fmt.Sprintf(VPCDependencies+`
resource "yandex_mdb_greenplum_cluster" "foo" {
	name        = "%s"
	description = "greenplum User Terraform Test"
	environment = "PRESTABLE"
	network_id  = yandex_vpc_network.foo.id

	zone = "ru-central1-b"
	subnet_id = yandex_vpc_subnet.foo.id
	assign_public_ip = false
	version = "6.25"
	
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

resource "yandex_mdb_greenplum_resource_group" "some_group1" {
	cluster_id     = yandex_mdb_greenplum_cluster.foo.id
	name           = "some_group1"
	cpu_rate_limit = 25
}

resource "yandex_mdb_greenplum_resource_group" "some_group2" {
	cluster_id     = yandex_mdb_greenplum_cluster.foo.id
	name           = "some_group2"
	cpu_rate_limit = 25
}
`, name)
}

// Create cluster, user and database
func testAccMDBGreenplumUserConfigStep1(name string) string {
	return testAccMDBGreenplumUserConfigStep0(name) + `
resource "yandex_mdb_greenplum_user" "alice" {
	cluster_id     = yandex_mdb_greenplum_cluster.foo.id
	name           = "alice"
	password       = "mysecureP@ssw0rd"
	resource_group = yandex_mdb_greenplum_resource_group.some_group2.name
}`
}

// Create another user and give resource_group
func testAccMDBGreenplumUserConfigStep2(name string) string {
	return testAccMDBGreenplumUserConfigStep1(name) + `
resource "yandex_mdb_greenplum_user" "bob" {
	cluster_id = yandex_mdb_greenplum_cluster.foo.id
	name       = "bob"
	password   = "mysecureP@ssw0rd"
	resource_group = yandex_mdb_greenplum_resource_group.some_group2.name
}`
}

// Change Alice's resource_group
func testAccMDBGreenplumUserConfigStep3(name string) string {
	return testAccMDBGreenplumUserConfigStep0(name) + `
resource "yandex_mdb_greenplum_user" "alice" {
	cluster_id = yandex_mdb_greenplum_cluster.foo.id
	name       = "alice"
	password   = "mysecureP@ssw0rd"
	resource_group = yandex_mdb_greenplum_resource_group.some_group1.name
}`
}
