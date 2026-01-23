package mdb_clickhouse_cluster_v2_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/kms_symmetric_key"
)

var ignoreAttrsSet = map[string]struct{}{
	// Senserive attributies:
	"admin_password":                        {},
	"clickhouse.config.kafka.sasl_password": {},
	"clickhouse.config.rabbitmq.password":   {},
}

var ignoreByPrefixAttrsSet = map[string]struct{}{
	"hosts": {}, // Keys are different between resource (alias) and datasource (fqdn).
}

func TestAccDataSourceMDBClickHouseClusterV2_byID(t *testing.T) {
	t.Parallel()

	chName := acctest.RandomWithPrefix("ds-ch-by-id")
	chDesc := "ClickHouseCluster Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBClickHouseClusterConfig(chName, chDesc, true),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceMDBClickHouseClusterAttributesCheck("data.yandex_mdb_clickhouse_cluster_v2.bar", "yandex_mdb_clickhouse_cluster_v2.foo"),
					resource.TestCheckResourceAttr("yandex_mdb_clickhouse_cluster_v2.foo", "name", chName),
					resource.TestCheckResourceAttr("data.yandex_mdb_clickhouse_cluster_v2.bar", "name", chName),
				),
			},
		},
	})
}

func TestAccDataSourceMDBClickHouseClusterV2_byName(t *testing.T) {
	t.Parallel()

	chName := acctest.RandomWithPrefix("ds-ch-by-name")
	chDesc := "ClickHouseCluster Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBClickHouseClusterConfig(chName, chDesc, false),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceMDBClickHouseClusterAttributesCheck("data.yandex_mdb_clickhouse_cluster_v2.bar", "yandex_mdb_clickhouse_cluster_v2.foo"),
					resource.TestCheckResourceAttr("yandex_mdb_clickhouse_cluster_v2.foo", "name", chName),
					resource.TestCheckResourceAttr("data.yandex_mdb_clickhouse_cluster_v2.bar", "name", chName),
				),
			},
		},
	})
}

func TestAccDataSourceMDBClickHouseClusterV2_diskEncryption(t *testing.T) {
	t.Parallel()

	chName := acctest.RandomWithPrefix("ds-ch-disk-encryption")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             resource.ComposeTestCheckFunc(testAccCheckMDBClickHouseClusterDestroy, kms_symmetric_key.TestAccCheckYandexKmsSymmetricKeyAllDestroyed),
		Steps: []resource.TestStep{
			{
				Config: testAccMDBClickHouseCluster_encrypted_disk(chName) + mdbClickHouseClusterByIDConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.yandex_mdb_clickhouse_cluster_v2.bar", "disk_encryption_key_id"),
				),
			},
		},
	})
}

func testAccDataSourceMDBClickHouseClusterAttributesCheck(
	datasourceName string,
	resourceName string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		if ds.Primary.ID != rs.Primary.ID {
			return fmt.Errorf("datasource ID does not match resource ID: %s and %s", ds.Primary.ID, rs.Primary.ID)
		}

		dsAttrs := ds.Primary.Attributes
		rsAttrs := rs.Primary.Attributes

		// resource -> datasource
		for k, rv := range rsAttrs {
			if shouldIgnoreAttrKey(k) {
				continue
			}

			dv, ok := dsAttrs[k]
			if !ok {
				return fmt.Errorf("attribute %q is present in resource state but missing in datasource state", k)
			}
			if dv != rv {
				return fmt.Errorf("attribute %q mismatch: datasource=%q resource=%q", k, dv, rv)
			}
		}

		// datasource -> resource
		for k, dv := range dsAttrs {
			if shouldIgnoreAttrKey(k) {
				continue
			}

			rv, ok := rsAttrs[k]
			if !ok {
				return fmt.Errorf("attribute %q is present in datasource state but missing in resource state", k)
			}
			if dv != rv {
				return fmt.Errorf("attribute %q mismatch: datasource=%q resource=%q", k, dv, rv)
			}
		}

		return nil
	}
}

func shouldIgnoreAttrKey(k string) bool {
	if _, ok := ignoreAttrsSet[k]; ok {
		return true
	}
	for prefix := range ignoreByPrefixAttrsSet {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

const mdbClickHouseClusterByIDConfig = `
data "yandex_mdb_clickhouse_cluster_v2" "bar" {
  cluster_id = "${yandex_mdb_clickhouse_cluster_v2.foo.id}"
}
`

const mdbClickHouseClusterByNameConfig = `
data "yandex_mdb_clickhouse_cluster_v2" "bar" {
  name = "${yandex_mdb_clickhouse_cluster_v2.foo.name}"
}
`

func testAccDataSourceMDBClickHouseClusterConfig(name, desc string, useDataID bool) string {
	resourceHCL := fmt.Sprintf(clickHouseVPCDependencies+"\n"+`
resource "yandex_mdb_clickhouse_cluster_v2" "foo" {
  name           	  = "%s"
  description    	  = "%s"
  environment 		  = "PRESTABLE"
  network_id     	  = "${yandex_vpc_network.mdb-ch-test-net.id}"
  admin_password      = "strong_password"
  deletion_protection = false

  labels = {
    test_key = "test_value"
  }

  version    = "%s"
  clickhouse = {
	  resources = {
		resource_preset_id = "s2.micro"
		disk_type_id       = "network-ssd"
		disk_size          = 10
	  }
  }

  hosts = {
    "ha" = {
	  type      = "CLICKHOUSE"
	  zone      = "ru-central1-a"
	  subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
    }
  }

  security_group_ids = ["${yandex_vpc_security_group.mdb-ch-test-sg-x.id}"]

  maintenance_window {
  	type = "ANYTIME"
  }
}
`,
		name,
		desc,
		chVersion,
	)

	if useDataID {
		return resourceHCL + mdbClickHouseClusterByIDConfig
	}
	return resourceHCL + mdbClickHouseClusterByNameConfig
}
