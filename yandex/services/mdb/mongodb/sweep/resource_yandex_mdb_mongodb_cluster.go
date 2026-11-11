package sweep

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	mongodbsdk "github.com/yandex-cloud/go-sdk/services/mdb/mongodb/v1"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/genproto/protobuf/field_mask"
)

func registerResourceYandexMdbMongodbCluster(add func(string, *resource.Sweeper)) {
	add(mongodbResourceType, &resource.Sweeper{
		Name: mongodbResourceType,
		F:    sweepAllMDBMongoDBCluster,
	})
}

func sweepAllMDBMongoDBCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := mongodbsdk.NewClusterClient(conf.SDK).List(conf.Context(), &mongodb.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: sdkcommon.DefaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting MongoDB clusters: %s", err)
	}

	result := &multierror.Group{}
	for _, c := range resp.Clusters {
		id := c.Id
		result.Go(func() error { return sweepMDBMongoDBCluster(conf, id) })
	}

	return result.Wait().ErrorOrNil()
}

func sweepMDBMongoDBCluster(conf *providerconfig.Config, id string) error {
	if !sweep.SweepWithRetry(sweepMDBMongoDBClusterOnce, conf, "MongoDB cluster", id) {
		return fmt.Errorf("failed to sweep MongoDB cluster %q", id)
	}
	return nil
}

func sweepMDBMongoDBClusterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(*schema.DefaultTimeout(30 * time.Minute))
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := mongodbsdk.NewClusterClient(conf.SDK).Update(ctx, &mongodb.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = sweep.HandleSweepOperationV2(ctx, op, err)
	if err != nil && !strings.EqualFold(sdkcommon.ErrorMessage(err), "no changes detected") {
		return err
	}

	deleteOp, err := mongodbsdk.NewClusterClient(conf.SDK).Delete(ctx, &mongodb.DeleteClusterRequest{
		ClusterId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, deleteOp, err)
}

const (
	mongodbRestoreBackupId          = "c9qvb4o0gnrh8ene82l7:mdb948pf5jgli93dbujp"
	mongodbRestoreBackupIdEncrypted = "c9q1omobes6h1t4bmm5b:mdbos8dlgnfe5nqvt5nl"

	mongodbResourceType = "yandex_mdb_mongodb_cluster"
	mongodbResourceFoo  = mongodbResourceType + ".foo"

	mongodbVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-d"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_security_group" "sg-x" {
  network_id     = "${yandex_vpc_network.foo.id}"
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

resource "yandex_vpc_security_group" "sg-y" {
  network_id     = "${yandex_vpc_network.foo.id}"
  
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

	resourceYandexMdbMongodbClusterTemplateText = mongodbVPCDependencies + `
resource "yandex_mdb_mongodb_cluster" "foo" {
  name        = "{{.ClusterName}}"
{{- if .ClusterDescription}}
  description = "{{.ClusterDescription}}"
{{- end}}
  environment = "{{.Environment}}"
  network_id  = "${yandex_vpc_network.foo.id}"

{{if .Restore}}
  restore {
	backup_id = "{{.Restore.BackupId}}"
	{{if .Restore.Time}}
	time = "{{.Restore.Time}}"
	{{end}}
  }
{{end}}

{{if .Lables}}
  labels = {
{{- range $key, $value := .Lables}}
    {{ $key }} = "{{ $value }}"
{{- end}}
  }
{{end}}

  cluster_config {
    version = "{{.Version}}"
{{ if .CompatibilityVersion}}
    feature_compatibility_version = "{{.CompatibilityVersion}}"
{{end}}
{{if .Access}}
	access {
	{{- range $key, $value := .Access}}
		{{ $key }} = "{{ $value }}"
	{{- end}}
	} 
{{end}}
{{if .PerformanceDiagnostics}}
	performance_diagnostics {
		enabled = "{{.PerformanceDiagnostics.enabled}}"
	}
{{end}}
{{if .AutocompactConfig}}
	autocompact_config {
		enabled = {{.AutocompactConfig.enabled}}
	{{- if .AutocompactConfig.compaction_type}}
		compaction_type = "{{.AutocompactConfig.compaction_type}}"
	{{- end}}
	{{- if .AutocompactConfig.target_free_space}}
		target_free_space = {{.AutocompactConfig.target_free_space}}
	{{- end}}
	{{- if .AutocompactConfig.bloat_percent}}
		bloat_percent = {{.AutocompactConfig.bloat_percent}}
	{{- end}}
	}
{{end}}
{{if .BackupWindow}}
    backup_window_start {
      hours = {{.BackupWindow.hours}}
      minutes = {{.BackupWindow.minutes}}
    }
{{end}}
{{if .BackupRetainPeriodDays}}
    backup_retain_period_days = {{.BackupRetainPeriodDays}}
{{end}}
{{if .Mongod}}
    mongod {
{{if .Mongod.AuditLog}}
      audit_log {
        filter = "{{escapeQuotations .Mongod.AuditLog.Filter}}"
      }
{{end}}
{{if .Mongod.SetParameter}}
      set_parameter {
				{{if .Mongod.SetParameter.EnableFlowControl}}
					enable_flow_control = {{.Mongod.SetParameter.EnableFlowControl}}
				{{end}}
				{{if .Mongod.SetParameter.AuditAuthorizationSuccess}}
					audit_authorization_success = {{.Mongod.SetParameter.AuditAuthorizationSuccess}}
				{{end}}
				{{if .Mongod.SetParameter.MinSnapshotHistoryWindowInSeconds}}
				    min_snapshot_history_window_in_seconds = {{.Mongod.SetParameter.MinSnapshotHistoryWindowInSeconds}}
				{{end}}
				{{if .Mongod.SetParameter.FlowControlTargetLagSeconds}}
				    flow_control_target_lag_seconds = {{.Mongod.SetParameter.FlowControlTargetLagSeconds}}
				{{end}}
				{{if .Mongod.SetParameter.FlowControlWarnThresholdSeconds}}
				    flow_control_warn_threshold_seconds = {{.Mongod.SetParameter.FlowControlWarnThresholdSeconds}}
				{{end}}
				{{if .Mongod.SetParameter.MigrateCloneInsertionBatchDelayMs}}
				    migrate_clone_insertion_batch_delay_ms = {{.Mongod.SetParameter.MigrateCloneInsertionBatchDelayMs}}
				{{end}}
				{{if .Mongod.SetParameter.MigrateCloneInsertionBatchSize}}
				    migrate_clone_insertion_batch_size = {{.Mongod.SetParameter.MigrateCloneInsertionBatchSize}}
				{{end}}
				{{if .Mongod.SetParameter.OrphanCleanupDelaySecs}}
				    orphan_cleanup_delay_secs = {{.Mongod.SetParameter.OrphanCleanupDelaySecs}}
				{{end}}
				{{if .Mongod.SetParameter.PersistedChunkCacheUpdateMaxBatchSize}}
				    persisted_chunk_cache_update_max_batch_size = {{.Mongod.SetParameter.PersistedChunkCacheUpdateMaxBatchSize}}
				{{end}}
				{{if .Mongod.SetParameter.RangeDeleterBatchDelayMs}}
				    range_deleter_batch_delay_ms = {{.Mongod.SetParameter.RangeDeleterBatchDelayMs}}
				{{end}}
				{{if .Mongod.SetParameter.RangeDeleterBatchSize}}
				    range_deleter_batch_size = {{.Mongod.SetParameter.RangeDeleterBatchSize}}
				{{end}}
				{{if .Mongod.SetParameter.MirrorReads}}
				  mirror_reads {
						{{if .Mongod.SetParameter.MirrorReads.SamplingRate}}
							sampling_rate = {{.Mongod.SetParameter.MirrorReads.SamplingRate}}
						{{end}}
						{{if .Mongod.SetParameter.MirrorReads.MaxTimeMs}}
							max_time_ms = {{.Mongod.SetParameter.MirrorReads.MaxTimeMs}}
						{{end}}
					}
				{{end}}
			}
{{end}}
{{if .Mongod.Net}}
      net {
        max_incoming_connections = "{{.Mongod.Net.MaxConnections}}"
      {{if .Mongod.Net.Compressors}}
				compressors = {{.Mongod.Net.Compressors}}
      {{end}}
      }
{{end}}
{{if .Mongod.Storage}}
	storage {
	{{if .Mongod.Storage.Journal}}
      journal {
        commit_interval = "{{.Mongod.Storage.Journal.CommitInterval}}"
      }
	{{end}}
	{{if .Mongod.Storage.WiredTiger}}
      wired_tiger {
        block_compressor = "{{.Mongod.Storage.WiredTiger.Compressor}}"
    	prefix_compression = "{{.Mongod.Storage.WiredTiger.PrefixCompression}}"
        {{if .Mongod.Storage.WiredTiger.CacheSizeGb}}
        	cache_size_gb = "{{.Mongod.Storage.WiredTiger.CacheSizeGb}}"
        {{end}}
        {{if .Mongod.Storage.WiredTiger.CacheSize}}
        	cache_size = "{{.Mongod.Storage.WiredTiger.CacheSize}}"
        {{end}}
      }
	{{end}}
	}
{{end}}
{{if .Mongod.OperationProfiling}}
	operation_profiling {
				{{if .Mongod.OperationProfiling.Mode}}
        		mode = "{{.Mongod.OperationProfiling.Mode}}"
				{{end}}
        slow_op_threshold = "{{.Mongod.OperationProfiling.OpThreshold}}"
				slow_op_sample_rate = "{{.Mongod.OperationProfiling.OpSampleRate}}"
	}
{{end}}
{{if .Mongod.Oplog}}
  oplog {
    {{if .Mongod.Oplog.MaxSizePercent}}max_size_percent = "{{.Mongod.Oplog.MaxSizePercent}}"{{end}}
    {{if .Mongod.Oplog.MinRetentionHours}}min_retention_hours = "{{.Mongod.Oplog.MinRetentionHours}}"{{end}}
  }
{{end}}
{{if .Mongod.ChainingAllowed}}
  chaining_allowed = "{{.Mongod.ChainingAllowed}}"
{{end}}
    }
{{end}}

{{if .Mongos}}
    mongos {
{{if .Mongos.Net}}
      net {
        max_incoming_connections = "{{.Mongos.Net.MaxConnections}}"
				{{if .Mongos.Net.Compressors}}
					compressors = {{.Mongos.Net.Compressors}}
      	{{end}}
      }
{{end}}
{{if .Mongos.AuditLog}}
      audit_log {
        filter = "{{escapeQuotations .Mongos.AuditLog.Filter}}"
      }
{{end}}
{{if .Mongos.ChunkSize}}
      chunk_size = "{{.Mongos.ChunkSize}}"
{{end}}
{{if .Mongos.SetParameter}}
      set_parameter {
        {{if .Mongos.SetParameter.AuditAuthorizationSuccess}}audit_authorization_success = {{.Mongos.SetParameter.AuditAuthorizationSuccess}}{{end}}
        {{if .Mongos.SetParameter.ReadHedgingMode}}read_hedging_mode = "{{.Mongos.SetParameter.ReadHedgingMode}}"{{end}}
        {{if .Mongos.SetParameter.ShardingTaskExecutorPoolMaxSize}}sharding_task_executor_pool_max_size = {{.Mongos.SetParameter.ShardingTaskExecutorPoolMaxSize}}{{end}}
        {{if .Mongos.SetParameter.ShardingTaskExecutorPoolMaxConnecting}}sharding_task_executor_pool_max_connecting = {{.Mongos.SetParameter.ShardingTaskExecutorPoolMaxConnecting}}{{end}}
        {{if .Mongos.SetParameter.ShardingTaskExecutorPoolMinSize}}sharding_task_executor_pool_min_size = {{.Mongos.SetParameter.ShardingTaskExecutorPoolMinSize}}{{end}}
        {{if .Mongos.SetParameter.ShardingTaskExecutorPoolReplicaSetMatching}}sharding_task_executor_pool_replica_set_matching = "{{.Mongos.SetParameter.ShardingTaskExecutorPoolReplicaSetMatching}}"{{end}}
        {{if .Mongos.SetParameter.ShardingTaskExecutorPoolHostTimeoutMs}}sharding_task_executor_pool_host_timeout_ms = {{.Mongos.SetParameter.ShardingTaskExecutorPoolHostTimeoutMs}}{{end}}
        {{if .Mongos.SetParameter.ShardingTaskExecutorPoolRefreshRequirementMs}}sharding_task_executor_pool_refresh_requirement_ms = {{.Mongos.SetParameter.ShardingTaskExecutorPoolRefreshRequirementMs}}{{end}}
        {{if .Mongos.SetParameter.ShardingTaskExecutorPoolRefreshTimeoutMs}}sharding_task_executor_pool_refresh_timeout_ms = {{.Mongos.SetParameter.ShardingTaskExecutorPoolRefreshTimeoutMs}}{{end}}
        {{if .Mongos.SetParameter.ShardingTaskExecutorPoolMaxSizeForConfigServers}}sharding_task_executor_pool_max_size_for_config_servers = {{.Mongos.SetParameter.ShardingTaskExecutorPoolMaxSizeForConfigServers}}{{end}}
        {{if .Mongos.SetParameter.ShardingTaskExecutorPoolMinSizeForConfigServers}}sharding_task_executor_pool_min_size_for_config_servers = {{.Mongos.SetParameter.ShardingTaskExecutorPoolMinSizeForConfigServers}}{{end}}
        {{if .Mongos.SetParameter.WarmMinConnectionsInShardingTaskExecutorPoolOnStartup}}warm_min_connections_in_sharding_task_executor_pool_on_startup = {{.Mongos.SetParameter.WarmMinConnectionsInShardingTaskExecutorPoolOnStartup}}{{end}}
        {{if .Mongos.SetParameter.WarmMinConnectionsInShardingTaskExecutorPoolOnStartupWaitMs}}warm_min_connections_in_sharding_task_executor_pool_on_startup_wait_ms = {{.Mongos.SetParameter.WarmMinConnectionsInShardingTaskExecutorPoolOnStartupWaitMs}}{{end}}
      }
{{end}}
{{if .Mongos.OperationProfiling}}
      operation_profiling {
        slow_op_threshold = "{{.Mongos.OperationProfiling.OpThreshold}}"
        slow_op_sample_rate = "{{.Mongos.OperationProfiling.OpSampleRate}}"
      }
{{end}}
    }
{{end}}

{{if .MongoCfg}}
    mongocfg {
{{if .MongoCfg.Net}}
      net {
        max_incoming_connections = "{{.MongoCfg.Net.MaxConnections}}"
      }
{{end}}
{{if .MongoCfg.OperationProfiling}}
      operation_profiling {
        {{if .MongoCfg.OperationProfiling.Mode}}
        		mode = "{{.MongoCfg.OperationProfiling.Mode}}"
				{{end}}
        slow_op_threshold = "{{.MongoCfg.OperationProfiling.OpThreshold}}"
      }
{{end}}
{{if .MongoCfg.SetParameter}}
      set_parameter {
        {{if .MongoCfg.SetParameter.EnableFlowControl}}enable_flow_control = {{.MongoCfg.SetParameter.EnableFlowControl}}{{end}}
        {{if .MongoCfg.SetParameter.AuditAuthorizationSuccess}}audit_authorization_success = {{.MongoCfg.SetParameter.AuditAuthorizationSuccess}}{{end}}
      }
{{end}}
{{if .MongoCfg.AuditLog}}
      audit_log {
        filter = "{{escapeQuotations .MongoCfg.AuditLog.Filter}}"
      }
{{end}}
{{if .MongoCfg.Oplog}}
      oplog {
        {{if .MongoCfg.Oplog.MaxSizePercent}}max_size_percent = "{{.MongoCfg.Oplog.MaxSizePercent}}"{{end}}
        {{if .MongoCfg.Oplog.MinRetentionHours}}min_retention_hours = "{{.MongoCfg.Oplog.MinRetentionHours}}"{{end}}
      }
{{end}}
{{if .MongoCfg.ChainingAllowed}}
      chaining_allowed = "{{.MongoCfg.ChainingAllowed}}"
{{end}}
    }
{{end}}
  }

{{range $i, $r := .Databases}}
  database {
    name = "{{.}}"
  }
{{- end}}

{{range $i, $r := .Users}}
  user {
    name     = "{{$r.Name}}"
    password = "{{$r.Password}}"
{{range $ii, $rr := $r.Permissions}}
    permission {
      database_name = "{{$rr.DatabaseName}}"
      {{if $rr.Roles -}}
      roles = [{{range $iii, $rrr := $rr.Roles}}{{if $iii}}, {{end}}"{{.}}"{{end}}]
      {{- end}}
    }
{{- end}}
  }
{{- end}}

{{if .Resources}}
  resources {
    resource_preset_id = "{{.Resources.ResourcePresetId}}"
    disk_size          = {{.Resources.DiskSize}}
    disk_type_id       = "{{.Resources.DiskTypeId}}"
  }
{{end}}

{{if .ResourcesMongod}}
  resources_mongod {
    resource_preset_id = "{{.ResourcesMongod.ResourcePresetId}}"
    disk_size          = {{.ResourcesMongod.DiskSize}}
    disk_type_id       = "{{.ResourcesMongod.DiskTypeId}}"
  }
{{end}}

{{if .ResourcesMongoCfg}}
  resources_mongocfg {
    resource_preset_id = "{{.ResourcesMongoCfg.ResourcePresetId}}"
    disk_size          = {{.ResourcesMongoCfg.DiskSize}}
    disk_type_id       = "{{.ResourcesMongoCfg.DiskTypeId}}"
  }
{{end}}

{{if .ResourcesMongos}}
  resources_mongos {
    resource_preset_id = "{{.ResourcesMongos.ResourcePresetId}}"
    disk_size          = {{.ResourcesMongos.DiskSize}}
    disk_type_id       = "{{.ResourcesMongos.DiskTypeId}}"
  }
{{end}}

{{if .ResourcesMongoInfra}}
  resources_mongoinfra {
    resource_preset_id = "{{.ResourcesMongoInfra.ResourcePresetId}}"
    disk_size          = {{.ResourcesMongoInfra.DiskSize}}
    disk_type_id       = "{{.ResourcesMongoInfra.DiskTypeId}}"
  }
{{end}}

{{if .DiskSizeAutoscalingMongod}}
  disk_size_autoscaling_mongod {
    disk_size_limit           = {{.DiskSizeAutoscalingMongod.DiskSizeLimit.Value}}
	{{if .DiskSizeAutoscalingMongod.PlannedUsageThreshold}}
    planned_usage_threshold   = {{.DiskSizeAutoscalingMongod.PlannedUsageThreshold.Value}}
	{{end}}
    emergency_usage_threshold = {{.DiskSizeAutoscalingMongod.EmergencyUsageThreshold.Value}}
  }

{{end}}
{{if .DiskSizeAutoscalingMongoCfg}}
  disk_size_autoscaling_mongocfg {
    disk_size_limit           = {{.DiskSizeAutoscalingMongoCfg.DiskSizeLimit.Value}}
	{{if .DiskSizeAutoscalingMongoCfg.PlannedUsageThreshold}}
    planned_usage_threshold   = {{.DiskSizeAutoscalingMongoCfg.PlannedUsageThreshold.Value}}
	{{end}}
    emergency_usage_threshold = {{.DiskSizeAutoscalingMongoCfg.EmergencyUsageThreshold.Value}}
  }
{{end}}

{{if .DiskSizeAutoscalingMongos}}
  disk_size_autoscaling_mongos {
    disk_size_limit           = {{.DiskSizeAutoscalingMongos.DiskSizeLimit.Value}}
	{{if .DiskSizeAutoscalingMongos.PlannedUsageThreshold}}
    planned_usage_threshold   = {{.DiskSizeAutoscalingMongos.PlannedUsageThreshold.Value}}
	{{end}}
    emergency_usage_threshold = {{.DiskSizeAutoscalingMongos.EmergencyUsageThreshold.Value}}
  }
{{end}}

{{if .DiskSizeAutoscalingMongoInfra}}
  disk_size_autoscaling_mongoinfra {
    disk_size_limit           = {{.DiskSizeAutoscalingMongoInfra.DiskSizeLimit.Value}}
	{{if .DiskSizeAutoscalingMongoInfra.PlannedUsageThreshold}}
    planned_usage_threshold   = {{.DiskSizeAutoscalingMongoInfra.PlannedUsageThreshold.Value}}
	{{end}}
    emergency_usage_threshold = {{.DiskSizeAutoscalingMongoInfra.EmergencyUsageThreshold.Value}}
  }
{{end}}

{{range $i, $r := .Hosts}}
  host {
    zone_id   = "{{$r.ZoneId}}"
    subnet_id = "{{$r.SubnetId}}"
{{if $r.Type}}
	type 	  = "{{$r.Type}}"
{{end}}
{{if $r.ShardName}}
	shard_name 	  = "{{$r.ShardName}}"
{{end}}
{{if $r.AssignPublicIp}}
	assign_public_ip = "{{$r.AssignPublicIp}}"
{{end}}
{{if $r.HostParameters}}
	host_parameters {
{{if hasKey $r.HostParameters "Hidden"}}
		hidden = "{{$r.HostParameters.Hidden}}"
{{end}}
{{if hasKey $r.HostParameters "Priority"}}
		priority = "{{$r.HostParameters.Priority}}"
{{end}}
{{if hasKey $r.HostParameters "Votes"}}
		votes = "{{$r.HostParameters.Votes}}"
{{end}}
{{if hasKey $r.HostParameters "SecondaryDelaySecs"}}
		secondary_delay_secs = "{{$r.HostParameters.SecondaryDelaySecs}}"
{{end}}
{{if $r.HostParameters.Tags}}
		tags = {
			{{- range $k, $v := $r.HostParameters.Tags}}
				{{$k}} = "{{$v}}"
			{{- end}}
		}
{{end}}
	}
{{end}}

  }
{{end}}

  security_group_ids = [{{range $i, $r := .SecurityGroupIds}}{{if $i}}, {{end}}"{{.}}"{{end}}]


  maintenance_window {
    type = "{{.MaintenanceWindow.Type}}"
    {{with .MaintenanceWindow.Day}}day  = "{{.}}"{{end}}
    {{with .MaintenanceWindow.Hour}}hour = {{.}}{{end}}
  }

  {{if isNotNil .DeletionProtection}}deletion_protection = {{.DeletionProtection}}{{end}}
}
`
)
