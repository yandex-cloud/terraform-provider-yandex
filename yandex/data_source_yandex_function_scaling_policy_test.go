package yandex

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

var functionScalingPolicyDataSource = "data.yandex_function_scaling_policy.test-function-scaling"

func TestAccDataSourceYandexFunctionScalingPolicy(t *testing.T) {
	t.Parallel()

	var function functions.Function
	var policies []*functions.ScalingPolicy

	functionName := acctest.RandomWithPrefix("tf-function")
	instancesLimit := acctest.RandIntRange(10, 100)
	requestsLimit := acctest.RandIntRange(10, 100)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			createYandexFunctionWithTagTestStep(functionName, "my_tag", &function),
			{
				Config: testYandexFunctionScalingPolicyDataSource(functionName, instancesLimit+10, requestsLimit+10, "my_tag", instancesLimit, requestsLimit),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionScalingPolicyExists(functionScalingPolicyResource, 2, &policies),
					resource.TestCheckResourceAttrSet(functionScalingPolicyDataSource, "function_id"),
					testYandexFunctionScalingPolicyDataSourceLimits(functionScalingPolicyDataSource, &policies),
				),
			},
		},
	})
}

func testYandexFunctionScalingPolicyDataSourceLimits(name string, policies *[]*functions.ScalingPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		attributes := rs.Primary.Attributes

		policiesCount, err := strconv.Atoi(attributes["policy.#"])
		if err != nil {
			return err
		}
		if policiesCount != len(*policies) {
			return fmt.Errorf("Incorrect scaling policies count: expected '%d' but found '%d'", len(*policies), policiesCount)
		}

	main:
		for i := 0; i < policiesCount; i++ {
			tag := attributes["policy."+strconv.Itoa(i)+".tag"]
			instancesLimit := attributes["policy."+strconv.Itoa(i)+".zone_instances_limit"]
			requestsLimit := attributes["policy."+strconv.Itoa(i)+".zone_requests_limit"]
			for _, policy := range *policies {
				if policy.Tag == tag {
					if strconv.Itoa(int(policy.ZoneInstancesLimit)) != instancesLimit {
						return fmt.Errorf("Incorrect scaling policy '%s' ZoneInstancesLimit: expected '%d' but found '%s'", tag, policy.ZoneInstancesLimit, instancesLimit)
					}
					if strconv.Itoa(int(policy.ZoneRequestsLimit)) != requestsLimit {
						return fmt.Errorf("Incorrect scaling policy '%s' ZoneRequestsLimit: expected '%d' but found '%s'", tag, policy.ZoneRequestsLimit, requestsLimit)
					}
					continue main
				}
			}
			return fmt.Errorf("Unexpected scaling policy with tag '%s'", tag)
		}

		return nil
	}
}

func testYandexFunctionScalingPolicyDataSource(functionName string, latestInstancesLimit int, latestRequestsLimit int, tag string, tagInstancesLimit int, tagRequestsLimit int) string {
	return fmt.Sprintf(`
resource "yandex_function" "test-function" {
  name       = "%s"
  user_hash  = "user_hash_new"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "test-fixtures/serverless/main.zip"
  }
}

resource "yandex_function_scaling_policy" "test-function-scaling" {
  function_id = yandex_function.test-function.id
  policy {
    tag = "$latest"
    zone_instances_limit = %d
    zone_requests_limit  = %d
  }
  policy {
    tag = "%s"
    zone_instances_limit = %d
    zone_requests_limit  = %d
  }
}

data "yandex_function_scaling_policy" "test-function-scaling" {
  function_id = yandex_function_scaling_policy.test-function-scaling.function_id
}
	`, functionName, latestInstancesLimit, latestRequestsLimit, tag, tagInstancesLimit, tagRequestsLimit)
}
