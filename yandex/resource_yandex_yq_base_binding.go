package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/client"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type BindingStrategy interface {
	FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.BindingSetting) error
	ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.BindingSetting, error)
}

func resourceYandexYQBaseBinding(strategy BindingStrategy, description string, s map[string]*schema.Schema) *schema.Resource {
	return &schema.Resource{
		Description:   description,
		Schema:        s,
		SchemaVersion: 0,
		CreateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
			return resourceYandexYQBaseBindingCreate(ctx, d, meta, strategy)
		},
		ReadContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
			return resourceYandexYQBaseBindingRead(ctx, d, meta, strategy)
		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
			return resourceYandexYQBaseBindingUpdate(ctx, d, meta, strategy)
		},
		DeleteContext: resourceYandexYQBaseBindingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: defaultTimeouts(),
	}
}

func resourceYandexYQBaseBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}, strategy BindingStrategy) diag.Diagnostics {
	config := meta.(*Config)
	err := executeYandexYQBaseBindingCreate(ctx, config.yqSdk.Client(), d, config, strategy)
	return diag.FromErr(err)
}

func resourceYandexYQBaseBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}, strategy BindingStrategy) diag.Diagnostics {
	config := meta.(*Config)
	err := executeYandexYQBaseBindingRead(ctx, config.yqSdk.Client(), d, config, strategy)
	return diag.FromErr(err)
}

func resourceYandexYQBaseBindingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, strategy BindingStrategy) diag.Diagnostics {
	config := meta.(*Config)
	err := executeYandexYQBaseBindingUpdate(ctx, config.yqSdk.Client(), d, config, strategy)
	return diag.FromErr(err)
}

func resourceYandexYQBaseBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	err := executeYandexYQBaseBindingDelete(ctx, config.yqSdk.Client(), d, config)
	return diag.FromErr(err)
}

func parseBindingContent(d *schema.ResourceData, strategy BindingStrategy) (*Ydb_FederatedQuery.BindingContent, error) {
	name := d.Get(AttributeName).(string)
	description := d.Get(AttributeDescription).(string)
	connectionId := d.Get(AttributeConnectionID).(string)

	setting, err := strategy.ExpandSetting(d)
	if err != nil {
		return nil, err
	}

	return &Ydb_FederatedQuery.BindingContent{
		Name:         name,
		ConnectionId: connectionId,
		Description:  description,
		Setting:      setting,
		Acl: &Ydb_FederatedQuery.Acl{
			Visibility: Ydb_FederatedQuery.Acl_SCOPE,
		},
	}, nil
}

func executeYandexYQBaseBindingCreate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	config *Config,
	strategy BindingStrategy,
) error {
	bindingContent, err := parseBindingContent(d, strategy)
	if err != nil {
		return err
	}

	req := Ydb_FederatedQuery.CreateBindingRequest{
		Content: bindingContent,
	}

	if err := performYandexYQBindingCreate(ctx, client, d, &req); err != nil {
		return err
	}

	return executeYandexYQBaseBindingRead(ctx, client, d, config, strategy)
}

func performYandexYQBindingCreate(
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

func executeYandexYQBaseBindingRead(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	_ *Config,
	strategy BindingStrategy,
) error {
	id := d.Id()

	req := &Ydb_FederatedQuery.DescribeBindingRequest{
		BindingId: id,
	}

	connectionRes, err := client.DescribeBinding(ctx, req)
	if err != nil {
		return err
	}

	return flattenYandexYQBaseBinding(d, connectionRes, strategy)
}

func flattenYandexYQBaseBinding(
	d *schema.ResourceData,
	connectionRes *Ydb_FederatedQuery.DescribeBindingResult,
	strategy BindingStrategy,
) error {
	if connectionRes == nil {
		d.SetId("")
		return nil
	}

	connection := connectionRes.GetBinding()

	if err := flattenYandexYQBaseBindingContent(d, connection.GetContent(), strategy); err != nil {
		return err
	}

	if err := flattenYandexYQCommonMeta(d, connection.GetMeta()); err != nil {
		return err
	}

	return nil
}

func flattenYandexYQBaseBindingContent(
	d *schema.ResourceData,
	content *Ydb_FederatedQuery.BindingContent,
	strategy BindingStrategy,
) error {
	d.Set(AttributeConnectionID, content.GetConnectionId())
	d.Set(AttributeName, content.GetName())
	d.Set(AttributeDescription, content.GetDescription())
	return strategy.FlattenSetting(d, content.GetSetting())
}

func executeYandexYQBaseBindingUpdate(
	ctx context.Context,
	client client.YQClient,
	d *schema.ResourceData,
	config *Config,
	strategy BindingStrategy,
) error {
	bindingContent, err := parseBindingContent(d, strategy)
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

	return executeYandexYQBaseBindingRead(ctx, client, d, config, strategy)
}

func executeYandexYQBaseBindingDelete(
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
