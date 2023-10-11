package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/monitoring/v3"
	"testing"
)

const (
	monitoringDashboardDataSource = "data.yandex_monitoring_dashboard.this"
)

func TestAccDataSourceMonitoringDashboard_byId(t *testing.T) {
	t.Parallel()

	str := `data yandex_monitoring_dashboard "this" {
		dashboard_id	= yandex_monitoring_dashboard.this.dashboard_id
	}`
	testAccDataSourceMonitoringDashboardWithDataSpecified(t, str, "by-id")
}

func TestAccDataSourceMonitoringDashboard_byName(t *testing.T) {
	t.Parallel()

	str := `data yandex_monitoring_dashboard "this" {
		name	= yandex_monitoring_dashboard.this.name
	}`
	testAccDataSourceMonitoringDashboardWithDataSpecified(t, str, "by-name")
}

func testAccDataSourceMonitoringDashboardWithDataSpecified(t *testing.T, dataString string, name string) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMonitoringDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMonitoringDashboard(dataString, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "name", fmt.Sprintf("local-id-data-source-%s", name)),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "description", "Dashboard description"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "title", "My title"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "labels.a", "b"),
					resource.TestCheckResourceAttrSet(monitoringDashboardDataSource, "dashboard_id"),
					resource.TestCheckResourceAttrSet(monitoringDashboardDataSource, "folder_id"),

					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.selectors", "a=b"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.2.hidden", "true"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.2.id", "param3"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.2.text.0.default_value", "abc"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.1.hidden", "true"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.1.id", "param2"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.1.label_values.0.multiselectable", "true"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.1.label_values.0.label_key", "key"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.1.label_values.0.default_values.0", "1"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.1.label_values.0.default_values.1", "2"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.0.hidden", "false"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.0.id", "param1"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.0.description", "param1 description"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.0.title", "title"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.0.custom.0.multiselectable", "true"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.0.custom.0.default_values.0", "1"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.0.custom.0.default_values.1", "2"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.0.custom.0.values.0", "1"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.0.custom.0.values.1", "2"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "parametrization.0.parameters.0.custom.0.values.2", "3"),

					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.0.text.0.text", "text here"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.0.position.0.h", "1"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.0.position.0.w", "1"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.0.position.0.x", "4"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.0.position.0.x", "4"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.2.title.0.text", "title here"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.description", "chart description"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.title", "title for chart"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.chart_id", "chart1id"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.display_legend", "true"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.freeze", "FREEZE_DURATION_HOUR"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.name_hiding_settings.0.positive", "true"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.name_hiding_settings.0.names.0", "a"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.name_hiding_settings.0.names.1", "b"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.queries.0.downsampling.0.disabled", "false"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.queries.0.downsampling.0.gap_filling", "GAP_FILLING_NULL"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.queries.0.downsampling.0.grid_aggregation", "GRID_AGGREGATION_COUNT"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.queries.0.downsampling.0.max_points", "100"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.queries.0.target.0.hidden", "true"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.queries.0.target.0.text_mode", "true"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.queries.0.target.0.query", "{service=monitoring}"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.series_overrides.0.name", "name"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.series_overrides.0.settings.0.color", "colorValue"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.series_overrides.0.settings.0.grow_down", "true"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.series_overrides.0.settings.0.name", "series_overrides name"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.series_overrides.0.settings.0.type", "SERIES_VISUALIZATION_TYPE_LINE"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.series_overrides.0.settings.0.yaxis_position", "YAXIS_POSITION_LEFT"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.series_overrides.0.settings.0.stack_name", "stack name"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.aggregation", "SERIES_AGGREGATION_AVG"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.interpolate", "INTERPOLATE_LEFT"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.type", "VISUALIZATION_TYPE_POINTS"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.normalize", "true"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.show_labels", "true"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.title", "visualization_settings title"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.green_value", "11"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.red_value", "22"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.violet_value", "33"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.yellow_value", "44"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.heatmap_settings.0.green_value", "1"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.heatmap_settings.0.red_value", "2"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.heatmap_settings.0.violet_value", "3"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.heatmap_settings.0.yellow_value", "4"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.max", "111"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.min", "11"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.title", "yaxis_settings left title"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.precision", "3"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.type", "YAXIS_TYPE_LOGARITHMIC"),
					resource.TestCheckResourceAttr(monitoringDashboardDataSource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.unit_format", "UNIT_CELSIUS"),
				),
			},
		},
	})
}

