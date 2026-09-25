package table

import (
	"bytes"
	"strconv"

	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers"
)

const (
	defaultRequestCapacity = 1024 // 1 KiB
)

func appendIndent(req []byte, indent int) []byte {
	req = append(req, bytes.Repeat([]byte{'\t'}, indent)...)
	return req
}

func PrepareCreateRequest(r *Resource) string { //nolint:gocyclo
	req := make([]byte, 0, defaultRequestCapacity)

	req = append(req, "CREATE TABLE `"...)
	req = helpers.AppendWithEscape(req, r.FullPath)
	req = append(req, "`("...)
	req = append(req, '\n')

	indent := 1
	for _, v := range r.Columns {
		req = appendIndent(req, indent)
		req = append(req, v.ToYQL()...)
		req = append(req, ',')
		req = append(req, '\n')
	}

	req = appendIndent(req, indent)
	req = append(req, "PRIMARY KEY"...)
	req = append(req, ' ')
	req = append(req, '(')
	for _, v := range r.PrimaryKey.Columns {
		req = append(req, '`')
		req = helpers.AppendWithEscape(req, v)
		req = append(req, '`')
		req = append(req, ',')
	}
	req[len(req)-1] = ')'
	req = append(req, '\n')
	if len(r.Family) > 0 {
		req[len(req)-1] = ','
		for _, v := range r.Family {
			req = append(req, '\n')
			req = appendIndent(req, indent)
			req = append(req, "FAMILY"...)
			req = append(req, ' ')
			req = append(req, '`')
			req = helpers.AppendWithEscape(req, v.Name)
			req = append(req, '`')
			req = append(req, '(')
			req = append(req, '\n')
			indent++
			req = appendIndent(req, indent)
			req = append(req, "DATA = "...)
			req = append(req, '"')
			req = helpers.AppendWithEscape(req, v.Data)
			req = append(req, '"')
			req = append(req, ',')
			req = append(req, '\n')
			req = appendIndent(req, indent)
			req = append(req, "COMPRESSION = "...)
			req = append(req, '"')
			req = helpers.AppendWithEscape(req, v.Compression)
			req = append(req, '"')
			req = append(req, '\n')
			indent--
			req = appendIndent(req, indent)
			req = append(req, ')')
			req = append(req, ',')
		}
		req[len(req)-1] = '\n'
	}
	req = append(req, ')')
	req = append(req, '\n')
	indent--

	// PARTITION BY HASH
	if r.PartitioningSettings != nil &&
		r.PartitioningSettings.PartitionBy != nil &&
		len(r.PartitioningSettings.PartitionBy.Columns) > 0 {
		req = appendIndent(req, indent)
		req = append(req, "PARTITION BY HASH"...)
		req = append(req, ' ')
		req = append(req, '(')
		for _, v := range r.PartitioningSettings.PartitionBy.Columns {
			req = append(req, '`')
			req = helpers.AppendWithEscape(req, v)
			req = append(req, '`')
			req = append(req, ',')
		}
		req[len(req)-1] = ')'
		req = append(req, '\n')
	}

	needWith := r.TTL != nil
	if len(r.Attributes) != 0 {
		needWith = true
	}
	if r.PartitioningSettings != nil {
		needWith = true
	}
	if r.ReplicationSettings != nil {
		needWith = true
	}
	if r.isStoreNeeded() {
		needWith = true
	}

	if !needWith {
		return string(req)
	}

	req = append(req, "WITH"...)
	req = append(req, ' ', '(', '\n')
	indent++
	needComma := false

	if r.isStoreNeeded() {
		if needComma {
			req = append(req, ',', '\n')
		}
		req = appendIndent(req, indent)
		req = append(req, "STORE = "...)
		req = append(req, r.storeYQLStmt()...)
		needComma = true
	}

	if r.TTL != nil {
		req = appendIndent(req, indent)
		req = append(req, "TTL = Interval(\""...)
		req = helpers.AppendWithEscape(req, r.TTL.ExpireInterval)
		req = append(req, '"')
		req = append(req, ')')
		req = append(req, " ON "...)
		req = append(req, '`')
		req = helpers.AppendWithEscape(req, r.TTL.ColumnName)
		req = append(req, '`')
		if r.TTL.Unit != "" {
			req = append(req, " AS "...)
			req = append(req, r.TTL.Unit...)
		}
		needComma = true
	}
	if r.PartitioningSettings != nil { //nolint:nestif
		if r.PartitioningSettings.ByLoad != nil {
			if needComma {
				req = append(req, ',', '\n')
			}
			req = appendIndent(req, indent)
			if *r.PartitioningSettings.ByLoad {
				req = append(req, "AUTO_PARTITIONING_BY_LOAD = ENABLED"...)
			} else {
				req = append(req, "AUTO_PARTITIONING_BY_LOAD = DISABLED"...)
			}
			needComma = true
		}
		if r.PartitioningSettings.BySize != nil {
			if needComma {
				req = append(req, ',', '\n')
			}
			req = appendIndent(req, indent)
			if *r.PartitioningSettings.BySize {
				req = append(req, "AUTO_PARTITIONING_BY_SIZE = ENABLED"...)
				if r.PartitioningSettings.PartitionSizeMb != nil {
					req = append(req, ',')
					req = append(req, '\n')
					req = appendIndent(req, indent)
					req = append(req, "AUTO_PARTITIONING_PARTITION_SIZE_MB = "...)
					req = strconv.AppendInt(req, int64(*r.PartitioningSettings.PartitionSizeMb), 10)
				}
			} else {
				req = append(req, "AUTO_PARTITIONING_BY_SIZE = DISABLED"...)
			}
			needComma = true
		}
		if r.PartitioningSettings.PartitionsCount != 0 {
			if needComma {
				req = append(req, ',', '\n')
			}
			req = appendIndent(req, indent)
			req = append(req, "UNIFORM_PARTITIONS = "...)
			req = strconv.AppendInt(req, int64(r.PartitioningSettings.PartitionsCount), 10)
			needComma = true
		}
		if len(r.PartitioningSettings.PartitionAtKeys) != 0 {
			if needComma {
				req = append(req, ',', '\n')
			}
			req = appendIndent(req, indent)
			req = append(req, "PARTITION_AT_KEYS = "...)
			if len(r.PartitioningSettings.PartitionAtKeys) > 1 {
				req = append(req, '(')
			}
			for i, v := range r.PartitioningSettings.PartitionAtKeys {
				req = append(req, '(')
				for ii, vv := range v.Keys {
					switch t := vv.(type) {
					case uint64:
						req = strconv.AppendUint(req, t, 10)
					case int64:
						req = strconv.AppendInt(req, t, 10)
					case bool:
						req = strconv.AppendBool(req, t)
					case string:
						req = append(req, '"')
						req = helpers.AppendWithEscape(req, t)
						req = append(req, '"')
					}
					if ii < len(v.Keys)-1 {
						req = append(req, ',')
					}
				}
				req = append(req, ')')
				if i < len(r.PartitioningSettings.PartitionAtKeys)-1 {
					req = append(req, ',')
				}
			}
			if len(r.PartitioningSettings.PartitionAtKeys) > 1 {
				req = append(req, ')')
			}
			needComma = true
		}
		if r.PartitioningSettings.MinPartitionsCount != 0 {
			if needComma {
				req = append(req, ',', '\n')
			}
			req = appendIndent(req, indent)
			req = append(req, "AUTO_PARTITIONING_MIN_PARTITIONS_COUNT = "...)
			req = strconv.AppendInt(req, int64(r.PartitioningSettings.MinPartitionsCount), 10)
			needComma = true
		}
		if r.PartitioningSettings.MaxPartitionsCount != 0 {
			if needComma {
				req = append(req, ',', '\n')
			}
			req = appendIndent(req, indent)
			req = append(req, "AUTO_PARTITIONING_MAX_PARTITIONS_COUNT = "...)
			req = strconv.AppendInt(req, int64(r.PartitioningSettings.MaxPartitionsCount), 10)
			needComma = true
		}
	}
	if r.ReplicationSettings != nil && r.ReplicationSettings.ReadReplicasSettings != "" {
		if needComma {
			req = append(req, ',', '\n')
		}
		req = appendIndent(req, indent)
		req = append(req, "READ_REPLICAS_SETTINGS = \""...)
		req = helpers.AppendWithEscape(req, r.ReplicationSettings.ReadReplicasSettings)
		req = append(req, '"')
		needComma = true
	}
	if r.EnableBloomFilter != nil {
		if needComma {
			req = append(req, ',', '\n')
		}
		needComma = true
		req = appendIndent(req, indent)
		req = append(req, "KEY_BLOOM_FILTER = "...)
		if *r.EnableBloomFilter {
			req = append(req, "ENABLED"...)
		} else {
			req = append(req, "DISABLED"...)
		}
	}

	// indent--
	_ = needComma

	req = append(req, '\n', ')')

	//	if len(r.ChangeFeeds) > 0 {
	//		req = append(req, ';', '\n')
	//		req = append(req, prepareCDCAlterQuery(r.Path, r.ChangeFeeds)...)
	//	}

	return string(req)
}

