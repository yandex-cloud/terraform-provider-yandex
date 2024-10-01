resource "yandex_api_gateway" "test-api-gateway" {
  name        = "some_name"
  description = "any description"
  labels = {
    label       = "label"
    empty-label = ""
  }
  custom_domains {
    fqdn           = "test.example.com"
    certificate_id = "<certificate_id_from_cert_manager>"
  }
  connectivity {
    network_id = "<dynamic network id>"
  }
  variables = {
    installation = "prod"
  }
  canary {
    weight = 20
    variables = {
      installation = "dev"
    }
  }
  log_options {
    log_group_id = "<log group id>"
    min_level    = "ERROR"
  }
  execution_timeout = "300"
  spec              = <<-EOT
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test API
x-yc-apigateway:
  variables:
    installation:
      default: "prod"
      enum:
       - "prod"
       - "dev"
paths:
  /hello:
    get:
      summary: Say hello
      operationId: hello
      parameters:
        - name: user
          in: query
          description: User name to appear in greetings
          required: false
          schema:
            type: string
            default: 'world'
      responses:
        '200':
          description: Greeting
          content:
            'text/plain':
              schema:
                type: "string"
      x-yc-apigateway-integration:
        type: dummy
        http_code: 200
        http_headers:
          'Content-Type': "text/plain"
        content:
          'text/plain': "Hello again, {user} from ${apigw.installation} release!\n"
EOT
}
