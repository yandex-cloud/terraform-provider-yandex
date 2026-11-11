package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
	k8ssdk "github.com/yandex-cloud/go-sdk/services/k8s/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	iamsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/sweep"
	svc_managed_kubernetes "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/k8s/pluginsdk"
)

func registerResourceYandexKubernetesCluster(add func(string, *resource.Sweeper)) {
	add("yandex_kubernetes_cluster", &resource.Sweeper{
		Name: "yandex_kubernetes_cluster",
		F:    sweepAllKubernetesClusters,
		Dependencies: []string{
			"yandex_kubernetes_node_group",
		},
	})
}

func sweepAllKubernetesClusters(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	var serviceAccountID string
	var depsCreated bool

	req := &k8s.ListClustersRequest{FolderId: conf.FolderID}
	client := k8ssdk.NewClusterClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		if !depsCreated {
			depsCreated = true
			serviceAccountID, err = iamsweep.CreateIAMServiceAccountForSweeper(conf)
			if err != nil {
				result = multierror.Append(result, err)
				break
			}
		}

		id := it.Value().GetId()
		if !updateKubernetesClusterWithSweeperDeps(conf, id, serviceAccountID) {
			result = multierror.Append(result,
				fmt.Errorf("failed to sweep (update with dependencies) Kubernetes Cluster %q", id))
			continue
		}

		if !sweepKubernetesCluster(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Kubernetes Cluster %q", id))
		}
	}

	return result.ErrorOrNil()
}

func updateKubernetesClusterWithSweeperDeps(conf *providerconfig.Config, clusterID, serviceAccountID string) bool {
	sweep.DebugLog("started updating Kubernetes Cluster %q", clusterID)

	client := k8ssdk.NewClusterClient(conf.SDK)

	for i := 1; i <= conf.MaxRetries; i++ {
		req := &k8s.UpdateClusterRequest{
			ClusterId:        clusterID,
			ServiceAccountId: serviceAccountID,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{
					"service_account_id",
				},
			},
		}

		_, err := client.Update(conf.Context(), req)
		if err != nil {
			sweep.DebugLog("[kubernetes cluster %q] update try #%d: %v", clusterID, i, err)
		} else {
			sweep.DebugLog("[kubernetes cluster %q] update try #%d: request was successfully sent", clusterID, i)
			return true
		}
	}

	sweep.DebugLog("[kubernetes cluster %q] update failed", clusterID)
	return false
}

func sweepKubernetesCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepKubernetesClusterOnce, conf, "Kubernetes Cluster", id)
}

func sweepKubernetesClusterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_managed_kubernetes.YandexKubernetesClusterDefaultTimeout)
	defer cancel()

	client := k8ssdk.NewClusterClient(conf.SDK)

	op, err := client.Delete(ctx, &k8s.DeleteClusterRequest{
		ClusterId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
