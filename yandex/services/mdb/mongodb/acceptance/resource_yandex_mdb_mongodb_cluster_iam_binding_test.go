package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	mongodbsdk "github.com/yandex-cloud/go-sdk/services/mdb/mongodb/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
)

const (
	mongodbIAMBindingResourceType = mongodbResourceType + "_iam_binding"
	mongodbIAMBindingResourceFoo  = mongodbIAMBindingResourceType + ".foo"
	mongodbIAMBindingResourceBar  = mongodbIAMBindingResourceType + ".bar"

	mongodbIAMClusterName = "tf-mongodb-cluster-access-bindings"
	mongodbIAMClusterDesc = "MongoDB Cluster Terraform Test AccessBindings"

	mongodbIAMRoleViewer = "managed-mongodb.viewer"
	mongodbIAMRoleEditor = "managed-mongodb.editor"
)

func TestAccMDBMongoDBClusterIamBinding_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()
	var (
		cluster    mongodb.Cluster
		configData = create8_0V0ConfigData()
		ctx        = context.Background()

		role   = mongodbIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)
	configData["ClusterDescription"] = mongodbIAMClusterDesc + " Basic"
	configData["ClusterName"] = acctest.RandomWithPrefix(mongodbIAMClusterName)
	configData["Environment"] = "PRESTABLE"
	configData["Hosts"] = mongoHosts

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMongoDBClusterIamBindingConfig(role, userID) + makeConfig(t, &configData, nil),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMDBMongoDBClusterExists(mongodbResourceFoo, &cluster, 2),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return mongodbsdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(mongodbIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccMDBMongoDBClusterIamBinding_multiple(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()
	var (
		cluster    mongodb.Cluster
		configData = create8_0V0ConfigData()
		ctx        = context.Background()

		roleFoo = mongodbIAMRoleViewer
		roleBar = mongodbIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)
	configData["ClusterDescription"] = mongodbIAMClusterDesc + " AddAndRemove"
	configData["ClusterName"] = acctest.RandomWithPrefix(mongodbIAMClusterName)
	configData["Environment"] = "PRESTABLE"
	configData["Hosts"] = mongoHosts

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Prepare cluster without IAM bindings
			{
				Config: makeConfig(t, &configData, nil),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMDBMongoDBClusterExists(mongodbResourceFoo, &cluster, 2),
					iam.TestAccCheckIamBindingEmpty(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return mongodbsdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, roleFoo),
				),
			},
			// Apply One IAM binding
			{
				Config: testAccMDBMongoDBClusterIamBindingConfig(roleFoo, userID) + makeConfig(t, &configData, nil),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return mongodbsdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, roleFoo, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(mongodbIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Apply Two IAM bindings
			{
				Config: testAccMDBMongoDBClusterIamBindingMultipleConfig(roleFoo, roleBar, userID) + makeConfig(t, &configData, nil),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return mongodbsdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return mongodbsdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(mongodbIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(mongodbIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all IAM bindings
			{
				Config: makeConfig(t, &configData, nil),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return mongodbsdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return mongodbsdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccMDBMongoDBClusterIamBindingConfig(role, userID string) string {
	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role       = "%s"
  members    = ["%s"]
}
`, mongodbIAMBindingResourceType, mongodbResourceFoo, role, userID)
}

func testAccMDBMongoDBClusterIamBindingMultipleConfig(roleFoo, roleBar, userID string) string {
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
		mongodbIAMBindingResourceType, mongodbResourceFoo, roleFoo, userID,
		mongodbIAMBindingResourceType, mongodbResourceFoo, roleBar, userID,
	)
}
