package mdb_redis_cluster_v2_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const diskEncryptionKeyResource = `
resource "yandex_kms_symmetric_key" "disk_encrypt" {}
`

func init() {
	resource.AddTestSweepers("yandex_mdb_redis_cluster_v2", &resource.Sweeper{
		Name: "yandex_mdb_redis_cluster_v2",
		F:    testSweepMDBRedisCluster,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

//todo need test for `move`
//todo need test for `restore`

// Test
// 1) Can create cluster without settings
// 2) Can update all settings
// 3) Can change flavor
func TestAccMDBRedisClusterV2_host_changes(t *testing.T) {
	t.Parallel()

	var r redis.Cluster
	redisName := acctest.RandomWithPrefix("tf-redis-1")
	redisDesc := "Redis Cluster Terraform Test #1"
	folderID := test.GetExampleFolderID()
	baseDiskSize := 16
	diskTypeId := "network-ssd"
	baseFlavor := "hm3-c2-m8"
	tlsEnabled := true
	version := "8.1-valkey"
	password := "12345678PP"

	ops := []Op{
		OpCreate,
		OpAddHosts,
		OpAddHosts,
		OpModifyHosts,
		OpDeleteHosts,
		OpDeleteHosts,
	}
	conf := testAccBaseConfig(redisName, redisDesc)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			//1 Create Redis Cluster
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Config: &config{
						Version:  &version,
						Password: &password,
					},
					TlsEnabled: &tlsEnabled,
					Hosts: map[string]host{
						"hst_0": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_1": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_2": {Zone: &defaultZone, SubnetId: &defaultSubnet},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 3, tlsEnabled, false, false, "ON"),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_0.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.replica_priority", "100"),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_1.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_1.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_1.replica_priority", "100"),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_2.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_2.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_2.replica_priority", "100"),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					testAccCheckMDBRedisOperations(redisResource, ops[:1]),
				),
			},
			//2
			mdbRedisClusterImportStep(redisResource),
			//3 Add New Host With Empty Subnet
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Hosts: map[string]host{
						"hst_0": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_1": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_2": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_3": {Zone: &defaultZone},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 4, tlsEnabled, false, false, "ON"),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_3.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_3.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_3.replica_priority", "100"),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					testAccCheckMDBRedisOperations(redisResource, ops[:2]),
				),
			},
			//4
			mdbRedisClusterImportStep(redisResource),
			//5 Impossible Change
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Hosts: map[string]host{
						"hst_0": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_1": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_2": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_3": {Zone: newPtr("ru-central1-b")},
					},
				}),
				ExpectError: regexp.MustCompile(".*Attributes shard_name, zone, subnet_id can't be changed.*"),
			},
			//6
			mdbRedisClusterImportStep(redisResource),
			//7 Rename Label - must be without changes
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Hosts: map[string]host{
						"hst_0": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_1": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_2": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_5": {Zone: &defaultZone},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 4, tlsEnabled, false, false, "ON"),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					resource.TestCheckNoResourceAttr(redisResource, "hosts.hst_3"),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_5.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_5.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_5.replica_priority", "100"),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					testAccCheckMDBRedisOperations(redisResource, ops[:2]),
				),
			},
			//8
			mdbRedisClusterImportStep(redisResource),
			//9 Two Remove One Update One Create
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Hosts: map[string]host{
						"hst_q1": {Zone: &defaultZone, SubnetId: &defaultSubnet, ReplicaPriority: newPtr(101), AssignPublicIp: newPtr(true)}, // this host will be updated
						"hst_q2": {Zone: &defaultZone, SubnetId: &defaultSubnet},                                                             // this host only change label
						"hst_q5": {Zone: newPtr("ru-central1-b"), SubnetId: &secondSubnet},                                                   // this host wil be recreated
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 3, tlsEnabled, false, false, "ON"),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					resource.TestCheckNoResourceAttr(redisResource, "hosts.hst_0"),
					resource.TestCheckNoResourceAttr(redisResource, "hosts.hst_1"),
					resource.TestCheckNoResourceAttr(redisResource, "hosts.hst_2"),
					resource.TestCheckNoResourceAttr(redisResource, "hosts.hst_5"),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_q1.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_q1.assign_public_ip", "true"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_q1.replica_priority", "101"),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_q2.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_q2.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_q2.replica_priority", "100"),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_q5.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_q5.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_q5.replica_priority", "100"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_q5.zone", "ru-central1-b"),
					testAccCheckMDBRedisOperations(redisResource, ops),
				),
			},
			//10
			mdbRedisClusterImportStep(redisResource),
		},
	})
}

