package yandex

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

//revive:disable:var-naming
func TestAccDataSourceYandexLogin_byLogin(t *testing.T) {
	testLogin1 := getExampleUserLogin1()
	testLogin2 := getExampleUserLogin2()
	var user1 iam.UserAccount
	var user2 iam.UserAccount

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckYandexLogin(testLogin1, testLogin2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.foo", "id"),
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.foo", "user_id"),
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.foo", "default_email"),
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.bar", "id"),
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.bar", "user_id"),
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.bar", "default_email"),
					testAccDataSourceYandexLoginExists("data.yandex_iam_user.foo", &user1),
					testAccDataSourceYandexLoginExists("data.yandex_iam_user.bar", &user2),
					testAccDataSourceYandexLoginHasLogin(&user1, testLogin1),
					testAccDataSourceYandexLoginHasLogin(&user2, testLogin2),
				),
			},
		},
	})
}

func TestAccDataSourceYandexLogin_byUserID(t *testing.T) {
	testID1 := getExampleUserID1()
	testID2 := getExampleUserID2()
	var user1 iam.UserAccount
	var user2 iam.UserAccount

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckYandexUserID(testID1, testID2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.foo", "id"),
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.foo", "login"),
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.foo", "default_email"),
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.bar", "id"),
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.bar", "login"),
					resource.TestCheckResourceAttrSet("data.yandex_iam_user.bar", "default_email"),
					testAccDataSourceYandexLoginExists("data.yandex_iam_user.foo", &user1),
					testAccDataSourceYandexLoginExists("data.yandex_iam_user.bar", &user2),
					testAccDataSourceYandexLoginHasID(&user1, testID1),
					testAccDataSourceYandexLoginHasID(&user2, testID2),
				),
			},
		},
	})
}

func TestAccDataSourceYandexLogin_invalidLogin(t *testing.T) {
	invalidLogin1 := acctest.RandomWithPrefix("some-login")
	invalidLogin2 := acctest.RandomWithPrefix("some-login")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckYandexLogin(invalidLogin1, invalidLogin2),
				ExpectError: regexp.MustCompile("login not found"),
			},
		},
	})
}

func testAccDataSourceYandexLoginExists(n string, user *iam.UserAccount) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.IAM().UserAccount().Get(context.Background(), &iam.GetUserAccountRequest{
			UserAccountId: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("User not found")
		}

		*user = *found

		return nil
	}
}

func testAccDataSourceYandexLoginHasLogin(user *iam.UserAccount, login string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		yaUser := user.GetYandexPassportUserAccount()

		if yaUser == nil {
			return fmt.Errorf("not Yandex passport user account")
		}

		if yaUser.Login != login {
			return fmt.Errorf("Expect login of user account %s, got %s", login, yaUser.Login)
		}

		return nil
	}
}

func testAccDataSourceYandexLoginHasID(user *iam.UserAccount, id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if user.Id != id {
			return fmt.Errorf("Expect id of user account %s, got %s", id, user.Id)
		}

		return nil
	}
}

func testAccCheckYandexLogin(login1, login2 string) string {
	return fmt.Sprintf(`
data "yandex_iam_user" "foo" {
  login = "%s"
}

data "yandex_iam_user" "bar" {
  login = "%s"
}
`, login1, login2)
}

func testAccCheckYandexUserID(userID1, userID2 string) string {
	return fmt.Sprintf(`
data "yandex_iam_user" "foo" {
  user_id = "%s"
}

data "yandex_iam_user" "bar" {
  user_id = "%s"
}
`, userID1, userID2)
}
