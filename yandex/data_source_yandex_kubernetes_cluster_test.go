package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	k8s "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
)

//revive:disable:var-naming
func TestAccDataSourceKubernetesClusterZonal_basic(t *testing.T) {
	clusterResource := clusterInfoWithNetworkAndMaintenancePolicies("testAccDataSourceKubernetesClusterZonalConfig_basic",
		true, true, dailyMaintenancePolicy)
	clusterResourceFullName := clusterResource.ResourceFullName(true)
	clusterDataSourceFullName := clusterResource.ResourceFullName(false)

	var cluster k8s.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesClusterZonalConfig_basic(clusterResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					testAccCheckResourceIDField(clusterDataSourceFullName, "cluster_id"),
					checkClusterAttributes(&cluster, &clusterResource, false),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesClusterRegional_basic(t *testing.T) {
	clusterResource := clusterInfoWithNetworkAndMaintenancePolicies("testAccDataSourceKubernetesClusterRegionalConfig_basic", false,
		false, weeklyMaintenancePolicy)
	clusterResourceFullName := clusterResource.ResourceFullName(true)
	clusterDataSourceFullName := clusterResource.ResourceFullName(false)

	var cluster k8s.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesClusterRegionalConfig_basic(clusterResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					testAccCheckResourceIDField(clusterDataSourceFullName, "cluster_id"),
					checkClusterAttributes(&cluster, &clusterResource, false),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
		},
	})
}

const dataClusterConfigTemplate = `
data "yandex_kubernetes_cluster" "{{.ClusterResourceName}}" {
  name = "${yandex_kubernetes_cluster.{{.ClusterResourceName}}.name}" 
}
`

func testAccDataSourceKubernetesClusterZonalConfig_basic(in resourceClusterInfo) string {
	resourceConfig := testAccKubernetesClusterZonalConfig_basic(in)
	resourceConfig += templateConfig(dataClusterConfigTemplate, in.Map())
	return resourceConfig
}

func testAccDataSourceKubernetesClusterRegionalConfig_basic(in resourceClusterInfo) string {
	resourceConfig := testAccKubernetesClusterRegionalConfig_basic(in)
	resourceConfig += templateConfig(dataClusterConfigTemplate, in.Map())
	return resourceConfig
}
