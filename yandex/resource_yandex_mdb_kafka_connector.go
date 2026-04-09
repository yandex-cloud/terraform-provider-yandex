package yandex

import (
	"fmt"
	"log"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"

	"google.golang.org/genproto/protobuf/field_mask"
)

func resourceYandexMDBKafkaConnector() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a connector of a Kafka cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kafka/concepts).",

		Create: resourceYandexMDBKafkaConnectorCreate,
		Read:   resourceYandexMDBKafkaConnectorRead,
		Update: resourceYandexMDBKafkaConnectorUpdate,
		Delete: resourceYandexMDBKafkaConnectorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the Kafka cluster.",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
				ForceNew:    true,
			},
			"tasks_max": {
				Type:        schema.TypeInt,
				Description: "The number of the connector's parallel working tasks. Default is the number of brokers.",
				Optional:    true,
			},
			"properties": {
				Type:        schema.TypeMap,
				Description: "Additional properties for connector.",
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"connector_config_mirrormaker": {
				Type:        schema.TypeList,
				Description: "Settings for MirrorMaker2 connector.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"topics": {
							Type:        schema.TypeString,
							Description: "The pattern for topic names to be replicated.",
							Required:    true,
						},
						"source_cluster": {
							Type:        schema.TypeList,
							Description: "Settings for source cluster.",
							Required:    true,
							MaxItems:    1,
							Elem:        resourceYandexMDBKafkaClusterConnectionSpec(),
						},
						"target_cluster": {
							Type:        schema.TypeList,
							Description: "Settings for target cluster.",
							Required:    true,
							MaxItems:    1,
							Elem:        resourceYandexMDBKafkaClusterConnectionSpec(),
						},
						"replication_factor": {
							Type:        schema.TypeInt,
							Description: "Replication factor for topics created in target cluster.",
							Required:    true,
						},
					},
				},
			},
			"connector_config_s3_sink": {
				Type:        schema.TypeList,
				Description: "Settings for S3 Sink connector.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"topics": {
							Type:        schema.TypeString,
							Description: "The pattern for topic names to be copied to s3 bucket.",
							Required:    true,
						},
						"file_compression_type": {
							Type:        schema.TypeString,
							Description: "Compression type for messages. Cannot be changed.",
							Required:    true,
							ForceNew:    true,
						},
						"file_max_records": {
							Type:        schema.TypeInt,
							Description: "Max records per file.",
							Optional:    true,
						},
						"s3_connection": {
							Type:        schema.TypeList,
							Description: "Settings for connection to s3-compatible storage.",
							Required:    true,
							MaxItems:    1,
							Elem:        resourceYandexMDBKafkaS3ConnectionSpec(),
						},
					},
				},
			},
			"connector_config_iceberg_sink": {
				Type:        schema.TypeList,
				Description: "Settings for Iceberg Sink connector.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"topics": {
							Type:        schema.TypeString,
							Description: "The pattern for topic names to be written to Iceberg tables.",
							Optional:    true,
						},
						"topics_regex": {
							Type:        schema.TypeString,
							Description: "Regex pattern for topic names to be written to Iceberg tables.",
							Optional:    true,
						},
						"control_topic": {
							Type:        schema.TypeString,
							Description: "Control topic name for Iceberg connector.",
							Optional:    true,
						},
						"metastore_connection": {
							Type:        schema.TypeList,
							Description: "Settings for connection to Hive Metastore.",
							Required:    true,
							MaxItems:    1,
							Elem:        resourceYandexMDBKafkaMetastoreConnectionSpec(),
						},
						"s3_connection": {
							Type:        schema.TypeList,
							Description: "Settings for connection to s3-compatible storage.",
							Required:    true,
							MaxItems:    1,
							Elem:        resourceYandexMDBKafkaIcebergS3ConnectionSpec(),
						},
						"static_tables": {
							Type:        schema.TypeList,
							Description: "Static table routing configuration. Cannot be changed after creation.",
							Optional:    true,
							ForceNew:    true,
							MaxItems:    1,
							Elem:        resourceYandexMDBKafkaStaticTablesSpec(),
						},
						"dynamic_tables": {
							Type:        schema.TypeList,
							Description: "Dynamic table routing configuration. Cannot be changed after creation.",
							Optional:    true,
							ForceNew:    true,
							MaxItems:    1,
							Elem:        resourceYandexMDBKafkaDynamicTablesSpec(),
						},
						"tables_config": {
							Type:        schema.TypeList,
							Description: "Optional table settings.",
							Optional:    true,
							MaxItems:    1,
							Elem:        resourceYandexMDBKafkaIcebergTablesConfigSpec(),
						},
						"control_config": {
							Type:        schema.TypeList,
							Description: "Optional control settings.",
							Optional:    true,
							MaxItems:    1,
							Elem:        resourceYandexMDBKafkaIcebergControlSpec(),
						},
					},
				},
			},
		},
	}

}

