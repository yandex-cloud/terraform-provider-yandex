package table

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
)

type tableDiff struct {
	TableName                 string
	ColumnsToAdd              []*Column
	NewTTLSettings            *TTL
	NewPartitioningSettings   *PartitioningSettings
	NewKeyBloomFilterSettings *bool
	ReadReplicasSettings      string
	OnlyResetTTL              bool
}

func checkColumnDiff(rcolumns []*Column, dcolumns []*Column) ([]*Column, error) {
	existingColumns := make(map[string]struct{})
	for _, v := range dcolumns {
		existingColumns[v.Name] = struct{}{}
	}
	resourceColumns := make(map[string]*Column)
	for _, v := range rcolumns {
		resourceColumns[v.Name] = v
	}

	var columnsToAdd []*Column
	var deletedColumns []string

	for k := range existingColumns {
		if _, ok := resourceColumns[k]; !ok {
			deletedColumns = append(deletedColumns, k)
		}
	}

	for k, v := range resourceColumns {
		if _, ok := existingColumns[k]; !ok {
			columnsToAdd = append(columnsToAdd, v)
		}
	}

	if len(deletedColumns) > 0 {
		return nil, fmt.Errorf("it is prohibited to delete columns with terraform. Columns for deletion: [%s]", strings.Join(deletedColumns, ","))
	}
	return columnsToAdd, nil
}

func compareIndexes(ridx *Index, didx options.IndexDescription) bool {
	if ridx.Name != didx.Name {
		return false
	}
	if len(ridx.Columns) != len(didx.IndexColumns) {
		return false
	}
	// TODO(shmel1k@): check index type, wait for go-sdk fix.
	for i := 0; i < len(ridx.Columns); i++ {
		if ridx.Columns[i] != didx.IndexColumns[i] {
			return false
		}
	}
	if len(ridx.Cover) != len(didx.DataColumns) {
		return false
	}
	mp1 := make(map[string]struct{})
	for _, v := range ridx.Cover {
		mp1[v] = struct{}{}
	}
	mp2 := make(map[string]struct{})
	for _, v := range didx.DataColumns {
		mp2[v] = struct{}{}
	}

	return reflect.DeepEqual(mp1, mp2)
}

func checkIndexDiff(rindexes []*Index, dindexes []options.IndexDescription) (toDrop []string, toCreate []*Index) {
	existingIndexes := make(map[string]struct{})
	for _, v := range dindexes {
		existingIndexes[v.Name] = struct{}{}
	}

	resourceIndexes := make(map[string]*Index)
	for _, v := range rindexes {
		resourceIndexes[v.Name] = v
	}

	for k := range existingIndexes {
		if _, ok := resourceIndexes[k]; !ok {
			toDrop = append(toDrop, k)
		}
	}

	for k, v := range resourceIndexes {
		if _, ok := existingIndexes[k]; !ok {
			toCreate = append(toCreate, v)
		} else {
			toCreate = append(toCreate, v)
			toDrop = append(toDrop, v.Name)
		}
	}

	return
}

func prepareTableDiff(d *schema.ResourceData) (*tableDiff, error) {
	diff := &tableDiff{}
	if d.HasChange("column") {
		o, n := d.GetChange("column")
		oColumns := expandColumns(o)
		nColumns := expandColumns(n)
		newColumns, err := checkColumnDiff(nColumns, oColumns)
		if err != nil {
			return nil, err
		}
		diff.ColumnsToAdd = newColumns
	}
	if d.HasChange("ttl") {
		diff.NewTTLSettings = expandTableTTLSettings(d)
		if diff.NewTTLSettings == nil {
			diff.OnlyResetTTL = true
		}
	}
	if d.HasChange("partitioning_settings") {
		var err error
		diff.NewPartitioningSettings, err = expandTablePartitioningPolicySettings(d, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to expand new partitioning settings: %w", err)
		}
	}
	if d.HasChange("key_bloom_filter") {
		val := false
		v, ok := d.GetOk("key_bloom_filter")
		if ok {
			val = v.(bool)
		}
		diff.NewKeyBloomFilterSettings = &val
	}
	if d.HasChange("read_replicas_settings") {
		if v, ok := d.GetOk("read_replicas_settings"); ok {
			diff.ReadReplicasSettings = v.(string)
		}
	}

	return diff, nil
}
