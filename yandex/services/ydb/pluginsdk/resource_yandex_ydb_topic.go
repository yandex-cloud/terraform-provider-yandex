package pluginsdk

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"

	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/topic"
)

//nolint:unused
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

func ResourceYandexYDBTopic() *schema.Resource {
	return &schema.Resource{
		Description: "Manage a YDB Topic. For more information, see [the official documentation](https://yandex.cloud/docs/ydb/concepts/#ydb).",

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
		config := meta.(*providerconfig.Config)
		token, err := config.GetIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token}, nil
	}
	return topic.ResourceCreateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTopicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*providerconfig.Config)
		token, err := config.GetIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token}, nil
	}
	return topic.ResourceReadFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTopicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*providerconfig.Config)
		token, err := config.GetIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token}, nil
	}
	return topic.ResourceUpdateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTopicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*providerconfig.Config)
		token, err := config.GetIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token}, nil
	}
	return topic.ResourceDeleteFunc(cb)(ctx, d, meta)
}
