package yandex

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	mongo_config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1/config"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
)

var supportedVersions = map[string]bool{
	"6.0-enterprise": true,
	"5.0-enterprise": true,
	"4.4-enterprise": true,
	"6.0":            true,
	"5.0":            true,
	"4.4":            true,
	"4.2":            true,
	"4.0":            true,
	"3.6":            true,
}

type MongodbSpecHelper struct {
	FlattenResources func(c *mongodb.ClusterConfig, d *schema.ResourceData) (map[string]interface{}, error)
	FlattenMongod    func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error)
	FlattenMongos    func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error)
	FlattenMongocfg  func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error)
	Expand           func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec
}

func GetMongodbSpecHelper(version string) *MongodbSpecHelper {
	switch version {
	case "6.0-enterprise":
		{
			return &MongodbSpecHelper{
				FlattenResources: func(c *mongodb.ClusterConfig, d *schema.ResourceData) (map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_6_0Enterprise).Mongodb_6_0Enterprise
					resources := map[string]interface{}{}
					if _, ok := d.GetOk("resources"); ok {
						if spec.Mongod != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongod.Resources)
							return resources, nil
						}
						if spec.Mongos != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongos.Resources)
							return resources, nil
						}
						if spec.Mongocfg != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongocfg.Resources)
							return resources, nil
						}
						if spec.Mongoinfra != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
							return resources, nil
						}
					} else {
						if spec.Mongod != nil {
							resources["resources_mongod"] = flattenMongoDBResources(spec.Mongod.Resources)
						}
						if spec.Mongos != nil {
							resources["resources_mongos"] = flattenMongoDBResources(spec.Mongos.Resources)
						}
						if spec.Mongocfg != nil {
							resources["resources_mongocfg"] = flattenMongoDBResources(spec.Mongocfg.Resources)
						}
						if spec.Mongoinfra != nil {
							resources["resources_mongoinfra"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
						}
					}
					if len(resources) == 0 {
						return nil, fmt.Errorf("Non empty service not found in mongo spec")
					}
					return resources, nil
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongod := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_6_0Enterprise).Mongodb_6_0Enterprise.Mongod
					if mongod != nil {
						user_config := mongod.GetConfig().GetUserConfig()
						default_config := mongod.GetConfig().GetDefaultConfig()

						result := map[string]interface{}{}

						if security := user_config.GetSecurity(); security != nil {
							flattenSecurity := map[string]interface{}{}
							if enableEncription := security.GetEnableEncryption(); enableEncription != nil {
								flattenSecurity["enable_encryption"] = enableEncription.GetValue()
							}
							if kmip := security.GetKmip(); kmip != nil {
								flattenKmip := map[string]interface{}{}
								flattenKmip["server_name"] = kmip.GetServerName()
								flattenKmip["port"] = int(kmip.GetPort().GetValue())
								flattenKmip["server_ca"] = kmip.GetServerCa()
								flattenKmip["client_certificate"] = d.Get("cluster_config.0.mongod.0.security.0.kmip.0.client_certificate")
								flattenKmip["key_identifier"] = kmip.GetKeyIdentifier()

								flattenSecurity["kmip"] = []map[string]interface{}{flattenKmip}
							}
							result["security"] = []map[string]interface{}{flattenSecurity}
						}

						if audit_log := user_config.GetAuditLog(); audit_log != nil {
							audit_log_data := map[string]interface{}{}
							if audit_log.GetFilter() != default_config.GetAuditLog().GetFilter() {
								audit_log_data["filter"] = audit_log.GetFilter()
							}
							if audit_log.GetRuntimeConfiguration() != nil {
								audit_log_data["runtime_configuration"] = audit_log.GetRuntimeConfiguration().GetValue()
							}
							result["audit_log"] = []map[string]interface{}{audit_log_data}
						}
						if setParameter := user_config.GetSetParameter(); setParameter != nil {
							setParameterData := map[string]interface{}{}
							if setParameter.GetAuditAuthorizationSuccess() != nil {
								setParameterData["audit_authorization_success"] = setParameter.GetAuditAuthorizationSuccess().GetValue()
							}
							if enableFlowControl := setParameter.GetEnableFlowControl(); enableFlowControl != nil {
								setParameterData["enable_flow_control"] = enableFlowControl.GetValue()
							}
							if minSnapshotHistoryWindowInSeconds := setParameter.GetMinSnapshotHistoryWindowInSeconds(); minSnapshotHistoryWindowInSeconds != nil {
								setParameterData["min_snapshot_history_window_in_seconds"] = minSnapshotHistoryWindowInSeconds.GetValue()
							}
							result["set_parameter"] = []map[string]interface{}{setParameterData}
						}

						if net := user_config.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongodConfig6_0Enterprise_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := user_config.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								if collectionConfig := wiredTiger.GetCollectionConfig(); collectionConfig != nil {
									if blockCompressor := collectionConfig.GetBlockCompressor(); blockCompressor != 0 {
										flattenWiredTiger["block_compressor"] = blockCompressor.String()
									}
								}
								if indexConfig := wiredTiger.GetIndexConfig(); indexConfig != nil {
									if prefixCompression := indexConfig.GetPrefixCompression(); prefixCompression != nil {
										flattenWiredTiger["prefix_compression"] = prefixCompression.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}

							if journal := storage.GetJournal(); journal != nil {
								flattenJournal := map[string]interface{}{}
								if commitInterval := journal.GetCommitInterval(); commitInterval != nil {
									flattenJournal["commit_interval"] = commitInterval.GetValue()
								}
								flattenStorage["journal"] = []map[string]interface{}{flattenJournal}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := user_config.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							if opSampleRate := opProfiling.GetSlowOpSampleRate(); opSampleRate != nil {
								flattenOpProfiling["slow_op_sample_rate"] = opSampleRate.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongos: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_6_0Enterprise).Mongodb_6_0Enterprise
					userConfig := mongodbConfig.Mongos.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongos().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongosConfig6_0Enterprise_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}
						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongocfg: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_6_0Enterprise).Mongodb_6_0Enterprise
					userConfig := mongodbConfig.Mongocfg.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongocfg().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := userConfig.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := userConfig.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					configMongod := mongo_config.MongodConfig6_0Enterprise{}
					configMongos := mongo_config.MongosConfig6_0Enterprise{}
					configMongoCfg := mongo_config.MongoCfgConfig6_0Enterprise{}

					security := mongo_config.MongodConfig6_0Enterprise_Security{}
					if enable_encryption := d.Get("cluster_config.0.mongod.0.security.0.enable_encryption"); enable_encryption != nil {
						security.SetEnableEncryption(&wrappers.BoolValue{Value: enable_encryption.(bool)})
					}
					kmip := mongo_config.MongodConfig6_0Enterprise_Security_KMIP{}
					if server_name := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.server_name"); server_name != nil {
						kmip.SetServerName(server_name.(string))
					}
					if port := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.port"); port != nil {
						kmip.SetPort(&wrappers.Int64Value{Value: int64(port.(int))})
					}
					if server_ca := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.server_ca"); server_ca != nil {
						kmip.SetServerCa(server_ca.(string))
					}
					if client_certificate := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.client_certificate"); client_certificate != nil {
						kmip.SetClientCertificate(client_certificate.(string))
					}
					if key_identifier := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.key_identifier"); key_identifier != nil {
						kmip.SetKeyIdentifier(key_identifier.(string))
					}
					security.SetKmip(&kmip)
					configMongod.SetSecurity(&security)
					audit_log := mongo_config.MongodConfig6_0Enterprise_AuditLog{}
					if filter := d.Get("cluster_config.0.mongod.0.audit_log.0.filter"); filter != nil {
						audit_log.SetFilter(filter.(string))
					}
					// Note: right now runtime_configuration unsupported, so we should comment this statement
					//if rt := d.Get("cluster_config.0.mongod.0.audit_log.0.runtime_configuration"); rt != nil {
					//	audit_log.SetRuntimeConfiguration(&wrappers.BoolValue{Value: rt.(bool)})
					//}
					configMongod.SetAuditLog(&audit_log)

					setParameter := mongo_config.MongodConfig6_0Enterprise_SetParameter{}
					if success := d.Get("cluster_config.0.mongod.0.set_parameter.0.audit_authorization_success"); success != nil {
						setParameter.SetAuditAuthorizationSuccess(&wrappers.BoolValue{Value: success.(bool)})
					}
					if flowControl, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter.0.enable_flow_control"); ok {
						setParameter.SetEnableFlowControl(&wrappers.BoolValue{Value: flowControl.(bool)})
					}
					if minSnapshotHistoryWindowInSeconds, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter.0.min_snapshot_history_window_in_seconds"); ok {
						setParameter.SetMinSnapshotHistoryWindowInSeconds(&wrappers.Int64Value{Value: int64(minSnapshotHistoryWindowInSeconds.(int))})
					}
					configMongod.SetSetParameter(&setParameter)
					if _, ok := d.GetOk("cluster_config.0.mongod.0.net"); ok {
						netMongod := mongo_config.MongodConfig6_0Enterprise_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongod.0.net.0.max_incoming_connections"); ok {
							netMongod.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongod.0.net.0.compressors"); ok {
							compressionMongod := mongo_config.MongodConfig6_0Enterprise_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongodConfig6_0Enterprise_Network_Compression_Compressor {
									compressorInt := mongo_config.MongodConfig6_0Enterprise_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongodConfig6_0Enterprise_Network_Compression_Compressor(compressorInt)
								})
							compressionMongod.SetCompressors(modifiedCompressors)
							netMongod.SetCompression(&compressionMongod)
						}
						configMongod.SetNet(&netMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling"); ok {
						opProfilingMongod := mongo_config.MongodConfig6_0Enterprise_OperationProfiling{}

						if mode, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongodConfig6_0_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongod.SetMode(mongo_config.MongodConfig6_0Enterprise_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongod.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}

						if opSampleRate, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate"); ok {
							opProfilingMongod.SetSlowOpSampleRate(&wrappers.DoubleValue{Value: opSampleRate.(float64)})
						}
						configMongod.SetOperationProfiling(&opProfilingMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.storage"); ok {
						engineConfigMongod := mongo_config.MongodConfig6_0Enterprise_Storage_WiredTiger_EngineConfig{}
						collectionConfigMongod := mongo_config.MongodConfig6_0Enterprise_Storage_WiredTiger_CollectionConfig{}
						indexConfigMongod := mongo_config.MongodConfig6_0Enterprise_Storage_WiredTiger_IndexConfig{}
						journalMongod := mongo_config.MongodConfig6_0Enterprise_Storage_Journal{}
						wiredTigerMongod := mongo_config.MongodConfig6_0Enterprise_Storage_WiredTiger{
							EngineConfig:     &engineConfigMongod,
							CollectionConfig: &collectionConfigMongod,
							IndexConfig:      &indexConfigMongod,
						}
						storageMongod := mongo_config.MongodConfig6_0Enterprise_Storage{
							WiredTiger: &wiredTigerMongod,
							Journal:    &journalMongod,
						}
						if cacheSize, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongod.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						if blockCompressor, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"); ok {
							blockCompressorInt := mongo_config.MongodConfig6_0Enterprise_Storage_WiredTiger_CollectionConfig_Compressor_value[strings.ToUpper(blockCompressor.(string))]
							collectionConfigMongod.SetBlockCompressor(
								mongo_config.MongodConfig6_0Enterprise_Storage_WiredTiger_CollectionConfig_Compressor(blockCompressorInt),
							)
						}
						if prefixCompression, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression"); ok {
							indexConfigMongod.SetPrefixCompression(&wrappers.BoolValue{Value: prefixCompression.(bool)})
						}
						if commitInterval, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"); ok {
							journalMongod.SetCommitInterval(&wrappers.Int64Value{Value: int64(commitInterval.(int))})
						}
						configMongod.SetStorage(&storageMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongos.0.net"); ok {
						netMongos := mongo_config.MongosConfig6_0Enterprise_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongos.0.net.0.max_incoming_connections"); ok {
							netMongos.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongos.0.net.0.compressors"); ok {
							compressionMongoS := mongo_config.MongosConfig6_0Enterprise_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongosConfig6_0Enterprise_Network_Compression_Compressor {
									compressorInt := mongo_config.MongosConfig6_0Enterprise_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongosConfig6_0Enterprise_Network_Compression_Compressor(compressorInt)
								})
							compressionMongoS.SetCompressors(modifiedCompressors)
							netMongos.SetCompression(&compressionMongoS)
						}
						configMongos.SetNet(&netMongos)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.net"); ok {
						netMongoCfg := mongo_config.MongoCfgConfig6_0Enterprise_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongocfg.0.net.0.max_incoming_connections"); ok {
							netMongoCfg.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongoCfg.SetNet(&netMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling"); ok {
						opProfilingMongoCfg := mongo_config.MongoCfgConfig6_0Enterprise_OperationProfiling{}
						if mode, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongoCfgConfig6_0Enterprise_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongoCfg.SetMode(mongo_config.MongoCfgConfig6_0Enterprise_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongoCfg.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						configMongoCfg.SetOperationProfiling(&opProfilingMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.storage"); ok {
						engineConfigMongoCfg := mongo_config.MongoCfgConfig6_0Enterprise_Storage_WiredTiger_EngineConfig{}
						wiredTigerMongoCfg := mongo_config.MongoCfgConfig6_0Enterprise_Storage_WiredTiger{EngineConfig: &engineConfigMongoCfg}
						storageMongoCfg := mongo_config.MongoCfgConfig6_0Enterprise_Storage{WiredTiger: &wiredTigerMongoCfg}

						if cacheSize, ok := d.GetOk("cluster_config.0.mongocfg.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongoCfg.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						configMongoCfg.SetStorage(&storageMongoCfg)
					}
					hostTypes := getSetOfHostTypes(d)
					var resourcesMongod, resourcesMongos, resourcesMongoCfg, resourcesMongoInfra *mongodb.Resources = getResources(d)
					var mongod *mongodb.MongodbSpec6_0Enterprise_Mongod
					var mongos *mongodb.MongodbSpec6_0Enterprise_Mongos
					var mongocfg *mongodb.MongodbSpec6_0Enterprise_MongoCfg
					var mongoinfra *mongodb.MongodbSpec6_0Enterprise_MongoInfra
					mongod = &mongodb.MongodbSpec6_0Enterprise_Mongod{
						Config:    &configMongod,
						Resources: resourcesMongod,
					}

					if _, ok := hostTypes["MONGOS"]; ok {
						mongos = &mongodb.MongodbSpec6_0Enterprise_Mongos{
							Config:    &configMongos,
							Resources: resourcesMongos,
						}
					}
					if _, ok := hostTypes["MONGOCFG"]; ok {
						mongocfg = &mongodb.MongodbSpec6_0Enterprise_MongoCfg{
							Config:    &configMongoCfg,
							Resources: resourcesMongoCfg,
						}
					}
					if _, ok := hostTypes["MONGOINFRA"]; ok {
						mongoinfra = &mongodb.MongodbSpec6_0Enterprise_MongoInfra{
							ConfigMongocfg: &configMongoCfg,
							ConfigMongos:   &configMongos,
							Resources:      resourcesMongoInfra,
						}
					}
					return &mongodb.ConfigSpec_MongodbSpec_6_0Enterprise{
						MongodbSpec_6_0Enterprise: &mongodb.MongodbSpec6_0Enterprise{
							Mongod:     mongod,
							Mongos:     mongos,
							Mongocfg:   mongocfg,
							Mongoinfra: mongoinfra,
						},
					}
				},
			}
		}
	case "6.0":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig, d *schema.ResourceData) (map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_6_0).Mongodb_6_0
					resources := map[string]interface{}{}
					if _, ok := d.GetOk("resources"); ok {
						if spec.Mongod != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongod.Resources)
							return resources, nil
						}
						if spec.Mongos != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongos.Resources)
							return resources, nil
						}
						if spec.Mongocfg != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongocfg.Resources)
							return resources, nil
						}
						if spec.Mongoinfra != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
							return resources, nil
						}
					} else {
						if spec.Mongod != nil {
							resources["resources_mongod"] = flattenMongoDBResources(spec.Mongod.Resources)
						}
						if spec.Mongos != nil {
							resources["resources_mongos"] = flattenMongoDBResources(spec.Mongos.Resources)
						}
						if spec.Mongocfg != nil {
							resources["resources_mongocfg"] = flattenMongoDBResources(spec.Mongocfg.Resources)
						}
						if spec.Mongoinfra != nil {
							resources["resources_mongoinfra"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
						}
					}
					if len(resources) == 0 {
						return nil, fmt.Errorf("Non empty service not found in mongo spec")
					}
					return resources, nil
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongod := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_6_0).Mongodb_6_0.Mongod
					if mongod != nil {
						user_config := mongod.GetConfig().GetUserConfig()
						result := map[string]interface{}{}
						if setParameter := user_config.GetSetParameter(); setParameter != nil {
							setParameterData := map[string]interface{}{}
							if enableFlowControl := setParameter.GetEnableFlowControl(); enableFlowControl != nil {
								setParameterData["enable_flow_control"] = enableFlowControl.GetValue()
							}
							if minSnapshotHistoryWindowInSeconds := setParameter.GetMinSnapshotHistoryWindowInSeconds(); minSnapshotHistoryWindowInSeconds != nil {
								setParameterData["min_snapshot_history_window_in_seconds"] = minSnapshotHistoryWindowInSeconds.GetValue()
							}
							result["set_parameter"] = []map[string]interface{}{setParameterData}
						}

						if net := user_config.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongodConfig6_0_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := user_config.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								if collectionConfig := wiredTiger.GetCollectionConfig(); collectionConfig != nil {
									if blockCompressor := collectionConfig.GetBlockCompressor(); blockCompressor != 0 {
										flattenWiredTiger["block_compressor"] = blockCompressor.String()
									}
								}
								if indexConfig := wiredTiger.GetIndexConfig(); indexConfig != nil {
									if prefixCompression := indexConfig.GetPrefixCompression(); prefixCompression != nil {
										flattenWiredTiger["prefix_compression"] = prefixCompression.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}

							if journal := storage.GetJournal(); journal != nil {
								flattenJournal := map[string]interface{}{}
								if commitInterval := journal.GetCommitInterval(); commitInterval != nil {
									flattenJournal["commit_interval"] = commitInterval.GetValue()
								}
								flattenStorage["journal"] = []map[string]interface{}{flattenJournal}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := user_config.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							if opSampleRate := opProfiling.GetSlowOpSampleRate(); opSampleRate != nil {
								flattenOpProfiling["slow_op_sample_rate"] = opSampleRate.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}
						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongos: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_6_0).Mongodb_6_0
					userConfig := mongodbConfig.Mongos.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongos().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongosConfig6_0_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}
						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongocfg: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_6_0).Mongodb_6_0
					userConfig := mongodbConfig.Mongocfg.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongocfg().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := userConfig.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := userConfig.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					configMongod := mongo_config.MongodConfig6_0{}
					configMongos := mongo_config.MongosConfig6_0{}
					configMongoCfg := mongo_config.MongoCfgConfig6_0{}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.net"); ok {
						netMongod := mongo_config.MongodConfig6_0_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongod.0.net.0.max_incoming_connections"); ok {
							netMongod.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongod.0.net.0.compressors"); ok {
							compressionMongod := mongo_config.MongodConfig6_0_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongodConfig6_0_Network_Compression_Compressor {
									compressorInt := mongo_config.MongodConfig6_0_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongodConfig6_0_Network_Compression_Compressor(compressorInt)
								})
							compressionMongod.SetCompressors(modifiedCompressors)
							netMongod.SetCompression(&compressionMongod)
						}
						configMongod.SetNet(&netMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter"); ok {
						setParameter := mongo_config.MongodConfig6_0_SetParameter{}
						if flowControl, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter.0.enable_flow_control"); ok {
							setParameter.SetEnableFlowControl(&wrappers.BoolValue{Value: flowControl.(bool)})
						}
						if minSnapshotHistoryWindowInSeconds, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter.0.min_snapshot_history_window_in_seconds"); ok {
							setParameter.SetMinSnapshotHistoryWindowInSeconds(&wrappers.Int64Value{Value: int64(minSnapshotHistoryWindowInSeconds.(int))})
						}
						configMongod.SetSetParameter(&setParameter)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling"); ok {
						opProfilingMongod := mongo_config.MongodConfig6_0_OperationProfiling{}

						if mode, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongodConfig6_0_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongod.SetMode(mongo_config.MongodConfig6_0_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongod.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}

						if opSampleRate, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate"); ok {
							opProfilingMongod.SetSlowOpSampleRate(&wrappers.DoubleValue{Value: opSampleRate.(float64)})
						}
						configMongod.SetOperationProfiling(&opProfilingMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.storage"); ok {
						engineConfigMongod := mongo_config.MongodConfig6_0_Storage_WiredTiger_EngineConfig{}
						collectionConfigMongod := mongo_config.MongodConfig6_0_Storage_WiredTiger_CollectionConfig{}
						indexConfigMongod := mongo_config.MongodConfig6_0_Storage_WiredTiger_IndexConfig{}
						journalMongod := mongo_config.MongodConfig6_0_Storage_Journal{}
						wiredTigerMongod := mongo_config.MongodConfig6_0_Storage_WiredTiger{
							EngineConfig:     &engineConfigMongod,
							CollectionConfig: &collectionConfigMongod,
							IndexConfig:      &indexConfigMongod,
						}
						storageMongod := mongo_config.MongodConfig6_0_Storage{
							WiredTiger: &wiredTigerMongod,
							Journal:    &journalMongod,
						}
						if cacheSize, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongod.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						if blockCompressor, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"); ok {
							blockCompressorInt := mongo_config.MongodConfig6_0_Storage_WiredTiger_CollectionConfig_Compressor_value[strings.ToUpper(blockCompressor.(string))]
							collectionConfigMongod.SetBlockCompressor(
								mongo_config.MongodConfig6_0_Storage_WiredTiger_CollectionConfig_Compressor(blockCompressorInt),
							)
						}
						if prefixCompression, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression"); ok {
							indexConfigMongod.SetPrefixCompression(&wrappers.BoolValue{Value: prefixCompression.(bool)})
						}
						if commitInterval, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"); ok {
							journalMongod.SetCommitInterval(&wrappers.Int64Value{Value: int64(commitInterval.(int))})
						}
						configMongod.SetStorage(&storageMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongos.0.net"); ok {
						netMongos := mongo_config.MongosConfig6_0_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongos.0.net.0.max_incoming_connections"); ok {
							netMongos.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongos.0.net.0.compressors"); ok {
							compressionMongoS := mongo_config.MongosConfig6_0_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongosConfig6_0_Network_Compression_Compressor {
									compressorInt := mongo_config.MongosConfig6_0_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongosConfig6_0_Network_Compression_Compressor(compressorInt)
								})
							compressionMongoS.SetCompressors(modifiedCompressors)
							netMongos.SetCompression(&compressionMongoS)
						}
						configMongos.SetNet(&netMongos)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.net"); ok {
						netMongoCfg := mongo_config.MongoCfgConfig6_0_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongocfg.0.net.0.max_incoming_connections"); ok {
							netMongoCfg.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongoCfg.SetNet(&netMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling"); ok {
						opProfilingMongoCfg := mongo_config.MongoCfgConfig6_0_OperationProfiling{}
						if mode, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongoCfgConfig6_0_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongoCfg.SetMode(mongo_config.MongoCfgConfig6_0_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongoCfg.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						configMongoCfg.SetOperationProfiling(&opProfilingMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.storage"); ok {
						engineConfigMongoCfg := mongo_config.MongoCfgConfig6_0_Storage_WiredTiger_EngineConfig{}
						wiredTigerMongoCfg := mongo_config.MongoCfgConfig6_0_Storage_WiredTiger{EngineConfig: &engineConfigMongoCfg}
						storageMongoCfg := mongo_config.MongoCfgConfig6_0_Storage{WiredTiger: &wiredTigerMongoCfg}

						if cacheSize, ok := d.GetOk("cluster_config.0.mongocfg.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongoCfg.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						configMongoCfg.SetStorage(&storageMongoCfg)
					}
					hostTypes := getSetOfHostTypes(d)
					var resourcesMongod, resourcesMongos, resourcesMongoCfg, resourcesMongoInfra *mongodb.Resources = getResources(d)
					var mongod *mongodb.MongodbSpec6_0_Mongod
					var mongos *mongodb.MongodbSpec6_0_Mongos
					var mongocfg *mongodb.MongodbSpec6_0_MongoCfg
					var mongoinfra *mongodb.MongodbSpec6_0_MongoInfra
					mongod = &mongodb.MongodbSpec6_0_Mongod{
						Config:    &configMongod,
						Resources: resourcesMongod,
					}

					if _, ok := hostTypes["MONGOS"]; ok {
						mongos = &mongodb.MongodbSpec6_0_Mongos{
							Config:    &configMongos,
							Resources: resourcesMongos,
						}
					}
					if _, ok := hostTypes["MONGOCFG"]; ok {
						mongocfg = &mongodb.MongodbSpec6_0_MongoCfg{
							Config:    &configMongoCfg,
							Resources: resourcesMongoCfg,
						}
					}
					if _, ok := hostTypes["MONGOINFRA"]; ok {
						mongoinfra = &mongodb.MongodbSpec6_0_MongoInfra{
							ConfigMongocfg: &configMongoCfg,
							ConfigMongos:   &configMongos,
							Resources:      resourcesMongoInfra,
						}
					}

					return &mongodb.ConfigSpec_MongodbSpec_6_0{
						MongodbSpec_6_0: &mongodb.MongodbSpec6_0{
							Mongod:     mongod,
							Mongos:     mongos,
							Mongocfg:   mongocfg,
							Mongoinfra: mongoinfra,
						},
					}
				},
			}
		}
	case "5.0-enterprise":
		{
			return &MongodbSpecHelper{
				FlattenResources: func(c *mongodb.ClusterConfig, d *schema.ResourceData) (map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0Enterprise).Mongodb_5_0Enterprise
					resources := map[string]interface{}{}
					if _, ok := d.GetOk("resources"); ok {
						if spec.Mongod != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongod.Resources)
							return resources, nil
						}
						if spec.Mongos != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongos.Resources)
							return resources, nil
						}
						if spec.Mongocfg != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongocfg.Resources)
							return resources, nil
						}
						if spec.Mongoinfra != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
							return resources, nil
						}
					} else {
						if spec.Mongod != nil {
							resources["resources_mongod"] = flattenMongoDBResources(spec.Mongod.Resources)
						}
						if spec.Mongos != nil {
							resources["resources_mongos"] = flattenMongoDBResources(spec.Mongos.Resources)
						}
						if spec.Mongocfg != nil {
							resources["resources_mongocfg"] = flattenMongoDBResources(spec.Mongocfg.Resources)
						}
						if spec.Mongoinfra != nil {
							resources["resources_mongoinfra"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
						}
					}
					if len(resources) == 0 {
						return nil, fmt.Errorf("Non empty service not found in mongo spec")
					}
					return resources, nil
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongod := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0Enterprise).Mongodb_5_0Enterprise.Mongod
					if mongod != nil {
						user_config := mongod.GetConfig().GetUserConfig()
						default_config := mongod.GetConfig().GetDefaultConfig()

						result := map[string]interface{}{}

						if security := user_config.GetSecurity(); security != nil {
							flattenSecurity := map[string]interface{}{}
							if enableEncription := security.GetEnableEncryption(); enableEncription != nil {
								flattenSecurity["enable_encryption"] = enableEncription.GetValue()
							}
							if kmip := security.GetKmip(); kmip != nil {
								flattenKmip := map[string]interface{}{}
								flattenKmip["server_name"] = kmip.GetServerName()
								flattenKmip["port"] = int(kmip.GetPort().GetValue())
								flattenKmip["server_ca"] = kmip.GetServerCa()
								flattenKmip["client_certificate"] = d.Get("cluster_config.0.mongod.0.security.0.kmip.0.client_certificate")
								flattenKmip["key_identifier"] = kmip.GetKeyIdentifier()

								flattenSecurity["kmip"] = []map[string]interface{}{flattenKmip}
							}
							result["security"] = []map[string]interface{}{flattenSecurity}
						}

						if audit_log := user_config.GetAuditLog(); audit_log != nil {
							audit_log_data := map[string]interface{}{}
							if audit_log.GetFilter() != default_config.GetAuditLog().GetFilter() {
								audit_log_data["filter"] = audit_log.GetFilter()
							}
							if audit_log.GetRuntimeConfiguration() != nil {
								audit_log_data["runtime_configuration"] = audit_log.GetRuntimeConfiguration().GetValue()
							}
							result["audit_log"] = []map[string]interface{}{audit_log_data}
						}
						if setParameter := user_config.GetSetParameter(); setParameter != nil {
							setParameterData := map[string]interface{}{}
							if setParameter.GetAuditAuthorizationSuccess() != nil {
								setParameterData["audit_authorization_success"] = setParameter.GetAuditAuthorizationSuccess().GetValue()
							}
							if enableFlowControl := setParameter.GetEnableFlowControl(); enableFlowControl != nil {
								setParameterData["enable_flow_control"] = enableFlowControl.GetValue()
							}
							if minSnapshotHistoryWindowInSeconds := setParameter.GetMinSnapshotHistoryWindowInSeconds(); minSnapshotHistoryWindowInSeconds != nil {
								setParameterData["min_snapshot_history_window_in_seconds"] = minSnapshotHistoryWindowInSeconds.GetValue()
							}
							result["set_parameter"] = []map[string]interface{}{setParameterData}
						}

						if net := user_config.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongodConfig5_0Enterprise_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := user_config.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								if collectionConfig := wiredTiger.GetCollectionConfig(); collectionConfig != nil {
									if blockCompressor := collectionConfig.GetBlockCompressor(); blockCompressor != 0 {
										flattenWiredTiger["block_compressor"] = blockCompressor.String()
									}
								}
								if indexConfig := wiredTiger.GetIndexConfig(); indexConfig != nil {
									if prefixCompression := indexConfig.GetPrefixCompression(); prefixCompression != nil {
										flattenWiredTiger["prefix_compression"] = prefixCompression.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}

							if journal := storage.GetJournal(); journal != nil {
								flattenJournal := map[string]interface{}{}
								if commitInterval := journal.GetCommitInterval(); commitInterval != nil {
									flattenJournal["commit_interval"] = commitInterval.GetValue()
								}
								flattenStorage["journal"] = []map[string]interface{}{flattenJournal}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := user_config.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							if opSampleRate := opProfiling.GetSlowOpSampleRate(); opSampleRate != nil {
								flattenOpProfiling["slow_op_sample_rate"] = opSampleRate.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongos: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0Enterprise).Mongodb_5_0Enterprise
					userConfig := mongodbConfig.Mongos.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongos().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongosConfig5_0Enterprise_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}
						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongocfg: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0Enterprise).Mongodb_5_0Enterprise
					userConfig := mongodbConfig.Mongocfg.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongocfg().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := userConfig.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := userConfig.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					configMongod := mongo_config.MongodConfig5_0Enterprise{}
					configMongos := mongo_config.MongosConfig5_0Enterprise{}
					configMongoCfg := mongo_config.MongoCfgConfig5_0Enterprise{}

					security := mongo_config.MongodConfig5_0Enterprise_Security{}
					if enable_encryption := d.Get("cluster_config.0.mongod.0.security.0.enable_encryption"); enable_encryption != nil {
						security.SetEnableEncryption(&wrappers.BoolValue{Value: enable_encryption.(bool)})
					}
					kmip := mongo_config.MongodConfig5_0Enterprise_Security_KMIP{}
					if server_name := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.server_name"); server_name != nil {
						kmip.SetServerName(server_name.(string))
					}
					if port := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.port"); port != nil {
						kmip.SetPort(&wrappers.Int64Value{Value: int64(port.(int))})
					}
					if server_ca := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.server_ca"); server_ca != nil {
						kmip.SetServerCa(server_ca.(string))
					}
					if client_certificate := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.client_certificate"); client_certificate != nil {
						kmip.SetClientCertificate(client_certificate.(string))
					}
					if key_identifier := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.key_identifier"); key_identifier != nil {
						kmip.SetKeyIdentifier(key_identifier.(string))
					}
					security.SetKmip(&kmip)
					configMongod.SetSecurity(&security)
					audit_log := mongo_config.MongodConfig5_0Enterprise_AuditLog{}
					if filter := d.Get("cluster_config.0.mongod.0.audit_log.0.filter"); filter != nil {
						audit_log.SetFilter(filter.(string))
					}
					// Note: right now runtime_configuration unsupported, so we should comment this statement
					//if rt := d.Get("cluster_config.0.mongod.0.audit_log.0.runtime_configuration"); rt != nil {
					//	audit_log.SetRuntimeConfiguration(&wrappers.BoolValue{Value: rt.(bool)})
					//}
					configMongod.SetAuditLog(&audit_log)

					setParameter := mongo_config.MongodConfig5_0Enterprise_SetParameter{}
					if success := d.Get("cluster_config.0.mongod.0.set_parameter.0.audit_authorization_success"); success != nil {
						setParameter.SetAuditAuthorizationSuccess(&wrappers.BoolValue{Value: success.(bool)})
					}
					if flowControl, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter.0.enable_flow_control"); ok {
						setParameter.SetEnableFlowControl(&wrappers.BoolValue{Value: flowControl.(bool)})
					}
					if minSnapshotHistoryWindowInSeconds, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter.0.min_snapshot_history_window_in_seconds"); ok {
						setParameter.SetMinSnapshotHistoryWindowInSeconds(&wrappers.Int64Value{Value: int64(minSnapshotHistoryWindowInSeconds.(int))})
					}
					configMongod.SetSetParameter(&setParameter)
					if _, ok := d.GetOk("cluster_config.0.mongod.0.net"); ok {
						netMongod := mongo_config.MongodConfig5_0Enterprise_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongod.0.net.0.max_incoming_connections"); ok {
							netMongod.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongod.0.net.0.compressors"); ok {
							netCompressionMongod := mongo_config.MongodConfig5_0Enterprise_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongodConfig5_0Enterprise_Network_Compression_Compressor {
									compressorInt := mongo_config.MongodConfig5_0Enterprise_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongodConfig5_0Enterprise_Network_Compression_Compressor(compressorInt)
								})
							netCompressionMongod.SetCompressors(modifiedCompressors)
							netMongod.SetCompression(&netCompressionMongod)
						}
						configMongod.SetNet(&netMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling"); ok {
						opProfilingMongod := mongo_config.MongodConfig5_0Enterprise_OperationProfiling{}

						if mode, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongodConfig5_0_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongod.SetMode(mongo_config.MongodConfig5_0Enterprise_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongod.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}

						if opSampleRate, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate"); ok {
							opProfilingMongod.SetSlowOpSampleRate(&wrappers.DoubleValue{Value: opSampleRate.(float64)})
						}

						configMongod.SetOperationProfiling(&opProfilingMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.storage"); ok {
						engineConfigMongod := mongo_config.MongodConfig5_0Enterprise_Storage_WiredTiger_EngineConfig{}
						collectionConfigMongod := mongo_config.MongodConfig5_0Enterprise_Storage_WiredTiger_CollectionConfig{}
						indexConfigMongod := mongo_config.MongodConfig5_0Enterprise_Storage_WiredTiger_IndexConfig{}
						journalMongod := mongo_config.MongodConfig5_0Enterprise_Storage_Journal{}
						wiredTigerMongod := mongo_config.MongodConfig5_0Enterprise_Storage_WiredTiger{
							EngineConfig:     &engineConfigMongod,
							CollectionConfig: &collectionConfigMongod,
							IndexConfig:      &indexConfigMongod,
						}
						storageMongod := mongo_config.MongodConfig5_0Enterprise_Storage{
							WiredTiger: &wiredTigerMongod,
							Journal:    &journalMongod,
						}
						if cacheSize, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongod.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						if blockCompressor, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"); ok {
							blockCompressorInt := mongo_config.MongodConfig5_0Enterprise_Storage_WiredTiger_CollectionConfig_Compressor_value[strings.ToUpper(blockCompressor.(string))]
							collectionConfigMongod.SetBlockCompressor(
								mongo_config.MongodConfig5_0Enterprise_Storage_WiredTiger_CollectionConfig_Compressor(blockCompressorInt),
							)
						}
						if prefixCompression, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression"); ok {
							indexConfigMongod.SetPrefixCompression(&wrappers.BoolValue{Value: prefixCompression.(bool)})
						}

						if commitInterval, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"); ok {
							journalMongod.SetCommitInterval(&wrappers.Int64Value{Value: int64(commitInterval.(int))})
						}
						configMongod.SetStorage(&storageMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongos.0.net"); ok {
						netMongos := mongo_config.MongosConfig5_0Enterprise_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongos.0.net.0.max_incoming_connections"); ok {
							netMongos.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongos.0.net.0.compressors"); ok {
							netCompressionMongos := mongo_config.MongosConfig5_0Enterprise_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongosConfig5_0Enterprise_Network_Compression_Compressor {
									compressorInt := mongo_config.MongosConfig5_0Enterprise_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongosConfig5_0Enterprise_Network_Compression_Compressor(compressorInt)
								})
							netCompressionMongos.SetCompressors(modifiedCompressors)
							netMongos.SetCompression(&netCompressionMongos)
						}
						configMongos.SetNet(&netMongos)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.net"); ok {
						netMongoCfg := mongo_config.MongoCfgConfig5_0Enterprise_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongocfg.0.net.0.max_incoming_connections"); ok {
							netMongoCfg.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongoCfg.SetNet(&netMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling"); ok {
						opProfilingMongoCfg := mongo_config.MongoCfgConfig5_0Enterprise_OperationProfiling{}
						if mode, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongoCfgConfig5_0Enterprise_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongoCfg.SetMode(mongo_config.MongoCfgConfig5_0Enterprise_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongoCfg.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						configMongoCfg.SetOperationProfiling(&opProfilingMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.storage"); ok {
						engineConfigMongoCfg := mongo_config.MongoCfgConfig5_0Enterprise_Storage_WiredTiger_EngineConfig{}
						wiredTigerMongoCfg := mongo_config.MongoCfgConfig5_0Enterprise_Storage_WiredTiger{EngineConfig: &engineConfigMongoCfg}
						storageMongoCfg := mongo_config.MongoCfgConfig5_0Enterprise_Storage{WiredTiger: &wiredTigerMongoCfg}

						if cacheSize, ok := d.GetOk("cluster_config.0.mongocfg.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongoCfg.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						configMongoCfg.SetStorage(&storageMongoCfg)
					}
					hostTypes := getSetOfHostTypes(d)
					var resourcesMongod, resourcesMongos, resourcesMongoCfg, resourcesMongoInfra *mongodb.Resources = getResources(d)
					var mongod *mongodb.MongodbSpec5_0Enterprise_Mongod
					var mongos *mongodb.MongodbSpec5_0Enterprise_Mongos
					var mongocfg *mongodb.MongodbSpec5_0Enterprise_MongoCfg
					var mongoinfra *mongodb.MongodbSpec5_0Enterprise_MongoInfra
					mongod = &mongodb.MongodbSpec5_0Enterprise_Mongod{
						Config:    &configMongod,
						Resources: resourcesMongod,
					}

					if _, ok := hostTypes["MONGOS"]; ok {
						mongos = &mongodb.MongodbSpec5_0Enterprise_Mongos{
							Config:    &configMongos,
							Resources: resourcesMongos,
						}
					}
					if _, ok := hostTypes["MONGOCFG"]; ok {
						mongocfg = &mongodb.MongodbSpec5_0Enterprise_MongoCfg{
							Config:    &configMongoCfg,
							Resources: resourcesMongoCfg,
						}
					}
					if _, ok := hostTypes["MONGOINFRA"]; ok {
						mongoinfra = &mongodb.MongodbSpec5_0Enterprise_MongoInfra{
							ConfigMongocfg: &configMongoCfg,
							ConfigMongos:   &configMongos,
							Resources:      resourcesMongoInfra,
						}
					}
					return &mongodb.ConfigSpec_MongodbSpec_5_0Enterprise{
						MongodbSpec_5_0Enterprise: &mongodb.MongodbSpec5_0Enterprise{
							Mongod:     mongod,
							Mongos:     mongos,
							Mongocfg:   mongocfg,
							Mongoinfra: mongoinfra,
						},
					}
				},
			}
		}
	case "5.0":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig, d *schema.ResourceData) (map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0).Mongodb_5_0
					resources := map[string]interface{}{}
					if _, ok := d.GetOk("resources"); ok {
						if spec.Mongod != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongod.Resources)
							return resources, nil
						}
						if spec.Mongos != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongos.Resources)
							return resources, nil
						}
						if spec.Mongocfg != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongocfg.Resources)
							return resources, nil
						}
						if spec.Mongoinfra != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
							return resources, nil
						}
					} else {
						if spec.Mongod != nil {
							resources["resources_mongod"] = flattenMongoDBResources(spec.Mongod.Resources)
						}
						if spec.Mongos != nil {
							resources["resources_mongos"] = flattenMongoDBResources(spec.Mongos.Resources)
						}
						if spec.Mongocfg != nil {
							resources["resources_mongocfg"] = flattenMongoDBResources(spec.Mongocfg.Resources)
						}
						if spec.Mongoinfra != nil {
							resources["resources_mongoinfra"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
						}
					}
					if len(resources) == 0 {
						return nil, fmt.Errorf("Non empty service not found in mongo spec")
					}
					return resources, nil
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongod := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0).Mongodb_5_0.Mongod
					if mongod != nil {
						user_config := mongod.GetConfig().GetUserConfig()
						result := map[string]interface{}{}

						if setParameter := user_config.GetSetParameter(); setParameter != nil {
							flattenSetParameter := map[string]interface{}{}
							if flowControl := setParameter.GetEnableFlowControl(); flowControl != nil {
								flattenSetParameter["enable_flow_control"] = flowControl.GetValue()
							}
							if minSnapshotHistoryWindowInSeconds := setParameter.GetMinSnapshotHistoryWindowInSeconds(); minSnapshotHistoryWindowInSeconds != nil {
								flattenSetParameter["min_snapshot_history_window_in_seconds"] = minSnapshotHistoryWindowInSeconds.GetValue()
							}
							result["set_parameter"] = []map[string]interface{}{flattenSetParameter}
						}

						if net := user_config.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongodConfig5_0_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := user_config.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								if collectionConfig := wiredTiger.GetCollectionConfig(); collectionConfig != nil {
									if blockCompressor := collectionConfig.GetBlockCompressor(); blockCompressor != 0 {
										flattenWiredTiger["block_compressor"] = blockCompressor.String()
									}
								}
								if indexConfig := wiredTiger.GetIndexConfig(); indexConfig != nil {
									if prefixCompression := indexConfig.GetPrefixCompression(); prefixCompression != nil {
										flattenWiredTiger["prefix_compression"] = prefixCompression.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}

							if journal := storage.GetJournal(); journal != nil {
								flattenJournal := map[string]interface{}{}
								if commitInterval := journal.GetCommitInterval(); commitInterval != nil {
									flattenJournal["commit_interval"] = commitInterval.GetValue()
								}
								flattenStorage["journal"] = []map[string]interface{}{flattenJournal}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := user_config.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							if opSampleRate := opProfiling.GetSlowOpSampleRate(); opSampleRate != nil {
								flattenOpProfiling["slow_op_sample_rate"] = opSampleRate.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongos: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0).Mongodb_5_0
					userConfig := mongodbConfig.Mongos.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongos().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongosConfig5_0_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}
						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongocfg: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0).Mongodb_5_0
					userConfig := mongodbConfig.Mongocfg.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongocfg().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := userConfig.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := userConfig.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					configMongod := mongo_config.MongodConfig5_0{}
					configMongos := mongo_config.MongosConfig5_0{}
					configMongoCfg := mongo_config.MongoCfgConfig5_0{}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.net"); ok {
						netMongod := mongo_config.MongodConfig5_0_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongod.0.net.0.max_incoming_connections"); ok {
							netMongod.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongod.0.net.0.compressors"); ok {
							netCompressionMongod := mongo_config.MongodConfig5_0_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongodConfig5_0_Network_Compression_Compressor {
									compressorInt := mongo_config.MongodConfig5_0_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongodConfig5_0_Network_Compression_Compressor(compressorInt)
								})
							netCompressionMongod.SetCompressors(modifiedCompressors)
							netMongod.SetCompression(&netCompressionMongod)
						}
						configMongod.SetNet(&netMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter"); ok {
						setParameter := mongo_config.MongodConfig5_0_SetParameter{}
						if flowControl, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter.0.enable_flow_control"); ok {
							setParameter.SetEnableFlowControl(&wrappers.BoolValue{Value: flowControl.(bool)})
						}
						if minSnapshotHistoryWindowInSeconds, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter.0.min_snapshot_history_window_in_seconds"); ok {
							setParameter.SetMinSnapshotHistoryWindowInSeconds(&wrappers.Int64Value{Value: int64(minSnapshotHistoryWindowInSeconds.(int))})
						}
						configMongod.SetSetParameter(&setParameter)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling"); ok {
						opProfilingMongod := mongo_config.MongodConfig5_0_OperationProfiling{}

						if mode, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongodConfig5_0_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongod.SetMode(mongo_config.MongodConfig5_0_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongod.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}

						if opSampleRate, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate"); ok {
							opProfilingMongod.SetSlowOpSampleRate(&wrappers.DoubleValue{Value: opSampleRate.(float64)})
						}

						configMongod.SetOperationProfiling(&opProfilingMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.storage"); ok {
						engineConfigMongod := mongo_config.MongodConfig5_0_Storage_WiredTiger_EngineConfig{}
						collectionConfigMongod := mongo_config.MongodConfig5_0_Storage_WiredTiger_CollectionConfig{}
						indexConfigMongod := mongo_config.MongodConfig5_0_Storage_WiredTiger_IndexConfig{}
						journalMongod := mongo_config.MongodConfig5_0_Storage_Journal{}
						wiredTigerMongod := mongo_config.MongodConfig5_0_Storage_WiredTiger{
							EngineConfig:     &engineConfigMongod,
							CollectionConfig: &collectionConfigMongod,
							IndexConfig:      &indexConfigMongod,
						}
						storageMongod := mongo_config.MongodConfig5_0_Storage{
							WiredTiger: &wiredTigerMongod,
							Journal:    &journalMongod,
						}
						if cacheSize, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongod.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						if blockCompressor, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"); ok {
							blockCompressorInt := mongo_config.MongodConfig5_0_Storage_WiredTiger_CollectionConfig_Compressor_value[strings.ToUpper(blockCompressor.(string))]
							collectionConfigMongod.SetBlockCompressor(
								mongo_config.MongodConfig5_0_Storage_WiredTiger_CollectionConfig_Compressor(blockCompressorInt),
							)
						}
						if prefixCompression, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression"); ok {
							indexConfigMongod.SetPrefixCompression(&wrappers.BoolValue{Value: prefixCompression.(bool)})
						}
						if commitInterval, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"); ok {
							journalMongod.SetCommitInterval(&wrappers.Int64Value{Value: int64(commitInterval.(int))})
						}
						configMongod.SetStorage(&storageMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongos.0.net"); ok {
						netMongos := mongo_config.MongosConfig5_0_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongos.0.net.0.max_incoming_connections"); ok {
							netMongos.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongos.0.net.0.compressors"); ok {
							netCompressionMongos := mongo_config.MongosConfig5_0_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongosConfig5_0_Network_Compression_Compressor {
									compressorInt := mongo_config.MongosConfig5_0_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongosConfig5_0_Network_Compression_Compressor(compressorInt)
								})
							netCompressionMongos.SetCompressors(modifiedCompressors)
							netMongos.SetCompression(&netCompressionMongos)
						}
						configMongos.SetNet(&netMongos)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.net"); ok {
						netMongoCfg := mongo_config.MongoCfgConfig5_0_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongocfg.0.net.0.max_incoming_connections"); ok {
							netMongoCfg.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongoCfg.SetNet(&netMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling"); ok {
						opProfilingMongoCfg := mongo_config.MongoCfgConfig5_0_OperationProfiling{}
						if mode, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongoCfgConfig5_0_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongoCfg.SetMode(mongo_config.MongoCfgConfig5_0_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongoCfg.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						configMongoCfg.SetOperationProfiling(&opProfilingMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.storage"); ok {
						engineConfigMongoCfg := mongo_config.MongoCfgConfig5_0_Storage_WiredTiger_EngineConfig{}
						wiredTigerMongoCfg := mongo_config.MongoCfgConfig5_0_Storage_WiredTiger{EngineConfig: &engineConfigMongoCfg}
						storageMongoCfg := mongo_config.MongoCfgConfig5_0_Storage{WiredTiger: &wiredTigerMongoCfg}

						if cacheSize, ok := d.GetOk("cluster_config.0.mongocfg.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongoCfg.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						configMongoCfg.SetStorage(&storageMongoCfg)
					}
					hostTypes := getSetOfHostTypes(d)
					var resourcesMongod, resourcesMongos, resourcesMongoCfg, resourcesMongoInfra *mongodb.Resources = getResources(d)
					var mongod *mongodb.MongodbSpec5_0_Mongod
					var mongos *mongodb.MongodbSpec5_0_Mongos
					var mongocfg *mongodb.MongodbSpec5_0_MongoCfg
					var mongoinfra *mongodb.MongodbSpec5_0_MongoInfra
					mongod = &mongodb.MongodbSpec5_0_Mongod{
						Config:    &configMongod,
						Resources: resourcesMongod,
					}

					if _, ok := hostTypes["MONGOS"]; ok {
						mongos = &mongodb.MongodbSpec5_0_Mongos{
							Config:    &configMongos,
							Resources: resourcesMongos,
						}
					}
					if _, ok := hostTypes["MONGOCFG"]; ok {
						mongocfg = &mongodb.MongodbSpec5_0_MongoCfg{
							Config:    &configMongoCfg,
							Resources: resourcesMongoCfg,
						}
					}
					if _, ok := hostTypes["MONGOINFRA"]; ok {
						mongoinfra = &mongodb.MongodbSpec5_0_MongoInfra{
							ConfigMongocfg: &configMongoCfg,
							ConfigMongos:   &configMongos,
							Resources:      resourcesMongoInfra,
						}
					}

					return &mongodb.ConfigSpec_MongodbSpec_5_0{
						MongodbSpec_5_0: &mongodb.MongodbSpec5_0{
							Mongod:     mongod,
							Mongos:     mongos,
							Mongocfg:   mongocfg,
							Mongoinfra: mongoinfra,
						},
					}
				},
			}
		}
	case "4.4-enterprise":
		{
			return &MongodbSpecHelper{
				FlattenResources: func(c *mongodb.ClusterConfig, d *schema.ResourceData) (map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4Enterprise).Mongodb_4_4Enterprise
					resources := map[string]interface{}{}
					if _, ok := d.GetOk("resources"); ok {
						if spec.Mongod != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongod.Resources)
							return resources, nil
						}
						if spec.Mongos != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongos.Resources)
							return resources, nil
						}
						if spec.Mongocfg != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongocfg.Resources)
							return resources, nil
						}
						if spec.Mongoinfra != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
							return resources, nil
						}
					} else {
						if spec.Mongod != nil {
							resources["resources_mongod"] = flattenMongoDBResources(spec.Mongod.Resources)
						}
						if spec.Mongos != nil {
							resources["resources_mongos"] = flattenMongoDBResources(spec.Mongos.Resources)
						}
						if spec.Mongocfg != nil {
							resources["resources_mongocfg"] = flattenMongoDBResources(spec.Mongocfg.Resources)
						}
						if spec.Mongoinfra != nil {
							resources["resources_mongoinfra"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
						}
					}
					if len(resources) == 0 {
						return nil, fmt.Errorf("Non empty service not found in mongo spec")
					}
					return resources, nil
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongod := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4Enterprise).Mongodb_4_4Enterprise.Mongod
					if mongod != nil {
						user_config := mongod.GetConfig().GetUserConfig()
						default_config := mongod.GetConfig().GetDefaultConfig()

						result := map[string]interface{}{}

						if security := user_config.GetSecurity(); security != nil {
							flattenSecurity := map[string]interface{}{}
							if enableEncription := security.GetEnableEncryption(); enableEncription != nil {
								flattenSecurity["enable_encryption"] = enableEncription.GetValue()
							}
							if kmip := security.GetKmip(); kmip != nil {
								flattenKmip := map[string]interface{}{}
								flattenKmip["server_name"] = kmip.GetServerName()
								flattenKmip["port"] = int(kmip.GetPort().GetValue())
								flattenKmip["server_ca"] = kmip.GetServerCa()
								flattenKmip["client_certificate"] = d.Get("cluster_config.0.mongod.0.security.0.kmip.0.client_certificate")
								flattenKmip["key_identifier"] = kmip.GetKeyIdentifier()

								flattenSecurity["kmip"] = []map[string]interface{}{flattenKmip}
							}
							result["security"] = []map[string]interface{}{flattenSecurity}
						}

						if audit_log := user_config.GetAuditLog(); audit_log != nil {
							audit_log_data := map[string]interface{}{}
							if audit_log.GetFilter() != default_config.GetAuditLog().GetFilter() {
								audit_log_data["filter"] = audit_log.GetFilter()
							}
							result["audit_log"] = []map[string]interface{}{audit_log_data}
						}
						if setParameter := user_config.GetSetParameter(); setParameter != nil {
							setParameterData := map[string]interface{}{}
							if setParameter.GetAuditAuthorizationSuccess() != nil {
								setParameterData["audit_authorization_success"] = setParameter.GetAuditAuthorizationSuccess().GetValue()
							}
							if flowControl := setParameter.GetEnableFlowControl(); flowControl != nil {
								setParameterData["enable_flow_control"] = flowControl.GetValue()
							}
							result["set_parameter"] = []map[string]interface{}{setParameterData}
						}

						if net := user_config.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongodConfig4_4Enterprise_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := user_config.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								if collectionConfig := wiredTiger.GetCollectionConfig(); collectionConfig != nil {
									if blockCompressor := collectionConfig.GetBlockCompressor(); blockCompressor != 0 {
										flattenWiredTiger["block_compressor"] = blockCompressor.String()
									}
								}
								if indexConfig := wiredTiger.GetIndexConfig(); indexConfig != nil {
									if prefixCompression := indexConfig.GetPrefixCompression(); prefixCompression != nil {
										flattenWiredTiger["prefix_compression"] = prefixCompression.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}

							if journal := storage.GetJournal(); journal != nil {
								flattenJournal := map[string]interface{}{}
								if commitInterval := journal.GetCommitInterval(); commitInterval != nil {
									flattenJournal["commit_interval"] = commitInterval.GetValue()
								}
								flattenStorage["journal"] = []map[string]interface{}{flattenJournal}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := user_config.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							if opSampleRate := opProfiling.GetSlowOpSampleRate(); opSampleRate != nil {
								flattenOpProfiling["slow_op_sample_rate"] = opSampleRate.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongos: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4Enterprise).Mongodb_4_4Enterprise
					userConfig := mongodbConfig.Mongos.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongos().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongosConfig4_4Enterprise_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}
						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongocfg: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4Enterprise).Mongodb_4_4Enterprise
					userConfig := mongodbConfig.Mongocfg.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongocfg().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := userConfig.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := userConfig.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					configMongod := mongo_config.MongodConfig4_4Enterprise{}
					configMongos := mongo_config.MongosConfig4_4Enterprise{}
					configMongoCfg := mongo_config.MongoCfgConfig4_4Enterprise{}

					security := mongo_config.MongodConfig4_4Enterprise_Security{}
					if enable_encryption := d.Get("cluster_config.0.mongod.0.security.0.enable_encryption"); enable_encryption != nil {
						security.SetEnableEncryption(&wrappers.BoolValue{Value: enable_encryption.(bool)})
					}
					kmip := mongo_config.MongodConfig4_4Enterprise_Security_KMIP{}
					if server_name := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.server_name"); server_name != nil {
						kmip.SetServerName(server_name.(string))
					}
					if port := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.port"); port != nil {
						kmip.SetPort(&wrappers.Int64Value{Value: int64(port.(int))})
					}
					if server_ca := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.server_ca"); server_ca != nil {
						kmip.SetServerCa(server_ca.(string))
					}
					if client_certificate := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.client_certificate"); client_certificate != nil {
						kmip.SetClientCertificate(client_certificate.(string))
					}
					if key_identifier := d.Get("cluster_config.0.mongod.0.security.0.kmip.0.key_identifier"); key_identifier != nil {
						kmip.SetKeyIdentifier(key_identifier.(string))
					}
					security.SetKmip(&kmip)
					configMongod.SetSecurity(&security)
					audit_log := mongo_config.MongodConfig4_4Enterprise_AuditLog{}
					if filter := d.Get("cluster_config.0.mongod.0.audit_log.0.filter"); filter != nil {
						audit_log.SetFilter(filter.(string))
					}
					// Note: right now runtime_configuration unsupported, so we should comment this statement
					//if rt := d.Get("cluster_config.0.mongod.0.audit_log.0.runtime_configuration"); rt != nil {
					//	audit_log.SetRuntimeConfiguration(&wrappers.BoolValue{Value: rt.(bool)})
					//}
					configMongod.SetAuditLog(&audit_log)

					set_paramenter := mongo_config.MongodConfig4_4Enterprise_SetParameter{}
					if success := d.Get("cluster_config.0.mongod.0.set_parameter.0.audit_authorization_success"); success != nil {
						set_paramenter.SetAuditAuthorizationSuccess(&wrappers.BoolValue{Value: success.(bool)})
					}
					if flowControl, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter.0.enable_flow_control"); ok {
						set_paramenter.SetEnableFlowControl(&wrappers.BoolValue{Value: flowControl.(bool)})
					}
					configMongod.SetSetParameter(&set_paramenter)
					if _, ok := d.GetOk("cluster_config.0.mongod.0.net"); ok {
						netMongod := mongo_config.MongodConfig4_4Enterprise_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongod.0.net.0.max_incoming_connections"); ok {
							netMongod.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongod.0.net.0.compressors"); ok {
							netCompressionMongod := mongo_config.MongodConfig4_4Enterprise_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongodConfig4_4Enterprise_Network_Compression_Compressor {
									compressorInt := mongo_config.MongodConfig4_4Enterprise_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongodConfig4_4Enterprise_Network_Compression_Compressor(compressorInt)
								})
							netCompressionMongod.SetCompressors(modifiedCompressors)
							netMongod.SetCompression(&netCompressionMongod)
						}
						configMongod.SetNet(&netMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling"); ok {
						opProfilingMongod := mongo_config.MongodConfig4_4Enterprise_OperationProfiling{}

						if mode, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongodConfig4_4Enterprise_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongod.SetMode(mongo_config.MongodConfig4_4Enterprise_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongod.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}

						if opSampleRate, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate"); ok {
							opProfilingMongod.SetSlowOpSampleRate(&wrappers.DoubleValue{Value: opSampleRate.(float64)})
						}
						configMongod.SetOperationProfiling(&opProfilingMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.storage"); ok {
						engineConfigMongod := mongo_config.MongodConfig4_4Enterprise_Storage_WiredTiger_EngineConfig{}
						collectionConfigMongod := mongo_config.MongodConfig4_4Enterprise_Storage_WiredTiger_CollectionConfig{}
						indexConfigMongod := mongo_config.MongodConfig4_4Enterprise_Storage_WiredTiger_IndexConfig{}
						journalMongod := mongo_config.MongodConfig4_4Enterprise_Storage_Journal{}
						wiredTigerMongod := mongo_config.MongodConfig4_4Enterprise_Storage_WiredTiger{
							EngineConfig:     &engineConfigMongod,
							CollectionConfig: &collectionConfigMongod,
							IndexConfig:      &indexConfigMongod,
						}
						storageMongod := mongo_config.MongodConfig4_4Enterprise_Storage{
							WiredTiger: &wiredTigerMongod,
							Journal:    &journalMongod,
						}
						if cacheSize, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongod.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						if blockCompressor, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"); ok {
							blockCompressorInt := mongo_config.MongodConfig4_4Enterprise_Storage_WiredTiger_CollectionConfig_Compressor_value[strings.ToUpper(blockCompressor.(string))]
							collectionConfigMongod.SetBlockCompressor(
								mongo_config.MongodConfig4_4Enterprise_Storage_WiredTiger_CollectionConfig_Compressor(blockCompressorInt),
							)
						}
						if prefixCompression, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression"); ok {
							indexConfigMongod.SetPrefixCompression(&wrappers.BoolValue{Value: prefixCompression.(bool)})
						}

						if commitInterval, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"); ok {
							journalMongod.SetCommitInterval(&wrappers.Int64Value{Value: int64(commitInterval.(int))})
						}
						configMongod.SetStorage(&storageMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongos.0.net"); ok {
						netMongos := mongo_config.MongosConfig4_4Enterprise_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongos.0.net.0.max_incoming_connections"); ok {
							netMongos.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongos.0.net.0.compressors"); ok {
							netCompressionMongos := mongo_config.MongosConfig4_4Enterprise_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongosConfig4_4Enterprise_Network_Compression_Compressor {
									compressorInt := mongo_config.MongosConfig4_4Enterprise_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongosConfig4_4Enterprise_Network_Compression_Compressor(compressorInt)
								})
							netCompressionMongos.SetCompressors(modifiedCompressors)
							netMongos.SetCompression(&netCompressionMongos)
						}
						configMongos.SetNet(&netMongos)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.net"); ok {
						netMongoCfg := mongo_config.MongoCfgConfig4_4Enterprise_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongocfg.0.net.0.max_incoming_connections"); ok {
							netMongoCfg.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongoCfg.SetNet(&netMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling"); ok {
						opProfilingMongoCfg := mongo_config.MongoCfgConfig4_4Enterprise_OperationProfiling{}
						if mode, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongoCfgConfig4_4Enterprise_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongoCfg.SetMode(mongo_config.MongoCfgConfig4_4Enterprise_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongoCfg.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						configMongoCfg.SetOperationProfiling(&opProfilingMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.storage"); ok {
						engineConfigMongoCfg := mongo_config.MongoCfgConfig4_4Enterprise_Storage_WiredTiger_EngineConfig{}
						wiredTigerMongoCfg := mongo_config.MongoCfgConfig4_4Enterprise_Storage_WiredTiger{EngineConfig: &engineConfigMongoCfg}
						storageMongoCfg := mongo_config.MongoCfgConfig4_4Enterprise_Storage{WiredTiger: &wiredTigerMongoCfg}

						if cacheSize, ok := d.GetOk("cluster_config.0.mongocfg.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongoCfg.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						configMongoCfg.SetStorage(&storageMongoCfg)
					}
					hostTypes := getSetOfHostTypes(d)
					var resourcesMongod, resourcesMongos, resourcesMongoCfg, resourcesMongoInfra *mongodb.Resources = getResources(d)
					var mongod *mongodb.MongodbSpec4_4Enterprise_Mongod
					var mongos *mongodb.MongodbSpec4_4Enterprise_Mongos
					var mongocfg *mongodb.MongodbSpec4_4Enterprise_MongoCfg
					var mongoinfra *mongodb.MongodbSpec4_4Enterprise_MongoInfra
					mongod = &mongodb.MongodbSpec4_4Enterprise_Mongod{
						Config:    &configMongod,
						Resources: resourcesMongod,
					}

					if _, ok := hostTypes["MONGOS"]; ok {
						mongos = &mongodb.MongodbSpec4_4Enterprise_Mongos{
							Config:    &configMongos,
							Resources: resourcesMongos,
						}
					}
					if _, ok := hostTypes["MONGOCFG"]; ok {
						mongocfg = &mongodb.MongodbSpec4_4Enterprise_MongoCfg{
							Config:    &configMongoCfg,
							Resources: resourcesMongoCfg,
						}
					}
					if _, ok := hostTypes["MONGOINFRA"]; ok {
						mongoinfra = &mongodb.MongodbSpec4_4Enterprise_MongoInfra{
							ConfigMongocfg: &configMongoCfg,
							ConfigMongos:   &configMongos,
							Resources:      resourcesMongoInfra,
						}
					}
					return &mongodb.ConfigSpec_MongodbSpec_4_4Enterprise{
						MongodbSpec_4_4Enterprise: &mongodb.MongodbSpec4_4Enterprise{
							Mongod:     mongod,
							Mongos:     mongos,
							Mongocfg:   mongocfg,
							Mongoinfra: mongoinfra,
						},
					}
				},
			}
		}
	case "4.4":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig, d *schema.ResourceData) (map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4).Mongodb_4_4
					resources := map[string]interface{}{}
					if _, ok := d.GetOk("resources"); ok {
						if spec.Mongod != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongod.Resources)
							return resources, nil
						}
						if spec.Mongos != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongos.Resources)
							return resources, nil
						}
						if spec.Mongocfg != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongocfg.Resources)
							return resources, nil
						}
						if spec.Mongoinfra != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
							return resources, nil
						}
					} else {
						if spec.Mongod != nil {
							resources["resources_mongod"] = flattenMongoDBResources(spec.Mongod.Resources)
						}
						if spec.Mongos != nil {
							resources["resources_mongos"] = flattenMongoDBResources(spec.Mongos.Resources)
						}
						if spec.Mongocfg != nil {
							resources["resources_mongocfg"] = flattenMongoDBResources(spec.Mongocfg.Resources)
						}
						if spec.Mongoinfra != nil {
							resources["resources_mongoinfra"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
						}
					}
					if len(resources) == 0 {
						return nil, fmt.Errorf("Non empty service not found in mongo spec")
					}
					return resources, nil
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongod := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4).Mongodb_4_4.Mongod
					if mongod != nil {
						user_config := mongod.GetConfig().GetUserConfig()
						result := map[string]interface{}{}

						if setParameter := user_config.GetSetParameter(); setParameter != nil {
							flattenSetParameter := map[string]interface{}{}
							if flowControl := setParameter.GetEnableFlowControl(); flowControl != nil {
								flattenSetParameter["enable_flow_control"] = flowControl.GetValue()
							}
							result["set_parameter"] = []map[string]interface{}{flattenSetParameter}
						}

						if net := user_config.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongodConfig4_4_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := user_config.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								if collectionConfig := wiredTiger.GetCollectionConfig(); collectionConfig != nil {
									if blockCompressor := collectionConfig.GetBlockCompressor(); blockCompressor != 0 {
										flattenWiredTiger["block_compressor"] = blockCompressor.String()
									}
								}
								if indexConfig := wiredTiger.GetIndexConfig(); indexConfig != nil {
									if prefixCompression := indexConfig.GetPrefixCompression(); prefixCompression != nil {
										flattenWiredTiger["prefix_compression"] = prefixCompression.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}

							if journal := storage.GetJournal(); journal != nil {
								flattenJournal := map[string]interface{}{}
								if commitInterval := journal.GetCommitInterval(); commitInterval != nil {
									flattenJournal["commit_interval"] = commitInterval.GetValue()
								}
								flattenStorage["journal"] = []map[string]interface{}{flattenJournal}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := user_config.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							if opSampleRate := opProfiling.GetSlowOpSampleRate(); opSampleRate != nil {
								flattenOpProfiling["slow_op_sample_rate"] = opSampleRate.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongos: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4).Mongodb_4_4
					userConfig := mongodbConfig.Mongos.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongos().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongosConfig4_4_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}
						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongocfg: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4).Mongodb_4_4
					userConfig := mongodbConfig.Mongocfg.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongocfg().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := userConfig.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := userConfig.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					configMongod := mongo_config.MongodConfig4_4{}
					configMongos := mongo_config.MongosConfig4_4{}
					configMongoCfg := mongo_config.MongoCfgConfig4_4{}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter"); ok {
						setParameter := mongo_config.MongodConfig4_4_SetParameter{}
						if flowControl, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter.0.enable_flow_control"); ok {
							setParameter.SetEnableFlowControl(&wrappers.BoolValue{Value: flowControl.(bool)})
						}
						configMongod.SetSetParameter(&setParameter)
					}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.net"); ok {
						netMongod := mongo_config.MongodConfig4_4_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongod.0.net.0.max_incoming_connections"); ok {
							netMongod.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongod.0.net.0.compressors"); ok {
							netCompressionMongod := mongo_config.MongodConfig4_4_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongodConfig4_4_Network_Compression_Compressor {
									compressorInt := mongo_config.MongodConfig4_4_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongodConfig4_4_Network_Compression_Compressor(compressorInt)
								})
							netCompressionMongod.SetCompressors(modifiedCompressors)
							netMongod.SetCompression(&netCompressionMongod)
						}
						configMongod.SetNet(&netMongod)
					}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling"); ok {
						opProfilingMongod := mongo_config.MongodConfig4_4_OperationProfiling{}

						if mode, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongodConfig4_4_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongod.SetMode(mongo_config.MongodConfig4_4_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongod.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						if opSampleRate, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate"); ok {
							opProfilingMongod.SetSlowOpSampleRate(&wrappers.DoubleValue{Value: opSampleRate.(float64)})
						}
						configMongod.SetOperationProfiling(&opProfilingMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.storage"); ok {
						engineConfigMongod := mongo_config.MongodConfig4_4_Storage_WiredTiger_EngineConfig{}
						collectionConfigMongod := mongo_config.MongodConfig4_4_Storage_WiredTiger_CollectionConfig{}
						indexConfigMongod := mongo_config.MongodConfig4_4_Storage_WiredTiger_IndexConfig{}
						journalMongod := mongo_config.MongodConfig4_4_Storage_Journal{}
						wiredTigerMongod := mongo_config.MongodConfig4_4_Storage_WiredTiger{
							EngineConfig:     &engineConfigMongod,
							CollectionConfig: &collectionConfigMongod,
							IndexConfig:      &indexConfigMongod,
						}
						storageMongod := mongo_config.MongodConfig4_4_Storage{
							WiredTiger: &wiredTigerMongod,
							Journal:    &journalMongod,
						}
						if cacheSize, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongod.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						if blockCompressor, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"); ok {
							blockCompressorInt := mongo_config.MongodConfig4_4_Storage_WiredTiger_CollectionConfig_Compressor_value[strings.ToUpper(blockCompressor.(string))]
							collectionConfigMongod.SetBlockCompressor(
								mongo_config.MongodConfig4_4_Storage_WiredTiger_CollectionConfig_Compressor(blockCompressorInt),
							)
						}
						if prefixCompression, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression"); ok {
							indexConfigMongod.SetPrefixCompression(&wrappers.BoolValue{Value: prefixCompression.(bool)})
						}

						if commitInterval, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"); ok {
							journalMongod.SetCommitInterval(&wrappers.Int64Value{Value: int64(commitInterval.(int))})
						}
						configMongod.SetStorage(&storageMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongos.0.net"); ok {
						netMongos := mongo_config.MongosConfig4_4_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongos.0.net.0.max_incoming_connections"); ok {
							netMongos.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongos.0.net.0.compressors"); ok {
							netCompressionMongos := mongo_config.MongosConfig4_4_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongosConfig4_4_Network_Compression_Compressor {
									compressorInt := mongo_config.MongosConfig4_4_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongosConfig4_4_Network_Compression_Compressor(compressorInt)
								})
							netCompressionMongos.SetCompressors(modifiedCompressors)
							netMongos.SetCompression(&netCompressionMongos)
						}
						configMongos.SetNet(&netMongos)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.net"); ok {
						netMongoCfg := mongo_config.MongoCfgConfig4_4_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongocfg.0.net.0.max_incoming_connections"); ok {
							netMongoCfg.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongoCfg.SetNet(&netMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling"); ok {
						opProfilingMongoCfg := mongo_config.MongoCfgConfig4_4_OperationProfiling{}
						if mode, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongoCfgConfig4_4_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongoCfg.SetMode(mongo_config.MongoCfgConfig4_4_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongoCfg.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						configMongoCfg.SetOperationProfiling(&opProfilingMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.storage"); ok {
						engineConfigMongoCfg := mongo_config.MongoCfgConfig4_4_Storage_WiredTiger_EngineConfig{}
						wiredTigerMongoCfg := mongo_config.MongoCfgConfig4_4_Storage_WiredTiger{EngineConfig: &engineConfigMongoCfg}
						storageMongoCfg := mongo_config.MongoCfgConfig4_4_Storage{WiredTiger: &wiredTigerMongoCfg}

						if cacheSize, ok := d.GetOk("cluster_config.0.mongocfg.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongoCfg.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						configMongoCfg.SetStorage(&storageMongoCfg)
					}
					hostTypes := getSetOfHostTypes(d)
					var resourcesMongod, resourcesMongos, resourcesMongoCfg, resourcesMongoInfra *mongodb.Resources = getResources(d)
					var mongod *mongodb.MongodbSpec4_4_Mongod
					var mongos *mongodb.MongodbSpec4_4_Mongos
					var mongocfg *mongodb.MongodbSpec4_4_MongoCfg
					var mongoinfra *mongodb.MongodbSpec4_4_MongoInfra
					mongod = &mongodb.MongodbSpec4_4_Mongod{
						Config:    &configMongod,
						Resources: resourcesMongod,
					}

					if _, ok := hostTypes["MONGOS"]; ok {
						mongos = &mongodb.MongodbSpec4_4_Mongos{
							Config:    &configMongos,
							Resources: resourcesMongos,
						}
					}
					if _, ok := hostTypes["MONGOCFG"]; ok {
						mongocfg = &mongodb.MongodbSpec4_4_MongoCfg{
							Config:    &configMongoCfg,
							Resources: resourcesMongoCfg,
						}
					}
					if _, ok := hostTypes["MONGOINFRA"]; ok {
						mongoinfra = &mongodb.MongodbSpec4_4_MongoInfra{
							ConfigMongocfg: &configMongoCfg,
							ConfigMongos:   &configMongos,
							Resources:      resourcesMongoInfra,
						}
					}
					return &mongodb.ConfigSpec_MongodbSpec_4_4{
						MongodbSpec_4_4: &mongodb.MongodbSpec4_4{
							Mongod:     mongod,
							Mongos:     mongos,
							Mongocfg:   mongocfg,
							Mongoinfra: mongoinfra,
						},
					}
				},
			}
		}
	case "4.2":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig, d *schema.ResourceData) (map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_2).Mongodb_4_2
					resources := map[string]interface{}{}
					if _, ok := d.GetOk("resources"); ok {
						if spec.Mongod != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongod.Resources)
							return resources, nil
						}
						if spec.Mongos != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongos.Resources)
							return resources, nil
						}
						if spec.Mongocfg != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongocfg.Resources)
							return resources, nil
						}
						if spec.Mongoinfra != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
							return resources, nil
						}
					} else {
						if spec.Mongod != nil {
							resources["resources_mongod"] = flattenMongoDBResources(spec.Mongod.Resources)
						}
						if spec.Mongos != nil {
							resources["resources_mongos"] = flattenMongoDBResources(spec.Mongos.Resources)
						}
						if spec.Mongocfg != nil {
							resources["resources_mongocfg"] = flattenMongoDBResources(spec.Mongocfg.Resources)
						}
						if spec.Mongoinfra != nil {
							resources["resources_mongoinfra"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
						}
					}
					if len(resources) == 0 {
						return nil, fmt.Errorf("Non empty service not found in mongo spec")
					}
					return resources, nil
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongod := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_2).Mongodb_4_2.Mongod
					if mongod != nil {
						user_config := mongod.GetConfig().GetUserConfig()
						result := map[string]interface{}{}

						if setParameter := user_config.GetSetParameter(); setParameter != nil {
							flattenSetParameter := map[string]interface{}{}
							if flowControl := setParameter.GetEnableFlowControl(); flowControl != nil {
								flattenSetParameter["enable_flow_control"] = flowControl.GetValue()
							}
							result["set_parameter"] = []map[string]interface{}{flattenSetParameter}
						}

						if net := user_config.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongodConfig4_2_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := user_config.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								if collectionConfig := wiredTiger.GetCollectionConfig(); collectionConfig != nil {
									if blockCompressor := collectionConfig.GetBlockCompressor(); blockCompressor != 0 {
										flattenWiredTiger["block_compressor"] = blockCompressor.String()
									}
								}
								if indexConfig := wiredTiger.GetIndexConfig(); indexConfig != nil {
									if prefixCompression := indexConfig.GetPrefixCompression(); prefixCompression != nil {
										flattenWiredTiger["prefix_compression"] = prefixCompression.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}

							if journal := storage.GetJournal(); journal != nil {
								flattenJournal := map[string]interface{}{}
								if commitInterval := journal.GetCommitInterval(); commitInterval != nil {
									flattenJournal["commit_interval"] = commitInterval.GetValue()
								}
								flattenStorage["journal"] = []map[string]interface{}{flattenJournal}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := user_config.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							if opSampleRate := opProfiling.GetSlowOpSampleRate(); opSampleRate != nil {
								flattenOpProfiling["slow_op_sample_rate"] = opSampleRate.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongos: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_2).Mongodb_4_2
					userConfig := mongodbConfig.Mongos.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongos().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							if compression := net.GetCompression(); compression != nil {
								if compressors := compression.GetCompressors(); compressors != nil {
									flattenNet["compressors"] = Map(compressors,
										func(f mongo_config.MongosConfig4_2_Network_Compression_Compressor) string {
											return f.String()
										})
								}
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}
						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongocfg: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongodbConfig := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_2).Mongodb_4_2
					userConfig := mongodbConfig.Mongocfg.GetConfig().GetUserConfig()
					if userConfig == nil {
						userConfig = mongodbConfig.Mongoinfra.GetConfigMongocfg().GetUserConfig()
					}
					if userConfig != nil {
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := userConfig.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := userConfig.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					configMongod := mongo_config.MongodConfig4_2{}
					configMongos := mongo_config.MongosConfig4_2{}
					configMongoCfg := mongo_config.MongoCfgConfig4_2{}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter"); ok {
						setParameter := mongo_config.MongodConfig4_2_SetParameter{}
						if flowControl, ok := d.GetOk("cluster_config.0.mongod.0.set_parameter.0.enable_flow_control"); ok {
							setParameter.SetEnableFlowControl(&wrappers.BoolValue{Value: flowControl.(bool)})
						}
						configMongod.SetSetParameter(&setParameter)
					}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.net"); ok {
						netMongod := mongo_config.MongodConfig4_2_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongod.0.net.0.max_incoming_connections"); ok {
							netMongod.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongod.0.net.0.compressors"); ok {
							netCompressionMongod := mongo_config.MongodConfig4_2_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongodConfig4_2_Network_Compression_Compressor {
									compressorInt := mongo_config.MongodConfig4_2_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongodConfig4_2_Network_Compression_Compressor(compressorInt)
								})
							netCompressionMongod.SetCompressors(modifiedCompressors)
							netMongod.SetCompression(&netCompressionMongod)
						}
						configMongod.SetNet(&netMongod)
					}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling"); ok {
						opProfilingMongod := mongo_config.MongodConfig4_2_OperationProfiling{}

						if mode, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongodConfig4_2_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongod.SetMode(mongo_config.MongodConfig4_2_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongod.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						if opSampleRate, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_sample_rate"); ok {
							opProfilingMongod.SetSlowOpSampleRate(&wrappers.DoubleValue{Value: opSampleRate.(float64)})
						}
						configMongod.SetOperationProfiling(&opProfilingMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.storage"); ok {
						engineConfigMongod := mongo_config.MongodConfig4_2_Storage_WiredTiger_EngineConfig{}
						collectionConfigMongod := mongo_config.MongodConfig4_2_Storage_WiredTiger_CollectionConfig{}
						indexConfigMongod := mongo_config.MongodConfig4_2_Storage_WiredTiger_IndexConfig{}
						journalMongod := mongo_config.MongodConfig4_2_Storage_Journal{}
						wiredTigerMongod := mongo_config.MongodConfig4_2_Storage_WiredTiger{
							EngineConfig:     &engineConfigMongod,
							CollectionConfig: &collectionConfigMongod,
							IndexConfig:      &indexConfigMongod,
						}
						storageMongod := mongo_config.MongodConfig4_2_Storage{
							WiredTiger: &wiredTigerMongod,
							Journal:    &journalMongod,
						}
						if cacheSize, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongod.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						if blockCompressor, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"); ok {
							blockCompressorInt := mongo_config.MongodConfig4_2_Storage_WiredTiger_CollectionConfig_Compressor_value[strings.ToUpper(blockCompressor.(string))]
							collectionConfigMongod.SetBlockCompressor(
								mongo_config.MongodConfig4_2_Storage_WiredTiger_CollectionConfig_Compressor(blockCompressorInt),
							)
						}
						if prefixCompression, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.prefix_compression"); ok {
							indexConfigMongod.SetPrefixCompression(&wrappers.BoolValue{Value: prefixCompression.(bool)})
						}

						if commitInterval, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"); ok {
							journalMongod.SetCommitInterval(&wrappers.Int64Value{Value: int64(commitInterval.(int))})
						}
						configMongod.SetStorage(&storageMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongos.0.net"); ok {
						netMongos := mongo_config.MongosConfig4_2_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongos.0.net.0.max_incoming_connections"); ok {
							netMongos.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						if compressors, ok := d.GetOk("cluster_config.0.mongos.0.net.0.compressors"); ok {
							netCompressionMongos := mongo_config.MongosConfig4_2_Network_Compression{}
							modifiedCompressors := Map(compressors.([]interface{}),
								func(f interface{}) mongo_config.MongosConfig4_2_Network_Compression_Compressor {
									compressorInt := mongo_config.MongosConfig4_2_Network_Compression_Compressor_value[strings.ToUpper(f.(string))]
									return mongo_config.MongosConfig4_2_Network_Compression_Compressor(compressorInt)
								})
							netCompressionMongos.SetCompressors(modifiedCompressors)
							netMongos.SetCompression(&netCompressionMongos)
						}
						configMongos.SetNet(&netMongos)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.net"); ok {
						netMongoCfg := mongo_config.MongoCfgConfig4_2_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongocfg.0.net.0.max_incoming_connections"); ok {
							netMongoCfg.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongoCfg.SetNet(&netMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling"); ok {
						opProfilingMongoCfg := mongo_config.MongoCfgConfig4_2_OperationProfiling{}
						if mode, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongoCfgConfig4_2_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongoCfg.SetMode(mongo_config.MongoCfgConfig4_2_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongoCfg.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						configMongoCfg.SetOperationProfiling(&opProfilingMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.storage"); ok {
						engineConfigMongoCfg := mongo_config.MongoCfgConfig4_2_Storage_WiredTiger_EngineConfig{}
						wiredTigerMongoCfg := mongo_config.MongoCfgConfig4_2_Storage_WiredTiger{EngineConfig: &engineConfigMongoCfg}
						storageMongoCfg := mongo_config.MongoCfgConfig4_2_Storage{WiredTiger: &wiredTigerMongoCfg}

						if cacheSize, ok := d.GetOk("cluster_config.0.mongocfg.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongoCfg.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						configMongoCfg.SetStorage(&storageMongoCfg)
					}
					hostTypes := getSetOfHostTypes(d)
					var resourcesMongod, resourcesMongos, resourcesMongoCfg, resourcesMongoInfra *mongodb.Resources = getResources(d)
					var mongod *mongodb.MongodbSpec4_2_Mongod
					var mongos *mongodb.MongodbSpec4_2_Mongos
					var mongocfg *mongodb.MongodbSpec4_2_MongoCfg
					var mongoinfra *mongodb.MongodbSpec4_2_MongoInfra
					mongod = &mongodb.MongodbSpec4_2_Mongod{
						Config:    &configMongod,
						Resources: resourcesMongod,
					}

					if _, ok := hostTypes["MONGOS"]; ok {
						mongos = &mongodb.MongodbSpec4_2_Mongos{
							Config:    &configMongos,
							Resources: resourcesMongos,
						}
					}
					if _, ok := hostTypes["MONGOCFG"]; ok {
						mongocfg = &mongodb.MongodbSpec4_2_MongoCfg{
							Config:    &configMongoCfg,
							Resources: resourcesMongoCfg,
						}
					}
					if _, ok := hostTypes["MONGOINFRA"]; ok {
						mongoinfra = &mongodb.MongodbSpec4_2_MongoInfra{
							ConfigMongocfg: &configMongoCfg,
							ConfigMongos:   &configMongos,
							Resources:      resourcesMongoInfra,
						}
					}
					return &mongodb.ConfigSpec_MongodbSpec_4_2{
						MongodbSpec_4_2: &mongodb.MongodbSpec4_2{
							Mongod:     mongod,
							Mongos:     mongos,
							Mongocfg:   mongocfg,
							Mongoinfra: mongoinfra,
						},
					}
				},
			}
		}
	case "4.0":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig, d *schema.ResourceData) (map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_0).Mongodb_4_0
					resources := map[string]interface{}{}
					if _, ok := d.GetOk("resources"); ok {
						if spec.Mongod != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongod.Resources)
							return resources, nil
						}
						if spec.Mongos != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongos.Resources)
							return resources, nil
						}
						if spec.Mongocfg != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongocfg.Resources)
							return resources, nil
						}
						if spec.Mongoinfra != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
							return resources, nil
						}
					} else {
						if spec.Mongod != nil {
							resources["resources_mongod"] = flattenMongoDBResources(spec.Mongod.Resources)
						}
						if spec.Mongos != nil {
							resources["resources_mongos"] = flattenMongoDBResources(spec.Mongos.Resources)
						}
						if spec.Mongocfg != nil {
							resources["resources_mongocfg"] = flattenMongoDBResources(spec.Mongocfg.Resources)
						}
						if spec.Mongoinfra != nil {
							resources["resources_mongoinfra"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
						}
					}
					if len(resources) == 0 {
						return nil, fmt.Errorf("Non empty service not found in mongo spec")
					}
					return resources, nil
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongod := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_0).Mongodb_4_0.Mongod
					if mongod != nil {
						user_config := mongod.GetConfig().GetUserConfig()
						result := map[string]interface{}{}

						if net := user_config.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := user_config.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								if collectionConfig := wiredTiger.GetCollectionConfig(); collectionConfig != nil {
									if blockCompressor := collectionConfig.GetBlockCompressor(); blockCompressor != 0 {
										flattenWiredTiger["block_compressor"] = blockCompressor.String()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}

							if journal := storage.GetJournal(); journal != nil {
								flattenJournal := map[string]interface{}{}
								if commitInterval := journal.GetCommitInterval(); commitInterval != nil {
									flattenJournal["commit_interval"] = commitInterval.GetValue()
								}
								flattenStorage["journal"] = []map[string]interface{}{flattenJournal}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := user_config.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongos: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongos := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_0).Mongodb_4_0.Mongos
					if mongos != nil {
						userConfig := mongos.GetConfig().GetUserConfig()
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}
						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongocfg: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongocfg := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_0).Mongodb_4_0.Mongocfg
					if mongocfg != nil {
						userConfig := mongocfg.GetConfig().GetUserConfig()
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := userConfig.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := userConfig.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					configMongod := mongo_config.MongodConfig4_0{}
					configMongos := mongo_config.MongosConfig4_0{}
					configMongoCfg := mongo_config.MongoCfgConfig4_0{}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.net"); ok {
						netMongod := mongo_config.MongodConfig4_0_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongod.0.net.0.max_incoming_connections"); ok {
							netMongod.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongod.SetNet(&netMongod)
					}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling"); ok {
						opProfilingMongod := mongo_config.MongodConfig4_0_OperationProfiling{}

						if mode, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongodConfig4_0_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongod.SetMode(mongo_config.MongodConfig4_0_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongod.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						configMongod.SetOperationProfiling(&opProfilingMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.storage"); ok {
						engineConfigMongod := mongo_config.MongodConfig4_0_Storage_WiredTiger_EngineConfig{}
						collectionConfigMongod := mongo_config.MongodConfig4_0_Storage_WiredTiger_CollectionConfig{}
						journalMongod := mongo_config.MongodConfig4_0_Storage_Journal{}
						wiredTigerMongod := mongo_config.MongodConfig4_0_Storage_WiredTiger{
							EngineConfig:     &engineConfigMongod,
							CollectionConfig: &collectionConfigMongod,
						}
						storageMongod := mongo_config.MongodConfig4_0_Storage{
							WiredTiger: &wiredTigerMongod,
							Journal:    &journalMongod,
						}
						if cacheSize, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongod.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						if blockCompressor, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"); ok {
							blockCompressorInt := mongo_config.MongodConfig4_0_Storage_WiredTiger_CollectionConfig_Compressor_value[strings.ToUpper(blockCompressor.(string))]
							collectionConfigMongod.SetBlockCompressor(
								mongo_config.MongodConfig4_0_Storage_WiredTiger_CollectionConfig_Compressor(blockCompressorInt),
							)
						}

						if commitInterval, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"); ok {
							journalMongod.SetCommitInterval(&wrappers.Int64Value{Value: int64(commitInterval.(int))})
						}
						configMongod.SetStorage(&storageMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongos.0.net"); ok {

						netMongos := mongo_config.MongosConfig4_0_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongos.0.net.0.max_incoming_connections"); ok {
							netMongos.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongos.SetNet(&netMongos)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.net"); ok {
						netMongoCfg := mongo_config.MongoCfgConfig4_0_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongocfg.0.net.0.max_incoming_connections"); ok {
							netMongoCfg.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongoCfg.SetNet(&netMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling"); ok {
						opProfilingMongoCfg := mongo_config.MongoCfgConfig4_0_OperationProfiling{}
						if mode, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongoCfgConfig4_0_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongoCfg.SetMode(mongo_config.MongoCfgConfig4_0_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongoCfg.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						configMongoCfg.SetOperationProfiling(&opProfilingMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.storage"); ok {
						engineConfigMongoCfg := mongo_config.MongoCfgConfig4_0_Storage_WiredTiger_EngineConfig{}
						wiredTigerMongoCfg := mongo_config.MongoCfgConfig4_0_Storage_WiredTiger{EngineConfig: &engineConfigMongoCfg}
						storageMongoCfg := mongo_config.MongoCfgConfig4_0_Storage{WiredTiger: &wiredTigerMongoCfg}

						if cacheSize, ok := d.GetOk("cluster_config.0.mongocfg.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongoCfg.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						configMongoCfg.SetStorage(&storageMongoCfg)
					}
					hostTypes := getSetOfHostTypes(d)
					var resourcesMongod, resourcesMongos, resourcesMongoCfg, resourcesMongoInfra *mongodb.Resources = getResources(d)
					var mongod *mongodb.MongodbSpec4_0_Mongod
					var mongos *mongodb.MongodbSpec4_0_Mongos
					var mongocfg *mongodb.MongodbSpec4_0_MongoCfg
					var mongoinfra *mongodb.MongodbSpec4_0_MongoInfra
					mongod = &mongodb.MongodbSpec4_0_Mongod{
						Config:    &configMongod,
						Resources: resourcesMongod,
					}

					if _, ok := hostTypes["MONGOS"]; ok {
						mongos = &mongodb.MongodbSpec4_0_Mongos{
							Config:    &configMongos,
							Resources: resourcesMongos,
						}
					}
					if _, ok := hostTypes["MONGOCFG"]; ok {
						mongocfg = &mongodb.MongodbSpec4_0_MongoCfg{
							Config:    &configMongoCfg,
							Resources: resourcesMongoCfg,
						}
					}
					if _, ok := hostTypes["MONGOINFRA"]; ok {
						mongoinfra = &mongodb.MongodbSpec4_0_MongoInfra{
							ConfigMongocfg: &configMongoCfg,
							ConfigMongos:   &configMongos,
							Resources:      resourcesMongoInfra,
						}
					}
					return &mongodb.ConfigSpec_MongodbSpec_4_0{
						MongodbSpec_4_0: &mongodb.MongodbSpec4_0{
							Mongod:     mongod,
							Mongos:     mongos,
							Mongocfg:   mongocfg,
							Mongoinfra: mongoinfra,
						},
					}
				},
			}
		}
	case "3.6":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig, d *schema.ResourceData) (map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_3_6).Mongodb_3_6
					resources := map[string]interface{}{}
					if _, ok := d.GetOk("resources"); ok {
						if spec.Mongod != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongod.Resources)
							return resources, nil
						}
						if spec.Mongos != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongos.Resources)
							return resources, nil
						}
						if spec.Mongocfg != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongocfg.Resources)
							return resources, nil
						}
						if spec.Mongoinfra != nil {
							resources["resources"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
							return resources, nil
						}
					} else {
						if spec.Mongod != nil {
							resources["resources_mongod"] = flattenMongoDBResources(spec.Mongod.Resources)
						}
						if spec.Mongos != nil {
							resources["resources_mongos"] = flattenMongoDBResources(spec.Mongos.Resources)
						}
						if spec.Mongocfg != nil {
							resources["resources_mongocfg"] = flattenMongoDBResources(spec.Mongocfg.Resources)
						}
						if spec.Mongoinfra != nil {
							resources["resources_mongoinfra"] = flattenMongoDBResources(spec.Mongoinfra.Resources)
						}
					}
					if len(resources) == 0 {
						return nil, fmt.Errorf("Non empty service not found in mongo spec")
					}
					return resources, nil
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongod := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_3_6).Mongodb_3_6.Mongod
					if mongod != nil {
						user_config := mongod.GetConfig().GetUserConfig()
						result := map[string]interface{}{}

						if net := user_config.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := user_config.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								if collectionConfig := wiredTiger.GetCollectionConfig(); collectionConfig != nil {
									if blockCompressor := collectionConfig.GetBlockCompressor(); blockCompressor != 0 {
										flattenWiredTiger["block_compressor"] = blockCompressor.String()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}

							if journal := storage.GetJournal(); journal != nil {
								flattenJournal := map[string]interface{}{}
								if commitInterval := journal.GetCommitInterval(); commitInterval != nil {
									flattenJournal["commit_interval"] = commitInterval.GetValue()
								}
								flattenStorage["journal"] = []map[string]interface{}{flattenJournal}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := user_config.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongos: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongos := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_3_6).Mongodb_3_6.Mongos
					if mongos != nil {
						userConfig := mongos.GetConfig().GetUserConfig()
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}
						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				FlattenMongocfg: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongocfg := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_3_6).Mongodb_3_6.Mongocfg
					if mongocfg != nil {
						userConfig := mongocfg.GetConfig().GetUserConfig()
						result := map[string]interface{}{}

						if net := userConfig.GetNet(); net != nil {
							flattenNet := map[string]interface{}{}
							if maxIncomingConnections := net.GetMaxIncomingConnections(); maxIncomingConnections != nil {
								flattenNet["max_incoming_connections"] = maxIncomingConnections.GetValue()
							}
							result["net"] = []map[string]interface{}{flattenNet}
						}

						if storage := userConfig.GetStorage(); storage != nil {
							flattenStorage := map[string]interface{}{}
							if wiredTiger := storage.GetWiredTiger(); wiredTiger != nil {
								flattenWiredTiger := map[string]interface{}{}
								if engineConfig := wiredTiger.GetEngineConfig(); engineConfig != nil {
									if cacheSize := engineConfig.GetCacheSizeGb(); cacheSize != nil {
										flattenWiredTiger["cache_size_gb"] = cacheSize.GetValue()
									}
								}
								flattenStorage["wired_tiger"] = []map[string]interface{}{flattenWiredTiger}
							}
							result["storage"] = []map[string]interface{}{flattenStorage}
						}

						if opProfiling := userConfig.GetOperationProfiling(); opProfiling != nil {
							flattenOpProfiling := map[string]interface{}{}
							if mode := opProfiling.GetMode(); mode != 0 {
								flattenOpProfiling["mode"] = mode.String()
							}
							if opThreshold := opProfiling.GetSlowOpThreshold(); opThreshold != nil {
								flattenOpProfiling["slow_op_threshold"] = opThreshold.GetValue()
							}
							result["operation_profiling"] = []map[string]interface{}{flattenOpProfiling}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					configMongod := mongo_config.MongodConfig3_6{}
					configMongos := mongo_config.MongosConfig3_6{}
					configMongoCfg := mongo_config.MongoCfgConfig3_6{}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.net"); ok {
						netMongod := mongo_config.MongodConfig3_6_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongod.0.net.0.max_incoming_connections"); ok {
							netMongod.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongod.SetNet(&netMongod)
					}

					if _, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling"); ok {
						opProfilingMongod := mongo_config.MongodConfig3_6_OperationProfiling{}

						if mode, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongodConfig3_6_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongod.SetMode(mongo_config.MongodConfig3_6_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongod.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongod.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						configMongod.SetOperationProfiling(&opProfilingMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongod.0.storage"); ok {
						engineConfigMongod := mongo_config.MongodConfig3_6_Storage_WiredTiger_EngineConfig{}
						collectionConfigMongod := mongo_config.MongodConfig3_6_Storage_WiredTiger_CollectionConfig{}
						journalMongod := mongo_config.MongodConfig3_6_Storage_Journal{}
						wiredTigerMongod := mongo_config.MongodConfig3_6_Storage_WiredTiger{
							EngineConfig:     &engineConfigMongod,
							CollectionConfig: &collectionConfigMongod,
						}
						storageMongod := mongo_config.MongodConfig3_6_Storage{
							WiredTiger: &wiredTigerMongod,
							Journal:    &journalMongod,
						}
						if cacheSize, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongod.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						if blockCompressor, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.wired_tiger.0.block_compressor"); ok {
							blockCompressorInt := mongo_config.MongodConfig3_6_Storage_WiredTiger_CollectionConfig_Compressor_value[strings.ToUpper(blockCompressor.(string))]
							collectionConfigMongod.SetBlockCompressor(
								mongo_config.MongodConfig3_6_Storage_WiredTiger_CollectionConfig_Compressor(blockCompressorInt),
							)
						}

						if commitInterval, ok := d.GetOk("cluster_config.0.mongod.0.storage.0.journal.0.commit_interval"); ok {
							journalMongod.SetCommitInterval(&wrappers.Int64Value{Value: int64(commitInterval.(int))})
						}
						configMongod.SetStorage(&storageMongod)
					}
					if _, ok := d.GetOk("cluster_config.0.mongos.0.net"); ok {

						netMongos := mongo_config.MongosConfig3_6_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongos.0.net.0.max_incoming_connections"); ok {
							netMongos.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongos.SetNet(&netMongos)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.net"); ok {
						netMongoCfg := mongo_config.MongoCfgConfig3_6_Network{}
						if maxConnections, ok := d.GetOk("cluster_config.0.mongocfg.0.net.0.max_incoming_connections"); ok {
							netMongoCfg.SetMaxIncomingConnections(&wrappers.Int64Value{Value: int64(maxConnections.(int))})
						}
						configMongoCfg.SetNet(&netMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling"); ok {
						opProfilingMongoCfg := mongo_config.MongoCfgConfig3_6_OperationProfiling{}
						if mode, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.mode"); ok {
							modeInt := mongo_config.MongoCfgConfig3_6_OperationProfiling_Mode_value[strings.ToUpper(mode.(string))]
							opProfilingMongoCfg.SetMode(mongo_config.MongoCfgConfig3_6_OperationProfiling_Mode(modeInt))
						}

						if opThreshold, ok := d.GetOk("cluster_config.0.mongocfg.0.operation_profiling.0.slow_op_threshold"); ok {
							opProfilingMongoCfg.SetSlowOpThreshold(&wrappers.Int64Value{Value: int64(opThreshold.(int))})
						}
						configMongoCfg.SetOperationProfiling(&opProfilingMongoCfg)
					}
					if _, ok := d.GetOk("cluster_config.0.mongocfg.0.storage"); ok {
						engineConfigMongoCfg := mongo_config.MongoCfgConfig3_6_Storage_WiredTiger_EngineConfig{}
						wiredTigerMongoCfg := mongo_config.MongoCfgConfig3_6_Storage_WiredTiger{EngineConfig: &engineConfigMongoCfg}
						storageMongoCfg := mongo_config.MongoCfgConfig3_6_Storage{WiredTiger: &wiredTigerMongoCfg}

						if cacheSize, ok := d.GetOk("cluster_config.0.mongocfg.0.storage.0.wired_tiger.0.cache_size_gb"); ok {
							engineConfigMongoCfg.SetCacheSizeGb(&wrappers.DoubleValue{Value: cacheSize.(float64)})
						}
						configMongoCfg.SetStorage(&storageMongoCfg)
					}
					hostTypes := getSetOfHostTypes(d)
					var resourcesMongod, resourcesMongos, resourcesMongoCfg, resourcesMongoInfra *mongodb.Resources = getResources(d)
					var mongod *mongodb.MongodbSpec3_6_Mongod
					var mongos *mongodb.MongodbSpec3_6_Mongos
					var mongocfg *mongodb.MongodbSpec3_6_MongoCfg
					var mongoinfra *mongodb.MongodbSpec3_6_MongoInfra
					mongod = &mongodb.MongodbSpec3_6_Mongod{
						Config:    &configMongod,
						Resources: resourcesMongod,
					}

					if _, ok := hostTypes["MONGOS"]; ok {
						mongos = &mongodb.MongodbSpec3_6_Mongos{
							Config:    &configMongos,
							Resources: resourcesMongos,
						}
					}
					if _, ok := hostTypes["MONGOCFG"]; ok {
						mongocfg = &mongodb.MongodbSpec3_6_MongoCfg{
							Config:    &configMongoCfg,
							Resources: resourcesMongoCfg,
						}
					}
					if _, ok := hostTypes["MONGOINFRA"]; ok {
						mongoinfra = &mongodb.MongodbSpec3_6_MongoInfra{
							ConfigMongocfg: &configMongoCfg,
							ConfigMongos:   &configMongos,
							Resources:      resourcesMongoInfra,
						}
					}
					return &mongodb.ConfigSpec_MongodbSpec_3_6{
						MongodbSpec_3_6: &mongodb.MongodbSpec3_6{
							Mongod:     mongod,
							Mongos:     mongos,
							Mongocfg:   mongocfg,
							Mongoinfra: mongoinfra,
						},
					}
				},
			}
		}
	}
	return nil
}

