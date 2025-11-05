package trino_catalog

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
)

func CatalogToState(ctx context.Context, catalog *trino.Catalog, state *CatalogModel) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("clusterToState: Trino cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("clusterToState: Received Trino cluster data: %+v", catalog))

	state.Name = types.StringValue(catalog.GetName())

	newDescription := types.StringValue(catalog.GetDescription())
	if !stringsAreEqual(state.Description, newDescription) {
		state.Description = newDescription
	}

	labels, diags := types.MapValueFrom(ctx, types.StringType, catalog.GetLabels())
	if diags.HasError() {
		return diags
	}
	state.Labels = labels

	switch connector := catalog.Connector.Type.(type) {
	case *trino.Connector_Postgresql:
		postgresqlObject, dd := postgresqlToModelObject(ctx, connector.Postgresql, state.Postgresql)
		diags.Append(dd...)
		state.Postgresql = postgresqlObject
	case *trino.Connector_Hive:
		hiveObject, dd := hiveToModelObject(ctx, connector.Hive, state.Hive)
		diags.Append(dd...)
		state.Hive = hiveObject
	case *trino.Connector_Hudi:
		hudiObject, dd := hudiToModelObject(ctx, connector.Hudi, state.Hudi)
		diags.Append(dd...)
		state.Hudi = hudiObject
	case *trino.Connector_Clickhouse:
		clickhouseObject, dd := clickhouseToModelObject(ctx, connector.Clickhouse, state.Clickhouse)
		diags.Append(dd...)
		state.Clickhouse = clickhouseObject
	case *trino.Connector_DeltaLake:
		deltaLakeObject, dd := deltaLakeToModelObject(ctx, connector.DeltaLake, state.DeltaLake)
		diags.Append(dd...)
		state.DeltaLake = deltaLakeObject
	case *trino.Connector_Iceberg:
		icebergObject, dd := icebergToModelObject(ctx, connector.Iceberg, state.Iceberg)
		diags.Append(dd...)
		state.Iceberg = icebergObject
	case *trino.Connector_Oracle:
		oracleObject, dd := oracleToModelObject(ctx, connector.Oracle, state.Oracle)
		diags.Append(dd...)
		state.Oracle = oracleObject
	case *trino.Connector_Sqlserver:
		sqlserverObject, dd := sqlserverToModelObject(ctx, connector.Sqlserver, state.Sqlserver)
		diags.Append(dd...)
		state.Sqlserver = sqlserverObject
	case *trino.Connector_Tpcds:
		tpcdsObject, dd := tpcdsToModelObject(ctx, connector.Tpcds, state.Tpcds)
		diags.Append(dd...)
		state.Tpcds = tpcdsObject
	case *trino.Connector_Tpch:
		tpchObject, dd := tpchToModelObject(ctx, connector.Tpch, state.Tpch)
		diags.Append(dd...)
		state.Tpch = tpchObject
	}

	return diags
}

func postgresqlToModelObject(ctx context.Context, postgresql *trino.PostgresqlConnector, stateObj types.Object) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	state := Postgresql{}
	if !stateObj.IsNull() && !stateObj.IsUnknown() {
		diags.Append(stateObj.As(ctx, &state, baseOptions)...)
	} else {
		state = NewPostgresqlNull()
	}

	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, postgresql.AdditionalProperties)
	diags.Append(dd...)
	if !mapsAreEqual(state.AdditionalProperties, additionalProperties) {
		state.AdditionalProperties = additionalProperties
	}

	switch connection := postgresql.Connection.Type.(type) {
	case *trino.PostgresqlConnection_OnPremise_:
		obPremiseObject, dd := onPremiseToModel(ctx, state.OnPremise, connection.OnPremise)
		diags.Append(dd...)
		state.OnPremise = obPremiseObject
	case *trino.PostgresqlConnection_ConnectionManager_:
		connectionManagerObject, dd := connectionManagerToModel(ctx, connection.ConnectionManager)
		diags.Append(dd...)
		state.ConnectionManager = connectionManagerObject
	}

	return types.ObjectValueFrom(ctx, PostgresqlT.AttributeTypes(), state)
}

