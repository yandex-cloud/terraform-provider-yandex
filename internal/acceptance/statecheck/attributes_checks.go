package statecheck

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func ResourceAttrValue(s *terraform.State, resourceName, attr string) (string, error) {
	instanceState, err := PrimaryInstanceState(s, resourceName)
	if err != nil {
		return "", err
	}
	return instanceState.Attributes[attr], nil
}

// primaryInstanceState returns the primary instance state for the given
// resource name in the root module.
func PrimaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return nil, fmt.Errorf("not found: %s in %s", name, s.RootModule().Path)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("no primary instance: %s in %s", name, s.RootModule().Path)
	}

	return is, nil
}

func CheckCreatedAtAttr(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		const createdAtAttrName = "created_at"
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		createdAt, ok := rs.Primary.Attributes[createdAtAttrName]
		if !ok {
			return fmt.Errorf("can't find '%s' attr for %s resource", createdAtAttrName, resourceName)
		}

		if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
			return fmt.Errorf("can't parse timestamp in attr '%s': %s", createdAtAttrName, createdAt)
		}
		return nil
	}
}

func CheckResourceIDField(resourceName string, idFieldName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if rs.Primary.Attributes[idFieldName] != rs.Primary.ID {
			return fmt.Errorf("Resource: %s id field: %s, doesn't match resource ID", resourceName, idFieldName)
		}

		return nil
	}
}

func CheckAttributeWithSuppress(suppressDiff schema.SchemaDiffSuppressFunc, resourceName string, attributePath string, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		startTime, ok := rs.Primary.Attributes[attributePath]
		if !ok {
			return fmt.Errorf("can't find '%s' attr for %s resource", attributePath, resourceName)
		}

		if !suppressDiff("", expectedValue, startTime, nil) {
			return fmt.Errorf("stored value: '%s' doesn't match expected value: '%s'", startTime, expectedValue)
		}

		return nil
	}
}
