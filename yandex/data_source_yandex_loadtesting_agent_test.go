package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	lt "github.com/yandex-cloud/go-genproto/yandex/cloud/loadtesting/api/v1"
	ltagent "github.com/yandex-cloud/go-genproto/yandex/cloud/loadtesting/api/v1/agent"
)

const agentDataSourceResource = "data.yandex_loadtesting_agent.test-lt-agent-ds"

func TestAccDataSourceLoadtestingAgent_byID(t *testing.T) {
	t.Parallel()

	agentName := acctest.RandomWithPrefix("ds-tf-loadtesting-agent")
	agentDescription := "Agent Desc for test"
	folderID := getExampleFolderID()

	var agent ltagent.Agent

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLoadtestingAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceYandexLoadtestingAgentConfigByID(agentName, agentDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceYandexLoadtestingAgentExists(agentDataSourceResource, &agent),
					testAccCheckResourceIDField(agentDataSourceResource, "agent_id"),
					resource.TestCheckResourceAttr(agentDataSourceResource, "name", agentName),
					resource.TestCheckResourceAttr(agentDataSourceResource, "description", agentDescription),
					resource.TestCheckResourceAttr(agentDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(agentDataSourceResource, "labels.purpose", "grpc-scenario"),
					resource.TestCheckResourceAttr(agentDataSourceResource, "labels.pandora", "0-5-20"),
					testAccCheckLoadtestingAgentLabel(&agent, "purpose", "grpc-scenario"),
					testAccCheckLoadtestingAgentLabel(&agent, "pandora", "0-5-20"),
					resource.TestCheckResourceAttrSet(agentDataSourceResource, "log_settings.0.log_group_id"),
					resource.TestCheckResourceAttrSet(agentDataSourceResource, "compute_instance_id"),
					resource.TestCheckResourceAttrSet(agentDataSourceResource, "compute_instance.0.service_account_id"),
					resource.TestCheckResourceAttr(agentDataSourceResource, "compute_instance.0.resources.0.memory", "4"),
					resource.TestCheckResourceAttr(agentDataSourceResource, "compute_instance.0.resources.0.cores", "4"),
					resource.TestCheckResourceAttr(agentDataSourceResource, "compute_instance.0.computed_labels.purpose", "loadtesting-agent"),
					resource.TestCheckResourceAttr(agentDataSourceResource, "compute_instance.0.computed_metadata.field1", "metavalue1"),
					resource.TestCheckResourceAttr(agentDataSourceResource, "compute_instance.0.computed_metadata.field2", "other value 2"),
					resource.TestCheckResourceAttr(agentDataSourceResource, "compute_instance.0.platform_id", "standard-v1"),
					resource.TestCheckResourceAttrSet(agentDataSourceResource, "compute_instance.0.computed_metadata.loadtesting-created"),
				),
			},
		},
	})
}

func testAccDataSourceYandexLoadtestingAgentExists(n string, agent *ltagent.Agent) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.Loadtesting().Agent().Get(context.Background(), &lt.GetAgentRequest{
			AgentId: ds.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != ds.Primary.ID {
			return fmt.Errorf("Loadtesting Agent not found")
		}

		*agent = *found

		return nil
	}
}

func testAccDataSourceYandexLoadtestingAgentConfigByID(name, desc string) string {
	return `
data "yandex_loadtesting_agent" "test-lt-agent-ds" {
  agent_id = "${yandex_loadtesting_agent.test-lt-agent.id}"
}
` + testAccLoadtestingAgentFull(name, desc)
}
