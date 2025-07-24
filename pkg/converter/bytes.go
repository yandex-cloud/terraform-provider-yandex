package converter

import (
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// DecodeBase64 decode string into bytes using base64 std encoding
func DecodeBase64(ts string, diags *diag.Diagnostics) []byte {
	if ts == "" {
		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(ts)
	if err != nil {
		diags.AddError(
			"Failed to decode base64 string",
			fmt.Sprintf("Failed to decode base64 string"),
		)
		return nil
	}

	return decoded
}
