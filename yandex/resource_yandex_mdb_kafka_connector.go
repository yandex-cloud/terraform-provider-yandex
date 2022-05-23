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
		log.Printf("[DEBUG] Creating Kafka topic: %+v", req)
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
		return fmt.Errorf("invalid topic resource id format: %q", d.Id())
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
	d.Set("cluster_id", clusterID)
	d.Set("name", conn.Name)
	d.Set("tasks_max", conn.TasksMax.GetValue())

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

	if _, ok := d.GetOk("connector_config_mirrormaker"); ok {
		connSpec.SetConnectorConfigMirrormaker(buildKafkaMirrorMakerSpec(d))
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

	if _, ok := d.GetOk("connector_config_mirrormaker"); ok {
		connSpec.SetConnectorConfigMirrormaker(buildKafkaMirrorMakerSpec(d))
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

var mdbKafkaConnectorUpdateFieldsMap = map[string]string{}

func init() {
	keyPrefix := ""
	valPrefix := "connector_spec."
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"tasks_max"] = valPrefix + "tasks_max"
	mdbKafkaConnectorUpdateFieldsMap[keyPrefix+"properties"] = valPrefix + "properties"

	keyPrefix = keyPrefix + "connector_config_mirrormaker.0."
	valPrefix = valPrefix + "connector_config_mirrormaker."
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
