package yandex

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

const (
	instanceResource = "yandex_compute_instance.foobar"
)

func init() {
	resource.AddTestSweepers("yandex_compute_instance", &resource.Sweeper{
		Name: "yandex_compute_instance",
		F:    testSweepComputeInstances,
		Dependencies: []string{
			"yandex_dataproc_cluster",
			"yandex_kubernetes_cluster",
			"yandex_compute_instance_group",
		},
	})
}

func testSweepComputeInstances(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListInstancesRequest{FolderId: conf.FolderID}
	it := conf.sdk.Compute().Instance().InstanceIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepComputeInstance(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Compute Instance %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputeInstance(conf *Config, id string) bool {
	return sweepWithRetry(sweepComputeInstanceOnce, conf, "Compute Instance", id)
}

func sweepComputeInstanceOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexComputeInstanceDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Compute().Instance().Delete(ctx, &compute.DeleteInstanceRequest{
		InstanceId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func computeInstanceImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:            instanceResource,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"allow_stopping_for_update"},
	}
}

func TestAccComputeInstance_basic1(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_basic(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(instanceResource, &instance),
					testAccCheckComputeInstanceHasInstanceID(&instance, instanceResource),
					testAccCheckComputeInstanceHasResources(&instance, 2, 100, 2),
					testAccCheckComputeInstanceIsPreemptible(&instance, false),
					testAccCheckComputeInstanceLabel(&instance, "my_key", "my_value"),
					testAccCheckComputeInstanceMetadata(&instance, "foo", "bar"),
					testAccCheckComputeInstanceMetadata(&instance, "baz", "qux"),
					testAccCheckCreatedAtAttr(instanceResource),
					resource.TestCheckResourceAttr(instanceResource, "metadata_options.#", "1"),
					resource.TestCheckResourceAttr(instanceResource, "metadata_options.0.gce_http_endpoint", "1"),
					resource.TestCheckResourceAttr(instanceResource, "metadata_options.0.aws_v1_http_endpoint", "1"),
					resource.TestCheckResourceAttr(instanceResource, "metadata_options.0.gce_http_token", "1"),
					resource.TestCheckResourceAttr(instanceResource, "metadata_options.0.aws_v1_http_token", "2"),
					resource.TestCheckResourceAttr("yandex_compute_instance.foobar", "hardware_generation.#", "1"),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_Gpus(t *testing.T) {
	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-gpus-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_gpus(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasInstanceID(&instance, instanceResource),
					testAccCheckComputeInstanceHasGpus(&instance, 1),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_basic2(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_basic2(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasResources(&instance, 2, 100, 2),
					testAccCheckComputeInstanceFqdn(&instance, instanceName),
					testAccCheckComputeInstanceMetadata(&instance, "foo", "bar"),
					testAccCheckCreatedAtAttr(instanceResource),
				),
			},
		},
	})
}

func TestAccComputeInstance_basic3(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_basic3(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceMetadata(&instance, "foo", "bar"),
					testAccCheckCreatedAtAttr(instanceResource),
				),
			},
		},
	})
}

func TestAccComputeInstance_basic4(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_basic4(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceMetadata(&instance, "foo", "bar"),
					testAccCheckCreatedAtAttr(instanceResource),
				),
			},
		},
	})
}

func TestAccComputeInstance_basic5(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_basic5(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceMetadata(&instance, "foo", "bar"),
					testAccCheckCreatedAtAttr(instanceResource),
				),
			},
		},
	})
}

func TestAccComputeInstance_basic6(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_basic6(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasResources(&instance, 2, 5, 0.5),
					testAccCheckCreatedAtAttr(instanceResource),
				),
			},
		},
	})
}

func TestAccComputeInstance_SecurityGroups(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_SecurityGroups(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasResources(&instance, 2, 5, 0.5),
					testAccCheckComputeInstanceHasSG(&instance),
					testAccCheckCreatedAtAttr(instanceResource),
				),
			},
		},
	})
}

func TestAccComputeInstance_NatIP(t *testing.T) {
	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_natIp(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasNatAddress(&instance),
				),
			},
		},
	})
}

func TestAccComputeInstance_attachedDisk(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))
	var diskName = fmt.Sprintf("disk-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_attachedDisk(diskName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceDisk(&instance, diskName, false, false),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_attachedDisk_sourceUrl(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))
	var diskName = fmt.Sprintf("disk-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_attachedDisk_sourceUrl(diskName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceDisk(&instance, diskName, false, false),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_attachedDisk_modeRo(t *testing.T) {
	t.Skip("Does not support READ_ONLY mode right now")
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))
	var diskName = fmt.Sprintf("disk-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_attachedDisk_modeRo(diskName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceDisk(&instance, diskName, false, false),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_attachedDiskUpdate(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))
	var diskName = fmt.Sprintf("disk-test-%s", acctest.RandString(10))
	var diskName2 = fmt.Sprintf("disk-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_attachedDisk(diskName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceDisk(&instance, diskName, false, false),
				),
			},
			// check attaching
			{
				Config: testAccComputeInstance_addAttachedDisk(diskName, diskName2, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceDisk(&instance, diskName, false, false),
					testAccCheckComputeInstanceDisk(&instance, diskName2, false, false),
				),
			},
			// check detaching
			{
				Config: testAccComputeInstance_detachDisk(diskName, diskName2, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceDisk(&instance, diskName, false, false),
				),
			},
		},
	})
}

func TestAccComputeInstance_attachedDiskDelete(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))
	var diskName = fmt.Sprintf("disk-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_delAttachedDisk(diskName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceAttachedDisks(&instance, diskName),
				),
			},
			// check attaching
			{
				Config: testAccComputeInstance_delAttachedDisk("", instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceAttachedDisks(&instance),
				),
			},
		},
	})

}

func TestAccComputeInstance_bootDisk_source(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))
	var diskName = fmt.Sprintf("disk-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_bootDisk_source(diskName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceBootDisk(&instance, diskName),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_bootDisk_size(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_bootDisk_size(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_bootDisk_type(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))
	var diskTypeID = "network-ssd"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_bootDisk_type(instanceName, diskTypeID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceBootDiskType(instanceName, diskTypeID),
				),
			},
		},
	})
}

func TestAccComputeInstance_forceNewAndChangeMetadata(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_basic(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
				),
			},
			{
				Config: testAccComputeInstance_forceNewAndChangeMetadata(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceMetadata(
						&instance, "qux", "true"),
				),
			},
		},
	})
}

func TestAccComputeInstance_update(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_basic(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
				),
			},
			{
				Config: testAccComputeInstance_update(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceMetadata(
						&instance, "bar", "baz"),
					testAccCheckComputeInstanceLabel(&instance, "only_me", "nothing_else"),
					testAccCheckComputeInstanceHasNoLabel(&instance, "my_key"),
					testAccCheckComputeInstanceHasNoLabel(&instance, "my_other_key"),
					testAccCheckComputeInstanceHasServiceAccount(&instance),
					resource.TestCheckResourceAttr(instanceResource, "metadata_options.#", "1"),
					resource.TestCheckResourceAttr(instanceResource, "metadata_options.0.gce_http_endpoint", "2"),
					resource.TestCheckResourceAttr(instanceResource, "metadata_options.0.aws_v1_http_endpoint", "2"),
					resource.TestCheckResourceAttr(instanceResource, "metadata_options.0.gce_http_token", "2"),
					resource.TestCheckResourceAttr(instanceResource, "metadata_options.0.aws_v1_http_token", "2"),
				),
			},
			{
				Config: testAccComputeInstance_update_add_dns(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasDnsRecord(&instance),
				),
			},
			{
				Config: testAccComputeInstance_update_add_natIp(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasNatAddress(&instance),
					testAccCheckComputeInstanceHasNoSG(&instance),
				),
			},
			{
				Config: testAccComputeInstance_update_add_SecurityGroups(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasNatAddress(&instance),
					testAccCheckComputeInstanceHasSG(&instance),
				),
			},
			{
				Config: testAccComputeInstance_update_remove_natIp_remove_SGs(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasNoNatAddress(&instance),
					testAccCheckComputeInstanceHasNoSG(&instance),
				),
			},
			{
				Config: testAccComputeInstance_update_to_basic(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
				),
			},
		},
	})
}

