package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	k8s "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
)

//revive:disable:var-naming
func TestAccDataSourceKubernetesNodeGroup_basic(t *testing.T) {
	t.Parallel()

	clusterResource := clusterInfo("testAccDataSourceKubernetesNodeGroupConfig_basic", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
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
					checkNodeGroupAttributes(&ng, &nodeResource, false),
					testAccCheckResourceIDField(nodeResourceFullName, "node_group_id"),
					testAccCheckCreatedAtAttr(nodeResourceFullName),
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
