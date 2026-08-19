package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	smartwebsecurity "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1"
)

func TestAccSmartwebsecuritySecurityProfileARLAttachment_deleteBeforeProfiles(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-sws-arl-attachment")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccSmartwebsecuritySecurityProfileARLAttachment(name, true),
				Check:  testAccCheckSmartwebsecuritySecurityProfileARLAttachment(true),
			},
			{
				Config: testAccSmartwebsecuritySecurityProfileARLAttachment(name, false),
				Check:  testAccCheckSmartwebsecuritySecurityProfileARLAttachment(false),
			},
		},
	})
}

func TestAccSmartwebsecuritySecurityProfileARLAttachment_inlineCompatibility(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-sws-arl-inline")
	config := testAccSmartwebsecuritySecurityProfileARLInline(name)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  testAccCheckSmartwebsecuritySecurityProfileARLAttachment(true),
			},
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccSmartwebsecuritySecurityProfileARLAttachment_migrateFromInline(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-sws-arl-migrate")
	var securityProfileID, arlProfileID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccSmartwebsecuritySecurityProfileARLInline(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCaptureResourceIDs(&securityProfileID, &arlProfileID),
					testAccCheckSmartwebsecuritySecurityProfileARLAttachment(true),
				),
			},
			{
				Config: testAccSmartwebsecuritySecurityProfileARLAttachmentMigrated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDsUnchanged(&securityProfileID, &arlProfileID),
					testAccCheckSmartwebsecuritySecurityProfileARLAttachment(true),
				),
			},
			{
				Config:             testAccSmartwebsecuritySecurityProfileARLAttachmentMigrated(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCaptureResourceIDs(securityProfileID, arlProfileID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		securityProfile, ok := s.RootModule().Resources["yandex_sws_security_profile.this"]
		if !ok {
			return fmt.Errorf("security profile resource not found in state")
		}
		arlProfile, ok := s.RootModule().Resources["yandex_sws_advanced_rate_limiter_profile.this"]
		if !ok {
			return fmt.Errorf("Advanced Rate Limiter profile resource not found in state")
		}
		*securityProfileID = securityProfile.Primary.ID
		*arlProfileID = arlProfile.Primary.ID
		return nil
	}
}

func testAccCheckResourceIDsUnchanged(securityProfileID, arlProfileID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		securityProfile, ok := s.RootModule().Resources["yandex_sws_security_profile.this"]
		if !ok {
			return fmt.Errorf("security profile resource not found in state")
		}
		arlProfile, ok := s.RootModule().Resources["yandex_sws_advanced_rate_limiter_profile.this"]
		if !ok {
			return fmt.Errorf("Advanced Rate Limiter profile resource not found in state")
		}
		if securityProfile.Primary.ID != *securityProfileID {
			return fmt.Errorf("Security Profile was replaced during migration: got %q, want %q", securityProfile.Primary.ID, *securityProfileID)
		}
		if arlProfile.Primary.ID != *arlProfileID {
			return fmt.Errorf("Advanced Rate Limiter profile was replaced during migration: got %q, want %q", arlProfile.Primary.ID, *arlProfileID)
		}
		return nil
	}
}

func testAccCheckSmartwebsecuritySecurityProfileARLAttachment(attached bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		const securityProfileResource = "yandex_sws_security_profile.this"

		rs, ok := s.RootModule().Resources[securityProfileResource]
		if !ok {
			return fmt.Errorf("resource %q not found in state", securityProfileResource)
		}

		conf, err := configForSweepers()
		if err != nil {
			return fmt.Errorf("creating SWS client: %w", err)
		}
		profile, err := conf.sdk.SmartWebSecurity().SecurityProfile().Get(conf.Context(), &smartwebsecurity.GetSecurityProfileRequest{
			SecurityProfileId: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("reading SWS security profile %q: %w", rs.Primary.ID, err)
		}

		if attached && profile.AdvancedRateLimiterProfileId == "" {
			return fmt.Errorf("SWS security profile %q has no Advanced Rate Limiter attachment", rs.Primary.ID)
		}
		if !attached && profile.AdvancedRateLimiterProfileId != "" {
			return fmt.Errorf("SWS security profile %q is still attached to Advanced Rate Limiter profile %q", rs.Primary.ID, profile.AdvancedRateLimiterProfileId)
		}
		return nil
	}
}

func testAccSmartwebsecuritySecurityProfileARLAttachment(targetName string, enabled bool) string {
	return fmt.Sprintf(`
resource "yandex_sws_advanced_rate_limiter_profile" "this" {
  count = %[2]t ? 1 : 0
  name  = "%[1]s-arl"

  advanced_rate_limiter_rule {
    name        = "rule1"
    priority    = 10
    description = "Attachment lifecycle test rule"
    dry_run     = true

    static_quota {
      action = "DENY"
      limit  = 10000000
      period = 1
    }
  }
}

resource "yandex_sws_security_profile" "this" {
  name                     = "%[1]s"
  default_action           = "ALLOW"
  disallow_data_processing = false
}

resource "yandex_sws_security_profile_advanced_rate_limiter_profile_attachment" "this" {
  count                            = %[2]t ? 1 : 0
  security_profile_id              = yandex_sws_security_profile.this.id
  advanced_rate_limiter_profile_id = yandex_sws_advanced_rate_limiter_profile.this[0].id
}
`, targetName, enabled)
}

func testAccSmartwebsecuritySecurityProfileARLInline(targetName string) string {
	return testAccSmartwebsecurityARLProfile(targetName) + fmt.Sprintf(`
resource "yandex_sws_security_profile" "this" {
  name                             = "%[1]s"
  default_action                   = "ALLOW"
  disallow_data_processing         = false
  advanced_rate_limiter_profile_id = yandex_sws_advanced_rate_limiter_profile.this.id
}
`, targetName)
}

func testAccSmartwebsecuritySecurityProfileARLAttachmentMigrated(targetName string) string {
	return testAccSmartwebsecurityARLProfile(targetName) + fmt.Sprintf(`
resource "yandex_sws_security_profile" "this" {
  name                     = "%[1]s"
  default_action           = "ALLOW"
  disallow_data_processing = false
}

resource "yandex_sws_security_profile_advanced_rate_limiter_profile_attachment" "this" {
  security_profile_id              = yandex_sws_security_profile.this.id
  advanced_rate_limiter_profile_id = yandex_sws_advanced_rate_limiter_profile.this.id
}
`, targetName)
}

func testAccSmartwebsecurityARLProfile(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_sws_advanced_rate_limiter_profile" "this" {
  name = "%[1]s-arl"

  advanced_rate_limiter_rule {
    name        = "rule1"
    priority    = 10
    description = "Attachment compatibility test rule"
    dry_run     = true

    static_quota {
      action = "DENY"
      limit  = 10000000
      period = 1
    }
  }
}
`, targetName)
}
