package mdb_greenplum_cluster_v2

import (
	"context"
	"strings"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

func TestGreenplumClusterPasswordWoSchema(t *testing.T) {
	var resp frameworkresource.SchemaResponse
	NewResource().Schema(context.Background(), frameworkresource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %#v", resp.Diagnostics)
	}
	if resp.Schema.Version != 1 {
		t.Fatalf("schema version = %d, want 1", resp.Schema.Version)
	}

	legacyPassword := resp.Schema.Attributes["user_password"].(schema.StringAttribute)
	if !legacyPassword.IsOptional() || legacyPassword.IsRequired() || !legacyPassword.IsSensitive() || len(legacyPassword.Validators) != 2 {
		t.Fatal("user_password must be optional, sensitive, length-limited, and mutually exclusive with user_password_wo")
	}

	writeOnlyPassword := resp.Schema.Attributes["user_password_wo"].(schema.StringAttribute)
	if !writeOnlyPassword.IsOptional() || !writeOnlyPassword.IsWriteOnly() || !writeOnlyPassword.IsSensitive() || len(writeOnlyPassword.Validators) != 2 {
		t.Fatal("user_password_wo must be optional, write-only, sensitive, length-limited, and require its version")
	}

	version := resp.Schema.Attributes["user_password_wo_version"].(schema.Int64Attribute)
	if !version.IsOptional() || version.IsWriteOnly() || len(version.Validators) != 1 {
		t.Fatal("user_password_wo_version must be an optional state attribute requiring the password")
	}
}

func TestGreenplumClusterPasswordForCreate(t *testing.T) {
	plan := &yandexMdbGreenplumClusterV2Model{UserPassword: types.StringValue("legacy-password")}
	if got := greenplumClusterPasswordForCreate(plan, types.StringValue("write-only-password")); got != "write-only-password" {
		t.Fatalf("password = %q, want write-only-password", got)
	}
	if got := greenplumClusterPasswordForCreate(plan, types.StringNull()); got != "legacy-password" {
		t.Fatalf("password = %q, want legacy-password", got)
	}
}

func TestGreenplumClusterPasswordChange(t *testing.T) {
	t.Run("version change rotates write-only password", func(t *testing.T) {
		state := &yandexMdbGreenplumClusterV2Model{UserPassword: types.StringNull(), UserPasswordWoVersion: types.Int64Value(1)}
		plan := &yandexMdbGreenplumClusterV2Model{UserPassword: types.StringNull(), UserPasswordWoVersion: types.Int64Value(2)}

		password, changed, diags := greenplumClusterPasswordChange(plan, state, types.StringValue("rotated-password"))
		if diags.HasError() || !changed || password != "rotated-password" {
			t.Fatalf("password change = (%q, %t, %#v), want rotated password", password, changed, diags)
		}
	})

	t.Run("same version ignores write-only password", func(t *testing.T) {
		state := &yandexMdbGreenplumClusterV2Model{UserPassword: types.StringNull(), UserPasswordWoVersion: types.Int64Value(1)}
		plan := &yandexMdbGreenplumClusterV2Model{UserPassword: types.StringNull(), UserPasswordWoVersion: types.Int64Value(1)}

		password, changed, diags := greenplumClusterPasswordChange(plan, state, types.StringValue("different-password"))
		if diags.HasError() || changed || password != "" {
			t.Fatalf("password change = (%q, %t, %#v), want no change", password, changed, diags)
		}
	})

	t.Run("version change requires write-only password", func(t *testing.T) {
		state := &yandexMdbGreenplumClusterV2Model{UserPassword: types.StringNull(), UserPasswordWoVersion: types.Int64Value(1)}
		plan := &yandexMdbGreenplumClusterV2Model{UserPassword: types.StringNull(), UserPasswordWoVersion: types.Int64Value(2)}

		_, _, diags := greenplumClusterPasswordChange(plan, state, types.StringNull())
		if !diags.HasError() {
			t.Fatal("greenplumClusterPasswordChange() diagnostics has no error")
		}
	})

	t.Run("legacy password change remains supported", func(t *testing.T) {
		state := &yandexMdbGreenplumClusterV2Model{UserPassword: types.StringValue("old-password"), UserPasswordWoVersion: types.Int64Null()}
		plan := &yandexMdbGreenplumClusterV2Model{UserPassword: types.StringValue("new-password"), UserPasswordWoVersion: types.Int64Null()}

		password, changed, diags := greenplumClusterPasswordChange(plan, state, types.StringNull())
		if diags.HasError() || !changed || password != "new-password" {
			t.Fatalf("password change = (%q, %t, %#v), want legacy password", password, changed, diags)
		}
	})
}

func TestGreenplumClusterPasswordRedaction(t *testing.T) {
	createReq := &greenplum.CreateClusterRequest{UserPassword: "create-secret"}
	redactedCreate := redactCreateClusterRequest(createReq)
	if createReq.GetUserPassword() != "create-secret" || redactedCreate.GetUserPassword() != redactedGreenplumPassword {
		t.Fatal("create request redaction mutated the source or did not redact the clone")
	}
	if dump := validate.ProtoDump(redactedCreate); strings.Contains(dump, "create-secret") {
		t.Fatalf("redacted create request contains the password: %s", dump)
	}

	updateReq := &greenplum.UpdateClusterRequest{UserPassword: "update-secret"}
	redactedUpdate := redactUpdateClusterRequest(updateReq)
	if updateReq.GetUserPassword() != "update-secret" || redactedUpdate.GetUserPassword() != redactedGreenplumPassword {
		t.Fatal("update request redaction mutated the source or did not redact the clone")
	}
	if dump := validate.ProtoDump(redactedUpdate); strings.Contains(dump, "update-secret") {
		t.Fatalf("redacted update request contains the password: %s", dump)
	}
}
