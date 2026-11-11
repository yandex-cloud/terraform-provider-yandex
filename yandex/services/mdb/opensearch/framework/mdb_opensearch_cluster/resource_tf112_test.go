//go:build tf1_12

package mdb_opensearch_cluster_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccMDBOpenSearchClusterAdminPasswordWo_TF1_12(t *testing.T) {
	var cluster opensearch.Cluster
	openSearchName := acctest.RandomWithPrefix("tf-opensearch-password-wo")
	openSearchDesc := "OpenSearch Cluster write-only password acceptance test"
	randInt := acctest.RandInt()
	openSearchResource := openSearchResourcePrefix + openSearchName
	baseConfig := testSingleAccMDBOpenSearchClusterConfig(openSearchName, openSearchDesc, "PRESTABLE", randInt)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { test.AccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBOpenSearchClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBOpenSearchClusterConfigWithAdminPasswordWo(
					baseConfig,
					"write_only_P@ssw0rd_1",
					1,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResource, &cluster, 1),
					resource.TestCheckNoResourceAttr(openSearchResource, "config.admin_password"),
					resource.TestCheckNoResourceAttr(openSearchResource, "config.admin_password_wo"),
					resource.TestCheckResourceAttr(openSearchResource, "config.admin_password_wo_version", "1"),
				),
			},
			{
				Config: testAccMDBOpenSearchClusterConfigWithAdminPasswordWo(
					baseConfig,
					"write_only_P@ssw0rd_1",
					1,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMDBOpenSearchClusterConfigWithPasswordFields(
					baseConfig,
					`admin_password            = "legacy_P@ssw0rd"
    admin_password_wo         = "write_only_P@ssw0rd"
    admin_password_wo_version = 1`,
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`one \(and only one\)`),
			},
			{
				Config: testAccMDBOpenSearchClusterConfigWithPasswordFields(
					baseConfig,
					`admin_password_wo = "write_only_P@ssw0rd"`,
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config: testAccMDBOpenSearchClusterConfigWithPasswordFields(
					baseConfig,
					`admin_password_wo_version = 1`,
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config: testAccMDBOpenSearchClusterConfigWithAdminPasswordWo(
					baseConfig,
					"write_only_P@ssw0rd_2",
					2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(openSearchResource, "config.admin_password"),
					resource.TestCheckNoResourceAttr(openSearchResource, "config.admin_password_wo"),
					resource.TestCheckResourceAttr(openSearchResource, "config.admin_password_wo_version", "2"),
				),
			},
		},
	})
}

func testAccMDBOpenSearchClusterConfigWithAdminPasswordWo(config, password string, version int) string {
	return testAccMDBOpenSearchClusterConfigWithPasswordFields(
		config,
		fmt.Sprintf(`admin_password_wo         = %q
    admin_password_wo_version = %d`, password, version),
	)
}

func testAccMDBOpenSearchClusterConfigWithPasswordFields(config, passwordFields string) string {
	return strings.Replace(config, `admin_password = "dummy_P@ssw0rd"`, passwordFields, 1)
}