func testAccDataSourceMonitoringDashboard(data string, name string) string {
	return fmt.Sprintf(`
	resource "yandex_monitoring_dashboard" "this" {
	  name        = "local-id-data-source-%s"
	  description = "Dashboard description"
	  title       = "My title"
	  labels      = {
		a = "b"
	  }
	  parametrization {
		selectors = "a=b"
		parameters {
		  description = "param1 description"
		  title       = "title"
		  hidden      = false
		  id          = "param1"
		  custom {
			default_values  = ["1", "2"]
			values          = ["1", "2", "3"]
			multiselectable = true
		  }
		}
		parameters {
		  hidden = true
		  id     = "param2"
		  label_values {
			default_values  = ["1", "2"]
			multiselectable = true
			label_key       = "key"
			selectors       = "a=b"
		  }
		}
		parameters {
		  hidden = true
		  id     = "param3"
		  text {
			default_value = "abc"
		  }
		}
	  }
	  widgets {
		text {
		  text = "text here"
		}
		position {
		  h = 1
		  w = 1
		  x = 4
		  y = 4
		}
	  }
	  widgets {
		chart {
		  description    = "chart description"
		  title          = "title for chart"
		  chart_id       = "chart1id"
		  display_legend = true
		  freeze         = "FREEZE_DURATION_HOUR"
		  name_hiding_settings {
			names    = ["a", "b"]
			positive = true
		  }
		  queries {
			downsampling {
			  disabled         = false
			  gap_filling      = "GAP_FILLING_NULL"
			  grid_aggregation = "GRID_AGGREGATION_COUNT"
			  max_points       = 100
			}
			target {
			  hidden    = true
			  text_mode = true
			  query     = "{service=monitoring}"
			}
		  }
		  series_overrides {
			name = "name"
			settings {
			  color          = "colorValue"
			  grow_down      = true
			  name           = "series_overrides name"
			  type           = "SERIES_VISUALIZATION_TYPE_LINE"
			  yaxis_position = "YAXIS_POSITION_LEFT"
			  stack_name     = "stack name"
			}
		  }
		  visualization_settings {
			aggregation = "SERIES_AGGREGATION_AVG"
			interpolate = "INTERPOLATE_LEFT"
			type        = "VISUALIZATION_TYPE_POINTS"
			normalize   = true
			show_labels = true
			title       = "visualization_settings title"
			color_scheme_settings {
			  gradient {
 				green_value  = "11"
			    red_value    = "22"
			    violet_value = "33"
			    yellow_value = "44"
			  }
			}
			heatmap_settings {
			  green_value  = "1"
			  red_value    = "2"
			  violet_value = "3"
			  yellow_value = "4"
			}
			yaxis_settings {
			  left {
				max         = "111"
				min         = "11"
				title       = "yaxis_settings left title"
				precision   = 3
				type        = "YAXIS_TYPE_LOGARITHMIC"
				unit_format = "UNIT_CELSIUS"
	
			  }
			}
		  }
		}
		position {
		  h = 100
		  w = 100
		  x = 6
		  y = 6
		}
	  }
	  widgets {
		title {
		  text = "title here"
		}
		position {
		  h = 1
		  w = 1
		  x = 1
		  y = 1
		}
	  }
	}

	%s`, name, data)
}

func testAccCheckMonitoringDashboardDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_monitoring_dashboard" {
			continue
		}
		ctx := context.Background()
		req := &monitoring.GetDashboardRequest{
			DashboardId: rs.Primary.ID,
		}
		_, err := config.sdk.Monitoring().Dashboard().Get(ctx, req)
		if err == nil {
			return fmt.Errorf("Dashboard still exists")
		}
	}

	return nil
}
