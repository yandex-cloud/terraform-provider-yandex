package pluginsdk

import (
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

type instanceAction int

const (
	instanceActionStop instanceAction = iota
	instanceActionStart
	instanceActionRestart
)

func (action instanceAction) String() string {
	switch action {
	case instanceActionStop:
		return "Stop"
	case instanceActionStart:
		return "Start"
	case instanceActionRestart:
		return "Restart"
	default:
		return "Unknown"
	}
}

func DiskSpecChanged(currDisk *compute.AttachedDisk, newSpec *compute.AttachedDiskSpec) bool {
	if currDisk == nil || newSpec == nil {
		return false
	}
	return !((int32(currDisk.Mode) == int32(newSpec.Mode) ||
		currDisk.Mode == compute.AttachedDisk_READ_WRITE && newSpec.Mode == compute.AttachedDiskSpec_MODE_UNSPECIFIED) &&
		(currDisk.DeviceName == newSpec.DeviceName ||
			currDisk.DeviceName == currDisk.DiskId && newSpec.DeviceName == "") &&
		currDisk.AutoDelete == newSpec.AutoDelete &&
		currDisk.DiskId == newSpec.GetDiskId())
}

func getSnapshotMinStorageSize(snapshotID string, config *providerconfig.Config) (size int64, err error) {
	ctx := config.Context()

	snapshot, err := computesdk.NewSnapshotClient(config.SDK).Get(ctx, &compute.GetSnapshotRequest{
		SnapshotId: snapshotID,
	})

	if err != nil {
		return 0, fmt.Errorf("Error on retrieve snapshot properties: %s", err)
	}

	return snapshot.DiskSize, nil
}

func getImageMinStorageSize(imageID string, config *providerconfig.Config) (size int64, err error) {
	ctx := config.Context()

	image, err := computesdk.NewImageClient(config.SDK).Get(ctx, &compute.GetImageRequest{
		ImageId: imageID,
	})

	if err != nil {
		return 0, fmt.Errorf("Error on retrieve image properties: %s", err)
	}

	return image.MinDiskSize, nil
}
