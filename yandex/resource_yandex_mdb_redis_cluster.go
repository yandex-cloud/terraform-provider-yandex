package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
)

const yandexMDBRedisClusterDefaultTimeout = 15 * time.Minute
const defaultMDBPageSize = 1000

func resourceYandexMDBRedisCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBRedisClusterCreate,
		Read:   resourceYandexMDBRedisClusterRead,
		Update: resourceYandexMDBRedisClusterUpdate,
		Delete: resourceYandexMDBRedisClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBRedisClusterDefaultTimeout),
			Update: schema.DefaultTimeout(yandexMDBRedisClusterDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexMDBRedisClusterDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"environment": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseRedisEnv),
			},
			"config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"maxmemory_policy": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateParsableValue(parseRedisMaxmemoryPolicy),
						},
					},
				},
			},
			"resources": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"disk_size": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"host": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:     schema.TypeString,
							Required: true,
						},
						"shard_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"sharded": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
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

	conf, err := expandRedisConfig(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding config while creating Redis Cluster: %s", err)
	}

	resources, err := expandRedisResources(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding resources on Redis Cluster create: %s", err)
	}

	configSpec := &redis.ConfigSpec{
		RedisSpec: conf,
		Resources: resources,
	}

	req := redis.CreateClusterRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		NetworkId:   d.Get("network_id").(string),
		Environment: env,
		ConfigSpec:  configSpec,
		HostSpecs:   hosts,
		Labels:      labels,
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

	hosts := []*redis.Host{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Redis().Cluster().ListHosts(ctx, &redis.ListClusterHostsRequest{
			ClusterId: d.Id(),
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return fmt.Errorf("Error while getting list of hosts for '%s': %s", d.Id(), err)
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	createdAt, err := getTimestamp(cluster.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)
	d.Set("sharded", cluster.Sharded)

	resources, err := flattenRedisResources(cluster.Config.Resources)
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
			"timeout":          conf.timeout,
			"maxmemory_policy": conf.maxmemoryPolicy,
			"password":         password,
		},
	})
	if err != nil {
		return err
	}

	if err := d.Set("resources", resources); err != nil {
		return err
	}

	// Do not change the state if only order of hosts differs.
	dHosts, err := expandRedisHosts(d)
	if err != nil {
		return err
	}

	sortRedisHosts(hosts, dHosts)

	hs, err := flattenRedisHosts(hosts)
	if err != nil {
		return err
	}

	if err := d.Set("host", hs); err != nil {
		return err
	}

	return d.Set("labels", cluster.Labels)
}

func resourceYandexMDBRedisClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	if d.HasChange("name") || d.HasChange("labels") || d.HasChange("description") || d.HasChange("resources") || d.HasChange("config") {
		if err := updateRedisClusterParams(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("host") {
		if err := updateRedisClusterHosts(d, meta); err != nil {
			return err
		}
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
	onDone := []func(){}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")

		onDone = append(onDone, func() {
			d.SetPartial("name")
		})
	}

	if d.HasChange("labels") {
		labelsProp, err := expandLabels(d.Get("labels"))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")

		onDone = append(onDone, func() {
			d.SetPartial("labels")
		})
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")

		onDone = append(onDone, func() {
			d.SetPartial("description")
		})
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

		onDone = append(onDone, func() {
			d.SetPartial("resources")
		})
	}

	if d.HasChange("config") {
		conf, err := expandRedisConfig(d)
		if err != nil {
			return err
		}

		if req.ConfigSpec == nil {
			req.ConfigSpec = &redis.ConfigSpec{}
		}

		req.ConfigSpec.RedisSpec = conf
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.redis_config_5_0")

		onDone = append(onDone, func() {
			d.SetPartial("config")
		})
	}

	err := makeRedisClusterUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	for _, f := range onDone {
		f()
	}
	return nil
}

func updateRedisClusterHosts(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	resp, err := config.sdk.MDB().Redis().Cluster().ListHosts(ctx, &redis.ListClusterHostsRequest{
		ClusterId: d.Id(),
		PageSize:  defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("Error while getting list of hosts for '%s': %s", d.Id(), err)
	}

	currHosts := resp.Hosts
	targetHosts, err := expandRedisHosts(d)
	if err != nil {
		return fmt.Errorf("Error while expanding hosts on Redis Cluster create: %s", err)
	}

	toDelete, toAdd := redisHostsDiff(currHosts, targetHosts)

	ctx, cancel = context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	for _, fqdn := range toDelete {
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

	for _, hs := range toAdd {
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

	d.SetPartial("host")
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
