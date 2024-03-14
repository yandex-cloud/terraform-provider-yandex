package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/audittrails/v1"
	"log"
	"time"
)

func resourceYandexAuditTrailsTrail() *schema.Resource {
	return &schema.Resource{
		ReadContext:   readTrailResource,
		CreateContext: createTrailResource,
		UpdateContext: updateTrailResource,
		DeleteContext: deleteTrailResource,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"trail_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"service_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_destination": {
				ExactlyOneOf: []string{
					"storage_destination",
					"logging_destination",
					"data_stream_destination",
				},
				Optional: true,
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"object_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"logging_destination": {
				ExactlyOneOf: []string{
					"storage_destination",
					"logging_destination",
					"data_stream_destination",
				},
				Optional: true,
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"data_stream_destination": {
				ExactlyOneOf: []string{
					"storage_destination",
					"logging_destination",
					"data_stream_destination",
				},
				Optional: true,
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"stream_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"filter": {
				Required: true,
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path_filter": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"any_filter": {
										ExactlyOneOf: []string{
											"filter.0.path_filter.0.any_filter",
											"filter.0.path_filter.0.some_filter",
										},
										Optional: true,
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"resource_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"resource_type": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"some_filter": {
										ExactlyOneOf: []string{
											"filter.0.path_filter.0.any_filter",
											"filter.0.path_filter.0.some_filter",
										},
										Optional: true,
										Type:     schema.TypeList,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"resource_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"resource_type": {
													Type:     schema.TypeString,
													Required: true,
												},
												"any_filters": {
													Type:     schema.TypeList,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"resource_id": {
																Type:     schema.TypeString,
																Required: true,
															},
															"resource_type": {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"event_filters": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"service": {
										Type:     schema.TypeString,
										Required: true,
									},
									"categories": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"plane": {
													Type:     schema.TypeString,
													Required: true,
												},
												"type": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"path_filter": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"any_filter": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"resource_id": {
																Type:     schema.TypeString,
																Required: true,
															},
															"resource_type": {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
												"some_filter": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"resource_id": {
																Type:     schema.TypeString,
																Required: true,
															},
															"resource_type": {
																Type:     schema.TypeString,
																Required: true,
															},
															"any_filters": {
																Type:     schema.TypeList,
																Required: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"resource_id": {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																		"resource_type": {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func deleteTrailResource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	id := data.Id()

	log.Printf("[DEBUG] Deleting Trail %q", id)

	req := &audittrails.DeleteTrailRequest{
		TrailId: id,
	}

	op, err := config.sdk.WrapOperation(config.sdk.AuditTrails().Trail().Delete(ctx, req))
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, data, fmt.Sprintf("Trail %q", id)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished deleting Trail %q", id)
	return nil
}

func updateTrailResource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	log.Printf("[DEBUG] Updating Trail %q", data.Id())

	labels, err := expandLabels(data.Get("labels"))
	if err != nil {
		return diag.FromErr(err)
	}

	filter, err := packResourceDataIntoFilter(data)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &audittrails.UpdateTrailRequest{
		TrailId:          data.Id(),
		Name:             data.Get("name").(string),
		Description:      data.Get("description").(string),
		Labels:           labels,
		ServiceAccountId: data.Get("service_account_id").(string),
		Destination:      packResourceDataIntoDestination(data),
		Filter:           filter,
	}

	err = retry.RetryContext(ctx, data.Timeout(schema.TimeoutRead), func() *retry.RetryError {
		operation, err := config.sdk.WrapOperation(config.sdk.AuditTrails().Trail().Update(ctx, req))
		if err != nil {
			return retryErrorForCode(err)
		}

		metadata, err := operation.Metadata()
		if err != nil {
			return retry.NonRetryableError(err)
		}

		trailMetadata := metadata.(*audittrails.UpdateTrailMetadata)
		data.SetId(trailMetadata.TrailId)

		err = operation.Wait(ctx)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		_, err = operation.Response()
		if err != nil {
			return nil
		}

		return nil // do not return any error in case if network call completed correctly
	})

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished updating Trail %q", data.Id())

	return readTrailResource(ctx, data, meta)
}

func createTrailResource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	log.Printf("[DEBUG] Creating Trail %q", data.Get("name").(string))

	labels, err := expandLabels(data.Get("labels"))
	if err != nil {
		return diag.FromErr(err)
	}

	folderID, err := getFolderID(data, config)
	if err != nil {
		return diag.FromErr(err)
	}

	filter, err := packResourceDataIntoFilter(data)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &audittrails.CreateTrailRequest{
		FolderId:         folderID,
		Name:             data.Get("name").(string),
		Description:      data.Get("description").(string),
		Labels:           labels,
		ServiceAccountId: data.Get("service_account_id").(string),
		Destination:      packResourceDataIntoDestination(data),
		Filter:           filter,
	}

	err = retry.RetryContext(ctx, data.Timeout(schema.TimeoutRead), func() *retry.RetryError {
		operation, err := config.sdk.WrapOperation(config.sdk.AuditTrails().Trail().Create(ctx, req))
		if err != nil {
			return retryErrorForCode(err)
		}

		metadata, err := operation.Metadata()
		if err != nil {
			return retry.NonRetryableError(err)
		}

		trailMetadata := metadata.(*audittrails.CreateTrailMetadata)
		data.SetId(trailMetadata.TrailId)

		err = operation.Wait(ctx)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		_, err = operation.Response()
		if err != nil {
			return nil
		}

		return nil // do not return any error in case if network call completed correctly
	})

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished creating Trail %q", data.Get("name").(string))

	return readTrailResource(ctx, data, meta)
}

func packResourceDataIntoFilter(data *schema.ResourceData) (*audittrails.Trail_Filter, error) {
	res := &audittrails.Trail_Filter{}

	_, exists := data.GetOk("filter.0.path_filter")
	if exists {
		pathFilter, err := packResourceDataIntoPathFilter(data, "filter.0.path_filter.0.")
		if err != nil {
			return nil, err
		}
		res.SetPathFilter(pathFilter)
	}

	eventFiltersField, ok := data.GetOk("filter.0.event_filters.#")

	var eventFiltersCount int
	if ok {
		eventFiltersCount = eventFiltersField.(int)
	} else {
		eventFiltersCount = 0
	}

	eventFilters := make([]*audittrails.Trail_EventFilterElement, eventFiltersCount)

	for i := 0; i < eventFiltersCount; i++ {
		prefix := fmt.Sprintf("filter.0.event_filters.%d.", i)
		eventFilterElement, err := packResourceDataIntoEventFilterElement(data, prefix)
		if err != nil {
			return nil, err
		}
		eventFilters[i] = eventFilterElement
	}

	res.SetEventFilter(&audittrails.Trail_EventFilter{
		Filters: eventFilters,
	})

	return res, nil
}

func packResourceDataIntoEventFilterElement(data *schema.ResourceData, namespace string) (*audittrails.Trail_EventFilterElement, error) {
	categoriesCount := data.Get(namespace + "categories.#").(int)
	categories := make([]*audittrails.Trail_EventFilterElementCategory, categoriesCount)
	for i := 0; i < categoriesCount; i++ {
		prefix := fmt.Sprintf("%scategories.%d.", namespace, i)
		categories[i] = &audittrails.Trail_EventFilterElementCategory{
			Type:  audittrails.Trail_EventAccessTypeFilter(audittrails.Trail_EventAccessTypeFilter_value[data.Get(prefix+"type").(string)]),
			Plane: audittrails.Trail_EventCategoryFilter(audittrails.Trail_EventCategoryFilter_value[data.Get(prefix+"plane").(string)]),
		}
	}

	pathFilter, err := packResourceDataIntoPathFilter(data, namespace+"path_filter.0.")
	if err != nil {
		return nil, err
	}

	return &audittrails.Trail_EventFilterElement{
		Service:    data.Get(namespace + "service").(string),
		Categories: categories,
		PathFilter: pathFilter,
	}, nil
}

func packResourceDataIntoPathFilter(data *schema.ResourceData, namespace string) (*audittrails.Trail_PathFilter, error) {
	_, anyDefined := data.GetOk(namespace + "any_filter")
	_, someDefined := data.GetOk(namespace + "some_filter")

	if anyDefined == someDefined {
		return nil, fmt.Errorf("exactly one of fields any_filter or some_filter should be specified at %s", namespace)
	}

	resRoot := &audittrails.Trail_PathFilterElement{}
	if anyDefined {
		resRoot.SetAnyFilter(&audittrails.Trail_PathFilterElementAny{
			Resource: packResourceDataIntoResource(data, namespace+"any_filter.0."),
		})
	}
	if someDefined {
		childNumber := data.Get(namespace + "some_filter.0.any_filters.#").(int)
		childFilters := make([]*audittrails.Trail_PathFilterElement, childNumber)
		for i := 0; i < childNumber; i++ {
			prefix := fmt.Sprintf("%ssome_filter.0.any_filters.%d.", namespace, i)
			childFilters[i] = &audittrails.Trail_PathFilterElement{}
			childFilters[i].SetAnyFilter(&audittrails.Trail_PathFilterElementAny{
				Resource: packResourceDataIntoResource(data, prefix),
			})
		}

		resRoot.SetSomeFilter(&audittrails.Trail_PathFilterElementSome{
			Resource: packResourceDataIntoResource(data, namespace+"some_filter.0."),
			Filters:  childFilters,
		})
	}
	return &audittrails.Trail_PathFilter{Root: resRoot}, nil
}

func packResourceDataIntoResource(data *schema.ResourceData, namespace string) *audittrails.Trail_Resource {
	return &audittrails.Trail_Resource{
		Type: data.Get(namespace + "resource_type").(string),
		Id:   data.Get(namespace + "resource_id").(string),
	}
}

func packResourceDataIntoDestination(data *schema.ResourceData) *audittrails.Trail_Destination {
	if _, exists := data.GetOk("storage_destination"); exists {
		return &audittrails.Trail_Destination{
			Destination: &audittrails.Trail_Destination_ObjectStorage{
				ObjectStorage: &audittrails.Trail_ObjectStorage{
					BucketId:     data.Get("storage_destination.0.bucket_name").(string),
					ObjectPrefix: data.Get("storage_destination.0.object_prefix").(string),
				},
			},
		}
	}

	if _, exists := data.GetOk("logging_destination"); exists {
		return &audittrails.Trail_Destination{
			Destination: &audittrails.Trail_Destination_CloudLogging{
				CloudLogging: &audittrails.Trail_CloudLogging{
					Destination: &audittrails.Trail_CloudLogging_LogGroupId{
						LogGroupId: data.Get("logging_destination.0.log_group_id").(string),
					},
				},
			},
		}
	}

	if _, exists := data.GetOk("data_stream_destination"); exists {
		return &audittrails.Trail_Destination{
			Destination: &audittrails.Trail_Destination_DataStream{
				DataStream: &audittrails.Trail_DataStream{
					DatabaseId: data.Get("data_stream_destination.0.database_id").(string),
					StreamName: data.Get("data_stream_destination.0.stream_name").(string),
				},
			},
		}
	}

	panic("This shouldn't happen due to ExactlyOneOf validation")
}

func readTrailResource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	id := data.Id()

	log.Printf("[DEBUG] Reading Trail %q", id)

	var unpackingErrors diag.Diagnostics

	err := retry.RetryContext(ctx, data.Timeout(schema.TimeoutRead), func() *retry.RetryError {
		trail, err := config.sdk.AuditTrails().Trail().Get(ctx, &audittrails.GetTrailRequest{
			TrailId: id,
		})
		if err != nil {
			return retryErrorForCode(err)
		}

		unpackingErrors = unpackProtoTrailIntoResourceData(trail, data)
		return nil // do not return any error in case if network call completed correctly
	})

	if err != nil {
		return diag.FromErr(handleNotFoundError(err, data, fmt.Sprintf("Trail %q", id)))
	}

	log.Printf("[DEBUG] Finished reading Trail %q", id)
	return unpackingErrors
}
