package yandex

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/fatih/structs"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
)

func init() {
	resource.AddTestSweepers("yandex_kubernetes_node_group", &resource.Sweeper{
		Name: "yandex_kubernetes_node_group",
		F:    testSweepKubernetesNodeGroups,
	})
}

func testSweepKubernetesNodeGroups(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &k8s.ListNodeGroupsRequest{FolderId: conf.FolderID}
	it := conf.sdk.Kubernetes().NodeGroup().NodeGroupIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepKubernetesNodeGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Kubernetes Node Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepKubernetesNodeGroup(conf *Config, id string) bool {
	return sweepWithRetry(sweepKubernetesNodeGroupOnce, conf, "Kubernetes Node Group", id)
}

func sweepKubernetesNodeGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexKubernetesNodeGroupDeleteTimeout)
	defer cancel()

	op, err := conf.sdk.Kubernetes().NodeGroup().Delete(ctx, &k8s.DeleteNodeGroupRequest{
		NodeGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func k8sNodeGroupImportStep(nodeResourceFullName string, ignored ...string) resource.TestStep {
	return resource.TestStep{
		ResourceName:            nodeResourceFullName,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: ignored,
	}
}

//revive:disable:var-naming
func TestAccKubernetesNodeGroup_basic(t *testing.T) {
	clusterResource := clusterInfo("testAccKubernetesNodeGroupConfig_basic", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResourceFullName := nodeResource.ResourceFullName(true)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, true, false),
				),
			},
			k8sNodeGroupImportStep(nodeResourceFullName),
		},
	})
}

func TestAccKubernetesNodeGroupDailyMaintenance_basic(t *testing.T) {
	clusterResource := clusterInfo("TestAccKubernetesNodeGroupDailyMaintenance_basic", true)
	nodeResource := nodeGroupInfoWithMaintenance(clusterResource.ClusterResourceName, true, true, dailyMaintenancePolicy)
	nodeResourceFullName := nodeResource.ResourceFullName(true)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, true, false),
				),
			},
			k8sNodeGroupImportStep(nodeResourceFullName),
		},
	})
}

func TestAccKubernetesNodeGroup_zero_cores(t *testing.T) {
	t.Parallel()

	clusterResource := clusterInfo("testAccKubernetesNodeGroupConfig_basic", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResource.Cores = "0"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				ExpectError: regexp.MustCompile("expected instance_template.0.resources.0.cores to be greater than"),
			},
		},
	})
}

func TestAccKubernetesNodeGroup_zero_memory(t *testing.T) {
	t.Parallel()

	clusterResource := clusterInfo("testAccKubernetesNodeGroupConfig_basic", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResource.Memory = "0"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				ExpectError: regexp.MustCompile("expected instance_template.0.resources.0.memory to be greater than"),
			},
		},
	})
}

func TestAccKubernetesNodeGroupNetworkInterfaces_basic(t *testing.T) {
	clusterResource := clusterInfoWithSecurityGroups("TestAccKubernetesNodeGroupNetworkInterfaces_basic", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResource.constructNetworkInterfaces(clusterResource.SubnetResourceNameA, clusterResource.SecurityGroupName)
	nodeResourceFullName := nodeResource.ResourceFullName(true)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, true, false),
				),
			},
			k8sNodeGroupImportStep(nodeResourceFullName),
		},
	})
}

