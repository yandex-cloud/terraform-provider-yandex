package mdb_clickhouse_database_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func mdbClickHouseDatabaseImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccMDBClickHouseDatabase_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-clickkhouse-database-basic")
	description := "ClickHouse database terraform resource test"

	dbResource1 := formatResourceName(chDBResourceName1)
	dbResource2 := formatResourceName(chDBResourceName2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBClickHouseDatabaseBasicConfig(clusterName, description, []string{chDBResourceName1}) + testAccMDBClickHouseDatabaseWithEngine(chDBResourceName2, "atomic"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterHasDatabases(chClusterResourceID, []string{chDBResourceName1, chDBResourceName2}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dbResource1, tfjsonpath.New("name"), knownvalue.StringExact(chDBResourceName1)),
					statecheck.ExpectKnownValue(dbResource2, tfjsonpath.New("name"), knownvalue.StringExact(chDBResourceName2)),
					statecheck.ExpectKnownValue(dbResource2, tfjsonpath.New("engine"), knownvalue.StringExact("atomic")),
					makeClickHouseDatabaseResourceIDComparer(dbResource1),
					makeClickHouseDatabaseResourceIDComparer(dbResource2),
				},
			},
			mdbClickHouseDatabaseImportStep(dbResource1),
			mdbClickHouseDatabaseImportStep(dbResource2),
			{
				Config: testAccMDBClickHouseDatabaseBasicConfig(clusterName, description, []string{chDBResourceName1}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterHasDatabases(chClusterResourceID, []string{chDBResourceName1}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					makeClickHouseDatabaseResourceIDComparer(dbResource1),
				},
			},
			{
				Config: testAccMDBClickHouseDatabaseBasicConfig(clusterName, description, []string{chDBResourceName3, chDBResourceName4}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterHasDatabases(chClusterResourceID, []string{chDBResourceName3, chDBResourceName4}),
				),
			},
		},
	})
}

func testAccMDBClickHouseDatabaseBasicConfig(name, description string, dbNames []string) string {
	result := testAccMDBClickHouseClusterConfigMain(name, description)
	for _, dbName := range dbNames {
		result = result + fmt.Sprintf(`

	resource "yandex_mdb_clickhouse_database" "%s" {
		cluster_id = %s
		name       = "%s"
	}
	
	`, dbName, chClusterResourceIDLink, dbName,
		)
	}
	return result
}

func testAccMDBClickHouseDatabaseWithEngine(dbName string, engine string) string {
	return fmt.Sprintf(`
	resource "yandex_mdb_clickhouse_database" "%s" {
		cluster_id = %s
		name       = "%s"
		engine = "%s"
	}
	`, dbName, chClusterResourceIDLink, dbName, engine,
	)
}

func testAccCheckMDBClickHouseDatabaseResourceIDField(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		expectedResourceId := resourceid.Construct(rs.Primary.Attributes["cluster_id"], rs.Primary.Attributes["name"])

		if expectedResourceId != rs.Primary.ID {
			return fmt.Errorf("Wrong resource %s id. Expected %s, got %s", resourceName, expectedResourceId, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMDBClickHouseClusterHasDatabases(r string, databases []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		cid := rs.Primary.ID

		resp, err := config.SDK.MDB().Clickhouse().Database().List(context.Background(), &clickhouse.ListDatabasesRequest{
			ClusterId: cid,
			PageSize:  100,
		})
		if err != nil {
			return err
		}
		dbs := []string{}
		for _, d := range resp.Databases {
			dbs = append(dbs, d.Name)
		}

		if len(dbs) != len(databases) {
			return fmt.Errorf("Expected %d dbs, found %d %v", len(databases), len(dbs), dbs)
		}

		sort.Strings(dbs)
		sort.Strings(databases)
		if fmt.Sprintf("%v", dbs) != fmt.Sprintf("%v", databases) {
			return fmt.Errorf("User has wrong databases, %v. Expected %v", dbs, databases)
		}
		return nil
	}
}
