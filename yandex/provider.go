package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const defaultEndpoint = "api.cloud.yandex.net:443"

// Global MutexKV
var mutexKV = mutexkv.NewMutexKV()

func Provider() terraform.ResourceProvider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_ENDPOINT", defaultEndpoint),
				Description: descriptions["endpoint"],
			},
			"folder_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_FOLDER_ID", nil),
				Description: descriptions["folder_id"],
			},
			"cloud_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_CLOUD_ID", nil),
				Description: descriptions["cloud_id"],
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_ZONE", nil),
				Description: descriptions["zone"],
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_TOKEN", nil),
				Description: descriptions["token"],
			},
			"service_account_key_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_SERVICE_ACCOUNT_KEY_FILE", nil),
				Description: descriptions["service_account_key_file"],
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_INSECURE", false),
				Description: descriptions["insecure"],
			},
			"plaintext": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_PLAINTEXT", false),
				Description: descriptions["plaintext"],
			},
			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5,
				Description: descriptions["max_retries"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"yandex_container_registry":       dataSourceYandexContainerRegistry(),
			"yandex_compute_disk":             dataSourceYandexComputeDisk(),
			"yandex_compute_image":            dataSourceYandexComputeImage(),
			"yandex_compute_instance":         dataSourceYandexComputeInstance(),
			"yandex_compute_instance_group":   dataSourceYandexComputeInstanceGroup(),
			"yandex_compute_snapshot":         dataSourceYandexComputeSnapshot(),
			"yandex_iam_policy":               dataSourceYandexIAMPolicy(),
			"yandex_iam_role":                 dataSourceYandexIAMRole(),
			"yandex_iam_service_account":      dataSourceYandexIAMServiceAccount(),
			"yandex_iam_user":                 dataSourceYandexIAMUser(),
			"yandex_resourcemanager_cloud":    dataSourceYandexResourceManagerCloud(),
			"yandex_resourcemanager_folder":   dataSourceYandexResourceManagerFolder(),
			"yandex_vpc_network":              dataSourceYandexVPCNetwork(),
			"yandex_vpc_route_table":          dataSourceYandexVPCRouteTable(),
			"yandex_vpc_subnet":               dataSourceYandexVPCSubnet(),
			"yandex_lb_network_load_balancer": dataSourceYandexLBNetworkLoadBalancer(),
			"yandex_lb_target_group":          dataSourceYandexLBTargetGroup(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"yandex_container_registry":                    resourceYandexContainerRegistry(),
			"yandex_compute_disk":                          resourceYandexComputeDisk(),
			"yandex_compute_image":                         resourceYandexComputeImage(),
			"yandex_compute_instance":                      resourceYandexComputeInstance(),
			"yandex_compute_instance_group":                resourceYandexComputeInstanceGroup(),
			"yandex_compute_snapshot":                      resourceYandexComputeSnapshot(),
			"yandex_iam_service_account":                   resourceYandexIAMServiceAccount(),
			"yandex_iam_service_account_api_key":           resourceYandexIAMServiceAccountAPIKey(),
			"yandex_iam_service_account_iam_binding":       resourceYandexIAMServiceAccountIAMBinding(),
			"yandex_iam_service_account_iam_member":        resourceYandexIAMServiceAccountIAMMember(),
			"yandex_iam_service_account_iam_policy":        resourceYandexIAMServiceAccountIAMPolicy(),
			"yandex_iam_service_account_key":               resourceYandexIAMServiceAccountKey(),
			"yandex_iam_service_account_static_access_key": resourceYandexIAMServiceAccountStaticAccessKey(),
			"yandex_resourcemanager_cloud_iam_binding":     resourceYandexResourceManagerCloudIAMBinding(),
			"yandex_resourcemanager_cloud_iam_member":      resourceYandexResourceManagerCloudIAMMember(),
			"yandex_resourcemanager_folder_iam_binding":    resourceYandexResourceManagerFolderIAMBinding(),
			"yandex_resourcemanager_folder_iam_member":     resourceYandexResourceManagerFolderIAMMember(),
			"yandex_resourcemanager_folder_iam_policy":     resourceYandexResourceManagerFolderIAMPolicy(),
			"yandex_vpc_network":                           resourceYandexVPCNetwork(),
			"yandex_vpc_route_table":                       resourceYandexVPCRouteTable(),
			"yandex_vpc_subnet":                            resourceYandexVPCSubnet(),
			"yandex_lb_network_load_balancer":              resourceYandexLBNetworkLoadBalancer(),
			"yandex_lb_target_group":                       resourceYandexLBTargetGroup(),
		},
	}
	provider.ConfigureFunc = providerConfigure(provider)

	return provider
}

var descriptions = map[string]string{
	"endpoint": "The API endpoint for Yandex.Cloud SDK client.",

	"folder_id": "The default folder ID where resources will be placed.",

	"cloud_id": "ID of Yandex.Cloud tenant.",

	"zone": "The zone where operations will take place. Examples\n" +
		"are ru-central1-a, ru-central2-c, etc.",

	"token": "The access token for API operations.",

	"service_account_key_file": "Path to file with Yandex.Cloud Service Account key.",

	"insecure": "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted," +
		"default value is `false`.",

	"plaintext": "Disable use of TLS. Default value is `false`.",

	"max_retries": "The maximum number of times an API request is being executed. \n" +
		"If the API request still fails, an error is thrown.",
}

func providerConfigure(provider *schema.Provider) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		config := Config{
			Token:                 d.Get("token").(string),
			ServiceAccountKeyFile: d.Get("service_account_key_file").(string),
			Zone:                  d.Get("zone").(string),
			FolderID:              d.Get("folder_id").(string),
			CloudID:               d.Get("cloud_id").(string),
			Endpoint:              d.Get("endpoint").(string),
			Plaintext:             d.Get("plaintext").(bool),
			Insecure:              d.Get("insecure").(bool),
			MaxRetries:            d.Get("max_retries").(int),
		}

		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}

		if err := config.initAndValidate(provider.StopContext(), terraformVersion); err != nil {
			return nil, err
		}

		return &config, nil
	}
}
