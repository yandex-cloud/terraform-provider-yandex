package mdb_clickhouse_user

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/chcommon"
	"google.golang.org/grpc/status"
)

const defaultUserAuthMethod = "password"

var (
	UserAuthMethod_name = map[int32]string{
		0: "unspecified",
		1: "password",
		2: "iam",
	}
	UserAuthMethod_value     = chcommon.MakeReversedMap(UserAuthMethod_name, clickhouse.AuthMethod_value)
	UserAuthMethod_validator = chcommon.MakeEnumNamesValidator(UserAuthMethod_name)

	DeletionProtectionMode_validator = []validator.String{stringvalidator.OneOf(
		clickhouse.DeletionProtectionMode_DELETION_PROTECTION_MODE_INHERITED.String(),
		clickhouse.DeletionProtectionMode_DELETION_PROTECTION_MODE_ENABLED.String(),
		clickhouse.DeletionProtectionMode_DELETION_PROTECTION_MODE_DISABLED.String(),
	)}
)

const defaultDeletionProtectionMode = clickhouse.DeletionProtectionMode_DELETION_PROTECTION_MODE_INHERITED

func getDeletionProtectionModeValue(mode types.String) clickhouse.DeletionProtectionMode {
	if mode.IsNull() || mode.IsUnknown() {
		return defaultDeletionProtectionMode
	}
	if value, ok := clickhouse.DeletionProtectionMode_value[mode.ValueString()]; ok {
		return clickhouse.DeletionProtectionMode(value)
	}
	return defaultDeletionProtectionMode
}

func getAuthMethodName(value clickhouse.AuthMethod) types.String {
	if name, ok := UserAuthMethod_name[int32(normalizeAuthMethod(value))]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getAuthMethodValue(name types.String) clickhouse.AuthMethod {
	if name.IsNull() || name.IsUnknown() || name.ValueString() == "" {
		return clickhouse.AuthMethod_AUTH_METHOD_PASSWORD
	}

	if value, ok := UserAuthMethod_value[name.ValueString()]; ok {
		return normalizeAuthMethod(clickhouse.AuthMethod(value))
	}

	return clickhouse.AuthMethod_AUTH_METHOD_UNSPECIFIED
}

func normalizeAuthMethod(value clickhouse.AuthMethod) clickhouse.AuthMethod {
	if value == clickhouse.AuthMethod_AUTH_METHOD_UNSPECIFIED {
		return clickhouse.AuthMethod_AUTH_METHOD_PASSWORD
	}
	return value
}

func errorMessage(err error) string {
	grpcStatus, _ := status.FromError(err)
	return grpcStatus.Message()
}
