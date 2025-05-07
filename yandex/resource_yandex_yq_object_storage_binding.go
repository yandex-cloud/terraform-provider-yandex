package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/client"
	os_binding "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/object_storage_binding"
)

func resourceYandexYQObjectStorageBinding() *schema.Resource {
	return &schema.Resource{
		Schema:        os_binding.ResourceSchema(),
		SchemaVersion: 0,
		CreateContext: resourceYandexYQObjectStorageBindingCreate,
		ReadContext:   resourceYandexYQObjectStorageBindingRead,
		UpdateContext: resourceYandexYQObjectStorageBindingUpdate,
		DeleteContext: resourceYandexYQObjectStorageBindingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: defaultTimeouts(),
	}
}

func resourceYandexYQObjectStorageBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	if err := executeYandexYQObjectStorageBindingCreate(ctx, config.yqSdk.Client(), d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQObjectStorageBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := executeYandexYQObjectStorageBindingRead(ctx, config.yqSdk.Client(), d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQObjectStorageBindingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	if err := executeYandexYQObjectStorageBindingUpdate(ctx, config.yqSdk.Client(), d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexYQObjectStorageBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	if err := executeYandexYQObjectStorageBindingDelete(ctx, config.yqSdk.Client(), d, config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func parseColumns(d *schema.ResourceData) ([]*Ydb.Column, error) {
	columnsRaw := d.Get(os_binding.AttributeColumn)
	if columnsRaw == nil {
		return nil, nil
	}

	raw := columnsRaw.([]interface{})
	columns := make([]*Ydb.Column, 0, len(raw))
	for _, rw := range raw {
		r := rw.(map[string]interface{})
		name := r[os_binding.AttributeColumnName].(string)
		//t := r[os_binding.AttributeColumnType].(string)
		//not_null := r[os_binding.AttributeColumnNotNull].(bool)

		columns = append(columns, &Ydb.Column{
			Name: name,
			Type: &Ydb.Type{
				Type: &Ydb.Type_TypeId{
					TypeId: Ydb.Type_STRING,
				},
			},
		})
	}

	return columns, nil

}

func parseBindingContent(d *schema.ResourceData) (*Ydb_FederatedQuery.BindingContent, error) {
	name := d.Get(os_binding.AttributeName).(string)
	description := d.Get(os_binding.AttributeDescription).(string)
	connectionId := d.Get(os_binding.AttributeConnectionID).(string)
	format := d.Get(os_binding.AttributeFormat).(string)
	compression := d.Get(os_binding.AttributeCompression).(string)
	pathPattern := d.Get(os_binding.AttributePathPattern).(string)

	columns, err := parseColumns(d)
	if err != nil {
		return nil, err
	}

	schema := &Ydb_FederatedQuery.Schema{
		Column: columns,
	}

	return &Ydb_FederatedQuery.BindingContent{
		Name:         name,
		ConnectionId: connectionId,
		Description:  description,
		Setting: &Ydb_FederatedQuery.BindingSetting{
			Binding: &Ydb_FederatedQuery.BindingSetting_ObjectStorage{
				ObjectStorage: &Ydb_FederatedQuery.ObjectStorageBinding{
					Subset: []*Ydb_FederatedQuery.ObjectStorageBinding_Subset{
						{
							Format:      format,
							Compression: compression,
							PathPattern: pathPattern,
							Schema:      schema,
						},
					},
				},
			},
		},
		Acl: &Ydb_FederatedQuery.Acl{
			Visibility: Ydb_FederatedQuery.Acl_SCOPE,
		},
	}, nil
}

func executeYandexYQObjectStorageBindingCreate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	config *Config,
) error {
	bindingContent, err := parseBindingContent(d)
	if err != nil {
		return err
	}

	req := Ydb_FederatedQuery.CreateBindingRequest{
		Content: bindingContent,
	}

	if err := performYandexYQObjectStorageBindingCreate(ctx, client, d, &req); err != nil {
		return err
	}

	return executeYandexYQObjectStorageBindingRead(ctx, client, d, config)
}

func performYandexYQObjectStorageBindingCreate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	req *Ydb_FederatedQuery.CreateBindingRequest,
) error {
	res, err := client.CreateBinding(ctx, req)
	if err != nil {
		return err
	}

	d.SetId(res.BindingId)

	return nil
}

func executeYandexYQObjectStorageBindingRead(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	_ *Config,
) error {
	id := d.Id()

	req := &Ydb_FederatedQuery.DescribeBindingRequest{
		BindingId: id,
	}

	connectionRes, err := client.DescribeBinding(ctx, req)
	if err != nil {
		return err
	}

	return flattenYandexYQObjectStorageBinding(d, connectionRes)
}

func flattenYandexYQObjectStorageBinding(
	d *schema.ResourceData,
	connectionRes *Ydb_FederatedQuery.DescribeBindingResult,
) error {
	if connectionRes == nil {
		d.SetId("")
		return nil
	}

	connection := connectionRes.GetBinding()

	if err := flattenYandexYQBindingContent(d, connection.GetContent()); err != nil {
		return err
	}

	if err := flattenYandexYQCommonMeta(d, connection.GetMeta()); err != nil {
		return err
	}

	return nil
}

func flattenYandexYQBindingContent(
	d *schema.ResourceData,
	content *Ydb_FederatedQuery.BindingContent,
) error {
	d.Set(os_binding.AttributeName, content.GetName())
	d.Set(os_binding.AttributeDescription, content.GetDescription())
	if err := flattenYandexYQBindingSetting(d, content.GetSetting()); err != nil {
		return err
	}

	return nil
}

func flattenYandexYQBindingSetting(
	d *schema.ResourceData,
	setting *Ydb_FederatedQuery.BindingSetting,
) error {
	return nil
}

func executeYandexYQObjectStorageBindingUpdate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	config *Config,
) error {
	bindingContent, err := parseBindingContent(d)
	if err != nil {
		return err
	}

	id := d.Id()

	req := &Ydb_FederatedQuery.ModifyBindingRequest{
		BindingId: id,
		Content:   bindingContent,
	}

	_, err = client.ModifyBinding(ctx, req)
	if err != nil {
		return err
	}

	return executeYandexYQObjectStorageBindingRead(ctx, client, d, config)
}

func executeYandexYQObjectStorageBindingDelete(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	_ *Config,
) error {
	id := d.Id()

	req := &Ydb_FederatedQuery.DeleteBindingRequest{
		BindingId: id,
	}

	_, err := client.DeleteBinding(ctx, req)
	return err
}
