package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
)

const (
	mysqlIAMBindingResourceType = mysqlResourceType + "_iam_binding"
	mysqlIAMBindingResourceFoo  = mysqlIAMBindingResourceType + ".foo"
	mysqlIAMBindingResourceBar  = mysqlIAMBindingResourceType + ".bar"

	mysqlIAMClusterName = "tf-mysql-cluster-access-bindings"
	mysqlIAMClusterDesc = "MySQL Cluster Terraform Test AccessBindings"

	mysqlIAMRoleViewer = "managed-mysql.viewer"
	mysqlIAMRoleEditor = "managed-mysql.editor"
)

func TestAccMDBMySQLClusterIamBinding_basic(t *testing.T) {
	t.Parallel()
	var (
		cluster          mysql.Cluster
		clusterDesc      = mysqlIAMClusterDesc + " Basic"
		clusterName      = acctest.RandomWithPrefix(mysqlIAMClusterName)
		ctx              = context.Background()
		deleteProtection = false
		environment      = "PRESTABLE"

		role   = mysqlIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMySQLClusterIamBindingConfig(role, userID, clusterName, clusterDesc, environment, deleteProtection),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResourceFoo, &cluster),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().MySQL().Cluster()
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(mysqlIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccMDBMySQLClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()
	var (
		cluster          mysql.Cluster
		clusterDesc      = mysqlIAMClusterDesc + " AddAndRemove"
		clusterName      = acctest.RandomWithPrefix(mysqlIAMClusterName)
		ctx              = context.Background()
		deleteProtection = false
		environment      = "PRESTABLE"

		roleFoo = mysqlIAMRoleViewer
		roleBar = mysqlIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccMDBMySQLClusterConfigMain(clusterName, clusterDesc, environment, deleteProtection),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLClusterExists(mysqlResourceFoo, &cluster),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().MySQL().Cluster()
					}, &cluster, roleFoo),
				),
			},
			// Apply One IAM binding
			{
				Config: testAccMDBMySQLClusterIamBindingConfig(roleFoo, userID, clusterName, clusterDesc, environment, deleteProtection),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().MySQL().Cluster()
					}, &cluster, roleFoo, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(mysqlIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Apply Two IAM bindings
			{
				Config: testAccMDBMySQLClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, clusterName, clusterDesc, environment, deleteProtection),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().MySQL().Cluster()
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().MySQL().Cluster()
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(mysqlIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(mysqlIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all IAM bindings
			{
				Config: testAccMDBMySQLClusterConfigMain(clusterName, clusterDesc, environment, deleteProtection),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().MySQL().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().MySQL().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccMDBMySQLClusterIamBindingConfig(role, userID, name, desc, environment string, deletionProtection bool) string {
	mainConfig := testAccMDBMySQLClusterConfigMain(name, desc, environment, deletionProtection)
	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role        = "%s"
  members     = ["%s"]
}
`, mysqlIAMBindingResourceType, mysqlResourceFoo, role, userID) + mainConfig
}

func testAccMDBMySQLClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, name, desc, environment string, deletionProtection bool) string {
	mainConfig := testAccMDBMySQLClusterConfigMain(name, desc, environment, deletionProtection)
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
		mysqlIAMBindingResourceType, mysqlResourceFoo, roleFoo, userID,
		mysqlIAMBindingResourceType, mysqlResourceFoo, roleBar, userID,
	) + mainConfig
}
