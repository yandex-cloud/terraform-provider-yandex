package datalens_workbook

import "github.com/hashicorp/terraform-plugin-framework/types"

// workbookModel is both the Terraform-Framework model and the wire DTO for
// the DataLens REST API. `tfsdk` tags drive schema mapping; `wire` tags drive
// JSON-RPC body shape (snake_case ↔ camelCase). Fields with `wire:"-"` are
// Terraform-only (id is sent via path/header, organization_id goes in headers).
type workbookModel struct {
	Id             types.String `tfsdk:"id"              wire:"workbookId"`
	OrganizationId types.String `tfsdk:"organization_id" wire:"-"`
	CollectionId   types.String `tfsdk:"collection_id"   wire:"collectionId"`
	Title          types.String `tfsdk:"title"           wire:"title"`
	Description    types.String `tfsdk:"description"     wire:"description,nullIfEmpty"`
	TenantId       types.String `tfsdk:"tenant_id"       wire:"tenantId"`
	Status         types.String `tfsdk:"status"          wire:"status"`
	CreatedBy      types.String `tfsdk:"created_by"      wire:"createdBy"`
	CreatedAt      types.String `tfsdk:"created_at"      wire:"createdAt"`
	UpdatedBy      types.String `tfsdk:"updated_by"      wire:"updatedBy"`
	UpdatedAt      types.String `tfsdk:"updated_at"      wire:"updatedAt"`
}
