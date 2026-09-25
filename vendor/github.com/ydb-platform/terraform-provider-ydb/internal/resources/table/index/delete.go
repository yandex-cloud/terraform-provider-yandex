package index

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"

	tbl "github.com/ydb-platform/terraform-provider-ydb/internal/table"
)

type dropIndexParams struct {
	name             string
	databaseEndpoint string
	tablePath        string
}

func (h *handler) dropIndex(ctx context.Context, params dropIndexParams) diag.Diagnostics {
	db, err := tbl.CreateDBConnection(ctx, tbl.ClientParams{
		DatabaseEndpoint: params.databaseEndpoint,
		AuthCreds:        h.authCreds,
	})
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "failed to initialize table client",
				Detail:   err.Error(),
			},
		}
	}
	defer func() {
		_ = db.Close(ctx)
	}()

	err = db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		return s.ExecuteSchemeQuery(ctx, prepareDropRequest(params.tablePath, params.name))
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func (h *handler) Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	indexResource, err := indexResourceSchemaToIndexResource(d)
	if err != nil {
		return diag.FromErr(err)
	}
	return h.dropIndex(ctx, dropIndexParams{
		name:             indexResource.Name,
		databaseEndpoint: indexResource.getConnectionString(),
		tablePath:        indexResource.getTablePath(),
	})
}
