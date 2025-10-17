package mdb_redis_cluster_v2

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Cluster struct {
	//----Attributes----
	ID                  types.String `tfsdk:"id"`
	ClusterID           types.String `tfsdk:"cluster_id"`
	Name                types.String `tfsdk:"name"`
	NetworkID           types.String `tfsdk:"network_id"`
	Environment         types.String `tfsdk:"environment"`
	Description         types.String `tfsdk:"description"`
	Sharded             types.Bool   `tfsdk:"sharded"`
	TlsEnabled          types.Bool   `tfsdk:"tls_enabled"`
	PersistenceMode     types.String `tfsdk:"persistence_mode"`
	AnnounceHostnames   types.Bool   `tfsdk:"announce_hostnames"`
	FolderID            types.String `tfsdk:"folder_id"`
	CreatedAt           types.String `tfsdk:"created_at"`
	DeletionProtection  types.Bool   `tfsdk:"deletion_protection"`
	AuthSentinel        types.Bool   `tfsdk:"auth_sentinel"`
	DiskEncryptionKeyId types.String `tfsdk:"disk_encryption_key_id"`

	Labels              types.Map    `tfsdk:"labels"`
	SecurityGroupIDs    types.Set    `tfsdk:"security_group_ids"`
	HostSpecs           types.Map    `tfsdk:"hosts"`
	Access              types.Object `tfsdk:"access"`
	DiskSizeAutoscaling types.Object `tfsdk:"disk_size_autoscaling"`
	MaintenanceWindow   types.Object `tfsdk:"maintenance_window"`
	Resources           types.Object `tfsdk:"resources"`
	Modules             types.Object `tfsdk:"modules"`

	Config   *Config        `tfsdk:"config"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type Access struct {
	DataLens types.Bool `tfsdk:"data_lens"`
	WebSql   types.Bool `tfsdk:"web_sql"`
}
type MaintenanceWindow struct {
	Type types.String `tfsdk:"type"`
	Day  types.String `tfsdk:"day"`
	Hour types.Int64  `tfsdk:"hour"`
}

var MaintenanceWindowType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type": types.StringType,
		"day":  types.StringType,
		"hour": types.Int64Type,
	},
}

var AccessType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"data_lens": types.BoolType,
		"web_sql":   types.BoolType,
	},
}

type DiskSizeAutoscaling struct {
	DiskSizeLimit           types.Int64 `tfsdk:"disk_size_limit"`
	PlannedUsageThreshold   types.Int64 `tfsdk:"planned_usage_threshold"`
	EmergencyUsageThreshold types.Int64 `tfsdk:"emergency_usage_threshold"`
}

var DiskSizeAutoscalingType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"disk_size_limit":           types.Int64Type,
		"planned_usage_threshold":   types.Int64Type,
		"emergency_usage_threshold": types.Int64Type,
	},
}

type ValkeyModules struct {
	ValkeySearch *ValkeySearch `tfsdk:"valkey_search"`
	ValkeyJson   *ValkeyJson   `tfsdk:"valkey_json"`
	ValkeyBloom  *ValkeyBloom  `tfsdk:"valkey_bloom"`
}

type ValkeySearch struct {
	Enabled       types.Bool  `tfsdk:"enabled"`
	ReaderThreads types.Int64 `tfsdk:"reader_threads"`
	WriterThreads types.Int64 `tfsdk:"writer_threads"`
}

type ValkeyJson struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type ValkeyBloom struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

var ValkeyModulesType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"valkey_search": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"enabled":        types.BoolType,
				"reader_threads": types.Int64Type,
				"writer_threads": types.Int64Type,
			},
		},
		"valkey_json": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"enabled": types.BoolType,
			},
		},
		"valkey_bloom": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"enabled": types.BoolType,
			},
		},
	},
}

type Config struct {
	Password                        types.String `tfsdk:"password"`
	Timeout                         types.Int64  `tfsdk:"timeout"`
	MaxmemoryPolicy                 types.String `tfsdk:"maxmemory_policy"`
	NotifyKeyspaceEvents            types.String `tfsdk:"notify_keyspace_events"`
	SlowlogLogSlowerThan            types.Int64  `tfsdk:"slowlog_log_slower_than"`
	SlowlogMaxLen                   types.Int64  `tfsdk:"slowlog_max_len"`
	Databases                       types.Int64  `tfsdk:"databases"`
	MaxmemoryPercent                types.Int64  `tfsdk:"maxmemory_percent"`
	ClientOutputBufferLimitNormal   types.String `tfsdk:"client_output_buffer_limit_normal"`
	ClientOutputBufferLimitPubsub   types.String `tfsdk:"client_output_buffer_limit_pubsub"`
	UseLuajit                       types.Bool   `tfsdk:"use_luajit"`
	IoThreadsAllowed                types.Bool   `tfsdk:"io_threads_allowed"`
	Version                         types.String `tfsdk:"version"`
	LuaTimeLimit                    types.Int64  `tfsdk:"lua_time_limit"`
	ReplBacklogSizePercent          types.Int64  `tfsdk:"repl_backlog_size_percent"`
	ClusterRequireFullCoverage      types.Bool   `tfsdk:"cluster_require_full_coverage"`
	ClusterAllowReadsWhenDown       types.Bool   `tfsdk:"cluster_allow_reads_when_down"`
	ClusterAllowPubsubshardWhenDown types.Bool   `tfsdk:"cluster_allow_pubsubshard_when_down"`
	LfuDecayTime                    types.Int64  `tfsdk:"lfu_decay_time"`
	LfuLogFactor                    types.Int64  `tfsdk:"lfu_log_factor"`
	TurnBeforeSwitchover            types.Bool   `tfsdk:"turn_before_switchover"`
	AllowDataLoss                   types.Bool   `tfsdk:"allow_data_loss"`
	BackupRetainPeriodDays          types.Int64  `tfsdk:"backup_retain_period_days"`
	BackupWindowStart               types.Object `tfsdk:"backup_window_start"`
	ZsetMaxListpackEntries          types.Int64  `tfsdk:"zset_max_listpack_entries"`
}