func TestAccComputeInstance_stopInstanceToUpdate(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			// Set fields that require stopping the instance to update
			{
				Config: testAccComputeInstance_stopInstanceToUpdate(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasResources(&instance, 2, 100, 2),
				),
			},
			computeInstanceImportStep(),
			// Check that instance resources was updated
			{
				Config: testAccComputeInstance_stopInstanceToUpdate2(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasPlatformID(&instance, "standard-v2"),
					testAccCheckComputeInstanceHasResources(&instance, 4, 100, 4),
				),
			},
			computeInstanceImportStep(),
			// Check that instance resources was updated
			{
				Config: testAccComputeInstance_stopInstanceToUpdate3(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasPlatformID(&instance, "standard-v2"),
					testAccCheckComputeInstanceHasResources(&instance, 4, 5, 1),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_stopInstanceToUpdateResourcesAndPlatform(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			// Set fields that require stopping the instance to update
			{
				Config: testAccComputeInstance_stopInstanceToUpdateResourcesAndPlatform(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasPlatformID(&instance, "standard-v2"),
					testAccCheckComputeInstanceHasResources(&instance, 2, 100, 2),
				),
			},
			computeInstanceImportStep(),
			// Check that instance resources was updated
			{
				Config: testAccComputeInstance_stopInstanceToUpdateResourcesAndPlatform2(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasPlatformID(&instance, "standard-v2"),
					testAccCheckComputeInstanceHasResources(&instance, 2, 50, 1),
				),
			},
		},
	})
}

func TestAccComputeInstance_stopInstanceToUpdateAttachDetachNetworkIfaces(t *testing.T) {
	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_stopInstanceToUpdate_attach_detach_NetworkInterfaces(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
				),
			},
			computeInstanceImportStep(),
			// Check that instance interfaces was updated
			{
				Config: testAccComputeInstance_stopInstanceToUpdate_attach_detach_NetworkInterfaces2(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasNetworkInterfaces(&instance, []string{"0", "2", "3"}),
				),
			},
		},
	})
}

func TestAccComputeInstance_subnet_auto(t *testing.T) {
	t.Skip("waiting implementation of yandex_vpc_network with auto provisioning of subnets")
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_subnet_auto(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasSubnet(&instance),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_subnet_custom(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_subnet_custom(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasSubnet(&instance),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_address_auto(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_address_auto(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasAnyAddress(&instance),
				),
			},
		},
	})
}

func TestAccComputeInstance_address_custom(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))
	var address = "10.0.200.200"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_address_custom(instanceName, address),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasAddress(&instance, address),
				),
			},
		},
	})
}

func TestAccComputeInstance_multiNic(t *testing.T) {
	t.Skip("Currently only one network interface is supported per instance")
	t.Parallel()

	var instance compute.Instance
	instanceName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))
	networkName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))
	subnetworkName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_multiNic(instanceName, networkName, subnetworkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(instanceResource, &instance),
					testAccCheckComputeInstanceHasMultiNic(&instance),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_preemptible(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-preemptible-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_preemptible(instanceName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceIsPreemptible(&instance, true),
					testAccCheckCreatedAtAttr(instanceResource),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_update_scheduling_policy(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-scheduling-policy-update-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_preemptible(instanceName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(instanceResource, &instance),
					testAccCheckComputeInstanceIsPreemptible(&instance, false),
				),
			},
			{
				Config: testAccComputeInstance_preemptible(instanceName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(instanceResource, &instance),
					resource.TestCheckResourceAttrPtr(instanceResource, "id", &instance.Id),
					testAccCheckComputeInstanceIsPreemptible(&instance, true),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_service_account(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-with-sa-%s", acctest.RandString(10))
	var saName = acctest.RandomWithPrefix("test-sa-for-vm")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_service_account(instanceName, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasServiceAccount(&instance),
					testAccCheckCreatedAtAttr(instanceResource),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_network_acceleration_type(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-with-ns-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			// create without setting acceleration type
			{
				Config: testAccComputeInstance_network_acceleration_type_empty(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasNetworkAccelerationType(&instance, compute.NetworkSettings_STANDARD),
				),
			},
			computeInstanceImportStep(),
			// set standard - nothing changes
			{
				Config: testAccComputeInstance_network_acceleration_type(instanceName, "standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasNetworkAccelerationType(&instance, compute.NetworkSettings_STANDARD),
				),
			},
			computeInstanceImportStep(),
			//change to software_accelerated
			{
				Config: testAccComputeInstance_network_acceleration_type(instanceName, "software_accelerated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasNetworkAccelerationType(&instance, compute.NetworkSettings_SOFTWARE_ACCELERATED),
				),
			},
			computeInstanceImportStep(),
			//clear
			{
				Config: testAccComputeInstance_network_acceleration_type_empty(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasNetworkAccelerationType(&instance, compute.NetworkSettings_STANDARD),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_nat_create_specific(t *testing.T) {
	t.Skip("Need address reservation api")
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-with-ns-%s", acctest.RandString(10))

	reservedAddress := "TODO: replace with reservation in config"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			// create with nat, not set address
			{
				Config: testAccComputeInstance_network_nat(instanceName, true, reservedAddress, false, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceNat(&instance, true, reservedAddress, false, ""),
				),
			},
		},
	})
}

func TestAccComputeInstance_nat(t *testing.T) {
	t.Skip("Need address reservation api")
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-with-ns-%s", acctest.RandString(10))

	reservedAddress1 := "TODO: replace with reservation in config"
	reservedAddress2 := "TODO: replace with reservation in config"

	testStep := func(nat1 bool, natAddress1 string, nat2 bool, natAddress2 string) resource.TestStep {
		return resource.TestStep{
			Config: testAccComputeInstance_network_nat(instanceName, nat1, natAddress1, nat2, natAddress2),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckComputeInstanceExists(instanceResource, &instance),
				testAccCheckComputeInstanceNat(&instance, nat1, natAddress1, nat2, natAddress2),
			),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			// create with nat, not set address
			testStep(true, "", false, ""),
			// set nat address for iface 1
			testStep(true, reservedAddress1, false, ""),
			// change nat address for iface 1
			testStep(true, reservedAddress2, false, ""),
			// add nat for iface2, drop specific address for iface1
			testStep(true, "", true, ""),
			// drop all nat
			testStep(false, "", false, ""),
			// add two specific addresses
			testStep(true, reservedAddress2, true, reservedAddress1),
			computeInstanceImportStep(),
		},
	})
}

func TestComputeInstancePlacementPolicyRequest(t *testing.T) {
	rawInstanceID := "test-instance-id"
	rawInstance := map[string]interface{}{
		"name":        "test-instance",
		"description": "test instance",
		"zone":        "ru-central1-d",
		"platform_id": "standard-v2",

		"resources": []interface{}{
			map[string]interface{}{
				"cores":  2,
				"memory": 2,
			},
		},

		"boot_disk": []interface{}{
			map[string]interface{}{
				"disk_id": "test-disk-id",
			},
		},
		"network_interface": []interface{}{
			map[string]interface{}{
				"subnet_id": "test-subnet-id",
			},
		},
	}

	instanceResourceWithPlacement := func(placement []interface{}) *schema.ResourceData {
		rawInstance["placement_policy"] = placement
		return schema.TestResourceDataRaw(t, resourceYandexComputeInstance().Schema, rawInstance)
	}

	cc := []struct {
		name            string
		placementPolicy []interface{}
		expectedPolicy  *compute.PlacementPolicy
		expectedPaths   []string
	}{
		{
			name: "update host affinity rules only",
			placementPolicy: []interface{}{
				map[string]interface{}{
					"host_affinity_rules": []interface{}{
						map[string]interface{}{
							"key": "yc.hostGroupId",
							"op":  "IN",
							"values": []interface{}{
								"test-hostgroup-id",
							},
						},
					},
				},
			},
			expectedPaths: []string{"placement_policy.host_affinity_rules"},
			expectedPolicy: &compute.PlacementPolicy{
				HostAffinityRules: []*compute.PlacementPolicy_HostAffinityRule{
					{
						Key:    "yc.hostGroupId",
						Op:     compute.PlacementPolicy_HostAffinityRule_IN,
						Values: []string{"test-hostgroup-id"},
					},
				},
			},
		},
		{
			name: "update placement group id only",
			placementPolicy: []interface{}{
				map[string]interface{}{
					"placement_group_id": "placement-group-id",
				},
			},
			expectedPaths: []string{"placement_policy.placement_group_id"},
			expectedPolicy: &compute.PlacementPolicy{
				PlacementGroupId: "placement-group-id",
			},
		},
		{
			name: "update placement group id and affinity rules",
			placementPolicy: []interface{}{
				map[string]interface{}{
					"placement_group_id": "placement-group-id",
					"host_affinity_rules": []interface{}{
						map[string]interface{}{
							"key": "yc.hostGroupId",
							"op":  "IN",
							"values": []interface{}{
								"test-hostgroup-id",
							},
						},
					},
				},
			},
			expectedPaths: []string{
				"placement_policy.placement_group_id",
				"placement_policy.host_affinity_rules",
			},
			expectedPolicy: &compute.PlacementPolicy{
				PlacementGroupId: "placement-group-id",
				HostAffinityRules: []*compute.PlacementPolicy_HostAffinityRule{
					{
						Key:    "yc.hostGroupId",
						Op:     compute.PlacementPolicy_HostAffinityRule_IN,
						Values: []string{"test-hostgroup-id"},
					},
				},
			},
		},
	}

	for _, c := range cc {
		t.Run(c.name, func(t *testing.T) {
			resourceData := instanceResourceWithPlacement(c.placementPolicy)
			resourceData.SetId(rawInstanceID)

			policy, paths := preparePlacementPolicyForUpdateRequest(resourceData)
			assert.Equal(t, c.expectedPolicy, policy)
			assert.Equal(t, c.expectedPaths, paths)
		})
	}
}

func TestAccComputeInstance_placement_host_rules(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	var hostID = os.Getenv("COMPUTE_HOST_ID")
	var hostGroupID = os.Getenv("COMPUTE_HOST_GROUP_ID")
	var placementGroupID = os.Getenv("COMPUTE_PLACEMENT_GROUP_ID")
	if hostID == "" || hostGroupID == "" || placementGroupID == "" {
		t.Skip("Required vars COMPUTE_HOST_ID, COMPUTE_PLACEMENT_GROUP_ID and COMPUTE_HOST_GROUP_ID are not set.")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_basic(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
				),
			},
			{
				Config: testAccComputeInstance_with_placement_host(instanceName, placementGroupID, hostID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasAffinityRules(&instance, map[string]string{"yc.hostId": hostID}),
					testAccCheckComputeInstanceHasPlacementGroupID(&instance, placementGroupID),
				),
			},
			{
				Config: testAccComputeInstance_with_placement_hostgroup(instanceName, placementGroupID, hostGroupID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasAffinityRules(&instance, map[string]string{"yc.hostGroupId": hostGroupID}),
					testAccCheckComputeInstanceHasPlacementGroupID(&instance, placementGroupID),
				),
			},
			{
				Config: testAccComputeInstance_with_placement_group(instanceName, placementGroupID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasAffinityRules(&instance, nil),
					testAccCheckComputeInstanceHasPlacementGroupID(&instance, placementGroupID),
				),
			},
			{
				Config: testAccComputeInstance_with_placement_empty(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasAffinityRules(&instance, nil),
					testAccCheckComputeInstanceHasPlacementGroupID(&instance, ""),
				),
			},
			{
				Config: testAccComputeInstance_with_placement_hostgroup(instanceName, placementGroupID, hostGroupID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasAffinityRules(&instance, map[string]string{"yc.hostGroupId": hostGroupID}),
					testAccCheckComputeInstanceHasPlacementGroupID(&instance, placementGroupID),
				),
			},
			{
				Config: testAccComputeInstance_with_placement_empty(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasAffinityRules(&instance, nil),
					testAccCheckComputeInstanceHasPlacementGroupID(&instance, ""),
				),
			},
		},
	})
}

func TestAccComputeInstance_move(t *testing.T) {
	t.Parallel()

	targetFolderID := os.Getenv("COMPUTE_TARGET_FOLDER")
	sourceFolderID := os.Getenv("YC_FOLDER_ID")
	if targetFolderID == "" {
		t.Skip("Required var COMPUTE_TARGET_FOLDER is not set.")
	}

	instanceName := fmt.Sprintf("instance-test-%s", acctest.RandString(10))
	var instance, instanceNew compute.Instance

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_with_folder(instanceName, "", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(instanceResource, &instance),
				),
			},
			{
				Config: testAccComputeInstance_with_folder(instanceName, targetFolderID, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(instanceResource, "folder_id", targetFolderID),
					resource.TestCheckResourceAttrPtr(instanceResource, "id", &instance.Id),
				),
			},
			{
				Config: testAccComputeInstance_with_folder(instanceName, sourceFolderID, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(instanceResource, "folder_id", sourceFolderID),
					testAccCheckComputeInstanceExists(instanceResource, &instanceNew),
					testAccCheckComputeInstancesNotEqual(&instance, &instanceNew),
				),
			},
		},
	})
}

