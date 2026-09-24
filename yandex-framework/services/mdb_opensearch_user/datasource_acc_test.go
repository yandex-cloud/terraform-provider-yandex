package mdb_opensearch_user_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccDataSourceMDBOpenSearchUser(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-opensearch-user-ds")
	dataSourceName := "data.yandex_mdb_opensearch_user.read_only"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBOpenSearchUserConfig(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", "read_only"),
					resource.TestCheckResourceAttrSet(dataSourceName, "cluster_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "connection_manager.connection_id"),
				),
			},
		},
	})
}

func testAccMDBOpenSearchUserConfig(clusterName string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "test" {}

resource "yandex_vpc_subnet" "test" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.test.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}

resource "yandex_mdb_opensearch_cluster" "test" {
  name        = %q
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.test.id

  config {
    admin_password = "dummy_P@ssw0rd"

    opensearch {
      node_groups {
        name             = "group0"
        assign_public_ip = false
        hosts_count      = 1
        subnet_ids       = [yandex_vpc_subnet.test.id]
        zone_ids         = ["ru-central1-d"]
        roles            = ["DATA", "MANAGER"]

        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }
    }
  }

  maintenance_window {
    type = "ANYTIME"
  }
}

data "yandex_mdb_opensearch_user" "read_only" {
  cluster_id = yandex_mdb_opensearch_cluster.test.id
  name       = "read_only"
}
`, clusterName)
}
