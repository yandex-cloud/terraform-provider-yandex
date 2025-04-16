package mdb_clickhouse_user_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_user"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func mdbClickHouseUserImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password",          // sensitive
			"generate_password", // does not return
		},
	}

}

func TestAccMDBClickHouseUser_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-clickhouse-user-basic")
	description := "Clickhouse User Terraform basic creation and updating Test"

	chUserResourceID1 := makeCHUserResource(chUserResourceName1)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBClickHouseUserConfig_basic_create(clusterName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseUserResourceIDField(chUserResourceID1),
					resource.TestCheckResourceAttr(chUserResourceID1, "name", chUserResourceName1),
					resource.TestCheckResourceAttr(chUserResourceID1, "generate_password", "false"),
					resource.TestCheckResourceAttr(chUserResourceID1, "connection_manager.%", "1"),
					testAccCheckMDBClickHouseClusterHasUsers(chClusterResourceID, []string{chUserResourceName1}),
				),
			},
			mdbClickHouseUserImportStep(chUserResourceID1),
			{
				Config: testAccMDBClickHouseUserConfig_basic_update(clusterName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseUserResourceIDField(chUserResourceID1),
					resource.TestCheckResourceAttr(chUserResourceID1, "name", chUserResourceName1),
					resource.TestCheckResourceAttr(chUserResourceID1, "generate_password", "false"),
					resource.TestCheckResourceAttr(chUserResourceID1, "connection_manager.%", "1"),
					testAccCheckMDBClickHouseUserHasDatabases(chUserResourceID1, []string{chDBResourceName1, chDBResourceName2}),
					resource.TestCheckResourceAttr(chUserResourceID1, "quota.0.interval_duration", "79800000"),
					resource.TestCheckResourceAttr(chUserResourceID1, "quota.0.queries", "5000"),
					resource.TestCheckResourceAttr(chUserResourceID1, "settings.readonly", "0"),
					resource.TestCheckResourceAttr(chUserResourceID1, "settings.allow_ddl", "true"),
					resource.TestCheckResourceAttr(chUserResourceID1, "settings.connect_timeout", "30000"),
					resource.TestCheckResourceAttr(chUserResourceID1, "settings.distributed_product_mode", "local"),
					resource.TestCheckResourceAttr(chUserResourceID1, "settings.max_block_size", "5008"),
					resource.TestCheckResourceAttr(chUserResourceID1, "settings.join_algorithm.0", "partial_merge"),
				),
			},
			mdbClickHouseUserImportStep(chUserResourceID1),
			{
				Config: testAccMDBClickHouseUserConfig_basic_delete(clusterName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseClusterHasUsers(chClusterResourceID, []string{}),
				),
			},
			{
				Config: testAccMDBClickHouseUserConfig_basic_several(clusterName, description, []string{chUserResourceName2, chUserResourceName3}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseUserResourceIDField(makeCHUserResource(chUserResourceName2)),
					testAccCheckMDBClickHouseClusterHasUsers(chClusterResourceID, []string{chUserResourceName2, chUserResourceName3}),
				),
			},
		},
	})
}

