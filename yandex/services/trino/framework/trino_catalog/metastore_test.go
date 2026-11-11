package trino_catalog

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	trino "github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
)

func TestMetastoreToAPI(t *testing.T) {
	t.Run("uri", func(t *testing.T) {
		hive := metastoreToAPI(Metastore{
			Uri:              types.StringValue("thrift://example.net:9083"),
			ManagedClusterId: types.StringNull(),
		}).GetHive()

		if hive.GetUri() != "thrift://example.net:9083" {
			t.Fatalf("uri = %q", hive.GetUri())
		}
		if _, ok := hive.GetConnection().(*trino.Metastore_HiveMetastore_ManagedClusterId); ok {
			t.Fatal("expected uri connection, got managed_cluster_id")
		}
	})

	t.Run("managed cluster id", func(t *testing.T) {
		hive := metastoreToAPI(Metastore{
			Uri:              types.StringNull(),
			ManagedClusterId: types.StringValue("mcid1"),
		}).GetHive()

		if hive.GetManagedClusterId() != "mcid1" {
			t.Fatalf("managed_cluster_id = %q", hive.GetManagedClusterId())
		}
		if _, ok := hive.GetConnection().(*trino.Metastore_HiveMetastore_Uri); ok {
			t.Fatal("expected managed_cluster_id connection, got uri")
		}
	})
}

func TestMetastoreToModel(t *testing.T) {
	ctx := context.Background()

	t.Run("uri", func(t *testing.T) {
		obj, diags := metastoreToModel(ctx, &trino.Metastore{
			Type: &trino.Metastore_Hive{Hive: &trino.Metastore_HiveMetastore{
				Connection: &trino.Metastore_HiveMetastore_Uri{Uri: "u"},
			}},
		})
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}

		var m Metastore
		obj.As(ctx, &m, baseOptions)
		if m.Uri.ValueString() != "u" {
			t.Fatalf("uri = %q", m.Uri.ValueString())
		}
		if !m.ManagedClusterId.IsNull() {
			t.Fatalf("managed_cluster_id should be null, got %q", m.ManagedClusterId.ValueString())
		}
	})

	t.Run("managed cluster id", func(t *testing.T) {
		obj, diags := metastoreToModel(ctx, &trino.Metastore{
			Type: &trino.Metastore_Hive{Hive: &trino.Metastore_HiveMetastore{
				Connection: &trino.Metastore_HiveMetastore_ManagedClusterId{ManagedClusterId: "mcid1"},
			}},
		})
		if diags.HasError() {
			t.Fatalf("diags: %v", diags)
		}

		var m Metastore
		obj.As(ctx, &m, baseOptions)
		if m.ManagedClusterId.ValueString() != "mcid1" {
			t.Fatalf("managed_cluster_id = %q", m.ManagedClusterId.ValueString())
		}
		if !m.Uri.IsNull() {
			t.Fatalf("uri should be null, got %q", m.Uri.ValueString())
		}
	})
}
