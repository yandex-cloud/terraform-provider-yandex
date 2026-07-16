package trino_catalog

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	trino "github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
)

func TestIcebergMetastoreToAPI(t *testing.T) {
	t.Run("hive uri", func(t *testing.T) {
		m, diags := icebergMetastoreToAPI(MetastoreIceberg{
			Uri:      types.StringValue("thrift://example.net:9083"),
			Protocol: types.StringNull(),
			RestUri:  types.StringNull(),
		})
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}
		hive := m.GetHive()
		if hive == nil {
			t.Fatal("expected hive metastore")
		}
		if hive.GetUri() != "thrift://example.net:9083" {
			t.Fatalf("uri = %q", hive.GetUri())
		}
		if hive.GetProtocol() != nil {
			t.Fatal("protocol should be nil when not set")
		}
	})

	t.Run("hive managed cluster id", func(t *testing.T) {
		m, diags := icebergMetastoreToAPI(MetastoreIceberg{
			ManagedClusterId: types.StringValue("mcid1"),
		})
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}
		hive := m.GetHive()
		if hive.GetManagedClusterId() != "mcid1" {
			t.Fatalf("managed_cluster_id = %q", hive.GetManagedClusterId())
		}
		if _, ok := hive.GetConnection().(*trino.Metastore_HiveMetastore_Uri); ok {
			t.Fatal("expected managed_cluster_id connection, got uri")
		}
	})

	t.Run("hive uri with rest protocol", func(t *testing.T) {
		m, diags := icebergMetastoreToAPI(MetastoreIceberg{
			Uri:      types.StringValue("http://rest.example.net"),
			Protocol: types.StringValue("rest"),
			RestUri:  types.StringNull(),
		})
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}
		if _, ok := m.GetHive().GetProtocol().GetType().(*trino.Metastore_HiveMetastore_Protocol_Rest); !ok {
			t.Fatal("expected iceberg-rest protocol")
		}
	})

	t.Run("hive managed cluster id with rest protocol", func(t *testing.T) {
		m, diags := icebergMetastoreToAPI(MetastoreIceberg{
			ManagedClusterId: types.StringValue("mcid1"),
			Protocol:         types.StringValue("rest"),
		})
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}
		hive := m.GetHive()
		if hive.GetManagedClusterId() != "mcid1" {
			t.Fatalf("managed_cluster_id = %q", hive.GetManagedClusterId())
		}
		if _, ok := hive.GetProtocol().GetType().(*trino.Metastore_HiveMetastore_Protocol_Rest); !ok {
			t.Fatal("expected iceberg-rest protocol together with managed_cluster_id")
		}
	})

	t.Run("rest metastore", func(t *testing.T) {
		m, diags := icebergMetastoreToAPI(MetastoreIceberg{
			Uri:      types.StringNull(),
			Protocol: types.StringNull(),
			RestUri:  types.StringValue("https://rest-catalog.example.net"),
		})
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}
		rest := m.GetRest()
		if rest == nil {
			t.Fatal("expected rest metastore")
		}
		if rest.GetUri() != "https://rest-catalog.example.net" {
			t.Fatalf("rest uri = %q", rest.GetUri())
		}
		if rest.GetAuthorization().GetNone() == nil {
			t.Fatal("expected none authorization")
		}
	})

	t.Run("rest_uri with protocol is an error", func(t *testing.T) {
		_, diags := icebergMetastoreToAPI(MetastoreIceberg{
			Uri:      types.StringNull(),
			Protocol: types.StringValue("rest"),
			RestUri:  types.StringValue("https://rest-catalog.example.net"),
		})
		if !diags.HasError() {
			t.Fatal("expected error for protocol set together with rest_uri")
		}
	})
}

