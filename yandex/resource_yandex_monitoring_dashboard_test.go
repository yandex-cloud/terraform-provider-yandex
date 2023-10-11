package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/monitoring/v3"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	monitoringDashboardResource = "yandex_monitoring_dashboard.this"
)

func init() {
	resource.AddTestSweepers("yandex_monitoring_dashboard", &resource.Sweeper{
		Name: "yandex_monitoring_dashboard",
		F:    testSweepMonitoringDashboard,
	})
}

func testSweepMonitoringDashboard(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	ctx := context.Background()
	req := &monitoring.ListDashboardsRequest{
		Container: &monitoring.ListDashboardsRequest_FolderId{
			FolderId: conf.FolderID,
		},
	}
	resp, err := conf.sdk.Monitoring().Dashboard().List(ctx, req)
	if err != nil {
		return fmt.Errorf("error getting monitoring dashboards: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Dashboards {
		if !sweepMonitoringDashboard(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep YDB monitoring dashboard %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMonitoringDashboard(conf *Config, id string) bool {
	return sweepWithRetry(sweepMonitoringDashboardOnce, conf, "Monitoring dashboard", id)
}

func sweepMonitoringDashboardOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexMonitoringDashboardDefaultTimeout)
	defer cancel()
	op, err := conf.sdk.Monitoring().Dashboard().Delete(ctx, &monitoring.DeleteDashboardRequest{
		DashboardId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccResourceMonitoringDashboard(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMonitoringDashboardDestroy,
		Steps: []resource.TestStep{
			{
				// creates dashboard with description
				Config: testAccResourceMonitoringDashboard("Dashboard description"),
				Check:  checkResourceMonitoringDashboardStep("Dashboard description"),
			},
			{
				// updates dashboard with description
				Config: testAccResourceMonitoringDashboard("Dashboard description 2"),
				Check:  checkResourceMonitoringDashboardStep("Dashboard description 2"),
			},
		},
	})
}

func checkResourceMonitoringDashboardStep(description string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccResourceMonitoringDashboardExists(),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "name", "local-id-resource"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "description", description),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "title", "My title"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "labels.a", "b"),
		resource.TestCheckResourceAttrSet(monitoringDashboardResource, "dashboard_id"),
		resource.TestCheckResourceAttrSet(monitoringDashboardResource, "folder_id"),

		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.selectors", "a=b"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.2.hidden", "true"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.2.id", "param3"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.2.text.0.default_value", "abc"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.1.hidden", "true"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.1.id", "param2"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.1.label_values.0.multiselectable", "true"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.1.label_values.0.label_key", "key"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.1.label_values.0.default_values.0", "1"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.1.label_values.0.default_values.1", "2"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.0.hidden", "false"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.0.id", "param1"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.0.description", "param1 description"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.0.title", "title"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.0.custom.0.multiselectable", "true"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.0.custom.0.default_values.0", "1"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.0.custom.0.default_values.1", "2"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.0.custom.0.values.0", "1"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.0.custom.0.values.1", "2"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "parametrization.0.parameters.0.custom.0.values.2", "3"),

		resource.TestCheckNoResourceAttr(monitoringDashboardResource, "widgets.3"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.0.text.0.text", "text here"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.0.position.0.h", "1"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.0.position.0.w", "1"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.0.position.0.x", "4"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.0.position.0.x", "4"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.2.title.0.text", "title here"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.2.title.0.size", "TITLE_SIZE_XS"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.description", "chart description"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.title", "title for chart"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.chart_id", "chart1id"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.display_legend", "true"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.freeze", "FREEZE_DURATION_HOUR"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.name_hiding_settings.0.positive", "true"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.name_hiding_settings.0.names.0", "a"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.name_hiding_settings.0.names.1", "b"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.queries.0.downsampling.0.disabled", "false"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.queries.0.downsampling.0.gap_filling", "GAP_FILLING_NULL"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.queries.0.downsampling.0.grid_aggregation", "GRID_AGGREGATION_COUNT"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.queries.0.downsampling.0.max_points", "100"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.queries.0.target.0.hidden", "true"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.queries.0.target.0.text_mode", "true"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.queries.0.target.0.query", "{service=monitoring}"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.series_overrides.0.name", "name"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.series_overrides.0.settings.0.color", "colorValue"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.series_overrides.0.settings.0.grow_down", "true"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.series_overrides.0.settings.0.name", "series_overrides name"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.series_overrides.0.settings.0.type", "SERIES_VISUALIZATION_TYPE_LINE"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.series_overrides.0.settings.0.yaxis_position", "YAXIS_POSITION_LEFT"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.series_overrides.0.settings.0.stack_name", "stack name"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.aggregation", "SERIES_AGGREGATION_AVG"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.interpolate", "INTERPOLATE_LEFT"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.type", "VISUALIZATION_TYPE_POINTS"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.normalize", "true"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.show_labels", "true"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.title", "visualization_settings title"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.green_value", "11"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.red_value", "22"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.violet_value", "33"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.yellow_value", "44"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.heatmap_settings.0.green_value", "1"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.heatmap_settings.0.red_value", "2"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.heatmap_settings.0.violet_value", "3"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.heatmap_settings.0.yellow_value", "4"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.max", "111"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.min", "11"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.title", "yaxis_settings left title"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.precision", "3"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.type", "YAXIS_TYPE_LOGARITHMIC"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.left.0.unit_format", "UNIT_CELSIUS"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.right.0.max", "22"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.right.0.min", "2"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.right.0.title", "yaxis_settings right title"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.right.0.precision", "2"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.right.0.type", "YAXIS_TYPE_LOGARITHMIC"),
		resource.TestCheckResourceAttr(monitoringDashboardResource, "widgets.1.chart.0.visualization_settings.0.yaxis_settings.0.right.0.unit_format", "UNIT_NONE"),
	)
}

func testAccResourceMonitoringDashboard(description string) string {
	return fmt.Sprintf(`
	resource "yandex_monitoring_dashboard" "this" {
	  name        = "local-id-resource"
	  description = "%s"
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
			  right {
				max         = "22"
				min         = "2"
				title       = "yaxis_settings right title"
				precision   = 2
				type        = "YAXIS_TYPE_LOGARITHMIC"
				unit_format = "UNIT_NONE"
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
		  size = "TITLE_SIZE_XS"
		}
		position {
		  h = 1
		  w = 1
		  x = 1
		  y = 1
		}
	  }
	}`, description)
}

func testAccResourceMonitoringDashboardExists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[monitoringDashboardResource]
		if !ok {
			return fmt.Errorf("Not found: %s", monitoringDashboardResource)
		}

		if rs.Primary.Attributes["dashboard_id"] == "" {
			return fmt.Errorf("No monitoring dashboard id specified!")
		}
		config := testAccProvider.Meta().(*Config)
		ctx := context.Background()
		req := &monitoring.GetDashboardRequest{
			DashboardId: rs.Primary.Attributes["dashboard_id"],
		}
		dashboard, err := config.sdk.Monitoring().Dashboard().Get(ctx, req)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Dashboard %s doesnt exists", rs.Primary.Attributes["dashboard_id"]))
		}
		if dashboard.Id != rs.Primary.Attributes["dashboard_id"] {
			return fmt.Errorf("Dashboard id mismatch")
		}
		return nil
	}
}
