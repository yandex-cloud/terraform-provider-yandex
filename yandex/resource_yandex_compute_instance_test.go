package yandex

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func computeInstanceImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:            "yandex_compute_instance.foobar",
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
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceHasInstanceID(&instance, "yandex_compute_instance.foobar"),
					testAccCheckComputeInstanceHasResources(&instance, 1, 100, 2),
					testAccCheckComputeInstanceIsPreemptible(&instance, false),
					testAccCheckComputeInstanceLabel(&instance, "my_key", "my_value"),
					testAccCheckComputeInstanceMetadata(&instance, "foo", "bar"),
					testAccCheckComputeInstanceMetadata(&instance, "baz", "qux"),
					testAccCheckCreatedAtAttr("yandex_compute_instance.foobar"),
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
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceHasResources(&instance, 1, 100, 2),
					testAccCheckComputeInstanceFqdn(&instance, instanceName),
					testAccCheckComputeInstanceMetadata(&instance, "foo", "bar"),
					testAccCheckCreatedAtAttr("yandex_compute_instance.foobar"),
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
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceMetadata(&instance, "foo", "bar"),
					testAccCheckCreatedAtAttr("yandex_compute_instance.foobar"),
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
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceMetadata(&instance, "foo", "bar"),
					testAccCheckCreatedAtAttr("yandex_compute_instance.foobar"),
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
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceMetadata(&instance, "foo", "bar"),
					testAccCheckCreatedAtAttr("yandex_compute_instance.foobar"),
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
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceHasResources(&instance, 2, 5, 0.5),
					testAccCheckCreatedAtAttr("yandex_compute_instance.foobar"),
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
						"yandex_compute_instance.foobar", &instance),
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
						"yandex_compute_instance.foobar", &instance),
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
						"yandex_compute_instance.foobar", &instance),
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
						"yandex_compute_instance.foobar", &instance),
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
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceDisk(&instance, diskName, false, false),
				),
			},
			// check attaching
			{
				Config: testAccComputeInstance_addAttachedDisk(diskName, diskName2, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceDisk(&instance, diskName, false, false),
					testAccCheckComputeInstanceDisk(&instance, diskName2, false, false),
				),
			},
			// check detaching
			{
				Config: testAccComputeInstance_detachDisk(diskName, diskName2, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceDisk(&instance, diskName, false, false),
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
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceBootDisk(&instance, diskName),
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
	var diskTypeID = "network-nvme"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_bootDisk_type(instanceName, diskTypeID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
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
						"yandex_compute_instance.foobar", &instance),
				),
			},
			{
				Config: testAccComputeInstance_forceNewAndChangeMetadata(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
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
						"yandex_compute_instance.foobar", &instance),
				),
			},
			{
				Config: testAccComputeInstance_update(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceMetadata(
						&instance, "bar", "baz"),
					testAccCheckComputeInstanceLabel(&instance, "only_me", "nothing_else"),
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
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceHasResources(&instance, 1, 100, 2),
				),
			},
			computeInstanceImportStep(),
			// Check that instance resources was updated
			{
				Config: testAccComputeInstance_stopInstanceToUpdate2(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceHasPlatformID(&instance, "standard-v1"),
					testAccCheckComputeInstanceHasResources(&instance, 2, 100, 4),
				),
			},
			computeInstanceImportStep(),
			// Check that instance resources was updated
			{
				Config: testAccComputeInstance_stopInstanceToUpdate3(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceHasPlatformID(&instance, "standard-v2"),
					testAccCheckComputeInstanceHasResources(&instance, 2, 5, 0.5),
				),
			},
			computeInstanceImportStep(),
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
						"yandex_compute_instance.foobar", &instance),
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
						"yandex_compute_instance.foobar", &instance),
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
						"yandex_compute_instance.foobar", &instance),
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
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceHasAddress(&instance, address),
				),
			},
		},
	})
}

