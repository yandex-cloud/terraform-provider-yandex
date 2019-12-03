package yandex

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/fatih/structs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	k8s "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
)

func k8sClusterImportStep(clusterResourceFullName string, ignored ...string) resource.TestStep {
	return resource.TestStep{
		ResourceName:            clusterResourceFullName,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: ignored,
	}
}

//revive:disable:var-naming
func TestAccKubernetesClusterZonal_basic(t *testing.T) {
	t.Parallel()

	clusterResource := clusterInfo("testAccKubernetesClusterZonalConfig_basic", true)
	clusterResourceFullName := clusterResource.ResourceFullName(true)

	var cluster k8s.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesClusterZonalConfig_basic(clusterResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterResource, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
			k8sClusterImportStep(clusterResourceFullName, "master.0.zonal"),
		},
	})
}

func TestAccKubernetesClusterZonalNoVersion_basic(t *testing.T) {
	t.Parallel()

	clusterResource := clusterInfo("TestAccKubernetesClusterZonalNoVersion_basic", true)
	clusterResource.MasterVersion = ""
	clusterResourceFullName := clusterResource.ResourceFullName(true)

	var cluster k8s.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesClusterZonalConfig_basic(clusterResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterResource, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
		},
	})
}

func TestAccKubernetesClusterRegional_basic(t *testing.T) {
	t.Parallel()

	clusterResource := clusterInfo("testAccKubernetesClusterRegionalConfig_basic", false)
	clusterResourceFullName := clusterResource.ResourceFullName(true)

	var cluster k8s.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesClusterRegionalConfig_basic(clusterResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterResource, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
			k8sClusterImportStep(clusterResourceFullName, "master.0.regional"),
		},
	})
}

func TestAccKubernetesClusterZonal_update(t *testing.T) {
	t.Parallel()

	clusterResource := clusterInfo("testAccKubernetesClusterZonalConfig_basic", true)
	clusterResourceFullName := clusterResource.ResourceFullName(true)

	clusterUpdatedResource := clusterResource

	clusterUpdatedResource.Name = safeResourceName("clusternewname")
	clusterUpdatedResource.Description = "new-description"
	clusterUpdatedResource.LabelKey = "new_label_key"
	clusterUpdatedResource.LabelValue = "new_label_value"
	// switch service accounts
	clusterUpdatedResource.ServiceAccountResourceName = clusterResource.NodeServiceAccountResourceName
	clusterUpdatedResource.NodeServiceAccountResourceName = clusterResource.ServiceAccountResourceName
	clusterUpdatedResource.TestDescription = "testAccKubernetesClusterZonalConfig_update"
	clusterUpdatedResource.MasterVersion = "1.14"

	var cluster k8s.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesClusterZonalConfig_basic(clusterResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterResource, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
			{
				Config: testAccKubernetesClusterZonalConfig_update(clusterResource, clusterUpdatedResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterUpdatedResource, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
		},
	})
}

func TestAccKubernetesClusterRegional_update(t *testing.T) {
	t.Parallel()

	clusterResource := clusterInfo("testAccKubernetesClusterRegionalConfig_basic", false)
	clusterResourceFullName := clusterResource.ResourceFullName(true)

	clusterUpdatedResource := clusterResource

	clusterUpdatedResource.Name = safeResourceName("clusternewname")
	clusterUpdatedResource.Description = "new-description"
	clusterUpdatedResource.LabelKey = "new_label_key"
	clusterUpdatedResource.LabelValue = "new_label_value"
	// switch service accounts
	clusterUpdatedResource.ServiceAccountResourceName = clusterResource.NodeServiceAccountResourceName
	clusterUpdatedResource.NodeServiceAccountResourceName = clusterResource.ServiceAccountResourceName
	clusterUpdatedResource.TestDescription = "testAccKubernetesClusterRegionalConfig_update"
	clusterUpdatedResource.MasterVersion = "1.14"

	var cluster k8s.Cluster

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesClusterRegionalConfig_basic(clusterResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterResource, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
			{
				Config: testAccKubernetesClusterRegionalConfig_update(clusterResource, clusterUpdatedResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterUpdatedResource, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
		},
	})
}

