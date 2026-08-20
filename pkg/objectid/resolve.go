package objectid

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-sdk/v2/pkg/sdkresolvers"
)

type ObjectResolverFunc func(name string, opts ...sdkresolvers.ResolveOption) sdkresolvers.Resolver

// this function can be only used to resolve objects that belong to some folder (have folder_id attribute)
// do not use this function to resolve cloud (or similar objects) ID by name.
func ResolveByNameAndFolderID(ctx context.Context, folderID, name string, resolverFunc ObjectResolverFunc) (string, diag.Diagnostic) {
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

	resolver := resolverFunc(name, sdkresolvers.FolderID(folderID))
	err := resolver.Run(ctx)

	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to resolve object ID",
			"Error while resolve object id: "+err.Error())
	}

	return resolver.ID(), nil
}
