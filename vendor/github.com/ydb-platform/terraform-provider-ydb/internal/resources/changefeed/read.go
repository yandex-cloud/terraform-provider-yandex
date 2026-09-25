package changefeed

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ydb "github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"

	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers"
	tbl "github.com/ydb-platform/terraform-provider-ydb/internal/table"
)

func (h *handler) Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cdcResource, err := changefeedResourceSchemaToChangefeedResource(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if cdcResource == nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "got nil resource, unreachable code",
			},
		}
	}
	db, err := tbl.CreateDBConnection(ctx, tbl.ClientParams{
		DatabaseEndpoint: cdcResource.getConnectionString(),
		AuthCreds:        h.authCreds,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	defer func() {
		_ = db.Close(ctx)
	}()

	var description options.Description
	err = db.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		description, err = s.DescribeTable(
			ctx,
			parseTablePathFromCDCEntity(cdcResource.Entity.GetFullEntityPath()),
			options.WithPartitionStats(),
			options.WithShardKeyBounds(),
			options.WithTableStats(),
		)
		return err
	})
	if err != nil {
		if ydb.IsOperationErrorSchemeError(err) {
			// NOTE(shmel1k@): marking as non-existing resource
			d.SetId("")
			return nil
		}
		return diag.Errorf("failed to describe table %q: %s", cdcResource.TablePath, err)
	}

	var cdcDescription options.ChangefeedDescription
	for _, v := range description.Changefeeds {
		if v.Name == cdcResource.Name {
			cdcDescription = v
			break
		}
	}
	if cdcDescription.Name == "" {
		// NOTE(shmel1k@): changefeed was not found.
		d.SetId("")
		return h.Create(ctx, d, meta)
	}

	topicDesc, err := db.Topic().Describe(ctx, helpers.TrimPath(cdcResource.Entity.GetEntityPath()))
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenCDCDescription(d, cdcResource, cdcDescription, topicDesc.Consumers))
}
