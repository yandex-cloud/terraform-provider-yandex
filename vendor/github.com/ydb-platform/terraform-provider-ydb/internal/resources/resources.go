package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Handler interface {
	Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
	Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
	Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
	Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics
}