func hiveToModelObject(ctx context.Context, hive *trino.HiveConnector, stateObj types.Object) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	state := Hive{}
	if !stateObj.IsNull() && !stateObj.IsUnknown() {
		diags.Append(stateObj.As(ctx, &state, baseOptions)...)
	} else {
		state = NewHiveNull()
	}

	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, hive.AdditionalProperties)
	diags.Append(dd...)
	if !mapsAreEqual(state.AdditionalProperties, additionalProperties) {
		state.AdditionalProperties = additionalProperties
	}

	// Handle metastore
	if hive.Metastore != nil && hive.Metastore.GetHive() != nil {
		metastore := Metastore{
			Uri: types.StringValue(hive.Metastore.GetHive().GetUri()),
		}
		metastoreObject, dd := types.ObjectValueFrom(ctx, MetastoreT.AttributeTypes(), metastore)
		diags.Append(dd...)
		state.Metastore = metastoreObject
	}

	// Handle file system
	if hive.Filesystem != nil {
		fileSystemObject, dd := fileSystemToModel(ctx, state.FileSystem, hive.Filesystem)
		diags.Append(dd...)
		state.FileSystem = fileSystemObject
	}

	return types.ObjectValueFrom(ctx, HiveT.AttributeTypes(), state)
}

func hudiToModelObject(ctx context.Context, hudi *trino.HudiConnector, stateObj types.Object) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	state := Hudi{}
	if !stateObj.IsNull() && !stateObj.IsUnknown() {
		diags.Append(stateObj.As(ctx, &state, baseOptions)...)
	} else {
		state = NewHudiNull()
	}

	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, hudi.AdditionalProperties)
	diags.Append(dd...)
	if !mapsAreEqual(state.AdditionalProperties, additionalProperties) {
		state.AdditionalProperties = additionalProperties
	}

	// Handle metastore
	if hudi.Metastore != nil && hudi.Metastore.GetHive() != nil {
		metastore := Metastore{
			Uri: types.StringValue(hudi.Metastore.GetHive().GetUri()),
		}
		metastoreObject, dd := types.ObjectValueFrom(ctx, MetastoreT.AttributeTypes(), metastore)
		diags.Append(dd...)
		state.Metastore = metastoreObject
	}

	// Handle file system
	if hudi.Filesystem != nil {
		fileSystemObject, dd := fileSystemToModel(ctx, state.FileSystem, hudi.Filesystem)
		diags.Append(dd...)
		state.FileSystem = fileSystemObject
	}

	return types.ObjectValueFrom(ctx, HudiT.AttributeTypes(), state)
}

func clickhouseToModelObject(ctx context.Context, clickhouse *trino.ClickhouseConnector, stateObj types.Object) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	state := Clickhouse{}
	if !stateObj.IsNull() && !stateObj.IsUnknown() {
		diags.Append(stateObj.As(ctx, &state, baseOptions)...)
	} else {
		state = NewClickhouseNull()
	}

	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, clickhouse.AdditionalProperties)
	diags.Append(dd...)
	if !mapsAreEqual(state.AdditionalProperties, additionalProperties) {
		state.AdditionalProperties = additionalProperties
	}

	switch connection := clickhouse.Connection.Type.(type) {
	case *trino.ClickhouseConnection_OnPremise_:
		obPremiseObject, dd := onPremiseToModel(ctx, state.OnPremise, connection.OnPremise)
		diags.Append(dd...)
		state.OnPremise = obPremiseObject
	case *trino.ClickhouseConnection_ConnectionManager_:
		connectionManagerObject, dd := connectionManagerToModel(ctx, connection.ConnectionManager)
		diags.Append(dd...)
		state.ConnectionManager = connectionManagerObject
	}

	return types.ObjectValueFrom(ctx, ClickhouseT.AttributeTypes(), state)
}

