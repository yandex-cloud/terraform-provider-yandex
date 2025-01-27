resource "yandex_serverless_container" "test-container-with-digest" {
  name   = "some_name"
  memory = 128
  image {
    url    = "cr.yandex/yc/test-image:v1"
    digest = "sha256:e1d772fa8795adac847a2420c87d0d2e3d38fb02f168cab8c0b5fe2fb95c47f4"
  }
}
