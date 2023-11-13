package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	yandexCMCertificateDefaultTimeout = 1 * time.Minute
)

func resourceYandexCMCertificate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexCMCertificateCreate,
		ReadContext:   resourceYandexCMCertificateRead,
		UpdateContext: resourceYandexCMCertificateUpdate,
		DeleteContext: resourceYandexCMCertificateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: yandexCMCertificateImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexCMCertificateDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexCMCertificateDefaultTimeout),
			Update: schema.DefaultTimeout(yandexCMCertificateDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexCMCertificateDefaultTimeout),
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domains": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      1,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ForceNew:      true,
				ConflictsWith: []string{"self_managed"},
			},
			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([-_0-9a-z]*)$"), ""), validation.StringLenBetween(0, 63)),
				},
				Set:      schema.HashString,
				Optional: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"self_managed": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"managed", "domains"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate": {
							Type:     schema.TypeString,
							Required: true,
						},
						"private_key": {
							Type:          schema.TypeString,
							Optional:      true,
							Sensitive:     true,
							ConflictsWith: []string{"self_managed.private_key_lockbox_secret"},
						},
						"private_key_lockbox_secret": {
							Type:          schema.TypeList,
							MaxItems:      1,
							Optional:      true,
							ConflictsWith: []string{"self_managed.private_key"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"key": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"managed": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"self_managed"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"challenge_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"DNS_CNAME", "DNS_TXT", "HTTP"}, false),
						},
						"challenge_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},

			// Exported attributes

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"issuer": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subject": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"serial": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"issued_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_after": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_before": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"challenges": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dns_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dns_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dns_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_content": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func getSelfManagedCertificateAndChain(d *schema.ResourceData, required bool) (string, string, diag.Diagnostics) {
	certificate, ok := d.GetOk("self_managed.0.certificate")
	if !ok {
		if required {
			return "", "", diag.Errorf("self_managed.certificate should be specified")
		}
		certificate = ""
	}
	return "", certificate.(string), nil
}

func getSelfManagedPrivateKey(ctx context.Context, d *schema.ResourceData, meta interface{}, required bool) (string, diag.Diagnostics) {
	config := meta.(*Config)
	privateKey, privateKeyOk := d.GetOk("self_managed.0.private_key")
	_, privateKeyLockboxOk := d.GetOk("self_managed.0.private_key_lockbox_secret")
	if !privateKeyOk && !privateKeyLockboxOk {
		if required {
			return "", diag.Errorf("either self_managed.private_key or self_managed.private_key_lockbox_secret should be specified")
		} else {
			return "", nil
		}
	}
	if privateKeyLockboxOk {
		lockboxId, ok := d.GetOk("self_managed.0.private_key_lockbox_secret.0.id")
		if !ok {
			return "", diag.Errorf("self_managed.private_key_lockbox_secret.id should be specified")
		}
		lockboxKey, ok := d.GetOk("self_managed.0.private_key_lockbox_secret.0.key")
		if !ok {
			return "", diag.Errorf("self_managed.private_key_lockbox_secret.key should be specified")
		}
		payload, err := config.sdk.LockboxPayload().Payload().Get(
			ctx,
			&lockbox.GetPayloadRequest{
				SecretId: lockboxId.(string),
			},
		)
		if err != nil {
			return "", diag.Errorf("error while requesting API to get secret: %s", err)
		}
		privateKey = nil
		for _, entry := range payload.Entries {
			if entry.Key == lockboxKey.(string) {
				privateKey = entry.Value.(*lockbox.Payload_Entry_TextValue).TextValue
			}
		}
		if privateKey == nil {
			return "", diag.Errorf("there is no secret key: %s", lockboxKey.(string))
		}
	}
	return privateKey.(string), nil
}