func deltaLakeToModelObject(ctx context.Context, deltaLake *trino.DeltaLakeConnector, stateObj types.Object) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	state := DeltaLake{}
	if !stateObj.IsNull() && !stateObj.IsUnknown() {
		diags.Append(stateObj.As(ctx, &state, baseOptions)...)
	} else {
		state = NewDeltaLakeNull()
	}

	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, deltaLake.AdditionalProperties)
	diags.Append(dd...)
	if !mapsAreEqual(state.AdditionalProperties, additionalProperties) {
		state.AdditionalProperties = additionalProperties
	}

	// Handle metastore
	if deltaLake.Metastore != nil && deltaLake.Metastore.GetHive() != nil {
		metastore := Metastore{
			Uri: types.StringValue(deltaLake.Metastore.GetHive().GetUri()),
		}
		metastoreObject, dd := types.ObjectValueFrom(ctx, MetastoreT.AttributeTypes(), metastore)
		diags.Append(dd...)
		state.Metastore = metastoreObject
	}

	// Handle file system
	if deltaLake.Filesystem != nil {
		fileSystemObject, dd := fileSystemToModel(ctx, state.FileSystem, deltaLake.Filesystem)
		diags.Append(dd...)
		state.FileSystem = fileSystemObject
	}

	return types.ObjectValueFrom(ctx, DeltaLakeT.AttributeTypes(), state)
}

func icebergToModelObject(ctx context.Context, iceberg *trino.IcebergConnector, stateObj types.Object) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	state := Iceberg{}
	if !stateObj.IsNull() && !stateObj.IsUnknown() {
		diags.Append(stateObj.As(ctx, &state, baseOptions)...)
	} else {
		state = NewIcebergNull()
	}

	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, iceberg.AdditionalProperties)
	diags.Append(dd...)
	if !mapsAreEqual(state.AdditionalProperties, additionalProperties) {
		state.AdditionalProperties = additionalProperties
	}

	// Handle metastore
	if iceberg.Metastore != nil && iceberg.Metastore.GetHive() != nil {
		metastore := Metastore{
			Uri: types.StringValue(iceberg.Metastore.GetHive().GetUri()),
		}
		metastoreObject, dd := types.ObjectValueFrom(ctx, MetastoreT.AttributeTypes(), metastore)
		diags.Append(dd...)
		state.Metastore = metastoreObject
	}

	// Handle file system
	if iceberg.Filesystem != nil {
		fileSystemObject, dd := fileSystemToModel(ctx, state.FileSystem, iceberg.Filesystem)
		diags.Append(dd...)
		state.FileSystem = fileSystemObject
	}

	return types.ObjectValueFrom(ctx, IcebergT.AttributeTypes(), state)
}

func oracleToModelObject(ctx context.Context, oracle *trino.OracleConnector, stateObj types.Object) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	state := Oracle{}
	if !stateObj.IsNull() && !stateObj.IsUnknown() {
		diags.Append(stateObj.As(ctx, &state, baseOptions)...)
	} else {
		state = NewOracleNull()
	}

	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, oracle.AdditionalProperties)
	diags.Append(dd...)
	state.AdditionalProperties = additionalProperties

	switch connection := oracle.Connection.Type.(type) {
	case *trino.OracleConnection_OnPremise_:
		obPremiseObject, dd := onPremiseToModel(ctx, state.OnPremise, connection.OnPremise)
		diags.Append(dd...)
		state.OnPremise = obPremiseObject
	}

	return types.ObjectValueFrom(ctx, OracleT.AttributeTypes(), state)
}

func sqlserverToModelObject(ctx context.Context, sqlserver *trino.SQLServerConnector, stateObj types.Object) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	state := Sqlserver{}
	if !stateObj.IsNull() && !stateObj.IsUnknown() {
		diags.Append(stateObj.As(ctx, &state, baseOptions)...)
	} else {
		state = NewSqlserverNull()
	}

	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, sqlserver.AdditionalProperties)
	diags.Append(dd...)
	state.AdditionalProperties = additionalProperties

	switch connection := sqlserver.Connection.Type.(type) {
	case *trino.SQLServerConnection_OnPremise_:
		obPremiseObject, dd := onPremiseToModel(ctx, state.OnPremise, connection.OnPremise)
		diags.Append(dd...)
		state.OnPremise = obPremiseObject
	}

	return types.ObjectValueFrom(ctx, SqlserverT.AttributeTypes(), state)
}

