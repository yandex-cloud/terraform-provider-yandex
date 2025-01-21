package objectid

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

type objectResolverFunc func(name string, opts ...sdkresolvers.ResolveOption) ycsdk.Resolver

// this function can be only used to resolve objects that belong to some folder (have folder_id attribute)
// do not use this function to resolve cloud (or similar objects) ID by name.
func ResolveByNameAndFolderID(ctx context.Context, sdk *ycsdk.SDK, folderID, name string, resolverFunc objectResolverFunc) (string, diag.Diagnostic) {
	if folderID == "" {
		return "", diag.NewErrorDiagnostic(
			"Failed to resolve object ID",
			"Non empty folder_id should be provided")
	}

	if name == "" {
		return "", diag.NewErrorDiagnostic(
			"Failed to resolve object ID",
			"Non empty name should be provided")
	}

	var objectID string
	resolver := resolverFunc(name, sdkresolvers.Out(&objectID), sdkresolvers.FolderID(folderID))

	err := sdk.Resolve(ctx, resolver)

	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to resolve object ID",
			"Error while resolve object id: "+err.Error())
	}

	return objectID, nil
}