func TestAccKubernetesNodeGroup_update(t *testing.T) {
	clusterResource := clusterInfo("TestAccKubernetesNodeGroup_update", true)
	clusterResource.MasterVersion = k8sTestUpdateVersion
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResource.Version = k8sTestVersion
	nodeResourceFullName := nodeResource.ResourceFullName(true)

	nodeUpdatedResource := nodeResource

	nodeUpdatedResource.Name = safeResourceName("clusternewname")
	nodeUpdatedResource.Description = "new-description"
	nodeUpdatedResource.Version = k8sTestUpdateVersion
	nodeUpdatedResource.LabelKey = "new_label_key"
	nodeUpdatedResource.LabelValue = "new_label_value"
	nodeUpdatedResource.NodeLabelKey = "new_node_label_key"
	nodeUpdatedResource.NodeLabelValue = "new_node_label_value"
	nodeUpdatedResource.Memory = "4"
	nodeUpdatedResource.Cores = "2"
	nodeUpdatedResource.DiskSize = "65"
	nodeUpdatedResource.Preemptible = "false"

	// update maintenance policy
	nodeUpdatedResource.constructMaintenancePolicyField(false, false, dailyMaintenancePolicy)
	// commented, because of current quotes for summary disk size
	//nodeUpdatedResource.FixedScale = "2"

	nodeUpdatedResource2 := nodeUpdatedResource
	nodeUpdatedResource2.constructMaintenancePolicyField(true, true, weeklyMaintenancePolicy)

	nodeUpdatedResource3 := nodeUpdatedResource2
	nodeUpdatedResource3.constructMaintenancePolicyField(true, true, emptyMaintenancePolicy)

	nodeUpdatedResource4 := nodeUpdatedResource3
	nodeUpdatedResource4.constructMaintenancePolicyField(false, true, weeklyMaintenancePolicySecond)

	nodeUpdatedResource5 := nodeUpdatedResource4
	nodeUpdatedResource5.constructMaintenancePolicyField(true, false, anyMaintenancePolicy)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, true, false),
				),
			},
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeUpdatedResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeUpdatedResource, true, false),
				),
			},
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeUpdatedResource2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeUpdatedResource2, true, false),
				),
			},
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeUpdatedResource3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeUpdatedResource3, true, false),
				),
			},
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeUpdatedResource4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeUpdatedResource4, true, false),
				),
			},
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeUpdatedResource5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeUpdatedResource5, true, false),
				),
			},
		},
	})
}

func TestAccKubernetesNodeGroupNetworkInterfaces_update(t *testing.T) {
	clusterResource := clusterInfoWithSecurityGroups("TestAccKubernetesNodeGroupNetworkInterfaces_update", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResource.constructNetworkInterfaces(clusterResource.SubnetResourceNameA, clusterResource.SecurityGroupName)
	nodeResourceFullName := nodeResource.ResourceFullName(true)

	nodeUpdatedResource := nodeResource
	nodeUpdatedResource.NetworkInterfaces = enableNAT

	nodeUpdatedResource2 := nodeUpdatedResource
	nodeUpdatedResource2.constructNetworkInterfaces(clusterResource.SubnetResourceNameA, "")

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, true, false),
				),
			},
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeUpdatedResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeUpdatedResource, true, false),
				),
			},
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeUpdatedResource2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeUpdatedResource2, true, false),
				),
			},
		},
	})
}

func TestAccKubernetesNodeGroup_autoscaled(t *testing.T) {
	clusterResource := clusterInfo("testAccKubernetesNodeGroupConfig_basic", true)
	nodeResource := nodeGroupInfoAutoscaled(clusterResource.ClusterResourceName)
	nodeResourceFullName := nodeResource.ResourceFullName(true)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesNodeGroupConfig_autoscaled(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, true, true),
				),
			},
			k8sNodeGroupImportStep(nodeResourceFullName),
		},
	})
}

func TestAccKubernetesNodeGroup_createPlacementGroup(t *testing.T) {
	clusterResource := clusterInfo("testAccKubernetesNodeGroupConfig_basic", true)
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResource.PlacementGroupId = "yandex_compute_placement_group.pg.id"
	nodeResourceFullName := nodeResource.ResourceFullName(true)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResource) + constPlacementGroupResource,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, true, false),
				),
			},
		},
	})
}

func TestAccKubernetesNodeGroup_dualStack(t *testing.T) {
	clusterResource := clusterInfoDualStack("TestAccKubernetesNodeGroup_dualStack", true)
	nodeResource := nodeGroupInfoDualStack(clusterResource.ClusterResourceName)

	nodeResource.constructNetworkInterfaces(clusterResource.SubnetResourceNameA, clusterResource.SecurityGroupName)
	nodeResourceFullName := nodeResource.ResourceFullName(true)

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
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResource, true, false),
				),
			},
		},
	})
}

