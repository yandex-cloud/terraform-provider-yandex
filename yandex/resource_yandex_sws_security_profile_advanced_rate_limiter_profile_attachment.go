package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	smartwebsecurity "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1"
	smartwebsecuritysdk "github.com/yandex-cloud/go-sdk/services/smartwebsecurity/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
)

const yandexSWSSecurityProfileARLAttachmentDefaultTimeout = 5 * time.Minute

func resourceYandexSWSSecurityProfileAdvancedRateLimiterProfileAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "Attaches an Advanced Rate Limiter profile to a Smart Web Security profile. Do not manage `advanced_rate_limiter_profile_id` in `yandex_sws_security_profile` when using this resource.",
		Create:      resourceYandexSWSSecurityProfileARLAttachmentCreate,
		Read:        resourceYandexSWSSecurityProfileARLAttachmentRead,
		Delete:      resourceYandexSWSSecurityProfileARLAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexSWSSecurityProfileARLAttachmentDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexSWSSecurityProfileARLAttachmentDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexSWSSecurityProfileARLAttachmentDefaultTimeout),
		},
		Schema: map[string]*schema.Schema{
			"security_profile_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Smart Web Security profile.",
			},
			"advanced_rate_limiter_profile_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Advanced Rate Limiter profile.",
			},
		},
	}
}

func resourceYandexSWSSecurityProfileARLAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	securityProfileID := d.Get("security_profile_id").(string)
	arlProfileID := d.Get("advanced_rate_limiter_profile_id").(string)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	profile, err := smartwebsecuritysdk.NewSecurityProfileClient(config.SDK).Get(ctx, &smartwebsecurity.GetSecurityProfileRequest{
		SecurityProfileId: securityProfileID,
	})
	if err != nil {
		return fmt.Errorf("reading Smart Web Security profile %q: %w", securityProfileID, err)
	}
	if current := profile.AdvancedRateLimiterProfileId; current != "" && current != arlProfileID {
		return fmt.Errorf("Smart Web Security profile %q is already attached to Advanced Rate Limiter profile %q", securityProfileID, current)
	}

	if profile.AdvancedRateLimiterProfileId != arlProfileID {
		if err := updateSWSSecurityProfileARLAttachment(ctx, config, securityProfileID, arlProfileID); err != nil {
			return err
		}
	}

	d.SetId(makeSWSSecurityProfileARLAttachmentID(securityProfileID, arlProfileID))
	return resourceYandexSWSSecurityProfileARLAttachmentRead(d, meta)
}

func resourceYandexSWSSecurityProfileARLAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	securityProfileID, arlProfileID, err := swsSecurityProfileARLAttachmentIDs(d)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	profile, err := smartwebsecuritysdk.NewSecurityProfileClient(config.SDK).Get(ctx, &smartwebsecurity.GetSecurityProfileRequest{
		SecurityProfileId: securityProfileID,
	})
	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading Smart Web Security profile %q: %w", securityProfileID, err)
	}
	if profile.AdvancedRateLimiterProfileId != arlProfileID {
		d.SetId("")
		return nil
	}

	if err := d.Set("security_profile_id", securityProfileID); err != nil {
		return err
	}
	return d.Set("advanced_rate_limiter_profile_id", arlProfileID)
}

func resourceYandexSWSSecurityProfileARLAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	securityProfileID, arlProfileID, err := swsSecurityProfileARLAttachmentIDs(d)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	profile, err := smartwebsecuritysdk.NewSecurityProfileClient(config.SDK).Get(ctx, &smartwebsecurity.GetSecurityProfileRequest{
		SecurityProfileId: securityProfileID,
	})
	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading Smart Web Security profile %q before detaching Advanced Rate Limiter profile: %w", securityProfileID, err)
	}
	if profile.AdvancedRateLimiterProfileId == arlProfileID {
		if err := updateSWSSecurityProfileARLAttachment(ctx, config, securityProfileID, ""); err != nil {
			return err
		}
	}

	d.SetId("")
	return nil
}

func updateSWSSecurityProfileARLAttachment(ctx context.Context, config *Config, securityProfileID, arlProfileID string) error {
	op, err := smartwebsecuritysdk.NewSecurityProfileClient(config.SDK).Update(ctx, &smartwebsecurity.UpdateSecurityProfileRequest{
		SecurityProfileId:            securityProfileID,
		AdvancedRateLimiterProfileId: arlProfileID,
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"advanced_rate_limiter_profile_id"},
		},
	})
	if err != nil {
		return fmt.Errorf("updating Advanced Rate Limiter attachment for Smart Web Security profile %q: %w", securityProfileID, err)
	}
	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("waiting for Advanced Rate Limiter attachment update for Smart Web Security profile %q: %w", securityProfileID, err)
	}
	return nil
}

func makeSWSSecurityProfileARLAttachmentID(securityProfileID, arlProfileID string) string {
	return securityProfileID + "/" + arlProfileID
}

func swsSecurityProfileARLAttachmentIDs(d *schema.ResourceData) (string, string, error) {
	securityProfileID, _ := d.Get("security_profile_id").(string)
	arlProfileID, _ := d.Get("advanced_rate_limiter_profile_id").(string)
	if securityProfileID != "" && arlProfileID != "" {
		return securityProfileID, arlProfileID, nil
	}

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid attachment ID %q: expected <security_profile_id>/<advanced_rate_limiter_profile_id>", d.Id())
	}
	return parts[0], parts[1], nil
}
