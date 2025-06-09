package mdb_redis_user_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	redisClusterResourceName   = "yandex_mdb_redis_cluster_v2.foo"
	redisUserResourceNameAlice = "yandex_mdb_redis_user.alice"
	redisUserResourceNameBob   = "yandex_mdb_redis_user.bob"

	VPCDependencies = `
	resource "yandex_vpc_network" "foo" {}
	
	resource "yandex_vpc_subnet" "foo" {
	  zone           = "ru-central1-a"
	  network_id     = yandex_vpc_network.foo.id
	  v4_cidr_blocks = ["10.1.0.0/24"]
	}
	`
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

type Permissions struct {
	Commands        string
	Categories      string
	Patterns        string
	PubSubChannels  string
	SanitizePayload string
}

// Test that a Redis User can be created, updated and destroyed
func TestAccMDBRedisUser_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-redis-user")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBRedisUserConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(redisUserResourceNameAlice, "name", "alice"),
					testAccCheckMDBRedisUserComparePermissions(t, "alice",
						Permissions{
							PubSubChannels:  "resetchannels",
							Patterns:        "allkeys",
							SanitizePayload: "sanitize-payload",
						}),
				),
			},
			mdbRedisUserImportStep(redisUserResourceNameAlice),
			{
				Config: testAccMDBRedisUserConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(redisUserResourceNameBob, "name", "bob"),
					testAccCheckMDBRedisUserComparePermissions(t, "bob",
						Permissions{
							Commands:        "+ping -set",
							Categories:      "-@all +@geo",
							Patterns:        "~456*",
							PubSubChannels:  "&123*",
							SanitizePayload: "sanitize-payload",
						}),
				),
			},
			mdbRedisUserImportStep(redisUserResourceNameAlice),
			mdbRedisUserImportStep(redisUserResourceNameBob),
			{
				Config: testAccMDBRedisUserConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(redisUserResourceNameAlice, "name", "alice"),
					testAccCheckMDBRedisUserComparePermissions(t, "alice",
						Permissions{
							Commands:        "+get",
							Categories:      "-@admin",
							Patterns:        "~4242*",
							PubSubChannels:  "&4242*",
							SanitizePayload: "skip-sanitize-payload",
						}),
				),
			},
			mdbRedisUserImportStep(redisUserResourceNameAlice),
		},
	})
}

func mdbRedisUserImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"passwords", // passwords are not returned
		},
	}
}

func testAccLoadRedisUser(s *terraform.State, username string) (*redis.User, error) {
	rs, ok := s.RootModule().Resources[redisClusterResourceName]

	if !ok {
		return nil, fmt.Errorf("resource %q not found", redisUserResourceNameAlice)
	}
	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no ID is set")
	}

	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	return config.SDK.MDB().Redis().User().Get(context.Background(), &redis.GetUserRequest{
		ClusterId: rs.Primary.ID,
		UserName:  username,
	})
}

func testAccCheckMDBRedisUserComparePermissions(t *testing.T, username string, expected Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		user, err := testAccLoadRedisUser(s, username)
		if err != nil {
			return err
		}
		actual := user.Permissions

		assert.Equal(t, expected.Commands, actual.Commands.GetValue())
		assert.Equal(t, expected.Categories, actual.Categories.GetValue())
		assert.Equal(t, expected.Patterns, actual.Patterns.GetValue())
		assert.Equal(t, expected.PubSubChannels, actual.PubSubChannels.GetValue())
		assert.Equal(t, expected.SanitizePayload, actual.SanitizePayload.GetValue())

		return nil
	}
}

func testAccMDBRedisUserConfigStep0(name string) string {
	return fmt.Sprintf(VPCDependencies+`
resource "yandex_mdb_redis_cluster_v2" "foo" {
	name        = "%s"
	description = "Redis User Terraform Test"
	environment = "PRESTABLE"
	network_id  = yandex_vpc_network.foo.id

	config = {
		password = "mySecre4tP@ssw0rd"
	    version = "7.2"
	}

	resources = {
    	resource_preset_id = "hm1.nano"
    	disk_size          = 16
  	}

	hosts = {
		"aaa" = {
			zone      = "ru-central1-a"
			subnet_id  = yandex_vpc_subnet.foo.id
		}
	}
}
`, name)
}

// Create cluster, user and database
func testAccMDBRedisUserConfigStep1(name string) string {
	return testAccMDBRedisUserConfigStep0(name) + `
resource "yandex_mdb_redis_user" "alice" {
	cluster_id = yandex_mdb_redis_cluster_v2.foo.id
	name       = "alice"
	passwords   = ["mysecureP@ssw0rd"]
	permissions = { pub_sub_channels = "resetchannels" }
}`
}

// Create another user and give permission to database
func testAccMDBRedisUserConfigStep2(name string) string {
	return testAccMDBRedisUserConfigStep1(name) + `
resource "yandex_mdb_redis_user" "bob" {
	cluster_id = yandex_mdb_redis_cluster_v2.foo.id
	name        = "bob"
	passwords   = ["mysecureP@ssw0rd"]
	permissions = {
    	commands = "+ping -set"
    	categories = "-@all +@geo"
		patterns = "~456*"
		pub_sub_channels = "&123*"
		sanitize_payload = "sanitize-payload"
  	}
	enabled = false
}`
}

// Change Alice's permissions
func testAccMDBRedisUserConfigStep3(name string) string {
	return testAccMDBRedisUserConfigStep0(name) + `
resource "yandex_mdb_redis_user" "alice" {
	cluster_id = yandex_mdb_redis_cluster_v2.foo.id
	name        = "alice"
	passwords   = ["mysecureP@ssw0rd"]
	permissions = {
    	commands = "+get"
    	categories = "-@admin"
		patterns = "~4242*"
		pub_sub_channels = "&4242*"
		sanitize_payload = "skip-sanitize-payload"
  	}
}`
}
