package mdb_clickhouse_cluster_v2_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/kms_symmetric_key"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/utils"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexMDBClickHouseClusterCreateTimeout = 30 * time.Minute
	yandexMDBClickHouseClusterDeleteTimeout = 15 * time.Minute
	yandexMDBClickHouseClusterUpdateTimeout = 60 * time.Minute

	chVersion        = "25.3"
	chUpdatedVersion = "25.8"

	chResourceKeeper       = "yandex_mdb_clickhouse_cluster_v2.keeper"
	chResourceCloudStorage = "yandex_mdb_clickhouse_cluster_v2.cloud"
	chResourceSharded      = "yandex_mdb_clickhouse_cluster_v2.bar"
	chResource             = "yandex_mdb_clickhouse_cluster_v2.foo"

	defaultMDBPageSize = 1000
)

var (
	maintenanceWindowAnytime = `
maintenance_window {
	type = "ANYTIME"
}
`
	maintenanceWindowWeekly = `
maintenance_window {
	type = "WEEKLY"
	day  = "FRI"
	hour = 20
}
`
)

var StorageEndpointUrl = getStorageEndpointUrl()

func getStorageEndpointUrl() string {
	rawUrl := os.Getenv("YC_STORAGE_ENDPOINT_URL")
	const protocol = "https://"
	if strings.HasPrefix(rawUrl, protocol) {
		return rawUrl
	}
	return fmt.Sprintf("%s%s", protocol, rawUrl)
}

