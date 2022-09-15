package yandex

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"github.com/yandex-cloud/go-sdk/operation"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexContainerRepositoryLifecyclePolicyDefaultTimeout = 5 * time.Minute

func resourceYandexContainerRepositoryLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexContainerRepositoryLifecyclePolicyCreate,
		ReadContext:   resourceYandexContainerRepositoryLifycyclePolicyRead,
		UpdateContext: resourceYandexContainerRepositoryLifycyclePolicyUpdate,
		DeleteContext: resourceYandexContainerRepositoryLifycyclePolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(yandexContainerRepositoryLifecyclePolicyDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"repository_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"status": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"rule": {
				Type:             schema.TypeList,
				Computed:         true,
				Optional:         true,
				DiffSuppressFunc: shouldSuppressDiffForContainerRepositoryLifecyclePolicyRules,

				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:         schema.TypeString,
							Computed:     true,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},

						"expire_period": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateFunc:     validateParsableValue(parseDuration),
							DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
							Default:          "",
						},

						"tag_regexp": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
							Default:      "",
						},

						"untagged": {
							Type:     schema.TypeBool,
							Computed: true,
							Optional: true,
						},

						"retained_top": {
							Type:         schema.TypeInt,
							Computed:     true,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
		},
	}
}

func shouldSuppressDiffForContainerRepositoryLifecyclePolicyRules(_, _, _ string, d *schema.ResourceData) bool {
	oldRules, newRules := d.GetChange("rule")

	old := oldRules.([]interface{})
	new := newRules.([]interface{})

	if old == nil || new == nil {
		return false
	}

	if len(old) != len(new) {
		return false
	}

	for _, o := range old {
		var foundEqual bool

		for _, n := range new {
			expand := func(m map[string]interface{}) containerregistry.LifecycleRule {
				duration, _ := parseDuration(m["expire_period"].(string))

				return containerregistry.LifecycleRule{
					Description:  m["description"].(string),
					ExpirePeriod: duration,
					TagRegexp:    m["tag_regexp"].(string),
					Untagged:     m["untagged"].(bool),
					RetainedTop:  int64(m["retained_top"].(int)),
				}
			}

			or := expand(o.(map[string]interface{}))
			nr := expand(n.(map[string]interface{}))

			if reflect.DeepEqual(or, nr) {
				foundEqual = true
				break
			}
		}

		if !foundEqual {
			return false
		}
	}

	return true
}

func resourceYandexContainerRepositoryLifecyclePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Creating Container Repository Lifecycle Policy (repository_id: %v)", d.Get("repository_id").(string))

	status, err := parseContainerRepositoryLifecyclePolicyStatus(d.Get("status").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	rules, err := expandContainerRepositoryLifecyclePolicyRules(d)
	if err != nil {
		return diag.FromErr(err)
	}

	createLifecyclePolicyRequest := &containerregistry.CreateLifecyclePolicyRequest{
		RepositoryId: d.Get("repository_id").(string),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Status:       status,
		Rules:        rules,
	}

	config := meta.(*Config)
	lifecyclePolicyService := config.sdk.ContainerRegistry().LifecyclePolicy()
	operation, err := config.sdk.WrapOperation(lifecyclePolicyService.Create(ctx, createLifecyclePolicyRequest))
	if err != nil {
		return diag.FromErr(err)
	}

	lifecyclePolicyId, err := unwrapContainerRepositoryLifecyclePolicyID(operation)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(lifecyclePolicyId)

	if err := waitContainerRepositoryLifecyclePolicyOperation(ctx, operation); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished creating Container Repository Lifecycle Policy (id: %v)", lifecyclePolicyId)

	return resourceYandexContainerRepositoryLifycyclePolicyRead(ctx, d, meta)
}

func resourceYandexContainerRepositoryLifycyclePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lifecyclePolicy, err := getLifecyclePolicy(ctx, d.Id(), meta)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", lifecyclePolicy.Name)
	d.Set("repository_id", lifecyclePolicy.RepositoryId)
	d.Set("created_at", getTimestamp(lifecyclePolicy.CreatedAt))
	d.Set("description", lifecyclePolicy.Description)
	d.Set("status", strings.ToLower(lifecyclePolicy.Status.String()))

	lifecyclePolicyRules := flattenContainerRepositoryLifecyclePolicyRules(lifecyclePolicy.Rules)
	if err := d.Set("rule", lifecyclePolicyRules); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexContainerRepositoryLifycyclePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lifecyclePolicyUpdateRequest := &containerregistry.UpdateLifecyclePolicyRequest{
		LifecyclePolicyId: d.Id(),
		UpdateMask:        &field_mask.FieldMask{},
	}

	if d.HasChange("name") {
		log.Printf("[DEBUG] name has changed: %v", d.Get("name").(string))

		lifecyclePolicyUpdateRequest.SetName(d.Get("name").(string))
		lifecyclePolicyUpdateRequest.UpdateMask.Paths = append(lifecyclePolicyUpdateRequest.UpdateMask.Paths, "name")
	}

	if d.HasChange("description") {
		log.Printf("[DEBUG] description has changed: %v", d.Get("description").(string))

		lifecyclePolicyUpdateRequest.SetDescription(d.Get("description").(string))
		lifecyclePolicyUpdateRequest.UpdateMask.Paths = append(lifecyclePolicyUpdateRequest.UpdateMask.Paths, "description")
	}

	if d.HasChange("status") {
		log.Printf("[DEBUG] status has changed: %v", d.Get("status").(string))

		status, err := parseContainerRepositoryLifecyclePolicyStatus(d.Get("status").(string))
		if err != nil {
			return diag.FromErr(err)
		}

		lifecyclePolicyUpdateRequest.SetStatus(status)
		lifecyclePolicyUpdateRequest.UpdateMask.Paths = append(lifecyclePolicyUpdateRequest.UpdateMask.Paths, "status")
	}

	if d.HasChange("rule") {
		log.Printf("[DEBUG] rules has changed: %v", d.Get("rules"))

		rules, err := expandContainerRepositoryLifecyclePolicyRules(d)
		if err != nil {
			return diag.FromErr(err)
		}

		lifecyclePolicyUpdateRequest.SetRules(rules)
		lifecyclePolicyUpdateRequest.UpdateMask.Paths = append(lifecyclePolicyUpdateRequest.UpdateMask.Paths, "rules")
	}

	config := meta.(*Config)
	lifecyclePolicyService := config.sdk.ContainerRegistry().LifecyclePolicy()
	operation, err := config.sdk.WrapOperation(lifecyclePolicyService.Update(ctx, lifecyclePolicyUpdateRequest))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := waitContainerRepositoryLifecyclePolicyOperation(ctx, operation); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexContainerRepositoryLifycyclePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lifecyclePolicyId := d.Id()

	log.Printf("[DEBUG] Deleting Container Repository Lifecycle Policy (id: %v)", lifecyclePolicyId)

	lifecyclePolicyDeleteRequest := &containerregistry.DeleteLifecyclePolicyRequest{LifecyclePolicyId: lifecyclePolicyId}

	config := meta.(*Config)
	lifecyclePolicyService := config.sdk.ContainerRegistry().LifecyclePolicy()
	operation, err := config.sdk.WrapOperation(lifecyclePolicyService.Delete(ctx, lifecyclePolicyDeleteRequest))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := waitContainerRepositoryLifecyclePolicyOperation(ctx, operation); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished deleting Container Repository Lifecycle Policy (id: %v)", lifecyclePolicyId)

	return nil
}

func unwrapContainerRepositoryLifecyclePolicyID(operation *operation.Operation) (string, error) {
	protoMetadata, err := operation.Metadata()
	if err != nil {
		return "", fmt.Errorf("failed to get Lifecycle Policy create operation metadata: %v", err)
	}

	createLifecyclePolicyMetadata, ok := protoMetadata.(*containerregistry.CreateLifecyclePolicyMetadata)
	if !ok {
		return "", fmt.Errorf("failed to get Lifecycle Policy ID from create operation metadata")
	}

	return createLifecyclePolicyMetadata.LifecyclePolicyId, nil
}

func waitContainerRepositoryLifecyclePolicyOperation(ctx context.Context, operation *operation.Operation) error {
	if err := operation.Wait(ctx); err != nil {
		return fmt.Errorf("error while waiting operation to create Lifecycle Policy: %s", err)
	}

	if _, err := operation.Response(); err != nil {
		return fmt.Errorf("failed to create Lifecycle Policy: %s", err)
	}

	return nil
}

func parseContainerRepositoryLifecyclePolicyStatus(str string) (containerregistry.LifecyclePolicy_Status, error) {
	status, ok := containerregistry.LifecyclePolicy_Status_value[strings.ToUpper(str)]
	if !ok {
		allowedStatuses := []string{containerregistry.LifecyclePolicy_ACTIVE.String(), containerregistry.LifecyclePolicy_DISABLED.String()}
		return containerregistry.LifecyclePolicy_STATUS_UNSPECIFIED,
			fmt.Errorf("provided invalid status (%v) for Lifecycle Policy, must be one of: %v", str, getJoinedKeys(allowedStatuses))
	}

	return containerregistry.LifecyclePolicy_Status(status), nil
}