func TestAccKubernetesNodeGroup_networkSettings(t *testing.T) {
	clusterResource := clusterInfo("TestAccKubernetesNodeGroup_networkSettings", true)

	nodeResourceStandard := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResourceSoftwareAcceleration := nodeResourceStandard

	nodeResourceStandard.NetworkAccelerationType = "standard"
	nodeResourceSoftwareAcceleration.NetworkAccelerationType = "software_accelerated"

	nodeResourceFullName := nodeResourceStandard.ResourceFullName(true)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResourceSoftwareAcceleration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResourceSoftwareAcceleration, true, false),
				),
			},
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResourceStandard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResourceStandard, true, false),
				),
			},
			k8sNodeGroupImportStep(nodeResourceFullName),
		},
	})
}

func TestAccKubernetesNodeGroup_containerRuntime(t *testing.T) {
	clusterResource := clusterInfo("TestAccKubernetesNodeGroup_containerRuntime", true)

	nodeResourceDocker := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResourceContainerd := nodeResourceDocker

	nodeResourceDocker.ContainerRuntimeType = "docker"
	nodeResourceContainerd.ContainerRuntimeType = "containerd"

	nodeResourceFullName := nodeResourceDocker.ResourceFullName(true)

	var ng k8s.NodeGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResourceDocker),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResourceDocker, true, false),
				),
			},
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResourceContainerd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeResourceContainerd, true, false),
				),
			},
		},
	})
}

func TestAccKubernetesNodeGroup_containerRuntime_invalid(t *testing.T) {
	t.Parallel()

	clusterResource := clusterInfo("TestAccKubernetesNodeGroup_containerRuntime_invalid", true)
	nodeResourceInvalid := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResourceInvalid.ContainerRuntimeType = "some_invalid_type"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeResourceInvalid),
				ExpectError: regexp.MustCompile("expected instance_template.0.container_runtime.0.type to be one of"),
			},
		},
	})
}

type resourceNodeGroupInfo struct {
	ClusterResourceName   string
	NodeGroupResourceName string
	Name                  string
	Description           string
	Version               string

	ContainerRuntimeType string

	Memory string
	Cores  string

	DiskSize         string
	Preemptible      string
	ScalePolicy      string
	PlacementGroupId string

	LabelKey   string
	LabelValue string

	NodeLabelKey   string
	NodeLabelValue string

	MaintenancePolicy string

	NetworkInterfaces string

	autoUpgrade       bool
	autoRepair        bool
	policy            maintenancePolicyType
	SecurityGroupName string
	SubnetName        string

	DualStack bool

	NetworkAccelerationType string
}

func nodeGroupInfo(clusterResourceName string) resourceNodeGroupInfo {
	return nodeGroupInfoWithMaintenance(clusterResourceName, true, true, anyMaintenancePolicy)
}

func nodeGroupInfoDualStack(clusterResourceName string) resourceNodeGroupInfo {
	ng := nodeGroupInfoWithMaintenance(clusterResourceName, true, true, anyMaintenancePolicy)
	ng.DualStack = true
	return ng
}

func nodeGroupInfoWithMaintenance(clusterResourceName string, autoUpgrade, autoRepair bool, policyType maintenancePolicyType) resourceNodeGroupInfo {
	info := resourceNodeGroupInfo{
		ClusterResourceName:   clusterResourceName,
		NodeGroupResourceName: randomResourceName("nodegroup"),
		Name:                  safeResourceName("nodegroupname"),
		Description:           "description",
		Version:               k8sTestVersion,
		Memory:                "2",
		Cores:                 "2",
		DiskSize:              "64",
		Preemptible:           "true",
		LabelKey:              "label_key",
		LabelValue:            "label_value",
		NodeLabelKey:          "node_label_key",
		NodeLabelValue:        "node_label_value",
		ScalePolicy:           fixedScalePolicy,
		NetworkInterfaces:     enableNAT,
	}

	info.constructMaintenancePolicyField(autoUpgrade, autoRepair, policyType)
	return info
}

