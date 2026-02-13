package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
)

const (
	kafkaIAMBindingResourceType = kfResourceType + "_iam_binding"
	kafkaIAMBindingResourceFoo  = kafkaIAMBindingResourceType + ".foo"
	kafkaIAMBindingResourceBar  = kafkaIAMBindingResourceType + ".bar"

	kafkaIAMClusterName = "tf-kafka-cluster-access-bindings"
	kafkaIAMClusterDesc = "Kafka Cluster Terraform Test AccessBindings"

	kafkaIAMRoleViewer = "managed-kafka.viewer"
	kafkaIAMRoleEditor = "managed-kafka.editor"
)

func TestAccMDBKafkaClusterIamBinding_basic(t *testing.T) {
	t.Parallel()

	var (
		cluster     kafka.Cluster
		clusterName = acctest.RandomWithPrefix(kafkaIAMClusterName)
		clusterDesc = kafkaIAMClusterDesc + " Basic"
		environment = "PRESTABLE"
		ctx         = context.Background()

		role   = kafkaIAMRoleViewer
		userID = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBKafkaClusterIamBindingConfig(role, userID, clusterName, clusterDesc, environment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBKafkaClusterExists(kfResourceFoo, &cluster, 1),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Kafka().Cluster()
					}, &cluster, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(kafkaIAMBindingResourceFoo, &cluster, role, "cluster_id"),
		},
	})
}

func TestAccMDBKafkaClusterIamBinding_multiple(t *testing.T) {
	t.Parallel()

	var (
		cluster     kafka.Cluster
		clusterName = acctest.RandomWithPrefix(kafkaIAMClusterName)
		clusterDesc = kafkaIAMClusterDesc + " AddAndRemove"
		environment = "PRESTABLE"
		ctx         = context.Background()

		roleFoo = kafkaIAMRoleViewer
		roleBar = kafkaIAMRoleEditor
		userID  = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			// Prepare cluster without IAM
			{
				Config: testAccMDBKafkaClusterConfigMain(clusterName, clusterDesc, environment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBKafkaClusterExists(kfResourceFoo, &cluster, 1),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Kafka().Cluster()
					}, &cluster, roleFoo),
				),
			},
			// Apply one IAM binding
			{
				Config: testAccMDBKafkaClusterIamBindingConfig(roleFoo, userID, clusterName, clusterDesc, environment),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(
						ctx,
						func() iam.BindingsGetter {
							cfg := testAccProvider.Meta().(*Config)
							return cfg.sdk.MDB().Kafka().Cluster()
						},
						&cluster,
						roleFoo,
						[]string{userID},
					),
				),
			},
			iam.IAMBindingImportTestStep(kafkaIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			// Apply two IAM bindings
			{
				Config: testAccMDBKafkaClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, clusterName, clusterDesc, environment),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Kafka().Cluster()
					}, &cluster, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Kafka().Cluster()
					}, &cluster, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(kafkaIAMBindingResourceFoo, &cluster, roleFoo, "cluster_id"),
			iam.IAMBindingImportTestStep(kafkaIAMBindingResourceBar, &cluster, roleBar, "cluster_id"),
			// Remove all IAM bindings
			{
				Config: testAccMDBKafkaClusterConfigMain(clusterName, clusterDesc, environment),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Kafka().Cluster()
					}, &cluster, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testAccProvider.Meta().(*Config)
						return cfg.sdk.MDB().Kafka().Cluster()
					}, &cluster, roleBar),
				),
			},
		},
	})
}

func testAccMDBKafkaClusterIamBindingConfig(role, userID, name, desc, environment string) string {
	mainConfig := testAccMDBKafkaClusterConfigMain(name, desc, environment)

	return fmt.Sprintf(`
resource "%s" "foo" {
  cluster_id = %s.id
  role       = "%s"
  members    = ["%s"]
}
`, kafkaIAMBindingResourceType, kfResourceFoo, role, userID) + mainConfig
}

func testAccMDBKafkaClusterIamBindingMultipleConfig(roleFoo, roleBar, userID, name, desc, environment string) string {
	mainConfig := testAccMDBKafkaClusterConfigMain(name, desc, environment)

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
		kafkaIAMBindingResourceType, kfResourceFoo, roleFoo, userID,
		kafkaIAMBindingResourceType, kfResourceFoo, roleBar, userID,
	) + mainConfig
}
