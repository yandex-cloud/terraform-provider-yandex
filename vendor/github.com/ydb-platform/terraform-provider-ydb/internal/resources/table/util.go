package table

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func isIntColumn(typ string) bool {
	return typ == "Int8" || typ == "Int16" || typ == "Int32" || typ == "Int64"
}

func isUintColumn(typ string) bool {
	return typ == "Uint8" || typ == "Uint16" || typ == "Uint32" || typ == "Uint64"
}

func isBoolColumn(typ string) bool {
	return typ == "Bool"
}

func isFloatColumn(typ string) bool {
	return typ == "Float" || typ == "Decimal" || typ == "Double"
}

func isUTF8Column(typ string) bool {
	return typ == "Utf8"
}

func isStringColumn(typ string) bool {
	return typ == "String" || typ == "Bytes" || typ == "Optional<String>"
}

func parsePartitionKey(k string, typ string) (interface{}, error) {
	if isIntColumn(typ) {
		return strconv.ParseInt(k, 10, 64)
	}
	if isUintColumn(typ) {
		return strconv.ParseUint(k, 10, 64)
	}
	if isFloatColumn(typ) {
		return strconv.ParseFloat(k, 64)
	}
	if isStringColumn(typ) || isUTF8Column(typ) {
		return k, nil
	}
	if isBoolColumn(typ) {
		return strconv.ParseBool(k)
	}
	return nil, fmt.Errorf("unknown column type %q", typ)
}

func expandColumns(cols interface{}) []*Column {
	columnsRaw := cols.(*schema.Set)
	columns := make([]*Column, 0, len(columnsRaw.List()))
	for _, v := range columnsRaw.List() {
		mp := v.(map[string]interface{})
		family := ""
		if f, ok := mp["family"].(string); ok {
			family = f
		}
		col := &Column{
			Name:   mp["name"].(string),
			Type:   mp["type"].(string),
			Family: family,
		}
		if notNull, ok := mp["not_null"]; ok {
			col.NotNull = notNull.(bool)
		}
		columns = append(columns, col)
	}

	return columns
}

func expandPrimaryKey(d *schema.ResourceData) []string {
	pkRaw := d.Get("primary_key").([]interface{})
	pk := make([]string, 0, len(pkRaw))
	for _, v := range pkRaw {
		pk = append(pk, v.(string))
	}
	return pk
}

func expandColumnFamilies(d *schema.ResourceData) []*Family {
	familiesRaw := d.Get("family")
	if familiesRaw == nil {
		return nil
	}

	raw := familiesRaw.([]interface{})
	families := make([]*Family, 0, len(raw))
	for _, rw := range raw {
		r := rw.(map[string]interface{})
		name := r["name"].(string)
		data := r["data"].(string)
		compression := r["compression"].(string)
		families = append(families, &Family{
			Name:        name,
			Data:        data,
			Compression: compression,
		})
	}

	return families
}

func expandAttributes(d *schema.ResourceData) map[string]string {
	attributesRaw := d.Get("attributes")
	attributes := make(map[string]string)
	if attributesRaw == nil {
		return attributes
	}

	// TODO(shmel1k@): think about sorting.
	raw := attributesRaw.(map[string]interface{})
	for k, v := range raw {
		attributes[k] = v.(string)
	}
	return attributes
}

// https://ydb.tech/docs/en/yql/reference/builtins/basic#syntax
func ttlToISO8601(ttl time.Duration) string {
	const (
		day  = 24 * time.Hour
		week = 7 * day
	)

	if ttl == 0 {
		return ""
	}

	result := make([]byte, 0, 16)
	result = append(result, "P"...)
	w := ttl / week
	if w > 0 {
		result = strconv.AppendInt(result, int64(w), 10)
		result = append(result, 'W')
	}

	ttl %= week
	d := ttl / day
	if d > 0 {
		result = strconv.AppendInt(result, int64(d), 10)
		result = append(result, 'D')
	}
	ttl %= day

	hasT := false
	h := ttl / time.Hour
	if h > 0 {
		hasT = true
		result = append(result, 'T')
		result = strconv.AppendInt(result, int64(h), 10)
		result = append(result, 'H')
	}
	ttl %= time.Hour

	minutes := ttl / time.Minute
	if minutes > 0 {
		if !hasT {
			result = append(result, 'T')
		}
		hasT = true
		result = strconv.AppendInt(result, int64(minutes), 10)
		result = append(result, 'M')
	}
	ttl %= time.Minute

	sec := ttl / time.Second
	if sec > 0 {
		if !hasT {
			result = append(result, 'T')
		}
		hasT = true
		_ = hasT
		result = strconv.AppendInt(result, int64(sec), 10)
		result = append(result, 'S')
	}
	return string(result)
}
