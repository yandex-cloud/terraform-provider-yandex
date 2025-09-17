package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const iamPolicyDescription string = "**IAM policy for a service account**\nWhen managing IAM roles, you can treat a service account either as a resource or as an identity. This resource is used to add IAM policy bindings to a service account resource to configure permissions that define who can edit the service account.\n\nThere are three different resources that help you manage your IAM policy for a service account. Each of these resources is used for a different use case:\n* [yandex_iam_service_account_iam_policy](iam_service_account_iam_policy.html): Authoritative. Sets the IAM policy for the service account and replaces any existing policy already attached.\n* [yandex_iam_service_account_iam_binding](iam_service_account_iam_binding.html): Authoritative for a given role. Updates the IAM policy to grant a role to a list of members. Other roles within the IAM policy for the service account are preserved.\n* [yandex_iam_service_account_iam_member](iam_service_account_iam_member.html): Non-authoritative. Updates the IAM policy to grant a role to a new member. Other members for the role of the service account are preserved.\n\n~> `yandex_iam_service_account_iam_policy` **cannot** be used in conjunction with `yandex_iam_service_account_iam_binding` and `yandex_iam_service_account_iam_member` or they will conflict over what your policy should be.\n\n~> `yandex_iam_service_account_iam_binding` resources **can be** used in conjunction with `yandex_iam_service_account_iam_member` resources **only if** they do not grant privileges to the same role.\n"

func resourceYandexIAMServiceAccountIAMPolicy() *schema.Resource {
	return resourceIamPolicy(
		IamServiceAccountSchema,
		newServiceAccountIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMServiceAccountDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamPolicyImport(serviceAccountIDParseFunc),
			}),
		WithDescription(iamPolicyDescription),
	)
}
