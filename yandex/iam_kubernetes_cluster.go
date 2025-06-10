package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

var IamKubernetesClusterSchema = map[string]*schema.Schema{
	"cluster_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The [Yandex Managed Service for Kubernetes](https://yandex.cloud/docs/managed-kubernetes/) cluster ID to apply a binding to.",
	},
}

type KubernetesClusterIamUpdater struct {
	clusterID string
	Config    *Config
}

func newKubernetesClusterIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &KubernetesClusterIamUpdater{
		clusterID: d.Get("cluster_id").(string),
		Config:    config,
	}, nil
}

func kubernetesClusterIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("cluster_id", d.Id())
	return nil
}

func (u *KubernetesClusterIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getKubernetesClusterBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *KubernetesClusterIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.clusterID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMKMSDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.Kubernetes().Cluster().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *KubernetesClusterIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexIAMKMSUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.clusterID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.Kubernetes().Cluster().UpdateAccessBindings(ctx, req))
		if err != nil {
			if reqID, ok := isRequestIDPresent(err); ok {
				log.Printf("[DEBUG] request ID is %s\n", reqID)
			}
			return fmt.Errorf("Error updating access bindings of %s: %w", u.DescribeResource(), err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error updating access bindings of %s: %w", u.DescribeResource(), err)
		}
	}

	return nil
}

func (u *KubernetesClusterIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-kubernetes-cluster-%s", u.clusterID)
}

func (u *KubernetesClusterIamUpdater) GetResourceID() string {
	return u.clusterID
}

func (u *KubernetesClusterIamUpdater) DescribeResource() string {
	return fmt.Sprintf("Kubernetes cluster '%s'", u.clusterID)
}

func getKubernetesClusterBindings(ctx context.Context, config *Config, clusterID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.sdk.Kubernetes().Cluster().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: clusterID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of %s: %w", clusterID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
