package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/monitoring/v3"
)

func parseMonitoringUnitFormat(str string) (monitoring.UnitFormat, error) {
	val, ok := monitoring.UnitFormat_value[str]
	if !ok {
		return monitoring.UnitFormat(0), invalidKeyError("unit_format", monitoring.UnitFormat_value, str)
	}
	return monitoring.UnitFormat(val), nil
}

func parseMonitoringChartWidgetXFreezeDuration(str string) (monitoring.ChartWidget_FreezeDuration, error) {
	val, ok := monitoring.ChartWidget_FreezeDuration_value[str]
	if !ok {
		return monitoring.ChartWidget_FreezeDuration(0), invalidKeyError("freeze_duration", monitoring.ChartWidget_FreezeDuration_value, str)
	}
	return monitoring.ChartWidget_FreezeDuration(val), nil
}

func parseMonitoringDownsamplingXGapFilling(str string) (monitoring.Downsampling_GapFilling, error) {
	val, ok := monitoring.Downsampling_GapFilling_value[str]
	if !ok {
		return monitoring.Downsampling_GapFilling(0), invalidKeyError("gap_filling", monitoring.Downsampling_GapFilling_value, str)
	}
	return monitoring.Downsampling_GapFilling(val), nil
}

func parseMonitoringDownsamplingXGridAggregation(str string) (monitoring.Downsampling_GridAggregation, error) {
	val, ok := monitoring.Downsampling_GridAggregation_value[str]
	if !ok {
		return monitoring.Downsampling_GridAggregation(0), invalidKeyError("grid_aggregation", monitoring.Downsampling_GridAggregation_value, str)
	}
	return monitoring.Downsampling_GridAggregation(val), nil
}

func parseMonitoringChartWidgetXSeriesOverridesXSeriesVisualizationType(str string) (monitoring.ChartWidget_SeriesOverrides_SeriesVisualizationType, error) {
	val, ok := monitoring.ChartWidget_SeriesOverrides_SeriesVisualizationType_value[str]
	if !ok {
		return monitoring.ChartWidget_SeriesOverrides_SeriesVisualizationType(0), invalidKeyError("series_visualization_type", monitoring.ChartWidget_SeriesOverrides_SeriesVisualizationType_value, str)
	}
	return monitoring.ChartWidget_SeriesOverrides_SeriesVisualizationType(val), nil
}

func parseMonitoringChartWidgetXSeriesOverridesXYaxisPosition(str string) (monitoring.ChartWidget_SeriesOverrides_YaxisPosition, error) {
	val, ok := monitoring.ChartWidget_SeriesOverrides_YaxisPosition_value[str]
	if !ok {
		return monitoring.ChartWidget_SeriesOverrides_YaxisPosition(0), invalidKeyError("yaxis_position", monitoring.ChartWidget_SeriesOverrides_YaxisPosition_value, str)
	}
	return monitoring.ChartWidget_SeriesOverrides_YaxisPosition(val), nil
}

func parseMonitoringChartWidgetXVisualizationSettingsXSeriesAggregation(str string) (monitoring.ChartWidget_VisualizationSettings_SeriesAggregation, error) {
	val, ok := monitoring.ChartWidget_VisualizationSettings_SeriesAggregation_value[str]
	if !ok {
		return monitoring.ChartWidget_VisualizationSettings_SeriesAggregation(0), invalidKeyError("series_aggregation", monitoring.ChartWidget_VisualizationSettings_SeriesAggregation_value, str)
	}
	return monitoring.ChartWidget_VisualizationSettings_SeriesAggregation(val), nil
}

func parseMonitoringChartWidgetXVisualizationSettingsXInterpolate(str string) (monitoring.ChartWidget_VisualizationSettings_Interpolate, error) {
	val, ok := monitoring.ChartWidget_VisualizationSettings_Interpolate_value[str]
	if !ok {
		return monitoring.ChartWidget_VisualizationSettings_Interpolate(0), invalidKeyError("interpolate", monitoring.ChartWidget_VisualizationSettings_Interpolate_value, str)
	}
	return monitoring.ChartWidget_VisualizationSettings_Interpolate(val), nil
}

func parseMonitoringChartWidgetXVisualizationSettingsXVisualizationType(str string) (monitoring.ChartWidget_VisualizationSettings_VisualizationType, error) {
	val, ok := monitoring.ChartWidget_VisualizationSettings_VisualizationType_value[str]
	if !ok {
		return monitoring.ChartWidget_VisualizationSettings_VisualizationType(0), invalidKeyError("visualization_type", monitoring.ChartWidget_VisualizationSettings_VisualizationType_value, str)
	}
	return monitoring.ChartWidget_VisualizationSettings_VisualizationType(val), nil
}

