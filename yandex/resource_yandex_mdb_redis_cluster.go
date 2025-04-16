package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1/config"
)

const (
	yandexMDBRedisClusterCreateTimeout = 45 * time.Minute
	yandexMDBRedisClusterUpdateTimeout = 60 * time.Minute
	yandexMDBRedisClusterDeleteTimeout = 20 * time.Minute
	defaultMDBPageSize                 = 1000
)

func resourceYandexMDBRedisCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Redis cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-redis/concepts).",

		Create: resourceYandexMDBRedisClusterCreate,
		Read:   resourceYandexMDBRedisClusterRead,
		Update: resourceYandexMDBRedisClusterUpdate,
		Delete: resourceYandexMDBRedisClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBRedisClusterCreateTimeout),
			Update: schema.DefaultTimeout(yandexMDBRedisClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBRedisClusterDeleteTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Required:    true,
				ForceNew:    true,
			},
			"environment": {
				Type:         schema.TypeString,
				Description:  "Deployment environment of the Redis cluster. Can be either `PRESTABLE` or `PRODUCTION`.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseRedisEnv),
			},
			"config": {
				Type:        schema.TypeList,
				Description: "Configuration of the Redis cluster.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"password": {
							Type:        schema.TypeString,
							Description: "Password for the Redis cluster.",
							Required:    true,
							Sensitive:   true,
						},
						"timeout": {
							Type:        schema.TypeInt,
							Description: "Close the connection after a client is idle for N seconds.",
							Optional:    true,
							Computed:    true,
						},
						"maxmemory_policy": {
							Type:        schema.TypeString,
							Description: "Redis key eviction policy for a dataset that reaches maximum memory. Can be any of the listed in [the official RedisDB documentation](https://docs.redislabs.com/latest/rs/administering/database-operations/eviction-policy/).",
							Optional:    true,
							Computed:    true,
						},
						"notify_keyspace_events": {
							Type:        schema.TypeString,
							Description: "Select the events that Redis will notify among a set of classes.",
							Optional:    true,
							Computed:    true,
						},
						"slowlog_log_slower_than": {
							Type:        schema.TypeInt,
							Description: "Log slow queries below this number in microseconds.",
							Optional:    true,
							Computed:    true,
						},
						"slowlog_max_len": {
							Type:        schema.TypeInt,
							Description: "Slow queries log length.",
							Optional:    true,
							Computed:    true,
						},
						"databases": {
							Type:        schema.TypeInt,
							Description: "Number of databases (changing requires redis-server restart).",
							Optional:    true,
							Computed:    true,
						},
						"maxmemory_percent": {
							Type:        schema.TypeInt,
							Description: "Redis maxmemory usage in percent",
							Optional:    true,
						},
						"client_output_buffer_limit_normal": {
							Type:        schema.TypeString,
							Description: "Normal clients output buffer limits. See [redis config file](https://github.com/redis/redis/blob/6.2/redis.conf#L1841).",
							Optional:    true,
							Computed:    true,
						},
						"client_output_buffer_limit_pubsub": {
							Type:        schema.TypeString,
							Description: "Pubsub clients output buffer limits. See [redis config file](https://github.com/redis/redis/blob/6.2/redis.conf#L1843).",
							Optional:    true,
							Computed:    true,
						},
						"use_luajit": {
							Type:        schema.TypeBool,
							Description: "Use JIT for lua scripts and functions.",
							Optional:    true,
							Computed:    true,
						},
						"io_threads_allowed": {
							Type:        schema.TypeBool,
							Description: "Allow Redis to use io-threads.",
							Optional:    true,
							Computed:    true,
						},
						"version": {
							Type:        schema.TypeString,
							Description: "Version of Redis.",
							Required:    true,
						},
						"lua_time_limit": {
							Type:        schema.TypeInt,
							Description: "Maximum time in milliseconds for Lua scripts.",
							Optional:    true,
						},
						"repl_backlog_size_percent": {
							Type:        schema.TypeInt,
							Description: "Replication backlog size as a percentage of flavor maxmemory.",
							Optional:    true,
						},
						"cluster_require_full_coverage": {
							Type:        schema.TypeBool,
							Description: "Controls whether all hash slots must be covered by nodes.",
							Optional:    true,
						},
						"cluster_allow_reads_when_down": {
							Type:        schema.TypeBool,
							Description: "Allows read operations when cluster is down.",
							Optional:    true,
						},
						"cluster_allow_pubsubshard_when_down": {
							Type:        schema.TypeBool,
							Description: "Permits Pub/Sub shard operations when cluster is down.",
							Optional:    true,
						},
						"lfu_decay_time": {
							Type:        schema.TypeInt,
							Description: "The time, in minutes, that must elapse in order for the key counter to be divided by two (or decremented if it has a value less <= 10).",
							Optional:    true,
						},
						"lfu_log_factor": {
							Type:        schema.TypeInt,
							Description: "Determines how the frequency counter represents key hits.",
							Optional:    true,
						},
						"turn_before_switchover": {
							Type:        schema.TypeBool,
							Description: "Allows to turn before switchover in RDSync.",
							Optional:    true,
						},
						"allow_data_loss": {
							Type:        schema.TypeBool,
							Description: "Allows some data to be lost in favor of faster switchover/restart by RDSync.",
							Optional:    true,
						},
						"zset_max_listpack_entries": {
							Type:        schema.TypeInt,
							Description: "Controls max number of entries in zset before conversion from memory-efficient listpack to CPU-efficient hash table and skiplist",
							Optional:    true,
						},
						"backup_window_start": {
							Type:        schema.TypeList,
							Description: "Time to start the daily backup, in the UTC timezone.",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hours": {
										Type:         schema.TypeInt,
										Description:  "The hour at which backup will be started.",
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 23),
									},
									"minutes": {
										Type:         schema.TypeInt,
										Description:  "The minute at which backup will be started.",
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 59),
									},
								},
							},
						},
					},
				},
			},
			"resources": {
				Type:        schema.TypeList,
				Description: "Resources allocated to hosts of the Redis cluster.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: "The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-redis/concepts).",
							Required:    true,
						},
						"disk_size": {
							Type:        schema.TypeInt,
							Description: "Volume of the storage available to a host, in gigabytes.",
							Required:    true,
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: "Type of the storage of Redis hosts - environment default is used if missing.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"disk_size_autoscaling": {
				Type:        schema.TypeList,
				Description: "Disk size autoscaling settings.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_size_limit": {
							Type:        schema.TypeInt,
							Description: "Limit of disk size after autoscaling (GiB).",
							Required:    true,
						},
						"planned_usage_threshold": {
							Type:        schema.TypeInt,
							Description: "Maintenance window autoscaling disk usage (percent).",
							Optional:    true,
						},
						"emergency_usage_threshold": {
							Type:        schema.TypeInt,
							Description: "Immediate autoscaling disk usage (percent).",
							Optional:    true,
						},
					},
				},
			},
			"access": {
				Type:        schema.TypeList,
				Description: "Access policy to the Redis cluster.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_lens": {
							Type:        schema.TypeBool,
							Description: "Allow access for DataLens. Can be either `true` or `false`.",
							Optional:    true,
							Computed:    true,
						},
						"web_sql": {
							Type:        schema.TypeBool,
							Description: "Allow access for Web SQL. Can be either `true` or `false`.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"host": {
				Type:        schema.TypeList,
				Description: "A host of the Redis cluster.",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:        schema.TypeString,
							Description: common.ResourceDescriptions["zone"],
							Required:    true,
						},
						"shard_name": {
							Type:        schema.TypeString,
							Description: "The name of the shard to which the host belongs.",
							Optional:    true,
							Computed:    true,
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Description: "The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.",
							Optional:    true,
							Computed:    true,
						},
						"fqdn": {
							Type:        schema.TypeString,
							Description: "The fully qualified domain name of the host.",
							Computed:    true,
						},
						"replica_priority": {
							Type:        schema.TypeInt,
							Description: "Replica priority of a current replica (usable for non-sharded only).",
							Optional:    true,
							Default:     defaultReplicaPriority,
						},
						"assign_public_ip": {
							Type:        schema.TypeBool,
							Description: "Sets whether the host should get a public IP address or not.",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"sharded": {
				Type:        schema.TypeBool,
				Description: "Redis Cluster mode enabled/disabled. Enables sharding when cluster non-sharded. If cluster is sharded - disabling is not allowed.",
				Optional:    true,
				Default:     false,
				ForceNew:    false,
			},
			"tls_enabled": {
				Type:        schema.TypeBool,
				Description: "TLS support mode enabled/disabled.",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"persistence_mode": {
				Type:         schema.TypeString,
				Description:  "Persistence mode. Possible values: `ON`, `OFF`.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateParsableValue(parsePersistenceMode),
			},
			"announce_hostnames": {
				Type:        schema.TypeBool,
				Description: "Announce fqdn instead of ip address.",
				Optional:    true,
			},
			"auth_sentinel": {
				Type:        schema.TypeBool,
				Description: "Allows to use ACL users to auth in sentinel",
				Optional:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: "Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-redis/api-ref/Cluster/).",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-redis/api-ref/Cluster/).",
				Computed:    true,
			},
			"security_group_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["security_group_ids"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
			},
			"maintenance_window": {
				Type:        schema.TypeList,
				Description: "Maintenance window settings.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Description:  "Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.",
							ValidateFunc: validation.StringInSlice([]string{"ANYTIME", "WEEKLY"}, false),
							Required:     true,
						},
						"day": {
							Type:         schema.TypeString,
							Description:  "Day of week for maintenance window if window type is weekly. Possible values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.",
							ValidateFunc: validateParsableValue(parseRedisWeekDay),
							Optional:     true,
						},
						"hour": {
							Type:         schema.TypeInt,
							Description:  "Hour of day in UTC time zone (1-24) for maintenance window if window type is weekly.",
							ValidateFunc: validation.IntBetween(1, 24),
							Optional:     true,
						},
					},
				},
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceYandexMDBRedisClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareCreateRedisRequest(d, config)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Redis().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Redis Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get redis create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*redis.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)
	log.Printf("[DEBUG] Creating Redis Cluster %q", md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create Redis Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Redis Cluster creation failed: %s", err)
	}

	return resourceYandexMDBRedisClusterRead(d, meta)
}