func init() {
	resource.AddTestSweepers("yandex_mdb_clickhouse_cluster_v2", &resource.Sweeper{
		Name: "yandex_mdb_clickhouse_cluster_v2",
		F:    testSweepMDBClickHouseCluster,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// Tests

// Test that a ClickHouse Cluster can be created, updated and destroyed
func TestAccMDBClickHouseCluster_basic(t *testing.T) {
	t.Parallel()

	var cluster clickhouse.Cluster
	clusterName := acctest.RandomWithPrefix("tf-clickhouse-basic")
	folderID := test.GetExampleFolderID()
	bucketName := acctest.RandomWithPrefix("tf-test-clickhouse-bucket")
	randInt := acctest.RandInt()

	basicConfig := fmt.Sprintf(`
environment = "PRESTABLE"

security_group_ids = ["${yandex_vpc_security_group.mdb-ch-test-sg-x.id}"]
backup_retain_period_days = 12
sql_user_management = false
sql_database_management = false

access = {
	web_sql 	  = true
	data_lens     = true
	metrika       = true
	serverless    = true
	data_transfer = true
	yandex_query  = true
}

# maintenance_window
%s
`,
		maintenanceWindowAnytime,
	)

	updatedConfig := fmt.Sprintf(`
environment = "PRODUCTION"

security_group_ids = ["${yandex_vpc_security_group.mdb-ch-test-sg-x.id}", "${yandex_vpc_security_group.mdb-ch-test-sg-y.id}"]
backup_retain_period_days = 13
sql_user_management = true
sql_database_management = true

access = {
	web_sql 	  = false
	data_lens     = false
	metrika       = false
	serverless    = false
	data_transfer = false
	yandex_query  = false
}

format_schema {
	name = "test_schema"
	type = "FORMAT_SCHEMA_TYPE_CAPNPROTO"
	uri  = "%s/${yandex_storage_bucket.tmp_bucket.bucket}/test.capnp"
}

# maintenance_window
%s
`,
		StorageEndpointUrl,
		maintenanceWindowWeekly,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			// Create ClickHouse Cluster
			{
				Config: testAccMDBClickHouseCluster_basic(clusterName, bucketName, randInt, basicConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &cluster, 1),
					resource.TestCheckResourceAttr(chResource, "name", clusterName),
					resource.TestCheckResourceAttr(chResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResource, "environment", "PRESTABLE"),
					testAccCheckCreatedAtAttr(chResource),
					resource.TestCheckResourceAttr(chResource, "version", chVersion),

					resource.TestCheckResourceAttr(chResource, "access.web_sql", "true"),
					resource.TestCheckResourceAttr(chResource, "access.data_lens", "true"),
					resource.TestCheckResourceAttr(chResource, "access.metrika", "true"),
					resource.TestCheckResourceAttr(chResource, "access.serverless", "true"),
					resource.TestCheckResourceAttr(chResource, "access.data_transfer", "true"),
					resource.TestCheckResourceAttr(chResource, "access.yandex_query", "true"),

					resource.TestCheckResourceAttr(chResource, "backup_window_start.hours", "0"),
					resource.TestCheckResourceAttr(chResource, "backup_window_start.minutes", "0"),

					testAccCheckMDBClickHouseClusterContainsLabel(&cluster, "test_key", "test_value"),
					resource.TestCheckResourceAttr(chResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttrSet(chResource, "service_account_id"),
					resource.TestCheckResourceAttrSet(chResource, "hosts.ha.fqdn"),
					resource.TestCheckResourceAttr(chResource, "maintenance_window.type", "ANYTIME"),
					resource.TestCheckResourceAttr(chResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(chResource, "backup_retain_period_days", "12"),
					resource.TestCheckNoResourceAttr(chResource, "disk_encryption_key_id"),
					resource.TestCheckResourceAttr(chResource, "sql_user_management", "false"),
					resource.TestCheckResourceAttr(chResource, "sql_database_management", "false"),

					testAccCheckMDBClickHouseClusterHasFormatSchemas(chResource, map[string]map[string]string{}),
				),
			},
			mdbClickHouseClusterImportStep(chResource),
			// Update ClickHouse Cluster with weekly maintenance_window
			{
				Config: testAccMDBClickHouseCluster_basic(clusterName, bucketName, randInt, updatedConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &cluster, 1),
					resource.TestCheckResourceAttr(chResource, "name", clusterName),
					resource.TestCheckResourceAttr(chResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResource, "environment", "PRODUCTION"),
					testAccCheckCreatedAtAttr(chResource),
					resource.TestCheckResourceAttr(chResource, "version", chVersion),

					resource.TestCheckResourceAttr(chResource, "access.web_sql", "false"),
					resource.TestCheckResourceAttr(chResource, "access.data_lens", "false"),
					resource.TestCheckResourceAttr(chResource, "access.metrika", "false"),
					resource.TestCheckResourceAttr(chResource, "access.serverless", "false"),
					resource.TestCheckResourceAttr(chResource, "access.data_transfer", "false"),
					resource.TestCheckResourceAttr(chResource, "access.yandex_query", "false"),

					resource.TestCheckResourceAttr(chResource, "backup_window_start.hours", "0"),
					resource.TestCheckResourceAttr(chResource, "backup_window_start.minutes", "0"),

					testAccCheckMDBClickHouseClusterContainsLabel(&cluster, "test_key", "test_value"),
					resource.TestCheckResourceAttr(chResource, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttrSet(chResource, "service_account_id"),
					resource.TestCheckResourceAttrSet(chResource, "hosts.ha.fqdn"),
					resource.TestCheckResourceAttr(chResource, "maintenance_window.type", "WEEKLY"),
					resource.TestCheckResourceAttr(chResource, "maintenance_window.day", "FRI"),
					resource.TestCheckResourceAttr(chResource, "maintenance_window.hour", "20"),
					resource.TestCheckResourceAttr(chResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(chResource, "backup_retain_period_days", "13"),
					resource.TestCheckNoResourceAttr(chResource, "disk_encryption_key_id"),
					resource.TestCheckResourceAttr(chResource, "sql_user_management", "true"),
					resource.TestCheckResourceAttr(chResource, "sql_database_management", "true"),

					testAccCheckMDBClickHouseClusterHasFormatSchemas(chResource, map[string]map[string]string{
						"test_schema": {
							"type": "FORMAT_SCHEMA_TYPE_CAPNPROTO",
							"uri":  fmt.Sprintf("%s/%s/test.capnp", StorageEndpointUrl, bucketName),
						},
					}),
				),
			},
			mdbClickHouseClusterImportStep(chResource),
		},
	})
}

// Test that a ClickHouse Cluster version and resources could be updated simultaneously.
func TestAccMDBClickHouseCluster_resources(t *testing.T) {
	var cluster clickhouse.Cluster
	chName := acctest.RandomWithPrefix("tf-clickhouse-cluster-resources")
	folderID := test.GetExampleFolderID()

	firstShardName := "shard1"

	clickHouseFirstResources := &clickhouse.Resources{
		ResourcePresetId: "s2.micro",
		DiskTypeId:       "network-ssd",
		DiskSize:         10737418240,
	}

	clickHouseSecondResources := &clickhouse.Resources{
		ResourcePresetId: "s2.small",
		DiskTypeId:       "network-ssd",
		DiskSize:         21474836480,
	}

	clickHouseThirdResources := &clickhouse.Resources{
		ResourcePresetId: "s2.micro",
		DiskTypeId:       "network-ssd",
		DiskSize:         21474836480,
	}

	zookeeperFirstResources := &clickhouse.Resources{
		ResourcePresetId: "s2.micro",
		DiskTypeId:       "network-ssd",
		DiskSize:         10737418240,
	}

	zookeeperSecondResources := &clickhouse.Resources{
		ResourcePresetId: "s2.small",
		DiskTypeId:       "network-ssd",
		DiskSize:         21474836480,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			// Create ClickHouse Cluster
			{
				Config: testAccMDBClickHouseCluster_resources(chName, firstShardName, chVersion, clickHouseFirstResources, zookeeperFirstResources, nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &cluster, 4),
					resource.TestCheckResourceAttr(chResource, "name", chName),
					resource.TestCheckResourceAttr(chResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResource, "version", chVersion),
					testAccCheckMDBClickHouseClusterHasResources(&cluster, clickHouseFirstResources.ResourcePresetId, clickHouseFirstResources.DiskTypeId, clickHouseFirstResources.DiskSize),
					testAccCheckMDBClickHouseShardHasResources(&cluster, firstShardName, clickHouseFirstResources.ResourcePresetId, clickHouseFirstResources.DiskTypeId, clickHouseFirstResources.DiskSize),
					testAccCheckMDBClickHouseZooKeeperSubclusterHasResources(&cluster, zookeeperFirstResources.ResourcePresetId, zookeeperFirstResources.DiskTypeId, zookeeperFirstResources.DiskSize),
					testAccCheckCreatedAtAttr(chResource)),
			},
			mdbClickHouseClusterImportStep(chResource),
			// Update ClickHouse version only
			{
				Config: testAccMDBClickHouseCluster_resources(chName, firstShardName, chUpdatedVersion, clickHouseFirstResources, zookeeperFirstResources, nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &cluster, 4),
					resource.TestCheckResourceAttr(chResource, "name", chName),
					resource.TestCheckResourceAttr(chResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResource, "version", chUpdatedVersion),
					testAccCheckMDBClickHouseClusterHasResources(&cluster, clickHouseFirstResources.ResourcePresetId, clickHouseFirstResources.DiskTypeId, clickHouseFirstResources.DiskSize),
					testAccCheckMDBClickHouseShardHasResources(&cluster, firstShardName, clickHouseFirstResources.ResourcePresetId, clickHouseFirstResources.DiskTypeId, clickHouseFirstResources.DiskSize),
					testAccCheckMDBClickHouseZooKeeperSubclusterHasResources(&cluster, zookeeperFirstResources.ResourcePresetId, zookeeperFirstResources.DiskTypeId, zookeeperFirstResources.DiskSize),
					testAccCheckCreatedAtAttr(chResource)),
			},
			mdbClickHouseClusterImportStep(chResource),
			// Update first shard and zookeeper resources. Changing the resource management shard-mode.
			{
				Config: testAccMDBClickHouseCluster_resources(chName, firstShardName, chVersion, nil, zookeeperSecondResources, clickHouseSecondResources),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &cluster, 4),
					resource.TestCheckResourceAttr(chResource, "name", chName),
					resource.TestCheckResourceAttr(chResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResource, "version", chVersion),
					testAccCheckMDBClickHouseClusterHasResources(&cluster, clickHouseSecondResources.ResourcePresetId, clickHouseSecondResources.DiskTypeId, clickHouseSecondResources.DiskSize),
					testAccCheckMDBClickHouseShardHasResources(&cluster, firstShardName, clickHouseSecondResources.ResourcePresetId, clickHouseSecondResources.DiskTypeId, clickHouseSecondResources.DiskSize),
					testAccCheckMDBClickHouseZooKeeperSubclusterHasResources(&cluster, zookeeperSecondResources.ResourcePresetId, zookeeperSecondResources.DiskTypeId, zookeeperSecondResources.DiskSize),
					testAccCheckCreatedAtAttr(chResource)),
			},
			mdbClickHouseClusterImportStep(chResource),
			// Plan should fail when both clickhouse.resources and shards[*].resources are set
			{
				Config:      testAccMDBClickHouseCluster_resources(chName, firstShardName, chVersion, clickHouseSecondResources, zookeeperSecondResources, clickHouseSecondResources),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)Invalid Attribute Combination|cannot be configured together|These attributes cannot be configured together`),
			},
			// Downgrade ClickHouse version and cluster resources. Changing the resource management cluster-mode.
			{
				Config: testAccMDBClickHouseCluster_resources(chName, firstShardName, chVersion, clickHouseThirdResources, zookeeperSecondResources, nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &cluster, 4),
					resource.TestCheckResourceAttr(chResource, "name", chName),
					resource.TestCheckResourceAttr(chResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResource, "version", chVersion),
					testAccCheckMDBClickHouseClusterHasResources(&cluster, clickHouseThirdResources.ResourcePresetId, clickHouseThirdResources.DiskTypeId, clickHouseThirdResources.DiskSize),
					testAccCheckMDBClickHouseShardHasResources(&cluster, firstShardName, clickHouseThirdResources.ResourcePresetId, clickHouseThirdResources.DiskTypeId, clickHouseThirdResources.DiskSize),
					testAccCheckMDBClickHouseZooKeeperSubclusterHasResources(&cluster, zookeeperSecondResources.ResourcePresetId, zookeeperSecondResources.DiskTypeId, zookeeperSecondResources.DiskSize),
					testAccCheckCreatedAtAttr(chResource)),
			},
			mdbClickHouseClusterImportStep(chResource),
		},
	})
}

func TestAccMDBClickHouseCluster_clickhouse_config(t *testing.T) {
	t.Parallel()

	var r clickhouse.Cluster
	chName := acctest.RandomWithPrefix("tf-clickhouse")
	folderID := test.GetExampleFolderID()

	configForFirstStep := &clickhouseConfig.ClickhouseConfig{
		MergeTree: &clickhouseConfig.ClickhouseConfig_MergeTree{
			ReplicatedDeduplicationWindow:                  &wrappers.Int64Value{Value: 1000},
			ReplicatedDeduplicationWindowSeconds:           &wrappers.Int64Value{Value: 1000},
			PartsToDelayInsert:                             &wrappers.Int64Value{Value: 110001},
			PartsToThrowInsert:                             &wrappers.Int64Value{Value: 11001},
			InactivePartsToDelayInsert:                     &wrappers.Int64Value{Value: 101},
			InactivePartsToThrowInsert:                     &wrappers.Int64Value{Value: 110},
			MaxReplicatedMergesInQueue:                     &wrappers.Int64Value{Value: 11000},
			NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge: &wrappers.Int64Value{Value: 15},
			MaxBytesToMergeAtMinSpaceInPool:                &wrappers.Int64Value{Value: 11000},
			MaxBytesToMergeAtMaxSpaceInPool:                &wrappers.Int64Value{Value: 16106127300},
			MinBytesForWidePart:                            &wrappers.Int64Value{Value: 0},
			MinRowsForWidePart:                             &wrappers.Int64Value{Value: 0},
			TtlOnlyDropParts:                               &wrappers.BoolValue{Value: false},
			MergeWithTtlTimeout:                            &wrappers.Int64Value{Value: 100005},
			MergeWithRecompressionTtlTimeout:               &wrappers.Int64Value{Value: 100006},
			MaxPartsInTotal:                                &wrappers.Int64Value{Value: 100007},
			MaxNumberOfMergesWithTtlInPool:                 &wrappers.Int64Value{Value: 1},
			CleanupDelayPeriod:                             &wrappers.Int64Value{Value: 120},
			NumberOfFreeEntriesInPoolToExecuteMutation:     &wrappers.Int64Value{Value: 30},
			MaxAvgPartSizeForTooManyParts:                  &wrappers.Int64Value{Value: 100009},
			MinAgeToForceMergeSeconds:                      &wrappers.Int64Value{Value: 100010},
			MinAgeToForceMergeOnPartitionOnly:              &wrappers.BoolValue{Value: false},
			MergeSelectingSleepMs:                          &wrappers.Int64Value{Value: 5001},
			MergeMaxBlockSize:                              &wrappers.Int64Value{Value: 5001},
			CheckSampleColumnIsCorrect:                     &wrappers.BoolValue{Value: true},
			MaxMergeSelectingSleepMs:                       &wrappers.Int64Value{Value: 50001},
			MaxCleanupDelayPeriod:                          &wrappers.Int64Value{Value: 201},
			DeduplicateMergeProjectionMode:                 clickhouseConfig.ClickhouseConfig_MergeTree_DEDUPLICATE_MERGE_PROJECTION_MODE_DROP,
			LightweightMutationProjectionMode:              clickhouseConfig.ClickhouseConfig_MergeTree_LIGHTWEIGHT_MUTATION_PROJECTION_MODE_REBUILD,
			MaterializeTtlRecalculateOnly:                  &wrappers.BoolValue{Value: true},
			FsyncAfterInsert:                               &wrappers.BoolValue{Value: true},
			FsyncPartDirectory:                             &wrappers.BoolValue{Value: true},
			MinCompressedBytesToFsyncAfterFetch:            &wrappers.Int64Value{Value: 1024},
			MinCompressedBytesToFsyncAfterMerge:            &wrappers.Int64Value{Value: 2048},
			MinRowsToFsyncAfterMerge:                       &wrappers.Int64Value{Value: 32},
		},
		Kafka: &clickhouseConfig.ClickhouseConfig_Kafka{
			SecurityProtocol:                 clickhouseConfig.ClickhouseConfig_Kafka_SECURITY_PROTOCOL_PLAINTEXT,
			SaslMechanism:                    clickhouseConfig.ClickhouseConfig_Kafka_SASL_MECHANISM_GSSAPI,
			SaslUsername:                     "user1",
			SaslPassword:                     "pass1",
			Debug:                            clickhouseConfig.ClickhouseConfig_Kafka_DEBUG_GENERIC,
			AutoOffsetReset:                  clickhouseConfig.ClickhouseConfig_Kafka_AUTO_OFFSET_RESET_SMALLEST,
			EnableSslCertificateVerification: &wrappers.BoolValue{Value: true},
			MaxPollIntervalMs:                &wrappers.Int64Value{Value: 300000},
			SessionTimeoutMs:                 &wrappers.Int64Value{Value: 45000},
		},
		Rabbitmq: &clickhouseConfig.ClickhouseConfig_Rabbitmq{
			Username: "rabbit_user",
			Password: "rabbit_pass",
			Vhost:    "old_clickhouse",
		},
		Compression: []*clickhouseConfig.ClickhouseConfig_Compression{
			{
				Method:           clickhouseConfig.ClickhouseConfig_Compression_ZSTD,
				MinPartSize:      1024,
				MinPartSizeRatio: 0.5,
				Level:            &wrappers.Int64Value{Value: 5},
			},
		},
		GraphiteRollup: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup{
			{
				Name: "rollup1",
				Patterns: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern{
					{
						Regexp:   "abc",
						Function: "func1",
						Retention: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention{
							{
								Age:       1000,
								Precision: 3,
							},
						},
					},
				},
				PathColumnName:    "path1",
				TimeColumnName:    "time1",
				ValueColumnName:   "value1",
				VersionColumnName: "version1",
			},
		},
		QueryMaskingRules: []*clickhouseConfig.ClickhouseConfig_QueryMaskingRule{
			{
				Name:    "name1",
				Regexp:  "regexp1",
				Replace: "replace1",
			},
		},
		QueryCache: &clickhouseConfig.ClickhouseConfig_QueryCache{
			MaxSizeInBytes:      &wrappers.Int64Value{Value: 1073741820},
			MaxEntries:          &wrappers.Int64Value{Value: 1020},
			MaxEntrySizeInBytes: &wrappers.Int64Value{Value: 1048570},
			MaxEntrySizeInRows:  &wrappers.Int64Value{Value: 20000000},
		},
		JdbcBridge: &clickhouseConfig.ClickhouseConfig_JdbcBridge{
			Host: "127.0.0.2",
			Port: &wrappers.Int64Value{Value: 8999},
		},
		LogLevel:                                  clickhouseConfig.ClickhouseConfig_TRACE,
		MaxConnections:                            &wrappers.Int64Value{Value: 512},
		MaxConcurrentQueries:                      &wrappers.Int64Value{Value: 100},
		KeepAliveTimeout:                          &wrappers.Int64Value{Value: 123000},
		UncompressedCacheSize:                     &wrappers.Int64Value{Value: 8096},
		MaxTableSizeToDrop:                        &wrappers.Int64Value{Value: 1024},
		MaxPartitionSizeToDrop:                    &wrappers.Int64Value{Value: 1024},
		Timezone:                                  "UTC",
		GeobaseUri:                                "",
		GeobaseEnabled:                            &wrappers.BoolValue{Value: false},
		QueryLogRetentionSize:                     &wrappers.Int64Value{Value: 1001},
		QueryLogRetentionTime:                     &wrappers.Int64Value{Value: 86400000},
		QueryThreadLogEnabled:                     &wrappers.BoolValue{Value: true},
		QueryThreadLogRetentionSize:               &wrappers.Int64Value{Value: 1002},
		QueryThreadLogRetentionTime:               &wrappers.Int64Value{Value: 86400000},
		PartLogRetentionSize:                      &wrappers.Int64Value{Value: 1003},
		PartLogRetentionTime:                      &wrappers.Int64Value{Value: 86400000},
		MetricLogEnabled:                          &wrappers.BoolValue{Value: true},
		MetricLogRetentionSize:                    &wrappers.Int64Value{Value: 1004},
		MetricLogRetentionTime:                    &wrappers.Int64Value{Value: 86400000},
		TraceLogEnabled:                           &wrappers.BoolValue{Value: true},
		TraceLogRetentionSize:                     &wrappers.Int64Value{Value: 1005},
		TraceLogRetentionTime:                     &wrappers.Int64Value{Value: 86400000},
		TextLogEnabled:                            &wrappers.BoolValue{Value: true},
		TextLogRetentionSize:                      &wrappers.Int64Value{Value: 1006},
		TextLogRetentionTime:                      &wrappers.Int64Value{Value: 86400000},
		OpentelemetrySpanLogEnabled:               &wrappers.BoolValue{Value: true},
		OpentelemetrySpanLogRetentionSize:         &wrappers.Int64Value{Value: 1007},
		OpentelemetrySpanLogRetentionTime:         &wrappers.Int64Value{Value: 86400000},
		QueryViewsLogEnabled:                      &wrappers.BoolValue{Value: true},
		QueryViewsLogRetentionSize:                &wrappers.Int64Value{Value: 1008},
		QueryViewsLogRetentionTime:                &wrappers.Int64Value{Value: 86400000},
		AsynchronousMetricLogEnabled:              &wrappers.BoolValue{Value: true},
		AsynchronousMetricLogRetentionSize:        &wrappers.Int64Value{Value: 1009},
		AsynchronousMetricLogRetentionTime:        &wrappers.Int64Value{Value: 86400000},
		SessionLogEnabled:                         &wrappers.BoolValue{Value: true},
		SessionLogRetentionSize:                   &wrappers.Int64Value{Value: 1010},
		SessionLogRetentionTime:                   &wrappers.Int64Value{Value: 86400000},
		ZookeeperLogEnabled:                       &wrappers.BoolValue{Value: true},
		ZookeeperLogRetentionSize:                 &wrappers.Int64Value{Value: 1011},
		ZookeeperLogRetentionTime:                 &wrappers.Int64Value{Value: 86400000},
		AsynchronousInsertLogEnabled:              &wrappers.BoolValue{Value: true},
		AsynchronousInsertLogRetentionSize:        &wrappers.Int64Value{Value: 1012},
		AsynchronousInsertLogRetentionTime:        &wrappers.Int64Value{Value: 86400000},
		TextLogLevel:                              clickhouseConfig.ClickhouseConfig_WARNING,
		BackgroundPoolSize:                        &wrappers.Int64Value{Value: 16},
		BackgroundSchedulePoolSize:                &wrappers.Int64Value{Value: 32},
		BackgroundFetchesPoolSize:                 &wrappers.Int64Value{Value: 8},
		BackgroundMovePoolSize:                    &wrappers.Int64Value{Value: 8},
		BackgroundDistributedSchedulePoolSize:     &wrappers.Int64Value{Value: 8},
		BackgroundBufferFlushSchedulePoolSize:     &wrappers.Int64Value{Value: 8},
		BackgroundCommonPoolSize:                  &wrappers.Int64Value{Value: 8},
		BackgroundMessageBrokerSchedulePoolSize:   &wrappers.Int64Value{Value: 9},
		BackgroundMergesMutationsConcurrencyRatio: &wrappers.Int64Value{Value: 3},
		TotalMemoryProfilerStep:                   &wrappers.Int64Value{Value: 4194304},
		DictionariesLazyLoad:                      &wrappers.BoolValue{Value: true},
		ProcessorsProfileLogEnabled:               &wrappers.BoolValue{Value: true},
		ProcessorsProfileLogRetentionSize:         &wrappers.Int64Value{Value: 1013},
		ProcessorsProfileLogRetentionTime:         &wrappers.Int64Value{Value: 86400000},
		ErrorLogEnabled:                           &wrappers.BoolValue{Value: true},
		ErrorLogRetentionSize:                     &wrappers.Int64Value{Value: 1014},
		ErrorLogRetentionTime:                     &wrappers.Int64Value{Value: 86400000},
		QueryMetricLogEnabled:                     &wrappers.BoolValue{Value: true},
		QueryMetricLogRetentionSize:               &wrappers.Int64Value{Value: 1015},
		QueryMetricLogRetentionTime:               &wrappers.Int64Value{Value: 90000000},
		TotalMemoryTrackerSampleProbability:       &wrappers.DoubleValue{Value: 0.123},
		AsyncInsertThreads:                        &wrappers.Int64Value{Value: 4},
		BackupThreads:                             &wrappers.Int64Value{Value: 2},
		RestoreThreads:                            &wrappers.Int64Value{Value: 2},
		MysqlProtocol:                             &wrappers.BoolValue{Value: true},
		AccessControlImprovements: &clickhouseConfig.ClickhouseConfig_AccessControlImprovements{
			SelectFromSystemDbRequiresGrant:          &wrappers.BoolValue{Value: false},
			SelectFromInformationSchemaRequiresGrant: &wrappers.BoolValue{Value: false},
		},
		CustomMacros: []*clickhouseConfig.ClickhouseConfig_Macro{
			{
				Name:  "macro1",
				Value: "value1",
			},
		},
	}

	configForSecondStep := &clickhouseConfig.ClickhouseConfig{
		MergeTree: &clickhouseConfig.ClickhouseConfig_MergeTree{
			ReplicatedDeduplicationWindow:                  &wrappers.Int64Value{Value: 100},
			ReplicatedDeduplicationWindowSeconds:           &wrappers.Int64Value{Value: 604800},
			PartsToDelayInsert:                             &wrappers.Int64Value{Value: 150},
			PartsToThrowInsert:                             &wrappers.Int64Value{Value: 12000},
			InactivePartsToDelayInsert:                     &wrappers.Int64Value{Value: 102},
			InactivePartsToThrowInsert:                     &wrappers.Int64Value{Value: 120},
			MaxReplicatedMergesInQueue:                     &wrappers.Int64Value{Value: 16},
			NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge: &wrappers.Int64Value{Value: 8},
			MaxBytesToMergeAtMinSpaceInPool:                &wrappers.Int64Value{Value: 1048576},
			MaxBytesToMergeAtMaxSpaceInPool:                &wrappers.Int64Value{Value: 16106127301},
			MinBytesForWidePart:                            &wrappers.Int64Value{Value: 512},
			MinRowsForWidePart:                             &wrappers.Int64Value{Value: 16},
			TtlOnlyDropParts:                               &wrappers.BoolValue{Value: true},
			MergeWithTtlTimeout:                            &wrappers.Int64Value{Value: 200010},
			MergeWithRecompressionTtlTimeout:               &wrappers.Int64Value{Value: 200012},
			MaxPartsInTotal:                                &wrappers.Int64Value{Value: 200014},
			MaxNumberOfMergesWithTtlInPool:                 &wrappers.Int64Value{Value: 2},
			CleanupDelayPeriod:                             &wrappers.Int64Value{Value: 240},
			NumberOfFreeEntriesInPoolToExecuteMutation:     &wrappers.Int64Value{Value: 40},
			MaxAvgPartSizeForTooManyParts:                  &wrappers.Int64Value{Value: 200018},
			MinAgeToForceMergeSeconds:                      &wrappers.Int64Value{Value: 200020},
			MinAgeToForceMergeOnPartitionOnly:              &wrappers.BoolValue{Value: true},
			MergeSelectingSleepMs:                          &wrappers.Int64Value{Value: 5002},
			MergeMaxBlockSize:                              &wrappers.Int64Value{Value: 5002},
			CheckSampleColumnIsCorrect:                     &wrappers.BoolValue{Value: false},
			MaxMergeSelectingSleepMs:                       &wrappers.Int64Value{Value: 100001},
			MaxCleanupDelayPeriod:                          &wrappers.Int64Value{Value: 401},
			DeduplicateMergeProjectionMode:                 clickhouseConfig.ClickhouseConfig_MergeTree_DEDUPLICATE_MERGE_PROJECTION_MODE_REBUILD,
			LightweightMutationProjectionMode:              clickhouseConfig.ClickhouseConfig_MergeTree_LIGHTWEIGHT_MUTATION_PROJECTION_MODE_DROP,
			MaterializeTtlRecalculateOnly:                  &wrappers.BoolValue{Value: false},
			FsyncAfterInsert:                               &wrappers.BoolValue{Value: false},
			FsyncPartDirectory:                             &wrappers.BoolValue{Value: false},
			MinCompressedBytesToFsyncAfterFetch:            &wrappers.Int64Value{Value: 4096},
			MinCompressedBytesToFsyncAfterMerge:            &wrappers.Int64Value{Value: 8192},
			MinRowsToFsyncAfterMerge:                       &wrappers.Int64Value{Value: 64},
		},
		Kafka: &clickhouseConfig.ClickhouseConfig_Kafka{
			SecurityProtocol:                 clickhouseConfig.ClickhouseConfig_Kafka_SECURITY_PROTOCOL_PLAINTEXT,
			SaslMechanism:                    clickhouseConfig.ClickhouseConfig_Kafka_SASL_MECHANISM_GSSAPI,
			SaslUsername:                     "user1",
			SaslPassword:                     "pass1",
			Debug:                            clickhouseConfig.ClickhouseConfig_Kafka_DEBUG_METADATA,
			AutoOffsetReset:                  clickhouseConfig.ClickhouseConfig_Kafka_AUTO_OFFSET_RESET_LARGEST,
			EnableSslCertificateVerification: &wrappers.BoolValue{Value: false},
			MaxPollIntervalMs:                &wrappers.Int64Value{Value: 400000},
			SessionTimeoutMs:                 &wrappers.Int64Value{Value: 60000},
		},
		Rabbitmq: &clickhouseConfig.ClickhouseConfig_Rabbitmq{
			Username: "rabbit_user",
			Password: "rabbit_pass2",
			Vhost:    "clickhouse",
		},
		Compression: []*clickhouseConfig.ClickhouseConfig_Compression{
			{
				Method:           clickhouseConfig.ClickhouseConfig_Compression_LZ4HC,
				MinPartSize:      2024,
				MinPartSizeRatio: 0.3,
				Level:            &wrappers.Int64Value{Value: 7},
			},
			{
				Method:           clickhouseConfig.ClickhouseConfig_Compression_ZSTD,
				MinPartSize:      4048,
				MinPartSizeRatio: 0.77,
				Level:            &wrappers.Int64Value{Value: 3},
			},
		},
		GraphiteRollup: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup{
			{
				Name: "rollup1",
				Patterns: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern{
					{
						Regexp:   "abc",
						Function: "func1",
						Retention: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention{
							{
								Age:       1000,
								Precision: 3,
							},
						},
					},
				},
				PathColumnName:    "path1",
				TimeColumnName:    "time1",
				ValueColumnName:   "value1",
				VersionColumnName: "version1",
			},
			{
				Name: "rollup2",
				Patterns: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern{
					{
						Regexp:   "abc",
						Function: "func3",
						Retention: []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention{
							{
								Age:       3000,
								Precision: 7,
							},
						},
					},
				},
				PathColumnName:    "path2",
				TimeColumnName:    "time2",
				ValueColumnName:   "value2",
				VersionColumnName: "version2",
			},
		},
		QueryMaskingRules: []*clickhouseConfig.ClickhouseConfig_QueryMaskingRule{
			{
				Name:    "name11",
				Regexp:  "regexp11",
				Replace: "replace11",
			},
			{
				Regexp: "regexp22",
			},
		},
		QueryCache: &clickhouseConfig.ClickhouseConfig_QueryCache{
			MaxSizeInBytes:      &wrappers.Int64Value{Value: 2073741820},
			MaxEntries:          &wrappers.Int64Value{Value: 2020},
			MaxEntrySizeInBytes: &wrappers.Int64Value{Value: 2048570},
			MaxEntrySizeInRows:  &wrappers.Int64Value{Value: 30000000},
		},
		JdbcBridge: &clickhouseConfig.ClickhouseConfig_JdbcBridge{
			Host: "127.0.0.3",
			Port: &wrappers.Int64Value{Value: 8998},
		},
		LogLevel:                                  clickhouseConfig.ClickhouseConfig_WARNING,
		MaxConnections:                            &wrappers.Int64Value{Value: 1024},
		MaxConcurrentQueries:                      &wrappers.Int64Value{Value: 200},
		KeepAliveTimeout:                          &wrappers.Int64Value{Value: 246000},
		UncompressedCacheSize:                     &wrappers.Int64Value{Value: 16192},
		MaxTableSizeToDrop:                        &wrappers.Int64Value{Value: 2048},
		MaxPartitionSizeToDrop:                    &wrappers.Int64Value{Value: 2048},
		Timezone:                                  "UTC",
		GeobaseUri:                                "",
		GeobaseEnabled:                            &wrappers.BoolValue{Value: false}, // make true when extensions are supported
		QueryLogRetentionSize:                     &wrappers.Int64Value{Value: 2001},
		QueryLogRetentionTime:                     &wrappers.Int64Value{Value: 86400000},
		QueryThreadLogEnabled:                     &wrappers.BoolValue{Value: true},
		QueryThreadLogRetentionSize:               &wrappers.Int64Value{Value: 2002},
		QueryThreadLogRetentionTime:               &wrappers.Int64Value{Value: 86400000},
		PartLogRetentionSize:                      &wrappers.Int64Value{Value: 2003},
		PartLogRetentionTime:                      &wrappers.Int64Value{Value: 86400000},
		MetricLogEnabled:                          &wrappers.BoolValue{Value: true},
		MetricLogRetentionSize:                    &wrappers.Int64Value{Value: 2004},
		MetricLogRetentionTime:                    &wrappers.Int64Value{Value: 86400000},
		TraceLogEnabled:                           &wrappers.BoolValue{Value: true},
		TraceLogRetentionSize:                     &wrappers.Int64Value{Value: 2005},
		TraceLogRetentionTime:                     &wrappers.Int64Value{Value: 86400000},
		TextLogEnabled:                            &wrappers.BoolValue{Value: true},
		TextLogRetentionSize:                      &wrappers.Int64Value{Value: 2006},
		TextLogRetentionTime:                      &wrappers.Int64Value{Value: 86400000},
		OpentelemetrySpanLogEnabled:               &wrappers.BoolValue{Value: true},
		OpentelemetrySpanLogRetentionSize:         &wrappers.Int64Value{Value: 2007},
		OpentelemetrySpanLogRetentionTime:         &wrappers.Int64Value{Value: 86400000},
		QueryViewsLogEnabled:                      &wrappers.BoolValue{Value: true},
		QueryViewsLogRetentionSize:                &wrappers.Int64Value{Value: 2008},
		QueryViewsLogRetentionTime:                &wrappers.Int64Value{Value: 86400000},
		AsynchronousMetricLogEnabled:              &wrappers.BoolValue{Value: true},
		AsynchronousMetricLogRetentionSize:        &wrappers.Int64Value{Value: 2009},
		AsynchronousMetricLogRetentionTime:        &wrappers.Int64Value{Value: 86400000},
		SessionLogEnabled:                         &wrappers.BoolValue{Value: true},
		SessionLogRetentionSize:                   &wrappers.Int64Value{Value: 2010},
		SessionLogRetentionTime:                   &wrappers.Int64Value{Value: 86400000},
		ZookeeperLogEnabled:                       &wrappers.BoolValue{Value: true},
		ZookeeperLogRetentionSize:                 &wrappers.Int64Value{Value: 2011},
		ZookeeperLogRetentionTime:                 &wrappers.Int64Value{Value: 86400000},
		AsynchronousInsertLogEnabled:              &wrappers.BoolValue{Value: true},
		AsynchronousInsertLogRetentionSize:        &wrappers.Int64Value{Value: 2012},
		AsynchronousInsertLogRetentionTime:        &wrappers.Int64Value{Value: 86400000},
		TextLogLevel:                              clickhouseConfig.ClickhouseConfig_ERROR,
		BackgroundPoolSize:                        &wrappers.Int64Value{Value: 32},
		BackgroundSchedulePoolSize:                &wrappers.Int64Value{Value: 64},
		BackgroundFetchesPoolSize:                 &wrappers.Int64Value{Value: 16},
		BackgroundMovePoolSize:                    &wrappers.Int64Value{Value: 16},
		BackgroundDistributedSchedulePoolSize:     &wrappers.Int64Value{Value: 16},
		BackgroundBufferFlushSchedulePoolSize:     &wrappers.Int64Value{Value: 16},
		BackgroundCommonPoolSize:                  &wrappers.Int64Value{Value: 16},
		BackgroundMessageBrokerSchedulePoolSize:   &wrappers.Int64Value{Value: 17},
		BackgroundMergesMutationsConcurrencyRatio: &wrappers.Int64Value{Value: 4},
		TotalMemoryProfilerStep:                   &wrappers.Int64Value{Value: 4194303},
		DictionariesLazyLoad:                      &wrappers.BoolValue{Value: false},
		ProcessorsProfileLogEnabled:               &wrappers.BoolValue{Value: false},
		ProcessorsProfileLogRetentionSize:         &wrappers.Int64Value{Value: 2013},
		ProcessorsProfileLogRetentionTime:         &wrappers.Int64Value{Value: 86400000},
		ErrorLogEnabled:                           &wrappers.BoolValue{Value: false},
		ErrorLogRetentionSize:                     &wrappers.Int64Value{Value: 2014},
		ErrorLogRetentionTime:                     &wrappers.Int64Value{Value: 86400000},
		QueryMetricLogEnabled:                     &wrappers.BoolValue{Value: false},
		QueryMetricLogRetentionSize:               &wrappers.Int64Value{Value: 2015},
		QueryMetricLogRetentionTime:               &wrappers.Int64Value{Value: 80000000},
		TotalMemoryTrackerSampleProbability:       &wrappers.DoubleValue{Value: 0.321},
		AsyncInsertThreads:                        &wrappers.Int64Value{Value: 8},
		BackupThreads:                             &wrappers.Int64Value{Value: 4},
		RestoreThreads:                            &wrappers.Int64Value{Value: 6},
		MysqlProtocol:                             &wrappers.BoolValue{Value: false},
		AccessControlImprovements: &clickhouseConfig.ClickhouseConfig_AccessControlImprovements{
			SelectFromSystemDbRequiresGrant:          &wrappers.BoolValue{Value: true},
			SelectFromInformationSchemaRequiresGrant: &wrappers.BoolValue{Value: true},
		},
		CustomMacros: []*clickhouseConfig.ClickhouseConfig_Macro{
			{
				Name:  "macro1",
				Value: "value1",
			},
			{
				Name:  "macro2",
				Value: "value2",
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBClickHouseCluster_clickhouse_config(chName, configForFirstStep),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &r, 1),
					resource.TestCheckResourceAttr(chResource, "name", chName),
					resource.TestCheckResourceAttr(chResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResource, "version", chVersion),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.log_level", "TRACE"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.max_connections", "512"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.max_concurrent_queries", "100"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.keep_alive_timeout", "123000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.uncompressed_cache_size", "8096"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.max_table_size_to_drop", "1024"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.max_partition_size_to_drop", "1024"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.timezone", "UTC"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.geobase_uri", ""),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.geobase_enabled", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_log_retention_size", "1001"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_thread_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_thread_log_retention_size", "1002"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_thread_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.part_log_retention_size", "1003"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.part_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.metric_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.metric_log_retention_size", "1004"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.metric_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.trace_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.trace_log_retention_size", "1005"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.trace_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.text_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.text_log_retention_size", "1006"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.text_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.opentelemetry_span_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.opentelemetry_span_log_retention_size", "1007"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.opentelemetry_span_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_views_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_views_log_retention_size", "1008"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_views_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_metric_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_metric_log_retention_size", "1009"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_metric_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.session_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.session_log_retention_size", "1010"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.session_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.zookeeper_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.zookeeper_log_retention_size", "1011"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.zookeeper_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_insert_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_insert_log_retention_size", "1012"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_insert_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.text_log_level", "WARNING"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_pool_size", "16"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_schedule_pool_size", "32"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_fetches_pool_size", "8"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_move_pool_size", "8"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_distributed_schedule_pool_size", "8"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_buffer_flush_schedule_pool_size", "8"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_common_pool_size", "8"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_message_broker_schedule_pool_size", "9"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_merges_mutations_concurrency_ratio", "3"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.total_memory_profiler_step", "4194304"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.dictionaries_lazy_load", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.processors_profile_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.processors_profile_log_retention_size", "1013"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.processors_profile_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.error_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.error_log_retention_size", "1014"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.error_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_metric_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_metric_log_retention_size", "1015"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_metric_log_retention_time", "90000000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.total_memory_tracker_sample_probability", "0.123"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.async_insert_threads", "4"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.backup_threads", "2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.restore_threads", "2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.mysql_protocol", "true"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.access_control_improvements.select_from_system_db_requires_grant", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.access_control_improvements.select_from_information_schema_requires_grant", "false"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.custom_macros.#", "1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.custom_macros.0.name", "macro1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.custom_macros.0.value", "value1"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.replicated_deduplication_window", "1000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.replicated_deduplication_window_seconds", "1000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.parts_to_delay_insert", "110001"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.parts_to_throw_insert", "11001"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.inactive_parts_to_delay_insert", "101"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.inactive_parts_to_throw_insert", "110"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_replicated_merges_in_queue", "11000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.number_of_free_entries_in_pool_to_lower_max_size_of_merge", "15"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_bytes_to_merge_at_min_space_in_pool", "11000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_bytes_to_merge_at_max_space_in_pool", "16106127300"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_bytes_for_wide_part", "0"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_rows_for_wide_part", "0"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.ttl_only_drop_parts", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.merge_with_ttl_timeout", "100005"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.merge_with_recompression_ttl_timeout", "100006"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_parts_in_total", "100007"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_number_of_merges_with_ttl_in_pool", "1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.cleanup_delay_period", "120"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.number_of_free_entries_in_pool_to_execute_mutation", "30"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_avg_part_size_for_too_many_parts", "100009"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_age_to_force_merge_seconds", "100010"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_age_to_force_merge_on_partition_only", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.merge_selecting_sleep_ms", "5001"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.merge_max_block_size", "5001"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.check_sample_column_is_correct", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_merge_selecting_sleep_ms", "50001"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_cleanup_delay_period", "201"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.deduplicate_merge_projection_mode", "DEDUPLICATE_MERGE_PROJECTION_MODE_DROP"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.lightweight_mutation_projection_mode", "LIGHTWEIGHT_MUTATION_PROJECTION_MODE_REBUILD"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.materialize_ttl_recalculate_only", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.fsync_after_insert", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.fsync_part_directory", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_compressed_bytes_to_fsync_after_fetch", "1024"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_compressed_bytes_to_fsync_after_merge", "2048"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_rows_to_fsync_after_merge", "32"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.security_protocol", "SECURITY_PROTOCOL_PLAINTEXT"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.sasl_mechanism", "SASL_MECHANISM_GSSAPI"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.sasl_username", "user1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.sasl_password", "pass1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.debug", "DEBUG_GENERIC"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.auto_offset_reset", "AUTO_OFFSET_RESET_SMALLEST"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.enable_ssl_certificate_verification", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.max_poll_interval_ms", "300000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.session_timeout_ms", "45000"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.rabbitmq.username", "rabbit_user"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.rabbitmq.password", "rabbit_pass"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.rabbitmq.vhost", "old_clickhouse"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.#", "1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.0.method", "ZSTD"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.0.min_part_size", "1024"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.0.min_part_size_ratio", "0.5"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.0.level", "5"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.#", "1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.name", "rollup1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.path_column_name", "path1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.time_column_name", "time1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.value_column_name", "value1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.version_column_name", "version1"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.#", "1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.0.regexp", "abc"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.0.function", "func1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.0.retention.#", "1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.0.retention.0.age", "1000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.0.retention.0.precision", "3"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_masking_rules.#", "1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_masking_rules.0.name", "name1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_masking_rules.0.regexp", "regexp1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_masking_rules.0.replace", "replace1"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_cache.max_size_in_bytes", "1073741820"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_cache.max_entries", "1020"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_cache.max_entry_size_in_bytes", "1048570"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_cache.max_entry_size_in_rows", "20000000"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.jdbc_bridge.host", "127.0.0.2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.jdbc_bridge.port", "8999"),

					testAccCheckCreatedAtAttr(chResource)),
			},
			mdbClickHouseClusterImportStep(chResource),
			{
				Config: testAccMDBClickHouseCluster_clickhouse_config(chName, configForSecondStep),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &r, 1),
					resource.TestCheckResourceAttr(chResource, "name", chName),
					resource.TestCheckResourceAttr(chResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResource, "version", chVersion),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.log_level", "WARNING"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.max_connections", "1024"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.max_concurrent_queries", "200"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.keep_alive_timeout", "246000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.uncompressed_cache_size", "16192"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.max_table_size_to_drop", "2048"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.max_partition_size_to_drop", "2048"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.timezone", "UTC"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.geobase_uri", ""),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.geobase_enabled", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_log_retention_size", "2001"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_thread_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_thread_log_retention_size", "2002"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_thread_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.part_log_retention_size", "2003"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.part_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.metric_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.metric_log_retention_size", "2004"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.metric_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.trace_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.trace_log_retention_size", "2005"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.trace_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.text_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.text_log_retention_size", "2006"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.text_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.opentelemetry_span_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.opentelemetry_span_log_retention_size", "2007"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.opentelemetry_span_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_views_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_views_log_retention_size", "2008"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_views_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_metric_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_metric_log_retention_size", "2009"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_metric_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.session_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.session_log_retention_size", "2010"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.session_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.zookeeper_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.zookeeper_log_retention_size", "2011"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.zookeeper_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_insert_log_enabled", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_insert_log_retention_size", "2012"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.asynchronous_insert_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.text_log_level", "ERROR"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_pool_size", "32"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_schedule_pool_size", "64"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_fetches_pool_size", "16"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_move_pool_size", "16"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_distributed_schedule_pool_size", "16"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_buffer_flush_schedule_pool_size", "16"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_common_pool_size", "16"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_message_broker_schedule_pool_size", "17"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.background_merges_mutations_concurrency_ratio", "4"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.total_memory_profiler_step", "4194303"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.dictionaries_lazy_load", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.processors_profile_log_enabled", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.processors_profile_log_retention_size", "2013"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.processors_profile_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.error_log_enabled", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.error_log_retention_size", "2014"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.error_log_retention_time", "86400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_metric_log_enabled", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_metric_log_retention_size", "2015"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_metric_log_retention_time", "80000000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.total_memory_tracker_sample_probability", "0.321"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.async_insert_threads", "8"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.backup_threads", "4"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.restore_threads", "6"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.mysql_protocol", "false"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.access_control_improvements.select_from_system_db_requires_grant", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.access_control_improvements.select_from_information_schema_requires_grant", "true"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.custom_macros.#", "2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.custom_macros.0.name", "macro1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.custom_macros.0.value", "value1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.custom_macros.1.name", "macro2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.custom_macros.1.value", "value2"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.replicated_deduplication_window", "100"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.replicated_deduplication_window_seconds", "604800"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.parts_to_delay_insert", "150"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.parts_to_throw_insert", "12000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.inactive_parts_to_delay_insert", "102"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.inactive_parts_to_throw_insert", "120"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_replicated_merges_in_queue", "16"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.number_of_free_entries_in_pool_to_lower_max_size_of_merge", "8"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_bytes_to_merge_at_min_space_in_pool", "1048576"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_bytes_to_merge_at_max_space_in_pool", "16106127301"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_bytes_for_wide_part", "512"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_rows_for_wide_part", "16"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.ttl_only_drop_parts", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.merge_with_ttl_timeout", "200010"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.merge_with_recompression_ttl_timeout", "200012"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_parts_in_total", "200014"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_number_of_merges_with_ttl_in_pool", "2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.cleanup_delay_period", "240"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.number_of_free_entries_in_pool_to_execute_mutation", "40"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_avg_part_size_for_too_many_parts", "200018"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_age_to_force_merge_seconds", "200020"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_age_to_force_merge_on_partition_only", "true"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.merge_selecting_sleep_ms", "5002"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.merge_max_block_size", "5002"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.check_sample_column_is_correct", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_merge_selecting_sleep_ms", "100001"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.max_cleanup_delay_period", "401"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.deduplicate_merge_projection_mode", "DEDUPLICATE_MERGE_PROJECTION_MODE_REBUILD"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.lightweight_mutation_projection_mode", "LIGHTWEIGHT_MUTATION_PROJECTION_MODE_DROP"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.materialize_ttl_recalculate_only", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.fsync_after_insert", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.fsync_part_directory", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_compressed_bytes_to_fsync_after_fetch", "4096"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_compressed_bytes_to_fsync_after_merge", "8192"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.merge_tree.min_rows_to_fsync_after_merge", "64"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.security_protocol", "SECURITY_PROTOCOL_PLAINTEXT"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.sasl_mechanism", "SASL_MECHANISM_GSSAPI"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.sasl_username", "user1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.sasl_password", "pass1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.debug", "DEBUG_METADATA"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.auto_offset_reset", "AUTO_OFFSET_RESET_LARGEST"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.enable_ssl_certificate_verification", "false"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.max_poll_interval_ms", "400000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.kafka.session_timeout_ms", "60000"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.rabbitmq.username", "rabbit_user"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.rabbitmq.password", "rabbit_pass2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.rabbitmq.vhost", "clickhouse"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.#", "2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.0.method", "LZ4HC"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.0.min_part_size", "2024"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.0.min_part_size_ratio", "0.3"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.0.level", "7"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.1.method", "ZSTD"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.1.min_part_size", "4048"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.1.min_part_size_ratio", "0.77"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.compression.1.level", "3"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.#", "2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.name", "rollup1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.path_column_name", "path1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.time_column_name", "time1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.value_column_name", "value1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.version_column_name", "version1"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.#", "1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.0.regexp", "abc"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.0.function", "func1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.0.retention.#", "1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.0.retention.0.age", "1000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.0.patterns.0.retention.0.precision", "3"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.1.name", "rollup2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.1.path_column_name", "path2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.1.time_column_name", "time2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.1.value_column_name", "value2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.1.version_column_name", "version2"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.1.patterns.#", "1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.1.patterns.0.regexp", "abc"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.1.patterns.0.function", "func3"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.1.patterns.0.retention.#", "1"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.1.patterns.0.retention.0.age", "3000"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.graphite_rollup.1.patterns.0.retention.0.precision", "7"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_masking_rules.#", "2"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_masking_rules.0.name", "name11"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_masking_rules.0.regexp", "regexp11"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_masking_rules.0.replace", "replace11"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_masking_rules.1.regexp", "regexp22"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_cache.max_size_in_bytes", "2073741820"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_cache.max_entries", "2020"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_cache.max_entry_size_in_bytes", "2048570"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.query_cache.max_entry_size_in_rows", "30000000"),

					resource.TestCheckResourceAttr(chResource, "clickhouse.config.jdbc_bridge.host", "127.0.0.3"),
					resource.TestCheckResourceAttr(chResource, "clickhouse.config.jdbc_bridge.port", "8998"),

					testAccCheckCreatedAtAttr(chResource)),
			},
			mdbClickHouseClusterImportStep(chResource),
		},
	})
}

/**
* Test that a sharded ClickHouse Cluster can be created, updated and destroyed.
* Also it checks changes shard's configuration.
 */
func TestAccMDBClickHouseCluster_sharded(t *testing.T) {
	t.Parallel()

	var cluster clickhouse.Cluster
	chName := acctest.RandomWithPrefix("tf-clickhouse-sharded")
	folderID := test.GetExampleFolderID()

	shardsWithShardGroupsFirstStep := `
hosts = {
	"h1" = {
		type      = "CLICKHOUSE"
		zone      = "ru-central1-a"
		subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
		shard_name = "shard1"
		assign_public_ip = false
	}
	"h2" = {
		type      = "CLICKHOUSE"
		zone      = "ru-central1-b"
		subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-b.id}"
		shard_name = "shard2"
		assign_public_ip = false
	}
}

