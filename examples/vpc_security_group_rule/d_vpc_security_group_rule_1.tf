//
// Get information about existing VPC Security Group Rule.
//
data "yandex_vpc_security_group_rule" "rule1" {
  security_group_binding = "my-sg-id"
  rule_id                = "my-rule-id"
}
