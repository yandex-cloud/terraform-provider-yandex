package common

var ResourceDescriptions = map[string]string{
	"id":                  "The resource identifier.",
	"folder_id":           "The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.",
	"name":                "The resource name.",
	"description":         "The resource description.",
	"labels":              "A set of key/value label pairs which assigned to resource.",
	"created_at":          "The creation timestamp of the resource.",
	"cloud_id":            "The `Cloud ID` which resource belongs to. If it is not provided, the default provider `cloud-id` is used.",
	"zone":                "The [availability zone](https://cloud.yandex.com/docs/overview/concepts/geo-scope) where resource is located. If it is not provided, the default provider zone will be used.",
	"deletion_protection": "The `true` value means that resource is protected from accidental deletion.",
	"security_group_ids":  "The list of security groups applied to resource or their components.",
	"service_account_id":  "[Service account](https://yandex.cloud/docs/iam/concepts/users/service-accounts) which linked to the resource.",
	"subnet_ids":          "The list of VPC subnets identifiers which resource is attached.",
	"network_id":          "The `VPC Network ID` of subnets which resource attached to.",
}
