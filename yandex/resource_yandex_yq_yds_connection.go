package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/client"
	yds_conn "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/yds_connection"
)

func resourceYandexYQYDSConnection() *schema.Resource {
	return &schema.Resource{
		Schema:        yds_conn.ResourceSchema(),
		SchemaVersion: 0,
		CreateContext: resourceYandexYQYDSConnectionCreate,
		ReadContext:   resourceYandexYQYDSConnectionRead,
		UpdateContext: resourceYandexYQYDSConnectionUpdate,
		DeleteContext: resourceYandexYQYDSConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: defaultTimeouts(),
	}
}

func resourceYandexYQYDSConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	if err := executeYandexYQYDSConnectionCreate(ctx, config.yqSdk.Client(), d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQYDSConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := executeYandexYQYDSConnectionRead(ctx, config.yqSdk.Client(), d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQYDSConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	if err := executeYandexYQYDSConnectionUpdate(ctx, config.yqSdk.Client(), d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQYDSConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	if err := executeYandexYQYDSConnectionDelete(ctx, config.yqSdk.Client(), d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func executeYandexYQYDSConnectionCreate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	config *Config,
) error {
	connectionName := d.Get(yds_conn.AttributeName).(string)
	serviceAccountID := d.Get(yds_conn.AttributeServiceAccountID).(string)
	databaseID := d.Get(yds_conn.AttributeDatabaseID).(string)
	description := d.Get(yds_conn.AttributeDescription).(string)

	auth := parseServiceIDToIAMAuth(serviceAccountID)

	req := Ydb_FederatedQuery.CreateConnectionRequest{
		Content: &Ydb_FederatedQuery.ConnectionContent{
			Name: connectionName,
			Acl: &Ydb_FederatedQuery.Acl{
				Visibility: Ydb_FederatedQuery.Acl_SCOPE,
			},
			Description: description,
			Setting: &Ydb_FederatedQuery.ConnectionSetting{
				Connection: &Ydb_FederatedQuery.ConnectionSetting_DataStreams{
					DataStreams: &Ydb_FederatedQuery.DataStreams{
						DatabaseId: databaseID,
						Auth:       auth,
					},
				},
			},
		},
	}

	if err := performYandexYQYDSConnectionCreate(ctx, client, d, &req); err != nil {
		return err
	}

	return executeYandexYQYDSConnectionRead(ctx, client, d, config)
}

func performYandexYQYDSConnectionCreate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	req *Ydb_FederatedQuery.CreateConnectionRequest,
) error {
	res, err := client.CreateConnection(ctx, req)
	if err != nil {
		return err
	}

	d.SetId(res.ConnectionId)

	return nil
}

func executeYandexYQYDSConnectionRead(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	_ *Config,
) error {
	id := d.Id()

	req := &Ydb_FederatedQuery.DescribeConnectionRequest{
		ConnectionId: id,
	}

	connectionRes, err := client.DescribeConnection(ctx, req)
	if err != nil {
		return err
	}

	return flattenYandexYQYDSConnection(d, connectionRes)
}

func flattenYandexYQYDSConnection(
	d *schema.ResourceData,
	connectionRes *Ydb_FederatedQuery.DescribeConnectionResult,
) error {
	if connectionRes == nil {
		d.SetId("")
		return nil
	}

	connection := connectionRes.GetConnection()

	if err := flattenYandexYQYDSConnectionContent(d, connection.GetContent()); err != nil {
		return err
	}

	if err := flattenYandexYQCommonMeta(d, connection.GetMeta()); err != nil {
		return err
	}

	return nil
}

func flattenYandexYQYDSConnectionContent(
	d *schema.ResourceData,
	content *Ydb_FederatedQuery.ConnectionContent,
) error {
	d.Set(yds_conn.AttributeName, content.GetName())
	d.Set(yds_conn.AttributeDescription, content.GetDescription())
	if err := flattenYandexYQYDSConnectionSetting(d, content.GetSetting()); err != nil {
		return err
	}

	return nil
}

func flattenYandexYQYDSConnectionSetting(
	d *schema.ResourceData,
	setting *Ydb_FederatedQuery.ConnectionSetting,
) error {
	dataStreamsSetting := setting.GetDataStreams()

	d.Set(yds_conn.AttributeDatabaseID, dataStreamsSetting.GetDatabaseId())

	if err := flattenYandexYQAuth2(d, dataStreamsSetting.GetAuth()); err != nil {
		return err
	}

	return nil
}

func flattenYandexYQAuth2(d *schema.ResourceData,
	auth *Ydb_FederatedQuery.IamAuth,
) error {
	serviceAccountID, err := iAMAuthToString(auth)
	if err != nil {
		return err
	}

	d.Set(yds_conn.AttributeServiceAccountID, serviceAccountID)

	return nil
}

func executeYandexYQYDSConnectionUpdate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	config *Config,
) error {
	connectionName := d.Get(yds_conn.AttributeName).(string)
	serviceAccountID := d.Get(yds_conn.AttributeServiceAccountID).(string)
	databaseID := d.Get(yds_conn.AttributeDatabaseID).(string)
	description := d.Get(yds_conn.AttributeDescription).(string)

	auth := parseServiceIDToIAMAuth(serviceAccountID)
	id := d.Id()

	req := &Ydb_FederatedQuery.ModifyConnectionRequest{
		ConnectionId: id,
		Content: &Ydb_FederatedQuery.ConnectionContent{
			Name:        connectionName,
			Description: description,
			Setting: &Ydb_FederatedQuery.ConnectionSetting{
				Connection: &Ydb_FederatedQuery.ConnectionSetting_DataStreams{
					DataStreams: &Ydb_FederatedQuery.DataStreams{
						DatabaseId: databaseID,
						Auth:       auth,
					},
				},
			},
			Acl: &Ydb_FederatedQuery.Acl{
				Visibility: Ydb_FederatedQuery.Acl_SCOPE,
			},
		},
	}

	if _, err := client.ModifyConnection(ctx, req); err != nil {
		return err
	}

	return executeYandexYQYDSConnectionRead(ctx, client, d, config)
}

func executeYandexYQYDSConnectionDelete(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	_ *Config,
) error {
	id := d.Id()

	req := &Ydb_FederatedQuery.DeleteConnectionRequest{
		ConnectionId: id,
	}

	_, err := client.DeleteConnection(ctx, req)
	return err
}
