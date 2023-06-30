package yandex

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

var IamMemberBaseSchema = map[string]*schema.Schema{
	"role": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
	"member": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validateIamMember,
	},
	// for test purposes, to compensate IAM operations delay
	"sleep_after": {
		Type:     schema.TypeInt,
		Optional: true,
		ForceNew: true,
	},
}

func validateIamMember(i interface{}, k string) (s []string, es []error) {
	chunks := strings.SplitN(i.(string), ":", 2)
	if len(chunks) == 1 || chunks[0] == "" || chunks[1] == "" {
		es = append(es, fmt.Errorf("expect 'member' value should be in TYPE:ID format, got '%v'", i.(string)))
	}
	return
}

func iamMemberImport(resourceIDParser resourceIDParserFunc) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
		if resourceIDParser == nil {
			return nil, errors.New("Import not supported for this IAM resource")
		}
		config := meta.(*Config)
		s := strings.Fields(d.Id())
		if len(s) != 3 {
			d.SetId("")
			return nil, fmt.Errorf("Wrong number of parts to Member id %s; expected 'resource_name role member'", s)
		}
		id, role, member := s[0], s[1], s[2]

		// |member| part must be in TYPE:ID format.
		chunks := strings.SplitN(member, ":", 2)
		if len(chunks) != 2 {
			d.SetId("")
			return nil, errors.New("Invalid member spec, must be in TYPE:ID format")
		}

		// Set the ID only to the first part so all IAM types can share the same resourceIDParserFunc.
		d.SetId(id)
		d.Set("role", role)
		d.Set("member", member)

		err := resourceIDParser(d, config)
		if err != nil {
			return nil, err
		}

		// Set the ID again so that the ID matches the ID it would have if it had been created via TF.
		// Use the current ID in case it changed in the resourceIDParserFunc.
		d.SetId(d.Id() + "/" + role + "/" + member)
		return []*schema.ResourceData{d}, nil
	}
}

func resourceIamMember(parentSpecificSchema map[string]*schema.Schema, newUpdaterFunc newResourceIamUpdaterFunc, opts ...SchemaOption) *schema.Resource {
	r := &schema.Resource{
		CreateContext: resourceIamMemberCreate(newUpdaterFunc),
		ReadContext:   resourceIamMemberRead(newUpdaterFunc),
		DeleteContext: resourceIamMemberDelete(newUpdaterFunc),

		Schema: mergeSchemas(IamMemberBaseSchema, parentSpecificSchema),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func getResourceIamMember(d *schema.ResourceData) *access.AccessBinding {
	member := d.Get("member").(string)
	role := d.Get("role").(string)

	return roleMemberToAccessBinding(role, member)
}

func resourceIamMemberCreate(newUpdaterFunc newResourceIamUpdaterFunc) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return diag.FromErr(err)
		}

		member := getResourceIamMember(d)
		err = iamPolicyReadModifyUpdate(ctx, updater, &PolicyDelta{
			Deltas: []*access.AccessBindingDelta{
				{
					Action:        access.AccessBindingAction_ADD,
					AccessBinding: member,
				},
			},
		})
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(updater.GetResourceID() + "/" + member.RoleId + "/" + canonicalMember(member))

		if v, ok := d.GetOk("sleep_after"); ok {
			time.Sleep(time.Second * time.Duration(v.(int)))
		}

		return resourceIamMemberRead(newUpdaterFunc)(ctx, d, meta)
	}
}

func resourceIamMemberRead(newUpdaterFunc newResourceIamUpdaterFunc) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return diag.FromErr(err)
		}

		eMember := getResourceIamMember(d)
		p, err := updater.GetResourceIamPolicy(ctx)
		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				log.Printf("[DEBUG]: Access binding of member %q with role %q does not exist for non-existent resource %s, removing from state.", canonicalMember(eMember), eMember.RoleId, updater.DescribeResource())
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG]: Retrieved access bindings of %s: %+v\n", updater.DescribeResource(), p)
		////
		role := d.Get("role").(string)

		var mBinding []*access.AccessBinding
		for _, b := range p.Bindings {
			if b.RoleId != role {
				continue
			}
			mBinding = append(mBinding, b)
		}

		if mBinding == nil {
			log.Printf("[DEBUG]: Access binding for role %q does not exist in access bindings of %s, removing member %q from state.", eMember.RoleId, updater.DescribeResource(), canonicalMember(eMember))
			d.SetId("")
			return nil
		}

		var member string
		for _, b := range mBinding {
			if canonicalMember(b) == canonicalMember(eMember) {
				member = canonicalMember(b)
			}
		}
		if member == "" {
			log.Printf("[DEBUG]: Member %q for binding for role %q does not exist in access bindings of %s, removing from state.", canonicalMember(eMember), eMember.RoleId, updater.DescribeResource())
			d.SetId("")
			return nil
		}
		d.Set("member", member)
		d.Set("role", role)
		return nil
	}
}

func resourceIamMemberDelete(newUpdaterFunc newResourceIamUpdaterFunc) schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return diag.FromErr(err)
		}

		member := getResourceIamMember(d)

		err = iamPolicyReadModifyUpdate(ctx, updater, &PolicyDelta{
			Deltas: []*access.AccessBindingDelta{
				{
					Action:        access.AccessBindingAction_REMOVE,
					AccessBinding: member,
				},
			},
		})
		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				log.Printf("[DEBUG]: Member %q for binding for role %q does not exist for non-existent resource %q.", canonicalMember(member), member.RoleId, updater.GetResourceID())
				return nil
			}
			return diag.FromErr(err)
		}

		return resourceIamMemberRead(newUpdaterFunc)(ctx, d, meta)
	}
}

func canonicalMember(ab *access.AccessBinding) string {
	return ab.Subject.Type + ":" + ab.Subject.Id
}
