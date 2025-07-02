package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/audittrails/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func dataSourceAuditTrailsTrailResourceListSchema() *schema.Schema {
	return &schema.Schema{
		Computed:    true,
		Type:        schema.TypeList,
		Description: "Structure describing that events will be gathered from the specified resource.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"resource_id": {
					Type:        schema.TypeString,
					Description: "Resource ID.",
					Computed:    true,
				},
				"resource_type": {
					Type:        schema.TypeString,
					Description: "Resource type.",
					Computed:    true,
				},
			},
		},
	}
}

func dataSourceAuditTrailsTrailResourcePathSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Deprecated.",
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"any_filter": {
					Computed:    true,
					Type:        schema.TypeList,
					Description: "Deprecated.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"resource_id": {
								Type:        schema.TypeString,
								Description: "Deprecated.",
								Computed:    true,
							},
							"resource_type": {
								Type:        schema.TypeString,
								Description: "Deprecated.",
								Computed:    true,
							},
						},
					},
				},
				"some_filter": {
					Computed:    true,
					Type:        schema.TypeList,
					Description: "Deprecated.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"resource_id": {
								Type:        schema.TypeString,
								Description: "Deprecated.",
								Computed:    true,
							},
							"resource_type": {
								Type:        schema.TypeString,
								Description: "Deprecated.",
								Computed:    true,
							},
							"any_filters": {
								Type:        schema.TypeList,
								Description: "Deprecated.",
								Computed:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"resource_id": {
											Type:        schema.TypeString,
											Description: "Deprecated.",
											Computed:    true,
										},
										"resource_type": {
											Type:        schema.TypeString,
											Description: "Deprecated.",
											Computed:    true,
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

func dataSourceYandexAuditTrailsTrail() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a trail. For information about the trail concept, see [official documentation](https://yandex.cloud/docs/audit-trails/concepts/trail).",
		ReadContext: readTrailDataSource,

		SchemaVersion: 1,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"trail_id": {
				Type:        schema.TypeString,
				Description: "Trail ID.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Computed:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"service_account_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["service_account_id"],
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexAuditTrailsTrail().Schema["status"].Description,
				Computed:    true,
			},
			"storage_destination": {
				Computed:    true,
				Type:        schema.TypeList,
				Description: "Structure describing destination bucket of the trail. Mutually exclusive with `logging_destination` and `data_stream_destination`.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_name": {
							Type:        schema.TypeString,
							Description: "Name of the [destination bucket](https://yandex.cloud/docs/storage/concepts/bucket).",
							Computed:    true,
						},
						"object_prefix": {
							Type:        schema.TypeString,
							Description: "Additional prefix of the uploaded objects. If not specified, objects will be uploaded with prefix equal to `trail_id`.",
							Computed:    true,
						},
					},
				},
			},
			"logging_destination": {
				Computed:    true,
				Type:        schema.TypeList,
				Description: "Structure describing destination log group of the trail. Mutually exclusive with `storage_destination` and `data_stream_destination`.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_id": {
							Type:        schema.TypeString,
							Description: "ID of the destination [Cloud Logging Group](https://yandex.cloud/docs/logging/concepts/log-group).",
							Computed:    true,
						},
					},
				},
			},
			"data_stream_destination": {
				Computed:    true,
				Type:        schema.TypeList,
				Description: "Structure describing destination data stream of the trail. Mutually exclusive with `logging_destination` and `storage_destination`.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_id": {
							Type:        schema.TypeString,
							Description: "ID of the [YDB](https://yandex.cloud/docs/ydb/concepts/resources) hosting the destination data stream.",
							Computed:    true,
						},
						"stream_name": {
							Type:        schema.TypeString,
							Description: "Name of the [YDS stream](https://yandex.cloud/docs/data-streams/concepts/glossary#stream-concepts) belonging to the specified YDB.",
							Computed:    true,
						},
					},
				},
			},
			"filtering_policy": {
				Computed:    true,
				Type:        schema.TypeList,
				Description: "Structure describing event filtering process for the trail. Mutually exclusive with `filter`. At least one of the `management_events_filter` or `data_events_filter` fields will be filled.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"management_events_filter": {
							Type:        schema.TypeList,
							Description: "Structure describing filtering process for management events.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_scope": dataSourceAuditTrailsTrailResourceListSchema(),
								},
							},
						},
						"data_events_filter": {
							Type:        schema.TypeList,
							Description: "Structure describing filtering process for the service-specific data events.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"service": {
										Type:        schema.TypeString,
										Description: "ID of the service which events will be gathered.",
										Computed:    true,
									},
									"included_events": {
										Type:        schema.TypeList,
										Description: "A list of events that will be gathered by the trail from this service. New events won't be gathered by default when this option is specified. Mutually exclusive with `excluded_events`.",
										Computed:    true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"excluded_events": {
										Type:        schema.TypeList,
										Description: "A list of events that won't be gathered by the trail from this service. New events will be automatically gathered when this option is specified. Mutually exclusive with `included_events`.",
										Computed:    true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"resource_scope": dataSourceAuditTrailsTrailResourceListSchema(),
									"dns_filter": {
										Type:        schema.TypeList,
										Description: "Specific filter for DNS service.",
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"include_nonrecursive_queries": {
													Type:        schema.TypeBool,
													Description: "All types of queries will be delivered.",
													Computed:    true,
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
			"filter": {
				Computed:    true,
				Type:        schema.TypeSet,
				Description: "Structure is deprecated. Use `filtering_policy` instead.",
				Deprecated:  "Use filtering_policy instead. This attribute will be removed",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path_filter": dataSourceAuditTrailsTrailResourcePathSchema(),
						"event_filters": {
							Type:        schema.TypeList,
							Description: "Deprecated.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"service": {
										Type:        schema.TypeString,
										Description: "Deprecated.",
										Computed:    true,
									},
									"categories": {
										Type:        schema.TypeList,
										Description: "Deprecated.",
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"plane": {
													Type:        schema.TypeString,
													Description: "Deprecated.",
													Computed:    true,
												},
												"type": {
													Type:        schema.TypeString,
													Description: "Deprecated.",
													Computed:    true,
												},
											},
										},
									},
									"path_filter": dataSourceAuditTrailsTrailResourcePathSchema(),
								},
							},
						},
					},
				},
			},
		},
	}
}

