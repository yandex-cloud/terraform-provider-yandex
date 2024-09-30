resource "yandex_cdn_origin_group" "my_group" {

  name = "My Origin group"

  use_next = true

  origin {
    source = "ya.ru"
  }

  origin {
    source = "yandex.ru"
  }

  origin {
    source = "goo.gl"
  }

  origin {
    source = "amazon.com"
    backup = false
  }
}
