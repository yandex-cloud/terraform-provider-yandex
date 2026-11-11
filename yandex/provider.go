package yandex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/yandex-cloud/terraform-provider-yandex/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/terraform-provider-yandex/version"
)

type ProviderOptions struct {
	ConfigurationContext context.Context
	IgnoreFolderID       bool
}

func NewPluginSDKProvider() (*schema.Provider, error) {
	return NewPluginSDKProviderWithOptions(ProviderOptions{})
}

func NewPluginSDKProviderWithOptions(options ProviderOptions) (*schema.Provider, error) {
	return buildPluginSDKProvider(options.IgnoreFolderID, options)
}

func buildPluginSDKProvider(emptyFolder bool, options ProviderOptions) (*schema.Provider, error) {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["endpoint"],
			},
			"yq_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["yq_endpoint"],
			},
			"datalens_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["datalens_endpoint"],
			},
			"folder_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["folder_id"],
			},
			"cloud_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["cloud_id"],
			},
			"organization_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["organization_id"],
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["region_id"],
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["zone"],
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: common.Descriptions["token"],
			},
			"service_account_key_file": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   common.Descriptions["service_account_key_file"],
				ConflictsWith: []string{"token"},
				ValidateFunc:  validateSAKey,
			},
			"storage_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["storage_endpoint"],
			},
			"storage_access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["storage_access_key"],
			},
			"storage_secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: common.Descriptions["storage_secret_key"],
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: common.Descriptions["insecure"],
			},
			"plaintext": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: common.Descriptions["plaintext"],
			},
			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: common.Descriptions["max_retries"],
			},
			"ymq_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["ymq_endpoint"],
			},
			"ymq_access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["ymq_access_key"],
			},
			"ymq_secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: common.Descriptions["ymq_secret_key"],
			},
			"shared_credentials_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["shared_credentials_file"],
			},
			"profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.Descriptions["profile"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"yandex_client_config": dataSourceYandexClientConfig(),
		},
	}

	if err := registerServices(provider); err != nil {
		return nil, err
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		if options.ConfigurationContext != nil {
			return providerConfigure(options.ConfigurationContext, d, provider, emptyFolder, false)
		}
		return providerConfigure(ctx, d, provider, emptyFolder, false)
	}

	return provider, nil
}

func setToDefaultIfNeeded(field string, osEnvName string, defaultVal string) string {
	if len(field) != 0 {
		return field
	}
	field = os.Getenv(osEnvName)
	if len(field) == 0 {
		field = defaultVal
	}
	return field
}

func setToDefaultBoolIfNeeded(osEnvName string, defaultVal bool) bool {
	field := os.Getenv(osEnvName)
	v, err := strconv.ParseBool(field)
	if err != nil {
		return v
	} else {
		return defaultVal
	}
}

// testConfig is used to avoid using StopContext duo to tests are run in parallel and context is cancelled randomly in tests
// there is same following issue https://github.com/hashicorp/terraform-plugin-sdk/issues/966
func providerConfigure(ctx context.Context, d *schema.ResourceData, p *schema.Provider, emptyFolder bool, testConfig bool) (interface{}, diag.Diagnostics) {
	config := Config{
		Endpoint:                       setToDefaultIfNeeded(d.Get("endpoint").(string), "YC_ENDPOINT", common.DefaultEndpoint),
		FolderID:                       setToDefaultIfNeeded(d.Get("folder_id").(string), "YC_FOLDER_ID", ""),
		CloudID:                        setToDefaultIfNeeded(d.Get("cloud_id").(string), "YC_CLOUD_ID", ""),
		OrganizationID:                 setToDefaultIfNeeded(d.Get("organization_id").(string), "YC_ORGANIZATION_ID", ""),
		Region:                         setToDefaultIfNeeded(d.Get("region_id").(string), "YC_REGION", common.DefaultRegion),
		Zone:                           setToDefaultIfNeeded(d.Get("zone").(string), "YC_ZONE", ""),
		Token:                          setToDefaultIfNeeded(d.Get("token").(string), "YC_TOKEN", ""),
		ServiceAccountKeyFileOrContent: setToDefaultIfNeeded(d.Get("service_account_key_file").(string), "YC_SERVICE_ACCOUNT_KEY_FILE", ""),
		StorageEndpoint:                setToDefaultIfNeeded(d.Get("storage_endpoint").(string), "YC_STORAGE_ENDPOINT_URL", common.DefaultStorageEndpoint),
		StorageAccessKey:               setToDefaultIfNeeded(d.Get("storage_access_key").(string), "YC_STORAGE_ACCESS_KEY", ""),
		StorageSecretKey:               setToDefaultIfNeeded(d.Get("storage_secret_key").(string), "YC_STORAGE_SECRET_KEY", ""),
		YMQEndpoint:                    setToDefaultIfNeeded(d.Get("ymq_endpoint").(string), "YC_MESSAGE_QUEUE_ENDPOINT", common.DefaultYMQEndpoint),
		YMQAccessKey:                   setToDefaultIfNeeded(d.Get("ymq_access_key").(string), "YC_MESSAGE_QUEUE_ACCESS_KEY", ""),
		YMQSecretKey:                   setToDefaultIfNeeded(d.Get("ymq_secret_key").(string), "YC_MESSAGE_QUEUE_SECRET_KEY", ""),

		Plaintext:             setToDefaultBoolIfNeeded("YC_PLAINTEXT", d.Get("plaintext").(bool)),
		Insecure:              setToDefaultBoolIfNeeded("YC_INSECURE", d.Get("insecure").(bool)),
		MaxRetries:            d.Get("max_retries").(int),
		SharedCredentialsFile: d.Get("shared_credentials_file").(string),
		Profile:               d.Get("profile").(string),
	}

	config.SetUserAgent(p.UserAgent("terraform-provider-yandex", version.ProviderVersion))

	if len(config.Profile) == 0 {
		config.Profile = "default"
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = common.DefaultMaxRetries
	}

	if emptyFolder {
		config.FolderID = ""
	}
	stopCtx, ok := schema.StopContext(ctx)
	if !ok {
		stopCtx = ctx
	}
	if testConfig {
		stopCtx = context.Background()
	}

	terraformVersion := p.TerraformVersion
	if terraformVersion == "" {
		// Terraform 0.12 introduced this field to the protocol
		// We can therefore assume that if it's missing it's 0.10 or 0.11
		terraformVersion = "0.11+compatible"
	}

	if err := config.InitAndValidate(stopCtx, terraformVersion, false); err != nil {
		return nil, diag.FromErr(err)
	}

	return &config, nil

}

func validateSAKey(v interface{}, k string) (warnings []string, errors []error) {
	if v == nil || v.(string) == "" {
		return
	}

	saKey := v.(string)
	// if this is a path to file and we can stat it, assume it's ok
	if _, err := os.Stat(saKey); err == nil {
		return
	}

	// else check for a valid json data value
	var f map[string]interface{}
	if err := json.Unmarshal([]byte(saKey), &f); err != nil {
		errors = append(errors, fmt.Errorf("JSON in %q are not valid: %s", saKey, err))
	}
	return
}
