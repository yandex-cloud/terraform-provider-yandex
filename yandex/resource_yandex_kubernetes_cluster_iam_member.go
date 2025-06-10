package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexKubernetesClusterIAMMember() *schema.Resource {
	return resourceIamMember(
		IamKubernetesClusterSchema,
		newKubernetesClusterIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMKMSDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamMemberImport(kubernetesClusterIDParseFunc),
			}),
		WithDescription("Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Managed Service for Kubernetes cluster.\n\n~> Roles controlled by `yandex_kubernetes_cluster_iam_binding` should not be assigned using `yandex_kubernetes_cluster_iam_member`.\n"),
	)
}
