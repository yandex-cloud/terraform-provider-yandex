package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexKubernetesClusterIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamKubernetesClusterSchema,
		newKubernetesClusterIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMKMSDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(kubernetesClusterIDParseFunc),
			}),
		WithDescription("Allows creation and management of a single binding within IAM policy for an existing Yandex Managed Service for Kubernetes cluster.\n\n~> Roles controlled by `yandex_kubernetes_cluster_iam_binding` should not be assigned using `yandex_kubernetes_cluster_iam_member`.\n\n~> When you delete `yandex_kubernetes_cluster_iam_binding` resource, the roles can be deleted from other users within the folder as well. Be careful!\n"),
	)
}
