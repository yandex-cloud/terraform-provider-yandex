package trino_catalog

import (
	"context"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func BuildCreateCatalogRequest(ctx context.Context, catalogModel *CatalogModel, providerConfig *config.State) (*trino.CreateCatalogRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	clusterID := catalogModel.ClusterId.ValueString()
	common, _, dd := buildCommonForCreateAndUpdate(ctx, catalogModel, nil)
	diags.Append(dd...)
	if diags.HasError() {
		return nil, diags
	}

	clusterCreateRequest := &trino.CreateCatalogRequest{
		ClusterId: clusterID,
		Catalog: &trino.CatalogSpec{
			Name:        common.Name,
			Connector:   common.Connector,
			Description: common.Description,
			Labels:      common.Labels,
		},
	}

	return clusterCreateRequest, diags
}

func BuildUpdateCatalogRequest(ctx context.Context, state *CatalogModel, plan *CatalogModel) (*trino.UpdateCatalogRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	common, updateMaskPaths, dd := buildCommonForCreateAndUpdate(ctx, plan, state)
	diags.Append(dd...)

	updateCatalogRequest := &trino.UpdateCatalogRequest{
		ClusterId:  state.ClusterId.ValueString(),
		CatalogId:  state.Id.ValueString(),
		UpdateMask: &field_mask.FieldMask{Paths: updateMaskPaths},
		Catalog: &trino.CatalogUpdateSpec{
			Name:        common.Name,
			Connector:   common.Connector,
			Description: common.Description,
			Labels:      common.Labels,
		},
	}

	return updateCatalogRequest, diags
}

type CommonForCreateAndUpdate struct {
	Name        string
	Description string
	Labels      map[string]string
	Connector   *trino.Connector
}

func buildCommonForCreateAndUpdate(ctx context.Context, plan, state *CatalogModel) (*CommonForCreateAndUpdate, []string, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	updateMaskPaths := make([]string, 0)

	if state != nil {
		if !plan.Name.Equal(state.Name) {
			updateMaskPaths = append(updateMaskPaths, "catalog.name")
		}
		if !stringsAreEqual(plan.Description, state.Description) {
			updateMaskPaths = append(updateMaskPaths, "catalog.description")
		}
	}

	labels := make(map[string]string, len(plan.Labels.Elements()))
	diags.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}
	if state != nil && !mapsAreEqual(plan.Labels, state.Labels) {
		updateMaskPaths = append(updateMaskPaths, "catalog.labels")
	}

	// Connector
	connector := &trino.Connector{}
	switch {
	case plan.Postgresql != nil:
		postgresql, dd := postgresqlConnectorToAPI(ctx, plan.Postgresql)
		diags.Append(dd...)
		connector.Type = &trino.Connector_Postgresql{Postgresql: postgresql}
		if state != nil && !plan.Postgresql.Equal(state.Postgresql) {
			updateMaskPaths = append(updateMaskPaths, "catalog.connector.postgresql")
		}
	case plan.Hive != nil:
		hive, dd := hiveConnectorToAPI(ctx, plan.Hive)
		diags.Append(dd...)
		connector.Type = &trino.Connector_Hive{Hive: hive}
		if state != nil && !plan.Hive.Equal(state.Hive) {
			updateMaskPaths = append(updateMaskPaths, "catalog.connector.hive")
		}
	case plan.Clickhouse != nil:
		clickhouse, dd := clickhouseConnectorToAPI(ctx, plan.Clickhouse)
		diags.Append(dd...)
		connector.Type = &trino.Connector_Clickhouse{Clickhouse: clickhouse}
		if state != nil && !plan.Clickhouse.Equal(state.Clickhouse) {
			updateMaskPaths = append(updateMaskPaths, "catalog.connector.clickhouse")
		}
	case plan.DeltaLake != nil:
		deltaLake, dd := deltaLakeConnectorToAPI(ctx, plan.DeltaLake)
		diags.Append(dd...)
		connector.Type = &trino.Connector_DeltaLake{DeltaLake: deltaLake}
		if state != nil && !plan.DeltaLake.Equal(state.DeltaLake) {
			updateMaskPaths = append(updateMaskPaths, "catalog.connector.delta_lake")
		}
	case plan.Iceberg != nil:
		iceberg, dd := icebergConnectorToAPI(ctx, plan.Iceberg)
		diags.Append(dd...)
		connector.Type = &trino.Connector_Iceberg{Iceberg: iceberg}
		if state != nil && !plan.Iceberg.Equal(state.Iceberg) {
			updateMaskPaths = append(updateMaskPaths, "catalog.connector.iceberg")
		}
	case plan.Hudi != nil:
		hudi, dd := hudiConnectorToAPI(ctx, plan.Hudi)
		diags.Append(dd...)
		connector.Type = &trino.Connector_Hudi{Hudi: hudi}
		if state != nil && !plan.Hudi.Equal(state.Hudi) {
			updateMaskPaths = append(updateMaskPaths, "catalog.connector.hudi")
		}
	case plan.Oracle != nil:
		oracle, dd := oracleConnectorToAPI(ctx, plan.Oracle)
		diags.Append(dd...)
		connector.Type = &trino.Connector_Oracle{Oracle: oracle}
		if state != nil && !plan.Oracle.Equal(state.Oracle) {
			updateMaskPaths = append(updateMaskPaths, "catalog.connector.oracle")
		}
	case plan.Sqlserver != nil:
		sqlserver, dd := sqlserverConnectorToAPI(ctx, plan.Sqlserver)
		diags.Append(dd...)
		connector.Type = &trino.Connector_Sqlserver{Sqlserver: sqlserver}
		if state != nil && !plan.Sqlserver.Equal(state.Sqlserver) {
			updateMaskPaths = append(updateMaskPaths, "catalog.connector.sqlserver")
		}
	case plan.Tpcds != nil:
		tpcds, dd := tpcdsConnectorToAPI(ctx, plan.Tpcds)
		diags.Append(dd...)
		connector.Type = &trino.Connector_Tpcds{Tpcds: tpcds}
		if state != nil && !plan.Tpcds.Equal(state.Tpcds) {
			updateMaskPaths = append(updateMaskPaths, "catalog.connector.tpcds")
		}
	case plan.Tpch != nil:
		tpch, dd := tpchConnectorToAPI(ctx, plan.Tpch)
		diags.Append(dd...)
		connector.Type = &trino.Connector_Tpch{Tpch: tpch}
		if state != nil && !plan.Tpch.Equal(state.Tpch) {
			updateMaskPaths = append(updateMaskPaths, "catalog.connector.tpch")
		}
	}
	if diags.HasError() {
		return nil, nil, diags
	}

	params := &CommonForCreateAndUpdate{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Labels:      labels,
		Connector:   connector,
	}

	return params, updateMaskPaths, diags
}

