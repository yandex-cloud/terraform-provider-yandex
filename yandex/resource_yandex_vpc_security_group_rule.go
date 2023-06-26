package yandex

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

func resourceYandexVpcSecurityGroupRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexVpcSecurityGroupRuleCreate,
		Read:   resourceYandexVpcSecurityGroupRuleRead,
		Update: resourceYandexVpcSecurityGroupRuleUpdate,
		Delete: resourceYandexVpcSecurityGroupRuleDelete,

		Importer: &schema.ResourceImporter{
			State: resourceYandexVpcSecurityGroupRuleImporterFunc,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexVPCSecurityGroupDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexVPCSecurityGroupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexVPCSecurityGroupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexVPCSecurityGroupDefaultTimeout),
		},

		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"direction": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ingress",
					"egress",
				}, false),
			},
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(-1, 65535),
				Default:      -1,
				ForceNew:     true,
			},
			"from_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(-1, 65535),
				Default:      -1,
				ForceNew:     true,
			},
			"to_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(-1, 65535),
				Default:      -1,
				ForceNew:     true,
			},
			"v4_cidr_blocks": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ForceNew:      true,
				ConflictsWith: []string{"security_group_id", "predefined_target"},
			},
			"v6_cidr_blocks": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ForceNew:      true,
				ConflictsWith: []string{"security_group_id", "predefined_target"},
			},
			"security_group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"predefined_target", "v4_cidr_blocks", "v6_cidr_blocks"},
			},
			"predefined_target": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"security_group_id", "v4_cidr_blocks", "v6_cidr_blocks"},
			},
			"security_group_binding": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceYandexVpcSecurityGroupRuleCreate(data *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), data.Timeout(schema.TimeoutCreate))
	defer cancel()

	ruleSpec, err := securityRuleResourceToSpec(data)
	if err != nil {
		return err
	}

	sgId := data.Get("security_group_binding").(string)

	mutexKV.Lock(sgId)
	defer mutexKV.Unlock(sgId)

	ruleId, err := addRuleToSecurityGroup(sgId, ruleSpec, config, ctx)
	if err != nil {
		return err
	}

	data.SetId(ruleId)

	return resourceYandexVpcSecurityGroupRuleRead(data, meta)
}

func resourceYandexVpcSecurityGroupRuleRead(data *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), data.Timeout(schema.TimeoutRead))
	defer cancel()

	sgId := data.Get("security_group_binding").(string)

	rule, err := findRule(data, config, ctx, sgId, data.Id())
	if err != nil {
		return err
	}

	return writeSecurityGroupRuleToData(rule, data)
}

func resourceYandexVpcSecurityGroupRuleUpdate(data *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	data.Partial(true)

	ctx, cancel := context.WithTimeout(config.Context(), data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	sgId := data.Get("security_group_binding").(string)

	mutexKV.Lock(sgId)
	defer mutexKV.Unlock(sgId)

	req := &vpc.UpdateSecurityGroupRuleRequest{
		RuleId:          data.Id(),
		SecurityGroupId: sgId,
		UpdateMask:      &field_mask.FieldMask{},
	}

	if data.HasChange("description") {
		req.Description = data.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	if data.HasChange("labels") {
		labelsProp, err := expandLabels(data.Get("labels"))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	if len(req.UpdateMask.Paths) > 0 {
		op, err := config.sdk.WrapOperation(config.sdk.VPC().SecurityGroup().UpdateRule(ctx, req))
		if err != nil {
			return err
		}

		err = op.Wait(ctx)
		if err != nil {
			return err
		}

		if _, err := op.Response(); err != nil {
			return err
		}
	}

	data.Partial(false)

	return resourceYandexVpcSecurityGroupRuleRead(data, meta)
}

func resourceYandexVpcSecurityGroupRuleDelete(data *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), data.Timeout(schema.TimeoutDelete))
	defer cancel()

	sgId := data.Get("security_group_binding").(string)

	mutexKV.Lock(sgId)
	defer mutexKV.Unlock(sgId)

	op, err := config.sdk.WrapOperation(config.sdk.VPC().SecurityGroup().UpdateRules(ctx, &vpc.UpdateSecurityGroupRulesRequest{
		SecurityGroupId: sgId,
		DeletionRuleIds: []string{data.Id()},
	}))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	data.SetId("")

	return nil
}

func addRuleToSecurityGroup(sgId string, ruleSpec *vpc.SecurityGroupRuleSpec, config *Config, ctx context.Context) (string, error) {
	op, err := config.sdk.WrapOperation(config.sdk.VPC().SecurityGroup().UpdateRules(ctx, &vpc.UpdateSecurityGroupRulesRequest{
		SecurityGroupId:   sgId,
		AdditionRuleSpecs: []*vpc.SecurityGroupRuleSpec{ruleSpec},
	}))
	if err != nil {
		return "", fmt.Errorf("error updating security group: %s", err)
	}

	meta, err := op.Metadata()
	if err != nil {
		return "", fmt.Errorf("failed to get metadata of update security group operation: %s", err)
	}

	updateMeta, ok := meta.(*vpc.UpdateSecurityGroupMetadata)
	if !ok {
		return "", fmt.Errorf("can't convert operation meta to update security group meta")
	}

	addedRuleIds := updateMeta.GetAddedRuleIds()

	if addedRuleIds == nil || len(addedRuleIds) != 1 {
		return "", fmt.Errorf("added rule ids list of update meta was nil or not singleton")
	}

	ruleId := updateMeta.GetAddedRuleIds()[0]

	err = op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("security group rules update failed: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return "", fmt.Errorf("security group rules update failed: %s", err)
	}

	return ruleId, nil
}

func findRule(data *schema.ResourceData, config *Config, ctx context.Context, sgId, ruleId string) (*vpc.SecurityGroupRule, error) {
	sg, err := config.sdk.VPC().SecurityGroup().Get(ctx, &vpc.GetSecurityGroupRequest{
		SecurityGroupId: sgId,
	})
	if err != nil {
		return nil, handleSecurityGroupNotFoundById(err, data, sgId)
	}

	rules := sg.Rules

	for i := range rules {
		if rules[i].Id == ruleId {
			return rules[i], nil
		}
	}

	return nil, fmt.Errorf("couldn't find rule %s in security group %s", data.Id(), sg.Id)
}

func securityRuleDescriptionToRuleSpec(dir string, v interface{}) (*vpc.SecurityGroupRuleSpec, error) {
	res, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("fail to cast %#v to map[string]interface{}", v)
	}

	sr := new(vpc.SecurityGroupRuleSpec)

	directionId := vpc.SecurityGroupRule_Direction_value[strings.ToUpper(dir)]
	sr.SetDirection(vpc.SecurityGroupRule_Direction(directionId))

	if v, ok := res["description"].(string); ok {
		sr.SetDescription(v)
	}

	if v, ok := res["security_group_id"].(string); ok && v != "" {
		sr.SetSecurityGroupId(v)
	}

	if v, ok := res["predefined_target"].(string); ok && v != "" {
		sr.SetPredefinedTarget(v)
	}

	if p, ok := res["protocol"].(string); ok {
		sr.SetProtocolName(strings.ToUpper(p))
	}

	if v, ok := res["labels"]; ok {
		labels, err := expandLabels(v)
		if err != nil {
			return sr, err
		}
		sr.SetLabels(labels)
	}

	if cidr, ok := securityRuleCidrsFromMap(res); ok {
		sr.SetCidrBlocks(cidr)
	}

	ports, err := securityRulePortsFromMap(res)
	if err != nil {
		return sr, err
	}
	sr.SetPorts(ports)

	return sr, nil
}

