package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"google.golang.org/protobuf/encoding/protojson"
)

func dataSourceYandexCMCertificateContent() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexCMCertificateContentRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"certificate_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"wait_validation": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"private_key_format": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"certificates": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceYandexCMCertificateContentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "certificate_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}
	id := d.Get("certificate_id").(string)

	_, certificateNameOk := d.GetOk("name")
	if certificateNameOk {
		id, err = resolveObjectID(config.Context(), config, d, sdkresolvers.CertificateResolver)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	privateKeyFormat := certificatemanager.PrivateKeyFormat_PRIVATE_KEY_FORMAT_UNSPECIFIED
	formatString := d.Get("private_key_format")
	if formatString != nil {
		if format, ok := certificatemanager.PrivateKeyFormat_value[formatString.(string)]; ok {
			privateKeyFormat = certificatemanager.PrivateKeyFormat(format)
		}
	}

	// TODO: SA1019: resource.RetryContext is deprecated: Use helper/retry package instead. This is required for migrating acceptance testing to terraform-plugin-testing. (staticcheck)
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		if d.Get("wait_validation").(bool) {
			req := &certificatemanager.GetCertificateRequest{
				CertificateId: id,
				View:          certificatemanager.CertificateView_BASIC,
			}
			log.Printf("[INFO] reading Certificate status: %s", protojson.Format(req))

			resp, err := config.sdk.Certificates().Certificate().Get(ctx, req)
			if err != nil {
				// TODO: SA1019: resource.NonRetryableError is deprecated: Use helper/retry package instead. This is required for migrating acceptance testing to terraform-plugin-testing. (staticcheck)
				return resource.NonRetryableError(handleNotFoundError(err, d, fmt.Sprintf("certificate %q", id)))
			}
			if resp.Status == certificatemanager.Certificate_VALIDATING ||
				resp.Status == certificatemanager.Certificate_RENEWING {
				// TODO: SA1019: resource.RetryableError is deprecated: Use helper/retry package instead. This is required for migrating acceptance testing to terraform-plugin-testing. (staticcheck)
				return resource.RetryableError(
					fmt.Errorf("certificate still %s", certificatemanager.Certificate_Status_name[int32(resp.Status)]),
				)
			}
			log.Printf("[INFO] certificate status is %s", certificatemanager.Certificate_Status_name[int32(resp.Status)])
		}
		resp, err := config.sdk.CertificatesData().CertificateContent().Get(ctx, &certificatemanager.GetCertificateContentRequest{
			CertificateId:    id,
			PrivateKeyFormat: privateKeyFormat,
		})
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if err := d.Set("certificate_id", resp.CertificateId); err != nil {
			log.Printf("[ERROR] failed set field certificate_id: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("certificates", convertStringArrToInterface(resp.CertificateChain)); err != nil {
			log.Printf("[ERROR] failed set field certificates: %s", err)
			return resource.NonRetryableError(err)
		}
		if err := d.Set("private_key", resp.PrivateKey); err != nil {
			log.Printf("[ERROR] failed set field private_key: %s", err)
			return resource.NonRetryableError(err)
		}
		d.SetId(resp.CertificateId)
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