func TestAccMDBClickHouseUser_settings(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-clickhouse-user-settings")
	description := "Clickhouse User Terraform full settings Test"

	settingsCreate := mdb_clickhouse_user.Setting{
		Readonly:                            types.Int64Value(0),
		AllowDdl:                            types.BoolValue(false),
		AllowIntrospectionFunctions:         types.BoolValue(false),
		ConnectTimeout:                      types.Int64Value(20000),
		ConnectTimeoutWithFailover:          types.Int64Value(100),
		ReceiveTimeout:                      types.Int64Value(400000),
		SendTimeout:                         types.Int64Value(400000),
		TimeoutBeforeCheckingExecutionSpeed: types.Int64Value(20000),
		InsertQuorum:                        types.Int64Value(2),
		InsertQuorumTimeout:                 types.Int64Value(30000),
		InsertQuorumParallel:                types.BoolValue(false),
		InsertNullAsDefault:                 types.BoolValue(false),
		SelectSequentialConsistency:         types.BoolValue(false),
		DeduplicateBlocksInDependentMaterializedViews: types.BoolValue(false),
		ReplicationAlterPartitionsSync:                types.Int64Value(2),
		MaxReplicaDelayForDistributedQueries:          types.Int64Value(300000),
		FallbackToStaleReplicasForDistributedQueries:  types.BoolValue(false),
		DistributedProductMode:                        types.StringValue("allow"),
		DistributedAggregationMemoryEfficient:         types.BoolValue(false),
		DistributedDdlTaskTimeout:                     types.Int64Value(360000),
		SkipUnavailableShards:                         types.BoolValue(false),
		CompileExpressions:                            types.BoolValue(false),
		MinCountToCompileExpression:                   types.Int64Value(2),
		MaxBlockSize:                                  types.Int64Value(32768),
		MinInsertBlockSizeRows:                        types.Int64Value(1048576),
		MinInsertBlockSizeBytes:                       types.Int64Value(268435456),
		MaxInsertBlockSize:                            types.Int64Value(268435456),
		MinBytesToUseDirectIo:                         types.Int64Value(52428800),
		UseUncompressedCache:                          types.BoolValue(false),
		MergeTreeMaxRowsToUseCache:                    types.Int64Value(1048576),
		MergeTreeMaxBytesToUseCache:                   types.Int64Value(2013265920),
		MergeTreeMinRowsForConcurrentRead:             types.Int64Value(163840),
		MergeTreeMinBytesForConcurrentRead:            types.Int64Value(251658240),
		MaxBytesBeforeExternalGroupBy:                 types.Int64Value(0),
		MaxBytesBeforeExternalSort:                    types.Int64Value(0),
		GroupByTwoLevelThreshold:                      types.Int64Value(100000),
		GroupByTwoLevelThresholdBytes:                 types.Int64Value(1000000000),
		Priority:                                      types.Int64Value(1),
		MaxThreads:                                    types.Int64Value(10),
		MaxMemoryUsage:                                types.Int64Value(21474836480),
		MaxMemoryUsageForUser:                         types.Int64Value(53687091200),
		MaxNetworkBandwidth:                           types.Int64Value(1073741824),
		MaxNetworkBandwidthForUser:                    types.Int64Value(2147483648),
		MaxPartitionsPerInsertBlock:                   types.Int64Value(150),
		MaxConcurrentQueriesForUser:                   types.Int64Value(100),
		ForceIndexByDate:                              types.BoolValue(false),
		ForcePrimaryKey:                               types.BoolValue(false),
		MaxRowsToRead:                                 types.Int64Value(1000000),
		MaxBytesToRead:                                types.Int64Value(2000000),
		ReadOverflowMode:                              types.StringValue("throw"),
		MaxRowsToGroupBy:                              types.Int64Value(1000001),
		GroupByOverflowMode:                           types.StringValue("any"),
		MaxRowsToSort:                                 types.Int64Value(1000002),
		MaxBytesToSort:                                types.Int64Value(2000002),
		SortOverflowMode:                              types.StringValue("break"),
		MaxResultRows:                                 types.Int64Value(1000003),
		MaxResultBytes:                                types.Int64Value(2000003),
		ResultOverflowMode:                            types.StringValue("throw"),
		MaxRowsInDistinct:                             types.Int64Value(1000004),
		MaxBytesInDistinct:                            types.Int64Value(2000004),
		DistinctOverflowMode:                          types.StringValue("break"),
		MaxRowsToTransfer:                             types.Int64Value(1000005),
		MaxBytesToTransfer:                            types.Int64Value(2000005),
		TransferOverflowMode:                          types.StringValue("throw"),
		MaxExecutionTime:                              types.Int64Value(600000),
		TimeoutOverflowMode:                           types.StringValue("break"),
		MaxRowsInSet:                                  types.Int64Value(1000006),
		MaxBytesInSet:                                 types.Int64Value(2000006),
		SetOverflowMode:                               types.StringValue("throw"),
		MaxRowsInJoin:                                 types.Int64Value(1000007),
		MaxBytesInJoin:                                types.Int64Value(2000007),
		JoinOverflowMode:                              types.StringValue("break"),
		AnyJoinDistinctRightTableKeys:                 types.BoolValue(false),
		MaxColumnsToRead:                              types.Int64Value(25),
		MaxTemporaryColumns:                           types.Int64Value(20),
		MaxTemporaryNonConstColumns:                   types.Int64Value(15),
		MaxQuerySize:                                  types.Int64Value(524288),
		MaxAstDepth:                                   types.Int64Value(2000),
		MaxAstElements:                                types.Int64Value(100000),
		MaxExpandedAstElements:                        types.Int64Value(1000000),
		MinExecutionSpeed:                             types.Int64Value(1000008),
		MinExecutionSpeedBytes:                        types.Int64Value(2000008),
		CountDistinctImplementation:                   types.StringValue("uniq_hll_12"),
		InputFormatValuesInterpretExpressions:         types.BoolValue(false),
		InputFormatDefaultsForOmittedFields:           types.BoolValue(false),
		InputFormatNullAsDefault:                      types.BoolValue(false),
		DateTimeInputFormat:                           types.StringValue("best_effort"),
		InputFormatWithNamesUseHeader:                 types.BoolValue(false),
		OutputFormatJsonQuote_64BitIntegers:           types.BoolValue(false),
		OutputFormatJsonQuoteDenormals:                types.BoolValue(false),
		DateTimeOutputFormat:                          types.StringValue("iso"),
		LowCardinalityAllowInNativeFormat:             types.BoolValue(false),
		AllowSuspiciousLowCardinalityTypes:            types.BoolValue(false),
		EmptyResultForAggregationByEmptySet:           types.BoolValue(false),
		HttpConnectionTimeout:                         types.Int64Value(3000),
		HttpReceiveTimeout:                            types.Int64Value(1800000),
		HttpSendTimeout:                               types.Int64Value(1900000),
		EnableHttpCompression:                         types.BoolValue(false),
		SendProgressInHttpHeaders:                     types.BoolValue(false),
		HttpHeadersProgressInterval:                   types.Int64Value(1000),
		AddHttpCorsHeader:                             types.BoolValue(false),
		CancelHttpReadonlyQueriesOnClientClose:        types.BoolValue(false),
		MaxHttpGetRedirects:                           types.Int64Value(10),
		JoinedSubqueryRequiresAlias:                   types.BoolValue(false),
		JoinUseNulls:                                  types.BoolValue(false),
		TransformNullIn:                               types.BoolValue(false),
		QuotaMode:                                     types.StringValue("keyed"),
		FlattenNested:                                 types.BoolValue(false),
		FormatRegexp:                                  types.StringValue("regexp"),
		FormatRegexpSkipUnmatched:                     types.BoolValue(false),
		AsyncInsert:                                   types.BoolValue(false),
		AsyncInsertThreads:                            types.Int64Value(10),
		WaitForAsyncInsert:                            types.BoolValue(false),
		WaitForAsyncInsertTimeout:                     types.Int64Value(200),
		AsyncInsertMaxDataSize:                        types.Int64Value(1024),
		AsyncInsertBusyTimeout:                        types.Int64Value(1000),
		AsyncInsertStaleTimeout:                       types.Int64Value(1000),
		MemoryProfilerStep:                            types.Int64Value(1024),
		MemoryProfilerSampleProbability:               types.Float64Value(0.8),
		MaxFinalThreads:                               types.Int64Value(16),
		InputFormatParallelParsing:                    types.BoolValue(false),
		InputFormatImportNestedJson:                   types.BoolValue(false),
		LocalFilesystemReadMethod:                     types.StringValue("read"),
		MaxReadBufferSize:                             types.Int64Value(10485780),
		InsertKeeperMaxRetries:                        types.Int64Value(21),
		MaxTemporaryDataOnDiskSizeForUser:             types.Int64Value(2147483652),
		MaxTemporaryDataOnDiskSizeForQuery:            types.Int64Value(1073741826),
		MaxParserDepth:                                types.Int64Value(1007),
		RemoteFilesystemReadMethod:                    types.StringValue("threadpool"),
		MemoryOvercommitRatioDenominator:              types.Int64Value(1073741828),
		MemoryOvercommitRatioDenominatorForUser:       types.Int64Value(2147483656),
		MemoryUsageOvercommitMaxWaitMicroseconds:      types.Int64Value(5000008),
		LogQueryThreads:                               types.BoolValue(false),
		MaxInsertThreads:                              types.Int64Value(10),
		UseHedgedRequests:                             types.BoolValue(false),
		IdleConnectionTimeout:                         types.Int64Value(400000),
		HedgedConnectionTimeoutMs:                     types.Int64Value(1000),
		LoadBalancing:                                 types.StringValue("first_or_random"),
		PreferLocalhostReplica:                        types.BoolValue(false),
	}

	settingsUpdate := mdb_clickhouse_user.Setting{
		Readonly:                            types.Int64Value(1),
		AllowDdl:                            types.BoolValue(true),
		AllowIntrospectionFunctions:         types.BoolValue(true),
		ConnectTimeout:                      types.Int64Value(21000),
		ConnectTimeoutWithFailover:          types.Int64Value(200),
		ReceiveTimeout:                      types.Int64Value(410000),
		SendTimeout:                         types.Int64Value(410000),
		TimeoutBeforeCheckingExecutionSpeed: types.Int64Value(21000),
		InsertQuorum:                        types.Int64Value(2),
		InsertQuorumTimeout:                 types.Int64Value(31000),
		InsertQuorumParallel:                types.BoolValue(true),
		InsertNullAsDefault:                 types.BoolValue(true),
		SelectSequentialConsistency:         types.BoolValue(true),
		DeduplicateBlocksInDependentMaterializedViews: types.BoolValue(true),
		ReplicationAlterPartitionsSync:                types.Int64Value(2),
		MaxReplicaDelayForDistributedQueries:          types.Int64Value(310000),
		FallbackToStaleReplicasForDistributedQueries:  types.BoolValue(true),
		DistributedProductMode:                        types.StringValue("local"),
		DistributedAggregationMemoryEfficient:         types.BoolValue(true),
		DistributedDdlTaskTimeout:                     types.Int64Value(370000),
		SkipUnavailableShards:                         types.BoolValue(true),
		CompileExpressions:                            types.BoolValue(true),
		MinCountToCompileExpression:                   types.Int64Value(2),
		MaxBlockSize:                                  types.Int64Value(31768),
		MinInsertBlockSizeRows:                        types.Int64Value(1148576),
		MinInsertBlockSizeBytes:                       types.Int64Value(278435456),
		MaxInsertBlockSize:                            types.Int64Value(278435456),
		MinBytesToUseDirectIo:                         types.Int64Value(53428800),
		UseUncompressedCache:                          types.BoolValue(true),
		MergeTreeMaxRowsToUseCache:                    types.Int64Value(1148576),
		MergeTreeMaxBytesToUseCache:                   types.Int64Value(2113265920),
		MergeTreeMinRowsForConcurrentRead:             types.Int64Value(173840),
		MergeTreeMinBytesForConcurrentRead:            types.Int64Value(261658240),
		MaxBytesBeforeExternalGroupBy:                 types.Int64Value(0),
		MaxBytesBeforeExternalSort:                    types.Int64Value(0),
		GroupByTwoLevelThreshold:                      types.Int64Value(110000),
		GroupByTwoLevelThresholdBytes:                 types.Int64Value(1100000000),
		Priority:                                      types.Int64Value(1),
		MaxThreads:                                    types.Int64Value(11),
		MaxMemoryUsage:                                types.Int64Value(22474836480),
		MaxMemoryUsageForUser:                         types.Int64Value(54687091200),
		MaxNetworkBandwidth:                           types.Int64Value(1173741824),
		MaxNetworkBandwidthForUser:                    types.Int64Value(2247483648),
		MaxPartitionsPerInsertBlock:                   types.Int64Value(160),
		MaxConcurrentQueriesForUser:                   types.Int64Value(110),
		ForceIndexByDate:                              types.BoolValue(true),
		ForcePrimaryKey:                               types.BoolValue(true),
		MaxRowsToRead:                                 types.Int64Value(1100000),
		MaxBytesToRead:                                types.Int64Value(2100000),
		ReadOverflowMode:                              types.StringValue("break"),
		MaxRowsToGroupBy:                              types.Int64Value(1100001),
		GroupByOverflowMode:                           types.StringValue("break"),
		MaxRowsToSort:                                 types.Int64Value(1100002),
		MaxBytesToSort:                                types.Int64Value(2100002),
		SortOverflowMode:                              types.StringValue("throw"),
		MaxResultRows:                                 types.Int64Value(1100003),
		MaxResultBytes:                                types.Int64Value(2100003),
		ResultOverflowMode:                            types.StringValue("break"),
		MaxRowsInDistinct:                             types.Int64Value(1100004),
		MaxBytesInDistinct:                            types.Int64Value(2100004),
		DistinctOverflowMode:                          types.StringValue("throw"),
		MaxRowsToTransfer:                             types.Int64Value(1100005),
		MaxBytesToTransfer:                            types.Int64Value(2100005),
		TransferOverflowMode:                          types.StringValue("break"),
		MaxExecutionTime:                              types.Int64Value(610000),
		TimeoutOverflowMode:                           types.StringValue("throw"),
		MaxRowsInSet:                                  types.Int64Value(1100006),
		MaxBytesInSet:                                 types.Int64Value(2100006),
		SetOverflowMode:                               types.StringValue("break"),
		MaxRowsInJoin:                                 types.Int64Value(1100007),
		MaxBytesInJoin:                                types.Int64Value(2100007),
		JoinOverflowMode:                              types.StringValue("throw"),
		AnyJoinDistinctRightTableKeys:                 types.BoolValue(true),
		MaxColumnsToRead:                              types.Int64Value(26),
		MaxTemporaryColumns:                           types.Int64Value(21),
		MaxTemporaryNonConstColumns:                   types.Int64Value(16),
		MaxQuerySize:                                  types.Int64Value(534288),
		MaxAstDepth:                                   types.Int64Value(3000),
		MaxAstElements:                                types.Int64Value(110000),
		MaxExpandedAstElements:                        types.Int64Value(1100000),
		MinExecutionSpeed:                             types.Int64Value(1100008),
		MinExecutionSpeedBytes:                        types.Int64Value(2100008),
		CountDistinctImplementation:                   types.StringValue("uniq_combined_64"),
		InputFormatValuesInterpretExpressions:         types.BoolValue(true),
		InputFormatDefaultsForOmittedFields:           types.BoolValue(true),
		InputFormatNullAsDefault:                      types.BoolValue(true),
		DateTimeInputFormat:                           types.StringValue("basic"),
		InputFormatWithNamesUseHeader:                 types.BoolValue(true),
		OutputFormatJsonQuote_64BitIntegers:           types.BoolValue(true),
		OutputFormatJsonQuoteDenormals:                types.BoolValue(true),
		DateTimeOutputFormat:                          types.StringValue("simple"),
		LowCardinalityAllowInNativeFormat:             types.BoolValue(true),
		AllowSuspiciousLowCardinalityTypes:            types.BoolValue(true),
		EmptyResultForAggregationByEmptySet:           types.BoolValue(true),
		HttpConnectionTimeout:                         types.Int64Value(4000),
		HttpReceiveTimeout:                            types.Int64Value(1900000),
		HttpSendTimeout:                               types.Int64Value(2000000),
		EnableHttpCompression:                         types.BoolValue(true),
		SendProgressInHttpHeaders:                     types.BoolValue(true),
		HttpHeadersProgressInterval:                   types.Int64Value(2000),
		AddHttpCorsHeader:                             types.BoolValue(true),
		CancelHttpReadonlyQueriesOnClientClose:        types.BoolValue(true),
		MaxHttpGetRedirects:                           types.Int64Value(11),
		JoinedSubqueryRequiresAlias:                   types.BoolValue(true),
		JoinUseNulls:                                  types.BoolValue(true),
		TransformNullIn:                               types.BoolValue(true),
		QuotaMode:                                     types.StringValue("default"),
		FlattenNested:                                 types.BoolValue(true),
		FormatRegexp:                                  types.StringValue("regexpp"),
		FormatRegexpSkipUnmatched:                     types.BoolValue(true),
		AsyncInsert:                                   types.BoolValue(true),
		AsyncInsertThreads:                            types.Int64Value(11),
		WaitForAsyncInsert:                            types.BoolValue(true),
		WaitForAsyncInsertTimeout:                     types.Int64Value(210),
		AsyncInsertMaxDataSize:                        types.Int64Value(2024),
		AsyncInsertBusyTimeout:                        types.Int64Value(2000),
		AsyncInsertStaleTimeout:                       types.Int64Value(3000),
		MemoryProfilerStep:                            types.Int64Value(1124),
		MemoryProfilerSampleProbability:               types.Float64Value(0.7),
		MaxFinalThreads:                               types.Int64Value(18),
		InputFormatParallelParsing:                    types.BoolValue(true),
		InputFormatImportNestedJson:                   types.BoolValue(true),
		LocalFilesystemReadMethod:                     types.StringValue("pread"),
		MaxReadBufferSize:                             types.Int64Value(11485780),
		InsertKeeperMaxRetries:                        types.Int64Value(22),
		MaxTemporaryDataOnDiskSizeForUser:             types.Int64Value(2247483652),
		MaxTemporaryDataOnDiskSizeForQuery:            types.Int64Value(1173741826),
		MaxParserDepth:                                types.Int64Value(1107),
		RemoteFilesystemReadMethod:                    types.StringValue("read"),
		MemoryOvercommitRatioDenominator:              types.Int64Value(1173741828),
		MemoryOvercommitRatioDenominatorForUser:       types.Int64Value(2247483656),
		MemoryUsageOvercommitMaxWaitMicroseconds:      types.Int64Value(5100008),
		LogQueryThreads:                               types.BoolValue(true),
		MaxInsertThreads:                              types.Int64Value(12),
		UseHedgedRequests:                             types.BoolValue(true),
		IdleConnectionTimeout:                         types.Int64Value(510000),
		HedgedConnectionTimeoutMs:                     types.Int64Value(3000),
		LoadBalancing:                                 types.StringValue("random"),
		PreferLocalhostReplica:                        types.BoolValue(true),
	}

	chResourceID := makeCHUserResource(chUserResourceName4)
	chUserName := chUserResourceName4

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBClickHouseUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBClickHouseUserWithFullSettings(clusterName, description, chUserName, chDBResourceName1, settingsCreate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseUserResourceIDField(chResourceID),
					resource.TestCheckResourceAttr(chResourceID, "name", chUserName),
					resource.TestCheckResourceAttr(chResourceID, "generate_password", "true"),
					resource.TestCheckResourceAttr(chResourceID, "connection_manager.%", "1"),
					testAccCheckMDBClickHouseUserSettingsSet(chResourceID, settingsCreate),
				),
			},
			mdbClickHouseUserImportStep(chResourceID),
			{
				Config: testAccMDBClickHouseUserWithFullSettings(clusterName, description, chUserName, chDBResourceName1, settingsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBClickHouseUserResourceIDField(chResourceID),
					resource.TestCheckResourceAttr(chResourceID, "name", chUserName),
					resource.TestCheckResourceAttr(chResourceID, "generate_password", "true"),
					resource.TestCheckResourceAttr(chResourceID, "connection_manager.%", "1"),
					testAccCheckMDBClickHouseUserSettingsSet(chResourceID, settingsUpdate),
				),
			},
			mdbClickHouseUserImportStep(chResourceID),
		},
	})
}

