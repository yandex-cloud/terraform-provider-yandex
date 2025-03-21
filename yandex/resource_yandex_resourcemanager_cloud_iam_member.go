package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexResourceManagerCloudIAMMember() *schema.Resource {
	return resourceIamMember(
		IamCloudSchema,
		newCloudIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexResourceManagerCloudDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamMemberImport(cloudIDParseFunc),
			}),
		WithDescription("Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Resource Manager cloud.\n\n~> Roles controlled by `yandex_resourcemanager_cloud_iam_binding` should not be assigned using `yandex_resourcemanager_cloud_iam_member`.\n\n~> When you delete `yandex_resourcemanager_cloud_iam_binding` resource, the roles can be deleted from other users within the cloud as well. Be careful!\n"),
	)
}