// Test
// 1) Can create cluster without settings
// 2) Can update all settings
// 3) Can change flavor
func TestAccMDBRedisClusterV2_create_without_settings(t *testing.T) {
	t.Parallel()

	var r redis.Cluster
	redisName := acctest.RandomWithPrefix("tf-redis-2")
	redisDesc := "Redis Cluster Terraform Test #2"
	redisDesc2 := "Redis Cluster Terraform Test #2 Updated"
	folderID := test.GetExampleFolderID()
	baseDiskSize := 16
	updatedDiskSize := 24
	diskTypeId := "network-ssd"
	baseFlavor := "hm3-c2-m8"
	updatedFlavor := "hm3-c2-m12"
	tlsEnabled := true
	normalLimits := "16777215 8388607 61"
	pubsubLimits := "16777214 8388606 62"
	version := "8.1-valkey"
	password := "12345678PP_OLD"

	nonShardedHosts := map[string]host{
		"hst_0": {Zone: &defaultZone, SubnetId: &defaultSubnet},
	}
	ops := []Op{
		OpCreate,
		OpModify,
		OpModify,
		OpModify,
		OpModify,
	}
	conf := testAccBaseConfig(redisName, redisDesc)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			//1 Create Redis Cluster1
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Config: &config{
						Version:  &version,
						Password: &password,
					},
					TlsEnabled:         &tlsEnabled,
					Hosts:              nonShardedHosts,
					DeletionProtection: newPtr(true),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1, tlsEnabled, false, false, "ON"),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_0.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.replica_priority", "100"),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					resource.TestCheckResourceAttr(redisResource, "deletion_protection", "true"),
					testAccCheckMDBRedisOperations(redisResource, ops[:1]),

					resource.TestCheckNoResourceAttr(redisResource, "security_group_ids"),
					resource.TestCheckResourceAttr(redisResource, "labels.%", "0"),
				),
			},
			//2
			mdbRedisClusterImportStep(redisResource),
			//3
			{
				Config: makeConfig(t, conf, testAccAllSettingsConfig(redisName, redisDesc2, version, baseDiskSize, diskTypeId, baseFlavor, nonShardedHosts)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1, tlsEnabled, true, true, "OFF"),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc2),
					resource.TestCheckResourceAttr(redisResource, "sharded", "false"),
					resource.TestCheckResourceAttr(redisResource, "environment", "PRESTABLE"),
					resource.TestCheckResourceAttr(redisResource, "tls_enabled", "true"),
					resource.TestCheckResourceAttr(redisResource, "announce_hostnames", "true"),
					resource.TestCheckResourceAttr(redisResource, "deletion_protection", "true"),
					resource.TestCheckResourceAttr(redisResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(redisResource, "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr(redisResource, "maintenance_window.day", "MON"),
					resource.TestCheckResourceAttr(redisResource, "maintenance_window.hour", "1"),
					resource.TestCheckResourceAttr(redisResource, "access.web_sql", "true"),
					resource.TestCheckResourceAttr(redisResource, "access.data_lens", "true"),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_0.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.replica_priority", "100"),
					testAccCheckMDBRedisClusterHasConfig(&r, "ALLKEYS_LRU", 100,
						"Elg", 5000, 19, 18, version,
						normalLimits, pubsubLimits, 70, 4444, 15, true, true, true,
						14, 13, true, true, true, true, 256),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					resource.TestCheckResourceAttr(redisResource, "labels.%", "2"),
					testAccCheckMDBRedisClusterContainsLabel(&r, "foo", "bar"),
					testAccCheckMDBRedisClusterContainsLabel(&r, "foo2", "bar2"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_retain_period_days", "12"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_window_start.hours", "10"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_window_start.minutes", "11"),
					resource.TestCheckResourceAttr(redisResource, "disk_size_autoscaling.disk_size_limit", fmt.Sprintf("%d", baseDiskSize*2)),
					resource.TestCheckResourceAttr(redisResource, "disk_size_autoscaling.emergency_usage_threshold", "83"),
					testAccCheckMDBRedisOperations(redisResource, ops[:3]),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_search.enabled", "false"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_json.enabled", "true"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_bloom.enabled", "false"),
				),
			},
			//4
			mdbRedisClusterImportStep(redisResource),
			//5
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					TlsEnabled: newPtr(false), // settings which require recreate cluster if changed
				}),
				ExpectError: regexp.MustCompile(`.*The operation was rejected because cluster has 'deletion_protection' =\s+ON.*`),
			},
			//6
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					TlsEnabled:         &tlsEnabled,
					DeletionProtection: newPtr(false),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1, tlsEnabled, true, true, "OFF"),
					resource.TestCheckResourceAttr(redisResource, "deletion_protection", "false"),
					testAccCheckMDBRedisOperations(redisResource, ops[:4]),
				),
			},
			//7
			mdbRedisClusterImportStep(redisResource),
			//8
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Resources: &hostResource{
						ResourcePresetId: &updatedFlavor,
						DiskSize:         &updatedDiskSize,
					},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(

					testAccCheckMDBRedisClusterExists(redisResource, &r, 1, tlsEnabled, true, true, "OFF"),
					testAccCheckMDBRedisClusterHasResources(&r, updatedFlavor, updatedDiskSize, diskTypeId),
					testAccCheckMDBRedisOperations(redisResource, ops),
				),
			},
			//9
			mdbRedisClusterImportStep(redisResource),
		},
	})
}

