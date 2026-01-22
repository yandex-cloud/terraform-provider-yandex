package mdb_clickhouse_user_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	chVersion               = "25.8"
	chClusterResourceID     = "yandex_mdb_clickhouse_cluster.sewage"
	chClusterResourceIDLink = "yandex_mdb_clickhouse_cluster.sewage.id"
	chDBResourceName1       = "pepperoni"
	chDBResourceName2       = "margarita"
	chUserResourceName0     = "splinter" // does not participate in the tests
	chUserResourceName1     = "leonardo"
	chUserResourceName2     = "michelangelo"
	chUserResourceName3     = "donatello"
	chUserResourceName4     = "raphael"

	clickHouseVPCDependencies = `
resource "yandex_vpc_network" "mdb-ch-test-net" {}

resource "yandex_vpc_subnet" "mdb-ch-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-ch-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

`
)

func testAccMDBClickHouseClusterConfigMain(name, desc string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+`
	resource "yandex_mdb_clickhouse_cluster" "sewage" {
	  name           = "%s"
	  description    = "%s"
	  environment    = "PRESTABLE"
	  version        = "%s"
	  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
	  admin_password = "strong_password"

	  labels = {
	    test_key = "test_value"
	  }

	  clickhouse {
	    resources {
	      resource_preset_id = "s2.micro"
	      disk_type_id       = "network-ssd"
	      disk_size          = 16
	    }
	  }

	  host {
	    type      = "CLICKHOUSE"
	    zone      = "ru-central1-a"
	    subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
		shard_name = "shard1"
	  }

	  timeouts {
		create = "1h"
		update = "1h"
		delete = "30m"
	  }

	  lifecycle {
	    ignore_changes = [user, database,]
	  }
	}

	resource "yandex_mdb_clickhouse_database" "pepperoni" {
		depends_on = [yandex_mdb_clickhouse_cluster.sewage]
		cluster_id = yandex_mdb_clickhouse_cluster.sewage.id
		name       = "%s"
	}

	resource "yandex_mdb_clickhouse_database" "margarita" {
		depends_on = [yandex_mdb_clickhouse_cluster.sewage]
		cluster_id = yandex_mdb_clickhouse_cluster.sewage.id
		name       = "%s"
	}
	
	`, name, desc, chVersion, chDBResourceName1, chDBResourceName2)
}

func testAccCheckMDBClickHouseUserDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_clickhouse_user" {
			continue
		}

		clusterId, userName, err := resourceid.Deconstruct(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = config.SDK.MDB().Clickhouse().User().Get(context.Background(), &clickhouse.GetUserRequest{
			ClusterId: clusterId,
			UserName:  userName,
		})

		if err == nil {
			return fmt.Errorf("Clickhouse user still exists")
		}
	}

	return nil
}

func testAccCheckMDBClickHouseUserResourceIDField(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		expectedResourceId := resourceid.Construct(rs.Primary.Attributes["cluster_id"], rs.Primary.Attributes["name"])

		if expectedResourceId != rs.Primary.ID {
			return fmt.Errorf("Wrong resource %s id. Expected %s, got %s", resourceName, expectedResourceId, rs.Primary.ID)
		}

		return nil
	}
}

func makeCHUserResource(name string) string {
	return fmt.Sprintf("yandex_mdb_clickhouse_user.%s", name)
}

func makeCHDBResource(name string) string {
	return fmt.Sprintf("yandex_mdb_clickhouse_database.%s", name)
}
