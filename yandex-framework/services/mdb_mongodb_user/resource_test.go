package mdb_mongodb_user_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"

	"golang.org/x/exp/slices"
)

const (
	mgClusterResourceName   = "yandex_mdb_mongodb_cluster.foo"
	mgUserResourceNameAlice = "yandex_mdb_mongodb_user.alice"
	mgUserResourceNameBob   = "yandex_mdb_mongodb_user.bob"

	VPCDependencies = `
	resource "yandex_vpc_network" "foo" {}
	
	resource "yandex_vpc_subnet" "foo" {
	  zone           = "ru-central1-a"
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

// Test that a MongoDB User can be created, updated and destroyed
func TestAccMDBMongoDBUser_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-mongodb-user")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMongoDBUserConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgUserResourceNameAlice, "name", "alice"),
					testAccCheckMDBMongoDBUserHasPermission(t, "alice", nil),
				),
			},
			mdbMongoDBUserImportStep(mgUserResourceNameAlice),
			{
				Config: testAccMDBMongoDBUserConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgUserResourceNameBob, "name", "bob"),
					testAccCheckMDBMongoDBUserHasPermission(t, "bob",
						[]Permission{{DatabaseName: "testdb", Roles: []string{"readWrite", "read"}}}),
				),
			},
			mdbMongoDBUserImportStep(mgUserResourceNameAlice),
			mdbMongoDBUserImportStep(mgUserResourceNameBob),
			{
				Config: testAccMDBMongoDBUserConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgUserResourceNameAlice, "name", "alice"),
					testAccCheckMDBMongoDBUserHasPermission(t, "alice",
						[]Permission{{DatabaseName: "testdb", Roles: []string{"readWrite"}}}),
				),
			},
			mdbMongoDBUserImportStep(mgUserResourceNameAlice),
		},
	})
}

func mdbMongoDBUserImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password", // password is not returned
		},
	}
}

func testAccLoadMongoDBUser(s *terraform.State, username string) (*mongodb.User, error) {
	rs, ok := s.RootModule().Resources[mgClusterResourceName]

	if !ok {
		return nil, fmt.Errorf("resource %q not found", mgUserResourceNameAlice)
	}
	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no ID is set")
	}

	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	return config.SDK.MDB().MongoDB().User().Get(context.Background(), &mongodb.GetUserRequest{
		ClusterId: rs.Primary.ID,
		UserName:  username,
	})
}

func testAccCheckMDBMongoDBUserHasPermission(t *testing.T, username string, expected []Permission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		user, err := testAccLoadMongoDBUser(s, username)
		if err != nil {
			return err
		}
		var actual []Permission
		for _, permission := range user.Permissions {
			slices.Sort(permission.Roles)
			actual = append(actual, Permission{DatabaseName: permission.DatabaseName, Roles: permission.Roles})
		}
		for _, permission := range expected {
			slices.Sort(permission.Roles)
		}

		cmp := func(a, b Permission) int {
			if a.DatabaseName > b.DatabaseName {
				return 1
			} else if a.DatabaseName < b.DatabaseName {
				return -1
			} else {
				return 0
			}
		}
		slices.SortFunc(expected, cmp)
		slices.SortFunc(actual, cmp)
		assert.Equal(t, expected, actual)

		return nil
	}
}

func testAccMDBMongoDBUserConfigStep0(name string) string {
	return fmt.Sprintf(VPCDependencies+`
resource "yandex_mdb_mongodb_cluster" "foo" {
	name        = "%s"
	description = "MongoDB User Terraform Test"
	environment = "PRESTABLE"
	network_id  = yandex_vpc_network.foo.id

	cluster_config {
	    version = "6.0"
	}

	host {
		zone_id      = "ru-central1-a"
		subnet_id  = yandex_vpc_subnet.foo.id
	}
	resources_mongod {
		  resource_preset_id = "s2.micro"
		  disk_size          = 10
		  disk_type_id       = "network-ssd"
	    }
}

resource "yandex_mdb_mongodb_database" "testdb" {
	cluster_id = yandex_mdb_mongodb_cluster.foo.id
	name       = "testdb"
}
`, name)
}

// Create cluster, user and database
func testAccMDBMongoDBUserConfigStep1(name string) string {
	return testAccMDBMongoDBUserConfigStep0(name) + `
resource "yandex_mdb_mongodb_user" "alice" {
	cluster_id = yandex_mdb_mongodb_cluster.foo.id
	name       = "alice"
	password   = "mysecureP@ssw0rd"
}`
}

// Create another user and give permission to database
func testAccMDBMongoDBUserConfigStep2(name string) string {
	return testAccMDBMongoDBUserConfigStep1(name) + `
resource "yandex_mdb_mongodb_user" "bob" {
	cluster_id = yandex_mdb_mongodb_cluster.foo.id
	name       = "bob"
	password   = "mysecureP@ssw0rd"
	permission {
    	database_name=yandex_mdb_mongodb_database.testdb.name
    	roles = ["readWrite", "read"]
  	}
}`
}

// Change Alice's permissions
func testAccMDBMongoDBUserConfigStep3(name string) string {
	return testAccMDBMongoDBUserConfigStep0(name) + `
resource "yandex_mdb_mongodb_user" "alice" {
	cluster_id = yandex_mdb_mongodb_cluster.foo.id
	name       = "alice"
	password   = "mysecureP@ssw0rd"
	permission {
    	database_name=yandex_mdb_mongodb_database.testdb.name
    	roles = ["readWrite"]
  	}
}`
}
