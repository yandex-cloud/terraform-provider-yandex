//
// Create a new Compute Instance and new IAM Binding for it.
//
resource "yandex_compute_instance" "vm1" {
  name        = "test"
  platform_id = "standard-v3"
  zone        = "ru-central1-a"

  resources {
    cores  = 2
    memory = 4
  }

  boot_disk {
    disk_id = yandex_compute_disk.boot-disk.id
  }

  network_interface {
    index     = 1
    subnet_id = yandex_vpc_subnet.foo.id
  }

  metadata = {
    foo      = "bar"
    ssh-keys = "ubuntu:${file("~/.ssh/id_ed25519.pub")}"
  }
}

resource "yandex_compute_instance_iam_binding" "editor" {
  image_id = data.yandex_compute_instance.vm1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
