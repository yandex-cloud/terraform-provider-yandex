package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
)

const nlbResource = "yandex_lb_network_load_balancer.test-nlb"

func init() {
	resource.AddTestSweepers("yandex_lb_network_load_balancer", &resource.Sweeper{
		Name: "yandex_lb_network_load_balancer",
		F:    testSweepLBNetworkLoadBalancers,
		Dependencies: []string{
			"yandex_lb_target_group",
		},
	})
}

func testSweepLBNetworkLoadBalancers(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &loadbalancer.ListNetworkLoadBalancersRequest{FolderId: conf.FolderID}
	it := conf.sdk.LoadBalancer().NetworkLoadBalancer().NetworkLoadBalancerIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepLBNetworkLoadBalancer(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Network Load Balancer %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepLBNetworkLoadBalancer(conf *Config, id string) bool {
	return sweepWithRetry(sweepLBNetworkLoadBalancerOnce, conf, "Network Load Balancer", id)
}

func sweepLBNetworkLoadBalancerOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexLBNetworkLoadBalancerDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.LoadBalancer().NetworkLoadBalancer().Delete(ctx, &loadbalancer.DeleteNetworkLoadBalancerRequest{
		NetworkLoadBalancerId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func networkLoadBalancerImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      nlbResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccLBNetworkLoadBalancer_basic(t *testing.T) {
	t.Parallel()

	var nlb loadbalancer.NetworkLoadBalancer
	nlbName := acctest.RandomWithPrefix("tf-network-load-balancer")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBNetworkLoadBalancerBasic(nlbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					resource.TestCheckResourceAttr(nlbResource, "name", nlbName),
					resource.TestCheckResourceAttrSet(nlbResource, "folder_id"),
					resource.TestCheckResourceAttr(nlbResource, "deletion_protection", "false"),
					testAccCheckLBNetworkLoadBalancerContainsLabel(&nlb, "tf-label", "tf-label-value"),
					testAccCheckLBNetworkLoadBalancerContainsLabel(&nlb, "empty-label", ""),
					testAccCheckCreatedAtAttr(nlbResource),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 0, 0, nil, nil),
				),
			},
			networkLoadBalancerImportStep(),
		},
	})
}

func TestAccLBNetworkLoadBalancer_deletion_protection(t *testing.T) {
	t.Parallel()

	var nlb loadbalancer.NetworkLoadBalancer
	nlbName := acctest.RandomWithPrefix("tf-network-load-balancer")
	nlbNewName := acctest.RandomWithPrefix("tf-network-load-balancer")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBNetworkLoadBalancerDeletionProtection(nlbName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					resource.TestCheckResourceAttr(nlbResource, "name", nlbName),
					resource.TestCheckResourceAttr(nlbResource, "deletion_protection", "true"),
				),
			},
			{
				Config: testAccLBNetworkLoadBalancerBasic(nlbNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					resource.TestCheckResourceAttr(nlbResource, "name", nlbNewName),
					resource.TestCheckResourceAttr(nlbResource, "deletion_protection", "true"),
				),
			},
			{
				Config: testAccLBNetworkLoadBalancerDeletionProtection(nlbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					resource.TestCheckResourceAttr(nlbResource, "name", nlbName),
					resource.TestCheckResourceAttr(nlbResource, "deletion_protection", "false"),
				),
			},
			networkLoadBalancerImportStep(),
		},
	})
}

func TestAccLBNetworkLoadBalancer_full(t *testing.T) {
	t.Parallel()

	var nlb loadbalancer.NetworkLoadBalancer
	listenerPath := ""
	atgPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBGeneralNLBTemplate(lbDefaultNLBValues(), false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testExistsElementWithAttrValue(
						nlbResource, "listener", "name", lbDefaultListenerName, &listenerPath,
					),
					checkWithState(
						func() resource.TestCheckFunc {
							return testCheckResourceSubAttrFn(
								nlbResource, &listenerPath,
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
						nlbResource, "attached_target_group", "healthcheck.0.name", lbDefaultHCHTTPName, &atgPath,
					),
					testCheckResourceSubAttrFn(
						nlbResource, &atgPath, "target_group_id", func(value string) error {
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
			networkLoadBalancerImportStep(),
		},
	})
}

func TestAccLBNetworkLoadBalancer_defaults(t *testing.T) {
	var nlb loadbalancer.NetworkLoadBalancer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBNetworkLoadBalancerDefaults(lbDefaultNLBValues()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
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
			networkLoadBalancerImportStep(),
		},
	})
}

