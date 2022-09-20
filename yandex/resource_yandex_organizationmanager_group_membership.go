package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const yandexOrganizationManagerGroupMembershipDefaultTimeout = 1 * time.Minute
const maximumUpdateDeltas = 1000

func resourceYandexOrganizationManagerGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexOrganizationManagerGroupMembershipCreate,
		ReadContext:   resourceYandexOrganizationManagerGroupMembershipRead,
		UpdateContext: resourceYandexOrganizationManagerGroupMembershipUpdate,
		DeleteContext: resourceYandexOrganizationManagerGroupMembershipDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexOrganizationManagerGroupMembershipDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexOrganizationManagerGroupMembershipDefaultTimeout),
			Update: schema.DefaultTimeout(yandexOrganizationManagerGroupMembershipDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexOrganizationManagerGroupMembershipDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"members": {
				Type:     schema.TypeSet,
				Set:      schema.HashString,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// Min returns the smaller of x or y.
func Min(x, y int64) int64 {
	if x > y {
		return y
	}
	return x
}

func updateGroupMembers(context context.Context, groupID string, members *schema.Set, meta interface{}, action organizationmanager.MemberDelta_MemberAction) error {
	// Nothing to update.
	if members.Len() == 0 {
		return nil
	}

	config := meta.(*Config)

	memberDeltas := make([]*organizationmanager.MemberDelta, 0, Min(int64(members.Len()), maximumUpdateDeltas))
	for i, member := range members.List() {
		memberDeltas = append(memberDeltas, &organizationmanager.MemberDelta{
			Action:    action,
			SubjectId: member.(string),
		})
		// API will reject requests with more than 1000 deltas.
		if len(memberDeltas) == maximumUpdateDeltas || i+1 == members.Len() {
			req := organizationmanager.UpdateGroupMembersRequest{
				GroupId:      groupID,
				MemberDeltas: memberDeltas,
			}

			op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().Group().UpdateMembers(context, &req))
			if err != nil {
				return fmt.Errorf("Error while requesting API to create GroupMembership: %s", err)
			}

			protoMetadata, err := op.Metadata()
			if err != nil {
				return fmt.Errorf("Error while get GroupMembership create operation metadata: %s", err)
			}

			// Nothing of value in UpdateGroupMembersMetadata.
			_, ok := protoMetadata.(*organizationmanager.UpdateGroupMembersMetadata)
			if !ok {
				return fmt.Errorf("could not get GroupMembership from create operation metadata")
			}

			err = op.Wait(context)
			if err != nil {
				return fmt.Errorf("Error while waiting operation to create GroupMembership: %s", err)
			}

			if _, err := op.Response(); err != nil {
				return fmt.Errorf("GroupMembership creation failed: %s", err)
			}
			memberDeltas = memberDeltas[:0]
		}
	}

	return nil
}

func resourceYandexOrganizationManagerGroupMembershipCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)
	members := d.Get("members").(*schema.Set)
	err := updateGroupMembers(context, groupID, members, meta, organizationmanager.MemberDelta_ADD)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("group-membership/" + groupID)
	return nil
}

func resourceYandexOrganizationManagerGroupMembershipRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	cloudMembers, err := getGroupMembers(context, config, d.Get("group_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	cloudMembersS := schema.NewSet(schema.HashString, nil)
	for _, member := range cloudMembers {
		cloudMembersS.Add(member.SubjectId)
	}
	currentMembers := d.Get("members").(*schema.Set)
	members := cloudMembersS.Intersection(currentMembers)
	return diag.FromErr(d.Set("members", members))
}

func resourceYandexOrganizationManagerGroupMembershipUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	oldM, newM := d.GetChange("members")
	oldS := oldM.(*schema.Set)
	newS := newM.(*schema.Set)

	groupID := d.Get("group_id").(string)

	// Remove the ones that have been deleted.
	removedS := oldS.Difference(newS)
	err := updateGroupMembers(context, groupID, removedS, meta, organizationmanager.MemberDelta_REMOVE)
	if err != nil {
		return diag.FromErr(err)
	}

	// Add the new ones.
	addedS := newS.Difference(oldS)
	return diag.FromErr(updateGroupMembers(context, groupID, addedS, meta, organizationmanager.MemberDelta_ADD))
}

func resourceYandexOrganizationManagerGroupMembershipDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)
	members := d.Get("members").(*schema.Set)
	return diag.FromErr(updateGroupMembers(context, groupID, members, meta, organizationmanager.MemberDelta_REMOVE))
}
