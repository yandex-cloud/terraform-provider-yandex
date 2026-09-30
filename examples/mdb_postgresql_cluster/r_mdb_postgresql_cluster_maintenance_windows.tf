// A weekly maintenance schedule with separate write-interrupting and other slots.
resource "yandex_mdb_postgresql_cluster" "scheduled" {
  name        = "scheduled-postgresql"
  environment = "PRODUCTION"
  network_id  = yandex_vpc_network.maintenance.id

  config {
    version = 15
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.maintenance.id
  }

  maintenance_windows {
    type = "WEEKLY"

    slot {
      day                            = "MON"
      start_time                     = "02:30:00"
      duration                       = "3h"
      allow_temporary_unavailability = true
    }

    slot {
      day                            = "THU"
      start_time                     = "01:30:00"
      duration                       = "2h"
      allow_temporary_unavailability = false
    }
  }
}

resource "yandex_vpc_network" "maintenance" {}

resource "yandex_vpc_subnet" "maintenance" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.maintenance.id
  v4_cidr_blocks = ["10.10.0.0/24"]
}
