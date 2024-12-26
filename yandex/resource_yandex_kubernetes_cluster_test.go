package yandex

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/fatih/structs"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
)

const (
	k8sTestVersion       = "1.28"
	k8sTestUpdateVersion = "1.29"
)

func init() {
	resource.AddTestSweepers("yandex_kubernetes_cluster", &resource.Sweeper{
		Name: "yandex_kubernetes_cluster",
		F:    testSweepKubernetesClusters,
		Dependencies: []string{
			"yandex_kubernetes_node_group",
			"yandex_kms_symmetric_key",
		},
	})
}

func testSweepKubernetesClusters(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	var serviceAccountID string
	var depsCreated bool

	req := &k8s.ListClustersRequest{FolderId: conf.FolderID}
	it := conf.sdk.Kubernetes().Cluster().ClusterIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		if !depsCreated {
			depsCreated = true
			serviceAccountID, err = createIAMServiceAccountForSweeper(conf)
			if err != nil {
				result = multierror.Append(result, err)
				break
			}
		}

		id := it.Value().GetId()
		if !updateKubernetesClusterWithSweeperDeps(conf, id, serviceAccountID) {
			result = multierror.Append(result,
				fmt.Errorf("failed to sweep (update with dependencies) Kubernetes Cluster %q", id))
			continue
		}

		if !sweepKubernetesCluster(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Kubernetes Cluster %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepKubernetesCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepKubernetesClusterOnce, conf, "Kubernetes Cluster", id)
}

func sweepKubernetesClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexKubernetesClusterDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Kubernetes().Cluster().Delete(ctx, &k8s.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func k8sClusterImportStep(clusterResourceFullName string, ignored ...string) resource.TestStep {
	return resource.TestStep{
		ResourceName:            clusterResourceFullName,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: ignored,
	}
}

func updateKubernetesClusterWithSweeperDeps(conf *Config, clusterID, serviceAccountID string) bool {
	debugLog("started updating Kubernetes Cluster %q", clusterID)

	client := conf.sdk.Kubernetes().Cluster()
	for i := 1; i <= conf.MaxRetries; i++ {
		req := &k8s.UpdateClusterRequest{
			ClusterId:        clusterID,
			ServiceAccountId: serviceAccountID,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{
					"service_account_id",
				},
			},
		}

		_, err := conf.sdk.WrapOperation(client.Update(conf.Context(), req))
		if err != nil {
			debugLog("[kubernetes cluster %q] update try #%d: %v", clusterID, i, err)
		} else {
			debugLog("[kubernetes cluster %q] update try #%d: request was successfully sent", clusterID, i)
			return true
		}
	}

	debugLog("[kubernetes cluster %q] update failed", clusterID)
	return false
}