func nodeGroupInfoAutoscaled(clusterResourceName string) resourceNodeGroupInfo {
	info := nodeGroupInfo(clusterResourceName)
	info.ScalePolicy = autoscaledScalePolicy
	return info
}

func (i *resourceNodeGroupInfo) Map() map[string]interface{} {
	return structs.Map(i)
}

func (i *resourceNodeGroupInfo) ResourceFullName(resource bool) string {
	if resource {
		return "yandex_kubernetes_node_group." + i.NodeGroupResourceName
	}

	return "data.yandex_kubernetes_node_group." + i.NodeGroupResourceName
}

func (i *resourceNodeGroupInfo) constructMaintenancePolicyField(autoUpgrade, autoRepair bool, policy maintenancePolicyType) {
	m := map[string]interface{}{
		"AutoUpgrade": autoUpgrade,
		"AutoRepair":  autoRepair,
	}

	i.autoUpgrade = autoUpgrade
	i.autoRepair = autoRepair
	i.policy = policy

	switch policy {
	case emptyMaintenancePolicy:
		i.MaintenancePolicy = ""
	case anyMaintenancePolicy:
		i.MaintenancePolicy = templateConfig(ngAnyMaintenancePolicyTemplate, m)
	case dailyMaintenancePolicy:
		i.MaintenancePolicy = templateConfig(ngDailyMaintenancePolicyTemplate, m)
	case weeklyMaintenancePolicy:
		i.MaintenancePolicy = templateConfig(ngWeeklyMaintenancePolicyTemplate, m)
	case weeklyMaintenancePolicySecond:
		i.MaintenancePolicy = templateConfig(ngWeeklyMaintenancePolicyTemplateSecond, m)
	}
}

func (i *resourceNodeGroupInfo) constructNetworkInterfaces(subnetName, securityGroupName string) {
	i.SubnetName = subnetName
	i.SecurityGroupName = securityGroupName
	if i.DualStack {
		i.NetworkInterfaces = fmt.Sprintf(networkInterfacesTemplateDualStack, subnetName, securityGroupName)
		return
	}

	subnetNameGetter := fmt.Sprintf("\"${yandex_vpc_subnet.%s.id}\"", i.SubnetName)
	securityGroupIDGetter := ""
	if securityGroupName != "" {
		securityGroupIDGetter = fmt.Sprintf("\"${yandex_vpc_security_group.%s.id}\"", i.SecurityGroupName)
	}

	i.NetworkInterfaces = fmt.Sprintf(networkInterfacesTemplate,
		subnetNameGetter,
		securityGroupIDGetter,
	)
}

func (i *resourceNodeGroupInfo) securityGroupName() string {
	return "yandex_vpc_security_group." + i.SecurityGroupName
}

func (i *resourceNodeGroupInfo) subnetName() string {
	return "yandex_vpc_subnet." + i.SubnetName
}

const ngAnyMaintenancePolicyTemplate = `
	maintenance_policy {
        auto_upgrade = {{.AutoUpgrade}}
        auto_repair  = {{.AutoRepair}}
    }
`

const ngDailyMaintenancePolicyTemplate = `
	maintenance_policy {
        auto_upgrade = {{.AutoUpgrade}}
        auto_repair  = {{.AutoRepair}}

        maintenance_window {
			start_time = "15:00"
			duration   = "3h"
		}
    }
`

const ngWeeklyMaintenancePolicyTemplate = `
	maintenance_policy {
        auto_upgrade = {{.AutoUpgrade}}
        auto_repair  = {{.AutoRepair}}

        maintenance_window {
            day		   = "monday"
			start_time = "15:00"
			duration   = "3h"
		}

        maintenance_window {
            day		   = "friday"
			start_time = "10:00"
			duration   = "4h"
		}
    }
`