func getResources(d *schema.ResourceData) (*mongodb.Resources, *mongodb.Resources, *mongodb.Resources, *mongodb.Resources) {
	// migration from resource to resource_*
	if _, ok := d.GetOk("resources_mongod"); !ok {
		resources := expandMongoDBResources(d)
		return resources, resources, resources, resources
	} else {
		return expandMongoDBResourcesWithType(d, "resources_mongod"),
			expandMongoDBResourcesWithType(d, "resources_mongos"),
			expandMongoDBResourcesWithType(d, "resources_mongocfg"),
			expandMongoDBResourcesWithType(d, "resources_mongoinfra")
	}
}

func getSetOfHostTypes(d *schema.ResourceData) map[string]struct{} {
	hosts := d.Get("host").([]interface{})

	var hostTypes = make(map[string]struct{})

	for _, host := range hosts {
		hostConfig := host.(map[string]interface{})
		hostTypes[strings.ToUpper(hostConfig["type"].(string))] = struct{}{}
	}
	return hostTypes
}

func flattenMongoDBClusterConfig(cc *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
	mongodbSpecHelper := GetMongodbSpecHelper(cc.Version)

	flattenMongod, err := mongodbSpecHelper.FlattenMongod(cc, d)
	if err != nil {
		return nil, err
	}

	flattenMongos, err := mongodbSpecHelper.FlattenMongos(cc, d)
	if err != nil {
		return nil, err
	}

	flattenMongocfg, err := mongodbSpecHelper.FlattenMongocfg(cc, d)
	if err != nil {
		return nil, err
	}

	flattenConfig := []map[string]interface{}{
		{
			"backup_window_start": []*map[string]interface{}{
				{
					"hours":   int(cc.BackupWindowStart.Hours),
					"minutes": int(cc.BackupWindowStart.Minutes),
				},
			},
			"backup_retain_period_days":     int(cc.GetBackupRetainPeriodDays().GetValue()),
			"feature_compatibility_version": cc.FeatureCompatibilityVersion,
			"version":                       cc.Version,
			"access": []interface{}{
				map[string]interface{}{
					"data_lens":     cc.Access.DataLens,
					"data_transfer": cc.Access.DataTransfer,
				},
			},
			"performance_diagnostics": []interface{}{
				map[string]interface{}{
					"enabled": cc.PerformanceDiagnostics.ProfilingEnabled,
				},
			},
			"mongod":   flattenMongod,
			"mongos":   flattenMongos,
			"mongocfg": flattenMongocfg,
		},
	}
	return flattenConfig, nil
}

