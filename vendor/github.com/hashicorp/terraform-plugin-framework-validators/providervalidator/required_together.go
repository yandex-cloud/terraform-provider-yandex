package providervalidator

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/internal/configvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// RequiredTogether checks that a set of path.Expression either has all known
// or all null values.
func RequiredTogether(expressions ...path.Expression) provider.ConfigValidator {
	return &configvalidator.RequiredTogetherValidator{
		PathExpressions: expressions,
	}
}
