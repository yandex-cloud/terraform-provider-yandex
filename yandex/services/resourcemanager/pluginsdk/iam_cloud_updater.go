package pluginsdk

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	resourcemanagersdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

const yandexResourceManagerCloudDefaultTimeout = time.Second * 60

type CloudIamUpdater struct {
	cloudID string
	Config  *providerconfig.Config
}

func NewCloudIamUpdater(cloudID string, config *providerconfig.Config) *CloudIamUpdater {
	return &CloudIamUpdater{cloudID: cloudID, Config: config}
}

func (u *CloudIamUpdater) GetResourceIamPolicy(ctx context.Context) (*sdkcommon.Policy, error) {
	bindings, err := GetCloudAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &sdkcommon.Policy{bindings}, nil
}

func (u *CloudIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *sdkcommon.Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.cloudID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexResourceManagerCloudDefaultTimeout)
	defer cancel()

	client := resourcemanagersdk.NewCloudClient(u.Config.SDK)

	op, err := client.SetAccessBindings(ctx, req)
	if err != nil {
		if reqID, ok := sdkcommon.IsRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *CloudIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *sdkcommon.PolicyDelta) error {
	bSize := sdkcommon.YandexResourceManagerCloudUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < sdkcommon.CountBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.cloudID,
			AccessBindingDeltas: deltas[i*bSize : sdkcommon.Min((i+1)*bSize, dLen)],
		}

		client := resourcemanagersdk.NewCloudClient(u.Config.SDK)

		op, err := client.UpdateAccessBindings(ctx, req)
		if err != nil {
			if reqID, ok := sdkcommon.IsRequestIDPresent(err); ok {
				log.Printf("[DEBUG] request ID is %s\n", reqID)
			}
			return fmt.Errorf("Error updating access bindings of %s: %w", u.DescribeResource(), err)
		}

		_, err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error updating access bindings of %s: %w", u.DescribeResource(), err)
		}
	}

	return nil
}

func (u *CloudIamUpdater) GetResourceID() string {
	return u.cloudID
}

func (u *CloudIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-cloud-%s", u.cloudID)
}

func (u *CloudIamUpdater) DescribeResource() string {
	return fmt.Sprintf("cloud %q", u.cloudID)
}

func GetCloudAccessBindings(ctx context.Context, config *providerconfig.Config, cloudID string) ([]*access.AccessBinding, error) {
	client := resourcemanagersdk.NewCloudClient(config.SDK)

	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: cloudID,
			PageSize:   sdkcommon.DefaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of cloud %s: %w", cloudID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
