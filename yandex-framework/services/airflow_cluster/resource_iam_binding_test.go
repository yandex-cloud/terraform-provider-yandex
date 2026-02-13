package airflow_cluster_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	afv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/airflow/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	airflowResourceType       = "yandex_airflow_cluster"
	airflowResourceFoo        = airflowResourceType + ".foo"
	airflowIAMBindingResource = airflowResourceType + "_iam_binding"
	airflowIAMBindingFoo      = airflowIAMBindingResource + ".foo"
	airflowIAMBindingBar      = airflowIAMBindingResource + ".bar"

	airflowIAMClusterNamePrefix = "cluster-access-bindings"

	airflowIAMRoleViewer = "managed-airflow.viewer"
	airflowIAMRoleEditor = "managed-airflow.editor"
)

func TestAccMDBAirflowClusterIamBinding_basic(t *testing.T) {
	t.Parallel()
	var (
		cluster    afv1.Cluster
		ctx        = context.Background()
		randSuffix = acctest.RandomWithPrefix(airflowIAMClusterNamePrefix)

		role   = airflowIAMRoleViewer
		member = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckAirflowClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAirflowClusterIamBindingConfig(t, randSuffix, role, member),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAirflowExists(airflowResourceFoo, &cluster),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Airflow().Cluster()
					}, &cluster, role, []string{member}),
				),
			},
			iam.IAMBindingImportTestStep(airflowIAMBindingFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccMDBAirflowClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()
	var (
		cluster    afv1.Cluster
		ctx        = context.Background()
		randSuffix = acctest.RandomWithPrefix(airflowIAMClusterNamePrefix)

		roleFoo = airflowIAMRoleViewer
		roleBar = airflowIAMRoleEditor
		member  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckAirflowClusterDestroy,
		Steps: []resource.TestStep{
			// cluster only, no IAM
			{
				Config: testAccAirflowClusterConfigOnly(t, randSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAirflowExists(airflowResourceFoo, &cluster),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Airflow().Cluster()
					}, &cluster, roleFoo),
				),
			},
			// one binding
			{
				Config: testAccAirflowClusterIamBindingConfig(t, randSuffix, roleFoo, member),
				Check: iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
					cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
					return cfg.SDK.Airflow().Cluster()
				}, &cluster, roleFoo, []string{member}),
			},
			iam.IAMBindingImportTestStep(airflowIAMBindingFoo, &cluster, roleFoo, "cluster_id"),
			// two bindings
			{
				Config: testAccAirflowClusterIamBindingMultipleConfig(t, randSuffix, roleFoo, roleBar, member),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Airflow().Cluster()
					}, &cluster, roleFoo, []string{member}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Airflow().Cluster()
					}, &cluster, roleBar, []string{member}),
				),
			},
			iam.IAMBindingImportTestStep(airflowIAMBindingFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(airflowIAMBindingBar, &cluster, roleBar, "cluster_id"),
			// remove all IAM
			{
				Config: testAccAirflowClusterConfigOnly(t, randSuffix),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Airflow().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*provider.Provider).GetConfig()
						return cfg.SDK.Airflow().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccAirflowClusterConfigOnly(t *testing.T, randSuffix string) string {
	return airflowClusterConfig(t, airflowClusterConfigParams{
		RandSuffix:     randSuffix,
		FolderID:       os.Getenv("YC_FOLDER_ID"),
		Webserver:      airflowComponentParams{Count: 1, ResourcePresetID: "c1-m4"},
		Scheduler:      airflowComponentParams{Count: 1, ResourcePresetID: "c1-m4"},
		Worker:         airflowWorkerParams{MinCount: 1, MaxCount: 1, ResourcePresetID: "c1-m4"},
		AirflowVersion: "2.10",
		ResourceName:   "foo",
	})
}

func testAccAirflowClusterIamBindingConfig(t *testing.T, randSuffix, role, member string) string {
	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role       = "%s"
  members    = ["%s"]
}
`,
		airflowIAMBindingResource, airflowResourceFoo, role, member) + testAccAirflowClusterConfigOnly(t, randSuffix)
}

func testAccAirflowClusterIamBindingMultipleConfig(t *testing.T, randSuffix, roleFoo, roleBar, member string) string {
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
		airflowIAMBindingResource, airflowResourceFoo, roleFoo, member,
		airflowIAMBindingResource, airflowResourceFoo, roleBar, member,
	) + testAccAirflowClusterConfigOnly(t, randSuffix)
}
