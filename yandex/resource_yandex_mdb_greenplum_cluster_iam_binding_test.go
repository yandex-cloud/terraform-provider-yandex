package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
)

const (
	greenplumIAMBindingResourceType = greenplumResourceType + "_iam_binding"
	greenplumIAMBindingResourceFoo  = greenplumIAMBindingResourceType + ".foo"
	greenplumIAMBindingResourceBar  = greenplumIAMBindingResourceType + ".bar"

	greenplumIAMClusterName = "tf-greenplum-access-bindings"
	greenplumIAMClusterDesc = "Greenplum Cluster Terraform Test AccessBindings"

	greenplumIAMRoleViewer = "managed-greenplum.viewer"
	greenplumIAMRoleEditor = "managed-greenplum.editor"
)

func TestAccMDBGreenplumClusterIamBinding_basic(t *testing.T) {
	t.Parallel()
	var (
		cluster     greenplum.Cluster
		clusterDesc = greenplumIAMClusterDesc + " Basic"
		clusterName = acctest.RandomWithPrefix(greenplumIAMClusterName)
		ctx         = context.Background()

		role   = greenplumIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBGreenplumClusterIamBindingConfig(role, userID, clusterName, clusterDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBGreenplumClusterExists(greenplumResourceFoo, &cluster, 2, 5),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Greenplum().Cluster()
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(greenplumIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccMDBGreenplumClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()
	var (
		cluster     greenplum.Cluster
		clusterDesc = greenplumIAMClusterDesc + " AddAndRemove"
		clusterName = acctest.RandomWithPrefix(greenplumIAMClusterName)
		ctx         = context.Background()

		roleFoo = greenplumIAMRoleViewer
		roleBar = greenplumIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			// Prepare cluster without IAM bindings
			{
				Config: testAccMDBGreenplumClusterConfigStep1(clusterName, clusterDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBGreenplumClusterExists(greenplumResourceFoo, &cluster, 2, 5),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Greenplum().Cluster()
					}, &cluster, roleFoo),
				),
			},
			// Apply single IAM binding
			{
				Config: testAccMDBGreenplumClusterIamBindingConfig(roleFoo, userID, clusterName, clusterDesc),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Greenplum().Cluster()
					}, &cluster, roleFoo, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(greenplumIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Apply two IAM bindings
			{
				Config: testAccMDBGreenplumClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, clusterName, clusterDesc),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Greenplum().Cluster()
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Greenplum().Cluster()
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(greenplumIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(greenplumIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all IAM bindings
			{
				Config: testAccMDBGreenplumClusterConfigStep1(clusterName, clusterDesc),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Greenplum().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Greenplum().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccMDBGreenplumClusterIamBindingConfig(
	role, userID, name, desc string,
) string {
	mainConfig := testAccMDBGreenplumClusterConfigStep1(name, desc)

	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role       = "%s"
  members    = ["%s"]
}
`, greenplumIAMBindingResourceType, greenplumResourceFoo, role, userID) + mainConfig
}

func testAccMDBGreenplumClusterIamBindingMultipleConfig(
	roleFoo, roleBar, userID, name, desc string,
) string {
	mainConfig := testAccMDBGreenplumClusterConfigStep1(name, desc)

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
		greenplumIAMBindingResourceType, greenplumResourceFoo, roleFoo, userID,
		greenplumIAMBindingResourceType, greenplumResourceFoo, roleBar, userID,
	) + mainConfig
}
