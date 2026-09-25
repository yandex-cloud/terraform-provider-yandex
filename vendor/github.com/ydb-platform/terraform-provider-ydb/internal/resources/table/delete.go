package table

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"

	tbl "github.com/ydb-platform/terraform-provider-ydb/internal/table"
)

func (h *handler) Delete(ctx context.Context, d *schema.ResourceData, cfg interface{}) diag.Diagnostics {
	tableResource, err := tableResourceSchemaToTableResource(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if tableResource == nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "got nil resource, unreachable code",
			},
		}
	}

	db, err := tbl.CreateDBConnection(ctx, tbl.ClientParams{
		DatabaseEndpoint: tableResource.getConnectionString(),
		AuthCreds:        h.authCreds,
	})
	if err != nil {
		return diag.Errorf("failed to initialize table client: %s", err)
	}
	defer func() {
		_ = db.Close(ctx)
	}()

	err = db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		query := PrepareDropTableRequest(tableResource.Path)
		return s.ExecuteSchemeQuery(ctx, query)
	})
	if err != nil {
		return diag.Errorf("failed to drop table %q: %s", tableResource.Path, err)
	}
	return nil
}