func parseMongoDBWeekDay(wd string) (mongodb.WeeklyMaintenanceWindow_WeekDay, error) {
	val, ok := mongodb.WeeklyMaintenanceWindow_WeekDay_value[wd]
	// do not allow WEEK_DAY_UNSPECIFIED
	if !ok || val == 0 {
		return mongodb.WeeklyMaintenanceWindow_WEEK_DAY_UNSPECIFIED,
			fmt.Errorf("value for 'day' should be one of %s, not `%s`",
				getJoinedKeys(getEnumValueMapKeysExt(mongodb.WeeklyMaintenanceWindow_WeekDay_value, true)), wd)
	}

	return mongodb.WeeklyMaintenanceWindow_WeekDay(val), nil
}

func expandMongoDBMaintenanceWindow(d *schema.ResourceData) (*mongodb.MaintenanceWindow, error) {
	mwType, ok := d.GetOk("maintenance_window.0.type")
	if !ok {
		return nil, nil
	}

	result := &mongodb.MaintenanceWindow{}

	switch mwType {
	case "ANYTIME":
		timeSet := false
		if _, ok := d.GetOk("maintenance_window.0.day"); ok {
			timeSet = true
		}
		if _, ok := d.GetOk("maintenance_window.0.hour"); ok {
			timeSet = true
		}
		if timeSet {
			return nil, fmt.Errorf("with ANYTIME type of maintenance window both DAY and HOUR should be omitted")
		}
		result.SetAnytime(&mongodb.AnytimeMaintenanceWindow{})

	case "WEEKLY":
		weekly := &mongodb.WeeklyMaintenanceWindow{}
		if val, ok := d.GetOk("maintenance_window.0.day"); ok {
			var err error
			weekly.Day, err = parseMongoDBWeekDay(val.(string))
			if err != nil {
				return nil, err
			}
		}
		if v, ok := d.GetOk("maintenance_window.0.hour"); ok {
			weekly.Hour = int64(v.(int))
		}

		result.SetWeeklyMaintenanceWindow(weekly)
	}

	return result, nil
}

