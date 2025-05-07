package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/client"
	os_conn "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/object_storage_connection"
)

func resourceYandexYQObjectStorageConnection() *schema.Resource {
	return &schema.Resource{
		Schema:        os_conn.ResourceSchema(),
		SchemaVersion: 0,
		CreateContext: resourceYandexYQObjectStorageConnectionCreate,
		ReadContext:   resourceYandexYQObjectStorageConnectionRead,
		UpdateContext: resourceYandexYQObjectStorageConnectionUpdate,
		DeleteContext: resourceYandexYQObjectStorageConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: defaultTimeouts(),
	}
}

func resourceYandexYQObjectStorageConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	if err := executeYandexYQObjectStorageConnectionCreate(ctx, config.yqSdk.Client(), d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQObjectStorageConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := executeYandexYQObjectStorageConnectionRead(ctx, config.yqSdk.Client(), d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQObjectStorageConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	if err := executeYandexYQObjectStorageConnectionUpdate(ctx, config.yqSdk.Client(), d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQObjectStorageConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	if err := executeYandexYQObjectStorageConnectionDelete(ctx, config.yqSdk.Client(), d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func executeYandexYQObjectStorageConnectionCreate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	config *Config,
) error {
	connectionName := d.Get(os_conn.AttributeName).(string)
	serviceAccountID := d.Get(os_conn.AttributeServiceAccountID).(string)
	bucket := d.Get(os_conn.AttributeBucket).(string)
	description := d.Get(os_conn.AttributeDescription).(string)

	auth := parseServiceIDToIAMAuth(serviceAccountID)

	req := Ydb_FederatedQuery.CreateConnectionRequest{
		Content: &Ydb_FederatedQuery.ConnectionContent{
			Name: connectionName,
			Acl: &Ydb_FederatedQuery.Acl{
				Visibility: Ydb_FederatedQuery.Acl_SCOPE,
			},
			Description: description,
			Setting: &Ydb_FederatedQuery.ConnectionSetting{
				Connection: &Ydb_FederatedQuery.ConnectionSetting_ObjectStorage{
					ObjectStorage: &Ydb_FederatedQuery.ObjectStorageConnection{
						Bucket: bucket,
						Auth:   auth,
					},
				},
			},
		},
	}

	if err := performYandexYQObjectStorageConnectionCreate(ctx, client, d, &req); err != nil {
		return err
	}

	return executeYandexYQObjectStorageConnectionRead(ctx, client, d, config)
}

func performYandexYQObjectStorageConnectionCreate(
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

func executeYandexYQObjectStorageConnectionRead(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	_ *Config,
) error {
	id := d.Id()

	req := &Ydb_FederatedQuery.DescribeConnectionRequest{
		ConnectionId: id,
	}

	connectionRes, err := performYandexYQObjectStorageConnectionRead(ctx, client, d, req)
	if err != nil {
		return err
	}

	return flattenYandexYQObjectStorageConnection(d, connectionRes)
}

func performYandexYQObjectStorageConnectionRead(
	ctx context.Context,
	client client.YQClient,
	_ *schema.ResourceData,
	req *Ydb_FederatedQuery.DescribeConnectionRequest,
) (*Ydb_FederatedQuery.DescribeConnectionResult, error) {
	res, err := client.DescribeConnection(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func flattenYandexYQObjectStorageConnection(
	d *schema.ResourceData,
	connectionRes *Ydb_FederatedQuery.DescribeConnectionResult,
) error {
	if connectionRes == nil {
		d.SetId("")
		return nil
	}

	connection := connectionRes.GetConnection()

	if err := flattenYandexYQConnectionContent(d, connection.GetContent()); err != nil {
		return err
	}

	if err := flattenYandexYQCommonMeta(d, connection.GetMeta()); err != nil {
		return err
	}

	return nil
}

func flattenYandexYQConnectionContent(
	d *schema.ResourceData,
	content *Ydb_FederatedQuery.ConnectionContent,
) error {
	d.Set(os_conn.AttributeName, content.GetName())
	d.Set(os_conn.AttributeDescription, content.GetDescription())
	if err := flattenYandexYQConnectionSetting(d, content.GetSetting()); err != nil {
		return err
	}

	return nil
}

func flattenYandexYQConnectionSetting(
	d *schema.ResourceData,
	setting *Ydb_FederatedQuery.ConnectionSetting,
) error {
	objectStorageSetting := setting.GetObjectStorage()

	d.Set(os_conn.AttributeBucket, objectStorageSetting.GetBucket())

	if err := flattenYandexYQAuth(d, objectStorageSetting.GetAuth()); err != nil {
		return err
	}

	return nil
}

func flattenYandexYQAuth(d *schema.ResourceData,
	auth *Ydb_FederatedQuery.IamAuth,
) error {
	serviceAccountID, err := iAMAuthToString(auth)
	if err != nil {
		return err
	}

	d.Set(os_conn.AttributeServiceAccountID, serviceAccountID)

	return nil
}

func executeYandexYQObjectStorageConnectionUpdate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	config *Config,
) error {
	connectionName := d.Get(os_conn.AttributeName).(string)
	serviceAccountID := d.Get(os_conn.AttributeServiceAccountID).(string)
	bucket := d.Get(os_conn.AttributeBucket).(string)
	description := d.Get(os_conn.AttributeDescription).(string)

	auth := parseServiceIDToIAMAuth(serviceAccountID)
	id := d.Id()

	req := &Ydb_FederatedQuery.ModifyConnectionRequest{
		ConnectionId: id,
		Content: &Ydb_FederatedQuery.ConnectionContent{
			Name:        connectionName,
			Description: description,
			Setting: &Ydb_FederatedQuery.ConnectionSetting{
				Connection: &Ydb_FederatedQuery.ConnectionSetting_ObjectStorage{
					ObjectStorage: &Ydb_FederatedQuery.ObjectStorageConnection{
						Bucket: bucket,
						Auth:   auth,
					},
				},
			},
			Acl: &Ydb_FederatedQuery.Acl{
				Visibility: Ydb_FederatedQuery.Acl_SCOPE,
			},
		},
	}

	if err := performYandexYQObjectStorageConnectionUpdate(ctx, client, d, req); err != nil {
		return err
	}

	return executeYandexYQObjectStorageConnectionRead(ctx, client, d, config)
}

func performYandexYQObjectStorageConnectionUpdate(
	ctx context.Context,
	client client.YQClient,
	_ *schema.ResourceData,
	req *Ydb_FederatedQuery.ModifyConnectionRequest,
) error {
	_, err := client.ModifyConnection(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func executeYandexYQObjectStorageConnectionDelete(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	_ *Config,
) error {
	id := d.Id()

	req := &Ydb_FederatedQuery.DeleteConnectionRequest{
		ConnectionId: id,
	}

	err := performYandexYQObjectStorageConnectionDelete(ctx, client, d, req)
	if err != nil {
		return err
	}

	return nil
}

func performYandexYQObjectStorageConnectionDelete(
	ctx context.Context,
	client client.YQClient,
	_ *schema.ResourceData,
	req *Ydb_FederatedQuery.DeleteConnectionRequest,
) error {
	_, err := client.DeleteConnection(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