shards = {
	shard1 = {
		weight = 11
	}
	shard2 = {
		weight = 22
	}
}

shard_group {
	name = "test_group"
	description = "test shard group"
	shard_names = [
		"shard1",
		"shard2",
	]
}

shard_group	{
	name = "test_group_2"
	description = "shard group to delete"
	shard_names = [
		"shard1",
	]
}
`

	firstShardResourcesSecondStep := &clickhouse.Resources{
		ResourcePresetId: "s2.small",
		DiskTypeId:       "network-ssd",
		DiskSize:         21474836480,
	}
	shardsWithShardGroupsSecondStep := fmt.Sprintf(`
hosts = {
	"h1" = {
		type      = "CLICKHOUSE"
		zone      = "ru-central1-a"
		subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
		shard_name = "shard1"
		assign_public_ip = true
	}
	"h3" = {
		type      = "CLICKHOUSE"
		zone      = "ru-central1-d"
		subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-d.id}"
		shard_name = "shard3"
		assign_public_ip = true
	}
}

shards = {
	shard1 = {
		weight = 110
		# resources
		%s
	}
	shard3 = {
		weight = 330
	}
}

shard_group {
	name = "test_group"
	description = "updated shard group"
	shard_names = [
		"shard1",
		"shard3",
	]
}
shard_group {
	name = "test_group_3"
	description = "new shard group"
	shard_names = [
		"shard1",
	]
}
`,
		buildResourcesHCL(firstShardResourcesSecondStep),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			// Create sharded ClickHouse Cluster
			{
				Config: testAccMDBClickHouseCluster_sharded(chName, shardsWithShardGroupsFirstStep),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResourceSharded, &cluster, 2),
					resource.TestCheckResourceAttr(chResourceSharded, "name", chName),
					resource.TestCheckResourceAttr(chResourceSharded, "folder_id", folderID),

					resource.TestCheckResourceAttr(chResourceSharded, "shards.shard1.weight", "11"),
					resource.TestCheckResourceAttr(chResourceSharded, "shards.shard2.weight", "22"),

					resource.TestCheckResourceAttrSet(chResourceSharded, "hosts.h1.fqdn"),
					resource.TestCheckResourceAttr(chResourceSharded, "hosts.h1.assign_public_ip", "false"),
					resource.TestCheckResourceAttrSet(chResourceSharded, "hosts.h2.fqdn"),
					resource.TestCheckResourceAttr(chResourceSharded, "hosts.h2.assign_public_ip", "false"),

					testAccCheckMDBClickHouseClusterHasShards(&cluster, []string{"shard1", "shard2"}),
					testAccCheckMDBClickHouseClusterHasShardGroups(&cluster, map[string][]string{
						"test_group":   {"shard1", "shard2"},
						"test_group_2": {"shard1"},
					}),
					testAccCheckCreatedAtAttr(chResourceSharded),
				),
			},
			mdbClickHouseClusterImportStep(chResourceSharded),
			// Add new shard, delete old shard
			{
				Config: testAccMDBClickHouseCluster_sharded(chName, shardsWithShardGroupsSecondStep),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResourceSharded, &cluster, 2),
					resource.TestCheckResourceAttr(chResourceSharded, "name", chName),
					resource.TestCheckResourceAttr(chResourceSharded, "folder_id", folderID),

					resource.TestCheckResourceAttr(chResourceSharded, "shards.shard1.weight", "110"),
					resource.TestCheckResourceAttr(chResourceSharded, "shards.shard3.weight", "330"),

					resource.TestCheckResourceAttrSet(chResourceSharded, "hosts.h1.fqdn"),
					resource.TestCheckResourceAttr(chResourceSharded, "hosts.h1.assign_public_ip", "true"),
					resource.TestCheckResourceAttrSet(chResourceSharded, "hosts.h3.fqdn"),
					resource.TestCheckResourceAttr(chResourceSharded, "hosts.h3.assign_public_ip", "true"),

					testAccCheckMDBClickHouseClusterHasResources(&cluster, firstShardResourcesSecondStep.ResourcePresetId, firstShardResourcesSecondStep.DiskTypeId, firstShardResourcesSecondStep.DiskSize),

					testAccCheckMDBClickHouseClusterHasShards(&cluster, []string{"shard1", "shard3"}),
					testAccCheckMDBClickHouseClusterHasShardGroups(&cluster, map[string][]string{
						"test_group":   {"shard1", "shard3"},
						"test_group_3": {"shard1"},
					}),
					testAccCheckCreatedAtAttr(chResourceSharded),
				),
			},
			mdbClickHouseClusterImportStep(chResourceSharded),
		},
	})
}

// Test that a Keeper-based ClickHouse Cluster can be created and destroyed
func TestAccMDBClickHouseCluster_keeper(t *testing.T) {
	t.Parallel()

	var r clickhouse.Cluster
	chName := acctest.RandomWithPrefix("tf-clickhouse-keeper")
	chDesc := "ClickHouse Cluster Keeper Test"
	folderID := test.GetExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			// Enable embedded_keeper
			{
				Config: testAccMDBClickHouseCluster_keeper(chName, chDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResourceKeeper, &r, 1),
					resource.TestCheckResourceAttr(chResourceKeeper, "name", chName),
					resource.TestCheckResourceAttr(chResourceKeeper, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResourceKeeper, "description", chDesc),
					resource.TestCheckResourceAttrSet(chResourceKeeper, "hosts.ha.fqdn"),
					testAccCheckCreatedAtAttr(chResourceKeeper)),
			},
			mdbClickHouseClusterImportStep(chResourceKeeper),
		},
	})
}

// Test that a ClickHouse Cluster with cloud storage can be created
func TestAccMDBClickHouseCluster_cloud_storage(t *testing.T) {
	t.Parallel()

	var r clickhouse.Cluster
	chName := acctest.RandomWithPrefix("tf-clickhouse-cloud-storage")
	chDesc := "ClickHouse Cloud Storage Cluster Terraform Test"
	folderID := test.GetExampleFolderID()
	bucketName := acctest.RandomWithPrefix("tf-test-clickhouse-bucket")
	rInt := acctest.RandInt()

	firstCloudStorage := `