func flattenMongoDBMaintenanceWindow(mw *mongodb.MaintenanceWindow) []map[string]interface{} {
	result := map[string]interface{}{}

	if val := mw.GetAnytime(); val != nil {
		result["type"] = "ANYTIME"
	}

	if val := mw.GetWeeklyMaintenanceWindow(); val != nil {
		result["type"] = "WEEKLY"
		result["day"] = val.Day.String()
		result["hour"] = val.Hour
	}

	return []map[string]interface{}{result}
}

func flattenMongoDBResources(m *mongodb.Resources) []map[string]interface{} {
	res := map[string]interface{}{}

	res["resource_preset_id"] = m.ResourcePresetId
	res["disk_size"] = toGigabytes(m.DiskSize)
	res["disk_type_id"] = m.DiskTypeId

	return []map[string]interface{}{res}
}

func flattenMongoDBHosts(hs []*mongodb.Host) ([]map[string]interface{}, error) {
	var res []map[string]interface{}

	for _, h := range hs {
		m := map[string]interface{}{}
		m["zone_id"] = h.ZoneId
		m["subnet_id"] = h.SubnetId
		m["name"] = h.Name
		m["role"] = h.Role.String()
		m["health"] = h.Health.String()
		m["assign_public_ip"] = h.AssignPublicIp
		m["shard_name"] = h.ShardName
		m["type"] = h.Type.String()
		res = append(res, m)
	}

	return res, nil
}