func resourceYandexMDBKafkaClusterConnectionSpec() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"alias": {
				Type:        schema.TypeString,
				Description: "Name of the cluster. Used also as a topic prefix.",
				Optional:    true,
			},
			"this_cluster": {
				Type:        schema.TypeList,
				Description: "Using this section in the cluster definition (source or target) means it's this cluster.",
				Optional:    true,
				Elem:        &schema.Resource{},
			},
			"external_cluster": {
				Type:        schema.TypeList,
				Description: "Connection settings for external cluster.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bootstrap_servers": {
							Type:        schema.TypeString,
							Description: "List of bootstrap servers to connect to cluster.",
							Required:    true,
						},
						"sasl_username": {
							Type:        schema.TypeString,
							Description: "Username to use in SASL authentification mechanism.",
							Optional:    true,
						},
						"sasl_password": {
							Type:        schema.TypeString,
							Description: "Password to use in SASL authentification mechanism",
							Optional:    true,
							Sensitive:   true,
						},
						"sasl_mechanism": {
							Type:        schema.TypeString,
							Description: "Type of SASL authentification mechanism to use.",
							Optional:    true,
						},
						"security_protocol": {
							Type:        schema.TypeString,
							Description: "Security protocol to use.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexMDBKafkaS3ConnectionSpec() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Type:        schema.TypeString,
				Description: "Name of the bucket in s3-compatible storage.",
				Required:    true,
			},
			"external_s3": {
				Type:        schema.TypeList,
				Description: "Connection params for external s3-compatible storage.",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint": {
							Type:        schema.TypeString,
							Description: "URL of s3-compatible storage.",
							Required:    true,
						},
						"access_key_id": {
							Type:        schema.TypeString,
							Description: "ID of aws-compatible static key.",
							Optional:    true,
						},
						"secret_access_key": {
							Type:        schema.TypeString,
							Description: "Secret key of aws-compatible static key.",
							Optional:    true,
							Sensitive:   true,
						},
						"region": {
							Type:        schema.TypeString,
							Description: "Region of s3-compatible storage. [Available region list](https://docs.aws.amazon.com/AWSJavaSDK/latest/javadoc/com/amazonaws/regions/Regions.html).",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexMDBKafkaMetastoreConnectionSpec() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"catalog_uri": {
				Type:        schema.TypeString,
				Description: "Thrift URI of Hive Metastore. Format: 'thrift://host:9083'",
				Required:    true,
			},
			"warehouse": {
				Type:        schema.TypeString,
				Description: "Warehouse root directory in S3. Format: 's3a://bucket-name/path/to/warehouse'",
				Required:    true,
			},
		},
	}
}

func resourceYandexMDBKafkaIcebergS3ConnectionSpec() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"external_s3": {
				Type:        schema.TypeList,
				Description: "Connection params for external s3-compatible storage.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint": {
							Type:        schema.TypeString,
							Description: "URL of s3-compatible storage.",
							Required:    true,
						},
						"access_key_id": {
							Type:        schema.TypeString,
							Description: "ID of aws-compatible static key.",
							Optional:    true,
						},
						"secret_access_key": {
							Type:        schema.TypeString,
							Description: "Secret key of aws-compatible static key.",
							Optional:    true,
							Sensitive:   true,
						},
						"region": {
							Type:        schema.TypeString,
							Description: "Region of s3-compatible storage.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexMDBKafkaStaticTablesSpec() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"tables": {
				Type:        schema.TypeString,
				Description: "List of tables, separated by ','.",
				Required:    true,
			},
		},
	}
}

func resourceYandexMDBKafkaDynamicTablesSpec() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"route_field": {
				Type:        schema.TypeString,
				Description: "Field in the message to define the target table.",
				Required:    true,
			},
		},
	}
}

