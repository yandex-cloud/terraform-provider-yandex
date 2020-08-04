package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	k8s "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
)

//revive:disable:var-naming
func TestAccDataSourceKubernetesNodeGroupDailyMaintenance_basic(t *testing.T) {
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroupDailyMaintenance_basic", true)
	nodeResource := nodeGroupInfoWithMaintenance(clusterResource.ClusterResourceName, true, false, dailyMaintenancePolicy)
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, false),
					testAccCheckResourceIDField(nodeResourceFullName, "node_group_id"),
					testAccCheckCreatedAtAttr(nodeResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesNodeGroupWeeklyMaintenance_basic(t *testing.T) {
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroupWeeklyMaintenance_basic", true)
	nodeResource := nodeGroupInfoWithMaintenance(clusterResource.ClusterResourceName, false, true, weeklyMaintenancePolicy)
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, false),
					testAccCheckResourceIDField(nodeResourceFullName, "node_group_id"),
					testAccCheckCreatedAtAttr(nodeResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesNodeGroup_autoscaled(t *testing.T) {
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroup_autoscaled", true)

	nodeResource := nodeGroupInfoAutoscaled(clusterResource.ClusterResourceName)
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, true),
				),
			},
		},
	})
}

const dataNodeGroupConfigTemplate = `
data "yandex_kubernetes_node_group" "{{.NodeGroupResourceName}}" {
  name = "${yandex_kubernetes_node_group.{{.NodeGroupResourceName}}.name}"
}
`

func testAccDataSourceKubernetesNodeGroupConfig_basic(cluster resourceClusterInfo, ng resourceNodeGroupInfo) string {
	resourceConfig := testAccKubernetesNodeGroupConfig_basic(cluster, ng)
	resourceConfig += templateConfig(dataNodeGroupConfigTemplate, ng.Map())
	return resourceConfig
}