func TestAccComputeInstance_address_ipv6(t *testing.T) {
	t.Skip("waiting ipv6 support in subnets")
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))
	var addressIpv6 = "fd00:aabb:ccdd:eeff::a"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_address_ipv6(instanceName, addressIpv6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceHasAddressV6(&instance, addressIpv6),
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
					testAccCheckComputeInstanceExists("yandex_compute_instance.foobar", &instance),
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
				Config: testAccComputeInstance_preemptible(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceIsPreemptible(&instance, true),
					testAccCheckCreatedAtAttr("yandex_compute_instance.foobar"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_service_account(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(
						"yandex_compute_instance.foobar", &instance),
					testAccCheckComputeInstanceHasServiceAccount(&instance),
					testAccCheckCreatedAtAttr("yandex_compute_instance.foobar"),
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

func testAccCheckComputeInstanceHasAddressV6(instance *compute.Instance, address string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, i := range instance.NetworkInterfaces {
			if i.PrimaryV6Address.Address != address {
				return fmt.Errorf("Wrong address found: expected %v, got %v", address, i.PrimaryV6Address.Address)
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

//revive:disable:var-naming
func testAccComputeInstance_basic(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  zone        = "ru-central1-a"

  resources {
    cores  = 1
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
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
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
  description = "testAccComputeInstance_basic2"
  zone        = "ru-central1-a"

  resources {
    cores  = 1
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
  zone        = "ru-central1-a"

  resources {
    cores  = 1
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
  zone        = "ru-central1-a"

  resources {
    cores  = 1
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
  zone        = "ru-central1-a"

  resources {
    cores  = 1
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

  resources {
    cores  = 1
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

// Update metadata and network_interface
func testAccComputeInstance_update(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"

  resources {
    cores  = 1
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
    bar            = "baz"
    startup-script = "echo Hello"
  }

  labels = {
    only_me = "nothing_else"
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

func testAccComputeInstance_natIp(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-c"

  resources {
    cores  = 1
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
  zone           = "ru-central1-c"
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

  allow_stopping_for_update = true

  resources {
    cores  = 1
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

  resources {
    cores  = 1
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

  resources {
    cores  = 1
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

  allow_stopping_for_update = true

  resources {
    cores  = 1
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

  allow_stopping_for_update = true

  resources {
    cores  = 1
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

  resources {
    cores  = 1
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

func testAccComputeInstance_bootDisk_type(instance string, diskType string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"

  resources {
    cores  = 1
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

  resources {
    cores  = 1
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

  resources {
    cores  = 1
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

  resources {
    cores  = 1
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

func testAccComputeInstance_address_ipv6(instance, address string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_vpc_network" "inst-test-network" {
  name = "inst-test-network-%s"
}

resource "yandex_vpc_subnet" "inst-test-subnet-v6" {
  name           = "inst-test-subnet-%s"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["10.0.200.0/24"]
  v6_cidr_blocks = ["fd00:aabb:ccdd:eeff::/64"]
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"

  resources {
    cores  = 1
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id    = "${yandex_vpc_subnet.inst-test-subnet-v6.id}"
    ipv6_address = "%s"
  }
}
`, acctest.RandString(10), acctest.RandString(10), instance, address)
}

func testAccComputeInstance_multiNic(instance, network, subnetwork string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-a"

  resources {
    cores  = 1
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

  allow_stopping_for_update = true

  resources {
    cores  = 1
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

// Update fields that require stopping the instance:
func testAccComputeInstance_stopInstanceToUpdate2(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name = "%s"
  zone = "ru-central1-b"

  allow_stopping_for_update = true

  resources {
    cores  = 2
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
    cores         = 2
    core_fraction = 5
    memory        = 0.5
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

func testAccComputeInstance_preemptible(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  zone        = "ru-central1-a"

  resources {
    cores  = 1
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

func testAccComputeInstance_service_account(instance string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name               = "%s"
  description        = "testAccComputeInstance_basic"
  zone               = "ru-central1-a"
  service_account_id = "${yandex_iam_service_account.sa-test.id}"

  resources {
    cores  = 1
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
  name        = "test-sa-for-vm"
  description = "Test SA for VM"
}
`, instance)
}
