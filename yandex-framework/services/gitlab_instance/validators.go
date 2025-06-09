package gitlab_instance

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var domainRegex = regexp.MustCompile(`^([a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)\.gitlab\.yandexcloud\.net$`)

func domainValidator() validator.String {
	return stringvalidator.RegexMatches(
		domainRegex,
		"Must be in gitlab.yandexcloud.net DNS zone",
	)
}

func resourcePresetValidator() validator.String {
	return stringvalidator.OneOf("s2.micro", "s2.small", "s2.medium", "s2.large")
}

func arValidator() validator.String {
	return stringvalidator.OneOf("NONE", "BASIC", "STANDARD", "ADVANCED")
}

func nameValidator() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^[a-z][a-z0-9-]{0,61}[a-z0-9]$`),
		"Can contain lowercase Latin letters, numbers, and dashes. The first character must be a letter, and the last character cannot be a dash.",
	)
}

func labelKeysValidator() validator.Map {
	return mapvalidator.KeysAre(
		stringvalidator.LengthBetween(1, 63),
		stringvalidator.RegexMatches(regexp.MustCompile(`[a-z][-_0-9a-z]*`),
			"It can contain lowercase letters of the Latin alphabet, numbers, "+
				"hyphens and underscores. And first character must be letter."),
	)
}

func labelValuesValidator() validator.Map {
	return mapvalidator.ValueStringsAre(
		stringvalidator.LengthBetween(1, 63),
		stringvalidator.RegexMatches(regexp.MustCompile(`[-_0-9a-z]+`),
			"It can contain lowercase letters of the Latin alphabet, numbers, "+
				"hyphens and underscores."),
	)
}
