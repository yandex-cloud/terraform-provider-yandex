package table

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/senseyeio/duration"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"

	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers"
)

type Column struct {
	Name    string
	Type    string
	Family  string
	NotNull bool
}

func (c *Column) ToYQL() string {
	buf := make([]byte, 0, 128)
	buf = append(buf, '`')
	buf = helpers.AppendWithEscape(buf, c.Name)
	buf = append(buf, '`')
	buf = append(buf, ' ')
	buf = helpers.AppendWithEscape(buf, c.Type)
	if c.Family != "" {
		buf = append(buf, ' ')
		buf = append(buf, "FAMILY "...)
		buf = append(buf, '`')
		buf = helpers.AppendWithEscape(buf, c.Family)
		buf = append(buf, '`')
	}
	if c.NotNull {
		buf = append(buf, " NOT NULL"...)
	}
	return string(buf)
}

type PrimaryKey struct {
	Columns []string
}

type Index struct {
	Name    string
	Type    string
	Columns []string
	Cover   []string
}

type TTL struct {
	ColumnName     string
	ExpireInterval string
	Unit           string
}

func (t *TTL) ToYQL() string {
	buf := make([]byte, 0, 64)
	buf = append(buf, "TTL = Interval(\""...)
	buf = helpers.AppendWithEscape(buf, t.ExpireInterval)
	buf = append(buf, '"')
	buf = append(buf, ')')
	buf = append(buf, " ON "...)
	buf = append(buf, '`')
	buf = helpers.AppendWithEscape(buf, t.ColumnName)
	buf = append(buf, '`')
	if t.Unit != "" {
		buf = append(buf, " AS "...)
		buf = append(buf, t.Unit...)
	}
	return string(buf)
}

type PartitionAtKeys struct {
	Keys []interface{}
}

type PartitioningSettings struct {
	PartitionBy        *PrimaryKey
	BySize             *bool
	ByLoad             *bool
	PartitionSizeMb    *int
	PartitionAtKeys    []*PartitionAtKeys
	PartitionsCount    int
	MinPartitionsCount int
	MaxPartitionsCount int
}

type ReplicationSettings struct {
	ReadReplicasSettings string
}

type Family struct {
	Name        string
	Data        string
	Compression string
}

type Resource struct {
	Entity *helpers.YDBEntity

	FullPath             string
	Path                 string
	DatabaseEndpoint     string
	Attributes           map[string]string
	Family               []*Family
	Columns              []*Column
	PrimaryKey           *PrimaryKey
	TTL                  *TTL
	ReplicationSettings  *ReplicationSettings
	PartitioningSettings *PartitioningSettings
	EnableBloomFilter    *bool
	StoreType            options.StoreType
}

func (r *Resource) getConnectionString() string {
	if r.DatabaseEndpoint != "" {
		return r.DatabaseEndpoint
	}
	return r.Entity.PrepareFullYDBEndpoint()
}

func (r *Resource) isStoreNeeded() bool {
	return r.StoreType != options.StoreTypeUnspecified
}

func (r *Resource) storeYQLStmt() string {
	switch r.StoreType {
	case options.StoreTypeRow:
		return "ROW"
	case options.StoreTypeColumn:
		return "COLUMN"
	}

	return ""
}

func expandTableTTLSettings(d *schema.ResourceData) (ttl *TTL) {
	v, ok := d.GetOk("ttl")
	if !ok {
		return
	}
	ttlSet := v.(*schema.Set)
	for _, l := range ttlSet.List() {
		m := l.(map[string]interface{})
		ttl = &TTL{}
		ttl.ColumnName = m["column_name"].(string)
		//		ttl.Mode = m["mode"].(string)
		ttl.ExpireInterval = m["expire_interval"].(string)
		ttl.Unit = m["unit"].(string)
	}
	return
}

func expandTableReplicasSettings(d *schema.ResourceData) (p *ReplicationSettings) {
	v, ok := d.GetOk("read_replicas_settings")
	if !ok {
		return
	}

	p = &ReplicationSettings{}
	p.ReadReplicasSettings = v.(string)
	return
}

func expandPartitionAtKeys(p []interface{}, primaryKeyColumns []*Column) ([]*PartitionAtKeys, error) {
	if len(p) == 0 || len(primaryKeyColumns) == 0 {
		return nil, nil
	}

	res := make([]*PartitionAtKeys, 0, len(p))
	for _, v := range p {
		vv := v.(map[string]interface{})
		keys := vv["keys"].([]interface{})
		pp := &PartitionAtKeys{}
		for i, k := range keys {
			if i == len(primaryKeyColumns) {
				return nil, fmt.Errorf("can not be more partition keys than primary key columns")
			}
			got, err := parsePartitionKey(k.(string), primaryKeyColumns[i].Type)
			if err != nil {
				return nil, err
			}
			pp.Keys = append(pp.Keys, got)
		}
		res = append(res, pp)
	}
	return res, nil
}