func testAccMDBClickHouseUserConfig_basic_create(name string, description string) string {
	return testAccMDBClickHouseClusterConfigMain(name, description) + fmt.Sprintf(`

	resource "yandex_mdb_clickhouse_user" "%s" {
		cluster_id = %s
		name       = "%s"
		password   = "mysecureP@ssw0rd"
		permission {
	      database_name = %s.name
	  	}
		settings {
		}
	}
	`, chUserResourceName1, chClusterResourceIDLink, chUserResourceName1, makeCHDBResource(chDBResourceName1))
}

func testAccMDBClickHouseUserConfig_basic_update(name string, description string) string {
	return testAccMDBClickHouseClusterConfigMain(name, description) + fmt.Sprintf(`

	resource "yandex_mdb_clickhouse_user" "%s" {
		cluster_id = %s
		name       = "%s"
		password   = "mysecureP@ssw0rd"
		permission {
	      database_name = %s.name
	  	}
		permission {
	      database_name = %s.name
	  	}
		quota {
		  interval_duration = 79800000
		  queries           = 5000
		}
		settings {
          readonly = 0
          allow_ddl = true
          connect_timeout = 30000
          distributed_product_mode = "local"
          join_algorithm = [ "partial_merge" ]
          max_block_size = 5008
		}

	}
	`, chUserResourceName1, chClusterResourceIDLink, chUserResourceName1,
		makeCHDBResource(chDBResourceName1), makeCHDBResource(chDBResourceName2))
}

