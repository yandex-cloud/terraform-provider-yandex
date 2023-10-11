package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

//revive:disable:var-naming
func TestAccDataSourceIAMRole_byID(t *testing.T) {
	testRoleID := getExampleRoleID()
	var role iam.Role

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckIAMRole_byID(testRoleID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.yandex_iam_role.foo", "id"),
					testAccDataSourceIAMRoleExists("data.yandex_iam_role.foo", &role),
				),
			},
		},
	})
}

func testAccDataSourceIAMRoleExists(n string, role *iam.Role) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.IAM().Role().Get(context.Background(), &iam.GetRoleRequest{
			RoleId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Role not found")
		}

		*role = *found

		return nil
	}
}

func testAccCheckIAMRole_byID(cloudID string) string {
	return fmt.Sprintf(`
data "yandex_iam_role" "foo" {
  role_id = "%s"
}
`, cloudID)
}
