package yandex

import (
	"context"

	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/table/changefeed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceYandexYDBTableChangefeed() *schema.Resource {
	return &schema.Resource{
		Description:   "Yandex Database [table changefeed](https://ydb.tech/en/docs/concepts/cdc), or Change Data Capture (CDC) resource, keeps you informed about changes in a given table. When you add, update, or delete a table row, the CDC mechanism generates a change record where it specifies the primary key of the row and writes it to the topic partition corresponding to this key. A [topic](https://ydb.tech/en/docs/concepts/topic) is an entity for storing unstructured messages and delivering them to multiple subscribers. Basically, a topic is a named set of messages.",
		Schema:        changefeed.ResourceSchema(),
		SchemaVersion: 0,
		CreateContext: resourceYandexYDBTableChangefeedCreate,
		ReadContext:   resourceYandexYDBTableChangefeedRead,
		UpdateContext: resourceYandexYDBTableChangefeedUpdate,
		DeleteContext: resourceYandexYDBTableChangefeedDelete,
		Importer: &schema.ResourceImporter{
			StateContext: changefeed.ResourceImportFunc,
		},
		Timeouts: ydbTimeouts(),
	}
}

func resourceYandexYDBTableChangefeedCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return changefeed.ResourceCreateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableChangefeedRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return changefeed.ResourceReadFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableChangefeedUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return changefeed.ResourceUpdateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableChangefeedDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return changefeed.ResourceDeleteFunc(cb)(ctx, d, meta)
}
