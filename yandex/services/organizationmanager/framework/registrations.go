package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/framework/organizationmanager_idp_application_oauth_application_assignment"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/framework/organizationmanager_idp_application_saml_application_assignment"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/framework/organizationmanager_mfa_enforcement_audience"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/framework/organizationmanager_mfa_enforcement_excluded_audience"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		organizationmanager_idp_application_oauth_application_assignment.NewResource,
		organizationmanager_idp_application_saml_application_assignment.NewResource,
		organizationmanager_mfa_enforcement_audience.NewResource,
		organizationmanager_mfa_enforcement_excluded_audience.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return nil
}
