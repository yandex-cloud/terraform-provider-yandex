package providervalidator

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/internal/configvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// Conflicting checks that a set of path.Expression, are not configured
// simultaneously.
func Conflicting(expressions ...path.Expression) provider.ConfigValidator {
	return &configvalidator.ConflictingValidator{
		PathExpressions: expressions,
	}
}