func expandMongoDBHosts(d *schema.ResourceData) ([]*mongodb.HostSpec, error) {
	var result []*mongodb.HostSpec
	hosts := d.Get("host").([]interface{})

	for _, v := range hosts {
		config := v.(map[string]interface{})
		host := expandMongoDBHost(config)
		result = append(result, host)
	}

	return result, nil
}

func expandMongoDBHost(config map[string]interface{}) *mongodb.HostSpec {
	host := &mongodb.HostSpec{}
	if v, ok := config["type"]; ok {
		host.Type = mongodb.Host_Type(mongodb.Host_Type_value[strings.ToUpper(v.(string))])
	}

	if v, ok := config["zone_id"]; ok {
		host.ZoneId = v.(string)
	}

	if v, ok := config["subnet_id"]; ok {
		host.SubnetId = v.(string)
	}

	if v, ok := config["shard_name"]; ok {
		host.ShardName = v.(string)
	}

	if v, ok := config["assign_public_ip"]; ok {
		host.AssignPublicIp = v.(bool)
	}
	return host
}

func parseMongoDBEnv(e string) (mongodb.Cluster_Environment, error) {
	v, ok := mongodb.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(mongodb.Cluster_Environment_value)), e)
	}
	return mongodb.Cluster_Environment(v), nil
}

func mongodbUserPermissionHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["database_name"]; ok {
		//goland:noinspection GoDeprecation (this comment suppress warning in Idea IDE about using Deprecated method)
		return hashcode.String(n.(string))
	}
	return 0
}

func mongodbUserHash(v interface{}) int {
	var buf bytes.Buffer

	m := v.(map[string]interface{})
	if n, ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", n.(string)))
	}
	if p, ok := m["password"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", p.(string)))
	}
	if ps, ok := m["permission"]; ok {
		buf.WriteString(fmt.Sprintf("%v-", ps.(*schema.Set).List()))
	}

	//goland:noinspection GoDeprecation (this comment suppress warning in Idea IDE about using Deprecated method)
	return hashcode.String(buf.String())
}

func mongodbDatabaseHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["name"]; ok {
		//goland:noinspection GoDeprecation (this comment suppress warning in Idea IDE about using Deprecated method)
		return hashcode.String(n.(string))
	}
	return 0
}

func mongodbUsersPasswords(users []*mongodb.UserSpec) map[string]string {
	result := map[string]string{}
	for _, u := range users {
		result[u.Name] = u.Password
	}
	return result
}

func flattenMongoDBUsers(users []*mongodb.User, passwords map[string]string) *schema.Set {
	result := schema.NewSet(mongodbUserHash, nil)

	for _, user := range users {
		u := map[string]interface{}{}
		u["name"] = user.Name

		perms := schema.NewSet(mongodbUserPermissionHash, nil)
		for _, perm := range user.Permissions {
			p := map[string]interface{}{}
			p["database_name"] = perm.DatabaseName
			p["roles"] = perm.Roles
			perms.Add(p)
		}
		u["permission"] = perms

		if p, ok := passwords[user.Name]; ok {
			u["password"] = p
		}
		result.Add(u)
	}
	return result
}

