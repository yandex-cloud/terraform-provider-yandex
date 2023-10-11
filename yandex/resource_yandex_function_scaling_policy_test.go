package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

var functionScalingPolicyResource = "yandex_function_scaling_policy.test-function-scaling"

func TestAccYandexFunctionScalingPolicy_single(t *testing.T) {
	t.Parallel()

	var policies []*functions.ScalingPolicy
	functionName := acctest.RandomWithPrefix("tf-function")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			singleYandexFunctionScalingPolicyTestStep(functionName, 2, 3, &policies),
			singleYandexFunctionScalingPolicyTestStep(functionName, 5, 6, &policies),
		},
	})
}

func TestAccYandexFunctionScalingPolicy_multiple(t *testing.T) {
	t.Parallel()

	var function functions.Function
	var policies []*functions.ScalingPolicy

	functionName := acctest.RandomWithPrefix("tf-function")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			createYandexFunctionWithTagTestStep(functionName, "my_tag", &function),
			multipleYandexFunctionScalingPolicyTestStep(functionName, "my_tag", 2, 3, &policies),
			multipleYandexFunctionScalingPolicyTestStep(functionName, "my_tag", 5, 6, &policies),
		},
	})
}

func createYandexFunctionWithTagTestStep(functionName, tag string, function *functions.Function) resource.TestStep {
	return resource.TestStep{
		Config: testYandexFunctionWithTag(functionName, tag),
		Check: resource.ComposeTestCheckFunc(
			testYandexFunctionExists(functionResource, function),
			resource.TestCheckResourceAttr(functionResource, "name", functionName),
		),
	}
}

func singleYandexFunctionScalingPolicyTestStep(functionName string, instancesLimit int, requestsLimit int, policies *[]*functions.ScalingPolicy) resource.TestStep {
	return resource.TestStep{
		Config: testYandexFunctionScalingPolicySingle(functionName, instancesLimit, requestsLimit),
		Check: resource.ComposeTestCheckFunc(
			testYandexFunctionScalingPolicyExists(functionScalingPolicyResource, 1, policies),
			testYandexFunctionScalingPolicyLimits(policies, "$latest", instancesLimit, requestsLimit),
		),
	}
}

func multipleYandexFunctionScalingPolicyTestStep(functionName, prevVersionTag string, instancesLimit int, requestsLimit int, policies *[]*functions.ScalingPolicy) resource.TestStep {
	return resource.TestStep{
		Config: testYandexFunctionScalingPolicyDouble(functionName, instancesLimit+10, requestsLimit+10, prevVersionTag, instancesLimit, requestsLimit),
		Check: resource.ComposeTestCheckFunc(
			testYandexFunctionScalingPolicyExists(functionScalingPolicyResource, 2, policies),
			testYandexFunctionScalingPolicyLimits(policies, "$latest", instancesLimit+10, requestsLimit+10),
			testYandexFunctionScalingPolicyLimits(policies, prevVersionTag, instancesLimit, requestsLimit),
		),
	}
}

func testYandexFunctionScalingPolicyExists(name string, expectedPoliciesCount int, policies *[]*functions.ScalingPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		functionID := rs.Primary.Attributes["function_id"]

		found, err := fetchFunctionScalingPolicies(context.Background(), config, functionID)
		if err != nil {
			return err
		}

		if len(found) != expectedPoliciesCount {
			return fmt.Errorf("Incorrect scaling policies count: expected '%d' but found '%d'", expectedPoliciesCount, len(found))
		}

		*policies = found
		return nil
	}
}

func testYandexFunctionScalingPolicyLimits(policies *[]*functions.ScalingPolicy, tag string, instancesLimit int, requestsLimit int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, policy := range *policies {
			if policy.Tag == tag {
				if policy.ZoneInstancesLimit != int64(instancesLimit) {
					return fmt.Errorf("Incorrect scaling policy '%s' ZoneInstancesLimit: expected '%d' but found '%d'", tag, instancesLimit, policy.ZoneInstancesLimit)
				}
				if policy.ZoneRequestsLimit != int64(requestsLimit) {
					return fmt.Errorf("Incorrect scaling policy '%s' ZoneRequestsLimit: expected '%d' but found '%d'", tag, requestsLimit, policy.ZoneRequestsLimit)
				}
				return nil
			}
		}
		return fmt.Errorf("Scaling policy with tag '%s' not found", tag)
	}
}

func testYandexFunctionWithTag(name string, tag string) string {
	return fmt.Sprintf(`
resource "yandex_function" "test-function" {
  name       = "%s"
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "test-fixtures/serverless/main.zip"
  }
  tags = ["%s"]
}
	`, name, tag)
}

func testYandexFunctionScalingPolicySingle(functionName string, instancesLimit int, requestsLimit int) string {
	return fmt.Sprintf(`
resource "yandex_function" "test-function" {
  name       = "%s"
  user_hash  = "user_hash"
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
}
	`, functionName, instancesLimit, requestsLimit)
}

func testYandexFunctionScalingPolicyDouble(functionName string, latestInstancesLimit int, latestRequestsLimit int, tag string, tagInstancesLimit int, tagRequestsLimit int) string {
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
	`, functionName, latestInstancesLimit, latestRequestsLimit, tag, tagInstancesLimit, tagRequestsLimit)
}
