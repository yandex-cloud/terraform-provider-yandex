---
subcategory: "Monitoring"
page_title: "Yandex: yandex_monitoring_dashboard"
description: |-
  Get information about a Yandex Monitoring dashboard.
---

# yandex_monitoring_dashboard (Data Source)

Get information about a Yandex Monitoring dashboard.

~> One of `dashboard_id` or `name` should be specified.

## Example usage

```terraform
//
// Get information about existing Monitoring Dashboard.
//
data "yandex_monitoring_dashboard" "my_dashboard" {
  dashboard_id = "some_instance_dashboard_id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `dashboard_id` (String) Dashboard ID.
- `description` (String) The resource description.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `name` (String) The resource name.

### Read-Only

- `id` (String) The ID of this resource.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `parametrization` (List of Object) Dashboard parametrization (see [below for nested schema](#nestedatt--parametrization))
- `title` (String) Dashboard title.
- `widgets` (List of Object) Widgets (see [below for nested schema](#nestedatt--widgets))

<a id="nestedatt--parametrization"></a>
### Nested Schema for `parametrization`

Read-Only:

- `parameters` (Block List) Dashboard parameters. (see [below for nested schema](#nestedobjatt--parametrization--parameters))

- `selectors` (String) Dashboard predefined parameters selector.


<a id="nestedobjatt--parametrization--parameters"></a>
### Nested Schema for `parametrization.parameters`

Read-Only:

- `custom` (Block List) Custom values parameter. Oneof: label_values, custom, text. (see [below for nested schema](#nestedobjatt--parametrization--parameters--custom))

- `description` (String) Parameter description.

- `hidden` (Boolean) UI-visibility

- `id` (String) Parameter identifier.

- `label_values` (Block List) Label values parameter. Oneof: label_values, custom, text. (see [below for nested schema](#nestedobjatt--parametrization--parameters--label_values))

- `text` (Block List) Text parameter. Oneof: label_values, custom, text. (see [below for nested schema](#nestedobjatt--parametrization--parameters--text))

- `title` (String) UI-visible title of the parameter.


<a id="nestedobjatt--parametrization--parameters--custom"></a>
### Nested Schema for `parametrization.parameters.custom`

Read-Only:

- `default_values` (List of String) Default value.

- `multiselectable` (Boolean) Specifies the multiselectable values of parameter.

- `values` (List of String) Parameter values.



<a id="nestedobjatt--parametrization--parameters--label_values"></a>
### Nested Schema for `parametrization.parameters.label_values`

Read-Only:

- `default_values` (List of String) Default value.

- `folder_id` (String) Folder ID.

- `label_key` (String) Label key to list label values.

- `multiselectable` (Boolean) Specifies the multiselectable values of parameter.

- `selectors` (String) Selectors to select metric label values.



<a id="nestedobjatt--parametrization--parameters--text"></a>
### Nested Schema for `parametrization.parameters.text`

Read-Only:

- `default_value` (String) Default value.





<a id="nestedatt--widgets"></a>
### Nested Schema for `widgets`

Read-Only:

- `chart` (Block List) Chart widget settings. (see [below for nested schema](#nestedobjatt--widgets--chart))

- `position` (Block List) Widget layout position. (see [below for nested schema](#nestedobjatt--widgets--position))

- `text` (Block List) Text widget settings. (see [below for nested schema](#nestedobjatt--widgets--text))

- `title` (Block List) Title widget settings. (see [below for nested schema](#nestedobjatt--widgets--title))


<a id="nestedobjatt--widgets--chart"></a>
### Nested Schema for `widgets.chart`

Read-Only:

- `chart_id` (String) Chart ID.

- `description` (String) Chart description in dashboard (not enabled in UI).

- `display_legend` (Boolean) Enable legend under chart.

- `freeze` (String) Fixed time interval for chart. Values:

- `name_hiding_settings` (Block List) Name hiding settings (see [below for nested schema](#nestedobjatt--widgets--chart--name_hiding_settings))

- `queries` (Block List) Queries settings. (see [below for nested schema](#nestedobjatt--widgets--chart--queries))

- `series_overrides` (Block List) Time series settings. (see [below for nested schema](#nestedobjatt--widgets--chart--series_overrides))

- `title` (String) Chart widget title.

- `visualization_settings` (Block List) Visualization settings. (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings))


<a id="nestedobjatt--widgets--chart--name_hiding_settings"></a>
### Nested Schema for `widgets.chart.name_hiding_settings`

Read-Only:

- `names` (List of String)
- `positive` (Boolean) True if we want to show concrete series names only, false if we want to hide concrete series names



<a id="nestedobjatt--widgets--chart--queries"></a>
### Nested Schema for `widgets.chart.queries`

Read-Only:

- `downsampling` (Block List) Downsampling settings (see [below for nested schema](#nestedobjatt--widgets--chart--queries--downsampling))

- `target` (Block List) Downsampling settings (see [below for nested schema](#nestedobjatt--widgets--chart--queries--target))


<a id="nestedobjatt--widgets--chart--queries--downsampling"></a>
### Nested Schema for `widgets.chart.queries.downsampling`

Read-Only:

- `disabled` (Boolean) Disable downsampling

- `gap_filling` (String) Parameters for filling gaps in data

- `grid_aggregation` (String) Function that is used for downsampling

- `grid_interval` (Number) Time interval (grid) for downsampling in milliseconds. Points in the specified range are aggregated into one time point

- `max_points` (Number) Maximum number of points to be returned



<a id="nestedobjatt--widgets--chart--queries--target"></a>
### Nested Schema for `widgets.chart.queries.target`

Read-Only:

- `hidden` (Boolean) Checks that target is visible or invisible

- `query` (String) Required. Query

- `text_mode` (Boolean) Text mode




<a id="nestedobjatt--widgets--chart--series_overrides"></a>
### Nested Schema for `widgets.chart.series_overrides`

Read-Only:

- `name` (String) Series name

- `settings` (Block List) Override settings (see [below for nested schema](#nestedobjatt--widgets--chart--series_overrides--settings))

- `target_index` (String) Target index


<a id="nestedobjatt--widgets--chart--series_overrides--settings"></a>
### Nested Schema for `widgets.chart.series_overrides.settings`

Read-Only:

- `color` (String) Series color or empty

- `grow_down` (Boolean) Stack grow down

- `name` (String) Series name or empty

- `stack_name` (String) Stack name or empty

- `type` (String) Type

- `yaxis_position` (String) Yaxis position




<a id="nestedobjatt--widgets--chart--visualization_settings"></a>
### Nested Schema for `widgets.chart.visualization_settings`

Read-Only:

- `aggregation` (String) Aggregation

- `color_scheme_settings` (Block List) Color scheme settings (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--color_scheme_settings))

- `heatmap_settings` (Block List) Heatmap settings (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--heatmap_settings))

- `interpolate` (String) Interpolate

- `normalize` (Boolean) Normalize

- `show_labels` (Boolean) Show chart labels

- `title` (String) Inside chart title

- `type` (String) Visualization type

- `yaxis_settings` (Block List) Y axis settings (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings))


<a id="nestedobjatt--widgets--chart--visualization_settings--color_scheme_settings"></a>
### Nested Schema for `widgets.chart.visualization_settings.color_scheme_settings`

Read-Only:

- `automatic` (Block List) Automatic color scheme (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--automatic))

- `gradient` (Block List) Gradient color scheme (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--gradient))

- `standard` (Block List) Standard color scheme (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--standard))


<a id="nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--automatic"></a>
### Nested Schema for `widgets.chart.visualization_settings.yaxis_settings.automatic`

Read-Only:



<a id="nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--gradient"></a>
### Nested Schema for `widgets.chart.visualization_settings.yaxis_settings.gradient`

Read-Only:

- `green_value` (String)
- `red_value` (String)
- `violet_value` (String)
- `yellow_value` (String)


<a id="nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--standard"></a>
### Nested Schema for `widgets.chart.visualization_settings.yaxis_settings.standard`

Read-Only:




<a id="nestedobjatt--widgets--chart--visualization_settings--heatmap_settings"></a>
### Nested Schema for `widgets.chart.visualization_settings.heatmap_settings`

Read-Only:

- `green_value` (String) Heatmap green value

- `red_value` (String) Heatmap red value

- `violet_value` (String) Heatmap violet_value

- `yellow_value` (String) Heatmap yellow value



<a id="nestedobjatt--widgets--chart--visualization_settings--yaxis_settings"></a>
### Nested Schema for `widgets.chart.visualization_settings.yaxis_settings`

Read-Only:

- `left` (Block List) Left Y axis settings (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--left))

- `right` (Block List) Right Y axis settings (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--right))


<a id="nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--left"></a>
### Nested Schema for `widgets.chart.visualization_settings.yaxis_settings.left`

Read-Only:

- `max` (String) Max value in extended number format or empty

- `min` (String) Min value in extended number format or empty

- `precision` (Number) Tick value precision (null as default, 0-7 in other cases)

- `title` (String) Title or empty

- `type` (String) Type

- `unit_format` (String) Unit format



<a id="nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--right"></a>
### Nested Schema for `widgets.chart.visualization_settings.yaxis_settings.right`

Read-Only:

- `max` (String) Max value in extended number format or empty

- `min` (String) Min value in extended number format or empty

- `precision` (Number) Tick value precision (null as default, 0-7 in other cases)

- `title` (String) Title or empty

- `type` (String) Type

- `unit_format` (String) Unit format






<a id="nestedobjatt--widgets--position"></a>
### Nested Schema for `widgets.position`

Read-Only:

- `h` (Number) Height.

- `w` (Number) Weight.

- `x` (Number) X-axis top-left corner coordinate.

- `y` (Number) Y-axis top-left corner coordinate.



<a id="nestedobjatt--widgets--text"></a>
### Nested Schema for `widgets.text`

Read-Only:

- `text` (String) Widget text.



<a id="nestedobjatt--widgets--title"></a>
### Nested Schema for `widgets.title`

Read-Only:

- `size` (String) Title size.

- `text` (String) Title text.

