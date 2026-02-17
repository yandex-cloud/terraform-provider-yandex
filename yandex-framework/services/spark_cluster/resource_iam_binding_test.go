package spark_cluster_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	sparkv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/spark/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	sparkIAMBindingResourceType = sparkResourceType + "_iam_binding"
	sparkIAMBindingResourceFoo  = sparkIAMBindingResourceType + ".foo"
	sparkIAMBindingResourceBar  = sparkIAMBindingResourceType + ".bar"

	sparkIAMClusterName = "tf-spark-cluster-access-bindings"
	sparkIAMClusterDesc = "Spark Cluster Terraform Test AccessBindings"

	sparkIAMRoleViewer = "managed-spark.viewer"
	sparkIAMRoleEditor = "managed-spark.editor"
)

func TestAccSparkClusterIamBinding_basic(t *testing.T) {
	t.Parallel()
	var (
		cluster     sparkv1.Cluster
		clusterDesc = sparkIAMClusterDesc + " Basic"
		clusterName = acctest.RandomWithPrefix(sparkIAMClusterName)
		ctx         = context.Background()

		role   = sparkIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSparkClusterIamBindingConfig(t, role, userID, clusterName, clusterDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSparkExists(sparkResourceType+".spark_cluster", &cluster),
					iam.TestAccCheckIamBindingContainsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Spark().Cluster()
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(sparkIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccSparkClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()
	var (
		cluster sparkv1.Cluster
		ctx     = context.Background()

		clusterName = acctest.RandomWithPrefix(sparkIAMClusterName)
		clusterDesc = sparkIAMClusterDesc + " AddAndRemove"

		roleFoo = sparkIAMRoleViewer
		roleBar = sparkIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		Steps: []resource.TestStep{
			// Prepare cluster
			{
				Config: sparkClusterConfig(t, sparkClusterConfigParams{
					RandSuffix:               clusterName,
					Description:              clusterDesc,
					DriverResourcePresetID:   "c2-m8",
					DriverSize:               1,
					ExecutorResourcePresetID: "c4-m16",
					ExecutorMinSize:          1,
					ExecutorMaxSize:          2,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSparkExists(sparkResourceType+".spark_cluster", &cluster),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Spark().Cluster()
					}, &cluster, roleFoo),
				),
			},

			// One binding
			{
				Config: testAccSparkClusterIamBindingConfig(t, roleFoo, userID, clusterName, clusterDesc),
				Check: iam.TestAccCheckIamBindingContainsMembers(ctx, func() iam.BindingsGetter {
					cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
					return cfg.SDK.Spark().Cluster()
				}, &cluster, roleFoo, []string{userID}),
			},
			iam.IAMBindingImportTestStep(sparkIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Two bindings
			{
				Config: testAccSparkClusterIamBindingMultipleConfig(t, roleFoo, roleBar, userID, clusterName, clusterDesc),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingContainsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Spark().Cluster()
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingContainsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Spark().Cluster()
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(sparkIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(sparkIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all
			{
				Config: sparkClusterConfig(t, sparkClusterConfigParams{
					RandSuffix:               clusterName,
					Description:              clusterDesc,
					DriverResourcePresetID:   "c2-m8",
					DriverSize:               1,
					ExecutorResourcePresetID: "c4-m16",
					ExecutorMinSize:          1,
					ExecutorMaxSize:          2,
				}),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Spark().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Spark().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccSparkClusterIamBindingConfig(t *testing.T, role, userID, name, desc string) string {
	main := sparkClusterConfig(t, sparkClusterConfigParams{
		RandSuffix:               name,
		Description:              desc,
		DriverResourcePresetID:   "c2-m8",
		DriverSize:               1,
		ExecutorResourcePresetID: "c4-m16",
		ExecutorMinSize:          1,
		ExecutorMaxSize:          2,
	})

	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role       = "%s"
  members    = ["%s"]
}
`, sparkIAMBindingResourceType, sparkResourceType+".spark_cluster", role, userID) + main
}

func testAccSparkClusterIamBindingMultipleConfig(t *testing.T, roleFoo, roleBar, userID, name, desc string) string {
	main := sparkClusterConfig(t, sparkClusterConfigParams{
		RandSuffix:               name,
		Description:              desc,
		DriverResourcePresetID:   "c2-m8",
		DriverSize:               1,
		ExecutorResourcePresetID: "c4-m16",
		ExecutorMinSize:          1,
		ExecutorMaxSize:          2,
	})

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
		sparkIAMBindingResourceType, sparkResourceType+".spark_cluster", roleFoo, userID,
		sparkIAMBindingResourceType, sparkResourceType+".spark_cluster", roleBar, userID,
	) + main
}
