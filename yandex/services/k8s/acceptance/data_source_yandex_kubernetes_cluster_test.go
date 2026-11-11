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
func TestAccDataSourceKubernetesClusterZonal_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfoWithSecurityGroupsNetworkAndMaintenancePolicies("testAccDataSourceKubernetesClusterZonalConfig_basic",
		true, true, dailyMaintenancePolicy)
	clusterResourceFullName := clusterResource.ResourceFullName(true)
	clusterDataSourceFullName := clusterResource.ResourceFullName(false)

	var cluster k8s.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesClusterZonalConfig_basic(clusterResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					statecheck.CheckResourceIDField(clusterDataSourceFullName, "cluster_id"),
					checkClusterAttributes(&cluster, &clusterResource, false),
					statecheck.CheckCreatedAtAttr(clusterResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesClusterRegional_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfoWithSecurityGroupsNetworkAndMaintenancePolicies("testAccDataSourceKubernetesClusterRegionalConfig_basic", false,
		false, weeklyMaintenancePolicy)
	clusterResourceFullName := clusterResource.ResourceFullName(true)
	clusterDataSourceFullName := clusterResource.ResourceFullName(false)

	var cluster k8s.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesClusterRegionalConfig_basic(clusterResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					statecheck.CheckResourceIDField(clusterDataSourceFullName, "cluster_id"),
					checkClusterAttributes(&cluster, &clusterResource, false),
					statecheck.CheckCreatedAtAttr(clusterResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesClusterRegional_externalIPv6Address(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfoExternalIPv6Address("testAccDataSourceKubernetesClusterRegional_externalIPv6Address")
	clusterResourceFullName := clusterResource.ResourceFullName(true)
	clusterDataSourceFullName := clusterResource.ResourceFullName(false)

	var cluster k8s.Cluster

	// All external IPv6 tests share the same subnets. Disallow concurrent execution.
	sdkcommon.MutexKV.Lock(clusterResource.SubnetResourceNameA)
	sdkcommon.MutexKV.Lock(clusterResource.SubnetResourceNameB)
	sdkcommon.MutexKV.Lock(clusterResource.SubnetResourceNameD)
	defer sdkcommon.MutexKV.Unlock(clusterResource.SubnetResourceNameA)
	defer sdkcommon.MutexKV.Unlock(clusterResource.SubnetResourceNameB)
	defer sdkcommon.MutexKV.Unlock(clusterResource.SubnetResourceNameD)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesClusterRegionalConfig_basic(clusterResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					statecheck.CheckResourceIDField(clusterDataSourceFullName, "cluster_id"),
					checkClusterAttributes(&cluster, &clusterResource, false),
					statecheck.CheckCreatedAtAttr(clusterResourceFullName),
				),
			},
		},
	})
}

func TestAccDataSourceKubernetesClusterZonal_networkImplementationCilium(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	clusterResource := clusterInfo("TestAccDataSourceKubernetesClusterZonal_networkImplementationCilium", true)

	clusterResourceFullName := clusterResource.ResourceFullName(true)
	clusterDataSourceFullName := clusterResource.ResourceFullName(false)
	clusterResource.NetworkImplementationCilium = true
	clusterResource.MasterVersion = k8sTestVersion

	var cluster k8s.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKubernetesClusterZonalConfig_basic(clusterResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					statecheck.CheckResourceIDField(clusterDataSourceFullName, "cluster_id"),
					checkClusterAttributes(&cluster, &clusterResource, false),
					statecheck.CheckCreatedAtAttr(clusterResourceFullName),
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
	resourceConfig += sdkcommon.TemplateConfig(dataClusterConfigTemplate, in.Map())
	return resourceConfig
}

func testAccDataSourceKubernetesClusterRegionalConfig_basic(in resourceClusterInfo) string {
	resourceConfig := testAccKubernetesClusterRegionalConfig_basic(in)
	resourceConfig += sdkcommon.TemplateConfig(dataClusterConfigTemplate, in.Map())
	return resourceConfig
}
