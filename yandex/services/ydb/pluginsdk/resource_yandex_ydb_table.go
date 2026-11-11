package pluginsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"

	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/table"
)

func ResourceYandexYDBTable() *schema.Resource {
	return &schema.Resource{
		Description: "Yandex Database table.",

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
		config := meta.(*providerconfig.Config)
		token, err := config.GetIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token}, nil
	}
	return table.ResourceCreateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*providerconfig.Config)
		token, err := config.GetIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token}, nil
	}
	return table.ResourceReadFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*providerconfig.Config)
		token, err := config.GetIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token}, nil
	}
	return table.ResourceUpdateFunc(cb)(ctx, d, meta)
}

func resourceYandexYDBTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cb := func(ctx context.Context) (auth.YdbCredentials, error) {
		config := meta.(*providerconfig.Config)
		token, err := config.GetIAMToken(ctx)
		if err != nil {
			return auth.YdbCredentials{}, err
		}
		return auth.YdbCredentials{Token: token}, nil
	}
	return table.ResourceDeleteFunc(cb)(ctx, d, meta)
}