cloud_storage = {
	enabled 			= false
	move_factor 		= 0.000000
	data_cache_enabled  = false
	data_cache_max_size = 0
	prefer_not_to_merge = false
}
`

	secondCloudStorage := `
cloud_storage = {
	enabled 			= true
	move_factor 		= 0.500000
	data_cache_enabled  = true
	data_cache_max_size = 214748364
	prefer_not_to_merge = true
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseClusterDestroy,
		Steps: []resource.TestStep{
			// Create ClickHouse Cluster with cloud storage
			{
				Config: testAccMDBClickHouseCluster_cloud_storage(chName, chDesc, bucketName, "", rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResourceCloudStorage, &r, 1),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "name", chName),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "description", chDesc),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "cloud_storage.enabled", "false"),
					testAccCheckCreatedAtAttr(chResourceCloudStorage)),
			},
			mdbClickHouseClusterImportStep(chResourceCloudStorage),
			// Update ClickHouse Cluster with cloud storage
			{
				Config: testAccMDBClickHouseCluster_cloud_storage(chName, chDesc, bucketName, firstCloudStorage, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResourceCloudStorage, &r, 1),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "name", chName),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "description", chDesc),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "cloud_storage.enabled", "false"),
					testAccCheckCreatedAtAttr(chResourceCloudStorage)),
			},
			mdbClickHouseClusterImportStep(chResourceCloudStorage),
			// Update ClickHouse Cluster with cloud storage with all params
			{
				Config: testAccMDBClickHouseCluster_cloud_storage(chName, chDesc, bucketName, secondCloudStorage, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResourceCloudStorage, &r, 1),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "name", chName),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "folder_id", folderID),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "description", chDesc),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "cloud_storage.enabled", "true"),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "cloud_storage.move_factor", "0.5"),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "cloud_storage.data_cache_enabled", "true"),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "cloud_storage.data_cache_max_size", "214748364"),
					resource.TestCheckResourceAttr(chResourceCloudStorage, "cloud_storage.prefer_not_to_merge", "true"),
					testAccCheckCreatedAtAttr(chResourceCloudStorage)),
			},
			mdbClickHouseClusterImportStep(chResourceCloudStorage),
		},
	})
}

