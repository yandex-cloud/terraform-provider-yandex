package sweep

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	lockboxsdk "github.com/yandex-cloud/go-sdk/services/lockbox/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_lockbox "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/lockbox/pluginsdk"
	"google.golang.org/grpc/codes"
)

func registerResourceYandexLockboxSecret(add func(string, *resource.Sweeper)) {
	add("yandex_lockbox_secret", &resource.Sweeper{
		Name: "yandex_lockbox_secret",
		F:    sweepAllLockboxSecret,
	})
}

func sweepAllLockboxSecret(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	client := lockboxsdk.NewSecretClient(conf.SDK)

	req := &lockbox.ListSecretsRequest{FolderId: conf.FolderID}
	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepLockboxSecret(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep lockbox secret %q", id))
		}
	}
	if err := it.Error(); err != nil {
		return err
	}

	return result.ErrorOrNil()
}

func sweepLockboxSecret(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepLockboxSecretOnce, conf, "Lockbox secret", id)
}

func sweepLockboxSecretOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_lockbox.YandexLockboxSecretDefaultTimeout)
	defer cancel()

	client := lockboxsdk.NewSecretClient(conf.SDK)

	op, err := client.Delete(ctx, &lockbox.DeleteSecretRequest{
		SecretId: id,
	})
	if err == nil {
		_, err = op.Wait(ctx)
	}
	if sdkcommon.IsStatusWithCode(err, codes.NotFound) {
		return nil
	}
	if sdkcommon.IsStatusWithCode(err, codes.FailedPrecondition) &&
		strings.Contains(sdkcommon.ErrorMessage(err), "management is only allowed to") {
		return nil
	}
	return err
}
