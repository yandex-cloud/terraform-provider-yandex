package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Resources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_compute_disk":              ResourceYandexComputeDisk(),
		"yandex_compute_image":             ResourceYandexComputeImage(),
		"yandex_compute_instance":          ResourceYandexComputeInstance(),
		"yandex_compute_instance_group":    ResourceYandexComputeInstanceGroup(),
		"yandex_compute_placement_group":   ResourceYandexComputePlacementGroup(),
		"yandex_compute_snapshot":          ResourceYandexComputeSnapshot(),
		"yandex_compute_snapshot_schedule": ResourceYandexComputeSnapshotSchedule(),
	}
}

func DataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"yandex_compute_disk":                 DataSourceYandexComputeDisk(),
		"yandex_compute_disk_placement_group": DataSourceYandexComputeDiskPlacementGroup(),
		"yandex_compute_filesystem":           DataSourceYandexComputeFilesystem(),
		"yandex_compute_gpu_cluster":          DataSourceYandexComputeGpuCluster(),
		"yandex_compute_image":                DataSourceYandexComputeImage(),
		"yandex_compute_instance":             DataSourceYandexComputeInstance(),
		"yandex_compute_instance_group":       DataSourceYandexComputeInstanceGroup(),
		"yandex_compute_placement_group":      DataSourceYandexComputePlacementGroup(),
		"yandex_compute_snapshot":             DataSourceYandexComputeSnapshot(),
		"yandex_compute_snapshot_schedule":    DataSourceYandexComputeSnapshotSchedule(),
	}
}
