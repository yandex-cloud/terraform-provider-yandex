package yandex_serverless_workflow_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/workflows/v1"
	workflowsv1sdk "github.com/yandex-cloud/go-sdk/services/serverless/workflows/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const serverlessWorkflowResource = "yandex_serverless_workflow.test-wf"

const minimalWorkflowSpecYAML = `yawl: "1.0"
start: s
steps:
  s:
    noOp: {}
`

const updatedWorkflowSpecYAML = `yawl: "1.0"
start: s
steps:
  s:
    noOp:
      output: updated
`

func init() {
	resource.AddTestSweepers("yandex_serverless_workflow", &resource.Sweeper{
		Name: "yandex_serverless_workflow",
		F:    testSweepServerlessWorkflow,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepServerlessWorkflow(_ string) error {
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

func TestAccServerlessWorkflow_basic(t *testing.T) {
	var wf workflows.Workflow
	name := acctest.RandomWithPrefix("tf-wf")
	desc := acctest.RandomWithPrefix("tf-wf-desc")
	labelKey := acctest.RandomWithPrefix("tf-wf-label")
	labelValue := acctest.RandomWithPrefix("tf-wf-label-value")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testYandexServerlessWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexServerlessWorkflowBasic(name, desc, labelKey, labelValue, minimalWorkflowSpecYAML),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessWorkflowExists(serverlessWorkflowResource, &wf),
					resource.TestCheckResourceAttr(serverlessWorkflowResource, "name", name),
					resource.TestCheckResourceAttr(serverlessWorkflowResource, "description", desc),
					resource.TestCheckResourceAttrSet(serverlessWorkflowResource, "folder_id"),
					resource.TestCheckResourceAttrSet(serverlessWorkflowResource, "workflow_id"),
					testYandexServerlessWorkflowContainsLabel(&wf, labelKey, labelValue),
					test.AccCheckCreatedAtAttr(serverlessWorkflowResource),
				),
			},
			serverlessWorkflowImportTestStep(),
		},
	})
}

func TestAccServerlessWorkflow_update(t *testing.T) {
	var wf workflows.Workflow
	var wfSpecificationUpdated workflows.Workflow
	var wfUpdated workflows.Workflow
	name := acctest.RandomWithPrefix("tf-wf")
	desc := acctest.RandomWithPrefix("tf-wf-desc")
	labelKey := acctest.RandomWithPrefix("tf-wf-label")
	labelValue := acctest.RandomWithPrefix("tf-wf-label-value")

	nameUpdated := acctest.RandomWithPrefix("tf-wf-1")
	descUpdated := acctest.RandomWithPrefix("tf-wf-desc-1")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-wf-label-1")
	labelValueUpdated := acctest.RandomWithPrefix("tf-wf-label-value-1")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testYandexServerlessWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexServerlessWorkflowBasic(name, desc, labelKey, labelValue, minimalWorkflowSpecYAML),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessWorkflowExists(serverlessWorkflowResource, &wf),
				),
			},
			serverlessWorkflowImportTestStep(),
			{
				Config: testYandexServerlessWorkflowBasic(name, desc, labelKey, labelValue, updatedWorkflowSpecYAML),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessWorkflowExists(serverlessWorkflowResource, &wfSpecificationUpdated),
					resource.TestCheckResourceAttrWith(serverlessWorkflowResource, "id", func(w *workflows.Workflow) resource.CheckResourceAttrWithFunc {
						return func(id string) error {
							if id == w.Id {
								return nil
							}
							return errors.New("invalid Serverless Workflow id")
						}
					}(&wf)),
					resource.TestCheckResourceAttr(serverlessWorkflowResource, "specification.spec_yaml", updatedWorkflowSpecYAML),
					testYandexServerlessWorkflowHasSpecification(&wfSpecificationUpdated, updatedWorkflowSpecYAML),
				),
			},
			{
				Config: testYandexServerlessWorkflowBasic(nameUpdated, descUpdated, labelKeyUpdated, labelValueUpdated, updatedWorkflowSpecYAML),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessWorkflowExists(serverlessWorkflowResource, &wfUpdated),
					resource.TestCheckResourceAttrWith(serverlessWorkflowResource, "id", func(w *workflows.Workflow) resource.CheckResourceAttrWithFunc {
						return func(id string) error {
							if id == w.Id {
								return nil
							}
							return errors.New("invalid Serverless Workflow id")
						}
					}(&wf)),
					resource.TestCheckResourceAttr(serverlessWorkflowResource, "name", nameUpdated),
					resource.TestCheckResourceAttr(serverlessWorkflowResource, "description", descUpdated),
					testYandexServerlessWorkflowContainsLabel(&wfUpdated, labelKeyUpdated, labelValueUpdated),
					resource.TestCheckResourceAttr(serverlessWorkflowResource, "specification.spec_yaml", updatedWorkflowSpecYAML),
					testYandexServerlessWorkflowHasSpecification(&wfUpdated, updatedWorkflowSpecYAML),
					test.AccCheckCreatedAtAttr(serverlessWorkflowResource),
				),
			},
			serverlessWorkflowImportTestStep(),
		},
	})
}

func testYandexServerlessWorkflowDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_serverless_workflow" {
			continue
		}

		_, err := testGetServerlessWorkflowByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Serverless Workflow still exists")
		}
	}

	return nil
}

func serverlessWorkflowImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      serverlessWorkflowResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testYandexServerlessWorkflowExists(name string, wf *workflows.Workflow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := testGetServerlessWorkflowByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Serverless Workflow not found")
		}

		*wf = *found
		return nil
	}
}

func testGetServerlessWorkflowByID(conf provider_config.Config, id string) (*workflows.Workflow, error) {
	req := workflows.GetWorkflowRequest{
		WorkflowId: id,
	}

	resp, err := workflowsv1sdk.NewWorkflowClient(conf.SDKv2).Get(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	return resp.GetWorkflow(), nil
}

func testYandexServerlessWorkflowContainsLabel(wf *workflows.Workflow, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := wf.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testYandexServerlessWorkflowHasSpecification(wf *workflows.Workflow, specYAML string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if actual := wf.GetSpecification().GetSpecYaml(); actual != specYAML {
			return fmt.Errorf("incorrect workflow specification: expected %q, got %q", specYAML, actual)
		}
		return nil
	}
}

func testYandexServerlessWorkflowBasic(name, desc, labelKey, labelValue, specYAML string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_workflow" "test-wf" {
  name        = "{{.name}}"
  description = "{{.description}}"
  folder_id   = "{{.folder_id}}"
  labels = {
    {{.label_key}} = "{{.label_value}}"
  }
  specification {
    spec_yaml = <<-EOT
{{.spec_yaml}}
EOT
  }
}`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   test.GetExampleFolderID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
		"spec_yaml":   specYAML,
	})
	return buf.String()
}
