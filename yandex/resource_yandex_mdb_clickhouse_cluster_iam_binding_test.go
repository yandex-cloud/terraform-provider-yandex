package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
)

const (
	clickhouseIAMBindingResourceType = chResourceType + "_iam_binding"
	clickhouseIAMBindingResourceFoo  = clickhouseIAMBindingResourceType + ".foo"
	clickhouseIAMBindingResourceBar  = clickhouseIAMBindingResourceType + ".bar"

	clickhouseIAMClusterName = "tf-clickhouse-cluster-access-bindings"
	clickhouseIAMClusterDesc = "ClickHouse Cluster Terraform Test AccessBindings"

	clickhouseIAMRoleViewer = "managed-clickhouse.viewer"
	clickhouseIAMRoleEditor = "managed-clickhouse.editor"
)

func TestAccMDBClickHouseClusterIamBinding_basic(t *testing.T) {
	t.Parallel()
	var (
		bucketName       = acctest.RandomWithPrefix("tf-test-clickhouse-bucket")
		cluster          clickhouse.Cluster
		clusterDesc      = clickhouseIAMClusterDesc + " Basic"
		clusterName      = acctest.RandomWithPrefix(clickhouseIAMClusterName)
		ctx              = context.Background()
		deleteProtection = false
		environment      = "PRESTABLE"
		rInt             = acctest.RandInt()

		role   = clickhouseIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBClickHouseClusterIamBindingConfig(
					role, userID, clusterName, clusterDesc,
					environment, deleteProtection, bucketName, rInt,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResourceFoo, &cluster, 1),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Clickhouse().Cluster()
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(clickhouseIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccMDBClickHouseClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()
	var (
		bucketName       = acctest.RandomWithPrefix("tf-test-clickhouse-bucket")
		cluster          clickhouse.Cluster
		clusterDesc      = clickhouseIAMClusterDesc + " AddAndRemove"
		clusterName      = acctest.RandomWithPrefix(clickhouseIAMClusterName)
		ctx              = context.Background()
		deleteProtection = false
		environment      = "PRESTABLE"
		rInt             = acctest.RandInt()

		roleFoo = clickhouseIAMRoleViewer
		roleBar = clickhouseIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			// Prepare cluster
			{
				Config: testAccMDBClickHouseClusterConfigMain(
					clusterName, clusterDesc, environment, deleteProtection,
					bucketName, rInt, MaintenanceWindowAnytime,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResourceFoo, &cluster, 1),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Clickhouse().Cluster()
					}, &cluster, roleFoo),
				),
			},
			// Apply one IAM binding
			{
				Config: testAccMDBClickHouseClusterIamBindingConfig(
					roleFoo, userID, clusterName, clusterDesc,
					environment, deleteProtection, bucketName, rInt,
				),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Clickhouse().Cluster()
					}, &cluster, roleFoo, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(clickhouseIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Apply two IAM bindings
			{
				Config: testAccMDBClickHouseClusterIamBindingMultipleConfig(
					roleFoo, roleBar, userID, clusterName, clusterDesc,
					environment, deleteProtection, bucketName, rInt,
				),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Clickhouse().Cluster()
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Clickhouse().Cluster()
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(clickhouseIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(clickhouseIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all IAM bindings
			{
				Config: testAccMDBClickHouseClusterConfigMain(
					clusterName, clusterDesc, environment, deleteProtection,
					bucketName, rInt, MaintenanceWindowAnytime,
				),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Clickhouse().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Clickhouse().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccMDBClickHouseClusterIamBindingConfig(role, userID, name, desc, environment string, deletionProtection bool, bucket string, randInt int) string {
	mainConfig := testAccMDBClickHouseClusterConfigMain(name, desc, environment, deletionProtection, bucket, randInt, MaintenanceWindowAnytime)
	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role        = "%s"
  members     = ["%s"]
}
`, clickhouseIAMBindingResourceType, chResourceFoo, role, userID) + mainConfig
}

func testAccMDBClickHouseClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, name, desc, environment string, deletionProtection bool, bucket string, randInt int) string {
	mainConfig := testAccMDBClickHouseClusterConfigMain(name, desc, environment, deletionProtection, bucket, randInt, MaintenanceWindowAnytime)
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
		clickhouseIAMBindingResourceType, chResourceFoo, roleFoo, userID,
		clickhouseIAMBindingResourceType, chResourceFoo, roleBar, userID,
	) + mainConfig
}
