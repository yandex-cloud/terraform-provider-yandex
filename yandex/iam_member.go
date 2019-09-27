package yandex

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

func iamMemberImport(resourceIDParser resourceIDParserFunc) schema.StateFunc {
	return func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
		if resourceIDParser == nil {
			return nil, errors.New("Import not supported for this IAM resource")
		}
		config := meta.(*Config)
		s := strings.Fields(d.Id())
		if len(s) != 3 {
			d.SetId("")
			return nil, fmt.Errorf("Wrong number of parts to Member id %s; expected 'resource_name role username'", s)
		}
		id, role, member := s[0], s[1], s[2]

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

func resourceIamMember(parentSpecificSchema map[string]*schema.Schema, newUpdaterFunc newResourceIamUpdaterFunc) *schema.Resource {
	return &schema.Resource{
		Create: resourceIamMemberCreate(newUpdaterFunc),
		Read:   resourceIamMemberRead(newUpdaterFunc),
		Delete: resourceIamMemberDelete(newUpdaterFunc),

		Schema: mergeSchemas(IamMemberBaseSchema, parentSpecificSchema),
	}
}

func resourceIamMemberWithImport(parentSpecificSchema map[string]*schema.Schema, newUpdaterFunc newResourceIamUpdaterFunc, resourceIDParser resourceIDParserFunc) *schema.Resource {
	r := resourceIamMember(parentSpecificSchema, newUpdaterFunc)
	r.Importer = &schema.ResourceImporter{
		State: iamMemberImport(resourceIDParser),
	}
	return r
}

func getResourceIamMember(d *schema.ResourceData) *access.AccessBinding {
	member := d.Get("member").(string)
	role := d.Get("role").(string)

	return roleMemberToAccessBinding(role, member)
}

func resourceIamMemberCreate(newUpdaterFunc newResourceIamUpdaterFunc) schema.CreateFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return err
		}

		p := getResourceIamMember(d)
		err = iamPolicyReadModifyWrite(updater, func(ep *Policy) error {
			// Merge the bindings together
			ep.Bindings = mergeBindings(append(ep.Bindings, p))
			return nil
		})
		if err != nil {
			return err
		}
		d.SetId(updater.GetResourceID() + "/" + p.RoleId + "/" + canonicalMember(p))

		if v, ok := d.GetOk("sleep_after"); ok {
			time.Sleep(time.Second * time.Duration(v.(int)))
		}

		return resourceIamMemberRead(newUpdaterFunc)(d, meta)
	}
}

func resourceIamMemberRead(newUpdaterFunc newResourceIamUpdaterFunc) schema.ReadFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return err
		}

		eMember := getResourceIamMember(d)
		p, err := updater.GetResourceIamPolicy()
		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				log.Printf("[DEBUG]: Binding of member %q with role %q does not exist for non-existent resource %s, removing from state.", canonicalMember(eMember), eMember.RoleId, updater.DescribeResource())
				d.SetId("")
				return nil
			}
			return err
		}
		log.Printf("[DEBUG]: Retrieved policy for %s: %+v\n", updater.DescribeResource(), p)
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
			log.Printf("[DEBUG]: Binding for role %q does not exist in policy of %s, removing member %q from state.", eMember.RoleId, updater.DescribeResource(), canonicalMember(eMember))
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
			log.Printf("[DEBUG]: Member %q for binding for role %q does not exist in policy of %s, removing from state.", canonicalMember(eMember), eMember.RoleId, updater.DescribeResource())
			d.SetId("")
			return nil
		}
		d.Set("member", member)
		d.Set("role", role)
		return nil
	}
}

func resourceIamMemberDelete(newUpdaterFunc newResourceIamUpdaterFunc) schema.DeleteFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return err
		}

		member := getResourceIamMember(d)
		role := member.RoleId

		err = iamPolicyReadModifyWrite(updater, func(p *Policy) error {
			var toRemovePos []int
			for pos, b := range p.Bindings {
				if b.RoleId != role {
					continue
				}
				if b.Subject.Type != member.Subject.Type || b.Subject.Id != member.Subject.Id {
					continue
				}
				toRemovePos = append(toRemovePos, pos)
				break
			}

			if len(toRemovePos) == 0 {
				log.Printf("[DEBUG]: Binding for role %q does not exist in policy of project %q, so member %q can't be on it.", member.RoleId, updater.GetResourceID(), canonicalMember(member))
				return nil
			}

			for _, pos := range toRemovePos {
				p.Bindings = append(p.Bindings[:pos], p.Bindings[pos+1:]...)
			}

			return nil

		})
		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				log.Printf("[DEBUG]: Member %q for binding for role %q does not exist for non-existent resource %q.", canonicalMember(member), member.RoleId, updater.GetResourceID())
				return nil
			}
			return err
		}

		return resourceIamMemberRead(newUpdaterFunc)(d, meta)
	}
}

func canonicalMember(ab *access.AccessBinding) string {
	return ab.Subject.Type + ":" + ab.Subject.Id
}
