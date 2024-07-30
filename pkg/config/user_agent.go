package config

import (
	"fmt"
	"strings"

	"github.com/yandex-cloud/terraform-provider-yandex/version"
)

const terraformURL = "https://www.terraform.io"

func BuildUserAgent(terraformVersion string, sweeper bool) string {
	// version is set during the release process to the release version of the binary
	// https://www.terraform.io/docs/configuration/providers.html#plugin-names-and-versions
	if sweeper {
		return "Terraform Sweeper"
	}

	return fmt.Sprintf("Terraform/%s (%s) terraform-provider-yandex/%s",
		terraformVersion, terraformURL, strings.TrimPrefix(version.ProviderVersion, "v"))
}