func TestAccMDBClickHouseCluster_encrypted_disk(t *testing.T) {
	t.Parallel()

	var r clickhouse.Cluster
	chName := acctest.RandomWithPrefix("tf-clickhouse-disk-encryption-create")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             resource.ComposeTestCheckFunc(testAccCheckMDBClickHouseClusterDestroy, kms_symmetric_key.TestAccCheckYandexKmsSymmetricKeyAllDestroyed),
		Steps: []resource.TestStep{
			// Create ClickHouse Cluster with disk encryption
			{
				Config: testAccMDBClickHouseCluster_encrypted_disk(chName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterExists(chResource, &r, 1),
					resource.TestCheckResourceAttrSet(chResource, "disk_encryption_key_id"),
				),
			},
			mdbClickHouseClusterImportStep(chResource),
		},
	})
}

// Test HCL configs

func testAccMDBClickHouseCluster_basic(name string, bucket string, randInt int, changeableConf string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+"\n"+clickhouseObjectStorageDependencies(bucket, randInt)+"\n"+`
resource "yandex_mdb_clickhouse_cluster_v2" "foo" {
  name           	  = "%s"
  description    	  = "ClickHouse basic tests"
  network_id     	  = "${yandex_vpc_network.mdb-ch-test-net.id}"
  admin_password      = "strong_password"
  deletion_protection = false

  labels = {
    test_key = "test_value"
  }

  service_account_id = "${yandex_iam_service_account.sa.id}"

  version = "%s"

  hosts = {
    "ha" = {
	  type      = "CLICKHOUSE"
	  zone      = "ru-central1-a"
	  subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
    }
  }

  # changeable config
  %s
}
`,
		name,
		chVersion,
		changeableConf,
	)
}

