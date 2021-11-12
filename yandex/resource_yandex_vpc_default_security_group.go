package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

const yandexVPCDefaultSecurityGroupDefaultTimeout = yandexVPCSecurityGroupDefaultTimeout

func yandexVPCDefaultSecurityGroupSchema() map[string]*schema.Schema {
	s := yandexVPCSecurityGroupSchema()

	// name field cannot be updated in default security group
	s["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}

	return s
}

func resourceYandexVPCDefaultSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexVPCDefaultSecurityGroupCreate,
		Read:   resourceYandexVPCDefaultSecurityGroupRead,
		Update: resourceYandexVPCDefaultSecurityGroupUpdate,
		Delete: resourceYandexVPCDefaultSecurityGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexVPCDefaultSecurityGroupDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexVPCDefaultSecurityGroupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexVPCDefaultSecurityGroupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexVPCDefaultSecurityGroupDefaultTimeout),
		},

		SchemaVersion: 0,
		Schema:        yandexVPCDefaultSecurityGroupSchema(),
	}
}

func resourceYandexVPCDefaultSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	networkId := d.Get("network_id").(string)
	network, err := config.sdk.VPC().Network().Get(ctx, &vpc.GetNetworkRequest{
		NetworkId: networkId,
	})
	if err != nil {
		return fmt.Errorf("error while receiving network %s: %s", networkId, err)
	}

	sgId := network.GetDefaultSecurityGroupId()

	if sgId == "" {
		return fmt.Errorf("network %s has no default security group", networkId)
	}

	upd := &vpc.UpdateSecurityGroupRequest{
		SecurityGroupId: sgId,
		UpdateMask:      &field_mask.FieldMask{},
	}

	upd.UpdateMask.Paths = []string{"labels", "description", "rule_specs"}

	upd.Description = d.Get("description").(string)

	if l, ok := d.GetOk("labels"); ok {
		labels, err := expandLabels(l)
		if err != nil {
			return err
		}
		upd.Labels = labels
	} else {
		upd.Labels = make(map[string]string)
	}

	upd.RuleSpecs = make([]*vpc.SecurityGroupRuleSpec, 0)

	for _, dir := range []string{"egress", "ingress"} {
		v, ok := d.GetOk(dir)
		if !ok {
			continue
		}

		for _, v := range v.(*schema.Set).List() {
			ruleSpec, err := securityRuleDescriptionToRuleSpec(dir, v)
			if err != nil {
				return err
			}

			upd.RuleSpecs = append(upd.RuleSpecs, ruleSpec)
		}
	}

	op, err := config.sdk.WrapOperation(config.sdk.VPC().SecurityGroup().Update(ctx, upd))
	if err != nil {
		return fmt.Errorf("error while updating security group %s: %s", sgId, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error updating security group %s: %s", sgId, err)
	}

	d.SetId(sgId)

	return resourceYandexVPCDefaultSecurityGroupRead(d, meta)
}

func resourceYandexVPCDefaultSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	return yandexVPCSecurityGroupRead(d, meta, d.Id())
}

func resourceYandexVPCDefaultSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceYandexVPCSecurityGroupUpdate(d, meta)
}

func resourceYandexVPCDefaultSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	// no need to actually delete sg or it rules
	d.SetId("")

	return nil
}
