---
layout: "yandex"
page_title: "Yandex: yandex_monitoring_dashboard"
sidebar_current: "docs-yandex-monitoring-dashboard"
description: |-
Allows management of a Yandex.Cloud Monitoring Dashboard.
---

# yandex\_monitoring\_dashboard

Get information about a Yandex Monitoring dashboard.

## Example Usage

```hcl
resource "yandex_monitoring_dashboard" "my-dashboard" {
  name        = "local-id-resource"
  description = "Description"
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
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the Dashboard.
* `folder_id` - Folder that the resource belongs to. If value is omitted, the default provider folder is used.
* `description` - Dashboard description.
* `title` - Dashboard title.
* `labels` - A set of key/value label pairs to assign to the Dashboard.
* `parametrization` - Dashboard parametrization
* `widgets` - Widgets

The `parametrization` block supports:

* `parameters` - parameters list.
* `selectors` - dashboard predefined parameters selector.

The `parameters` block supports:

* `description` - Parameter description.
* `hidden` - UI-visibility.
* `id` - (Required) Parameter identifier
* `title` - UI-visible title of the parameter.
* `custom` - Custom values parameter. Oneof: label_values, custom, text.
* `label_values` - Label values parameter. Oneof: label_values, custom, text.
* `text` - Text parameter. Oneof: label_values, custom, text.

The `custom` block supports:

* `default_values` - Default value.
* `multiselectable` - Specifies the multiselectable values of parameter.
* `values` - Parameter values.

The `label_values` block supports:

* `default_values` - Default value.
* `folder_id` - Labels folder ID.
* `label_key` - (Required) Label key to list label values.
* `multiselectable` - Specifies the multiselectable values of parameter.
* `selectors` - (Required) Selectors to select metric label values.

The `text` block supports:

* `default_value` - Default value.

The `widgets` block supports:

* `position` - Widget position.
* `text` - Text widget settings. Oneof: text, title or chart.
* `title` - Title widget settings. Oneof: text, title or chart.
* `chart` - Chart widget settings. Oneof: text, title or chart.

The `position` block supports:

* `h` - Height.
* `w` - Width.
* `x` - X-axis top-left corner coordinate.
* `y` - Y-axis top-left corner coordinate.

The `text` block supports:

* `text` - Widget text.

The `title` block supports:

* `text` - Title text.
* `size` - Title size. Values: 
  - TITLE_SIZE_XS: Extra small size.
  - TITLE_SIZE_S: Small size.
  - TITLE_SIZE_M: Middle size.
  - TITLE_SIZE_L: Large size.

The `chart` block supports:

* `chart_id` - Chart ID.
* `description` - Chart description in dashboard (not enabled in UI).
* `display_legend` - Enable legend under chart.
* `freeze` - Fixed time interval for chart. Values:
  - FREEZE_DURATION_HOUR: Last hour.
  - FREEZE_DURATION_DAY: Last day = last 24 hours.
  - FREEZE_DURATION_WEEK: Last 7 days.
  - FREEZE_DURATION_MONTH: Last 31 days.
* `name_hiding_settings` - Names settings.
* `queries` - (Required) Queries settings.
* `series_overrides` Time series settings.
* `title` - Chart widget title.
* `visualization_settings` - Visualization settings.

The `name_hiding_settings` block supports:

* `names` - Series name.
* `positive` - True if we want to show concrete series names only, false if we want to hide concrete series names.

The `queries` block supports:

* `downsampling` - Downsamplang settings.
* `target` - Query targets.

The `downsampling` block supports:

* `disabled` - Disable downsampling.
* `gap_filling` - Parameters for filling gaps in data.
* `grid_aggregation` - Function that is used for downsampling.
* `grid_interval` - Time interval (grid) for downsampling in milliseconds. Points in the specified range are aggregated into one time point
* `max_points` - Maximum number of points to be returned.

The `target` block supports:

* `hidden` - Checks that target is visible or invisible.
* `query` - Query.
* `text_mode` - Text mode enabled.

The `series_overrides` block supports:

* `name` - Series name. Oneof: name or target_index.
* `settings` - Override settings.
* `target_index` - Series index. Oneof: name or target_index.

The `settings` block supports:

* `color` - Series color or empty.
* `grow_down` - Stack grow down.
* `name` - Series name or empty.
* `stack_name` - Stack name or empty.
* `type` - Type.
* `yaxis_position` - Yaxis position.

The `visualization_settings` block supports:

* `aggregation` - Aggregation. Values:
  - SERIES_AGGREGATION_UNSPECIFIED: Not specified (avg by default).
  - SERIES_AGGREGATION_AVG: Average.
  - SERIES_AGGREGATION_MIN: Minimum.
  - SERIES_AGGREGATION_MAX: Maximum.
  - SERIES_AGGREGATION_LAST: Last non-NaN value.
  - SERIES_AGGREGATION_SUM: Sum.
* `color_scheme_settings` - Color settings.
* `heatmap_settings` - Heatmap settings.
* `interpolate` - Interpolate values. Values:
  - INTERPOLATE_UNSPECIFIED: Not specified (linear by default).
  - INTERPOLATE_LINEAR: Linear.
  - INTERPOLATE_LEFT: Left.
  - INTERPOLATE_RIGHT: Right.
* `normalize` - Normalize values.
* `show_labels` - Show chart labels.
* `title` - Inside chart title.
* `type` - Visualization type. Values:
  - VISUALIZATION_TYPE_UNSPECIFIED: Not specified (line by default).
  - VISUALIZATION_TYPE_LINE: Line chart.
  - VISUALIZATION_TYPE_STACK: Stack chart.
  - VISUALIZATION_TYPE_COLUMN: Points as columns chart.
  - VISUALIZATION_TYPE_POINTS: Points.
  - VISUALIZATION_TYPE_PIE: Pie aggregation chart.
  - VISUALIZATION_TYPE_BARS: Bars aggregation chart.
  - VISUALIZATION_TYPE_DISTRIBUTION: Distribution aggregation chart.
  - VISUALIZATION_TYPE_HEATMAP: Heatmap aggregation chart.
* `yaxis_settings` - Y axis settings.

The `color_scheme_settings` block supports:

* `automatic` - Automatic color scheme. Oneof: automatic, standard or gradient.
* `gradient` - Gradient color scheme. Oneof: automatic, standard or gradient.
* `standard` - Standard color scheme. Oneof: automatic, standard or gradient.

The `gradient` block supports:

* `green_value` - Gradient green value.
* `red_value` - Gradient red value.
* `violet_value` - Gradient violet value.
* `yellow_value` - Gradient yellow value.

The `heatmap_settings` block supports:

* `green_value` - Heatmap green value.
* `red_value` - Heatmap red value.
* `violet_value` - Heatmap violet value.
* `yellow_value` - Heatmap yellow value.

The `yaxis_settings` block supports:

* `left` - Left yaxis config.
* `right` Right yaxis config.

The `left` block supports:

* `max` - Max value in extended number format or empty.
* `min` - Min value in extended number format or empty.
* `precision` - Tick value precision (null as default, 0-7 in other cases).
* `title` -Title or empty.
* `type` - Type. Values:
  - YAXIS_TYPE_UNSPECIFIED: Not specified (linear by default).
  - YAXIS_TYPE_LINEAR: Linear.
  - YAXIS_TYPE_LOGARITHMIC: Logarithmic.
* `unit_format` - Unit format. Values:
  - UNIT_NONE: Misc. None (show tick values as-is).
  - UNIT_COUNT: Count.
  - UNIT_PERCENT: Percent (0-100).
  - UNIT_PERCENT_UNIT: Percent (0-1).
  - UNIT_NANOSECONDS: Time. Nanoseconds (ns).
  - UNIT_MICROSECONDS: Microseconds (µs).
  - UNIT_MILLISECONDS: Milliseconds (ms).
  - UNIT_SECONDS: Seconds (s).
  - UNIT_MINUTES: Minutes (m).
  - UNIT_HOURS: Hours (h).
  - UNIT_DAYS: Days (d).
  - UNIT_BITS_SI: Data (SI). Bits (SI).
  - UNIT_BYTES_SI: Bytes (SI).
  - UNIT_KILOBYTES: Kilobytes (KB).
  - UNIT_MEGABYTES: Megabytes (MB).
  - UNIT_GIGABYTES: Gigabytes (GB).
  - UNIT_TERABYTES: Terabytes (TB)
  - UNIT_PETABYTES: Petabytes (PB).
  - UNIT_EXABYTES: Exabytes (EB).
  - UNIT_BITS_IEC: Data (IEC). Bits (IEC).
  - UNIT_BYTES_IEC: Bytes (IEC).
  - UNIT_KIBIBYTES: Kibibytes (KiB).
  - UNIT_MEBIBYTES: Mebibytes (MiB).
  - UNIT_GIBIBYTES: Gigibytes (GiB).
  - UNIT_TEBIBYTES: Tebibytes (TiB).
  - UNIT_PEBIBYTES: Pebibytes (PiB).
  - UNIT_EXBIBYTES: Exbibytes (EiB).
  - UNIT_REQUESTS_PER_SECOND: Throughput. Requests per second (reqps).
  - UNIT_OPERATIONS_PER_SECOND: Operations per second (ops).
  - UNIT_WRITES_PER_SECOND: Writes per second (wps).
  - UNIT_READS_PER_SECOND: Reads per second (rps).
  - UNIT_PACKETS_PER_SECOND: Packets per second (pps).
  - UNIT_IO_OPERATIONS_PER_SECOND: IO operations per second (iops).
  - UNIT_COUNTS_PER_SECOND: Counts per second (counts/sec).
  - UNIT_BITS_SI_PER_SECOND: Data Rate (SI). Bits (SI) per second (bits/sec).
  - UNIT_BYTES_SI_PER_SECOND: Bytes (SI) per second (bytes/sec).
  - UNIT_KILOBITS_PER_SECOND: Kilobits per second (KBits/sec).
  - UNIT_KILOBYTES_PER_SECOND: Kilobytes per second (KB/sec).
  - UNIT_MEGABITS_PER_SECOND: Megabits per second (MBits/sec).
  - UNIT_MEGABYTES_PER_SECOND: Megabytes per second (MB/sec).
  - UNIT_GIGABITS_PER_SECOND: Gigabits per second (GBits/sec).
  - UNIT_GIGABYTES_PER_SECOND: Gigabytes per second (GB/sec).
  - UNIT_TERABITS_PER_SECOND: Terabits per second (TBits/sec).
  - UNIT_TERABYTES_PER_SECOND: Terabytes per second (TB/sec).
  - UNIT_PETABITS_PER_SECOND: Petabits per second (Pbits/sec).
  - UNIT_PETABYTES_PER_SECOND: Petabytes per second (PB/sec).
  - UNIT_BITS_IEC_PER_SECOND: Data Rate (IEC). Bits (IEC) per second (bits/sec).
  - UNIT_BYTES_IEC_PER_SECOND: Bytes (IEC) per second (bytes/sec).
  - UNIT_KIBIBITS_PER_SECOND: Kibibits per second (KiBits/sec).
  - UNIT_KIBIBYTES_PER_SECOND: Kibibytes per second (KiB/sec).
  - UNIT_MEBIBITS_PER_SECOND: Mebibits per second (MiBits/sec).
  - UNIT_MEBIBYTES_PER_SECOND: Mebibytes per second (MiB/sec).
  - UNIT_GIBIBITS_PER_SECOND: Gibibits per second (GiBits/sec).
  - UNIT_GIBIBYTES_PER_SECOND: Gibibytes per second (GiB/sec).
  - UNIT_TEBIBITS_PER_SECOND: Tebibits per second (TiBits/sec).
  - UNIT_TEBIBYTES_PER_SECOND: Tebibytes per second (TiB/sec).
  - UNIT_PEBIBITS_PER_SECOND: Pebibits per second (PiBits/sec).
  - UNIT_PEBIBYTES_PER_SECOND: Pebibytes per second (PiB/sec).
  - UNIT_DATETIME_UTC: Date & time. Datetime (UTC).
  - UNIT_DATETIME_LOCAL: Datetime (local).
  - UNIT_HERTZ: Frequency. Hertz (Hz).
  - UNIT_KILOHERTZ: Kilohertz (KHz).
  - UNIT_MEGAHERTZ: Megahertz (MHz).
  - UNIT_GIGAHERTZ: Gigahertz (GHz).
  - UNIT_DOLLAR: Currency. Dollar.
  - UNIT_EURO: Euro.
  - UNIT_ROUBLE: Rouble.
  - UNIT_CELSIUS: Temperature. Celsius (°C).
  - UNIT_FAHRENHEIT: Fahrenheit (°F).
  - UNIT_KELVIN: Kelvin (K).
  - UNIT_FLOP_PER_SECOND: Computation. Flop per second (FLOP/sec).
  - UNIT_KILOFLOP_PER_SECOND: Kiloflop per second (KFLOP/sec).
  - UNIT_MEGAFLOP_PER_SECOND: Megaflop per second (MFLOP/sec).
  - UNIT_GIGAFLOP_PER_SECOND: Gigaflop per second (GFLOP/sec).
  - UNIT_PETAFLOP_PER_SECOND: Petaflop per second (PFLOP/sec).
  - UNIT_EXAFLOP_PER_SECOND: Exaflop per second (EFLOP/sec).
  - UNIT_METERS_PER_SECOND: Velocity. Meters per second (m/sec).
  - UNIT_KILOMETERS_PER_HOUR: Kilometers per hour (km/h).
  - UNIT_MILES_PER_HOUR: Miles per hour (mi/h).
  - UNIT_MILLIMETER: Length. Millimeter.
  - UNIT_CENTIMETER: Centimeter.
  - UNIT_METER: Meter.
  - UNIT_KILOMETER: Kilometer.
  - UNIT_MILE: Mile.
  - UNIT_PPM: Concentration. Parts per million (ppm).
  - UNIT_EVENTS_PER_SECOND: Events per second
  - UNIT_PACKETS: Packets
  - UNIT_DBM: dBm (dbm)
  - UNIT_VIRTUAL_CPU: Virtual CPU cores based on CPU time (vcpu)
  - UNIT_MESSAGES_PER_SECOND: Messages per second (mps)

The `right` block supports:

* `max` - Max value in extended number format or empty.
* `min` - Min value in extended number format or empty.
* `precision` - Tick value precision (null as default, 0-7 in other cases).
* `title` -Title or empty.
* `type` - Type.
* `unit_format` - Unit format.

## Timeouts

This resource provides the following configuration options for
[timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default is 2 minute.
- `update` - Default is 2 minute.
- `delete` - Default is 2 minute.

## Import

A Monitoring dashboard can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_monitoring_dashboard.default dashboard_id
```