func testAccMDBClickHouseUserConfig_basic_delete(name, description string) string {
	return testAccMDBClickHouseClusterConfigMain(name, description)
}

func testAccMDBClickHouseUserConfig_basic_several(name, description string, userNames []string) string {
	planAll := testAccMDBClickHouseClusterConfigMain(name, description)
	for _, userName := range userNames {
		planAll = planAll + fmt.Sprintf(`
	resource "yandex_mdb_clickhouse_user" "%s" {
		cluster_id = %s
		name       = "%s"
		password   = "mysecureP@ssw0rd"
		permission {
	      database_name = %s.name
	  	}
		settings {
		}
	}
	`, userName, chClusterResourceIDLink, userName, makeCHDBResource(chDBResourceName1))
	}

	return planAll
}

func testAccMDBClickHouseUserWithFullSettings(name, desc, userName, dbName string, settings mdb_clickhouse_user.Setting) string {
	return testAccMDBClickHouseClusterConfigMain(name, desc) + fmt.Sprintf(`
   resource "yandex_mdb_clickhouse_user" "%s" {
    cluster_id = %s
    name     = "%s"
    generate_password = "true"
    permission {
      database_name = %s.name
    }
    settings {
      join_algorithm = ["hash", "auto"]
      readonly = %d
      allow_ddl = %t
      allow_introspection_functions = %t
      connect_timeout = %d
      connect_timeout_with_failover = %d
      receive_timeout = %d
      send_timeout = %d
      timeout_before_checking_execution_speed = %d
      insert_quorum = %d
      insert_quorum_timeout = %d
      insert_quorum_parallel = %t
      insert_null_as_default = %t
      select_sequential_consistency = %t
      deduplicate_blocks_in_dependent_materialized_views = %t
      replication_alter_partitions_sync = %d
      max_replica_delay_for_distributed_queries = %d
      fallback_to_stale_replicas_for_distributed_queries = %t
      distributed_product_mode = "%s"
      distributed_aggregation_memory_efficient = %t
      distributed_ddl_task_timeout = %d
      skip_unavailable_shards = %t
      compile_expressions = %t
      min_count_to_compile_expression = %d
      max_block_size = %d
      min_insert_block_size_rows = %d
      min_insert_block_size_bytes = %d
      max_insert_block_size = %d
      min_bytes_to_use_direct_io = %d
      use_uncompressed_cache = %t
      merge_tree_max_rows_to_use_cache = %d
      merge_tree_max_bytes_to_use_cache = %d
      merge_tree_min_rows_for_concurrent_read = %d
      merge_tree_min_bytes_for_concurrent_read = %d
      max_bytes_before_external_group_by = %d
      max_bytes_before_external_sort = %d
      group_by_two_level_threshold = %d
      group_by_two_level_threshold_bytes = %d
      priority = %d
      max_threads = %d
      max_memory_usage = %d
      max_memory_usage_for_user = %d
      max_network_bandwidth = %d
      max_network_bandwidth_for_user = %d
      max_partitions_per_insert_block = %d
      max_concurrent_queries_for_user = %d
      force_index_by_date = %t
      force_primary_key = %t
      max_rows_to_read = %d
      max_bytes_to_read = %d
      read_overflow_mode = "%s"
      max_rows_to_group_by = %d
      group_by_overflow_mode = "%s"
      max_rows_to_sort = %d
      max_bytes_to_sort = %d
      sort_overflow_mode = "%s"
      max_result_rows = %d
      max_result_bytes = %d
      result_overflow_mode = "%s"
      max_rows_in_distinct = %d
      max_bytes_in_distinct = %d
      distinct_overflow_mode = "%s"
      max_rows_to_transfer = %d
      max_bytes_to_transfer = %d
      transfer_overflow_mode = "%s"
      max_execution_time = %d
      timeout_overflow_mode = "%s"
      max_rows_in_set = %d
      max_bytes_in_set = %d
      set_overflow_mode = "%s"
      max_rows_in_join = %d
      max_bytes_in_join = %d
      join_overflow_mode = "%s"
      any_join_distinct_right_table_keys = %t
      max_columns_to_read = %d
      max_temporary_columns = %d
      max_temporary_non_const_columns = %d
      max_query_size = %d
      max_ast_depth = %d
      max_ast_elements = %d
      max_expanded_ast_elements = %d
      min_execution_speed = %d
      min_execution_speed_bytes = %d
      count_distinct_implementation = "%s"
      input_format_values_interpret_expressions = %t
      input_format_defaults_for_omitted_fields = %t
      input_format_null_as_default = %t
      date_time_input_format = "%s"
      input_format_with_names_use_header = %t
      output_format_json_quote_64bit_integers = %t
      output_format_json_quote_denormals = %t
      date_time_output_format = "%s"
      low_cardinality_allow_in_native_format = %t
      allow_suspicious_low_cardinality_types = %t
      empty_result_for_aggregation_by_empty_set = %t
      http_connection_timeout = %d
      http_receive_timeout = %d
      http_send_timeout = %d
      enable_http_compression = %t
      send_progress_in_http_headers = %t
      http_headers_progress_interval = %d
      add_http_cors_header = %t
      cancel_http_readonly_queries_on_client_close = %t
      max_http_get_redirects = %d
      joined_subquery_requires_alias = %t
      join_use_nulls = %t
      transform_null_in = %t
      quota_mode = "%s"
      flatten_nested = %t
      format_regexp = "%s"
      format_regexp_skip_unmatched = %t
      async_insert = %t
      async_insert_threads = %d
      wait_for_async_insert = %t
      wait_for_async_insert_timeout = %d
      async_insert_max_data_size = %d
      async_insert_busy_timeout = %d
      async_insert_stale_timeout = %d
      memory_profiler_step = %d
      memory_profiler_sample_probability = %f
      max_final_threads = %d
      input_format_parallel_parsing = %t
      input_format_import_nested_json = %t
      local_filesystem_read_method = "%s"
      max_read_buffer_size = %d
      insert_keeper_max_retries = %d
      max_temporary_data_on_disk_size_for_user = %d
      max_temporary_data_on_disk_size_for_query = %d
      max_parser_depth = %d
      remote_filesystem_read_method = "%s"
      memory_overcommit_ratio_denominator = %d
      memory_overcommit_ratio_denominator_for_user = %d
      memory_usage_overcommit_max_wait_microseconds = %d
      log_query_threads = %t
      max_insert_threads = %d
      use_hedged_requests = %t
      idle_connection_timeout = %d
      hedged_connection_timeout_ms = %d
      load_balancing = "%s"
      prefer_localhost_replica = %t
    }
  }
`, userName, chClusterResourceIDLink, userName, makeCHDBResource(dbName),
		settings.Readonly.ValueInt64(),
		settings.AllowDdl.ValueBool(),
		settings.AllowIntrospectionFunctions.ValueBool(),
		settings.ConnectTimeout.ValueInt64(),
		settings.ConnectTimeoutWithFailover.ValueInt64(),
		settings.ReceiveTimeout.ValueInt64(),
		settings.SendTimeout.ValueInt64(),
		settings.TimeoutBeforeCheckingExecutionSpeed.ValueInt64(),
		settings.InsertQuorum.ValueInt64(),
		settings.InsertQuorumTimeout.ValueInt64(),
		settings.InsertQuorumParallel.ValueBool(),
		settings.InsertNullAsDefault.ValueBool(),
		settings.SelectSequentialConsistency.ValueBool(),
		settings.DeduplicateBlocksInDependentMaterializedViews.ValueBool(),
		settings.ReplicationAlterPartitionsSync.ValueInt64(),
		settings.MaxReplicaDelayForDistributedQueries.ValueInt64(),
		settings.FallbackToStaleReplicasForDistributedQueries.ValueBool(),
		settings.DistributedProductMode.ValueString(),
		settings.DistributedAggregationMemoryEfficient.ValueBool(),
		settings.DistributedDdlTaskTimeout.ValueInt64(),
		settings.SkipUnavailableShards.ValueBool(),
		settings.CompileExpressions.ValueBool(),
		settings.MinCountToCompileExpression.ValueInt64(),
		settings.MaxBlockSize.ValueInt64(),
		settings.MinInsertBlockSizeRows.ValueInt64(),
		settings.MinInsertBlockSizeBytes.ValueInt64(),
		settings.MaxInsertBlockSize.ValueInt64(),
		settings.MinBytesToUseDirectIo.ValueInt64(),
		settings.UseUncompressedCache.ValueBool(),
		settings.MergeTreeMaxRowsToUseCache.ValueInt64(),
		settings.MergeTreeMaxBytesToUseCache.ValueInt64(),
		settings.MergeTreeMinRowsForConcurrentRead.ValueInt64(),
		settings.MergeTreeMinBytesForConcurrentRead.ValueInt64(),
		settings.MaxBytesBeforeExternalGroupBy.ValueInt64(),
		settings.MaxBytesBeforeExternalSort.ValueInt64(),
		settings.GroupByTwoLevelThreshold.ValueInt64(),
		settings.GroupByTwoLevelThresholdBytes.ValueInt64(),
		settings.Priority.ValueInt64(),
		settings.MaxThreads.ValueInt64(),
		settings.MaxMemoryUsage.ValueInt64(),
		settings.MaxMemoryUsageForUser.ValueInt64(),
		settings.MaxNetworkBandwidth.ValueInt64(),
		settings.MaxNetworkBandwidthForUser.ValueInt64(),
		settings.MaxPartitionsPerInsertBlock.ValueInt64(),
		settings.MaxConcurrentQueriesForUser.ValueInt64(),
		settings.ForceIndexByDate.ValueBool(),
		settings.ForcePrimaryKey.ValueBool(),
		settings.MaxRowsToRead.ValueInt64(),
		settings.MaxBytesToRead.ValueInt64(),
		settings.ReadOverflowMode.ValueString(),
		settings.MaxRowsToGroupBy.ValueInt64(),
		settings.GroupByOverflowMode.ValueString(),
		settings.MaxRowsToSort.ValueInt64(),
		settings.MaxBytesToSort.ValueInt64(),
		settings.SortOverflowMode.ValueString(),
		settings.MaxResultRows.ValueInt64(),
		settings.MaxResultBytes.ValueInt64(),
		settings.ResultOverflowMode.ValueString(),
		settings.MaxRowsInDistinct.ValueInt64(),
		settings.MaxBytesInDistinct.ValueInt64(),
		settings.DistinctOverflowMode.ValueString(),
		settings.MaxRowsToTransfer.ValueInt64(),
		settings.MaxBytesToTransfer.ValueInt64(),
		settings.TransferOverflowMode.ValueString(),
		settings.MaxExecutionTime.ValueInt64(),
		settings.TimeoutOverflowMode.ValueString(),
		settings.MaxRowsInSet.ValueInt64(),
		settings.MaxBytesInSet.ValueInt64(),
		settings.SetOverflowMode.ValueString(),
		settings.MaxRowsInJoin.ValueInt64(),
		settings.MaxBytesInJoin.ValueInt64(),
		settings.JoinOverflowMode.ValueString(),
		settings.AnyJoinDistinctRightTableKeys.ValueBool(),
		settings.MaxColumnsToRead.ValueInt64(),
		settings.MaxTemporaryColumns.ValueInt64(),
		settings.MaxTemporaryNonConstColumns.ValueInt64(),
		settings.MaxQuerySize.ValueInt64(),
		settings.MaxAstDepth.ValueInt64(),
		settings.MaxAstElements.ValueInt64(),
		settings.MaxExpandedAstElements.ValueInt64(),
		settings.MinExecutionSpeed.ValueInt64(),
		settings.MinExecutionSpeedBytes.ValueInt64(),
		settings.CountDistinctImplementation.ValueString(),
		settings.InputFormatValuesInterpretExpressions.ValueBool(),
		settings.InputFormatDefaultsForOmittedFields.ValueBool(),
		settings.InputFormatNullAsDefault.ValueBool(),
		settings.DateTimeInputFormat.ValueString(),
		settings.InputFormatWithNamesUseHeader.ValueBool(),
		settings.OutputFormatJsonQuote_64BitIntegers.ValueBool(),
		settings.OutputFormatJsonQuoteDenormals.ValueBool(),
		settings.DateTimeOutputFormat.ValueString(),
		settings.LowCardinalityAllowInNativeFormat.ValueBool(),
		settings.AllowSuspiciousLowCardinalityTypes.ValueBool(),
		settings.EmptyResultForAggregationByEmptySet.ValueBool(),
		settings.HttpConnectionTimeout.ValueInt64(),
		settings.HttpReceiveTimeout.ValueInt64(),
		settings.HttpSendTimeout.ValueInt64(),
		settings.EnableHttpCompression.ValueBool(),
		settings.SendProgressInHttpHeaders.ValueBool(),
		settings.HttpHeadersProgressInterval.ValueInt64(),
		settings.AddHttpCorsHeader.ValueBool(),
		settings.CancelHttpReadonlyQueriesOnClientClose.ValueBool(),
		settings.MaxHttpGetRedirects.ValueInt64(),
		settings.JoinedSubqueryRequiresAlias.ValueBool(),
		settings.JoinUseNulls.ValueBool(),
		settings.TransformNullIn.ValueBool(),
		settings.QuotaMode.ValueString(),
		settings.FlattenNested.ValueBool(),
		settings.FormatRegexp.ValueString(),
		settings.FormatRegexpSkipUnmatched.ValueBool(),
		settings.AsyncInsert.ValueBool(),
		settings.AsyncInsertThreads.ValueInt64(),
		settings.WaitForAsyncInsert.ValueBool(),
		settings.WaitForAsyncInsertTimeout.ValueInt64(),
		settings.AsyncInsertMaxDataSize.ValueInt64(),
		settings.AsyncInsertBusyTimeout.ValueInt64(),
		settings.AsyncInsertStaleTimeout.ValueInt64(),
		settings.MemoryProfilerStep.ValueInt64(),
		settings.MemoryProfilerSampleProbability.ValueFloat64(),
		settings.MaxFinalThreads.ValueInt64(),
		settings.InputFormatParallelParsing.ValueBool(),
		settings.InputFormatImportNestedJson.ValueBool(),
		settings.LocalFilesystemReadMethod.ValueString(),
		settings.MaxReadBufferSize.ValueInt64(),
		settings.InsertKeeperMaxRetries.ValueInt64(),
		settings.MaxTemporaryDataOnDiskSizeForUser.ValueInt64(),
		settings.MaxTemporaryDataOnDiskSizeForQuery.ValueInt64(),
		settings.MaxParserDepth.ValueInt64(),
		settings.RemoteFilesystemReadMethod.ValueString(),
		settings.MemoryOvercommitRatioDenominator.ValueInt64(),
		settings.MemoryOvercommitRatioDenominatorForUser.ValueInt64(),
		settings.MemoryUsageOvercommitMaxWaitMicroseconds.ValueInt64(),
		settings.LogQueryThreads.ValueBool(),
		settings.MaxInsertThreads.ValueInt64(),
		settings.UseHedgedRequests.ValueBool(),
		settings.IdleConnectionTimeout.ValueInt64(),
		settings.HedgedConnectionTimeoutMs.ValueInt64(),
		settings.LoadBalancing.ValueString(),
		settings.PreferLocalhostReplica.ValueBool(),
	)
}

