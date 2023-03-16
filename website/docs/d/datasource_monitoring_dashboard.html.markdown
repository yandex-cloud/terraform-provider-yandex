---
layout: "yandex"
page_title: "Yandex: yandex_monitoring_dashboard"
sidebar_current: "docs-yandex-datasource-monitoring-dashboard"
description: |-
Get information about a Yandex Monitoring dashboard.
---

# yandex\_monitoring\_dashboard

Get information about a Yandex Monitoring dashboard.

## Example Usage

```hcl
data "yandex_monitoring_dashboard" "my_dashboard" {
  dashboard_id = "some_instance_dashboard_id"
}
```

## Argument Reference

The following arguments are supported:

* `dashboard_id` (Optional) - Dashboard ID.
* `name` - (Optional) - Name of the Dashboard.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

~> **NOTE:** One of `dashboard_id` or `name` should be specified.

## Attributes Reference

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
* `id` - Parameter identifier
* `title` - UI-visible title of the parameter.
* `custom` - Custom values parameter.
* `label_values` - Label values parameter.
* `text` - Text parameter.

The `custom` block supports:

* `default_values` - Default value.
* `multiselectable` - Specifies the multiselectable values of parameter.
* `values` - Parameter values.

The `label_values` block supports:

* `default_values` - Default value.
* `folder_id` - Labels folder ID.
* `label_key` - Label key to list label values.
* `multiselectable` - Specifies the multiselectable values of parameter.
* `selectors` - Selectors to select metric label values.

The `text` block supports:

* `default_value` - Default value.

The `widgets` block supports:

* `position` - Widget position.
* `text` - Text widget settings.
* `title` - Title widget settings.
* `chart` - Chart widget settings.

The `position` block supports:

* `h` - Height.
* `w` - Width.
* `x` - X-axis top-left corner coordinate.
* `y` - Y-axis top-left corner coordinate.

The `text` block supports:

* `text` - Widget text.

The `title` block supports:

* `text` - Title text.
* `size` - Title size.

The `chart` block supports:

* `chart_id` - Chart ID.
* `description` - Chart description in dashboard (not enabled in UI).
* `display_legend` - Enable legend under chart.
* `freeze` - Fixed time interval for chart.
* `name_hiding_settings` - Names settings.
* `queries` - Queries settings.
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

* `name` - Series name.
* `settings` - Override settings.
* `target_index` - Series index.

The `settings` block supports:

* `color` - Series color or empty.
* `grow_down` - Stack grow down.
* `name` - Series name or empty.
* `stack_name` - Stack name or empty.
* `type` - Type.
* `yaxis_position` - Yaxis positio

The `visualization_settings` block supports:

* `aggregation` - Aggregation.
* `color_scheme_settings` - Color settings.
* `heatmap_settings` - Heatmap settings.
* `interpolate` - Interpolate values.
* `normalize` - Normalize values.
* `show_labels` - Show chart labels.
* `title` - Inside chart title.
* `type` - Visualization type.
* `yaxis_settings` - Y axis settings.

The `color_scheme_settings` block supports:

* `automatic` - Automatic color scheme.
* `gradient` - Gradient color scheme.
* `standard` - Standard color scheme.

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
* `type` - Type.
* `unit_format` - Unit format.

The `right` block supports:

* `max` - Max value in extended number format or empty.
* `min` - Min value in extended number format or empty.
* `precision` - Tick value precision (null as default, 0-7 in other cases).
* `title` -Title or empty.
* `type` - Type.
* `unit_format` - Unit format.