func postgresqlConnectorToAPI(ctx context.Context, postgresql *Postgresql) (*trino.PostgresqlConnector, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	additionalProperties := make(map[string]string, len(postgresql.AdditionalProperties.Elements()))
	diags.Append(postgresql.AdditionalProperties.ElementsAs(ctx, &additionalProperties, false)...)

	connector := &trino.PostgresqlConnector{
		Connection:           &trino.PostgresqlConnection{},
		AdditionalProperties: additionalProperties,
	}

	switch {
	case !postgresql.OnPremise.IsNull() && !postgresql.OnPremise.IsUnknown():
		onPremise := OnPremise{}
		diags.Append(postgresql.OnPremise.As(ctx, &onPremise, baseOptions)...)
		connector.Connection.Type = &trino.PostgresqlConnection_OnPremise_{
			OnPremise: &trino.PostgresqlConnection_OnPremise{
				ConnectionUrl: onPremise.ConnectionUrl.ValueString(),
				UserName:      onPremise.UserName.ValueString(),
				Password:      onPremise.Password.ValueString(),
			},
		}
	case !postgresql.ConnectionManager.IsNull() && !postgresql.ConnectionManager.IsUnknown():
		connectionManager := ConnectionManager{}
		diags.Append(postgresql.ConnectionManager.As(ctx, &connectionManager, baseOptions)...)
		connectionProperties := make(map[string]string, len(connectionManager.ConnectionProperties.Elements()))
		diags.Append(connectionManager.ConnectionProperties.ElementsAs(ctx, &connectionProperties, false)...)

		connector.Connection.Type = &trino.PostgresqlConnection_ConnectionManager_{
			ConnectionManager: &trino.PostgresqlConnection_ConnectionManager{
				ConnectionId:         connectionManager.ConnectionId.ValueString(),
				Database:             connectionManager.Database.ValueString(),
				ConnectionProperties: connectionProperties,
			},
		}
	}

	return connector, diags
}