func TestComputeInstanceLocalDisksRequest(t *testing.T) {
	rawInstanceID := "test-instance-id"
	rawInstance := map[string]interface{}{
		"name":        "test-instance",
		"description": "test instance",
		"zone":        "ru-central1-d",
		"platform_id": "standard-v2",

		"resources": []interface{}{
			map[string]interface{}{
				"cores":  2,
				"memory": 2,
			},
		},

		"boot_disk": []interface{}{
			map[string]interface{}{
				"disk_id": "test-disk-id",
			},
		},
		"network_interface": []interface{}{
			map[string]interface{}{
				"subnet_id": "test-subnet-id",
			},
		},
	}

	instanceResourceWithLocalDisks := func(localDiskSpec []interface{}) *schema.ResourceData {
		rawInstance["local_disk"] = localDiskSpec
		return schema.TestResourceDataRaw(t, resourceYandexComputeInstance().Schema, rawInstance)
	}

	cc := []struct {
		name           string
		localDisksSpec []interface{}
		expected       []*compute.AttachedLocalDiskSpec
	}{
		{
			name:           "create instance without local disks",
			localDisksSpec: []interface{}{},
			expected:       nil,
		},
		{
			name: "create instance with local disk",
			localDisksSpec: []interface{}{
				map[string]interface{}{
					"size_bytes": 100,
				},
			},
			expected: []*compute.AttachedLocalDiskSpec{
				{
					Size: 100,
				},
			},
		},
	}

	for _, c := range cc {
		t.Run(c.name, func(t *testing.T) {
			resourceData := instanceResourceWithLocalDisks(c.localDisksSpec)
			resourceData.SetId(rawInstanceID)

			config := Config{FolderID: "folder-id"}
			req, err := prepareCreateInstanceRequest(resourceData, &config)
			assert.NoError(t, err)
			assert.Equal(t, c.expected, req.LocalDiskSpecs)
		})
	}
}

func TestAccComputeInstance_local_disks(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	var diskSize = os.Getenv("COMPUTE_LOCAL_DISK_SIZE")
	if diskSize == "" {
		t.Skip("Required var COMPUTE_LOCAL_DISK_SIZE is not set.")
	}

	diskSizeBytes, err := strconv.Atoi(diskSize)
	if err != nil {
		t.Errorf("parse disk size: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_local_disks(instanceName, diskSizeBytes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						instanceResource, &instance),
					testAccCheckComputeInstanceHasLocalDisk(&instance, int64(diskSizeBytes)),
				),
			},
		},
	})
}