func TestIcebergMetastoreToModel(t *testing.T) {
	ctx := context.Background()
	nullState := types.ObjectNull(MetastoreIcebergT.AttrTypes)

	t.Run("hive managed cluster id", func(t *testing.T) {
		obj, diags := icebergMetastoreToModel(ctx, &trino.Metastore{
			Type: &trino.Metastore_Hive{Hive: &trino.Metastore_HiveMetastore{
				Connection: &trino.Metastore_HiveMetastore_ManagedClusterId{ManagedClusterId: "mcid1"},
			}},
		}, nullState)
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}
		var m MetastoreIceberg
		obj.As(ctx, &m, baseOptions)
		if m.ManagedClusterId.ValueString() != "mcid1" {
			t.Fatalf("managed_cluster_id = %q", m.ManagedClusterId.ValueString())
		}
		if !m.Uri.IsNull() {
			t.Fatal("uri should be null")
		}
	})

	t.Run("hive uri", func(t *testing.T) {
		obj, diags := icebergMetastoreToModel(ctx, &trino.Metastore{
			Type: &trino.Metastore_Hive{Hive: &trino.Metastore_HiveMetastore{
				Connection: &trino.Metastore_HiveMetastore_Uri{Uri: "u"},
			}},
		}, nullState)
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}
		var m MetastoreIceberg
		obj.As(ctx, &m, baseOptions)
		if m.Uri.ValueString() != "u" {
			t.Fatalf("uri = %q", m.Uri.ValueString())
		}
		if !m.RestUri.IsNull() {
			t.Fatal("rest_uri should be null")
		}
		if !m.Protocol.IsNull() {
			t.Fatal("protocol should be null for thrift default when unset")
		}
	})

	t.Run("hive uri with rest protocol", func(t *testing.T) {
		obj, diags := icebergMetastoreToModel(ctx, &trino.Metastore{
			Type: &trino.Metastore_Hive{Hive: &trino.Metastore_HiveMetastore{
				Connection: &trino.Metastore_HiveMetastore_Uri{Uri: "u"},
				Protocol: &trino.Metastore_HiveMetastore_Protocol{
					Type: &trino.Metastore_HiveMetastore_Protocol_Rest{
						Rest: &trino.Metastore_HiveMetastore_Protocol_IcebergRest{},
					},
				},
			}},
		}, nullState)
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}
		var m MetastoreIceberg
		obj.As(ctx, &m, baseOptions)
		if m.Protocol.ValueString() != "rest" {
			t.Fatalf("protocol = %q", m.Protocol.ValueString())
		}
	})

	t.Run("hive managed cluster id with rest protocol", func(t *testing.T) {
		obj, diags := icebergMetastoreToModel(ctx, &trino.Metastore{
			Type: &trino.Metastore_Hive{Hive: &trino.Metastore_HiveMetastore{
				Connection: &trino.Metastore_HiveMetastore_ManagedClusterId{ManagedClusterId: "mcid1"},
				Protocol: &trino.Metastore_HiveMetastore_Protocol{
					Type: &trino.Metastore_HiveMetastore_Protocol_Rest{
						Rest: &trino.Metastore_HiveMetastore_Protocol_IcebergRest{},
					},
				},
			}},
		}, nullState)
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}
		var m MetastoreIceberg
		obj.As(ctx, &m, baseOptions)
		if m.ManagedClusterId.ValueString() != "mcid1" {
			t.Fatalf("managed_cluster_id = %q", m.ManagedClusterId.ValueString())
		}
		if m.Protocol.ValueString() != "rest" {
			t.Fatalf("protocol = %q", m.Protocol.ValueString())
		}
		if !m.Uri.IsNull() {
			t.Fatal("uri should be null")
		}
	})

	t.Run("rest metastore", func(t *testing.T) {
		obj, diags := icebergMetastoreToModel(ctx, &trino.Metastore{
			Type: &trino.Metastore_Rest{Rest: &trino.Metastore_RestMetastore{Uri: "r"}},
		}, nullState)
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}
		var m MetastoreIceberg
		obj.As(ctx, &m, baseOptions)
		if m.RestUri.ValueString() != "r" {
			t.Fatalf("rest_uri = %q", m.RestUri.ValueString())
		}
		if !m.Uri.IsNull() {
			t.Fatal("uri should be null")
		}
	})
}