func prepareCreateRedisRequest(d *schema.ResourceData, meta *Config) (*redis.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	sharded := d.Get("sharded").(bool)

	if err != nil {
		return nil, fmt.Errorf("Error while expanding labels on Redis Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating Redis Cluster: %s", err)
	}

	hosts, err := expandRedisHosts(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding hosts on Redis Cluster create: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseRedisEnv(e)
	if err != nil {
		return nil, fmt.Errorf("Error resolving environment while creating Redis Cluster: %s", err)
	}

	conf, version, err := expandRedisConfig(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding config while creating Redis Cluster: %s", err)
	}

	resources, err := expandRedisResources(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding resources on Redis Cluster create: %s", err)
	}

	dsa, err := expandRedisDiskSizeAutoscaling(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding disk size autoscaling on Redis Cluster create: %s", err)
	}

	configSpec := &redis.ConfigSpec{
		Redis:               conf,
		Resources:           resources,
		Version:             version,
		DiskSizeAutoscaling: dsa,
		Access:              expandRedisAccess(d),
	}

	configSpec.BackupWindowStart = expandMDBBackupWindowStart(d, "config.0.backup_window_start.0")

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding network id on Redis Cluster create: %s", err)
	}

	persistenceMode, err := parsePersistenceMode(d.Get("persistence_mode"))
	if err != nil {
		return nil, fmt.Errorf("Error resolving persistence_mode while creating Redis Cluster: %s", err)
	}

	mw, err := expandRedisMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding maintenance window on Redis Cluster create: %s", err)
	}

	req := redis.CreateClusterRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		NetworkId:          networkID,
		Environment:        env,
		ConfigSpec:         configSpec,
		HostSpecs:          hosts,
		Labels:             labels,
		Sharded:            sharded,
		TlsEnabled:         &wrappers.BoolValue{Value: d.Get("tls_enabled").(bool)},
		PersistenceMode:    persistenceMode,
		AnnounceHostnames:  d.Get("announce_hostnames").(bool),
		SecurityGroupIds:   securityGroupIds,
		DeletionProtection: d.Get("deletion_protection").(bool),
		MaintenanceWindow:  mw,
		AuthSentinel:       d.Get("auth_sentinel").(bool),
	}
	return &req, nil
}

func resourceYandexMDBRedisClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().Redis().Cluster().Get(ctx, &redis.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	hosts, err := listRedisHosts(ctx, config, d)
	if err != nil {
		return err
	}

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)
	d.Set("sharded", cluster.Sharded)
	d.Set("tls_enabled", cluster.TlsEnabled)
	d.Set("persistence_mode", cluster.GetPersistenceMode().String())
	d.Set("announce_hostnames", cluster.AnnounceHostnames)
	d.Set("auth_sentinel", cluster.AuthSentinel)

	resources, err := flattenRedisResources(cluster.Config.Resources)
	if err != nil {
		return err
	}

	ac := flattenRedisAccess(cluster.GetConfig().GetAccess())
	if err := d.Set("access", ac); err != nil {
		return err
	}

	dsa, err := flattenRedisDiskSizeAutoscaling(cluster.Config.DiskSizeAutoscaling)
	if err != nil {
		return err
	}

	conf := extractRedisConfig(cluster.Config)
	password := ""
	if v, ok := d.GetOk("config.0.password"); ok {
		password = v.(string)
	}

	err = d.Set("config", []map[string]interface{}{
		{
			"timeout":                             conf.timeout,
			"maxmemory_policy":                    conf.maxmemoryPolicy,
			"notify_keyspace_events":              conf.notifyKeyspaceEvents,
			"slowlog_log_slower_than":             conf.slowlogLogSlowerThan,
			"slowlog_max_len":                     conf.slowlogMaxLen,
			"databases":                           conf.databases,
			"maxmemory_percent":                   conf.maxmemoryPercent,
			"version":                             conf.version,
			"password":                            password,
			"client_output_buffer_limit_normal":   conf.clientOutputBufferLimitNormal,
			"client_output_buffer_limit_pubsub":   conf.clientOutputBufferLimitPubsub,
			"lua_time_limit":                      conf.luaTimeLimit,
			"repl_backlog_size_percent":           conf.replBacklogSizePercent,
			"cluster_require_full_coverage":       conf.clusterRequireFullCoverage,
			"cluster_allow_reads_when_down":       conf.clusterAllowReadsWhenDown,
			"cluster_allow_pubsubshard_when_down": conf.clusterAllowPubsubshardWhenDown,
			"lfu_decay_time":                      conf.lfuDecayTime,
			"lfu_log_factor":                      conf.lfuLogFactor,
			"turn_before_switchover":              conf.turnBeforeSwitchover,
			"allow_data_loss":                     conf.allowDataLoss,
			"use_luajit":                          conf.useLuajit,
			"zset_max_listpack_entries":           conf.zsetMaxListpackEntries,
			"io_threads_allowed":                  conf.ioThreadsAllowed,
			"backup_window_start":                 flattenMDBBackupWindowStart(cluster.GetConfig().GetBackupWindowStart()),
		},
	})
	if err != nil {
		return err
	}

	if err := d.Set("resources", resources); err != nil {
		return err
	}

	if err := d.Set("disk_size_autoscaling", dsa); err != nil {
		return err
	}

	// Do not change the state if only order of hosts differs.
	dHosts, err := expandRedisHosts(d)
	if err != nil {
		return err
	}

	sortRedisHosts(cluster.Sharded, hosts, dHosts)

	hs, err := flattenRedisHosts(cluster.Sharded, hosts)
	if err != nil {
		return err
	}

	if err := d.Set("host", hs); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	mw := flattenRedisMaintenanceWindow(cluster.MaintenanceWindow)
	if err := d.Set("maintenance_window", mw); err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.DeletionProtection)

	return d.Set("labels", cluster.Labels)
}

func resourceYandexMDBRedisClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)
	if err := setRedisFolderID(d, meta); err != nil {
		return err
	}

	if d.HasChange("resources.0.disk_type_id") {
		return fmt.Errorf("Changing disk_type_id is not supported for Redis Cluster. Id: %v", d.Id())
	}
	config := meta.(*Config)

	if d.HasChange("sharded") {
		if !d.Get("sharded").(bool) {
			return fmt.Errorf("Disabling sharding on Redis Cluster is not supported, Id: %q", d.Id())
		}
		err := enableShardingRedis(context.Background(), config, d)
		if err != nil {
			return err
		}
	}

	if err := updateRedisClusterParams(d, meta); err != nil {
		return err
	}

	if err := updateRedisClusterHosts(d, meta); err != nil {
		return err
	}

	d.Partial(false)
	return resourceYandexMDBRedisClusterRead(d, meta)
}

func updateRedisClusterParams(d *schema.ResourceData, meta interface{}) error {
	req := &redis.UpdateClusterRequest{
		ClusterId: d.Id(),
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{},
		},
	}
	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")

	}

	if d.HasChange("persistence_mode") {
		mode, err := parsePersistenceMode(d.Get("persistence_mode"))
		if err != nil {
			return err
		}

		req.PersistenceMode = mode
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "persistence_mode")

	}

	if d.HasChange("announce_hostnames") {
		req.AnnounceHostnames = d.Get("announce_hostnames").(bool)

		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "announce_hostnames")

	}

	if d.HasChange("auth_sentinel") {
		req.AuthSentinel = d.Get("auth_sentinel").(bool)

		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "auth_sentinel")
	}

	if d.HasChange("labels") {
		labelsProp, err := expandLabels(d.Get("labels"))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")

	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")

	}

	if d.HasChange("resources") {
		res, err := expandRedisResources(d)
		if err != nil {
			return err
		}

		if req.ConfigSpec == nil {
			req.ConfigSpec = &redis.ConfigSpec{}
		}

		req.ConfigSpec.Resources = res
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.resources")

	}

	if d.HasChange("disk_size_autoscaling") {
		res, err := expandRedisDiskSizeAutoscaling(d)
		if err != nil {
			return err
		}

		if req.ConfigSpec == nil {
			req.ConfigSpec = &redis.ConfigSpec{}
		}

		req.ConfigSpec.DiskSizeAutoscaling = res
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.disk_size_autoscaling")
	}

	var password string
	if d.HasChange("config") {
		conf, ver, err := expandRedisConfig(d)
		if err != nil {
			return err
		}

		if req.ConfigSpec == nil {
			req.ConfigSpec = &redis.ConfigSpec{}
		}

		// Password change cannot be mixed with other updates
		if conf.Password != "" {
			password = conf.Password
			conf.Password = ""
		}

		req.ConfigSpec.Redis = conf
		fields := [...]string{
			"timeout",
			"maxmemory_policy",
			"notify_keyspace_events",
			"slowlog_log_slower_than",
			"slowlog_max_len",
			"databases",
			"maxmemory_percent",
			"client_output_buffer_limit_normal",
			"client_output_buffer_limit_pubsub",
			"lua_time_limit",
			"repl_backlog_size_percent",
			"cluster_require_full_coverage",
			"cluster_allow_reads_when_down",
			"cluster_allow_pubsubshard_when_down",
			"lfu_decay_time",
			"lfu_log_factor",
			"turn_before_switchover",
			"allow_data_loss",
			"use_luajit",
			"io_threads_allowed",
			"zset_max_listpack_entries",
		}
		for _, field := range fields {
			fullPath := "config_spec.redis." + field
			if d.HasChange("config.0." + field) {
				req.UpdateMask.Paths = append(req.UpdateMask.Paths, fullPath)
			}
		}
		if d.HasChange("config.0.version") {
			req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.version")
			req.ConfigSpec.Version = ver
		}
		if d.HasChange("config.0.backup_window_start") {
			if req.ConfigSpec == nil {
				req.ConfigSpec = &redis.ConfigSpec{}
			}
			req.ConfigSpec.BackupWindowStart = expandMDBBackupWindowStart(d, "config.0.backup_window_start.0")
			req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.backup_window_start")
		}
	}

	if d.HasChange("security_group_ids") {
		securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

		req.SecurityGroupIds = securityGroupIds
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "security_group_ids")

	}

	if d.HasChange("deletion_protection") {
		req.DeletionProtection = d.Get("deletion_protection").(bool)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "deletion_protection")

	}

	if d.HasChange("maintenance_window") {
		mw, err := expandRedisMaintenanceWindow(d)
		if err != nil {
			return err
		}
		req.MaintenanceWindow = mw
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "maintenance_window")

	}
	if d.HasChange("access") {
		if req.ConfigSpec == nil {
			req.ConfigSpec = &redis.ConfigSpec{}
		}
		req.ConfigSpec.Access = expandRedisAccess(d)
		if d.HasChange("access.0.web_sql") {
			req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.access.web_sql")
		}
		if d.HasChange("access.0.data_lens") {
			req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.access.data_lens")
		}
	}

	if len(req.UpdateMask.Paths) == 0 && password == "" {
		return nil
	} else if len(req.UpdateMask.Paths) != 0 {
		err := makeRedisClusterUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}
	}

	// Password change cannot be mixed with other updates
	if d.HasChange("config.0.password") && password != "" {
		reqPasswordUpdate := &redis.UpdateClusterRequest{
			ClusterId: d.Id(),
			ConfigSpec: &redis.ConfigSpec{
				Redis: &config.RedisConfig{Password: password},
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"config_spec.redis.password"},
			},
		}
		err := makeRedisClusterUpdateRequest(reqPasswordUpdate, d, meta)
		if err != nil {
			return err
		}
	}

	return nil
}

