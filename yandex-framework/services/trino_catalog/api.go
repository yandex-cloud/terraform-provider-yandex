package trino_catalog

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-sdk/services/trino/v1"
	"time"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func CreateCatalog(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *trino.CreateCatalogRequest) (string, diag.Diagnostic) {
	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*trinosdk.CatalogCreateOperation, error) {
		return trinosdk.NewCatalogClient(sdk).Create(ctx, req)
	})
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino catalog",
			"Error while requesting API to create Trino catalog: "+err.Error(),
		)
	}

	_, err = op.WaitInterval(ctx, func(int) time.Duration { return 5 * time.Second })
	if err != nil {
		return "", diag.NewErrorDiagnostic(
			"Failed to create Trino catalog",
			"Error while requesting API to create Trino catalog. Failed to wait: "+err.Error(),
		)
	}

	md := op.Metadata()

	return md.CatalogId, nil
}

func GetCatalogByID(ctx context.Context, sdk *ycsdk.SDK, catalogID, cid string) (*trino.Catalog, diag.Diagnostic) {
	catalog, err := trinosdk.NewCatalogClient(sdk).Get(ctx, &trino.GetCatalogRequest{
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

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*trinosdk.CatalogUpdateOperation, error) {
		return trinosdk.NewCatalogClient(sdk).Update(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to update Trino catalog", err.Error())
	}
	return nil
}

func GetCatalogByName(ctx context.Context, sdk *ycsdk.SDK, clusterId, catalogName string) (string, diag.Diagnostic) {
	catalogs, err := trinosdk.NewCatalogClient(sdk).List(ctx, &trino.ListCatalogsRequest{
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

	op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*trinosdk.CatalogDeleteOperation, error) {
		return trinosdk.NewCatalogClient(sdk).Delete(ctx, req)
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if err != nil {
		return diag.NewErrorDiagnostic("Failed to delete Trino catalog", err.Error())
	}
	return nil
}
