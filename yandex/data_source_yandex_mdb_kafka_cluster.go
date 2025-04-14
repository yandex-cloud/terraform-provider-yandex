package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexMDBKafkaCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Managed Kafka cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).\n\n~> Either `cluster_id` or `name` should be specified.\n",

		Read: dataSourceYandexMDBKafkaClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the Kafka cluster.",
				Computed:    true,
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Computed:    true,
				Optional:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Computed:    true,
			},
			"environment": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBKafkaCluster().Schema["environment"].Description,
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBKafkaCluster().Schema["health"].Description,
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBKafkaCluster().Schema["status"].Description,
				Computed:    true,
			},
			"config": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBKafkaCluster().Schema["config"].Description,
				Optional:    true,
				MaxItems:    1,
				Elem:        resourceYandexMDBKafkaClusterConfig(),
			},
			"subnet_ids": {
				Type:        schema.TypeList,
				Description: common.ResourceDescriptions["subnet_ids"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"topic": {
				Type:        schema.TypeList,
				Description: "List of kafka topics.",
				Optional:    true,
				Elem:        resourceYandexMDBKafkaClusterTopicBlock(),
				Deprecated:  useResourceInstead("topic", "yandex_mdb_kafka_topic"),
			},
			"user": {
				Type:        schema.TypeSet,
				Description: "List of kafka users.",
				Optional:    true,
				Set:         kafkaUserHash,
				Elem:        resourceYandexMDBKafkaClusterUserBlock(),
				Deprecated:  useResourceInstead("user", "yandex_mdb_kafka_user"),
			},
			"security_group_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["security_group_ids"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Computed:    true,
			},
			"host_group_ids": {
				Type:        schema.TypeSet,
				Description: resourceYandexMDBKafkaCluster().Schema["host_group_ids"].Description,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Computed:    true,
			},
			"host": {
				Type:        schema.TypeSet,
				Description: resourceYandexMDBKafkaCluster().Schema["host"].Description,
				Computed:    true,
				Set:         kafkaHostHash,
				Elem:        resourceYandexMDBKafkaHost(),
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Computed:    true,
				Optional:    true,
			},
			"maintenance_window": {
				Type:        schema.TypeList,
				Description: "Maintenance policy of the Kafka cluster.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Description: "Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`. A day and hour of window need to be specified with weekly window.",
							Computed:    true,
						},
						"day": {
							Type:        schema.TypeString,
							Description: "Day of the week (in `DDD` format). Allowed values: `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`.",
							Computed:    true,
						},
						"hour": {
							Type:        schema.TypeInt,
							Description: "Hour of the day in UTC (in `HH` format). Allowed value is between 1 and 24.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexMDBKafkaClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.KafkaClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Kafka Cluster by name: %v", err)
		}
	}

	cluster, err := config.sdk.MDB().Kafka().Cluster().Get(ctx, &kafka.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	if err := d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
	d.Set("cluster_id", cluster.Id)
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)

	cfg, err := flattenKafkaConfig(cluster)
	if err != nil {
		return err
	}
	if err := d.Set("config", cfg); err != nil {
		return err
	}

	topics, err := listKafkaTopics(ctx, config, clusterID)
	if err != nil {
		return err
	}
	if err := d.Set("topic", flattenKafkaTopics(topics)); err != nil {
		return err
	}

	users, err := listKafkaUsers(ctx, config, clusterID)
	if err != nil {
		return err
	}
	if err := d.Set("user", flattenKafkaUsers(users, nil)); err != nil {
		return err
	}

	hosts, err := listKafkaHosts(ctx, config, clusterID)
	if err != nil {
		return err
	}
	if err := d.Set("host", flattenKafkaHosts(hosts)); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	if err := d.Set("host_group_ids", cluster.HostGroupIds); err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.DeletionProtection)

	maintenanceWindow, err := flattenKafkaMaintenanceWindow(cluster.MaintenanceWindow)
	if err != nil {
		return err
	}
	if err := d.Set("maintenance_window", maintenanceWindow); err != nil {
		return err
	}

	d.SetId(cluster.Id)
	return nil
}