func testAccCheckMDBClickHouseUserSettingsSet(chUserID string, settings mdb_clickhouse_user.Setting) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(chUserID, "settings.readonly", settings.Readonly.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.allow_ddl", settings.AllowDdl.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.allow_introspection_functions", settings.AllowIntrospectionFunctions.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.connect_timeout", settings.ConnectTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.connect_timeout_with_failover", settings.ConnectTimeoutWithFailover.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.receive_timeout", settings.ReceiveTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.send_timeout", settings.SendTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.timeout_before_checking_execution_speed", settings.TimeoutBeforeCheckingExecutionSpeed.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.insert_quorum", settings.InsertQuorum.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.insert_quorum_timeout", settings.InsertQuorumTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.insert_quorum_parallel", settings.InsertQuorumParallel.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.insert_null_as_default", settings.InsertNullAsDefault.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.select_sequential_consistency", settings.SelectSequentialConsistency.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.deduplicate_blocks_in_dependent_materialized_views", settings.DeduplicateBlocksInDependentMaterializedViews.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.replication_alter_partitions_sync", settings.ReplicationAlterPartitionsSync.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_replica_delay_for_distributed_queries", settings.MaxReplicaDelayForDistributedQueries.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.fallback_to_stale_replicas_for_distributed_queries", settings.FallbackToStaleReplicasForDistributedQueries.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.distributed_product_mode", settings.DistributedProductMode.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.distributed_aggregation_memory_efficient", settings.DistributedAggregationMemoryEfficient.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.distributed_ddl_task_timeout", settings.DistributedDdlTaskTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.skip_unavailable_shards", settings.SkipUnavailableShards.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.compile_expressions", settings.CompileExpressions.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.min_count_to_compile_expression", settings.MinCountToCompileExpression.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_block_size", settings.MaxBlockSize.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.min_insert_block_size_rows", settings.MinInsertBlockSizeRows.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.min_insert_block_size_bytes", settings.MinInsertBlockSizeBytes.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_insert_block_size", settings.MaxInsertBlockSize.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.min_bytes_to_use_direct_io", settings.MinBytesToUseDirectIo.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.use_uncompressed_cache", settings.UseUncompressedCache.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.merge_tree_max_rows_to_use_cache", settings.MergeTreeMaxRowsToUseCache.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.merge_tree_max_bytes_to_use_cache", settings.MergeTreeMaxBytesToUseCache.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.merge_tree_min_rows_for_concurrent_read", settings.MergeTreeMinRowsForConcurrentRead.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.merge_tree_min_bytes_for_concurrent_read", settings.MergeTreeMinBytesForConcurrentRead.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_bytes_before_external_group_by", settings.MaxBytesBeforeExternalGroupBy.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_bytes_before_external_sort", settings.MaxBytesBeforeExternalSort.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.group_by_two_level_threshold", settings.GroupByTwoLevelThreshold.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.group_by_two_level_threshold_bytes", settings.GroupByTwoLevelThresholdBytes.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.priority", settings.Priority.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_threads", settings.MaxThreads.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_memory_usage", settings.MaxMemoryUsage.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_memory_usage_for_user", settings.MaxMemoryUsageForUser.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_network_bandwidth", settings.MaxNetworkBandwidth.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_network_bandwidth_for_user", settings.MaxNetworkBandwidthForUser.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_partitions_per_insert_block", settings.MaxPartitionsPerInsertBlock.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_concurrent_queries_for_user", settings.MaxConcurrentQueriesForUser.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.force_index_by_date", settings.ForceIndexByDate.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.force_primary_key", settings.ForcePrimaryKey.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_rows_to_read", settings.MaxRowsToRead.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_bytes_to_read", settings.MaxBytesToRead.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.read_overflow_mode", settings.ReadOverflowMode.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_rows_to_group_by", settings.MaxRowsToGroupBy.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.group_by_overflow_mode", settings.GroupByOverflowMode.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_rows_to_sort", settings.MaxRowsToSort.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_bytes_to_sort", settings.MaxBytesToSort.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.sort_overflow_mode", settings.SortOverflowMode.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_result_rows", settings.MaxResultRows.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_result_bytes", settings.MaxResultBytes.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.result_overflow_mode", settings.ResultOverflowMode.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_rows_in_distinct", settings.MaxRowsInDistinct.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_bytes_in_distinct", settings.MaxBytesInDistinct.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.distinct_overflow_mode", settings.DistinctOverflowMode.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_rows_to_transfer", settings.MaxRowsToTransfer.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_bytes_to_transfer", settings.MaxBytesToTransfer.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.transfer_overflow_mode", settings.TransferOverflowMode.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_execution_time", settings.MaxExecutionTime.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.timeout_overflow_mode", settings.TimeoutOverflowMode.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_rows_in_set", settings.MaxRowsInSet.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_bytes_in_set", settings.MaxBytesInSet.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.set_overflow_mode", settings.SetOverflowMode.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_rows_in_join", settings.MaxRowsInJoin.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_bytes_in_join", settings.MaxBytesInJoin.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.join_overflow_mode", settings.JoinOverflowMode.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.any_join_distinct_right_table_keys", settings.AnyJoinDistinctRightTableKeys.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_columns_to_read", settings.MaxColumnsToRead.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_temporary_columns", settings.MaxTemporaryColumns.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_temporary_non_const_columns", settings.MaxTemporaryNonConstColumns.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_query_size", settings.MaxQuerySize.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_ast_depth", settings.MaxAstDepth.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_ast_elements", settings.MaxAstElements.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_expanded_ast_elements", settings.MaxExpandedAstElements.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.min_execution_speed", settings.MinExecutionSpeed.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.min_execution_speed_bytes", settings.MinExecutionSpeedBytes.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.count_distinct_implementation", settings.CountDistinctImplementation.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.input_format_values_interpret_expressions", settings.InputFormatValuesInterpretExpressions.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.input_format_defaults_for_omitted_fields", settings.InputFormatDefaultsForOmittedFields.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.input_format_null_as_default", settings.InputFormatNullAsDefault.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.date_time_input_format", settings.DateTimeInputFormat.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.input_format_with_names_use_header", settings.InputFormatWithNamesUseHeader.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.output_format_json_quote_64bit_integers", settings.OutputFormatJsonQuote_64BitIntegers.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.output_format_json_quote_denormals", settings.OutputFormatJsonQuoteDenormals.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.date_time_output_format", settings.DateTimeOutputFormat.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.low_cardinality_allow_in_native_format", settings.LowCardinalityAllowInNativeFormat.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.allow_suspicious_low_cardinality_types", settings.AllowSuspiciousLowCardinalityTypes.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.empty_result_for_aggregation_by_empty_set", settings.EmptyResultForAggregationByEmptySet.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.http_connection_timeout", settings.HttpConnectionTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.http_receive_timeout", settings.HttpReceiveTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.http_send_timeout", settings.HttpSendTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.enable_http_compression", settings.EnableHttpCompression.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.send_progress_in_http_headers", settings.SendProgressInHttpHeaders.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.http_headers_progress_interval", settings.HttpHeadersProgressInterval.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.add_http_cors_header", settings.AddHttpCorsHeader.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.cancel_http_readonly_queries_on_client_close", settings.CancelHttpReadonlyQueriesOnClientClose.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_http_get_redirects", settings.MaxHttpGetRedirects.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.joined_subquery_requires_alias", settings.JoinedSubqueryRequiresAlias.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.join_use_nulls", settings.JoinUseNulls.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.transform_null_in", settings.TransformNullIn.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.quota_mode", settings.QuotaMode.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.flatten_nested", settings.FlattenNested.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.format_regexp", settings.FormatRegexp.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.format_regexp_skip_unmatched", settings.FormatRegexpSkipUnmatched.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.async_insert", settings.AsyncInsert.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.async_insert_threads", settings.AsyncInsertThreads.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.wait_for_async_insert", settings.WaitForAsyncInsert.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.wait_for_async_insert_timeout", settings.WaitForAsyncInsertTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.async_insert_max_data_size", settings.AsyncInsertMaxDataSize.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.async_insert_busy_timeout", settings.AsyncInsertBusyTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.async_insert_stale_timeout", settings.AsyncInsertStaleTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.memory_profiler_step", settings.MemoryProfilerStep.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.memory_profiler_sample_probability", fmt.Sprintf("%.1f", settings.MemoryProfilerSampleProbability.ValueFloat64())),
		resource.TestCheckResourceAttr(chUserID, "settings.max_final_threads", settings.MaxFinalThreads.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.input_format_parallel_parsing", settings.InputFormatParallelParsing.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.input_format_import_nested_json", settings.InputFormatImportNestedJson.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.local_filesystem_read_method", settings.LocalFilesystemReadMethod.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_read_buffer_size", settings.MaxReadBufferSize.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.insert_keeper_max_retries", settings.InsertKeeperMaxRetries.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_temporary_data_on_disk_size_for_user", settings.MaxTemporaryDataOnDiskSizeForUser.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_temporary_data_on_disk_size_for_query", settings.MaxTemporaryDataOnDiskSizeForQuery.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_parser_depth", settings.MaxParserDepth.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.remote_filesystem_read_method", settings.RemoteFilesystemReadMethod.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.memory_overcommit_ratio_denominator", settings.MemoryOvercommitRatioDenominator.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.memory_overcommit_ratio_denominator_for_user", settings.MemoryOvercommitRatioDenominatorForUser.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.memory_usage_overcommit_max_wait_microseconds", settings.MemoryUsageOvercommitMaxWaitMicroseconds.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.log_query_threads", settings.LogQueryThreads.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.max_insert_threads", settings.MaxInsertThreads.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.use_hedged_requests", settings.UseHedgedRequests.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.idle_connection_timeout", settings.IdleConnectionTimeout.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.hedged_connection_timeout_ms", settings.HedgedConnectionTimeoutMs.String()),
		resource.TestCheckResourceAttr(chUserID, "settings.load_balancing", settings.LoadBalancing.ValueString()),
		resource.TestCheckResourceAttr(chUserID, "settings.prefer_localhost_replica", settings.PreferLocalhostReplica.String()),
	)
}