// Test
// 1) Can create cluster with all settings
// 2) Can update all settings
func TestAccMDBRedisClusterV2_create_with_settings(t *testing.T) {
	t.Parallel()

	var r redis.Cluster
	redisName := acctest.RandomWithPrefix("tf-redis-3")
	redisNameUp := acctest.RandomWithPrefix("tf-redis-3-updated")
	redisDesc := "Redis Cluster Terraform Test #3"
	redisDescUp := "Redis Cluster Terraform Test #3 Updated"
	folderID := test.GetExampleFolderID()
	baseDiskSize := 368
	diskTypeId := "local-ssd"
	baseFlavor := "hm3-c2-m8"
	tlsEnabled := true
	normalLimits := "16777215 8388607 61"
	pubsubLimits := "16777214 8388606 62"
	normalUpdatedLimits := "16777212 8388605 63"
	pubsubUpdatedLimits := "33554432 16777216 60"
	pubIpSet := true
	pubIpUnset := false
	baseReplicaPriority := 100
	updatedReplicaPriority := 51
	version := "8.1-valkey"
	ops := []Op{
		OpCreate,
		OpModify,
		OpModify,
		OpModify,
	}

	nonShardedHosts := map[string]host{
		"hst_0": {Zone: &defaultZone, SubnetId: &defaultSubnet, ReplicaPriority: &baseReplicaPriority, AssignPublicIp: &pubIpUnset},
		"hst_1": {Zone: &defaultZone, SubnetId: &defaultSubnet, ReplicaPriority: &updatedReplicaPriority, AssignPublicIp: &pubIpSet},
		"hst_2": {Zone: &defaultZone, SubnetId: &defaultSubnet, ReplicaPriority: &baseReplicaPriority, AssignPublicIp: &pubIpUnset},
	}
	confg := testAccAllSettingsConfig(redisName, redisDesc, version, baseDiskSize, diskTypeId, baseFlavor, nonShardedHosts)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			// Create Redis Cluster
			{
				Config: makeConfig(t, confg, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 3, tlsEnabled, true, true, "OFF"),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					resource.TestCheckResourceAttr(redisResource, "sharded", "false"),
					resource.TestCheckResourceAttr(redisResource, "environment", "PRESTABLE"),
					resource.TestCheckResourceAttr(redisResource, "tls_enabled", "true"),
					resource.TestCheckResourceAttr(redisResource, "announce_hostnames", "true"),
					resource.TestCheckResourceAttr(redisResource, "auth_sentinel", "true"),
					resource.TestCheckResourceAttr(redisResource, "deletion_protection", "true"),
					resource.TestCheckResourceAttr(redisResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(redisResource, "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr(redisResource, "maintenance_window.day", "MON"),
					resource.TestCheckResourceAttr(redisResource, "maintenance_window.hour", "1"),
					resource.TestCheckResourceAttr(redisResource, "access.web_sql", "true"),
					resource.TestCheckResourceAttr(redisResource, "access.data_lens", "true"),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_0.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.assign_public_ip", fmt.Sprintf("%t", pubIpUnset)),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.replica_priority", fmt.Sprintf("%d", baseReplicaPriority)),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_1.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_1.assign_public_ip", fmt.Sprintf("%t", pubIpSet)),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_1.replica_priority", fmt.Sprintf("%d", updatedReplicaPriority)),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_2.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_2.assign_public_ip", fmt.Sprintf("%t", pubIpUnset)),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_2.replica_priority", fmt.Sprintf("%d", baseReplicaPriority)),
					testAccCheckMDBRedisClusterHasConfig(&r, "ALLKEYS_LRU", 100,
						"Elg", 5000, 19, 18, version,
						normalLimits, pubsubLimits, 70, 4444, 15, true, true, true,
						14, 13, true, true, true, true, 256),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					resource.TestCheckResourceAttr(redisResource, "labels.%", "2"),
					testAccCheckMDBRedisClusterContainsLabel(&r, "foo", "bar"),
					testAccCheckMDBRedisClusterContainsLabel(&r, "foo2", "bar2"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_retain_period_days", "12"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_window_start.hours", "10"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_window_start.minutes", "11"),
					resource.TestCheckResourceAttr(redisResource, "disk_size_autoscaling.disk_size_limit", fmt.Sprintf("%d", baseDiskSize*2)),
					resource.TestCheckResourceAttr(redisResource, "disk_size_autoscaling.emergency_usage_threshold", "83"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_search.enabled", "false"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_json.enabled", "true"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_bloom.enabled", "false"),
					testAccCheckMDBRedisOperations(redisResource, ops[:1]),
				),
			},
			mdbRedisClusterImportStep(redisResource),
			// Update All Settings Cluster
			{
				Config: makeConfig(t, confg, testAccAllSettingsConfigChanged(redisNameUp, redisDescUp, version, baseDiskSize, diskTypeId, baseFlavor, nonShardedHosts)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 3, tlsEnabled, false, false, "ON"),
					resource.TestCheckResourceAttr(redisResource, "name", redisNameUp),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDescUp),
					resource.TestCheckResourceAttr(redisResource, "sharded", "false"),
					resource.TestCheckResourceAttr(redisResource, "environment", "PRESTABLE"),
					resource.TestCheckResourceAttr(redisResource, "tls_enabled", "true"),
					resource.TestCheckResourceAttr(redisResource, "announce_hostnames", "false"),
					resource.TestCheckResourceAttr(redisResource, "auth_sentinel", "false"),
					resource.TestCheckResourceAttr(redisResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(redisResource, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(redisResource, "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr(redisResource, "maintenance_window.day", "FRI"),
					resource.TestCheckResourceAttr(redisResource, "maintenance_window.hour", "2"),
					resource.TestCheckResourceAttr(redisResource, "access.web_sql", "false"),
					resource.TestCheckResourceAttr(redisResource, "access.data_lens", "false"),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_0.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.assign_public_ip", fmt.Sprintf("%t", pubIpUnset)),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.replica_priority", fmt.Sprintf("%d", baseReplicaPriority)),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_1.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_1.assign_public_ip", fmt.Sprintf("%t", pubIpSet)),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_1.replica_priority", fmt.Sprintf("%d", updatedReplicaPriority)),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_2.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_2.assign_public_ip", fmt.Sprintf("%t", pubIpUnset)),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_2.replica_priority", fmt.Sprintf("%d", baseReplicaPriority)),
					testAccCheckMDBRedisClusterHasConfig(&r, "VOLATILE_LFU", 101,
						"Ex", 5001, 20, 21, version,
						normalUpdatedLimits, pubsubUpdatedLimits, 71, 4440, 16, false, false, false,
						22, 23, false, false, false, false, 128),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					resource.TestCheckResourceAttr(redisResource, "labels.%", "2"),
					testAccCheckMDBRedisClusterContainsLabel(&r, "qwe", "rty"),
					testAccCheckMDBRedisClusterContainsLabel(&r, "foo2", "bar2"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_retain_period_days", "31"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_window_start.hours", "20"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_window_start.minutes", "15"),
					resource.TestCheckResourceAttr(redisResource, "disk_size_autoscaling.disk_size_limit", fmt.Sprintf("%d", baseDiskSize*3)),
					resource.TestCheckResourceAttr(redisResource, "disk_size_autoscaling.emergency_usage_threshold", "84"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_search.enabled", "true"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_search.reader_threads", "3"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_search.writer_threads", "3"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_json.enabled", "true"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_bloom.enabled", "true"),
					testAccCheckMDBRedisOperations(redisResource, ops[:3]),
				),
			},
			mdbRedisClusterImportStep(redisResource),
			// Update to empty (nothing changes)
			{
				Config: makeConfig(t, testAccBaseConfig(redisNameUp, redisDescUp), &redisConfigTest{
					Config: &config{
						Version:  &version,
						Password: newPtr("12345678PQ"),
					},
					Hosts: nonShardedHosts,

					Resources: &hostResource{
						ResourcePresetId: newPtr(baseFlavor),
						DiskSize:         newPtr(baseDiskSize),
						DiskTypeId:       newPtr(diskTypeId),
					},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 3, tlsEnabled, false, false, "ON"),
					resource.TestCheckResourceAttr(redisResource, "name", redisNameUp),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDescUp),
					resource.TestCheckResourceAttr(redisResource, "sharded", "false"),
					resource.TestCheckResourceAttr(redisResource, "environment", "PRESTABLE"),
					resource.TestCheckResourceAttr(redisResource, "tls_enabled", "true"),
					resource.TestCheckResourceAttr(redisResource, "announce_hostnames", "false"),
					resource.TestCheckResourceAttr(redisResource, "deletion_protection", "false"),
					resource.TestCheckNoResourceAttr(redisResource, "security_group_ids"),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_0.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.assign_public_ip", fmt.Sprintf("%t", pubIpUnset)),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_0.replica_priority", fmt.Sprintf("%d", baseReplicaPriority)),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_1.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_1.assign_public_ip", fmt.Sprintf("%t", pubIpSet)),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_1.replica_priority", fmt.Sprintf("%d", updatedReplicaPriority)),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_2.fqdn"),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_2.assign_public_ip", fmt.Sprintf("%t", pubIpUnset)),
					resource.TestCheckResourceAttr(redisResource, "hosts.hst_2.replica_priority", fmt.Sprintf("%d", baseReplicaPriority)),
					testAccCheckMDBRedisClusterHasConfig(&r, "VOLATILE_LFU", 101,
						"Ex", 5001, 20, 21, version,
						normalUpdatedLimits, pubsubUpdatedLimits, 71, 4440, 16, false, false, false,
						22, 23, false, false, false, false, 128),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					resource.TestCheckResourceAttr(redisResource, "labels.%", "0"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_retain_period_days", "31"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_window_start.hours", "20"),
					resource.TestCheckResourceAttr(redisResource, "config.backup_window_start.minutes", "15"),
					testAccCheckMDBRedisOperations(redisResource, ops),
				),
			},
			mdbRedisClusterImportStep(redisResource),
			// Decrease disk size (nothing changes)
			{
				Config: makeConfig(t, testAccBaseConfig(redisNameUp, redisDescUp), &redisConfigTest{
					Config: &config{
						Version:  &version,
						Password: newPtr("12345678PQ"),
					},
					Hosts: nonShardedHosts,

					Resources: &hostResource{
						ResourcePresetId: newPtr(baseFlavor),
						DiskSize:         newPtr(baseDiskSize - 3),
						DiskTypeId:       newPtr(diskTypeId),
					},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 3, tlsEnabled, false, false, "ON"),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					resource.TestCheckResourceAttr(redisResource, "resources.disk_size", strconv.Itoa(baseDiskSize)),
					testAccCheckMDBRedisOperations(redisResource, ops),
				),
			},
			mdbRedisClusterImportStep(redisResource),
		},
	})

}