func TestAccKubernetesCluster_wrong(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccKubernetesClusterConfig_wrong(),
				ExpectError: regexp.MustCompile("conflicts with master.0"),
			},
		},
	})
}

func randomResourceName(tp string) string {
	return fmt.Sprintf("test_%s_%s", tp, acctest.RandString(10))
}

// iam uses strict validation for SA names
func safeResourceName(tp string) string {
	return fmt.Sprintf("test%s%s", tp, acctest.RandString(10))
}

func clusterInfo(testDesc string, zonal bool) resourceClusterInfo {
	return resourceClusterInfo{
		ClusterResourceName:            randomResourceName("cluster"),
		FolderID:                       os.Getenv("YC_FOLDER_ID"),
		Name:                           safeResourceName("clustername"),
		Description:                    "description",
		MasterVersion:                  "1.13",
		LabelKey:                       "label_key",
		LabelValue:                     "label_value",
		TestDescription:                testDesc,
		NetworkResourceName:            randomResourceName("network"),
		SubnetResourceNameA:            randomResourceName("subnet"),
		SubnetResourceNameB:            randomResourceName("subnet"),
		SubnetResourceNameC:            randomResourceName("subnet"),
		ServiceAccountResourceName:     safeResourceName("serviceaccount"),
		NodeServiceAccountResourceName: safeResourceName("nodeserviceaccount"),
		ReleaseChannel:                 k8s.ReleaseChannel_STABLE.String(),
		zonal:                          zonal,
	}
}

type clusterResourceIDs struct {
	networkResourceID            string
	subnetAResourceID            string
	subnetBResourceID            string
	subnetCResourceID            string
	serviceAccountResourceID     string
	nodeServiceAccountResourceID string
}

func getClusterResourcesIds(s *terraform.State, info *resourceClusterInfo) (ids clusterResourceIDs, err error) {
	ids.networkResourceID, err = getResourceID(info.networkResourceName(), s)
	if err != nil {
		return
	}

	ids.subnetAResourceID, err = getResourceID(info.subnetAResourceName(), s)
	if err != nil {
		return
	}

	ids.subnetBResourceID, err = getResourceID(info.subnetBResourceName(), s)
	if err != nil {
		return
	}

	ids.subnetCResourceID, err = getResourceID(info.subnetCResourceName(), s)
	if err != nil {
		return
	}

	ids.serviceAccountResourceID, err = getResourceID(info.serviceAccountResourceName(), s)
	if err != nil {
		return
	}

	ids.nodeServiceAccountResourceID, err = getResourceID(info.nodeServiceAccountResourceName(), s)
	if err != nil {
		return
	}

	return
}