func TestAccComputeInstance_filesystem(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var fs compute.Filesystem
	var newFs compute.Filesystem

	var instanceName = acctest.RandomWithPrefix("tf-test")
	var fsName = acctest.RandomWithPrefix("tf-test")
	var newFsName = acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			// Create instance with a filesystem attached to it
			{
				Config: testAccComputeInstance_filesystem(fsName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeFilesystemExists("yandex_compute_filesystem.foobar", &fs),
					testAccCheckComputeInstanceFilesystem(&instance, []string{fsName}),
				),
			},
			// Attach one more filesystem
			{
				Config: testAccComputeInstance_attachFilesystem(fsName, newFsName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists("yandex_compute_instance.foobar", &instance),
					testAccCheckComputeFilesystemExists("yandex_compute_filesystem.foobar", &fs),
					testAccCheckComputeFilesystemExists("yandex_compute_filesystem.foobar2", &newFs),
					testAccCheckComputeInstanceFilesystem(&instance, []string{fsName, newFsName}),
				),
			},
			// Detach filesystem
			{
				Config: testAccComputeInstance_detachFilesystem(fsName, newFsName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists("yandex_compute_instance.foobar", &instance),
					testAccCheckComputeFilesystemExists("yandex_compute_filesystem.foobar2", &newFs),
					testAccCheckComputeInstanceFilesystem(&instance, []string{newFsName}),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_GpuCluster(t *testing.T) {
	t.Skip("We don't have resources in Prod cluster for this test")
	t.Parallel()

	var instance compute.Instance
	var gpuCluster compute.GpuCluster

	var instanceName = acctest.RandomWithPrefix("tf-test")
	var gpuClusterName = acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			// Create instance within a GPU cluster
			{
				Config: testAccComputeInstance_GpuCluster(gpuClusterName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeGpuClusterExists("yandex_compute_gpu_cluster.foobar", &gpuCluster),
					testAccCheckComputeInstanceGpuCluster(&instance, &gpuCluster.Id),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_Nat(t *testing.T) {
	t.Parallel()

	var instance compute.Instance

	var instanceName = acctest.RandomWithPrefix("tf-test")
	var addressName = acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			// Create instance with a filesystem attached to it
			{
				Config: testAccComputeInstance_nat(addressName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists("yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceHasNatAddress(&instance),
				),
			},
			{
				Config: testAccComputeInstance_remove_nat(addressName, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists("yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceHasNoNatAddress(&instance),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func TestAccComputeInstance_Maintenance(t *testing.T) {
	t.Parallel()

	var instance compute.Instance

	var instanceName = acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_Maintenance(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstance_Maintenance(&instance),
				),
			},
			computeInstanceImportStep(),
		},
	})
}

func testAccCheckComputeInstanceDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_instance" {
			continue
		}

		_, err := config.sdk.Compute().Instance().Get(context.Background(), &compute.GetInstanceRequest{
			InstanceId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Instance still exists")
		}
	}

	return nil
}

func testAccCheckComputeInstanceExists(n string, instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.Compute().Instance().Get(context.Background(), &compute.GetInstanceRequest{
			InstanceId: rs.Primary.ID,
			View:       compute.InstanceView_FULL,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Instance not found")
		}

		*instance = *found

		return nil
	}
}

func testAccCheckComputeInstanceMetadata(
	instance *compute.Instance,
	k string, v string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instance.Metadata == nil {
			return fmt.Errorf("no metadata")
		}

		mv, ok := instance.Metadata[k]
		if !ok {
			return fmt.Errorf("metadata not found for key '%s'", k)
		}

		if v != mv {
			return fmt.Errorf("bad value for %s: %s", k, mv)
		}

		return nil
	}
}

func testAccCheckComputeInstanceFqdn(instance *compute.Instance, hostname string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instance.Fqdn == "" {
			return fmt.Errorf("no fqdn defined for instance")
		}

		re := regexp.MustCompile(hostname)
		if !re.MatchString(instance.Fqdn) {
			return fmt.Errorf("instance fqdn didn't match '%s', got '%s'", hostname, instance.Fqdn)
		}

		return nil
	}
}

func testAccCheckComputeInstanceDisk(instance *compute.Instance, diskName string, delete bool, boot bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		diskResolver := sdkresolvers.DiskResolver(diskName, sdkresolvers.FolderID(config.FolderID))
		if err := config.sdk.Resolve(context.Background(), diskResolver); err != nil {
			return fmt.Errorf("Error while resolve disk name to ID: %s", err)
		}

		sourceDiskID := diskResolver.ID()

		if instance.BootDisk == nil && instance.SecondaryDisks == nil {
			return fmt.Errorf("no disks")
		}

		if boot {
			if instance.BootDisk.DiskId == sourceDiskID && instance.BootDisk.AutoDelete == delete {
				return nil
			}
		} else {
			for _, disk := range instance.SecondaryDisks {
				if disk.DiskId == sourceDiskID && disk.AutoDelete == delete {
					return nil
				}
			}
		}

		return fmt.Errorf("Disk not found: %s", diskName)
	}
}

func testAccCheckComputeInstanceAttachedDisks(instance *compute.Instance, diskNames ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		instanceDiskIDs := make(map[string]struct{})
		for _, disk := range instance.SecondaryDisks {
			instanceDiskIDs[disk.DiskId] = struct{}{}
		}

		for i := 0; i < len(diskNames); i++ {
			diskResolver := sdkresolvers.DiskResolver(diskNames[i], sdkresolvers.FolderID(config.FolderID))
			if err := config.sdk.Resolve(context.Background(), diskResolver); err != nil {
				return fmt.Errorf("Error while resolve disk name to ID: %s", err)
			}

			diskID := diskResolver.ID()
			if _, ok := instanceDiskIDs[diskID]; !ok {
				return fmt.Errorf("Disk %s is expected to be attached", diskID)
			}

			delete(instanceDiskIDs, diskID)
		}

		if len(instanceDiskIDs) > 0 {
			extraDiskIDs := make([]string, 0, len(instanceDiskIDs))
			for extraDiskID := range instanceDiskIDs {
				extraDiskIDs = append(extraDiskIDs, extraDiskID)
			}
			return fmt.Errorf("Instance contains more disks that expected: %s", extraDiskIDs)
		}

		return nil
	}
}

func testAccCheckComputeInstanceHasInstanceID(instance *compute.Instance, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		remote := instance.Id
		local := rs.Primary.ID

		if remote != local {
			return fmt.Errorf("Instance id stored does not match: remote has %#v but local has %#v", remote,
				local)
		}

		return nil
	}
}

func testAccCheckComputeInstanceBootDisk(instance *compute.Instance, source string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instance.BootDisk == nil {
			return fmt.Errorf("no disks")
		}

		config := testAccProvider.Meta().(*Config)

		diskResolver := sdkresolvers.DiskResolver(source, sdkresolvers.FolderID(config.FolderID))
		if err := config.sdk.Resolve(context.Background(), diskResolver); err != nil {
			return fmt.Errorf("Error while resolve disk name to ID: %s", err)
		}

		sourceDiskID := diskResolver.ID()

		if instance.BootDisk.DiskId == sourceDiskID {
			return nil
		}

		return fmt.Errorf("Boot disk not found with source %q", source)
	}
}

func testAccCheckComputeInstanceBootDiskType(instanceName string, diskType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		instanceResolver := sdkresolvers.InstanceResolver(instanceName, sdkresolvers.FolderID(config.FolderID))
		if err := config.sdk.Resolve(context.Background(), instanceResolver); err != nil {
			log.Printf("error while resolve instance: %s", err)
		}

		instance, err := config.sdk.Compute().Instance().Get(context.Background(), &compute.GetInstanceRequest{
			InstanceId: instanceResolver.ID(),
		})
		if err != nil {
			log.Printf("error while get instance: %s", err)
			return err
		}

		disk, err := config.sdk.Compute().Disk().Get(context.Background(), &compute.GetDiskRequest{
			DiskId: instance.BootDisk.DiskId,
		})

		if err != nil {
			return err
		}
		if disk.TypeId == diskType {
			return nil
		}

		return fmt.Errorf("Boot disk not found with type %q", diskType)
	}
}

func testAccCheckComputeInstanceLabel(instance *compute.Instance, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instance.Labels == nil {
			return fmt.Errorf("no labels found on instance %s", instance.Name)
		}

		v, ok := instance.Labels[key]
		if !ok {
			return fmt.Errorf("No label found with key %s on instance %s", key, instance.Name)
		}
		if v != value {
			return fmt.Errorf("Expected value '%s' but found value '%s' for label '%s' on instance %s", value, v, key, instance.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceHasNoLabel(instance *compute.Instance, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instance.Labels == nil {
			return nil
		}

		_, ok := instance.Labels[key]
		if ok {
			return fmt.Errorf("There is label '%s' on instance %s but should not be", key, instance.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceHasNetworkInterfaces(instance *compute.Instance, networkInterfacesIndexes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceNetworkIfaces := instance.GetNetworkInterfaces()
		if len(instanceNetworkIfaces) != len(networkInterfacesIndexes) {
			return fmt.Errorf("Wrong instance network interfaces count: expected %d, got %d", len(networkInterfacesIndexes), len(instanceNetworkIfaces))
		}
		for index, iface := range instanceNetworkIfaces {
			if iface.Index != networkInterfacesIndexes[index] {
				return fmt.Errorf("Wrong instance network interface resource: expected index #%s, got index #%s",
					iface.Index, networkInterfacesIndexes[index])
			}
		}
		return nil
	}
}

func testAccCheckComputeInstanceHasPlatformID(instance *compute.Instance, platformID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instance.PlatformId != platformID {
			return fmt.Errorf("Wrong instance platform_id: expected %s, got %s", platformID, instance.PlatformId)
		}
		return nil
	}
}

func testAccCheckComputeInstanceHasResources(instance *compute.Instance, cores, coreFraction int64, memoryGB float64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resources := instance.GetResources()
		if resources.Cores != cores {
			return fmt.Errorf("Wrong instance Cores resource: expected %d, got %d", cores, resources.Cores)
		}
		if resources.CoreFraction != coreFraction {
			return fmt.Errorf("Wrong instance Cores Fraction resource: expected %d, got %d", coreFraction, resources.CoreFraction)
		}
		memoryBytes := toBytesFromFloat(memoryGB)
		if resources.Memory != memoryBytes {
			return fmt.Errorf("Wrong instance Memory resource: expected %f, got %d", memoryGB, toGigabytes(resources.Memory))
		}
		return nil
	}
}

func testAccCheckComputeInstanceHasSG(instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ni := instance.GetNetworkInterfaces()[0]
		if ni.SecurityGroupIds == nil || len(ni.SecurityGroupIds) == 0 {
			return fmt.Errorf("invalid network_interface.security_group_ids value in instance group %s", instance.Name)
		}
		return nil
	}
}

func testAccCheckComputeInstanceHasNoSG(instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ni := instance.GetNetworkInterfaces()[0]
		if ni.SecurityGroupIds != nil || len(ni.SecurityGroupIds) != 0 {
			return fmt.Errorf("invalid network_interface.security_group_ids value in instance group %s", instance.Name)
		}
		return nil
	}
}

func testAccCheckComputeInstanceHasGpus(instance *compute.Instance, gpus int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resources := instance.GetResources()
		if resources.Gpus != gpus {
			return fmt.Errorf("Wrong instance Gpus resource: expected %d, got %d", gpus, resources.Gpus)
		}
		return nil
	}
}