func expandTablePartitioningPolicySettings(d *schema.ResourceData, columns []*Column, primaryKeyColumns []string) (p *PartitioningSettings, err error) {
	v, ok := d.GetOk("partitioning_settings")
	if !ok {
		return
	}

	p = &PartitioningSettings{}

	pk := make(map[string]struct{})
	for _, v := range primaryKeyColumns {
		pk[v] = struct{}{}
	}

	primaryKeyCols := make([]*Column, 0, len(primaryKeyColumns))
	for _, v := range columns {
		if _, ok := pk[v.Name]; ok {
			primaryKeyCols = append(primaryKeyCols, v)
		}
	}

	pList := v.([]interface{})
	for _, l := range pList {
		m := l.(map[string]interface{})
		if partitionsCount, ok := m["uniform_partitions"].(int); ok && partitionsCount != 0 {
			p.PartitionsCount = partitionsCount
		}
		if explicitPartitions, ok := m["partition_at_keys"].([]interface{}); ok {
			p.PartitionAtKeys, err = expandPartitionAtKeys(explicitPartitions, primaryKeyCols)
			if err != nil {
				return nil, err
			}
		}
		if minPartitionsCount, ok := m["auto_partitioning_min_partitions_count"].(int); ok && minPartitionsCount != 0 {
			p.MinPartitionsCount = minPartitionsCount
		}
		if maxPartitionsCount, ok := m["auto_partitioning_max_partitions_count"].(int); ok && maxPartitionsCount != 0 {
			p.MaxPartitionsCount = maxPartitionsCount
		}
		if byLoad, ok := m["auto_partitioning_by_load"].(bool); ok {
			p.ByLoad = &byLoad
		}
		if bySize, ok := m["auto_partitioning_by_size_enabled"].(bool); ok {
			p.BySize = &bySize
		}
		if partitionSizeMb, ok := m["auto_partitioning_partition_size_mb"].(int); ok && partitionSizeMb != 0 {
			p.PartitionSizeMb = &partitionSizeMb
		}
		if partitionBy, ok := m["partition_by"].([]any); ok {
			p.PartitionBy = &PrimaryKey{
				Columns: make([]string, 0, len(partitionBy)),
			}
			for _, v := range partitionBy {
				p.PartitionBy.Columns = append(p.PartitionBy.Columns, v.(string))
			}
		}
	}

	return p, nil
}

func expandTableStore(d *schema.ResourceData) options.StoreType {
	switch d.Get("store") {
	case "column":
		return options.StoreTypeColumn
	case "row":
		return options.StoreTypeRow
	default:
		return options.StoreTypeUnspecified
	}
}

func tableResourceSchemaToTableResource(d *schema.ResourceData) (*Resource, error) {
	var entity *helpers.YDBEntity
	var err error
	if d.Id() != "" {
		entity, err = helpers.ParseYDBEntityID(d.Id())
		if err != nil {
			return nil, fmt.Errorf("failed to parse table entity: %w", err)
		}
	}

	columns := expandColumns(d.Get("column"))
	pk := expandPrimaryKey(d)
	families := expandColumnFamilies(d)
	attributes := expandAttributes(d)
	ttl := expandTableTTLSettings(d)

	databaseEndpoint := d.Get("connection_string").(string)
	databaseURL, err := url.Parse(databaseEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database endpoint: %w", err)
	}

	partitioningSettings, err := expandTablePartitioningPolicySettings(d, columns, pk)
	if err != nil {
		return nil, fmt.Errorf("failed to expand table partitioning settings: %w", err)
	}

	replicasSettings := expandTableReplicasSettings(d)

	var bloomFilterEnabled *bool
	if v, ok := d.GetOk("key_bloom_filter"); ok {
		b := v.(bool)
		bloomFilterEnabled = &b
	}

	var path string
	if entity != nil {
		path = entity.GetEntityPath()
		databaseEndpoint = entity.PrepareFullYDBEndpoint()
		path = databaseEndpoint + "/" + path
	} else {
		path = databaseURL.Query().Get("database") + "/" + d.Get("path").(string)
		databaseEndpoint = d.Get("connection_string").(string)
	}

	return &Resource{
		Entity:           entity,
		FullPath:         path,
		Path:             helpers.TrimPath(d.Get("path").(string)),
		DatabaseEndpoint: databaseEndpoint,
		Attributes:       attributes,
		Family:           families,
		// ChangeFeeds:      cdcSettings,
		Columns: columns,
		PrimaryKey: &PrimaryKey{
			Columns: pk,
		},
		TTL:                  ttl,
		PartitioningSettings: partitioningSettings,
		ReplicationSettings:  replicasSettings,
		EnableBloomFilter:    bloomFilterEnabled,
		StoreType:            expandTableStore(d),
	}, nil
}