func TestAccLBNetworkLoadBalancer_update(t *testing.T) {
	var nlb loadbalancer.NetworkLoadBalancer
	nlbDefaults := lbDefaultNLBValues()
	nlbUpdated := map[string]interface{}{
		"NLBName":               nlbDefaults["NLBName"],
		"TGName":                nlbDefaults["TGName"],
		"NLBDescr":              fmt.Sprintf("%s-updated", lbDefaultNLBDescription),
		"RegionID":              lbDefaultRegionID,
		"NLBType":               lbDefaultNLBType,
		"BaseTemplate":          nlbDefaults["BaseTemplate"],
		"ListenerName":          fmt.Sprintf("%s-updated", lbDefaultListenerName),
		"ListenerPort":          int64(8090),
		"ListenerTargetPort":    int64(8090),
		"ListenerProtocol":      lbDefaultListenerProtocol,
		"ListenerIPVersion":     lbDefaultListenerIPVersion,
		"HTTPName":              fmt.Sprintf("%s-updated", lbDefaultHCHTTPName),
		"HTTPInterval":          3,
		"HTTPTimeout":           2,
		"HTTPHealthyTreshold":   3,
		"HTTPUnhealthyTreshold": 3,
		"HTTPPort":              8090,
		"HTTPPath":              "/new_ping",
	}
	updatedListenerChecker := func(ls *loadbalancer.Listener) error {
		return checkLBListener(
			ls, fmt.Sprintf("%s-updated", lbDefaultListenerName),
			8090, 8090,
		)
	}
	updatedATGChecker := func(atg *loadbalancer.AttachedTargetGroup) error {
		return checkLBAttachedTargetGroup(
			atg, fmt.Sprintf("%s-updated", lbDefaultHCHTTPName),
			3, 2, 3, 3, 8090, "/new_ping",
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBGeneralNLBTemplate(nlbDefaults, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(
						&nlb, 1, 1,
						func(ls *loadbalancer.Listener) error {
							return checkLBListener(
								ls, lbDefaultListenerName, lbDefaultListenerPort, lbDefaultListenerTargetPort,
							)
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
			{
				Config: testAccLBGeneralNLBTemplate(nlbUpdated, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 1, 1, updatedListenerChecker, updatedATGChecker),
				),
			},
			{
				Config: testAccLBGeneralNLBTemplate(nlbUpdated, false, false, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 0, 1, updatedListenerChecker, updatedATGChecker),
				),
			},
			{
				Config: testAccLBGeneralNLBTemplate(nlbUpdated, false, false, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 0, 0, updatedListenerChecker, updatedATGChecker),
				),
			},
			{
				Config: testAccLBGeneralNLBTemplate(nlbUpdated, false, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 0, 0, updatedListenerChecker, updatedATGChecker),
				),
			},
			networkLoadBalancerImportStep(),
		},
	})
}