func testAccCheckComputeInstanceHasSubnet(instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, i := range instance.NetworkInterfaces {
			if i.SubnetId == "" {
				return fmt.Errorf("no subnet")
			}
		}

		return nil
	}
}

func testAccCheckComputeInstanceHasAnyAddress(instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, i := range instance.NetworkInterfaces {
			if i.PrimaryV4Address.Address == "" {
				return fmt.Errorf("no address")
			}
		}

		return nil
	}
}

func testAccCheckComputeInstanceHasAddress(instance *compute.Instance, address string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, i := range instance.NetworkInterfaces {
			if i.PrimaryV4Address.Address != address {
				return fmt.Errorf("Wrong address found: expected %v, got %v", address, i.PrimaryV4Address.Address)
			}
		}

		return nil
	}
}

func testAccCheckComputeInstanceHasNatAddress(instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, i := range instance.NetworkInterfaces {
			if i.PrimaryV4Address.OneToOneNat == nil || i.PrimaryV4Address.OneToOneNat.Address == "" {
				return fmt.Errorf("No NAT address assigned")
			}
		}

		return nil
	}
}

func testAccCheckComputeInstanceHasDnsRecord(instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, i := range instance.NetworkInterfaces {
			if len(i.GetPrimaryV4Address().GetDnsRecords()) == 0 && len(i.GetPrimaryV6Address().GetDnsRecords()) == 0 {
				return fmt.Errorf("No DNS records assigned")
			}
		}

		return nil
	}
}

func testAccCheckComputeInstanceHasNoNatAddress(instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, i := range instance.NetworkInterfaces {
			if i.PrimaryV4Address.OneToOneNat != nil {
				return fmt.Errorf("NAT address assigned")
			}
		}

		return nil
	}
}

func testAccCheckComputeInstancesNotEqual(instanceOld, instanceNew *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instanceOld.Id == instanceNew.Id {
			return fmt.Errorf("Instance was not changed.")
		}
		return nil
	}
}

//nolint:unused
func testAccCheckComputeInstanceHasMultiNic(instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(instance.NetworkInterfaces) < 2 {
			return fmt.Errorf("only saw %d nics", len(instance.NetworkInterfaces))
		}

		return nil
	}
}

func testAccCheckComputeInstanceIsPreemptible(instance *compute.Instance, expect bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instance.SchedulingPolicy.Preemptible != expect {
			return fmt.Errorf("instance preemptible attr wrong: expected %v, got %v", expect, instance.SchedulingPolicy.Preemptible)
		}
		return nil
	}
}

func testAccCheckComputeInstanceHasServiceAccount(instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instance.ServiceAccountId == "" {
			return fmt.Errorf("No Service Account assigned to instance")
		}

		return nil
	}
}

func testAccCheckComputeInstanceHasNetworkAccelerationType(instance *compute.Instance, expected compute.NetworkSettings_Type) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instance.NetworkSettings.Type != expected {
			return fmt.Errorf("Unexpected network acceleration type, actual = %v, expected = %v", instance.NetworkSettings.Type, expected)
		}

		return nil
	}
}

func testAccCheckComputeInstanceNat(instance *compute.Instance, expectedNat1 bool, expectedNatAddress1 string, expectedNat2 bool, expectedNatAddress2 string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(instance.NetworkInterfaces) != 2 {
			return fmt.Errorf("Unexpected count of network interfaces, actual = %d, expected = %d", len(instance.NetworkInterfaces), 1)
		}

		err := testIfaceNat(instance.NetworkInterfaces[0], 0, expectedNat1, expectedNatAddress1)
		if err != nil {
			return err
		}
		return testIfaceNat(instance.NetworkInterfaces[1], 1, expectedNat2, expectedNatAddress2)
	}
}

func testIfaceNat(iface *compute.NetworkInterface, index int, expectedNat bool, expectedNatAddress string) error {
	if iface.PrimaryV4Address.OneToOneNat == nil {
		if expectedNat {
			return fmt.Errorf("Expected nat on the interface %d", index)
		}
		return nil
	}
	if !expectedNat {
		return fmt.Errorf("Unexpected nat on the interface %d", index)
	}
	if expectedNatAddress != "" && expectedNatAddress != iface.PrimaryV4Address.OneToOneNat.Address {
		return fmt.Errorf("Unexpected nat address on the interface %d, expected = %v, actual = %v", index, expectedNatAddress, iface.PrimaryV4Address.OneToOneNat.Address)
	}
	return nil
}

func testAccCheckComputeInstanceHasAffinityRules(instance *compute.Instance, ruleParams map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		placement := instance.PlacementPolicy
		if placement.HostAffinityRules == nil && len(ruleParams) == 0 {
			return nil
		}

		if placement.HostAffinityRules != nil && len(ruleParams) != len(placement.HostAffinityRules) {
			return fmt.Errorf("wrong host affinity rules count")
		}

		for _, rule := range placement.HostAffinityRules {
			if _, ok := ruleParams[rule.Key]; !ok {
				return fmt.Errorf("unexpected rule key: %s", rule.Key)
			}

			if len(rule.Values) != 1 || ruleParams[rule.Key] != rule.Values[0] {
				return fmt.Errorf("unexpected rule value: %s", rule.Values[0])
			}
		}
		return nil
	}
}

func testAccCheckComputeInstanceHasPlacementGroupID(instance *compute.Instance, placementGroupID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		placement := instance.PlacementPolicy
		if placement.PlacementGroupId != placementGroupID {
			return fmt.Errorf("Unexpected placement group id: expected %q, got %q", placementGroupID, placement.PlacementGroupId)
		}
		return nil
	}
}

func testAccCheckComputeInstanceHasLocalDisk(instance *compute.Instance, localDiskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(instance.LocalDisks) != 1 {
			return fmt.Errorf("Unexpected number of local disks: expected 1, got %d", len(instance.LocalDisks))
		}
		if instance.LocalDisks[0].Size != localDiskSize {
			return fmt.Errorf("Unexpected local disk size: expected %d, got %d", localDiskSize, instance.LocalDisks[0].Size)
		}
		return nil
	}
}

func testAccCheckComputeInstanceFilesystem(instance *compute.Instance, fsNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		instanceAttachedFs := make(map[string]struct{})
		for _, fs := range instance.Filesystems {
			instanceAttachedFs[fs.FilesystemId] = struct{}{}
		}

		if len(instanceAttachedFs) != len(fsNames) {
			return fmt.Errorf("Unexpected number of filesystems attached to instance: expected %d, got %d",
				len(instanceAttachedFs), len(fsNames))
		}

		for _, fsName := range fsNames {
			fsResolver := sdkresolvers.FilesystemResolver(fsName, sdkresolvers.FolderID(config.FolderID))
			if err := config.sdk.Resolve(context.Background(), fsResolver); err != nil {
				return fmt.Errorf("Error while resolve filesystem name to ID: %s", err)
			}

			fsID := fsResolver.ID()
			if _, ok := instanceAttachedFs[fsID]; !ok {
				return fmt.Errorf("Filesystem %q is expected to be attached", fsID)
			}
		}
		return nil
	}
}

func testAccCheckComputeInstanceGpuCluster(instance *compute.Instance, expectedID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var gpuClusterID string
		if settings := instance.GpuSettings; settings != nil {
			gpuClusterID = settings.GpuClusterId
		}
		if gpuClusterID != *expectedID {
			return fmt.Errorf("Unexpected GPU cluster ID, actual = %v, expected = %v", gpuClusterID, *expectedID)
		}
		return nil
	}
}

