package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/monitoring/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

const yandexMonitoringDashboardDefaultTimeout = 2 * time.Minute

func resourceYandexMonitoringDashboard() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMonitoringDashboardCreate,
		ReadContext:   resourceMonitoringDashboardRead,
		UpdateContext: resourceMonitoringDashboardUpdate,
		DeleteContext: resourceMonitoringDashboardDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				d.Set("dashboard_id", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMonitoringDashboardDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexMonitoringDashboardDefaultTimeout),
			Update: schema.DefaultTimeout(yandexMonitoringDashboardDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexMonitoringDashboardDefaultTimeout),
		},
		SchemaVersion: 1,
		Description:   "Monitoring dashboard",
		Schema: map[string]*schema.Schema{
			"dashboard_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Dashboard ID",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Dashboard description",
			},
			"folder_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "Folder ID",
			},
			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:         schema.HashString,
				Computed:    true,
				Optional:    true,
				Description: "Dashboard labels",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
													Optional:    true,
													Description: "Default value",
												},
												"multiselectable": {
													Type:        schema.TypeBool,
													Optional:    true,
													Description: "Specifies the multiselectable values of parameter",
												},
												"values": {
													Type: schema.TypeList,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
													Optional:    true,
													Description: "Parameter values",
												},
											},
										},
										Optional:    true,
										Description: "Custom parameter",
									},
									"description": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Parameter description",
									},
									"hidden": {
										Type:        schema.TypeBool,
										Optional:    true,
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
													Optional:    true,
													Description: "Default value",
												},
												"folder_id": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Folder ID",
												},
												"label_key": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Required. Label key to list label values",
												},
												"multiselectable": {
													Type:        schema.TypeBool,
													Optional:    true,
													Description: "Specifies the multiselectable values of parameter",
												},
												"selectors": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Required. Selectors to select metric label values",
												},
											},
										},
										Optional:    true,
										Description: "Label values parameter",
									},
									"id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Parameter identifier",
									},
									"text": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"default_value": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Default value",
												},
											},
										},
										Optional:    true,
										Description: "Text parameter",
									},
									"title": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "UI-visible title of the parameter",
									},
								},
							},
							Optional:    true,
							Description: "Dashboard parameter",
						},
						"selectors": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Predefined selectors",
						},
					},
				},
				Optional:    true,
				Computed:    true,
				Description: "Dashboard parametrization",
			},
			"title": {
				Type:        schema.TypeString,
				Optional:    true,
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
										Optional:    true,
										Description: "Chart ID",
									},
									"description": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Chart description in dashboard (not enabled in UI)",
									},
									"display_legend": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Enable legend under chart",
									},
									"freeze": {
										Type:        schema.TypeString,
										Optional:    true,
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
													Optional:    true,
													Description: "",
												},
												"positive": {
													Type:        schema.TypeBool,
													Optional:    true,
													Description: "True if we want to show concrete series names only, false if we want to hide concrete series names",
												},
											},
										},
										Optional:    true,
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
																Optional:    true,
																Description: "Disable downsampling",
															},
															"gap_filling": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Parameters for filling gaps in data",
															},
															"grid_aggregation": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Function that is used for downsampling",
															},
															"grid_interval": {
																Type:     schema.TypeInt,
																Optional: true,
																Description: "Time interval (grid) for downsampling in milliseconds. " +
																	"Points in the specified range are aggregated into one time point",
															},

															"max_points": {
																Type:        schema.TypeInt,
																Optional:    true,
																Description: "Maximum number of points to be returned",
															},
														},
													},
													Optional:    true,
													Description: "Downsampling settings",
												},

												"target": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"hidden": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "Checks that target is visible or invisible",
															},

															"query": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Required. Query",
															},

															"text_mode": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "Text mode",
															},
														},
													},
													Optional:    true,
													Description: "Downsampling settings",
												},
											},
										},
										Optional:    true,
										Description: "Queries",
									},
									"series_overrides": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Series name",
												},

												"settings": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"color": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Series color or empty",
															},

															"grow_down": {
																Type:        schema.TypeBool,
																Optional:    true,
																Description: "Stack grow down",
															},

															"name": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Series name or empty",
															},

															"stack_name": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Stack name or empty",
															},

															"type": {
																Type:        schema.TypeString,
																Optional:    true,
																Computed:    true,
																Description: "Type",
															},

															"yaxis_position": {
																Type:        schema.TypeString,
																Optional:    true,
																Computed:    true,
																Description: "Yaxis position",
															},
														},
													},
													Optional:    true,
													Description: "Override settings",
												},

												"target_index": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Target index",
												},
											},
										},
										Optional:    true,
										Description: "",
									},

									"title": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Chart widget title",
									},

									"visualization_settings": {
										Type: schema.TypeList,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"aggregation": {
													Type:        schema.TypeString,
													Optional:    true,
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
																Optional:    true,
																Description: "Automatic color scheme",
															},

															"gradient": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"green_value": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Gradient green value",
																		},

																		"red_value": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Gradient red value",
																		},

																		"violet_value": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Gradient violet_value",
																		},

																		"yellow_value": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Gradient yellow value",
																		},
																	},
																},
																Optional:    true,
																Description: "Gradient color scheme",
															},

															"standard": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{},
																},
																Optional:    true,
																Description: "Standard color scheme",
															},
														},
													},
													Optional:    true,
													Description: "Color scheme settings",
												},

												"heatmap_settings": {
													Type: schema.TypeList,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"green_value": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Heatmap green value",
															},

															"red_value": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Heatmap red value",
															},

															"violet_value": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Heatmap violet_value",
															},

															"yellow_value": {
																Type:        schema.TypeString,
																Optional:    true,
																Description: "Heatmap yellow value",
															},
														},
													},
													Optional:    true,
													Description: "Heatmap settings",
												},

												"interpolate": {
													Type:        schema.TypeString,
													Optional:    true,
													Computed:    true,
													Description: "Interpolate",
												},

												"normalize": {
													Type:        schema.TypeBool,
													Optional:    true,
													Description: "Normalize",
												},

												"show_labels": {
													Type:        schema.TypeBool,
													Optional:    true,
													Description: "Show chart labels",
												},

												"title": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Inside chart title",
												},

												"type": {
													Type:        schema.TypeString,
													Optional:    true,
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
																			Optional:    true,
																			Description: "Max value in extended number format or empty",
																		},

																		"min": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Min value in extended number format or empty",
																		},

																		"precision": {
																			Type:        schema.TypeInt,
																			Optional:    true,
																			Description: "Tick value precision (null as default, 0-7 in other cases)",
																		},

																		"title": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Title or empty",
																		},

																		"type": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Computed:    true,
																			Description: "Type",
																		},

																		"unit_format": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Computed:    true,
																			Description: "Unit format",
																		},
																	},
																},
																Optional:    true,
																Description: "Left Y axis settings",
															},

															"right": {
																Type: schema.TypeList,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"max": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Max value in extended number format or empty",
																		},

																		"min": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Min value in extended number format or empty",
																		},

																		"precision": {
																			Type:        schema.TypeInt,
																			Optional:    true,
																			Description: "Tick value precision (null as default, 0-7 in other cases)",
																		},

																		"title": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Title or empty",
																		},

																		"type": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Type",
																		},

																		"unit_format": {
																			Type:        schema.TypeString,
																			Optional:    true,
																			Description: "Unit format",
																		},
																	},
																},
																Optional:    true,
																Description: "Right Y axis settings",
															},
														},
													},
													Optional:    true,
													Description: "Y axis settings",
												},
											},
										},
										Optional:    true,
										Description: "Visualization settings",
									},
								},
							},
							Optional:    true,
							Description: "Chart widget",
						},
						"position": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"h": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Required. Height",
									},

									"w": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Required. Weight",
									},

									"x": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Required. X-axis top-left corner coordinate",
									},

									"y": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Required. Y-axis top-left corner coordinate",
									},
								},
							},
							Optional:    true,
							Description: "Required. Widget layout position",
						},
						"text": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"text": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Text",
									},
								},
							},
							Optional:    true,
							Description: "Text widget",
						},
						"title": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
										Description: "Title size",
									},

									"text": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Title text",
									},
								},
							},
							Optional:    true,
							Description: "Title widget",
						},
					},
				},
				Optional:    true,
				Description: "Widgets",
			},
		},
	}
}

func resourceMonitoringDashboardCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error getting folder ID while creating dashboard: %s", err))
	}
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error expanding labels while creating dashboard: %s", err))
	}
	widgets, err := expandDashboardWidgetsSlice(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error expanding widgets while creating dashboard: %s", err))
	}
	parametrization, err := expandDashboardParametrization(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error expanding parametrization while creating dashboard: %s", err))
	}

	req := &monitoring.CreateDashboardRequest{
		Name: d.Get("name").(string),
		Container: &monitoring.CreateDashboardRequest_FolderId{
			FolderId: folderID,
		},
		Description:     d.Get("description").(string),
		Labels:          labels,
		Title:           d.Get("title").(string),
		Widgets:         widgets,
		Parametrization: parametrization,
	}

	log.Printf("[DEBUG] Creating Monitoring dashboard %+v", req)

	ctx = wrapMonitoringGrpcContext(ctx)
	op, err := config.sdk.WrapOperation(config.sdk.Monitoring().Dashboard().Create(ctx, req))
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error while creating dashboard %s: %s", req.Name, err))
	}
	ID := ""
	if op.Error() != nil {
		identifier, success, err := tryHandleConflictError(ctx, config.sdk.Monitoring().Dashboard(), req, op.ErrorStatus())
		if err != nil {
			return diag.FromErr(err)
		}
		if !success {
			return diag.FromErr(fmt.Errorf("Error while handle creating dashboard conflict %s: %s", req.Name, op.Error()))
		}
		ID = identifier
	} else {
		if err = op.Wait(ctx); err != nil {
			return diag.FromErr(fmt.Errorf("Error while waiting create dashboard %s: %s", req.Name, err))
		}
		res, err := op.Response()
		if err != nil {
			return diag.FromErr(fmt.Errorf("Error while unmarshal response of created dashboard %s: %s", req.Name, err))
		}
		dashboard, _ := res.(*monitoring.Dashboard)
		ID = dashboard.Id
	}
	d.Set("dashboard_id", ID)
	d.SetId(ID)
	return resourceMonitoringDashboardRead(ctx, d, meta)
}

