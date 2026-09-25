package providervalidator

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/internal/configvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// ExactlyOneOf checks that a set of path.Expression does not have more than
// one known value.
func ExactlyOneOf(expressions ...path.Expression) provider.ConfigValidator {
	return &configvalidator.ExactlyOneOfValidator{
		PathExpressions: expressions,
	}
}
