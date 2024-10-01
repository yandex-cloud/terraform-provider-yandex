data "yandex_organizationmanager_os_login_settings" "my_os_login_settings_settings" {
  organization_id = "some_organization_id"
}

output "my_organization_ssh_certificates_enabled" {
  value = "data.yandex_organizationmanager_os_login_settings.my_os_login_settings.ssh_certificate_settings.0.enabled"
}