const ngWeeklyMaintenancePolicyTemplateSecond = `
	maintenance_policy {
        auto_upgrade = {{.AutoUpgrade}}
        auto_repair  = {{.AutoRepair}}

        maintenance_window {
            day		   = "monday"
			start_time = "15:00"
			duration   = "5h"
		}

        maintenance_window {
            day		   = "friday"
			start_time = "12:00"
			duration   = "4h"
		}
    }
`

const (
	autoscaledMinSize     = 1
	autoscaledMaxSize     = 3
	autoscaledInitialSize = 2
)

var autoscaledScalePolicy = fmt.Sprintf(`
  scale_policy {
    auto_scale {
      min = %d
      max = %d
      initial = %d
    }
  }
`, autoscaledMinSize, autoscaledMaxSize, autoscaledInitialSize)

const fixedScaleSize = 1

var fixedScalePolicy = fmt.Sprintf(`
  scale_policy {
    fixed_scale {
      size = %d
    }
  }
`, fixedScaleSize)

const nodeGroupConfigTemplate = `
resource "yandex_kubernetes_node_group" "{{.NodeGroupResourceName}}" {
  cluster_id = "${yandex_kubernetes_cluster.{{.ClusterResourceName}}.id}"
  name        = "{{.Name}}"
  description = "{{.Description}}"

  labels = {
	{{.LabelKey}} = "{{.LabelValue}}"
  }
  node_labels = {
	{{.NodeLabelKey}} = "{{.NodeLabelValue}}"
  }

  instance_template {
    platform_id = "standard-v2"

	{{if .ContainerRuntimeType}}
	container_runtime {
		type = "{{.ContainerRuntimeType}}"
	}
	{{end}}

    {{.NetworkInterfaces}}

    resources {
      memory = {{.Memory}}
      cores  = {{.Cores}}
    }

    boot_disk {
      type = "network-hdd"
      size     = {{.DiskSize}}
    }

    scheduling_policy {
      preemptible = {{.Preemptible}}
    }
    {{if .PlacementGroupId}}
    placement_policy {
      placement_group_id = {{.PlacementGroupId}}
    }
    {{end}}

    {{if .NetworkAccelerationType}}
	network_acceleration_type = "{{.NetworkAccelerationType}}"
	{{end}}
  }

  {{.ScalePolicy}}
  
  allocation_policy {
    location {
      zone = "ru-central1-a"
    }
  }

  version = "{{.Version}}"

  {{.MaintenancePolicy}}

  node_taints = [
    "key1=value1:NoSchedule"
  ]
  allowed_unsafe_sysctls = [
    "kernel.msg*",
    "net.core.somaxconn",
  ]
}
`

const enableNAT = "nat = true"

var networkInterfacesTemplate = `
  network_interface {
	nat = true
    subnet_ids = [%s]
	security_group_ids = [%s]
  }
`

var networkInterfacesTemplateDualStack = `
  network_interface {
	ipv4 = true
	ipv6 = true
    subnet_ids = ["%s"]
    security_group_ids = ["%s"]
  }
`

// language=tf
const constPlacementGroupResource = `
resource yandex_compute_placement_group pg {
}
`

func testAccKubernetesNodeGroupConfig_basic(cluster resourceClusterInfo, ng resourceNodeGroupInfo) string {
	deps := testAccKubernetesClusterZonalConfig_basic(cluster)
	return deps + templateConfig(nodeGroupConfigTemplate, ng.Map())
}

func testAccKubernetesNodeGroupConfig_autoscaled(cluster resourceClusterInfo, ng resourceNodeGroupInfo) string {
	deps := testAccKubernetesClusterZonalConfig_basic(cluster)
	return deps + templateConfig(nodeGroupConfigTemplate, ng.Map())
}

