package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/providervalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/airflow_cluster"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/billing_cloud_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/compute_disk_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/compute_disk_placement_group_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/compute_filesystem_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/compute_gpu_cluster_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/compute_image_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/compute_instance_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/compute_placement_group_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/compute_snapshot_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/compute_snapshot_schedule_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/datasphere_community"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/datasphere_community_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/datasphere_project"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/datasphere_project_iam_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/kubernetes_marketplace_helm_release"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_mongodb_database"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_mongodb_user"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_postgresql_cluster_beta"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/vpc_security_group_rule"
	// "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/vpc_security_group"
)

type saKeyValidator struct{}

func (v saKeyValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Validate Service Account Key")
}

func (v saKeyValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Validate Service Account Key")
}

func (v saKeyValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	saKey := req.ConfigValue.ValueString()
	if len(saKey) == 0 {
		return
	}
	if _, err := os.Stat(saKey); err == nil {
		return
	}
	var _f map[string]interface{}
	if err := json.Unmarshal([]byte(saKey), &_f); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid SA Key",
			fmt.Sprintf("JSON in %q are not valid: %s", saKey, err),
		)
	}
}

type Provider struct {
	emptyFolder bool
	config      provider_config.Config
}

func NewFrameworkProvider() provider.Provider {
	return &Provider{}
}

func (p *Provider) ConfigValidators(ctx context.Context) []provider.ConfigValidator {
	return []provider.ConfigValidator{
		providervalidator.Conflicting(
			path.MatchRoot("token"),
			path.MatchRoot("service_account_key_file"),
		),
	}
}

func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "yandex"
}

func (p *Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["endpoint"],
			},
			"folder_id": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["folder_id"],
			},
			"cloud_id": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["cloud_id"],
			},
			"organization_id": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["organization_id"],
			},
			"region_id": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["region_id"],
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["zone"],
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: common.Descriptions["token"],
			},
			"service_account_key_file": schema.StringAttribute{ // TODO: finish
				Optional:    true,
				Description: common.Descriptions["service_account_key_file"],
				Validators: []validator.String{
					saKeyValidator{},
				},
			},
			"storage_endpoint": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["storage_endpoint"],
			},
			"storage_access_key": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["storage_access_key"],
			},
			"storage_secret_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: common.Descriptions["storage_secret_key"],
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: common.Descriptions["insecure"],
			},
			"plaintext": schema.BoolAttribute{
				Optional:    true,
				Description: common.Descriptions["plaintext"],
			},
			"max_retries": schema.Int64Attribute{
				Optional:    true,
				Description: common.Descriptions["max_retries"],
			},
			"ymq_endpoint": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["ymq_endpoint"],
			},
			"ymq_access_key": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["ymq_access_key"],
			},
			"ymq_secret_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: common.Descriptions["ymq_secret_key"],
			},
			"shared_credentials_file": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["shared_credentials_file"],
			},
			"profile": schema.StringAttribute{
				Optional:    true,
				Description: common.Descriptions["profile"],
			},
		},
	}
}

func setToDefaultIfNeeded(field types.String, osEnvName string, defaultVal string) types.String {
	if len(field.ValueString()) != 0 {
		return field
	}
	field = types.StringValue(os.Getenv(osEnvName))
	if len(field.ValueString()) == 0 {
		field = types.StringValue(defaultVal)
	}
	return field
}

func setToDefaultBoolIfNeeded(field types.Bool, osEnvName string, defaultVal bool) types.Bool {
	if field.IsUnknown() || field.IsNull() {
		env := os.Getenv(osEnvName)
		v, err := strconv.ParseBool(env)
		if err != nil {
			return types.BoolValue(v)
		}
		return types.BoolValue(defaultVal)
	}
	return field
}

