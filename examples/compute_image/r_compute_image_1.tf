//
// Create a new Compute Image.
//
resource "yandex_compute_image" "foo-image" {
  name       = "my-custom-image"
  source_url = "https://storage.yandexcloud.net/lucky-images/kube-it.img"
}

// You can use "data.yandex_compute_image.my_image.id" identifier 
// as reference to existing resource.
resource "yandex_compute_instance" "vm" {
  name = "vm-from-custom-image"

  # ...

  boot_disk {
    initialize_params {
      image_id = yandex_compute_image.foo-image.id
    }
  }
}
