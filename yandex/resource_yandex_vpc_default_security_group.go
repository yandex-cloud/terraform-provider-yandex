package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const yandexVPCDefaultSecurityGroupDefaultTimeout = yandexVPCSecurityGroupDefaultTimeout

func yandexVPCDefaultSecurityGroupSchema() map[string]*schema.Schema {
	s := yandexVPCSecurityGroupSchema()

	// name field cannot be updated in default security group
	s["name"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: common.ResourceDescriptions["name"] + " Cannot be updated.",
		Computed:    true,
	}

	return s
}

func resourceYandexVPCDefaultSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Default Security Group within the Yandex Cloud. For more information, see the official documentation of [security group](https://yandex.cloud/docs/vpc/concepts/security-groups) or [default security group](https://yandex.cloud/docs/vpc/concepts/security-groups#default-security-group).\n\n~> This resource is not intended for managing security group in general case. To manage normal security group use [yandex_vpc_security_group](vpc_security_group.html)\n\nWhen [network](https://yandex.cloud/docs/vpc/concepts/network) is created, a non-removable security group, called a *default security group*, is automatically attached to it. Life time of default security group cannot be controlled, so in fact the resource `yandex_vpc_default_security_group` does not create or delete any security groups, instead it simply takes or releases control of the default security group.\n\n~> When Terraform takes over management of the default security group, it **deletes** all info in it (including security group rules) and replace it with specified configuration. When Terraform drops the management (i.e. when resource is deleted from statefile and management), the state of the security group **remains the same** as it was before the deletion.\n\n~> Duplicating a resource (specifying same `network_id` for two different default security groups) will cause errors in the apply stage of your's configuration.\n",
		Create:      resourceYandexVPCDefaultSecurityGroupCreate,
		Read:        resourceYandexVPCDefaultSecurityGroupRead,
		Update:      resourceYandexVPCDefaultSecurityGroupUpdate,
		Delete:      resourceYandexVPCDefaultSecurityGroupDelete,

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