func setDefaults(config provider_config.State) provider_config.State {
	config.Endpoint = setToDefaultIfNeeded(config.Endpoint, "YC_ENDPOINT", common.DefaultEndpoint)
	config.FolderID = setToDefaultIfNeeded(config.FolderID, "YC_FOLDER_ID", "")
	config.CloudID = setToDefaultIfNeeded(config.CloudID, "YC_CLOUD_ID", "")
	config.OrganizationID = setToDefaultIfNeeded(config.OrganizationID, "YC_ORGANIZATION_ID", "")
	config.Region = setToDefaultIfNeeded(config.Region, "YC_REGION", common.DefaultRegion)
	config.Zone = setToDefaultIfNeeded(config.Zone, "YC_ZONE", "")
	config.Token = setToDefaultIfNeeded(config.Token, "YC_TOKEN", "")
	config.ServiceAccountKeyFileOrContent = setToDefaultIfNeeded(config.ServiceAccountKeyFileOrContent, "YC_SERVICE_ACCOUNT_KEY_FILE", "")
	config.StorageEndpoint = setToDefaultIfNeeded(config.StorageEndpoint, "YC_STORAGE_ENDPOINT_URL", common.DefaultStorageEndpoint)
	config.StorageAccessKey = setToDefaultIfNeeded(config.StorageAccessKey, "YC_STORAGE_ACCESS_KEY", "")
	config.StorageSecretKey = setToDefaultIfNeeded(config.StorageSecretKey, "YC_STORAGE_SECRET_KEY", "")
	config.YMQEndpoint = setToDefaultIfNeeded(config.YMQEndpoint, "YC_MESSAGE_QUEUE_ENDPOINT", common.DefaultYMQEndpoint)
	config.YMQAccessKey = setToDefaultIfNeeded(config.YMQAccessKey, "YC_MESSAGE_QUEUE_ACCESS_KEY", "")
	config.YMQSecretKey = setToDefaultIfNeeded(config.YMQSecretKey, "YC_MESSAGE_QUEUE_SECRET_KEY", "")

	config.Insecure = setToDefaultBoolIfNeeded(config.Insecure, "YC_INSECURE", false)
	config.Plaintext = setToDefaultBoolIfNeeded(config.Plaintext, "YC_PLAINTEXT", false)

	if config.MaxRetries.IsUnknown() || config.MaxRetries.IsNull() {
		config.MaxRetries = types.Int64Value(common.DefaultMaxRetries)
	}
	if config.Profile.IsUnknown() || config.Profile.IsNull() {
		config.Profile = types.StringValue("default")
	}

	return config
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Unmarshal config
	p.config = provider_config.Config{}
	resp.Diagnostics.Append(req.Config.Get(ctx, &p.config.ProviderState)...)
	p.config.UserAgent = types.StringValue(req.TerraformVersion)
	p.config.ProviderState = setDefaults(p.config.ProviderState)
	if p.emptyFolder {
		p.config.ProviderState.FolderID = types.StringValue("")
	}

	if err := p.config.InitAndValidate(ctx, req.TerraformVersion, false); err != nil {
		resp.Diagnostics.AddError("Failed to configure", err.Error())
	}
	resp.ResourceData = &p.config
	resp.DataSourceData = &p.config
}

func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return billing_cloud_binding.NewResource(
				billing_cloud_binding.BindingServiceInstanceCloudType,
				billing_cloud_binding.BindingServiceInstanceCloudIdFieldName)
		},
		datasphere_project.NewResource,
		datasphere_project_iam_binding.NewIamBinding,
		datasphere_community.NewResource,
		datasphere_community_iam_binding.NewIamBinding,
		mdb_mongodb_database.NewResource,
		mdb_mongodb_user.NewResource,
		mdb_opensearch_cluster.NewResource,
		compute_disk_iam_binding.NewIamBinding,
		compute_disk_placement_group_iam_binding.NewIamBinding,
		compute_filesystem_iam_binding.NewIamBinding,
		compute_gpu_cluster_iam_binding.NewIamBinding,
		compute_image_iam_binding.NewIamBinding,
		compute_instance_iam_binding.NewIamBinding,
		compute_placement_group_iam_binding.NewIamBinding,
		compute_snapshot_iam_binding.NewIamBinding,
		compute_snapshot_schedule_iam_binding.NewIamBinding,
		airflow_cluster.NewResource,
		vpc_security_group_rule.NewResource,
		mdb_postgresql_cluster_beta.NewPostgreSQLClusterResourceBeta,
		kubernetes_marketplace_helm_release.NewResource,
	}
}

func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource {
			return billing_cloud_binding.NewDataSource(
				billing_cloud_binding.BindingServiceInstanceCloudType,
				billing_cloud_binding.BindingServiceInstanceCloudIdFieldName)
		},
		datasphere_project.NewDataSource,
		datasphere_community.NewDataSource,
		mdb_mongodb_database.NewDataSource,
		mdb_mongodb_user.NewDataSource,
		mdb_opensearch_cluster.NewDataSource,
		vpc_security_group_rule.NewDataSource,
	}
}

func (p *Provider) GetConfig() provider_config.Config {
	return p.config
}
