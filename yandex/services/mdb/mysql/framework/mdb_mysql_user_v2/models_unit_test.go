package mdb_mysql_user_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysql "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	mdbv1 "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/v1"
)

func TestSpecToStateUsesUserConnectionManagerForLegacyConnectionManager(t *testing.T) {
	var state User
	var diags diag.Diagnostics

	specToState(context.Background(), &mysql.User{
		ClusterId: "cluster",
		Name:      "user",
		UserConnectionManager: &mdbv1.UserConnectionManager{
			ConnectionId: "connection",
		},
	}, &state, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	value, ok := state.ConnectionManager.Elements()["connection_id"]
	if !ok || !value.Equal(types.StringValue("connection")) {
		t.Errorf("legacy connection manager = %v, want connection ID %q", state.ConnectionManager, "connection")
	}
}
