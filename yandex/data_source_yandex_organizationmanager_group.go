package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexOrganizationManagerGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexOrganizationManagerGroupRead,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"members": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func dataSourceYandexOrganizationManagerGroupRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "group_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	organizationID, err := getOrganizationID(d, config)
	if err != nil {
		return diag.Errorf("Error getting organization ID while reading Group: %s", err)
	}

	groupID := d.Get("group_id").(string)
	federationName, ok := d.GetOk("name")

	if ok {
		groupID, err = resolveGroupIDByName(context, config, federationName.(string), organizationID)
		if err != nil {
			return diag.Errorf("failed to resolve data source Group by name: %v", err)
		}
	}

	err = flattenGroup(context, groupID, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	cloudMembers, err := getGroupMembers(context, config, groupID)
	if err != nil {
		return diag.FromErr(err)
	}
	var members []interface{}
	for _, m := range cloudMembers {
		member := map[string]interface{}{
			"id":   m.SubjectId,
			"type": m.SubjectType,
		}
		members = append(members, member)
	}
	err = d.Set("members", members)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(groupID)
	return nil
}

func resolveGroupIDByName(ctx context.Context, config *Config, groupName, organizationID string) (string, error) {
	var objectID string
	resolver := sdkresolvers.OrganizationGroupResolver(groupName, sdkresolvers.OrganizationID(organizationID), sdkresolvers.Out(&objectID))

	err := config.sdk.Resolve(ctx, resolver)
	if err != nil {
		return "", err
	}

	return objectID, nil
}

func getGroupMembers(context context.Context, config *Config, groupID string) ([]*organizationmanager.GroupMember, error) {
	var token string
	result := make([]*organizationmanager.GroupMember, 0, 100)
	for {
		req := &organizationmanager.ListGroupMembersRequest{
			GroupId:   groupID,
			PageSize:  defaultListSize,
			PageToken: token,
		}

		resp, err := config.sdk.OrganizationManager().Group().ListMembers(context, req)
		if err != nil {
			return nil, err
		}
		result = append(result, resp.Members...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return result, nil
}
