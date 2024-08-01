package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	"log"
)

// Logic to store sensitive values of resources into Lockbox, to avoid leaking those values to the Terraform state

var lockboxOutputEntryKeyPrefix = "entry_for_"
var lockboxOutputAttr = "output_to_lockbox"
var lockboxOutputVersionIdAttr = lockboxOutputAttr + "_version_id"

// ExtendWithOutputToLockbox adds output_to_lockbox attributes, used by ManageOutputToLockbox and ValidateChangeInOutputToLockbox
func ExtendWithOutputToLockbox(resourceSchema map[string]*schema.Schema, sensitiveAttrs []string) map[string]*schema.Schema {
	outputToLockboxSchema := map[string]*schema.Schema{
		"secret_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "secret where to add the version with the sensitive values",
		},
	}

	for _, sensitiveAttr := range sensitiveAttrs {
		outputToLockboxSchema[outputToLockboxAttrForSensitiveAttr(sensitiveAttr)] = &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			Description: "entry that will store the value of " + sensitiveAttr,
		}
	}

	resourceSchema[lockboxOutputAttr] = &schema.Schema{
		Type:        schema.TypeList,
		Description: "option to create a Lockbox secret version from sensitive outputs",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: outputToLockboxSchema,
		},
	}

	resourceSchema[lockboxOutputVersionIdAttr] = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "version generated, that will contain the sensitive outputs",
	}

	return resourceSchema
}

// ManageOutputToLockbox moves sensitive values between state and Lockbox.
// If output_to_lockbox is added: moves sensitive attributes to a Lockbox secret, and removes the sensitive values from the state.
// If output_to_lockbox is removed: restores the sensitive attributes from the Lockbox secret.
// This method should be called at the end of the resource Read method.
func ManageOutputToLockbox(ctx context.Context, d *schema.ResourceData, config *Config, sensitiveAttrs []string) error {
	secretOldID, secretID := getChangeAsString(d, lockboxOutputAttr+".0.secret_id")
	if secretOldID == secretID {
		return nil
	}

	if secretOldID == "" {
		log.Printf("[DEBUG] output_to_lockbox added, so move sensitive attributes %v to a new version in secret: %s", sensitiveAttrs, secretID)
		return moveSensitiveAttrsToNewLockboxVersion(ctx, d, config, sensitiveAttrs, secretID)
	} else if secretID == "" {
		versionOldID, _ := getChangeAsString(d, lockboxOutputVersionIdAttr)
		log.Printf("[DEBUG] output_to_lockbox removed, so restoring sensitive fields %v from secret/version: %s/%s", sensitiveAttrs, secretOldID, versionOldID)
		return restoreSensitiveValuesFromLockboxVersion(ctx, d, config, sensitiveAttrs, secretOldID, versionOldID)
	} else {
		// This might happen if we forgot to call ValidateChangeInOutputToLockbox from the Update method of the resource
		return fmt.Errorf("unexpected change in secret_id, from %s to %s; this seems like a bug in the resource", secretOldID, secretID)
	}
}

// ValidateChangeInOutputToLockbox looks for invalid changes. Should be called in the resource Update method.
func ValidateChangeInOutputToLockbox(d *schema.ResourceData, sensitiveAttrs []string) error {
	secretOldID, secretID := getChangeAsString(d, lockboxOutputAttr+".0.secret_id")
	commonMessage := "first remove output_to_lockbox block to restore sensitive values from the Lockbox secret to the resource state"
	if secretOldID != secretID {
		// check that secret wasn't changed (we only allow adding output_to_lockbox or removing it)
		if secretOldID != "" && secretID != "" {
			return fmt.Errorf("changing secret_id is not allowed; " + commonMessage)
		}
		return nil
	}
	if secretID == "" {
		return nil
	}
	// check that entry keys are not changed (we only allow adding output_to_lockbox or removing it)
	for _, sensitiveAttr := range sensitiveAttrs {
		entryKeyOld, entryKeyNew := getEntryKeyForSensitiveAttr(d, sensitiveAttr)
		if entryKeyOld != entryKeyNew {
			return fmt.Errorf("changing entry keys is not allowed; " + commonMessage)
		}
	}
	return nil
}

