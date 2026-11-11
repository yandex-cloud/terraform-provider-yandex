package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
	certificatemanagersdk "github.com/yandex-cloud/go-sdk/services/certificatemanager/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_certificate_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/certificatemanager/pluginsdk"
	"google.golang.org/grpc/codes"
)

func registerResourceYandexCmCertificate(add func(string, *resource.Sweeper)) {
	add("yandex_cm_certificate", &resource.Sweeper{
		Name: "yandex_cm_certificate",
		F:    sweepAllCMCertificate,
	})
}

func sweepAllCMCertificate(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &certificatemanager.ListCertificatesRequest{FolderId: conf.FolderID}
	client := certificatemanagersdk.NewCertificateClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		certificate := it.Value()

		if len(certificate.Labels) > 0 {
			if certificate.Labels["sweeper-skip-deletion"] == "1" {
				continue
			}
		}

		id := certificate.GetId()
		if !sweepCMCertificate(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep certificate %q", id))
		}
	}
	if err := it.Error(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func sweepCMCertificate(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepCMCertificateOnce, conf, "Certificate", id)
}

func sweepCMCertificateOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_certificate_manager.YandexCMCertificateDefaultTimeout)
	defer cancel()

	client := certificatemanagersdk.NewCertificateClient(conf.SDK)

	op, err := client.Delete(ctx, &certificatemanager.DeleteCertificateRequest{
		CertificateId: id,
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