// Test
// 1) Can create cluster with explicitly disabled modules
// 2) Can update module settings
func TestAccMDBRedisClusterV2_create_with_disabled_modules(t *testing.T) {
	t.Parallel()

	var r redis.Cluster
	redisName := acctest.RandomWithPrefix("tf-valkey-modules")
	redisDesc := "Valkey Cluster Terraform Test valkey modules"
	folderID := test.GetExampleFolderID()
	baseDiskSize := 16
	diskTypeId := "network-ssd"
	baseFlavor := "hm3-c2-m8"
	tlsEnabled := true
	version := "9.0-valkey"
	password := "12345678PP"

	nonShardedHosts := map[string]host{
		"hst_0": {Zone: &defaultZone, SubnetId: &defaultSubnet},
	}
	ops := []Op{
		OpCreate,
		OpModify,
	}
	conf := testAccModulesDisabledConfig(redisName, redisDesc)
	modulesDisabledConfig := makeConfig(t, conf, &redisConfigTest{
		Config: &config{
			Version:  &version,
			Password: &password,
		},
		TlsEnabled: &tlsEnabled,
		Hosts:      nonShardedHosts,
	})
	modulesExplicitlyDisabledConfig := makeConfig(t, conf, &redisConfigTest{
		Config: &config{
			Version:  &version,
			Password: &password,
		},
		Modules: &valkeyModules{
			ValkeySearch: &valkeySearch{
				Enabled: newPtr(false),
			},
			ValkeyJson: &valkeyJson{
				Enabled: newPtr(false),
			},
			ValkeyBloom: &valkeyBloom{
				Enabled: newPtr(false),
			},
		},
		TlsEnabled: &tlsEnabled,
		Hosts:      nonShardedHosts,
	})
	modulesEnabledConfig := makeConfig(t, conf, &redisConfigTest{
		Config: &config{
			Version:  &version,
			Password: &password,
		},
		Modules: &valkeyModules{
			ValkeySearch: &valkeySearch{
				Enabled:       newPtr(true),
				ReaderThreads: newPtr(4),
				WriterThreads: newPtr(4),
			},
			ValkeyJson: &valkeyJson{
				Enabled: newPtr(true),
			},
			ValkeyBloom: &valkeyBloom{
				Enabled: newPtr(true),
			},
		},
		TlsEnabled: &tlsEnabled,
		Hosts:      nonShardedHosts,
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			//1
			{
				Config: modulesDisabledConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1, tlsEnabled, false, false, "ON"),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_0.fqdn"),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					testAccCheckMDBRedisOperations(redisResource, ops[:1]),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_search.enabled", "false"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_json.enabled", "false"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_bloom.enabled", "false"),
				),
			},
			//2
			mdbRedisClusterImportStep(redisResource),
			//3
			{
				Config: modulesExplicitlyDisabledConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1, tlsEnabled, false, false, "ON"),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_0.fqdn"),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					testAccCheckMDBRedisOperations(redisResource, ops[:1]),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_search.enabled", "false"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_json.enabled", "false"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_bloom.enabled", "false"),
				),
			},
			//4
			mdbRedisClusterImportStep(redisResource),
			//5
			{
				Config: modulesEnabledConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1, tlsEnabled, false, false, "ON"),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					resource.TestCheckResourceAttrSet(redisResource, "hosts.hst_0.fqdn"),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					testAccCheckMDBRedisOperations(redisResource, ops[:2]),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_search.enabled", "true"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_search.reader_threads", "4"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_search.writer_threads", "4"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_json.enabled", "true"),
					resource.TestCheckResourceAttr(redisResource, "modules.valkey_bloom.enabled", "true"),
				),
			},
			//6
			mdbRedisClusterImportStep(redisResource),
			//7
			{
				Config:      modulesExplicitlyDisabledConfig,
				ExpectError: regexp.MustCompile(".*module can not be disabled.*"),
			},
			//8
			mdbRedisClusterImportStep(redisResource),
		},
	})
}

