package acceptance

import (
	"context"
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	loadbalancersdk "github.com/yandex-cloud/go-sdk/services/loadbalancer/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

const nlbDataSourceResource = "data.yandex_lb_network_load_balancer.test-nlb-ds"

func TestAccDataSourceLBNetworkLoadBalancer_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	nlbName := acctest.RandomWithPrefix("tf-nlb")
	nlbDesc := "Description for test"
	folderID := acceptance.FolderID()

	var nlb loadbalancer.NetworkLoadBalancer

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLBNetworkLoadBalancerConfigByID(nlbName, nlbDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccDataSourceLBNetworkLoadBalancerExists(nlbDataSourceResource, &nlb),
					statecheck.CheckResourceIDField(nlbDataSourceResource, "network_load_balancer_id"),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "name", nlbName),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "description", nlbDesc),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "listener.#", "0"),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "attached_target_group.#", "0"),
					statecheck.CheckCreatedAtAttr(nlbDataSourceResource),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 0, 0, nil, nil),
				),
			},
		},
	})
}

func TestAccDataSourceLBNetworkLoadBalancer_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	nlbName := acctest.RandomWithPrefix("tf-nlb")
	nlbDesc := "Description for test"
	folderID := acceptance.FolderID()

	var nlb loadbalancer.NetworkLoadBalancer

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLBNetworkLoadBalancerConfigByName(nlbName, nlbDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccDataSourceLBNetworkLoadBalancerExists(nlbDataSourceResource, &nlb),
					statecheck.CheckResourceIDField(nlbDataSourceResource, "network_load_balancer_id"),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "name", nlbName),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "description", nlbDesc),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "listener.#", "0"),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "attached_target_group.#", "0"),
					statecheck.CheckCreatedAtAttr(nlbDataSourceResource),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 0, 0, nil, nil),
				),
			},
		},
	})
}

