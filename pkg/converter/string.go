package converter

import "github.com/hashicorp/terraform-plugin-framework/types"

func StringValueWithState(str string, state types.String) types.String {
	if str == "" {
		return state
	}
	return types.StringValue(str)
}

func SetUnknownStringValue(state types.String) types.String {
	if state.IsUnknown() {
		return types.StringNull()
	}
	return state
}