func tryHandleConflictError(ctx context.Context, client monitoring.DashboardServiceClient, getReq *monitoring.CreateDashboardRequest, error *status.Status) (string, bool, error) {
	if error.Code() == codes.AlreadyExists {
		req := &monitoring.ListDashboardsRequest{
			Container: &monitoring.ListDashboardsRequest_FolderId{
				FolderId: getReq.GetFolderId(),
			},
			Filter:   fmt.Sprintf("name=\"%s\"", getReq.Name),
			PageSize: 2,
		}
		ctx = wrapMonitoringGrpcContext(ctx)
		response, err := client.List(ctx, req)
		if err != nil {
			return "", false, err
		}
		if len(response.Dashboards) != 1 {
			return "", false, fmt.Errorf("Failed to find dashboard: %+v, %+v", req, response)
		}
		return response.Dashboards[0].Id, true, nil
	}
	return "", false, nil
}

func resourceMonitoringDashboardRead(ctxParent context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx := wrapMonitoringGrpcContext(ctxParent)
	dashboardID := d.Get("dashboard_id").(string)
	req := &monitoring.GetDashboardRequest{
		DashboardId: dashboardID,
	}

	log.Printf("[DEBUG] Reading Monitoring dashboard %+v", req)
	dashboard, err := config.sdk.Monitoring().Dashboard().Get(ctx, req)

	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			log.Printf("[DEBUG] Monitoring dashboard (%s) was not found", d.Get("name").(string))
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	err = monitoringDashboardToState(dashboard, d)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceMonitoringDashboardUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error expanding labels while updating dashboard: %s", err))
	}
	widgets, err := expandDashboardWidgetsSlice(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error expanding widgets while updating dashboard: %s", err))
	}
	parametrization, err := expandDashboardParametrization(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error expanding parametrization while updating dashboard: %s", err))
	}

	req := &monitoring.UpdateDashboardRequest{
		DashboardId:     d.Id(),
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		Labels:          labels,
		Title:           d.Get("title").(string),
		Widgets:         widgets,
		Parametrization: parametrization,
		Etag:            "-1",
	}

	log.Printf("[DEBUG] Updating Monitoring dashboard %+v", req)

	ctx = wrapMonitoringGrpcContext(ctx)
	op, err := config.sdk.WrapOperation(config.sdk.Monitoring().Dashboard().Update(ctx, req))
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error while updating dashboard %s: %s", req.Name, err))
	}
	if err = op.Wait(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("Error while waiting updating dashboard %s: %s", req.Name, err))
	}
	if op.Error() != nil {
		return diag.FromErr(fmt.Errorf("Error while updating updating %s: %s", req.Name, op.Error()))
	}
	return resourceMonitoringDashboardRead(ctx, d, meta)
}

func resourceMonitoringDashboardDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	name := d.Get("name").(string)
	req := &monitoring.DeleteDashboardRequest{
		DashboardId: d.Id(),
		Etag:        "-1",
	}
	log.Printf("[DEBUG] Deleting Monitoring dashboard %+v", req)

	ctx = wrapMonitoringGrpcContext(ctx)
	op, err := config.sdk.WrapOperation(config.sdk.Monitoring().Dashboard().Delete(ctx, req))
	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			log.Printf("[WARN] Removing %s because resource doesn't exist anymore", name)
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err = op.Wait(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("Error while waiting deleting dashboard %s: %s", d.Id(), err))
	}
	if op.Error() != nil {
		return diag.FromErr(fmt.Errorf("Error while deleting updating %s: %s", name, op.Error()))
	}
	return nil
}
