package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/audittrails/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func resourceAuditTrailsTrailResourceSchema() *schema.Resource {
	return &schema.Resource{
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
	}
}

func resourceYandexAuditTrailsTrail() *schema.Resource {
	return &schema.Resource{
		ReadContext:   readTrailResource,
		CreateContext: createTrailResource,
		UpdateContext: updateTrailResource,
		DeleteContext: deleteTrailResource,

		SchemaVersion: 1,

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
			"filtering_policy": {
				ExactlyOneOf: []string{
					"filtering_policy",
					"filter",
				},
				Optional: true,
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"management_events_filter": {
							AtLeastOneOf: []string{
								"filtering_policy.0.management_events_filter",
								"filtering_policy.0.data_events_filter",
							},
							Optional: true,
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_scope": {
										Required: true,
										Type:     schema.TypeList,
										MinItems: 1,
										Elem:     resourceAuditTrailsTrailResourceSchema(),
									},
								},
							},
						},
						"data_events_filter": {
							AtLeastOneOf: []string{
								"filtering_policy.0.management_events_filter",
								"filtering_policy.0.data_events_filter",
							},
							Optional: true,
							Type:     schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"service": {
										Required: true,
										Type:     schema.TypeString,
									},
									"included_events": {
										Optional: true,
										Type:     schema.TypeList,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"excluded_events": {
										Optional: true,
										Type:     schema.TypeList,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"resource_scope": {
										Required: true,
										Type:     schema.TypeList,
										MinItems: 1,
										Elem:     resourceAuditTrailsTrailResourceSchema(),
									},
								},
							},
						},
					},
				},
			},
			"filter": {
				ExactlyOneOf: []string{
					"filtering_policy",
					"filter",
				},
				Optional:   true,
				Type:       schema.TypeList,
				MaxItems:   1,
				Deprecated: "Configure filtering_policy instead. This attribute will be removed",
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
										Elem:     resourceAuditTrailsTrailResourceSchema(),
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
													Elem:     resourceAuditTrailsTrailResourceSchema(),
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
													Elem:     resourceAuditTrailsTrailResourceSchema(),
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
																Elem:     resourceAuditTrailsTrailResourceSchema(),
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
	ctx = tflog.SetField(ctx, "trail_id", id)

	tflog.Debug(ctx, "Deleting trail")

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

	tflog.Debug(ctx, "Finished deleting trail")
	return nil
}

func updateTrailResource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx = tflog.SetField(ctx, "trail_id", data.Id())

	tflog.Debug(ctx, "Updating trail")

	labels, err := expandLabels(data.Get("labels"))
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
		FilteringPolicy:  packResourceDataIntoFilteringPolicy(data),
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"name", "description", "labels", "service_account_id", "destination", "filtering_policy", "filter"},
		},
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

	tflog.Debug(ctx, "Finished updating trail")

	return readTrailResource(ctx, data, meta)
}

func createTrailResource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx = tflog.SetField(ctx, "trail_name", data.Get("name"))

	tflog.Debug(ctx, "Creating trail")

	labels, err := expandLabels(data.Get("labels"))
	if err != nil {
		return diag.FromErr(err)
	}

	folderID, err := getFolderID(data, config)
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
		FilteringPolicy:  packResourceDataIntoFilteringPolicy(data),
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

	tflog.Debug(ctx, "Finished creating trail")

	return readTrailResource(ctx, data, meta)
}

func packResourceDataIntoFilteringPolicy(data *schema.ResourceData) *audittrails.Trail_FilteringPolicy {
	res := &audittrails.Trail_FilteringPolicy{}

	if _, newFilterUsed := data.GetOk("filtering_policy"); newFilterUsed {

		_, filteringPolicyExists := data.GetOk("filtering_policy.0.management_events_filter")
		if filteringPolicyExists {
			managementFilter := packResourceDataIntoManagementFilter(data, "filtering_policy.0.management_events_filter.0.")
			res.SetManagementEventsFilter(managementFilter)
		}

		_, filteringPolicyExists = data.GetOk("filtering_policy.0.data_events_filter")
		if filteringPolicyExists {
			dataEventsFilters := packResourceDataIntoDataEventsFilters(data, "filtering_policy.0.data_events_filter.")
			res.SetDataEventsFilters(dataEventsFilters)
		}
	}

	if _, oldFilterUsed := data.GetOk("filter"); oldFilterUsed {
		return packResourceDataIntoFilter(data)
	}

	return res
}

