package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/dataproc/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
)

const (
	dataprocIAMBindingResourceType = dataprocResourceType + "_iam_binding"
	dataprocIAMBindingResourceFoo  = dataprocIAMBindingResourceType + ".foo"
	dataprocIAMBindingResourceBar  = dataprocIAMBindingResourceType + ".bar"

	dataprocIAMClusterName = "tf-dataproc-cluster-access-bindings"
	dataprocIAMClusterDesc = "Dataproc Cluster Terraform Test AccessBindings"

	dataprocIAMRoleViewer = "dataproc.viewer"
	dataprocIAMRoleEditor = "dataproc.editor"
)

func TestAccDataprocClusterIamBinding_basic(t *testing.T) {
	t.Parallel()

	var (
		cluster        dataproc.Cluster
		ctx            = context.Background()
		templateParams = defaultDataprocConfigParams(t)

		role   = dataprocIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)
	templateParams.Name = acctest.RandomWithPrefix(dataprocIAMClusterName)
	templateParams.Description = dataprocIAMClusterDesc + " Basic"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccDataprocClusterIamBindingConfig(t, role, userID, templateParams),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataprocClusterExists(dataprocResourceType+".tf-dataproc-cluster", &cluster),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.Dataproc().Cluster()
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(dataprocIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccDataprocClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()

	var (
		cluster        dataproc.Cluster
		ctx            = context.Background()
		templateParams = defaultDataprocConfigParams(t)

		roleFoo = dataprocIAMRoleViewer
		roleBar = dataprocIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)
	templateParams.Name = acctest.RandomWithPrefix(dataprocIAMClusterName)
	templateParams.Description = dataprocIAMClusterDesc + " AddAndRemove"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			// Prepare cluster without IAM
			{
				Config: testAccDataprocClusterConfig(t, templateParams),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataprocClusterExists(dataprocResourceType+".tf-dataproc-cluster", &cluster),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.Dataproc().Cluster()
					}, &cluster, roleFoo),
				),
			},
			// One binding
			{
				Config: testAccDataprocClusterIamBindingConfig(t, roleFoo, userID, templateParams),
				Check: iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
					cfg := testAccProvider.Meta().(*Config)
					return cfg.sdk.Dataproc().Cluster()
				}, &cluster, roleFoo, []string{userID}),
			},
			iam.IAMBindingImportTestStep(dataprocIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Two bindings
			{
				Config: testAccDataprocClusterIamBindingMultipleConfig(t, roleFoo, roleBar, userID, templateParams),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.Dataproc().Cluster()
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.Dataproc().Cluster()
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(dataprocIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(dataprocIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all bindings
			{
				Config: testAccDataprocClusterConfig(t, templateParams),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.Dataproc().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.Dataproc().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccDataprocClusterIamBindingConfig(t *testing.T, role, userID string, templateParams dataprocTFConfigParams) string {
	main := testAccDataprocClusterConfig(t, templateParams)

	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role       = "%s"
  members    = ["%s"]
}
`, dataprocIAMBindingResourceType, dataprocResourceType+".tf-dataproc-cluster", role, userID) + main
}

func testAccDataprocClusterIamBindingMultipleConfig(t *testing.T, roleFoo, roleBar, userID string, templateParams dataprocTFConfigParams) string {
	main := testAccDataprocClusterConfig(t, templateParams)

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
		dataprocIAMBindingResourceType, dataprocResourceType+".tf-dataproc-cluster", roleFoo, userID,
		dataprocIAMBindingResourceType, dataprocResourceType+".tf-dataproc-cluster", roleBar, userID,
	) + main
}
