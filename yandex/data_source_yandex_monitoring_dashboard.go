package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/monitoring/v3"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"log"
)

func dataSourceYandexMonitoringDashboard() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexMonitoringDashboardRead,
		Schema: map[string]*schema.Schema{
			"dashboard_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Dashboard ID",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Dashboard description",
			},
			"folder_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Folder ID",
			},
			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:         schema.HashString,
				Computed:    true,
				Description: "Dashboard labels",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Dashboard name, used as local identifier in folder_id",
			},
			"parametrization": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parameters": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"custom": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"default_values": {
													Type: schema.TypeList,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
													Computed:    true,
													Description: "Default values from values",
												},
												"multiselectable": {
													Type:        schema.TypeBool,
													Computed:    true,
													Description: "Specifies the multiselectable values of parameter",
												},
												"values": {
													Type: schema.TypeList,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
													Computed:    true,
													Description: "Parameter values",
												},
											},
										},
										Computed:    true,
										Description: "Custom parameter",
									},
									"description": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Parameter description",
									},
									"hidden": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "UI-visibility",
									},
									"label_values": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"default_values": {
													Type: schema.TypeList,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
													Computed:    true,
													Description: "Default value",
												},
												"folder_id": {
													Type:        schema.TypeString,
													Computed:    true,
													Optional:    true,
													Description: "Folder ID",
												},
												"label_key": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Label key to list label values",
												},
												"multiselectable": {
													Type:        schema.TypeBool,
													Computed:    true,
													Description: "Specifies the multiselectable values of parameter",
												},
												"selectors": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Selectors to select metric label values",
												},
											},
										},
										Computed:    true,
										Description: "Label values parameter",
									},
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Parameter identifier",
									},
									"text": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"default_value": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Default value",
												},
											},
										},
										Computed:    true,
										Description: "Text parameter",
									},
									"title": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "UI-visible title of the parameter",
									},
								},
							},
							Computed:    true,
							Description: "Dashboard parameter",
						},
						"selectors": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Predefined selectors",
						},
					},
				},
				Computed:    true,
				Description: "Dashboard parametrization",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Dashboard title",
			},
			"widgets": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"chart": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"chart_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Chart ID",
									},
									"description": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Chart description in dashboard (not enabled in UI)",
									},
									"display_legend": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Enable legend under chart",
									},
									"freeze": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Fixed time interval for chart",
									},
									"name_hiding_settings": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"names": {
													Type: schema.TypeList,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
													Computed:    true,
													Description: "",
												},
												"positive": {
													Type:        schema.TypeBool,
													Computed:    true,
													Description: "True if we want to show concrete series names only, false if we want to hide concrete series names",
												},
											},
										},
										Computed:    true,
										Description: "Name hiding settings",
									},
									"queries": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"downsampling": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"disabled": {
																Type:        schema.TypeBool,
																Computed:    true,
																Description: "Disable downsampling",
															},
															"gap_filling": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Parameters for filling gaps in data",
															},
															"grid_aggregation": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Function that is used for downsampling",
															},
															"grid_interval": {
																Type:     schema.TypeInt,
																Computed: true,
																Description: "Time interval (grid) for downsampling in milliseconds. " +
																	"Points in the specified range are aggregated into one time point",
															},

															"max_points": {
																Type:        schema.TypeInt,
																Computed:    true,
																Description: "Maximum number of points to be returned",
															},
														},
													},
													Computed:    true,
													Description: "Downsampling settings",
												},

												"target": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"hidden": {
																Type:        schema.TypeBool,
																Computed:    true,
																Description: "Checks that target is visible or invisible",
															},

															"query": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Query",
															},

															"text_mode": {
																Type:        schema.TypeBool,
																Computed:    true,
																Description: "Text mode",
															},
														},
													},
													Computed:    true,
													Description: "Downsampling settings",
												},
											},
										},
										Computed:    true,
										Description: "Queries",
									},
									"series_overrides": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Series name",
												},

												"settings": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"color": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Series color or empty",
															},

															"grow_down": {
																Type:        schema.TypeBool,
																Computed:    true,
																Description: "Stack grow down",
															},

															"name": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Series name or empty",
															},

															"stack_name": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Stack name or empty",
															},

															"type": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Type",
															},

															"yaxis_position": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Yaxis position",
															},
														},
													},
													Computed:    true,
													Description: "Override settings",
												},

												"target_index": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Target index",
												},
											},
										},
										Computed:    true,
										Description: "",
									},
									"title": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Chart widget title",
									},
									"visualization_settings": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"aggregation": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Aggregation",
												},

												"color_scheme_settings": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"automatic": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{},
																},
																Computed:    true,
																Description: "Automatic color scheme",
															},

															"gradient": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"green_value": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Gradient green value",
																		},

																		"red_value": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Gradient red value",
																		},

																		"violet_value": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Gradient violet value",
																		},

																		"yellow_value": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Gradient yellow value",
																		},
																	},
																},
																Computed:    true,
																Description: "Gradient color scheme",
															},

															"standard": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{},
																},
																Computed:    true,
																Description: "Standard color scheme",
															},
														},
													},
													Computed:    true,
													Description: "Color scheme settings",
												},

												"heatmap_settings": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"green_value": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Heatmap green value",
															},

															"red_value": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Heatmap red value",
															},

															"violet_value": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Heatmap violet_value",
															},

															"yellow_value": {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Heatmap yellow value",
															},
														},
													},
													Computed:    true,
													Description: "Heatmap settings",
												},

												"interpolate": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Interpolate",
												},

												"normalize": {
													Type:        schema.TypeBool,
													Computed:    true,
													Description: "Normalize",
												},

												"show_labels": {
													Type:        schema.TypeBool,
													Computed:    true,
													Description: "Show chart labels",
												},

												"title": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Inside chart title",
												},

												"type": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Visualization type",
												},

												"yaxis_settings": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"left": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"max": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Max value in extended number format or empty",
																		},

																		"min": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Min value in extended number format or empty",
																		},

																		"precision": {
																			Type:        schema.TypeInt,
																			Computed:    true,
																			Description: "Tick value precision (null as default, 0-7 in other cases)",
																		},

																		"title": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Title or empty",
																		},

																		"type": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Type",
																		},

																		"unit_format": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Unit format",
																		},
																	},
																},
																Computed:    true,
																Description: "Left Y axis settings",
															},

															"right": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"max": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Max value in extended number format or empty",
																		},

																		"min": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Min value in extended number format or empty",
																		},

																		"precision": {
																			Type:        schema.TypeInt,
																			Computed:    true,
																			Description: "Tick value precision (null as default, 0-7 in other cases)",
																		},

																		"title": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Title or empty",
																		},

																		"type": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Type",
																		},

																		"unit_format": {
																			Type:        schema.TypeString,
																			Computed:    true,
																			Description: "Unit format",
																		},
																	},
																},
																Computed:    true,
																Description: "Right Y axis settings",
															},
														},
													},
													Computed:    true,
													Description: "Y axis settings",
												},
											},
										},
										Computed:    true,
										Description: "Visualization settings",
									},
								},
							},
							Computed:    true,
							Description: "Chart widget",
						},
						"position": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"h": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Height",
									},

									"w": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Width",
									},

									"x": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "X-axis top-left corner coordinate",
									},

									"y": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Y-axis top-left corner coordinate",
									},
								},
							},
							Computed:    true,
							Description: "Widget layout position",
						},
						"text": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"text": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Text",
									},
								},
							},
							Computed:    true,
							Description: "Text widget",
						},
						"title": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Title size",
									},

									"text": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Title text",
									},
								},
							},
							Computed:    true,
							Description: "Title widget",
						},
					},
				},
				Computed:    true,
				Description: "Widgets",
			},
		},
	}
}

