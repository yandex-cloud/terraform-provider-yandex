//go:build tf1_12

package mdb_redis_cluster_v2_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccMDBRedisClusterV2PasswordWo_TF1_12(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-redis-v2-password-wo")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { test.AccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBRedisClusterV2PasswordWoConfig(t, clusterName, "initialP@ssw0rd", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(redisResource, "config.password"),
					resource.TestCheckNoResourceAttr(redisResource, "config.password_wo"),
					resource.TestCheckResourceAttr(redisResource, "config.password_wo_version", "1"),
				),
			},
			{
				Config:             testAccMDBRedisClusterV2PasswordWoConfig(t, clusterName, "initialP@ssw0rd", 1),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:      testAccMDBRedisClusterV2PasswordFieldsConfig(t, clusterName, newPtr("legacyP@ssw0rd"), newPtr("writeOnlyP@ssw0rd"), newPtr(1)),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`one \(and only one\)`),
			},
			{
				Config:      testAccMDBRedisClusterV2PasswordFieldsConfig(t, clusterName, nil, newPtr("writeOnlyP@ssw0rd"), nil),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config:      testAccMDBRedisClusterV2PasswordFieldsConfig(t, clusterName, nil, nil, newPtr(1)),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified when`),
			},
			{
				Config: testAccMDBRedisClusterV2PasswordWoConfig(t, clusterName, "rotatedP@ssw0rd", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(redisResource, "config.password"),
					resource.TestCheckNoResourceAttr(redisResource, "config.password_wo"),
					resource.TestCheckResourceAttr(redisResource, "config.password_wo_version", "2"),
				),
			},
		},
	})
}

func testAccMDBRedisClusterV2PasswordWoConfig(t *testing.T, name, password string, version int) string {
	return testAccMDBRedisClusterV2PasswordFieldsConfig(t, name, nil, &password, &version)
}

func testAccMDBRedisClusterV2PasswordFieldsConfig(t *testing.T, name string, password, passwordWo *string, passwordWoVersion *int) string {
	return makeConfig(t, &redisConfigTest{
		Name:        &name,
		Environment: newPtr("PRESTABLE"),
		Resources: &hostResource{
			ResourcePresetId: newPtr("hm3-c2-m8"),
			DiskSize:         newPtr(16),
			DiskTypeId:       newPtr("network-ssd"),
		},
		Hosts: map[string]host{
			"aaa": {Zone: newPtr("ru-central1-d"), SubnetId: newPtr("${yandex_vpc_subnet.foo.id}")},
		},
		Config: &config{
			Password:          password,
			PasswordWo:        passwordWo,
			PasswordWoVersion: passwordWoVersion,
			Version:           newPtr("9.1-valkey"),
		},
	})
}
