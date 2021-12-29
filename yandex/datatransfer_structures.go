package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/datatransfer/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datatransfer/v1/endpoint"
)

func parseEndpointObjectTransferStage(stage string) (endpoint.ObjectTransferStage, error) {
	val, ok := endpoint.ObjectTransferStage_value[stage]
	if !ok {
		return endpoint.ObjectTransferStage(0), fmt.Errorf("value for 'object_transfer_stage' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(endpoint.ObjectTransferStage_value)), stage)
	}
	return endpoint.ObjectTransferStage(val), nil
}

func expandEndpointSettings(d *schema.ResourceData) (*datatransfer.EndpointSettings, error) {
	val := new(datatransfer.EndpointSettings)

	if _, ok := d.GetOk("settings.0.mysql_source"); ok {
		mysqlSource, err := expandEndpointSettingsMysqlSource(d)
		if err != nil {
			return nil, err
		}

		val.SetMysqlSource(mysqlSource)
	}

	if _, ok := d.GetOk("settings.0.postgres_source"); ok {
		postgresSource, err := expandEndpointSettingsPostgresSource(d)
		if err != nil {
			return nil, err
		}

		val.SetPostgresSource(postgresSource)
	}

	if _, ok := d.GetOk("settings.0.mysql_target"); ok {
		mysqlTarget, err := expandEndpointSettingsMysqlTarget(d)
		if err != nil {
			return nil, err
		}

		val.SetMysqlTarget(mysqlTarget)
	}

	if _, ok := d.GetOk("settings.0.postgres_target"); ok {
		postgresTarget, err := expandEndpointSettingsPostgresTarget(d)
		if err != nil {
			return nil, err
		}

		val.SetPostgresTarget(postgresTarget)
	}

	empty := new(datatransfer.EndpointSettings)
	if proto.Equal(val, empty) {
		return nil, nil
	}

	return val, nil
}

func expandEndpointSettingsMysqlSource(d *schema.ResourceData) (*endpoint.MysqlSource, error) {
	val := new(endpoint.MysqlSource)

	if _, ok := d.GetOk("settings.0.mysql_source.0.connection"); ok {
		connection, err := expandEndpointSettingsMysqlSourceConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.user"); ok {
		val.SetUser(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mysql_source.0.password"); ok {
		password, err := expandEndpointSettingsMysqlSourcePassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.include_tables_regex"); ok {
		includeTablesRegex := expandStringSlice(v.([]interface{}))
		val.SetIncludeTablesRegex(includeTablesRegex)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.exclude_tables_regex"); ok {
		excludeTablesRegex := expandStringSlice(v.([]interface{}))
		val.SetExcludeTablesRegex(excludeTablesRegex)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.timezone"); ok {
		val.SetTimezone(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mysql_source.0.object_transfer_settings"); ok {
		objectTransferSettings, err := expandEndpointSettingsMysqlSourceObjectTransferSettings(d)
		if err != nil {
			return nil, err
		}

		val.SetObjectTransferSettings(objectTransferSettings)
	}

	return val, nil
}

func expandEndpointSettingsMysqlSourceConnection(d *schema.ResourceData) (*endpoint.MysqlConnection, error) {
	val := new(endpoint.MysqlConnection)

	if v, ok := d.GetOk("settings.0.mysql_source.0.connection.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise"); ok {
		onPremise, err := expandEndpointSettingsMysqlSourceConnectionOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	return val, nil
}

func expandEndpointSettingsMysqlSourceConnectionOnPremise(d *schema.ResourceData) (*endpoint.OnPremiseMysql, error) {
	val := new(endpoint.OnPremiseMysql)

	if v, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.hosts"); ok {
		hosts := expandStringSlice(v.([]interface{}))
		val.SetHosts(hosts)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.port"); ok {
		val.SetPort(int64(v.(int)))
	}

	if _, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandEndpointSettingsMysqlSourceConnectionOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsMysqlSourceConnectionOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandEndpointSettingsMysqlTarget(d *schema.ResourceData) (*endpoint.MysqlTarget, error) {
	val := new(endpoint.MysqlTarget)

	if _, ok := d.GetOk("settings.0.mysql_target.0.connection"); ok {
		connection, err := expandEndpointSettingsMysqlTargetConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.user"); ok {
		val.SetUser(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mysql_target.0.password"); ok {
		password, err := expandEndpointSettingsMysqlTargetPassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.sql_mode"); ok {
		val.SetSqlMode(v.(string))
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.skip_constraint_checks"); ok {
		val.SetSkipConstraintChecks(v.(bool))
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.timezone"); ok {
		val.SetTimezone(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsMysqlTargetConnection(d *schema.ResourceData) (*endpoint.MysqlConnection, error) {
	val := new(endpoint.MysqlConnection)

	if v, ok := d.GetOk("settings.0.mysql_target.0.connection.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise"); ok {
		onPremise, err := expandEndpointSettingsMysqlTargetConnectionOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	return val, nil
}

func expandEndpointSettingsMysqlTargetConnectionOnPremise(d *schema.ResourceData) (*endpoint.OnPremiseMysql, error) {
	val := new(endpoint.OnPremiseMysql)

	if v, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.hosts"); ok {
		hosts := expandStringSlice(v.([]interface{}))
		val.SetHosts(hosts)
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.port"); ok {
		val.SetPort(int64(v.(int)))
	}

	if _, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandEndpointSettingsMysqlTargetConnectionOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsMysqlTargetConnectionOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsMysqlTargetPassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.mysql_target.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsMysqlSourcePassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.mysql_source.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsMysqlSourceObjectTransferSettings(d *schema.ResourceData) (*endpoint.MysqlObjectTransferSettings, error) {
	val := new(endpoint.MysqlObjectTransferSettings)

	if v, ok := d.GetOk("settings.0.mysql_source.0.object_transfer_settings.0.view"); ok {
		objectTransferStage, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetView(objectTransferStage)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.object_transfer_settings.0.routine"); ok {
		objectTransferStage_, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetRoutine(objectTransferStage_)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.object_transfer_settings.0.trigger"); ok {
		objectTransferStage__, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetTrigger(objectTransferStage__)
	}

	return val, nil
}

func expandEndpointSettingsPostgresSource(d *schema.ResourceData) (*endpoint.PostgresSource, error) {
	val := new(endpoint.PostgresSource)

	if _, ok := d.GetOk("settings.0.postgres_source.0.connection"); ok {
		connection, err := expandEndpointSettingsPostgresSourceConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.user"); ok {
		val.SetUser(v.(string))
	}

	if _, ok := d.GetOk("settings.0.postgres_source.0.password"); ok {
		password, err := expandEndpointSettingsPostgresSourcePassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.include_tables"); ok {
		includeTables := expandStringSlice(v.([]interface{}))
		val.SetIncludeTables(includeTables)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.exclude_tables"); ok {
		excludeTables := expandStringSlice(v.([]interface{}))
		val.SetExcludeTables(excludeTables)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.slot_gigabyte_lag_limit"); ok {
		val.SetSlotByteLagLimit(toBytes(v.(int)))
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.service_schema"); ok {
		val.SetServiceSchema(v.(string))
	}

	if _, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings"); ok {
		objectTransferSettings, err := expandEndpointSettingsPostgresSourceObjectTransferSettings(d)
		if err != nil {
			return nil, err
		}

		val.SetObjectTransferSettings(objectTransferSettings)
	}

	return val, nil
}

func expandEndpointSettingsPostgresTarget(d *schema.ResourceData) (*endpoint.PostgresTarget, error) {
	val := new(endpoint.PostgresTarget)

	if _, ok := d.GetOk("settings.0.postgres_target.0.connection"); ok {
		connection, err := expandEndpointSettingsPostgresTargetConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.postgres_target.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.postgres_target.0.user"); ok {
		val.SetUser(v.(string))
	}

	if _, ok := d.GetOk("settings.0.postgres_target.0.password"); ok {
		password, err := expandEndpointSettingsPostgresTargetPassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	return val, nil
}

func expandEndpointSettingsPostgresTargetConnection(d *schema.ResourceData) (*endpoint.PostgresConnection, error) {
	val := new(endpoint.PostgresConnection)

	if v, ok := d.GetOk("settings.0.postgres_target.0.connection.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise"); ok {
		onPremise, err := expandEndpointSettingsPostgresTargetConnectionOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	return val, nil
}

func expandEndpointSettingsPostgresTargetConnectionOnPremise(d *schema.ResourceData) (*endpoint.OnPremisePostgres, error) {
	val := new(endpoint.OnPremisePostgres)

	if v, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.hosts"); ok {
		hosts := expandStringSlice(v.([]interface{}))
		val.SetHosts(hosts)
	}

	if v, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.port"); ok {
		val.SetPort(int64(v.(int)))
	}

	if _, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandEndpointSettingsPostgresTargetConnectionOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	if v, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsPostgresTargetConnectionOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsPostgresTargetPassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.postgres_target.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsPostgresSourceConnection(d *schema.ResourceData) (*endpoint.PostgresConnection, error) {
	val := new(endpoint.PostgresConnection)

	if v, ok := d.GetOk("settings.0.postgres_source.0.connection.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise"); ok {
		onPremise, err := expandEndpointSettingsPostgresSourceConnectionOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	return val, nil
}

func expandEndpointSettingsPostgresSourceConnectionOnPremise(d *schema.ResourceData) (*endpoint.OnPremisePostgres, error) {
	val := new(endpoint.OnPremisePostgres)

	if v, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.hosts"); ok {
		hosts := expandStringSlice(v.([]interface{}))
		val.SetHosts(hosts)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.port"); ok {
		val.SetPort(int64(v.(int)))
	}

	if _, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandEndpointSettingsPostgresSourceConnectionOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsPostgresSourceConnectionOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsPostgresSourcePassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.postgres_source.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandEndpointSettingsPostgresSourceObjectTransferSettings(d *schema.ResourceData) (*endpoint.PostgresObjectTransferSettings, error) {
	val := new(endpoint.PostgresObjectTransferSettings)

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.sequence"); ok {
		objectTransferStage, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetSequence(objectTransferStage)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.sequence_owned_by"); ok {
		objectTransferStage_, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetSequenceOwnedBy(objectTransferStage_)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.table"); ok {
		objectTransferStage__, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetTable(objectTransferStage__)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.primary_key"); ok {
		objectTransferStage___, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetPrimaryKey(objectTransferStage___)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.fk_constraint"); ok {
		objectTransferStage____, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetFkConstraint(objectTransferStage____)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.default_values"); ok {
		objectTransferStage_____, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetDefaultValues(objectTransferStage_____)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.constraint"); ok {
		objectTransferStage______, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetConstraint(objectTransferStage______)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.index"); ok {
		objectTransferStage_______, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetIndex(objectTransferStage_______)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.view"); ok {
		objectTransferStage________, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetView(objectTransferStage________)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.function"); ok {
		objectTransferStage_________, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetFunction(objectTransferStage_________)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.trigger"); ok {
		objectTransferStage__________, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetTrigger(objectTransferStage__________)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.type"); ok {
		objectTransferStage___________, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(objectTransferStage___________)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.rule"); ok {
		objectTransferStage____________, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetRule(objectTransferStage____________)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.collation"); ok {
		objectTransferStage_____________, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetCollation(objectTransferStage_____________)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.policy"); ok {
		objectTransferStage______________, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetPolicy(objectTransferStage______________)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.cast"); ok {
		objectTransferStage_______________, err := parseEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetCast(objectTransferStage_______________)
	}

	return val, nil
}

func flattenDatatransferSettings(d *schema.ResourceData, v *datatransfer.EndpointSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	postgresSource, err := flattenDatatransferSettingsPostgresSource(d, v.GetPostgresSource())
	if err != nil {
		return nil, err
	}
	m["postgres_source"] = postgresSource

	postgresTarget, err := flattenDatatransferEndpointPostgresTarget(d, v.GetPostgresTarget())
	if err != nil {
		return nil, err
	}
	m["postgres_target"] = postgresTarget

	mysqlSource, err := flattenDatatransferSettingsMysqlSource(d, v.GetMysqlSource())
	if err != nil {
		return nil, err
	}
	m["mysql_source"] = mysqlSource

	mysqlTarget, err := flattenDatatransferEndpointMysqlTarget(d, v.GetMysqlTarget())
	if err != nil {
		return nil, err
	}
	m["mysql_target"] = mysqlTarget

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferSettingsPostgresSource(d *schema.ResourceData, v *endpoint.PostgresSource) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connection, err := flattenDatatransferEndpointPostgresConnection(v.Connection)
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["database"] = v.Database
	m["exclude_tables"] = v.ExcludeTables
	m["include_tables"] = v.IncludeTables
	objectTransferSettings, err := flattenDatatransferPostgresObjectTransferSettings(v.ObjectTransferSettings)
	if err != nil {
		return nil, err
	}
	m["object_transfer_settings"] = objectTransferSettings
	if password, ok := d.GetOk("settings.0.postgres_source.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["service_schema"] = v.ServiceSchema
	m["slot_gigabyte_lag_limit"] = toGigabytes(v.SlotByteLagLimit)
	m["user"] = v.User

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointPostgresTarget(d *schema.ResourceData, v *endpoint.PostgresTarget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connection, err := flattenDatatransferEndpointPostgresConnection(v.Connection)
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["database"] = v.Database
	if password, ok := d.GetOk("settings.0.postgres_target.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["user"] = v.User

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferSettingsMysqlSource(d *schema.ResourceData, v *endpoint.MysqlSource) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connection, err := flattenDatatransferEndpointMysqlConnection(v.Connection)
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["database"] = v.Database
	m["exclude_tables_regex"] = v.ExcludeTablesRegex
	m["include_tables_regex"] = v.IncludeTablesRegex
	objectTransferSettings, err := flattenDatatransferEndpointMysqlObjectTransferSettings(v.ObjectTransferSettings)
	if err != nil {
		return nil, err
	}
	m["object_transfer_settings"] = objectTransferSettings
	if password, ok := d.GetOk("settings.0.mysql_source.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["timezone"] = v.Timezone
	m["user"] = v.User

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointMysqlTarget(d *schema.ResourceData, v *endpoint.MysqlTarget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connection, err := flattenDatatransferEndpointMysqlConnection(v.Connection)
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["database"] = v.Database
	if password, ok := d.GetOk("settings.0.mysql_target.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["skip_constraint_checks"] = v.SkipConstraintChecks
	m["sql_mode"] = v.SqlMode
	m["timezone"] = v.Timezone
	m["user"] = v.User

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointPostgresConnection(v *endpoint.PostgresConnection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["mdb_cluster_id"] = v.GetMdbClusterId()
	onPremise, err := flattenEndpointSettingspostgresSourceconnectiononPremise(v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferPostgresObjectTransferSettings(v *endpoint.PostgresObjectTransferSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["cast"] = v.Cast.String()
	m["collation"] = v.Collation.String()
	m["constraint"] = v.Constraint.String()
	m["default_values"] = v.DefaultValues.String()
	m["fk_constraint"] = v.FkConstraint.String()
	m["function"] = v.Function.String()
	m["index"] = v.Index.String()
	m["policy"] = v.Policy.String()
	m["primary_key"] = v.PrimaryKey.String()
	m["rule"] = v.Rule.String()
	m["sequence"] = v.Sequence.String()
	m["sequence_owned_by"] = v.SequenceOwnedBy.String()
	m["table"] = v.Table.String()
	m["trigger"] = v.Trigger.String()
	m["type"] = v.Type.String()
	m["view"] = v.View.String()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointMysqlConnection(v *endpoint.MysqlConnection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["mdb_cluster_id"] = v.GetMdbClusterId()
	onPremise, err := flattenEndpointSettingsmysqlSourceconnectiononPremise(v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointMysqlObjectTransferSettings(v *endpoint.MysqlObjectTransferSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["routine"] = v.Routine.String()
	m["trigger"] = v.Trigger.String()
	m["view"] = v.View.String()

	return []map[string]interface{}{m}, nil
}

func flattenEndpointSettingspostgresSourceconnectiononPremise(v *endpoint.OnPremisePostgres) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["hosts"] = v.Hosts
	m["port"] = v.Port
	m["subnet_id"] = v.SubnetId
	tlsMode, err := flattenrDatatransferEndpointTLSMode(v.TlsMode)
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenEndpointSettingsmysqlSourceconnectiononPremise(v *endpoint.OnPremiseMysql) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["hosts"] = v.Hosts
	m["port"] = v.Port
	m["subnet_id"] = v.SubnetId
	tlsMode, err := flattenrDatatransferEndpointTLSMode(v.TlsMode)
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenrDatatransferEndpointTLSMode(v *endpoint.TLSMode) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	disabled, err := flattenGoogleProtobufEmpty(v.GetDisabled())
	if err != nil {
		return nil, err
	}
	m["disabled"] = disabled
	enabled, err := flattenDatatransferEndpointTLSConfig(v.GetEnabled())
	if err != nil {
		return nil, err
	}
	m["enabled"] = enabled

	return []map[string]interface{}{m}, nil
}

func flattenGoogleProtobufEmpty(v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointTLSConfig(v *endpoint.TLSConfig) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ca_certificate"] = v.CaCertificate

	return []map[string]interface{}{m}, nil
}

func parseDatatransferTransferType(str string) (datatransfer.TransferType, error) {
	val, ok := datatransfer.TransferType_value[str]
	if !ok {
		return datatransfer.TransferType(0), fmt.Errorf("value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(datatransfer.TransferType_value)), str)
	}
	return datatransfer.TransferType(val), nil
}
