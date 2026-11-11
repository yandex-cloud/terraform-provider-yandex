package pluginsdk

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/dataproc/v1"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func expandDataprocCreateClusterConfigSpec(d *schema.ResourceData) (*dataproc.CreateClusterConfigSpec, error) {
	subclusters, err := expandDataprocSubclustersSpec(d)
	if err != nil {
		return nil, err
	}
	hadoop, err := expandDataprocHadoopConfig(d)
	if err != nil {
		return nil, err
	}
	return &dataproc.CreateClusterConfigSpec{
		VersionId:       d.Get("cluster_config.0.version_id").(string),
		Hadoop:          hadoop,
		SubclustersSpec: subclusters,
	}, nil
}

func expandDataprocHadoopConfig(d *schema.ResourceData) (*dataproc.HadoopConfig, error) {
	initActions, err := expandDataprocInitActions(d)
	if err != nil {
		return nil, err
	}
	return &dataproc.HadoopConfig{
		Services:              expandDataprocServices(d),
		Properties:            expandDataprocProperties(d),
		SshPublicKeys:         expandDataprocSSHPublicKeys(d),
		InitializationActions: initActions,
		OsloginEnabled:        d.Get("cluster_config.0.hadoop.0.oslogin").(bool),
	}, nil
}

func expandDataprocServices(d *schema.ResourceData) []dataproc.HadoopConfig_Service {
	set := d.Get("cluster_config.0.hadoop.0.services").(*schema.Set)
	serviceNames := sdkcommon.ConvertStringSet(set)
	sort.Strings(serviceNames)
	services := make([]dataproc.HadoopConfig_Service, len(serviceNames))

	for i, serviceName := range serviceNames {
		// service name is checked by validation
		serviceID := dataproc.HadoopConfig_Service_value[serviceName]
		services[i] = dataproc.HadoopConfig_Service(serviceID)
	}

	return services
}

func expandDataprocProperties(d *schema.ResourceData) map[string]string {
	v := d.Get("cluster_config.0.hadoop.0.properties").(map[string]interface{})
	return convertStringMap(v)
}

func expandDataprocSSHPublicKeys(d *schema.ResourceData) []string {
	v := d.Get("cluster_config.0.hadoop.0.ssh_public_keys").(*schema.Set)
	return sdkcommon.ConvertStringSet(v)
}

func expandDataprocInitActions(d *schema.ResourceData) ([]*dataproc.InitializationAction, error) {
	list := d.Get("cluster_config.0.hadoop.0.initialization_action").([]interface{})
	initActions := make([]*dataproc.InitializationAction, len(list))
	for index, element := range list {
		elementMap := element.(map[string]interface{})
		initAction := &dataproc.InitializationAction{
			Uri: elementMap["uri"].(string),
		}

		if elementMap["timeout"].(string) != "" {
			timeout, err := strconv.Atoi(elementMap["timeout"].(string))
			if err != nil {
				return nil, err
			}
			initAction.Timeout = int64(timeout)
		}

		argsList := elementMap["args"].([]interface{})
		if argsList != nil {
			args := make([]string, len(argsList))
			for listIndex, arg := range argsList {
				args[listIndex] = arg.(string)
			}
			initAction.Args = args
		}

		initActions[index] = initAction
	}
	return initActions, nil
}

func expandDataprocSubclustersSpec(d *schema.ResourceData) ([]*dataproc.CreateSubclusterConfigSpec, error) {
	rootKey := "cluster_config.0.subcluster_spec"
	list := d.Get(rootKey).([]interface{})
	subclusters := make([]*dataproc.CreateSubclusterConfigSpec, len(list))
	for index, element := range list {
		subcluster, err := expandDataprocSubclusterSpec(element)
		if err != nil {
			return nil, err
		}
		subclusters[index] = subcluster
	}

	return subclusters, nil
}

