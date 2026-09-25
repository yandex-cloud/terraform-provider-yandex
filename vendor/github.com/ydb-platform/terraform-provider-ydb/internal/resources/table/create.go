package table

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"

	tbl "github.com/ydb-platform/terraform-provider-ydb/internal/table"
)

func (h *handler) Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		DatabaseEndpoint: tableResource.DatabaseEndpoint,
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

	q := PrepareCreateRequest(tableResource)
	err = db.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		return s.ExecuteSchemeQuery(ctx, q)
	})
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "failed to create table",
				Detail:   err.Error(),
			},
		}
	}

	id := tableResource.DatabaseEndpoint + "?path=" + tableResource.Path
	d.SetId(id)

	return h.Read(ctx, d, meta)
}
