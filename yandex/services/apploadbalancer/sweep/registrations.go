package sweep

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func Register(add func(string, *resource.Sweeper)) {
	registerResourceYandexAlbBackendGroup(add)
	registerResourceYandexAlbHttpRouter(add)
	registerResourceYandexAlbLoadBalancer(add)
	registerResourceYandexAlbTargetGroup(add)
}