func checkClusterAttributes(cluster *k8s.Cluster, info *resourceClusterInfo, rs bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ids, err := getClusterResourcesIds(s, info)
		if err != nil {
			return err
		}

		master := cluster.GetMaster()
		zonalMaster := master.GetZonalMaster()
		regionalMaster := master.GetRegionalMaster()
		versionInfo := master.GetVersionInfo()
		if master == nil || versionInfo == nil || (regionalMaster == nil && zonalMaster == nil) {
			return fmt.Errorf("failed to get cluster master specs")
		}

		if info.zonal && zonalMaster == nil {
			return fmt.Errorf("expected zonal cluster, but got regional")
		}

		if !info.zonal && regionalMaster == nil {
			return fmt.Errorf("expected regional cluster, but got zonal")
		}

		resourceFullName := info.ResourceFullName(rs)
		checkFuncsAr := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(resourceFullName, "service_account_id", ids.serviceAccountResourceID),
			resource.TestCheckResourceAttr(resourceFullName, "node_service_account_id", ids.nodeServiceAccountResourceID),
			resource.TestCheckResourceAttr(resourceFullName, "network_id", ids.networkResourceID),

			resource.TestCheckResourceAttr(resourceFullName, "name", info.Name),
			resource.TestCheckResourceAttr(resourceFullName, "description", info.Description),

			resource.TestCheckResourceAttr(resourceFullName, "release_channel", info.ReleaseChannel),
			resource.TestCheckResourceAttr(resourceFullName, "release_channel", cluster.ReleaseChannel.String()),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.version_info.0.current_version", versionInfo.GetCurrentVersion()),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.version_info.0.new_revision_available", strconv.FormatBool(versionInfo.GetNewRevisionAvailable())),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.version_info.0.new_revision_summary", versionInfo.GetNewRevisionSummary()),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.version_info.0.version_deprecated", strconv.FormatBool(versionInfo.GetVersionDeprecated())),

			resource.TestCheckResourceAttr(resourceFullName, "name", cluster.Name),
			resource.TestCheckResourceAttr(resourceFullName, "description", cluster.Description),
			resource.TestCheckResourceAttr(resourceFullName, "service_account_id", cluster.ServiceAccountId),
			resource.TestCheckResourceAttr(resourceFullName, "node_service_account_id", cluster.NodeServiceAccountId),
			resource.TestCheckResourceAttr(resourceFullName, "network_id", cluster.NetworkId),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.internal_v4_endpoint", master.GetEndpoints().GetInternalV4Endpoint()),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.external_v4_endpoint", master.GetEndpoints().GetExternalV4Endpoint()),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.cluster_ca_certificate", master.GetMasterAuth().GetClusterCaCertificate()),
			testAccCheckClusterLabel(cluster, info, rs),
		}

		if zonalMaster != nil {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "master.0.zonal.0.zone",
					zonalMaster.GetZoneId()),
				resource.TestCheckResourceAttr(resourceFullName, "master.0.external_v4_address",
					zonalMaster.GetExternalV4Address()),
				resource.TestCheckResourceAttr(resourceFullName, "master.0.internal_v4_address",
					zonalMaster.GetInternalV4Address()),
			)

			if rs {
				checkFuncsAr = append(checkFuncsAr,
					resource.TestCheckResourceAttr(resourceFullName, "master.0.zonal.0.subnet_id",
						ids.subnetAResourceID),
				)
			}
		}

		if regionalMaster != nil {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.region",
					regionalMaster.GetRegionId()),
				resource.TestCheckResourceAttr(resourceFullName, "master.0.external_v4_address",
					regionalMaster.GetExternalV4Address()),
				resource.TestCheckResourceAttr(resourceFullName, "master.0.internal_v4_address",
					regionalMaster.GetInternalV4Address()),
			)

			if rs {
				checkFuncsAr = append(checkFuncsAr,
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.0.subnet_id",
						ids.subnetAResourceID),
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.1.subnet_id",
						ids.subnetBResourceID),
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.2.subnet_id",
						ids.subnetCResourceID),
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.0.zone",
						"ru-central1-a"),
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.1.zone",
						"ru-central1-b"),
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.2.zone",
						"ru-central1-c"),
				)
			}

		}

		if rs {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "master.0.public_ip", "true"),
				resource.TestCheckResourceAttr(resourceFullName, "master.0.version", cluster.GetMaster().GetVersion()),
			)

			if info.MasterVersion != "" {
				checkFuncsAr = append(checkFuncsAr,
					resource.TestCheckResourceAttr(resourceFullName, "master.0.version", info.MasterVersion))
			}
		}

		return resource.ComposeTestCheckFunc(checkFuncsAr...)(s)
	}
}

func testAccCheckClusterLabel(cluster *k8s.Cluster, info *resourceClusterInfo, rs bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(cluster.Labels) != 1 {
			return fmt.Errorf("should be exactly one label on Kubernetes cluster %s", cluster.Name)
		}

		v, ok := cluster.Labels[info.LabelKey]
		if !ok {
			return fmt.Errorf("no label found with key %s on Kubernetes cluster %s", info.LabelKey, cluster.Name)
		}
		if v != info.LabelValue {
			return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on Kubernetes cluster %s",
				info.LabelValue, v, info.LabelKey, cluster.Name)
		}

		objName := info.ResourceFullName(rs)
		labelPath := fmt.Sprintf("labels.%s", info.LabelKey)

		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(objName, "labels.%", "1"),
			resource.TestCheckResourceAttr(objName, labelPath, info.LabelValue))(s)
	}
}

type resourceClusterInfo struct {
	ClusterResourceName string
	FolderID            string
	Name                string
	Description         string
	MasterVersion       string

	LabelKey   string
	LabelValue string

	TestDescription     string
	NetworkResourceName string

	SubnetResourceNameA string
	SubnetResourceNameB string
	SubnetResourceNameC string

	ServiceAccountResourceName     string
	NodeServiceAccountResourceName string
	ReleaseChannel                 string

	zonal bool
}