func checkNodeGroupAttributes(ng *k8s.NodeGroup, info *resourceNodeGroupInfo, rs bool, autoscaled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		versionInfo := ng.GetVersionInfo()
		scalePolicy := ng.GetScalePolicy()
		locations := ng.GetAllocationPolicy().GetLocations()
		tpl := ng.GetNodeTemplate()
		if tpl == nil || versionInfo == nil || scalePolicy == nil || len(locations) != 1 {
			return fmt.Errorf("failed to get kubernetes node group specs info")
		}

		resourceFullName := info.ResourceFullName(rs)
		checkFuncsAr := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(resourceFullName, "cluster_id", ng.ClusterId),

			resource.TestCheckResourceAttr(resourceFullName, "name", info.Name),
			resource.TestCheckResourceAttr(resourceFullName, "name", ng.Name),

			resource.TestCheckResourceAttr(resourceFullName, "description", info.Description),
			resource.TestCheckResourceAttr(resourceFullName, "description", ng.Description),

			resource.TestCheckResourceAttr(resourceFullName, "instance_group_id", ng.GetInstanceGroupId()),

			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.resources.0.memory", info.Memory),
			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.resources.0.cores", info.Cores),
			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.resources.0.cores",
				strconv.Itoa(int(tpl.GetResourcesSpec().GetCores()))),

			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.boot_disk.0.type", "network-hdd"),
			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.boot_disk.0.type",
				tpl.GetBootDiskSpec().GetDiskTypeId()),

			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.boot_disk.0.size", info.DiskSize),

			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.platform_id", "standard-v2"),
			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.platform_id", tpl.GetPlatformId()),

			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.scheduling_policy.0.preemptible", info.Preemptible),
			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.scheduling_policy.0.preemptible",
				strconv.FormatBool(tpl.GetSchedulingPolicy().GetPreemptible())),
			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_acceleration_type",
				strings.ToLower(tpl.NetworkSettings.Type.String())),

			resource.TestCheckResourceAttr(resourceFullName, "version_info.0.current_version",
				versionInfo.GetCurrentVersion()),
			resource.TestCheckResourceAttr(resourceFullName, "version_info.0.new_revision_available",
				strconv.FormatBool(versionInfo.GetNewRevisionAvailable())),
			resource.TestCheckResourceAttr(resourceFullName, "version_info.0.new_revision_summary",
				versionInfo.GetNewRevisionSummary()),
			resource.TestCheckResourceAttr(resourceFullName, "version_info.0.version_deprecated",
				strconv.FormatBool(versionInfo.GetVersionDeprecated())),

			resource.TestCheckResourceAttr(resourceFullName, "allocation_policy.0.location.#", "1"),
			resource.TestCheckResourceAttr(resourceFullName, "allocation_policy.0.location.0.zone", locations[0].GetZoneId()),
			resource.TestCheckResourceAttr(resourceFullName, "allocation_policy.0.location.0.subnet_id", locations[0].GetSubnetId()),

			resource.TestCheckResourceAttr(resourceFullName, "maintenance_policy.0.auto_upgrade", strconv.FormatBool(ng.GetMaintenancePolicy().GetAutoUpgrade())),
			resource.TestCheckResourceAttr(resourceFullName, "maintenance_policy.0.auto_repair", strconv.FormatBool(ng.GetMaintenancePolicy().GetAutoRepair())),

			testAccCheckNodeGroupLabel(ng, info, rs),
			testAccCheckCreatedAtAttr(resourceFullName),

			testCheckResourceMap(resourceFullName, "node_labels", ng.GetNodeLabels()),
			testCheckResourceList(resourceFullName, "node_taints", formatTaints(ng.GetNodeTaints())),
			testCheckResourceList(resourceFullName, "allowed_unsafe_sysctls", ng.GetAllowedUnsafeSysctls()),

			resource.TestCheckResourceAttr(resourceFullName, "deploy_policy.0.max_unavailable", strconv.Itoa(int(ng.GetDeployPolicy().GetMaxUnavailable()))),
			resource.TestCheckResourceAttr(resourceFullName, "deploy_policy.0.max_expansion", strconv.Itoa(int(ng.GetDeployPolicy().GetMaxExpansion()))),

			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.0.ipv4",
				strconv.FormatBool(tpl.NetworkInterfaceSpecs[0].PrimaryV4AddressSpec != nil)),
			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.0.ipv6",
				strconv.FormatBool(tpl.NetworkInterfaceSpecs[0].PrimaryV6AddressSpec != nil)),
			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.0.nat",
				strconv.FormatBool(
					tpl.NetworkInterfaceSpecs[0].PrimaryV4AddressSpec != nil &&
						tpl.NetworkInterfaceSpecs[0].PrimaryV4AddressSpec.OneToOneNatSpec != nil)),
		}

		if info.PlacementGroupId != "" {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName,
					"instance_template.0.placement_policy.0.placement_group_id",
					tpl.PlacementPolicy.PlacementGroupId),
			)
		}

		if info.ContainerRuntimeType != "" {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.container_runtime.0.type",
					strings.ToLower(tpl.GetContainerRuntimeSettings().GetType().String())),
				resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.container_runtime.0.type",
					strings.ToLower(info.ContainerRuntimeType)),
			)
		}

		if info.NetworkInterfaces != enableNAT {
			subnetID := info.SubnetName
			if !info.DualStack {
				var err error
				subnetID, err = getResourceID(info.subnetName(), s)
				if err != nil {
					return err
				}
			}

			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.#", "1"),
				resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.0.subnet_ids.#", "1"),
				resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.0.subnet_ids.0", subnetID),
				resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.0.subnet_ids.0", ng.NodeTemplate.NetworkInterfaceSpecs[0].SubnetIds[0]),
			)

			if info.SecurityGroupName != "" {
				securityGroupID := info.SecurityGroupName
				if !info.DualStack {
					var err error
					securityGroupID, err = getResourceID(info.securityGroupName(), s)
					if err != nil {
						return err
					}
				}

				checkFuncsAr = append(checkFuncsAr,
					resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.0.security_group_ids.0", securityGroupID),
					resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.0.security_group_ids.0", ng.NodeTemplate.NetworkInterfaceSpecs[0].SecurityGroupIds[0]),
				)
			}
		} else {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.#", "1"),
				resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.network_interface.0.nat", "true"),
			)
		}

		if info.policy != emptyMaintenancePolicy {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "maintenance_policy.0.auto_upgrade", strconv.FormatBool(info.autoUpgrade)),
				resource.TestCheckResourceAttr(resourceFullName, "maintenance_policy.0.auto_repair", strconv.FormatBool(info.autoRepair)),
			)
		}

		const maintenanceWindowPrefix = "maintenance_policy.0.maintenance_window."
		switch info.policy {
		case anyMaintenancePolicy:
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, maintenanceWindowPrefix+"#", "0"),
			)
		case dailyMaintenancePolicy:
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, maintenanceWindowPrefix+"#", "1"),
				testAccCheckMaintenanceWindow(resourceFullName, maintenanceWindowPrefix, "", "15:00", "3h"),
			)
		case weeklyMaintenancePolicy:
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, maintenanceWindowPrefix+"#", "2"),
				testAccCheckMaintenanceWindow(resourceFullName, maintenanceWindowPrefix, "monday", "15:00", "3h"),
				testAccCheckMaintenanceWindow(resourceFullName, maintenanceWindowPrefix, "friday", "10:00", "4h"),
			)
		case weeklyMaintenancePolicySecond:
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, maintenanceWindowPrefix+"#", "2"),
				testAccCheckMaintenanceWindow(resourceFullName, maintenanceWindowPrefix, "monday", "15:00", "5h"),
				testAccCheckMaintenanceWindow(resourceFullName, maintenanceWindowPrefix, "friday", "12:00", "4h"),
			)
		}

		if !autoscaled {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "scale_policy.0.fixed_scale.0.size", strconv.Itoa(int(scalePolicy.GetFixedScale().GetSize()))),
				resource.TestCheckResourceAttr(resourceFullName, "scale_policy.0.fixed_scale.0.size", strconv.Itoa(fixedScaleSize)),
			)
		} else {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "scale_policy.0.auto_scale.0.min", strconv.Itoa(int(scalePolicy.GetAutoScale().GetMinSize()))),
				resource.TestCheckResourceAttr(resourceFullName, "scale_policy.0.auto_scale.0.min", strconv.Itoa(autoscaledMinSize)),
				resource.TestCheckResourceAttr(resourceFullName, "scale_policy.0.auto_scale.0.max", strconv.Itoa(int(scalePolicy.GetAutoScale().GetMaxSize()))),
				resource.TestCheckResourceAttr(resourceFullName, "scale_policy.0.auto_scale.0.max", strconv.Itoa(autoscaledMaxSize)),
				resource.TestCheckResourceAttr(resourceFullName, "scale_policy.0.auto_scale.0.initial", strconv.Itoa(int(scalePolicy.GetAutoScale().GetInitialSize()))),
				resource.TestCheckResourceAttr(resourceFullName, "scale_policy.0.auto_scale.0.initial", strconv.Itoa(autoscaledInitialSize)),
			)
		}

		if rs {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "version", info.Version),
				resource.TestCheckResourceAttr(resourceFullName, "version", ng.GetVersionInfo().GetCurrentVersion()),
			)
		}

		return resource.ComposeTestCheckFunc(checkFuncsAr...)(s)
	}
}