func hiveConnectorToAPI(ctx context.Context, hive *Hive) (*trino.HiveConnector, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	additionalProperties := make(map[string]string, len(hive.AdditionalProperties.Elements()))
	diags.Append(hive.AdditionalProperties.ElementsAs(ctx, &additionalProperties, false)...)

	metastore := Metastore{}
	diags.Append(hive.Metastore.As(ctx, &metastore, baseOptions)...)

	fileSystem := FileSystem{}
	diags.Append(hive.FileSystem.As(ctx, &fileSystem, baseOptions)...)

	connector := &trino.HiveConnector{
		Metastore: &trino.Metastore{
			Type: &trino.Metastore_Hive{
				Hive: &trino.Metastore_HiveMetastore{
					Connection: &trino.Metastore_HiveMetastore_Uri{
						Uri: metastore.Uri.ValueString(),
					},
				},
			},
		},
		Filesystem:           &trino.FileSystem{},
		AdditionalProperties: additionalProperties,
	}

	switch {
	case !fileSystem.S3.IsNull() && !fileSystem.S3.IsUnknown():
		s3 := S3{}
		diags.Append(fileSystem.S3.As(ctx, &s3, baseOptions)...)

		connector.Filesystem.Type = &trino.FileSystem_S3{
			S3: &trino.FileSystem_S3FileSystem{},
		}
	case !fileSystem.ExternalS3.IsNull() && !fileSystem.ExternalS3.IsUnknown():
		externalS3 := ExternalS3{}
		diags.Append(fileSystem.ExternalS3.As(ctx, &externalS3, baseOptions)...)

		connector.Filesystem.Type = &trino.FileSystem_ExternalS3{
			ExternalS3: &trino.FileSystem_ExternalS3FileSystem{
				AwsAccessKey: externalS3.AwsAccessKey.ValueString(),
				AwsSecretKey: externalS3.AwsSecretKey.ValueString(),
				AwsEndpoint:  externalS3.AwsEndpoint.ValueString(),
				AwsRegion:    externalS3.AwsRegion.ValueString(),
			},
		}
	}

	return connector, diags
}

func clickhouseConnectorToAPI(ctx context.Context, clickhouse *Clickhouse) (*trino.ClickhouseConnector, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	additionalProperties := make(map[string]string, len(clickhouse.AdditionalProperties.Elements()))
	diags.Append(clickhouse.AdditionalProperties.ElementsAs(ctx, &additionalProperties, false)...)

	connector := &trino.ClickhouseConnector{
		Connection:           &trino.ClickhouseConnection{},
		AdditionalProperties: additionalProperties,
	}

	switch {
	case !clickhouse.OnPremise.IsNull() && !clickhouse.OnPremise.IsUnknown():
		onPremise := OnPremise{}
		diags.Append(clickhouse.OnPremise.As(ctx, &onPremise, baseOptions)...)
		connector.Connection.Type = &trino.ClickhouseConnection_OnPremise_{
			OnPremise: &trino.ClickhouseConnection_OnPremise{
				ConnectionUrl: onPremise.ConnectionUrl.ValueString(),
				UserName:      onPremise.UserName.ValueString(),
				Password:      onPremise.Password.ValueString(),
			},
		}
	case !clickhouse.ConnectionManager.IsNull() && !clickhouse.ConnectionManager.IsUnknown():
		connectionManager := ConnectionManager{}
		diags.Append(clickhouse.ConnectionManager.As(ctx, &connectionManager, baseOptions)...)
		connectionProperties := make(map[string]string, len(connectionManager.ConnectionProperties.Elements()))
		diags.Append(connectionManager.ConnectionProperties.ElementsAs(ctx, &connectionProperties, false)...)

		connector.Connection.Type = &trino.ClickhouseConnection_ConnectionManager_{
			ConnectionManager: &trino.ClickhouseConnection_ConnectionManager{
				ConnectionId:         connectionManager.ConnectionId.ValueString(),
				Database:             connectionManager.Database.ValueString(),
				ConnectionProperties: connectionProperties,
			},
		}
	}

	return connector, diags
}

