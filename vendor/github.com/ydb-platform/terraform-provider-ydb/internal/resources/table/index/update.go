package index

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func (h *handler) Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// NOTE(shmel1k@): currently all parameters are 'force new', so only read can be done here.
	err := h.dropIndex(ctx, dropIndexParams{
		name:             d.Get("name").(string),
		databaseEndpoint: d.Get("connection_string").(string),
		tablePath:        d.Get("table_path").(string),
	})
	if err != nil {
		return err
	}

	return h.Create(ctx, d, meta)
}