func flattenTablePartitioningSettings(d *schema.ResourceData, settings options.PartitioningSettings) []interface{} {
	output := make([]interface{}, 0, 1)
	partitioningSettings := make(map[string]interface{})
	partitioningSettings["auto_partitioning_by_load"] = settings.PartitioningByLoad == options.FeatureEnabled
	partitioningSettings["auto_partitioning_by_size_enabled"] = settings.PartitioningBySize == options.FeatureEnabled
	partitioningSettings["auto_partitioning_partition_size_mb"] = settings.PartitionSizeMb
	partitioningSettings["auto_partitioning_min_partitions_count"] = settings.MinPartitionsCount
	partitioningSettings["auto_partitioning_max_partitions_count"] = settings.MaxPartitionsCount
	partitioningSettings["partition_by"] = settings.PartitionBy
	pList := d.Get("partitioning_settings").([]interface{})
	for _, l := range pList {
		m := l.(map[string]interface{})
		partitioningSettings["partition_at_keys"] = m["partition_at_keys"]
		partitioningSettings["uniform_partitions"] = m["uniform_partitions"]
	}

	output = append(output, partitioningSettings)
	return output
}

func unwrapType(t types.Type) (typ string, notNull bool) {
	yqlStr := t.Yql()
	notNull = true

	if strings.HasPrefix(yqlStr, "Optional<") {
		notNull = false
		yqlStr = strings.TrimPrefix(yqlStr, "Optional<")
		yqlStr = strings.TrimSuffix(yqlStr, ">")
	}

	typ = yqlStr

	// need to discuss type migration
	/*if typ == "String" { //nolint
		typ = "Bytes" //nolint
	}*/

	return typ, notNull
}

func flattenTableDescription(d *schema.ResourceData, desc options.Description, entity *helpers.YDBEntity) (err error) {
	err = d.Set("path", entity.GetEntityPath())
	if err != nil {
		return
	}
	err = d.Set("connection_string", entity.PrepareFullYDBEndpoint())
	if err != nil {
		return
	}

	cols := make([]interface{}, 0, len(desc.Columns))
	for _, col := range desc.Columns {
		mp := make(map[string]interface{})
		mp["name"] = col.Name
		mp["type"], mp["not_null"] = unwrapType(col.Type)
		mp["family"] = col.Family
		cols = append(cols, mp)
	}
	err = d.Set("column", cols)
	if err != nil {
		return
	}

	pk := make([]interface{}, 0, len(desc.PrimaryKey))
	for _, p := range desc.PrimaryKey {
		pk = append(pk, p)
	}
	err = d.Set("primary_key", pk)
	if err != nil {
		return
	}

	var store string
	switch desc.StoreType {
	case options.StoreTypeRow:
		store = "row"
	case options.StoreTypeColumn:
		store = "column"
	}
	err = d.Set("store", store)
	if err != nil {
		return
	}

	if desc.TimeToLiveSettings != nil {
		var ttlSettings []interface{}

		// for explaine some variations "zero" interval to ISO8601
		v, ok := d.GetOk("ttl")
		interval := ttlToISO8601(time.Duration(desc.TimeToLiveSettings.ExpireAfterSeconds) * time.Second)
		if interval == "" && ok {
			ttl := v.(*schema.Set)
			// only one ttl
			ttlOpts := ttl.List()[0].(map[string]interface{})
			interval = ttlOpts["expire_interval"].(string)
			d, _ := duration.ParseISO8601(interval)
			if !d.IsZero() {
				interval = ""
			}
		}

		ttlSettings = append(ttlSettings, map[string]interface{}{
			"column_name":     desc.TimeToLiveSettings.ColumnName,
			"expire_interval": interval,
			"unit":            helpers.YDBUnitToUnit(desc.TimeToLiveSettings.ColumnUnit.ToYDB().String()),
		})
		err = d.Set("ttl", ttlSettings)
		if err != nil {
			return
		}
	}

	attributes := make(map[string]interface{})
	for k, v := range desc.Attributes {
		attributes[k] = v
	}
	err = d.Set("attributes", attributes)
	if err != nil {
		return
	}
	err = d.Set("partitioning_settings", flattenTablePartitioningSettings(d, desc.PartitioningSettings))
	if err != nil {
		return
	}

	err = d.Set("key_bloom_filter", desc.KeyBloomFilter == options.FeatureEnabled)
	if err != nil {
		return
	}
	if desc.ReadReplicaSettings.Type == options.ReadReplicasAnyAzReadReplicas {
		err = d.Set("read_replicas_settings", fmt.Sprintf("ANY_AZ:%d", desc.ReadReplicaSettings.Count))
	} else {
		err = d.Set("read_replicas_settings", fmt.Sprintf("PER_AZ:%d", desc.ReadReplicaSettings.Count))
	}
	return err
}
