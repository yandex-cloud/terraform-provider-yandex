package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	redissdk "github.com/yandex-cloud/go-sdk/services/mdb/redis/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
)

const (
	redisIAMBindingResourceType = redisResourceType + "_iam_binding"
	redisIAMBindingResourceFoo  = redisIAMBindingResourceType + ".foo"
	redisIAMBindingResourceBar  = redisIAMBindingResourceType + ".bar"

	redisIAMClusterName = "tf-redis-cluster-access-bindings"
	redisIAMClusterDesc = "Redis Cluster Terraform Test AccessBindings"

	redisIAMRoleViewer = "managed-redis.viewer"
	redisIAMRoleEditor = "managed-redis.editor"
)

func TestAccMDBRedisClusterIamBinding_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()
	var (
		cluster            redis.Cluster
		clusterName        = acctest.RandomWithPrefix(redisIAMClusterName)
		clusterDesc        = redisIAMClusterDesc + " Basic"
		ctx                = context.Background()
		deletionProtection = false
		environment        = "PRESTABLE"

		role   = redisIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccMDBRedisClusterIamBindingConfig(role, userID, clusterName, clusterDesc, environment, deletionProtection),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMDBRedisClusterExists(redisResourceFoo, &cluster, 1, true, true, true, "ON"),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return redissdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(redisIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccMDBRedisClusterIamBinding_multiple(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var (
		announceHostnames  = true
		authSentinel       = true
		cluster            redis.Cluster
		clusterDesc        = redisIAMClusterDesc + " AddAndRemove"
		clusterName        = acctest.RandomWithPrefix(redisIAMClusterName)
		ctx                = context.Background()
		deletionProtection = false
		environment        = "PRESTABLE"
		tlsEnabled         = true

		roleFoo = redisIAMRoleViewer
		roleBar = redisIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Prepare cluster without IAM bindings
			{
				Config: testAccMDBRedisClusterConfigMain(
					clusterName, clusterDesc, environment, deletionProtection,
					&tlsEnabled, &announceHostnames, &authSentinel,
					"", "8.0-valkey", "hm3-c2-m8", 16, "",
					"16777215 8388607 61", "16777214 8388606 62",
					[]*bool{nil}, []*int{nil},
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMDBRedisClusterExists(redisResourceFoo, &cluster, 1, true, true, true, "ON"),
					iam.TestAccCheckIamBindingEmpty(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return redissdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, roleFoo),
				),
			},
			// Apply one IAM binding
			{
				Config: testAccMDBRedisClusterIamBindingConfig(
					roleFoo, userID, clusterName, clusterDesc, environment, deletionProtection,
				),
				Check: iam.TestAccCheckIamBindingEqualsMembers(ctx, func() (iam.BindingsGetter, error) {
					cfg, configErr := suite.Config()
					if configErr != nil {
						return nil, configErr
					}

					return redissdk.NewClusterClient(cfg.SDK), nil
				}, &cluster, roleFoo, []string{userID}),
			},
			iam.IAMBindingImportTestStep(redisIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Apply two IAM bindings
			{
				Config: testAccMDBRedisClusterIamBindingMultipleConfig(
					roleFoo, roleBar, userID, clusterName, clusterDesc, environment, deletionProtection,
				),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return redissdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return redissdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(redisIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(redisIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all IAM bindings
			{
				Config: testAccMDBRedisClusterConfigMain(
					clusterName, clusterDesc, environment, deletionProtection,
					&tlsEnabled, &announceHostnames, &authSentinel,
					"", "8.0-valkey", "hm3-c2-m8", 16, "",
					"16777215 8388607 61", "16777214 8388606 62",
					[]*bool{nil}, []*int{nil},
				),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return redissdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() (iam.BindingsGetter, error) {
						cfg, configErr := suite.Config()
						if configErr != nil {
							return nil, configErr
						}

						return redissdk.NewClusterClient(cfg.SDK), nil
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccMDBRedisClusterIamBindingConfig(role, userID, name, desc, environment string, deletionProtection bool) string {
	announceHostnames := true
	authSentinel := true
	tlsEnabled := true

	mainConfig := testAccMDBRedisClusterConfigMain(
		name, desc, environment, deletionProtection,
		&tlsEnabled, &announceHostnames, &authSentinel,
		"", "8.0-valkey", "hm3-c2-m8", 16, "",
		"16777215 8388607 61", "16777214 8388606 62",
		[]*bool{nil}, []*int{nil},
	)

	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role       = "%s"
  members    = ["%s"]
}
`, redisIAMBindingResourceType, redisResourceFoo, role, userID) + mainConfig
}

func testAccMDBRedisClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, name, desc, environment string, deletionProtection bool) string {
	announceHostnames := true
	authSentinel := true
	tlsEnabled := true

	mainConfig := testAccMDBRedisClusterConfigMain(
		name, desc, environment, deletionProtection,
		&tlsEnabled, &announceHostnames, &authSentinel,
		"", "8.0-valkey", "hm3-c2-m8", 16, "",
		"16777215 8388607 61", "16777214 8388606 62",
		[]*bool{nil}, []*int{nil},
	)

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
		redisIAMBindingResourceType, redisResourceFoo, roleFoo, userID,
		redisIAMBindingResourceType, redisResourceFoo, roleBar, userID,
	) + mainConfig
}
