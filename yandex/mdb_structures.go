package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/googleapis/type/timeofday"
)

var weeklyMaintenanceWindow_WeekDay_value = map[string]int32{
	"WEEK_DAY_UNSPECIFIED": 0,
	"MON":                  1,
	"TUE":                  2,
	"WED":                  3,
	"THU":                  4,
	"FRI":                  5,
	"SAT":                  6,
	"SUN":                  7,
}

func mdbMaintenanceWindowSchemaValidateFunc(v interface{}, k string) (s []string, es []error) {
	dayString := v.(string)
	day, ok := weeklyMaintenanceWindow_WeekDay_value[dayString]
	if !ok || day == 0 {
		es = append(es, fmt.Errorf(`expected %s value should be one of ("MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"). Current value is %v`, k, v))
		return
	}

	return
}

func flattenMDBBackupWindowStart(t *timeofday.TimeOfDay) []interface{} {
	if t == nil {
		return nil
	}

	out := map[string]interface{}{}

	out["hours"] = int(t.Hours)
	out["minutes"] = int(t.Minutes)

	return []interface{}{out}
}

func expandMDBBackupWindowStart(d *schema.ResourceData, path string) *timeofday.TimeOfDay {
	out := &timeofday.TimeOfDay{}
	hours := fmt.Sprintf(path + ".hours")
	minutes := fmt.Sprintf(path + ".minutes")

	if v, ok := d.GetOk(hours); ok {
		out.Hours = int32(v.(int))
	}

	if v, ok := d.GetOk(minutes); ok {
		out.Minutes = int32(v.(int))
	}

	return out
}