func parseMonitoringChartWidgetXVisualizationSettingsXYaxisType(str string) (monitoring.ChartWidget_VisualizationSettings_YaxisType, error) {
	val, ok := monitoring.ChartWidget_VisualizationSettings_YaxisType_value[str]
	if !ok {
		return monitoring.ChartWidget_VisualizationSettings_YaxisType(0), invalidKeyError("yaxis_type", monitoring.ChartWidget_VisualizationSettings_YaxisType_value, str)
	}
	return monitoring.ChartWidget_VisualizationSettings_YaxisType(val), nil
}

func parseMonitoringTitleWidgetXTitleSize(str string) (monitoring.TitleWidget_TitleSize, error) {
	val, ok := monitoring.TitleWidget_TitleSize_value[str]
	if !ok {
		return monitoring.TitleWidget_TitleSize(0), invalidKeyError("title_size", monitoring.TitleWidget_TitleSize_value, str)
	}
	return monitoring.TitleWidget_TitleSize(val), nil
}

func expandDashboardWidgetsSlice(d *schema.ResourceData) ([]*monitoring.Widget, error) {
	count := d.Get("widgets.#").(int)
	slice := make([]*monitoring.Widget, count)

	for i := 0; i < count; i++ {
		widgets, err := expandDashboardWidgets(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = widgets
	}

	return slice, nil
}

func expandDashboardWidgets(d *schema.ResourceData, indexes ...interface{}) (*monitoring.Widget, error) {
	val := new(monitoring.Widget)

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.position", indexes...)); ok {
		position, err := expandDashboardWidgetsPosition(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPosition(position)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.text", indexes...)); ok {
		text, err := expandDashboardWidgetsText(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetText(text)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.title", indexes...)); ok {
		title, err := expandDashboardWidgetsTitle(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetTitle(title)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart", indexes...)); ok {
		chart, err := expandDashboardWidgetsChart(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetChart(chart)
	}

	return val, nil
}

func expandDashboardWidgetsPosition(d *schema.ResourceData, indexes ...interface{}) (*monitoring.Widget_LayoutPosition, error) {
	val := new(monitoring.Widget_LayoutPosition)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.position.0.x", indexes...)); ok {
		val.SetX(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.position.0.y", indexes...)); ok {
		val.SetY(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.position.0.w", indexes...)); ok {
		val.SetW(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.position.0.h", indexes...)); ok {
		val.SetH(int64(v.(int)))
	}

	return val, nil
}

func expandDashboardWidgetsText(d *schema.ResourceData, indexes ...interface{}) (*monitoring.TextWidget, error) {
	val := new(monitoring.TextWidget)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.text.0.text", indexes...)); ok {
		val.SetText(v.(string))
	}

	return val, nil
}

func expandDashboardWidgetsTitle(d *schema.ResourceData, indexes ...interface{}) (*monitoring.TitleWidget, error) {
	val := new(monitoring.TitleWidget)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.title.0.text", indexes...)); ok {
		val.SetText(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.title.0.size", indexes...)); ok {
		titleSize, err := parseMonitoringTitleWidgetXTitleSize(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetSize(titleSize)
	}

	return val, nil
}

func expandDashboardWidgetsChart(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget, error) {
	val := new(monitoring.ChartWidget)

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.queries", indexes...)); ok {
		queries, err := expandDashboardWidgetsChartQueries(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings", indexes...)); ok {
		visualizationSettings, err := expandDashboardWidgetsChartVisualizationSettings(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetVisualizationSettings(visualizationSettings)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.series_overrides", indexes...)); ok {
		seriesOverrides, err := expandDashboardWidgetsChartSeriesOverridesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSeriesOverrides(seriesOverrides)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.name_hiding_settings", indexes...)); ok {
		nameHidingSettings, err := expandDashboardWidgetsChartNameHidingSettings(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetNameHidingSettings(nameHidingSettings)
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.chart_id", indexes...)); ok {
		val.SetId(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.title", indexes...)); ok {
		val.SetTitle(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.display_legend", indexes...)); ok {
		val.SetDisplayLegend(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.freeze", indexes...)); ok {
		freezeDuration, err := parseMonitoringChartWidgetXFreezeDuration(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetFreeze(freezeDuration)
	}

	return val, nil
}

func expandDashboardWidgetsChartQueries(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_Queries, error) {
	val := new(monitoring.ChartWidget_Queries)

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.queries.0.target", indexes...)); ok {
		targets, err := expandDashboardWidgetsChartQueriesTargetsSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetTargets(targets)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.queries.0.downsampling", indexes...)); ok {
		downsampling, err := expandDashboardWidgetsChartQueriesDownsampling(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetDownsampling(downsampling)
	}

	return val, nil
}

func expandDashboardWidgetsChartQueriesTargetsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*monitoring.ChartWidget_Queries_Target, error) {
	count := d.Get(fmt.Sprintf("widgets.%d.chart.0.queries.0.target.#", indexes...)).(int)
	slice := make([]*monitoring.ChartWidget_Queries_Target, count)

	for i := 0; i < count; i++ {
		targets, err := expandDashboardWidgetsChartQueriesTargets(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = targets
	}

	return slice, nil
}

func expandDashboardWidgetsChartQueriesTargets(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_Queries_Target, error) {
	val := new(monitoring.ChartWidget_Queries_Target)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.queries.0.target.%d.query", indexes...)); ok {
		val.SetQuery(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.queries.0.target.%d.text_mode", indexes...)); ok {
		val.SetTextMode(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.queries.0.target.%d.hidden", indexes...)); ok {
		val.SetHidden(v.(bool))
	}

	return val, nil
}

func expandDashboardWidgetsChartQueriesDownsampling(d *schema.ResourceData, indexes ...interface{}) (*monitoring.Downsampling, error) {
	val := new(monitoring.Downsampling)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.queries.0.downsampling.0.max_points", indexes...)); ok {
		val.SetMaxPoints(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.queries.0.downsampling.0.grid_interval", indexes...)); ok {
		val.SetGridInterval(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.queries.0.downsampling.0.disabled", indexes...)); ok {
		val.SetDisabled(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.queries.0.downsampling.0.grid_aggregation", indexes...)); ok {
		gridAggregation, err := parseMonitoringDownsamplingXGridAggregation(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetGridAggregation(gridAggregation)
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.queries.0.downsampling.0.gap_filling", indexes...)); ok {
		gapFilling, err := parseMonitoringDownsamplingXGapFilling(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetGapFilling(gapFilling)
	}

	return val, nil
}

func expandDashboardWidgetsChartVisualizationSettings(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_VisualizationSettings, error) {
	val := new(monitoring.ChartWidget_VisualizationSettings)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.type", indexes...)); ok {
		visualizationType, err := parseMonitoringChartWidgetXVisualizationSettingsXVisualizationType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(visualizationType)
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.normalize", indexes...)); ok {
		val.SetNormalize(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.interpolate", indexes...)); ok {
		interpolate, err := parseMonitoringChartWidgetXVisualizationSettingsXInterpolate(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetInterpolate(interpolate)
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.aggregation", indexes...)); ok {
		seriesAggregation, err := parseMonitoringChartWidgetXVisualizationSettingsXSeriesAggregation(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetAggregation(seriesAggregation)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.color_scheme_settings", indexes...)); ok {
		colorSchemeSettings, err := expandDashboardWidgetsChartVisualizationSettingsColorSchemeSettings(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetColorSchemeSettings(colorSchemeSettings)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.heatmap_settings", indexes...)); ok {
		heatmapSettings, err := expandDashboardWidgetsChartVisualizationSettingsHeatmapSettings(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeatmapSettings(heatmapSettings)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings", indexes...)); ok {
		yaxisSettings, err := expandDashboardWidgetsChartVisualizationSettingsYaxisSettings(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetYaxisSettings(yaxisSettings)
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.title", indexes...)); ok {
		val.SetTitle(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.show_labels", indexes...)); ok {
		val.SetShowLabels(v.(bool))
	}

	return val, nil
}

func expandDashboardWidgetsChartVisualizationSettingsColorSchemeSettings(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings, error) {
	val := new(monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings)

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.color_scheme_settings.0.automatic", indexes...)); ok {
		automatic, err := expandDashboardWidgetsChartVisualizationSettingsColorSchemeSettingsAutomatic()
		if err != nil {
			return nil, err
		}

		val.SetAutomatic(automatic)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.color_scheme_settings.0.standard", indexes...)); ok {
		standard, err := expandDashboardWidgetsChartVisualizationSettingsColorSchemeSettingsStandard()
		if err != nil {
			return nil, err
		}

		val.SetStandard(standard)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.color_scheme_settings.0.gradient", indexes...)); ok {
		gradient, err := expandDashboardWidgetsChartVisualizationSettingsColorSchemeSettingsGradient(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGradient(gradient)
	}

	return val, nil
}

func expandDashboardWidgetsChartVisualizationSettingsColorSchemeSettingsAutomatic() (*monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings_AutomaticColorScheme, error) {
	val := new(monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings_AutomaticColorScheme)
	return val, nil
}

func expandDashboardWidgetsChartVisualizationSettingsColorSchemeSettingsStandard() (*monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings_StandardColorScheme, error) {
	val := new(monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings_StandardColorScheme)
	return val, nil
}

func expandDashboardWidgetsChartVisualizationSettingsColorSchemeSettingsGradient(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings_GradientColorScheme, error) {
	val := new(monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings_GradientColorScheme)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.green_value", indexes...)); ok {
		val.SetGreenValue(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.yellow_value", indexes...)); ok {
		val.SetYellowValue(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.red_value", indexes...)); ok {
		val.SetRedValue(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.color_scheme_settings.0.gradient.0.violet_value", indexes...)); ok {
		val.SetVioletValue(v.(string))
	}

	return val, nil
}

func expandDashboardWidgetsChartVisualizationSettingsHeatmapSettings(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_VisualizationSettings_HeatmapSettings, error) {
	val := new(monitoring.ChartWidget_VisualizationSettings_HeatmapSettings)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.heatmap_settings.0.green_value", indexes...)); ok {
		val.SetGreenValue(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.heatmap_settings.0.yellow_value", indexes...)); ok {
		val.SetYellowValue(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.heatmap_settings.0.red_value", indexes...)); ok {
		val.SetRedValue(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.heatmap_settings.0.violet_value", indexes...)); ok {
		val.SetVioletValue(v.(string))
	}

	return val, nil
}

func expandDashboardWidgetsChartVisualizationSettingsYaxisSettings(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_VisualizationSettings_YaxisSettings, error) {
	val := new(monitoring.ChartWidget_VisualizationSettings_YaxisSettings)

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.left", indexes...)); ok {
		left, err := expandDashboardWidgetsChartVisualizationSettingsYaxisSettingsLeft(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetLeft(left)
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.right", indexes...)); ok {
		right, err := expandDashboardWidgetsChartVisualizationSettingsYaxisSettingsRight(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRight(right)
	}

	return val, nil
}

func expandDashboardWidgetsChartVisualizationSettingsYaxisSettingsLeft(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_VisualizationSettings_Yaxis, error) {
	val := new(monitoring.ChartWidget_VisualizationSettings_Yaxis)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.left.0.type", indexes...)); ok {
		yaxisType, err := parseMonitoringChartWidgetXVisualizationSettingsXYaxisType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(yaxisType)
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.left.0.title", indexes...)); ok {
		val.SetTitle(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.left.0.min", indexes...)); ok {
		val.SetMin(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.left.0.max", indexes...)); ok {
		val.SetMax(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.left.0.unit_format", indexes...)); ok {
		unitFormat, err := parseMonitoringUnitFormat(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetUnitFormat(unitFormat)
	}

	if v, ok := d.GetOkExists(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.left.0.precision", indexes...)); ok {
		val.SetPrecision(&wrapperspb.Int64Value{
			Value: int64(v.(int)),
		})
	}

	return val, nil
}

func expandDashboardWidgetsChartVisualizationSettingsYaxisSettingsRight(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_VisualizationSettings_Yaxis, error) {
	val := new(monitoring.ChartWidget_VisualizationSettings_Yaxis)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.right.0.type", indexes...)); ok {
		yaxisType, err := parseMonitoringChartWidgetXVisualizationSettingsXYaxisType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(yaxisType)
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.right.0.title", indexes...)); ok {
		val.SetTitle(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.right.0.min", indexes...)); ok {
		val.SetMin(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.right.0.max", indexes...)); ok {
		val.SetMax(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.right.0.unit_format", indexes...)); ok {
		unitFormat, err := parseMonitoringUnitFormat(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetUnitFormat(unitFormat)
	}

	if v, ok := d.GetOkExists(fmt.Sprintf("widgets.%d.chart.0.visualization_settings.0.yaxis_settings.0.right.0.precision", indexes...)); ok {
		val.SetPrecision(&wrapperspb.Int64Value{
			Value: int64(v.(int)),
		})
	}

	return val, nil
}

func expandDashboardWidgetsChartSeriesOverridesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*monitoring.ChartWidget_SeriesOverrides, error) {
	count := d.Get(fmt.Sprintf("widgets.%d.chart.0.series_overrides.#", indexes...)).(int)
	slice := make([]*monitoring.ChartWidget_SeriesOverrides, count)

	for i := 0; i < count; i++ {
		seriesOverrides, err := expandDashboardWidgetsChartSeriesOverrides(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = seriesOverrides
	}

	return slice, nil
}

func expandDashboardWidgetsChartSeriesOverrides(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_SeriesOverrides, error) {
	val := new(monitoring.ChartWidget_SeriesOverrides)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.series_overrides.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.series_overrides.%d.target_index", indexes...)); ok {
		val.SetTargetIndex(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.series_overrides.%d.settings", indexes...)); ok {
		settings, err := expandDashboardWidgetsChartSeriesOverridesSettings(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSettings(settings)
	}

	return val, nil
}

func expandDashboardWidgetsChartSeriesOverridesSettings(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_SeriesOverrides_SeriesOverrideSettings, error) {
	val := new(monitoring.ChartWidget_SeriesOverrides_SeriesOverrideSettings)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.series_overrides.%d.settings.0.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.series_overrides.%d.settings.0.color", indexes...)); ok {
		val.SetColor(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.series_overrides.%d.settings.0.type", indexes...)); ok {
		seriesVisualizationType, err := parseMonitoringChartWidgetXSeriesOverridesXSeriesVisualizationType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(seriesVisualizationType)
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.series_overrides.%d.settings.0.stack_name", indexes...)); ok {
		val.SetStackName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.series_overrides.%d.settings.0.grow_down", indexes...)); ok {
		val.SetGrowDown(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.series_overrides.%d.settings.0.yaxis_position", indexes...)); ok {
		yaxisPosition, err := parseMonitoringChartWidgetXSeriesOverridesXYaxisPosition(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetYaxisPosition(yaxisPosition)
	}

	return val, nil
}

func expandDashboardWidgetsChartNameHidingSettings(d *schema.ResourceData, indexes ...interface{}) (*monitoring.ChartWidget_NameHidingSettings, error) {
	val := new(monitoring.ChartWidget_NameHidingSettings)

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.name_hiding_settings.0.positive", indexes...)); ok {
		val.SetPositive(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("widgets.%d.chart.0.name_hiding_settings.0.names", indexes...)); ok {
		names := expandStringSlice(v.([]interface{}))
		val.SetNames(names)
	}

	return val, nil
}

func expandDashboardParametrization(d *schema.ResourceData) (*monitoring.Parametrization, error) {
	val := new(monitoring.Parametrization)

	if _, ok := d.GetOk("parametrization.0.parameters"); ok {
		parameters, err := expandDashboardParametrizationParametersSlice(d)
		if err != nil {
			return nil, err
		}

		val.SetParameters(parameters)
	}

	if v, ok := d.GetOk("parametrization.0.selectors"); ok {
		val.SetSelectors(v.(string))
	}

	empty := new(monitoring.Parametrization)
	if proto.Equal(val, empty) {
		return nil, nil
	}

	return val, nil
}

func expandDashboardParametrizationParametersSlice(d *schema.ResourceData) ([]*monitoring.Parameter, error) {
	count := d.Get("parametrization.0.parameters.#").(int)
	slice := make([]*monitoring.Parameter, count)

	for i := 0; i < count; i++ {
		parameters, err := expandDashboardParametrizationParameters(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = parameters
	}

	return slice, nil
}

func expandDashboardParametrizationParameters(d *schema.ResourceData, indexes ...interface{}) (*monitoring.Parameter, error) {
	val := new(monitoring.Parameter)

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.id", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.title", indexes...)); ok {
		val.SetTitle(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.label_values", indexes...)); ok {
		labelValues, err := expandDashboardParametrizationParametersLabelValues(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetLabelValues(labelValues)
	}

	if _, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.custom", indexes...)); ok {
		custom, err := expandDashboardParametrizationParametersCustom(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCustom(custom)
	}

	if _, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.text", indexes...)); ok {
		text, err := expandDashboardParametrizationParametersText(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetText(text)
	}

	if _, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.integer_parameter", indexes...)); ok {
		integerParameter, err := expandDashboardParametrizationParametersIntegerParameter(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIntegerParameter(integerParameter)
	}

	if _, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.double_parameter", indexes...)); ok {
		doubleParameter, err := expandDashboardParametrizationParametersDoubleParameter(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetDoubleParameter(doubleParameter)
	}

	if _, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.text_values", indexes...)); ok {
		textValues, err := expandDashboardParametrizationParametersTextValues(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetTextValues(textValues)
	}

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.hidden", indexes...)); ok {
		val.SetHidden(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	return val, nil
}

func expandDashboardParametrizationParametersLabelValues(d *schema.ResourceData, indexes ...interface{}) (*monitoring.LabelValuesParameter, error) {
	val := new(monitoring.LabelValuesParameter)

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.label_values.0.folder_id", indexes...)); ok {
		val.SetFolderId(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.label_values.0.selectors", indexes...)); ok {
		val.SetSelectors(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.label_values.0.label_key", indexes...)); ok {
		val.SetLabelKey(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.label_values.0.multiselectable", indexes...)); ok {
		val.SetMultiselectable(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.label_values.0.default_values", indexes...)); ok {
		defaultValues := expandStringSlice(v.([]interface{}))
		val.SetDefaultValues(defaultValues)
	}

	return val, nil
}

func expandDashboardParametrizationParametersCustom(d *schema.ResourceData, indexes ...interface{}) (*monitoring.CustomParameter, error) {
	val := new(monitoring.CustomParameter)

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.custom.0.values", indexes...)); ok {
		values := expandStringSlice(v.([]interface{}))
		val.SetValues(values)
	}

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.custom.0.multiselectable", indexes...)); ok {
		val.SetMultiselectable(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.custom.0.default_values", indexes...)); ok {
		defaultValues := expandStringSlice(v.([]interface{}))
		val.SetDefaultValues(defaultValues)
	}

	return val, nil
}

func expandDashboardParametrizationParametersText(d *schema.ResourceData, indexes ...interface{}) (*monitoring.TextParameter, error) {
	val := new(monitoring.TextParameter)

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.text.0.default_value", indexes...)); ok {
		val.SetDefaultValue(v.(string))
	}

	return val, nil
}

func expandDashboardParametrizationParametersIntegerParameter(d *schema.ResourceData, indexes ...interface{}) (*monitoring.IntegerParameter, error) {
	val := new(monitoring.IntegerParameter)

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.integer_parameter.0.default_value", indexes...)); ok {
		val.SetDefaultValue(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.integer_parameter.0.unit_format", indexes...)); ok {
		unitFormat, err := parseMonitoringUnitFormat(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetUnitFormat(unitFormat)
	}

	return val, nil
}

func expandDashboardParametrizationParametersDoubleParameter(d *schema.ResourceData, indexes ...interface{}) (*monitoring.DoubleParameter, error) {
	val := new(monitoring.DoubleParameter)

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.double_parameter.0.default_value", indexes...)); ok {
		val.SetDefaultValue(v.(float64))
	}

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.double_parameter.0.unit_format", indexes...)); ok {
		unitFormat, err := parseMonitoringUnitFormat(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetUnitFormat(unitFormat)
	}

	return val, nil
}

func expandDashboardParametrizationParametersTextValues(d *schema.ResourceData, indexes ...interface{}) (*monitoring.TextValuesParameter, error) {
	val := new(monitoring.TextValuesParameter)

	if v, ok := d.GetOk(fmt.Sprintf("parametrization.0.parameters.%d.text_values.0.default_values", indexes...)); ok {
		defaultValues := expandStringSlice(v.([]interface{}))
		val.SetDefaultValues(defaultValues)
	}

	return val, nil
}

func flattenMonitoringParametrization(v *monitoring.Parametrization) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	parameter, err := flattenMonitoringParametrizationParameterSlice(v.Parameters)
	if err != nil {
		return nil, err
	}
	m["parameters"] = parameter
	m["selectors"] = v.Selectors

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringParametrizationParameterSlice(vs []*monitoring.Parameter) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		parameter, err := flattenMonitoringParameter(v)
		if err != nil {
			return nil, err
		}

		if len(parameter) != 0 {
			s = append(s, parameter[0])
		}
	}

	return s, nil
}

func flattenMonitoringParameter(v *monitoring.Parameter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	custom, err := flattenMonitoringCustomParameter(v.GetCustom())
	if err != nil {
		return nil, err
	}
	m["custom"] = custom
	m["description"] = v.Description
	m["hidden"] = v.Hidden
	labelValues, err := flattenMonitoringLabelValuesParameter(v.GetLabelValues())
	if err != nil {
		return nil, err
	}
	m["label_values"] = labelValues
	m["id"] = v.Name
	text, err := flattenMonitoringTextParameter(v.GetText())
	if err != nil {
		return nil, err
	}
	m["text"] = text
	m["title"] = v.Title

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringCustomParameter(v *monitoring.CustomParameter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["default_values"] = v.DefaultValues
	m["multiselectable"] = v.Multiselectable
	m["values"] = v.Values

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringLabelValuesParameter(v *monitoring.LabelValuesParameter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["default_values"] = v.DefaultValues
	m["folder_id"] = v.GetFolderId()
	m["label_key"] = v.LabelKey
	m["multiselectable"] = v.Multiselectable
	m["selectors"] = v.Selectors

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringTextParameter(v *monitoring.TextParameter) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["default_value"] = v.DefaultValue

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringWidgetSlice(vs []*monitoring.Widget) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		widget, err := flattenMonitoringWidget(v)
		if err != nil {
			return nil, err
		}

		if len(widget) != 0 {
			s = append(s, widget[0])
		}
	}

	return s, nil
}

func flattenMonitoringWidget(v *monitoring.Widget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})
	chart, err := flattenMonitoringChartWidget(v.GetChart())
	if err != nil {
		return nil, err
	}
	m["chart"] = chart
	position, err := flattenMonitoringWidgetLayoutPosition(v.Position)
	if err != nil {
		return nil, err
	}
	m["position"] = position
	text, err := flattenMonitoringTextWidget(v.GetText())
	if err != nil {
		return nil, err
	}
	m["text"] = text
	title, err := flattenMonitoringTitleWidget(v.GetTitle())
	if err != nil {
		return nil, err
	}
	m["title"] = title

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringChartWidget(v *monitoring.ChartWidget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["description"] = v.Description
	m["display_legend"] = v.DisplayLegend
	m["freeze"] = v.Freeze.String()
	nameHidingSettings, err := flattenMonitoringChartWidgetNameHidingSettings(v.NameHidingSettings)
	if err != nil {
		return nil, err
	}
	m["name_hiding_settings"] = nameHidingSettings
	queries, err := flattenMonitoringChartWidgetQueries(v.Queries)
	if err != nil {
		return nil, err
	}
	m["queries"] = queries
	seriesOverrides, err := flattenMonitoringWidgetChartSeriesOverridesSlice(v.SeriesOverrides)
	if err != nil {
		return nil, err
	}
	m["series_overrides"] = seriesOverrides
	m["title"] = v.Title
	m["chart_id"] = v.Id
	visualizationSettings, err := flattenMonitoringChartWidgetVisualizationSettings(v.VisualizationSettings)
	if err != nil {
		return nil, err
	}
	m["visualization_settings"] = visualizationSettings

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringChartWidgetNameHidingSettings(v *monitoring.ChartWidget_NameHidingSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["names"] = v.Names
	m["positive"] = v.Positive

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringChartWidgetQueries(v *monitoring.ChartWidget_Queries) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	downsampling, err := flattenMonitoringDownsampling(v.Downsampling)
	if err != nil {
		return nil, err
	}
	m["downsampling"] = downsampling
	target, err := flattenMonitoringWidgetChartQueriesTargetSlice(v.Targets)
	if err != nil {
		return nil, err
	}
	m["target"] = target

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringDownsampling(v *monitoring.Downsampling) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["disabled"] = v.GetDisabled()
	m["gap_filling"] = v.GapFilling.String()
	m["grid_aggregation"] = v.GridAggregation.String()
	m["grid_interval"] = v.GetGridInterval()
	m["max_points"] = v.GetMaxPoints()

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringWidgetChartQueriesTargetSlice(vs []*monitoring.ChartWidget_Queries_Target) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		target, err := flattenMonitoringChartWidgetQueriesTarget(v)
		if err != nil {
			return nil, err
		}

		if len(target) != 0 {
			s = append(s, target[0])
		}
	}

	return s, nil
}

func flattenMonitoringChartWidgetQueriesTarget(v *monitoring.ChartWidget_Queries_Target) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["hidden"] = v.Hidden
	m["query"] = v.Query
	m["text_mode"] = v.TextMode

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringWidgetChartSeriesOverridesSlice(vs []*monitoring.ChartWidget_SeriesOverrides) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		seriesOverrides, err := flattenMonitoringChartWidgetSeriesOverrides(v)
		if err != nil {
			return nil, err
		}

		if len(seriesOverrides) != 0 {
			s = append(s, seriesOverrides[0])
		}
	}

	return s, nil
}

func flattenMonitoringChartWidgetSeriesOverrides(v *monitoring.ChartWidget_SeriesOverrides) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["name"] = v.GetName()
	settings, err := flattenMonitoringChartWidgetSeriesOverridesSeriesOverrideSettings(v.Settings)
	if err != nil {
		return nil, err
	}
	m["settings"] = settings
	m["target_index"] = v.GetTargetIndex()

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringChartWidgetSeriesOverridesSeriesOverrideSettings(v *monitoring.ChartWidget_SeriesOverrides_SeriesOverrideSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["color"] = v.Color
	m["grow_down"] = v.GrowDown
	m["name"] = v.Name
	m["stack_name"] = v.StackName
	m["type"] = v.Type.String()
	m["yaxis_position"] = v.YaxisPosition.String()

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringChartWidgetVisualizationSettings(v *monitoring.ChartWidget_VisualizationSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["aggregation"] = v.Aggregation.String()
	colorSchemeSettings, err := flattenMonitoringChartWidgetVisualizationSettingsColorSchemeSettings(v.ColorSchemeSettings)
	if err != nil {
		return nil, err
	}
	m["color_scheme_settings"] = colorSchemeSettings
	heatmapSettings, err := flattenMonitoringChartWidgetVisualizationSettingsHeatmapSettings(v.HeatmapSettings)
	if err != nil {
		return nil, err
	}
	m["heatmap_settings"] = heatmapSettings
	m["interpolate"] = v.Interpolate.String()
	m["normalize"] = v.Normalize
	m["show_labels"] = v.ShowLabels
	m["title"] = v.Title
	m["type"] = v.Type.String()
	yaxisSettings, err := flattenMonitoringChartWidgetVisualizationSettingsYaxisSettings(v.YaxisSettings)
	if err != nil {
		return nil, err
	}
	m["yaxis_settings"] = yaxisSettings

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringChartWidgetVisualizationSettingsColorSchemeSettings(v *monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	automatic, err := ChartWidgetVisualizationSettingsColorSchemeSettingsAutomaticColorScheme(v.GetAutomatic())
	if err != nil {
		return nil, err
	}
	m["automatic"] = automatic
	gradient, err := ChartWidgetVisualizationSettingsColorSchemeSettingsGradientColorScheme(v.GetGradient())
	if err != nil {
		return nil, err
	}
	m["gradient"] = gradient
	standard, err := ChartWidgetVisualizationSettingsColorSchemeSettingsStandardColorScheme(v.GetStandard())
	if err != nil {
		return nil, err
	}
	m["standard"] = standard

	return []map[string]interface{}{m}, nil
}

func ChartWidgetVisualizationSettingsColorSchemeSettingsAutomaticColorScheme(v *monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings_AutomaticColorScheme) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func ChartWidgetVisualizationSettingsColorSchemeSettingsGradientColorScheme(v *monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings_GradientColorScheme) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["green_value"] = v.GreenValue
	m["red_value"] = v.RedValue
	m["violet_value"] = v.VioletValue
	m["yellow_value"] = v.YellowValue

	return []map[string]interface{}{m}, nil
}

func ChartWidgetVisualizationSettingsColorSchemeSettingsStandardColorScheme(v *monitoring.ChartWidget_VisualizationSettings_ColorSchemeSettings_StandardColorScheme) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringChartWidgetVisualizationSettingsHeatmapSettings(v *monitoring.ChartWidget_VisualizationSettings_HeatmapSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["green_value"] = v.GreenValue
	m["red_value"] = v.RedValue
	m["violet_value"] = v.VioletValue
	m["yellow_value"] = v.YellowValue

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringChartWidgetVisualizationSettingsYaxisSettings(v *monitoring.ChartWidget_VisualizationSettings_YaxisSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	left, err := flattenMonitoringChartWidgetVisualizationSettingsYaxis(v.Left)
	if err != nil {
		return nil, err
	}
	m["left"] = left
	right, err := flattenMonitoringChartWidgetVisualizationSettingsYaxis(v.Right)
	if err != nil {
		return nil, err
	}
	m["right"] = right

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringChartWidgetVisualizationSettingsYaxis(v *monitoring.ChartWidget_VisualizationSettings_Yaxis) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["max"] = v.Max
	m["min"] = v.Min
	var precision interface{}
	if v.Precision != nil {
		precision = v.Precision.GetValue()
	}
	m["precision"] = precision
	m["title"] = v.Title
	m["type"] = v.Type.String()
	m["unit_format"] = v.UnitFormat.String()

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringWidgetLayoutPosition(v *monitoring.Widget_LayoutPosition) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["h"] = v.H
	m["w"] = v.W
	m["x"] = v.X
	m["y"] = v.Y

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringTextWidget(v *monitoring.TextWidget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["text"] = v.Text

	return []map[string]interface{}{m}, nil
}

func flattenMonitoringTitleWidget(v *monitoring.TitleWidget) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["size"] = v.Size.String()
	m["text"] = v.Text

	return []map[string]interface{}{m}, nil
}
