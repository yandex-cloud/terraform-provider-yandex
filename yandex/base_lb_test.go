package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

const lbDefaultNLBDescription = "nlb-descriprion"
const lbDefaultListenerName = "nlb-listener"
const lbDefaultListenerPort = int64(8080)
const lbDefaultListenerProtocol = "tcp"
const lbDefaultListenerIPVersion = "ipv4"
const lbDefaultListenerTargetPort = int64(8080)
const lbDefaultHCHTTPName = "http"
const lbDefaultHCHTTPInterval = 2
const lbDefaultHCHTTPTimeout = 1
const lbDefaultHCHTTPHealthyTreshold = 2
const lbDefaultHCHTTPUnhealthyTreshold = 2
const lbDefaultHCHTTPPort = 8080
const lbDefaultHCHTTPPath = "/ping"
const lbDefaultRegionID = "ru-central1"
const lbDefaultNLBType = "external"

type lbAttachedTargetGroupChecker func(*loadbalancer.AttachedTargetGroup) error
type lbListenerChecker func(*loadbalancer.Listener) error

func lbDefaultNLBValues() map[string]interface{} {
	return map[string]interface{}{
		"NLBName":               acctest.RandomWithPrefix("tf-nlb"),
		"TGName":                acctest.RandomWithPrefix("tf-tg"),
		"NLBDescr":              lbDefaultNLBDescription,
		"RegionID":              lbDefaultRegionID,
		"NLBType":               lbDefaultNLBType,
		"BaseTemplate":          testAccLBBaseTemplate(acctest.RandomWithPrefix("tf-instance")),
		"ListenerName":          lbDefaultListenerName,
		"ListenerPort":          lbDefaultListenerPort,
		"ListenerTargetPort":    lbDefaultListenerTargetPort,
		"ListenerProtocol":      lbDefaultListenerProtocol,
		"ListenerIPVersion":     lbDefaultListenerIPVersion,
		"HTTPName":              lbDefaultHCHTTPName,
		"HTTPInterval":          lbDefaultHCHTTPInterval,
		"HTTPTimeout":           lbDefaultHCHTTPTimeout,
		"HTTPHealthyTreshold":   lbDefaultHCHTTPHealthyTreshold,
		"HTTPUnhealthyTreshold": lbDefaultHCHTTPUnhealthyTreshold,
		"HTTPPort":              lbDefaultHCHTTPPort,
		"HTTPPath":              lbDefaultHCHTTPPath,
	}
}

func testAccCheckLBTargetGroupValues(tg *loadbalancer.TargetGroup, expectedInstanceNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		subnetIPMap, err := getSubnetIPMap(expectedInstanceNames)
		if err != nil {
			return err
		}

		if tg.GetRegionId() != "ru-central1" {
			return fmt.Errorf("invalid RegionId value in target group %s", tg.Name)
		}

		if len(tg.GetTargets()) != len(expectedInstanceNames) {
			return fmt.Errorf("invalid count of targets in target group %s", tg.Name)
		}

		for _, t := range tg.GetTargets() {
			if addresses, ok := subnetIPMap[t.GetSubnetId()]; ok {
				addressExists := false
				for _, a := range addresses {
					if a == t.GetAddress() {
						addressExists = true
						break
					}
				}
				if !addressExists {
					return fmt.Errorf("invalid Target's Address %s in target group %s", t.GetAddress(), tg.Name)
				}
			} else {
				return fmt.Errorf("invalid Target's SubnetID %s in target group %s", t.GetSubnetId(), tg.Name)
			}
		}

		return nil
	}
}

func getSubnetIPMap(instanceNames []string) (map[string][]string, error) {
	result := make(map[string][]string)
	config := testAccProvider.Meta().(*Config)
	ctx := context.Background()

	for _, instanceName := range instanceNames {
		instanceID, err := resolveObjectIDByNameAndFolderID(ctx, config, instanceName, config.FolderID, sdkresolvers.InstanceResolver)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve data source instance by name: %v", err)
		}
		instance, err := config.sdk.Compute().Instance().Get(ctx, &compute.GetInstanceRequest{
			InstanceId: instanceID,
			View:       compute.InstanceView_FULL,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get instance with ID: %v", instanceID)
		}

		ifs := instance.GetNetworkInterfaces()
		if len(ifs) == 0 {
			return nil, fmt.Errorf("target instance %s doesn't have network interface", instanceName)
		}
		subnetID := ifs[0].GetSubnetId()
		address := ifs[0].GetPrimaryV4Address().GetAddress()
		result[subnetID] = append(result[subnetID], address)
	}

	return result, nil
}