func flattenMongoDBDatabases(dbs []*mongodb.Database) *schema.Set {
	result := schema.NewSet(mongodbDatabaseHash, nil)

	for _, d := range dbs {
		m := make(map[string]interface{})
		m["name"] = d.Name
		result.Add(m)
	}
	return result
}

func expandMongoDBUser(u map[string]interface{}) *mongodb.UserSpec {
	user := &mongodb.UserSpec{}

	if v, ok := u["name"]; ok {
		user.Name = v.(string)
	}

	if v, ok := u["password"]; ok {
		user.Password = v.(string)
	}

	if v, ok := u["permission"]; ok {
		user.Permissions = expandMongoDBUserPermissions(v.(*schema.Set))
	}

	return user
}

func expandMongoDBUserSpecs(d *schema.ResourceData) ([]*mongodb.UserSpec, error) {
	var result []*mongodb.UserSpec
	users := d.Get("user").(*schema.Set)

	for _, u := range users.List() {
		m := u.(map[string]interface{})

		result = append(result, expandMongoDBUser(m))
	}

	return result, nil
}

func expandMongoDBUserPermissions(ps *schema.Set) []*mongodb.Permission {
	var result []*mongodb.Permission

	for _, p := range ps.List() {
		m := p.(map[string]interface{})
		permission := &mongodb.Permission{}
		if v, ok := m["database_name"]; ok {
			permission.DatabaseName = v.(string)
		}

		if v, ok := m["roles"]; ok {
			roles := make([]string, len(v.([]interface{})))
			for n, item := range v.([]interface{}) {
				roles[n] = item.(string)
			}

			permission.Roles = roles
		}
		result = append(result, permission)
	}
	return result
}

