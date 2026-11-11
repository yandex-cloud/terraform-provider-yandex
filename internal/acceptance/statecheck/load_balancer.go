package statecheck

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func WithState(fn func() resource.TestCheckFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return fn()(s)
	}
}

func CheckResourceSubAttrFn(resourceName string, path *string, field string, checkfn func(string) error) resource.TestCheckFunc {
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
		value, ok := is.Attributes[fullPath]
		if !ok {
			return fmt.Errorf("Can't find path %s in resource: %s", fullPath, resourceName)
		}

		err := checkfn(value)
		if err != nil {
			return err
		}

		return nil
	}
}

func ExistsElementWithAttrTrimmedValue(resourceName, path, field, value string, fullPath *string) resource.TestCheckFunc {
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
		re, err := regexp.Compile(reStr)
		if err != nil {
			return err
		}
		for k, v := range is.Attributes {
			trimmedValue := strings.TrimSpace(v)
			if re.MatchString(k) && trimmedValue == value {
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

func ExistsFirstElementWithAttr(resourceName, path, field string, fullPath *string) resource.TestCheckFunc {
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

		for k := range is.Attributes {
			reStr := fmt.Sprintf(`(%s\.\d+)\.%s`, path, field)
			re := regexp.MustCompile(reStr)
			if re.MatchString(k) {
				sm := re.FindStringSubmatch(k)
				*fullPath = sm[1]
				return nil
			}
		}

		return fmt.Errorf(
			"Can't find key %s.*.%s in resource: %s", path, field, resourceName,
		)
	}
}