func expandDataprocSubclusterSpec(element interface{}) (*dataproc.CreateSubclusterConfigSpec, error) {
	subclusterSpec := element.(map[string]interface{})
	roleName := subclusterSpec["role"].(string)
	roleID := dataproc.Role_value[roleName]
	resourcesSpec := subclusterSpec["resources"].([]interface{})[0]

	subcluster := &dataproc.CreateSubclusterConfigSpec{
		Role:           dataproc.Role(roleID),
		Name:           subclusterSpec["name"].(string),
		SubnetId:       subclusterSpec["subnet_id"].(string),
		HostsCount:     int64(subclusterSpec["hosts_count"].(int)),
		Resources:      expandDataprocResources(resourcesSpec),
		AssignPublicIp: subclusterSpec["assign_public_ip"].(bool),
	}
	if v, ok := subclusterSpec["autoscaling_config"]; ok {
		autoscalingConfigs := v.([]interface{})
		if len(autoscalingConfigs) > 0 {
			autoscalingConfig, err := expandDataprocAutoscalingConfig(autoscalingConfigs[0])
			if err != nil {
				return nil, err
			}
			subcluster.AutoscalingConfig = autoscalingConfig
		}
	}

	return subcluster, nil
}

func expandDataprocResources(r interface{}) *dataproc.Resources {
	resources := &dataproc.Resources{}
	resourcesMap := r.(map[string]interface{})

	if v, ok := resourcesMap["resource_preset_id"]; ok {
		resources.ResourcePresetId = v.(string)
	}
	if v, ok := resourcesMap["disk_size"]; ok {
		resources.DiskSize = sdkcommon.ToBytes(v.(int))
	}
	if v, ok := resourcesMap["disk_type_id"]; ok {
		resources.DiskTypeId = v.(string)
	}
	return resources
}

func stringToDuration(s string) (*duration.Duration, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return &duration.Duration{Seconds: int64(i)}, nil
}

func expandAutoscalingDurationField(autoscalingConfigMap map[string]interface{}, fieldName string) (*duration.Duration, error) {
	if v, ok := autoscalingConfigMap[fieldName]; ok {
		durationSeconds := v.(string)
		if durationSeconds != "" {
			return stringToDuration(durationSeconds)
		}
	}
	return nil, nil
}

func expandDataprocAutoscalingConfig(r interface{}) (*dataproc.AutoscalingConfig, error) {
	autoscalingConfig := &dataproc.AutoscalingConfig{}
	autoscalingConfigMap := r.(map[string]interface{})
	log.Printf("[DEBUG] autoscalingConfigMap = %v", autoscalingConfigMap)
	if v, ok := autoscalingConfigMap["max_hosts_count"]; ok {
		if v.(int) >= 0 {
			autoscalingConfig.MaxHostsCount = int64(v.(int))
		}
	}
	if v, ok := autoscalingConfigMap["preemptible"]; ok {
		autoscalingConfig.Preemptible = v.(bool)
	}

	durationValue, err := expandAutoscalingDurationField(autoscalingConfigMap, "measurement_duration")
	if err != nil {
		return nil, err
	}
	autoscalingConfig.MeasurementDuration = durationValue

	durationValue, err = expandAutoscalingDurationField(autoscalingConfigMap, "warmup_duration")
	if err != nil {
		return nil, err
	}
	autoscalingConfig.WarmupDuration = durationValue

	durationValue, err = expandAutoscalingDurationField(autoscalingConfigMap, "stabilization_duration")
	if err != nil {
		return nil, err
	}
	autoscalingConfig.StabilizationDuration = durationValue

	if v, ok := autoscalingConfigMap["cpu_utilization_target"]; ok {
		value := v.(string)
		if value != "" {
			valueInt, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}

			autoscalingConfig.CpuUtilizationTarget = float64(valueInt)
		}
	}
	if v, ok := autoscalingConfigMap["decommission_timeout"]; ok {
		durationSeconds := v.(string)
		if durationSeconds != "" {
			durationSecondsInt, err := strconv.Atoi(durationSeconds)
			if err != nil {
				return nil, err
			}

			autoscalingConfig.DecommissionTimeout = int64(durationSecondsInt)
		}
	}

	return autoscalingConfig, nil
}