func prepareAddColumnsQuery(tableName string, columnsToAdd []*Column) string {
	req := []byte("ALTER TABLE `")
	req = helpers.AppendWithEscape(req, tableName)
	req = append(req, '`', ' ')
	for i := 0; i < len(columnsToAdd); i++ {
		req = append(req, "ADD COLUMN "...)
		req = append(req, columnsToAdd[i].ToYQL()...)
		if i != len(columnsToAdd)-1 {
			req = append(req, ',', ' ')
		}
	}

	return string(req)
}

func prepareDropColumnsQuery(tableName string, columnsToDrop []string) string {
	req := make([]byte, 0, defaultRequestCapacity)
	req = append(req, "ALTER TABLE `"...)
	req = helpers.AppendWithEscape(req, tableName)
	req = append(req, '`', ' ')
	for i := 0; i < len(columnsToDrop); i++ {
		req = append(req, "DROP COLUMN `"...)
		req = helpers.AppendWithEscape(req, columnsToDrop[i])
		req = append(req, '`')
		if i != len(columnsToDrop)-1 {
			req = append(req, ',', ' ')
		}
	}

	return string(req)
}

func prepareResetTTLQuery(tableName string) string {
	buf := make([]byte, 0, 64)
	buf = append(buf, "ALTER TABLE `"...)
	buf = helpers.AppendWithEscape(buf, tableName)
	buf = append(buf, '`', ' ')
	buf = append(buf, "RESET (TTL)"...)
	return string(buf)
}

