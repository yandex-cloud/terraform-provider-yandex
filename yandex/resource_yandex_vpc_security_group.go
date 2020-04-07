package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

const yandexVPCSecurityGroupDefaultTimeout = 3 * time.Minute

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

		Schema: map[string]*schema.Schema{
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
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"direction": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"INGRESS", "EGRESS"}, false),
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"protocol_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"protocol_number": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
						"from_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
						"to_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
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
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexVPCSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := config.sdk

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating security group: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating security group: %s", err)
	}

	rules, err := expandSecurityGroupRulesSpec(d)
	if err != nil {
		return fmt.Errorf("Error getting rules while creating security group: %s", err)
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

	op, err := sdk.WrapOperation(sdk.VPC().SecurityGroup().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create security group: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get security group create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*vpc.CreateSecurityGroupMetadata)
	if !ok {
		return fmt.Errorf("could not get SecurityGroup ID from create operation metadata")
	}

	d.SetId(md.SecurityGroupId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create security group: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Security group creation failed: %s", err)
	}

	return resourceYandexVPCSecurityGroupRead(d, meta)
}

func resourceYandexVPCSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := config.sdk

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	securityGroup, err := sdk.VPC().SecurityGroup().Get(ctx, &vpc.GetSecurityGroupRequest{
		SecurityGroupId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Security group %q", d.Get("name").(string)))
	}

	createdAt, err := getTimestamp(securityGroup.GetCreatedAt())
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("name", securityGroup.GetName())
	d.Set("folder_id", securityGroup.GetFolderId())
	d.Set("network_id", securityGroup.GetNetworkId())
	d.Set("description", securityGroup.GetDescription())
	d.Set("status", securityGroup.GetStatus())
	d.Set("labels", securityGroup.GetLabels())

	rules, err := flattenSecurityGroupRulesSpec(securityGroup.Rules)
	if err != nil {
		return err
	}

	return d.Set("rule", rules)
}

func resourceYandexVPCSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := config.sdk

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
		op, err := sdk.WrapOperation(sdk.VPC().SecurityGroup().Update(ctx, req))
		if err != nil {
			return fmt.Errorf("Error while requesting API to update Security group %q: %s", d.Id(), err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error updating Security group %q: %s", d.Id(), err)
		}

		for _, v := range req.UpdateMask.Paths {
			d.SetPartial(v)
		}
	}

	if d.HasChange("rule") {
		if err := resourceYandexVPCSecurityGroupUpdateRules(ctx, d, config); err != nil {
			return err
		}
		d.SetPartial("rule")
	}

	d.Partial(false)

	return resourceYandexVPCSecurityGroupRead(d, meta)
}

func resourceYandexVPCSecurityGroupUpdateRules(ctx context.Context, d *schema.ResourceData, config *Config) error {
	sdk := config.sdk

	sg, err := sdk.VPC().SecurityGroup().Get(ctx, &vpc.GetSecurityGroupRequest{
		SecurityGroupId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Security group %q", d.Id()))
	}

	newRules := make([]*vpc.SecurityGroupRuleSpec, 0)
	delRules := make([]string, 0)

	rulenum := d.Get("rule.#").(int)
	ruleIds := make([]string, rulenum)

	for i := 0; i < rulenum; i++ {
		key := fmt.Sprintf("rule.%d", i)

		if v, ok := d.GetOk(key + ".id"); ok {
			ruleIds[i] = v.(string)

			if d.HasChange(key) {
				r, err := expandSecurityGroupRuleSpec(d, key)
				if err != nil {
					return err
				}
				newRules = append(newRules, r)
				delRules = append(delRules, v.(string))
			}
		} else {
			r, err := expandSecurityGroupRuleSpec(d, key)
			if err != nil {
				return err
			}
			newRules = append(newRules, r)
		}
	}

	for _, r := range sg.Rules {
		found := false
		for _, id := range ruleIds {
			if r.Id == id {
				found = true
				break
			}
		}

		if !found {
			delRules = append(delRules, r.Id)
		}
	}

	req := &vpc.UpdateSecurityGroupRulesRequest{
		SecurityGroupId:   d.Id(),
		AdditionRuleSpecs: newRules,
		DeletionRuleIds:   delRules,
	}
	op, err := sdk.WrapOperation(sdk.VPC().SecurityGroup().UpdateRules(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Security group rules %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Security group rules %q: %s", d.Id(), err)
	}

	return nil
}

func resourceYandexVPCSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := config.sdk

	req := &vpc.DeleteSecurityGroupRequest{
		SecurityGroupId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := sdk.WrapOperation(sdk.VPC().SecurityGroup().Delete(ctx, req))
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
