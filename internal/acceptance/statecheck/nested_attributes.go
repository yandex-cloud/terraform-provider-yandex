package statecheck

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func RandomResourceName() string {
	return fmt.Sprintf("tf-test-%s", acctest.RandString(10))
}

func CheckResourceAttrWithValueFactory(name, key string, valueFactory func() string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't get resource '%s'", name)
		}

		instanceState := resourceState.Primary
		if instanceState == nil {
			return fmt.Errorf("there is no primary instance state for '%s'", name)
		}

		expected := valueFactory()
		if expected == "0" && (strings.HasSuffix(key, ".#") || strings.HasSuffix(key, ".%")) {
			return fmt.Errorf("testAccCheckResourceAttrWithValueFactory does not know how to perform empty check =(")
		}

		actual, ok := instanceState.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: there is no '%s' attribute", name, key)
		}

		if expected != actual {
			return fmt.Errorf("%s: expected attribute '%s' to have value '%s', got '%s'", name, key, expected, actual)
		}

		return nil
	}
}

func TestCheckResourceSubAttr(resourceName string, path *string, field string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", resourceName, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", resourceName, ms.Path)
		}

		fullPath := fmt.Sprintf("%s.%s", *path, field)
		actualValue, ok := is.Attributes[fullPath]
		if !ok {
			return fmt.Errorf("Can't find path %s in resource: %s", fullPath, resourceName)
		}

		if actualValue != value {
			return fmt.Errorf(
				"Can't match values for path %s in resource: %s. %s != %s", fullPath, resourceName, value, actualValue,
			)
		}

		return nil
	}
}

func TestExistsElementWithAttrValue(resourceName, path, field, value string, fullPath *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", resourceName, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", resourceName, ms.Path)
		}
		reStr := fmt.Sprintf(`(%s\.\d+)\.%s`, path, field)
		re := regexp.MustCompile(reStr)

		for k, v := range is.Attributes {
			if re.MatchString(k) && v == value {
				sm := re.FindStringSubmatch(k)
				*fullPath = sm[1]
				return nil
			}
		}

		return fmt.Errorf(
			"Can't find key %s.*.%s in resource: %s with value %s", path, field, resourceName, value,
		)
	}
}