//revive:disable:var-naming
func testAccComputeInstance_basic(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  metadata_options {
	gce_http_endpoint = 1
	aws_v1_http_endpoint = 1
	gce_http_token = 1
	aws_v1_http_token = 2
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_gpus(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  zone        = "ru-central1-b"
  platform_id = "gpu-standard-v1"

  resources {
    cores  = 8
    memory = 96
    gpus   = 1
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_basic2(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  hostname    = "%s"
  platform_id = "standard-v2"
  description = "testAccComputeInstance_basic2"
  zone        = "ru-central1-a"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
 `, instance, instance)
}

func testAccComputeInstance_basic3(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic3"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_basic4(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic4"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_basic5(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic5"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_basic6(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic6"
  zone        = "ru-central1-b"
  platform_id = "standard-v2"

  resources {
    cores         = 2
    core_fraction = 5
    memory        = 0.5
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = data.yandex_compute_image.ubuntu.id
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.inst-test-subnet.id
    dns_record {
      fqdn = "myhost1.internal."
    }
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.inst-test-network.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_SecurityGroups(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic6"
  zone        = "ru-central1-b"
  platform_id = "standard-v2"

  resources {
    cores         = 2
    core_fraction = 5
    memory        = 0.5
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id          = "${yandex_vpc_subnet.inst-test-subnet.id}"
    security_group_ids = ["${yandex_vpc_security_group.sg1.id}"]
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_security_group" "sg1" {
  depends_on  = ["yandex_vpc_network.inst-test-network"]
  name        = "tf-test-sg-1"
  description = "description"
  network_id  = "${yandex_vpc_network.inst-test-network.id}"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }

  ingress {
    description    = "rule1 description"
    protocol       = "TCP"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    port           = 8080
  }

  egress {
    description    = "rule2 description"
    protocol       = "ANY"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    from_port      = 8090
    to_port        = 8099
  }
}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

// Update zone to ForceNew, and change metadata k/v entirely
// Generates diff mismatch
func testAccComputeInstance_forceNewAndChangeMetadata(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    qux = "true"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

// Update metadata, network_interface, service account id, metadata_options
func testAccComputeInstance_update(instance string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%[1]s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  allow_stopping_for_update = true

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id  = "${yandex_vpc_subnet.inst-update-test-subnet.id}"
    ip_address = ""
  }

  metadata = {
    bar            = "baz"
    startup-script = "echo Hello"
  }

  metadata_options {
	gce_http_endpoint = 2
	aws_v1_http_endpoint = 2
	gce_http_token = 2
	aws_v1_http_token = 2
  }

  labels = {
    only_me = "nothing_else"
  }

  service_account_id = "${yandex_iam_service_account.inst-test-sa.id}"
}

resource "yandex_iam_service_account" "inst-test-sa" {
  name        = "%[1]s"
  description = "instance update test service account"
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "inst-update-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["10.0.0.0/24"]
}
`, instance)
}

// Update network_interface
func testAccComputeInstance_update_add_SecurityGroups(instance string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%[1]s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id          = "${yandex_vpc_subnet.inst-update-test-subnet.id}"
    nat                = true
    security_group_ids = ["${yandex_vpc_security_group.sg1.id}"]
  }

  metadata = {
    bar            = "baz"
    startup-script = "echo Hello"
  }

  labels = {
    only_me = "nothing_else"
  }

  service_account_id = "${yandex_iam_service_account.inst-test-sa.id}"
}

resource "yandex_iam_service_account" "inst-test-sa" {
  name        = "%[1]s"
  description = "instance update test service account"
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_security_group" "sg1" {
  depends_on  = ["yandex_vpc_network.inst-test-network"]
  name        = "tf-test-sg-2"
  description = "description"
  network_id  = "${yandex_vpc_network.inst-test-network.id}"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }

  ingress {
    description    = "rule1 description"
    protocol       = "TCP"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    port           = 8080
  }

  egress {
    description    = "rule2 description"
    protocol       = "ANY"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    from_port      = 8090
    to_port        = 8099
  }
}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "inst-update-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["10.0.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_update_add_natIp(instance string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name                      = "%[1]s"
  zone                      = "ru-central1-a"
  platform_id               = "standard-v2"
  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-update-test-subnet.id}"
    nat       = true
  }

  metadata = {
    bar            = "baz"
    startup-script = "echo Hello"
  }

  labels = {
    only_me = "nothing_else"
  }

  service_account_id = "${yandex_iam_service_account.inst-test-sa.id}"
}

resource "yandex_iam_service_account" "inst-test-sa" {
  name        = "%[1]s"
  description = "instance update test service account"
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "inst-update-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["10.0.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_update_add_dns(instance string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name                      = "%[1]s"
  zone                      = "ru-central1-a"
  platform_id               = "standard-v2"
  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu.id
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.inst-update-test-subnet.id
    dns_record {
      fqdn = "%[1]s.fakezone."
    }
  }

  metadata = {
    bar            = "baz"
    startup-script = "echo Hello"
  }

  labels = {
    only_me = "nothing_else"
  }

  service_account_id = yandex_iam_service_account.inst-test-sa.id
}

resource "yandex_iam_service_account" "inst-test-sa" {
  name        = "%[1]s"
  description = "instance update test service account"
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.inst-test-network.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "inst-update-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.inst-test-network.id
  v4_cidr_blocks = ["10.0.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_update_remove_natIp_remove_SGs(instance string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%[1]s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id          = "${yandex_vpc_subnet.inst-update-test-subnet.id}"
    nat                = false
    security_group_ids = []
  }

  metadata = {
    bar            = "baz"
    startup-script = "echo Hello"
  }

  labels = {
    only_me = "nothing_else"
  }

  service_account_id = "${yandex_iam_service_account.inst-test-sa.id}"
}

resource "yandex_iam_service_account" "inst-test-sa" {
  name        = "%[1]s"
  description = "instance update test service account"
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_security_group" "sg1" {
  depends_on  = ["yandex_vpc_network.inst-test-network"]
  name        = "tf-test-sg-2"
  description = "description"
  network_id  = "${yandex_vpc_network.inst-test-network.id}"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }

  ingress {
    description    = "rule1 description"
    protocol       = "TCP"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    port           = 8080
  }

  egress {
    description    = "rule2 description"
    protocol       = "ANY"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    from_port      = 8090
    to_port        = 8099
  }
}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "inst-update-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["10.0.0.0/24"]
}
`, instance)
}

// Attach/detach network_interface
func testAccComputeInstance_stopInstanceToUpdate_attach_detach_NetworkInterfaces(instance string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%[1]s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id          = "${yandex_vpc_subnet.inst-update-test-subnet.id}"
    security_group_ids = ["${yandex_vpc_security_group.sg1.id}"]
  }

  network_interface {
    subnet_id          = "${yandex_vpc_subnet.inst-update-test-subnet.id}"
    security_group_ids = ["${yandex_vpc_security_group.sg1.id}"]
  }

  metadata = {
    bar            = "baz"
    startup-script = "echo Hello"
  }

  labels = {
    only_me = "nothing_else"
  }

  service_account_id = "${yandex_iam_service_account.inst-test-sa.id}"
}

resource "yandex_iam_service_account" "inst-test-sa" {
  name        = "%[1]s"
  description = "instance update test service account"
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_security_group" "sg1" {
  depends_on  = ["yandex_vpc_network.inst-test-network"]
  name        = "tf-test-sg-2"
  description = "description"
  network_id  = "${yandex_vpc_network.inst-test-network.id}"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }

  ingress {
    description    = "rule1 description"
    protocol       = "TCP"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    port           = 8080
  }

  egress {
    description    = "rule2 description"
    protocol       = "ANY"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    from_port      = 8090
    to_port        = 8099
  }
}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "inst-update-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["10.0.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_stopInstanceToUpdate_attach_detach_NetworkInterfaces2(instance string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%[1]s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
	index              = 0
    subnet_id          = "${yandex_vpc_subnet.inst-update-test-subnet.id}"
    security_group_ids = ["${yandex_vpc_security_group.sg1.id}"]
  }

  network_interface {
	index              = 2
    subnet_id          = "${yandex_vpc_subnet.inst-update-test-subnet.id}"
    security_group_ids = ["${yandex_vpc_security_group.sg1.id}"]
  }

  network_interface {
	index              = 3
    subnet_id          = "${yandex_vpc_subnet.inst-update-test-subnet.id}"
    nat                = true
    security_group_ids = ["${yandex_vpc_security_group.sg1.id}"]
  }

  metadata = {
    bar            = "baz"
    startup-script = "echo Hello"
  }

  labels = {
    only_me = "nothing_else"
  }

  service_account_id = "${yandex_iam_service_account.inst-test-sa.id}"
}

resource "yandex_iam_service_account" "inst-test-sa" {
  name        = "%[1]s"
  description = "instance update test service account"
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_security_group" "sg1" {
  depends_on  = ["yandex_vpc_network.inst-test-network"]
  name        = "tf-test-sg-2"
  description = "description"
  network_id  = "${yandex_vpc_network.inst-test-network.id}"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }

  ingress {
    description    = "rule1 description"
    protocol       = "TCP"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    port           = 8080
  }

  egress {
    description    = "rule2 description"
    protocol       = "ANY"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    from_port      = 8090
    to_port        = 8099
  }
}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "inst-update-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["10.0.0.0/24"]
}
`, instance)
}