func testAccCheckMDBClickHouseUserHasDatabases(r string, databases []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		if rs.Type != "yandex_mdb_clickhouse_user" {
			return fmt.Errorf("Invalid resource type: %s", rs.Type)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		clusterId, userName, err := resourceid.Deconstruct(rs.Primary.ID)
		if err != nil {
			return err
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		resp, err := config.SDK.MDB().Clickhouse().User().Get(context.Background(), &clickhouse.GetUserRequest{
			ClusterId: clusterId,
			UserName:  userName,
		})
		if err != nil {
			return err
		}
		dbs := []string{}
		for _, p := range resp.Permissions {
			dbs = append(dbs, p.DatabaseName)
		}

		if len(dbs) != len(databases) {
			return fmt.Errorf("Expected %d dbs, found %d", len(databases), len(dbs))
		}

		sort.Strings(dbs)
		sort.Strings(databases)
		if fmt.Sprintf("%v", dbs) != fmt.Sprintf("%v", databases) {
			return fmt.Errorf("User has wrong databases, %v. Expected %v", dbs, databases)
		}

		return nil
	}
}

func testAccCheckMDBClickHouseClusterHasUsers(r string, users []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}

		cid := rs.Primary.ID

		resp, err := config.SDK.MDB().Clickhouse().User().List(context.Background(), &clickhouse.ListUsersRequest{
			ClusterId: cid,
			PageSize:  100,
		})
		if err != nil {
			return err
		}
		cUsers := []string{}
		for _, u := range resp.Users {
			cUsers = append(cUsers, u.Name)
		}

		if len(cUsers) != len(users) {
			return fmt.Errorf("Cluster %s has different number of users. Expected %d dbs, found %d %v", cid, len(users), len(cUsers), cUsers)
		}

		sort.Strings(cUsers)
		sort.Strings(users)
		if fmt.Sprintf("%v", cUsers) != fmt.Sprintf("%v", users) {
			return fmt.Errorf("Cluster %s has wrong users. Expected: %v got: %v. ", cid, users, cUsers)
		}
		return nil
	}
}
