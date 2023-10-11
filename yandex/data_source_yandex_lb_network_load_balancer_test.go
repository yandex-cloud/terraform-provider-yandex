package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
)

const nlbDataSourceResource = "data.yandex_lb_network_load_balancer.test-nlb-ds"

func TestAccDataSourceLBNetworkLoadBalancer_byID(t *testing.T) {
	t.Parallel()

	nlbName := acctest.RandomWithPrefix("tf-nlb")
	nlbDesc := "Description for test"
	folderID := getExampleFolderID()

	var nlb loadbalancer.NetworkLoadBalancer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLBNetworkLoadBalancerConfigByID(nlbName, nlbDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceLBNetworkLoadBalancerExists(nlbDataSourceResource, &nlb),
					testAccCheckResourceIDField(nlbDataSourceResource, "network_load_balancer_id"),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "name", nlbName),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "description", nlbDesc),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "listener.#", "0"),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "attached_target_group.#", "0"),
					testAccCheckCreatedAtAttr(nlbDataSourceResource),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 0, 0, nil, nil),
				),
			},
		},
	})
}

func TestAccDataSourceLBNetworkLoadBalancer_byName(t *testing.T) {
	t.Parallel()

	nlbName := acctest.RandomWithPrefix("tf-nlb")
	nlbDesc := "Description for test"
	folderID := getExampleFolderID()

	var nlb loadbalancer.NetworkLoadBalancer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLBNetworkLoadBalancerConfigByName(nlbName, nlbDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceLBNetworkLoadBalancerExists(nlbDataSourceResource, &nlb),
					testAccCheckResourceIDField(nlbDataSourceResource, "network_load_balancer_id"),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "name", nlbName),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "description", nlbDesc),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "listener.#", "0"),
					resource.TestCheckResourceAttr(nlbDataSourceResource, "attached_target_group.#", "0"),
					testAccCheckCreatedAtAttr(nlbDataSourceResource),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 0, 0, nil, nil),
				),
			},
		},
	})
}

func TestAccDataSourceLBNetworkLoadBalancer_full(t *testing.T) {
	t.Parallel()

	var nlb loadbalancer.NetworkLoadBalancer
	listenerPath := ""
	atgPath := ""
	nlbValues := lbDefaultNLBValues()
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBGeneralNLBTemplate(nlbValues, true, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceLBNetworkLoadBalancerExists(nlbDataSourceResource, &nlb),
					testAccCheckResourceIDField(nlbDataSourceResource, "network_load_balancer_id"),
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
					testExistsElementWithAttrValue(
						nlbDataSourceResource, "listener", "name", lbDefaultListenerName, &listenerPath,
					),
					testCheckResourceSubAttr(
						nlbDataSourceResource, &listenerPath, "port", fmt.Sprintf("%d", lbDefaultListenerPort),
					),
					testCheckResourceSubAttr(
						nlbDataSourceResource, &listenerPath, "target_port", fmt.Sprintf("%d", lbDefaultListenerTargetPort),
					),
					testCheckResourceSubAttr(
						nlbDataSourceResource, &listenerPath, "protocol", lbDefaultListenerProtocol,
					),
					checkWithState(
						func() resource.TestCheckFunc {
							return testCheckResourceSubAttr(
								nlbDataSourceResource, &listenerPath,
								"external_address_spec.0.ip_version",
								lbDefaultListenerIPVersion,
							)
						},
					),
					checkWithState(
						func() resource.TestCheckFunc {
							return testCheckResourceSubAttrFn(
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
					testExistsElementWithAttrValue(
						nlbDataSourceResource, "attached_target_group", "healthcheck.0.name", lbDefaultHCHTTPName, &atgPath,
					),
					testCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.timeout", fmt.Sprintf("%d", lbDefaultHCHTTPTimeout),
					),
					testCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.interval", fmt.Sprintf("%d", lbDefaultHCHTTPInterval),
					),
					testCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.healthy_threshold", fmt.Sprintf("%d", lbDefaultHCHTTPHealthyTreshold),
					),
					testCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.unhealthy_threshold", fmt.Sprintf("%d", lbDefaultHCHTTPUnhealthyTreshold),
					),
					testCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.tcp_options.#", "0",
					),
					testCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.http_options.0.port", fmt.Sprintf("%d", lbDefaultHCHTTPPort),
					),
					testCheckResourceSubAttr(
						nlbDataSourceResource, &atgPath, "healthcheck.0.http_options.0.path", lbDefaultHCHTTPPath,
					),
					testCheckResourceSubAttrFn(
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

func testAccDataSourceLBNetworkLoadBalancerExists(n string, nlb *loadbalancer.NetworkLoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.LoadBalancer().NetworkLoadBalancer().Get(context.Background(), &loadbalancer.GetNetworkLoadBalancerRequest{
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
