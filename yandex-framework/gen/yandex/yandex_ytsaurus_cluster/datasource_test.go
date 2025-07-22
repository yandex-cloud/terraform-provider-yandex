package yandex_ytsaurus_cluster_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const testYtsaurusClusterDataSourceName = "data.yandex_ytsaurus_cluster.test-cluster-data"

func TestAccYtsaurusClusterDataSource(t *testing.T) {
	var (
		clusterName = test.ResourceName(63)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYtsaurusClusterDataSourceConfig(clusterName),
				Check: resource.ComposeTestCheckFunc(
					test.YtsaurusClusterExists(testYtsaurusClusterDataSourceName),
					resource.TestCheckResourceAttr(testYtsaurusClusterDataSourceName, "name", clusterName),
					resource.TestCheckResourceAttrSet(testYtsaurusClusterDataSourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testYtsaurusClusterDataSourceName, "created_by"),
					resource.TestCheckResourceAttrSet(testYtsaurusClusterDataSourceName, "endpoints.ui"),
					resource.TestCheckResourceAttrSet(testYtsaurusClusterDataSourceName, "endpoints.external_http_proxy_balancer"),
					resource.TestCheckResourceAttrSet(testYtsaurusClusterDataSourceName, "endpoints.internal_http_proxy_alias"),
					resource.TestCheckResourceAttrSet(testYtsaurusClusterDataSourceName, "endpoints.internal_rpc_proxy_alias"),
					resource.TestCheckResourceAttr(testYtsaurusClusterDataSourceName, "status", "RUNNING"),
					test.AccCheckCreatedAtAttr(testYtsaurusClusterDataSourceName),
				),
			},
			ytsaurusClusterImportTestStep(),
		},
	})
}

func testYtsaurusClusterDataSourceConfig(clusterName string) string {
	return fmt.Sprintf(`
data "yandex_ytsaurus_cluster" "test-cluster-data" {
  cluster_id = yandex_ytsaurus_cluster.test-cluster.id
}

resource "yandex_vpc_network" "test-network" {}

resource "yandex_vpc_subnet" "test-subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.test-network.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_security_group" "test-security-group" {
  network_id = yandex_vpc_network.test-network.id

  ingress {
    protocol       = "TCP"
    description    = "healthchecks"
    port           = 30080
    v4_cidr_blocks = ["198.18.235.0/24", "198.18.248.0/24"]
  }
}

resource "yandex_ytsaurus_cluster" "test-cluster" {
  name = "%s"

  zone_id			 = "ru-central1-a"
  subnet_id			 = yandex_vpc_subnet.test-subnet.id
  security_group_ids = [yandex_vpc_security_group.test-security-group.id]

  spec = {
	storage = {
	  hdd = {
	  	size_gb = 100
		count 	= 3
	  }
	  ssd = {
	  	size_gb = 100
		type 	= "network-ssd"
		count 	= 3
	  }
	}
	compute = [{
	  preset = "c8-m32"
	  disks = [{
	  	type 	= "network-ssd"
		size_gb = 50
	  }]
	  scale_policy = {
	  	fixed = {
		  size = 1
		}
	  }
	}]
	tablet = {
      preset = "c8-m16"
	  count = 3
	}
	proxy = {
	  http = {
	  	count = 1
	  }
	  rpc = {
	  	count = 1
	  }
	}
  }
}
`, clusterName)
}