func (i *resourceClusterInfo) ResourceFullName(resource bool) string {
	if resource {
		return "yandex_kubernetes_cluster." + i.ClusterResourceName
	}

	return "data.yandex_kubernetes_cluster." + i.ClusterResourceName
}

func (i *resourceClusterInfo) Map() map[string]interface{} {
	return structs.Map(i)
}

func (i *resourceClusterInfo) networkResourceName() string {
	return "yandex_vpc_network." + i.NetworkResourceName
}

func (i *resourceClusterInfo) subnetAResourceName() string {
	return "yandex_vpc_subnet." + i.SubnetResourceNameA
}

func (i *resourceClusterInfo) subnetBResourceName() string {
	return "yandex_vpc_subnet." + i.SubnetResourceNameB
}

func (i *resourceClusterInfo) subnetCResourceName() string {
	return "yandex_vpc_subnet." + i.SubnetResourceNameC
}

func (i *resourceClusterInfo) serviceAccountResourceName() string {
	return "yandex_iam_service_account." + i.ServiceAccountResourceName
}

func (i *resourceClusterInfo) nodeServiceAccountResourceName() string {
	return "yandex_iam_service_account." + i.NodeServiceAccountResourceName
}

const zonalClusterConfigTemplate = `
resource "yandex_kubernetes_cluster" "{{.ClusterResourceName}}" {
  depends_on         = [
	"yandex_resourcemanager_folder_iam_member.{{.ServiceAccountResourceName}}",
	"yandex_resourcemanager_folder_iam_member.{{.NodeServiceAccountResourceName}}"
  ]

  name        = "{{.Name}}"
  description = "{{.Description}}"

  network_id = "${yandex_vpc_network.{{.NetworkResourceName}}.id}"

  master {
    version = "{{.MasterVersion}}" 
    zonal {
  	  zone = "${yandex_vpc_subnet.{{.SubnetResourceNameA}}.zone}" 
	  subnet_id = "${yandex_vpc_subnet.{{.SubnetResourceNameA}}.id}"
    }
  
    public_ip = true 
  }

  service_account_id = "${yandex_iam_service_account.{{.ServiceAccountResourceName}}.id}"
  node_service_account_id = "${yandex_iam_service_account.{{.NodeServiceAccountResourceName}}.id}"

  labels = {
	{{.LabelKey}} = "{{.LabelValue}}"
  }

  release_channel = "{{.ReleaseChannel}}"
}
`
const regionalClusterConfigTemplate = `
resource "yandex_kubernetes_cluster" "{{.ClusterResourceName}}" {
  depends_on         = [
	"yandex_resourcemanager_folder_iam_member.{{.ServiceAccountResourceName}}",
	"yandex_resourcemanager_folder_iam_member.{{.NodeServiceAccountResourceName}}"
  ]

  name        = "{{.Name}}"
  description = "{{.Description}}"

  network_id = "${yandex_vpc_network.{{.NetworkResourceName}}.id}"

  master {
	version = "{{.MasterVersion}}"
    regional {
  	  region = "ru-central1"
      location {
          zone = "${yandex_vpc_subnet.{{.SubnetResourceNameA}}.zone}"
          subnet_id = "${yandex_vpc_subnet.{{.SubnetResourceNameA}}.id}"
	  }
      location {
          zone = "${yandex_vpc_subnet.{{.SubnetResourceNameB}}.zone}"
          subnet_id = "${yandex_vpc_subnet.{{.SubnetResourceNameB}}.id}"
	  }
      location {
          zone = "${yandex_vpc_subnet.{{.SubnetResourceNameC}}.zone}"
          subnet_id = "${yandex_vpc_subnet.{{.SubnetResourceNameC}}.id}"
	  }
    }
  
    public_ip = true 
  }

  service_account_id = "${yandex_iam_service_account.{{.ServiceAccountResourceName}}.id}"
  node_service_account_id = "${yandex_iam_service_account.{{.NodeServiceAccountResourceName}}.id}"

  labels = {
	{{.LabelKey}} = "{{.LabelValue}}"
  }

  release_channel = "{{.ReleaseChannel}}"
}
`