func packResourceDataIntoDataEventsFilters(data *schema.ResourceData, namespace string) []*audittrails.Trail_DataEventsFiltering {
	res := []*audittrails.Trail_DataEventsFiltering{}

	numberOfFilters := data.Get(namespace + "#").(int)
	for i := 0; i < numberOfFilters; i++ {
		filterNamespace := fmt.Sprintf("%s%d.", namespace, i)

		filter := &audittrails.Trail_DataEventsFiltering{
			ResourceScopes: packResourceDataIntoResourceScopes(data, filterNamespace+"resource_scope."),
			Service:        data.Get(filterNamespace + "service").(string),
		}

		_, exists := data.GetOk(filterNamespace + "included_events")
		if exists {
			includedEvents := packResourceDataIntoEventTypes(data, filterNamespace+"included_events.")
			filter.SetIncludedEvents(includedEvents)
		}

		_, exists = data.GetOk(filterNamespace + "excluded_events")
		if exists {
			excludedEvents := packResourceDataIntoEventTypes(data, filterNamespace+"excluded_events.")
			filter.SetExcludedEvents(excludedEvents)
		}

		res = append(res, filter)
	}

	return res
}

func packResourceDataIntoEventTypes(data *schema.ResourceData, namespace string) *audittrails.Trail_EventTypes {
	res := []string{}

	numberOfTypes := data.Get(namespace + "#").(int)
	for i := 0; i < numberOfTypes; i++ {
		eventTypePath := fmt.Sprintf("%s%d", namespace, i)
		eventType := data.Get(eventTypePath).(string)
		res = append(res, eventType)
	}

	return &audittrails.Trail_EventTypes{
		EventTypes: res,
	}
}

func packResourceDataIntoFilter(data *schema.ResourceData) *audittrails.Trail_FilteringPolicy {
	res := &audittrails.Trail_FilteringPolicy{}

	_, exists := data.GetOk("filter.0.path_filter")
	if exists {
		pathFilter := packResourceDataIntoPathFilter(data, "filter.0.path_filter.0.")
		res.SetManagementEventsFilter(&audittrails.Trail_ManagementEventsFiltering{
			ResourceScopes: pathFilterToResourceScopes(pathFilter),
		})
	}

	eventFiltersField, ok := data.GetOk("filter.0.event_filters.#")

	var eventFiltersCount int
	if ok {
		eventFiltersCount = eventFiltersField.(int)
	} else {
		eventFiltersCount = 0
	}

	eventFilters := make([]*audittrails.Trail_DataEventsFiltering, eventFiltersCount)

	for i := 0; i < eventFiltersCount; i++ {
		prefix := fmt.Sprintf("filter.0.event_filters.%d.", i)
		eventFilterElement := packResourceDataIntoEventFilterElement(data, prefix)
		eventFilters[i] = eventFilterElement
	}

	res.SetDataEventsFilters(eventFilters)

	return res
}

func packResourceDataIntoEventFilterElement(data *schema.ResourceData, namespace string) *audittrails.Trail_DataEventsFiltering {
	pathFilter := packResourceDataIntoPathFilter(data, namespace+"path_filter.0.")

	return &audittrails.Trail_DataEventsFiltering{
		Service:        data.Get(namespace + "service").(string),
		ResourceScopes: pathFilterToResourceScopes(pathFilter),
	}
}

func pathFilterToResourceScopes(pathFilter *audittrails.Trail_PathFilter) []*audittrails.Trail_Resource {
	if anyFilter := pathFilter.Root.GetAnyFilter(); anyFilter != nil {
		return []*audittrails.Trail_Resource{anyFilter.Resource}
	}

	if someFilter := pathFilter.Root.GetSomeFilter(); someFilter != nil {
		result := []*audittrails.Trail_Resource{}
		anyFilters := someFilter.GetFilters()
		for _, anyFilter := range anyFilters {
			result = append(result, anyFilter.GetAnyFilter().GetResource())
		}
		return result
	}

	panic("Shouldn't happen due to internal terraform resource validations")
}

func packResourceDataIntoPathFilter(data *schema.ResourceData, namespace string) *audittrails.Trail_PathFilter {
	_, anyDefined := data.GetOk(namespace + "any_filter")
	_, someDefined := data.GetOk(namespace + "some_filter")

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
	return &audittrails.Trail_PathFilter{Root: resRoot}
}

func packResourceDataIntoManagementFilter(data *schema.ResourceData, namespace string) *audittrails.Trail_ManagementEventsFiltering {
	return &audittrails.Trail_ManagementEventsFiltering{
		ResourceScopes: packResourceDataIntoResourceScopes(data, namespace+"resource_scope."),
	}
}

func packResourceDataIntoResourceScopes(data *schema.ResourceData, namespace string) []*audittrails.Trail_Resource {
	res := []*audittrails.Trail_Resource{}

	numberOfScopes := data.Get(namespace + "#").(int)
	for i := 0; i < numberOfScopes; i++ {
		resourceNamespace := fmt.Sprintf("%s%d.", namespace, i)
		res = append(res, packResourceDataIntoResource(data, resourceNamespace))
	}
	return res
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
	ctx = tflog.SetField(ctx, "trail_id", id)

	tflog.Debug(ctx, "Reading trail")

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

	tflog.Debug(ctx, "Finished reading trail")

	return unpackingErrors
}