func TestAccDataSourceLBNetworkLoadBalancer_full(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var nlb loadbalancer.NetworkLoadBalancer
	listenerPath := ""
	atgPath := ""
	nlbValues := lbDefaultNLBValues()
	folderID := acceptance.FolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBGeneralNLBTemplate(nlbValues, true, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccDataSourceLBNetworkLoadBalancerExists(nlbDataSourceResource, &nlb),
					statecheck.CheckResourceIDField(nlbDataSourceResource, "network_load_balancer_id"),
					resource.TestCheckResourceAttr(
						nlbDataSourceResource, "name", nlbValues["NLBName"].(string),
					),
					resource.TestCheckResourceAttr(
						nlbDataSourceResource, "description", lbDefaultNLBDescription,
					),
					resource.TestCheckResourceAttr(
						nlbDataSourceResource, "folder_id", folderID,
					),
					resource.TestCheckResourceAttr(
						nlbDataSourceResource, "deletion_protection", "false",
					),
					resource.TestCheckResourceAttr(
						nlbDataSourceResource, "listener.#", "1",
					),
					resource.TestCheckResourceAttr(
						nlbDataSourceResource, "attached_target_group.#", "1",
					),
					statecheck.TestExistsElementWithAttrValue(
						nlbDataSourceResource, "listener", "name", lbDefaultListenerName, &listenerPath,
					),
					statecheck.TestCheckResourceSubAttr(
						nlbDataSourceResource, &listenerPath, "port", fmt.Sprintf("%d", lbDefaultListenerPort),
					),
					statecheck.TestCheckResourceSubAttr(
						nlbDataSourceResource, &listenerPath, "target_port", fmt.Sprintf("%d", lbDefaultListenerTargetPort),
					),
					statecheck.TestCheckResourceSubAttr(
						nlbDataSourceResource, &listenerPath, "protocol", lbDefaultListenerProtocol,
					),
					statecheck.WithState(
						func() resource.TestCheckFunc {
							return statecheck.TestCheckResourceSubAttr(
								nlbDataSourceResource, &listenerPath,
								"external_address_spec.0.ip_version",
								lbDefaultListenerIPVersion,
							)
						},
					),
					statecheck.WithState(
						func() resource.TestCheckFunc {
							return statecheck.CheckResourceSubAttrFn(
								nlbDataSourceResource, &listenerPath,
								"external_address_spec.0.address",
								func(value string) error {
									address := nlb.GetListeners()[0].GetAddress()
									if value != address {
										return fmt.Errorf("NetworkLoadBalancer's listener's address doesn't match. %s != %s", value, address)
									}
									return nil
								},
							)
						},
					),
					statecheck.TestExistsElementWithAttrValue(
						nlbDataSourceResource, "attached_target_group", "healthcheck.0.name", lbDefaultHCHTTPName, &atgPath,
					),
					statecheck.TestCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.timeout", fmt.Sprintf("%d", lbDefaultHCHTTPTimeout),
					),
					statecheck.TestCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.interval", fmt.Sprintf("%d", lbDefaultHCHTTPInterval),
					),
					statecheck.TestCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.healthy_threshold", fmt.Sprintf("%d", lbDefaultHCHTTPHealthyTreshold),
					),
					statecheck.TestCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.unhealthy_threshold", fmt.Sprintf("%d", lbDefaultHCHTTPUnhealthyTreshold),
					),
					statecheck.TestCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.tcp_options.#", "0",
					),
					statecheck.TestCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.http_options.0.port", fmt.Sprintf("%d", lbDefaultHCHTTPPort),
					),
					statecheck.TestCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.http_options.0.path", lbDefaultHCHTTPPath,
					),
					statecheck.CheckResourceSubAttrFn(
						nlbDataSourceResource, &atgPath, "target_group_id", func(value string) error {
							targetGroupID := nlb.GetAttachedTargetGroups()[0].TargetGroupId
							if value != targetGroupID {
								return fmt.Errorf("NetworkLoadBalancer's atg's target_group_id doesn't match. %s != %s", value, targetGroupID)
							}
							return nil
						},
					),
					testAccCheckLBNetworkLoadBalancerValues(
						&nlb, 1, 1,
						func(ls *loadbalancer.Listener) error {
							return checkLBListener(ls, lbDefaultListenerName, lbDefaultListenerPort, lbDefaultListenerTargetPort)
						},
						func(atg *loadbalancer.AttachedTargetGroup) error {
							return checkLBAttachedTargetGroup(
								atg, lbDefaultHCHTTPName,
								lbDefaultHCHTTPInterval, lbDefaultHCHTTPTimeout,
								lbDefaultHCHTTPHealthyTreshold, lbDefaultHCHTTPUnhealthyTreshold,
								lbDefaultHCHTTPPort, lbDefaultHCHTTPPath,
							)
						},
					),
				),
			},
		},
	})
}

func (suite testSuite) testAccDataSourceLBNetworkLoadBalancerExists(n string, nlb *loadbalancer.NetworkLoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := loadbalancersdk.NewNetworkLoadBalancerClient(config.SDK)

		found, err := client.Get(context.Background(), &loadbalancer.GetNetworkLoadBalancerRequest{
			NetworkLoadBalancerId: ds.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != ds.Primary.ID {
			return fmt.Errorf("NetworkLoadBalancer not found")
		}

		*nlb = *found

		return nil
	}
}

func testAccDataSourceLBNetworkLoadBalancerConfigByID(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_lb_network_load_balancer" "test-nlb-ds" {
  network_load_balancer_id = "${yandex_lb_network_load_balancer.test-nlb.id}"
}

resource "yandex_lb_network_load_balancer" "test-nlb" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}

func testAccDataSourceLBNetworkLoadBalancerConfigByName(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_lb_network_load_balancer" "test-nlb-ds" {
  name = "${yandex_lb_network_load_balancer.test-nlb.name}"
}

resource "yandex_lb_network_load_balancer" "test-nlb" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}
