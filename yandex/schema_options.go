package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

type SchemaOption func(timeouts *schema.Resource)

func WithTimeout(timeout *schema.ResourceTimeout) func(*schema.Resource) {
	return func(r *schema.Resource) {
		r.Timeouts = timeout
	}
}

func WithImporter(importer *schema.ResourceImporter) func(resource *schema.Resource) {
	return func(r *schema.Resource) {
		r.Importer = importer
	}
}
