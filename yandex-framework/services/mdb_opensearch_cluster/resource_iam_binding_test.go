package mdb_opensearch_cluster_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	openSearchIAMBindingResourceType = openSearchResourceType + "_iam_binding"
	openSearchIAMBindingResourceFoo  = openSearchIAMBindingResourceType + ".foo"
	openSearchIAMBindingResourceBar  = openSearchIAMBindingResourceType + ".bar"

	openSearchIAMClusterName = "tf-opensearch-cluster-access-bindings"
	openSearchIAMClusterDesc = "OpenSearch Cluster Terraform Test AccessBindings"

	openSearchIAMRoleViewer = "managed-opensearch.viewer"
	openSearchIAMRoleEditor = "managed-opensearch.editor"
)

func TestAccMDBOpenSearchClusterIamBinding_basic(t *testing.T) {
	t.Parallel()
	var (
		cluster     opensearch.Cluster
		clusterName = acctest.RandomWithPrefix(openSearchIAMClusterName)
		clusterDesc = openSearchIAMClusterDesc + " Basic"
		ctx         = context.Background()
		randInt     = acctest.RandInt()
		environment = "PRESTABLE"

		role   = openSearchIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBOpenSearchClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBOpenSearchClusterIamBindingConfig(role, userID, clusterName, clusterDesc, environment, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResourcePrefix+clusterName, &cluster, 1),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().OpenSearch().Cluster()
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(openSearchIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccMDBOpenSearchClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()
	var (
		cluster     opensearch.Cluster
		clusterName = acctest.RandomWithPrefix(openSearchIAMClusterName)
		clusterDesc = openSearchIAMClusterDesc + " AddAndRemove"
		ctx         = context.Background()
		randInt     = acctest.RandInt()
		environment = "PRESTABLE"

		roleFoo = openSearchIAMRoleViewer
		roleBar = openSearchIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBOpenSearchClusterDestroy,
		Steps: []resource.TestStep{
			// Prepare cluster
			{
				Config: testSingleAccMDBOpenSearchClusterConfig(clusterName, clusterDesc, environment, randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBOpenSearchClusterExists(openSearchResourcePrefix+clusterName, &cluster, 1),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().OpenSearch().Cluster()
					}, &cluster, roleFoo),
				),
			},

			// Apply one IAM binding
			{
				Config: testAccMDBOpenSearchClusterIamBindingConfig(roleFoo, userID, clusterName, clusterDesc, environment, randInt),
				Check: iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
					cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
					return cfg.SDK.MDB().OpenSearch().Cluster()
				}, &cluster, roleFoo, []string{userID}),
			},
			iam.IAMBindingImportTestStep(openSearchIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Apply two IAM bindings
			{
				Config: testAccMDBOpenSearchClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, clusterName, clusterDesc, environment, randInt),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().OpenSearch().Cluster()
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().OpenSearch().Cluster()
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(openSearchIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(openSearchIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all IAM bindings
			{
				Config: testSingleAccMDBOpenSearchClusterConfig(clusterName, clusterDesc, environment, randInt),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().OpenSearch().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().OpenSearch().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccMDBOpenSearchClusterIamBindingConfig(role, userID, name, desc, environment string, randInt int,
) string {
	mainConfig := testSingleAccMDBOpenSearchClusterConfig(name, desc, environment, randInt)

	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role       = "%s"
  members    = ["%s"]
}
`, openSearchIAMBindingResourceType, openSearchResourcePrefix+name, role, userID) + mainConfig
}

func testAccMDBOpenSearchClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, name, desc, environment string, randInt int) string {
	mainConfig := testSingleAccMDBOpenSearchClusterConfig(name, desc, environment, randInt)

	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role       = "%s"
  members    = ["%s"]
}

resource "%s" "bar" {
  cluster_id = %s.id
  role       = "%s"
  members    = ["%s"]
}
`,
		openSearchIAMBindingResourceType, openSearchResourcePrefix+name, roleFoo, userID,
		openSearchIAMBindingResourceType, openSearchResourcePrefix+name, roleBar, userID,
	) + mainConfig
}
