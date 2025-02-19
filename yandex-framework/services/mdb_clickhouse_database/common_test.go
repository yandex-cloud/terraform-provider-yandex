package mdb_clickhouse_database_test

import (
	"context"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	chVersion               = "24.3"
	chClusterResourceID     = "yandex_mdb_clickhouse_cluster.sewage"
	chClusterResourceIDLink = "yandex_mdb_clickhouse_cluster.sewage.id"
	chDBResourceName0       = "splinter" // does not participate in the tests
	chDBResourceName1       = "leonardo"
	chDBResourceName2       = "michelangelo"
	chDBResourceName3       = "donatello"
	chDBResourceName4       = "raphael"
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

	`, name, desc, chVersion)
}

const clickHouseVPCDependencies = `
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

func testAccCheckMDBClickHouseDatabaseDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_clickhouse_database" {
			continue
		}

		clusterId, dbName, err := resourceid.Deconstruct(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = config.SDK.MDB().Clickhouse().Database().Get(context.Background(), &clickhouse.GetDatabaseRequest{
			ClusterId:    clusterId,
			DatabaseName: dbName,
		})

		if err == nil {
			return fmt.Errorf("Clickhouse database still exists")
		}
	}

	return nil
}

func formatResourceName(name string) string {
	return fmt.Sprintf("yandex_mdb_clickhouse_database.%s", name)
}

type resourceIDCompare struct {
	resourceAddress     string
	attributeFirstPath  tfjsonpath.Path
	attributeSecondPath tfjsonpath.Path
}

func makeClickHouseDatabaseResourceIDComparer(resourceID string) statecheck.StateCheck {
	return &resourceIDCompare{
		resourceAddress:     resourceID,
		attributeFirstPath:  tfjsonpath.New("cluster_id"),
		attributeSecondPath: tfjsonpath.New("name"),
	}
}

func (e *resourceIDCompare) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	var resource *tfjson.StateResource

	if req.State == nil {
		resp.Error = fmt.Errorf("state is nil")

		return
	}

	if req.State.Values == nil {
		resp.Error = fmt.Errorf("state does not contain any state values")

		return
	}

	if req.State.Values.RootModule == nil {
		resp.Error = fmt.Errorf("state does not contain a root module")

		return
	}

	for _, r := range req.State.Values.RootModule.Resources {
		if e.resourceAddress == r.Address {
			resource = r

			break
		}
	}

	if resource == nil {
		resp.Error = fmt.Errorf("%s - Resource not found in state", e.resourceAddress)

		return
	}

	idState, err := tfjsonpath.Traverse(resource.AttributeValues, tfjsonpath.New("id"))
	if err != nil {
		resp.Error = err
		return
	}

	if idState == nil || idState.(string) == "" {
		resp.Error = fmt.Errorf("ID for resource %s is not setted", e.resourceAddress)
		return
	}

	firstPartState, err := tfjsonpath.Traverse(resource.AttributeValues, e.attributeFirstPath)

	if err != nil {
		resp.Error = err
		return
	}

	secondPartState, err := tfjsonpath.Traverse(resource.AttributeValues, e.attributeSecondPath)
	if err != nil {
		resp.Error = err
		return
	}

	expectedResourceId := resourceid.Construct(firstPartState.(string), secondPartState.(string))

	if expectedResourceId != idState.(string) {
		resp.Error = fmt.Errorf("Wrong resource id. Expected %v, got %v", expectedResourceId, idState)
		return
	}
}
