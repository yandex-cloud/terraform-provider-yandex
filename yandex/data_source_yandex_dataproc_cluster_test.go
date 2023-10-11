package yandex

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceDataprocCluster_byName(t *testing.T) {
	templateParams := defaultDataprocConfigParams(t)
	config := testAccDataprocClusterConfig(t, templateParams) + `
		data "yandex_dataproc_cluster" "tf-dataproc-cluster" {
		  name = "${yandex_dataproc_cluster.tf-dataproc-cluster.name}"
		}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataprocClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: testAccDataSourceDataprocClusterCheck(
					"data.yandex_dataproc_cluster.tf-dataproc-cluster",
					"yandex_dataproc_cluster.tf-dataproc-cluster",
					templateParams.Name),
			},
		},
	})
}

func TestAccDataSourceDataprocCluster_byId(t *testing.T) {
	templateParams := defaultDataprocConfigParams(t)
	config := testAccDataprocClusterConfig(t, templateParams) + `
		data "yandex_dataproc_cluster" "tf-dataproc-cluster" {
		  cluster_id = "${yandex_dataproc_cluster.tf-dataproc-cluster.id}"
		}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDataprocClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: testAccDataSourceDataprocClusterCheck(
					"data.yandex_dataproc_cluster.tf-dataproc-cluster",
					"yandex_dataproc_cluster.tf-dataproc-cluster",
					templateParams.Name),
			},
		},
	})
}

func testAccDataSourceDataprocClusterCheck(datasourceName, resourceName, clusterName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceDataprocClusterAttributesCheck(datasourceName, resourceName),
		testAccCheckCreatedAtAttr(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "cluster_config.0.subcluster_spec.#", "2"),
		resource.TestCheckResourceAttr(datasourceName, "cluster_config.0.version_id", "2.0"),
		resource.TestCheckResourceAttr(datasourceName, "description",
			"Dataproc Cluster created by Terraform"),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", getExampleFolderID()),
		resource.TestCheckResourceAttr(datasourceName, "labels.created_by", "terraform"),
		resource.TestCheckResourceAttr(datasourceName, "name", clusterName),
		resource.TestCheckResourceAttr(datasourceName, "zone_id", testDataprocZone),
	)
}

func testAccDataSourceDataprocClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
			return fmt.Errorf("instance `data source` ID does not match `resource` ID: %s and %s", ds.Primary.ID, rs.Primary.ID)
		}

		dataSourceAttributes := prepareDataprocAttributesForCompare(ds.Primary.Attributes)
		resourceAttributes := prepareDataprocAttributesForCompare(rs.Primary.Attributes)

		for key, resourceValue := range resourceAttributes {
			dataSourceValue, ok := dataSourceAttributes[key]
			if !ok {
				return fmt.Errorf("data source %q doesn't have attribute %q", datasourceName, key)
			}
			if dataSourceValue != resourceValue {
				return fmt.Errorf("value mismatch for attribute %q, data source value is %q, "+
					"resource value is %q", key, dataSourceValue, resourceValue)
			}
		}

		return nil
	}
}

// substitute subcluster's index with its name because index is not stable
func prepareDataprocAttributesForCompare(attributes map[string]string) map[string]string {
	subclusterNameByIndex := make(map[string]string)
	re := regexp.MustCompile(`^cluster_config\.0\.subcluster_spec\.(\d+)\.name$`)
	for key, value := range attributes {
		if re.MatchString(key) {
			matches := re.FindSubmatch([]byte(key))
			index := string(matches[1])
			subclusterNameByIndex[index] = value
		}
	}

	re = regexp.MustCompile(`^cluster_config\.0\.subcluster_spec\.(\d+)\.(.*)$`)
	fixedAttributes := make(map[string]string)
	for key, value := range attributes {
		fixedKey := key
		if re.MatchString(key) {
			matches := re.FindSubmatch([]byte(key))
			name := subclusterNameByIndex[string(matches[1])]
			fixedKey = fmt.Sprintf("cluster_config.0.subcluster_spec.%s.%s", name, string(matches[2]))
		}
		fixedAttributes[fixedKey] = value
	}
	return fixedAttributes
}
