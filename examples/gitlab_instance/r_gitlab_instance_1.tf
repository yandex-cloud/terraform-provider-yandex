//
// Create a new GitLab instance.
//
resource "yandex_gitlab_instance" "my_gitlab_instance" {
  name                      = "gitlab-created-with-terraform"
  resource_preset_id        = "s2.micro"
  disk_size                 = 30
  admin_login               = "gitlab-user"
  admin_email               = "gitlab-user@example.com"
  domain                    = "gitlab-terraform.gitlab.yandexcloud.net"
  subnet_id                 = yandex_vpc_subnet.a.id
  backup_retain_period_days = 7
}
