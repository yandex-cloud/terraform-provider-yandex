package pluginsdk_test

import (
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	svc_compute "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/compute/pluginsdk"
)

func Test_diskSpecChanged(t *testing.T) {
	tests := []struct {
		name     string
		currDisk *compute.AttachedDisk
		newSpec  *compute.AttachedDiskSpec
		want     bool
	}{
		{
			name:     "empty",
			currDisk: &compute.AttachedDisk{},
			newSpec:  &compute.AttachedDiskSpec{},
			want:     false,
		},
		{
			name: "unchanged",
			currDisk: &compute.AttachedDisk{
				DeviceName: "device-name",
				DiskId:     "disk-id",
				Mode:       compute.AttachedDisk_READ_WRITE,
				AutoDelete: true,
			},
			newSpec: &compute.AttachedDiskSpec{
				DeviceName: "device-name",
				Mode:       compute.AttachedDiskSpec_READ_WRITE,
				AutoDelete: true,
				Disk: &compute.AttachedDiskSpec_DiskId{
					DiskId: "disk-id",
				},
			},
			want: false,
		},
		{
			name: "different device name",
			currDisk: &compute.AttachedDisk{
				DeviceName: "device-name1",
				DiskId:     "disk-id",
				Mode:       compute.AttachedDisk_READ_WRITE,
				AutoDelete: true,
			},
			newSpec: &compute.AttachedDiskSpec{
				DeviceName: "device-name2",
				Mode:       compute.AttachedDiskSpec_READ_WRITE,
				AutoDelete: true,
				Disk: &compute.AttachedDiskSpec_DiskId{
					DiskId: "disk-id",
				},
			},
			want: true,
		},
		{
			name: "empty new device name",
			currDisk: &compute.AttachedDisk{
				DeviceName: "device-name1",
				DiskId:     "disk-id",
				Mode:       compute.AttachedDisk_READ_WRITE,
				AutoDelete: true,
			},
			newSpec: &compute.AttachedDiskSpec{
				DeviceName: "",
				Mode:       compute.AttachedDiskSpec_READ_WRITE,
				AutoDelete: true,
				Disk: &compute.AttachedDiskSpec_DiskId{
					DiskId: "disk-id",
				},
			},
			want: true,
		},
		{
			name: "empty new device name not changed",
			currDisk: &compute.AttachedDisk{
				DeviceName: "disk-id",
				DiskId:     "disk-id",
				Mode:       compute.AttachedDisk_READ_WRITE,
				AutoDelete: true,
			},
			newSpec: &compute.AttachedDiskSpec{
				DeviceName: "",
				Mode:       compute.AttachedDiskSpec_READ_WRITE,
				AutoDelete: true,
				Disk: &compute.AttachedDiskSpec_DiskId{
					DiskId: "disk-id",
				},
			},
			want: false,
		},
		{
			name: "different mode",
			currDisk: &compute.AttachedDisk{
				DeviceName: "device-name",
				DiskId:     "disk-id",
				Mode:       compute.AttachedDisk_READ_WRITE,
				AutoDelete: true,
			},
			newSpec: &compute.AttachedDiskSpec{
				DeviceName: "device-name",
				Mode:       compute.AttachedDiskSpec_READ_ONLY,
				AutoDelete: true,
				Disk: &compute.AttachedDiskSpec_DiskId{
					DiskId: "disk-id",
				},
			},
			want: true,
		},
		{
			name: "empty mode unchanged",
			currDisk: &compute.AttachedDisk{
				DeviceName: "device-name",
				DiskId:     "disk-id",
				Mode:       compute.AttachedDisk_READ_WRITE,
				AutoDelete: true,
			},
			newSpec: &compute.AttachedDiskSpec{
				DeviceName: "device-name",
				Mode:       compute.AttachedDiskSpec_MODE_UNSPECIFIED,
				AutoDelete: true,
				Disk: &compute.AttachedDiskSpec_DiskId{
					DiskId: "disk-id",
				},
			},
			want: false,
		},
		{
			name: "empty mode changed",
			currDisk: &compute.AttachedDisk{
				DeviceName: "device-name",
				DiskId:     "disk-id",
				Mode:       compute.AttachedDisk_READ_ONLY,
				AutoDelete: true,
			},
			newSpec: &compute.AttachedDiskSpec{
				DeviceName: "device-name",
				Mode:       compute.AttachedDiskSpec_MODE_UNSPECIFIED,
				AutoDelete: true,
				Disk: &compute.AttachedDiskSpec_DiskId{
					DiskId: "disk-id",
				},
			},
			want: true,
		},
		{
			name: "different auto delete",
			currDisk: &compute.AttachedDisk{
				DeviceName: "device-name",
				DiskId:     "disk-id",
				Mode:       compute.AttachedDisk_READ_WRITE,
				AutoDelete: true,
			},
			newSpec: &compute.AttachedDiskSpec{
				DeviceName: "device-name",
				Mode:       compute.AttachedDiskSpec_READ_WRITE,
				AutoDelete: false,
				Disk: &compute.AttachedDiskSpec_DiskId{
					DiskId: "disk-id",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc_compute.DiskSpecChanged(tt.currDisk, tt.newSpec); got != tt.want {
				t.Errorf("diskSpecChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}
