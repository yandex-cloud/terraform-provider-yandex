package sweep

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	resourcemanagersdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func registerResourceYandexResourcemanagerFolder(add func(string, *resource.Sweeper)) {
	add("yandex_resourcemanager_folder", &resource.Sweeper{
		Name:         "yandex_resourcemanager_folder",
		F:            sweepAllFolders,
		Dependencies: []string{},
	})
}

func sweepAllFolders(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &resourcemanager.ListFoldersRequest{CloudId: conf.CloudID}
	client := resourcemanagersdk.NewFolderClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		if !strings.HasPrefix(it.Value().Name, folderPrefix) {
			continue
		}
		id := it.Value().GetId()
		if !sweep.SweepWithRetry(sweepFolderOnce, conf, "Folder", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Folder %q", id))
		}
	}
	if err := it.Error(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

const folderPrefix = "tfacc"

func sweepFolderOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(30 * time.Minute)
	defer cancel()

	client := resourcemanagersdk.NewFolderClient(conf.SDK)

	op, err := client.Delete(ctx, &resourcemanager.DeleteFolderRequest{
		FolderId:    id,
		DeleteAfter: timestamppb.Now(),
	})
	if err != nil {
		if sdkcommon.IsStatusWithCode(err, codes.NotFound) {
			return nil
		}
		return err
	}
	_, err = op.Wait(ctx)
	return err
}