func resourceYandexMDBKafkaIcebergTablesConfigSpec() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"default_commit_branch": {
				Type:        schema.TypeString,
				Description: "Default Git-like branch name for Iceberg commits. Default: 'main'",
				Optional:    true,
			},
			"default_id_columns": {
				Type:        schema.TypeString,
				Description: "List of columns used as identifiers for upsert operations, separated by ','.",
				Optional:    true,
			},
			"default_partition_by": {
				Type:        schema.TypeString,
				Description: "Comma-separated list of columns or transform expressions for table partitioning.",
				Optional:    true,
			},
			"evolve_schema_enabled": {
				Type:        schema.TypeBool,
				Description: "Enable automatic schema evolution. Default: false",
				Optional:    true,
			},
			"schema_force_optional": {
				Type:        schema.TypeBool,
				Description: "Force all columns to be nullable. Default: false",
				Optional:    true,
			},
			"schema_case_insensitive": {
				Type:        schema.TypeBool,
				Description: "Enable case-insensitive field name matching. Default: false",
				Optional:    true,
			},
		},
	}
}

func resourceYandexMDBKafkaIcebergControlSpec() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"group_id_prefix": {
				Type:        schema.TypeString,
				Description: "Consumer group ID prefix for control topic. Default: 'cg-control'",
				Optional:    true,
			},
			"commit_interval_ms": {
				Type:        schema.TypeInt,
				Description: "Interval between commits in milliseconds. Default: 300000 (5 minutes)",
				Optional:    true,
			},
			"commit_timeout_ms": {
				Type:        schema.TypeInt,
				Description: "Commit operation timeout in milliseconds. Default: 30000 (30 seconds)",
				Optional:    true,
			},
			"commit_threads": {
				Type:        schema.TypeInt,
				Description: "Number of threads for commit operations. Default: cores * 2",
				Optional:    true,
			},
			"transactional_prefix": {
				Type:        schema.TypeString,
				Description: "Prefix for transactional operations. Default: ''",
				Optional:    true,
			},
		},
	}
}

func resourceYandexMDBKafkaConnectorCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	connectorSpec, err := buildKafkaConnectorSpec(d)
	if err != nil {
		return err
	}

	req := &kafka.CreateConnectorRequest{
		ClusterId:     d.Get("cluster_id").(string),
		ConnectorSpec: connectorSpec,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Creating Kafka connector: %+v", req)
		return config.sdk.MDB().Kafka().Connector().Create(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to create Kafka connector: %s", err)
	}

	conectorName := constructResourceId(req.ClusterId, req.ConnectorSpec.Name)
	d.SetId(conectorName)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for Kafka conector create operation: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("kafka conector creation failed: %s", err)
	}
	log.Printf("[DEBUG] Finished creating Kafka conector %q", conectorName)

	return resourceYandexMDBKafkaConnectorRead(d, meta)
}

func resourceYandexMDBKafkaConnectorRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid connector resource id format: %q", d.Id())
	}

	clusterID := parts[0]
	connectorName := parts[1]
	conn, err := config.sdk.MDB().Kafka().Connector().Get(ctx, &kafka.GetConnectorRequest{
		ClusterId:     clusterID,
		ConnectorName: connectorName,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Connector %q", connectorName))
	}
	if err = d.Set("cluster_id", clusterID); err != nil {
		return err
	}
	if err = d.Set("name", conn.Name); err != nil {
		return err
	}
	if err = d.Set("tasks_max", conn.TasksMax.GetValue()); err != nil {
		return err
	}
	if err = d.Set("properties", conn.Properties); err != nil {
		return err
	}

	switch conn.GetConnectorConfig().(type) {
	case *kafka.Connector_ConnectorConfigMirrormaker:
		cfg, err := flattenKafkaConnectorMirrormaker(conn.GetConnectorConfigMirrormaker())
		if err != nil {
			return err
		}
		if err = d.Set("connector_config_mirrormaker", cfg); err != nil {
			return err
		}
	case *kafka.Connector_ConnectorConfigS3Sink:
		cfg, err := flattenKafkaConnectorS3Sink(conn.GetConnectorConfigS3Sink(), d)
		if err != nil {
			return err
		}
		if err = d.Set("connector_config_s3_sink", cfg); err != nil {
			return err
		}
	case *kafka.Connector_ConnectorConfigIcebergSink:
		cfg, err := flattenKafkaConnectorIcebergSink(conn.GetConnectorConfigIcebergSink(), d)
		if err != nil {
			return err
		}
		if err = d.Set("connector_config_iceberg_sink", cfg); err != nil {
			return err
		}
	default:
		return fmt.Errorf("this type of connector is not supported by current version of terraform provider")
	}
	return nil
}

func resourceYandexMDBKafkaConnectorUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	var updatePath []string
	for field, path := range mdbKafkaConnectorUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}
	if len(updatePath) == 0 {
		return nil
	}

	connSpec, err := buildKafkaConnectorUpdateSpec(d)
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	connName := d.Get("name").(string)
	request := &kafka.UpdateConnectorRequest{
		ClusterId:     clusterID,
		ConnectorName: connName,
		ConnectorSpec: connSpec,
		UpdateMask:    &field_mask.FieldMask{Paths: updatePath},
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending connector update request: %+v", request)
		return config.sdk.MDB().Kafka().Connector().Update(ctx, request)
	})
	if err != nil {
		return fmt.Errorf("error while requesting API to update connector %q in Kafka Cluster %q: %s",
			connName, clusterID, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating connector in Kafka Cluster %q: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Finished updating Kafka connector %q", connName)

	return resourceYandexMDBKafkaConnectorRead(d, meta)
}

func resourceYandexMDBKafkaConnectorDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	connName := d.Get("name").(string)
	clusterID := d.Get("cluster_id").(string)
	request := &kafka.DeleteConnectorRequest{
		ClusterId:     clusterID,
		ConnectorName: connName,
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Deleting Kafka connector %q", connName)
		return config.sdk.MDB().Kafka().Connector().Delete(ctx, request)
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Kafka connector %q", connName))
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting connector %q from Kafka Cluster %q: %s", connName, clusterID, err)
	}

	log.Printf("[DEBUG] Finished deleting Kafka connector %q", connName)
	return nil
}

func getWrapedInt64(d *schema.ResourceData, name string) *wrappers.Int64Value {
	val, ok := d.GetOk(name)
	if !ok {
		return nil
	}
	valInt := val.(int)
	return &wrappers.Int64Value{Value: int64(valInt)}
}

func buildKafkaConnectorSpec(d *schema.ResourceData) (*kafka.ConnectorSpec, error) {
	connSpec := &kafka.ConnectorSpec{
		Name:     d.Get("name").(string),
		TasksMax: getWrapedInt64(d, "tasks_max"),
	}

	props, err := expandLabels(d.Get("properties"))
	if err != nil {
		return nil, fmt.Errorf("error expanding properties while creating connector: %s", err)
	}
	connSpec.Properties = props
	var countOfSpecificConnectorConfigs int64
	if _, ok := d.GetOk("connector_config_mirrormaker"); ok {
		connSpec.SetConnectorConfigMirrormaker(buildKafkaMirrorMakerSpec(d))
		countOfSpecificConnectorConfigs++
	}
	if _, ok := d.GetOk("connector_config_s3_sink"); ok {
		connSpec.SetConnectorConfigS3Sink(buildKafkaS3SinkConnectorSpec(d))
		countOfSpecificConnectorConfigs++
	}
	if _, ok := d.GetOk("connector_config_iceberg_sink"); ok {
		connSpec.SetConnectorConfigIcebergSink(buildKafkaIcebergSinkConnectorSpec(d))
		countOfSpecificConnectorConfigs++
	}
	if countOfSpecificConnectorConfigs == 0 {
		return nil, fmt.Errorf("connector-specific config must be specified")
	} else if countOfSpecificConnectorConfigs > 1 {
		return nil, fmt.Errorf("must be specified only one connector-specific config")
	}
	return connSpec, nil
}

func buildKafkaConnectorUpdateSpec(d *schema.ResourceData) (*kafka.UpdateConnectorSpec, error) {
	connSpec := &kafka.UpdateConnectorSpec{
		TasksMax: getWrapedInt64(d, "tasks_max"),
	}

	props, err := expandLabels(d.Get("properties"))
	if err != nil {
		return nil, fmt.Errorf("error expanding properties while creating connector: %s", err)
	}
	connSpec.Properties = props

	var countOfSpecificConnectorConfigs int64
	if _, ok := d.GetOk("connector_config_mirrormaker"); ok {
		connSpec.SetConnectorConfigMirrormaker(buildKafkaMirrorMakerSpec(d))
		countOfSpecificConnectorConfigs++
	}
	if _, ok := d.GetOk("connector_config_s3_sink"); ok {
		connSpec.SetConnectorConfigS3Sink(buildKafkaS3SinkConnectorSpecUpdate(d))
		countOfSpecificConnectorConfigs++
	}
	if _, ok := d.GetOk("connector_config_iceberg_sink"); ok {
		connSpec.SetConnectorConfigIcebergSink(buildKafkaIcebergSinkConnectorSpecUpdate(d))
		countOfSpecificConnectorConfigs++
	}
	if countOfSpecificConnectorConfigs > 1 {
		return nil, fmt.Errorf("must be specified only one connector-specific config")
	}
	return connSpec, nil
}

