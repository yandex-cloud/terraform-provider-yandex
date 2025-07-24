//
// Create a new MDB Greenplum database resource group.
//
resource "yandex_mdb_greenplum_resource_group" "my_resource_group" {
  cluster_id     = yandex_mdb_greenplum_cluster.my_cluster.id
  name           = "alice"
  password       = "password"
  resource_group = "default_group"
}

resource "yandex_mdb_greenplum_cluster" "my_cluster" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