func FlattenDataprocClusterConfig(cluster *dataproc.Cluster, subclusters []*dataproc.Subcluster) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"version_id":      cluster.Config.VersionId,
			"hadoop":          flattenDataprocHadoopConfig(cluster.Config.Hadoop),
			"subcluster_spec": flattenDataprocSubclusters(subclusters),
		},
	}
}

func flattenDataprocHadoopConfig(config *dataproc.HadoopConfig) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"services":              flattenDataprocServices(config.Services),
			"properties":            config.Properties,
			"ssh_public_keys":       config.SshPublicKeys,
			"initialization_action": flattenInitActions(config.InitializationActions),
			"oslogin":               config.OsloginEnabled,
		},
	}
}

func flattenInitActions(actions []*dataproc.InitializationAction) []interface{} {
	result := make([]interface{}, len(actions))
	for index, action := range actions {
		result[index] = map[string]interface{}{
			"uri":     action.Uri,
			"args":    action.Args,
			"timeout": strconv.Itoa(int(action.Timeout)),
		}
	}
	return result
}

func flattenDataprocServices(services []dataproc.HadoopConfig_Service) []string {
	serviceNames := make([]string, len(services))
	for idx, service := range services {
		serviceNames[idx] = service.String()
	}
	return serviceNames
}

func flattenDataprocSubclusters(subclusters []*dataproc.Subcluster) []interface{} {
	result := make([]interface{}, len(subclusters))
	for idx, subcluster := range subclusters {
		result[idx] = flattenDataprocSubcluster(subcluster)
	}
	return result
}

func flattenDataprocSubcluster(subcluster *dataproc.Subcluster) map[string]interface{} {
	result := map[string]interface{}{
		"id":               subcluster.Id,
		"name":             subcluster.Name,
		"role":             subcluster.Role.String(),
		"resources":        flattenDataprocResources(subcluster.Resources),
		"subnet_id":        subcluster.SubnetId,
		"hosts_count":      subcluster.HostsCount,
		"assign_public_ip": subcluster.AssignPublicIp,
	}
	if subcluster.AutoscalingConfig != nil {
		result["autoscaling_config"] = flattenDataprocAutoscalingConfig(subcluster.AutoscalingConfig)
	}
	return result
}

func flattenDataprocResources(r *dataproc.Resources) []map[string]interface{} {
	res := map[string]interface{}{}

	res["resource_preset_id"] = r.ResourcePresetId
	res["disk_type_id"] = r.DiskTypeId
	res["disk_size"] = sdkcommon.ToGigabytes(r.DiskSize)

	return []map[string]interface{}{res}
}

func flattenDataprocAutoscalingConfig(r *dataproc.AutoscalingConfig) []map[string]interface{} {
	res := map[string]interface{}{}

	res["max_hosts_count"] = int(r.MaxHostsCount)
	res["preemptible"] = r.Preemptible
	res["measurement_duration"] = strconv.Itoa(int(r.MeasurementDuration.Seconds))
	res["warmup_duration"] = strconv.Itoa(int(r.WarmupDuration.Seconds))
	res["stabilization_duration"] = strconv.Itoa(int(r.StabilizationDuration.Seconds))
	res["cpu_utilization_target"] = strconv.FormatFloat(r.CpuUtilizationTarget, 'f', 0, 64)
	res["decommission_timeout"] = strconv.Itoa(int(r.DecommissionTimeout))

	return []map[string]interface{}{res}
}

func parseDataprocEnv(e string) (dataproc.Cluster_Environment, error) {
	v, ok := dataproc.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`", sdkcommon.GetJoinedKeys(sdkcommon.GetEnumValueMapKeys(dataproc.Cluster_Environment_value)), e)
	}
	return dataproc.Cluster_Environment(v), nil
}
