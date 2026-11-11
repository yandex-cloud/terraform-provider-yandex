package sweep

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func Register(add func(string, *resource.Sweeper)) {
	registerResourceYandexComputeDiskPlacementGroup(add)
	registerResourceYandexComputeDisk(add)
	registerResourceYandexComputeFilesystem(add)
	registerResourceYandexComputeGpuCluster(add)
	registerResourceYandexComputeInstanceGroup(add)
	registerResourceYandexComputeInstance(add)
	registerResourceYandexComputePlacementGroup(add)
}
