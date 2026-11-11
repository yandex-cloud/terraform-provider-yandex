package pluginsdk

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	oidcsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1/workload/oidc"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

const yandexIAMWorkloadIdentityOidcFederationUpdateAccessBindingsBatchSize = 1000

var IamWorkloadIdentityOidcFederationSchema = map[string]*schema.Schema{
	"federation_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "ID of the OIDC workload identity federation to apply a binding to.",
	},
}

type WorkloadIdentityOidcFederationIamUpdater struct {
	federationID string
	Config       *providerconfig.Config
}

func newWorkloadIdentityOidcFederationIamUpdater(d *schema.ResourceData, config *providerconfig.Config) (sdkcommon.ResourceIamUpdater, error) {
	return &WorkloadIdentityOidcFederationIamUpdater{
		federationID: d.Get("federation_id").(string),
		Config:       config,
	}, nil
}

func workloadIdentityOidcFederationIDParseFunc(d *schema.ResourceData, _ *providerconfig.Config) error {
	d.Set("federation_id", d.Id())
	return nil
}

func (u *WorkloadIdentityOidcFederationIamUpdater) GetResourceIamPolicy(ctx context.Context) (*sdkcommon.Policy, error) {
	bindings, err := GetWorkloadIdentityOidcFederationAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &sdkcommon.Policy{bindings}, nil
}

func (u *WorkloadIdentityOidcFederationIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *sdkcommon.Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.federationID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), 1*time.Minute)
	defer cancel()

	client := oidcsdk.NewFederationClient(u.Config.SDK)

	op, err := client.SetAccessBindings(ctx, req)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *WorkloadIdentityOidcFederationIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *sdkcommon.PolicyDelta) error {
	bSize := yandexIAMWorkloadIdentityOidcFederationUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < sdkcommon.CountBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.federationID,
			AccessBindingDeltas: deltas[i*bSize : sdkcommon.Min((i+1)*bSize, dLen)],
		}

		client := oidcsdk.NewFederationClient(u.Config.SDK)

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

func (u *WorkloadIdentityOidcFederationIamUpdater) GetResourceID() string {
	return u.federationID
}

func (u *WorkloadIdentityOidcFederationIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-workload-identity-oidc-federation-%s", u.federationID)
}

func (u *WorkloadIdentityOidcFederationIamUpdater) DescribeResource() string {
	return fmt.Sprintf("workload identity OIDC federation '%s'", u.federationID)
}

func GetWorkloadIdentityOidcFederationAccessBindings(ctx context.Context, config *providerconfig.Config, federationID string) ([]*access.AccessBinding, error) {
	client := oidcsdk.NewFederationClient(config.SDK)

	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: federationID,
			PageSize:   sdkcommon.DefaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of workload identity OIDC federation %s: %w", federationID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