func resourceYandexCMCertificateCreateSelfManaged(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error while get labels: %s", err)
	}

	certificate, chain, errDiag := getSelfManagedCertificateAndChain(d, true)
	if errDiag != nil {
		return errDiag
	}
	privateKey, errDiag := getSelfManagedPrivateKey(ctx, d, meta, true)
	if errDiag != nil {
		return errDiag
	}

	op, err := config.sdk.WrapOperation(config.sdk.Certificates().Certificate().Create(
		ctx,
		&certificatemanager.CreateCertificateRequest{
			FolderId:           folderID,
			Name:               d.Get("name").(string),
			Description:        d.Get("description").(string),
			Labels:             labels,
			Certificate:        certificate,
			Chain:              chain,
			PrivateKey:         privateKey,
			DeletionProtection: d.Get("deletion_protection").(bool),
		},
	))
	if err != nil {
		return diag.Errorf("error while requesting API to create certificate: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("error while getting operation metadata of create certificate: %s", err)
	}

	md, ok := protoMetadata.(*certificatemanager.CreateCertificateMetadata)
	if !ok {
		return diag.Errorf("could not get Certificate Id from create operation metadata")
	}

	d.SetId(md.CertificateId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("error while waiting operation to create certificate: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("certificate creation failed: %s", err)
	}

	log.Printf("[INFO] created Certificate with ID: %s", d.Id())

	return resourceYandexCMCertificateRead(ctx, d, meta)
}

type challengeType int

const (
	CHALLENGE_TYPE_DNS_CNAME challengeType = 0
	CHALLENGE_TYPE_DNS_TXT   challengeType = 1
	CHALLENGE_TYPE_HTTP      challengeType = 2
)

func parseChallengeType(challengeTypeStr string) (challengeType, error) {
	switch challengeTypeStr {
	case "DNS_CNAME":
		return CHALLENGE_TYPE_DNS_CNAME, nil
	case "DNS_TXT":
		return CHALLENGE_TYPE_DNS_TXT, nil
	case "HTTP":
		return CHALLENGE_TYPE_HTTP, nil
	}
	return 0, fmt.Errorf("unknown challenge type: %s", challengeTypeStr)
}

func challengeTypeToCMChallengeType(challengeType challengeType) certificatemanager.ChallengeType {
	switch challengeType {
	case CHALLENGE_TYPE_DNS_CNAME:
		return certificatemanager.ChallengeType_DNS
	case CHALLENGE_TYPE_DNS_TXT:
		return certificatemanager.ChallengeType_DNS
	case CHALLENGE_TYPE_HTTP:
		return certificatemanager.ChallengeType_HTTP
	}
	return certificatemanager.ChallengeType_CHALLENGE_TYPE_UNSPECIFIED
}

func resourceYandexCMCertificateCreateManagedByLetsEncrypt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("error while get labels: %s", err)
	}

	challengeTypeStr, ok := d.GetOk("managed.0.challenge_type")
	if !ok {
		return diag.Errorf("managed.challenge_type should be specified")
	}
	challengeType, err := parseChallengeType(challengeTypeStr.(string))
	if err != nil {
		return diag.FromErr(err)
	}

	domainsIntf, ok := d.GetOk("domains")
	if !ok {
		return diag.Errorf("domains should be specified")
	}

	var domains []string
	for _, v := range domainsIntf.([]interface{}) {
		domains = append(domains, v.(string))
	}

	op, err := config.sdk.WrapOperation(config.sdk.Certificates().Certificate().RequestNew(
		ctx,
		&certificatemanager.RequestNewCertificateRequest{
			FolderId:           folderID,
			Name:               d.Get("name").(string),
			Description:        d.Get("description").(string),
			Labels:             labels,
			DeletionProtection: d.Get("deletion_protection").(bool),
			Domains:            domains,
			ChallengeType:      challengeTypeToCMChallengeType(challengeType),
		},
	))
	if err != nil {
		return diag.Errorf("error while requesting API to request certificate: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("error while getting operation metadata of request certificate: %s", err)
	}

	md, ok := protoMetadata.(*certificatemanager.RequestNewCertificateMetadata)
	if !ok {
		return diag.Errorf("could not get Certificate Id from request operation metadata")
	}

	d.SetId(md.CertificateId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("error while waiting operation to request certificate: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("certificate request failed: %s", err)
	}
	log.Printf("[INFO] requested Certificate with ID: %s", d.Id())
	d.Partial(true)
	result := yandexCMCertificateRead(d.Id(), ctx, d, meta, false)
	d.Partial(false)
	return result
}

func resourceYandexCMCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, selfManagedOk := d.GetOk("self_managed")
	_, managedOk := d.GetOk("managed")

	if selfManagedOk {
		return resourceYandexCMCertificateCreateSelfManaged(ctx, d, meta)
	}
	if managedOk {
		return resourceYandexCMCertificateCreateManagedByLetsEncrypt(ctx, d, meta)
	}
	return diag.Errorf("either self_managed or managed should be specified")
}

func resourceYandexCMCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return yandexCMCertificateRead(d.Id(), ctx, d, meta, false)
}

func yandexCMCertificateRead(id string, ctx context.Context, d *schema.ResourceData, meta interface{}, fromDataSource bool) diag.Diagnostics {
	config := meta.(*Config)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		req := &certificatemanager.GetCertificateRequest{
			CertificateId: id,
			View:          certificatemanager.CertificateView_FULL,
		}
		log.Printf("[INFO] reading Certificate: %s", protojson.Format(req))

		resp, err := config.sdk.Certificates().Certificate().Get(ctx, req)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		if resp.Status == certificatemanager.Certificate_VALIDATING ||
			resp.Status == certificatemanager.Certificate_RENEWING {
			if fromDataSource {
				if d.Get("wait_validation").(bool) {
					return resource.RetryableError(
						fmt.Errorf("certificate still %s", certificatemanager.Certificate_Status_name[int32(resp.Status)]),
					)
				}
			} else if _, ok := d.GetOk("managed"); ok {
				if resp.Challenges == nil || len(resp.Challenges) == 0 {
					return resource.RetryableError(
						fmt.Errorf("certificate challenges still being created"),
					)
				} else {
					for _, challenge := range resp.Challenges {
						if challenge.Challenge == nil {
							return resource.RetryableError(
								fmt.Errorf("certificate challenges still being created"),
							)
						}
					}
				}
			}
		}

		if err := d.Set("folder_id", resp.FolderId); err != nil {
			log.Printf("[ERROR] failed set field folder_id: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("created_at", getTimestamp(resp.CreatedAt)); err != nil {
			log.Printf("[ERROR] failed set field created_at: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("name", resp.Name); err != nil {
			log.Printf("[ERROR] failed set field name: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("description", resp.Description); err != nil {
			log.Printf("[ERROR] failed set field description: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("labels", resp.Labels); err != nil {
			log.Printf("[ERROR] failed set field labels: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("type", resp.Type.String()); err != nil {
			log.Printf("[ERROR] failed set field type: %s", err)
			return resource.NonRetryableError(err)
		}
		if fromDataSource {
			// In the resource, this value might differ from the original `domains`, so we don't set it
			// We could decide to set the output domains to another attribute.
			// Or use DiffSuppressFunc, to change the way domains are compared.
			if err := d.Set("domains", resp.Domains); err != nil {
				log.Printf("[ERROR] failed set field domains: %s", err)
				return resource.NonRetryableError(err)
			}
		}
		if err := d.Set("status", resp.Status.String()); err != nil {
			log.Printf("[ERROR] failed set field status: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("issuer", resp.Issuer); err != nil {
			log.Printf("[ERROR] failed set field issuer: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("subject", resp.Subject); err != nil {
			log.Printf("[ERROR] failed set field subject: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("serial", resp.Serial); err != nil {
			log.Printf("[ERROR] failed set field serial: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("updated_at", getTimestamp(resp.UpdatedAt)); err != nil {
			log.Printf("[ERROR] failed set field updated_at: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("issued_at", getTimestamp(resp.IssuedAt)); err != nil {
			log.Printf("[ERROR] failed set field issued_at: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("not_after", getTimestamp(resp.NotAfter)); err != nil {
			log.Printf("[ERROR] failed set field not_after: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("not_before", getTimestamp(resp.NotBefore)); err != nil {
			log.Printf("[ERROR] failed set field not_before: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("deletion_protection", resp.DeletionProtection); err != nil {
			log.Printf("[ERROR] failed set field deletion_protection: %s", err)
			return resource.NonRetryableError(err)
		}
		var challengeType = CHALLENGE_TYPE_DNS_CNAME
		challengeTypeStr, ok := d.GetOk("managed.0.challenge_type")
		if ok {
			challengeType, err = parseChallengeType(challengeTypeStr.(string))
			if err != nil {
				log.Printf("[ERROR] failed parse field managed.challenge_type: %s", err)
				return resource.NonRetryableError(err)
			}
		}
		switch resp.Type {
		case certificatemanager.CertificateType_MANAGED:
			needReadChallenges := resp.Status == certificatemanager.Certificate_VALIDATING || resp.Status == certificatemanager.Certificate_RENEWING

			if needReadChallenges && len(resp.Challenges) == 0 {
				log.Printf("[WARN] the service did not provide challenges, but should have")
			}

			if needReadChallenges || len(resp.Challenges) > 0 {
				var challenges []interface{}
				var exists = make(map[string]bool)
				var key string
				for _, challenge := range resp.Challenges {
					var flChallenge map[string]interface{}
					switch challenge.Type {
					case certificatemanager.ChallengeType_DNS:
						dnsChallenge := challenge.Challenge.(*certificatemanager.Challenge_DnsChallenge).DnsChallenge
						if challengeType == CHALLENGE_TYPE_DNS_CNAME && strings.ToUpper(dnsChallenge.Type) == "CNAME" ||
							challengeType == CHALLENGE_TYPE_DNS_TXT && strings.ToUpper(dnsChallenge.Type) == "TXT" {
							flChallenge = map[string]interface{}{
								"dns_name":  dnsChallenge.Name,
								"dns_type":  dnsChallenge.Type,
								"dns_value": dnsChallenge.Value,
							}
							key = dnsChallenge.Name + " " + dnsChallenge.Type + " " + dnsChallenge.Value
						} else {
							continue
						}
					case certificatemanager.ChallengeType_HTTP:
						if challengeType == CHALLENGE_TYPE_HTTP {
							httpChallenge := challenge.Challenge.(*certificatemanager.Challenge_HttpChallenge).HttpChallenge
							flChallenge = map[string]interface{}{
								"http_url":     httpChallenge.Url,
								"http_content": httpChallenge.Content,
							}
							key = httpChallenge.Url + " " + httpChallenge.Content
						} else {
							continue
						}
					default:
						continue
					}
					if exists[key] {
						continue
					}
					flChallenge["created_at"] = getTimestamp(challenge.CreatedAt)
					flChallenge["domain"] = challenge.Domain
					flChallenge["message"] = challenge.Message
					flChallenge["type"] = certificatemanager.ChallengeType_name[int32(challenge.Type)]
					flChallenge["updated_at"] = getTimestamp(challenge.UpdatedAt)
					exists[key] = true
					challenges = append(challenges, flChallenge)
				}
				if err := d.Set("challenges", challenges); err != nil {
					log.Printf("[ERROR] failed set field challenges: %s", err)
					return resource.NonRetryableError(err)
				}
				_, isManaged := d.GetOk("managed")
				if !fromDataSource && isManaged {
					if challengeCount, ok := d.GetOk("managed.0.challenge_count"); ok {
						if len(challenges) != challengeCount {
							log.Printf("[ERROR] managed.challenge_count must be equals %d", len(challenges))
							return resource.NonRetryableError(fmt.Errorf("managed.challenge_count must be equals %d", len(challenges)))
						}
					}
				}
			} else {
				log.Printf("[INFO] the challenges update will be skipped, because service did not transmit them")
			}
		case certificatemanager.CertificateType_IMPORTED:
			var challenges []interface{}
			if err := d.Set("challenges", challenges); err != nil {
				log.Printf("[ERROR] failed set field challenges: %s", err)
				return resource.NonRetryableError(err)
			}
		}

		d.SetId(resp.Id)

		log.Printf("[INFO] read Certificate with ID: %s", d.Id())
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceYandexCMCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &certificatemanager.UpdateCertificateRequest{
		CertificateId: d.Id(),
		UpdateMask:    &field_mask.FieldMask{},
	}

	d.Partial(true)

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	if d.HasChange("labels") {
		labelsProp, err := expandLabels(d.Get("labels"))
		if err != nil {
			return diag.Errorf("error while get labels: %s", err)
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	if d.HasChange("deletion_protection") {
		req.DeletionProtection = d.Get("deletion_protection").(bool)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "deletion_protection")
	}

	if d.HasChange("self_managed.0.certificate") ||
		d.HasChange("self_managed.0.private_key") ||
		d.HasChange("self_managed.0.private_key_lockbox_secret.0.id") ||
		d.HasChange("self_managed.0.private_key_lockbox_secret.0.key") {
		certificate, chain, errDiag := getSelfManagedCertificateAndChain(d, false)
		if errDiag != nil {
			return errDiag
		}
		req.Certificate = certificate
		req.Chain = chain

		privateKey, errDiag := getSelfManagedPrivateKey(ctx, d, meta, false)
		if errDiag != nil {
			return errDiag
		}
		req.PrivateKey = privateKey
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "certificate", "chain", "private_key")
	}

	if len(req.UpdateMask.Paths) > 0 {
		op, err := config.sdk.WrapOperation(config.sdk.Certificates().Certificate().Update(ctx, req))
		if err != nil {
			return diag.Errorf("error while requesting API to update certificate: %s", err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return diag.Errorf("error while waiting operation to update certificate: %s", err)

		}
		if _, err := op.Response(); err != nil {
			return diag.Errorf("certificate update failed: %s", err)
		}
	}
	d.Partial(false)
	log.Printf("[INFO] updated certificate with ID: %s", d.Id())
	return resourceYandexCMCertificateRead(ctx, d, meta)
}

func resourceYandexCMCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &certificatemanager.DeleteCertificateRequest{
		CertificateId: d.Id(),
	}

	log.Printf("[INFO] deleting certificate: %s", protojson.Format(req))

	op, err := config.sdk.WrapOperation(config.sdk.Certificates().Certificate().Delete(ctx, req))
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("certificate %q", d.Id())))
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] deleted certificate with ID: %s", d.Id())
	return nil
}

func yandexCMCertificateImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	config := m.(*Config)

	req := &certificatemanager.GetCertificateRequest{
		CertificateId: d.Id(),
		View:          certificatemanager.CertificateView_FULL,
	}
	resp, err := config.sdk.Certificates().Certificate().Get(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Type == certificatemanager.CertificateType_MANAGED {
		if err := d.Set("domains", resp.Domains); err != nil {
			log.Printf("[ERROR] failed set field domains: %s", err)
			return nil, err
		}
		managed := make(map[string]interface{})
		managed["challenge_count"] = 0
		if err := d.Set("managed", []interface{}{managed}); err != nil {
			log.Printf("[ERROR] failed set field managed: %s", err)
			return nil, err
		}

	}
	return []*schema.ResourceData{d}, nil
}