func TestAccLBNetworkLoadBalancer_update_healthcheck(t *testing.T) {
	var nlb loadbalancer.NetworkLoadBalancer
	nlbDefaults := lbDefaultNLBValues()

	nlbUpdatedNewPing := lbDefaultNLBValues()
	nlbUpdatedNewPing["HTTPPath"] = "/new_ping"

	nlbUpdatedNewPort := copyNlbSettings(nlbUpdatedNewPing)
	nlbUpdatedNewPort["HTTPPort"] = 8090

	nlbUpdatedNewUnhealthyTreshold := copyNlbSettings(nlbUpdatedNewPort)
	nlbUpdatedNewUnhealthyTreshold["HTTPUnhealthyTreshold"] = 7

	nlbUpdatedNewHealthyTreshold := copyNlbSettings(nlbUpdatedNewUnhealthyTreshold)
	nlbUpdatedNewHealthyTreshold["HTTPHealthyTreshold"] = 9

	nlbUpdatedNewHTTPInterval := copyNlbSettings(nlbUpdatedNewHealthyTreshold)
	nlbUpdatedNewHTTPInterval["HTTPInterval"] = 30

	nlbUpdatedNewHTTPTimeout := copyNlbSettings(nlbUpdatedNewHTTPInterval)
	nlbUpdatedNewHTTPTimeout["HTTPTimeout"] = 25

	updatedListenerChecker := func(ls *loadbalancer.Listener) error {
		return nil
	}

	updatedATGChecker := func(settings map[string]interface{}) func(atg *loadbalancer.AttachedTargetGroup) error {
		return func(atg *loadbalancer.AttachedTargetGroup) error {
			return checkLBAttachedTargetGroup(
				atg,
				settings["HTTPName"].(string),
				int64(settings["HTTPInterval"].(int)),
				int64(settings["HTTPTimeout"].(int)),
				int64(settings["HTTPUnhealthyTreshold"].(int)),
				int64(settings["HTTPHealthyTreshold"].(int)),
				int64(settings["HTTPPort"].(int)),
				settings["HTTPPath"].(string),
			)
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBGeneralNLBTemplate(nlbDefaults, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(
						&nlb, 1, 1,
						func(ls *loadbalancer.Listener) error {
							return checkLBListener(
								ls, lbDefaultListenerName, lbDefaultListenerPort, lbDefaultListenerTargetPort,
							)
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
			{
				Config: testAccLBGeneralNLBTemplate(nlbUpdatedNewPing, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 1, 1, updatedListenerChecker, updatedATGChecker(nlbUpdatedNewPing)),
				),
			},
			{
				Config: testAccLBGeneralNLBTemplate(nlbUpdatedNewPort, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 1, 1, updatedListenerChecker, updatedATGChecker(nlbUpdatedNewPort)),
				),
			},
			{
				Config: testAccLBGeneralNLBTemplate(nlbUpdatedNewUnhealthyTreshold, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 1, 1, updatedListenerChecker, updatedATGChecker(nlbUpdatedNewUnhealthyTreshold)),
				),
			},
			{
				Config: testAccLBGeneralNLBTemplate(nlbUpdatedNewHealthyTreshold, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 1, 1, updatedListenerChecker, updatedATGChecker(nlbUpdatedNewHealthyTreshold)),
				),
			},
			{
				Config: testAccLBGeneralNLBTemplate(nlbUpdatedNewHTTPInterval, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 1, 1, updatedListenerChecker, updatedATGChecker(nlbUpdatedNewHTTPInterval)),
				),
			},
			{
				Config: testAccLBGeneralNLBTemplate(nlbUpdatedNewHTTPTimeout, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 1, 1, updatedListenerChecker, updatedATGChecker(nlbUpdatedNewHTTPTimeout)),
				),
			},
			networkLoadBalancerImportStep(),
		},
	})
}

