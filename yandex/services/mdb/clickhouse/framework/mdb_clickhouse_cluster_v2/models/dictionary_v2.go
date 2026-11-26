package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/framework/mdb_clickhouse_cluster_v2/utils"
)

type DictionaryPostgresqlSourceV2 struct {
	Db                  types.String `tfsdk:"db"`
	Table               types.String `tfsdk:"table"`
	Replicas            types.List   `tfsdk:"replicas"`
	Port                types.Int64  `tfsdk:"port"`
	User                types.String `tfsdk:"user"`
	Password            types.String `tfsdk:"password"`
	InvalidateQuery     types.String `tfsdk:"invalidate_query"`
	SslMode             types.String `tfsdk:"ssl_mode"`
	Where               types.String `tfsdk:"where"`
	Query               types.String `tfsdk:"query"`
	BackgroundReconnect types.Bool   `tfsdk:"background_reconnect"`
}

var DictionaryPostgresqlSourceV2AttrTypes = map[string]attr.Type{
	"db":                   types.StringType,
	"table":                types.StringType,
	"replicas":             types.ListType{ElemType: types.ObjectType{AttrTypes: DictionaryPostgresqlReplicaV2AttrTypes}},
	"port":                 types.Int64Type,
	"user":                 types.StringType,
	"password":             types.StringType,
	"invalidate_query":     types.StringType,
	"ssl_mode":             types.StringType,
	"where":                types.StringType,
	"query":                types.StringType,
	"background_reconnect": types.BoolType,
}

type DictionaryPostgresqlReplicaV2 struct {
	Host     types.String `tfsdk:"host"`
	Priority types.Int64  `tfsdk:"priority"`
	Port     types.Int64  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
}

var DictionaryPostgresqlReplicaV2AttrTypes = map[string]attr.Type{
	"host":     types.StringType,
	"priority": types.Int64Type,
	"port":     types.Int64Type,
	"user":     types.StringType,
	"password": types.StringType,
}

type DictionaryYtSourceV2 struct {
	HttpProxyUrls types.List   `tfsdk:"http_proxy_urls"`
	CypressPath   types.String `tfsdk:"cypress_path"`
	OauthToken    types.String `tfsdk:"oauth_token"`
}

var DictionaryYtSourceV2AttrTypes = map[string]attr.Type{
	"http_proxy_urls": types.ListType{ElemType: types.StringType},
	"cypress_path":    types.StringType,
	"oauth_token":     types.StringType,
}

