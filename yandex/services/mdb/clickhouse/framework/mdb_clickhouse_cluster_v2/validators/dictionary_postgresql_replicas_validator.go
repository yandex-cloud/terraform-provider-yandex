package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/framework/mdb_clickhouse_cluster_v2/models"
)

type DictionaryPostgresqlReplicasValidator struct{}

func (DictionaryPostgresqlReplicasValidator) Description(context.Context) string {
	return "PostgreSQL replicas must have unique host, effective port and effective user combinations."
}

func (v DictionaryPostgresqlReplicasValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v DictionaryPostgresqlReplicasValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var source models.DictionaryPostgresqlSourceV2
	resp.Diagnostics.Append(req.ConfigValue.As(ctx, &source, datasize.DefaultOpts)...)
	if resp.Diagnostics.HasError() || source.Replicas.IsNull() || source.Replicas.IsUnknown() {
		return
	}

	type endpoint struct {
		host string
		port int64
		user string
	}
	seen := make(map[endpoint]int)
	for i, value := range source.Replicas.Elements() {
		if value.IsNull() || value.IsUnknown() {
			continue
		}
		var replica models.DictionaryPostgresqlReplicaV2
		resp.Diagnostics.Append(value.(types.Object).As(ctx, &replica, datasize.DefaultOpts)...)
		if resp.Diagnostics.HasError() {
			return
		}

		port, user := replica.Port, replica.User
		if !port.IsUnknown() && port.ValueInt64() == 0 {
			port = source.Port
			if port.IsNull() {
				port = types.Int64Value(5432)
			}
		}
		if !user.IsUnknown() && user.ValueString() == "" {
			user = source.User
		}
		if replica.Host.IsUnknown() || port.IsUnknown() || user.IsUnknown() {
			continue
		}

		key := endpoint{host: replica.Host.ValueString(), port: port.ValueInt64(), user: user.ValueString()}
		if previous, exists := seen[key]; exists {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName("replicas").AtListIndex(i),
				"Duplicate PostgreSQL replica",
				fmt.Sprintf("Replicas %d and %d have the same host, effective port and effective user. %s", previous, i, v.Description(ctx)),
			)
			return
		}
		seen[key] = i
	}
}
