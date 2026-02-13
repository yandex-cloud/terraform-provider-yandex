package trino_cluster_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	trinov1 "github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	trinoIAMBindingResourceType = trinoResourceType + "_iam_binding"
	trinoIAMBindingResourceFoo  = trinoIAMBindingResourceType + ".foo"
	trinoIAMBindingResourceBar  = trinoIAMBindingResourceType + ".bar"

	trinoIAMClusterName = "tf-trino-cluster-access-bindings"

	trinoIAMRoleViewer = "managed-trino.viewer"
	trinoIAMRoleEditor = "managed-trino.editor"
)

func TestAccTrinoClusterIamBinding_basic(t *testing.T) {
	t.Parallel()
	var (
		cluster     trinov1.Cluster
		clusterName = acctest.RandomWithPrefix(trinoIAMClusterName)
		ctx         = context.Background()

		role   = trinoIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTrinoClusterIamBindingConfig(t, role, userID, clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoExists(trinoResourceType+".trino_cluster", &cluster),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Trino().Cluster()
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(trinoIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccTrinoClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()
	var (
		cluster     trinov1.Cluster
		clusterName = acctest.RandomWithPrefix(trinoIAMClusterName)
		ctx         = context.Background()

		roleFoo = trinoIAMRoleViewer
		roleBar = trinoIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		Steps: []resource.TestStep{
			// Prepare cluster
			{
				Config: trinoClusterConfig(t, trinoClusterConfigParams{
					RandSuffix: clusterName,
					FolderID:   os.Getenv("YC_FOLDER_ID"),
					Coordinator: trinoComponentParams{
						ResourcePresetID: "c4-m16",
					},
					Worker: trinoWorkerParams{
						ResourcePresetID: "c4-m16",
						FixedScale: &FixedScaleParams{
							Count: 1,
						},
					},
					Version: "468",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrinoExists(trinoResourceType+".trino_cluster", &cluster),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Trino().Cluster()
					}, &cluster, roleFoo),
				),
			},
			// One IAM binding
			{
				Config: testAccTrinoClusterIamBindingConfig(t, roleFoo, userID, clusterName),
				Check: iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
					cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
					return cfg.SDK.Trino().Cluster()
				}, &cluster, roleFoo, []string{userID}),
			},
			iam.IAMBindingImportTestStep(trinoIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Two IAM bindings
			{
				Config: testAccTrinoClusterIamBindingMultipleConfig(t, roleFoo, roleBar, userID, clusterName),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Trino().Cluster()
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Trino().Cluster()
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(trinoIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(trinoIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all IAM bindings
			{
				Config: trinoClusterConfig(t, trinoClusterConfigParams{
					RandSuffix: clusterName,
					FolderID:   os.Getenv("YC_FOLDER_ID"),
					Coordinator: trinoComponentParams{
						ResourcePresetID: "c4-m16",
					},
					Worker: trinoWorkerParams{
						ResourcePresetID: "c4-m16",
						FixedScale: &FixedScaleParams{
							Count: 1,
						},
					},
					Version: "468",
				}),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Trino().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Trino().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccTrinoClusterIamBindingConfig(t *testing.T, role, userID, name string) string {
	main := trinoClusterConfig(t, trinoClusterConfigParams{
		RandSuffix: name,
		FolderID:   os.Getenv("YC_FOLDER_ID"),
		Coordinator: trinoComponentParams{
			ResourcePresetID: "c4-m16",
		},
		Worker: trinoWorkerParams{
			ResourcePresetID: "c4-m16",
			FixedScale: &FixedScaleParams{
				Count: 1,
			},
		},
		Version: "468",
	})

	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role       = "%s"
  members    = ["%s"]
}
`, trinoIAMBindingResourceType, trinoResourceType+".trino_cluster", role, userID) + main
}

func testAccTrinoClusterIamBindingMultipleConfig(t *testing.T, roleFoo, roleBar, userID, name string) string {
	main := trinoClusterConfig(t, trinoClusterConfigParams{
		RandSuffix: name,
		FolderID:   os.Getenv("YC_FOLDER_ID"),
		Coordinator: trinoComponentParams{
			ResourcePresetID: "c4-m16",
		},
		Worker: trinoWorkerParams{
			ResourcePresetID: "c4-m16",
			FixedScale: &FixedScaleParams{
				Count: 1,
			},
		},
		Version: "468",
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
		trinoIAMBindingResourceType, trinoResourceType+".trino_cluster", roleFoo, userID,
		trinoIAMBindingResourceType, trinoResourceType+".trino_cluster", roleBar, userID,
	) + main
}
