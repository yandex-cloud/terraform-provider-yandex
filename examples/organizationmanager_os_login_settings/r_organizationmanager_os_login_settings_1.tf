resource "yandex_organizationmanager_os_login_settings" "my_os_login_settings" {
  organization_id = "sdf4*********3fr"
  user_ssh_key_settings {
    enabled               = true
    allow_manage_own_keys = true
  }
  ssh_certificate_settings {
    enabled = true
  }
}