func flattenDictionaryPostgresqlSourceV2(ctx context.Context, source *clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSourceV2, previous types.Object, diags *diag.Diagnostics) types.Object {
	if source == nil {
		return types.ObjectNull(DictionaryPostgresqlSourceV2AttrTypes)
	}
	var prev DictionaryPostgresqlSourceV2
	if !previous.IsNull() && !previous.IsUnknown() {
		diags.Append(previous.As(ctx, &prev, datasize.DefaultOpts)...)
		if diags.HasError() {
			return types.ObjectNull(DictionaryPostgresqlSourceV2AttrTypes)
		}
	}

	var prevReplicas []DictionaryPostgresqlReplicaV2
	if !prev.Replicas.IsNull() && !prev.Replicas.IsUnknown() {
		diags.Append(prev.Replicas.ElementsAs(ctx, &prevReplicas, false)...)
		if diags.HasError() {
			return types.ObjectNull(DictionaryPostgresqlSourceV2AttrTypes)
		}
	}
	previousByEndpoint := make(map[postgresqlReplicaEndpoint]DictionaryPostgresqlReplicaV2, len(prevReplicas))
	previousAPI := expandDictionaryPostgresqlSourceV2(ctx, previous, diags)
	if diags.HasError() {
		return types.ObjectNull(DictionaryPostgresqlSourceV2AttrTypes)
	}
	for i, key := range postgresqlReplicaV2Keys(previousAPI) {
		previousByEndpoint[key] = prevReplicas[i]
	}
	keys := postgresqlReplicaV2Keys(source)

	replicas := types.ListNull(types.ObjectType{AttrTypes: DictionaryPostgresqlReplicaV2AttrTypes})
	if len(source.Replicas) > 0 || (!prev.Replicas.IsNull() && !prev.Replicas.IsUnknown()) {
		values := make([]DictionaryPostgresqlReplicaV2, 0, len(source.Replicas))
		for i, replica := range source.Replicas {
			old := previousByEndpoint[keys[i]]
			values = append(values, DictionaryPostgresqlReplicaV2{
				Host:     types.StringValue(replica.GetHost()),
				Priority: flattenDictionaryV2OptionalInt64(replica.GetPriority(), old.Priority),
				Port:     flattenDictionaryV2OptionalInt64(replica.GetPort(), old.Port),
				User:     flattenDictionaryV2OptionalString(replica.GetUser(), old.User),
				Password: flattenDictionaryV2Secret(replica.GetPassword(), old.Password),
			})
		}
		var d diag.Diagnostics
		replicas, d = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DictionaryPostgresqlReplicaV2AttrTypes}, values)
		diags.Append(d...)
	}

	sslMode := types.StringValue(source.SslMode.String())
	if source.SslMode == clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSourceV2_SSL_MODE_UNSPECIFIED && (prev.SslMode.IsNull() || prev.SslMode.IsUnknown()) {
		sslMode = types.StringNull()
	}

	obj, d := types.ObjectValueFrom(ctx, DictionaryPostgresqlSourceV2AttrTypes, DictionaryPostgresqlSourceV2{
		Db:                  types.StringValue(source.Db),
		Table:               flattenDictionaryV2OptionalString(source.Table, prev.Table),
		Replicas:            replicas,
		Port:                types.Int64Value(source.Port),
		User:                flattenDictionaryV2OptionalString(source.User, prev.User),
		Password:            flattenDictionaryV2Secret(source.Password, prev.Password),
		InvalidateQuery:     flattenDictionaryV2OptionalString(source.InvalidateQuery, prev.InvalidateQuery),
		SslMode:             sslMode,
		Where:               flattenDictionaryV2OptionalString(source.Where, prev.Where),
		Query:               flattenDictionaryV2OptionalString(source.Query, prev.Query),
		BackgroundReconnect: mdbcommon.FlattenBoolWrapper(ctx, source.BackgroundReconnect, diags),
	})
	diags.Append(d...)
	return obj
}

// Identify replicas by their host and effective connection settings.
type postgresqlReplicaEndpoint struct {
	host string
	port int64
	user string
}

func postgresqlReplicaV2Keys(source *clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSourceV2) []postgresqlReplicaEndpoint {
	if source == nil {
		return nil
	}
	keys := make([]postgresqlReplicaEndpoint, 0, len(source.Replicas))
	for _, replica := range source.Replicas {
		key := postgresqlReplicaEndpoint{host: replica.GetHost(), port: source.Port, user: source.User}
		if replica.GetPort() != 0 {
			key.port = replica.GetPort()
		}
		if replica.GetUser() != "" {
			key.user = replica.GetUser()
		}
		keys = append(keys, key)
	}
	return keys
}

// RestoreDictionaryPostgresqlV2Passwords fills secrets omitted by the read API
// before comparing a dictionary with its desired configuration.
func RestoreDictionaryPostgresqlV2Passwords(current, previous *clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSourceV2) {
	if current == nil || previous == nil {
		return
	}
	if current.Password == "" {
		current.Password = previous.Password
	}
	passwords := make(map[postgresqlReplicaEndpoint]string, len(previous.Replicas))
	for i, key := range postgresqlReplicaV2Keys(previous) {
		passwords[key] = previous.Replicas[i].GetPassword()
	}
	for i, key := range postgresqlReplicaV2Keys(current) {
		if replica := current.Replicas[i]; replica != nil && replica.Password == "" {
			replica.Password = passwords[key]
		}
	}
}

