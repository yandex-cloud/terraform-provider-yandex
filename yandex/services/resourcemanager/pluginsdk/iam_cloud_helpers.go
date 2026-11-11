package pluginsdk

import (
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	resourcemanagersdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func cloudIDOfFolderID(config *providerconfig.Config, folderID string) (string, error) {
	client := resourcemanagersdk.NewFolderClient(config.SDK)

	folder, err := client.Get(config.Context(), &resourcemanager.GetFolderRequest{
		FolderId: folderID,
	})
	if err != nil {
		return "", err
	}
	return folder.CloudId, nil
}

func lockCloudByFolderID(config *providerconfig.Config, folderID string) (func(), error) {
	cloudID, err := cloudIDOfFolderID(config, folderID)
	if err != nil {
		return nil, fmt.Errorf("error getting cloud ID of `folder_id` %s: %s", folderID, err)
	}
	c := CloudIamUpdater{cloudID: cloudID}
	mutexKey := c.GetMutexKey()
	sdkcommon.MutexKV.Lock(mutexKey)
	return func() {
		sdkcommon.MutexKV.Unlock(mutexKey)
	}, nil
}