func testAccCheckNodeGroupLabel(ng *k8s.NodeGroup, info *resourceNodeGroupInfo, rs bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(ng.Labels) != 1 {
			return fmt.Errorf("should be exactly one label on kubernetes node group %s", ng.Name)
		}

		v, ok := ng.Labels[info.LabelKey]
		if !ok {
			return fmt.Errorf("no label found with key %s on kubernetes node group %s", info.LabelKey, ng.Name)
		}
		if v != info.LabelValue {
			return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on kubernetes node group %s",
				info.LabelValue, v, info.LabelKey, ng.Name)
		}

		objName := info.ResourceFullName(rs)
		labelPath := fmt.Sprintf("labels.%s", info.LabelKey)

		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(objName, "labels.%", "1"),
			resource.TestCheckResourceAttr(objName, labelPath, info.LabelValue))(s)
	}
}

func testAccCheckKubernetesNodeGroupExists(n string, ng *k8s.NodeGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.Kubernetes().NodeGroup().Get(context.Background(), &k8s.GetNodeGroupRequest{
			NodeGroupId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Kubernetes node group not found")
		}

		*ng = *found
		return nil
	}
}

func testAccCheckKubernetesNodeGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_kubernetes_node_group" {
			continue
		}

		_, err := config.sdk.Kubernetes().NodeGroup().Get(context.Background(), &k8s.GetNodeGroupRequest{
			NodeGroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Kubernetes node group still exists")
		}
	}

	return nil
}

