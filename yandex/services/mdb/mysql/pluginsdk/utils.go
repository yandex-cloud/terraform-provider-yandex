package pluginsdk

import (
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func sortInterfaceListByResourceData(listToSort []interface{}, d *schema.ResourceData, entityName string, cmpFieldName string) {
	templateList, ok := d.GetOk(entityName)
	if !ok || templateList == nil {
		return
	}
	SortInterfaceListByTemplate(listToSort, templateList.([]interface{}), cmpFieldName)
}

func SortInterfaceListByTemplate(listToSort []interface{}, templateList []interface{}, cmpFieldName string) {
	if len(templateList) == 0 || len(listToSort) == 0 {
		return
	}

	sortRule := map[string]int{}

	for i := range templateList {
		sortRule[GetField(templateList[i], cmpFieldName)] = i
	}

	sort.Slice(listToSort, func(i int, j int) bool {
		return lessInterfaceList(listToSort, cmpFieldName, i, j, sortRule)
	})
}

func lessInterfaceList(list []interface{}, name string, i int, j int, sortRule map[string]int) bool {
	nameI := GetField(list[i], name)
	nameJ := GetField(list[j], name)

	posI, okI := sortRule[nameI]
	posJ, okJ := sortRule[nameJ]

	if okI && okJ {
		return posI < posJ
	}

	if okI {
		return true
	}

	if okJ {
		return false
	}

	return nameI < nameJ
}

func GetField(value interface{}, field string) string {
	return (value.(map[string]interface{}))[field].(string)
}