func deltaLakeConnectorToAPI(ctx context.Context, deltaLake *DeltaLake) (*trino.DeltaLakeConnector, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	additionalProperties := make(map[string]string, len(deltaLake.AdditionalProperties.Elements()))
	diags.Append(deltaLake.AdditionalProperties.ElementsAs(ctx, &additionalProperties, false)...)

	metastore := Metastore{}
	diags.Append(deltaLake.Metastore.As(ctx, &metastore, baseOptions)...)

	fileSystem := FileSystem{}
	diags.Append(deltaLake.FileSystem.As(ctx, &fileSystem, baseOptions)...)

	connector := &trino.DeltaLakeConnector{
		Metastore: &trino.Metastore{
			Type: &trino.Metastore_Hive{
				Hive: &trino.Metastore_HiveMetastore{
					Connection: &trino.Metastore_HiveMetastore_Uri{
						Uri: metastore.Uri.ValueString(),
					},
				},
			},
		},
		Filesystem:           &trino.FileSystem{},
		AdditionalProperties: additionalProperties,
	}

	switch {
	case !fileSystem.S3.IsNull() && !fileSystem.S3.IsUnknown():
		s3 := S3{}
		diags.Append(fileSystem.S3.As(ctx, &s3, baseOptions)...)

		connector.Filesystem.Type = &trino.FileSystem_S3{
			S3: &trino.FileSystem_S3FileSystem{},
		}
	case !fileSystem.ExternalS3.IsNull() && !fileSystem.ExternalS3.IsUnknown():
		externalS3 := ExternalS3{}
		diags.Append(fileSystem.ExternalS3.As(ctx, &externalS3, baseOptions)...)

		connector.Filesystem.Type = &trino.FileSystem_ExternalS3{
			ExternalS3: &trino.FileSystem_ExternalS3FileSystem{
				AwsAccessKey: externalS3.AwsAccessKey.ValueString(),
				AwsSecretKey: externalS3.AwsSecretKey.ValueString(),
				AwsEndpoint:  externalS3.AwsEndpoint.ValueString(),
				AwsRegion:    externalS3.AwsRegion.ValueString(),
			},
		}
	}

	return connector, diags
}

func icebergConnectorToAPI(ctx context.Context, iceberg *Iceberg) (*trino.IcebergConnector, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	additionalProperties := make(map[string]string, len(iceberg.AdditionalProperties.Elements()))
	diags.Append(iceberg.AdditionalProperties.ElementsAs(ctx, &additionalProperties, false)...)

	metastore := Metastore{}
	diags.Append(iceberg.Metastore.As(ctx, &metastore, baseOptions)...)

	fileSystem := FileSystem{}
	diags.Append(iceberg.FileSystem.As(ctx, &fileSystem, baseOptions)...)

	connector := &trino.IcebergConnector{
		Metastore: &trino.Metastore{
			Type: &trino.Metastore_Hive{
				Hive: &trino.Metastore_HiveMetastore{
					Connection: &trino.Metastore_HiveMetastore_Uri{
						Uri: metastore.Uri.ValueString(),
					},
				},
			},
		},
		Filesystem:           &trino.FileSystem{},
		AdditionalProperties: additionalProperties,
	}

	switch {
	case !fileSystem.S3.IsNull() && !fileSystem.S3.IsUnknown():
		s3 := S3{}
		diags.Append(fileSystem.S3.As(ctx, &s3, baseOptions)...)

		connector.Filesystem.Type = &trino.FileSystem_S3{
			S3: &trino.FileSystem_S3FileSystem{},
		}
	case !fileSystem.ExternalS3.IsNull() && !fileSystem.ExternalS3.IsUnknown():
		externalS3 := ExternalS3{}
		diags.Append(fileSystem.ExternalS3.As(ctx, &externalS3, baseOptions)...)

		connector.Filesystem.Type = &trino.FileSystem_ExternalS3{
			ExternalS3: &trino.FileSystem_ExternalS3FileSystem{
				AwsAccessKey: externalS3.AwsAccessKey.ValueString(),
				AwsSecretKey: externalS3.AwsSecretKey.ValueString(),
				AwsEndpoint:  externalS3.AwsEndpoint.ValueString(),
				AwsRegion:    externalS3.AwsRegion.ValueString(),
			},
		}
	}

	return connector, diags
}

func hudiConnectorToAPI(ctx context.Context, hudi *Hudi) (*trino.HudiConnector, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	additionalProperties := make(map[string]string, len(hudi.AdditionalProperties.Elements()))
	diags.Append(hudi.AdditionalProperties.ElementsAs(ctx, &additionalProperties, false)...)

	metastore := Metastore{}
	diags.Append(hudi.Metastore.As(ctx, &metastore, baseOptions)...)

	fileSystem := FileSystem{}
	diags.Append(hudi.FileSystem.As(ctx, &fileSystem, baseOptions)...)

	connector := &trino.HudiConnector{
		Metastore: &trino.Metastore{
			Type: &trino.Metastore_Hive{
				Hive: &trino.Metastore_HiveMetastore{
					Connection: &trino.Metastore_HiveMetastore_Uri{
						Uri: metastore.Uri.ValueString(),
					},
				},
			},
		},
		Filesystem:           &trino.FileSystem{},
		AdditionalProperties: additionalProperties,
	}

	switch {
	case !fileSystem.S3.IsNull() && !fileSystem.S3.IsUnknown():
		s3 := S3{}
		diags.Append(fileSystem.S3.As(ctx, &s3, baseOptions)...)

		connector.Filesystem.Type = &trino.FileSystem_S3{
			S3: &trino.FileSystem_S3FileSystem{},
		}
	case !fileSystem.ExternalS3.IsNull() && !fileSystem.ExternalS3.IsUnknown():
		externalS3 := ExternalS3{}
		diags.Append(fileSystem.ExternalS3.As(ctx, &externalS3, baseOptions)...)

		connector.Filesystem.Type = &trino.FileSystem_ExternalS3{
			ExternalS3: &trino.FileSystem_ExternalS3FileSystem{
				AwsAccessKey: externalS3.AwsAccessKey.ValueString(),
				AwsSecretKey: externalS3.AwsSecretKey.ValueString(),
				AwsEndpoint:  externalS3.AwsEndpoint.ValueString(),
				AwsRegion:    externalS3.AwsRegion.ValueString(),
			},
		}
	}

	return connector, diags
}

