package yandex

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/fatih/structs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	k8s "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
)

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
	t.Parallel()

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
					checkNodeGroupAttributes(&ng, &nodeResource, true),
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

func TestAccKubernetesNodeGroup_update(t *testing.T) {
	t.Parallel()

	clusterResource := clusterInfo("testAccKubernetesNodeGroupConfig_basic", true)
	clusterResource.ReleaseChannel = k8s.ReleaseChannel_REGULAR.String()
	clusterResource.MasterVersion = "1.15"
	nodeResource := nodeGroupInfo(clusterResource.ClusterResourceName)
	nodeResource.Version = "1.14"
	nodeResourceFullName := nodeResource.ResourceFullName(true)

	nodeUpdatedResource := nodeResource

	nodeUpdatedResource.Name = safeResourceName("clusternewname")
	nodeUpdatedResource.Description = "new-description"
	nodeUpdatedResource.Version = "1.15"
	nodeUpdatedResource.LabelKey = "new_label_key"
	nodeUpdatedResource.LabelValue = "new_label_value"
	nodeUpdatedResource.Memory = "4"
	nodeUpdatedResource.Cores = "2"
	nodeUpdatedResource.DiskSize = "65"
	nodeUpdatedResource.Preemptible = "false"

	// commented, because of current quotes for summary disk size
	//nodeUpdatedResource.FixedScale = "2"

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
					checkNodeGroupAttributes(&ng, &nodeResource, true),
				),
			},
			{
				Config: testAccKubernetesNodeGroupConfig_basic(clusterResource, nodeUpdatedResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesNodeGroupExists(nodeResourceFullName, &ng),
					checkNodeGroupAttributes(&ng, &nodeUpdatedResource, true),
				),
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

	Memory string
	Cores  string

	DiskSize    string
	Preemptible string
	FixedScale  string

	LabelKey   string
	LabelValue string
}

func nodeGroupInfo(clusterResourceName string) resourceNodeGroupInfo {
	return resourceNodeGroupInfo{
		ClusterResourceName:   clusterResourceName,
		NodeGroupResourceName: randomResourceName("nodegroup"),
		Name:                  safeResourceName("nodegroupname"),
		Description:           "description",
		Version:               "1.13",
		Memory:                "2",
		Cores:                 "1",
		DiskSize:              "64",
		Preemptible:           "true",
		FixedScale:            "1",
		LabelKey:              "label_key",
		LabelValue:            "label_value",
	}
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

const nodeGroupConfigTemplate = `
resource "yandex_kubernetes_node_group" "{{.NodeGroupResourceName}}" {
  cluster_id = "${yandex_kubernetes_cluster.{{.ClusterResourceName}}.id}"
  name        = "{{.Name}}"
  description = "{{.Description}}"

  labels = {
	{{.LabelKey}} = "{{.LabelValue}}"
  }

  instance_template {
    platform_id = "standard-v1"
    nat = "true"

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
  }

  scale_policy {
    fixed_scale {
      size = {{.FixedScale}}
    }
  }
  
  allocation_policy {
    location {
      zone = "ru-central1-a"
    }
  }

  version = {{.Version}}
}
`

func testAccKubernetesNodeGroupConfig_basic(cluster resourceClusterInfo, ng resourceNodeGroupInfo) string {
	deps := testAccKubernetesClusterZonalConfig_basic(cluster)
	return deps + templateConfig(nodeGroupConfigTemplate, ng.Map())
}

func checkNodeGroupAttributes(ng *k8s.NodeGroup, info *resourceNodeGroupInfo, rs bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		versionInfo := ng.GetVersionInfo()
		scalePolicy := ng.GetScalePolicy().GetFixedScale()
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

			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.platform_id", "standard-v1"),
			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.platform_id", tpl.GetPlatformId()),

			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.nat", "true"),
			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.nat",
				strconv.FormatBool(tpl.GetV4AddressSpec().GetOneToOneNatSpec().GetIpVersion() == k8s.IpVersion_IPV4)),

			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.scheduling_policy.0.preemptible", info.Preemptible),
			resource.TestCheckResourceAttr(resourceFullName, "instance_template.0.scheduling_policy.0.preemptible",
				strconv.FormatBool(tpl.GetSchedulingPolicy().GetPreemptible())),

			resource.TestCheckResourceAttr(resourceFullName, "version_info.0.current_version",
				versionInfo.GetCurrentVersion()),
			resource.TestCheckResourceAttr(resourceFullName, "version_info.0.new_revision_available",
				strconv.FormatBool(versionInfo.GetNewRevisionAvailable())),
			resource.TestCheckResourceAttr(resourceFullName, "version_info.0.new_revision_summary",
				versionInfo.GetNewRevisionSummary()),
			resource.TestCheckResourceAttr(resourceFullName, "version_info.0.version_deprecated",
				strconv.FormatBool(versionInfo.GetVersionDeprecated())),

			resource.TestCheckResourceAttr(resourceFullName, "scale_policy.0.fixed_scale.0.size", strconv.Itoa(int(scalePolicy.GetSize()))),
			resource.TestCheckResourceAttr(resourceFullName, "scale_policy.0.fixed_scale.0.size", info.FixedScale),

			resource.TestCheckResourceAttr(resourceFullName, "allocation_policy.0.location.#", "1"),
			resource.TestCheckResourceAttr(resourceFullName, "allocation_policy.0.location.0.zone", locations[0].GetZoneId()),
			resource.TestCheckResourceAttr(resourceFullName, "allocation_policy.0.location.0.subnet_id", locations[0].GetSubnetId()),

			testAccCheckNodeGroupLabel(ng, info, rs),
			testAccCheckCreatedAtAttr(resourceFullName),
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
