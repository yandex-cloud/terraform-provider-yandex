package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

func testAccALBGeneralTGTemplate(tgName, tgDesc, baseTemplate string, targetsCount int) string {
	targets := make([]string, targetsCount)
	for i := 1; i <= targetsCount; i++ {
		targets[i-1] = fmt.Sprintf("test-instance-%d", i)
	}
	return templateConfig(`
		resource "yandex_alb_target_group" "test-tg" {
		  name		    = "{{.TGName}}"
		  description	= "{{.TGDescription}}"

		  labels = {
			tf-label    = "tf-label-value"
			empty-label = ""
		  }

		{{range .Targets}}
		  target {
			subnet_id	= "${yandex_vpc_subnet.test-subnet.id}"
			ip_address		= "${yandex_compute_instance.{{.}}.network_interface.0.ip_address}"
		  }
		{{end}}
		}

		{{.BaseTemplate}}
		`,
		map[string]interface{}{
			"TGName":        tgName,
			"TGDescription": tgDesc,
			"BaseTemplate":  baseTemplate,
			"Targets":       targets,
		},
	)
}

func testAccCheckALBTargetGroupValues(tg *apploadbalancer.TargetGroup, expectedInstanceNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		subnetIPMap, err := getSubnetIPMap(expectedInstanceNames)
		if err != nil {
			return err
		}

		if len(tg.GetTargets()) != len(expectedInstanceNames) {
			return fmt.Errorf("invalid count of targets in target group %s", tg.Name)
		}

		for _, t := range tg.GetTargets() {
			if addresses, ok := subnetIPMap[t.GetSubnetId()]; ok {
				addressExists := false
				for _, a := range addresses {
					if a == t.GetIpAddress() {
						addressExists = true
						break
					}
				}
				if !addressExists {
					return fmt.Errorf("invalid Target's Address %s in target group %s", t.GetIpAddress(), tg.Name)
				}
			} else {
				return fmt.Errorf("invalid Target's SubnetID %s in target group %s", t.GetSubnetId(), tg.Name)
			}
		}

		return nil
	}
}

func testAccALBBaseTemplate(instanceName string) string {
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
