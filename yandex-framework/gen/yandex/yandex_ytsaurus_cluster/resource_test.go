package yandex_ytsaurus_cluster_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ytsaurus/v1"
	ytsaurusv1sdk "github.com/yandex-cloud/go-sdk/services/ytsaurus/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func init() {
	resource.AddTestSweepers("yandex_ytsaurus_cluster", &resource.Sweeper{
		Name:         "yandex_ytsaurus_cluster",
		F:            testSweepCluster,
		Dependencies: []string{},
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepCluster(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := ytsaurusv1sdk.NewClusterClient(conf.SDKv2).List(context.Background(), &ytsaurus.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error getting YTsaurus clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepYtsaurusCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep YTsaurus cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepYtsaurusCluster(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepYtsaurusClusterOnce, conf, "YTsaurus cluster", id)
}

func sweepYtsaurusClusterOnce(conf *provider_config.Config, id string) error {
	op, err := ytsaurusv1sdk.NewClusterClient(conf.SDKv2).Delete(context.Background(), &ytsaurus.DeleteClusterRequest{
		ClusterId: id,
	})
	_, err = op.Wait(context.Background())
	return err
}

func TestAccYtsaurusCluster_full(t *testing.T) {
	var (
		clusterName        = test.ResourceName(63)
		clusterNameUpdated = test.ResourceName(63)

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
		CheckDestroy:             test.AccCheckProjectDestroy,
		Steps: []resource.TestStep{
			ytsaurusClusterBaseTestStep(clusterName, clusterDesc, labelKey, labelValue, 1),
			ytsaurusClusterImportTestStep(),
			ytsaurusClusterBaseTestStep(clusterNameUpdated, clusterDescUpdated, labelKeyUpdated, labelValueUpdated, 2),
			ytsaurusClusterImportTestStep(),
		},
	})
}

func ytsaurusClusterBaseTestStep(clusterName, clusterDesc, labelKey, labelValue string, execNodeCount int64) resource.TestStep {
	return resource.TestStep{
		Config: testYtsaurusClusterFullConfig(clusterName, clusterDesc, labelKey, labelValue, execNodeCount),
		Check: resource.ComposeTestCheckFunc(
			test.YtsaurusClusterExists(test.YtsaurusClusterResourceName),
			resource.TestCheckResourceAttr(test.YtsaurusClusterResourceName, "name", clusterName),
			resource.TestCheckResourceAttr(test.YtsaurusClusterResourceName, "description", clusterDesc),
			resource.TestCheckResourceAttr(test.YtsaurusClusterResourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
			resource.TestCheckResourceAttr(test.YtsaurusClusterResourceName, "spec.compute.0.scale_policy.fixed.size", fmt.Sprint(execNodeCount)),
			resource.TestCheckResourceAttrSet(test.YtsaurusClusterResourceName, "created_at"),
			resource.TestCheckResourceAttrSet(test.YtsaurusClusterResourceName, "created_by"),
			resource.TestCheckResourceAttrSet(test.YtsaurusClusterResourceName, "endpoints.ui"),
			resource.TestCheckResourceAttrSet(test.YtsaurusClusterResourceName, "endpoints.external_http_proxy_balancer"),
			resource.TestCheckResourceAttrSet(test.YtsaurusClusterResourceName, "endpoints.internal_http_proxy_alias"),
			resource.TestCheckResourceAttrSet(test.YtsaurusClusterResourceName, "endpoints.internal_rpc_proxy_alias"),
			resource.TestCheckResourceAttr(test.YtsaurusClusterResourceName, "status", "RUNNING"),
			test.AccCheckCreatedAtAttr(test.YtsaurusClusterResourceName),
		),
	}
}

func ytsaurusClusterImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      test.YtsaurusClusterResourceName,
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
