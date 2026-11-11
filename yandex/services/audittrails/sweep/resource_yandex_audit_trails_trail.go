package sweep

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/audittrails/v1"
	audittrailssdk "github.com/yandex-cloud/go-sdk/services/audittrails/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/grpc/codes"
)

func registerResourceYandexAuditTrailsTrail(add func(string, *resource.Sweeper)) {
	add("yandex_audit_trails_trail", &resource.Sweeper{
		Name: "yandex_audit_trails_trail",
		F:    sweepAllAuditTrails,
	})
}

// Sweep function deletes all trails (used resources are deleted by their sweepers)
func sweepAllAuditTrails(_ string) error {
	log.Printf("[DEBUG] Sweeping Audit Trails Trail")
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}

	req := &audittrails.ListTrailsRequest{FolderId: conf.FolderID}
	client := audittrailssdk.NewTrailClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	for it.Next() {
		id := it.Value().GetId()

		if !sweepAuditTrailsTrail(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Audit Trails Trail %q", id))
		}
	}
	if err := it.Error(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func sweepAuditTrailsTrail(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepAuditTrailsTrailOnce, conf, "Audit Trails Trail", id)
}

func sweepAuditTrailsTrailOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(5 * time.Minute)
	defer cancel()

	client := audittrailssdk.NewTrailClient(conf.SDK)

	op, err := client.Delete(ctx, &audittrails.DeleteTrailRequest{
		TrailId: id,
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
