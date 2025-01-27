resource "yandex_compute_image" "image1" {
  name       = "my-custom-image"
  source_url = "https://storage.yandexcloud.net/lucky-images/kube-it.img"
}

resource "yandex_compute_image_iam_binding" "editor" {
  image_id = data.yandex_compute_image.image1.id

  role = "editor"

  members = [
    "userAccount:some_user_id",
  ]
}
