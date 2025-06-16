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
		if state.Postgresql == nil {
			state.Postgresql = NewPostgresqlNull()
		}
		diags.Append(postgresqlToModel(ctx, connector.Postgresql, state.Postgresql)...)
	case *trino.Connector_Hive:
		if state.Hive == nil {
			state.Hive = NewHiveNull()
		}
		diags.Append(hiveToModel(ctx, connector.Hive, state.Hive)...)
	case *trino.Connector_Clickhouse:
		if state.Clickhouse == nil {
			state.Clickhouse = NewClickhouseNull()
		}
		diags.Append(clickhouseToModel(ctx, connector.Clickhouse, state.Clickhouse)...)
	case *trino.Connector_DeltaLake:
		if state.DeltaLake == nil {
			state.DeltaLake = NewDeltaLakeNull()
		}
		diags.Append(deltaLakeToModel(ctx, connector.DeltaLake, state.DeltaLake)...)
	case *trino.Connector_Iceberg:
		if state.Iceberg == nil {
			state.Iceberg = NewIcebergNull()
		}
		diags.Append(icebergToModel(ctx, connector.Iceberg, state.Iceberg)...)
	case *trino.Connector_Oracle:
		if state.Oracle == nil {
			state.Oracle = NewOracleNull()
		}
		diags.Append(oracleToModel(ctx, connector.Oracle, state.Oracle)...)
	case *trino.Connector_Sqlserver:
		if state.Sqlserver == nil {
			state.Sqlserver = NewSqlserverNull()
		}
		diags.Append(sqlserverToModel(ctx, connector.Sqlserver, state.Sqlserver)...)
	case *trino.Connector_Tpcds:
		if state.Tpcds == nil {
			state.Tpcds = NewTpcdsNull()
		}
		diags.Append(tpcdsToModel(ctx, connector.Tpcds, state.Tpcds)...)
	case *trino.Connector_Tpch:
		if state.Tpch == nil {
			state.Tpch = NewTpchNull()
		}
		diags.Append(tpchToModel(ctx, connector.Tpch, state.Tpch)...)
	}

	return diags
}

func postgresqlToModel(ctx context.Context, postgresql *trino.PostgresqlConnector, state *Postgresql) diag.Diagnostics {
	diags := diag.Diagnostics{}
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

	return diags
}

func hiveToModel(ctx context.Context, hive *trino.HiveConnector, state *Hive) diag.Diagnostics {
	diags := diag.Diagnostics{}
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

	return diags
}

func clickhouseToModel(ctx context.Context, clickhouse *trino.ClickhouseConnector, state *Clickhouse) diag.Diagnostics {
	diags := diag.Diagnostics{}
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

	return diags
}

func deltaLakeToModel(ctx context.Context, deltaLake *trino.DeltaLakeConnector, state *DeltaLake) diag.Diagnostics {
	diags := diag.Diagnostics{}
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

	return diags
}

func icebergToModel(ctx context.Context, iceberg *trino.IcebergConnector, state *Iceberg) diag.Diagnostics {
	diags := diag.Diagnostics{}
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

	return diags
}

func oracleToModel(ctx context.Context, oracle *trino.OracleConnector, state *Oracle) diag.Diagnostics {
	diags := diag.Diagnostics{}
	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, oracle.AdditionalProperties)
	diags.Append(dd...)
	state.AdditionalProperties = additionalProperties

	switch connection := oracle.Connection.Type.(type) {
	case *trino.OracleConnection_OnPremise_:
		obPremiseObject, dd := onPremiseToModel(ctx, state.OnPremise, connection.OnPremise)
		diags.Append(dd...)
		state.OnPremise = obPremiseObject
	}

	return diags
}

func sqlserverToModel(ctx context.Context, sqlserver *trino.SQLServerConnector, state *Sqlserver) diag.Diagnostics {
	diags := diag.Diagnostics{}
	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, sqlserver.AdditionalProperties)
	diags.Append(dd...)
	state.AdditionalProperties = additionalProperties

	switch connection := sqlserver.Connection.Type.(type) {
	case *trino.SQLServerConnection_OnPremise_:
		obPremiseObject, dd := onPremiseToModel(ctx, state.OnPremise, connection.OnPremise)
		diags.Append(dd...)
		state.OnPremise = obPremiseObject
	}

	return diags
}

func tpcdsToModel(ctx context.Context, tpcds *trino.TPCDSConnector, state *Tpcds) diag.Diagnostics {
	diags := diag.Diagnostics{}
	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, tpcds.AdditionalProperties)
	diags.Append(dd...)
	state.AdditionalProperties = additionalProperties

	return diags
}

func tpchToModel(ctx context.Context, tpch *trino.TPCHConnector, state *Tpch) diag.Diagnostics {
	diags := diag.Diagnostics{}
	additionalProperties, dd := types.MapValueFrom(ctx, types.StringType, tpch.AdditionalProperties)
	diags.Append(dd...)
	state.AdditionalProperties = additionalProperties

	return diags
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
