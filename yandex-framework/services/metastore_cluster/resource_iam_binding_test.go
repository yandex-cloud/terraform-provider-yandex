package metastore_cluster_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/metastore/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	metastoreIAMBindingResourceType = metastoreResourceType + "_iam_binding"
	metastoreIAMBindingResourceFoo  = metastoreIAMBindingResourceType + ".foo"
	metastoreIAMBindingResourceBar  = metastoreIAMBindingResourceType + ".bar"

	metastoreIAMClusterName = "cluster-access-bindings"
	metastoreIAMClusterDesc = "Metastore Cluster Terraform Test AccessBindings"

	metastoreIAMRoleViewer = "managed-metastore.viewer"
	metastoreIAMRoleEditor = "managed-metastore.editor"
)

func TestAccMDBMetastoreClusterIamBinding_basic(t *testing.T) {
	t.Parallel()
	var (
		cluster     metastore.Cluster
		clusterName = acctest.RandomWithPrefix(metastoreIAMClusterName)
		ctx         = context.Background()
		description = metastoreIAMClusterDesc + " Basic"

		role   = metastoreIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckMetastoreClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMetastoreClusterIamBindingConfig(t, role, userID, clusterName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetastoreExists(metastoreResourceType+".metastore_cluster", &cluster),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Metastore().Cluster()
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(metastoreIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccMDBMetastoreClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()
	var (
		cluster     metastore.Cluster
		clusterName = acctest.RandomWithPrefix(metastoreIAMClusterName)
		ctx         = context.Background()
		description = metastoreIAMClusterDesc + " AddAndRemove"

		roleFoo = metastoreIAMRoleViewer
		roleBar = metastoreIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckMetastoreClusterDestroy,
		Steps: []resource.TestStep{
			// Prepare cluster
			{
				Config: metastoreClusterConfig(t, metastoreClusterConfigParams{
					Description:    newOptional(description),
					RandSuffix:     clusterName,
					FolderID:       os.Getenv("YC_FOLDER_ID"),
					SubnetIDVar:    "yandex_vpc_subnet.metastore-a.id",
					ResourcePreset: "c2-m4",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetastoreExists(metastoreResourceType+".metastore_cluster", &cluster),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Metastore().Cluster()
					}, &cluster, roleFoo),
				),
			},
			// Apply one IAM binding
			{
				Config: testAccMDBMetastoreClusterIamBindingConfig(t, roleFoo, userID, clusterName, description),
				Check: iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
					cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
					return cfg.SDK.Metastore().Cluster()
				}, &cluster, roleFoo, []string{userID}),
			},
			iam.IAMBindingImportTestStep(metastoreIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Apply two IAM bindings
			{
				Config: testAccMDBMetastoreClusterIamBindingMultipleConfig(t, roleFoo, roleBar, userID, clusterName, description),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Metastore().Cluster()
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Metastore().Cluster()
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(metastoreIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(metastoreIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all IAM bindings
			{
				Config: metastoreClusterConfig(t, metastoreClusterConfigParams{
					Description:    newOptional(description),
					RandSuffix:     clusterName,
					FolderID:       os.Getenv("YC_FOLDER_ID"),
					SubnetIDVar:    "yandex_vpc_subnet.metastore-a.id",
					ResourcePreset: "c2-m4",
				}),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Metastore().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Metastore().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccMDBMetastoreClusterIamBindingConfig(t *testing.T, role, userID, nameSuffix, description string) string {
	mainConfig := metastoreClusterConfig(t, metastoreClusterConfigParams{
		Description:    newOptional(description),
		FolderID:       os.Getenv("YC_FOLDER_ID"),
		RandSuffix:     nameSuffix,
		ResourcePreset: "c2-m4",
		SubnetIDVar:    "yandex_vpc_subnet.metastore-a.id",
	})

	return fmt.Sprintf(`
resource "%s" "foo" {
	cluster_id = yandex_metastore_cluster.metastore_cluster.id
	role       = "%s"
	members    = ["%s"]
}
`, metastoreIAMBindingResourceType, role, userID) + mainConfig
}

func testAccMDBMetastoreClusterIamBindingMultipleConfig(t *testing.T, roleFoo, roleBar, userID, nameSuffix, description string) string {
	mainConfig := metastoreClusterConfig(t, metastoreClusterConfigParams{
		Description:    newOptional(description),
		FolderID:       os.Getenv("YC_FOLDER_ID"),
		RandSuffix:     nameSuffix,
		ResourcePreset: "c2-m4",
		SubnetIDVar:    "yandex_vpc_subnet.metastore-a.id",
	})

	return fmt.Sprintf(`
resource "%s" "foo" {
	cluster_id = yandex_metastore_cluster.metastore_cluster.id
	role       = "%s"
	members    = ["%s"]
}

resource "%s" "bar" {
	cluster_id = yandex_metastore_cluster.metastore_cluster.id
	role       = "%s"
	members    = ["%s"]
}
`,
		metastoreIAMBindingResourceType, roleFoo, userID,
		metastoreIAMBindingResourceType, roleBar, userID,
	) + mainConfig
}
