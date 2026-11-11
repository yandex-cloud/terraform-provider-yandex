package pluginsdk

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	containerregistrysdk "github.com/yandex-cloud/go-sdk/services/containerregistry/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

const yandexIAMContainerRegistryDefaultTimeout = 1 * time.Minute
const yandexContainerRegistryUpdateAccessBindingsBatchSize = 1000

var IamContainerRegistrySchema = map[string]*schema.Schema{
	"registry_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The [Yandex Container Registry](https://yandex.cloud/docs/container-registry/) ID to apply a binding to.",
	},
}

type ContainerRegistryIamUpdater struct {
	registryID string
	Config     *providerconfig.Config
}

func newContainerRegistryIamUpdater(d *schema.ResourceData, config *providerconfig.Config) (sdkcommon.ResourceIamUpdater, error) {
	return &ContainerRegistryIamUpdater{
		registryID: d.Get("registry_id").(string),
		Config:     config,
	}, nil
}

func containerRegistryIDParseFunc(d *schema.ResourceData, _ *providerconfig.Config) error {
	d.Set("registry_id", d.Id())
	return nil
}

func (u *ContainerRegistryIamUpdater) GetResourceIamPolicy(ctx context.Context) (*sdkcommon.Policy, error) {
	bindings, err := GetContainerRegistryAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &sdkcommon.Policy{bindings}, nil
}

func (u *ContainerRegistryIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *sdkcommon.Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.registryID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMContainerRegistryDefaultTimeout)
	defer cancel()

	client := containerregistrysdk.NewRegistryClient(u.Config.SDK)

	op, err := client.SetAccessBindings(ctx, req)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	if _, err = op.Wait(ctx); err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *ContainerRegistryIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *sdkcommon.PolicyDelta) error {
	bSize := yandexContainerRegistryUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)
	client := containerregistrysdk.NewRegistryClient(u.Config.SDK)

	for i := 0; i < sdkcommon.CountBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.registryID,
			AccessBindingDeltas: deltas[i*bSize : sdkcommon.Min((i+1)*bSize, dLen)],
		}

		op, err := client.UpdateAccessBindings(ctx, req)
		if err != nil {
			if reqID, ok := sdkcommon.IsRequestIDPresent(err); ok {
				log.Printf("[DEBUG] request ID is %s\n", reqID)
			}
			return fmt.Errorf("Error updating access bindings of %s: %w", u.DescribeResource(), err)
		}

		if _, err = op.Wait(ctx); err != nil {
			return fmt.Errorf("Error updating access bindings of %s: %w", u.DescribeResource(), err)
		}
	}

	return nil
}

func (u *ContainerRegistryIamUpdater) GetResourceID() string {
	return u.registryID
}

func (u *ContainerRegistryIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-container-registry-%s", u.registryID)
}

func (u *ContainerRegistryIamUpdater) DescribeResource() string {
	return fmt.Sprintf("Container Registry '%s'", u.registryID)
}

func GetContainerRegistryAccessBindings(ctx context.Context, config *providerconfig.Config, registryID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""
	client := containerregistrysdk.NewRegistryClient(config.SDK)

	for {
		resp, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: registryID,
			PageSize:   sdkcommon.DefaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of Container Registry %s: %w", registryID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
