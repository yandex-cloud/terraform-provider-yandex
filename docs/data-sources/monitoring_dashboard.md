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

- `parameters` (List of Object) (see [below for nested schema](#nestedobjatt--parametrization--parameters))
- `selectors` (String)

<a id="nestedobjatt--parametrization--parameters"></a>
### Nested Schema for `parametrization.parameters`

Read-Only:

- `custom` (List of Object) (see [below for nested schema](#nestedobjatt--parametrization--parameters--custom))
- `description` (String)
- `hidden` (Boolean)
- `id` (String)
- `label_values` (List of Object) (see [below for nested schema](#nestedobjatt--parametrization--parameters--label_values))
- `text` (List of Object) (see [below for nested schema](#nestedobjatt--parametrization--parameters--text))
- `title` (String)

<a id="nestedobjatt--parametrization--parameters--custom"></a>
### Nested Schema for `parametrization.parameters.custom`

Read-Only:

- `default_values` (List of String)
- `multiselectable` (Boolean)
- `values` (List of String)


<a id="nestedobjatt--parametrization--parameters--label_values"></a>
### Nested Schema for `parametrization.parameters.label_values`

Read-Only:

- `default_values` (List of String)
- `folder_id` (String)
- `label_key` (String)
- `multiselectable` (Boolean)
- `selectors` (String)


<a id="nestedobjatt--parametrization--parameters--text"></a>
### Nested Schema for `parametrization.parameters.text`

Read-Only:

- `default_value` (String)




<a id="nestedatt--widgets"></a>
### Nested Schema for `widgets`

Read-Only:

- `chart` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart))
- `position` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--position))
- `text` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--text))
- `title` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--title))

<a id="nestedobjatt--widgets--chart"></a>
### Nested Schema for `widgets.chart`

Read-Only:

- `chart_id` (String)
- `description` (String)
- `display_legend` (Boolean)
- `freeze` (String)
- `name_hiding_settings` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--name_hiding_settings))
- `queries` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--queries))
- `series_overrides` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--series_overrides))
- `title` (String)
- `visualization_settings` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings))

<a id="nestedobjatt--widgets--chart--name_hiding_settings"></a>
### Nested Schema for `widgets.chart.name_hiding_settings`

Read-Only:

- `names` (List of String)
- `positive` (Boolean)


<a id="nestedobjatt--widgets--chart--queries"></a>
### Nested Schema for `widgets.chart.queries`

Read-Only:

- `downsampling` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--queries--downsampling))
- `target` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--queries--target))

<a id="nestedobjatt--widgets--chart--queries--downsampling"></a>
### Nested Schema for `widgets.chart.queries.downsampling`

Read-Only:

- `disabled` (Boolean)
- `gap_filling` (String)
- `grid_aggregation` (String)
- `grid_interval` (Number)
- `max_points` (Number)


<a id="nestedobjatt--widgets--chart--queries--target"></a>
### Nested Schema for `widgets.chart.queries.target`

Read-Only:

- `hidden` (Boolean)
- `query` (String)
- `text_mode` (Boolean)



<a id="nestedobjatt--widgets--chart--series_overrides"></a>
### Nested Schema for `widgets.chart.series_overrides`

Read-Only:

- `name` (String)
- `settings` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--series_overrides--settings))
- `target_index` (String)

<a id="nestedobjatt--widgets--chart--series_overrides--settings"></a>
### Nested Schema for `widgets.chart.series_overrides.settings`

Read-Only:

- `color` (String)
- `grow_down` (Boolean)
- `name` (String)
- `stack_name` (String)
- `type` (String)
- `yaxis_position` (String)



<a id="nestedobjatt--widgets--chart--visualization_settings"></a>
### Nested Schema for `widgets.chart.visualization_settings`

Read-Only:

- `aggregation` (String)
- `color_scheme_settings` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--color_scheme_settings))
- `heatmap_settings` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--heatmap_settings))
- `interpolate` (String)
- `normalize` (Boolean)
- `show_labels` (Boolean)
- `title` (String)
- `type` (String)
- `yaxis_settings` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings))

<a id="nestedobjatt--widgets--chart--visualization_settings--color_scheme_settings"></a>
### Nested Schema for `widgets.chart.visualization_settings.color_scheme_settings`

Read-Only:

- `automatic` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--automatic))
- `gradient` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--gradient))
- `standard` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--standard))

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

- `green_value` (String)
- `red_value` (String)
- `violet_value` (String)
- `yellow_value` (String)


<a id="nestedobjatt--widgets--chart--visualization_settings--yaxis_settings"></a>
### Nested Schema for `widgets.chart.visualization_settings.yaxis_settings`

Read-Only:

- `left` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--left))
- `right` (List of Object) (see [below for nested schema](#nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--right))

<a id="nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--left"></a>
### Nested Schema for `widgets.chart.visualization_settings.yaxis_settings.left`

Read-Only:

- `max` (String)
- `min` (String)
- `precision` (Number)
- `title` (String)
- `type` (String)
- `unit_format` (String)


<a id="nestedobjatt--widgets--chart--visualization_settings--yaxis_settings--right"></a>
### Nested Schema for `widgets.chart.visualization_settings.yaxis_settings.right`

Read-Only:

- `max` (String)
- `min` (String)
- `precision` (Number)
- `title` (String)
- `type` (String)
- `unit_format` (String)





<a id="nestedobjatt--widgets--position"></a>
### Nested Schema for `widgets.position`

Read-Only:

- `h` (Number)
- `w` (Number)
- `x` (Number)
- `y` (Number)


<a id="nestedobjatt--widgets--text"></a>
### Nested Schema for `widgets.text`

Read-Only:

- `text` (String)


<a id="nestedobjatt--widgets--title"></a>
### Nested Schema for `widgets.title`

Read-Only:

- `size` (String)
- `text` (String)
