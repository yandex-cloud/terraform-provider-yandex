package sweep

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/gitlab/v1"
	gitlabsdk "github.com/yandex-cloud/go-sdk/services/gitlab/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	pageSize      = 1000
	deleteTimeout = 30 * time.Minute
	updateTimeout = 30 * time.Minute
)

func registerGitlabInstanceSweepers(add func(string, *resource.Sweeper)) {
	add("yandex_gitlab_instance", &resource.Sweeper{
		Name: "yandex_gitlab_instance",
		F:    sweepAllGitlabInstance,
	})
}

func sweepAllGitlabInstance(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := gitlabsdk.NewInstanceClient(conf.SDKv2).List(context.Background(), &gitlab.ListInstancesRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
		PageSize: pageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Gitlab instances: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Instances {
		if !sweepGitlabInstance(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Gitlab instance %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepGitlabInstance(conf *config.Config, id string) bool {
	return testhelpers.SweepWithRetry(sweepGitlabInstanceOnce, conf, "Gitalb instance", id)
}

func sweepGitlabInstanceOnce(conf *config.Config, id string) error {
	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	ctxUpd, cancelUpd := context.WithTimeout(context.Background(), updateTimeout)
	defer cancelUpd()
	op, err := gitlabsdk.NewInstanceClient(conf.SDKv2).Update(ctxUpd, &gitlab.UpdateInstanceRequest{
		InstanceId:         id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = testhelpers.HandleSweepOperation(ctxUpd, op, err)
	if err != nil && !strings.EqualFold(testhelpers.ErrorMessage(err), "no changes detected") {
		return err
	}

	ctxDel, cancelDel := context.WithTimeout(context.Background(), deleteTimeout)
	defer cancelDel()
	deleteOp, err := gitlabsdk.NewInstanceClient(conf.SDKv2).Delete(ctxDel, &gitlab.DeleteInstanceRequest{
		InstanceId: id,
	})
	return testhelpers.HandleSweepOperation(ctxDel, deleteOp, err)
}