// Test
// 1) Sharding can be enabled
// 2) Sharding can't be disabled
// 3) Shards can be created when enabling sharding
// dont need to check in this tests update config
func TestAccMDBRedisClusterV2_enable_sharding(t *testing.T) {
	t.Parallel()

	var r redis.Cluster
	redisName := acctest.RandomWithPrefix("tf-redis-4")
	redisDesc := "Redis Cluster Enabling Sharding Test #4"
	folderID := test.GetExampleFolderID()
	baseDiskSize := 16
	diskTypeId := "network-ssd"
	baseFlavor := "hm3-c2-m8"
	tlsEnabled := false
	persistenceMode := "ON"
	announceHostnames := false
	authSentinel := false
	password := "12345678P"
	version := "8.1-valkey"
	ops := []Op{
		OpCreate,
		OpEnableSh,
		OpAddShard,
		OpRebalance,
		OpDeleteHosts,
	}
	conf := testAccBaseConfig(redisName, redisDesc)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			// Create Redis Cluster
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Config: &config{
						Version:  &version,
						Password: &password,
					},
					Hosts: map[string]host{
						"hst_1": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_2": {Zone: &defaultZone, SubnetId: &defaultSubnet},
						"hst_3": {Zone: &defaultZone, SubnetId: &defaultSubnet},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 3, tlsEnabled, announceHostnames, authSentinel, persistenceMode),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					resource.TestCheckResourceAttr(redisResource, "sharded", "false"),
					testAccCheckMDBRedisOperations(redisResource, ops[:1]),
				),
			},
			mdbRedisClusterImportStep(redisResource),
			// Enable sharding and add one shard
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Hosts: map[string]host{
						"hst_4": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("shard1")},
						"hst_5": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("shard1")},
						"hst_6": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("second")},
						"hst_7": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("second")},
					},
					Sharded: newPtr(true),
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 4, tlsEnabled, announceHostnames, authSentinel, persistenceMode),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					testAccCheckMDBRedisClusterHasShards(&r, []string{"shard1", "second"}),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					resource.TestCheckResourceAttr(redisResource, "sharded", "true"),
					testAccCheckMDBRedisOperations(redisResource, ops),
				),
			},
			mdbRedisClusterImportStep(redisResource),
			// Changing labels works without operations
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Hosts: map[string]host{
						"hst_70": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("second")},
						"hst_60": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("second")},
						"hst_40": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("shard1")},
						"hst_50": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("shard1")},
					},
					Sharded: newPtr(true),
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 4, tlsEnabled, announceHostnames, authSentinel, persistenceMode),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					testAccCheckMDBRedisClusterHasShards(&r, []string{"shard1", "second"}),
					testAccCheckMDBRedisClusterHasResources(&r, baseFlavor, baseDiskSize, diskTypeId),
					resource.TestCheckResourceAttr(redisResource, "sharded", "true"),
					testAccCheckMDBRedisOperations(redisResource, ops),
				),
			},
			mdbRedisClusterImportStep(redisResource),
			// Disabling sharding not works
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Sharded: newPtr(false),
				}),
				ExpectError: regexp.MustCompile(".*Disabling sharding on Redis Cluster is not supported, Id:.*"),
			},
			mdbRedisClusterImportStep(redisResource),
		},
	})
}