//revive:disable:var-naming
func TestAccKubernetesClusterZonal_basic(t *testing.T) {
	clusterResource := clusterInfoWithNetworkPolicy("testAccKubernetesClusterZonalConfig_basic", true)
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

func TestAccKubernetesClusterZonalSecurityGroups_basic(t *testing.T) {
	clusterResource := clusterInfoWithSecurityGroups("TestAccKubernetesClusterZonalSecurityGroups_basic", true)
	clusterResourceFullName := clusterResource.ResourceFullName(true)

	clusterUpdatedResource := clusterResource
	clusterUpdatedResource.SecurityGroups = ""
	clusterUpdatedResource.TestDescription = "testAccKubernetesClusterZonalConfig_update"
	clusterUpdatedResource.MasterVersion = k8sTestUpdateVersion

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

func TestAccKubernetesClusterZonalDailyMaintenance_basic(t *testing.T) {
	clusterResource := clusterInfoWithMaintenance("TestAccKubernetesClusterZonalDailyMaintenance_basic",
		true, true, dailyMaintenancePolicy)

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

func TestAccKubernetesClusterZonalWeeklyMaintenance_basic(t *testing.T) {
	clusterResource := clusterInfoWithMaintenance("TestAccKubernetesClusterZonalWeeklyMaintenance_basic",
		true, false, weeklyMaintenancePolicy)
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
	clusterResource := clusterInfoWithNetworkPolicy("testAccKubernetesClusterRegionalConfig_basic", false)
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

func TestAccKubernetesClusterRegional_externalIPv6Address(t *testing.T) {
	clusterResource := clusterInfoExternalIPv6Address("testAccKubernetesClusterRegional_externalIPv6Address")
	clusterResourceFullName := clusterResource.ResourceFullName(true)

	var cluster k8s.Cluster

	// All external IPv6 tests share the same subnets. Disallow concurrent execution.
	mutexKV.Lock(clusterResource.SubnetResourceNameA)
	mutexKV.Lock(clusterResource.SubnetResourceNameB)
	mutexKV.Lock(clusterResource.SubnetResourceNameC)
	defer mutexKV.Unlock(clusterResource.SubnetResourceNameA)
	defer mutexKV.Unlock(clusterResource.SubnetResourceNameB)
	defer mutexKV.Unlock(clusterResource.SubnetResourceNameC)

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
	clusterUpdatedResource.MasterVersion = k8sTestUpdateVersion

	// update maintenance policy
	clusterUpdatedResource.constructMaintenancePolicyField(false, dailyMaintenancePolicy)

	// test update of weekly maintenance policy (change start time && duration, without changing the 'days')
	clusterUpdatedResource2 := clusterUpdatedResource
	clusterUpdatedResource2.constructMaintenancePolicyField(true, weeklyMaintenancePolicy)

	clusterUpdatedResource3 := clusterUpdatedResource2
	clusterUpdatedResource3.constructMaintenancePolicyField(true, weeklyMaintenancePolicySecond)

	clusterUpdatedResource4 := clusterUpdatedResource3
	clusterUpdatedResource4.constructMaintenancePolicyField(false, anyMaintenancePolicy)

	clusterUpdatedResource5 := clusterUpdatedResource4
	clusterUpdatedResource5.constructMaintenancePolicyField(true, emptyMaintenancePolicy)

	clusterUpdatedResource6 := clusterUpdatedResource5
	clusterUpdatedResource6.constructMaintenancePolicyField(true, weeklyMaintenancePolicySecond)

	clusterUpdatedResource7 := clusterUpdatedResource6
	clusterUpdatedResource7.constructMaintenancePolicyField(true, emptyMaintenancePolicy)

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
			{
				Config: testAccKubernetesClusterZonalConfig_update(clusterResource, clusterUpdatedResource2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterUpdatedResource2, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
			{
				Config: testAccKubernetesClusterZonalConfig_update(clusterResource, clusterUpdatedResource3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterUpdatedResource3, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
			{
				Config: testAccKubernetesClusterZonalConfig_update(clusterResource, clusterUpdatedResource4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterUpdatedResource4, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
			{
				Config: testAccKubernetesClusterZonalConfig_update(clusterResource, clusterUpdatedResource5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterUpdatedResource5, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
			{
				Config: testAccKubernetesClusterZonalConfig_update(clusterResource, clusterUpdatedResource6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterUpdatedResource6, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
			{
				Config: testAccKubernetesClusterZonalConfig_update(clusterResource, clusterUpdatedResource7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceFullName, &cluster),
					checkClusterAttributes(&cluster, &clusterUpdatedResource7, true),
					testAccCheckCreatedAtAttr(clusterResourceFullName),
				),
			},
		},
	})
}

func TestAccKubernetesClusterRegional_update(t *testing.T) {
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
	clusterUpdatedResource.MasterVersion = k8sTestUpdateVersion

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

func TestAccKubernetesClusterZonal_networkImplementationCilium(t *testing.T) {
	clusterResource := clusterInfo("TestAccKubernetesClusterZonal_networkImplementationCilium", true)
	clusterResourceFullName := clusterResource.ResourceFullName(true)
	clusterResource.NetworkImplementationCilium = true
	clusterResource.MasterVersion = k8sTestVersion

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

func TestAccKubernetesClusterZonal_masterLogging_defaultDestination(t *testing.T) {
	clusterResource := clusterInfo("TestAccKubernetesClusterZonal_masterLogging_defaultDestination", true)
	clusterResource.constructMasterLoggingField(masterLoggingParams{})
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

func TestAccKubernetesClusterZonal_masterLogging_toPrecreatedLogGroup(t *testing.T) {
	clusterResource := clusterInfo("TestAccKubernetesClusterZonal_masterLogging_toPrecreatedLogGroup", true)
	clusterResource.constructMasterLoggingField(masterLoggingParams{LogToPrecreatedLogGroup: true})
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

func TestAccKubernetesClusterRegional_masterLogging_defaultDestination(t *testing.T) {
	clusterResource := clusterInfo("TestAccKubernetesClusterRegional_masterLogging_defaultDestination", false)
	clusterResource.constructMasterLoggingField(masterLoggingParams{})
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
		},
	})
}

func TestAccKubernetesClusterRegional_masterLogging_toPrecreatedLogGroup(t *testing.T) {
	clusterResource := clusterInfo("TestAccKubernetesClusterRegional_masterLogging_toPrecreatedLogGroup", false)
	clusterResource.constructMasterLoggingField(masterLoggingParams{LogToPrecreatedLogGroup: true})
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
		},
	})
}

// log_group_id and folder_id must not be set simultaneously
func TestAccKubernetesCluster_masterLogging_wrong(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccKubernetesClusterConfig_masterLogging_wrong(),
				ExpectError: regexp.MustCompile("Conflicting configuration arguments"),
			},
		},
	})
}

type masterLoggingParams struct {
	LogToPrecreatedLogGroup bool
}

func randomResourceName(tp string) string {
	return fmt.Sprintf("test_%s_%s", tp, acctest.RandString(10))
}

// iam uses strict validation for SA names
func safeResourceName(tp string) string {
	return fmt.Sprintf("test%s%s", tp, acctest.RandString(10))
}

func clusterInfo(testDesc string, zonal bool) resourceClusterInfo {
	return clusterInfoWithMaintenance(testDesc, zonal, true, anyMaintenancePolicy)
}

// fillSharedEmptyResourceClusterInfoParams ensures that some cluster info parameters has fake values if they're empty.
// It is needed for shared resources that are locked by mutexKV, so we don't have locks on the same empty string in unit
// tests where those env variables are not set.
func fillSharedEmptyResourceClusterInfoParams(ci resourceClusterInfo) resourceClusterInfo {
	if ci.SubnetResourceNameA == "" {
		ci.SubnetResourceNameA = "SubnetResourceNameA"
	}
	if ci.SubnetResourceNameB == "" {
		ci.SubnetResourceNameB = "SubnetResourceNameB"
	}
	if ci.SubnetResourceNameC == "" {
		ci.SubnetResourceNameC = "SubnetResourceNameC"
	}
	return ci
}

func clusterInfoDualStack(testDesc string, zonal bool) resourceClusterInfo {
	ci := clusterInfo(testDesc, zonal)

	// Use existing resources rather than creating new ones for dual stack clusters.
	//
	// K8S_SUBNET_NETWORK_ID - network with dual-stack subnets without external ivp6 for k8s cluster and node-groups.
	// K8S_SUBNET_A_ID - dual-stack subnet without external ivp6 in location "a" for k8s cluster and node-groups.
	// K8S_SUBNET_B_ID - dual-stack subnet without external ivp6 in location "b" for k8s cluster and node-groups.
	// K8S_SUBNET_C_ID - dual-stack subnet without external ivp6 in location "c" for k8s cluster and node-groups.
	// K8S_SECURITY_GROUP_ID - security-group of K8S_SUBNET_NETWORK_ID that can be used to test k8s cluster with security-group.
	// K8S_NETWORK_FOLDER_ID - folder with dual-stack resources to add folder iam member to use those resources.
	ci.NetworkResourceName = os.Getenv("K8S_SUBNET_NETWORK_ID")
	ci.SubnetResourceNameA = os.Getenv("K8S_SUBNET_A_ID")
	ci.SubnetResourceNameB = os.Getenv("K8S_SUBNET_B_ID")
	ci.SubnetResourceNameC = os.Getenv("K8S_SUBNET_C_ID")
	ci.SecurityGroupName = os.Getenv("K8S_SECURITY_GROUP_ID")
	ci.NetworkFolderID = os.Getenv("K8S_NETWORK_FOLDER_ID")
	ci.ClusterIPv6Range = "fc00::/96"
	ci.ServiceIPv6Range = "fc01::/112"
	ci.ClusterIPv4Range = "10.20.0.0/16"
	ci.ServiceIPv4Range = "10.21.0.0/16"
	ci.DualStack = true

	return fillSharedEmptyResourceClusterInfoParams(ci)
}

func clusterInfoExternalIPv6Address(testDesc string) resourceClusterInfo {
	ci := clusterInfo(testDesc, false)

	// Use existing resources rather than creating new ones for clusters with external ipv6 address.
	//
	// K8S_E2E_IPV6_SUBNET_NETWORK_ID - network with dual-stack external ivp6 subnets for k8s cluster and node-groups.
	// K8S_E2E_IPV6_SUBNET_A_ID - dual-stack subnet with external ivp6 in location "a" for k8s cluster and node-groups.
	// K8S_E2E_IPV6_SUBNET_B_ID - dual-stack subnet with external ivp6 in location "b" for k8s cluster and node-groups.
	// K8S_E2E_IPV6_SUBNET_C_ID - dual-stack subnet with external ivp6 in location "c" for k8s cluster and node-groups.
	// K8S_E2E_IPV6_SECURITY_GROUP_ID - security-group of K8S_E2E_IPV6_SUBNET_NETWORK_ID that can be used to test k8s cluster with security-group.
	// K8S_E2E_IPV6_ADDRESS - IPv6 address that can be used as K8S cluster endpoint.
	// K8S_NETWORK_FOLDER_ID - folder with dual-stack external ivp6 resources to add folder iam member to use those resources.
	ci.NetworkResourceName = os.Getenv("K8S_E2E_IPV6_SUBNET_NETWORK_ID")
	ci.SubnetResourceNameA = os.Getenv("K8S_E2E_IPV6_SUBNET_A_ID")
	ci.SubnetResourceNameB = os.Getenv("K8S_E2E_IPV6_SUBNET_B_ID")
	ci.SubnetResourceNameC = os.Getenv("K8S_E2E_IPV6_SUBNET_C_ID")
	ci.SecurityGroupName = os.Getenv("K8S_E2E_IPV6_SECURITY_GROUP_ID")
	ci.ExternalIPv6Address = os.Getenv("K8S_E2E_IPV6_ADDRESS")
	ci.NetworkFolderID = os.Getenv("K8S_NETWORK_FOLDER_ID")
	ci.ClusterIPv4Range = "10.8.0.0/16"
	ci.ServiceIPv4Range = "10.9.0.0/16"

	return fillSharedEmptyResourceClusterInfoParams(ci)
}

func clusterInfoWithMaintenance(testDesc string, zonal bool, autoUpgrade bool, policyType maintenancePolicyType) resourceClusterInfo {
	res := resourceClusterInfo{
		ClusterResourceName:            randomResourceName("cluster"),
		FolderID:                       getExampleFolderID(),
		Name:                           safeResourceName("clustername"),
		Description:                    "description",
		MasterVersion:                  k8sTestVersion,
		LabelKey:                       "label_key",
		LabelValue:                     "label_value",
		TestDescription:                testDesc,
		NetworkResourceName:            randomResourceName("network"),
		SubnetResourceNameA:            randomResourceName("subnet"),
		SubnetResourceNameB:            randomResourceName("subnet"),
		SubnetResourceNameC:            randomResourceName("subnet"),
		SecurityGroupName:              randomResourceName("sg"),
		ServiceAccountResourceName:     safeResourceName("serviceaccount"),
		NodeServiceAccountResourceName: safeResourceName("nodeserviceaccount"),
		ReleaseChannel:                 k8s.ReleaseChannel_RAPID.String(),
		zonal:                          zonal,
		KMSKeyResourceName:             randomResourceName("key"),
	}

	res.constructMaintenancePolicyField(autoUpgrade, policyType)
	return res
}

func clusterInfoWithNetworkPolicy(testDesc string, zonal bool) resourceClusterInfo {
	res := clusterInfo(testDesc, zonal)
	res.constructNetworkPolicyField(k8s.NetworkPolicy_CALICO)
	return res
}

func clusterInfoWithSecurityGroups(testDesc string, zonal bool) resourceClusterInfo {
	res := clusterInfo(testDesc, zonal)
	res.constructSecurityGroupsField()
	return res
}

func clusterInfoWithSecurityGroupsNetworkAndMaintenancePolicies(testDesc string, zonal bool, autoUpgrade bool, policyType maintenancePolicyType) resourceClusterInfo {
	res := clusterInfoWithMaintenance(testDesc, zonal, autoUpgrade, policyType)
	res.constructNetworkPolicyField(k8s.NetworkPolicy_CALICO)
	res.constructSecurityGroupsField()

	return res
}

type clusterResourceIDs struct {
	networkResourceID            string
	subnetAResourceID            string
	subnetBResourceID            string
	subnetCResourceID            string
	securityGroupID              string
	serviceAccountResourceID     string
	nodeServiceAccountResourceID string
	kmsKeyResourceID             string
}

func clusterVPCDepsPrecreated(info *resourceClusterInfo) bool {
	return info.DualStack || info.ExternalIPv6Address != ""
}

func getClusterResourcesIds(s *terraform.State, info *resourceClusterInfo) (ids clusterResourceIDs, err error) {
	if !clusterVPCDepsPrecreated(info) {
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

		ids.securityGroupID, err = getResourceID(info.securityGroupName(), s)
		if err != nil {
			return
		}
	}

	ids.serviceAccountResourceID, err = getResourceID(info.serviceAccountResourceName(), s)
	if err != nil {
		return
	}

	ids.nodeServiceAccountResourceID, err = getResourceID(info.nodeServiceAccountResourceName(), s)
	if err != nil {
		return
	}

	ids.kmsKeyResourceID, err = getResourceID(info.kmsKeyResourceName(), s)
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

		if info.NetworkImplementationCilium == true && cluster.GetNetworkImplementation() == nil {
			return fmt.Errorf("expected network implementation, but got none")
		}

		if !info.zonal && regionalMaster == nil {
			return fmt.Errorf("expected regional cluster, but got zonal")
		}

		resourceFullName := info.ResourceFullName(rs)
		checkFuncsAr := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(resourceFullName, "service_account_id", ids.serviceAccountResourceID),
			resource.TestCheckResourceAttr(resourceFullName, "node_service_account_id", ids.nodeServiceAccountResourceID),

			resource.TestCheckResourceAttr(resourceFullName, "name", info.Name),
			resource.TestCheckResourceAttr(resourceFullName, "description", info.Description),

			resource.TestCheckResourceAttr(resourceFullName, "release_channel", info.ReleaseChannel),
			resource.TestCheckResourceAttr(resourceFullName, "release_channel", cluster.ReleaseChannel.String()),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.version_info.0.current_version", versionInfo.GetCurrentVersion()),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.version_info.0.new_revision_available", strconv.FormatBool(versionInfo.GetNewRevisionAvailable())),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.version_info.0.new_revision_summary", versionInfo.GetNewRevisionSummary()),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.version_info.0.version_deprecated", strconv.FormatBool(versionInfo.GetVersionDeprecated())),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.maintenance_policy.0.auto_upgrade", strconv.FormatBool(master.GetMaintenancePolicy().GetAutoUpgrade())),

			resource.TestCheckResourceAttr(resourceFullName, "name", cluster.Name),
			resource.TestCheckResourceAttr(resourceFullName, "description", cluster.Description),
			resource.TestCheckResourceAttr(resourceFullName, "service_account_id", cluster.ServiceAccountId),
			resource.TestCheckResourceAttr(resourceFullName, "node_service_account_id", cluster.NodeServiceAccountId),
			resource.TestCheckResourceAttr(resourceFullName, "network_id", cluster.NetworkId),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.internal_v4_endpoint", master.GetEndpoints().GetInternalV4Endpoint()),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.external_v4_endpoint", master.GetEndpoints().GetExternalV4Endpoint()),
			resource.TestCheckResourceAttr(resourceFullName, "master.0.cluster_ca_certificate", master.GetMasterAuth().GetClusterCaCertificate()),
			resource.TestCheckResourceAttr(resourceFullName, "kms_provider.0.key_id", ids.kmsKeyResourceID),
			resource.TestCheckResourceAttr(resourceFullName, "kms_provider.0.key_id", cluster.GetKmsProvider().GetKeyId()),
			testAccCheckClusterLabel(cluster, info, rs),

			resource.TestCheckResourceAttr(resourceFullName,
				"cluster_ipv4_range", cluster.GetIpAllocationPolicy().ClusterIpv4CidrBlock),
			resource.TestCheckResourceAttr(resourceFullName,
				"cluster_ipv6_range", cluster.GetIpAllocationPolicy().ClusterIpv6CidrBlock),
			resource.TestCheckResourceAttr(resourceFullName,
				"node_ipv4_cidr_mask_size", strconv.Itoa(int(cluster.GetIpAllocationPolicy().GetNodeIpv4CidrMaskSize()))),
			resource.TestCheckResourceAttr(resourceFullName,
				"service_ipv4_range", cluster.GetIpAllocationPolicy().GetServiceIpv4CidrBlock()),
			resource.TestCheckResourceAttr(resourceFullName,
				"service_ipv6_range", cluster.GetIpAllocationPolicy().GetServiceIpv6CidrBlock()),
		}

		if !clusterVPCDepsPrecreated(info) {
			resource.TestCheckResourceAttr(resourceFullName, "network_id", ids.networkResourceID)
		}
		if info.ExternalIPv6Address != "" {
			resource.TestCheckResourceAttr(resourceFullName, "master.0.external_v6_endpoint", master.GetEndpoints().GetExternalV6Endpoint())
		}

		if networkImplementation := cluster.GetNetworkImplementation(); networkImplementation != nil {
			switch networkImplementation.(type) {
			case *k8s.Cluster_Cilium:
				resource.TestCheckResourceAttrSet(resourceFullName, "network_implementation.0.cilium.0")
			}
		}

		if info.SecurityGroups != "" && !clusterVPCDepsPrecreated(info) {
			resource.TestCheckResourceAttr(resourceFullName, "master.0.security_groups_ids.0", ids.securityGroupID)
		} else {
			resource.TestCheckResourceAttr(resourceFullName, "master.0.security_groups_ids.#", "0")
		}

		if info.policy != emptyMaintenancePolicy {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "master.0.maintenance_policy.0.auto_upgrade", strconv.FormatBool(info.autoUpgrade)))
		}

		if npp := info.networkPolicyProvider; npp != k8s.NetworkPolicy_PROVIDER_UNSPECIFIED {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "network_policy_provider", npp.String()),
				resource.TestCheckResourceAttr(resourceFullName, "network_policy_provider", cluster.GetNetworkPolicy().GetProvider().String()),
			)
		}

		const maintenanceWindowPrefix = "master.0.maintenance_policy.0.maintenance_window."
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
			if info.ExternalIPv6Address != "" {
				resource.TestCheckResourceAttr(resourceFullName, "master.0.external_v6_address",
					regionalMaster.GetExternalV6Address())
			}

			if rs {
				checkFuncsAr = append(checkFuncsAr,
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.0.zone",
						"ru-central1-a"),
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.1.zone",
						"ru-central1-b"),
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.2.zone",
						"ru-central1-d"),
				)

				if !clusterVPCDepsPrecreated(info) {
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.0.subnet_id",
						ids.subnetAResourceID)
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.1.subnet_id",
						ids.subnetBResourceID)
					resource.TestCheckResourceAttr(resourceFullName, "master.0.regional.0.location.2.subnet_id",
						ids.subnetCResourceID)
				}
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

		if info.MasterLogging != "" {
			checkFuncsAr = append(checkFuncsAr,
				resource.TestCheckResourceAttr(resourceFullName, "master.0.master_logging.0.enabled", fmt.Sprintf("%t", master.GetMasterLogging().GetEnabled())),
				resource.TestCheckResourceAttr(resourceFullName, "master.0.master_logging.0.log_group_id", master.GetMasterLogging().GetLogGroupId()),
				resource.TestCheckResourceAttr(resourceFullName, "master.0.master_logging.0.folder_id", master.GetMasterLogging().GetFolderId()),
				resource.TestCheckResourceAttr(resourceFullName, "master.0.master_logging.0.kube_apiserver_enabled", fmt.Sprintf("%t", master.GetMasterLogging().GetKubeApiserverEnabled())),
				resource.TestCheckResourceAttr(resourceFullName, "master.0.master_logging.0.cluster_autoscaler_enabled", fmt.Sprintf("%t", master.GetMasterLogging().GetClusterAutoscalerEnabled())),
				resource.TestCheckResourceAttr(resourceFullName, "master.0.master_logging.0.events_enabled", fmt.Sprintf("%t", master.GetMasterLogging().GetEventsEnabled())),
				resource.TestCheckResourceAttr(resourceFullName, "master.0.master_logging.0.audit_enabled", fmt.Sprintf("%t", master.GetMasterLogging().GetAuditEnabled())),
			)
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

func errorResourceCheckFunc(err error) resource.TestCheckFunc {
	return func(*terraform.State) error {
		return err
	}
}

func testAccCheckMaintenanceWindow(resourceFullName string, maintenanceWindowPrefix string, day, startTime, duration string) resource.TestCheckFunc {
	st, err := parseDayTime(startTime)
	if err != nil {
		return errorResourceCheckFunc(err)
	}

	du, err := parseDuration(duration)
	if err != nil {
		return errorResourceCheckFunc(err)
	}

	// can't use shouldSuppressDiffFor function here, thus, using regexp, to match either value
	// from config (resources tests) or value from api (datasource tests)
	m := map[string]*regexp.Regexp{
		"day":        regexp.MustCompile(fmt.Sprintf("\\Q%v\\E", day)),
		"start_time": regexp.MustCompile(fmt.Sprintf("\\Q%v\\E|\\Q%v\\E", startTime, st)),
		"duration":   regexp.MustCompile(fmt.Sprintf("\\Q%v\\E|\\Q%v\\E", duration, du)),
	}
	return resource.TestMatchTypeSetElemNestedAttrs(resourceFullName, maintenanceWindowPrefix+"*", m)
}

type maintenancePolicyType int

const (
	anyMaintenancePolicy          maintenancePolicyType = 0
	dailyMaintenancePolicy        maintenancePolicyType = 1
	weeklyMaintenancePolicy       maintenancePolicyType = 2
	weeklyMaintenancePolicySecond maintenancePolicyType = 3
	emptyMaintenancePolicy        maintenancePolicyType = 4
)

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

	MaintenancePolicy string

	NetworkPolicy         string
	networkPolicyProvider k8s.NetworkPolicy_Provider

	zonal bool

	autoUpgrade bool
	policy      maintenancePolicyType

	KMSKeyResourceName string

	SecurityGroups    string
	SecurityGroupName string

	ClusterIPv4Range string
	ClusterIPv6Range string
	ServiceIPv4Range string
	ServiceIPv6Range string

	// For dual stack clusters and clusters with external ipv6.
	NetworkFolderID     string
	ExternalIPv6Address string
	DualStack           bool

	NetworkImplementationCilium bool

	// folder_id is not tested here since folder deletion takes too long
	MasterLogging                     string
	MasterLoggingLogGroupResourceName string
}

