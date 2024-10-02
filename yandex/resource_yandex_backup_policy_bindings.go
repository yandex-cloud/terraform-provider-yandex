package yandex

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	backuppb "github.com/yandex-cloud/go-genproto/yandex/cloud/backup/v1"
	"google.golang.org/grpc/codes"
)

func resourceYandexBackupPolicyBindings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexBackupPolicyBindingsCreate,
		ReadContext:   resourceYandexBackupPolicyBindingsRead,
		DeleteContext: resourceYandexBackupPolicyBindingsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexBackupDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexBackupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexBackupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexBackupDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"policy_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// COMPUTED ONLY VALUES

			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"processing": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexBackupPolicyBindingsCreate(ctx context.Context, d *schema.ResourceData, meta any) (diagnostics diag.Diagnostics) {
	config := meta.(*Config)

	err := checkBackupProviderActivated(ctx, config)
	if err != nil {
		return diag.Errorf("Listing active Cloud Backup providers: %s", err)
	}

	policyID := d.Get("policy_id").(string)
	instanceID := d.Get("instance_id").(string)
	id := makeBackupPolicyBindingsID(policyID, instanceID)

	log.Printf("[INFO]: Starting to create Cloud Backup Policy Bindings with id=%q", id)

	// Should not wait for operation completeness, since policy can be applied after a long period of time
	operation, err := createBackupPolicyBindingsWithRetry(ctx, config, policyID, instanceID)
	if err != nil {
		return diag.Errorf("Requesting API to create Cloud Backup Policy Bindings: %s", err)
	}

	d.SetId(id)
	log.Printf("[INFO]: Created Cloud Backup Bindings with id=%q, operation_id=%q", d.Id(), operation.Id())

	return resourceYandexBackupPolicyBindingsRead(ctx, d, meta)
}

func resourceYandexBackupPolicyBindingsRead(ctx context.Context, d *schema.ResourceData, meta any) (diagnostics diag.Diagnostics) {
	config := meta.(*Config)

	log.Printf("[INFO]: Starting to fetch Cloud Backup Policy Bindings with id=%q", d.Id())

	policyID, instanceID, err := parseBackupPolicyBindingsID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	application, err := getBackupPolicyApplication(ctx, config, policyID, instanceID)
	if err != nil {
		if isStatusWithCode(err, codes.NotFound) || errors.Is(err, errBackupPolicyBindingsNotFound) {
			log.Printf("[INFO]: Policy binding with id=%q does not exist, removing from state.", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	log.Printf("[INFO]: Fetched Cloud Backup Policy Bindings application %v", application.String())

	if err = flattenBackupPolicyApplication(d, application); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexBackupPolicyBindingsDelete(ctx context.Context, d *schema.ResourceData, meta any) (diagnostics diag.Diagnostics) {
	config := meta.(*Config)

	log.Printf("[INFO]: Starting to delete Cloud Backup Policy Bindings with id=%q", d.Id())

	policyID, instanceID, err := parseBackupPolicyBindingsID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	operation, err := config.sdk.WrapOperation(config.sdk.Backup().Policy().Revoke(ctx, &backuppb.RevokeRequest{
		PolicyId:          policyID,
		ComputeInstanceId: instanceID,
	}))
	if err != nil {
		err = handleNotFoundError(err, d, d.Id())
		return diag.FromErr(err)
	}

	err = operation.Wait(ctx)
	if err != nil {
		return diag.Errorf("waiting operation for completes: %s", err)
	}

	return resourceYandexBackupPolicyBindingsRead(ctx, d, meta)
}