func oracleConnectorToAPI(ctx context.Context, oracle *Oracle) (*trino.OracleConnector, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	additionalProperties := make(map[string]string, len(oracle.AdditionalProperties.Elements()))
	diags.Append(oracle.AdditionalProperties.ElementsAs(ctx, &additionalProperties, false)...)

	connector := &trino.OracleConnector{
		Connection:           &trino.OracleConnection{},
		AdditionalProperties: additionalProperties,
	}

	switch {
	case !oracle.OnPremise.IsNull() && !oracle.OnPremise.IsUnknown():
		onPremise := OnPremise{}
		diags.Append(oracle.OnPremise.As(ctx, &onPremise, baseOptions)...)
		connector.Connection.Type = &trino.OracleConnection_OnPremise_{
			OnPremise: &trino.OracleConnection_OnPremise{
				ConnectionUrl: onPremise.ConnectionUrl.ValueString(),
				UserName:      onPremise.UserName.ValueString(),
				Password:      onPremise.Password.ValueString(),
			},
		}
	}

	return connector, diags
}

func sqlserverConnectorToAPI(ctx context.Context, sqlserver *Sqlserver) (*trino.SQLServerConnector, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	additionalProperties := make(map[string]string, len(sqlserver.AdditionalProperties.Elements()))
	diags.Append(sqlserver.AdditionalProperties.ElementsAs(ctx, &additionalProperties, false)...)

	connector := &trino.SQLServerConnector{
		Connection:           &trino.SQLServerConnection{},
		AdditionalProperties: additionalProperties,
	}

	switch {
	case !sqlserver.OnPremise.IsNull() && !sqlserver.OnPremise.IsUnknown():
		onPremise := OnPremise{}
		diags.Append(sqlserver.OnPremise.As(ctx, &onPremise, baseOptions)...)
		connector.Connection.Type = &trino.SQLServerConnection_OnPremise_{
			OnPremise: &trino.SQLServerConnection_OnPremise{
				ConnectionUrl: onPremise.ConnectionUrl.ValueString(),
				UserName:      onPremise.UserName.ValueString(),
				Password:      onPremise.Password.ValueString(),
			},
		}
	}

	return connector, diags
}

func tpcdsConnectorToAPI(ctx context.Context, tpcds *Tpcds) (*trino.TPCDSConnector, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	additionalProperties := make(map[string]string, len(tpcds.AdditionalProperties.Elements()))
	diags.Append(tpcds.AdditionalProperties.ElementsAs(ctx, &additionalProperties, false)...)

	connector := &trino.TPCDSConnector{
		AdditionalProperties: additionalProperties,
	}

	return connector, diags
}

func tpchConnectorToAPI(ctx context.Context, tpch *Tpch) (*trino.TPCHConnector, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	additionalProperties := make(map[string]string, len(tpch.AdditionalProperties.Elements()))
	diags.Append(tpch.AdditionalProperties.ElementsAs(ctx, &additionalProperties, false)...)

	connector := &trino.TPCHConnector{
		AdditionalProperties: additionalProperties,
	}

	return connector, diags
}

func mapsAreEqual(map1, map2 types.Map) bool {
	if map1.Equal(map2) {
		return true
	}
	// if one of map is null and the other is empty then we assume that they are equal
	if len(map1.Elements()) == 0 && len(map2.Elements()) == 0 {
		return true
	}
	return false
}

func stringsAreEqual(str1, str2 types.String) bool {
	if str1.Equal(str2) {
		return true
	}
	// if one of strings is null and the other is empty then we assume that they are equal
	if str1.ValueString() == "" && str2.ValueString() == "" {
		return true
	}
	return false
}
