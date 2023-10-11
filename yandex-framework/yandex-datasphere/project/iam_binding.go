package yandex_datasphere_project

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	iam_binding "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/iam"
)

func NewIamBinding() resource.Resource {
	return iam_binding.NewIamBinding(newProjectIamUpdater())
}
