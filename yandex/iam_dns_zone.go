package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	dnssdk "github.com/yandex-cloud/go-sdk/services/dns/v1"
)

const yandexIAMDnsZoneDefaultTimeout = 1 * time.Minute
const yandexIAMDnsZoneUpdateAccessBindingsBatchSize = 1000

var IamDnsZoneSchema = map[string]*schema.Schema{
	"dns_zone_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The [DNS](https://yandex.cloud/docs/dns/) Zone ID to apply a binding to.",
	},
}

type DnsZoneIamUpdater struct {
	dnsZoneId string
	Config    *Config
}

func newDnsZoneIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &DnsZoneIamUpdater{
		dnsZoneId: d.Get("dns_zone_id").(string),
		Config:    config,
	}, nil
}

func dnsZoneIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("dns_zone_id", d.Id())
	return nil
}

func (u *DnsZoneIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getDnsZoneAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *DnsZoneIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	client := dnssdk.NewDnsZoneClient(u.Config.SDK)

	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.dnsZoneId,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMDnsZoneDefaultTimeout)
	defer cancel()

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

func (u *DnsZoneIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	client := dnssdk.NewDnsZoneClient(u.Config.SDK)

	bSize := yandexIAMDnsZoneUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.dnsZoneId,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := client.UpdateAccessBindings(ctx, req)
		if err != nil {
			if reqID, ok := isRequestIDPresent(err); ok {
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

func (u *DnsZoneIamUpdater) GetResourceID() string {
	return u.dnsZoneId
}

func (u *DnsZoneIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("dns-zone-%s", u.dnsZoneId)
}

func (u *DnsZoneIamUpdater) DescribeResource() string {
	return fmt.Sprintf("DNS Zone '%s'", u.dnsZoneId)
}

func getDnsZoneAccessBindings(ctx context.Context, config *Config, databaseID string) ([]*access.AccessBinding, error) {
	client := dnssdk.NewDnsZoneClient(config.SDK)

	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: databaseID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of DNS Zone %s: %w", databaseID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