func (i *resourceClusterInfo) constructMaintenancePolicyField(autoUpgrade bool, policy maintenancePolicyType) {
	m := map[string]interface{}{
		"AutoUpgrade": autoUpgrade,
	}

	i.autoUpgrade = autoUpgrade
	i.policy = policy

	switch policy {
	case emptyMaintenancePolicy:
		i.MaintenancePolicy = ""
	case anyMaintenancePolicy:
		i.MaintenancePolicy = templateConfig(anyMaintenancePolicyTemplate, m)
	case dailyMaintenancePolicy:
		i.MaintenancePolicy = templateConfig(dailyMaintenancePolicyTemplate, m)
	case weeklyMaintenancePolicy:
		i.MaintenancePolicy = templateConfig(weeklyMaintenancePolicyTemplate, m)
	case weeklyMaintenancePolicySecond:
		i.MaintenancePolicy = templateConfig(weeklyMaintenancePolicyTemplateSecond, m)
	}
}

func (i *resourceClusterInfo) constructNetworkPolicyField(npp k8s.NetworkPolicy_Provider) {
	if npp != k8s.NetworkPolicy_PROVIDER_UNSPECIFIED {
		i.networkPolicyProvider = npp
		i.NetworkPolicy = fmt.Sprintf("network_policy_provider = \"%s\"", npp.String())
	}
}

