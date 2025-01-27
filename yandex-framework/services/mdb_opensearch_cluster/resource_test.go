package mdb_opensearch_cluster_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	pc "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/plancheck"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	openSearchResourcePrefix                = "yandex_mdb_opensearch_cluster."
	yandexMDBOpenSearchClusterDeleteTimeout = 30 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_mdb_opensearch_cluster", &resource.Sweeper{
		Name: "yandex_mdb_opensearch_cluster",
		F:    testSweepMDBOpenSearchCluster,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepMDBOpenSearchCluster(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.SDK.MDB().OpenSearch().Cluster().List(
		context.Background(),
		&opensearch.ListClustersRequest{
			FolderId: conf.ProviderState.FolderID.ValueString(),
		})
	if err != nil {
		return fmt.Errorf("error getting OpenSearch clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBOpenSearchCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep OpenSearch cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBOpenSearchCluster(conf *config.Config, id string) bool {
	return test.SweepWithRetry(sweepMDBOpenSearchClusterOnce, conf, "OpenSearch cluster", id)
}

func sweepMDBOpenSearchClusterOnce(conf *config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexMDBOpenSearchClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.SDK.MDB().OpenSearch().Cluster().Update(ctx, &opensearch.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = test.HandleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(test.ErrorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.SDK.MDB().OpenSearch().Cluster().Delete(ctx, &opensearch.DeleteClusterRequest{
		ClusterId: id,
	})
	return test.HandleSweepOperation(ctx, conf, op, err)
}

func mdbOpenSearchClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"health",                // volatile value
			"config.admin_password", // not importable
		},
	}

}

func TestAccMDBOpenSearchCluster_single(t *testing.T) {
	var r opensearch.Cluster
	openSearchName := acctest.RandomWithPrefix("tf-opensearch-single")
	openSearchDesc := "OpenSearch Cluster Terraform Test"
	randInt := acctest.RandInt()
	folderID := test.GetExampleFolderID()
	openSearchResource := openSearchResourcePrefix + openSearchName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBOpenSearchClusterDestroy,
		Steps: []resource.TestStep{
			// Create OpenSearch Cluster
			{
				Config: testSingleAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 1),
					resource.TestCheckResourceAttr(openSearchResource, "name", openSearchName),
					resource.TestCheckResourceAttr(openSearchResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(openSearchResource, "description", openSearchDesc),
					resource.TestCheckResourceAttr(openSearchResource, "config.admin_password", "password"),
					resource.TestCheckResourceAttrSet(openSearchResource, "service_account_id"),
					resource.TestCheckResourceAttr(openSearchResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(openSearchResource, "hosts.#", "1"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.0.fqdn"),
					test.AccCheckCreatedAtAttr(openSearchResource),
					testAccCheckMDBOpenSearchSubnetsAndZonesCount(&r, 3),
					testAccCheckMDBOpenSearchClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckMDBOpenSearchClusterDataNodeHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.day", "FRI"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.hour", "20"),
				),
			},
			// update AuthSettings
			{
				Config: testSamlAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", randInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 2),
					resource.TestCheckResourceAttr(openSearchResource, "name", openSearchName),
					resource.TestCheckResourceAttr(openSearchResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(openSearchResource, "description", openSearchDesc),
					resource.TestCheckResourceAttr(openSearchResource, "config.admin_password", "password"),
					resource.TestCheckResourceAttrSet(openSearchResource, "service_account_id"),
					resource.TestCheckResourceAttr(openSearchResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(openSearchResource, "hosts.#", "2"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.0.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.1.fqdn"),
					test.AccCheckCreatedAtAttr(openSearchResource),
					testAccCheckMDBOpenSearchSubnetsAndZonesCount(&r, 3),
					testAccCheckMDBOpenSearchClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckMDBOpenSearchClusterDataNodeHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.day", "FRI"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.hour", "20"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.enabled", "true"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.idp_entity_id", "https://test_identity_provider.com"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.idp_metadata_file_content", "<EntityDescriptor entityID=\"https://test_identity_provider.com\"></EntityDescriptor>"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.sp_entity_id", "https://some.db.yandex.net"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.dashboards_url", "https://dashboards.example.com"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(openSearchResource, tfjsonpath.New("hosts")),
					},
				},
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
		},
	})
}

func TestAccMDBOpenSearchCluster_simple(t *testing.T) {
	var r opensearch.Cluster
	openSearchName := acctest.RandomWithPrefix("tf-opensearch-simple")
	openSearchDesc := "OpenSearch Cluster Terraform Test"
	randInt := acctest.RandInt()
	folderID := test.GetExampleFolderID()
	openSearchResource := openSearchResourcePrefix + openSearchName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBOpenSearchClusterDestroy,
		Steps: []resource.TestStep{
			// Create OpenSearch Cluster
			{
				Config: testSimpleAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 5),
					resource.TestCheckResourceAttr(openSearchResource, "name", openSearchName),
					resource.TestCheckResourceAttr(openSearchResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(openSearchResource, "description", openSearchDesc),
					resource.TestCheckResourceAttr(openSearchResource, "config.admin_password", "password"),
					resource.TestCheckResourceAttrSet(openSearchResource, "service_account_id"),
					resource.TestCheckResourceAttr(openSearchResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(openSearchResource, "hosts.#", "5"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.0.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.1.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.2.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.3.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.4.fqdn"),
					test.AccCheckCreatedAtAttr(openSearchResource),
					testAccCheckMDBOpenSearchSubnetsAndZonesCount(&r, 3),
					testAccCheckMDBOpenSearchClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckMDBOpenSearchClusterDataNodeHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.day", "FRI"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.hour", "20"),
				),
			},
		},
	})
}

