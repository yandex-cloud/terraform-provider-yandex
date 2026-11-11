package trino_cluster

import (
	"context"
	"testing"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"
)

func TestEventListenersResourceSchema(t *testing.T) {
	ctx := context.Background()
	schema := ClusterResourceSchema(ctx)
	diags := schema.ValidateImplementation(ctx)
	require.False(t, diags.HasError(), "%v", diags)
	listeners := schema.Attributes["event_listeners"].(rsschema.SingleNestedAttribute)
	require.True(t, listeners.IsOptional())
	require.False(t, listeners.IsComputed())
	require.Len(t, listeners.Attributes, 1)
	dataCatalog := listeners.Attributes["data_catalog"].(rsschema.SingleNestedAttribute)
	require.True(t, dataCatalog.IsOptional())
	require.Empty(t, dataCatalog.Attributes)
	require.True(t, listeners.GetType().Equal((EventListenersValue{}).Type(ctx)))
}

func TestEventListenersDataSourceSchema(t *testing.T) {
	ctx := context.Background()
	schema := ClusterDataSourceSchema(ctx)
	diags := schema.ValidateImplementation(ctx)
	require.False(t, diags.HasError(), "%v", diags)
	listeners := schema.Attributes["event_listeners"].(dsschema.SingleNestedAttribute)
	require.True(t, listeners.IsComputed())
	require.Len(t, listeners.Attributes, 1)
	dataCatalog := listeners.Attributes["data_catalog"].(dsschema.SingleNestedAttribute)
	require.True(t, dataCatalog.IsComputed())
	require.Empty(t, dataCatalog.Attributes)
	require.True(t, listeners.GetType().Equal((EventListenersValue{}).Type(ctx)))
}