func testAccMDBClickHouseCluster_resources(name, firstShardName, version string, clickHouseResources, zookeeperResources, shardResources *clickhouse.Resources) string {
	return fmt.Sprintf(clickHouseVPCDependencies+"\n"+`
resource "yandex_mdb_clickhouse_cluster_v2" "foo" {
  name           = "%s"
  description    = "Cluster resources ans version"
  environment    = "PRESTABLE"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"

  version = "%s"
  clickhouse = {
	  # resources
	  %s
  }
	
  zookeeper = {
	# resources
	%s
  }

  hosts = {
    "za" = {
	  type      = "KEEPER"
	  zone      = "ru-central1-a"
	  subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
    }
	"zb" = {
	  type      = "KEEPER"
	  zone      = "ru-central1-b"
	  subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-b.id}"
    }
	"zd" = {
	  type      = "KEEPER"
	  zone      = "ru-central1-d"
	  subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-d.id}"
    }
    "ha" = {
	  type       = "CLICKHOUSE"
	  zone       = "ru-central1-a"
	  subnet_id  = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
	  shard_name = "shard1"
    }
  }

  shards = {
	%s = {
		weight = 11
		# resources
		%s
	}
  }

  # maintenance_window
  %s
}
`,
		name,
		version,
		buildResourcesHCL(clickHouseResources),
		buildResourcesHCL(zookeeperResources),
		firstShardName,
		buildResourcesHCL(shardResources),
		maintenanceWindowAnytime,
	)
}

func testAccMDBClickHouseCluster_clickhouse_config(name string, config *clickhouseConfig.ClickhouseConfig) string {
	return fmt.Sprintf(clickHouseVPCDependencies+"\n"+`
resource "yandex_mdb_clickhouse_cluster_v2" "foo" {
  name           = "%s"
  description    = "ClickHouse config"
  environment    = "PRESTABLE"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"

  version = "%s"
  clickhouse = {
	# clickhouse config
	%s

	resources = {
		resource_preset_id = "s2.micro"
		disk_type_id       = "network-ssd"
		disk_size          = 10
	}
  }

  hosts = {
    "ha" = {
	  type      = "CLICKHOUSE"
	  zone      = "ru-central1-a"
	  subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
    }
  }

  # maintenance_window
  %s

  # deletion_protection = true
}
`,
		name,
		chVersion,
		buildClickhouseConfigHCL(config),
		maintenanceWindowAnytime,
	)
}

func testAccMDBClickHouseCluster_sharded(name, shards string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+"\n"+`
resource "yandex_mdb_clickhouse_cluster_v2" "bar" {
  name           = "%s"
  description    = "ClickHouse Sharded Cluster Terraform Test"
  environment    = "PRESTABLE"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"

  # hosts, shards, shard groups
  %s

  # maintenance_window
  %s
}
`,
		name,
		shards,
		maintenanceWindowAnytime,
	)
}

func testAccMDBClickHouseCluster_keeper(name, desc string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+"\n"+`
resource "yandex_mdb_clickhouse_cluster_v2" "keeper" {
  name           = "%s"
  description    = "%s"
  environment    = "PRESTABLE"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
  embedded_keeper = true

  clickhouse = {
	  resources = {
		resource_preset_id = "s2.micro"
		disk_type_id       = "network-ssd"
		disk_size          = 10
	  }
  }

  hosts = {
    "ha" = {
	  type      = "CLICKHOUSE"
	  zone      = "ru-central1-a"
	  subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
    }
  }

  # maintenance_window
  %s
}
`,
		name,
		desc,
		maintenanceWindowAnytime,
	)
}

func testAccMDBClickHouseCluster_cloud_storage(name, desc, bucket, cloudStorage string, randInt int) string {
	return fmt.Sprintf(clickHouseVPCDependencies+"\n"+clickhouseObjectStorageDependencies(bucket, randInt)+"\n"+`
resource "yandex_mdb_clickhouse_cluster_v2" "cloud" {
  name           = "%s"
  description    = "%s"
  environment    = "PRESTABLE"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"

  clickhouse = {
	  resources = {
		resource_preset_id = "s2.micro"
		disk_type_id       = "network-ssd"
		disk_size          = 10
	  }
  }

  hosts = {
    "ha" = {
	  type      = "CLICKHOUSE"
	  zone      = "ru-central1-a"
	  subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
    }
  }

  # maintenance_window
  %s

  # cloud_storage
  %s
}
`,
		name,
		desc,
		maintenanceWindowAnytime,
		cloudStorage,
	)
}

func testAccMDBClickHouseCluster_encrypted_disk(name string) string {
	return fmt.Sprintf(clickHouseVPCDependencies+"\n"+`
resource "yandex_kms_symmetric_key" "disk_encrypt" {}

resource "yandex_mdb_clickhouse_cluster_v2" "foo" {
  name           = "%s"
  description    = "Encrypted cluster"
  environment    = "PRESTABLE"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"

  clickhouse = {
	  resources = {
		resource_preset_id = "s2.micro"
		disk_type_id       = "network-ssd"
		disk_size          = 10
	  }
  }

  hosts = {
    "ha" = {
	  type      = "CLICKHOUSE"
	  zone      = "ru-central1-a"
	  subnet_id = "${yandex_vpc_subnet.mdb-ch-test-subnet-a.id}"
    }
  }

  # maintenance_window
  %s

  disk_encryption_key_id = "${yandex_kms_symmetric_key.disk_encrypt.id}"
}
`,
		name,
		maintenanceWindowAnytime,
	)
}

// Utils

func testAccCheckMDBClickHouseClusterHasResources(r *clickhouse.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.Clickhouse.Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("Expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskTypeId != diskType {
			return fmt.Errorf("Expected disk type '%s', got '%s'", diskType, rs.DiskTypeId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("Expected disk size '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

func testAccCheckMDBClickHouseShardHasResources(r *clickhouse.Cluster, shardName string, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*provider.Provider).GetConfig()

		shard, err := config.SDK.MDB().Clickhouse().Cluster().GetShard(context.Background(), &clickhouse.GetClusterShardRequest{
			ClusterId: r.Id,
			ShardName: shardName,
		})
		if err != nil {
			return err
		}

		shardResources := shard.Config.Clickhouse.Resources
		if shardResources.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("Expected resource preset id '%s', got '%s'", resourcePresetID, shardResources.ResourcePresetId)
		}
		if shardResources.DiskTypeId != diskType {
			return fmt.Errorf("Expected disk type '%s', got '%s'", diskType, shardResources.DiskTypeId)
		}
		if shardResources.DiskSize != diskSize {
			return fmt.Errorf("Expected disk size '%d', got '%d'", diskSize, shardResources.DiskSize)
		}
		return nil
	}
}

func testAccCheckMDBClickHouseZooKeeperSubclusterHasResources(r *clickhouse.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.Zookeeper.Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("Expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskTypeId != diskType {
			return fmt.Errorf("Expected disk type '%s', got '%s'", diskType, rs.DiskTypeId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("Expected disk size '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

func testAccCheckMDBClickHouseClusterHasShards(r *clickhouse.Cluster, shards []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*provider.Provider).GetConfig()

		resp, err := config.SDK.MDB().Clickhouse().Cluster().ListShards(context.Background(), &clickhouse.ListClusterShardsRequest{
			ClusterId: r.Id,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Shards) != len(shards) {
			return fmt.Errorf("Expected %d shards, got %d", len(shards), len(resp.Shards))
		}
		for _, s := range shards {
			found := false
			for _, rs := range resp.Shards {
				if s == rs.Name {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("Shard '%s' not found", s)
			}
		}
		return nil
	}
}

func testAccCheckMDBClickHouseClusterHasShardGroups(r *clickhouse.Cluster, shardGroups map[string][]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*provider.Provider).GetConfig()

		resp, err := config.SDK.MDB().Clickhouse().Cluster().ListShardGroups(context.Background(), &clickhouse.ListClusterShardGroupsRequest{
			ClusterId: r.Id,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.ShardGroups) != len(shardGroups) {
			return fmt.Errorf("Expected %d shard groups, got %d", len(shardGroups), len(resp.ShardGroups))
		}
		for name, shards := range shardGroups {
			found := false
			for _, rs := range resp.ShardGroups {
				if name == rs.Name {
					found = true
					if !reflect.DeepEqual(shards, rs.ShardNames) {
						return fmt.Errorf("Shards in group %s not match, expexted %s, got %s", name, fmt.Sprint(shards), fmt.Sprint(rs.ShardNames))
					}
				}
			}
			if !found {
				return fmt.Errorf("Shard group '%s' not found", s)
			}
		}
		return nil
	}
}

func testAccCheckCreatedAtAttr(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		const createdAtAttrName = "created_at"
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		createdAt, ok := rs.Primary.Attributes[createdAtAttrName]
		if !ok {
			return fmt.Errorf("can't find '%s' attr for %s resource", createdAtAttrName, resourceName)
		}

		if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
			return fmt.Errorf("can't parse timestamp in attr '%s': %s", createdAtAttrName, createdAt)
		}
		return nil
	}
}

func testAccCheckMDBClickHouseClusterContainsLabel(r *clickhouse.Cluster, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := r.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccCheckMDBClickHouseClusterExists(n string, r *clickhouse.Cluster, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*provider.Provider).GetConfig()

		found, err := config.SDK.MDB().Clickhouse().Cluster().Get(context.Background(), &clickhouse.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("ClickHouse Cluster not found")
		}

		*r = *found

		resp, err := config.SDK.MDB().Clickhouse().Cluster().ListHosts(context.Background(), &clickhouse.ListClusterHostsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Hosts) != hosts {
			return fmt.Errorf("Expected %d hosts, got %d", hosts, len(resp.Hosts))
		}

		return nil
	}
}

func testAccCheckMDBClickHouseClusterDestroy(s *terraform.State) error {
	config := test.AccProvider.(*provider.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_clickhouse_cluster_v2" {
			continue
		}

		_, err := config.SDK.MDB().Clickhouse().Cluster().Get(context.Background(), &clickhouse.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("ClickHouse Cluster still exists")
		}
	}

	return nil
}