func (i *resourceClusterInfo) constructMasterLoggingField(params masterLoggingParams) {
	ctx := map[string]interface{}{}
	if params.LogToPrecreatedLogGroup {
		logGroupName := safeResourceName("loggroup")
		i.MasterLoggingLogGroupResourceName = logGroupName
		ctx["MasterLoggingLogGroupResourceName"] = logGroupName
	}
	i.MasterLogging = templateConfig(masterLoggingTemplate, ctx)
}

func (i *resourceClusterInfo) constructSecurityGroupsField() {
	i.SecurityGroups = fmt.Sprintf("security_group_ids = [\"${yandex_vpc_security_group.%s.id}\"]", i.SecurityGroupName)
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

func (i *resourceClusterInfo) securityGroupName() string {
	return "yandex_vpc_security_group." + i.SecurityGroupName
}

func (i *resourceClusterInfo) serviceAccountResourceName() string {
	return "yandex_iam_service_account." + i.ServiceAccountResourceName
}

func (i *resourceClusterInfo) nodeServiceAccountResourceName() string {
	return "yandex_iam_service_account." + i.NodeServiceAccountResourceName
}

func (i *resourceClusterInfo) kmsKeyResourceName() string {
	return "yandex_kms_symmetric_key." + i.KMSKeyResourceName
}

const anyMaintenancePolicyTemplate = `
	maintenance_policy {
        auto_upgrade = {{.AutoUpgrade}}
    }
`

const dailyMaintenancePolicyTemplate = `
	maintenance_policy {
        auto_upgrade = {{.AutoUpgrade}}
        
        maintenance_window {
			start_time = "15:00"
			duration   = "3h"
		}
    }
`

const weeklyMaintenancePolicyTemplate = `
	maintenance_policy {
        auto_upgrade = {{.AutoUpgrade}}
        
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

// used to test update for start time and duration, without changing of 'days'
const weeklyMaintenancePolicyTemplateSecond = `
	maintenance_policy {
        auto_upgrade = {{.AutoUpgrade}}

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

const masterLoggingTemplate = `
	master_logging {
		enabled = true
		kube_apiserver_enabled = true
		cluster_autoscaler_enabled = true
		events_enabled = true
		audit_enabled = true
{{if .MasterLoggingLogGroupResourceName}}
		log_group_id = yandex_logging_group.{{.MasterLoggingLogGroupResourceName}}.id
{{end}}
	}
`

const zonalClusterConfigTemplate = `
resource "yandex_kubernetes_cluster" "{{.ClusterResourceName}}" {
  depends_on         = [
	"yandex_resourcemanager_folder_iam_member.{{.ServiceAccountResourceName}}",
{{if .DualStack}}
	"yandex_resourcemanager_folder_iam_member.{{.ServiceAccountResourceName}}_existing_network",
{{end}}
	"yandex_resourcemanager_folder_iam_member.{{.NodeServiceAccountResourceName}}"
  ]

  name        = "{{.Name}}"
  description = "{{.Description}}"

{{if .DualStack}}
  network_id = "{{.NetworkResourceName}}"
{{else}}
  network_id = "${yandex_vpc_network.{{.NetworkResourceName}}.id}"
{{end}}

  master {
    version = "{{.MasterVersion}}" 
    zonal {
{{if .DualStack}}
  	  zone = "ru-central1-a"
	  subnet_id = "{{.SubnetResourceNameA}}"
{{else}}
  	  zone = "${yandex_vpc_subnet.{{.SubnetResourceNameA}}.zone}" 
	  subnet_id = "${yandex_vpc_subnet.{{.SubnetResourceNameA}}.id}"
{{end}}
    }
  
    public_ip = true
    
    {{.MaintenancePolicy}}

    {{.SecurityGroups}}

	{{.MasterLogging}}
  }

  service_account_id = "${yandex_iam_service_account.{{.ServiceAccountResourceName}}.id}"
  node_service_account_id = "${yandex_iam_service_account.{{.NodeServiceAccountResourceName}}.id}"

  labels = {
	{{.LabelKey}} = "{{.LabelValue}}"
  }

  release_channel = "{{.ReleaseChannel}}"

  {{.NetworkPolicy}}

  kms_provider {
    key_id = "${yandex_kms_symmetric_key.{{.KMSKeyResourceName}}.id}"
  }
  {{if .ClusterIPv4Range}}
  cluster_ipv4_range = "{{.ClusterIPv4Range}}"
  {{end}}
  {{if .ClusterIPv6Range}}
  cluster_ipv6_range = "{{.ClusterIPv6Range}}"
  {{end}}
  {{if .ServiceIPv4Range}}
  service_ipv4_range = "{{.ServiceIPv4Range}}"
  {{end}}
  {{if .ServiceIPv6Range}}
  service_ipv6_range = "{{.ServiceIPv6Range}}"
  {{end}}

  {{if .NetworkImplementationCilium}}
  network_implementation {
    cilium {
    }
  }
  {{end}}
}
`

const regionalClusterConfigTemplate = `
resource "yandex_kubernetes_cluster" "{{.ClusterResourceName}}" {
  depends_on         = [
    "yandex_resourcemanager_folder_iam_member.{{.ServiceAccountResourceName}}",
{{if or .DualStack .ExternalIPv6Address}}
    "yandex_resourcemanager_folder_iam_member.{{.ServiceAccountResourceName}}_existing_network",
{{end}}
    "yandex_resourcemanager_folder_iam_member.{{.NodeServiceAccountResourceName}}"
  ]

  name        = "{{.Name}}"
  description = "{{.Description}}"
{{if or .DualStack .ExternalIPv6Address}}
  network_id = "{{.NetworkResourceName}}"
{{else}}
  network_id = "${yandex_vpc_network.{{.NetworkResourceName}}.id}"
{{end}}

  master {
    version = "{{.MasterVersion}}"
    regional {
      region = "ru-central1"
{{if or .DualStack .ExternalIPv6Address}}
      location {
        zone = "ru-central1-a"
        subnet_id = "{{.SubnetResourceNameA}}"
      }
      location {
        zone = "ru-central1-b"
        subnet_id = "{{.SubnetResourceNameB}}"
      }
      location {
        zone = "ru-central1-d"
        subnet_id = "{{.SubnetResourceNameC}}"
      }
{{else}}
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
{{end}}
    }
  
    public_ip = true

    {{.SecurityGroups}}

    {{.MaintenancePolicy}}

    {{.MasterLogging}}

{{if .ExternalIPv6Address}}
    external_v6_address = "{{.ExternalIPv6Address}}"
{{end}}
  }

  service_account_id = "${yandex_iam_service_account.{{.ServiceAccountResourceName}}.id}"
  node_service_account_id = "${yandex_iam_service_account.{{.NodeServiceAccountResourceName}}.id}"

  labels = {
	{{.LabelKey}} = "{{.LabelValue}}"
  }

  release_channel = "{{.ReleaseChannel}}"

  {{.NetworkPolicy}}

  kms_provider {
    key_id = "${yandex_kms_symmetric_key.{{.KMSKeyResourceName}}.id}"
  }
}
`

const clusterDependenciesConfigTemplate = `
{{if or .DualStack .ExternalIPv6Address}}
// Use existing infrastructure for dual stack clusters or clusters with external ipv6 address.
{{else}}
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
  zone = "ru-central1-d"
  network_id     = "${yandex_vpc_network.{{.NetworkResourceName}}.id}"
  v4_cidr_blocks = ["192.168.2.0/24"]
}

resource "yandex_vpc_default_security_group" "default-network-sg" {
  description = "{{.TestDescription}}"
  network_id  = "${yandex_vpc_network.{{.NetworkResourceName}}.id}"

  ingress {
      protocol = "ANY"
      v4_cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
      protocol = "ANY"
      v4_cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "yandex_vpc_security_group" "{{.SecurityGroupName}}" {
  description = "{{.TestDescription}}"
  network_id  = "${yandex_vpc_network.{{.NetworkResourceName}}.id}"

  ingress {
      protocol = "ANY"
      v4_cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
      protocol = "ANY"
      v4_cidr_blocks = ["0.0.0.0/0"]
  }
}
{{end}}

resource "yandex_iam_service_account" "{{.ServiceAccountResourceName}}" {
  name = "{{.ServiceAccountResourceName}}"
  description = "{{.TestDescription}}"
}

{{if or .DualStack .ExternalIPv6Address}}
resource "yandex_resourcemanager_folder_iam_member" "{{.ServiceAccountResourceName}}_existing_network" {
  folder_id   = "{{.NetworkFolderID}}"
  member      = "serviceAccount:${yandex_iam_service_account.{{.ServiceAccountResourceName}}.id}"
  role        = "editor"
  sleep_after = 30
}
{{end}}

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

resource "yandex_kms_symmetric_key" "{{.KMSKeyResourceName}}" {
  name        = "{{.KMSKeyResourceName}}"
  description = "{{.TestDescription}}"
}

{{if .MasterLoggingLogGroupResourceName}}
resource "yandex_logging_group" "{{.MasterLoggingLogGroupResourceName}}" {
  name      = "{{.MasterLoggingLogGroupResourceName}}"
  folder_id = "{{.FolderID}}"
}
{{end}}
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

func testAccKubernetesClusterConfig_masterLogging_wrong() string {
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

    master_logging {
      enabled = true
      log_group_id = "log-group-id"
      folder_id = "folder-id"
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
