package common

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// MySQLConfigFieldName returns the version-specific configuration field in the API.
func MySQLConfigFieldName(version string) (string, error) {
	switch version {
	case "5.7":
		return "mysql_config_5_7", nil
	case "8.0":
		return "mysql_config_8_0", nil
	case "8.4":
		return "mysql_config_8_4", nil
	case "9.7":
		return "mysql_config_9_7", nil
	default:
		return "", fmt.Errorf("unsupported MySQL version: %s", version)
	}
}

// MySQL97UpgradeSettingsPaths includes every target setting because the upgrade
// can add user settings absent from both the old state and the plan.
// RemoveUnchangedSettings narrows these paths against the post-upgrade response.
func MySQL97UpgradeSettingsPaths() []string {
	fields := (&config.MysqlConfig9_7{}).ProtoReflect().Descriptor().Fields()
	paths := make([]string, fields.Len())
	for i := range paths {
		paths[i] = "config_spec.mysql_config_9_7." + string(fields.Get(i).Name())
	}
	return paths
}

// RemoveUnchangedSettings compares the pending settings with the user config
// read after an upgrade. The backend may already have migrated their values and
// rejects an Update that changes nothing. Defaults/effective values cannot be
// used here: setting an explicit value and resetting it are different operations.
func RemoveUnchangedSettings(request *mysql.UpdateClusterRequest, current *mysql.ClusterConfig) error {
	spec := request.GetConfigSpec().ProtoReflect()
	configField := spec.WhichOneof(spec.Descriptor().Oneofs().ByName("mysql_config"))
	if configField == nil {
		return nil
	}
	prefix := "config_spec." + string(configField.Name()) + "."
	if !slices.ContainsFunc(request.GetUpdateMask().GetPaths(), func(path string) bool { return strings.HasPrefix(path, prefix) }) {
		return nil
	}

	clusterConfig := current.ProtoReflect()
	configSetField := clusterConfig.Descriptor().Fields().ByName(configField.Name())
	if configSetField == nil || !clusterConfig.Has(configSetField) {
		return fmt.Errorf("cluster response after upgrade has no %s", configField.Name())
	}
	configSet := clusterConfig.Get(configSetField).Message()
	actual := configSet.Get(configSet.Descriptor().Fields().ByName("user_config")).Message()
	desired := spec.Get(configField).Message()

	paths := make([]string, 0, len(request.UpdateMask.Paths))
	for _, path := range request.UpdateMask.Paths {
		name, setting := strings.CutPrefix(path, prefix)
		if !setting {
			paths = append(paths, path)
			continue
		}
		field := desired.Descriptor().Fields().ByName(protoreflect.Name(name))
		if field == nil {
			return fmt.Errorf("unsupported MySQL setting in update mask: %s", path)
		}
		equal := desired.Has(field) == actual.Has(field) && desired.Get(field).Equal(actual.Get(field))
		if name == "sql_mode" {
			// SQL modes form a set; the API may return them in another order.
			equal = maps.Equal(sqlModesSet(desired.Get(field).List()), sqlModesSet(actual.Get(field).List()))
		}
		if !equal {
			paths = append(paths, path)
		}
	}
	request.UpdateMask.Paths = paths
	return nil
}

func sqlModesSet(modes protoreflect.List) map[protoreflect.EnumNumber]struct{} {
	result := make(map[protoreflect.EnumNumber]struct{}, modes.Len())
	for i := 0; i < modes.Len(); i++ {
		result[modes.Get(i).Enum()] = struct{}{}
	}
	return result
}
