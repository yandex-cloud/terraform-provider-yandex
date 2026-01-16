package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/utils"
)

type Kafka struct {
	SecurityProtocol                 types.String `tfsdk:"security_protocol"`
	SaslMechanism                    types.String `tfsdk:"sasl_mechanism"`
	SaslUsername                     types.String `tfsdk:"sasl_username"`
	SaslPassword                     types.String `tfsdk:"sasl_password"`
	EnableSslCertificateVerification types.Bool   `tfsdk:"enable_ssl_certificate_verification"`
	MaxPollIntervalMs                types.Int64  `tfsdk:"max_poll_interval_ms"`
	SessionTimeoutMs                 types.Int64  `tfsdk:"session_timeout_ms"`
	Debug                            types.String `tfsdk:"debug"`
	AutoOffsetReset                  types.String `tfsdk:"auto_offset_reset"`
}

var KafkaAttrTypes = map[string]attr.Type{
	"security_protocol":                   types.StringType,
	"sasl_mechanism":                      types.StringType,
	"sasl_username":                       types.StringType,
	"sasl_password":                       types.StringType,
	"enable_ssl_certificate_verification": types.BoolType,
	"max_poll_interval_ms":                types.Int64Type,
	"session_timeout_ms":                  types.Int64Type,
	"debug":                               types.StringType,
	"auto_offset_reset":                   types.StringType,
}

func flattenKafka(ctx context.Context, state *Cluster, kafka *clickhouseConfig.ClickhouseConfig_Kafka, diags *diag.Diagnostics) types.Object {
	if kafka == nil {
		return types.ObjectNull(KafkaAttrTypes)
	}

	var stateKafka Kafka
	if state != nil && !state.ClickHouse.IsNull() && !state.ClickHouse.IsUnknown() {
		var stateClickHouse Clickhouse
		diags.Append(state.ClickHouse.As(ctx, &stateClickHouse, datasize.UnhandledOpts)...)

		var stateClickHouseConfig ClickhouseConfig
		if !stateClickHouse.Config.IsNull() && !stateClickHouse.Config.IsUnknown() {
			diags.Append(stateClickHouse.Config.As(ctx, &stateClickHouseConfig, datasize.UnhandledOpts)...)
		}

		if !stateClickHouseConfig.Kafka.IsNull() && !stateClickHouseConfig.Kafka.IsUnknown() {
			diags.Append(stateClickHouseConfig.Kafka.As(ctx, &stateKafka, datasize.UnhandledOpts)...)
		}
	}

	obj, d := types.ObjectValueFrom(
		ctx, KafkaAttrTypes, Kafka{
			SecurityProtocol:                 types.StringValue(kafka.SecurityProtocol.Enum().String()),
			SaslMechanism:                    types.StringValue(kafka.SaslMechanism.Enum().String()),
			SaslUsername:                     types.StringValue(kafka.SaslUsername),
			SaslPassword:                     stateKafka.SaslPassword,
			EnableSslCertificateVerification: mdbcommon.FlattenBoolWrapper(ctx, kafka.EnableSslCertificateVerification, diags),
			MaxPollIntervalMs:                mdbcommon.FlattenInt64Wrapper(ctx, kafka.MaxPollIntervalMs, diags),
			SessionTimeoutMs:                 mdbcommon.FlattenInt64Wrapper(ctx, kafka.SessionTimeoutMs, diags),
			Debug:                            types.StringValue(kafka.Debug.Enum().String()),
			AutoOffsetReset:                  types.StringValue(kafka.AutoOffsetReset.Enum().String()),
		},
	)
	diags.Append(d...)

	return obj
}

func expandKafka(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_Kafka {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var kafka Kafka
	diags.Append(c.As(ctx, &kafka, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	securityProtocolValue := utils.ExpandEnum("security_protocol", kafka.SecurityProtocol.ValueString(), clickhouseConfig.ClickhouseConfig_Kafka_SecurityProtocol_value, diags)
	if diags.HasError() {
		return nil
	}

	saslMechanismValue := utils.ExpandEnum("sasl_mechanism", kafka.SaslMechanism.ValueString(), clickhouseConfig.ClickhouseConfig_Kafka_SaslMechanism_value, diags)
	if diags.HasError() {
		return nil
	}

	debugValue := utils.ExpandEnum("debug", kafka.Debug.ValueString(), clickhouseConfig.ClickhouseConfig_Kafka_Debug_value, diags)
	if diags.HasError() {
		return nil
	}

	autoOffsetResetValue := utils.ExpandEnum("auto_offset_reset", kafka.AutoOffsetReset.ValueString(), clickhouseConfig.ClickhouseConfig_Kafka_AutoOffsetReset_value, diags)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_Kafka{
		SecurityProtocol:                 clickhouseConfig.ClickhouseConfig_Kafka_SecurityProtocol(*securityProtocolValue),
		SaslMechanism:                    clickhouseConfig.ClickhouseConfig_Kafka_SaslMechanism(*saslMechanismValue),
		SaslUsername:                     kafka.SaslUsername.ValueString(),
		SaslPassword:                     kafka.SaslPassword.ValueString(),
		EnableSslCertificateVerification: mdbcommon.ExpandBoolWrapper(ctx, kafka.EnableSslCertificateVerification, diags),
		MaxPollIntervalMs:                mdbcommon.ExpandInt64Wrapper(ctx, kafka.MaxPollIntervalMs, diags),
		SessionTimeoutMs:                 mdbcommon.ExpandInt64Wrapper(ctx, kafka.SessionTimeoutMs, diags),
		Debug:                            clickhouseConfig.ClickhouseConfig_Kafka_Debug(*debugValue),
		AutoOffsetReset:                  clickhouseConfig.ClickhouseConfig_Kafka_AutoOffsetReset(*autoOffsetResetValue),
	}
}