const clusterDependenciesConfigTemplate = `
resource "yandex_vpc_network" "{{.NetworkResourceName}}" {
  description = "{{.TestDescription}}"
}

resource "yandex_vpc_subnet" "{{.SubnetResourceNameA}}" {
  description = "{{.TestDescription}}"
  zone = "ru-central1-a"
  network_id     = "${yandex_vpc_network.{{.NetworkResourceName}}.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "{{.SubnetResourceNameB}}" {
  description = "{{.TestDescription}}"
  zone = "ru-central1-b"
  network_id     = "${yandex_vpc_network.{{.NetworkResourceName}}.id}"
  v4_cidr_blocks = ["192.168.1.0/24"]
}

resource "yandex_vpc_subnet" "{{.SubnetResourceNameC}}" {
  description = "{{.TestDescription}}"
  zone = "ru-central1-c"
  network_id     = "${yandex_vpc_network.{{.NetworkResourceName}}.id}"
  v4_cidr_blocks = ["192.168.2.0/24"]
}

resource "yandex_iam_service_account" "{{.ServiceAccountResourceName}}" {
  name = "{{.ServiceAccountResourceName}}"
  description = "{{.TestDescription}}"
}

resource "yandex_resourcemanager_folder_iam_member" "{{.ServiceAccountResourceName}}" {
  folder_id   = "{{.FolderID}}"
  member      = "serviceAccount:${yandex_iam_service_account.{{.ServiceAccountResourceName}}.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_iam_service_account" "{{.NodeServiceAccountResourceName}}" {
  name = "{{.NodeServiceAccountResourceName}}"
  description = "{{.TestDescription}}"
}

resource "yandex_resourcemanager_folder_iam_member" "{{.NodeServiceAccountResourceName}}" {
  folder_id   = "{{.FolderID}}"
  member      = "serviceAccount:${yandex_iam_service_account.{{.NodeServiceAccountResourceName}}.id}"
  role        = "editor"
  sleep_after = 30
}
`

func testAccKubernetesClusterZonalConfig_update(orig, new resourceClusterInfo) string {
	config := templateConfig(zonalClusterConfigTemplate, new.Map()) + templateConfig(clusterDependenciesConfigTemplate, orig.Map())
	return config
}

func testAccKubernetesClusterZonalConfig_basic(in resourceClusterInfo) string {
	m := in.Map()
	config := templateConfig(clusterDependenciesConfigTemplate, m) + templateConfig(zonalClusterConfigTemplate, m)
	return config
}

func testAccKubernetesClusterRegionalConfig_update(orig, new resourceClusterInfo) string {
	config := templateConfig(regionalClusterConfigTemplate, new.Map()) + templateConfig(clusterDependenciesConfigTemplate, orig.Map())
	return config
}

func testAccKubernetesClusterRegionalConfig_basic(in resourceClusterInfo) string {
	m := in.Map()
	config := templateConfig(clusterDependenciesConfigTemplate, m) + templateConfig(regionalClusterConfigTemplate, m)
	return config
}

func testAccKubernetesClusterConfig_wrong() string {
	return `
resource "yandex_kubernetes_cluster" "this" {
  name        = "foo"
  description = "bar"

  network_id = "net-id"

  service_account_id = "sa-id"
  node_service_account_id = "node-sa-id"

  master {
    zonal {
  	  zone = "ru-central1-a" 
	  subnet_id = "subnet-id"
    }
    regional {
  	  region = "ru-central1"
      location {
          zone = "zone-a"
          subnet_id = "subnet-a"
	  }
      location {
          zone = "zone-b"
          subnet_id = "subnet-b"
	  }
      location {
          zone = "zone-c"
          subnet_id = "subnet-c"
	  }
    }
  
    public_ip = true 
  }
}
`
}

func testAccCheckKubernetesClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_kubernetes_cluster" {
			continue
		}

		_, err := config.sdk.Kubernetes().Cluster().Get(context.Background(), &k8s.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Kubernetes cluster still exists")
		}
	}

	return nil
}

func testAccCheckKubernetesClusterExists(n string, cluster *k8s.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.Kubernetes().Cluster().Get(context.Background(), &k8s.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Kubernetes cluster not found")
		}

		*cluster = *found
		return nil
	}
}
