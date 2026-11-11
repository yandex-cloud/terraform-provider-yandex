package sweep

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	resourcemanagersdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_iam "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/pluginsdk"
	svc_resource_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/pluginsdk"
)

// NOTE(dxan): function may return non-empty string and non-nil error. Example:
// Resource is successfully created, but wait fails: the function returns id and wait error
func CreateIAMServiceAccountForSweeper(conf *providerconfig.Config) (string, error) {
	ctx, cancel := conf.ContextWithTimeout(svc_iam.YandexIAMServiceAccountDefaultTimeout)
	defer cancel()
	client := iamsdk.NewServiceAccountClient(conf.SDK)

	op, err := client.Create(ctx, &iam.CreateServiceAccountRequest{
		FolderId:    conf.FolderID,
		Name:        acctest.RandomWithPrefix("sweeper"),
		Description: "created by sweeper",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create service account: %v", err)
	}

	md := op.Metadata()
	id := md.ServiceAccountId

	_, err = op.Wait(ctx)
	if err != nil {
		return id, fmt.Errorf("error while waiting for create service account operation: %v", err)
	}

	err = assignEditorRoleToSweeperServiceAccount(conf, id)
	if err != nil {
		return id, err
	}

	return md.ServiceAccountId, nil
}

func assignEditorRoleToSweeperServiceAccount(conf *providerconfig.Config, saID string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_resource_manager.YandexResourceManagerFolderDefaultTimeout)
	defer cancel()
	const role_EDITOR = "editor"
	client := resourcemanagersdk.NewFolderClient(conf.SDK)

	op, err := client.UpdateAccessBindings(ctx, &access.UpdateAccessBindingsRequest{
		ResourceId: conf.FolderID,
		AccessBindingDeltas: []*access.AccessBindingDelta{
			{
				Action: access.AccessBindingAction_ADD,
				AccessBinding: &access.AccessBinding{
					RoleId: role_EDITOR,
					Subject: &access.Subject{
						Id:   saID,
						Type: "serviceAccount",
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to assign '%s' role to the service account %q: %v", role_EDITOR, saID, err)
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for grant access bindings operation '%q': %v", op.ID(), err)
	}

	sweep.DebugLog("Service account '%s' was granted role '%s' to folder ID '%s'", saID, role_EDITOR, conf.FolderID)

	return nil
}

func SweepIAMServiceAccount(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepIAMServiceAccountOnce, conf, "IAM Service Account", id)
}

func sweepIAMServiceAccountOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(1 * time.Minute)
	defer cancel()

	client := iamsdk.NewServiceAccountClient(conf.SDK)

	op, err := client.Delete(ctx, &iam.DeleteServiceAccountRequest{
		ServiceAccountId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}

func registerResourceYandexIamServiceAccount(add func(string, *resource.Sweeper)) {
	add("yandex_iam_service_account", &resource.Sweeper{
		Name: "yandex_iam_service_account",
		F:    sweepAllIAMServiceAccounts,
		Dependencies: []string{
			"yandex_dataproc_cluster",
			"yandex_kubernetes_cluster",
			"yandex_compute_instance_group",
			"yandex_audit_trails_trail",
			"yandex_mdb_opensearch_cluster",
		},
	})
}

func sweepAllIAMServiceAccounts(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &iam.ListServiceAccountsRequest{FolderId: conf.FolderID}
	client := iamsdk.NewServiceAccountClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !SweepIAMServiceAccount(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep IAM service account %q", id))
		}
	}

	return result.ErrorOrNil()
}