// creates a new Lockbox version with the values of the sensitive fields, and clears those sensitive values from the state
func moveSensitiveAttrsToNewLockboxVersion(ctx context.Context, d *schema.ResourceData, config *Config, sensitiveAttrs []string, secretID string) error {
	var entries []*lockbox.PayloadEntryChange
	for _, sensitiveAttr := range sensitiveAttrs {
		entry := new(lockbox.PayloadEntryChange)
		_, entryKey := getEntryKeyForSensitiveAttr(d, sensitiveAttr) // get new value, since output_to_lockbox was added
		log.Printf("[DEBUG] - sensitive attribute '%s' will be stored in entry key '%s'", sensitiveAttr, entryKey)
		entry.SetKey(entryKey)
		entry.SetTextValue(d.Get(sensitiveAttr).(string))
		entries = append(entries, entry)

		// clear sensitive value from state
		if err := d.Set(sensitiveAttr, nil); err != nil {
			log.Printf("[ERROR] failed to clear sensitive field '%s': %s", sensitiveAttr, err)
			return err
		}
	}

	log.Printf("[DEBUG] adding entries for sensitive attributes %v to secret %s", sensitiveAttrs, secretID)

	lockboxVersion, err := addLockboxVersion(ctx, config, secretID, entries)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] created version %s", lockboxVersion.GetId())

	if err = d.Set(lockboxOutputVersionIdAttr, lockboxVersion.GetId()); err != nil {
		log.Printf("[ERROR] lockbox version %s was created, but it wasn't possible to set field output_to_lockbox_version_id: %v", lockboxVersion.GetId(), err)
		return err
	}
	return nil
}

// retrieves sensitive values from Lockbox version, and puts them into the state
func restoreSensitiveValuesFromLockboxVersion(ctx context.Context, d *schema.ResourceData, config *Config, sensitiveAttrs []string, secretOldID string, versionOldID string) error {
	lockboxVersion, err := getLockboxVersion(ctx, config, secretOldID, versionOldID)
	if err != nil {
		return err
	}

	for _, sensitiveAttr := range sensitiveAttrs {
		entryKey, _ := getEntryKeyForSensitiveAttr(d, sensitiveAttr) // get old value, since output_to_lockbox was removed
		entry := findEntryForKey(entryKey, lockboxVersion)
		if entry != nil {
			sensitiveValue := entry.GetTextValue()
			if err = d.Set(sensitiveAttr, sensitiveValue); err != nil {
				log.Printf("[ERROR] failed to restore sensitive field '%s': %s", sensitiveAttr, err)
				return err
			}
		} else {
			log.Printf("[WARN] couldn't restore value for sensitive attribute '%s' because entry key '%s' doesn't exist in secret/version: %s/%s", sensitiveAttr, entryKey, secretOldID, versionOldID)
		}
	}
	return nil
}

// entry key that stores the value for sensitiveAttr (old/new value)
func getEntryKeyForSensitiveAttr(d *schema.ResourceData, sensitiveAttr string) (string, string) {
	outputAttr := outputToLockboxAttrForSensitiveAttr(sensitiveAttr)
	return getChangeAsString(d, lockboxOutputAttr+".0."+outputAttr)
}

// name of the attribute (inside output_to_lockbox) that indicates the entry key for sensitiveAttr
func outputToLockboxAttrForSensitiveAttr(sensitiveAttr string) string {
	return lockboxOutputEntryKeyPrefix + sensitiveAttr
}

func getLockboxVersion(ctx context.Context, config *Config, secretID, versionID string) (*lockbox.Payload, error) {
	req := &lockbox.GetPayloadRequest{
		SecretId:  secretID,
		VersionId: versionID,
	}
	log.Printf("[DEBUG] GetPayloadRequest: %s", protoDump(req))

	return config.sdk.LockboxPayload().Payload().Get(ctx, req)
}

func findEntryForKey(key string, payload *lockbox.Payload) *lockbox.Payload_Entry {
	for _, entry := range payload.Entries {
		if entry.GetKey() == key {
			return entry
		}
	}
	return nil
}

func addLockboxVersion(ctx context.Context, config *Config, secretID string, entries []*lockbox.PayloadEntryChange) (*lockbox.Version, error) {
	req := &lockbox.AddVersionRequest{
		SecretId:       secretID,
		PayloadEntries: entries,
	}

	op, err := config.sdk.WrapOperation(config.sdk.LockboxSecret().Secret().AddVersion(ctx, req))
	if err != nil {
		return nil, err
	}
	err = op.Wait(ctx)
	if err != nil {
		return nil, err
	}
	version, err := op.Response()
	if err != nil {
		return nil, err
	}
	return version.(*lockbox.Version), nil
}

func getChangeAsString(d *schema.ResourceData, key string) (string, string) {
	oldValue, newValue := d.GetChange(key)
	return castToStringOrEmpty(oldValue), castToStringOrEmpty(newValue)
}

func castToStringOrEmpty(v interface{}) string {
	if v == nil {
		return ""
	}
	return v.(string)
}