// Test
// 1) Sharded cluster can be Created
// 2) Shards can be deleted/created
// 3) Sharded cluster can be upgraded
// dont need to check in this tests update config
func TestAccMDBRedisClusterV2_sharded(t *testing.T) {
	t.Parallel()

	var r redis.Cluster
	redisName := acctest.RandomWithPrefix("tf-redis-5")
	desc := "Sharded Redis Cluster Terraform Test #5"
	folderID := test.GetExampleFolderID()
	baseDiskSize := 16
	diskTypeId := "network-ssd"
	tlsEnabled := false
	announceHostnames := false
	authSentinel := false
	persistenceMode := "ON"
	password := "12345678P"
	conf := testAccBaseConfig(redisName, desc)
	version := "8.0-valkey"
	ops := []Op{
		OpCreate,
		OpAddShard,
		OpRebalance,
		OpDeleteShard,
		OpModify,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBRedisClusterDestroy,
		Steps: []resource.TestStep{
			// Create Redis Cluster
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Config: &config{
						Version:  &version,
						Password: &password,
					},
					Hosts: map[string]host{
						"hst_0": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("first")},
						"hst_2": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("second")},
						"hst_3": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("third")},
					},
					Sharded: newPtr(true),
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 3, tlsEnabled, announceHostnames, authSentinel, persistenceMode),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					testAccCheckMDBRedisClusterHasShards(&r, []string{"first", "second", "third"}),
					testAccCheckMDBRedisClusterHasResources(&r, "hm3-c2-m8", baseDiskSize, diskTypeId),
					testAccCheckMDBRedisOperations(redisResource, ops[:1]),
				),
			},
			// Can't change shard in host
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Hosts: map[string]host{
						"hst_0": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("first")},
						"hst_2": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("second")},
						"hst_3": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("new")},
					},
				}),
				ExpectError: regexp.MustCompile(".*Attributes shard_name, zone, subnet_id can't be changed.*"),
			},
			mdbRedisClusterImportStep(redisResource),
			// Add new shard, delete old shard
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Hosts: map[string]host{
						"hst_00": {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("first")},
						"hst_7":  {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("second")},
						"hst_4":  {Zone: &defaultZone, SubnetId: &defaultSubnet, ShardName: newPtr("new")},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 3, tlsEnabled, announceHostnames, authSentinel, persistenceMode),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					testAccCheckMDBRedisClusterHasShards(&r, []string{"first", "second", "new"}),
					testAccCheckMDBRedisClusterHasResources(&r, "hm3-c2-m8", baseDiskSize, diskTypeId),
					testAccCheckMDBRedisOperations(redisResource, ops[:4]),
				),
			},
			mdbRedisClusterImportStep(redisResource),
			// Upgrade check
			{
				Config: makeConfig(t, conf, &redisConfigTest{
					Config: &config{
						Version:  newPtr("8.1-valkey"),
						Password: &password,
					}}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 3, tlsEnabled, announceHostnames, authSentinel, persistenceMode),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					testAccCheckMDBRedisClusterHasShards(&r, []string{"first", "second", "new"}),
					testAccCheckMDBRedisClusterHasResources(&r, "hm3-c2-m8", baseDiskSize, diskTypeId),
					testAccCheckMDBRedisOperations(redisResource, ops),
				),
			},
			mdbRedisClusterImportStep(redisResource),
		},
	})
}

