package mdb_sharded_postgresql_cluster_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	spqrIAMBindingResourceType = yandexMDBShardedPostgreSQLClusterResourceType + "_iam_binding"
	spqrIAMBindingResourceFoo  = spqrIAMBindingResourceType + ".foo"
	spqrIAMBindingResourceBar  = spqrIAMBindingResourceType + ".bar"

	spqrIAMClusterName = "tf-spqr-access-bindings"
	spqrIAMClusterDesc = "SPQR Terraform Test AccessBindings"

	spqrIAMRoleViewer = "managed-spqr.viewer"
	spqrIAMRoleEditor = "managed-spqr.editor"

	spqrResources = `
		resource_preset_id = "s2.micro"
		disk_size          = 10
		disk_type_id       = "network-hdd"
	`
	spqrLabels = `
		key1 = "value1"
		key2 = "value2"
		key3 = "value3"
    `
)

func TestAccMDBShardedPostgreSQLClusterIamBinding_basic(t *testing.T) {
	t.Parallel()
	var (
		cluster     spqr.Cluster
		clusterName = acctest.RandomWithPrefix(spqrIAMClusterName)
		ctx         = context.Background()
		description = spqrIAMClusterDesc + " Basic"
		environment = "PRESTABLE"

		role   = spqrIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBShardedPostgreSQLClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBShardedPostgreSQLClusterIamBindingConfig(role, userID, "foo", clusterName, description, environment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsAndParseMDBShardedPostgreSQLCluster(yandexMDBShardedPostgreSQLClusterResourceType+".foo", &cluster, 1),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().SPQR().Cluster()
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(spqrIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccMDBShardedPostgreSQLClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()
	var (
		cluster     spqr.Cluster
		clusterName = acctest.RandomWithPrefix(spqrIAMClusterName)
		ctx         = context.Background()
		description = spqrIAMClusterDesc + " AddAndRemove"
		environment = "PRESTABLE"

		roleFoo = spqrIAMRoleViewer
		roleBar = spqrIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBShardedPostgreSQLClusterDestroy,
		Steps: []resource.TestStep{
			// Prepare cluster
			{
				Config: testAccMDBShardedPostgreSQLClusterBasic("foo", clusterName, description, environment, spqrLabels, spqrResources),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsAndParseMDBShardedPostgreSQLCluster(yandexMDBShardedPostgreSQLClusterResourceType+".foo", &cluster, 1),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().SPQR().Cluster()
					}, &cluster, roleFoo),
				),
			},
			// Apply one IAM binding
			{
				Config: testAccMDBShardedPostgreSQLClusterIamBindingConfig(roleFoo, userID, "foo", clusterName, description, environment),
				Check: iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
					cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
					return cfg.SDK.MDB().SPQR().Cluster()
				}, &cluster, roleFoo, []string{userID}),
			},
			iam.IAMBindingImportTestStep(spqrIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Apply two IAM bindings
			{
				Config: testAccMDBShardedPostgreSQLClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, "foo", clusterName, description, environment),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().SPQR().Cluster()
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().SPQR().Cluster()
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(spqrIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(spqrIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all IAM bindings
			{
				Config: testAccMDBShardedPostgreSQLClusterBasic("foo", clusterName, description, environment, spqrLabels, spqrResources),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().SPQR().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.MDB().SPQR().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccMDBShardedPostgreSQLClusterIamBindingConfig(role, userID, resourceName, clusterName, desc, environment string) string {
	mainConfig := testAccMDBShardedPostgreSQLClusterBasic(resourceName, clusterName, desc, environment, spqrLabels, spqrResources)

	return fmt.Sprintf(`
resource "%s" "foo" {
	cluster_id = %s.id
	role       = "%s"
	members    = ["%s"]
}
`, spqrIAMBindingResourceType, yandexMDBShardedPostgreSQLClusterResourceType+"."+resourceName, role, userID) + mainConfig
}

func testAccMDBShardedPostgreSQLClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, resourceName, clusterName, desc, environment string) string {
	mainConfig := testAccMDBShardedPostgreSQLClusterBasic(resourceName, clusterName, desc, environment, spqrLabels, spqrResources)

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
		spqrIAMBindingResourceType, yandexMDBShardedPostgreSQLClusterResourceType+"."+resourceName, roleFoo, userID,
		spqrIAMBindingResourceType, yandexMDBShardedPostgreSQLClusterResourceType+"."+resourceName, roleBar, userID,
	) + mainConfig
}
