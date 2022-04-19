// Code generated with gentf. DO NOT EDIT.
package yandex

import (
	fmt "fmt"

	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	datatransfer "github.com/yandex-cloud/go-genproto/yandex/cloud/datatransfer/v1"
	endpoint "github.com/yandex-cloud/go-genproto/yandex/cloud/datatransfer/v1/endpoint"
	proto "google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func expandDatatransferEndpointSettings(d *schema.ResourceData) (*datatransfer.EndpointSettings, error) {
	val := new(datatransfer.EndpointSettings)

	if _, ok := d.GetOk("settings.0.mongo_source"); ok {
		mongoSource, err := expandDatatransferEndpointSettingsMongoSource(d)
		if err != nil {
			return nil, err
		}

		val.SetMongoSource(mongoSource)
	}

	if _, ok := d.GetOk("settings.0.mongo_target"); ok {
		mongoTarget, err := expandDatatransferEndpointSettingsMongoTarget(d)
		if err != nil {
			return nil, err
		}

		val.SetMongoTarget(mongoTarget)
	}

	if _, ok := d.GetOk("settings.0.mysql_source"); ok {
		mysqlSource, err := expandDatatransferEndpointSettingsMysqlSource(d)
		if err != nil {
			return nil, err
		}

		val.SetMysqlSource(mysqlSource)
	}

	if _, ok := d.GetOk("settings.0.mysql_target"); ok {
		mysqlTarget, err := expandDatatransferEndpointSettingsMysqlTarget(d)
		if err != nil {
			return nil, err
		}

		val.SetMysqlTarget(mysqlTarget)
	}

	if _, ok := d.GetOk("settings.0.postgres_source"); ok {
		postgresSource, err := expandDatatransferEndpointSettingsPostgresSource(d)
		if err != nil {
			return nil, err
		}

		val.SetPostgresSource(postgresSource)
	}

	if _, ok := d.GetOk("settings.0.postgres_target"); ok {
		postgresTarget, err := expandDatatransferEndpointSettingsPostgresTarget(d)
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

func expandDatatransferEndpointSettingsPostgresTarget(d *schema.ResourceData) (*endpoint.PostgresTarget, error) {
	val := new(endpoint.PostgresTarget)

	if _, ok := d.GetOk("settings.0.postgres_target.0.connection"); ok {
		connection, err := expandDatatransferEndpointSettingsPostgresTargetConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.postgres_target.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if _, ok := d.GetOk("settings.0.postgres_target.0.password"); ok {
		password, err := expandDatatransferEndpointSettingsPostgresTargetPassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.postgres_target.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.postgres_target.0.user"); ok {
		val.SetUser(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresTargetPassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.postgres_target.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresTargetConnection(d *schema.ResourceData) (*endpoint.PostgresConnection, error) {
	val := new(endpoint.PostgresConnection)

	if v, ok := d.GetOk("settings.0.postgres_target.0.connection.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise"); ok {
		onPremise, err := expandDatatransferEndpointSettingsPostgresTargetConnectionOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresTargetConnectionOnPremise(d *schema.ResourceData) (*endpoint.OnPremisePostgres, error) {
	val := new(endpoint.OnPremisePostgres)

	if v, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.hosts"); ok {
		val.SetHosts(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.port"); ok {
		val.SetPort(int64(v.(int)))
	}

	if v, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.postgres_target.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresSource(d *schema.ResourceData) (*endpoint.PostgresSource, error) {
	val := new(endpoint.PostgresSource)

	if _, ok := d.GetOk("settings.0.postgres_source.0.connection"); ok {
		connection, err := expandDatatransferEndpointSettingsPostgresSourceConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.exclude_tables"); ok {
		val.SetExcludeTables(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.include_tables"); ok {
		val.SetIncludeTables(expandStringSlice(v.([]interface{})))
	}

	if _, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings"); ok {
		objectTransferSettings, err := expandDatatransferEndpointSettingsPostgresSourceObjectTransferSettings(d)
		if err != nil {
			return nil, err
		}

		val.SetObjectTransferSettings(objectTransferSettings)
	}

	if _, ok := d.GetOk("settings.0.postgres_source.0.password"); ok {
		password, err := expandDatatransferEndpointSettingsPostgresSourcePassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.service_schema"); ok {
		val.SetServiceSchema(v.(string))
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.slot_gigabyte_lag_limit"); ok {
		val.SetSlotByteLagLimit(toBytes(v.(int)))
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.user"); ok {
		val.SetUser(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresSourcePassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.postgres_source.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresSourceObjectTransferSettings(d *schema.ResourceData) (*endpoint.PostgresObjectTransferSettings, error) {
	val := new(endpoint.PostgresObjectTransferSettings)

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.cast"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetCast(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.collation"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetCollation(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.constraint"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetConstraint(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.default_values"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetDefaultValues(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.fk_constraint"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetFkConstraint(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.function"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetFunction(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.index"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetIndex(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.materialized_view"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetMaterializedView(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.policy"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetPolicy(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.primary_key"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetPrimaryKey(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.rule"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetRule(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.sequence"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetSequence(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.sequence_owned_by"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetSequenceOwnedBy(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.table"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetTable(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.trigger"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetTrigger(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.type"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(vv)
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.view"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetView(vv)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresSourceConnection(d *schema.ResourceData) (*endpoint.PostgresConnection, error) {
	val := new(endpoint.PostgresConnection)

	if v, ok := d.GetOk("settings.0.postgres_source.0.connection.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise"); ok {
		onPremise, err := expandDatatransferEndpointSettingsPostgresSourceConnectionOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresSourceConnectionOnPremise(d *schema.ResourceData) (*endpoint.OnPremisePostgres, error) {
	val := new(endpoint.OnPremisePostgres)

	if v, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.hosts"); ok {
		val.SetHosts(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.port"); ok {
		val.SetPort(int64(v.(int)))
	}

	if v, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.postgres_source.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlTarget(d *schema.ResourceData) (*endpoint.MysqlTarget, error) {
	val := new(endpoint.MysqlTarget)

	if _, ok := d.GetOk("settings.0.mysql_target.0.connection"); ok {
		connection, err := expandDatatransferEndpointSettingsMysqlTargetConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mysql_target.0.password"); ok {
		password, err := expandDatatransferEndpointSettingsMysqlTargetPassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.skip_constraint_checks"); ok {
		val.SetSkipConstraintChecks(v.(bool))
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.sql_mode"); ok {
		val.SetSqlMode(v.(string))
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.timezone"); ok {
		val.SetTimezone(v.(string))
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.user"); ok {
		val.SetUser(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlTargetPassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.mysql_target.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlTargetConnection(d *schema.ResourceData) (*endpoint.MysqlConnection, error) {
	val := new(endpoint.MysqlConnection)

	if v, ok := d.GetOk("settings.0.mysql_target.0.connection.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise"); ok {
		onPremise, err := expandDatatransferEndpointSettingsMysqlTargetConnectionOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlTargetConnectionOnPremise(d *schema.ResourceData) (*endpoint.OnPremiseMysql, error) {
	val := new(endpoint.OnPremiseMysql)

	if v, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.hosts"); ok {
		val.SetHosts(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.port"); ok {
		val.SetPort(int64(v.(int)))
	}

	if v, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.mysql_target.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlSource(d *schema.ResourceData) (*endpoint.MysqlSource, error) {
	val := new(endpoint.MysqlSource)

	if _, ok := d.GetOk("settings.0.mysql_source.0.connection"); ok {
		connection, err := expandDatatransferEndpointSettingsMysqlSourceConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.exclude_tables_regex"); ok {
		val.SetExcludeTablesRegex(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.include_tables_regex"); ok {
		val.SetIncludeTablesRegex(expandStringSlice(v.([]interface{})))
	}

	if _, ok := d.GetOk("settings.0.mysql_source.0.object_transfer_settings"); ok {
		objectTransferSettings, err := expandDatatransferEndpointSettingsMysqlSourceObjectTransferSettings(d)
		if err != nil {
			return nil, err
		}

		val.SetObjectTransferSettings(objectTransferSettings)
	}

	if _, ok := d.GetOk("settings.0.mysql_source.0.password"); ok {
		password, err := expandDatatransferEndpointSettingsMysqlSourcePassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.service_database"); ok {
		val.SetServiceDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.timezone"); ok {
		val.SetTimezone(v.(string))
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.user"); ok {
		val.SetUser(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlSourcePassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.mysql_source.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlSourceObjectTransferSettings(d *schema.ResourceData) (*endpoint.MysqlObjectTransferSettings, error) {
	val := new(endpoint.MysqlObjectTransferSettings)

	if v, ok := d.GetOk("settings.0.mysql_source.0.object_transfer_settings.0.routine"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetRoutine(vv)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.object_transfer_settings.0.trigger"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetTrigger(vv)
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.object_transfer_settings.0.view"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetView(vv)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlSourceConnection(d *schema.ResourceData) (*endpoint.MysqlConnection, error) {
	val := new(endpoint.MysqlConnection)

	if v, ok := d.GetOk("settings.0.mysql_source.0.connection.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise"); ok {
		onPremise, err := expandDatatransferEndpointSettingsMysqlSourceConnectionOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlSourceConnectionOnPremise(d *schema.ResourceData) (*endpoint.OnPremiseMysql, error) {
	val := new(endpoint.OnPremiseMysql)

	if v, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.hosts"); ok {
		val.SetHosts(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.port"); ok {
		val.SetPort(int64(v.(int)))
	}

	if v, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.mysql_source.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsMongoTarget(d *schema.ResourceData) (*endpoint.MongoTarget, error) {
	val := new(endpoint.MongoTarget)

	if v, ok := d.GetOk("settings.0.mongo_target.0.cleanup_policy"); ok {
		vv, err := parseDatatransferEndpointCleanupPolicy(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetCleanupPolicy(vv)
	}

	if _, ok := d.GetOk("settings.0.mongo_target.0.connection"); ok {
		connection, err := expandDatatransferEndpointSettingsMongoTargetConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.mongo_target.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.mongo_target.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.mongo_target.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoTargetConnection(d *schema.ResourceData) (*endpoint.MongoConnection, error) {
	val := new(endpoint.MongoConnection)

	if _, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options"); ok {
		connectionOptions, err := expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptions(d)
		if err != nil {
			return nil, err
		}

		val.SetConnectionOptions(connectionOptions)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptions(d *schema.ResourceData) (*endpoint.MongoConnectionOptions, error) {
	val := new(endpoint.MongoConnectionOptions)

	if v, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.auth_source"); ok {
		val.SetAuthSource(v.(string))
	}

	if v, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.on_premise"); ok {
		onPremise, err := expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	if _, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.password"); ok {
		password, err := expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsPassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.user"); ok {
		val.SetUser(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsPassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremise(d *schema.ResourceData) (*endpoint.OnPremiseMongo, error) {
	val := new(endpoint.OnPremiseMongo)

	if v, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.on_premise.0.hosts"); ok {
		val.SetHosts(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.on_premise.0.port"); ok {
		val.SetPort(int64(v.(int)))
	}

	if v, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.on_premise.0.replica_set"); ok {
		val.SetReplicaSet(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsMongoSource(d *schema.ResourceData) (*endpoint.MongoSource, error) {
	val := new(endpoint.MongoSource)

	if _, ok := d.GetOk("settings.0.mongo_source.0.collections"); ok {
		collections, err := expandDatatransferEndpointSettingsMongoSourceCollectionsSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetCollections(collections)
	}

	if _, ok := d.GetOk("settings.0.mongo_source.0.connection"); ok {
		connection, err := expandDatatransferEndpointSettingsMongoSourceConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if _, ok := d.GetOk("settings.0.mongo_source.0.excluded_collections"); ok {
		excludedCollections, err := expandDatatransferEndpointSettingsMongoSourceExcludedCollectionsSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetExcludedCollections(excludedCollections)
	}

	if v, ok := d.GetOk("settings.0.mongo_source.0.secondary_preferred_mode"); ok {
		val.SetSecondaryPreferredMode(v.(bool))
	}

	if v, ok := d.GetOk("settings.0.mongo_source.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.mongo_source.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoSourceExcludedCollectionsSlice(d *schema.ResourceData) ([]*endpoint.MongoCollection, error) {
	count := d.Get("settings.0.mongo_source.0.excluded_collections.#").(int)
	slice := make([]*endpoint.MongoCollection, count)

	for i := 0; i < count; i++ {
		expandedItem, err := expandDatatransferEndpointSettingsMongoSourceExcludedCollections(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsMongoSourceExcludedCollections(d *schema.ResourceData, indexes ...interface{}) (*endpoint.MongoCollection, error) {
	val := new(endpoint.MongoCollection)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.mongo_source.0.excluded_collections.%d.collection_name", indexes...)); ok {
		val.SetCollectionName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.mongo_source.0.excluded_collections.%d.database_name", indexes...)); ok {
		val.SetDatabaseName(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoSourceConnection(d *schema.ResourceData) (*endpoint.MongoConnection, error) {
	val := new(endpoint.MongoConnection)

	if _, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options"); ok {
		connectionOptions, err := expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptions(d)
		if err != nil {
			return nil, err
		}

		val.SetConnectionOptions(connectionOptions)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptions(d *schema.ResourceData) (*endpoint.MongoConnectionOptions, error) {
	val := new(endpoint.MongoConnectionOptions)

	if v, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.auth_source"); ok {
		val.SetAuthSource(v.(string))
	}

	if v, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.on_premise"); ok {
		onPremise, err := expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	if _, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.password"); ok {
		password, err := expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsPassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.user"); ok {
		val.SetUser(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsPassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremise(d *schema.ResourceData) (*endpoint.OnPremiseMongo, error) {
	val := new(endpoint.OnPremiseMongo)

	if v, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.on_premise.0.hosts"); ok {
		val.SetHosts(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.on_premise.0.port"); ok {
		val.SetPort(int64(v.(int)))
	}

	if v, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.on_premise.0.replica_set"); ok {
		val.SetReplicaSet(v.(string))
	}

	if _, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsMongoSourceCollectionsSlice(d *schema.ResourceData) ([]*endpoint.MongoCollection, error) {
	count := d.Get("settings.0.mongo_source.0.collections.#").(int)
	slice := make([]*endpoint.MongoCollection, count)

	for i := 0; i < count; i++ {
		expandedItem, err := expandDatatransferEndpointSettingsMongoSourceCollections(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsMongoSourceCollections(d *schema.ResourceData, indexes ...interface{}) (*endpoint.MongoCollection, error) {
	val := new(endpoint.MongoCollection)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.mongo_source.0.collections.%d.collection_name", indexes...)); ok {
		val.SetCollectionName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.mongo_source.0.collections.%d.database_name", indexes...)); ok {
		val.SetDatabaseName(v.(string))
	}

	return val, nil
}

func flattenDatatransferEndpointSettings(d *schema.ResourceData, v *datatransfer.EndpointSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	mongoSource, err := flattenDatatransferEndpointSettingsMongoSource(d, v.GetMongoSource())
	if err != nil {
		return nil, err
	}
	m["mongo_source"] = mongoSource

	mongoTarget, err := flattenDatatransferEndpointSettingsMongoTarget(d, v.GetMongoTarget())
	if err != nil {
		return nil, err
	}
	m["mongo_target"] = mongoTarget

	mysqlSource, err := flattenDatatransferEndpointSettingsMysqlSource(d, v.GetMysqlSource())
	if err != nil {
		return nil, err
	}
	m["mysql_source"] = mysqlSource

	mysqlTarget, err := flattenDatatransferEndpointSettingsMysqlTarget(d, v.GetMysqlTarget())
	if err != nil {
		return nil, err
	}
	m["mysql_target"] = mysqlTarget

	postgresSource, err := flattenDatatransferEndpointSettingsPostgresSource(d, v.GetPostgresSource())
	if err != nil {
		return nil, err
	}
	m["postgres_source"] = postgresSource

	postgresTarget, err := flattenDatatransferEndpointSettingsPostgresTarget(d, v.GetPostgresTarget())
	if err != nil {
		return nil, err
	}
	m["postgres_target"] = postgresTarget

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresTarget(d *schema.ResourceData, v *endpoint.PostgresTarget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connection, err := flattenDatatransferEndpointSettingsPostgresTargetConnection(d, v.GetConnection())
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["database"] = v.GetDatabase()
	if password, ok := d.GetOk("settings.0.postgres_target.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["security_groups"] = v.GetSecurityGroups()
	m["user"] = v.GetUser()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresTargetConnection(d *schema.ResourceData, v *endpoint.PostgresConnection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["mdb_cluster_id"] = v.GetMdbClusterId()

	onPremise, err := flattenDatatransferEndpointSettingsPostgresTargetConnectionOnPremise(d, v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresTargetConnectionOnPremise(d *schema.ResourceData, v *endpoint.OnPremisePostgres) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["hosts"] = v.GetHosts()
	m["port"] = v.GetPort()
	m["subnet_id"] = v.GetSubnetId()

	tlsMode, err := flattenDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsMode(d, v.GetTlsMode())
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsMode(d *schema.ResourceData, v *endpoint.TLSMode) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	disabled, err := flattenDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeDisabled(d, v.GetDisabled())
	if err != nil {
		return nil, err
	}
	m["disabled"] = disabled

	enabled, err := flattenDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeEnabled(d, v.GetEnabled())
	if err != nil {
		return nil, err
	}
	m["enabled"] = enabled

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData, v *endpoint.TLSConfig) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ca_certificate"] = v.GetCaCertificate()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresTargetConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresSource(d *schema.ResourceData, v *endpoint.PostgresSource) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connection, err := flattenDatatransferEndpointSettingsPostgresSourceConnection(d, v.GetConnection())
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["database"] = v.GetDatabase()
	m["exclude_tables"] = v.GetExcludeTables()
	m["include_tables"] = v.GetIncludeTables()

	objectTransferSettings, err := flattenDatatransferEndpointSettingsPostgresSourceObjectTransferSettings(d, v.GetObjectTransferSettings())
	if err != nil {
		return nil, err
	}
	m["object_transfer_settings"] = objectTransferSettings
	if password, ok := d.GetOk("settings.0.postgres_source.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["security_groups"] = v.GetSecurityGroups()
	m["service_schema"] = v.GetServiceSchema()
	m["slot_gigabyte_lag_limit"] = toGigabytes(v.GetSlotByteLagLimit())
	m["user"] = v.GetUser()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresSourceObjectTransferSettings(d *schema.ResourceData, v *endpoint.PostgresObjectTransferSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["cast"] = v.GetCast().String()
	m["collation"] = v.GetCollation().String()
	m["constraint"] = v.GetConstraint().String()
	m["default_values"] = v.GetDefaultValues().String()
	m["fk_constraint"] = v.GetFkConstraint().String()
	m["function"] = v.GetFunction().String()
	m["index"] = v.GetIndex().String()
	m["materialized_view"] = v.GetMaterializedView().String()
	m["policy"] = v.GetPolicy().String()
	m["primary_key"] = v.GetPrimaryKey().String()
	m["rule"] = v.GetRule().String()
	m["sequence"] = v.GetSequence().String()
	m["sequence_owned_by"] = v.GetSequenceOwnedBy().String()
	m["table"] = v.GetTable().String()
	m["trigger"] = v.GetTrigger().String()
	m["type"] = v.GetType().String()
	m["view"] = v.GetView().String()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresSourceConnection(d *schema.ResourceData, v *endpoint.PostgresConnection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["mdb_cluster_id"] = v.GetMdbClusterId()

	onPremise, err := flattenDatatransferEndpointSettingsPostgresSourceConnectionOnPremise(d, v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresSourceConnectionOnPremise(d *schema.ResourceData, v *endpoint.OnPremisePostgres) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["hosts"] = v.GetHosts()
	m["port"] = v.GetPort()
	m["subnet_id"] = v.GetSubnetId()

	tlsMode, err := flattenDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsMode(d, v.GetTlsMode())
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsMode(d *schema.ResourceData, v *endpoint.TLSMode) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	disabled, err := flattenDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeDisabled(d, v.GetDisabled())
	if err != nil {
		return nil, err
	}
	m["disabled"] = disabled

	enabled, err := flattenDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeEnabled(d, v.GetEnabled())
	if err != nil {
		return nil, err
	}
	m["enabled"] = enabled

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData, v *endpoint.TLSConfig) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ca_certificate"] = v.GetCaCertificate()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresSourceConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlTarget(d *schema.ResourceData, v *endpoint.MysqlTarget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connection, err := flattenDatatransferEndpointSettingsMysqlTargetConnection(d, v.GetConnection())
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["database"] = v.GetDatabase()
	if password, ok := d.GetOk("settings.0.mysql_target.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["security_groups"] = v.GetSecurityGroups()
	m["skip_constraint_checks"] = v.GetSkipConstraintChecks()
	m["sql_mode"] = v.GetSqlMode()
	m["timezone"] = v.GetTimezone()
	m["user"] = v.GetUser()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlTargetConnection(d *schema.ResourceData, v *endpoint.MysqlConnection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["mdb_cluster_id"] = v.GetMdbClusterId()

	onPremise, err := flattenDatatransferEndpointSettingsMysqlTargetConnectionOnPremise(d, v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlTargetConnectionOnPremise(d *schema.ResourceData, v *endpoint.OnPremiseMysql) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["hosts"] = v.GetHosts()
	m["port"] = v.GetPort()
	m["subnet_id"] = v.GetSubnetId()

	tlsMode, err := flattenDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsMode(d, v.GetTlsMode())
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsMode(d *schema.ResourceData, v *endpoint.TLSMode) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	disabled, err := flattenDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeDisabled(d, v.GetDisabled())
	if err != nil {
		return nil, err
	}
	m["disabled"] = disabled

	enabled, err := flattenDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeEnabled(d, v.GetEnabled())
	if err != nil {
		return nil, err
	}
	m["enabled"] = enabled

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData, v *endpoint.TLSConfig) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ca_certificate"] = v.GetCaCertificate()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlTargetConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlSource(d *schema.ResourceData, v *endpoint.MysqlSource) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connection, err := flattenDatatransferEndpointSettingsMysqlSourceConnection(d, v.GetConnection())
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["database"] = v.GetDatabase()
	m["exclude_tables_regex"] = v.GetExcludeTablesRegex()
	m["include_tables_regex"] = v.GetIncludeTablesRegex()

	objectTransferSettings, err := flattenDatatransferEndpointSettingsMysqlSourceObjectTransferSettings(d, v.GetObjectTransferSettings())
	if err != nil {
		return nil, err
	}
	m["object_transfer_settings"] = objectTransferSettings
	if password, ok := d.GetOk("settings.0.mysql_source.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["security_groups"] = v.GetSecurityGroups()
	m["service_database"] = v.GetServiceDatabase()
	m["timezone"] = v.GetTimezone()
	m["user"] = v.GetUser()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlSourceObjectTransferSettings(d *schema.ResourceData, v *endpoint.MysqlObjectTransferSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["routine"] = v.GetRoutine().String()
	m["trigger"] = v.GetTrigger().String()
	m["view"] = v.GetView().String()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlSourceConnection(d *schema.ResourceData, v *endpoint.MysqlConnection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["mdb_cluster_id"] = v.GetMdbClusterId()

	onPremise, err := flattenDatatransferEndpointSettingsMysqlSourceConnectionOnPremise(d, v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlSourceConnectionOnPremise(d *schema.ResourceData, v *endpoint.OnPremiseMysql) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["hosts"] = v.GetHosts()
	m["port"] = v.GetPort()
	m["subnet_id"] = v.GetSubnetId()

	tlsMode, err := flattenDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsMode(d, v.GetTlsMode())
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsMode(d *schema.ResourceData, v *endpoint.TLSMode) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	disabled, err := flattenDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeDisabled(d, v.GetDisabled())
	if err != nil {
		return nil, err
	}
	m["disabled"] = disabled

	enabled, err := flattenDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeEnabled(d, v.GetEnabled())
	if err != nil {
		return nil, err
	}
	m["enabled"] = enabled

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData, v *endpoint.TLSConfig) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ca_certificate"] = v.GetCaCertificate()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMysqlSourceConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoTarget(d *schema.ResourceData, v *endpoint.MongoTarget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["cleanup_policy"] = v.GetCleanupPolicy().String()

	connection, err := flattenDatatransferEndpointSettingsMongoTargetConnection(d, v.GetConnection())
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["database"] = v.GetDatabase()
	m["security_groups"] = v.GetSecurityGroups()
	m["subnet_id"] = v.GetSubnetId()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoTargetConnection(d *schema.ResourceData, v *endpoint.MongoConnection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connectionOptions, err := flattenDatatransferEndpointSettingsMongoTargetConnectionConnectionOptions(d, v.GetConnectionOptions())
	if err != nil {
		return nil, err
	}
	m["connection_options"] = connectionOptions

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoTargetConnectionConnectionOptions(d *schema.ResourceData, v *endpoint.MongoConnectionOptions) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["auth_source"] = v.GetAuthSource()
	m["mdb_cluster_id"] = v.GetMdbClusterId()

	onPremise, err := flattenDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremise(d, v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise
	if password, ok := d.GetOk("settings.0.mongo_target.0.connection.0.connection_options.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["user"] = v.GetUser()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremise(d *schema.ResourceData, v *endpoint.OnPremiseMongo) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["hosts"] = v.GetHosts()
	m["port"] = v.GetPort()
	m["replica_set"] = v.GetReplicaSet()

	tlsMode, err := flattenDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsMode(d, v.GetTlsMode())
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsMode(d *schema.ResourceData, v *endpoint.TLSMode) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	disabled, err := flattenDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsModeDisabled(d, v.GetDisabled())
	if err != nil {
		return nil, err
	}
	m["disabled"] = disabled

	enabled, err := flattenDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsModeEnabled(d, v.GetEnabled())
	if err != nil {
		return nil, err
	}
	m["enabled"] = enabled

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsModeEnabled(d *schema.ResourceData, v *endpoint.TLSConfig) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ca_certificate"] = v.GetCaCertificate()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoTargetConnectionConnectionOptionsOnPremiseTlsModeDisabled(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoSource(d *schema.ResourceData, v *endpoint.MongoSource) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	collections, err := flattenDatatransferEndpointSettingsMongoSourceCollectionsSlice(d, v.GetCollections())
	if err != nil {
		return nil, err
	}
	m["collections"] = collections

	connection, err := flattenDatatransferEndpointSettingsMongoSourceConnection(d, v.GetConnection())
	if err != nil {
		return nil, err
	}
	m["connection"] = connection

	excludedCollections, err := flattenDatatransferEndpointSettingsMongoSourceExcludedCollectionsSlice(d, v.GetExcludedCollections())
	if err != nil {
		return nil, err
	}
	m["excluded_collections"] = excludedCollections
	m["secondary_preferred_mode"] = v.GetSecondaryPreferredMode()
	m["security_groups"] = v.GetSecurityGroups()
	m["subnet_id"] = v.GetSubnetId()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoSourceExcludedCollectionsSlice(d *schema.ResourceData, v []*endpoint.MongoCollection) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsMongoSourceExcludedCollections(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsMongoSourceExcludedCollections(d *schema.ResourceData, v *endpoint.MongoCollection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["collection_name"] = v.GetCollectionName()
	m["database_name"] = v.GetDatabaseName()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoSourceConnection(d *schema.ResourceData, v *endpoint.MongoConnection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connectionOptions, err := flattenDatatransferEndpointSettingsMongoSourceConnectionConnectionOptions(d, v.GetConnectionOptions())
	if err != nil {
		return nil, err
	}
	m["connection_options"] = connectionOptions

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoSourceConnectionConnectionOptions(d *schema.ResourceData, v *endpoint.MongoConnectionOptions) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["auth_source"] = v.GetAuthSource()
	m["mdb_cluster_id"] = v.GetMdbClusterId()

	onPremise, err := flattenDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremise(d, v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise
	if password, ok := d.GetOk("settings.0.mongo_source.0.connection.0.connection_options.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["user"] = v.GetUser()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremise(d *schema.ResourceData, v *endpoint.OnPremiseMongo) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["hosts"] = v.GetHosts()
	m["port"] = v.GetPort()
	m["replica_set"] = v.GetReplicaSet()

	tlsMode, err := flattenDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsMode(d, v.GetTlsMode())
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsMode(d *schema.ResourceData, v *endpoint.TLSMode) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	disabled, err := flattenDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsModeDisabled(d, v.GetDisabled())
	if err != nil {
		return nil, err
	}
	m["disabled"] = disabled

	enabled, err := flattenDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsModeEnabled(d, v.GetEnabled())
	if err != nil {
		return nil, err
	}
	m["enabled"] = enabled

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsModeEnabled(d *schema.ResourceData, v *endpoint.TLSConfig) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ca_certificate"] = v.GetCaCertificate()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoSourceConnectionConnectionOptionsOnPremiseTlsModeDisabled(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMongoSourceCollectionsSlice(d *schema.ResourceData, v []*endpoint.MongoCollection) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsMongoSourceCollections(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsMongoSourceCollections(d *schema.ResourceData, v *endpoint.MongoCollection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["collection_name"] = v.GetCollectionName()
	m["database_name"] = v.GetDatabaseName()

	return []map[string]interface{}{m}, nil
}

// Enum types parsers

func parseDatatransferTransferStatus(str string) (datatransfer.TransferStatus, error) {
	val, ok := datatransfer.TransferStatus_value[str]
	if !ok {
		return datatransfer.TransferStatus(0), fmt.Errorf(
			"value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(datatransfer.TransferStatus_value)),
			str,
		)
	}
	return datatransfer.TransferStatus(val), nil
}

func parseDatatransferTransferType(str string) (datatransfer.TransferType, error) {
	val, ok := datatransfer.TransferType_value[str]
	if !ok {
		return datatransfer.TransferType(0), fmt.Errorf(
			"value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(datatransfer.TransferType_value)),
			str,
		)
	}
	return datatransfer.TransferType(val), nil
}

func parseDatatransferEndpointCleanupPolicy(str string) (endpoint.CleanupPolicy, error) {
	val, ok := endpoint.CleanupPolicy_value[str]
	if !ok {
		return endpoint.CleanupPolicy(0), fmt.Errorf(
			"value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(endpoint.CleanupPolicy_value)),
			str,
		)
	}
	return endpoint.CleanupPolicy(val), nil
}

func parseDatatransferEndpointClickhouseCleanupPolicy(str string) (endpoint.ClickhouseCleanupPolicy, error) {
	val, ok := endpoint.ClickhouseCleanupPolicy_value[str]
	if !ok {
		return endpoint.ClickhouseCleanupPolicy(0), fmt.Errorf(
			"value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(endpoint.ClickhouseCleanupPolicy_value)),
			str,
		)
	}
	return endpoint.ClickhouseCleanupPolicy(val), nil
}

func parseDatatransferEndpointObjectTransferStage(str string) (endpoint.ObjectTransferStage, error) {
	val, ok := endpoint.ObjectTransferStage_value[str]
	if !ok {
		return endpoint.ObjectTransferStage(0), fmt.Errorf(
			"value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(endpoint.ObjectTransferStage_value)),
			str,
		)
	}
	return endpoint.ObjectTransferStage(val), nil
}