func expandMongoDBDatabases(d *schema.ResourceData) ([]*mongodb.DatabaseSpec, error) {
	var result []*mongodb.DatabaseSpec
	dbs := d.Get("database").(*schema.Set).List()

	for _, d := range dbs {
		m := d.(map[string]interface{})
		db := &mongodb.DatabaseSpec{}

		if v, ok := m["name"]; ok {
			db.Name = v.(string)
		}

		result = append(result, db)
	}
	return result, nil
}

func expandMongoDBResources(d *schema.ResourceData) *mongodb.Resources {
	return expandMongoDBResourcesWithType(d, "resources")
}

func expandMongoDBResourcesWithType(d *schema.ResourceData, hostType string) *mongodb.Resources {
	if _, ok := d.GetOk(hostType); !ok {
		return nil
	}
	res := mongodb.Resources{
		DiskSize:         toBytes(d.Get(hostType + ".0.disk_size").(int)),
		DiskTypeId:       d.Get(hostType + ".0.disk_type_id").(string),
		ResourcePresetId: d.Get(hostType + ".0.resource_preset_id").(string),
	}
	return &res
}

func expandMongoDBBackupWindowStart(d *schema.ResourceData) *timeofday.TimeOfDay {
	res := timeofday.TimeOfDay{
		Hours:   int32(d.Get("cluster_config.0.backup_window_start.0.hours").(int)),
		Minutes: int32(d.Get("cluster_config.0.backup_window_start.0.minutes").(int)),
	}

	return &res
}

func expandMongoDBBackupRetainPeriod(d *schema.ResourceData) *wrappers.Int64Value {
	if backupRetainPeriod, ok := d.GetOk("cluster_config.0.backup_retain_period_days"); ok {
		return &wrappers.Int64Value{Value: int64(backupRetainPeriod.(int))}
	}
	return nil
}

func mongodbDatabasesDiff(currDBs []*mongodb.Database, targetDBs []*mongodb.DatabaseSpec) ([]string, []string) {
	m := map[string]bool{}
	var toAdd []string
	toDelete := map[string]bool{}
	for _, db := range currDBs {
		toDelete[db.Name] = true
		m[db.Name] = true
	}

	for _, db := range targetDBs {
		delete(toDelete, db.Name)
		if _, ok := m[db.Name]; !ok {
			toAdd = append(toAdd, db.Name)
		}
	}

	var toDel []string
	for u := range toDelete {
		toDel = append(toDel, u)
	}

	return toDel, toAdd
}

func checkSupportedVersion(version string) error {
	_, ok := supportedVersions[version]
	if !ok {
		expected := reflect.ValueOf(supportedVersions).MapKeys()
		return fmt.Errorf("Wrong MongoDB version: required either %v, got %s", expected, version)
	}
	return nil
}

func extractVersion(d *schema.ResourceData) (string, error) {
	version := d.Get("cluster_config.0.version").(string)
	err := checkSupportedVersion(version)
	if err != nil {
		return "", err
	}
	return version, nil
}

func flattendVersion(version string) string {
	result := strings.Replace(version, ".", "_", -1)
	result = strings.Replace(result, "-", "_", -1)
	return result
}

func Map[F, T any](s []F, f func(F) T) []T {
	r := make([]T, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}
