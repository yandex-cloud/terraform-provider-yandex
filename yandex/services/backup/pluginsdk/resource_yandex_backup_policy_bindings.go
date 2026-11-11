package pluginsdk

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	backuppb "github.com/yandex-cloud/go-genproto/yandex/cloud/backup/v1"
	backupsdk "github.com/yandex-cloud/go-sdk/services/backup/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/grpc/codes"
)

func ResourceYandexBackupPolicyBindings() *schema.Resource {
	return &schema.Resource{
		Description:   "Allows management of [Yandex Cloud Attach and Detach VM](https://yandex.cloud/docs/backup/operations/policy-vm/attach-and-detach-vm).\n\n ~> Cloud Backup Provider must be activated in order to manipulate with policies. \n",
		CreateContext: resourceYandexBackupPolicyBindingsCreate,
		ReadContext:   resourceYandexBackupPolicyBindingsRead,
		DeleteContext: resourceYandexBackupPolicyBindingsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(YandexBackupDefaultTimeout),
			Read:   schema.DefaultTimeout(YandexBackupDefaultTimeout),
			Update: schema.DefaultTimeout(YandexBackupDefaultTimeout),
			Delete: schema.DefaultTimeout(YandexBackupDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Compute Cloud instance ID.",
			},

			"policy_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Backup Policy ID.",
			},

			// COMPUTED ONLY VALUES

			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag is specifies whether the policy application is enabled. May be `false` if Processing flag is `true`.",
			},

			"processing": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag that specifies whether the policy is in the process of binding to an instance.",
			},

			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: common.ResourceDescriptions["created_at"],
			},
		},
	}
}

func resourceYandexBackupPolicyBindingsCreate(ctx context.Context, d *schema.ResourceData, meta any) (diagnostics diag.Diagnostics) {
	config := meta.(*providerconfig.Config)

	err := checkBackupProviderActivated(ctx, config)
	if err != nil {
		return diag.Errorf("Listing active Cloud Backup providers: %s", err)
	}

	policyID := d.Get("policy_id").(string)
	instanceID := d.Get("instance_id").(string)
	id := MakeBackupPolicyBindingsID(policyID, instanceID)

	log.Printf("[INFO]: Starting to create Cloud Backup Policy Bindings with id=%q", id)

	// Should not wait for operation completeness, since policy can be applied after a long period of time
	operation, err := createBackupPolicyBindingsWithRetry(ctx, config, policyID, instanceID)
	if err != nil {
		return diag.Errorf("Requesting API to create Cloud Backup Policy Bindings: %s", err)
	}

	d.SetId(id)
	log.Printf("[INFO]: Created Cloud Backup Bindings with id=%q, operation_id=%q", d.Id(), operation.ID())

	return resourceYandexBackupPolicyBindingsRead(ctx, d, meta)
}

func resourceYandexBackupPolicyBindingsRead(ctx context.Context, d *schema.ResourceData, meta any) (diagnostics diag.Diagnostics) {
	config := meta.(*providerconfig.Config)

	log.Printf("[INFO]: Starting to fetch Cloud Backup Policy Bindings with id=%q", d.Id())

	policyID, instanceID, err := ParseBackupPolicyBindingsID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	application, err := GetBackupPolicyApplication(ctx, config, policyID, instanceID)
	if err != nil {
		if sdkcommon.IsStatusWithCode(err, codes.NotFound) || errors.Is(err, errBackupPolicyBindingsNotFound) {
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
	config := meta.(*providerconfig.Config)

	log.Printf("[INFO]: Starting to delete Cloud Backup Policy Bindings with id=%q", d.Id())

	policyID, instanceID, err := ParseBackupPolicyBindingsID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := backupsdk.NewPolicyClient(config.SDK)

	operation, err := client.Revoke(ctx, &backuppb.RevokeRequest{
		PolicyId:          policyID,
		ComputeInstanceId: instanceID,
	})
	if err != nil {
		err = sdkcommon.HandleNotFoundError(err, d, d.Id())
		return diag.FromErr(err)
	}

	_, err = operation.Wait(ctx)
	if err != nil {
		return diag.Errorf("waiting operation for completes: %s", err)
	}

	return resourceYandexBackupPolicyBindingsRead(ctx, d, meta)
}