func buildKafkaMirrorMakerSpec(d *schema.ResourceData) *kafka.ConnectorConfigMirrorMakerSpec {
	return &kafka.ConnectorConfigMirrorMakerSpec{
		SourceCluster:     buildKafkaClusterConnectionSpec(d, "connector_config_mirrormaker.0.source_cluster.0."),
		TargetCluster:     buildKafkaClusterConnectionSpec(d, "connector_config_mirrormaker.0.target_cluster.0."),
		Topics:            d.Get("connector_config_mirrormaker.0.topics").(string),
		ReplicationFactor: getWrapedInt64(d, "connector_config_mirrormaker.0.replication_factor"),
	}
}

func buildKafkaClusterConnectionSpec(d *schema.ResourceData, prefixKey string) *kafka.ClusterConnectionSpec {
	key := func(key string) string {
		return fmt.Sprintf("%s%s", prefixKey, key)
	}
	spec := &kafka.ClusterConnectionSpec{
		Alias: d.Get(key("alias")).(string),
	}
	if _, ok := d.GetOk(key("this_cluster")); ok {
		spec.ClusterConnection = &kafka.ClusterConnectionSpec_ThisCluster{}
	}
	if _, ok := d.GetOk(key("external_cluster")); ok {
		spec.ClusterConnection = &kafka.ClusterConnectionSpec_ExternalCluster{
			ExternalCluster: &kafka.ExternalClusterConnectionSpec{
				BootstrapServers: d.Get(key("external_cluster.0.bootstrap_servers")).(string),
				SaslUsername:     d.Get(key("external_cluster.0.sasl_username")).(string),
				SaslPassword:     d.Get(key("external_cluster.0.sasl_password")).(string),
				SaslMechanism:    d.Get(key("external_cluster.0.sasl_mechanism")).(string),
				SecurityProtocol: d.Get(key("external_cluster.0.security_protocol")).(string),
			},
		}
	}
	return spec
}

func buildKafkaS3SinkConnectorSpec(d *schema.ResourceData) *kafka.ConnectorConfigS3SinkSpec {
	return &kafka.ConnectorConfigS3SinkSpec{
		S3Connection:        buildS3ConnectionSpec(d, "connector_config_s3_sink.0.s3_connection.0."),
		Topics:              d.Get("connector_config_s3_sink.0.topics").(string),
		FileCompressionType: d.Get("connector_config_s3_sink.0.file_compression_type").(string),
		FileMaxRecords:      getWrapedInt64(d, "connector_config_s3_sink.0.file_max_records"),
	}
}

func buildKafkaS3SinkConnectorSpecUpdate(d *schema.ResourceData) *kafka.UpdateConnectorConfigS3SinkSpec {
	return &kafka.UpdateConnectorConfigS3SinkSpec{
		S3Connection:   buildS3ConnectionSpec(d, "connector_config_s3_sink.0.s3_connection.0."),
		Topics:         d.Get("connector_config_s3_sink.0.topics").(string),
		FileMaxRecords: getWrapedInt64(d, "connector_config_s3_sink.0.file_max_records"),
	}
}

