package yandex_ytsaurus_cluster_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccYtsaurusCluster_full(t *testing.T) {
	var (
		clusterName = test.ResourceName(63)

		clusterDesc        = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		clusterDescUpdated = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey           = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelKeyUpdated    = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue         = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
		labelValueUpdated  = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			ytsaurusClusterBaseTestStep(clusterName, clusterDesc, labelKey, labelValue, 1),
			ytsaurusClusterImportTestStep(),
			ytsaurusClusterBaseTestStep(clusterName, clusterDescUpdated, labelKeyUpdated, labelValueUpdated, 2),
			ytsaurusClusterImportTestStep(),
		},
	})
}

func ytsaurusClusterBaseTestStep(clusterName, clusterDesc, labelKey, labelValue string, execNodeCount int64) resource.TestStep {
	return resource.TestStep{
		Config: testYtsaurusClusterFullConfig(clusterName, clusterDesc, labelKey, labelValue, execNodeCount),
		Check: resource.ComposeTestCheckFunc(
			ytsaurusClusterExists(ytsaurusClusterResourceName),
			resource.TestCheckResourceAttr(ytsaurusClusterResourceName, "name", clusterName),
			resource.TestCheckResourceAttr(ytsaurusClusterResourceName, "description", clusterDesc),
			resource.TestCheckResourceAttr(ytsaurusClusterResourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
			resource.TestCheckResourceAttr(ytsaurusClusterResourceName, "spec.compute.0.scale_policy.fixed.size", fmt.Sprint(execNodeCount)),
			resource.TestCheckResourceAttrSet(ytsaurusClusterResourceName, "created_at"),
			resource.TestCheckResourceAttrSet(ytsaurusClusterResourceName, "created_by"),
			resource.TestCheckResourceAttrSet(ytsaurusClusterResourceName, "endpoints.ui"),
			resource.TestCheckResourceAttrSet(ytsaurusClusterResourceName, "endpoints.external_http_proxy_balancer"),
			resource.TestCheckResourceAttrSet(ytsaurusClusterResourceName, "endpoints.internal_http_proxy_alias"),
			resource.TestCheckResourceAttrSet(ytsaurusClusterResourceName, "endpoints.internal_rpc_proxy_alias"),
			resource.TestCheckResourceAttr(ytsaurusClusterResourceName, "status", "RUNNING"),
			test.AccCheckCreatedAtAttr(ytsaurusClusterResourceName),
		),
	}
}

func ytsaurusClusterImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      ytsaurusClusterResourceName,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testYtsaurusClusterFullConfig(clusterName, clusterDesc, labelKey, labelValue string, execNodeCount int64) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "test-network" {}

resource "yandex_vpc_subnet" "test-subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.test-network.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_security_group" "test-security-group" {
  network_id = yandex_vpc_network.test-network.id

  ingress {
    protocol       = "TCP"
    description    = "healthchecks"
    port           = 30080
    v4_cidr_blocks = ["198.18.235.0/24", "198.18.248.0/24"]
  }
}

resource "yandex_ytsaurus_cluster" "test-cluster" {
  name = "%s"
  description = "%s"

  labels = {
	%s = "%s"
  }

  zone_id			 = "ru-central1-a"
  subnet_id			 = yandex_vpc_subnet.test-subnet.id
  security_group_ids = [yandex_vpc_security_group.test-security-group.id]

  spec = {
	storage = {
	  hdd = {
	  	size_gb = 100
		count 	= 3
	  }
	  ssd = {
	  	size_gb = 100
		type 	= "network-ssd"
		count 	= 3
		changelogs = {
		  size_gb = 10
		}
	  }
	}
	compute = [{
	  preset = "c8-m32"
	  disks = [{
	  	type 	= "network-ssd"
		size_gb = 50
	  }]
	  scale_policy = {
	  	fixed = {
		  size = %d
		}
	  }
	}]
	tablet = {
      preset = "c8-m16"
	  count = 3
	}
	proxy = {
	  http = {
	  	count = 1
	  }
	  rpc = {
	  	count = 1
	  }
	}
  }
}
`, clusterName, clusterDesc, labelKey, labelValue, execNodeCount)
}
