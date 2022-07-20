package yandex

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
)

const ydbDatabaseResource = "yandex_ydb_database_serverless.test-database"

func importYDBDatabaseIDFunc(database *ydb.Database, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return database.Id + " " + role, nil
	}
}

func TestAccYDBDatabaseIamBinding_basic(t *testing.T) {
	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database")

	role := "ydb.viewer"
	userID := "system:allAuthenticatedUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccYDBDatabaseIamBindingBasic(databaseName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testYandexYDBDatabaseServerlessExists(ydbDatabaseResource, &database),
					testAccCheckYDBDatabaseIam(ydbDatabaseResource, role, []string{userID}),
				),
			},
			{
				ResourceName:      "yandex_ydb_database_iam_binding.viewer",
				ImportStateIdFunc: importYDBDatabaseIDFunc(&database, role),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccYDBDatabaseIamBinding_remove(t *testing.T) {
	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database")

	role := "ydb.viewer"
	userID := "system:allAuthenticatedUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccYDBDatabase(databaseName),
				Check: resource.ComposeTestCheckFunc(
					testYandexYDBDatabaseServerlessExists(ydbDatabaseResource, &database),
					testAccCheckYDBDatabaseEmptyIam(ydbDatabaseResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccYDBDatabaseIamBindingBasic(databaseName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYDBDatabaseIam(ydbDatabaseResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccYDBDatabase(databaseName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYDBDatabaseEmptyIam(ydbDatabaseResource),
				),
			},
		},
	})
}

func testAccYDBDatabaseIamBindingBasic(databaseName, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_ydb_database_serverless" "test-database" {
  name       = "%s"
}

resource "yandex_ydb_database_iam_binding" "viewer" {
  database_id = yandex_ydb_database_serverless.test-database.id
  role        = "%s"
  members     = ["%s"]
}
`, databaseName, role, userID)
}

func testAccYDBDatabase(databaseName string) string {
	return fmt.Sprintf(`
resource "yandex_ydb_database_serverless" "test-database" {
  name       = "%s"
}
`, databaseName)
}

func testAccCheckYDBDatabaseEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getYDBDatabaseResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckYDBDatabaseIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getYDBDatabaseResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		var roleMembers []string
		for _, binding := range bindings {
			if binding.RoleId == role {
				member := binding.Subject.Type + ":" + binding.Subject.Id
				roleMembers = append(roleMembers, member)
			}
		}
		sort.Strings(members)
		sort.Strings(roleMembers)

		if reflect.DeepEqual(members, roleMembers) {
			return nil
		}

		return fmt.Errorf("Binding found but expected members is %v, got %v", members, roleMembers)
	}
}

func getYDBDatabaseResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := testAccProvider.Meta().(*Config)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return getYDBDatabaseAccessBindings(config, rs.Primary.ID)
}
