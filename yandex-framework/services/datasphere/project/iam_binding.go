package project

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/accessbinding"
)

func NewIamBinding() resource.Resource {
	return accessbinding.NewIamBinding(newProjectIamUpdater())
}
