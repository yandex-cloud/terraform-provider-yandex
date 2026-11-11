package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/grpc/codes"

	backuppb "github.com/yandex-cloud/go-genproto/yandex/cloud/backup/v1"
	backupsdk "github.com/yandex-cloud/go-sdk/services/backup/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_backup "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/backup/pluginsdk"
)

func registerResourceYandexBackupPolicy(add func(string, *resource.Sweeper)) {
	add("yandex_backup_policy", &resource.Sweeper{
		Name:         "yandex_backup_policy",
		F:            sweepAllBackupPolicy,
		Dependencies: []string{},
	})
}

func sweepAllBackupPolicy(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &backuppb.ListPoliciesRequest{FolderId: conf.FolderID}
	client := backupsdk.NewPolicyClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}

	for it.Next() {
		id := it.Value().GetId()
		if !sweepBackupPolicy(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Backup Policy %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepBackupPolicy(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetryByFunc(conf, "Backup Policy", func(conf *providerconfig.Config) error {
		ctx, cancel := conf.ContextWithTimeout(svc_backup.YandexBackupDefaultTimeout)
		defer cancel()

		client := backupsdk.NewPolicyClient(conf.SDK)

		op, err := client.Delete(ctx, &backuppb.DeletePolicyRequest{
			PolicyId: id,
		})
		if err != nil {
			if sdkcommon.IsStatusWithCode(err, codes.NotFound) {
				return nil
			}
			return err
		}
		_, err = op.Wait(ctx)
		return err
	})
}
