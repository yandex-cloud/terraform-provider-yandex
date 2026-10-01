package yandex_ydb_database_iam_member_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	ydbsdk "github.com/yandex-cloud/go-sdk/services/ydb/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	ydbDatabaseResource = "yandex_ydb_database_serverless.test-database"
	ydbLocationID       = "ru-central1"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccYDBDatabaseIAMMember_lifecycle(t *testing.T) {
	databaseName := acctest.RandomWithPrefix("tf-ydb-database")
	role := "ydb.viewer"
	firstMember := "system:allUsers"
	secondMember := "system:allAuthenticatedUsers"
	ctx := context.Background()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccYDBDatabaseIAMMemberConfig(databaseName, role, firstMember, secondMember),
				Check:  testAccCheckYDBDatabaseIAMMembers(ctx, role, []string{firstMember, secondMember}),
			},
			{
				Config: testAccYDBDatabaseIAMMemberConfig(databaseName, role, secondMember),
				Check:  testAccCheckYDBDatabaseIAMMembers(ctx, role, []string{secondMember}),
			},
			{
				ResourceName:                         "yandex_ydb_database_iam_member.viewer_0",
				ImportStateIdFunc:                    importYDBDatabaseIAMMemberIDFunc(role, secondMember),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "database_id",
				ImportStateVerifyIgnore:              []string{"sleep_after"},
			},
			{
				Config: testAccYDBDatabaseConfig(databaseName),
				Check:  testAccCheckYDBDatabaseIAMMembers(ctx, role, nil),
			},
		},
	})
}

func ydbDatabaseClient() ydbsdk.DatabaseClient {
	config := test.AccProvider.(*provider.Provider).GetConfig()
	return ydbsdk.NewDatabaseClient(config.SDKv2)
}

func testAccCheckYDBDatabaseIAMMembers(ctx context.Context, role string, members []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		databaseID, err := ydbDatabaseIDFromState(state)
		if err != nil {
			return err
		}

		response, err := ydbDatabaseClient().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: databaseID,
		})
		if err != nil {
			return fmt.Errorf("get access bindings for YDB database %q: %w", databaseID, err)
		}

		var actual []string
		for _, binding := range response.AccessBindings {
			if binding.RoleId == role {
				actual = append(actual, binding.Subject.Type+":"+binding.Subject.Id)
			}
		}

		expected := slices.Clone(members)
		slices.Sort(actual)
		slices.Sort(expected)
		if !slices.Equal(actual, expected) {
			return fmt.Errorf("members mismatch for role %q: expected %v, got %v", role, expected, actual)
		}

		return nil
	}
}

func importYDBDatabaseIAMMemberIDFunc(role, member string) func(*terraform.State) (string, error) {
	return func(state *terraform.State) (string, error) {
		databaseID, err := ydbDatabaseIDFromState(state)
		if err != nil {
			return "", err
		}

		return databaseID + "," + role + "," + member, nil
	}
}

func ydbDatabaseIDFromState(state *terraform.State) (string, error) {
	databaseResource, ok := state.RootModule().Resources[ydbDatabaseResource]
	if !ok {
		return "", fmt.Errorf("YDB database resource %q not found in state", ydbDatabaseResource)
	}
	if databaseResource.Primary.ID == "" {
		return "", fmt.Errorf("YDB database resource %q has no ID", ydbDatabaseResource)
	}

	return databaseResource.Primary.ID, nil
}

func testAccYDBDatabaseIAMMemberConfig(databaseName, role string, members ...string) string {
	config := testAccYDBDatabaseConfig(databaseName)
	for i, member := range members {
		config += fmt.Sprintf(`
resource "yandex_ydb_database_iam_member" "viewer_%d" {
  database_id = yandex_ydb_database_serverless.test-database.id
  role        = %q
  member      = %q
}
`, i, role, member)
	}
	return config
}

func testAccYDBDatabaseConfig(databaseName string) string {
	return fmt.Sprintf(`
resource "yandex_ydb_database_serverless" "test-database" {
  name        = %q
  location_id = %q
}
`, databaseName, ydbLocationID)
}