func testAccCheckMDBClickHouseClusterHasFormatSchemas(r string, targetSchemas map[string]map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*provider.Provider).GetConfig()

		resp, err := config.SDK.MDB().Clickhouse().FormatSchema().List(context.Background(), &clickhouse.ListFormatSchemasRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}
		formatSchemas := resp.FormatSchemas

		if len(formatSchemas) != len(targetSchemas) {
			return fmt.Errorf("expected %d format schemas, found %d", len(formatSchemas), len(targetSchemas))
		}

		for _, s := range formatSchemas {
			ts, ok := targetSchemas[s.Name]
			if !ok {
				return fmt.Errorf("unexpected format schema: %s", s.Name)
			}

			if s.Type.String() != ts["type"] {
				return fmt.Errorf("format schema %s has wrong type, %v. expected %v", s.Name, s.Type.String(), ts["type"])
			}

			if s.Uri != ts["uri"] {
				return fmt.Errorf("format schema %s has wrong uri, %v. expected %v", s.Name, s.Uri, ts["uri"])
			}
		}

		return nil
	}
}

// Steps

func mdbClickHouseClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"user",                       // passwords are not returned
			"host",                       // zookeeper hosts are not imported by default
			"zookeeper",                  // zookeeper spec is not imported by default
			"health",                     // volatile value
			"copy_schema_on_new_hosts",   // special parameter
			"admin_password",             // passwords are not returned
			"clickhouse.config.kafka",    // passwords are not returned
			"clickhouse.config.rabbitmq", // passwords are not returned
		},
	}
}

// Sweep logic