func prepareSetNewTTLSettingsQuery(tableName string, settings *TTL) string {
	buf := make([]byte, 0, 64)
	buf = append(buf, "ALTER TABLE `"...)
	buf = helpers.AppendWithEscape(buf, tableName)
	buf = append(buf, '`', ' ')
	buf = append(buf, "SET ("...)
	buf = append(buf, settings.ToYQL()...)
	buf = append(buf, ')')
	return string(buf)
}

func prepareKeyBloomFilterQuery(tableName string, enabled bool) string {
	buf := make([]byte, 0, 64)
	buf = append(buf, "ALTER TABLE `"...)
	buf = helpers.AppendWithEscape(buf, tableName)
	buf = append(buf, '`', ' ')
	buf = append(buf, "SET (\n"...)
	buf = append(buf, "KEY_BLOOM_FILTER = "...)
	if enabled {
		buf = append(buf, "ENABLED"...)
	} else {
		buf = append(buf, "DISABLED"...)
	}
	buf = append(buf, ')')
	return string(buf)
}

func prepareNewPartitioningSettingsQuery(
	tableName string,
	settings *PartitioningSettings,
	readReplicaSettings string,
) string {
	buf := make([]byte, 0, 64)
	buf = append(buf, "ALTER TABLE `"...)
	buf = helpers.AppendWithEscape(buf, tableName)
	buf = append(buf, '`', ' ')
	buf = append(buf, "SET (\n"...)

	// TODO(shmel1k@): remove copypaste.
	needComma := false
	if settings != nil && settings.ByLoad != nil {
		buf = append(buf, "AUTO_PARTITIONING_BY_LOAD = "...)
		val := *settings.ByLoad
		if val {
			buf = append(buf, "ENABLED"...)
		} else {
			buf = append(buf, "DISABLED"...)
		}
		needComma = true
	}
	if settings != nil && settings.BySize != nil { //nolint:nestif
		if needComma {
			buf = append(buf, ',', '\n')
		}
		needComma = true
		if *settings.BySize {
			buf = append(buf, "AUTO_PARTITIONING_BY_SIZE = ENABLED"...)
			if settings.PartitionSizeMb != nil {
				buf = append(buf, ',', '\n')
				buf = append(buf, "AUTO_PARTITIONING_PARTITION_SIZE_MB = "...)
				buf = strconv.AppendInt(buf, int64(*settings.PartitionSizeMb), 10)
			}
		} else {
			buf = append(buf, "AUTO_PARTITIONING_BY_SIZE = DISABLED"...)
		}
	}
	if settings != nil && settings.MinPartitionsCount != 0 {
		if needComma {
			buf = append(buf, ',', '\n')
		}
		buf = append(buf, "AUTO_PARTITIONING_MIN_PARTITIONS_COUNT = "...)
		buf = strconv.AppendInt(buf, int64(settings.MinPartitionsCount), 10)
		needComma = true
	}
	if settings != nil && settings.MaxPartitionsCount != 0 {
		if needComma {
			buf = append(buf, ',', '\n')
		}
		buf = append(buf, "AUTO_PARTITIONING_MAX_PARTITIONS_COUNT = "...)
		buf = strconv.AppendInt(buf, int64(settings.MaxPartitionsCount), 10)
	}
	if readReplicaSettings != "" {
		if needComma {
			buf = append(buf, ',', '\n')
		}
		buf = append(buf, "READ_REPLICAS_SETTINGS = \""...)
		buf = helpers.AppendWithEscape(buf, readReplicaSettings)
		buf = append(buf, '"')
	}
	buf = append(buf, '\n', ')')

	return string(buf)
}

