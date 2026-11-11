package sweep

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func Register(add func(string, *resource.Sweeper)) {
	registerResourceYandexMdbGreenplumCluster(add)
	registerMdbGreenplumClusterV2Resource(add)
}