func handleSecurityGroupNotFoundById(err error, data *schema.ResourceData, id string) error {
	return handleNotFoundError(err, data, fmt.Sprintf("Security group %s", id))
}

func securityRuleResourceToSpec(data *schema.ResourceData) (*vpc.SecurityGroupRuleSpec, error) {
	sr := new(vpc.SecurityGroupRuleSpec)

	dir := data.Get("direction").(string)
	sr.SetDirection(vpc.SecurityGroupRule_Direction(vpc.SecurityGroupRule_Direction_value[strings.ToUpper(dir)]))

	if v, ok := data.Get("description").(string); ok {
		sr.SetDescription(v)
	}

	if v, ok := data.Get("security_group_id").(string); ok && v != "" {
		sr.SetSecurityGroupId(v)
	}

	if v, ok := data.Get("predefined_target").(string); ok && v != "" {
		sr.SetPredefinedTarget(v)
	}

	if p, ok := data.Get("protocol").(string); ok {
		sr.SetProtocolName(strings.ToUpper(p))
	}

	if v, ok := data.GetOk("labels"); ok {
		labels, err := expandLabels(v)
		if err != nil {
			return sr, err
		}
		sr.SetLabels(labels)
	}

	if cidr, ok := securityRuleCidrsFromResourceData(data); ok {
		sr.SetCidrBlocks(cidr)
	}

	ports, err := securityRulePortsFromResourceData(data)
	if err != nil {
		return sr, err
	}
	sr.SetPorts(ports)

	return sr, nil
}

func writeSecurityGroupRuleToData(rule *vpc.SecurityGroupRule, data *schema.ResourceData) error {
	if err := data.Set("protocol", rule.GetProtocolName()); err != nil {
		return err
	}

	if err := data.Set("description", rule.GetDescription()); err != nil {
		return err
	}

	if err := data.Set("labels", rule.GetLabels()); err != nil {
		return err
	}

	port, fromPort, toPort := flattenSecurityGroupRulesProto(rule)
	if err := data.Set("port", port); err != nil {
		return err
	}
	if err := data.Set("from_port", fromPort); err != nil {
		return err
	}
	if err := data.Set("to_port", toPort); err != nil {
		return err
	}

	if err := data.Set("security_group_id", rule.GetSecurityGroupId()); err != nil {
		return nil
	}

	if err := data.Set("predefined_target", rule.GetPredefinedTarget()); err != nil {
		return nil
	}

	if err := data.Set("direction", strings.ToLower(rule.GetDirection().String())); err != nil {
		return nil
	}

	if cidr := rule.GetCidrBlocks(); cidr != nil {
		if cidr.V4CidrBlocks != nil {
			if err := data.Set("v4_cidr_blocks", convertStringArrToInterface(cidr.V4CidrBlocks)); err != nil {
				return err
			}
		}

		if cidr.V6CidrBlocks != nil {
			if err := data.Set("v6_cidr_blocks", convertStringArrToInterface(cidr.V6CidrBlocks)); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceYandexVpcSecurityGroupRuleImporterFunc(data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	sgId, ruleId, err := deconstructResourceId(data.Id())
	if err != nil {
		return nil, err
	}

	data.Set("security_group_binding", sgId)
	data.SetId(ruleId)

	return []*schema.ResourceData{data}, nil
}
