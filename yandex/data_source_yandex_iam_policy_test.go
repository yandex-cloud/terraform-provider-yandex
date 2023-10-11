package yandex

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceYandexIAMPolicy(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceYandexIAMPolicy(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.yandex_iam_policy.foo", "policy_data"),
					testPolicyUnmarshal("data.yandex_iam_policy.foo"),
				),
			},
		},
	})
}

//revive:disable:var-naming
func TestAccDataSourceYandexIAMPolicy_invalidConfig(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeImageDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceYandexIAMPolicy_invalidConfig(),
				ExpectError: regexp.MustCompile("expect 'member' value should be in TYPE:ID format"),
			},
		},
	})
}

func testPolicyUnmarshal(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policyData := rs.Primary.Attributes["policy_data"]
		var policy Policy

		err := json.Unmarshal([]byte(policyData), &policy)

		return err
	}
}

func testAccDataSourceYandexIAMPolicy() string {
	return `
data "yandex_iam_policy" "foo" {
  binding {
    role = "editor"

    members = [
      "userAccount:some_user_id_1",
      "userAccount:some_user_id_2",
    ]
  }

  binding {
    role = "owner"

    members = [
      "userAccount:some_user_id_1",
      "userAccount:some_user_id_2",
    ]
  }
}
`
}

func testAccDataSourceYandexIAMPolicy_invalidConfig() string {
	return `
data "yandex_iam_policy" "foo" {
  binding {
    role = "role_editor"

    members = [
      "user_user1@yandex.ru",
      "user2@yandex.ru",
    ]
  }

  binding {
    role = "role_owner"

    members = [
      ":user10@yandex.ru",
    ]
  }
}
`
}
