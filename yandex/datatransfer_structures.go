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

	if _, ok := d.GetOk("settings.0.clickhouse_source"); ok {
		clickhouseSource, err := expandDatatransferEndpointSettingsClickhouseSource(d)
		if err != nil {
			return nil, err
		}

		val.SetClickhouseSource(clickhouseSource)
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target"); ok {
		clickhouseTarget, err := expandDatatransferEndpointSettingsClickhouseTarget(d)
		if err != nil {
			return nil, err
		}

		val.SetClickhouseTarget(clickhouseTarget)
	}

	if _, ok := d.GetOk("settings.0.kafka_source"); ok {
		kafkaSource, err := expandDatatransferEndpointSettingsKafkaSource(d)
		if err != nil {
			return nil, err
		}

		val.SetKafkaSource(kafkaSource)
	}

	if _, ok := d.GetOk("settings.0.kafka_target"); ok {
		kafkaTarget, err := expandDatatransferEndpointSettingsKafkaTarget(d)
		if err != nil {
			return nil, err
		}

		val.SetKafkaTarget(kafkaTarget)
	}

	if _, ok := d.GetOk("settings.0.metrika_source"); ok {
		metrikaSource, err := expandDatatransferEndpointSettingsMetrikaSource(d)
		if err != nil {
			return nil, err
		}

		val.SetMetrikaSource(metrikaSource)
	}

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

	if _, ok := d.GetOk("settings.0.ydb_source"); ok {
		ydbSource, err := expandDatatransferEndpointSettingsYdbSource(d)
		if err != nil {
			return nil, err
		}

		val.SetYdbSource(ydbSource)
	}

	if _, ok := d.GetOk("settings.0.ydb_target"); ok {
		ydbTarget, err := expandDatatransferEndpointSettingsYdbTarget(d)
		if err != nil {
			return nil, err
		}

		val.SetYdbTarget(ydbTarget)
	}

	if _, ok := d.GetOk("settings.0.yds_source"); ok {
		ydsSource, err := expandDatatransferEndpointSettingsYdsSource(d)
		if err != nil {
			return nil, err
		}

		val.SetYdsSource(ydsSource)
	}

	if _, ok := d.GetOk("settings.0.yds_target"); ok {
		ydsTarget, err := expandDatatransferEndpointSettingsYdsTarget(d)
		if err != nil {
			return nil, err
		}

		val.SetYdsTarget(ydsTarget)
	}

	empty := new(datatransfer.EndpointSettings)
	if proto.Equal(val, empty) {
		return nil, nil
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsTarget(d *schema.ResourceData) (*endpoint.YDSTarget, error) {
	val := new(endpoint.YDSTarget)

	if v, ok := d.GetOk("settings.0.yds_target.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.yds_target.0.endpoint"); ok {
		val.SetEndpoint(v.(string))
	}

	if v, ok := d.GetOk("settings.0.yds_target.0.save_tx_order"); ok {
		val.SetSaveTxOrder(v.(bool))
	}

	if v, ok := d.GetOk("settings.0.yds_target.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if _, ok := d.GetOk("settings.0.yds_target.0.serializer"); ok {
		serializer, err := expandDatatransferEndpointSettingsYdsTargetSerializer(d)
		if err != nil {
			return nil, err
		}

		val.SetSerializer(serializer)
	}

	if v, ok := d.GetOk("settings.0.yds_target.0.service_account_id"); ok {
		val.SetServiceAccountId(v.(string))
	}

	if v, ok := d.GetOk("settings.0.yds_target.0.stream"); ok {
		val.SetStream(v.(string))
	}

	if v, ok := d.GetOk("settings.0.yds_target.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsTargetSerializer(d *schema.ResourceData) (*endpoint.Serializer, error) {
	val := new(endpoint.Serializer)

	if _, ok := d.GetOk("settings.0.yds_target.0.serializer.0.serializer_auto"); ok {
		serializerAuto, err := expandDatatransferEndpointSettingsYdsTargetSerializerSerializerAuto(d)
		if err != nil {
			return nil, err
		}

		val.SetSerializerAuto(serializerAuto)
	}

	if _, ok := d.GetOk("settings.0.yds_target.0.serializer.0.serializer_debezium"); ok {
		serializerDebezium, err := expandDatatransferEndpointSettingsYdsTargetSerializerSerializerDebezium(d)
		if err != nil {
			return nil, err
		}

		val.SetSerializerDebezium(serializerDebezium)
	}

	if _, ok := d.GetOk("settings.0.yds_target.0.serializer.0.serializer_json"); ok {
		serializerJson, err := expandDatatransferEndpointSettingsYdsTargetSerializerSerializerJson(d)
		if err != nil {
			return nil, err
		}

		val.SetSerializerJson(serializerJson)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsTargetSerializerSerializerJson(d *schema.ResourceData) (*endpoint.SerializerJSON, error) {
	val := new(endpoint.SerializerJSON)

	return val, nil
}

func expandDatatransferEndpointSettingsYdsTargetSerializerSerializerDebezium(d *schema.ResourceData) (*endpoint.SerializerDebezium, error) {
	val := new(endpoint.SerializerDebezium)

	if _, ok := d.GetOk("settings.0.yds_target.0.serializer.0.serializer_debezium.0.serializer_parameters"); ok {
		serializerParameters, err := expandDatatransferEndpointSettingsYdsTargetSerializerSerializerDebeziumSerializerParametersSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetSerializerParameters(serializerParameters)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsTargetSerializerSerializerDebeziumSerializerParametersSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.DebeziumSerializerParameter, error) {
	count := d.Get("settings.0.yds_target.0.serializer.0.serializer_debezium.0.serializer_parameters.#").(int)
	slice := make([]*endpoint.DebeziumSerializerParameter, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsYdsTargetSerializerSerializerDebeziumSerializerParameters(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsYdsTargetSerializerSerializerDebeziumSerializerParameters(d *schema.ResourceData, indexes ...interface{}) (*endpoint.DebeziumSerializerParameter, error) {
	val := new(endpoint.DebeziumSerializerParameter)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_target.0.serializer.0.serializer_debezium.0.serializer_parameters.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_target.0.serializer.0.serializer_debezium.0.serializer_parameters.%d.value", indexes...)); ok {
		val.SetValue(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsTargetSerializerSerializerAuto(d *schema.ResourceData) (*endpoint.SerializerAuto, error) {
	val := new(endpoint.SerializerAuto)

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSource(d *schema.ResourceData) (*endpoint.YDSSource, error) {
	val := new(endpoint.YDSSource)

	if v, ok := d.GetOk("settings.0.yds_source.0.allow_ttl_rewind"); ok {
		val.SetAllowTtlRewind(v.(bool))
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.consumer"); ok {
		val.SetConsumer(v.(string))
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.endpoint"); ok {
		val.SetEndpoint(v.(string))
	}

	if _, ok := d.GetOk("settings.0.yds_source.0.parser"); ok {
		parser, err := expandDatatransferEndpointSettingsYdsSourceParser(d)
		if err != nil {
			return nil, err
		}

		val.SetParser(parser)
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.service_account_id"); ok {
		val.SetServiceAccountId(v.(string))
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.stream"); ok {
		val.SetStream(v.(string))
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.yds_source.0.supported_codecs"); ok {
		supportedCodecs, err := expandDatatransferEndpointSettingsYdsSourceSupportedCodecsSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetSupportedCodecs(supportedCodecs)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSourceSupportedCodecsSlice(d *schema.ResourceData) ([]endpoint.YdsCompressionCodec, error) {
	count := d.Get("settings.0.yds_source.0.supported_codecs.#").(int)
	slice := make([]endpoint.YdsCompressionCodec, count)

	for i := 0; i < count; i++ {
		item := d.Get(fmt.Sprintf("settings.0.yds_source.0.supported_codecs.%d", i))
		expandedItem, err := parseDatatransferEndpointYdsCompressionCodec(item.(string))
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsYdsSourceParser(d *schema.ResourceData) (*endpoint.Parser, error) {
	val := new(endpoint.Parser)

	if _, ok := d.GetOk("settings.0.yds_source.0.parser.0.audit_trails_v1_parser"); ok {
		auditTrailsV1Parser, err := expandDatatransferEndpointSettingsYdsSourceParserAuditTrailsV1Parser(d)
		if err != nil {
			return nil, err
		}

		val.SetAuditTrailsV1Parser(auditTrailsV1Parser)
	}

	if _, ok := d.GetOk("settings.0.yds_source.0.parser.0.cloud_logging_parser"); ok {
		cloudLoggingParser, err := expandDatatransferEndpointSettingsYdsSourceParserCloudLoggingParser(d)
		if err != nil {
			return nil, err
		}

		val.SetCloudLoggingParser(cloudLoggingParser)
	}

	if _, ok := d.GetOk("settings.0.yds_source.0.parser.0.json_parser"); ok {
		jsonParser, err := expandDatatransferEndpointSettingsYdsSourceParserJsonParser(d)
		if err != nil {
			return nil, err
		}

		val.SetJsonParser(jsonParser)
	}

	if _, ok := d.GetOk("settings.0.yds_source.0.parser.0.tskv_parser"); ok {
		tskvParser, err := expandDatatransferEndpointSettingsYdsSourceParserTskvParser(d)
		if err != nil {
			return nil, err
		}

		val.SetTskvParser(tskvParser)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserTskvParser(d *schema.ResourceData) (*endpoint.GenericParserCommon, error) {
	val := new(endpoint.GenericParserCommon)

	if v, ok := d.GetOk("settings.0.yds_source.0.parser.0.tskv_parser.0.add_rest_column"); ok {
		val.SetAddRestColumn(v.(bool))
	}

	if _, ok := d.GetOk("settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema"); ok {
		dataSchema, err := expandDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchema(d)
		if err != nil {
			return nil, err
		}

		val.SetDataSchema(dataSchema)
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.parser.0.tskv_parser.0.null_keys_allowed"); ok {
		val.SetNullKeysAllowed(v.(bool))
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.parser.0.tskv_parser.0.unescape_string_values"); ok {
		val.SetUnescapeStringValues(v.(bool))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchema(d *schema.ResourceData) (*endpoint.DataSchema, error) {
	val := new(endpoint.DataSchema)

	if _, ok := d.GetOk("settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema.0.fields"); ok {
		fields, err := expandDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFields(d)
		if err != nil {
			return nil, err
		}

		val.SetFields(fields)
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema.0.json_fields"); ok {
		val.SetJsonFields(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFields(d *schema.ResourceData) (*endpoint.FieldList, error) {
	val := new(endpoint.FieldList)

	if _, ok := d.GetOk("settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields"); ok {
		fields, err := expandDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFieldsFieldsSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetFields(fields)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFieldsFieldsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.ColSchema, error) {
	count := d.Get("settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.#").(int)
	slice := make([]*endpoint.ColSchema, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFieldsFields(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFieldsFields(d *schema.ResourceData, indexes ...interface{}) (*endpoint.ColSchema, error) {
	val := new(endpoint.ColSchema)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.%d.key", indexes...)); ok {
		val.SetKey(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.%d.path", indexes...)); ok {
		val.SetPath(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.%d.required", indexes...)); ok {
		val.SetRequired(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.%d.type", indexes...)); ok {
		vv, err := parseDatatransferEndpointColumnType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(vv)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserJsonParser(d *schema.ResourceData) (*endpoint.GenericParserCommon, error) {
	val := new(endpoint.GenericParserCommon)

	if v, ok := d.GetOk("settings.0.yds_source.0.parser.0.json_parser.0.add_rest_column"); ok {
		val.SetAddRestColumn(v.(bool))
	}

	if _, ok := d.GetOk("settings.0.yds_source.0.parser.0.json_parser.0.data_schema"); ok {
		dataSchema, err := expandDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchema(d)
		if err != nil {
			return nil, err
		}

		val.SetDataSchema(dataSchema)
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.parser.0.json_parser.0.null_keys_allowed"); ok {
		val.SetNullKeysAllowed(v.(bool))
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.parser.0.json_parser.0.unescape_string_values"); ok {
		val.SetUnescapeStringValues(v.(bool))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchema(d *schema.ResourceData) (*endpoint.DataSchema, error) {
	val := new(endpoint.DataSchema)

	if _, ok := d.GetOk("settings.0.yds_source.0.parser.0.json_parser.0.data_schema.0.fields"); ok {
		fields, err := expandDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFields(d)
		if err != nil {
			return nil, err
		}

		val.SetFields(fields)
	}

	if v, ok := d.GetOk("settings.0.yds_source.0.parser.0.json_parser.0.data_schema.0.json_fields"); ok {
		val.SetJsonFields(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFields(d *schema.ResourceData) (*endpoint.FieldList, error) {
	val := new(endpoint.FieldList)

	if _, ok := d.GetOk("settings.0.yds_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields"); ok {
		fields, err := expandDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFieldsFieldsSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetFields(fields)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFieldsFieldsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.ColSchema, error) {
	count := d.Get("settings.0.yds_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.#").(int)
	slice := make([]*endpoint.ColSchema, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFieldsFields(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFieldsFields(d *schema.ResourceData, indexes ...interface{}) (*endpoint.ColSchema, error) {
	val := new(endpoint.ColSchema)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.%d.key", indexes...)); ok {
		val.SetKey(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.%d.path", indexes...)); ok {
		val.SetPath(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.%d.required", indexes...)); ok {
		val.SetRequired(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.yds_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.%d.type", indexes...)); ok {
		vv, err := parseDatatransferEndpointColumnType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(vv)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserCloudLoggingParser(d *schema.ResourceData) (*endpoint.CloudLoggingParser, error) {
	val := new(endpoint.CloudLoggingParser)

	return val, nil
}

func expandDatatransferEndpointSettingsYdsSourceParserAuditTrailsV1Parser(d *schema.ResourceData) (*endpoint.AuditTrailsV1Parser, error) {
	val := new(endpoint.AuditTrailsV1Parser)

	return val, nil
}

func expandDatatransferEndpointSettingsYdbTarget(d *schema.ResourceData) (*endpoint.YdbTarget, error) {
	val := new(endpoint.YdbTarget)

	if v, ok := d.GetOk("settings.0.ydb_target.0.cleanup_policy"); ok {
		vv, err := parseDatatransferEndpointYdbCleanupPolicy(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetCleanupPolicy(vv)
	}

	if v, ok := d.GetOk("settings.0.ydb_target.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.ydb_target.0.default_compression"); ok {
		vv, err := parseDatatransferEndpointYdbDefaultCompression(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetDefaultCompression(vv)
	}

	if v, ok := d.GetOk("settings.0.ydb_target.0.instance"); ok {
		val.SetInstance(v.(string))
	}

	if v, ok := d.GetOk("settings.0.ydb_target.0.is_table_column_oriented"); ok {
		val.SetIsTableColumnOriented(v.(bool))
	}

	if v, ok := d.GetOk("settings.0.ydb_target.0.path"); ok {
		val.SetPath(v.(string))
	}

	if v, ok := d.GetOk("settings.0.ydb_target.0.sa_key_content"); ok {
		val.SetSaKeyContent(v.(string))
	}

	if v, ok := d.GetOk("settings.0.ydb_target.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.ydb_target.0.service_account_id"); ok {
		val.SetServiceAccountId(v.(string))
	}

	if v, ok := d.GetOk("settings.0.ydb_target.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsYdbSource(d *schema.ResourceData) (*endpoint.YdbSource, error) {
	val := new(endpoint.YdbSource)

	if v, ok := d.GetOk("settings.0.ydb_source.0.changefeed_custom_name"); ok {
		val.SetChangefeedCustomName(v.(string))
	}

	if v, ok := d.GetOk("settings.0.ydb_source.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.ydb_source.0.instance"); ok {
		val.SetInstance(v.(string))
	}

	if v, ok := d.GetOk("settings.0.ydb_source.0.paths"); ok {
		val.SetPaths(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.ydb_source.0.sa_key_content"); ok {
		val.SetSaKeyContent(v.(string))
	}

	if v, ok := d.GetOk("settings.0.ydb_source.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.ydb_source.0.service_account_id"); ok {
		val.SetServiceAccountId(v.(string))
	}

	if v, ok := d.GetOk("settings.0.ydb_source.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsPostgresTarget(d *schema.ResourceData) (*endpoint.PostgresTarget, error) {
	val := new(endpoint.PostgresTarget)

	if v, ok := d.GetOk("settings.0.postgres_target.0.cleanup_policy"); ok {
		vv, err := parseDatatransferEndpointCleanupPolicy(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetCleanupPolicy(vv)
	}

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

	if v, ok := d.GetOk("settings.0.postgres_source.0.object_transfer_settings.0.sequence_set"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetSequenceSet(vv)
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

	if v, ok := d.GetOk("settings.0.mysql_target.0.cleanup_policy"); ok {
		vv, err := parseDatatransferEndpointCleanupPolicy(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetCleanupPolicy(vv)
	}

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

	if v, ok := d.GetOk("settings.0.mysql_target.0.service_database"); ok {
		val.SetServiceDatabase(v.(string))
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

	if v, ok := d.GetOk("settings.0.mysql_source.0.object_transfer_settings.0.tables"); ok {
		vv, err := parseDatatransferEndpointObjectTransferStage(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetTables(vv)
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

func expandDatatransferEndpointSettingsMongoSourceExcludedCollectionsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.MongoCollection, error) {
	count := d.Get("settings.0.mongo_source.0.excluded_collections.#").(int)
	slice := make([]*endpoint.MongoCollection, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsMongoSourceExcludedCollections(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
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

func expandDatatransferEndpointSettingsMongoSourceCollectionsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.MongoCollection, error) {
	count := d.Get("settings.0.mongo_source.0.collections.#").(int)
	slice := make([]*endpoint.MongoCollection, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsMongoSourceCollections(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
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

func expandDatatransferEndpointSettingsMetrikaSource(d *schema.ResourceData) (*endpoint.MetrikaSource, error) {
	val := new(endpoint.MetrikaSource)

	if v, ok := d.GetOk("settings.0.metrika_source.0.counter_ids"); ok {
		val.SetCounterIds(expandInt64Slice(v.([]interface{})))
	}

	if _, ok := d.GetOk("settings.0.metrika_source.0.streams"); ok {
		streams, err := expandDatatransferEndpointSettingsMetrikaSourceStreamsSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetStreams(streams)
	}

	if _, ok := d.GetOk("settings.0.metrika_source.0.token"); ok {
		token, err := expandDatatransferEndpointSettingsMetrikaSourceToken(d)
		if err != nil {
			return nil, err
		}

		val.SetToken(token)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMetrikaSourceToken(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.metrika_source.0.token.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsMetrikaSourceStreamsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.MetrikaStream, error) {
	count := d.Get("settings.0.metrika_source.0.streams.#").(int)
	slice := make([]*endpoint.MetrikaStream, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsMetrikaSourceStreams(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsMetrikaSourceStreams(d *schema.ResourceData, indexes ...interface{}) (*endpoint.MetrikaStream, error) {
	val := new(endpoint.MetrikaStream)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.metrika_source.0.streams.%d.columns", indexes...)); ok {
		val.SetColumns(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.metrika_source.0.streams.%d.type", indexes...)); ok {
		vv, err := parseDatatransferEndpointMetrikaStreamType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(vv)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTarget(d *schema.ResourceData) (*endpoint.KafkaTarget, error) {
	val := new(endpoint.KafkaTarget)

	if _, ok := d.GetOk("settings.0.kafka_target.0.auth"); ok {
		auth, err := expandDatatransferEndpointSettingsKafkaTargetAuth(d)
		if err != nil {
			return nil, err
		}

		val.SetAuth(auth)
	}

	if _, ok := d.GetOk("settings.0.kafka_target.0.connection"); ok {
		connection, err := expandDatatransferEndpointSettingsKafkaTargetConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.kafka_target.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if _, ok := d.GetOk("settings.0.kafka_target.0.serializer"); ok {
		serializer, err := expandDatatransferEndpointSettingsKafkaTargetSerializer(d)
		if err != nil {
			return nil, err
		}

		val.SetSerializer(serializer)
	}

	if _, ok := d.GetOk("settings.0.kafka_target.0.topic_settings"); ok {
		topicSettings, err := expandDatatransferEndpointSettingsKafkaTargetTopicSettings(d)
		if err != nil {
			return nil, err
		}

		val.SetTopicSettings(topicSettings)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetTopicSettings(d *schema.ResourceData) (*endpoint.KafkaTargetTopicSettings, error) {
	val := new(endpoint.KafkaTargetTopicSettings)

	if _, ok := d.GetOk("settings.0.kafka_target.0.topic_settings.0.topic"); ok {
		topic, err := expandDatatransferEndpointSettingsKafkaTargetTopicSettingsTopic(d)
		if err != nil {
			return nil, err
		}

		val.SetTopic(topic)
	}

	if v, ok := d.GetOk("settings.0.kafka_target.0.topic_settings.0.topic_prefix"); ok {
		val.SetTopicPrefix(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetTopicSettingsTopic(d *schema.ResourceData) (*endpoint.KafkaTargetTopic, error) {
	val := new(endpoint.KafkaTargetTopic)

	if v, ok := d.GetOk("settings.0.kafka_target.0.topic_settings.0.topic.0.save_tx_order"); ok {
		val.SetSaveTxOrder(v.(bool))
	}

	if v, ok := d.GetOk("settings.0.kafka_target.0.topic_settings.0.topic.0.topic_name"); ok {
		val.SetTopicName(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetSerializer(d *schema.ResourceData) (*endpoint.Serializer, error) {
	val := new(endpoint.Serializer)

	if _, ok := d.GetOk("settings.0.kafka_target.0.serializer.0.serializer_auto"); ok {
		serializerAuto, err := expandDatatransferEndpointSettingsKafkaTargetSerializerSerializerAuto(d)
		if err != nil {
			return nil, err
		}

		val.SetSerializerAuto(serializerAuto)
	}

	if _, ok := d.GetOk("settings.0.kafka_target.0.serializer.0.serializer_debezium"); ok {
		serializerDebezium, err := expandDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebezium(d)
		if err != nil {
			return nil, err
		}

		val.SetSerializerDebezium(serializerDebezium)
	}

	if _, ok := d.GetOk("settings.0.kafka_target.0.serializer.0.serializer_json"); ok {
		serializerJson, err := expandDatatransferEndpointSettingsKafkaTargetSerializerSerializerJson(d)
		if err != nil {
			return nil, err
		}

		val.SetSerializerJson(serializerJson)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetSerializerSerializerJson(d *schema.ResourceData) (*endpoint.SerializerJSON, error) {
	val := new(endpoint.SerializerJSON)

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebezium(d *schema.ResourceData) (*endpoint.SerializerDebezium, error) {
	val := new(endpoint.SerializerDebezium)

	if _, ok := d.GetOk("settings.0.kafka_target.0.serializer.0.serializer_debezium.0.serializer_parameters"); ok {
		serializerParameters, err := expandDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebeziumSerializerParametersSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetSerializerParameters(serializerParameters)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebeziumSerializerParametersSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.DebeziumSerializerParameter, error) {
	count := d.Get("settings.0.kafka_target.0.serializer.0.serializer_debezium.0.serializer_parameters.#").(int)
	slice := make([]*endpoint.DebeziumSerializerParameter, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebeziumSerializerParameters(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebeziumSerializerParameters(d *schema.ResourceData, indexes ...interface{}) (*endpoint.DebeziumSerializerParameter, error) {
	val := new(endpoint.DebeziumSerializerParameter)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_target.0.serializer.0.serializer_debezium.0.serializer_parameters.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_target.0.serializer.0.serializer_debezium.0.serializer_parameters.%d.value", indexes...)); ok {
		val.SetValue(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetSerializerSerializerAuto(d *schema.ResourceData) (*endpoint.SerializerAuto, error) {
	val := new(endpoint.SerializerAuto)

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetConnection(d *schema.ResourceData) (*endpoint.KafkaConnectionOptions, error) {
	val := new(endpoint.KafkaConnectionOptions)

	if v, ok := d.GetOk("settings.0.kafka_target.0.connection.0.cluster_id"); ok {
		val.SetClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.kafka_target.0.connection.0.on_premise"); ok {
		onPremise, err := expandDatatransferEndpointSettingsKafkaTargetConnectionOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetConnectionOnPremise(d *schema.ResourceData) (*endpoint.OnPremiseKafka, error) {
	val := new(endpoint.OnPremiseKafka)

	if v, ok := d.GetOk("settings.0.kafka_target.0.connection.0.on_premise.0.broker_urls"); ok {
		val.SetBrokerUrls(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.kafka_target.0.connection.0.on_premise.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.kafka_target.0.connection.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.kafka_target.0.connection.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.kafka_target.0.connection.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.kafka_target.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetAuth(d *schema.ResourceData) (*endpoint.KafkaAuth, error) {
	val := new(endpoint.KafkaAuth)

	if _, ok := d.GetOk("settings.0.kafka_target.0.auth.0.no_auth"); ok {
		noAuth, err := expandDatatransferEndpointSettingsKafkaTargetAuthNoAuth(d)
		if err != nil {
			return nil, err
		}

		val.SetNoAuth(noAuth)
	}

	if _, ok := d.GetOk("settings.0.kafka_target.0.auth.0.sasl"); ok {
		sasl, err := expandDatatransferEndpointSettingsKafkaTargetAuthSasl(d)
		if err != nil {
			return nil, err
		}

		val.SetSasl(sasl)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetAuthSasl(d *schema.ResourceData) (*endpoint.KafkaSaslSecurity, error) {
	val := new(endpoint.KafkaSaslSecurity)

	if v, ok := d.GetOk("settings.0.kafka_target.0.auth.0.sasl.0.mechanism"); ok {
		vv, err := parseDatatransferEndpointKafkaMechanism(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetMechanism(vv)
	}

	if _, ok := d.GetOk("settings.0.kafka_target.0.auth.0.sasl.0.password"); ok {
		password, err := expandDatatransferEndpointSettingsKafkaTargetAuthSaslPassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.kafka_target.0.auth.0.sasl.0.user"); ok {
		val.SetUser(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetAuthSaslPassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.kafka_target.0.auth.0.sasl.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaTargetAuthNoAuth(d *schema.ResourceData) (*endpoint.NoAuth, error) {
	val := new(endpoint.NoAuth)

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSource(d *schema.ResourceData) (*endpoint.KafkaSource, error) {
	val := new(endpoint.KafkaSource)

	if _, ok := d.GetOk("settings.0.kafka_source.0.auth"); ok {
		auth, err := expandDatatransferEndpointSettingsKafkaSourceAuth(d)
		if err != nil {
			return nil, err
		}

		val.SetAuth(auth)
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.connection"); ok {
		connection, err := expandDatatransferEndpointSettingsKafkaSourceConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.parser"); ok {
		parser, err := expandDatatransferEndpointSettingsKafkaSourceParser(d)
		if err != nil {
			return nil, err
		}

		val.SetParser(parser)
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.topic_name"); ok {
		val.SetTopicName(v.(string))
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.topic_names"); ok {
		val.SetTopicNames(expandStringSlice(v.([]interface{})))
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.transformer"); ok {
		transformer, err := expandDatatransferEndpointSettingsKafkaSourceTransformer(d)
		if err != nil {
			return nil, err
		}

		val.SetTransformer(transformer)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceTransformer(d *schema.ResourceData) (*endpoint.DataTransformationOptions, error) {
	val := new(endpoint.DataTransformationOptions)

	if v, ok := d.GetOk("settings.0.kafka_source.0.transformer.0.buffer_flush_interval"); ok {
		val.SetBufferFlushInterval(v.(string))
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.transformer.0.buffer_size"); ok {
		val.SetBufferSize(v.(string))
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.transformer.0.cloud_function"); ok {
		val.SetCloudFunction(v.(string))
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.transformer.0.invocation_timeout"); ok {
		val.SetInvocationTimeout(v.(string))
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.transformer.0.number_of_retries"); ok {
		val.SetNumberOfRetries(int64(v.(int)))
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.transformer.0.service_account_id"); ok {
		val.SetServiceAccountId(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParser(d *schema.ResourceData) (*endpoint.Parser, error) {
	val := new(endpoint.Parser)

	if _, ok := d.GetOk("settings.0.kafka_source.0.parser.0.audit_trails_v1_parser"); ok {
		auditTrailsV1Parser, err := expandDatatransferEndpointSettingsKafkaSourceParserAuditTrailsV1Parser(d)
		if err != nil {
			return nil, err
		}

		val.SetAuditTrailsV1Parser(auditTrailsV1Parser)
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.parser.0.cloud_logging_parser"); ok {
		cloudLoggingParser, err := expandDatatransferEndpointSettingsKafkaSourceParserCloudLoggingParser(d)
		if err != nil {
			return nil, err
		}

		val.SetCloudLoggingParser(cloudLoggingParser)
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.parser.0.json_parser"); ok {
		jsonParser, err := expandDatatransferEndpointSettingsKafkaSourceParserJsonParser(d)
		if err != nil {
			return nil, err
		}

		val.SetJsonParser(jsonParser)
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.parser.0.tskv_parser"); ok {
		tskvParser, err := expandDatatransferEndpointSettingsKafkaSourceParserTskvParser(d)
		if err != nil {
			return nil, err
		}

		val.SetTskvParser(tskvParser)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserTskvParser(d *schema.ResourceData) (*endpoint.GenericParserCommon, error) {
	val := new(endpoint.GenericParserCommon)

	if v, ok := d.GetOk("settings.0.kafka_source.0.parser.0.tskv_parser.0.add_rest_column"); ok {
		val.SetAddRestColumn(v.(bool))
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema"); ok {
		dataSchema, err := expandDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchema(d)
		if err != nil {
			return nil, err
		}

		val.SetDataSchema(dataSchema)
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.parser.0.tskv_parser.0.null_keys_allowed"); ok {
		val.SetNullKeysAllowed(v.(bool))
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.parser.0.tskv_parser.0.unescape_string_values"); ok {
		val.SetUnescapeStringValues(v.(bool))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchema(d *schema.ResourceData) (*endpoint.DataSchema, error) {
	val := new(endpoint.DataSchema)

	if _, ok := d.GetOk("settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema.0.fields"); ok {
		fields, err := expandDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFields(d)
		if err != nil {
			return nil, err
		}

		val.SetFields(fields)
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema.0.json_fields"); ok {
		val.SetJsonFields(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFields(d *schema.ResourceData) (*endpoint.FieldList, error) {
	val := new(endpoint.FieldList)

	if _, ok := d.GetOk("settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields"); ok {
		fields, err := expandDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFieldsFieldsSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetFields(fields)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFieldsFieldsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.ColSchema, error) {
	count := d.Get("settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.#").(int)
	slice := make([]*endpoint.ColSchema, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFieldsFields(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFieldsFields(d *schema.ResourceData, indexes ...interface{}) (*endpoint.ColSchema, error) {
	val := new(endpoint.ColSchema)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.%d.key", indexes...)); ok {
		val.SetKey(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.%d.path", indexes...)); ok {
		val.SetPath(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.%d.required", indexes...)); ok {
		val.SetRequired(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_source.0.parser.0.tskv_parser.0.data_schema.0.fields.0.fields.%d.type", indexes...)); ok {
		vv, err := parseDatatransferEndpointColumnType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(vv)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserJsonParser(d *schema.ResourceData) (*endpoint.GenericParserCommon, error) {
	val := new(endpoint.GenericParserCommon)

	if v, ok := d.GetOk("settings.0.kafka_source.0.parser.0.json_parser.0.add_rest_column"); ok {
		val.SetAddRestColumn(v.(bool))
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.parser.0.json_parser.0.data_schema"); ok {
		dataSchema, err := expandDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchema(d)
		if err != nil {
			return nil, err
		}

		val.SetDataSchema(dataSchema)
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.parser.0.json_parser.0.null_keys_allowed"); ok {
		val.SetNullKeysAllowed(v.(bool))
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.parser.0.json_parser.0.unescape_string_values"); ok {
		val.SetUnescapeStringValues(v.(bool))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchema(d *schema.ResourceData) (*endpoint.DataSchema, error) {
	val := new(endpoint.DataSchema)

	if _, ok := d.GetOk("settings.0.kafka_source.0.parser.0.json_parser.0.data_schema.0.fields"); ok {
		fields, err := expandDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFields(d)
		if err != nil {
			return nil, err
		}

		val.SetFields(fields)
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.parser.0.json_parser.0.data_schema.0.json_fields"); ok {
		val.SetJsonFields(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFields(d *schema.ResourceData) (*endpoint.FieldList, error) {
	val := new(endpoint.FieldList)

	if _, ok := d.GetOk("settings.0.kafka_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields"); ok {
		fields, err := expandDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFieldsFieldsSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetFields(fields)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFieldsFieldsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.ColSchema, error) {
	count := d.Get("settings.0.kafka_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.#").(int)
	slice := make([]*endpoint.ColSchema, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFieldsFields(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFieldsFields(d *schema.ResourceData, indexes ...interface{}) (*endpoint.ColSchema, error) {
	val := new(endpoint.ColSchema)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.%d.key", indexes...)); ok {
		val.SetKey(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.%d.path", indexes...)); ok {
		val.SetPath(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.%d.required", indexes...)); ok {
		val.SetRequired(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.kafka_source.0.parser.0.json_parser.0.data_schema.0.fields.0.fields.%d.type", indexes...)); ok {
		vv, err := parseDatatransferEndpointColumnType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(vv)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserCloudLoggingParser(d *schema.ResourceData) (*endpoint.CloudLoggingParser, error) {
	val := new(endpoint.CloudLoggingParser)

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceParserAuditTrailsV1Parser(d *schema.ResourceData) (*endpoint.AuditTrailsV1Parser, error) {
	val := new(endpoint.AuditTrailsV1Parser)

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceConnection(d *schema.ResourceData) (*endpoint.KafkaConnectionOptions, error) {
	val := new(endpoint.KafkaConnectionOptions)

	if v, ok := d.GetOk("settings.0.kafka_source.0.connection.0.cluster_id"); ok {
		val.SetClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.connection.0.on_premise"); ok {
		onPremise, err := expandDatatransferEndpointSettingsKafkaSourceConnectionOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceConnectionOnPremise(d *schema.ResourceData) (*endpoint.OnPremiseKafka, error) {
	val := new(endpoint.OnPremiseKafka)

	if v, ok := d.GetOk("settings.0.kafka_source.0.connection.0.on_premise.0.broker_urls"); ok {
		val.SetBrokerUrls(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.connection.0.on_premise.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.connection.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.kafka_source.0.connection.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.connection.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.kafka_source.0.connection.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceAuth(d *schema.ResourceData) (*endpoint.KafkaAuth, error) {
	val := new(endpoint.KafkaAuth)

	if _, ok := d.GetOk("settings.0.kafka_source.0.auth.0.no_auth"); ok {
		noAuth, err := expandDatatransferEndpointSettingsKafkaSourceAuthNoAuth(d)
		if err != nil {
			return nil, err
		}

		val.SetNoAuth(noAuth)
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.auth.0.sasl"); ok {
		sasl, err := expandDatatransferEndpointSettingsKafkaSourceAuthSasl(d)
		if err != nil {
			return nil, err
		}

		val.SetSasl(sasl)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceAuthSasl(d *schema.ResourceData) (*endpoint.KafkaSaslSecurity, error) {
	val := new(endpoint.KafkaSaslSecurity)

	if v, ok := d.GetOk("settings.0.kafka_source.0.auth.0.sasl.0.mechanism"); ok {
		vv, err := parseDatatransferEndpointKafkaMechanism(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetMechanism(vv)
	}

	if _, ok := d.GetOk("settings.0.kafka_source.0.auth.0.sasl.0.password"); ok {
		password, err := expandDatatransferEndpointSettingsKafkaSourceAuthSaslPassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.kafka_source.0.auth.0.sasl.0.user"); ok {
		val.SetUser(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceAuthSaslPassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.kafka_source.0.auth.0.sasl.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsKafkaSourceAuthNoAuth(d *schema.ResourceData) (*endpoint.NoAuth, error) {
	val := new(endpoint.NoAuth)

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTarget(d *schema.ResourceData) (*endpoint.ClickhouseTarget, error) {
	val := new(endpoint.ClickhouseTarget)

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.alt_names"); ok {
		altNames, err := expandDatatransferEndpointSettingsClickhouseTargetAltNamesSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetAltNames(altNames)
	}

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.cleanup_policy"); ok {
		vv, err := parseDatatransferEndpointClickhouseCleanupPolicy(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetCleanupPolicy(vv)
	}

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.clickhouse_cluster_name"); ok {
		val.SetClickhouseClusterName(v.(string))
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.connection"); ok {
		connection, err := expandDatatransferEndpointSettingsClickhouseTargetConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.sharding"); ok {
		sharding, err := expandDatatransferEndpointSettingsClickhouseTargetSharding(d)
		if err != nil {
			return nil, err
		}

		val.SetSharding(sharding)
	}

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetSharding(d *schema.ResourceData) (*endpoint.ClickhouseSharding, error) {
	val := new(endpoint.ClickhouseSharding)

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.sharding.0.column_value_hash"); ok {
		columnValueHash, err := expandDatatransferEndpointSettingsClickhouseTargetShardingColumnValueHash(d)
		if err != nil {
			return nil, err
		}

		val.SetColumnValueHash(columnValueHash)
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.sharding.0.custom_mapping"); ok {
		customMapping, err := expandDatatransferEndpointSettingsClickhouseTargetShardingCustomMapping(d)
		if err != nil {
			return nil, err
		}

		val.SetCustomMapping(customMapping)
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.sharding.0.round_robin"); ok {
		roundRobin, err := expandDatatransferEndpointSettingsClickhouseTargetShardingRoundRobin(d)
		if err != nil {
			return nil, err
		}

		val.SetRoundRobin(roundRobin)
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.sharding.0.transfer_id"); ok {
		transferId, err := expandDatatransferEndpointSettingsClickhouseTargetShardingTransferId(d)
		if err != nil {
			return nil, err
		}

		val.SetTransferId(transferId)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetShardingTransferId(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetShardingRoundRobin(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetShardingCustomMapping(d *schema.ResourceData) (*endpoint.ClickhouseSharding_ColumnValueMapping, error) {
	val := new(endpoint.ClickhouseSharding_ColumnValueMapping)

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.sharding.0.custom_mapping.0.column_name"); ok {
		val.SetColumnName(v.(string))
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.sharding.0.custom_mapping.0.mapping"); ok {
		mapping, err := expandDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMappingSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetMapping(mapping)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMappingSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.ClickhouseSharding_ColumnValueMapping_ValueToShard, error) {
	count := d.Get("settings.0.clickhouse_target.0.sharding.0.custom_mapping.0.mapping.#").(int)
	slice := make([]*endpoint.ClickhouseSharding_ColumnValueMapping_ValueToShard, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMapping(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMapping(d *schema.ResourceData, indexes ...interface{}) (*endpoint.ClickhouseSharding_ColumnValueMapping_ValueToShard, error) {
	val := new(endpoint.ClickhouseSharding_ColumnValueMapping_ValueToShard)

	if _, ok := d.GetOk(fmt.Sprintf("settings.0.clickhouse_target.0.sharding.0.custom_mapping.0.mapping.%d.column_value", indexes...)); ok {
		columnValue, err := expandDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMappingColumnValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetColumnValue(columnValue)
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.clickhouse_target.0.sharding.0.custom_mapping.0.mapping.%d.shard_name", indexes...)); ok {
		val.SetShardName(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMappingColumnValue(d *schema.ResourceData, indexes ...interface{}) (*endpoint.ColumnValue, error) {
	val := new(endpoint.ColumnValue)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.clickhouse_target.0.sharding.0.custom_mapping.0.mapping.%d.column_value.0.string_value", indexes...)); ok {
		val.SetStringValue(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetShardingColumnValueHash(d *schema.ResourceData) (*endpoint.ClickhouseSharding_ColumnValueHash, error) {
	val := new(endpoint.ClickhouseSharding_ColumnValueHash)

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.sharding.0.column_value_hash.0.column_name"); ok {
		val.SetColumnName(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetConnection(d *schema.ResourceData) (*endpoint.ClickhouseConnection, error) {
	val := new(endpoint.ClickhouseConnection)

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options"); ok {
		connectionOptions, err := expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptions(d)
		if err != nil {
			return nil, err
		}

		val.SetConnectionOptions(connectionOptions)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptions(d *schema.ResourceData) (*endpoint.ClickhouseConnectionOptions, error) {
	val := new(endpoint.ClickhouseConnectionOptions)

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise"); ok {
		onPremise, err := expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.password"); ok {
		password, err := expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsPassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.user"); ok {
		val.SetUser(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsPassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremise(d *schema.ResourceData) (*endpoint.OnPremiseClickhouse, error) {
	val := new(endpoint.OnPremiseClickhouse)

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.http_port"); ok {
		val.SetHttpPort(int64(v.(int)))
	}

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.native_port"); ok {
		val.SetNativePort(int64(v.(int)))
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.shards"); ok {
		shards, err := expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseShardsSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetShards(shards)
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseShardsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.ClickhouseShard, error) {
	count := d.Get("settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.shards.#").(int)
	slice := make([]*endpoint.ClickhouseShard, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseShards(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseShards(d *schema.ResourceData, indexes ...interface{}) (*endpoint.ClickhouseShard, error) {
	val := new(endpoint.ClickhouseShard)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.shards.%d.hosts", indexes...)); ok {
		val.SetHosts(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.clickhouse_target.0.connection.0.connection_options.0.on_premise.0.shards.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetAltNamesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.AltName, error) {
	count := d.Get("settings.0.clickhouse_target.0.alt_names.#").(int)
	slice := make([]*endpoint.AltName, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsClickhouseTargetAltNames(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsClickhouseTargetAltNames(d *schema.ResourceData, indexes ...interface{}) (*endpoint.AltName, error) {
	val := new(endpoint.AltName)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.clickhouse_target.0.alt_names.%d.from_name", indexes...)); ok {
		val.SetFromName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.clickhouse_target.0.alt_names.%d.to_name", indexes...)); ok {
		val.SetToName(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseSource(d *schema.ResourceData) (*endpoint.ClickhouseSource, error) {
	val := new(endpoint.ClickhouseSource)

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.clickhouse_cluster_name"); ok {
		val.SetClickhouseClusterName(v.(string))
	}

	if _, ok := d.GetOk("settings.0.clickhouse_source.0.connection"); ok {
		connection, err := expandDatatransferEndpointSettingsClickhouseSourceConnection(d)
		if err != nil {
			return nil, err
		}

		val.SetConnection(connection)
	}

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.exclude_tables"); ok {
		val.SetExcludeTables(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.include_tables"); ok {
		val.SetIncludeTables(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.security_groups"); ok {
		val.SetSecurityGroups(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.subnet_id"); ok {
		val.SetSubnetId(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseSourceConnection(d *schema.ResourceData) (*endpoint.ClickhouseConnection, error) {
	val := new(endpoint.ClickhouseConnection)

	if _, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options"); ok {
		connectionOptions, err := expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptions(d)
		if err != nil {
			return nil, err
		}

		val.SetConnectionOptions(connectionOptions)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptions(d *schema.ResourceData) (*endpoint.ClickhouseConnectionOptions, error) {
	val := new(endpoint.ClickhouseConnectionOptions)

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.database"); ok {
		val.SetDatabase(v.(string))
	}

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.mdb_cluster_id"); ok {
		val.SetMdbClusterId(v.(string))
	}

	if _, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise"); ok {
		onPremise, err := expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremise(d)
		if err != nil {
			return nil, err
		}

		val.SetOnPremise(onPremise)
	}

	if _, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.password"); ok {
		password, err := expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsPassword(d)
		if err != nil {
			return nil, err
		}

		val.SetPassword(password)
	}

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.user"); ok {
		val.SetUser(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsPassword(d *schema.ResourceData) (*endpoint.Secret, error) {
	val := new(endpoint.Secret)

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.password.0.raw"); ok {
		val.SetRaw(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremise(d *schema.ResourceData) (*endpoint.OnPremiseClickhouse, error) {
	val := new(endpoint.OnPremiseClickhouse)

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.http_port"); ok {
		val.SetHttpPort(int64(v.(int)))
	}

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.native_port"); ok {
		val.SetNativePort(int64(v.(int)))
	}

	if _, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.shards"); ok {
		shards, err := expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseShardsSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetShards(shards)
	}

	if _, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.tls_mode"); ok {
		tlsMode, err := expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsMode(d)
		if err != nil {
			return nil, err
		}

		val.SetTlsMode(tlsMode)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsMode(d *schema.ResourceData) (*endpoint.TLSMode, error) {
	val := new(endpoint.TLSMode)

	if _, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.disabled"); ok {
		disabled, err := expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsModeDisabled(d)
		if err != nil {
			return nil, err
		}

		val.SetDisabled(disabled)
	}

	if _, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled"); ok {
		enabled, err := expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsModeEnabled(d)
		if err != nil {
			return nil, err
		}

		val.SetEnabled(enabled)
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsModeEnabled(d *schema.ResourceData) (*endpoint.TLSConfig, error) {
	val := new(endpoint.TLSConfig)

	if v, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.tls_mode.0.enabled.0.ca_certificate"); ok {
		val.SetCaCertificate(v.(string))
	}

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsModeDisabled(d *schema.ResourceData) (*emptypb.Empty, error) {
	val := new(emptypb.Empty)

	return val, nil
}

func expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseShardsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*endpoint.ClickhouseShard, error) {
	count := d.Get("settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.shards.#").(int)
	slice := make([]*endpoint.ClickhouseShard, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseShards(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseShards(d *schema.ResourceData, indexes ...interface{}) (*endpoint.ClickhouseShard, error) {
	val := new(endpoint.ClickhouseShard)

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.shards.%d.hosts", indexes...)); ok {
		val.SetHosts(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("settings.0.clickhouse_source.0.connection.0.connection_options.0.on_premise.0.shards.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	return val, nil
}

func expandDatatransferTransferRuntime(d *schema.ResourceData) (*datatransfer.Runtime, error) {
	val := new(datatransfer.Runtime)

	if _, ok := d.GetOk("runtime.0.yc_runtime"); ok {
		ycRuntime, err := expandDatatransferTransferRuntimeYcRuntime(d)
		if err != nil {
			return nil, err
		}

		val.SetYcRuntime(ycRuntime)
	}

	empty := new(datatransfer.Runtime)
	if proto.Equal(val, empty) {
		return nil, nil
	}

	return val, nil
}

func expandDatatransferTransferRuntimeYcRuntime(d *schema.ResourceData) (*datatransfer.YcRuntime, error) {
	val := new(datatransfer.YcRuntime)

	if v, ok := d.GetOk("runtime.0.yc_runtime.0.job_count"); ok {
		val.SetJobCount(int64(v.(int)))
	}

	if _, ok := d.GetOk("runtime.0.yc_runtime.0.upload_shard_params"); ok {
		uploadShardParams, err := expandDatatransferTransferRuntimeYcRuntimeUploadShardParams(d)
		if err != nil {
			return nil, err
		}

		val.SetUploadShardParams(uploadShardParams)
	}

	return val, nil
}

func expandDatatransferTransferRuntimeYcRuntimeUploadShardParams(d *schema.ResourceData) (*datatransfer.ShardingUploadParams, error) {
	val := new(datatransfer.ShardingUploadParams)

	if v, ok := d.GetOk("runtime.0.yc_runtime.0.upload_shard_params.0.job_count"); ok {
		val.SetJobCount(int64(v.(int)))
	}

	if v, ok := d.GetOk("runtime.0.yc_runtime.0.upload_shard_params.0.process_count"); ok {
		val.SetProcessCount(int64(v.(int)))
	}

	return val, nil
}

func expandDatatransferTransferTransformation(d *schema.ResourceData) (*datatransfer.Transformation, error) {
	val := new(datatransfer.Transformation)

	if _, ok := d.GetOk("transformation.0.transformers"); ok {
		transformers, err := expandDatatransferTransferTransformationTransformersSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetTransformers(transformers)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersSlice(d *schema.ResourceData, indexes ...interface{}) ([]*datatransfer.Transformer, error) {
	err := validateOneofDatatransferTransferTransformationTransformers(d)
	if err != nil {
		return nil, err
	}

	count := d.Get("transformation.0.transformers.#").(int)
	slice := make([]*datatransfer.Transformer, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferTransferTransformationTransformers(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferTransferTransformationTransformers(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.Transformer, error) {
	val := new(datatransfer.Transformer)

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.convert_to_string", indexes...)); ok {
		convertToString, err := expandDatatransferTransferTransformationTransformersConvertToString(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetConvertToString(convertToString)
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_columns", indexes...)); ok {
		filterColumns, err := expandDatatransferTransferTransformationTransformersFilterColumns(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetFilterColumns(filterColumns)
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_rows", indexes...)); ok {
		filterRows, err := expandDatatransferTransferTransformationTransformersFilterRows(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetFilterRows(filterRows)
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.mask_field", indexes...)); ok {
		maskField, err := expandDatatransferTransferTransformationTransformersMaskField(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetMaskField(maskField)
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.rename_tables", indexes...)); ok {
		renameTables, err := expandDatatransferTransferTransformationTransformersRenameTables(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRenameTables(renameTables)
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.replace_primary_key", indexes...)); ok {
		replacePrimaryKey, err := expandDatatransferTransferTransformationTransformersReplacePrimaryKey(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetReplacePrimaryKey(replacePrimaryKey)
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.sharder_transformer", indexes...)); ok {
		sharderTransformer, err := expandDatatransferTransferTransformationTransformersSharderTransformer(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSharderTransformer(sharderTransformer)
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.table_splitter_transformer", indexes...)); ok {
		tableSplitterTransformer, err := expandDatatransferTransferTransformationTransformersTableSplitterTransformer(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetTableSplitterTransformer(tableSplitterTransformer)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersTableSplitterTransformer(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.TableSplitterTransformer, error) {
	val := new(datatransfer.TableSplitterTransformer)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.table_splitter_transformer.0.columns", indexes...)); ok {
		val.SetColumns(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.table_splitter_transformer.0.splitter", indexes...)); ok {
		val.SetSplitter(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.table_splitter_transformer.0.tables", indexes...)); ok {
		tables, err := expandDatatransferTransferTransformationTransformersTableSplitterTransformerTables(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetTables(tables)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersTableSplitterTransformerTables(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.TablesFilter, error) {
	val := new(datatransfer.TablesFilter)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.table_splitter_transformer.0.tables.0.exclude_tables", indexes...)); ok {
		val.SetExcludeTables(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.table_splitter_transformer.0.tables.0.include_tables", indexes...)); ok {
		val.SetIncludeTables(expandStringSlice(v.([]interface{})))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersSharderTransformer(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.SharderTransformer, error) {
	val := new(datatransfer.SharderTransformer)

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.sharder_transformer.0.columns", indexes...)); ok {
		columns, err := expandDatatransferTransferTransformationTransformersSharderTransformerColumns(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetColumns(columns)
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.sharder_transformer.0.shards_count", indexes...)); ok {
		val.SetShardsCount(int64(v.(int)))
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.sharder_transformer.0.tables", indexes...)); ok {
		tables, err := expandDatatransferTransferTransformationTransformersSharderTransformerTables(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetTables(tables)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersSharderTransformerTables(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.TablesFilter, error) {
	val := new(datatransfer.TablesFilter)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.sharder_transformer.0.tables.0.exclude_tables", indexes...)); ok {
		val.SetExcludeTables(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.sharder_transformer.0.tables.0.include_tables", indexes...)); ok {
		val.SetIncludeTables(expandStringSlice(v.([]interface{})))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersSharderTransformerColumns(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.ColumnsFilter, error) {
	val := new(datatransfer.ColumnsFilter)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.sharder_transformer.0.columns.0.exclude_columns", indexes...)); ok {
		val.SetExcludeColumns(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.sharder_transformer.0.columns.0.include_columns", indexes...)); ok {
		val.SetIncludeColumns(expandStringSlice(v.([]interface{})))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersReplacePrimaryKey(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.ReplacePrimaryKeyTransformer, error) {
	val := new(datatransfer.ReplacePrimaryKeyTransformer)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.replace_primary_key.0.keys", indexes...)); ok {
		val.SetKeys(expandStringSlice(v.([]interface{})))
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.replace_primary_key.0.tables", indexes...)); ok {
		tables, err := expandDatatransferTransferTransformationTransformersReplacePrimaryKeyTables(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetTables(tables)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersReplacePrimaryKeyTables(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.TablesFilter, error) {
	val := new(datatransfer.TablesFilter)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.replace_primary_key.0.tables.0.exclude_tables", indexes...)); ok {
		val.SetExcludeTables(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.replace_primary_key.0.tables.0.include_tables", indexes...)); ok {
		val.SetIncludeTables(expandStringSlice(v.([]interface{})))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersRenameTables(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.RenameTablesTransformer, error) {
	val := new(datatransfer.RenameTablesTransformer)

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.rename_tables.0.rename_tables", indexes...)); ok {
		renameTables, err := expandDatatransferTransferTransformationTransformersRenameTablesRenameTablesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRenameTables(renameTables)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersRenameTablesRenameTablesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*datatransfer.RenameTable, error) {
	countPath := fmt.Sprintf("transformation.0.transformers.%d.rename_tables.0.rename_tables", indexes...)
	count := d.Get(fmt.Sprintf("%s.#", countPath)).(int)
	slice := make([]*datatransfer.RenameTable, count)

	for i := 0; i < count; i++ {
		indexes = append(indexes, i)
		expandedItem, err := expandDatatransferTransferTransformationTransformersRenameTablesRenameTables(d, indexes...)
		if err != nil {
			return nil, err
		}

		slice[i] = expandedItem
		indexes = indexes[:len(indexes)-1]
	}

	return slice, nil
}

func expandDatatransferTransferTransformationTransformersRenameTablesRenameTables(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.RenameTable, error) {
	val := new(datatransfer.RenameTable)

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.rename_tables.0.rename_tables.%d.new_name", indexes...)); ok {
		newName, err := expandDatatransferTransferTransformationTransformersRenameTablesRenameTablesNewName(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetNewName(newName)
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.rename_tables.0.rename_tables.%d.original_name", indexes...)); ok {
		originalName, err := expandDatatransferTransferTransformationTransformersRenameTablesRenameTablesOriginalName(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetOriginalName(originalName)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersRenameTablesRenameTablesOriginalName(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.Table, error) {
	val := new(datatransfer.Table)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.rename_tables.0.rename_tables.%d.original_name.0.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.rename_tables.0.rename_tables.%d.original_name.0.name_space", indexes...)); ok {
		val.SetNameSpace(v.(string))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersRenameTablesRenameTablesNewName(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.Table, error) {
	val := new(datatransfer.Table)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.rename_tables.0.rename_tables.%d.new_name.0.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.rename_tables.0.rename_tables.%d.new_name.0.name_space", indexes...)); ok {
		val.SetNameSpace(v.(string))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersMaskField(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.MaskFieldTransformer, error) {
	val := new(datatransfer.MaskFieldTransformer)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.mask_field.0.columns", indexes...)); ok {
		val.SetColumns(expandStringSlice(v.([]interface{})))
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.mask_field.0.function", indexes...)); ok {
		function, err := expandDatatransferTransferTransformationTransformersMaskFieldFunction(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetFunction(function)
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.mask_field.0.tables", indexes...)); ok {
		tables, err := expandDatatransferTransferTransformationTransformersMaskFieldTables(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetTables(tables)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersMaskFieldTables(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.TablesFilter, error) {
	val := new(datatransfer.TablesFilter)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.mask_field.0.tables.0.exclude_tables", indexes...)); ok {
		val.SetExcludeTables(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.mask_field.0.tables.0.include_tables", indexes...)); ok {
		val.SetIncludeTables(expandStringSlice(v.([]interface{})))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersMaskFieldFunction(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.MaskFunction, error) {
	val := new(datatransfer.MaskFunction)

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.mask_field.0.function.0.mask_function_hash", indexes...)); ok {
		maskFunctionHash, err := expandDatatransferTransferTransformationTransformersMaskFieldFunctionMaskFunctionHash(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetMaskFunctionHash(maskFunctionHash)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersMaskFieldFunctionMaskFunctionHash(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.MaskFunctionHash, error) {
	val := new(datatransfer.MaskFunctionHash)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.mask_field.0.function.0.mask_function_hash.0.user_defined_salt", indexes...)); ok {
		val.SetUserDefinedSalt(v.(string))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersFilterRows(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.FilterRowsTransformer, error) {
	val := new(datatransfer.FilterRowsTransformer)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_rows.0.filter", indexes...)); ok {
		val.SetFilter(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_rows.0.tables", indexes...)); ok {
		tables, err := expandDatatransferTransferTransformationTransformersFilterRowsTables(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetTables(tables)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersFilterRowsTables(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.TablesFilter, error) {
	val := new(datatransfer.TablesFilter)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_rows.0.tables.0.exclude_tables", indexes...)); ok {
		val.SetExcludeTables(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_rows.0.tables.0.include_tables", indexes...)); ok {
		val.SetIncludeTables(expandStringSlice(v.([]interface{})))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersFilterColumns(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.FilterColumnsTransformer, error) {
	val := new(datatransfer.FilterColumnsTransformer)

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_columns.0.columns", indexes...)); ok {
		columns, err := expandDatatransferTransferTransformationTransformersFilterColumnsColumns(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetColumns(columns)
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_columns.0.tables", indexes...)); ok {
		tables, err := expandDatatransferTransferTransformationTransformersFilterColumnsTables(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetTables(tables)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersFilterColumnsTables(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.TablesFilter, error) {
	val := new(datatransfer.TablesFilter)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_columns.0.tables.0.exclude_tables", indexes...)); ok {
		val.SetExcludeTables(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_columns.0.tables.0.include_tables", indexes...)); ok {
		val.SetIncludeTables(expandStringSlice(v.([]interface{})))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersFilterColumnsColumns(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.ColumnsFilter, error) {
	val := new(datatransfer.ColumnsFilter)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_columns.0.columns.0.exclude_columns", indexes...)); ok {
		val.SetExcludeColumns(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_columns.0.columns.0.include_columns", indexes...)); ok {
		val.SetIncludeColumns(expandStringSlice(v.([]interface{})))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersConvertToString(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.ToStringTransformer, error) {
	val := new(datatransfer.ToStringTransformer)

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.convert_to_string.0.columns", indexes...)); ok {
		columns, err := expandDatatransferTransferTransformationTransformersConvertToStringColumns(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetColumns(columns)
	}

	if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.convert_to_string.0.tables", indexes...)); ok {
		tables, err := expandDatatransferTransferTransformationTransformersConvertToStringTables(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetTables(tables)
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersConvertToStringTables(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.TablesFilter, error) {
	val := new(datatransfer.TablesFilter)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.convert_to_string.0.tables.0.exclude_tables", indexes...)); ok {
		val.SetExcludeTables(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.convert_to_string.0.tables.0.include_tables", indexes...)); ok {
		val.SetIncludeTables(expandStringSlice(v.([]interface{})))
	}

	return val, nil
}

func expandDatatransferTransferTransformationTransformersConvertToStringColumns(d *schema.ResourceData, indexes ...interface{}) (*datatransfer.ColumnsFilter, error) {
	val := new(datatransfer.ColumnsFilter)

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.convert_to_string.0.columns.0.exclude_columns", indexes...)); ok {
		val.SetExcludeColumns(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.convert_to_string.0.columns.0.include_columns", indexes...)); ok {
		val.SetIncludeColumns(expandStringSlice(v.([]interface{})))
	}

	return val, nil
}

func validateOneofDatatransferTransferTransformationTransformers(d *schema.ResourceData) error {
	var filledOneofs []string
	count := d.Get("transformation.0.transformers.#").(int)
	for i := 0; i < count; i++ {
		filledOneofs = []string{}
		indexes := []interface{}{i}
		if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.convert_to_string", indexes...)); ok {
			filledOneofs = append(filledOneofs, "convert_to_string")
		}
		if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_columns", indexes...)); ok {
			filledOneofs = append(filledOneofs, "filter_columns")
		}
		if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.filter_rows", indexes...)); ok {
			filledOneofs = append(filledOneofs, "filter_rows")
		}
		if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.mask_field", indexes...)); ok {
			filledOneofs = append(filledOneofs, "mask_field")
		}
		if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.rename_tables", indexes...)); ok {
			filledOneofs = append(filledOneofs, "rename_tables")
		}
		if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.replace_primary_key", indexes...)); ok {
			filledOneofs = append(filledOneofs, "replace_primary_key")
		}
		if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.sharder_transformer", indexes...)); ok {
			filledOneofs = append(filledOneofs, "sharder_transformer")
		}
		if _, ok := d.GetOk(fmt.Sprintf("transformation.0.transformers.%d.table_splitter_transformer", indexes...)); ok {
			filledOneofs = append(filledOneofs, "table_splitter_transformer")
		}
		if len(filledOneofs) != 1 {
			return fmt.Errorf("expected exactly 1 specified transformer in each element of transformers, got %d %v", len(filledOneofs), filledOneofs)
		}
	}
	return nil
}

func flattenDatatransferEndpointSettings(d *schema.ResourceData, v *datatransfer.EndpointSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	clickhouseSource, err := flattenDatatransferEndpointSettingsClickhouseSource(d, v.GetClickhouseSource())
	if err != nil {
		return nil, err
	}
	m["clickhouse_source"] = clickhouseSource

	clickhouseTarget, err := flattenDatatransferEndpointSettingsClickhouseTarget(d, v.GetClickhouseTarget())
	if err != nil {
		return nil, err
	}
	m["clickhouse_target"] = clickhouseTarget

	kafkaSource, err := flattenDatatransferEndpointSettingsKafkaSource(d, v.GetKafkaSource())
	if err != nil {
		return nil, err
	}
	m["kafka_source"] = kafkaSource

	kafkaTarget, err := flattenDatatransferEndpointSettingsKafkaTarget(d, v.GetKafkaTarget())
	if err != nil {
		return nil, err
	}
	m["kafka_target"] = kafkaTarget

	metrikaSource, err := flattenDatatransferEndpointSettingsMetrikaSource(d, v.GetMetrikaSource())
	if err != nil {
		return nil, err
	}
	m["metrika_source"] = metrikaSource

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

	ydbSource, err := flattenDatatransferEndpointSettingsYdbSource(d, v.GetYdbSource())
	if err != nil {
		return nil, err
	}
	m["ydb_source"] = ydbSource

	ydbTarget, err := flattenDatatransferEndpointSettingsYdbTarget(d, v.GetYdbTarget())
	if err != nil {
		return nil, err
	}
	m["ydb_target"] = ydbTarget

	ydsSource, err := flattenDatatransferEndpointSettingsYdsSource(d, v.GetYdsSource())
	if err != nil {
		return nil, err
	}
	m["yds_source"] = ydsSource

	ydsTarget, err := flattenDatatransferEndpointSettingsYdsTarget(d, v.GetYdsTarget())
	if err != nil {
		return nil, err
	}
	m["yds_target"] = ydsTarget

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsTarget(d *schema.ResourceData, v *endpoint.YDSTarget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["database"] = v.GetDatabase()
	m["endpoint"] = v.GetEndpoint()
	m["save_tx_order"] = v.GetSaveTxOrder()
	m["security_groups"] = v.GetSecurityGroups()

	serializer, err := flattenDatatransferEndpointSettingsYdsTargetSerializer(d, v.GetSerializer())
	if err != nil {
		return nil, err
	}
	m["serializer"] = serializer
	m["service_account_id"] = v.GetServiceAccountId()
	m["stream"] = v.GetStream()
	m["subnet_id"] = v.GetSubnetId()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsTargetSerializer(d *schema.ResourceData, v *endpoint.Serializer) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	serializerAuto, err := flattenDatatransferEndpointSettingsYdsTargetSerializerSerializerAuto(d, v.GetSerializerAuto())
	if err != nil {
		return nil, err
	}
	m["serializer_auto"] = serializerAuto

	serializerDebezium, err := flattenDatatransferEndpointSettingsYdsTargetSerializerSerializerDebezium(d, v.GetSerializerDebezium())
	if err != nil {
		return nil, err
	}
	m["serializer_debezium"] = serializerDebezium

	serializerJson, err := flattenDatatransferEndpointSettingsYdsTargetSerializerSerializerJson(d, v.GetSerializerJson())
	if err != nil {
		return nil, err
	}
	m["serializer_json"] = serializerJson

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsTargetSerializerSerializerJson(d *schema.ResourceData, v *endpoint.SerializerJSON) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsTargetSerializerSerializerDebezium(d *schema.ResourceData, v *endpoint.SerializerDebezium) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	serializerParameters, err := flattenDatatransferEndpointSettingsYdsTargetSerializerSerializerDebeziumSerializerParametersSlice(d, v.GetSerializerParameters())
	if err != nil {
		return nil, err
	}
	m["serializer_parameters"] = serializerParameters

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsTargetSerializerSerializerDebeziumSerializerParametersSlice(d *schema.ResourceData, v []*endpoint.DebeziumSerializerParameter) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsYdsTargetSerializerSerializerDebeziumSerializerParameters(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsYdsTargetSerializerSerializerDebeziumSerializerParameters(d *schema.ResourceData, v *endpoint.DebeziumSerializerParameter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["key"] = v.GetKey()
	m["value"] = v.GetValue()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsTargetSerializerSerializerAuto(d *schema.ResourceData, v *endpoint.SerializerAuto) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSource(d *schema.ResourceData, v *endpoint.YDSSource) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["allow_ttl_rewind"] = v.GetAllowTtlRewind()
	m["consumer"] = v.GetConsumer()
	m["database"] = v.GetDatabase()
	m["endpoint"] = v.GetEndpoint()

	parser, err := flattenDatatransferEndpointSettingsYdsSourceParser(d, v.GetParser())
	if err != nil {
		return nil, err
	}
	m["parser"] = parser
	m["security_groups"] = v.GetSecurityGroups()
	m["service_account_id"] = v.GetServiceAccountId()
	m["stream"] = v.GetStream()
	m["subnet_id"] = v.GetSubnetId()

	ydsCompressionCodec, err := flattenDatatransferEndpointSettingsYdsSourceSupportedCodecsSlice(v.GetSupportedCodecs())
	if err != nil {
		return nil, err
	}
	m["supported_codecs"] = ydsCompressionCodec

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSourceSupportedCodecsSlice(v []endpoint.YdsCompressionCodec) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		s = append(s, item.String())
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParser(d *schema.ResourceData, v *endpoint.Parser) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	auditTrailsV1Parser, err := flattenDatatransferEndpointSettingsYdsSourceParserAuditTrailsV1Parser(d, v.GetAuditTrailsV1Parser())
	if err != nil {
		return nil, err
	}
	m["audit_trails_v1_parser"] = auditTrailsV1Parser

	cloudLoggingParser, err := flattenDatatransferEndpointSettingsYdsSourceParserCloudLoggingParser(d, v.GetCloudLoggingParser())
	if err != nil {
		return nil, err
	}
	m["cloud_logging_parser"] = cloudLoggingParser

	jsonParser, err := flattenDatatransferEndpointSettingsYdsSourceParserJsonParser(d, v.GetJsonParser())
	if err != nil {
		return nil, err
	}
	m["json_parser"] = jsonParser

	tskvParser, err := flattenDatatransferEndpointSettingsYdsSourceParserTskvParser(d, v.GetTskvParser())
	if err != nil {
		return nil, err
	}
	m["tskv_parser"] = tskvParser

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserTskvParser(d *schema.ResourceData, v *endpoint.GenericParserCommon) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["add_rest_column"] = v.GetAddRestColumn()

	dataSchema, err := flattenDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchema(d, v.GetDataSchema())
	if err != nil {
		return nil, err
	}
	m["data_schema"] = dataSchema
	m["null_keys_allowed"] = v.GetNullKeysAllowed()
	m["unescape_string_values"] = v.GetUnescapeStringValues()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchema(d *schema.ResourceData, v *endpoint.DataSchema) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	fields, err := flattenDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFields(d, v.GetFields())
	if err != nil {
		return nil, err
	}
	m["fields"] = fields
	m["json_fields"] = v.GetJsonFields()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFields(d *schema.ResourceData, v *endpoint.FieldList) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	fields, err := flattenDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFieldsFieldsSlice(d, v.GetFields())
	if err != nil {
		return nil, err
	}
	m["fields"] = fields

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFieldsFieldsSlice(d *schema.ResourceData, v []*endpoint.ColSchema) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFieldsFields(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserTskvParserDataSchemaFieldsFields(d *schema.ResourceData, v *endpoint.ColSchema) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["key"] = v.GetKey()
	m["name"] = v.GetName()
	m["path"] = v.GetPath()
	m["required"] = v.GetRequired()
	m["type"] = v.GetType().String()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserJsonParser(d *schema.ResourceData, v *endpoint.GenericParserCommon) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["add_rest_column"] = v.GetAddRestColumn()

	dataSchema, err := flattenDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchema(d, v.GetDataSchema())
	if err != nil {
		return nil, err
	}
	m["data_schema"] = dataSchema
	m["null_keys_allowed"] = v.GetNullKeysAllowed()
	m["unescape_string_values"] = v.GetUnescapeStringValues()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchema(d *schema.ResourceData, v *endpoint.DataSchema) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	fields, err := flattenDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFields(d, v.GetFields())
	if err != nil {
		return nil, err
	}
	m["fields"] = fields
	m["json_fields"] = v.GetJsonFields()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFields(d *schema.ResourceData, v *endpoint.FieldList) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	fields, err := flattenDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFieldsFieldsSlice(d, v.GetFields())
	if err != nil {
		return nil, err
	}
	m["fields"] = fields

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFieldsFieldsSlice(d *schema.ResourceData, v []*endpoint.ColSchema) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFieldsFields(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserJsonParserDataSchemaFieldsFields(d *schema.ResourceData, v *endpoint.ColSchema) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["key"] = v.GetKey()
	m["name"] = v.GetName()
	m["path"] = v.GetPath()
	m["required"] = v.GetRequired()
	m["type"] = v.GetType().String()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserCloudLoggingParser(d *schema.ResourceData, v *endpoint.CloudLoggingParser) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdsSourceParserAuditTrailsV1Parser(d *schema.ResourceData, v *endpoint.AuditTrailsV1Parser) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdbTarget(d *schema.ResourceData, v *endpoint.YdbTarget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["cleanup_policy"] = v.GetCleanupPolicy().String()
	m["database"] = v.GetDatabase()
	m["default_compression"] = v.GetDefaultCompression().String()
	m["instance"] = v.GetInstance()
	m["is_table_column_oriented"] = v.GetIsTableColumnOriented()
	m["path"] = v.GetPath()
	m["sa_key_content"] = v.GetSaKeyContent()
	m["security_groups"] = v.GetSecurityGroups()
	m["service_account_id"] = v.GetServiceAccountId()
	m["subnet_id"] = v.GetSubnetId()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsYdbSource(d *schema.ResourceData, v *endpoint.YdbSource) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["changefeed_custom_name"] = v.GetChangefeedCustomName()
	m["database"] = v.GetDatabase()
	m["instance"] = v.GetInstance()
	m["paths"] = v.GetPaths()
	m["sa_key_content"] = v.GetSaKeyContent()
	m["security_groups"] = v.GetSecurityGroups()
	m["service_account_id"] = v.GetServiceAccountId()
	m["subnet_id"] = v.GetSubnetId()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsPostgresTarget(d *schema.ResourceData, v *endpoint.PostgresTarget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["cleanup_policy"] = v.GetCleanupPolicy().String()

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
	m["sequence_set"] = v.GetSequenceSet().String()
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

	m["cleanup_policy"] = v.GetCleanupPolicy().String()

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
	m["service_database"] = v.GetServiceDatabase()
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
	m["tables"] = v.GetTables().String()
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

func flattenDatatransferEndpointSettingsMetrikaSource(d *schema.ResourceData, v *endpoint.MetrikaSource) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["counter_ids"] = v.GetCounterIds()

	streams, err := flattenDatatransferEndpointSettingsMetrikaSourceStreamsSlice(d, v.GetStreams())
	if err != nil {
		return nil, err
	}
	m["streams"] = streams
	if token, ok := d.GetOk("settings.0.metrika_source.0.token.0.raw"); ok {
		m["token"] = []map[string]interface{}{{"raw": token}}
	}

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsMetrikaSourceStreamsSlice(d *schema.ResourceData, v []*endpoint.MetrikaStream) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsMetrikaSourceStreams(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsMetrikaSourceStreams(d *schema.ResourceData, v *endpoint.MetrikaStream) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["columns"] = v.GetColumns()
	m["type"] = v.GetType().String()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTarget(d *schema.ResourceData, v *endpoint.KafkaTarget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	auth, err := flattenDatatransferEndpointSettingsKafkaTargetAuth(d, v.GetAuth())
	if err != nil {
		return nil, err
	}
	m["auth"] = auth

	connection, err := flattenDatatransferEndpointSettingsKafkaTargetConnection(d, v.GetConnection())
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["security_groups"] = v.GetSecurityGroups()

	serializer, err := flattenDatatransferEndpointSettingsKafkaTargetSerializer(d, v.GetSerializer())
	if err != nil {
		return nil, err
	}
	m["serializer"] = serializer

	topicSettings, err := flattenDatatransferEndpointSettingsKafkaTargetTopicSettings(d, v.GetTopicSettings())
	if err != nil {
		return nil, err
	}
	m["topic_settings"] = topicSettings

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetTopicSettings(d *schema.ResourceData, v *endpoint.KafkaTargetTopicSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	topic, err := flattenDatatransferEndpointSettingsKafkaTargetTopicSettingsTopic(d, v.GetTopic())
	if err != nil {
		return nil, err
	}
	m["topic"] = topic
	m["topic_prefix"] = v.GetTopicPrefix()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetTopicSettingsTopic(d *schema.ResourceData, v *endpoint.KafkaTargetTopic) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["save_tx_order"] = v.GetSaveTxOrder()
	m["topic_name"] = v.GetTopicName()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetSerializer(d *schema.ResourceData, v *endpoint.Serializer) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	serializerAuto, err := flattenDatatransferEndpointSettingsKafkaTargetSerializerSerializerAuto(d, v.GetSerializerAuto())
	if err != nil {
		return nil, err
	}
	m["serializer_auto"] = serializerAuto

	serializerDebezium, err := flattenDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebezium(d, v.GetSerializerDebezium())
	if err != nil {
		return nil, err
	}
	m["serializer_debezium"] = serializerDebezium

	serializerJson, err := flattenDatatransferEndpointSettingsKafkaTargetSerializerSerializerJson(d, v.GetSerializerJson())
	if err != nil {
		return nil, err
	}
	m["serializer_json"] = serializerJson

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetSerializerSerializerJson(d *schema.ResourceData, v *endpoint.SerializerJSON) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebezium(d *schema.ResourceData, v *endpoint.SerializerDebezium) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	serializerParameters, err := flattenDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebeziumSerializerParametersSlice(d, v.GetSerializerParameters())
	if err != nil {
		return nil, err
	}
	m["serializer_parameters"] = serializerParameters

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebeziumSerializerParametersSlice(d *schema.ResourceData, v []*endpoint.DebeziumSerializerParameter) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebeziumSerializerParameters(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetSerializerSerializerDebeziumSerializerParameters(d *schema.ResourceData, v *endpoint.DebeziumSerializerParameter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["key"] = v.GetKey()
	m["value"] = v.GetValue()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetSerializerSerializerAuto(d *schema.ResourceData, v *endpoint.SerializerAuto) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetConnection(d *schema.ResourceData, v *endpoint.KafkaConnectionOptions) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["cluster_id"] = v.GetClusterId()

	onPremise, err := flattenDatatransferEndpointSettingsKafkaTargetConnectionOnPremise(d, v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetConnectionOnPremise(d *schema.ResourceData, v *endpoint.OnPremiseKafka) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["broker_urls"] = v.GetBrokerUrls()
	m["subnet_id"] = v.GetSubnetId()

	tlsMode, err := flattenDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsMode(d, v.GetTlsMode())
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsMode(d *schema.ResourceData, v *endpoint.TLSMode) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	disabled, err := flattenDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsModeDisabled(d, v.GetDisabled())
	if err != nil {
		return nil, err
	}
	m["disabled"] = disabled

	enabled, err := flattenDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsModeEnabled(d, v.GetEnabled())
	if err != nil {
		return nil, err
	}
	m["enabled"] = enabled

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData, v *endpoint.TLSConfig) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ca_certificate"] = v.GetCaCertificate()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetAuth(d *schema.ResourceData, v *endpoint.KafkaAuth) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	noAuth, err := flattenDatatransferEndpointSettingsKafkaTargetAuthNoAuth(d, v.GetNoAuth())
	if err != nil {
		return nil, err
	}
	m["no_auth"] = noAuth

	sasl, err := flattenDatatransferEndpointSettingsKafkaTargetAuthSasl(d, v.GetSasl())
	if err != nil {
		return nil, err
	}
	m["sasl"] = sasl

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetAuthSasl(d *schema.ResourceData, v *endpoint.KafkaSaslSecurity) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["mechanism"] = v.GetMechanism().String()
	if password, ok := d.GetOk("settings.0.kafka_target.0.auth.0.sasl.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["user"] = v.GetUser()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaTargetAuthNoAuth(d *schema.ResourceData, v *endpoint.NoAuth) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSource(d *schema.ResourceData, v *endpoint.KafkaSource) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	auth, err := flattenDatatransferEndpointSettingsKafkaSourceAuth(d, v.GetAuth())
	if err != nil {
		return nil, err
	}
	m["auth"] = auth

	connection, err := flattenDatatransferEndpointSettingsKafkaSourceConnection(d, v.GetConnection())
	if err != nil {
		return nil, err
	}
	m["connection"] = connection

	parser, err := flattenDatatransferEndpointSettingsKafkaSourceParser(d, v.GetParser())
	if err != nil {
		return nil, err
	}
	m["parser"] = parser
	m["security_groups"] = v.GetSecurityGroups()
	m["topic_name"] = v.GetTopicName()
	m["topic_names"] = v.GetTopicNames()

	transformer, err := flattenDatatransferEndpointSettingsKafkaSourceTransformer(d, v.GetTransformer())
	if err != nil {
		return nil, err
	}
	m["transformer"] = transformer

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceTransformer(d *schema.ResourceData, v *endpoint.DataTransformationOptions) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["buffer_flush_interval"] = v.GetBufferFlushInterval()
	m["buffer_size"] = v.GetBufferSize()
	m["cloud_function"] = v.GetCloudFunction()
	m["invocation_timeout"] = v.GetInvocationTimeout()
	m["number_of_retries"] = v.GetNumberOfRetries()
	m["service_account_id"] = v.GetServiceAccountId()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParser(d *schema.ResourceData, v *endpoint.Parser) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	auditTrailsV1Parser, err := flattenDatatransferEndpointSettingsKafkaSourceParserAuditTrailsV1Parser(d, v.GetAuditTrailsV1Parser())
	if err != nil {
		return nil, err
	}
	m["audit_trails_v1_parser"] = auditTrailsV1Parser

	cloudLoggingParser, err := flattenDatatransferEndpointSettingsKafkaSourceParserCloudLoggingParser(d, v.GetCloudLoggingParser())
	if err != nil {
		return nil, err
	}
	m["cloud_logging_parser"] = cloudLoggingParser

	jsonParser, err := flattenDatatransferEndpointSettingsKafkaSourceParserJsonParser(d, v.GetJsonParser())
	if err != nil {
		return nil, err
	}
	m["json_parser"] = jsonParser

	tskvParser, err := flattenDatatransferEndpointSettingsKafkaSourceParserTskvParser(d, v.GetTskvParser())
	if err != nil {
		return nil, err
	}
	m["tskv_parser"] = tskvParser

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserTskvParser(d *schema.ResourceData, v *endpoint.GenericParserCommon) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["add_rest_column"] = v.GetAddRestColumn()

	dataSchema, err := flattenDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchema(d, v.GetDataSchema())
	if err != nil {
		return nil, err
	}
	m["data_schema"] = dataSchema
	m["null_keys_allowed"] = v.GetNullKeysAllowed()
	m["unescape_string_values"] = v.GetUnescapeStringValues()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchema(d *schema.ResourceData, v *endpoint.DataSchema) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	fields, err := flattenDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFields(d, v.GetFields())
	if err != nil {
		return nil, err
	}
	m["fields"] = fields
	m["json_fields"] = v.GetJsonFields()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFields(d *schema.ResourceData, v *endpoint.FieldList) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	fields, err := flattenDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFieldsFieldsSlice(d, v.GetFields())
	if err != nil {
		return nil, err
	}
	m["fields"] = fields

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFieldsFieldsSlice(d *schema.ResourceData, v []*endpoint.ColSchema) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFieldsFields(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserTskvParserDataSchemaFieldsFields(d *schema.ResourceData, v *endpoint.ColSchema) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["key"] = v.GetKey()
	m["name"] = v.GetName()
	m["path"] = v.GetPath()
	m["required"] = v.GetRequired()
	m["type"] = v.GetType().String()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserJsonParser(d *schema.ResourceData, v *endpoint.GenericParserCommon) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["add_rest_column"] = v.GetAddRestColumn()

	dataSchema, err := flattenDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchema(d, v.GetDataSchema())
	if err != nil {
		return nil, err
	}
	m["data_schema"] = dataSchema
	m["null_keys_allowed"] = v.GetNullKeysAllowed()
	m["unescape_string_values"] = v.GetUnescapeStringValues()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchema(d *schema.ResourceData, v *endpoint.DataSchema) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	fields, err := flattenDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFields(d, v.GetFields())
	if err != nil {
		return nil, err
	}
	m["fields"] = fields
	m["json_fields"] = v.GetJsonFields()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFields(d *schema.ResourceData, v *endpoint.FieldList) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	fields, err := flattenDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFieldsFieldsSlice(d, v.GetFields())
	if err != nil {
		return nil, err
	}
	m["fields"] = fields

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFieldsFieldsSlice(d *schema.ResourceData, v []*endpoint.ColSchema) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFieldsFields(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserJsonParserDataSchemaFieldsFields(d *schema.ResourceData, v *endpoint.ColSchema) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["key"] = v.GetKey()
	m["name"] = v.GetName()
	m["path"] = v.GetPath()
	m["required"] = v.GetRequired()
	m["type"] = v.GetType().String()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserCloudLoggingParser(d *schema.ResourceData, v *endpoint.CloudLoggingParser) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceParserAuditTrailsV1Parser(d *schema.ResourceData, v *endpoint.AuditTrailsV1Parser) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceConnection(d *schema.ResourceData, v *endpoint.KafkaConnectionOptions) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["cluster_id"] = v.GetClusterId()

	onPremise, err := flattenDatatransferEndpointSettingsKafkaSourceConnectionOnPremise(d, v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceConnectionOnPremise(d *schema.ResourceData, v *endpoint.OnPremiseKafka) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["broker_urls"] = v.GetBrokerUrls()
	m["subnet_id"] = v.GetSubnetId()

	tlsMode, err := flattenDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsMode(d, v.GetTlsMode())
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsMode(d *schema.ResourceData, v *endpoint.TLSMode) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	disabled, err := flattenDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsModeDisabled(d, v.GetDisabled())
	if err != nil {
		return nil, err
	}
	m["disabled"] = disabled

	enabled, err := flattenDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsModeEnabled(d, v.GetEnabled())
	if err != nil {
		return nil, err
	}
	m["enabled"] = enabled

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsModeEnabled(d *schema.ResourceData, v *endpoint.TLSConfig) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ca_certificate"] = v.GetCaCertificate()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceConnectionOnPremiseTlsModeDisabled(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceAuth(d *schema.ResourceData, v *endpoint.KafkaAuth) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	noAuth, err := flattenDatatransferEndpointSettingsKafkaSourceAuthNoAuth(d, v.GetNoAuth())
	if err != nil {
		return nil, err
	}
	m["no_auth"] = noAuth

	sasl, err := flattenDatatransferEndpointSettingsKafkaSourceAuthSasl(d, v.GetSasl())
	if err != nil {
		return nil, err
	}
	m["sasl"] = sasl

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceAuthSasl(d *schema.ResourceData, v *endpoint.KafkaSaslSecurity) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["mechanism"] = v.GetMechanism().String()
	if password, ok := d.GetOk("settings.0.kafka_source.0.auth.0.sasl.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["user"] = v.GetUser()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsKafkaSourceAuthNoAuth(d *schema.ResourceData, v *endpoint.NoAuth) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTarget(d *schema.ResourceData, v *endpoint.ClickhouseTarget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	altNames, err := flattenDatatransferEndpointSettingsClickhouseTargetAltNamesSlice(d, v.GetAltNames())
	if err != nil {
		return nil, err
	}
	m["alt_names"] = altNames
	m["cleanup_policy"] = v.GetCleanupPolicy().String()
	m["clickhouse_cluster_name"] = v.GetClickhouseClusterName()

	connection, err := flattenDatatransferEndpointSettingsClickhouseTargetConnection(d, v.GetConnection())
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["security_groups"] = v.GetSecurityGroups()

	sharding, err := flattenDatatransferEndpointSettingsClickhouseTargetSharding(d, v.GetSharding())
	if err != nil {
		return nil, err
	}
	m["sharding"] = sharding
	m["subnet_id"] = v.GetSubnetId()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetSharding(d *schema.ResourceData, v *endpoint.ClickhouseSharding) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	columnValueHash, err := flattenDatatransferEndpointSettingsClickhouseTargetShardingColumnValueHash(d, v.GetColumnValueHash())
	if err != nil {
		return nil, err
	}
	m["column_value_hash"] = columnValueHash

	customMapping, err := flattenDatatransferEndpointSettingsClickhouseTargetShardingCustomMapping(d, v.GetCustomMapping())
	if err != nil {
		return nil, err
	}
	m["custom_mapping"] = customMapping

	roundRobin, err := flattenDatatransferEndpointSettingsClickhouseTargetShardingRoundRobin(d, v.GetRoundRobin())
	if err != nil {
		return nil, err
	}
	m["round_robin"] = roundRobin

	transferId, err := flattenDatatransferEndpointSettingsClickhouseTargetShardingTransferId(d, v.GetTransferId())
	if err != nil {
		return nil, err
	}
	m["transfer_id"] = transferId

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetShardingTransferId(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetShardingRoundRobin(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetShardingCustomMapping(d *schema.ResourceData, v *endpoint.ClickhouseSharding_ColumnValueMapping) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["column_name"] = v.GetColumnName()

	mapping, err := flattenDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMappingSlice(d, v.GetMapping())
	if err != nil {
		return nil, err
	}
	m["mapping"] = mapping

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMappingSlice(d *schema.ResourceData, v []*endpoint.ClickhouseSharding_ColumnValueMapping_ValueToShard) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMapping(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMapping(d *schema.ResourceData, v *endpoint.ClickhouseSharding_ColumnValueMapping_ValueToShard) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	columnValue, err := flattenDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMappingColumnValue(d, v.GetColumnValue())
	if err != nil {
		return nil, err
	}
	m["column_value"] = columnValue
	m["shard_name"] = v.GetShardName()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetShardingCustomMappingMappingColumnValue(d *schema.ResourceData, v *endpoint.ColumnValue) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["string_value"] = v.GetStringValue()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetShardingColumnValueHash(d *schema.ResourceData, v *endpoint.ClickhouseSharding_ColumnValueHash) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["column_name"] = v.GetColumnName()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetConnection(d *schema.ResourceData, v *endpoint.ClickhouseConnection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connectionOptions, err := flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptions(d, v.GetConnectionOptions())
	if err != nil {
		return nil, err
	}
	m["connection_options"] = connectionOptions

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptions(d *schema.ResourceData, v *endpoint.ClickhouseConnectionOptions) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["database"] = v.GetDatabase()
	m["mdb_cluster_id"] = v.GetMdbClusterId()

	onPremise, err := flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremise(d, v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise
	if password, ok := d.GetOk("settings.0.clickhouse_target.0.connection.0.connection_options.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["user"] = v.GetUser()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremise(d *schema.ResourceData, v *endpoint.OnPremiseClickhouse) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["http_port"] = v.GetHttpPort()
	m["native_port"] = v.GetNativePort()

	shards, err := flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseShardsSlice(d, v.GetShards())
	if err != nil {
		return nil, err
	}
	m["shards"] = shards

	tlsMode, err := flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsMode(d, v.GetTlsMode())
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsMode(d *schema.ResourceData, v *endpoint.TLSMode) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	disabled, err := flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsModeDisabled(d, v.GetDisabled())
	if err != nil {
		return nil, err
	}
	m["disabled"] = disabled

	enabled, err := flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsModeEnabled(d, v.GetEnabled())
	if err != nil {
		return nil, err
	}
	m["enabled"] = enabled

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsModeEnabled(d *schema.ResourceData, v *endpoint.TLSConfig) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ca_certificate"] = v.GetCaCertificate()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseTlsModeDisabled(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseShardsSlice(d *schema.ResourceData, v []*endpoint.ClickhouseShard) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseShards(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetConnectionConnectionOptionsOnPremiseShards(d *schema.ResourceData, v *endpoint.ClickhouseShard) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["hosts"] = v.GetHosts()
	m["name"] = v.GetName()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetAltNamesSlice(d *schema.ResourceData, v []*endpoint.AltName) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsClickhouseTargetAltNames(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsClickhouseTargetAltNames(d *schema.ResourceData, v *endpoint.AltName) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["from_name"] = v.GetFromName()
	m["to_name"] = v.GetToName()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseSource(d *schema.ResourceData, v *endpoint.ClickhouseSource) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["clickhouse_cluster_name"] = v.GetClickhouseClusterName()

	connection, err := flattenDatatransferEndpointSettingsClickhouseSourceConnection(d, v.GetConnection())
	if err != nil {
		return nil, err
	}
	m["connection"] = connection
	m["exclude_tables"] = v.GetExcludeTables()
	m["include_tables"] = v.GetIncludeTables()
	m["security_groups"] = v.GetSecurityGroups()
	m["subnet_id"] = v.GetSubnetId()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseSourceConnection(d *schema.ResourceData, v *endpoint.ClickhouseConnection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	connectionOptions, err := flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptions(d, v.GetConnectionOptions())
	if err != nil {
		return nil, err
	}
	m["connection_options"] = connectionOptions

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptions(d *schema.ResourceData, v *endpoint.ClickhouseConnectionOptions) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["database"] = v.GetDatabase()
	m["mdb_cluster_id"] = v.GetMdbClusterId()

	onPremise, err := flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremise(d, v.GetOnPremise())
	if err != nil {
		return nil, err
	}
	m["on_premise"] = onPremise
	if password, ok := d.GetOk("settings.0.clickhouse_source.0.connection.0.connection_options.0.password.0.raw"); ok {
		m["password"] = []map[string]interface{}{{"raw": password}}
	}
	m["user"] = v.GetUser()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremise(d *schema.ResourceData, v *endpoint.OnPremiseClickhouse) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["http_port"] = v.GetHttpPort()
	m["native_port"] = v.GetNativePort()

	shards, err := flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseShardsSlice(d, v.GetShards())
	if err != nil {
		return nil, err
	}
	m["shards"] = shards

	tlsMode, err := flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsMode(d, v.GetTlsMode())
	if err != nil {
		return nil, err
	}
	m["tls_mode"] = tlsMode

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsMode(d *schema.ResourceData, v *endpoint.TLSMode) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	disabled, err := flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsModeDisabled(d, v.GetDisabled())
	if err != nil {
		return nil, err
	}
	m["disabled"] = disabled

	enabled, err := flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsModeEnabled(d, v.GetEnabled())
	if err != nil {
		return nil, err
	}
	m["enabled"] = enabled

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsModeEnabled(d *schema.ResourceData, v *endpoint.TLSConfig) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ca_certificate"] = v.GetCaCertificate()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseTlsModeDisabled(d *schema.ResourceData, v *emptypb.Empty) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseShardsSlice(d *schema.ResourceData, v []*endpoint.ClickhouseShard) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseShards(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferEndpointSettingsClickhouseSourceConnectionConnectionOptionsOnPremiseShards(d *schema.ResourceData, v *endpoint.ClickhouseShard) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["hosts"] = v.GetHosts()
	m["name"] = v.GetName()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferRuntime(d *schema.ResourceData, v *datatransfer.Runtime) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	ycRuntime, err := flattenDatatransferTransferRuntimeYcRuntime(d, v.GetYcRuntime())
	if err != nil {
		return nil, err
	}
	m["yc_runtime"] = ycRuntime

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferRuntimeYcRuntime(d *schema.ResourceData, v *datatransfer.YcRuntime) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["job_count"] = v.GetJobCount()

	uploadShardParams, err := flattenDatatransferTransferRuntimeYcRuntimeUploadShardParams(d, v.GetUploadShardParams())
	if err != nil {
		return nil, err
	}
	m["upload_shard_params"] = uploadShardParams

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferRuntimeYcRuntimeUploadShardParams(d *schema.ResourceData, v *datatransfer.ShardingUploadParams) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["job_count"] = v.GetJobCount()
	m["process_count"] = v.GetProcessCount()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformation(d *schema.ResourceData, v *datatransfer.Transformation) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	transformers, err := flattenDatatransferTransferTransformationTransformersSlice(d, v.GetTransformers())
	if err != nil {
		return nil, err
	}
	m["transformers"] = transformers

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersSlice(d *schema.ResourceData, v []*datatransfer.Transformer) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferTransferTransformationTransformers(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferTransferTransformationTransformers(d *schema.ResourceData, v *datatransfer.Transformer) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	convertToString, err := flattenDatatransferTransferTransformationTransformersConvertToString(d, v.GetConvertToString())
	if err != nil {
		return nil, err
	}
	m["convert_to_string"] = convertToString

	filterColumns, err := flattenDatatransferTransferTransformationTransformersFilterColumns(d, v.GetFilterColumns())
	if err != nil {
		return nil, err
	}
	m["filter_columns"] = filterColumns

	filterRows, err := flattenDatatransferTransferTransformationTransformersFilterRows(d, v.GetFilterRows())
	if err != nil {
		return nil, err
	}
	m["filter_rows"] = filterRows

	maskField, err := flattenDatatransferTransferTransformationTransformersMaskField(d, v.GetMaskField())
	if err != nil {
		return nil, err
	}
	m["mask_field"] = maskField

	renameTables, err := flattenDatatransferTransferTransformationTransformersRenameTables(d, v.GetRenameTables())
	if err != nil {
		return nil, err
	}
	m["rename_tables"] = renameTables

	replacePrimaryKey, err := flattenDatatransferTransferTransformationTransformersReplacePrimaryKey(d, v.GetReplacePrimaryKey())
	if err != nil {
		return nil, err
	}
	m["replace_primary_key"] = replacePrimaryKey

	sharderTransformer, err := flattenDatatransferTransferTransformationTransformersSharderTransformer(d, v.GetSharderTransformer())
	if err != nil {
		return nil, err
	}
	m["sharder_transformer"] = sharderTransformer

	tableSplitterTransformer, err := flattenDatatransferTransferTransformationTransformersTableSplitterTransformer(d, v.GetTableSplitterTransformer())
	if err != nil {
		return nil, err
	}
	m["table_splitter_transformer"] = tableSplitterTransformer

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersTableSplitterTransformer(d *schema.ResourceData, v *datatransfer.TableSplitterTransformer) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["columns"] = v.GetColumns()
	m["splitter"] = v.GetSplitter()

	tables, err := flattenDatatransferTransferTransformationTransformersTableSplitterTransformerTables(d, v.GetTables())
	if err != nil {
		return nil, err
	}
	m["tables"] = tables

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersTableSplitterTransformerTables(d *schema.ResourceData, v *datatransfer.TablesFilter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exclude_tables"] = v.GetExcludeTables()
	m["include_tables"] = v.GetIncludeTables()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersSharderTransformer(d *schema.ResourceData, v *datatransfer.SharderTransformer) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	columns, err := flattenDatatransferTransferTransformationTransformersSharderTransformerColumns(d, v.GetColumns())
	if err != nil {
		return nil, err
	}
	m["columns"] = columns
	m["shards_count"] = v.GetShardsCount()

	tables, err := flattenDatatransferTransferTransformationTransformersSharderTransformerTables(d, v.GetTables())
	if err != nil {
		return nil, err
	}
	m["tables"] = tables

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersSharderTransformerTables(d *schema.ResourceData, v *datatransfer.TablesFilter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exclude_tables"] = v.GetExcludeTables()
	m["include_tables"] = v.GetIncludeTables()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersSharderTransformerColumns(d *schema.ResourceData, v *datatransfer.ColumnsFilter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exclude_columns"] = v.GetExcludeColumns()
	m["include_columns"] = v.GetIncludeColumns()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersReplacePrimaryKey(d *schema.ResourceData, v *datatransfer.ReplacePrimaryKeyTransformer) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["keys"] = v.GetKeys()

	tables, err := flattenDatatransferTransferTransformationTransformersReplacePrimaryKeyTables(d, v.GetTables())
	if err != nil {
		return nil, err
	}
	m["tables"] = tables

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersReplacePrimaryKeyTables(d *schema.ResourceData, v *datatransfer.TablesFilter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exclude_tables"] = v.GetExcludeTables()
	m["include_tables"] = v.GetIncludeTables()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersRenameTables(d *schema.ResourceData, v *datatransfer.RenameTablesTransformer) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	renameTables, err := flattenDatatransferTransferTransformationTransformersRenameTablesRenameTablesSlice(d, v.GetRenameTables())
	if err != nil {
		return nil, err
	}
	m["rename_tables"] = renameTables

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersRenameTablesRenameTablesSlice(d *schema.ResourceData, v []*datatransfer.RenameTable) ([]interface{}, error) {
	s := make([]interface{}, 0, len(v))

	for _, item := range v {
		flattenedItem, err := flattenDatatransferTransferTransformationTransformersRenameTablesRenameTables(d, item)
		if err != nil {
			return nil, err
		}

		if len(flattenedItem) != 0 {
			s = append(s, flattenedItem[0])
		}
	}

	return s, nil
}

func flattenDatatransferTransferTransformationTransformersRenameTablesRenameTables(d *schema.ResourceData, v *datatransfer.RenameTable) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	newName, err := flattenDatatransferTransferTransformationTransformersRenameTablesRenameTablesNewName(d, v.GetNewName())
	if err != nil {
		return nil, err
	}
	m["new_name"] = newName

	originalName, err := flattenDatatransferTransferTransformationTransformersRenameTablesRenameTablesOriginalName(d, v.GetOriginalName())
	if err != nil {
		return nil, err
	}
	m["original_name"] = originalName

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersRenameTablesRenameTablesOriginalName(d *schema.ResourceData, v *datatransfer.Table) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["name"] = v.GetName()
	m["name_space"] = v.GetNameSpace()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersRenameTablesRenameTablesNewName(d *schema.ResourceData, v *datatransfer.Table) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["name"] = v.GetName()
	m["name_space"] = v.GetNameSpace()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersMaskField(d *schema.ResourceData, v *datatransfer.MaskFieldTransformer) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["columns"] = v.GetColumns()

	function, err := flattenDatatransferTransferTransformationTransformersMaskFieldFunction(d, v.GetFunction())
	if err != nil {
		return nil, err
	}
	m["function"] = function

	tables, err := flattenDatatransferTransferTransformationTransformersMaskFieldTables(d, v.GetTables())
	if err != nil {
		return nil, err
	}
	m["tables"] = tables

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersMaskFieldTables(d *schema.ResourceData, v *datatransfer.TablesFilter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exclude_tables"] = v.GetExcludeTables()
	m["include_tables"] = v.GetIncludeTables()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersMaskFieldFunction(d *schema.ResourceData, v *datatransfer.MaskFunction) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	maskFunctionHash, err := flattenDatatransferTransferTransformationTransformersMaskFieldFunctionMaskFunctionHash(d, v.GetMaskFunctionHash())
	if err != nil {
		return nil, err
	}
	m["mask_function_hash"] = maskFunctionHash

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersMaskFieldFunctionMaskFunctionHash(d *schema.ResourceData, v *datatransfer.MaskFunctionHash) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["user_defined_salt"] = v.GetUserDefinedSalt()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersFilterRows(d *schema.ResourceData, v *datatransfer.FilterRowsTransformer) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["filter"] = v.GetFilter()

	tables, err := flattenDatatransferTransferTransformationTransformersFilterRowsTables(d, v.GetTables())
	if err != nil {
		return nil, err
	}
	m["tables"] = tables

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersFilterRowsTables(d *schema.ResourceData, v *datatransfer.TablesFilter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exclude_tables"] = v.GetExcludeTables()
	m["include_tables"] = v.GetIncludeTables()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersFilterColumns(d *schema.ResourceData, v *datatransfer.FilterColumnsTransformer) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	columns, err := flattenDatatransferTransferTransformationTransformersFilterColumnsColumns(d, v.GetColumns())
	if err != nil {
		return nil, err
	}
	m["columns"] = columns

	tables, err := flattenDatatransferTransferTransformationTransformersFilterColumnsTables(d, v.GetTables())
	if err != nil {
		return nil, err
	}
	m["tables"] = tables

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersFilterColumnsTables(d *schema.ResourceData, v *datatransfer.TablesFilter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exclude_tables"] = v.GetExcludeTables()
	m["include_tables"] = v.GetIncludeTables()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersFilterColumnsColumns(d *schema.ResourceData, v *datatransfer.ColumnsFilter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exclude_columns"] = v.GetExcludeColumns()
	m["include_columns"] = v.GetIncludeColumns()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersConvertToString(d *schema.ResourceData, v *datatransfer.ToStringTransformer) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	columns, err := flattenDatatransferTransferTransformationTransformersConvertToStringColumns(d, v.GetColumns())
	if err != nil {
		return nil, err
	}
	m["columns"] = columns

	tables, err := flattenDatatransferTransferTransformationTransformersConvertToStringTables(d, v.GetTables())
	if err != nil {
		return nil, err
	}
	m["tables"] = tables

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersConvertToStringTables(d *schema.ResourceData, v *datatransfer.TablesFilter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exclude_tables"] = v.GetExcludeTables()
	m["include_tables"] = v.GetIncludeTables()

	return []map[string]interface{}{m}, nil
}

func flattenDatatransferTransferTransformationTransformersConvertToStringColumns(d *schema.ResourceData, v *datatransfer.ColumnsFilter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exclude_columns"] = v.GetExcludeColumns()
	m["include_columns"] = v.GetIncludeColumns()

	return []map[string]interface{}{m}, nil
}

// Enum types parsers

func parseDatatransferTransferTransferStatus(str string) (datatransfer.TransferStatus, error) {
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

func parseDatatransferTransferTransferType(str string) (datatransfer.TransferType, error) {
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

func parseDatatransferEndpointColumnType(str string) (endpoint.ColumnType, error) {
	val, ok := endpoint.ColumnType_value[str]
	if !ok {
		return endpoint.ColumnType(0), fmt.Errorf(
			"value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(endpoint.ColumnType_value)),
			str,
		)
	}
	return endpoint.ColumnType(val), nil
}

func parseDatatransferEndpointKafkaMechanism(str string) (endpoint.KafkaMechanism, error) {
	val, ok := endpoint.KafkaMechanism_value[str]
	if !ok {
		return endpoint.KafkaMechanism(0), fmt.Errorf(
			"value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(endpoint.KafkaMechanism_value)),
			str,
		)
	}
	return endpoint.KafkaMechanism(val), nil
}

func parseDatatransferEndpointMetrikaStreamType(str string) (endpoint.MetrikaStreamType, error) {
	val, ok := endpoint.MetrikaStreamType_value[str]
	if !ok {
		return endpoint.MetrikaStreamType(0), fmt.Errorf(
			"value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(endpoint.MetrikaStreamType_value)),
			str,
		)
	}
	return endpoint.MetrikaStreamType(val), nil
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

func parseDatatransferEndpointYdbCleanupPolicy(str string) (endpoint.YdbCleanupPolicy, error) {
	val, ok := endpoint.YdbCleanupPolicy_value[str]
	if !ok {
		return endpoint.YdbCleanupPolicy(0), fmt.Errorf(
			"value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(endpoint.YdbCleanupPolicy_value)),
			str,
		)
	}
	return endpoint.YdbCleanupPolicy(val), nil
}

func parseDatatransferEndpointYdbDefaultCompression(str string) (endpoint.YdbDefaultCompression, error) {
	val, ok := endpoint.YdbDefaultCompression_value[str]
	if !ok {
		return endpoint.YdbDefaultCompression(0), fmt.Errorf(
			"value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(endpoint.YdbDefaultCompression_value)),
			str,
		)
	}
	return endpoint.YdbDefaultCompression(val), nil
}

func parseDatatransferEndpointYdsCompressionCodec(str string) (endpoint.YdsCompressionCodec, error) {
	val, ok := endpoint.YdsCompressionCodec_value[str]
	if !ok {
		return endpoint.YdsCompressionCodec(0), fmt.Errorf(
			"value for 'transfer_type' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(endpoint.YdsCompressionCodec_value)),
			str,
		)
	}
	return endpoint.YdsCompressionCodec(val), nil
}