func buildS3ConnectionSpec(d *schema.ResourceData, prefixKey string) *kafka.S3ConnectionSpec {
	key := func(key string) string {
		return fmt.Sprintf("%s%s", prefixKey, key)
	}
	spec := &kafka.S3ConnectionSpec{
		BucketName: d.Get(key("bucket_name")).(string),
	}
	if _, ok := d.GetOk(key("external_s3")); ok {
		spec.Storage = &kafka.S3ConnectionSpec_ExternalS3{
			ExternalS3: &kafka.ExternalS3StorageSpec{
				AccessKeyId:     d.Get(key("external_s3.0.access_key_id")).(string),
				SecretAccessKey: d.Get(key("external_s3.0.secret_access_key")).(string),
				Endpoint:        d.Get(key("external_s3.0.endpoint")).(string),
				Region:          d.Get(key("external_s3.0.region")).(string),
			},
		}
	}
	return spec
}
func buildKafkaIcebergSinkConnectorSpec(d *schema.ResourceData) *kafka.ConnectorConfigIcebergSinkSpec {
	spec := &kafka.ConnectorConfigIcebergSinkSpec{
		MetastoreConnection: buildMetastoreConnectionSpec(d, "connector_config_iceberg_sink.0.metastore_connection.0."),
		S3Connection:        buildIcebergS3ConnectionSpec(d, "connector_config_iceberg_sink.0.s3_connection.0."),
	}

	// Topics source (topics or topics_regex)
	if v, ok := d.GetOk("connector_config_iceberg_sink.0.topics"); ok {
		spec.TopicsSource = &kafka.ConnectorConfigIcebergSinkSpec_Topics{
			Topics: v.(string),
		}
	}
	if v, ok := d.GetOk("connector_config_iceberg_sink.0.topics_regex"); ok {
		spec.TopicsSource = &kafka.ConnectorConfigIcebergSinkSpec_TopicsRegex{
			TopicsRegex: v.(string),
		}
	}

	// Control topic
	spec.ControlTopic = d.Get("connector_config_iceberg_sink.0.control_topic").(string)

	// Table routing (static_tables or dynamic_tables)
	if _, ok := d.GetOk("connector_config_iceberg_sink.0.static_tables"); ok {
		spec.TableRouting = &kafka.ConnectorConfigIcebergSinkSpec_StaticTables{
			StaticTables: &kafka.StaticTablesSpec{
				Tables: d.Get("connector_config_iceberg_sink.0.static_tables.0.tables").(string),
			},
		}
	}
	if _, ok := d.GetOk("connector_config_iceberg_sink.0.dynamic_tables"); ok {
		spec.TableRouting = &kafka.ConnectorConfigIcebergSinkSpec_DynamicTables{
			DynamicTables: &kafka.DynamicTablesSpec{
				RouteField: d.Get("connector_config_iceberg_sink.0.dynamic_tables.0.route_field").(string),
			},
		}
	}

	// Tables config
	if _, ok := d.GetOk("connector_config_iceberg_sink.0.tables_config"); ok {
		spec.TablesConfig = buildIcebergTablesConfigSpec(d, "connector_config_iceberg_sink.0.tables_config.0.")
	}

	// Control config
	if _, ok := d.GetOk("connector_config_iceberg_sink.0.control_config"); ok {
		spec.ControlConfig = buildIcebergControlSpec(d, "connector_config_iceberg_sink.0.control_config.0.")
	}

	return spec
}

func buildKafkaIcebergSinkConnectorSpecUpdate(d *schema.ResourceData) *kafka.UpdateConnectorConfigIcebergSinkSpec {
	spec := &kafka.UpdateConnectorConfigIcebergSinkSpec{
		MetastoreConnection: buildMetastoreConnectionSpec(d, "connector_config_iceberg_sink.0.metastore_connection.0."),
		S3Connection:        buildIcebergS3ConnectionSpec(d, "connector_config_iceberg_sink.0.s3_connection.0."),
	}

	// Topics source (topics or topics_regex)
	if v, ok := d.GetOk("connector_config_iceberg_sink.0.topics"); ok {
		spec.TopicsSource = &kafka.UpdateConnectorConfigIcebergSinkSpec_Topics{
			Topics: v.(string),
		}
	}
	if v, ok := d.GetOk("connector_config_iceberg_sink.0.topics_regex"); ok {
		spec.TopicsSource = &kafka.UpdateConnectorConfigIcebergSinkSpec_TopicsRegex{
			TopicsRegex: v.(string),
		}
	}

	// Control topic
	if v, ok := d.GetOk("connector_config_iceberg_sink.0.control_topic"); ok {
		spec.ControlTopic = v.(string)
	}

	// Tables config
	if _, ok := d.GetOk("connector_config_iceberg_sink.0.tables_config"); ok {
		spec.TablesConfig = buildIcebergTablesConfigSpec(d, "connector_config_iceberg_sink.0.tables_config.0.")
	}

	// Control config
	if _, ok := d.GetOk("connector_config_iceberg_sink.0.control_config"); ok {
		spec.ControlConfig = buildIcebergControlSpec(d, "connector_config_iceberg_sink.0.control_config.0.")
	}

	return spec
}

func buildMetastoreConnectionSpec(d *schema.ResourceData, prefixKey string) *kafka.MetastoreConnectionSpec {
	key := func(key string) string {
		return fmt.Sprintf("%s%s", prefixKey, key)
	}
	return &kafka.MetastoreConnectionSpec{
		CatalogUri: d.Get(key("catalog_uri")).(string),
		Warehouse:  d.Get(key("warehouse")).(string),
	}
}