func testAccCheckLBNetworkLoadBalancerValues(nlb *loadbalancer.NetworkLoadBalancer, expectedListeners, expectedAtgs int, lsChecker lbListenerChecker, atgChecker lbAttachedTargetGroupChecker) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if nlb.GetRegionId() != "ru-central1" {
			return fmt.Errorf("invalid RegionId value in network load balancer %s", nlb.Name)
		}

		if nlb.GetType() != loadbalancer.NetworkLoadBalancer_EXTERNAL {
			return fmt.Errorf("invalid Type value in network load balancer %s", nlb.Name)
		}

		if len(nlb.GetListeners()) != expectedListeners {
			return fmt.Errorf("invalid count of listeners in network load balancer %s", nlb.Name)
		}

		if lsChecker != nil {
			for _, ls := range nlb.GetListeners() {
				if err := lsChecker(ls); err != nil {
					return err
				}
			}
		}

		if len(nlb.GetAttachedTargetGroups()) != expectedAtgs {
			return fmt.Errorf("invalid count of attached target groups in network load balancer %s", nlb.Name)
		}

		if atgChecker != nil {
			for _, atg := range nlb.GetAttachedTargetGroups() {
				if err := atgChecker(atg); err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func checkLBListener(ls *loadbalancer.Listener, name string, port, targetPort int64) error {
	if ls.GetName() != name {
		return fmt.Errorf("invalid Name value in network load balancer listener %s", ls.Name)
	}

	if ls.GetPort() != port {
		return fmt.Errorf("invalid Port value in network load balancer listener %s", ls.Name)
	}

	if ls.GetTargetPort() != targetPort {
		return fmt.Errorf("invalid TargetPort value in network load balancer listener %s", ls.Name)
	}

	if ls.GetProtocol() != loadbalancer.Listener_TCP {
		return fmt.Errorf("invalid Protocol value in network load balancer listener %s", ls.Name)
	}

	if ls.GetAddress() == "" {
		return fmt.Errorf("invalid Address value in network load balancer listener %s", ls.Name)
	}

	return nil
}

func checkLBAttachedTargetGroup(atg *loadbalancer.AttachedTargetGroup, hcName string, hcInterval, hcTimeout, hcUnhealthyTreshold, hcHealthyTreshold, hcPort int64, hcPath string) error {
	hcs := atg.GetHealthChecks()

	if atg.GetTargetGroupId() == "" {
		return fmt.Errorf("invalid TargetGroupID value in network load balancer attached target group %s", atg.TargetGroupId)
	}

	if len(hcs) != 1 {
		return fmt.Errorf("invalid healthcheck count in network load balancer attached target group %s", atg.TargetGroupId)
	}

	hc := hcs[0]
	if hc.GetName() != hcName {
		return fmt.Errorf("invalid Name value in network load balancer healthcheck %s", hc.Name)
	}

	if hc.GetInterval().GetSeconds() != hcInterval {
		return fmt.Errorf("invalid Interval value in network load balancer healthcheck %s", hc.Name)
	}

	if hc.GetTimeout().GetSeconds() != hcTimeout {
		return fmt.Errorf("invalid Timeout value in network load balancer healthcheck %s", hc.Name)
	}

	if hc.GetUnhealthyThreshold() != hcUnhealthyTreshold {
		return fmt.Errorf("invalid UnhealthyThreshold value in network load balancer healthcheck %s", hc.Name)
	}

	if hc.GetHealthyThreshold() != hcHealthyTreshold {
		return fmt.Errorf("invalid HealthyThreshold value in network load balancer healthcheck %s", hc.Name)
	}

	if hc.GetHttpOptions() == nil {
		return fmt.Errorf("invalid HttpOptions value in network load balancer healthcheck %s", hc.Name)
	}

	if hc.GetHttpOptions().GetPort() != hcPort {
		return fmt.Errorf("invalid HttpOptions.Port value in network load balancer healthcheck %s", hc.Name)
	}

	if hc.GetHttpOptions().GetPath() != hcPath {
		return fmt.Errorf("invalid HttpOptions.Path value in network load balancer healthcheck %s", hc.Name)
	}

	return nil
}

func testAccLBGeneralNLBTemplate(ctx map[string]interface{}, isDataSource, isListener, isATG, isTG bool) string {
	ctx["IsListener"] = isListener
	ctx["IsATG"] = isATG
	ctx["IsTG"] = isTG
	ctx["IsDataSource"] = isDataSource

	return templateConfig(`
		{{ if .IsDataSource }}
		data "yandex_lb_network_load_balancer" "test-nlb-ds" {
		  name = "${yandex_lb_network_load_balancer.test-nlb.name}"
		}
		{{ end }}

		resource "yandex_lb_network_load_balancer" "test-nlb" {
		  name			= "{{.NLBName}}"
		  description	= "{{.NLBDescr}}"
		  region_id		= "{{.RegionID}}"
		  type			= "{{.NLBType}}"

		  labels = {
			tf-label		= "tf-label-value"
			empty-label		= ""
		  }
		  {{ if .IsListener }}
		  listener {
			name		= "{{.ListenerName}}"
			port		= {{.ListenerPort}}
			target_port = {{.ListenerTargetPort}}
			protocol	= "{{.ListenerProtocol}}"
			external_address_spec {
			  ip_version = "{{.ListenerIPVersion}}"
			}
		  }
		  {{ end }}

		  {{ if .IsATG }}
		  attached_target_group {
			target_group_id = "${yandex_lb_target_group.test-target-group.id}"

			healthcheck {
			  name					= "{{.HTTPName}}"
			  interval				= {{.HTTPInterval}}
			  timeout				= {{.HTTPTimeout}}
			  unhealthy_threshold	= {{.HTTPUnhealthyTreshold}}
			  healthy_threshold		= {{.HTTPHealthyTreshold}}
			  http_options {
				port = {{.HTTPPort}}
				path = "{{.HTTPPath}}"
			  }
			}
		  }
		  {{ end }}
		}

		{{ if .IsTG }}
		resource "yandex_lb_target_group" "test-target-group" {
		  name		= "{{.TGName}}"
		  region_id = "{{.RegionID}}"

		  target {
			subnet_id	= "${yandex_vpc_subnet.test-subnet.id}"
			address		= "${yandex_compute_instance.test-instance-1.network_interface.0.ip_address}"
		  }

		  target {
			subnet_id	= "${yandex_vpc_subnet.test-subnet.id}"
			address		= "${yandex_compute_instance.test-instance-2.network_interface.0.ip_address}"
		  }
		}
		{{ end }}

		{{.BaseTemplate}}
		`,
		ctx,
	)
}

func testAccLBGeneralTGTemplate(tgName, tgDesc, baseTemplate string, targetsCount int, isDataSource bool) string {
	targets := make([]string, targetsCount)
	for i := 1; i <= targetsCount; i++ {
		targets[i-1] = fmt.Sprintf("test-instance-%d", i)
	}
	return templateConfig(`
		{{ if .IsDataSource }}
		data "yandex_lb_target_group" "test-tg-ds" {
		  name = "${yandex_lb_target_group.test-tg.name}"
		}
		{{ end }}

		resource "yandex_lb_target_group" "test-tg" {
		  name		    = "{{.TGName}}"
		  description	= "{{.TGDescription}}"
		  region_id     = "{{.RegionID}}"

		  labels = {
			tf-label    = "tf-label-value"
			empty-label = ""
		  }

		{{range .Targets}}
		  target {
			subnet_id	= "${yandex_vpc_subnet.test-subnet.id}"
			address		= "${yandex_compute_instance.{{.}}.network_interface.0.ip_address}"
		  }
		{{end}}
		}

		{{.BaseTemplate}}
		`,
		map[string]interface{}{
			"TGName":        tgName,
			"TGDescription": tgDesc,
			"RegionID":      lbDefaultRegionID,
			"BaseTemplate":  baseTemplate,
			"Targets":       targets,
			"IsDataSource":  isDataSource,
		},
	)
}

func testAccLBBaseTemplate(instanceName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "test-image" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "test-instance-1" {
  name        = "%[1]s-1"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  resources {
    cores         = 2
    core_fraction = 20
    memory        = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.test-image.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.test-subnet.id}"
  }

  scheduling_policy {
    preemptible = true
  }
}

resource "yandex_compute_instance" "test-instance-2" {
  name        = "%[1]s-2"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  resources {
    cores         = 2
    core_fraction = 20
    memory        = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.test-image.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.test-subnet.id}"
  }

  scheduling_policy {
    preemptible = true
  }
}

resource "yandex_vpc_network" "test-network" {}

resource "yandex_vpc_subnet" "test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instanceName,
	)
}

func TestLBTemplateConfig(t *testing.T) {
	out := templateConfig(
		"{{.key1}}, {{.key2}}, {{.key3}}",
		map[string]interface{}{"key1": "val1", "key2": 2},
		map[string]interface{}{"key3": "val3"},
	)
	assert.Equal(t, out, "val1, 2, val3")
}