func PrepareAlterRequest(diff *tableDiff) string {
	if diff == nil {
		return ""
	}

	req := make([]byte, 0, defaultRequestCapacity)
	needSemiColon := false
	if len(diff.ColumnsToAdd) > 0 {
		if needSemiColon {
			req = append(req, ';', '\n')
		}
		needSemiColon = true
		req = append(req, prepareAddColumnsQuery(diff.TableName, diff.ColumnsToAdd)...)
	}
	if diff.NewTTLSettings != nil {
		if needSemiColon {
			req = append(req, ';', '\n')
		}
		needSemiColon = true
		req = append(req, prepareResetTTLQuery(diff.TableName)...)
		req = append(req, ';', '\n')
		req = append(req, prepareSetNewTTLSettingsQuery(diff.TableName, diff.NewTTLSettings)...)
	}
	if diff.OnlyResetTTL {
		needSemiColon = true
		req = append(req, prepareResetTTLQuery(diff.TableName)...)
	}
	if diff.NewPartitioningSettings != nil || diff.ReadReplicasSettings != "" {
		if needSemiColon {
			req = append(req, ';', '\n')
		}
		req = append(req, prepareNewPartitioningSettingsQuery(diff.TableName, diff.NewPartitioningSettings, diff.ReadReplicasSettings)...)
		needSemiColon = true
	}

	if diff.NewKeyBloomFilterSettings != nil {
		if needSemiColon {
			req = append(req, ';', '\n')
		}
		req = append(req, prepareKeyBloomFilterQuery(diff.TableName, *diff.NewKeyBloomFilterSettings)...)
		needSemiColon = true
	}

	_ = needSemiColon

	return string(req)
}

func PrepareDropTableRequest(tableName string) string {
	buf := make([]byte, 0, 64)
	buf = append(buf, "DROP TABLE `"...)
	buf = helpers.AppendWithEscape(buf, tableName)
	buf = append(buf, '`')
	return string(buf)
}
