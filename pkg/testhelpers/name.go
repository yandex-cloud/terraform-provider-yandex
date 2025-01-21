package testhelpers

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

var testPrefix = "tf-test"

func GenerateNameForResource(suffixLen int) string {
	return fmt.Sprintf("%s-%s", testPrefix, acctest.RandString(suffixLen))
}

func IsTestResource(name string) bool {
	return strings.HasPrefix(name, testPrefix)
}

func TestPrefix() string {
	return testPrefix
}
