package validate

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func FolderID(folderID basetypes.StringValue, config *config.State) (string, diag.Diagnostic) {
	if folderID.ValueString() != "" {
		return folderID.ValueString(), nil
	}

	if config.FolderID.ValueString() != "" {
		return config.FolderID.ValueString(), nil
	}

	return "", diag.NewErrorDiagnostic(
		"Failed to determine folder_id",
		"Error while determine folder_id: please set 'folder_id' key in this resource or at provider level",
	)
}

func NetworkId(networkID basetypes.StringValue, config *config.State) (string, diag.Diagnostic) {
	if config.Endpoint.ValueString() == common.DefaultEndpoint && len(networkID.ValueString()) == 0 {
		return "", diag.NewErrorDiagnostic(
			"Failed to validate network_id field",
			"Error while validating network_id field: empty network_id field",
		)
	}
	return networkID.ValueString(), nil
}
