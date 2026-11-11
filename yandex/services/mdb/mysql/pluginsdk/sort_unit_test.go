package pluginsdk_test

import (
	"testing"

	svc_mysql "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mysql/pluginsdk"
)

func TestSortInterfaceListByTemplate(t *testing.T) {

	name := "some_key"

	listToSort := []interface{}{
		map[string]interface{}{name: "a"},
		map[string]interface{}{name: "d"},
		map[string]interface{}{name: "b"},
		map[string]interface{}{name: "c"},
		map[string]interface{}{name: "h"},
	}
	templateList := []interface{}{
		map[string]interface{}{name: "b"},
		map[string]interface{}{name: "c"},
		map[string]interface{}{name: "d"},
		map[string]interface{}{name: "e"},
	}

	checkList := []string{"b", "c", "d", "a", "h"}

	svc_mysql.SortInterfaceListByTemplate(listToSort, templateList, name)

	for i, v := range checkList {
		if svc_mysql.GetField(listToSort[i], name) != v {
			t.Errorf("sortInterfaceListByTemplate: after sort %v value should be \"%v\" but value is \"%v\"", i, v, svc_mysql.GetField(listToSort[i], name))
		}
	}
}

func TestSortInterfaceListByTemplateNoIntersection(t *testing.T) {

	name := "some_key"

	listToSort := []interface{}{
		map[string]interface{}{name: "a"},
		map[string]interface{}{name: "d"},
		map[string]interface{}{name: "b"},
		map[string]interface{}{name: "c"},
		map[string]interface{}{name: "h"},
	}
	templateList := []interface{}{
		map[string]interface{}{name: "m"},
		map[string]interface{}{name: "n"},
		map[string]interface{}{name: "o"},
		map[string]interface{}{name: "p"},
	}

	checkList := []string{"a", "b", "c", "d", "h"}

	svc_mysql.SortInterfaceListByTemplate(listToSort, templateList, name)

	for i, v := range checkList {
		if svc_mysql.GetField(listToSort[i], name) != v {
			t.Errorf("sortInterfaceListByTemplate: after sort %v value should be \"%v\" but value is \"%v\"", i, v, svc_mysql.GetField(listToSort[i], name))
		}
	}
}

func TestSortInterfaceListByTemplateEmptyTemplate(t *testing.T) {

	name := "some_key"

	listToSort := []interface{}{
		map[string]interface{}{name: "a"},
		map[string]interface{}{name: "d"},
		map[string]interface{}{name: "b"},
		map[string]interface{}{name: "c"},
		map[string]interface{}{name: "h"},
	}
	templateList := []interface{}{}

	// NO sorting
	checkList := []string{"a", "d", "b", "c", "h"}

	svc_mysql.SortInterfaceListByTemplate(listToSort, templateList, name)

	for i, v := range checkList {
		if svc_mysql.GetField(listToSort[i], name) != v {
			t.Errorf("sortInterfaceListByTemplate: after sort %v value should be \"%v\" but value is \"%v\"", i, v, svc_mysql.GetField(listToSort[i], name))
		}
	}
}