func retryErrorForCode(err error) *retry.RetryError {
	grpcCode := status.Code(err)

	retryableCodes := []codes.Code{
		codes.Unavailable,
	}

	if slices.Contains(retryableCodes, grpcCode) {
		return retry.RetryableError(err)
	} else {
		return retry.NonRetryableError(err)
	}
}

func readTrailDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	id := data.Get("trail_id").(string)

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
		// do not use handleNotFoundError (that deletes resource upon not found error) in the data source
		// quote from terraform docs:
		// Data resources that are designed to return state for a singular
		// infrastructure component should conventionally return an error if that
		// infrastructure does not exist and omit any calls to the
		// SetId method.
		return diag.FromErr(err)
	}

	return unpackingErrors
}

func unpackProtoTrailIntoResourceData(trail *audittrails.Trail, data *schema.ResourceData) diag.Diagnostics {
	result := diag.Diagnostics{}

	data.SetId(trail.GetId())

	result = setAndAppendError(data, "name", trail.GetName(), result)
	result = setAndAppendError(data, "folder_id", trail.GetFolderId(), result)
	result = setAndAppendError(data, "description", trail.GetDescription(), result)
	result = setAndAppendError(data, "labels", trail.GetLabels(), result)
	result = setAndAppendError(data, "service_account_id", trail.GetServiceAccountId(), result)
	result = setAndAppendError(data, "status", trail.GetStatus().String(), result)

	if dataStreamDestination := trail.GetDestination().GetDataStream(); dataStreamDestination != nil {
		dataStream := map[string]string{
			"database_id": dataStreamDestination.GetDatabaseId(),
			"stream_name": dataStreamDestination.GetStreamName(),
		}
		result = setAndAppendError(data, "data_stream_destination", []interface{}{dataStream}, result)
	} else {
		result = setAndAppendError(data, "data_stream_destination", nil, result)
	}

	if loggingDestination := trail.GetDestination().GetCloudLogging(); loggingDestination != nil {
		logGroup := map[string]string{
			"log_group_id": loggingDestination.GetLogGroupId(),
		}
		result = setAndAppendError(data, "logging_destination", []interface{}{logGroup}, result)
	} else {
		result = setAndAppendError(data, "logging_destination", nil, result)
	}

	if storageDestination := trail.GetDestination().GetObjectStorage(); storageDestination != nil {
		bucket := map[string]string{
			"bucket_name":   storageDestination.GetBucketId(),
			"object_prefix": storageDestination.GetObjectPrefix(),
		}
		result = setAndAppendError(data, "storage_destination", []interface{}{bucket}, result)
	} else {
		result = setAndAppendError(data, "storage_destination", nil, result)
	}

	flatTrailFilteringPolicy := map[string]interface{}{}

	filteringPolicy := trail.GetFilteringPolicy()
	flatDataEventFilters := []map[string]interface{}{}
	for _, dataEventFilter := range filteringPolicy.GetDataEventsFilters() {
		flatDataEventFilter := map[string]interface{}{}

		flatDataEventFilter["service"] = dataEventFilter.GetService()
		flatDataEventFilter["resource_scope"] = unpackProtoResourceScopesIntoResourceData(dataEventFilter.GetResourceScopes())

		if excludedEvents := dataEventFilter.GetExcludedEvents(); excludedEvents != nil {
			flatDataEventFilter["excluded_events"] = unpackEventTypesIntoResourceData(excludedEvents)
		}
		if includedEvents := dataEventFilter.GetIncludedEvents(); includedEvents != nil {
			flatDataEventFilter["included_events"] = unpackEventTypesIntoResourceData(includedEvents)
		}
		if dnsFilter := dataEventFilter.GetDnsFilter(); dnsFilter != nil {
			dns_filter := []interface{}{
				map[string]bool{
					"include_nonrecursive_queries": dnsFilter.GetIncludeNonrecursiveQueries(),
				},
			}
			flatDataEventFilter["dns_filter"] = dns_filter
		}
		flatDataEventFilters = append(flatDataEventFilters, flatDataEventFilter)
	}
	if len(flatDataEventFilters) > 0 {
		flatTrailFilteringPolicy["data_events_filter"] = flatDataEventFilters
	}

	managementFilter := filteringPolicy.GetManagementEventsFilter()
	if len(managementFilter.GetResourceScopes()) > 0 {
		flatManagementFilter := map[string]interface{}{}
		flatManagementFilter["resource_scope"] = unpackProtoResourceScopesIntoResourceData(managementFilter.GetResourceScopes())

		flatTrailFilteringPolicy["management_events_filter"] = []interface{}{flatManagementFilter}
	}

	result = setAndAppendError(data, "filtering_policy", []interface{}{flatTrailFilteringPolicy}, result)

	return result
}

func unpackEventTypesIntoResourceData(eventTypes *audittrails.Trail_EventTypes) []string {
	return eventTypes.GetEventTypes()
}

func unpackProtoResourceScopesIntoResourceData(resources []*audittrails.Trail_Resource) []interface{} {
	flatResourceScopes := []interface{}{}
	for _, resource := range resources {
		flatResourceScopes = append(flatResourceScopes, unpackProtoResourceIntoResourceData(resource))
	}
	return flatResourceScopes
}

func unpackProtoResourceIntoResourceData(resource *audittrails.Trail_Resource) interface{} {
	return map[string]string{
		"resource_id":   resource.GetId(),
		"resource_type": resource.GetType(),
	}
}

func setAndAppendError(data *schema.ResourceData, key string, value interface{}, accumulator diag.Diagnostics) diag.Diagnostics {
	if err := data.Set(key, value); err != nil {
		return append(accumulator, diagnosticFromError(err))
	}
	return accumulator
}

func diagnosticFromError(err error) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: 0,
		Summary:  err.Error(),
	}
}