func testSweepMDBClickHouseCluster(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.SDK.MDB().Clickhouse().Cluster().List(context.Background(), &clickhouse.ListClustersRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting ClickHouse clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBClickHouseCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep ClickHouse cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBClickHouseCluster(conf *config.Config, id string) bool {
	return test.SweepWithRetry(sweepMDBClickHouseClusterOnce, conf, "ClickHouse cluster", id)
}

func sweepMDBClickHouseClusterOnce(conf *config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexMDBClickHouseClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.SDK.MDB().Clickhouse().Cluster().Update(ctx, &clickhouse.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = test.HandleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(test.ErrorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.SDK.MDB().Clickhouse().Cluster().Delete(ctx, &clickhouse.DeleteClusterRequest{
		ClusterId: id,
	})
	return test.HandleSweepOperation(ctx, conf, op, err)
}

// Build HCL functions

const clickHouseVPCDependencies = `
resource "yandex_vpc_network" "mdb-ch-test-net" {}

resource "yandex_vpc_subnet" "mdb-ch-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-ch-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-ch-test-subnet-d" {
  zone           = "ru-central1-d"
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_security_group" "mdb-ch-test-sg-x" {
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
  ingress {
    protocol          = "ANY"
    description       = "Allow incoming traffic from members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
  egress {
    protocol          = "ANY"
    description       = "Allow outgoing traffic to members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
}

resource "yandex_vpc_security_group" "mdb-ch-test-sg-y" {
  network_id     = "${yandex_vpc_network.mdb-ch-test-net.id}"
  
  ingress {
    protocol          = "ANY"
    description       = "Allow incoming traffic from members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
  egress {
    protocol          = "ANY"
    description       = "Allow outgoing traffic to members of the same security group"
    from_port         = 0
    to_port           = 65535
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }
}
`

func clickhouseObjectStorageDependencies(bucket string, randInt int) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "sa" {
	name = "test-sa-for-tf-test-%[1]d"
}

resource "yandex_resourcemanager_folder_iam_member" "binding" {
	folder_id   = "%[2]s"
	member      = "serviceAccount:${yandex_iam_service_account.sa.id}"
	role        = "storage.admin"
	sleep_after = 30
}

resource "yandex_iam_service_account_static_access_key" "sa-key" {
	service_account_id = "${yandex_iam_service_account.sa.id}"

	depends_on = [
		yandex_resourcemanager_folder_iam_member.binding
	]
}

resource "yandex_storage_bucket" "tmp_bucket" {
  bucket = "%[3]s"
  acl    = "public-read"

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key

  force_destroy = true
}

resource "yandex_storage_object" "test_capnp" {
  bucket = yandex_storage_bucket.tmp_bucket.bucket

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key

  key     = "test.capnp"
  content = "# This is a comment."

  acl = "public-read"

  depends_on = [
    yandex_storage_bucket.tmp_bucket
  ]
}

resource "yandex_storage_object" "test_capnp2" {
  bucket = yandex_storage_bucket.tmp_bucket.bucket

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key

  key     = "test2.capnp"
  content = "# This is a comment."
}

resource "yandex_storage_object" "test_proto" {
  bucket = yandex_storage_bucket.tmp_bucket.bucket

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key

  key     = "test.proto"
  content = "# This is a comment."
}

resource "yandex_storage_object" "test_ml_model" {
  bucket = yandex_storage_bucket.tmp_bucket.bucket

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key

  key     = "model.bin"
  content_base64 = <<EOT
Q0JNMUgBAAAMAAAACAAMAAQACAAIAAAACAAAAEgAAAASAAAARmxhYnVmZmVyc01vZGVsX3YxAAA
AACoASAAEAAgADAAQABQAGAAcACAAJAAoACwAMAA0ADgAAAAAADwAQABEACoAAAABAAAAjAAAAI
AAAAB0AAAA1AAAAKQAAACQAAAAiAAAAEwAAAAwAAAAeAAAACQAAACEAAAAeAAAAAwAAABcAAAAc
AAAAAEAAAAAAAAAAADgPwAAAAACAAAAAAAAAAAAJEAAAAAAAAAkQAAAAAACAAAA2Ymd2ImdyL/Z
iZ3YiZ3IPwEAAAAAAAAAAQAAAAEAAAABAAAAAAAAAAEAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAQAAABAAAAAMABAAAAAEAAgADAAMAAAAAAAAAAAAAAAEAAAAAQAAAAAAAD8AAAAA
EOT
}
`, randInt, test.GetExampleFolderID(), bucket)
}

func buildResourcesHCL(resources *clickhouse.Resources) string {
	if resources == nil {
		return ""
	}

	return fmt.Sprintf(`
resources = {
	resource_preset_id = "%s"
	disk_type_id       = "%s"
	disk_size          = %d
}
`,
		resources.ResourcePresetId,
		resources.DiskTypeId,
		utils.ToGigabytes(resources.DiskSize))
}

func buildClickhouseConfigHCL(config *clickhouseConfig.ClickhouseConfig) string {
	if config == nil {
		return ""
	}

	return fmt.Sprintf(`
config = {
	log_level		                        	  = "%s"
	max_connections                         	  = %d
	max_concurrent_queries                  	  = %d
	keep_alive_timeout                      	  = %d
	uncompressed_cache_size                 	  = %d
	max_table_size_to_drop                  	  = %d
	max_partition_size_to_drop              	  = %d
	timezone                                	  = "%s"
	geobase_uri                             	  = "%s"
	geobase_enabled                         	  = %t
	query_log_retention_size                	  = %d
	query_log_retention_time                	  = %d
	query_thread_log_enabled                	  = %t
	query_thread_log_retention_size         	  = %d
	query_thread_log_retention_time         	  = %d
	part_log_retention_size                 	  = %d
	part_log_retention_time                 	  = %d
	metric_log_enabled                      	  = %t
	metric_log_retention_size               	  = %d
	metric_log_retention_time               	  = %d
	trace_log_enabled                       	  = %t
	trace_log_retention_size                	  = %d
	trace_log_retention_time                	  = %d
	text_log_enabled                        	  = %t
	text_log_retention_size                 	  = %d
	text_log_retention_time                 	  = %d
	opentelemetry_span_log_enabled          	  = %t
	opentelemetry_span_log_retention_size   	  = %d
	opentelemetry_span_log_retention_time   	  = %d
	query_views_log_enabled                 	  = %t
	query_views_log_retention_size          	  = %d
	query_views_log_retention_time          	  = %d
	asynchronous_metric_log_enabled         	  = %t
	asynchronous_metric_log_retention_size  	  = %d
	asynchronous_metric_log_retention_time  	  = %d
	session_log_enabled                     	  = %t
	session_log_retention_size              	  = %d
	session_log_retention_time              	  = %d
	zookeeper_log_enabled                   	  = %t
	zookeeper_log_retention_size            	  = %d
	zookeeper_log_retention_time            	  = %d
	asynchronous_insert_log_enabled         	  = %t
	asynchronous_insert_log_retention_size  	  = %d
	asynchronous_insert_log_retention_time  	  = %d
	processors_profile_log_enabled          	  = %t
	processors_profile_log_retention_size   	  = %d
	processors_profile_log_retention_time   	  = %d
	error_log_enabled                       	  = %t
	error_log_retention_size                	  = %d
	error_log_retention_time                	  = %d
	query_metric_log_enabled                      = %t
	query_metric_log_retention_size               = %d
	query_metric_log_retention_time               = %d
	text_log_level                          	  = "%s"
	background_pool_size                    	  = %d
	background_schedule_pool_size           	  = %d
	background_fetches_pool_size            	  = %d
	background_move_pool_size               	  = %d
	background_distributed_schedule_pool_size     = %d
	background_buffer_flush_schedule_pool_size    = %d
	background_common_pool_size             	  = %d
	background_message_broker_schedule_pool_size  = %d
	background_merges_mutations_concurrency_ratio = %d
	# default_database                        	  = "..."
	total_memory_profiler_step              	  = %d
	total_memory_tracker_sample_probability       = %f
	dictionaries_lazy_load                  	  = %t
	async_insert_threads                   	 	  = %d
	backup_threads                          	  = %d
	restore_threads                         	  = %d
	mysql_protocol                          	  = %t

	# access_control_improvements
	%s

	# merge_tree
	%s

	# kafka
	%s

	# rabbitmq
	%s

	# compression
	%s

	# graphite_rollup
	%s

	# query_masking_rules
	%s

	# query_cache
	%s

	# jdbc_bridge
	%s

	# custom_macros
	%s
}
`,
		config.LogLevel.String(),
		config.MaxConnections.GetValue(),
		config.MaxConcurrentQueries.GetValue(),
		config.KeepAliveTimeout.GetValue(),
		config.UncompressedCacheSize.GetValue(),
		config.MaxTableSizeToDrop.GetValue(),
		config.MaxPartitionSizeToDrop.GetValue(),
		config.Timezone,
		config.GeobaseUri,
		config.GeobaseEnabled.GetValue(),
		config.QueryLogRetentionSize.GetValue(),
		config.QueryLogRetentionTime.GetValue(),
		config.QueryThreadLogEnabled.GetValue(),
		config.QueryThreadLogRetentionSize.GetValue(),
		config.QueryThreadLogRetentionTime.GetValue(),
		config.PartLogRetentionSize.GetValue(),
		config.PartLogRetentionTime.GetValue(),
		config.MetricLogEnabled.GetValue(),
		config.MetricLogRetentionSize.GetValue(),
		config.MetricLogRetentionTime.GetValue(),
		config.TraceLogEnabled.GetValue(),
		config.TraceLogRetentionSize.GetValue(),
		config.TraceLogRetentionTime.GetValue(),
		config.TextLogEnabled.GetValue(),
		config.TextLogRetentionSize.GetValue(),
		config.TextLogRetentionTime.GetValue(),
		config.OpentelemetrySpanLogEnabled.GetValue(),
		config.OpentelemetrySpanLogRetentionSize.GetValue(),
		config.OpentelemetrySpanLogRetentionTime.GetValue(),
		config.QueryViewsLogEnabled.GetValue(),
		config.QueryViewsLogRetentionSize.GetValue(),
		config.QueryViewsLogRetentionTime.GetValue(),
		config.AsynchronousMetricLogEnabled.GetValue(),
		config.AsynchronousMetricLogRetentionSize.GetValue(),
		config.AsynchronousMetricLogRetentionTime.GetValue(),
		config.SessionLogEnabled.GetValue(),
		config.SessionLogRetentionSize.GetValue(),
		config.SessionLogRetentionTime.GetValue(),
		config.ZookeeperLogEnabled.GetValue(),
		config.ZookeeperLogRetentionSize.GetValue(),
		config.ZookeeperLogRetentionTime.GetValue(),
		config.AsynchronousInsertLogEnabled.GetValue(),
		config.AsynchronousInsertLogRetentionSize.GetValue(),
		config.AsynchronousInsertLogRetentionTime.GetValue(),
		config.ProcessorsProfileLogEnabled.GetValue(),
		config.ProcessorsProfileLogRetentionSize.GetValue(),
		config.ProcessorsProfileLogRetentionTime.GetValue(),
		config.ErrorLogEnabled.GetValue(),
		config.ErrorLogRetentionSize.GetValue(),
		config.ErrorLogRetentionTime.GetValue(),
		config.QueryMetricLogEnabled.GetValue(),
		config.QueryMetricLogRetentionSize.GetValue(),
		config.QueryMetricLogRetentionTime.GetValue(),
		config.TextLogLevel.String(),
		config.BackgroundPoolSize.GetValue(),
		config.BackgroundSchedulePoolSize.GetValue(),
		config.BackgroundFetchesPoolSize.GetValue(),
		config.BackgroundMovePoolSize.GetValue(),
		config.BackgroundDistributedSchedulePoolSize.GetValue(),
		config.BackgroundBufferFlushSchedulePoolSize.GetValue(),
		config.BackgroundCommonPoolSize.GetValue(),
		config.BackgroundMessageBrokerSchedulePoolSize.GetValue(),
		config.BackgroundMergesMutationsConcurrencyRatio.GetValue(),
		config.TotalMemoryProfilerStep.GetValue(),
		config.TotalMemoryTrackerSampleProbability.GetValue(),
		config.DictionariesLazyLoad.GetValue(),
		config.AsyncInsertThreads.GetValue(),
		config.BackupThreads.GetValue(),
		config.RestoreThreads.GetValue(),
		config.MysqlProtocol.GetValue(),
		buildAccessControlImprovementsHCL(config.AccessControlImprovements),
		buildMergeTreeHCL(config.MergeTree),
		buildKafkaHCL(config.Kafka),
		buildRabbitMqHCL(config.Rabbitmq),
		buildCompressionsHCL(config.Compression),
		buildGraphiteRollupsHCL(config.GraphiteRollup),
		buildQueryMaskingRulesHCL(config.QueryMaskingRules),
		buildQueryCacheHCL(config.QueryCache),
		buildJdbcBridgeHCL(config.JdbcBridge),
		buildCustomMacrosHCL(config.CustomMacros),
	)
}

func buildMergeTreeHCL(mergeTree *clickhouseConfig.ClickhouseConfig_MergeTree) string {
	if mergeTree == nil {
		return ""
	}

	return fmt.Sprintf(`
merge_tree = {
	replicated_deduplication_window                           = %d
	replicated_deduplication_window_seconds                   = %d
	parts_to_delay_insert                                     = %d
	parts_to_throw_insert                                     = %d
	max_replicated_merges_in_queue                            = %d
	number_of_free_entries_in_pool_to_lower_max_size_of_merge = %d
	max_bytes_to_merge_at_min_space_in_pool                   = %d
	max_bytes_to_merge_at_max_space_in_pool                   = %d
	inactive_parts_to_delay_insert                            = %d
	inactive_parts_to_throw_insert                            = %d
	min_bytes_for_wide_part                                   = %d
	min_rows_for_wide_part                                    = %d
	ttl_only_drop_parts                                       = %t
	merge_with_ttl_timeout                                    = %d
	merge_with_recompression_ttl_timeout                      = %d
	max_parts_in_total                                        = %d
	max_number_of_merges_with_ttl_in_pool                     = %d
	cleanup_delay_period                                      = %d
	number_of_free_entries_in_pool_to_execute_mutation        = %d
	max_avg_part_size_for_too_many_parts                      = %d
	min_age_to_force_merge_seconds                            = %d
	min_age_to_force_merge_on_partition_only                  = %t
	merge_selecting_sleep_ms                                  = %d
	check_sample_column_is_correct                            = %t
	merge_max_block_size                                      = %d
	max_merge_selecting_sleep_ms                              = %d
	max_cleanup_delay_period                                  = %d
	deduplicate_merge_projection_mode                         = %q
	lightweight_mutation_projection_mode                      = %q
	materialize_ttl_recalculate_only                          = %t
	fsync_after_insert                                        = %t
	fsync_part_directory                                      = %t
	min_compressed_bytes_to_fsync_after_fetch                 = %d
	min_compressed_bytes_to_fsync_after_merge                 = %d
	min_rows_to_fsync_after_merge                             = %d
}
`,
		mergeTree.ReplicatedDeduplicationWindow.GetValue(),
		mergeTree.ReplicatedDeduplicationWindowSeconds.GetValue(),
		mergeTree.PartsToDelayInsert.GetValue(),
		mergeTree.PartsToThrowInsert.GetValue(),
		mergeTree.MaxReplicatedMergesInQueue.GetValue(),
		mergeTree.NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge.GetValue(),
		mergeTree.MaxBytesToMergeAtMinSpaceInPool.GetValue(),
		mergeTree.MaxBytesToMergeAtMaxSpaceInPool.GetValue(),
		mergeTree.InactivePartsToDelayInsert.GetValue(),
		mergeTree.InactivePartsToThrowInsert.GetValue(),
		mergeTree.MinBytesForWidePart.GetValue(),
		mergeTree.MinRowsForWidePart.GetValue(),
		mergeTree.TtlOnlyDropParts.GetValue(),
		mergeTree.MergeWithTtlTimeout.GetValue(),
		mergeTree.MergeWithRecompressionTtlTimeout.GetValue(),
		mergeTree.MaxPartsInTotal.GetValue(),
		mergeTree.MaxNumberOfMergesWithTtlInPool.GetValue(),
		mergeTree.CleanupDelayPeriod.GetValue(),
		mergeTree.NumberOfFreeEntriesInPoolToExecuteMutation.GetValue(),
		mergeTree.MaxAvgPartSizeForTooManyParts.GetValue(),
		mergeTree.MinAgeToForceMergeSeconds.GetValue(),
		mergeTree.MinAgeToForceMergeOnPartitionOnly.GetValue(),
		mergeTree.MergeSelectingSleepMs.GetValue(),
		mergeTree.CheckSampleColumnIsCorrect.GetValue(),
		mergeTree.MergeMaxBlockSize.GetValue(),
		mergeTree.MaxMergeSelectingSleepMs.GetValue(),
		mergeTree.MaxCleanupDelayPeriod.GetValue(),
		mergeTree.DeduplicateMergeProjectionMode.String(),
		mergeTree.LightweightMutationProjectionMode.String(),
		mergeTree.MaterializeTtlRecalculateOnly.GetValue(),
		mergeTree.FsyncAfterInsert.GetValue(),
		mergeTree.FsyncPartDirectory.GetValue(),
		mergeTree.MinCompressedBytesToFsyncAfterFetch.GetValue(),
		mergeTree.MinCompressedBytesToFsyncAfterMerge.GetValue(),
		mergeTree.MinRowsToFsyncAfterMerge.GetValue(),
	)
}

func buildKafkaHCL(kafka *clickhouseConfig.ClickhouseConfig_Kafka) string {
	if kafka == nil {
		return ""
	}

	return fmt.Sprintf(`
kafka = {
	security_protocol                   = "%s"
	sasl_mechanism                      = "%s"
	sasl_username                       = "%s"
	sasl_password                       = "%s"
	enable_ssl_certificate_verification = %t
	max_poll_interval_ms                = %d
	session_timeout_ms                  = %d
	debug                               = "%s"
	auto_offset_reset                   = "%s"
}
`,
		kafka.SecurityProtocol.String(),
		kafka.SaslMechanism.String(),
		kafka.SaslUsername,
		kafka.SaslPassword,
		kafka.EnableSslCertificateVerification.GetValue(),
		kafka.MaxPollIntervalMs.GetValue(),
		kafka.SessionTimeoutMs.GetValue(),
		kafka.Debug.String(),
		kafka.AutoOffsetReset.String(),
	)
}

func buildRabbitMqHCL(rabbitmq *clickhouseConfig.ClickhouseConfig_Rabbitmq) string {
	if rabbitmq == nil {
		return ""
	}

	return fmt.Sprintf(`
rabbitmq = {
	username = "%s"
	password = "%s"
	vhost 	 = "%s"
}
`,
		rabbitmq.Username,
		rabbitmq.Password,
		rabbitmq.Vhost,
	)
}

func buildCompressionsHCL(compressions []*clickhouseConfig.ClickhouseConfig_Compression) string {
	var result string
	for _, v := range compressions {
		result += fmt.Sprintf(`
{
	method 				= "%s"
	min_part_size 		= %d
	min_part_size_ratio = %f
	level 				= %d
},
`,
			v.Method.String(),
			v.MinPartSize,
			v.MinPartSizeRatio,
			v.Level.GetValue(),
		)
	}

	return fmt.Sprintf(`
compression = [
	%s
]
	`,
		result,
	)
}

func buildGraphiteRollupsHCL(graphiteRollups []*clickhouseConfig.ClickhouseConfig_GraphiteRollup) string {
	var result string
	for _, v := range graphiteRollups {
		result += fmt.Sprintf(`
{
	name = "%s"
	# patterns
	%s
	path_column_name    = "%s"
	time_column_name    = "%s"
	value_column_name   = "%s"
	version_column_name = "%s"
},
`,
			v.Name,
			buildGraphiteRollupPatternsHCL(v.Patterns),
			v.PathColumnName,
			v.TimeColumnName,
			v.ValueColumnName,
			v.VersionColumnName,
		)
	}

	return fmt.Sprintf(`
graphite_rollup = [
	%s
]
	`,
		result,
	)
}

func buildGraphiteRollupPatternsHCL(patterns []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern) string {
	var result string
	for _, v := range patterns {
		result += fmt.Sprintf(`
{
	regexp   = "%s"
	function = "%s"
	# retention
	%s
},
`,
			v.Regexp,
			v.Function,
			buildGraphiteRollupRetentionsHCL(v.Retention),
		)
	}

	return fmt.Sprintf(`
patterns = [
	%s
]
	`,
		result,
	)
}

func buildGraphiteRollupRetentionsHCL(retentions []*clickhouseConfig.ClickhouseConfig_GraphiteRollup_Pattern_Retention) string {
	var result string
	for _, v := range retentions {
		result += fmt.Sprintf(`
{
	age       = %d
	precision = %d
},
`,
			v.Age,
			v.Precision,
		)
	}

	return fmt.Sprintf(`
retention = [
	%s
]
	`,
		result,
	)
}

func buildQueryMaskingRulesHCL(rules []*clickhouseConfig.ClickhouseConfig_QueryMaskingRule) string {
	var result string
	for _, v := range rules {
		result += fmt.Sprintf(`
{
	name 				= "%s"
	regexp 	        	= "%s"
	replace             = "%s"
},
`,
			v.Name,
			v.Regexp,
			v.Replace,
		)
	}

	return fmt.Sprintf(`
query_masking_rules = [
	%s
]
	`,
		result,
	)
}

func buildQueryCacheHCL(queryCache *clickhouseConfig.ClickhouseConfig_QueryCache) string {
	if queryCache == nil {
		return ""
	}

	return fmt.Sprintf(`
query_cache = {
	max_size_in_bytes                           = %d
	max_entries                   				= %d
	max_entry_size_in_bytes                     = %d
	max_entry_size_in_rows                      = %d
}
`,
		queryCache.MaxSizeInBytes.GetValue(),
		queryCache.MaxEntries.GetValue(),
		queryCache.MaxEntrySizeInBytes.GetValue(),
		queryCache.MaxEntrySizeInRows.GetValue(),
	)
}

func buildJdbcBridgeHCL(jdbcBridge *clickhouseConfig.ClickhouseConfig_JdbcBridge) string {
	if jdbcBridge == nil {
		return ""
	}

	return fmt.Sprintf(`
jdbc_bridge = {
	host                    = "%s"
	port                    = %d
}
`,
		jdbcBridge.Host,
		jdbcBridge.Port.GetValue(),
	)
}

func buildAccessControlImprovementsHCL(ac *clickhouseConfig.ClickhouseConfig_AccessControlImprovements) string {
	if ac == nil {
		return ""
	}

	return fmt.Sprintf(`
access_control_improvements = {
	select_from_system_db_requires_grant          = %t
	select_from_information_schema_requires_grant = %t
}
`,
		ac.SelectFromSystemDbRequiresGrant.GetValue(),
		ac.SelectFromInformationSchemaRequiresGrant.GetValue(),
	)
}

func buildCustomMacrosHCL(macros []*clickhouseConfig.ClickhouseConfig_Macro) string {
	var result string
	for _, v := range macros {
		result += fmt.Sprintf(`
{
	name 				= "%s"
	value 	        	= "%s"
},
`,
			v.Name,
			v.Value,
		)
	}

	return fmt.Sprintf(`
custom_macros = [
	%s
]
	`,
		result,
	)
}
