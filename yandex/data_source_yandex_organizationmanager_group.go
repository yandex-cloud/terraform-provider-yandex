package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	sdkresolversv2 "github.com/yandex-cloud/go-sdk/v2/pkg/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexOrganizationManagerGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Cloud Organization Manager Group. For more information, see [the official documentation](https://yandex.cloud/docs/organization/manage-groups).\n\n~> One of `group_id` or `name` should be specified.",

		ReadContext: dataSourceYandexOrganizationManagerGroupRead,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:        schema.TypeString,
				Description: "ID of a Group.",
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},
			"organization_id": {
				Type:        schema.TypeString,
				Description: "Organization that the Group belongs to. If value is omitted, the default provider organization is used.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"members": {
				Type:        schema.TypeList,
				Description: "A list of members of the Group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "The ID of the member.",
							Computed:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "The type of the member.",
							Computed:    true,
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
	client := organizationmanagersdk.NewGroupClient(config.SDK)

	resolver := organizationmanagersdk.GroupResolver(groupName, client, sdkresolversv2.OrganizationID(organizationID))
	if err := resolver.Run(ctx); err != nil {
		return "", err
	}
	if err := resolver.Err(); err != nil {
		return "", err
	}
	return resolver.ID(), nil
}

func getGroupMembers(context context.Context, config *Config, groupID string) ([]*organizationmanager.GroupMember, error) {
	client := organizationmanagersdk.NewGroupClient(config.SDK)

	var token string
	result := make([]*organizationmanager.GroupMember, 0, 100)
	for {
		req := &organizationmanager.ListGroupMembersRequest{
			GroupId:   groupID,
			PageSize:  defaultListSize,
			PageToken: token,
		}

		resp, err := client.ListMembers(context, req)
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

func flattenGroup(context context.Context, groupID string, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	client := organizationmanagersdk.NewGroupClient(config.SDK)

	group, err := client.Get(context,
		&organizationmanager.GetGroupRequest{
			GroupId: groupID,
		})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Group %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(group.CreatedAt))
	d.Set("name", group.Name)
	d.Set("organization_id", group.OrganizationId)
	return d.Set("description", group.Description)
}
