package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMCMDefaultTimeout = 1 * time.Minute
const yandexIAMCMUpdateAccessBindingsBatchSize = 1000

var IamCMCertificateSchema = map[string]*schema.Schema{
	"certificate_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type CMCertificateIamUpdater struct {
	certificateId string
	Config        *Config
}

func newCMCertificateIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &CMCertificateIamUpdater{
		certificateId: d.Get("certificate_id").(string),
		Config:        config,
	}, nil
}

func CMCertificateIDParseFunc(d *schema.ResourceData, _ *Config) error {
	err := d.Set("certificate_id", d.Id())
	if err != nil {
		return err
	}
	return nil
}

func (u *CMCertificateIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getCMCertificateAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *CMCertificateIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.certificateId,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMCMDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.Certificates().Certificate().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *CMCertificateIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexIAMCMUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.certificateId,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.Certificates().Certificate().UpdateAccessBindings(ctx, req))
		if err != nil {
			if reqID, ok := isRequestIDPresent(err); ok {
				log.Printf("[DEBUG] request ID is %s\n", reqID)
			}
			return fmt.Errorf("error updating access bindings of %s: %w", u.DescribeResource(), err)
		}

		err = op.Wait(ctx)
		if err != nil {
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

func getCMCertificateAccessBindings(ctx context.Context, config *Config, certificateId string) ([]*access.AccessBinding, error) {
	var bindings []*access.AccessBinding
	pageToken := ""

	for {
		resp, err := config.sdk.Certificates().Certificate().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: certificateId,
			PageSize:   defaultListSize,
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