func formatTaints(taints []*k8s.Taint) []string {
	var formatted []string
	for _, t := range taints {
		var effect string
		switch t.Effect {
		case k8s.Taint_NO_EXECUTE:
			effect = "NoExecute"
		case k8s.Taint_NO_SCHEDULE:
			effect = "NoSchedule"
		case k8s.Taint_PREFER_NO_SCHEDULE:
			effect = "PreferNoSchedule"
		}
		formatted = append(formatted, fmt.Sprintf("%s=%s:%s", t.Key, t.Value, effect))
	}
	return formatted
}

func testCheckResourceMap(objName string, key string, m map[string]string) resource.TestCheckFunc {
	checkFuncs := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(objName, fmt.Sprintf("%s.%%", key), strconv.Itoa(len(m))),
	}

	for k, v := range m {
		labelPath := fmt.Sprintf("%s.%s", key, k)
		checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr(objName, labelPath, v))
	}

	return resource.ComposeTestCheckFunc(checkFuncs...)
}

func testCheckResourceList(objName string, key string, values []string) resource.TestCheckFunc {
	checkFuncs := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(objName, fmt.Sprintf("%s.#", key), strconv.Itoa(len(values))),
	}

	for i, v := range values {
		labelPath := fmt.Sprintf("%s.%d", key, i)
		checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr(objName, labelPath, v))
	}

	return resource.ComposeTestCheckFunc(checkFuncs...)
}
