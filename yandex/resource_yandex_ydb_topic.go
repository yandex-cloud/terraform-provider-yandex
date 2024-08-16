package yandex

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/topic"
)

func defaultTimeouts() *schema.ResourceTimeout {
	return &schema.ResourceTimeout{
		Create:  schema.DefaultTimeout(time.Minute * 20),
		Read:    schema.DefaultTimeout(time.Minute * 20),
		Update:  schema.DefaultTimeout(time.Minute * 20),
		Delete:  schema.DefaultTimeout(time.Minute * 20),
		Default: schema.DefaultTimeout(time.Minute * 20),
	}
}

func ydbTimeouts() *schema.ResourceTimeout {
	return &schema.ResourceTimeout{
		Create:  schema.DefaultTimeout(time.Minute * 1),
		Read:    schema.DefaultTimeout(time.Minute * 1),
		Update:  schema.DefaultTimeout(time.Minute * 1),
		Delete:  schema.DefaultTimeout(time.Minute * 1),
		Default: schema.DefaultTimeout(time.Minute * 1),
	}
}

func resourceYandexYDBTopic() *schema.Resource {
	return &schema.Resource{
		Schema:        topic.ResourceSchema(),
		SchemaVersion: 0,
		CreateContext: resourceYandexYDBTopicCreate,
		ReadContext:   resourceYandexYDBTopicRead,
		UpdateContext: resourceYandexYDBTopicUpdate,
		DeleteContext: resourceYandexYDBTopicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: ydbTimeouts(),
	}
}

func resourceYandexYDBTopicCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return topic.ResourceCreateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTopicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return topic.ResourceReadFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTopicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return topic.ResourceUpdateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTopicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*Config)
		token, err := config.sdk.CreateIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token.IamToken}, nil
	}
	return topic.ResourceDeleteFunc(cb)(ctx, d, meta)
}
