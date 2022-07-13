package yandex

import (
	"fmt"
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

func resourceAccessBinding(parentSpecificSchema map[string]*schema.Schema, newUpdaterFunc newResourceIamUpdaterFunc) *schema.Resource {
	return &schema.Resource{
		Create: resourceAccessBindingCreate(newUpdaterFunc),
		Read:   resourceAccessBindingRead(newUpdaterFunc, true),
		Update: resourceAccessBindingUpdate(newUpdaterFunc),
		Delete: resourceAccessBindingDelete(newUpdaterFunc),
		Schema: mergeSchemas(accessBindingSchema, parentSpecificSchema),
	}
}

func resourceIamBindingWithImport(parentSpecificSchema map[string]*schema.Schema, newUpdaterFunc newResourceIamUpdaterFunc, resourceIDParser resourceIDParserFunc) *schema.Resource {
	r := resourceAccessBinding(parentSpecificSchema, newUpdaterFunc)
	r.Importer = &schema.ResourceImporter{
		State: iamBindingImport(resourceIDParser),
	}
	return r
}

func resourceAccessBindingCreate(newUpdaterFunc newResourceIamUpdaterFunc) schema.CreateFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return err
		}

		p := getResourceIamBindings(d)
		err = iamPolicyReadModifyWrite(updater, func(ep *Policy) error {
			// Creating a binding does not remove existing members if they are not in the provided members list.
			// This prevents removing existing permission without the user's knowledge.
			// Instead, a diff is shown in that case after creation. Subsequent calls to update will remove any
			// existing members not present in the provided list.
			ep.Bindings = mergeBindings(append(ep.Bindings, p...))
			return nil
		})
		if err != nil {
			return err
		}

		role := p[0].RoleId
		d.SetId(updater.GetResourceID() + "/" + role)

		if v, ok := d.GetOk("sleep_after"); ok {
			time.Sleep(time.Second * time.Duration(v.(int)))
		}

		return resourceAccessBindingRead(newUpdaterFunc, true)(d, meta)
	}
}

func resourceAccessBindingRead(newUpdaterFunc newResourceIamUpdaterFunc, check bool) schema.ReadFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return err
		}

		role := d.Get("role").(string)
		eBindings := getResourceIamBindings(d)

		p, err := updater.GetResourceIamPolicy()
		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				return fmt.Errorf("Binding for role %q not found for non-existent resource %s.", role, updater.DescribeResource())
			}

			return err
		}
		log.Printf("[DEBUG]: Retrieved policy for %s: %+v", updater.DescribeResource(), p)

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

		if check && len(mBindings) == 0 {
			return fmt.Errorf("Binding for role %q not found in policy for %s.", role, updater.DescribeResource())
		}

		if err := d.Set("members", roleToMembersList(role, mBindings)); err != nil {
			return err
		}
		return nil
	}
}

func resourceAccessBindingUpdate(newUpdaterFunc newResourceIamUpdaterFunc) schema.UpdateFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return err
		}

		bindings := getResourceIamBindings(d)
		role := d.Get("role").(string)

		err = iamPolicyReadModifyWrite(updater, func(p *Policy) error {
			p.Bindings = removeRoleFromBindings(role, p.Bindings)
			p.Bindings = append(p.Bindings, bindings...)
			return nil
		})
		if err != nil {
			return err
		}

		return resourceAccessBindingRead(newUpdaterFunc, true)(d, meta)
	}
}

func resourceAccessBindingDelete(newUpdaterFunc newResourceIamUpdaterFunc) schema.DeleteFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return err
		}

		binding := getResourceIamBindings(d)
		role := binding[0].RoleId

		err = iamPolicyReadModifyWrite(updater, func(p *Policy) error {
			p.Bindings = removeRoleFromBindings(role, p.Bindings)
			return nil
		})

		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				log.Printf("[DEBUG]: Resource %s is missing or deleted, marking policy binding as deleted", updater.DescribeResource())
				return nil
			}
			return err
		}

		return resourceAccessBindingRead(newUpdaterFunc, false)(d, meta)
	}
}

func iamBindingImport(resourceIDParser resourceIDParserFunc) schema.StateFunc {
	return func(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
		if resourceIDParser == nil {
			return nil, fmt.Errorf("Import not supported for this IAM resource")
		}
		config := m.(*Config)
		s := strings.Fields(d.Id())
		if len(s) != 2 {
			d.SetId("")
			return nil, fmt.Errorf("Wrong number of parts to Binding id %s; expected 'resource_name role'", s)
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
