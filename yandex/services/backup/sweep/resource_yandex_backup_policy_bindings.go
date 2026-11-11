package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	backuppb "github.com/yandex-cloud/go-genproto/yandex/cloud/backup/v1"
	backupsdk "github.com/yandex-cloud/go-sdk/services/backup/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_backup "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/backup/pluginsdk"
	"google.golang.org/grpc/codes"
)

func registerResourceYandexBackupPolicyBindings(add func(string, *resource.Sweeper)) {
	add("yandex_backup_policy_bindings", &resource.Sweeper{
		Name: "yandex_backup_policy_bindings",
		F:    sweepAllBackupPolicyBindings,
	})
}

func sweepAllBackupPolicyBindings(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &backuppb.ListApplicationsRequest{
		Id: &backuppb.ListApplicationsRequest_FolderId{
			FolderId: conf.FolderID,
		},
	}
	client := backupsdk.NewPolicyClient(conf.SDK)

	it := client.ApplicationsIterator(conf.Context(), req)
	result := &multierror.Error{}

	for it.Next() {
		id := svc_backup.MakeBackupPolicyBindingsID(it.Value().GetPolicyId(), it.Value().GetComputeInstanceId())
		if !sweepBackupPolicyBindings(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Backup Policy %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepBackupPolicyBindings(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetryByFunc(conf, "Backup Policy Bindings", func(conf *providerconfig.Config) error {
		ctx, cancel := conf.ContextWithTimeout(svc_backup.YandexBackupDefaultTimeout)
		defer cancel()

		policyID, instanceID, err := svc_backup.ParseBackupPolicyBindingsID(id)
		if err != nil {
			return err
		}

		client := backupsdk.NewPolicyClient(conf.SDK)

		op, err := client.Revoke(ctx, &backuppb.RevokeRequest{
			PolicyId:          policyID,
			ComputeInstanceId: instanceID,
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
