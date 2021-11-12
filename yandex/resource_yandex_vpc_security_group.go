package yandex

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexVPCSecurityGroupDefaultTimeout = 3 * time.Minute

func yandexVPCSecurityGroupSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"network_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"folder_id": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
			ForceNew: true,
		},

		"name": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
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

		"ingress": {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			Elem:     resourceYandexSecurityGroupRule(),
			Set:      resourceYandexVPCSecurityGroupRuleHash,
		},

		"egress": {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			Elem:     resourceYandexSecurityGroupRule(),
			Set:      resourceYandexVPCSecurityGroupRuleHash,
		},

		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},

		"created_at": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceYandexVPCSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexVPCSecurityGroupCreate,
		Read:   resourceYandexVPCSecurityGroupRead,
		Update: resourceYandexVPCSecurityGroupUpdate,
		Delete: resourceYandexVPCSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexVPCSecurityGroupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexVPCSecurityGroupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexVPCSecurityGroupDefaultTimeout),
		},

		SchemaVersion: 0,
		Schema:        yandexVPCSecurityGroupSchema(),
	}
}

func resourceYandexSecurityGroupRule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
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
			},
			"from_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(-1, 65535),
				Default:      -1,
			},
			"to_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(-1, 65535),
				Default:      -1,
			},
			"v4_cidr_blocks": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"v6_cidr_blocks": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"predefined_target": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexVPCSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("error expanding labels while creating security group: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("error getting folder ID while creating security group: %s", err)
	}

	rules, err := expandSecurityGroupRulesSpec(d)
	if err != nil {
		return fmt.Errorf("error getting rules while creating security group: %s", err)
	}

	req := vpc.CreateSecurityGroupRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		NetworkId:   d.Get("network_id").(string),
		RuleSpecs:   rules,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().SecurityGroup().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create security group: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while get security group create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*vpc.CreateSecurityGroupMetadata)
	if !ok {
		return fmt.Errorf("could not get SecurityGroup ID from create operation metadata")
	}

	d.SetId(md.SecurityGroupId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting operation to create security group: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("security group creation failed: %s", err)
	}

	return resourceYandexVPCSecurityGroupRead(d, meta)
}

func resourceYandexVPCSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	return yandexVPCSecurityGroupRead(d, meta, d.Id())
}

func yandexVPCSecurityGroupRead(d *schema.ResourceData, meta interface{}, id string) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	securityGroup, err := config.sdk.VPC().SecurityGroup().Get(ctx, &vpc.GetSecurityGroupRequest{
		SecurityGroupId: id,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Security group %q", d.Get("name").(string)))
	}

	if err := d.Set("created_at", getTimestamp(securityGroup.GetCreatedAt())); err != nil {
		return err
	}
	if err := d.Set("name", securityGroup.GetName()); err != nil {
		return err
	}
	if err := d.Set("folder_id", securityGroup.GetFolderId()); err != nil {
		return err
	}
	if err := d.Set("network_id", securityGroup.GetNetworkId()); err != nil {
		return err
	}
	if err := d.Set("description", securityGroup.GetDescription()); err != nil {
		return err
	}
	if err := d.Set("status", securityGroup.GetStatus().String()); err != nil {
		return err
	}

	ingress, egress := flattenSecurityGroupRulesSpec(securityGroup.Rules)

	if err := d.Set("ingress", ingress); err != nil {
		return err
	}
	if err := d.Set("egress", egress); err != nil {
		return err
	}

	return d.Set("labels", securityGroup.GetLabels())
}

func resourceYandexVPCSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	d.Partial(true)

	req := &vpc.UpdateSecurityGroupRequest{
		SecurityGroupId: d.Id(),
		UpdateMask:      &field_mask.FieldMask{},
	}

	if d.HasChange("labels") {
		labelsProp, err := expandLabels(d.Get("labels"))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	if len(req.UpdateMask.Paths) > 0 {
		op, err := config.sdk.WrapOperation(config.sdk.VPC().SecurityGroup().Update(ctx, req))
		if err != nil {
			return fmt.Errorf("error while requesting API to update Security group %q: %s", d.Id(), err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error updating Security group %q: %s", d.Id(), err)
		}

	}

	if d.HasChange("egress") || d.HasChange("ingress") {
		if err := resourceYandexVPCSecurityGroupUpdateRules(ctx, d, config); err != nil {
			return err
		}

	}

	d.Partial(false)

	return resourceYandexVPCSecurityGroupRead(d, meta)
}

func resourceYandexVPCSecurityGroupUpdateRules(ctx context.Context, d *schema.ResourceData, config *Config) error {
	sg, err := config.sdk.VPC().SecurityGroup().Get(ctx, &vpc.GetSecurityGroupRequest{
		SecurityGroupId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Security group %q", d.Id()))
	}

	cloudRules := map[string]*vpc.SecurityGroupRule{}

	for _, r := range sg.Rules {
		cloudRules[r.Id] = r
	}

	newRules := make([]*vpc.SecurityGroupRuleSpec, 0)
	delRules := make([]string, 0)
	ruleIds := make([]string, 0)

	for _, dir := range []string{"egress", "ingress"} {
		if v, ok := d.GetOk(dir); ok {
			for _, v := range v.(*schema.Set).List() {
				rule, ok := v.(map[string]interface{})
				if !ok {
					return fmt.Errorf("fail to cast %#v to map[string]interface{}", v)
				}

				if id, ok := rule["id"].(string); ok && id != "" {
					// existed rule
					if cloudRule, ok := cloudRules[id]; ok {
						ruleSpec, err := securityRuleDescriptionToRuleSpec(dir, v)
						if err != nil {
							return err
						}

						if ruleChanged(cloudRule, ruleSpec) {
							delRules = append(delRules, id)
							newRules = append(newRules, ruleSpec)
						} else {
							ruleIds = append(ruleIds, id)
						}

					} else {
						return fmt.Errorf("no rule with id %s on cloud", id)
					}

					ruleIds = append(ruleIds, id)
				} else {
					// new rule
					ruleSpec, err := securityRuleDescriptionToRuleSpec(dir, v)
					if err != nil {
						return err
					}
					newRules = append(newRules, ruleSpec)
				}
			}
		}
	}

	for cid := range cloudRules {
		found := false
		for _, id := range ruleIds {
			if cid == id {
				found = true
				break
			}
		}

		if !found {
			delRules = append(delRules, cid)
		}
	}

	req := &vpc.UpdateSecurityGroupRulesRequest{
		SecurityGroupId:   d.Id(),
		AdditionRuleSpecs: newRules,
		DeletionRuleIds:   delRules,
	}
	op, err := config.sdk.WrapOperation(config.sdk.VPC().SecurityGroup().UpdateRules(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update Security group rules %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error updating Security group rules %q: %s", d.Id(), err)
	}

	return nil
}

func ruleChanged(r1 *vpc.SecurityGroupRule, r2 *vpc.SecurityGroupRuleSpec) bool {
	if r1.GetDescription() != r2.GetDescription() {
		return true
	}

	if !reflect.DeepEqual(r1.GetLabels(), r2.GetLabels()) {
		return true
	}

	if r1.GetDirection() != r2.GetDirection() {
		return true
	}

	if !reflect.DeepEqual(r1.GetPorts(), r2.GetPorts()) {
		return true
	}

	if !reflect.DeepEqual(r1.GetCidrBlocks(), r2.GetCidrBlocks()) {
		return true
	}

	if r1.GetProtocolName() != r2.GetProtocolName() {
		return true
	}

	if r1.GetProtocolNumber() != r2.GetProtocolNumber() {
		return true
	}

	if r1.GetSecurityGroupId() != r2.GetSecurityGroupId() {
		return true
	}

	if r1.GetPredefinedTarget() != r2.GetPredefinedTarget() {
		return true
	}

	return false
}

func resourceYandexVPCSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req := &vpc.DeleteSecurityGroupRequest{
		SecurityGroupId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().SecurityGroup().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Security group %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	return nil
}

var hashableRuleNames = []string{
	"direction",
	"port",
	"from_port",
	"to_port",
	"security_group_id",
	"predefined_target",
	"description",
}

var toUpperCaseHashableRuleNames = []string{
	"protocol",
}

func resourceYandexVPCSecurityGroupRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}

	for _, name := range toUpperCaseHashableRuleNames {
		if v, ok := m[name]; ok {
			buf.WriteString(fmt.Sprintf("%v-", strings.ToUpper(v.(string))))
		}
	}

	for _, name := range hashableRuleNames {
		if v, ok := m[name]; ok {
			buf.WriteString(fmt.Sprintf("%v-", v))
		}
	}

	for _, name := range []string{"v4_cidr_blocks", "v6_cidr_blocks"} {
		if v, ok := m[name]; ok {
			arr := v.([]interface{})
			for _, c := range arr {
				buf.WriteString(c.(string))
			}
		}
	}

	return hashcode.String(buf.String())
}
