package acceptance

import (
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	k8s "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

//revive:disable:var-naming
func TestAccDataSourceKubernetesNodeGroupDailyMaintenance_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroupDailyMaintenance_basic", true)
	nodeResource := nodeGroupInfoWithMaintenance(clusterResource.ClusterResourceName, true, false, dailyMaintenancePolicy)
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, false),
					statecheck.CheckResourceIDField(nodeResourceFullName, "node_group_id"),
					statecheck.CheckCreatedAtAttr(nodeResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesNodeGroupWeeklyMaintenance_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroupWeeklyMaintenance_basic", true)
	nodeResource := nodeGroupInfoWithMaintenance(clusterResource.ClusterResourceName, false, true, weeklyMaintenancePolicy)
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, false),
					statecheck.CheckResourceIDField(nodeResourceFullName, "node_group_id"),
					statecheck.CheckCreatedAtAttr(nodeResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesNodeGroupNetworkInterfaces_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfoWithSecurityGroups("TestAccDataSourceKubernetesNodeGroupNetworkInterfaces_basic", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResource.constructNetworkInterfaces(clusterResource.SubnetResourceNameA, clusterResource.SecurityGroupName)
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, false),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesNodeGroup_autoscaled(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroup_autoscaled", true)

	nodeResource := nodeGroupInfoAutoscaled(clusterResource.ClusterResourceName)
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, true),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesNodeGroupDailyMaintenance_placementGroup(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroupDailyMaintenance_basic", true)
	nodeResource := nodeGroupInfoWithMaintenance(clusterResource.ClusterResourceName, true, false, dailyMaintenancePolicy)
	nodeResource.PlacementGroupId = "yandex_compute_placement_group.pg.id"
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource) + constPlacementGroupResource,
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, false),
					statecheck.CheckResourceIDField(nodeResourceFullName, "node_group_id"),
					statecheck.CheckCreatedAtAttr(nodeResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesNodeGroup_dualStack(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfoDualStack("TestAccDataSourceKubernetesNodeGroup_dualStack", true)
	nodeResource := nodeGroupInfoDualStack(clusterResource.ClusterResourceName)

	nodeResource.constructNetworkInterfaces(clusterResource.SubnetResourceNameA, clusterResource.SecurityGroupName)
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	// All dual stack test share the same subnet. Disallow concurrent execution.
	sdkcommon.MutexKV.Lock(clusterResource.SubnetResourceNameA)
	defer sdkcommon.MutexKV.Unlock(clusterResource.SubnetResourceNameA)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, false),
					statecheck.CheckResourceIDField(nodeResourceFullName, "node_group_id"),
					statecheck.CheckCreatedAtAttr(nodeResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesNodeGroup_networkSettingsSoftwareAccelerated(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroup_networkSettingsSoftwareAccelerated", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResourceFullName := nodeResource.ResourceFullName(false)
	nodeResource.NetworkAccelerationType = "software_accelerated"

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, false),
					statecheck.CheckResourceIDField(nodeResourceFullName, "node_group_id"),
					statecheck.CheckCreatedAtAttr(nodeResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesNodeGroup_containerRuntimeContainerd(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroup_containerRuntimeContainerd", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, false),
					statecheck.CheckResourceIDField(nodeResourceFullName, "node_group_id"),
					statecheck.CheckCreatedAtAttr(nodeResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesNodeGroupIPv4DNSFQDN_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfoWithSecurityGroups("TestAccDataSourceKubernetesNodeGroupIPv4DNSFQDN_basic", true)
	nodeResource := nodeGroupInfoIPv4DNSFQDN(clusterResource.ClusterResourceName)
	nodeResource.constructNetworkInterfaces(clusterResource.SubnetResourceNameA, clusterResource.SecurityGroupName)
	nodeResourceFullName := nodeResource.ResourceFullName(false)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, false, false),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesNodeGroupContainerNetwork_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfo("TestAccDataSourceKubernetesNodeGroupContainerNetwork_basic", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResourceFullName := nodeResource.ResourceFullName(true)
	nodeResource.PodMTU = 8910

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
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
	resourceConfig += sdkcommon.TemplateConfig(dataNodeGroupConfigTemplate, ng.Map())
	return resourceConfig
}
