package yandex

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
)

const (
	pgIAMBindingResourceType = pgResourceType + "_iam_binding"
	pgIAMBindingResourceFoo  = pgIAMBindingResourceType + ".foo"
	pgIAMBindingResourceBar  = pgIAMBindingResourceType + ".bar"

	pgIAMClusterName = "tf-postgresql-cluster-access-bindings"
	pgIAMClusterDesc = "PostgreSQL Cluster Terraform Test AccessBindings"

	pgIAMRoleViewer = "managed-postgresql.viewer"
	pgIAMRoleEditor = "managed-postgresql.editor"
)

func TestAccMDBPostgreSQLClusterIamBinding_basic(t *testing.T) {
	t.Parallel()
	var (
		cluster     postgresql.Cluster
		clusterDesc = pgIAMClusterDesc + " Basic"
		clusterName = acctest.RandomWithPrefix(pgIAMClusterName)
		ctx         = context.Background()
		environment = "PRESTABLE"
		version     = postgresql_versions[rand.Intn(len(postgresql_versions))]

		role   = pgIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBPostgreSQLClusterIamBindingConfig(
					role, userID, clusterName, clusterDesc, environment, version, 10, false,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(pgResourceFoo, &cluster, 1),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().PostgreSQL().Cluster()
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(pgIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccMDBPostgreSQLClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()
	var (
		cluster     postgresql.Cluster
		clusterDesc = pgIAMClusterDesc + " AddAndRemove"
		clusterName = acctest.RandomWithPrefix(pgIAMClusterName)
		ctx         = context.Background()
		environment = "PRESTABLE"
		version     = postgresql_versions[rand.Intn(len(postgresql_versions))]

		roleFoo = pgIAMRoleViewer
		roleBar = pgIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccMDBPGClusterConfigMain(
					clusterName, clusterDesc, environment, version, 10, false,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBPGClusterExists(pgResourceFoo, &cluster, 1),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().PostgreSQL().Cluster()
					}, &cluster, roleFoo),
				),
			},
			// Apply One IAM binding
			{
				Config: testAccMDBPostgreSQLClusterIamBindingConfig(
					roleFoo, userID, clusterName, clusterDesc, environment, version, 10, false,
				),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().PostgreSQL().Cluster()
					}, &cluster, roleFoo, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(pgIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Apply Two IAM bindings
			{
				Config: testAccMDBPostgreSQLClusterIamBindingMultipleConfig(
					roleFoo, roleBar, userID, clusterName, clusterDesc, environment, version, 10, false,
				),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().PostgreSQL().Cluster()
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().PostgreSQL().Cluster()
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(pgIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(pgIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all IAM bindings
			{
				Config: testAccMDBPGClusterConfigMain(
					clusterName, clusterDesc, environment, version, 10, false,
				),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().PostgreSQL().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().PostgreSQL().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccMDBPostgreSQLClusterIamBindingConfig(role, userID, name, desc, environment, version string, diskSize int32, deletionProtection bool) string {
	mainConfig := testAccMDBPGClusterConfigMain(
		name, desc, environment, version, diskSize, deletionProtection,
	)
	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role        = "%s"
  members     = ["%s"]
}
`, pgIAMBindingResourceType, pgResourceFoo, role, userID) + mainConfig
}

func testAccMDBPostgreSQLClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, name, desc, environment, version string, diskSize int32, deletionProtection bool) string {
	mainConfig := testAccMDBPGClusterConfigMain(
		name, desc, environment, version, diskSize, deletionProtection,
	)
	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role        = "%s"
  members     = ["%s"]
}

resource "%s" "bar" {
  cluster_id = %s.id
  role        = "%s"
  members     = ["%s"]
}
`,
		pgIAMBindingResourceType, pgResourceFoo, roleFoo, userID,
		pgIAMBindingResourceType, pgResourceFoo, roleBar, userID,
	) + mainConfig
}
