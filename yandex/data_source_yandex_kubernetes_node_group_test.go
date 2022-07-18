package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

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

func TestAccDataSourceKubernetesNodeGroupNetworkInterfaces_basic(t *testing.T) {
	clusterResource := clusterInfoWithSecurityGroups("TestAccDataSourceKubernetesNodeGroupNetworkInterfaces_basic", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResource.constructNetworkInterfaces(clusterResource.SubnetResourceNameA, clusterResource.SecurityGroupName)
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

func TestAccDataSourceKubernetesNodeGroupDailyMaintenance_placementGroup(t *testing.T) {
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroupDailyMaintenance_basic", true)
	nodeResource := nodeGroupInfoWithMaintenance(clusterResource.ClusterResourceName, true, false, dailyMaintenancePolicy)
	nodeResource.PlacementGroupId = "yandex_compute_placement_group.pg.id"
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource) + constPlacementGroupResource,
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

func TestAccDataSourceKubernetesNodeGroup_dualStack(t *testing.T) {
	clusterResource := clusterInfoDualStack("TestAccDataSourceKubernetesNodeGroup_dualStack", true)
	nodeResource := nodeGroupInfoDualStack(clusterResource.ClusterResourceName)

	nodeResource.constructNetworkInterfaces(clusterResource.SubnetResourceNameA, clusterResource.SecurityGroupName)
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	// All dual stack test share the same subnet. Disallow concurrent execution.
	mutexKV.Lock(clusterResource.SubnetResourceNameA)
	defer mutexKV.Unlock(clusterResource.SubnetResourceNameA)

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

func TestAccDataSourceKubernetesNodeGroup_networkSettingsSoftwareAccelerated(t *testing.T) {
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroup_networkSettingsSoftwareAccelerated", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResourceFullName := nodeResource.ResourceFullName(false)
	nodeResource.NetworkAccelerationType = "software_accelerated"

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

func TestAccDataSourceKubernetesNodeGroup_containerRuntimeContainerd(t *testing.T) {
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroup_containerRuntimeContainerd", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResourceFullName := nodeResource.ResourceFullName(false)
	nodeResource.ContainerRuntimeType = "containerd"

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

func TestAccDataSourceKubernetesNodeGroupIPv4DNSFQDN_basic(t *testing.T) {
	clusterResource := clusterInfoWithSecurityGroups("TestAccDataSourceKubernetesNodeGroupIPv4DNSFQDN_basic", true)
	nodeResource := nodeGroupInfoIPv4DNSFQDN(clusterResource.ClusterResourceName)
	nodeResource.constructNetworkInterfaces(clusterResource.SubnetResourceNameA, clusterResource.SecurityGroupName)
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
