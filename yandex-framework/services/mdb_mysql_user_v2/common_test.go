package mdb_mysql_user_v2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	mysqlUserV2ResourceName  = "yandex_mdb_mysql_user_v2.testuser"
	mysqlUserV2ResourceName1 = "yandex_mdb_mysql_user_v2.testuser1"
	mysqlClusterV2Resource   = "yandex_mdb_mysql_cluster_v2.foo"
)

const mysqlUserV2VPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}
`

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func mdbMySQLUserV2ImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password",
			"generate_password",
		},
	}
}

func testAccCheckMDBMySQLUserV2Exists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		clusterID, userName, err := resourceid.Deconstruct(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = config.SDK.MDB().MySQL().User().Get(
			context.Background(),
			&mysql.GetUserRequest{
				ClusterId: clusterID,
				UserName:  userName,
			},
		)
		if err != nil {
			return fmt.Errorf("MySQL user %q not found in cluster %q: %v", userName, clusterID, err)
		}
		return nil
	}
}

func testAccCheckMDBMySQLUserV2Destroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_mysql_user_v2" {
			continue
		}
		clusterID, userName, err := resourceid.Deconstruct(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to deconstruct resource ID %s: %w", rs.Primary.ID, err)
		}
		_, err = config.SDK.MDB().MySQL().User().Get(
			context.Background(),
			&mysql.GetUserRequest{
				ClusterId: clusterID,
				UserName:  userName,
			},
		)
		if err == nil {
			return fmt.Errorf(
				"MySQL user %q in cluster %q still exists",
				userName, clusterID,
			)
		}
	}
	return nil
}

func testAccCheckMDBMySQLUserV2ResourceIDField(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}
		expectedID := resourceid.Construct(
			rs.Primary.Attributes["cluster_id"],
			rs.Primary.Attributes["name"],
		)
		if expectedID != rs.Primary.ID {
			return fmt.Errorf(
				"wrong resource %s id: expected %s, got %s",
				resourceName, expectedID, rs.Primary.ID,
			)
		}
		return nil
	}
}

func testAccLoadMySQLUserV2(s *terraform.State, userName string) (*mysql.User, error) {
	rs, ok := s.RootModule().Resources[mysqlClusterV2Resource]
	if !ok {
		return nil, fmt.Errorf("resource %q not found", mysqlClusterV2Resource)
	}
	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no ID is set for cluster resource")
	}

	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	return config.SDK.MDB().MySQL().User().Get(
		context.Background(),
		&mysql.GetUserRequest{
			ClusterId: rs.Primary.ID,
			UserName:  userName,
		},
	)
}

func testAccCheckMDBMySQLClusterHasUserV2(
	t *testing.T,
	userName string,
	deletionProtectionMode string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		user, err := testAccLoadMySQLUserV2(s, userName)
		if err != nil {
			return err
		}
		if user.Name != userName {
			return fmt.Errorf("expected user name %q, got %q", userName, user.Name)
		}
		if user.DeletionProtectionMode.String() != deletionProtectionMode {
			return fmt.Errorf(
				"expected deletion_protection_mode %q, got %q",
				deletionProtectionMode,
				user.DeletionProtectionMode.String(),
			)
		}
		return nil
	}
}

func clusterConfigForUserTests(clusterName string) string {
	return fmt.Sprintf(mysqlUserV2VPCDependencies+`
resource "yandex_mdb_mysql_cluster_v2" "foo" {
  name        = "%s"
  description = "MySQL User V2 Terraform Test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  resources {
    resource_preset_id = "s2.micro"
    disk_type_id       = "network-ssd"
    disk_size          = 16
  }

  hosts = {
    "host1" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
    }
    "host2" = {
      zone      = "ru-central1-b"
      subnet_id = yandex_vpc_subnet.bar.id
    }
  }
}

resource "yandex_mdb_mysql_database_v2" "testdb" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "testdb"
}

resource "yandex_mdb_mysql_database_v2" "new_testdb" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "new_testdb"
  depends_on = [yandex_mdb_mysql_database_v2.testdb]
}
`, clusterName)
}
