package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/client"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type ConnectionStrategy interface {
	FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.ConnectionSetting) error
	ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.ConnectionSetting, error)
}

func resourceYandexYQBaseConnection(strategy ConnectionStrategy, s map[string]*schema.Schema) *schema.Resource {
	return &schema.Resource{
		Schema:        s,
		SchemaVersion: 0,
		CreateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
			return resourceYandexYQBaseConnectionCreate(ctx, d, meta, strategy)
		},
		ReadContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
			return resourceYandexYQBaseConnectionRead(ctx, d, meta, strategy)
		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
			return resourceYandexYQBaseConnectionUpdate(ctx, d, meta, strategy)
		},
		DeleteContext: resourceYandexYQBaseConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: defaultTimeouts(),
	}
}

func resourceYandexYQBaseConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}, strategy ConnectionStrategy) diag.Diagnostics {
	config := meta.(*Config)
	err := executeYandexYQBaseConnectionCreate(ctx, config.yqSdk.Client(), d, config, strategy)
	return diag.FromErr(err)
}

func resourceYandexYQBaseConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}, strategy ConnectionStrategy) diag.Diagnostics {
	config := meta.(*Config)
	err := executeYandexYQBaseConnectionRead(ctx, config.yqSdk.Client(), d, config, strategy)
	return diag.FromErr(err)
}

func resourceYandexYQBaseConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, strategy ConnectionStrategy) diag.Diagnostics {
	config := meta.(*Config)
	err := executeYandexYQBaseConnectionUpdate(ctx, config.yqSdk.Client(), d, config, strategy)
	return diag.FromErr(err)
}

func resourceYandexYQBaseConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	err := executeYandexYQBaseConnectionDelete(ctx, config.yqSdk.Client(), d, config)
	return diag.FromErr(err)
}

func executeYandexYQBaseConnectionCreate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	config *Config,
	strategy ConnectionStrategy,
) error {
	connectionName := d.Get(AttributeName).(string)
	description := d.Get(AttributeDescription).(string)

	setting, err := strategy.ExpandSetting(d)
	if err != nil {
		return err
	}

	req := Ydb_FederatedQuery.CreateConnectionRequest{
		Content: &Ydb_FederatedQuery.ConnectionContent{
			Name:        connectionName,
			Description: description,
			Setting:     setting,
			Acl: &Ydb_FederatedQuery.Acl{
				Visibility: Ydb_FederatedQuery.Acl_SCOPE,
			},
		},
	}

	if err := performYandexYQConnectionCreate(ctx, client, d, &req); err != nil {
		return err
	}

	return executeYandexYQBaseConnectionRead(ctx, client, d, config, strategy)
}

func performYandexYQConnectionCreate(
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

func executeYandexYQBaseConnectionRead(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	_ *Config,
	strategy ConnectionStrategy,
) error {
	id := d.Id()

	req := &Ydb_FederatedQuery.DescribeConnectionRequest{
		ConnectionId: id,
	}

	connectionRes, err := client.DescribeConnection(ctx, req)
	if err != nil {
		return err
	}

	return flattenYandexYQBaseConnection(d, connectionRes, strategy)
}

func flattenYandexYQBaseConnection(
	d *schema.ResourceData,
	connectionRes *Ydb_FederatedQuery.DescribeConnectionResult,
	strategy ConnectionStrategy,
) error {
	if connectionRes == nil {
		d.SetId("")
		return nil
	}

	connection := connectionRes.GetConnection()

	if err := flattenYandexYQConnectionContent(d, connection.GetContent(), strategy); err != nil {
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
	strategy ConnectionStrategy,
) error {
	d.Set(AttributeName, content.GetName())
	d.Set(AttributeDescription, content.GetDescription())
	if err := strategy.FlattenSetting(d, content.GetSetting()); err != nil {
		return err
	}

	return nil
}

func executeYandexYQBaseConnectionUpdate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	config *Config,
	strategy ConnectionStrategy,
) error {
	connectionName := d.Get(AttributeName).(string)
	description := d.Get(AttributeDescription).(string)

	id := d.Id()
	setting, err := strategy.ExpandSetting(d)
	if err != nil {
		return err
	}

	req := &Ydb_FederatedQuery.ModifyConnectionRequest{
		ConnectionId: id,
		Content: &Ydb_FederatedQuery.ConnectionContent{
			Name:        connectionName,
			Description: description,
			Setting:     setting,
			Acl: &Ydb_FederatedQuery.Acl{
				Visibility: Ydb_FederatedQuery.Acl_SCOPE,
			},
		},
	}

	if _, err := client.ModifyConnection(ctx, req); err != nil {
		return err
	}

	return executeYandexYQBaseConnectionRead(ctx, client, d, config, strategy)
}

func executeYandexYQBaseConnectionDelete(
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
