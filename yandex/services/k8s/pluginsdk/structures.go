package pluginsdk

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/googleapis/type/dayofweek"
	"google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func flattenK8SNodeGroupDNSRecords(specs []*k8s.DnsRecordSpec) []map[string]interface{} {
	res := make([]map[string]interface{}, len(specs))

	for i, spec := range specs {
		res[i] = map[string]interface{}{
			"fqdn":        spec.Fqdn,
			"dns_zone_id": spec.DnsZoneId,
			"ttl":         int(spec.Ttl),
			"ptr":         spec.Ptr,
		}
	}

	return res
}

func expandSubnetIds(v interface{}) []string {
	if v == nil {
		return nil
	}
	var m []string
	subnetIdsSet := v.(*schema.Set)
	for _, val := range subnetIdsSet.List() {
		m = append(m, val.(string))
	}
	return m
}

func expandK8SNodeGroupDNSRecords(data []interface{}) []*k8s.DnsRecordSpec {
	recs := make([]*k8s.DnsRecordSpec, len(data))

	for i, raw := range data {
		d := raw.(map[string]interface{})
		r := &k8s.DnsRecordSpec{Fqdn: d["fqdn"].(string)}
		if s, ok := d["dns_zone_id"]; ok {
			r.DnsZoneId = s.(string)
		}
		if s, ok := d["ttl"]; ok {
			r.Ttl = int64(s.(int))
		}
		if s, ok := d["ptr"]; ok {
			r.Ptr = s.(bool)
		}
		recs[i] = r
	}

	return recs
}

func flattenMaintenanceWindow(mw *k8s.MaintenanceWindow) (*schema.Set, error) {
	maintenanceWindow := &schema.Set{F: dayOfWeekHash}
	if mw != nil {
		switch p := mw.GetPolicy().(type) {
		case *k8s.MaintenanceWindow_Anytime:
			// do nothing
		case *k8s.MaintenanceWindow_DailyMaintenanceWindow:
			dailyPolicy := map[string]interface{}{
				"start_time": formatTimeOfDay(p.DailyMaintenanceWindow.GetStartTime()),
				"duration":   sdkcommon.FormatDuration(p.DailyMaintenanceWindow.GetDuration()),
			}
			maintenanceWindow.Add(dailyPolicy)
		case *k8s.MaintenanceWindow_WeeklyMaintenanceWindow:
			for _, v := range p.WeeklyMaintenanceWindow.GetDaysOfWeek() {
				for _, d := range v.GetDays() {
					dailyPolicy := map[string]interface{}{
						"day":        strings.ToLower(d.String()),
						"start_time": formatTimeOfDay(v.GetStartTime()),
						"duration":   sdkcommon.FormatDuration(v.GetDuration()),
					}
					if maintenanceWindow.Contains(dailyPolicy) {
						return nil, fmt.Errorf("duplication for day of week found in maintenance window")
					}
					maintenanceWindow.Add(dailyPolicy)
				}
			}
		default:
			return nil, fmt.Errorf("unsupported Kubernetes master maintenance policy type")
		}
	}

	return maintenanceWindow, nil
}

func expandMaintenanceWindow(days []interface{}) (*k8s.MaintenanceWindow, error) {
	if len(days) == 0 {
		return nil, nil
	}

	windows := []*dayMaintenanceWindow{}
	parsedDays := map[dayofweek.DayOfWeek]struct{}{}
	dailyWindowSpecified := false

	for _, v := range days {
		window, err := expandDayMaintenanceWindow(v.(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		if window.day == dayofweek.DayOfWeek_DAY_OF_WEEK_UNSPECIFIED {
			dailyWindowSpecified = true
		}

		// duplicate day from config. can be either, any day, or specific day.
		if _, ok := parsedDays[window.day]; ok {
			return nil, fmt.Errorf("can not specify two time intervals for one day")
		}

		parsedDays[window.day] = struct{}{}
		windows = append(windows, window)
	}

	if dailyWindowSpecified {
		if len(windows) != 1 {
			return nil, fmt.Errorf("can not use daily and weekly maintenance window policies simultaneously")
		}

		return &k8s.MaintenanceWindow{
			Policy: &k8s.MaintenanceWindow_DailyMaintenanceWindow{
				DailyMaintenanceWindow: &k8s.DailyMaintenanceWindow{
					StartTime: windows[0].startTime,
					Duration:  windows[0].duration,
				},
			},
		}, nil
	}

	var dw []*k8s.DaysOfWeekMaintenanceWindow
	for _, w := range windows {
		dw = append(dw, &k8s.DaysOfWeekMaintenanceWindow{
			Days:      []dayofweek.DayOfWeek{w.day},
			StartTime: w.startTime,
			Duration:  w.duration,
		})
	}

	return &k8s.MaintenanceWindow{
		Policy: &k8s.MaintenanceWindow_WeeklyMaintenanceWindow{
			WeeklyMaintenanceWindow: &k8s.WeeklyMaintenanceWindow{
				DaysOfWeek: dw,
			},
		},
	}, nil
}

func expandDayMaintenanceWindow(daySpec map[string]interface{}) (*dayMaintenanceWindow, error) {
	var (
		window dayMaintenanceWindow
		err    error
	)

	// special case. Terraform fills fields in Set resource, that are not present in original user config!
	if dayName, ok := daySpec["day"]; ok && dayName != "" {
		if window.day, err = parseDayOfWeek(dayName.(string)); err != nil {
			return nil, err
		}
	}

	if window.startTime, err = ParseDayTime(daySpec["start_time"].(string)); err != nil {
		return nil, err
	}

	if window.duration, err = sdkcommon.ParseDuration(daySpec["duration"].(string)); err != nil {
		return nil, err
	}

	return &window, nil
}

type dayMaintenanceWindow struct {
	day       dayofweek.DayOfWeek
	startTime *timeofday.TimeOfDay
	duration  *duration.Duration
}
