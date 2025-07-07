package gitlab_instance_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/gitlab/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	pageSize      = 1000
	deleteTimeout = 30 * time.Minute
	updateTimeout = 30 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_gitlab_instance", &resource.Sweeper{
		Name: "yandex_gitlab_instance",
		F:    testSweepGitlabInstance,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepGitlabInstance(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.SDK.Gitlab().Instance().List(context.Background(), &gitlab.ListInstancesRequest{
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
	op, err := conf.SDK.Gitlab().Instance().Update(ctxUpd, &gitlab.UpdateInstanceRequest{
		InstanceId:         id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = testhelpers.HandleSweepOperation(ctxUpd, conf, op, err)
	if err != nil && !strings.EqualFold(testhelpers.ErrorMessage(err), "no changes detected") {
		return err
	}

	ctxDel, cancelDel := context.WithTimeout(context.Background(), deleteTimeout)
	defer cancelDel()
	op, err = conf.SDK.Gitlab().Instance().Delete(ctxDel, &gitlab.DeleteInstanceRequest{
		InstanceId: id,
	})
	return testhelpers.HandleSweepOperation(ctxDel, conf, op, err)
}
