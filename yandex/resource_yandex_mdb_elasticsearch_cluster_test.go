package yandex

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/elasticsearch/v1"
)

const elasticsearchResource = "yandex_mdb_elasticsearch_cluster.foo"

func init() {
	resource.AddTestSweepers("yandex_mdb_elasticsearch_cluster", &resource.Sweeper{
		Name: "yandex_mdb_elasticsearch_cluster",
		F:    testSweepMDBElasticsearchCluster,
	})
}

func testSweepMDBElasticsearchCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.MDB().ElasticSearch().Cluster().List(conf.Context(), &elasticsearch.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Elasticsearch clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBElasticsearchCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Elasticsearch cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBElasticsearchCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBElasticsearchClusterOnce, conf, "Elasticsearch cluster", id)
}

func sweepMDBElasticsearchClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexMDBElasticsearchClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.sdk.MDB().ElasticSearch().Cluster().Update(ctx, &elasticsearch.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.sdk.MDB().ElasticSearch().Cluster().Delete(ctx, &elasticsearch.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func mdbElasticsearchClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"health",                  // volatile value
			"config.0.admin_password", // not importable
			"host",                    // host name not importable
		},
	}

}

func TestAccMDBElasticsearchCluster_basic(t *testing.T) {
	t.Parallel()

	var r elasticsearch.Cluster
	elasticsearchName := acctest.RandomWithPrefix("tf-elasticsearch")
	elasticsearchDesc := "Elasticsearch Cluster Terraform Test"
	randInt := acctest.RandInt()
	folderID := getExampleFolderID()
	elasticsearchDesc2 := "Elasticsearch Cluster Terraform Test Updated"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBElasticsearchClusterDestroy,
		Steps: []resource.TestStep{
			// Create Elasticsearch Cluster
			{
				Config: testAccMDBElasticsearchClusterConfig(elasticsearchName, elasticsearchDesc, "PRESTABLE", true, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBElasticsearchClusterExists(elasticsearchResource, &r, 5),
					resource.TestCheckResourceAttr(elasticsearchResource, "name", elasticsearchName),
					resource.TestCheckResourceAttr(elasticsearchResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(elasticsearchResource, "description", elasticsearchDesc),
					resource.TestCheckResourceAttr(elasticsearchResource, "config.0.admin_password", "password"),
					resource.TestCheckResourceAttrSet(elasticsearchResource, "service_account_id"),
					resource.TestCheckResourceAttr(elasticsearchResource, "deletion_protection", "true"),
					// resource.TestCheckResourceAttrSet(elasticsearchResource, "host.0.fqdn"),
					testAccCheckCreatedAtAttr(elasticsearchResource),
					testAccCheckMDBElasticsearchClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckMDBElasticsearchClusterDataNodeHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckMDBElasticsearchClusterMasterNodeHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckMDBElasticsearchClusterHasPlugins(&r, "analysis-icu", "repository-s3"),
					func(s *terraform.State) error {
						time.Sleep(2 * time.Minute)
						return nil
					},
					resource.TestCheckResourceAttr(elasticsearchResource, "maintenance_window.0.type", "WEEKLY"),
					resource.TestCheckResourceAttr(elasticsearchResource, "maintenance_window.0.day", "FRI"),
					resource.TestCheckResourceAttr(elasticsearchResource, "maintenance_window.0.hour", "20"),
				),
			},
			mdbElasticsearchClusterImportStep(elasticsearchResource),
			// uncheck 'deletion_protection'
			{
				Config: testAccMDBElasticsearchClusterConfig(elasticsearchName, elasticsearchDesc, "PRESTABLE", false, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBElasticsearchClusterExists(elasticsearchResource, &r, 5),
					resource.TestCheckResourceAttr(elasticsearchResource, "deletion_protection", "false"),
				),
			},
			mdbElasticsearchClusterImportStep(elasticsearchResource),
			// check 'deletion_protection'
			{
				Config: testAccMDBElasticsearchClusterConfig(elasticsearchName, elasticsearchDesc, "PRESTABLE", true, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBElasticsearchClusterExists(elasticsearchResource, &r, 5),
					resource.TestCheckResourceAttr(elasticsearchResource, "deletion_protection", "true"),
				),
			},
			mdbElasticsearchClusterImportStep(elasticsearchResource),
			// test 'deletion_protection
			{
				Config:      testAccMDBElasticsearchClusterConfig(elasticsearchName, elasticsearchDesc, "PRODUCTION", true, randInt),
				ExpectError: regexp.MustCompile(".*The operation was rejected because cluster has 'deletion_protection' = ON.*"),
			},
			// uncheck 'deletion_protection'
			{
				Config: testAccMDBElasticsearchClusterConfig(elasticsearchName, elasticsearchDesc, "PRESTABLE", false, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBElasticsearchClusterExists(elasticsearchResource, &r, 5),
					resource.TestCheckResourceAttr(elasticsearchResource, "deletion_protection", "false"),
				),
			},
			mdbElasticsearchClusterImportStep(elasticsearchResource),
			// Update Elasticsearch Cluster
			{
				Config: testAccMDBElasticsearchClusterConfigUpdated(elasticsearchName, elasticsearchDesc2, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBElasticsearchClusterExists(elasticsearchResource, &r, 6),
					resource.TestCheckResourceAttr(elasticsearchResource, "name", elasticsearchName),
					resource.TestCheckResourceAttr(elasticsearchResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(elasticsearchResource, "description", elasticsearchDesc2),
					resource.TestCheckResourceAttr(elasticsearchResource, "service_account_id", ""),
					testAccCheckCreatedAtAttr(elasticsearchResource),
					testAccCheckMDBElasticsearchClusterContainsLabel(&r, "test_key2", "test_value2"),
					testAccCheckMDBElasticsearchClusterDataNodeHasResources(&r, "m2.small", "network-ssd", 11*1024*1024*1024),
					testAccCheckMDBElasticsearchClusterMasterNodeHasResources(&r, "m2.micro", "network-ssd", 11*1024*1024*1024),
					testAccCheckMDBElasticsearchClusterHasPlugins(&r, "repository-s3"),
					func(s *terraform.State) error {
						time.Sleep(time.Minute)
						return nil
					},
					resource.TestCheckResourceAttr(elasticsearchResource, "maintenance_window.0.type", "ANYTIME"),
				),
			},
			mdbElasticsearchClusterImportStep(elasticsearchResource),
		},
	})
}

