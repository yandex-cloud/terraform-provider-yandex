package trino_cluster

import (
	"context"
	"testing"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestAdditionalPropertiesResourceSchema(t *testing.T) {
	ctx := context.Background()
	schema := ClusterResourceSchema(ctx)
	diags := schema.ValidateImplementation(ctx)
	require.False(t, diags.HasError(), "%v", diags)
	properties := schema.Attributes["additional_properties"].(rsschema.MapAttribute)
	require.True(t, properties.IsOptional())
	require.False(t, properties.IsComputed())
	require.Equal(t, types.StringType, properties.ElementType)
}

func TestAdditionalPropertiesDataSourceSchema(t *testing.T) {
	ctx := context.Background()
	schema := ClusterDataSourceSchema(ctx)
	diags := schema.ValidateImplementation(ctx)
	require.False(t, diags.HasError(), "%v", diags)
	properties := schema.Attributes["additional_properties"].(dsschema.MapAttribute)
	require.True(t, properties.IsComputed())
	require.False(t, properties.IsOptional())
	require.Equal(t, types.StringType, properties.ElementType)
}
