package a

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

var _ = schema.StringAttribute{} // want "description field should be configured"

var _ = schema.Int64Attribute{
	Description: "This is a valid description",
}

var _ = schema.ObjectAttribute{} // want "description field should be configured"

var _ = schema.ListAttribute{
	Description: "Another valid attribute",
}
