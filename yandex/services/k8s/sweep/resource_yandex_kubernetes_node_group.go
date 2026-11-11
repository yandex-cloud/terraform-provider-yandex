package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
	k8ssdk "github.com/yandex-cloud/go-sdk/services/k8s/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_managed_kubernetes "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/k8s/pluginsdk"
)

func registerResourceYandexKubernetesNodeGroup(add func(string, *resource.Sweeper)) {
	add("yandex_kubernetes_node_group", &resource.Sweeper{
		Name: "yandex_kubernetes_node_group",
		F:    sweepAllKubernetesNodeGroups,
	})
}

func sweepAllKubernetesNodeGroups(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &k8s.ListNodeGroupsRequest{FolderId: conf.FolderID}
	client := k8ssdk.NewNodeGroupClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepKubernetesNodeGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Kubernetes Node Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepKubernetesNodeGroup(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepKubernetesNodeGroupOnce, conf, "Kubernetes Node Group", id)
}

func sweepKubernetesNodeGroupOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_managed_kubernetes.YandexKubernetesNodeGroupDeleteTimeout)
	defer cancel()

	client := k8ssdk.NewNodeGroupClient(conf.SDK)

	op, err := client.Delete(ctx, &k8s.DeleteNodeGroupRequest{
		NodeGroupId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
