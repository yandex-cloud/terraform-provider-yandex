package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type Rabbitmq struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Vhost    types.String `tfsdk:"vhost"`
}

var RabbitmqAttrTypes = map[string]attr.Type{
	"username": types.StringType,
	"password": types.StringType,
	"vhost":    types.StringType,
}

func flattenRabbitmq(ctx context.Context, state *Cluster, rabbitmq *clickhouseConfig.ClickhouseConfig_Rabbitmq, diags *diag.Diagnostics) types.Object {
	if rabbitmq == nil {
		return types.ObjectNull(RabbitmqAttrTypes)
	}

	var stateRabbitmq Rabbitmq
	if state != nil && !state.ClickHouse.IsNull() && !state.ClickHouse.IsUnknown() {
		var stateClickHouse Clickhouse
		diags.Append(state.ClickHouse.As(ctx, &stateClickHouse, datasize.UnhandledOpts)...)

		var stateClickHouseConfig ClickhouseConfig
		if !stateClickHouse.Config.IsNull() && !stateClickHouse.Config.IsUnknown() {
			diags.Append(stateClickHouse.Config.As(ctx, &stateClickHouseConfig, datasize.UnhandledOpts)...)
		}

		if !stateClickHouseConfig.Rabbitmq.IsNull() && !stateClickHouseConfig.Rabbitmq.IsUnknown() {
			diags.Append(stateClickHouseConfig.Rabbitmq.As(ctx, &stateRabbitmq, datasize.UnhandledOpts)...)
		}
	}

	obj, d := types.ObjectValueFrom(
		ctx, RabbitmqAttrTypes, Rabbitmq{
			Username: types.StringValue(rabbitmq.Username),
			Password: stateRabbitmq.Password,
			Vhost:    types.StringValue(rabbitmq.Vhost),
		},
	)
	diags.Append(d...)

	return obj
}

func expandRabbitmq(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_Rabbitmq {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var rabbitmq Rabbitmq
	diags.Append(c.As(ctx, &rabbitmq, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_Rabbitmq{
		Username: rabbitmq.Username.ValueString(),
		Password: rabbitmq.Password.ValueString(),
		Vhost:    rabbitmq.Vhost.ValueString(),
	}
}
