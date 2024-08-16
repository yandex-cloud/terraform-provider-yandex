package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/table"
)

func resourceYandexYDBTable() *schema.Resource {
	return &schema.Resource{
		Schema:        table.ResourceSchema(),
		SchemaVersion: 0,
		CreateContext: resourceYandexYDBTableCreate,
		ReadContext:   resourceYandexYDBTableRead,
		UpdateContext: resourceYandexYDBTableUpdate,
		DeleteContext: resourceYandexYDBTableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: ydbTimeouts(),
	}
}

func resourceYandexYDBTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return table.ResourceCreateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return table.ResourceReadFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return table.ResourceUpdateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return table.ResourceDeleteFunc(cb)(ctx, d, meta)
}
