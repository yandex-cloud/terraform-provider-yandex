package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

var accessBindingSchema = map[string]*schema.Schema{
	"role": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
	"members": {
		Type:     schema.TypeSet,
		Required: true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validateIamMember,
		},
	},
	// for test purposes, to compensate IAM operations delay
	"sleep_after": {
		Type:     schema.TypeInt,
		Optional: true,
		ForceNew: false,
	},
}

func resourceIamBinding(parentSpecificSchema map[string]*schema.Schema, newUpdaterFunc newResourceIamUpdaterFunc, opts ...SchemaOption) *schema.Resource {
	r := &schema.Resource{
		CreateContext: resourceAccessBindingCreate(newUpdaterFunc),
		ReadContext:   resourceAccessBindingRead(newUpdaterFunc, false),
		UpdateContext: resourceAccessBindingUpdate(newUpdaterFunc),
		DeleteContext: resourceAccessBindingDelete(newUpdaterFunc),
		Schema:        mergeSchemas(accessBindingSchema, parentSpecificSchema),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func resourceAccessBindingCreate(newUpdaterFunc newResourceIamUpdaterFunc) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return diag.FromErr(err)
		}

		p := getResourceIamBindings(d)
		err = iamPolicyReadModifySet(ctx, updater, func(ep *Policy) error {
			// Creating a binding does not remove existing members if they are not in the provided members list.
			// This prevents removing existing permission without the user's knowledge.
			// Instead, a diff is shown in that case after creation. Subsequent calls to update will remove any
			// existing members not present in the provided list.
			ep.Bindings = mergeBindings(append(ep.Bindings, p...))
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}

		role := d.Get("role").(string)
		d.SetId(updater.GetResourceID() + "/" + role)

		if v, ok := d.GetOk("sleep_after"); ok {
			time.Sleep(time.Second * time.Duration(v.(int)))
		}

		return resourceAccessBindingRead(newUpdaterFunc, true)(ctx, d, meta)
	}
}

func resourceAccessBindingRead(newUpdaterFunc newResourceIamUpdaterFunc, check bool) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return diag.FromErr(err)
		}

		role := d.Get("role").(string)
		eBindings := getResourceIamBindings(d)

		p, err := updater.GetResourceIamPolicy(ctx)
		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				if check {
					return diag.FromErr(fmt.Errorf("Access bindings for role %q not found for non-existent resource %s.", role, updater.DescribeResource()))
				} else {
					log.Printf("[DEBUG]: Access bindings for role %q not found for non-existent resource %s, removing from state.", role, updater.DescribeResource())
					d.SetId("")
					return nil
				}
			}

			return diag.FromErr(err)
		}
		log.Printf("[DEBUG]: Retrieved access bindings of %s: %+v", updater.DescribeResource(), p)

		var mBindings []*access.AccessBinding
		for _, b := range p.Bindings {
			if b.RoleId != role {
				continue
			}
			if len(eBindings) != 0 {
				for _, e := range eBindings {
					if canonicalMember(e) != canonicalMember(b) {
						continue
					}
					mBindings = append(mBindings, b)
				}
			} else {
				mBindings = append(mBindings, b)
			}
		}

		if len(mBindings) == 0 {
			if check {
				return diag.FromErr(fmt.Errorf("Access bindings for role %q not found in access bindings of %s.", role, updater.DescribeResource()))
			} else {
				log.Printf("[DEBUG]: Access bindings for role %q not found in access bindings of %s, removing from state.", role, updater.DescribeResource())
				d.SetId("")
				return nil
			}
		}

		if err := d.Set("members", roleToMembersList(role, mBindings)); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
}

func resourceAccessBindingUpdate(newUpdaterFunc newResourceIamUpdaterFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return diag.FromErr(err)
		}

		bindings := getResourceIamBindings(d)
		role := d.Get("role").(string)

		err = iamPolicyReadModifySet(ctx, updater, func(p *Policy) error {
			p.Bindings = removeRoleFromBindings(role, p.Bindings)
			p.Bindings = append(p.Bindings, bindings...)
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}

		return resourceAccessBindingRead(newUpdaterFunc, true)(ctx, d, meta)
	}
}

func resourceAccessBindingDelete(newUpdaterFunc newResourceIamUpdaterFunc) schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return diag.FromErr(err)
		}

		binding := getResourceIamBindings(d)
		if len(binding) == 0 {
			log.Printf("[DEBUG]: Resource %s is missing or deleted, marking policy binding as deleted", updater.DescribeResource())
			return nil
		}
		role := d.Get("role").(string)

		err = iamPolicyReadModifySet(ctx, updater, func(p *Policy) error {
			p.Bindings = removeRoleFromBindings(role, p.Bindings)
			return nil
		})

		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				log.Printf("[DEBUG]: Resource %s is missing or deleted, marking policy binding as deleted", updater.DescribeResource())
				return nil
			}

			return diag.FromErr(err)
		}

		return resourceAccessBindingRead(newUpdaterFunc, false)(ctx, d, meta)
	}
}

func iamBindingImport(resourceIDParser resourceIDParserFunc) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
		if resourceIDParser == nil {
			return nil, fmt.Errorf("Import not supported for this IAM resource")
		}
		config := m.(*Config)
		s := strings.Fields(d.Id())
		if len(s) != 2 {
			d.SetId("")
			return nil, fmt.Errorf("Import not supported for this IAM resource")
		}
		id, roleID := s[0], s[1]

		// Set the ID only to the first part so all IAM types can share the same resourceIDParserFunc.
		d.SetId(id)
		d.Set("role", roleID)
		err := resourceIDParser(d, config)
		if err != nil {
			return nil, err
		}

		// Set the ID again so that the ID matches the ID it would have if it had been created via TF.
		// Use the current ID in case it changed in the resourceIDParserFunc.
		d.SetId(d.Id() + "/" + roleID)
		return []*schema.ResourceData{d}, nil
	}
}

// all bindings use same Role
func getResourceIamBindings(d *schema.ResourceData) []*access.AccessBinding {
	members := d.Get("members").(*schema.Set)
	role := d.Get("role").(string)

	result := make([]*access.AccessBinding, members.Len())

	for i, member := range convertStringSet(members) {
		result[i] = roleMemberToAccessBinding(role, member)
	}
	return result
}