func TestAccLBNetworkLoadBalancer_update_listener(t *testing.T) {
	var nlb loadbalancer.NetworkLoadBalancer
	nlbDefaults := lbDefaultNLBValues()

	newName := copyNlbSettings(nlbDefaults)
	newName["ListenerName"] = newName["ListenerName"].(string) + "1"

	newPort := copyNlbSettings(nlbDefaults)
	newPort["ListenerPort"] = newPort["ListenerPort"].(int64) + 1

	newTargetPort := copyNlbSettings(nlbDefaults)
	newTargetPort["ListenerTargetPort"] = newPort["ListenerTargetPort"].(int64) + 1

	updatedListenerChecker := func(settings map[string]interface{}) func(ls *loadbalancer.Listener) error {
		return func(ls *loadbalancer.Listener) error {
			return checkLBListener(
				ls,
				settings["ListenerName"].(string),
				settings["ListenerPort"].(int64),
				settings["ListenerTargetPort"].(int64),
			)
		}
	}

	updatedATGChecker := func(atg *loadbalancer.AttachedTargetGroup) error {
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBGeneralNLBTemplate(nlbDefaults, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(
						&nlb, 1, 1,
						func(ls *loadbalancer.Listener) error {
							return checkLBListener(
								ls, lbDefaultListenerName, lbDefaultListenerPort, lbDefaultListenerTargetPort,
							)
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
			{
				Config: testAccLBGeneralNLBTemplate(newName, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 1, 1, updatedListenerChecker(newName), updatedATGChecker),
				),
			},
			{
				Config: testAccLBGeneralNLBTemplate(newPort, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 1, 1, updatedListenerChecker(newPort), updatedATGChecker),
				),
			},
			{
				Config: testAccLBGeneralNLBTemplate(newTargetPort, false, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBNetworkLoadBalancerExists(nlbResource, &nlb),
					testAccCheckLBNetworkLoadBalancerValues(&nlb, 1, 1, updatedListenerChecker(newTargetPort), updatedATGChecker),
				),
			},
			networkLoadBalancerImportStep(),
		},
	})
}

func copyNlbSettings(settings map[string]interface{}) map[string]interface{} {
	newSettings := map[string]interface{}{}
	for k, v := range settings {
		newSettings[k] = v
	}
	return newSettings
}

func testAccCheckLBNetworkLoadBalancerDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_lb_network_load_balancer" {
			continue
		}

		_, err := config.sdk.LoadBalancer().NetworkLoadBalancer().Get(context.Background(), &loadbalancer.GetNetworkLoadBalancerRequest{
			NetworkLoadBalancerId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("NetworkLoadBalancer still exists")
		}
	}

	return nil
}

func testAccCheckLBNetworkLoadBalancerExists(n string, nlb *loadbalancer.NetworkLoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.LoadBalancer().NetworkLoadBalancer().Get(context.Background(), &loadbalancer.GetNetworkLoadBalancerRequest{
			NetworkLoadBalancerId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("NetworkLoadBalancer not found")
		}

		*nlb = *found

		return nil
	}
}

func testAccCheckLBNetworkLoadBalancerContainsLabel(nlb *loadbalancer.NetworkLoadBalancer, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := nlb.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccLBNetworkLoadBalancerBasic(name string) string {
	return fmt.Sprintf(`
		resource "yandex_lb_network_load_balancer" "test-nlb" {
		  name			= "%s"
		  description	= "nlb-descr"

		  labels = {
			tf-label    = "tf-label-value"
			empty-label = ""
		  }
		}
		`, name,
	)
}

func testAccLBNetworkLoadBalancerDeletionProtection(name string, deletionProtection bool) string {
	return fmt.Sprintf(`
		resource "yandex_lb_network_load_balancer" "test-nlb" {
		  name					= "%s"
		  description			= "nlb-descr"
		  deletion_protection 	= "%t"
		}
		`, name, deletionProtection,
	)
}

func testAccLBNetworkLoadBalancerDefaults(ctx map[string]interface{}) string {
	return templateConfig(`
		resource "yandex_lb_network_load_balancer" "test-nlb" {
		  name = "{{.NLBName}}"

		  listener {
			name = "{{.ListenerName}}"
			port = {{.ListenerPort}}
			external_address_spec {
			  ip_version = "{{.ListenerIPVersion}}"
			}
		  }

		  attached_target_group {
			target_group_id = "${yandex_lb_target_group.test-target-group.id}"

			healthcheck {
			  name = "{{.HTTPName}}"
			  http_options {
				port = {{.HTTPPort}}
				path = "{{.HTTPPath}}"
			  }
			}
		  }
		}

		resource "yandex_lb_target_group" "test-target-group" {
		  name		= "{{.TGName}}"

		  target {
			subnet_id	= "${yandex_vpc_subnet.test-subnet.id}"
			address		= "${yandex_compute_instance.test-instance-1.network_interface.0.ip_address}"
		  }
		}

		{{.BaseTemplate}}
		`, ctx,
	)
}
