package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"

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

	client, err := config.yqSdk.ObjectStorageConnectionCaller(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = executeYandexYQObjectStorageConnectionCreate(ctx, client, d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQObjectStorageConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	client, err := config.yqSdk.ObjectStorageConnectionCaller(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	err = executeYandexYQObjectStorageConnectionRead(ctx, client, d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQObjectStorageConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	client, err := config.yqSdk.ObjectStorageConnectionCaller(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = executeYandexYQObjectStorageConnectionUpdate(ctx, client, d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQObjectStorageConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	client, err := config.yqSdk.ObjectStorageConnectionCaller(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = executeYandexYQObjectStorageConnectionDelete(ctx, client, d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func executeYandexYQObjectStorageConnectionCreate(
	ctx context.Context,
	client os_conn.ObjectStorageClient,
	d *schema.ResourceData,
	config *Config,
) error {
	connectionName := d.Get(os_conn.AttributeConnectionName).(string)
	serviceAccountID := d.Get(os_conn.AttributeServiceAccountID).(string)
	bucket := d.Get(os_conn.AttributeBucket).(string)
	description := d.Get(os_conn.AttributeDescription).(string)
	visibilityString := d.Get(os_conn.AttributeVisibility).(string)

	auth := parseServiceIDToIAMAuth(serviceAccountID)
	visibility, err := parseVisibilityToAclVisibility(visibilityString)
	if err != nil {
		return err
	}

	req := Ydb_FederatedQuery.CreateConnectionRequest{
		Content: &Ydb_FederatedQuery.ConnectionContent{
			Name: connectionName,
			Acl: &Ydb_FederatedQuery.Acl{
				Visibility: visibility,
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
	client os_conn.ObjectStorageClient,
	d *schema.ResourceData,
	req *Ydb_FederatedQuery.CreateConnectionRequest,
) error {
	res, err := client.CreateStorageConnection(ctx, req)
	if err != nil {
		return err
	}

	d.SetId(res.ConnectionId)

	return nil
}

func executeYandexYQObjectStorageConnectionRead(
	ctx context.Context,
	client os_conn.ObjectStorageClient,
	d *schema.ResourceData,
	_ *Config,
) error {
	connectionID := d.Id()

	req := &Ydb_FederatedQuery.DescribeConnectionRequest{
		ConnectionId: connectionID,
	}

	connectionRes, err := performYandexYQObjectStorageConnectionRead(ctx, client, d, req)
	if err != nil {
		return err
	}

	return flattenYandexYQObjectStorageConnection(d, connectionRes)
}

func performYandexYQObjectStorageConnectionRead(
	ctx context.Context,
	client os_conn.ObjectStorageClient,
	_ *schema.ResourceData,
	req *Ydb_FederatedQuery.DescribeConnectionRequest,
) (*Ydb_FederatedQuery.DescribeConnectionResult, error) {
	res, err := client.DescribeStorageConnection(ctx, req)
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

func flattenYandexYQCommonMeta(
	d *schema.ResourceData,
	meta *Ydb_FederatedQuery.CommonMeta,
) error {
	d.SetId(meta.GetId())

	return nil
}

func flattenYandexYQConnectionContent(
	d *schema.ResourceData,
	content *Ydb_FederatedQuery.ConnectionContent,
) error {
	d.Set(os_conn.AttributeConnectionName, content.GetName())
	d.Set(os_conn.AttributeDescription, content.GetDescription())
	if err := falttenYandexYQConnectionSetting(d, content.GetSetting()); err != nil {
		return err
	}

	if err := flattenYandexYQAcl(d, content.GetAcl()); err != nil {
		return err
	}

	return nil
}

func falttenYandexYQConnectionSetting(
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

func flattenYandexYQAcl(
	d *schema.ResourceData,
	acl *Ydb_FederatedQuery.Acl,
) error {
	visibilityString, err := visibilityToString(acl.GetVisibility())
	if err != nil {
		return err
	}

	d.Set(os_conn.AttributeVisibility, visibilityString)
	return nil
}

func executeYandexYQObjectStorageConnectionUpdate(
	ctx context.Context,
	client os_conn.ObjectStorageClient,
	d *schema.ResourceData,
	config *Config,
) error {
	connectionName := d.Get(os_conn.AttributeConnectionName).(string)
	serviceAccountID := d.Get(os_conn.AttributeServiceAccountID).(string)
	bucket := d.Get(os_conn.AttributeBucket).(string)
	description := d.Get(os_conn.AttributeDescription).(string)
	visibilityString := d.Get(os_conn.AttributeVisibility).(string)

	auth := parseServiceIDToIAMAuth(serviceAccountID)
	visibility, err := parseVisibilityToAclVisibility(visibilityString)
	if err != nil {
		return err
	}

	connectionID := d.Id()

	req := &Ydb_FederatedQuery.ModifyConnectionRequest{
		ConnectionId: connectionID,
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
				Visibility: visibility,
			},
		},
	}

	if err = performYandexYQObjectStorageConnectionUpdate(ctx, client, d, req); err != nil {
		return err
	}

	return executeYandexYQObjectStorageConnectionRead(ctx, client, d, config)
}

func performYandexYQObjectStorageConnectionUpdate(
	ctx context.Context,
	client os_conn.ObjectStorageClient,
	_ *schema.ResourceData,
	req *Ydb_FederatedQuery.ModifyConnectionRequest,
) error {
	_, err := client.ModifyStorageConnection(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func executeYandexYQObjectStorageConnectionDelete(
	ctx context.Context,
	client os_conn.ObjectStorageClient,
	d *schema.ResourceData,
	_ *Config,
) error {
	connectionID := d.Id()

	req := &Ydb_FederatedQuery.DeleteConnectionRequest{
		ConnectionId: connectionID,
	}

	err := performYandexYQObjectStorageConnectionDelete(ctx, client, d, req)
	if err != nil {
		return err
	}

	return nil
}

func performYandexYQObjectStorageConnectionDelete(
	ctx context.Context,
	client os_conn.ObjectStorageClient,
	_ *schema.ResourceData,
	req *Ydb_FederatedQuery.DeleteConnectionRequest,
) error {
	_, err := client.DeleteStorageConnection(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
