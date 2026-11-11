package pluginsdk

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	certificatemanagersdk "github.com/yandex-cloud/go-sdk/services/certificatemanager/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

const yandexIAMCMDefaultTimeout = 1 * time.Minute
const yandexIAMCMUpdateAccessBindingsBatchSize = 1000

var IamCMCertificateSchema = map[string]*schema.Schema{
	"certificate_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The [Certificate](https://yandex.cloud/docs/certificate-manager/) ID to apply a binding to.",
	},
}

type CMCertificateIamUpdater struct {
	certificateId string
	Config        *providerconfig.Config
}

func newCMCertificateIamUpdater(d *schema.ResourceData, config *providerconfig.Config) (sdkcommon.ResourceIamUpdater, error) {
	return &CMCertificateIamUpdater{
		certificateId: d.Get("certificate_id").(string),
		Config:        config,
	}, nil
}

func CMCertificateIDParseFunc(d *schema.ResourceData, _ *providerconfig.Config) error {
	err := d.Set("certificate_id", d.Id())
	if err != nil {
		return err
	}
	return nil
}

func (u *CMCertificateIamUpdater) GetResourceIamPolicy(ctx context.Context) (*sdkcommon.Policy, error) {
	bindings, err := GetCMCertificateAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &sdkcommon.Policy{bindings}, nil
}

func (u *CMCertificateIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *sdkcommon.Policy) error {
	client := certificatemanagersdk.NewCertificateClient(u.Config.SDK)

	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.certificateId,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMCMDefaultTimeout)
	defer cancel()

	op, err := client.SetAccessBindings(ctx, req)
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	if _, err = op.Wait(ctx); err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *CMCertificateIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *sdkcommon.PolicyDelta) error {
	client := certificatemanagersdk.NewCertificateClient(u.Config.SDK)

	bSize := yandexIAMCMUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < sdkcommon.CountBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.certificateId,
			AccessBindingDeltas: deltas[i*bSize : sdkcommon.Min((i+1)*bSize, dLen)],
		}

		op, err := client.UpdateAccessBindings(ctx, req)
		if err != nil {
			if reqID, ok := sdkcommon.IsRequestIDPresent(err); ok {
				log.Printf("[DEBUG] request ID is %s\n", reqID)
			}
			return fmt.Errorf("error updating access bindings of %s: %w", u.DescribeResource(), err)
		}

		if _, err = op.Wait(ctx); err != nil {
			return fmt.Errorf("error updating access bindings of %s: %w", u.DescribeResource(), err)
		}
	}

	return nil
}

func (u *CMCertificateIamUpdater) GetResourceID() string {
	return u.certificateId
}

func (u *CMCertificateIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-cm-certificate-%s", u.certificateId)
}

func (u *CMCertificateIamUpdater) DescribeResource() string {
	return fmt.Sprintf("Certificate Manager certificate '%s'", u.certificateId)
}

func GetCMCertificateAccessBindings(ctx context.Context, config *providerconfig.Config, certificateId string) ([]*access.AccessBinding, error) {
	client := certificatemanagersdk.NewCertificateClient(config.SDK)

	var bindings []*access.AccessBinding
	pageToken := ""

	for {
		resp, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: certificateId,
			PageSize:   sdkcommon.DefaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("error retrieving access bindings of Certificate Manager certificate %s: %w", certificateId, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