//revive:disable:var-naming
func testAccComputeInstance_update_to_basic(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-update-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "inst-update-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["10.0.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_natIp(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-d"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
    nat       = true
  }

  metadata = {
    foo = "bar"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-d"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_attachedDisk(disk, instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "%s"
  size     = 10
  zone     = "ru-central1-a"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  secondary_disk {
    disk_id = "${yandex_compute_disk.foobar.id}"
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, disk, instance)
}

func testAccComputeInstance_attachedDisk_sourceUrl(disk, instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "%s"
  size     = 10
  zone     = "ru-central1-a"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  secondary_disk {
    disk_id = "${yandex_compute_disk.foobar.id}"
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, disk, instance)
}

//nolint:unused
func testAccComputeInstance_attachedDisk_modeRo(disk, instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "%s"
  size     = 10
  zone     = "ru-central1-a"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    auto_delete = false

    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  secondary_disk {
    disk_id = "${yandex_compute_disk.foobar.id}"
    mode    = "READ_ONLY"
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, disk, instance)
}

func testAccComputeInstance_addAttachedDisk(disk, disk2, instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "%s"
  size     = 10
  zone     = "ru-central1-a"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
}

resource "yandex_compute_disk" "foobar2" {
  name     = "%s"
  size     = 10
  zone     = "ru-central1-a"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  secondary_disk {
    disk_id = "${yandex_compute_disk.foobar.id}"
  }

  secondary_disk {
    disk_id = "${yandex_compute_disk.foobar2.id}"
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, disk, disk2, instance)
}

func testAccComputeInstance_detachDisk(disk, disk2, instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "%s"
  size     = 10
  zone     = "ru-central1-a"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
}

resource "yandex_compute_disk" "foobar2" {
  name     = "%s"
  size     = 10
  zone     = "ru-central1-a"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  secondary_disk {
    disk_id = "${yandex_compute_disk.foobar.id}"
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, disk, disk2, instance)
}

func testAccComputeInstance_bootDisk_source(disk, instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_disk" "foobar" {
  name     = "%s"
  size     = 10
  zone     = "ru-central1-a"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    disk_id = "${yandex_compute_disk.foobar.id}"
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, disk, instance)
}

func testAccComputeInstance_bootDisk_size(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "centos7" {
  family = "centos-7"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.centos7.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_bootDisk_type(instance string, diskType string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
      type     = "%s"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance, diskType)
}

func testAccComputeInstance_delAttachedDisk(disk, instance string) string {
	var diskSpec, secDiskSpec string
	if disk != "" {
		diskSpec = fmt.Sprintf(`
resource "yandex_compute_disk" "foobar" {
  name     = "%s"
  size     = 10
  zone     = "ru-central1-a"
  image_id = "${data.yandex_compute_image.ubuntu.id}"
}`, disk)
		secDiskSpec = `
  secondary_disk {
    disk_id = "${yandex_compute_disk.foobar.id}"
  }`
	}
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

%s

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

%s

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, diskSpec, instance, secDiskSpec)
}

//nolint:unused
func testAccComputeInstance_subnet_auto(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "u_image" {
  family = "ubuntu-1804-lts"
}

resource "yandex_vpc_network" "inst-test-network" {
  name = "inst-test-network-%s"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  boot_disk {
    initialize_params {
      image_id = "${yandex_compute_image.u_image.id}"
    }
  }

  network_interface {
    network = "${yandex_vpc_network.inst-test-network.name}"
  }
}
`, acctest.RandString(10), instance)
}

func testAccComputeInstance_subnet_custom(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_vpc_network" "inst-test-network" {
  name = "inst-test-network-%s"
}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  name           = "inst-test-subnet-%s"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}
`, acctest.RandString(10), acctest.RandString(10), instance)
}

func testAccComputeInstance_address_auto(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_vpc_network" "inst-test-network" {
  name = "inst-test-network-%s"
}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  name           = "inst-test-subnet-%s"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}
`, acctest.RandString(10), acctest.RandString(10), instance)
}

func testAccComputeInstance_address_custom(instance, address string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_vpc_network" "inst-test-network" {
  name = "inst-test-network-%s"
}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  name           = "inst-test-subnet-%s"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["10.0.200.0/24"]
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id  = "${yandex_vpc_subnet.inst-test-subnet.id}"
    ip_address = "%s"
  }
}
`, acctest.RandString(10), acctest.RandString(10), instance, address)
}

//nolint:unused
func testAccComputeInstance_multiNic(instance, network, subnetwork string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet2.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {
  name = "%s"
}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  name           = "first-%s"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "inst-test-subnet2" {
  name           = "second-%s"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.2.0/24"]
}
`, instance, network, subnetwork, subnetwork)
}

// Set fields that require stopping the instance: 'resources'
func testAccComputeInstance_stopInstanceToUpdate(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-b"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_stopInstanceToUpdateResourcesAndPlatform(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  zone        = "ru-central1-b"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores         = 2
    core_fraction = 100
    memory        = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_stopInstanceToUpdateResourcesAndPlatform2(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  zone        = "ru-central1-b"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores         = 2
    core_fraction = 50
    memory        = 1
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

// Update fields that require stopping the instance:
func testAccComputeInstance_stopInstanceToUpdate2(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-b"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores  = 4
    memory = 4
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

// Update platform_id and resources
func testAccComputeInstance_stopInstanceToUpdate3(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  zone        = "ru-central1-b"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores         = 4
    core_fraction = 5
    memory        = 1
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_preemptible(instance string, preemptible bool) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  zone        = "ru-central1-a"
  platform_id = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  scheduling_policy {
    preemptible = %t
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance, preemptible)
}

func testAccComputeInstance_service_account(instance, sa string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name               = "%s"
  description        = "testAccComputeInstance_basic"
  zone               = "ru-central1-a"
  platform_id        = "standard-v2"
  service_account_id = "${yandex_iam_service_account.sa-test.id}"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  scheduling_policy {
    preemptible = true
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "sa-test" {
  name        = "%s"
  description = "Test SA for VM"
}
`, instance, sa)
}

func testAccComputeInstance_network_acceleration_type(instance string, accelerationType string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name               = "%s"
  description        = "testAccComputeInstance_basic"
  zone               = "ru-central1-a"
  platform_id        = "standard-v2"

  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_acceleration_type = "%s"

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  scheduling_policy {
    preemptible = true
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance, accelerationType)
}

func testAccComputeInstance_network_acceleration_type_empty(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name               = "%s"
  description        = "testAccComputeInstance_basic"
  platform_id        = "standard-v2"
  zone               = "ru-central1-a"

  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  scheduling_policy {
    preemptible = true
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstance_network_nat(instance string, nat1 bool, natAddress1 string, nat2 bool, natAddress2 string) string {
	addressStr1 := ""
	if natAddress1 != "" {
		addressStr1 = "nat_ip_address = \"" + natAddress1 + "\""
	}
	addressStr2 := ""
	if natAddress2 != "" {
		addressStr2 = "nat_ip_address = \"" + natAddress2 + "\""
	}
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name               = "%s"
  description        = "testAccComputeInstance_basic"
  zone               = "ru-central1-d"
  platform_id        = "standard-v2"

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
	nat = %v
    %s
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
	nat = %v
    %s
  }

  scheduling_policy {
    preemptible = true
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-d"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance, nat1, addressStr1, nat2, addressStr2)
}

func testAccComputeInstance_with_placement_host(instance, placementGroupID string, hostID string) string {
	return testAccComputeInstance_with_placement_policy(instance, placementGroupID, "yc.hostId", hostID)
}

func testAccComputeInstance_with_placement_hostgroup(instance, placementGroupID string, hostGroupID string) string {
	return testAccComputeInstance_with_placement_policy(instance, placementGroupID, "yc.hostGroupId", hostGroupID)
}

func testAccComputeInstance_with_placement_group(instance, placementGroupID string) string {
	return testAccComputeInstance_with_placement_policy(instance, placementGroupID)
}

func testAccComputeInstance_with_placement_empty(instance string) string {
	return testAccComputeInstance_with_placement_policy(instance, "")
}

func testAccComputeInstance_with_placement_policy(instance string, placementGroupID string, ruleOpts ...string) string {
	placement := testAccComputeInstance_placement_policy(placementGroupID, ruleOpts...)
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  platform_id = "standard-v2"
  zone        = "ru-central1-d"
  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }

%s
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-d"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance, placement)
}

func testAccComputeInstance_placement_policy(placementGroupID string, ruleOpts ...string) string {
	placementGroup := "placement_group_id = \"\""
	hostAffinity := "host_affinity_rules = []"

	if placementGroupID != "" {
		placementGroup = fmt.Sprintf(`
placement_group_id = "%s"
`, placementGroupID)
	}

	if ruleOpts != nil {
		key := ruleOpts[0]
		value := ruleOpts[1]
		hostAffinity = fmt.Sprintf(`
host_affinity_rules {
	key = "%s"
	op = "IN"
	values = ["%s"]
}
`, key, value)
	}

	return fmt.Sprintf(`
  placement_policy {
	%s
	%s
}
`, placementGroup, hostAffinity)
}

func testAccComputeInstance_with_folder(instance string, folderID string, allowRecreate bool) string {
	var folderAttr string
	if folderID != "" {
		folderAttr = fmt.Sprintf(`  folder_id = "%s"`, folderID)
	}
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_with_folder"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"
  %s

  allow_recreate            = %t
  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    auto_delete = false
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance, folderAttr, allowRecreate)
}

