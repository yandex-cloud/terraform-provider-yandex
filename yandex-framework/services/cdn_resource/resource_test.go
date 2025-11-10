package cdn_resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

// TestGetFolderID_FromModel verifies that folder_id is taken from the resource model when set
func TestGetFolderID_FromModel(t *testing.T) {
	var diags diag.Diagnostics

	resource := &cdnResourceResource{
		providerConfig: &config.Config{
			ProviderState: config.State{
				FolderID: types.StringValue("provider-folder-123"),
			},
		},
	}

	model := &CDNResourceModel{
		FolderID: types.StringValue("resource-folder-456"),
	}

	result := resource.getFolderID(model, &diags)

	require.False(t, diags.HasError(), "getFolderID should not produce errors")
	assert.Equal(t, "resource-folder-456", result,
		"folder_id should be taken from resource model, not provider config")
}

// TestGetFolderID_FromProvider verifies that folder_id falls back to provider config when not set in resource
func TestGetFolderID_FromProvider(t *testing.T) {
	var diags diag.Diagnostics

	resource := &cdnResourceResource{
		providerConfig: &config.Config{
			ProviderState: config.State{
				FolderID: types.StringValue("provider-folder-123"),
			},
		},
	}

	model := &CDNResourceModel{
		FolderID: types.StringNull(), // Not set in resource
	}

	result := resource.getFolderID(model, &diags)

	require.False(t, diags.HasError(), "getFolderID should not produce errors")
	assert.Equal(t, "provider-folder-123", result,
		"folder_id should fall back to provider config when not set in resource")
}

// TestGetFolderID_EmptyInModelTakesProvider verifies that empty string is treated as not set
func TestGetFolderID_EmptyInModelTakesProvider(t *testing.T) {
	var diags diag.Diagnostics

	resource := &cdnResourceResource{
		providerConfig: &config.Config{
			ProviderState: config.State{
				FolderID: types.StringValue("provider-folder-123"),
			},
		},
	}

	model := &CDNResourceModel{
		FolderID: types.StringValue(""), // Empty string in resource
	}

	result := resource.getFolderID(model, &diags)

	require.False(t, diags.HasError(), "getFolderID should not produce errors")
	assert.Equal(t, "provider-folder-123", result,
		"empty folder_id in resource should fall back to provider config")
}

// TestGetFolderID_Missing verifies that error is produced when folder_id is not set anywhere
func TestGetFolderID_Missing(t *testing.T) {
	var diags diag.Diagnostics

	resource := &cdnResourceResource{
		providerConfig: &config.Config{
			ProviderState: config.State{
				FolderID: types.StringNull(), // Not set in provider
			},
		},
	}

	model := &CDNResourceModel{
		FolderID: types.StringNull(), // Not set in resource
	}

	result := resource.getFolderID(model, &diags)

	require.True(t, diags.HasError(), "getFolderID should produce error when folder_id is not set anywhere")
	assert.Equal(t, "", result, "folder_id should be empty string when error occurs")
	assert.Contains(t, diags.Errors()[0].Summary(), "folder_id is required",
		"error message should indicate folder_id is required")
}

// TestGetFolderID_BothMissing verifies error when both resource and provider have empty folder_id
func TestGetFolderID_BothMissing(t *testing.T) {
	var diags diag.Diagnostics

	resource := &cdnResourceResource{
		providerConfig: &config.Config{
			ProviderState: config.State{
				FolderID: types.StringValue(""), // Empty in provider
			},
		},
	}

	model := &CDNResourceModel{
		FolderID: types.StringValue(""), // Empty in resource
	}

	result := resource.getFolderID(model, &diags)

	require.True(t, diags.HasError(), "getFolderID should produce error when folder_id is empty everywhere")
	assert.Equal(t, "", result, "folder_id should be empty string when error occurs")
}
