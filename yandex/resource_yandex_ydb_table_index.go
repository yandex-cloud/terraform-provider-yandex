package yandex

import (
	"context"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/table/index"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceYandexYDBTableIndex() *schema.Resource {
	return &schema.Resource{
		Schema:        index.ResourceSchema(),
		SchemaVersion: 0,
		CreateContext: resourceYandexYDBTableIndexCreate,
		ReadContext:   resourceYandexYDBTableIndexRead,
		UpdateContext: resourceYandexYDBTableIndexUpdate,
		DeleteContext: resourceYandexYDBTableIndexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: ydbTimeouts(),
	}
}

func resourceYandexYDBTableIndexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return index.ResourceCreateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableIndexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return index.ResourceReadFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableIndexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return index.ResourceUpdateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableIndexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return index.ResourceDeleteFunc(cb)(ctx, d, meta)
}