func testAccComputeInstance_local_disks(instance string, diskSize int) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  platform_id = "standard-v2"
  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  local_disk {
    size_bytes = %d
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance, diskSize)
}

func testAccComputeInstance_filesystem(fs, instance string) string {
	return fmt.Sprintf(`
	data "yandex_compute_image" "ubuntu" {
		family = "ubuntu-1804-lts"
	}

	resource "yandex_compute_filesystem" "foobar" {
		name        = "%s"
        description = "simple filesystem"
		size        = 15
        type        = "network-hdd"
	}

	resource "yandex_compute_instance" "foobar" {
		name = "%s"
		zone = "ru-central1-a"
		platform_id = "standard-v2"

		allow_stopping_for_update = true

		resources {
			cores  = 2
			memory = 2
		}

		boot_disk {
			initialize_params {
				image_id = "${data.yandex_compute_image.ubuntu.id}"
			}
		}

		filesystem {
			filesystem_id = "${yandex_compute_filesystem.foobar.id}"
		}

		network_interface {
			subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
		}
	}

	resource "yandex_vpc_network" "inst-test-network" {}

	resource "yandex_vpc_subnet" "inst-test-subnet" {
		zone           = "ru-central1-a"
		network_id     = "${yandex_vpc_network.inst-test-network.id}"
		v4_cidr_blocks = ["192.168.0.0/24"]
}
`, fs, instance)
}

func testAccComputeInstance_GpuCluster(gpuCluster, instance string) string {
	return fmt.Sprintf(`
	data "yandex_compute_image" "ubuntu" {
		family = "ubuntu-2004-lts"
	}

	resource "yandex_compute_gpu_cluster" "foobar" {
		name              = "%s"
        interconnect_type = "infiniband"
		zone              = "ru-central1-a"
	}

	resource "yandex_compute_instance" "foobar" {
		name = "%s"
		zone = "ru-central1-a"
		platform_id = "gpu-standard-v3"

		resources {
			gpus   = 8
			cores  = 224
			memory = 952
		}

		boot_disk {
			initialize_params {
				image_id = data.yandex_compute_image.ubuntu.id
			}
		}

		network_interface {
			subnet_id = yandex_vpc_subnet.inst-test-subnet.id
		}

		gpu_cluster_id = yandex_compute_gpu_cluster.foobar.id
	}

	resource "yandex_vpc_network" "inst-test-network" {}

	resource "yandex_vpc_subnet" "inst-test-subnet" {
		zone           = "ru-central1-a"
		network_id     = yandex_vpc_network.inst-test-network.id
		v4_cidr_blocks = ["192.168.0.0/24"]
}
`, gpuCluster, instance)
}

// testAccComputeInstance_Maintenance creates an instance with maintenance grace_policy and maintenance_grace_period
func testAccComputeInstance_Maintenance(instance string) string {
	return fmt.Sprintf(`
	data "yandex_compute_image" "ubuntu" {
		family = "ubuntu-2004-lts"
	}

	resource "yandex_compute_instance" "foobar" {
		name = "%s"
		zone = "ru-central1-a"

		resources {
			cores  = 2
			memory = 2
		}

		boot_disk {
			initialize_params {
				image_id = data.yandex_compute_image.ubuntu.id
			}
		}

		network_interface {
			subnet_id = yandex_vpc_subnet.inst-test-subnet.id
		}

		maintenance_policy = "migrate"
		maintenance_grace_period = "1s"
	}

	resource "yandex_vpc_network" "inst-test-network" {}

	resource "yandex_vpc_subnet" "inst-test-subnet" {
		zone           = "ru-central1-a"
		network_id     = yandex_vpc_network.inst-test-network.id
		v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccCheckComputeInstance_Maintenance(instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instance.MaintenancePolicy != compute.MaintenancePolicy_MIGRATE {
			return fmt.Errorf("unexpected maintenance_policy: %q", instance.MaintenancePolicy)
		}
		if instance.MaintenanceGracePeriod == nil {
			return fmt.Errorf("unexpected maintenance_grace_period: nil")
		}
		if duration := instance.MaintenanceGracePeriod.AsDuration(); duration != time.Second {
			return fmt.Errorf("unexpected maintenance_grace_period: %s", duration)
		}

		return nil
	}
}

func testAccComputeInstance_attachFilesystem(fs, newFs, instance string) string {
	return fmt.Sprintf(`
	data "yandex_compute_image" "ubuntu" {
		family = "ubuntu-1804-lts"
	}

	resource "yandex_compute_filesystem" "foobar" {
		name        = "%s"
        description = "simple filesystem"
		size        = 15
        type        = "network-hdd"
	}

	resource "yandex_compute_filesystem" "foobar2" {
		name        = "%s"
        description = "simple filesystem"
		size        = 15
        type        = "network-hdd"
	}

	resource "yandex_compute_instance" "foobar" {
		name = "%s"
		zone = "ru-central1-a"
		platform_id = "standard-v2"

		allow_stopping_for_update = true

		resources {
			cores  = 2
			memory = 2
		}

		boot_disk {
			initialize_params {
				image_id = "${data.yandex_compute_image.ubuntu.id}"
			}
		}

		filesystem {
			filesystem_id = "${yandex_compute_filesystem.foobar.id}"
		}

		filesystem {
			filesystem_id = "${yandex_compute_filesystem.foobar2.id}"
		}

		network_interface {
			subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
		}
	}

	resource "yandex_vpc_network" "inst-test-network" {}

	resource "yandex_vpc_subnet" "inst-test-subnet" {
		zone           = "ru-central1-a"
		network_id     = "${yandex_vpc_network.inst-test-network.id}"
		v4_cidr_blocks = ["192.168.0.0/24"]
}
`, fs, newFs, instance)
}

func testAccComputeInstance_detachFilesystem(fs, newFs, instance string) string {
	return fmt.Sprintf(`
	data "yandex_compute_image" "ubuntu" {
		family = "ubuntu-1804-lts"
	}

	resource "yandex_compute_filesystem" "foobar" {
		name        = "%s"
        description = "simple filesystem"
		size        = 15
        type        = "network-hdd"
	}

	resource "yandex_compute_filesystem" "foobar2" {
		name        = "%s"
        description = "simple filesystem"
		size        = 15
        type        = "network-hdd"
	}

	resource "yandex_compute_instance" "foobar" {
		name = "%s"
		zone = "ru-central1-a"
		platform_id = "standard-v2"

		resources {
			cores  = 2
			memory = 2
		}

		boot_disk {
			initialize_params {
				image_id = "${data.yandex_compute_image.ubuntu.id}"
			}
		}

		filesystem {
			filesystem_id = "${yandex_compute_filesystem.foobar2.id}"
		}

		network_interface {
			subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
		}
	}

	resource "yandex_vpc_network" "inst-test-network" {}

	resource "yandex_vpc_subnet" "inst-test-subnet" {
		zone           = "ru-central1-a"
		network_id     = "${yandex_vpc_network.inst-test-network.id}"
		v4_cidr_blocks = ["192.168.0.0/24"]
}
`, fs, newFs, instance)
}

func testAccComputeInstance_nat(addressName, instanceName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
    family = "ubuntu-1804-lts"
}
resource "yandex_vpc_address" "test" {
  name = "%s"
  external_ipv4_address {
    zone_id = "ru-central1-a"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  platform_id = "standard-v1"
  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  resources {
    cores  = 2
    memory = 4
  }

  network_interface {
    subnet_id      = "${yandex_vpc_subnet.inst-test-subnet.id}"
    nat            = true
    nat_ip_address = yandex_vpc_address.test.external_ipv4_address[0].address
  }
}
`, addressName, instanceName)
}

func testAccComputeInstance_remove_nat(addressName, instanceName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
    family = "ubuntu-1804-lts"
}
resource "yandex_vpc_address" "test" {
  name = "%s"
  external_ipv4_address {
    zone_id = "ru-central1-a"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  platform_id = "standard-v1"
  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  resources {
    cores  = 2
    memory = 4
  }

  network_interface {
    subnet_id      = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }
}
`, addressName, instanceName)
}