/*func TestAccMDBRedisClusterV2_diskEncryption(t *testing.T) {
	t.Parallel()

	var r redis.Cluster
	redisName := acctest.RandomWithPrefix("tf-redis-disk-encryption")
	redisDesc := "Redis Cluster Terraform Test Disk Encryption"
	folderID := test.GetExampleFolderID()
	version := "8.1-valkey"
	password := "12345678PP"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             resource.ComposeTestCheckFunc(testAccCheckMDBRedisClusterDestroy, kms_symmetric_key.TestAccCheckYandexKmsSymmetricKeyAllDestroyed),
		Steps: []resource.TestStep{
			// Create Redis Cluster with disk encryption
			{
				Config: diskEncryptionKeyResource + makeConfig(t, testAccBaseConfig(redisName, redisDesc), &redisConfigTest{
					Config: &config{
						Version:  &version,
						Password: &password,
					},
					Hosts: map[string]host{
						"hst_0": {Zone: &defaultZone, SubnetId: &defaultSubnet},
					},
					DiskEncryptionKeyId: newPtr("${yandex_kms_symmetric_key.disk_encrypt.id}"),
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBRedisClusterExists(redisResource, &r, 1, false, false, false, "ON"),
					resource.TestCheckResourceAttr(redisResource, "name", redisName),
					resource.TestCheckResourceAttr(redisResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(redisResource, "description", redisDesc),
					resource.TestCheckResourceAttrSet(redisResource, "disk_encryption_key_id"),
				),
			},
			mdbRedisClusterImportStep(redisResource),
		},
	})
}*/
