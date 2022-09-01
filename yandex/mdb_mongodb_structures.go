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
	"5.0-enterprise": true,
	"4.4-enterprise": true,
	"5.0":            true,
	"4.4":            true,
	"4.2":            true,
	"4.0":            true,
	"3.6":            true,
}

type MongodbSpecHelper struct {
	FlattenResources func(c *mongodb.ClusterConfig) ([]map[string]interface{}, error)
	FlattenMongod    func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error)
	Expand           func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec
}

func GetMongodbSpecHelper(version string) *MongodbSpecHelper {
	switch version {
	case "5.0-enterprise":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig) ([]map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0Enterprise).Mongodb_5_0Enterprise
					if spec.Mongod != nil {
						return flattenMongoDBResources(spec.Mongod.Resources), nil
					}
					if spec.Mongos != nil {
						return flattenMongoDBResources(spec.Mongos.Resources), nil
					}
					if spec.Mongocfg != nil {
						return flattenMongoDBResources(spec.Mongocfg.Resources), nil
					}
					return nil, fmt.Errorf("Non empty service not found in mongo spec")
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
						if set_parameter := user_config.GetSetParameter(); set_parameter != nil {
							set_parameter_data := map[string]interface{}{}
							if set_parameter.GetAuditAuthorizationSuccess() != nil {
								set_parameter_data["audit_authorization_success"] = set_parameter.GetAuditAuthorizationSuccess().GetValue()
							}
							result["set_parameter"] = []map[string]interface{}{set_parameter_data}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					config := mongo_config.MongodConfig5_0Enterprise{}

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
					config.SetSecurity(&security)

					audit_log := mongo_config.MongodConfig5_0Enterprise_AuditLog{}
					if filter := d.Get("cluster_config.0.mongod.0.audit_log.0.filter"); filter != nil {
						audit_log.SetFilter(filter.(string))
					}
					// Note: right now runtime_configuration unsupported, so we should comment this statement
					//if rt := d.Get("cluster_config.0.mongod.0.audit_log.0.runtime_configuration"); rt != nil {
					//	audit_log.SetRuntimeConfiguration(&wrappers.BoolValue{Value: rt.(bool)})
					//}
					config.SetAuditLog(&audit_log)

					set_paramenter := mongo_config.MongodConfig5_0Enterprise_SetParameter{}
					if success := d.Get("cluster_config.0.mongod.0.set_parameter.0.audit_authorization_success"); success != nil {
						set_paramenter.SetAuditAuthorizationSuccess(&wrappers.BoolValue{Value: success.(bool)})
					}
					config.SetSetParameter(&set_paramenter)

					resources := expandMongoDBResources(d)

					return &mongodb.ConfigSpec_MongodbSpec_5_0Enterprise{
						MongodbSpec_5_0Enterprise: &mongodb.MongodbSpec5_0Enterprise{
							Mongod: &mongodb.MongodbSpec5_0Enterprise_Mongod{
								Config:    &config,
								Resources: resources,
							},
							Mongos: &mongodb.MongodbSpec5_0Enterprise_Mongos{
								Resources: resources,
							},
							Mongocfg: &mongodb.MongodbSpec5_0Enterprise_MongoCfg{
								Resources: resources,
							},
						},
					}
				},
			}
		}
	case "5.0":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig) ([]map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0).Mongodb_5_0
					if spec.Mongod != nil {
						return flattenMongoDBResources(spec.Mongod.Resources), nil
					}
					if spec.Mongos != nil {
						return flattenMongoDBResources(spec.Mongos.Resources), nil
					}
					if spec.Mongocfg != nil {
						return flattenMongoDBResources(spec.Mongocfg.Resources), nil
					}
					return nil, fmt.Errorf("Non empty service not found in mongo spec")
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					resources := expandMongoDBResources(d)
					return &mongodb.ConfigSpec_MongodbSpec_5_0{
						MongodbSpec_5_0: &mongodb.MongodbSpec5_0{
							Mongod: &mongodb.MongodbSpec5_0_Mongod{
								Resources: resources,
							},
							Mongos: &mongodb.MongodbSpec5_0_Mongos{
								Resources: resources,
							},
							Mongocfg: &mongodb.MongodbSpec5_0_MongoCfg{
								Resources: resources,
							},
						},
					}
				},
			}
		}
	case "4.4-enterprise":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig) ([]map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4Enterprise).Mongodb_4_4Enterprise
					if spec.Mongod != nil {
						return flattenMongoDBResources(spec.Mongod.Resources), nil
					}
					if spec.Mongos != nil {
						return flattenMongoDBResources(spec.Mongos.Resources), nil
					}
					if spec.Mongocfg != nil {
						return flattenMongoDBResources(spec.Mongocfg.Resources), nil
					}
					return nil, fmt.Errorf("Non empty service not found in mongo spec")
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					mongod := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4Enterprise).Mongodb_4_4Enterprise.Mongod
					if mongod != nil {
						user_config := mongod.GetConfig().GetUserConfig()
						default_config := mongod.GetConfig().GetDefaultConfig()

						result := map[string]interface{}{}

						if security := user_config.GetSecurity(); security != nil {
							flattenSecurity := map[string]interface{}{}
							if enableEncryption := security.GetEnableEncryption(); enableEncryption != nil {
								flattenSecurity["enable_encryption"] = enableEncryption.GetValue()
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
						if set_parameter := user_config.GetSetParameter(); set_parameter != nil {
							set_parameter_data := map[string]interface{}{}
							if set_parameter.GetAuditAuthorizationSuccess() != nil {
								set_parameter_data["audit_authorization_success"] = set_parameter.GetAuditAuthorizationSuccess().GetValue()
							}
							result["set_parameter"] = []map[string]interface{}{set_parameter_data}
						}

						return []map[string]interface{}{result}, nil
					}
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					config := mongo_config.MongodConfig4_4Enterprise{}

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
					config.SetSecurity(&security)

					audit_log := mongo_config.MongodConfig4_4Enterprise_AuditLog{}
					if filter := d.Get("cluster_config.0.mongod.0.audit_log.0.filter"); filter != nil {
						audit_log.SetFilter(filter.(string))
					}
					config.SetAuditLog(&audit_log)

					set_paramenter := mongo_config.MongodConfig4_4Enterprise_SetParameter{}
					if success := d.Get("cluster_config.0.mongod.0.set_parameter.0.audit_authorization_success"); success != nil {
						set_paramenter.SetAuditAuthorizationSuccess(&wrappers.BoolValue{Value: success.(bool)})
					}
					config.SetSetParameter(&set_paramenter)

					resources := expandMongoDBResources(d)

					return &mongodb.ConfigSpec_MongodbSpec_4_4Enterprise{
						MongodbSpec_4_4Enterprise: &mongodb.MongodbSpec4_4Enterprise{
							Mongod: &mongodb.MongodbSpec4_4Enterprise_Mongod{
								Config:    &config,
								Resources: resources,
							},
							Mongos: &mongodb.MongodbSpec4_4Enterprise_Mongos{
								Resources: resources,
							},
							Mongocfg: &mongodb.MongodbSpec4_4Enterprise_MongoCfg{
								Resources: resources,
							},
						},
					}
				},
			}
		}
	case "4.4":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig) ([]map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4).Mongodb_4_4
					if spec.Mongod != nil {
						return flattenMongoDBResources(spec.Mongod.Resources), nil
					}
					if spec.Mongos != nil {
						return flattenMongoDBResources(spec.Mongos.Resources), nil
					}
					if spec.Mongocfg != nil {
						return flattenMongoDBResources(spec.Mongocfg.Resources), nil
					}
					return nil, fmt.Errorf("Non empty service not found in mongo spec")
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					resources := expandMongoDBResources(d)
					return &mongodb.ConfigSpec_MongodbSpec_4_4{
						MongodbSpec_4_4: &mongodb.MongodbSpec4_4{
							Mongod: &mongodb.MongodbSpec4_4_Mongod{
								Resources: resources,
							},
							Mongos: &mongodb.MongodbSpec4_4_Mongos{
								Resources: resources,
							},
							Mongocfg: &mongodb.MongodbSpec4_4_MongoCfg{
								Resources: resources,
							},
						},
					}
				},
			}
		}
	case "4.2":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig) ([]map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_2).Mongodb_4_2
					if spec.Mongod != nil {
						return flattenMongoDBResources(spec.Mongod.Resources), nil
					}
					if spec.Mongos != nil {
						return flattenMongoDBResources(spec.Mongos.Resources), nil
					}
					if spec.Mongocfg != nil {
						return flattenMongoDBResources(spec.Mongocfg.Resources), nil
					}
					return nil, fmt.Errorf("Non empty service not found in mongo spec")
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					resources := expandMongoDBResources(d)
					return &mongodb.ConfigSpec_MongodbSpec_4_2{
						MongodbSpec_4_2: &mongodb.MongodbSpec4_2{
							Mongod: &mongodb.MongodbSpec4_2_Mongod{
								Resources: resources,
							},
							Mongos: &mongodb.MongodbSpec4_2_Mongos{
								Resources: resources,
							},
							Mongocfg: &mongodb.MongodbSpec4_2_MongoCfg{
								Resources: resources,
							},
						},
					}
				},
			}
		}
	case "4.0":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig) ([]map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_0).Mongodb_4_0
					if spec.Mongod != nil {
						return flattenMongoDBResources(spec.Mongod.Resources), nil
					}
					if spec.Mongos != nil {
						return flattenMongoDBResources(spec.Mongos.Resources), nil
					}
					if spec.Mongocfg != nil {
						return flattenMongoDBResources(spec.Mongocfg.Resources), nil
					}
					return nil, fmt.Errorf("Non empty service not found in mongo spec")
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					resources := expandMongoDBResources(d)
					return &mongodb.ConfigSpec_MongodbSpec_4_0{
						MongodbSpec_4_0: &mongodb.MongodbSpec4_0{
							Mongod: &mongodb.MongodbSpec4_0_Mongod{
								Resources: resources,
							},
							Mongos: &mongodb.MongodbSpec4_0_Mongos{
								Resources: resources,
							},
							Mongocfg: &mongodb.MongodbSpec4_0_MongoCfg{
								Resources: resources,
							},
						},
					}
				},
			}
		}
	case "3.6":
		{
			return &MongodbSpecHelper{

				FlattenResources: func(c *mongodb.ClusterConfig) ([]map[string]interface{}, error) {
					spec := c.Mongodb.(*mongodb.ClusterConfig_Mongodb_3_6).Mongodb_3_6
					if spec.Mongod != nil {
						return flattenMongoDBResources(spec.Mongod.Resources), nil
					}
					if spec.Mongos != nil {
						return flattenMongoDBResources(spec.Mongos.Resources), nil
					}
					if spec.Mongocfg != nil {
						return flattenMongoDBResources(spec.Mongocfg.Resources), nil
					}
					return nil, fmt.Errorf("Non empty service not found in mongo spec")
				},

				FlattenMongod: func(c *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
					return []map[string]interface{}{}, nil
				},

				Expand: func(d *schema.ResourceData) mongodb.ConfigSpec_MongodbSpec {
					resources := expandMongoDBResources(d)
					return &mongodb.ConfigSpec_MongodbSpec_3_6{
						MongodbSpec_3_6: &mongodb.MongodbSpec3_6{
							Mongod: &mongodb.MongodbSpec3_6_Mongod{
								Resources: resources,
							},
							Mongos: &mongodb.MongodbSpec3_6_Mongos{
								Resources: resources,
							},
							Mongocfg: &mongodb.MongodbSpec3_6_MongoCfg{
								Resources: resources,
							},
						},
					}
				},
			}
		}
	}
	return nil
}

func flattenMongoDBClusterConfig(cc *mongodb.ClusterConfig, d *schema.ResourceData) ([]map[string]interface{}, error) {
	mongodbSpecHelper := GetMongodbSpecHelper(cc.Version)

	flattenMongod, err := mongodbSpecHelper.FlattenMongod(cc, d)
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
			"feature_compatibility_version": cc.FeatureCompatibilityVersion,
			"version":                       cc.Version,
			"access": []interface{}{
				map[string]interface{}{
					"data_lens":     cc.Access.DataLens,
					"data_transfer": cc.Access.DataTransfer,
				},
			},
			"mongod": flattenMongod,
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
		host.Type = mongodb.Host_Type(mongodb.Host_Type_value[v.(string)])
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
	res := mongodb.Resources{
		DiskSize:         toBytes(d.Get("resources.0.disk_size").(int)),
		DiskTypeId:       d.Get("resources.0.disk_type_id").(string),
		ResourcePresetId: d.Get("resources.0.resource_preset_id").(string),
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
