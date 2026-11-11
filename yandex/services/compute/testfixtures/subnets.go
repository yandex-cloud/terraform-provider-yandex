package testfixtures

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func SubnetIPMap(config *providerconfig.Config, instanceNames []string) (map[string][]string, error) {
	result := make(map[string][]string)
	ctx := context.Background()

	for _, instanceName := range instanceNames {
		instanceID, err := sdkcommon.ResolveObjectIDByNameAndFolderIDV2(ctx, instanceName, config.FolderID, sdkcommon.ResolverWithClient(computesdk.NewInstanceClient(config.SDK), computesdk.InstanceResolver))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve data source instance by name: %v", err)
		}
		instance, err := computesdk.NewInstanceClient(config.SDK).Get(ctx, &compute.GetInstanceRequest{
			InstanceId: instanceID,
			View:       compute.InstanceView_FULL,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get instance with ID: %v", instanceID)
		}

		ifs := instance.GetNetworkInterfaces()
		if len(ifs) == 0 {
			return nil, fmt.Errorf("target instance %s doesn't have network interface", instanceName)
		}
		subnetID := ifs[0].GetSubnetId()
		address := ifs[0].GetPrimaryV4Address().GetAddress()
		result[subnetID] = append(result[subnetID], address)
	}

	return result, nil
}
