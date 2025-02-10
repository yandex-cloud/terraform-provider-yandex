//
// Create a new IoT Core Device.
//
resource "yandex_iot_core_device" "my_device" {
  registry_id = "are1sampleregistryid11"
  name        = "some_name"
  description = "any description"
  aliases = {
    "some_alias1/subtopic" = "$devices/{id}/events/somesubtopic",
    "some_alias2/subtopic" = "$devices/{id}/events/aaa/bbb",
  }
  passwords = [
    "my-password1",
    "my-password2"
  ]
  certificates = [
    "public part of certificate1",
    "public part of certificate2"
  ]
}