func expandDictionaryPostgresqlSourceV2(ctx context.Context, value types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSourceV2 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	var source DictionaryPostgresqlSourceV2
	diags.Append(value.As(ctx, &source, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}
	sslMode := utils.ExpandEnum("ssl_mode", source.SslMode.ValueString(), clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSourceV2_SslMode_value, diags)
	if diags.HasError() {
		return nil
	}
	var replicas []*clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSourceV2_Replica
	if !source.Replicas.IsNull() && !source.Replicas.IsUnknown() {
		var values []DictionaryPostgresqlReplicaV2
		diags.Append(source.Replicas.ElementsAs(ctx, &values, false)...)
		if diags.HasError() {
			return nil
		}
		for _, replica := range values {
			replicas = append(replicas, &clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSourceV2_Replica{
				Host:     replica.Host.ValueString(),
				Priority: replica.Priority.ValueInt64(),
				Port:     replica.Port.ValueInt64(),
				User:     replica.User.ValueString(),
				Password: replica.Password.ValueString(),
			})
		}
	}
	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSourceV2{
		Db:                  source.Db.ValueString(),
		Table:               source.Table.ValueString(),
		Replicas:            replicas,
		Port:                source.Port.ValueInt64(),
		User:                source.User.ValueString(),
		Password:            source.Password.ValueString(),
		InvalidateQuery:     source.InvalidateQuery.ValueString(),
		SslMode:             clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSourceV2_SslMode(*sslMode),
		Where:               source.Where.ValueString(),
		Query:               source.Query.ValueString(),
		BackgroundReconnect: mdbcommon.ExpandBoolWrapper(ctx, source.BackgroundReconnect, diags),
	}
}

func flattenDictionaryYtSourceV2(ctx context.Context, source *clickhouseConfig.ClickhouseConfig_ExternalDictionary_YtSourceV2, previous types.Object, diags *diag.Diagnostics) types.Object {
	if source == nil {
		return types.ObjectNull(DictionaryYtSourceV2AttrTypes)
	}
	var prev DictionaryYtSourceV2
	if !previous.IsNull() && !previous.IsUnknown() {
		diags.Append(previous.As(ctx, &prev, datasize.DefaultOpts)...)
		if diags.HasError() {
			return types.ObjectNull(DictionaryYtSourceV2AttrTypes)
		}
	}
	urls, d := types.ListValueFrom(ctx, types.StringType, source.HttpProxyUrls)
	diags.Append(d...)
	obj, d := types.ObjectValueFrom(ctx, DictionaryYtSourceV2AttrTypes, DictionaryYtSourceV2{
		HttpProxyUrls: urls,
		CypressPath:   types.StringValue(source.CypressPath),
		OauthToken:    flattenDictionaryV2Secret(source.OauthToken, prev.OauthToken),
	})
	diags.Append(d...)
	return obj
}

func expandDictionaryYtSourceV2(ctx context.Context, value types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_YtSourceV2 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	var source DictionaryYtSourceV2
	diags.Append(value.As(ctx, &source, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}
	var urls []string
	diags.Append(source.HttpProxyUrls.ElementsAs(ctx, &urls, false)...)
	if diags.HasError() {
		return nil
	}
	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_YtSourceV2{
		HttpProxyUrls: urls,
		CypressPath:   source.CypressPath.ValueString(),
		OauthToken:    source.OauthToken.ValueString(),
	}
}

// Protobuf scalar defaults do not distinguish omitted values from explicit zero values.
// Preserve the previous Terraform representation when the API returns such a default.
func flattenDictionaryV2OptionalString(value string, previous types.String) types.String {
	if value == "" && (previous.IsNull() || previous.IsUnknown()) {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func flattenDictionaryV2OptionalInt64(value int64, previous types.Int64) types.Int64 {
	if value == 0 && (previous.IsNull() || previous.IsUnknown()) {
		return types.Int64Null()
	}
	return types.Int64Value(value)
}

func flattenDictionaryV2Secret(value string, previous types.String) types.String {
	if value == "" && !previous.IsNull() && !previous.IsUnknown() {
		return previous
	}
	return flattenDictionaryV2OptionalString(value, previous)
}
