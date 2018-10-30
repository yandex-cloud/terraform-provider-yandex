package yandex

import (
	"github.com/hashicorp/terraform/helper/mutexkv"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

const defaultEnpoint = "api.cloud.yandex.net:443"

// Global MutexKV
var mutexKV = mutexkv.NewMutexKV()

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("YC_ENDPOINT", defaultEnpoint),
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
		},

		DataSourcesMap: map[string]*schema.Resource{
			"yandex_compute_disk":           dataSourceYandexComputeDisk(),
			"yandex_compute_image":          dataSourceYandexComputeImage(),
			"yandex_compute_instance":       dataSourceYandexComputeInstance(),
			"yandex_compute_snapshot":       dataSourceYandexComputeSnapshot(),
			"yandex_iam_policy":             dataSourceYandexIAMPolicy(),
			"yandex_iam_role":               dataSourceYandexIAMRole(),
			"yandex_iam_service_account":    dataSourceYandexIAMServiceAccount(),
			"yandex_iam_user":               dataSourceYandexIAMUser(),
			"yandex_resourcemanager_cloud":  dataSourceYandexResourceManagerCloud(),
			"yandex_resourcemanager_folder": dataSourceYandexResourceManagerFolder(),
			"yandex_vpc_network":            dataSourceYandexVPCNetwork(),
			"yandex_vpc_subnet":             dataSourceYandexVPCSubnet(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"yandex_compute_disk":                       resourceYandexComputeDisk(),
			"yandex_compute_image":                      resourceYandexComputeImage(),
			"yandex_compute_instance":                   resourceYandexComputeInstance(),
			"yandex_compute_snapshot":                   resourceYandexComputeSnapshot(),
			"yandex_iam_service_account":                resourceYandexIAMServiceAccount(),
			"yandex_iam_service_account_iam_binding":    ResourceIamBindingWithImport(IamServiceAccountSchema, NewServiceAccountIamUpdater, ServiceAccountIDParseFunc),
			"yandex_iam_service_account_iam_member":     ResourceIamMemberWithImport(IamServiceAccountSchema, NewServiceAccountIamUpdater, ServiceAccountIDParseFunc),
			"yandex_iam_service_account_iam_policy":     ResourceIamPolicyWithImport(IamServiceAccountSchema, NewServiceAccountIamUpdater, ServiceAccountIDParseFunc),
			"yandex_iam_service_account_key":            resourceYandexIAMServiceAccountKey(),
			"yandex_resourcemanager_cloud_iam_binding":  ResourceIamBindingWithImport(IamCloudSchema, NewCloudIamUpdater, CloudIDParseFunc),
			"yandex_resourcemanager_cloud_iam_member":   ResourceIamMemberWithImport(IamCloudSchema, NewCloudIamUpdater, CloudIDParseFunc),
			"yandex_resourcemanager_cloud_iam_policy":   ResourceIamPolicyWithImport(IamCloudSchema, NewCloudIamUpdater, CloudIDParseFunc),
			"yandex_resourcemanager_folder_iam_binding": ResourceIamBindingWithImport(IamFolderSchema, NewFolderIamUpdater, FolderIDParseFunc),
			"yandex_resourcemanager_folder_iam_member":  ResourceIamMemberWithImport(IamFolderSchema, NewFolderIamUpdater, FolderIDParseFunc),
			"yandex_resourcemanager_folder_iam_policy":  ResourceIamPolicyWithImport(IamFolderSchema, NewFolderIamUpdater, FolderIDParseFunc),
			"yandex_vpc_network":                        resourceYandexVPCNetwork(),
			"yandex_vpc_subnet":                         resourceYandexVPCSubnet(),
		},

		ConfigureFunc: providerConfigure,
	}
}

var descriptions = map[string]string{
	"endpoint": "The API endpoint for Yandex Cloud SDK client",

	"folder_id": "The default folder ID where resources will be placed",

	"cloud_id": "ID of Yandex Cloud tenant",

	"zone": "The zone where operations will take place. Examples\n" +
		"are ru-central1-a, ru-central2-c, etc.",

	"token": "The access token for API operations.",

	"insecure": "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted," +
		"default value is `false`",

	"plaintext": "Disable use of TLS. Default value is `false`",
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Token:     d.Get("token").(string),
		Zone:      d.Get("zone").(string),
		FolderID:  d.Get("folder_id").(string),
		CloudID:   d.Get("cloud_id").(string),
		Endpoint:  d.Get("endpoint").(string),
		Plaintext: d.Get("plaintext").(bool),
		Insecure:  d.Get("insecure").(bool),
	}

	if err := config.initAndValidate(); err != nil {
		return nil, err
	}

	return &config, nil
}
