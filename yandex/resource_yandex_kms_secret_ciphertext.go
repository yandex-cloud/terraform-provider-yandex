package yandex

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
)

const (
	yandexKMSSecretCiphertextDefaultTimeout = 1 * time.Minute
)

func resourceYandexKMSSecretCiphertext() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexKMSSecretCiphertextCreate,
		Read:   resourceYandexKMSSecretCiphertextRead,
		Delete: resourceYandexKMSSecretCiphertextDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexKMSSecretCiphertextDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexKMSSecretCiphertextDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexKMSSecretCiphertextDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"aad_context": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(0, 8192),
				ForceNew:     true,
				Optional:     true,
			},

			"plaintext": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(0, 32768),
				Required:     true,
				ForceNew:     true,
				Sensitive:    true,
			},

			"ciphertext": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexKMSSecretCiphertextCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	req := &kms.SymmetricEncryptRequest{
		KeyId:      d.Get("key_id").(string),
		Plaintext:  []byte(d.Get("plaintext").(string)),
		AadContext: []byte(d.Get("aad_context").(string)),
	}

	resp, err := config.sdk.KMSCrypto().SymmetricCrypto().Encrypt(ctx, req)
	if err != nil {
		return fmt.Errorf("Error while requesting API to encrypt data with KMS symmetric key: %s", err)
	}

	ciphertext := base64.StdEncoding.EncodeToString(resp.Ciphertext)
	d.Set("ciphertext", ciphertext)

	h := sha256.New()
	_, err = h.Write(resp.Ciphertext)
	if err != nil {
		return fmt.Errorf("Error while hashing ciphertext with sha256: %s", err)
	}
	hashedCiphertext := h.Sum(nil)

	id := fmt.Sprintf("%s/%x", resp.KeyId, hashedCiphertext)

	d.SetId(id)

	return resourceYandexKMSSecretCiphertextRead(d, meta)
}

func resourceYandexKMSSecretCiphertextRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	resp, err := config.sdk.KMS().SymmetricKey().Get(ctx, &kms.GetSymmetricKeyRequest{
		KeyId: d.Get("key_id").(string),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("KMS Symmetric Key %q", d.Id()))
	}

	if err != nil {
		return fmt.Errorf("Error while requesting API to get KMS symmetric key: %s", err)
	}

	if resp == nil {
		fmt.Printf("[DEBUG] Removing yandex_kms_secret_ciphertext because related key no longer exists.")
		d.SetId("")
		return nil
	}

	return nil
}

func resourceYandexKMSSecretCiphertextDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")

	return nil
}