func tpcdsToModelObject(ctx context.Context, tpcds *trino.TPCDSConnector, stateObj types.Object) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	state := Tpcds{}
	if !stateObj.IsNull() && !stateObj.IsUnknown() {
		diags.Append(stateObj.As(ctx, &state, baseOptions)...)
	} else {
		state = NewTpcdsNull()
	}

	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, tpcds.AdditionalProperties)
	diags.Append(dd...)
	state.AdditionalProperties = additionalProperties

	return types.ObjectValueFrom(ctx, TpcdsT.AttributeTypes(), state)
}

func tpchToModelObject(ctx context.Context, tpch *trino.TPCHConnector, stateObj types.Object) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	state := Tpch{}
	if !stateObj.IsNull() && !stateObj.IsUnknown() {
		diags.Append(stateObj.As(ctx, &state, baseOptions)...)
	} else {
		state = NewTpchNull()
	}

	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, tpch.AdditionalProperties)
	diags.Append(dd...)
	state.AdditionalProperties = additionalProperties

	return types.ObjectValueFrom(ctx, TpchT.AttributeTypes(), state)
}

func fileSystemToModel(ctx context.Context, state types.Object, apiFileSystem *trino.FileSystem) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	fileSystem := NewFileSystemNull()
	// This code allows as to set set sensitive fields from the state. Because the State could be nil, the error is ignored.
	_ = state.As(ctx, &fileSystem, baseOptions)
	switch fs := apiFileSystem.Type.(type) {
	case *trino.FileSystem_S3:
		s3 := S3{}
		s3Object, dd := types.ObjectValueFrom(ctx, S3T.AttributeTypes(), s3)
		diags.Append(dd...)
		fileSystem.S3 = s3Object
	case *trino.FileSystem_ExternalS3:
		externalS3Object, dd := externalS3ToModel(ctx, fileSystem, fs)
		diags.Append(dd...)
		fileSystem.ExternalS3 = externalS3Object
	}
	fileSystemObject, dd := types.ObjectValueFrom(ctx, FileSystemT.AttributeTypes(), fileSystem)
	return fileSystemObject, dd
}

func externalS3ToModel(ctx context.Context, fileSystem FileSystem, fs *trino.FileSystem_ExternalS3) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	externalS3 := ExternalS3{}
	// This code allows as to set set sensitive fields from the state. Because the State could be nil, the error is ignored.
	_ = fileSystem.ExternalS3.As(ctx, &externalS3, baseOptions)

	externalS3.AwsEndpoint = types.StringValue(fs.ExternalS3.AwsEndpoint)
	externalS3.AwsRegion = types.StringValue(fs.ExternalS3.AwsRegion)

	externalS3Object, dd := types.ObjectValueFrom(ctx, ExternalS3T.AttributeTypes(), externalS3)
	diags.Append(dd...)
	return externalS3Object, diags
}

type onPremiseAPI interface {
	GetConnectionUrl() string
	GetUserName() string
}

func onPremiseToModel(ctx context.Context, state types.Object, apiOnPremise onPremiseAPI) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	onPremise := OnPremise{}
	// This code allows as to set password as in the state. Because the State could be nil, the error is ignored.
	_ = state.As(ctx, &onPremise, baseOptions)

	onPremise.ConnectionUrl = types.StringValue(apiOnPremise.GetConnectionUrl())
	onPremise.UserName = types.StringValue(apiOnPremise.GetUserName())

	obPremiseObject, dd := types.ObjectValueFrom(ctx, OnPremiseT.AttributeTypes(), onPremise)
	diags.Append(dd...)
	return obPremiseObject, diags
}

type connectionManagerAPI interface {
	GetConnectionId() string
	GetDatabase() string
	GetConnectionProperties() map[string]string
}

func connectionManagerToModel(ctx context.Context, apiConnectionManager connectionManagerAPI) (types.Object, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	connectionProperties, dd := types.MapValueFrom(ctx, types.StringType, apiConnectionManager.GetConnectionProperties())
	diags.Append(dd...)

	connectionManager := ConnectionManager{
		ConnectionId:         types.StringValue(apiConnectionManager.GetConnectionId()),
		Database:             types.StringValue(apiConnectionManager.GetDatabase()),
		ConnectionProperties: connectionProperties,
	}

	connectionManagerObject, dd := types.ObjectValueFrom(ctx, ConnectionManagerT.AttributeTypes(), connectionManager)
	diags.Append(dd...)
	return connectionManagerObject, diags
}