func testAccCheckMDBElasticsearchClusterExists(n string, r *elasticsearch.Cluster, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.MDB().ElasticSearch().Cluster().Get(context.Background(), &elasticsearch.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Elasticsearch Cluster not found")
		}

		*r = *found

		resp, err := config.sdk.MDB().ElasticSearch().Cluster().ListHosts(context.Background(), &elasticsearch.ListClusterHostsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Hosts) != hosts {
			return fmt.Errorf("Expected %d hosts, got %d", hosts, len(resp.Hosts))
		}

		return nil
	}
}

func testAccCheckMDBElasticsearchClusterDataNodeHasResources(r *elasticsearch.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.Elasticsearch.DataNode.Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskTypeId != diskType {
			return fmt.Errorf("expected disk type '%s', got '%s'", diskType, rs.DiskTypeId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("expected disk size '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

func testAccCheckMDBElasticsearchClusterMasterNodeHasResources(r *elasticsearch.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.Elasticsearch.MasterNode.Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskTypeId != diskType {
			return fmt.Errorf("expected disk type '%s', got '%s'", diskType, rs.DiskTypeId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("expected disk size '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

func testAccCheckMDBElasticsearchClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_elasticsearch_cluster" {
			continue
		}

		_, err := config.sdk.MDB().ElasticSearch().Cluster().Get(context.Background(), &elasticsearch.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Elasticsearch Cluster still exists")
		}
	}

	return nil
}

func testAccCheckMDBElasticsearchClusterContainsLabel(r *elasticsearch.Cluster, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := r.Labels[key]
		if !ok {
			return fmt.Errorf("expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccCheckMDBElasticsearchClusterHasPlugins(r *elasticsearch.Cluster, plugins ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		p := r.Config.Elasticsearch.Plugins
		sort.Strings(p)
		sort.Strings(plugins)
		if !reflect.DeepEqual(p, plugins) {
			return fmt.Errorf("incorrect cluster plugins: expected '%s' but found '%s'", plugins, p)
		}
		return nil
	}
}

func testAccMDBElasticsearchClusterConfig(name, desc, environment string, deletionProtection bool, randInt int) string {
	return testAccCommonIamDependenciesEditorConfig(randInt) + fmt.Sprintf("\n"+elasticsearchVPCDependencies+`

locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-c",
  ]
}

resource "yandex_mdb_elasticsearch_cluster" "foo" {
  name        = "%s"
  description = "%s"
  labels = {
    test_key  = "test_value"
  }
  environment = "%s"
  network_id  = "${yandex_vpc_network.mdb-elasticsearch-test-net.id}"
  security_group_ids = [yandex_vpc_security_group.mdb-elasticsearch-test-sg-x.id]
  service_account_id = "${yandex_iam_service_account.sa.id}"
  deletion_protection = %t

  config {

    admin_password = "password"

    data_node {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 10
      }
    }

    master_node {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 10
      }
    }

    plugins = ["analysis-icu", "repository-s3"]
  }

  dynamic "host" {
    for_each = toset(range(0,2))
    content {
      name = "datanode${host.value}"
      zone = local.zones[(host.value)%%3]
      type = "DATA_NODE"
      assign_public_ip = true
    }
  }

  dynamic "host" {
    for_each = toset(range(0,3))
    content {
      name = "masternode${host.value}"
      zone = local.zones[host.value%%3]
      type = "MASTER_NODE"
    }
  }

  depends_on = [
    yandex_vpc_subnet.mdb-elasticsearch-test-subnet-a,
    yandex_vpc_subnet.mdb-elasticsearch-test-subnet-b,
    yandex_vpc_subnet.mdb-elasticsearch-test-subnet-c,
  ]

  maintenance_window {
    type = "WEEKLY"
    day  = "FRI"
    hour = 20
  }
}
`, name, desc, environment, deletionProtection)
}

func testAccMDBElasticsearchClusterConfigUpdated(name, desc string, randInt int) string {
	return testAccCommonIamDependenciesEditorConfig(randInt) + fmt.Sprintf("\n"+elasticsearchVPCDependencies+`

locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-c",
  ]
}

resource "yandex_mdb_elasticsearch_cluster" "foo" {
  name        = "%s"
  description = "%s"
  labels = {
    test_key2  = "test_value2"
  }
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-elasticsearch-test-net.id}"
  security_group_ids = [yandex_vpc_security_group.mdb-elasticsearch-test-sg-x.id, yandex_vpc_security_group.mdb-elasticsearch-test-sg-y.id]
  service_account_id = ""

  config {

    admin_password = "password_updated"

    data_node {
      resources {
        resource_preset_id = "m2.small"
        disk_type_id       = "network-ssd"
        disk_size          = 11
      }
    }

    master_node {
      resources {
        resource_preset_id = "m2.micro"
        disk_type_id       = "network-ssd"
        disk_size          = 11
      }
    }

    plugins = ["repository-s3"]

  }

  dynamic "host" {
    for_each = toset(range(0,3))
    content {
      name = "datanode${host.value}"
      zone = local.zones[(host.value)%%3]
      type = "DATA_NODE"
      assign_public_ip = true
    }
  }

  dynamic "host" {
    for_each = toset(range(0,3))
    content {
      name = "masternode${host.value}"
      zone = local.zones[host.value%%3]
      type = "MASTER_NODE"
    }
  }

  depends_on = [
    yandex_vpc_subnet.mdb-elasticsearch-test-subnet-a,
    yandex_vpc_subnet.mdb-elasticsearch-test-subnet-b,
    yandex_vpc_subnet.mdb-elasticsearch-test-subnet-c,
  ]

  maintenance_window {
    type = "ANYTIME"
  }
}
`, name, desc)
}

const elasticsearchVPCDependencies = `
resource "yandex_vpc_network" "mdb-elasticsearch-test-net" {}

resource "yandex_vpc_security_group" "mdb-elasticsearch-test-sg-x" {
  network_id     = "${yandex_vpc_network.mdb-elasticsearch-test-net.id}"
  ingress {
    protocol          = "ANY"
    description       = "Allow incoming traffic from members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
  egress {
    protocol          = "ANY"
    description       = "Allow outgoing traffic to members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
}

resource "yandex_vpc_security_group" "mdb-elasticsearch-test-sg-y" {
  network_id     = "${yandex_vpc_network.mdb-elasticsearch-test-net.id}"
  
  ingress {
    protocol          = "ANY"
    description       = "Allow incoming traffic from members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
  egress {
    protocol          = "ANY"
    description       = "Allow outgoing traffic to members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
}

resource "yandex_vpc_subnet" "mdb-elasticsearch-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.mdb-elasticsearch-test-net.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-elasticsearch-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.mdb-elasticsearch-test-net.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-elasticsearch-test-subnet-c" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.mdb-elasticsearch-test-net.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}
`
