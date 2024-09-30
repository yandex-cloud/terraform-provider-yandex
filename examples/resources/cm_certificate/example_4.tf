resource "yandex_cm_certificate" "example" {
  name = "example"

  self_managed {
    certificate = "-----BEGIN CERTIFICATE----- ... -----END CERTIFICATE----- \n -----BEGIN CERTIFICATE----- ... -----END CERTIFICATE-----"
    private_key = "-----BEGIN RSA PRIVATE KEY----- ... -----END RSA PRIVATE KEY-----"
  }
}
