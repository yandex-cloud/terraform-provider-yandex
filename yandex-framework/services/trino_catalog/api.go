package trino_catalog

import (
	"context"
	"fmt"
	"time"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func CreateCatalog(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *trino.CreateCatalogRequest) (string, diag.Diagnostic) {
	op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
		return sdk.Trino().Catalog().Create(ctx, req)
	})
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino catalog",
			"Error while requesting API to create Trino catalog: "+err.Error(),
		)
	}

	err = op.WaitInterval(ctx, 5*time.Second)
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino catalog",
			"Error while requesting API to create Trino catalog. Failed to wait: "+err.Error(),
		)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino catalog",
			"Failed to unmarshal metadata: "+err.Error(),
		)
	}

	md, ok := protoMetadata.(*trino.CreateCatalogMetadata)
	if !ok {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino catalog",
			"Failed to convert response metadata to CreateCatalogMetadata",
		)
	}

	return md.CatalogId, nil
}

func GetCatalogByID(ctx context.Context, sdk *ycsdk.SDK, catalogID, cid string) (*trino.Catalog, diag.Diagnostic) {
	catalog, err := sdk.Trino().Catalog().Get(ctx, &trino.GetCatalogRequest{
		ClusterId: cid,
		CatalogId: catalogID,
	})
	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil, nil
		}

		return nil, diag.NewErrorDiagnostic(
			"Failed to read Trino catalog",
			"Error while requesting API to get Trino catalog: "+err.Error(),
		)
	}
	return catalog, nil
}

func UpdateCatalog(ctx context.Context, sdk *ycsdk.SDK, req *trino.UpdateCatalogRequest) diag.Diagnostic {
	if req == nil || req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil
	}

	return waitOperation(ctx, sdk, "update Trino catalog", func() (*operation.Operation, error) {
		return sdk.Trino().Catalog().Update(ctx, req)
	})
}

func GetCatalogByName(ctx context.Context, sdk *ycsdk.SDK, clusterId, catalogName string) (string, diag.Diagnostic) {
	catalogs, err := sdk.Trino().Catalog().List(ctx, &trino.ListCatalogsRequest{
		ClusterId: clusterId,
	})
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to list Trino catalogs",
			"Error while requesting API to list Trino catalogs: "+err.Error(),
		)
	}

	for _, catalog := range catalogs.Catalogs {
		if catalog.Name == catalogName {
			return catalog.Id, nil
		}
	}

	return "", diag.NewErrorDiagnostic(
		"Catalog not found",
		fmt.Sprintf("Trino catalog with name '%s' not found in cluster %s", catalogName, clusterId),
	)
}

func DeleteCatalog(ctx context.Context, sdk *ycsdk.SDK, catalogID, cid string) diag.Diagnostic {
	req := &trino.DeleteCatalogRequest{
		ClusterId: cid,
		CatalogId: catalogID,
	}

	return waitOperation(ctx, sdk, "delete Trino catalog", func() (*operation.Operation, error) {
		return sdk.Trino().Catalog().Delete(ctx, req)
	})
}

func waitOperation(ctx context.Context, sdk *ycsdk.SDK, action string, callback func() (*operation.Operation, error)) diag.Diagnostic {
	op, err := retry.ConflictingOperation(ctx, sdk, callback)

	if err == nil {
		err = op.Wait(ctx)
	}

	if err != nil {
		return diag.NewErrorDiagnostic(fmt.Sprintf("Failed to %s", action), err.Error())
	}

	return nil
}