func buildIcebergS3ConnectionSpec(d *schema.ResourceData, prefixKey string) *kafka.IcebergS3ConnectionSpec {
	key := func(key string) string {
		return fmt.Sprintf("%s%s", prefixKey, key)
	}
	spec := &kafka.IcebergS3ConnectionSpec{}
	if _, ok := d.GetOk(key("external_s3")); ok {
		spec.Storage = &kafka.IcebergS3ConnectionSpec_ExternalS3{
			ExternalS3: &kafka.ExternalIcebergS3StorageSpec{
				AccessKeyId:     d.Get(key("external_s3.0.access_key_id")).(string),
				SecretAccessKey: d.Get(key("external_s3.0.secret_access_key")).(string),
				Endpoint:        d.Get(key("external_s3.0.endpoint")).(string),
				Region:          d.Get(key("external_s3.0.region")).(string),
			},
		}
	}
	return spec
}

func buildIcebergTablesConfigSpec(d *schema.ResourceData, prefixKey string) *kafka.IcebergTablesConfigSpec {
	key := func(key string) string {
		return fmt.Sprintf("%s%s", prefixKey, key)
	}
	spec := &kafka.IcebergTablesConfigSpec{}

	if v, ok := d.GetOk(key("default_commit_branch")); ok {
		spec.DefaultCommitBranch = v.(string)
	}
	if v, ok := d.GetOk(key("default_id_columns")); ok {
		spec.DefaultIdColumns = v.(string)
	}
	if v, ok := d.GetOk(key("default_partition_by")); ok {
		spec.DefaultPartitionBy = v.(string)
	}
	if v, ok := d.GetOk(key("evolve_schema_enabled")); ok {
		spec.EvolveSchemaEnabled = v.(bool)
	}
	if v, ok := d.GetOk(key("schema_force_optional")); ok {
		spec.SchemaForceOptional = v.(bool)
	}
	if v, ok := d.GetOk(key("schema_case_insensitive")); ok {
		spec.SchemaCaseInsensitive = v.(bool)
	}

	return spec
}

func buildIcebergControlSpec(d *schema.ResourceData, prefixKey string) *kafka.IcebergControlSpec {
	key := func(key string) string {
		return fmt.Sprintf("%s%s", prefixKey, key)
	}
	spec := &kafka.IcebergControlSpec{}

	if v, ok := d.GetOk(key("group_id_prefix")); ok {
		spec.GroupIdPrefix = v.(string)
	}
	if v, ok := d.GetOk(key("commit_interval_ms")); ok {
		spec.CommitIntervalMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(key("commit_timeout_ms")); ok {
		spec.CommitTimeoutMs = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(key("commit_threads")); ok {
		spec.CommitThreads = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(key("transactional_prefix")); ok {
		spec.TransactionalPrefix = v.(string)
	}

	return spec
}

var mdbKafkaConnectorUpdateFieldsMap = map[string]string{}

func init() {
	keyPrefix := ""
	valPrefix := "connector_spec."
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"tasks_max"] = valPrefix + "tasks_max"
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"properties"] = valPrefix + "properties"
	addMirrormakerUpdatePathsToFieldsMap(keyPrefix, valPrefix)
	addS3SinkUpdatePathsToFieldsMap(keyPrefix, valPrefix)
	addIcebergSinkUpdatePathsToFieldsMap(keyPrefix, valPrefix)
}

func addMirrormakerUpdatePathsToFieldsMap(commonKeyPrefix string, commonValPrefix string) {
	keyPrefix := commonKeyPrefix + "connector_config_mirrormaker.0."
	valPrefix := commonValPrefix + "connector_config_mirrormaker."
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"topics"] = valPrefix + "topics"
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"replication_factor"] = valPrefix + "replication_factor"

	for _, source := range []string{"source_cluster", "target_cluster"} {
		keyPrefix := keyPrefix + source + ".0."
		valPrefix := valPrefix + source + "."
		mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"alias"] = valPrefix + "alias"

		keyPrefix = keyPrefix + "external_cluster.0."
		valPrefix = valPrefix + "external_cluster."
		mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"bootstrap_servers"] = valPrefix + "bootstrap_servers"
		mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"sasl_username"] = valPrefix + "sasl_username"
		mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"sasl_password"] = valPrefix + "sasl_password"
		mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"sasl_mechanism"] = valPrefix + "sasl_mechanism"
		mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"security_protocol"] = valPrefix + "security_protocol"
	}
}

func addS3SinkUpdatePathsToFieldsMap(commonKeyPrefix string, commonValPrefix string) {
	keyPrefix := commonKeyPrefix + "connector_config_s3_sink.0."
	valPrefix := commonValPrefix + "connector_config_s3_sink."
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"topics"] = valPrefix + "topics"
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"file_max_records"] = valPrefix + "file_max_records"

	keyPrefix = keyPrefix + "s3_connection" + ".0."
	valPrefix = valPrefix + "s3_connection" + "."
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"bucket_name"] = valPrefix + "bucket_name"

	keyPrefix = keyPrefix + "external_s3.0."
	valPrefix = valPrefix + "external_s3."
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"access_key_id"] = valPrefix + "access_key_id"
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"secret_access_key"] = valPrefix + "secret_access_key"
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"endpoint"] = valPrefix + "endpoint"
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"region"] = valPrefix + "region"
}

