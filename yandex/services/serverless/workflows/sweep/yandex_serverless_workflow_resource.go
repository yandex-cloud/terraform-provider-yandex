package sweep

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/workflows/v1"
	workflowsv1sdk "github.com/yandex-cloud/go-sdk/services/serverless/workflows/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func registerYandexServerlessWorkflowResource(add func(string, *resource.Sweeper)) {
	add("yandex_serverless_workflow", &resource.Sweeper{
		Name: "yandex_serverless_workflow",
		F:    sweepAllServerlessWorkflow,
	})
}

func sweepAllServerlessWorkflow(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &workflows.ListWorkflowsRequest{FolderId: conf.ProviderState.FolderID.ValueString()}
	resp, err := workflowsv1sdk.NewWorkflowClient(conf.SDKv2).List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting workflows: %s", err)
	}

	result := &multierror.Error{}
	for _, wf := range resp.Workflows {
		id := wf.GetId()
		if !test.SweepWithRetry(sweepServerlessWorkflowOnce, conf, "Serverless Workflow", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Serverless Workflow %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepServerlessWorkflowOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	op, err := workflowsv1sdk.NewWorkflowClient(conf.SDKv2).Delete(ctx, &workflows.DeleteWorkflowRequest{
		WorkflowId: id,
	})
	if err != nil {
		return err
	}
	_, err = op.Wait(ctx)
	return err
}
