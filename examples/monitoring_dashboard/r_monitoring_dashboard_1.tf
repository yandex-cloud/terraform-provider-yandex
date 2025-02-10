//
// Create a new Monitoring Dashboard.
//
resource "yandex_monitoring_dashboard" "my-dashboard" {
  name        = "local-id-resource"
  description = "Description"
  title       = "My title"
  labels = {
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
}