func dataSourceYandexMonitoringDashboardRead(ctxParent context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx := wrapMonitoringGrpcContext(ctxParent)
	dashboardID := d.Get("dashboard_id").(string)
	_, dashboardNameOk := d.GetOk("name")

	if dashboardNameOk {
		dashboardIDInner, err := resolveObjectID(ctx, config, d, sdkresolvers.MonitoringDashboardResolver)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to resolve data source Dashboard by name: %v", err))
		}
		dashboardID = dashboardIDInner
	}

	req := &monitoring.GetDashboardRequest{
		DashboardId: dashboardID,
	}

	log.Printf("[DEBUG] Reading dashboard %+v", req)
	dashboard, err := config.sdk.Monitoring().Dashboard().Get(ctx, req)

	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			return diag.FromErr(fmt.Errorf("dashboard not found: %s", dashboardID))
		}
		return diag.FromErr(err)
	}
	err = monitoringDashboardToState(dashboard, d)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func monitoringDashboardToState(dashboard *monitoring.Dashboard, d *schema.ResourceData) error {
	parametrization, err := flattenMonitoringParametrization(dashboard.GetParametrization())
	if err != nil {
		return err
	}
	widget, err := flattenMonitoringWidgetSlice(dashboard.GetWidgets())
	if err != nil {
		return err
	}

	d.Set("dashboard_id", dashboard.Id)
	d.Set("name", dashboard.Name)
	d.Set("description", dashboard.GetDescription())
	d.Set("folder_id", dashboard.GetFolderId())
	d.Set("name", dashboard.GetName())
	d.Set("title", dashboard.GetTitle())

	if err := d.Set("parametrization", parametrization); err != nil {
		return err
	}
	if err := d.Set("widgets", widget); err != nil {
		return err
	}
	if err := d.Set("labels", dashboard.Labels); err != nil {
		return err
	}

	d.SetId(dashboard.Id)
	return nil
}

func wrapMonitoringGrpcContext(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"clientId": "terraform-provider-yandex"}))
}