func TestAccMDBOpenSearchCluster_saml(t *testing.T) {
	var r opensearch.Cluster
	openSearchName := acctest.RandomWithPrefix("tf-opensearch-saml")
	openSearchDesc := "OpenSearch Cluster Terraform Test"
	randInt := acctest.RandInt()
	folderID := test.GetExampleFolderID()
	openSearchResource := openSearchResourcePrefix + openSearchName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBOpenSearchClusterDestroy,
		Steps: []resource.TestStep{
			// Create OpenSearch Cluster with enabled saml auth
			{
				Config: testSamlAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", randInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 2),
					resource.TestCheckResourceAttr(openSearchResource, "name", openSearchName),
					resource.TestCheckResourceAttr(openSearchResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(openSearchResource, "description", openSearchDesc),
					resource.TestCheckResourceAttr(openSearchResource, "config.admin_password", "password"),
					resource.TestCheckResourceAttrSet(openSearchResource, "service_account_id"),
					resource.TestCheckResourceAttr(openSearchResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(openSearchResource, "hosts.#", "2"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.0.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.1.fqdn"),
					test.AccCheckCreatedAtAttr(openSearchResource),
					testAccCheckMDBOpenSearchSubnetsAndZonesCount(&r, 3),
					testAccCheckMDBOpenSearchClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckMDBOpenSearchClusterDataNodeHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckMDBOpenSearchClusterDashboardsHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.day", "FRI"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.hour", "20"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.enabled", "true"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.idp_entity_id", "https://test_identity_provider.com"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.idp_metadata_file_content", "<EntityDescriptor entityID=\"https://test_identity_provider.com\"></EntityDescriptor>"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.sp_entity_id", "https://some.db.yandex.net"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.dashboards_url", "https://dashboards.example.com"),
				),
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
			// disable saml auth
			{
				Config: testSamlAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", randInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 2),
					resource.TestCheckResourceAttr(openSearchResource, "name", openSearchName),
					resource.TestCheckResourceAttr(openSearchResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(openSearchResource, "description", openSearchDesc),
					resource.TestCheckResourceAttr(openSearchResource, "config.admin_password", "password"),
					resource.TestCheckResourceAttrSet(openSearchResource, "service_account_id"),
					resource.TestCheckResourceAttr(openSearchResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(openSearchResource, "hosts.#", "2"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.0.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.1.fqdn"),
					test.AccCheckCreatedAtAttr(openSearchResource),
					testAccCheckMDBOpenSearchSubnetsAndZonesCount(&r, 3),
					testAccCheckMDBOpenSearchClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckMDBOpenSearchClusterDataNodeHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckMDBOpenSearchClusterDashboardsHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.day", "FRI"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.hour", "20"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.enabled", "false"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.idp_entity_id", "https://test_identity_provider.com"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.idp_metadata_file_content", "<EntityDescriptor entityID=\"https://test_identity_provider.com\"></EntityDescriptor>"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.sp_entity_id", "https://some.db.yandex.net"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.dashboards_url", "https://dashboards.example.com"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						pc.ExpectNoChangesAt(openSearchResource, tfjsonpath.New("hosts")),
					},
				},
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
			// change dashboards disk
			{
				Config: testSamlDashboardFlavorAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", randInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 2),
					resource.TestCheckResourceAttr(openSearchResource, "name", openSearchName),
					resource.TestCheckResourceAttr(openSearchResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(openSearchResource, "description", openSearchDesc),
					resource.TestCheckResourceAttr(openSearchResource, "config.admin_password", "password"),
					resource.TestCheckResourceAttrSet(openSearchResource, "service_account_id"),
					resource.TestCheckResourceAttr(openSearchResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(openSearchResource, "hosts.#", "2"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.0.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.1.fqdn"),
					test.AccCheckCreatedAtAttr(openSearchResource),
					testAccCheckMDBOpenSearchSubnetsAndZonesCount(&r, 3),
					testAccCheckMDBOpenSearchClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckMDBOpenSearchClusterDataNodeHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckMDBOpenSearchClusterDashboardsHasResources(&r, "s2.micro", "network-ssd", 11*1024*1024*1024),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.day", "FRI"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.hour", "20"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.enabled", "false"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.idp_entity_id", "https://test_identity_provider.com"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.idp_metadata_file_content", "<EntityDescriptor entityID=\"https://test_identity_provider.com\"></EntityDescriptor>"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.sp_entity_id", "https://some.db.yandex.net"),
					resource.TestCheckResourceAttr(openSearchResource, "auth_settings.saml.dashboards_url", "https://dashboards.example.com"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(openSearchResource, tfjsonpath.New("hosts")),
					},
				},
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
		},
	})
}

func TestAccMDBOpenSearchCluster_basic(t *testing.T) {
	var r opensearch.Cluster
	openSearchName := acctest.RandomWithPrefix("tf-opensearch")
	openSearchDesc := "OpenSearch Cluster Terraform Test"
	randInt := acctest.RandInt()
	folderID := test.GetExampleFolderID()
	openSearchDesc2 := "OpenSearch Cluster Terraform Test Updated"
	openSearchResource := openSearchResourcePrefix + openSearchName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBOpenSearchClusterDestroy,
		Steps: []resource.TestStep{
			// Create OpenSearch Cluster
			{
				//Config: testAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", false, randInt),
				Config: testAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", true, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 2),
					resource.TestCheckResourceAttr(openSearchResource, "name", openSearchName),
					resource.TestCheckResourceAttr(openSearchResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(openSearchResource, "description", openSearchDesc),
					resource.TestCheckResourceAttr(openSearchResource, "config.admin_password", "password"),
					resource.TestCheckResourceAttrSet(openSearchResource, "service_account_id"),
					resource.TestCheckResourceAttr(openSearchResource, "deletion_protection", "true"),
					resource.TestCheckResourceAttr(openSearchResource, "hosts.#", "2"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.0.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.1.fqdn"),
					test.AccCheckCreatedAtAttr(openSearchResource),
					testAccCheckMDBOpenSearchSubnetsAndZonesCount(&r, 3),
					testAccCheckMDBOpenSearchClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckMDBOpenSearchClusterDataNodeHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckMDBOpenSearchClusterDashboardsHasResources(&r, "s2.micro", "network-ssd", 10*1024*1024*1024),
					testAccCheckMDBOpenSearchClusterHasPlugins(&r, "analysis-icu", "repository-s3"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.day", "FRI"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.hour", "20"),
				),
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
			// uncheck 'deletion_protection'
			{
				Config: testAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", false, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 2),
					resource.TestCheckResourceAttr(openSearchResource, "deletion_protection", "false"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						pc.ExpectNoChangesAt(openSearchResource, tfjsonpath.New("hosts")),
					},
				},
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
			// check 'deletion_protection'
			{
				Config: testAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", true, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 2),
					resource.TestCheckResourceAttr(openSearchResource, "deletion_protection", "true"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						pc.ExpectNoChangesAt(openSearchResource, tfjsonpath.New("hosts")),
					},
				},
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
			// test 'deletion_protection
			{
				Config:      testAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRODUCTION", true, randInt),
				ExpectError: regexp.MustCompile(`.*The\soperation\swas\srejected\sbecause\scluster\shas\s'deletion_protection'\s=\sON.*`),
			},
			// uncheck 'deletion_protection'
			{
				Config: testAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", false, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 2),
					resource.TestCheckResourceAttr(openSearchResource, "deletion_protection", "false"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						pc.ExpectNoChangesAt(openSearchResource, tfjsonpath.New("hosts")),
					},
				},
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
			//Networks remove
			{
				Config: testAccMDBOpenSearchClusterConfigNetworksRemove(openSearchName, openSearchDesc2, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 2),
					testAccCheckMDBOpenSearchSubnetsAndZonesCount(&r, 2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(openSearchResource, tfjsonpath.New("hosts")),
					},
				},
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
			//Update OpenSearch Cluster (with Networks restore)
			{
				Config: testAccMDBOpenSearchClusterConfigUpdated(openSearchName, openSearchDesc2, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 5),
					resource.TestCheckResourceAttr(openSearchResource, "name", openSearchName),
					resource.TestCheckResourceAttr(openSearchResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(openSearchResource, "description", openSearchDesc2),
					resource.TestCheckResourceAttr(openSearchResource, "hosts.#", "5"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.0.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.1.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.2.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.3.fqdn"),
					resource.TestCheckResourceAttrSet(openSearchResource, "hosts.4.fqdn"),
					test.AccCheckCreatedAtAttr(openSearchResource),
					testAccCheckMDBOpenSearchSubnetsAndZonesCount(&r, 3),
					testAccCheckMDBOpenSearchClusterContainsLabel(&r, "test_key2", "test_value2"),
					testAccCheckMDBOpenSearchClusterDataNodeHasResources(&r, "s2.small", "network-ssd", 11*1024*1024*1024),
					testAccCheckMDBOpenSearchClusterDashboardsHasResources(&r, "s2.small", "network-ssd", 11*1024*1024*1024),
					testAccCheckMDBOpenSearchClusterHasPlugins(&r, "repository-s3"),
					resource.TestCheckResourceAttr(openSearchResource, "maintenance_window.type", "ANYTIME"),
					func(s *terraform.State) error {
						time.Sleep(1 * time.Minute)
						return nil
					},
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(openSearchResource, tfjsonpath.New("hosts")),
					},
				},
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
			//Add nodegroups
			{
				Config: testAccMDBOpenSearchClusterConfigWithManagerGroup(openSearchName, openSearchDesc2, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 12),
					resource.TestCheckResourceAttr(openSearchResource, "hosts.#", "12"),
					test.AccCheckCreatedAtAttr(openSearchResource),
					func(s *terraform.State) error {
						time.Sleep(1 * time.Minute)
						return nil
					},
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(openSearchResource, tfjsonpath.New("hosts")),
					},
				},
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
			//Remove nodegroups
			{
				Config: testAccMDBOpenSearchClusterConfigRemoveGroup(openSearchName, openSearchDesc2, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &r, 11),
					resource.TestCheckResourceAttr(openSearchResource, "hosts.#", "11"),
					// check role "manager" was removed
					func(s *terraform.State) error {
						for _, ng := range r.Config.Opensearch.NodeGroups {
							if ng.Name == "datamaster0" && (len(ng.Roles) != 1 || ng.Roles[0].String() != "DATA") {
								return fmt.Errorf("role 'DATA' was not set for nodegroup 'datamaster0'")
							}

						}
						return nil
					},
					func(s *terraform.State) error {
						time.Sleep(1 * time.Minute)
						return nil
					},
					test.AccCheckCreatedAtAttr(openSearchResource),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(openSearchResource, tfjsonpath.New("hosts")),
					},
				},
			},
			mdbOpenSearchClusterImportStep(openSearchResource),
		},
	})
}

func testAccCheckMDBOpenSearchClusterExists(n string, r *opensearch.Cluster, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*provider.Provider).GetConfig()

		found, err := config.SDK.MDB().OpenSearch().Cluster().Get(context.Background(), &opensearch.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("OpenSearch Cluster not found")
		}

		//TODO: should we change it?
		*r = *found

		resp, err := config.SDK.MDB().OpenSearch().Cluster().ListHosts(context.Background(), &opensearch.ListClusterHostsRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if len(resp.Hosts) != hosts {
			return fmt.Errorf("Expected %d hosts, got %d", hosts, len(resp.Hosts))
		}

		return nil
	}
}

func testAccCheckMDBOpenSearchClusterDataNodeHasResources(r *opensearch.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.Opensearch.NodeGroups[0].Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("OpenSearch expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskTypeId != diskType {
			return fmt.Errorf("expected disk type '%s', got '%s'", diskType, rs.DiskTypeId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("expected disk size '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

func testAccCheckMDBOpenSearchClusterDashboardsHasResources(r *opensearch.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := *r.Config.Dashboards.NodeGroups[0].Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("Dashboards expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskTypeId != diskType {
			return fmt.Errorf("expected disk type '%s', got '%s'", diskType, rs.DiskTypeId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("expected disk size '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

func testAccCheckMDBOpenSearchClusterDestroy(s *terraform.State) error {
	config := test.AccProvider.(*provider.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_opensearch_cluster" {
			continue
		}

		_, err := config.SDK.MDB().OpenSearch().Cluster().Get(context.Background(), &opensearch.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("OpenSearch Cluster still exists")
		}
	}

	return nil
}

func testAccCheckMDBOpenSearchClusterContainsLabel(r *opensearch.Cluster, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := r.Labels[key]
		if !ok {
			return fmt.Errorf("expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccCheckMDBOpenSearchClusterHasPlugins(r *opensearch.Cluster, plugins ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		p := r.Config.Opensearch.Plugins
		sort.Strings(p)
		sort.Strings(plugins)
		if !reflect.DeepEqual(p, plugins) {
			return fmt.Errorf("incorrect cluster plugins: expected '%s' but found '%s'", plugins, p)
		}
		return nil
	}
}

func testAccCheckMDBOpenSearchSubnetsAndZonesCount(r *opensearch.Cluster, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, ng := range r.Config.Opensearch.GetNodeGroups() {
			if len(ng.SubnetIds) != count {
				return fmt.Errorf("incorrect subnets count: expected '%d' but found '%d'", count, len(ng.SubnetIds))
			}

			if len(ng.ZoneIds) != count {
				return fmt.Errorf("incorrect zones count: expected '%d' but found '%d'", count, len(ng.ZoneIds))
			}
		}

		return nil
	}
}

func testSingleAccMDBOpenSearchClusterConfig(name, desc, environment string, randInt int) string {
	return openSearchIAMDependencies(randInt) + fmt.Sprintf("\n"+openSearchVPCDependencies+`

locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-d",
  ]
}

resource "yandex_mdb_opensearch_cluster" "%[1]s" {
  name        = "%[1]s"
  description = "%s"
  labels = {
    test_key  = "test_value"
  }
  environment = "%s"
  network_id  = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  security_group_ids = [yandex_vpc_security_group.mdb-opensearch-test-sg-x.id]
  service_account_id = "${yandex_iam_service_account.sa.id}"
  deletion_protection = false

  config {

    admin_password = "password"

    opensearch {
      node_groups {
        name             = "datamaster0"
        assign_public_ip = false
        hosts_count      = 1
        zone_ids         = local.zones
        roles = ["data","manager"]
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }
    }
  }

  depends_on = [
    yandex_vpc_subnet.mdb-opensearch-test-subnet-a,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-b,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-d,
  ]

  maintenance_window {
    type = "WEEKLY"
    day  = "FRI"
    hour = 20
  }

  timeouts {
    create = "1h"
    update = "2h"
  }
}
`, name, desc, environment)
}

func testSamlAccMDBOpenSearchClusterConfig(name, desc, environment string, randInt int, enabled bool) string {
	return openSearchIAMDependencies(randInt) + fmt.Sprintf("\n"+openSearchVPCDependencies+`

locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-d",
  ]
}

resource "yandex_mdb_opensearch_cluster" "%[1]s" {
  name        = "%[1]s"
  description = "%s"
  labels = {
    test_key  = "test_value"
  }
  environment = "%s"
  network_id  = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  security_group_ids = [yandex_vpc_security_group.mdb-opensearch-test-sg-x.id]
  service_account_id = "${yandex_iam_service_account.sa.id}"
  deletion_protection = false

  config {

    admin_password = "password"

    opensearch {
      node_groups {
        name             = "datamaster0"
        assign_public_ip = false
        hosts_count      = 1
        zone_ids         = local.zones
        subnet_ids = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles = ["DATA","MANAGER"]
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }
    }

    dashboards {
      node_groups {
        name = "dash0"
        assign_public_ip     = false
        hosts_count          = 1
        zone_ids             = local.zones  
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        resources {
          resource_preset_id   = "s2.micro"
          disk_size            = 10737418240
          disk_type_id         = "network-ssd"
        }
      }
    }
  }

  auth_settings = {
    saml = {
      enabled = %t
      idp_entity_id = "https://test_identity_provider.com"
      idp_metadata_file_content = "<EntityDescriptor entityID=\"https://test_identity_provider.com\"></EntityDescriptor>"
      sp_entity_id = "https://some.db.yandex.net",
      dashboards_url = "https://dashboards.example.com"
    }
  }

  depends_on = [
    yandex_vpc_subnet.mdb-opensearch-test-subnet-a,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-b,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-d,
  ]

  maintenance_window {
    type = "WEEKLY"
    day  = "FRI"
    hour = 20
  }

  timeouts {
    create = "1h"
    update = "2h"
  }
}
`, name, desc, environment, enabled)
}

func testSamlDashboardFlavorAccMDBOpenSearchClusterConfig(name, desc, environment string, randInt int, enabled bool) string {
	return openSearchIAMDependencies(randInt) + fmt.Sprintf("\n"+openSearchVPCDependencies+`

locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-d",
  ]
}

resource "yandex_mdb_opensearch_cluster" "%[1]s" {
  name        = "%[1]s"
  description = "%s"
  labels = {
    test_key  = "test_value"
  }
  environment = "%s"
  network_id  = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  security_group_ids = [yandex_vpc_security_group.mdb-opensearch-test-sg-x.id]
  service_account_id = "${yandex_iam_service_account.sa.id}"
  deletion_protection = false

  config {

    admin_password = "password"

    opensearch {
      node_groups {
        name             = "datamaster0"
        assign_public_ip = false
        hosts_count      = 1
        zone_ids         = local.zones
        subnet_ids = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles = ["DATA","MANAGER"]
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }
    }

    dashboards {
      node_groups {
        name = "dash0"
        assign_public_ip     = false
        hosts_count          = 1
        zone_ids             = local.zones  
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        resources {
          resource_preset_id   = "s2.micro"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
    }
  }

  auth_settings = {
    saml = {
      enabled = %t
      idp_entity_id = "https://test_identity_provider.com"
      idp_metadata_file_content = "<EntityDescriptor entityID=\"https://test_identity_provider.com\"></EntityDescriptor>"
      sp_entity_id = "https://some.db.yandex.net",
      dashboards_url = "https://dashboards.example.com"
    }
  }

  depends_on = [
    yandex_vpc_subnet.mdb-opensearch-test-subnet-a,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-b,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-d,
  ]

  maintenance_window {
    type = "WEEKLY"
    day  = "FRI"
    hour = 20
  }

  timeouts {
    create = "1h"
    update = "2h"
  }
}
`, name, desc, environment, enabled)
}

func testSimpleAccMDBOpenSearchClusterConfig(name, desc, environment string, randInt int) string {
	return openSearchIAMDependencies(randInt) + fmt.Sprintf("\n"+openSearchVPCDependencies+`

locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-d",
  ]
}

resource "yandex_mdb_opensearch_cluster" "%[1]s" {
  name        = "%[1]s"
  description = "%s"
  labels = {
    test_key  = "test_value"
  }
  environment = "%s"
  network_id  = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  security_group_ids = [yandex_vpc_security_group.mdb-opensearch-test-sg-x.id]
  service_account_id = "${yandex_iam_service_account.sa.id}"
  deletion_protection = false

  config {

    admin_password = "password"

    opensearch {
      node_groups {
        name             = "data1"
        assign_public_ip = false
        hosts_count      = 1
        zone_ids         = local.zones
        subnet_ids = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles = ["data"]
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }
      node_groups {
        name             = "manager0"
        assign_public_ip = false
        hosts_count      = 3
        zone_ids         = local.zones
        subnet_ids = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles = ["manager"]
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }
      node_groups {
        name             = "data0"
        assign_public_ip = false
        hosts_count      = 1
        zone_ids         = local.zones
        subnet_ids = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles = ["data"]
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }
    }
  }

  depends_on = [
    yandex_vpc_subnet.mdb-opensearch-test-subnet-a,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-b,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-d,
  ]

  maintenance_window {
    type = "WEEKLY"
    day  = "FRI"
    hour = 20
  }

  timeouts {
    create = "1h"
    update = "2h"
  }
}
`, name, desc, environment)
}

func testAccMDBOpenSearchClusterConfig(name, desc, environment string, deletionProtection bool, randInt int) string {
	return openSearchIAMDependencies(randInt) + fmt.Sprintf("\n"+openSearchVPCDependencies+`

locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-d",
  ]
}

resource "yandex_mdb_opensearch_cluster" "%[1]s" {
  name        = "%[1]s"
  description = "%s"
  labels = {
    test_key  = "test_value"
  }
  environment = "%s"
  network_id  = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  security_group_ids = [yandex_vpc_security_group.mdb-opensearch-test-sg-x.id]
  service_account_id = "${yandex_iam_service_account.sa.id}"
  deletion_protection = %t

  config {

    admin_password = "password"

    opensearch {
      node_groups {
        name = "datamaster0"
        assign_public_ip     = false
        hosts_count          = 1
        zone_ids             = local.zones
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles                = ["DATA", "MANAGER"]
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }
      plugins = ["analysis-icu", "repository-s3"]
    }

    dashboards {
      node_groups {
        name = "dash0"
        assign_public_ip     = false
        hosts_count          = 1
        zone_ids             = local.zones  
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        resources {
          resource_preset_id   = "s2.micro"
          disk_size            = 10737418240
          disk_type_id         = "network-ssd"
        }
      }
    }
  }

  depends_on = [
    yandex_vpc_subnet.mdb-opensearch-test-subnet-a,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-b,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-d,
  ]

  maintenance_window {
    type = "WEEKLY"
    day  = "FRI"
    hour = 20
  }

  timeouts {
    create = "1h"
    update = "2h"
  }
}
`, name, desc, environment, deletionProtection)
}

func testAccMDBOpenSearchClusterConfigNetworksRemove(name, desc string, randInt int) string {
	return openSearchIAMDependencies(randInt) + fmt.Sprintf("\n"+openSearchVPCDependencies+`

locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
  ]
}

resource "yandex_mdb_opensearch_cluster" "%[1]s" {
  name        = "%[1]s"
  description = "%s"
  labels = {
    test_key  = "test_value"
  }
  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  security_group_ids = [yandex_vpc_security_group.mdb-opensearch-test-sg-x.id]
  service_account_id = "${yandex_iam_service_account.sa.id}"
  deletion_protection = false

  config {

    admin_password = "password"

    opensearch {
      node_groups {
        name = "datamaster0"
        assign_public_ip     = false
        hosts_count          = 1
        zone_ids             = local.zones
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
        ]
        roles                = ["DATA", "MANAGER"]
        resources {
          resource_preset_id = "s2.micro"
          disk_size          = 10737418240
          disk_type_id       = "network-ssd"
        }
      }
      plugins = ["analysis-icu", "repository-s3"]
    }

    dashboards {
      node_groups {
        name = "dash0"
        assign_public_ip     = false
        hosts_count          = 1
        zone_ids             = local.zones  
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
        ]
        resources {
          resource_preset_id   = "s2.micro"
          disk_size            = 10737418240
          disk_type_id         = "network-ssd"
        }
      }
    }
  }

  depends_on = [
    yandex_vpc_subnet.mdb-opensearch-test-subnet-a,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-b,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-d,
  ]

  maintenance_window {
    type = "WEEKLY"
    day  = "FRI"
    hour = 20
  }

  timeouts {
    create = "1h"
    update = "2h"
  }
}
`, name, desc)
}

func testAccMDBOpenSearchClusterConfigNetworksRestore(name, desc, environment string, randInt int) string {
	return testAccMDBOpenSearchClusterConfig(name, desc, environment, false, randInt)
}

func testAccMDBOpenSearchClusterConfigUpdated(name, desc string, randInt int) string {
	return openSearchIAMDependencies(randInt) + fmt.Sprintf("\n"+openSearchVPCDependencies+`

locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-d",
  ]
}

resource "yandex_mdb_opensearch_cluster" "%[1]s" {
  name        = "%[1]s"
  description = "%s"
  labels = {
    test_key2  = "test_value2"
  }

  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  security_group_ids = [yandex_vpc_security_group.mdb-opensearch-test-sg-x.id, yandex_vpc_security_group.mdb-opensearch-test-sg-y.id]

  config {

    admin_password = "password_updated"

    opensearch {
      node_groups {
        name = "datamaster0"
        assign_public_ip     = false
        hosts_count          = 3
        zone_ids             = local.zones
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles                = ["DATA", "MANAGER"]
        resources {
          resource_preset_id   = "s2.small"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
      plugins = ["repository-s3"]
    }

    dashboards {
      node_groups {
        name = "dash0"
        assign_public_ip     = false
        hosts_count          = 2
        zone_ids             = local.zones  
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        resources {
          resource_preset_id   = "s2.small"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
    }
  }

  depends_on = [
    yandex_vpc_subnet.mdb-opensearch-test-subnet-a,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-b,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-d,
  ]

  maintenance_window {
    type = "ANYTIME"
  }

  timeouts {
    create = "1h"
    update = "2h"
  }
}
`, name, desc)
}

func testAccMDBOpenSearchClusterConfigWithManagerGroup(name, desc string, randInt int) string {
	return openSearchIAMDependencies(randInt) + fmt.Sprintf("\n"+openSearchVPCDependencies+`

locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-d",
  ]
}

resource "yandex_mdb_opensearch_cluster" "%[1]s" {
  name        = "%[1]s"
  description = "%s"
  labels = {
    test_key2  = "test_value2"
  }

  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  security_group_ids = [yandex_vpc_security_group.mdb-opensearch-test-sg-x.id, yandex_vpc_security_group.mdb-opensearch-test-sg-y.id]

  config {

    admin_password = "password_updated"

    opensearch {
      node_groups {
        name = "data1"
        assign_public_ip     = false
        hosts_count          = 1
        zone_ids             = local.zones
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles                = ["DATA"]
        resources {
          resource_preset_id   = "s2.small"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
      node_groups {
        name = "data2"
        assign_public_ip     = false
        hosts_count          = 1
        zone_ids             = local.zones
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles                = ["DATA"]
        resources {
          resource_preset_id   = "s2.small"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
      node_groups {
        name = "datamaster0"
        assign_public_ip     = false
        hosts_count          = 3
        zone_ids             = local.zones
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles                = ["DATA", "MANAGER"]
        resources {
          resource_preset_id   = "s2.small"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
      node_groups {
        name = "master"
        assign_public_ip     = false
        hosts_count          = 5
        zone_ids             = local.zones
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles                = ["MANAGER"]
        resources {
          resource_preset_id   = "s2.micro"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
      plugins = ["repository-s3"]
    }

    dashboards {
      node_groups {
        name = "dash0"
        assign_public_ip     = false
        hosts_count          = 2
        zone_ids             = local.zones  
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        resources {
          resource_preset_id   = "s2.small"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
    }
  }

  depends_on = [
    yandex_vpc_subnet.mdb-opensearch-test-subnet-a,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-b,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-d,
  ]

  maintenance_window {
    type = "ANYTIME"
  }

  timeouts {
    create = "1h"
    update = "2h"
  }
}
`, name, desc)
}

func testAccMDBOpenSearchClusterConfigRemoveGroup(name, desc string, randInt int) string {
	return openSearchIAMDependencies(randInt) + fmt.Sprintf("\n"+openSearchVPCDependencies+`

locals {
  zones = [
    "ru-central1-a",
    "ru-central1-b",
    "ru-central1-d",
  ]
}

resource "yandex_mdb_opensearch_cluster" "%[1]s" {
  name        = "%[1]s"
  description = "%s"
  labels = {
    test_key2  = "test_value2"
  }

  environment = "PRESTABLE"
  network_id  = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  security_group_ids = [yandex_vpc_security_group.mdb-opensearch-test-sg-x.id, yandex_vpc_security_group.mdb-opensearch-test-sg-y.id]

  config {

    admin_password = "password_updated"

    opensearch {
      node_groups {
        name = "data1"
        assign_public_ip     = false
        hosts_count          = 1
        zone_ids             = local.zones
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles                = ["DATA"]
        resources {
          resource_preset_id   = "s2.small"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
      node_groups {
        name = "datamaster0"
        assign_public_ip     = false
        hosts_count          = 3
        zone_ids             = local.zones
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles                = ["DATA"]
        resources {
          resource_preset_id   = "s2.small"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
      node_groups {
        name = "master"
        assign_public_ip     = false
        hosts_count          = 5
        zone_ids             = local.zones
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        roles                = ["MANAGER"]
        resources {
          resource_preset_id   = "s2.micro"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
      plugins = ["repository-s3"]
    }

    dashboards {
      node_groups {
        name = "dash0"
        assign_public_ip     = false
        hosts_count          = 2
        zone_ids             = local.zones  
        subnet_ids           = [
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-a.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-b.id}",
          "${yandex_vpc_subnet.mdb-opensearch-test-subnet-d.id}",
        ]
        resources {
          resource_preset_id   = "s2.micro"
          disk_size            = 11811160064
          disk_type_id         = "network-ssd"
        }
      }
    }
  }

  depends_on = [
    yandex_vpc_subnet.mdb-opensearch-test-subnet-a,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-b,
    yandex_vpc_subnet.mdb-opensearch-test-subnet-d,
  ]

  maintenance_window {
    type = "ANYTIME"
  }

  timeouts {
    create = "1h"
    update = "2h"
  }
}
`, name, desc)
}

func openSearchIAMDependencies(randInt int) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "sa" {
	name = "test-sa-for-tf-test-%[1]d"
}

resource "yandex_resourcemanager_folder_iam_member" "binding" {
	folder_id   = "%[2]s"
	member      = "serviceAccount:${yandex_iam_service_account.sa.id}"
	role        = "editor"
	sleep_after = 30
}

resource "yandex_iam_service_account_static_access_key" "sa-key" {
	service_account_id = "${yandex_iam_service_account.sa.id}"

	depends_on = [
		yandex_resourcemanager_folder_iam_member.binding
	]
}
`, randInt, test.GetExampleFolderID())
}

const openSearchVPCDependencies = `
resource "yandex_vpc_network" "mdb-opensearch-test-net" {}

resource "yandex_vpc_security_group" "mdb-opensearch-test-sg-x" {
  network_id     = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  ingress {
    protocol          = "ANY"
    description       = "Allow incoming traffic from members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
  egress {
    protocol          = "ANY"
    description       = "Allow outgoing traffic to members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
}

resource "yandex_vpc_security_group" "mdb-opensearch-test-sg-y" {
  network_id     = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  
  ingress {
    protocol          = "ANY"
    description       = "Allow incoming traffic from members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
  egress {
    protocol          = "ANY"
    description       = "Allow outgoing traffic to members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
}

resource "yandex_vpc_subnet" "mdb-opensearch-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-opensearch-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-opensearch-test-subnet-d" {
  zone           = "ru-central1-d"
  network_id     = "${yandex_vpc_network.mdb-opensearch-test-net.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}
`
