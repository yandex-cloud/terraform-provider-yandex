package yandex

import (
	"fmt"
	"log"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"

	"google.golang.org/genproto/protobuf/field_mask"
)

func resourceYandexMDBKafkaConnector() *schema.Resource {
	return &schema.Resource{
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tasks_max": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"properties": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"connector_config_mirrormaker": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"topics": {
							Type:     schema.TypeString,
							Required: true,
						},
						"source_cluster": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem:     resourceYandexMDBKafkaClusterConnectionSpec(),
						},
						"target_cluster": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem:     resourceYandexMDBKafkaClusterConnectionSpec(),
						},
						"replication_factor": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"connector_config_s3_sink": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"topics": {
							Type:     schema.TypeString,
							Required: true,
						},
						"file_compression_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"file_max_records": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"s3_connection": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem:     resourceYandexMDBKafkaS3ConnectionSpec(),
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
				Type:     schema.TypeString,
				Optional: true,
			},
			"this_cluster": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Resource{},
			},
			"external_cluster": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bootstrap_servers": {
							Type:     schema.TypeString,
							Required: true,
						},
						"sasl_username": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"sasl_password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"sasl_mechanism": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"security_protocol": {
							Type:     schema.TypeString,
							Optional: true,
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
				Type:     schema.TypeString,
				Required: true,
			},
			"external_s3": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint": {
							Type:     schema.TypeString,
							Required: true,
						},
						"access_key_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"secret_access_key": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
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
		cfg, err := flattenKafkaConnectorS3Sink(conn.GetConnectorConfigS3Sink())
		if err != nil {
			return err
		}
		if err = d.Set("connector_config_s3_sink", cfg); err != nil {
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

var mdbKafkaConnectorUpdateFieldsMap = map[string]string{}

func init() {
	keyPrefix := ""
	valPrefix := "connector_spec."
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"tasks_max"] = valPrefix + "tasks_max"
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"properties"] = valPrefix + "properties"
	addMirrormakerUpdatePathsToFieldsMap(keyPrefix, valPrefix)
	addS3SinkUpdatePathsToFieldsMap(keyPrefix, valPrefix)
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
