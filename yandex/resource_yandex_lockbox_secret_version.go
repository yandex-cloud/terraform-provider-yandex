package yandex

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	yandexLockboxSecretVersionDefaultTimeout = 1 * time.Minute
)

// To prevent parallelism when replacing one version with another
var resourceYandexLockboxSecretVersionMutex sync.Mutex

func resourceYandexLockboxSecretVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceYandexLockboxSecretVersionRead,
		CreateContext: resourceYandexLockboxSecretVersionCreate,
		DeleteContext: resourceYandexLockboxSecretVersionDelete,
		// UpdateContext: nil, // updates are not supported, all fields have ForceNew: true

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexLockboxSecretVersionDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexLockboxSecretVersionDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexLockboxSecretVersionDefaultTimeout),
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"entries": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`^([-_./\\@0-9a-zA-Z]+)$`), ""), validation.StringLenBetween(0, 256)),
						},

						"text_value": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(0, 65536),
						},

						"command": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"path": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"env": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Optional: true,
										ForceNew: true,
									},
									"args": {
										Type:     schema.TypeList,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
				ForceNew: true,
				Optional: true,
			},

			"secret_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
		},
	}
}

func resourceYandexLockboxSecretVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	versionPayloadEntries, err := expandLockboxSecretVersionEntriesSlice(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	diagErr := resourceYandexLockboxSecretVersionCreateAux(ctx, versionPayloadEntries, d, config)
	if diagErr != nil {
		return diagErr
	}

	return resourceYandexLockboxSecretVersionRead(ctx, d, meta)
}

func resourceYandexLockboxSecretVersionCreateAux(ctx context.Context, versionPayloadEntries []*lockbox.PayloadEntryChange, d *schema.ResourceData, config *Config) diag.Diagnostics {
	resourceYandexLockboxSecretVersionMutex.Lock()
	defer resourceYandexLockboxSecretVersionMutex.Unlock()

	secret, err := config.sdk.LockboxSecret().Secret().Get(ctx, &lockbox.GetSecretRequest{
		SecretId: d.Get("secret_id").(string),
	})
	if err != nil {
		return diag.Errorf("could not get secret %v: %s", d.Get("secret_id"), err)
	}

	var currentVersionId string

	if secret.CurrentVersion != nil && secret.CurrentVersion.Status == lockbox.Version_ACTIVE {
		currentVersionId = secret.CurrentVersion.Id
	}

	var currentEntries []*lockbox.Payload_Entry

	if currentVersionId != "" {
		getPayloadReq := &lockbox.GetPayloadRequest{
			SecretId:  d.Get("secret_id").(string),
			VersionId: currentVersionId,
		}
		getPayloadReq.VersionId = currentVersionId

		log.Printf("[INFO] getting Lockbox payload (to compare entries): %s", protojson.Format(getPayloadReq))

		payload, err := config.sdk.LockboxPayload().Payload().Get(ctx, getPayloadReq)
		if err != nil {
			return diag.Errorf("could not get payload from secret %v and version %v: %s", getPayloadReq.SecretId, getPayloadReq.VersionId, err)
		}

		log.Printf("[INFO] read Lockbox payload (to compare entries) with VersionID: %s", payload.GetVersionId())

		currentEntries = payload.Entries
	}

	req := &lockbox.AddVersionRequest{
		SecretId: d.Get("secret_id").(string),
		// Make sure we're taking this version as reference, since addEntryChangesForRemovedKeys will
		// remove the entries in payload.Entries that versionPayloadEntries doesn't contain anymore.
		Description: d.Get("description").(string),
	}
	if currentVersionId != "" {
		req.BaseVersionId = currentVersionId
	}
	if secret.PayloadSpecification == nil {
		if currentEntries != nil {
			req.PayloadEntries = addEntryChangesForRemovedKeys(currentEntries, versionPayloadEntries)
		} else {
			req.PayloadEntries = versionPayloadEntries
		}
	}

	log.Printf("[INFO] adding Lockbox version for secret with ID: %s, base version ID: %s", req.SecretId, req.BaseVersionId)

	op, err := config.sdk.WrapOperation(config.sdk.LockboxSecret().Secret().AddVersion(ctx, req))
	if err != nil {
		return diag.Errorf("error while requesting API to add version: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("error while getting operation metadata of add secret version: %s", err)
	}

	md, ok := protoMetadata.(*lockbox.AddVersionMetadata)
	if !ok {
		return diag.Errorf("could not get Secret ID from create operation metadata")
	}

	d.SetId(md.VersionId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("error while waiting operation to add secret version: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("add secret version failed: %s", err)
	}

	log.Printf("[INFO] added Lockbox version with ID: %s", d.Id())

	return nil
}

func resourceYandexLockboxSecretVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	id := d.Id()
	req := &lockbox.GetPayloadRequest{
		SecretId:  d.Get("secret_id").(string),
		VersionId: id,
	}

	log.Printf("[INFO] reading Lockbox version: %s", protojson.Format(req))

	_, err := config.sdk.LockboxPayload().Payload().Get(ctx, req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("secret version payload %q", id)))
	}

	log.Printf("[INFO] read Lockbox version with ID: %s", id)

	return nil
}

func resourceYandexLockboxSecretVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceYandexLockboxSecretVersionMutex.Lock()
	defer resourceYandexLockboxSecretVersionMutex.Unlock()

	config := meta.(*Config)

	req := &lockbox.ScheduleVersionDestructionRequest{
		SecretId:  d.Get("secret_id").(string),
		VersionId: d.Id(),
	}

	log.Printf("[INFO] scheduling destruction of Lockbox version: %s", protojson.Format(req))

	op, err := config.sdk.WrapOperation(config.sdk.LockboxSecret().Secret().ScheduleVersionDestruction(ctx, req))
	if err != nil {
		return diag.Errorf("error while requesting API to destroy version: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("error while waiting operation to destroy version: %s", err)
	}

	log.Printf("[INFO] successfully scheduled destruction of Lockbox version with ID: %s", d.Id())

	return nil
}

// Removed keys will be added as lockbox.PayloadEntryChange with an empty value, to remove the key.
// We're simulating the behavior of PayloadChangeKind.FULL, but this option is not available in the public API.
func addEntryChangesForRemovedKeys(currentEntries []*lockbox.Payload_Entry, entryChanges []*lockbox.PayloadEntryChange) []*lockbox.PayloadEntryChange {
	for _, entry := range currentEntries {
		keyFound := false
		for _, change := range entryChanges {
			if change.Key == entry.Key {
				keyFound = true
				break
			}
		}
		if !keyFound { // key was removed
			entryChanges = append(entryChanges, &lockbox.PayloadEntryChange{
				Key:   entry.Key,
				Value: &lockbox.PayloadEntryChange_TextValue{TextValue: ""}, // empty value to remove the key
			})
		}
	}
	return entryChanges
}
