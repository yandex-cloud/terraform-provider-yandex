package mdb_clickhouse_user

import (
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
)

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
