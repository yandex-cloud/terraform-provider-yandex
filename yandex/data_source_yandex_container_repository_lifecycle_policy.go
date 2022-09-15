package yandex

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
)

func dataSourceYandexContainerRepositoryLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexContainerRepositoryLifecyclePolicyRead,

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(yandexContainerRepositoryLifecyclePolicyDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"lifecycle_policy_id"},
			},

			"repository_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"lifecycle_policy_id"},
			},

			"lifecycle_policy_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name", "repository_id"},
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"rule": {
				Type:     schema.TypeList,
				Computed: true,

				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"expire_period": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"tag_regexp": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"untagged": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"retained_top": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexContainerRepositoryLifecyclePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := resolveLifecyclePolicyID(ctx, d, meta); err != nil {
		return diag.FromErr(err)
	}

	// ID were put into state in `resolveLifecyclePolicyID` function ^^^
	lifecyclePolicy, err := getLifecyclePolicy(ctx, d.Id(), meta)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", lifecyclePolicy.Name)
	d.Set("lifecycle_policy_id", lifecyclePolicy.Id)
	d.Set("repository_id", lifecyclePolicy.RepositoryId)
	d.Set("created_at", getTimestamp(lifecyclePolicy.CreatedAt))
	d.Set("description", lifecyclePolicy.Description)
	d.Set("status", strings.ToLower(lifecyclePolicy.Status.String()))
	if err := d.Set("rule", flattenContainerRepositoryLifecyclePolicyRules(lifecyclePolicy.Rules)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// resolveLifecyclePolicyID - resolve Lifecycle Policy ID and put it into state.
func resolveLifecyclePolicyID(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	// check one of two required patameters (lifecycle_policy_id or name + repository_id) by which we'll resole lifecycle_policy
	if err := checkOneOf(d, "name", "lifecycle_policy_id"); err != nil {
		return err
	}

	// by ID
	if lpid, ok := d.GetOk("lifecycle_policy_id"); ok {
		d.SetId(lpid.(string))

		return nil
	}

	// by name + repository_id
	if err := checkEveryOf(d, "name", "repository_id"); err != nil {
		return err
	}

	config, ok := meta.(*Config)
	if !ok {
		return errors.New("failed to cast meta to config")
	}

	listLifecyclePoliciesRequest := &containerregistry.ListLifecyclePoliciesRequest{
		Id: &containerregistry.ListLifecyclePoliciesRequest_RepositoryId{
			RepositoryId: d.Get("repository_id").(string),
		},
	}
	lifecyclePolicyIterator := config.sdk.ContainerRegistry().LifecyclePolicy().LifecyclePolicyIterator(ctx, listLifecyclePoliciesRequest)

	for lifecyclePolicyIterator.Next() {
		if err := lifecyclePolicyIterator.Error(); err != nil {
			return err
		}

		lifecyclePolicy := lifecyclePolicyIterator.Value()
		if lifecyclePolicy.Name == d.Get("name").(string) {
			d.SetId(lifecyclePolicy.Id)

			return nil
		}
	}

	return fmt.Errorf("no Lifecycle Policies with name (%v) found in repository (%v)", d.Get("name").(string), d.Get("repository_id").(string))
}

func getLifecyclePolicy(ctx context.Context, lifecyclePolicyID string, meta interface{}) (*containerregistry.LifecyclePolicy, error) {
	if lifecyclePolicyID == "" {
		return nil, errors.New("empty lifecycle_policy_id")
	}

	config, ok := meta.(*Config)
	if !ok {
		return nil, errors.New("failed to cast meta to config")
	}

	lifecyclePolicyService := config.sdk.ContainerRegistry().LifecyclePolicy()
	getLifecyclePolicyRequest := &containerregistry.GetLifecyclePolicyRequest{
		LifecyclePolicyId: lifecyclePolicyID,
	}

	return lifecyclePolicyService.Get(ctx, getLifecyclePolicyRequest)
}