func addIcebergSinkUpdatePathsToFieldsMap(commonKeyPrefix string, commonValPrefix string) {
	keyPrefix := commonKeyPrefix + "connector_config_iceberg_sink.0."
	valPrefix := commonValPrefix + "connector_config_iceberg_sink."

	// Topics source
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"topics"] = valPrefix + "topics"
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"topics_regex"] = valPrefix + "topics_regex"

	// Control topic
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"control_topic"] = valPrefix + "control_topic"

	// Metastore connection
	metastoreKeyPrefix := keyPrefix + "metastore_connection.0."
	metastoreValPrefix := valPrefix + "metastore_connection."
	mdbKafkaConnectorUpdateFieldsMap[metastoreKeyPrefix+"catalog_uri"] = metastoreValPrefix + "catalog_uri"
	mdbKafkaConnectorUpdateFieldsMap[metastoreKeyPrefix+"warehouse"] = metastoreValPrefix + "warehouse"

	// S3 connection
	s3KeyPrefix := keyPrefix + "s3_connection.0."
	s3ValPrefix := valPrefix + "s3_connection."

	s3ExternalKeyPrefix := s3KeyPrefix + "external_s3.0."
	s3ExternalValPrefix := s3ValPrefix + "external_s3."
	mdbKafkaConnectorUpdateFieldsMap[s3ExternalKeyPrefix+"access_key_id"] = s3ExternalValPrefix + "access_key_id"
	mdbKafkaConnectorUpdateFieldsMap[s3ExternalKeyPrefix+"secret_access_key"] = s3ExternalValPrefix + "secret_access_key"
	mdbKafkaConnectorUpdateFieldsMap[s3ExternalKeyPrefix+"endpoint"] = s3ExternalValPrefix + "endpoint"
	mdbKafkaConnectorUpdateFieldsMap[s3ExternalKeyPrefix+"region"] = s3ExternalValPrefix + "region"

	// Tables config
	tablesConfigKeyPrefix := keyPrefix + "tables_config.0."
	tablesConfigValPrefix := valPrefix + "tables_config."
	mdbKafkaConnectorUpdateFieldsMap[tablesConfigKeyPrefix+"default_commit_branch"] = tablesConfigValPrefix + "default_commit_branch"
	mdbKafkaConnectorUpdateFieldsMap[tablesConfigKeyPrefix+"default_id_columns"] = tablesConfigValPrefix + "default_id_columns"
	mdbKafkaConnectorUpdateFieldsMap[tablesConfigKeyPrefix+"default_partition_by"] = tablesConfigValPrefix + "default_partition_by"
	mdbKafkaConnectorUpdateFieldsMap[tablesConfigKeyPrefix+"evolve_schema_enabled"] = tablesConfigValPrefix + "evolve_schema_enabled"
	mdbKafkaConnectorUpdateFieldsMap[tablesConfigKeyPrefix+"schema_force_optional"] = tablesConfigValPrefix + "schema_force_optional"
	mdbKafkaConnectorUpdateFieldsMap[tablesConfigKeyPrefix+"schema_case_insensitive"] = tablesConfigValPrefix + "schema_case_insensitive"

	// Control config
	controlConfigKeyPrefix := keyPrefix + "control_config.0."
	controlConfigValPrefix := valPrefix + "control_config."
	mdbKafkaConnectorUpdateFieldsMap[controlConfigKeyPrefix+"group_id_prefix"] = controlConfigValPrefix + "group_id_prefix"
	mdbKafkaConnectorUpdateFieldsMap[controlConfigKeyPrefix+"commit_interval_ms"] = controlConfigValPrefix + "commit_interval_ms"
	mdbKafkaConnectorUpdateFieldsMap[controlConfigKeyPrefix+"commit_timeout_ms"] = controlConfigValPrefix + "commit_timeout_ms"
	mdbKafkaConnectorUpdateFieldsMap[controlConfigKeyPrefix+"commit_threads"] = controlConfigValPrefix + "commit_threads"
	mdbKafkaConnectorUpdateFieldsMap[controlConfigKeyPrefix+"transactional_prefix"] = controlConfigValPrefix + "transactional_prefix"
}