func addHosts(ctx context.Context, d *schema.ResourceData, config *Config, sharded bool, currShards []*redis.Shard,
	toAdd map[string][]*redis.HostSpec) error {
	var err error
	for shardName, specs := range toAdd {
		shardExists := false
		for _, s := range currShards {
			if s.Name == shardName {
				shardExists = true
				break
			}
		}
		if sharded && !shardExists {
			err = createRedisShard(ctx, config, d, shardName, specs)
			if err != nil {
				return err
			}
		} else {
			err = createRedisHosts(ctx, config, d, specs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func enableShardingRedis(ctx context.Context, config *Config, d *schema.ResourceData) error {
	op, err := config.sdk.WrapOperation(config.sdk.MDB().Redis().Cluster().EnableSharding(ctx, &redis.EnableShardingClusterRequest{ClusterId: d.Id()}))
	if err != nil {
		return fmt.Errorf("error while requesting API to enable sharding Redis Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while enabling sharding Redis Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteHosts(ctx context.Context, d *schema.ResourceData, config *Config, sharded bool, targetHosts []*redis.HostSpec,
	toDelete map[string][]string) error {
	var err error
	for shardName, fqdns := range toDelete {
		deleteShard := true
		for _, th := range targetHosts {
			if th.ShardName == shardName {
				deleteShard = false
				break
			}
		}
		if sharded && deleteShard {
			err = deleteRedisShard(ctx, config, d, shardName)
			if err != nil {
				return err
			}
		} else {
			err = deleteRedisHosts(ctx, config, d, fqdns)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func updateRedisClusterHosts(d *schema.ResourceData, meta interface{}) error {
	if !d.HasChange("host") {
		return nil
	}

	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	sharded := d.Get("sharded").(bool)

	currHosts, err := listRedisHosts(ctx, config, d)
	if err != nil {
		return err
	}

	targetHosts, err := expandRedisHosts(d)
	if err != nil {
		return fmt.Errorf("Error while expanding hosts on Redis Cluster create: %s", err)
	}

	currShards, err := listRedisShards(ctx, config, d)
	if err != nil {
		return err
	}

	toDelete, toUpdate, toAdd, err := redisHostsDiff(sharded, currHosts, targetHosts)
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	err = addHosts(ctx, d, config, sharded, currShards, toAdd)
	if err != nil {
		return err
	}

	err = updateHosts(ctx, d, config, toUpdate)
	if err != nil {
		return err
	}

	err = deleteHosts(ctx, d, config, sharded, targetHosts, toDelete)
	if err != nil {
		return err
	}

	return nil
}

//nolint:unused
func updateRedisMaintenanceWindow(ctx context.Context, config *Config, d *schema.ResourceData, mw *redis.MaintenanceWindow) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Redis().Cluster().Update(ctx, &redis.UpdateClusterRequest{
			ClusterId:         d.Id(),
			MaintenanceWindow: mw,
			UpdateMask:        &field_mask.FieldMask{Paths: []string{"maintenance_window"}},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update maintenance window in Redis Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating maintenance window in Redis Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func listRedisHosts(ctx context.Context, config *Config, d *schema.ResourceData) ([]*redis.Host, error) {
	hosts := []*redis.Host{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Redis().Cluster().ListHosts(ctx, &redis.ListClusterHostsRequest{
			ClusterId: d.Id(),
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of hosts for '%s': %s", d.Id(), err)
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return hosts, nil
}

func listRedisShards(ctx context.Context, config *Config, d *schema.ResourceData) ([]*redis.Shard, error) {
	shards := []*redis.Shard{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Redis().Cluster().ListShards(ctx, &redis.ListClusterShardsRequest{
			ClusterId: d.Id(),
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of shards for '%s': %s", d.Id(), err)
		}
		shards = append(shards, resp.Shards...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return shards, nil
}

func createRedisShard(ctx context.Context, config *Config, d *schema.ResourceData, shardName string, hostSpecs []*redis.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Redis().Cluster().AddShard(ctx, &redis.AddClusterShardRequest{
			ClusterId: d.Id(),
			ShardName: shardName,
			HostSpecs: hostSpecs,
		}),
	)
	if err != nil {
		return fmt.Errorf("Error while requesting API to add shard to Redis Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while adding shard to Redis Cluster %q: %s", d.Id(), err)
	}
	op, err = config.sdk.WrapOperation(
		config.sdk.MDB().Redis().Cluster().Rebalance(ctx, &redis.RebalanceClusterRequest{
			ClusterId: d.Id(),
		}),
	)
	if err != nil {
		return fmt.Errorf("Error while requesting API to rebalance the Redis Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while rebalancing the Redis Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createRedisHosts(ctx context.Context, config *Config, d *schema.ResourceData, specs []*redis.HostSpec) error {
	for _, hs := range specs {
		op, err := config.sdk.WrapOperation(
			config.sdk.MDB().Redis().Cluster().AddHosts(ctx, &redis.AddClusterHostsRequest{
				ClusterId: d.Id(),
				HostSpecs: []*redis.HostSpec{hs},
			}),
		)
		if err != nil {
			return fmt.Errorf("Error while requesting API to add host to Redis Cluster %q: %s", d.Id(), err)
		}
		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error while adding host to Redis Cluster %q: %s", d.Id(), err)
		}
	}
	return nil
}

type HostUpdateInfo struct {
	HostName        string
	ReplicaPriority *wrappers.Int64Value
	AssignPublicIp  bool
	UpdateMask      *field_mask.FieldMask
}

func getHostUpdateInfo(sharded bool, fqdn string, oldPriority *wrapperspb.Int64Value, oldAssignPublicIp bool,
	newPriority *wrapperspb.Int64Value, newAssignPublicIp bool) (*HostUpdateInfo, error) {
	var maskPaths []string
	if newPriority != nil && oldPriority != nil && oldPriority.Value != newPriority.Value {
		if sharded {
			return nil, fmt.Errorf("modifying replica priority in hosts of sharded clusters is not supported: %s", fqdn)
		}
		maskPaths = append(maskPaths, "replica_priority")
	}
	if oldAssignPublicIp != newAssignPublicIp {
		maskPaths = append(maskPaths, "assign_public_ip")
	}

	if len(maskPaths) == 0 {
		return nil, nil
	}
	res := &HostUpdateInfo{
		HostName:        fqdn,
		ReplicaPriority: newPriority,
		AssignPublicIp:  newAssignPublicIp,
		UpdateMask:      &field_mask.FieldMask{Paths: maskPaths},
	}
	return res, nil
}

func updateRedisHost(ctx context.Context, config *Config, d *schema.ResourceData, host *HostUpdateInfo) error {
	request := &redis.UpdateClusterHostsRequest{
		ClusterId: d.Id(),
		UpdateHostSpecs: []*redis.UpdateHostSpec{
			{
				HostName:        host.HostName,
				AssignPublicIp:  host.AssignPublicIp,
				ReplicaPriority: host.ReplicaPriority,
				UpdateMask:      host.UpdateMask,
			},
		},
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending Redis cluster update hosts request: %+v", request)
		return config.sdk.MDB().Redis().Cluster().UpdateHosts(ctx, request)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to update host for Redis Cluster %q - host %v: %s", d.Id(), host.HostName, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating host for Redis Cluster %q - host %v: %s", d.Id(), host.HostName, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating host for Redis Cluster %q - host %v failed: %s", d.Id(), host.HostName, err)
	}

	return nil
}

func updateHosts(ctx context.Context, d *schema.ResourceData, config *Config, specs map[string][]*HostUpdateInfo) error {
	for _, hostInfos := range specs {
		for _, hostInfo := range hostInfos {
			if err := updateRedisHost(ctx, config, d, hostInfo); err != nil {
				return err
			}
		}
	}

	return nil
}

func deleteRedisShard(ctx context.Context, config *Config, d *schema.ResourceData, shardName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Redis().Cluster().DeleteShard(ctx, &redis.DeleteClusterShardRequest{
			ClusterId: d.Id(),
			ShardName: shardName,
		}),
	)
	if err != nil {
		return fmt.Errorf("Error while requesting API to delete shard from Redis Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while deleting shard from Redis Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteRedisHosts(ctx context.Context, config *Config, d *schema.ResourceData, fqdns []string) error {
	for _, fqdn := range fqdns {
		op, err := config.sdk.WrapOperation(
			config.sdk.MDB().Redis().Cluster().DeleteHosts(ctx, &redis.DeleteClusterHostsRequest{
				ClusterId: d.Id(),
				HostNames: []string{fqdn},
			}),
		)
		if err != nil {
			return fmt.Errorf("Error while requesting API to delete host %s from Redis Cluster %q: %s", fqdn, d.Id(), err)
		}
		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error while deleting host %s from Redis Cluster %q: %s", fqdn, d.Id(), err)
		}
	}
	return nil
}

func makeRedisClusterUpdateRequest(req *redis.UpdateClusterRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Redis().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Redis Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Redis Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func resourceYandexMDBRedisClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Redis Cluster %q", d.Id())

	req := &redis.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Redis().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Redis Cluster %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Redis Cluster %q", d.Id())
	return nil
}

func setRedisFolderID(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().Redis().Cluster().Get(ctx, &redis.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Id()))
	}

	folderID, ok := d.GetOk("folder_id")
	if !ok {
		return nil
	}
	if folderID == "" {
		return nil
	}

	if cluster.FolderId != folderID {
		request := &redis.MoveClusterRequest{
			ClusterId:           d.Id(),
			DestinationFolderId: folderID.(string),
		}
		op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
			log.Printf("[DEBUG] Sending Redis cluster move request: %+v", request)
			return config.sdk.MDB().Redis().Cluster().Move(ctx, request)
		})
		if err != nil {
			return fmt.Errorf("error while requesting API to move Redis Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error while moving Redis Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		if _, err := op.Response(); err != nil {
			return fmt.Errorf("moving Redis Cluster %q to folder %v failed: %s", d.Id(), folderID, err)
		}

	}

	return nil
}
