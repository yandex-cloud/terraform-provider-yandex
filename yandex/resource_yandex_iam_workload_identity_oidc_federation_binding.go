package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexIAMWorkloadIdentityOidcFederationIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamWorkloadIdentityOidcFederationSchema,
		newWorkloadIdentityOidcFederationIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMWorkloadIdentityOidcFederationDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(workloadIdentityOidcFederationIDParseFunc),
			}),
		WithDescription("Allows creation and management of a single binding within IAM policy for an existing IAM workload identity OIDC federations"),
	)
}
