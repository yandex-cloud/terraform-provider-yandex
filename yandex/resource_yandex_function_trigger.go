package yandex

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/triggers/v1"
)

const (
	yandexFunctionTriggerDefaultTimeout = 5 * time.Minute

	triggerTypeIoT               = "iot"
	triggerTypeMessageQueue      = "message_queue"
	triggerTypeObjectStorage     = "object_storage"
	triggerTypeContainerRegistry = "container_registry"
	triggerTypeTimer             = "timer"
	triggerTypeLogGroup          = "log_group"
	triggerTypeLogging           = "logging"
	triggerTypeYDS               = "data_streams"
	triggerTypeMail              = "mail"
)

var functionTriggerTypesList = []string{
	triggerTypeIoT,
	triggerTypeMessageQueue,
	triggerTypeObjectStorage,
	triggerTypeContainerRegistry,
	triggerTypeTimer,
	triggerTypeLogGroup,
	triggerTypeLogging,
	triggerTypeYDS,
	triggerTypeMail,
}

var levelNameToEnum = map[string]logging.LogLevel_Level{
	"debug": logging.LogLevel_DEBUG,
	"error": logging.LogLevel_ERROR,
	"fatal": logging.LogLevel_FATAL,
	"info":  logging.LogLevel_INFO,
	"trace": logging.LogLevel_TRACE,
	"warn":  logging.LogLevel_WARN,
}

var levelEnumToName = map[logging.LogLevel_Level]string{
	logging.LogLevel_DEBUG: "debug",
	logging.LogLevel_ERROR: "error",
	logging.LogLevel_FATAL: "fatal",
	logging.LogLevel_INFO:  "info",
	logging.LogLevel_TRACE: "trace",
	logging.LogLevel_WARN:  "warn",
}

func resourceYandexFunctionTrigger() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexFunctionTriggerCreate,
		Read:   resourceYandexFunctionTriggerRead,
		Update: resourceYandexFunctionTriggerUpdate,
		Delete: resourceYandexFunctionTriggerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexFunctionTriggerDefaultTimeout),
			Update: schema.DefaultTimeout(yandexFunctionTriggerDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexFunctionTriggerDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"function": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"container"},
				ExactlyOneOf:  []string{"function", "container"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"tag": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"retry_attempts": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"retry_interval": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"container": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"function"},
				ExactlyOneOf:  []string{"function", "container"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"path": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"retry_attempts": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"retry_interval": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			triggerTypeIoT: {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: functionTriggerConflictingTypes(triggerTypeIoT),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"registry_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"device_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"topic": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Required: true,
						},
						"batch_size": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			triggerTypeMessageQueue: {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: functionTriggerConflictingTypes(triggerTypeMessageQueue),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Required: true,
						},

						"batch_size": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"visibility_timeout": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			triggerTypeObjectStorage: {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: functionTriggerConflictingTypes(triggerTypeObjectStorage),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"suffix": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"create": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"update": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"delete": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Required: true,
						},
						"batch_size": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			triggerTypeContainerRegistry: {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: functionTriggerConflictingTypes(triggerTypeContainerRegistry),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"registry_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"image_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tag": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"create_image": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"delete_image": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"create_image_tag": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"delete_image_tag": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Required: true,
						},
						"batch_size": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			triggerTypeYDS: {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: functionTriggerConflictingTypes(triggerTypeYDS),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stream_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"database": {
							Type:     schema.TypeString,
							Required: true,
						},
						"service_account_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Required: true,
						},
						"batch_size": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			triggerTypeTimer: {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: functionTriggerConflictingTypes(triggerTypeTimer),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cron_expression": {
							Type:     schema.TypeString,
							Required: true,
						},
						"payload": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			triggerTypeMail: {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: functionTriggerConflictingTypes(triggerTypeMail),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attachments_bucket_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"service_account_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Required: true,
						},
						"batch_size": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			triggerTypeLogGroup: {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: functionTriggerConflictingTypes(triggerTypeLogGroup),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							MinItems: 1,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"batch_size": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			triggerTypeLogging: {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: functionTriggerConflictingTypes(triggerTypeLogging),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"resource_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							MinItems: 0,
						},

						"resource_types": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							MinItems: 0,
						},

						"levels": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							MinItems: 0,
						},

						"stream_names": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							MinItems: 0,
						},

						"batch_cutoff": {
							Type:     schema.TypeString,
							Required: true,
						},

						"batch_size": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"dlq": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func constructRule(d *schema.ResourceData) (*triggers.Trigger_Rule, error) {
	var invokeType string
	if _, ok := d.GetOk("function"); ok {
		invokeType = "function"
	}
	if _, ok := d.GetOk("container"); ok {
		invokeType = "container"
	}

	retrySettings, err := expandRetrySettings(d, invokeType)
	if err != nil {
		return nil, err
	}

	dlqSettings, err := expandDLQSettings(d)
	if err != nil {
		return nil, err
	}

	var getInvokeFunctionWithRetry = func() *triggers.InvokeFunctionWithRetry {
		return &triggers.InvokeFunctionWithRetry{
			FunctionId:       d.Get("function.0.id").(string),
			FunctionTag:      d.Get("function.0.tag").(string),
			ServiceAccountId: d.Get("function.0.service_account_id").(string),
			RetrySettings:    retrySettings,
			DeadLetterQueue:  dlqSettings,
		}
	}

	var getInvokeFunctionOnce = func() *triggers.InvokeFunctionOnce {
		return &triggers.InvokeFunctionOnce{
			FunctionId:       d.Get("function.0.id").(string),
			FunctionTag:      d.Get("function.0.tag").(string),
			ServiceAccountId: d.Get("function.0.service_account_id").(string),
		}
	}

	var getInvokeContainerWithRetry = func() *triggers.InvokeContainerWithRetry {
		return &triggers.InvokeContainerWithRetry{
			ContainerId:      d.Get("container.0.id").(string),
			Path:             d.Get("container.0.path").(string),
			ServiceAccountId: d.Get("container.0.service_account_id").(string),
			RetrySettings:    retrySettings,
			DeadLetterQueue:  dlqSettings,
		}
	}

	var getInvokeContainerOnce = func() *triggers.InvokeContainerOnce {
		return &triggers.InvokeContainerOnce{
			ContainerId:      d.Get("container.0.id").(string),
			Path:             d.Get("container.0.path").(string),
			ServiceAccountId: d.Get("container.0.service_account_id").(string),
		}
	}

	if _, ok := d.GetOk(triggerTypeIoT); ok {
		iot := &triggers.Trigger_Rule_IotMessage{
			IotMessage: &triggers.Trigger_IoTMessage{
				RegistryId: d.Get("iot.0.registry_id").(string),
				DeviceId:   d.Get("iot.0.device_id").(string),
				MqttTopic:  d.Get("iot.0.topic").(string),
			},
		}

		if invokeType == "function" {
			iot.IotMessage.Action = &triggers.Trigger_IoTMessage_InvokeFunction{
				InvokeFunction: getInvokeFunctionWithRetry(),
			}
		} else if invokeType == "container" {
			iot.IotMessage.Action = &triggers.Trigger_IoTMessage_InvokeContainer{
				InvokeContainer: getInvokeContainerWithRetry(),
			}
		}

		batch, err := expandBatchSettings(d, "iot.0")
		if err != nil {
			return nil, err
		}
		iot.IotMessage.BatchSettings = batch

		return &triggers.Trigger_Rule{Rule: iot}, nil
	} else if _, ok := d.GetOk(triggerTypeMessageQueue); ok {
		if err := checkDisableRetrySettingsForMessageQueueTrigger(d, invokeType); err != nil {
			return nil, err
		}
		batch, err := expandBatchSettings(d, "message_queue.0")
		if err != nil {
			return nil, err
		}

		messageQueue := &triggers.Trigger_MessageQueue{
			QueueId:          d.Get("message_queue.0.queue_id").(string),
			ServiceAccountId: d.Get("message_queue.0.service_account_id").(string),
			BatchSettings:    batch,
		}

		if invokeType == "function" {
			messageQueue.Action = &triggers.Trigger_MessageQueue_InvokeFunction{
				InvokeFunction: getInvokeFunctionOnce(),
			}
		} else if invokeType == "container" {
			messageQueue.Action = &triggers.Trigger_MessageQueue_InvokeContainer{
				InvokeContainer: getInvokeContainerOnce(),
			}
		}

		if _, ok := d.GetOk("message_queue.0.visibility_timeout"); ok {
			timeout, err := strconv.ParseInt(d.Get("message_queue.0.visibility_timeout").(string), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("Cannot define message_queue.visibility_timeout for Yandex Cloud Functions Trigger: %s", err)
			}
			messageQueue.VisibilityTimeout = &duration.Duration{Seconds: timeout}
		}

		messageQueueRule := &triggers.Trigger_Rule_MessageQueue{MessageQueue: messageQueue}
		return &triggers.Trigger_Rule{Rule: messageQueueRule}, nil
	} else if _, ok := d.GetOk(triggerTypeObjectStorage); ok {
		events := make([]triggers.Trigger_ObjectStorageEventType, 0)
		eventsName := map[string]triggers.Trigger_ObjectStorageEventType{
			"object_storage.0.create": triggers.Trigger_OBJECT_STORAGE_EVENT_TYPE_CREATE_OBJECT,
			"object_storage.0.update": triggers.Trigger_OBJECT_STORAGE_EVENT_TYPE_UPDATE_OBJECT,
			"object_storage.0.delete": triggers.Trigger_OBJECT_STORAGE_EVENT_TYPE_DELETE_OBJECT,
		}
		for k, v := range eventsName {
			if d.Get(k).(bool) {
				events = append(events, v)
			}
		}

		storageTrigger := &triggers.Trigger_ObjectStorage{
			BucketId:  d.Get("object_storage.0.bucket_id").(string),
			Prefix:    d.Get("object_storage.0.prefix").(string),
			Suffix:    d.Get("object_storage.0.suffix").(string),
			EventType: events,
		}

		if invokeType == "function" {
			storageTrigger.Action = &triggers.Trigger_ObjectStorage_InvokeFunction{
				InvokeFunction: getInvokeFunctionWithRetry(),
			}
		} else if invokeType == "container" {
			storageTrigger.Action = &triggers.Trigger_ObjectStorage_InvokeContainer{
				InvokeContainer: getInvokeContainerWithRetry(),
			}
		}

		batch, err := expandBatchSettings(d, "object_storage.0")
		if err != nil {
			return nil, err
		}
		storageTrigger.BatchSettings = batch
		storageRule := &triggers.Trigger_Rule_ObjectStorage{ObjectStorage: storageTrigger}
		return &triggers.Trigger_Rule{Rule: storageRule}, nil
	} else if _, ok := d.GetOk(triggerTypeContainerRegistry); ok {
		events := make([]triggers.Trigger_ContainerRegistryEventType, 0)
		eventsName := map[string]triggers.Trigger_ContainerRegistryEventType{
			"container_registry.0.create_image":     triggers.Trigger_CONTAINER_REGISTRY_EVENT_TYPE_CREATE_IMAGE,
			"container_registry.0.delete_image":     triggers.Trigger_CONTAINER_REGISTRY_EVENT_TYPE_DELETE_IMAGE,
			"container_registry.0.create_image_tag": triggers.Trigger_CONTAINER_REGISTRY_EVENT_TYPE_CREATE_IMAGE_TAG,
			"container_registry.0.delete_image_tag": triggers.Trigger_CONTAINER_REGISTRY_EVENT_TYPE_DELETE_IMAGE_TAG,
		}
		for k, v := range eventsName {
			if d.Get(k).(bool) {
				events = append(events, v)
			}
		}

		crTrigger := &triggers.Trigger_ContainerRegistry{
			RegistryId: d.Get("container_registry.0.registry_id").(string),
			ImageName:  d.Get("container_registry.0.image_name").(string),
			Tag:        d.Get("container_registry.0.tag").(string),
			EventType:  events,
		}

		if invokeType == "function" {
			crTrigger.Action = &triggers.Trigger_ContainerRegistry_InvokeFunction{
				InvokeFunction: getInvokeFunctionWithRetry(),
			}
		} else if invokeType == "container" {
			crTrigger.Action = &triggers.Trigger_ContainerRegistry_InvokeContainer{
				InvokeContainer: getInvokeContainerWithRetry(),
			}
		}

		batch, err := expandBatchSettings(d, "container_registry.0")
		if err != nil {
			return nil, err
		}
		crTrigger.BatchSettings = batch
		storageRule := &triggers.Trigger_Rule_ContainerRegistry{ContainerRegistry: crTrigger}
		return &triggers.Trigger_Rule{Rule: storageRule}, nil
	} else if _, ok := d.GetOk(triggerTypeTimer); ok {
		timer := triggers.Trigger_Timer{
			CronExpression: d.Get("timer.0.cron_expression").(string),
		}
		if v, ok := d.GetOk("timer.0.payload"); ok {
			timer.Payload = v.(string)
		}

		if retrySettings != nil || dlqSettings != nil {
			if invokeType == "function" {
				timer.Action = &triggers.Trigger_Timer_InvokeFunctionWithRetry{
					InvokeFunctionWithRetry: getInvokeFunctionWithRetry(),
				}
			} else if invokeType == "container" {
				timer.Action = &triggers.Trigger_Timer_InvokeContainerWithRetry{
					InvokeContainerWithRetry: getInvokeContainerWithRetry(),
				}
			}
		} else {
			timer.Action = &triggers.Trigger_Timer_InvokeFunction{
				InvokeFunction: getInvokeFunctionOnce(),
			}
		}

		timerRule := &triggers.Trigger_Rule_Timer{Timer: &timer}
		return &triggers.Trigger_Rule{Rule: timerRule}, nil
	} else if _, ok := d.GetOk(triggerTypeLogGroup); ok {
		cloudLogs := &triggers.Trigger_CloudLogs{
			LogGroupId: convertStringSet(d.Get("log_group.0.log_group_ids").(*schema.Set)),
		}

		if invokeType == "function" {
			cloudLogs.Action = &triggers.Trigger_CloudLogs_InvokeFunction{
				InvokeFunction: getInvokeFunctionWithRetry(),
			}
		} else if invokeType == "container" {
			cloudLogs.Action = &triggers.Trigger_CloudLogs_InvokeContainer{
				InvokeContainer: getInvokeContainerWithRetry(),
			}
		}

		batch, err := expandBatchSettings(d, "log_group.0")
		if err != nil {
			return nil, err
		}
		if batch != nil {
			cloudLogs.BatchSettings = &triggers.CloudLogsBatchSettings{
				Size:   batch.Size,
				Cutoff: batch.Cutoff,
			}
		}
		return &triggers.Trigger_Rule{
			Rule: &triggers.Trigger_Rule_CloudLogs{CloudLogs: cloudLogs},
		}, nil
	} else if _, ok := d.GetOk(triggerTypeYDS); ok {
		yds := &triggers.DataStream{
			Stream:           d.Get("data_streams.0.stream_name").(string),
			Database:         d.Get("data_streams.0.database").(string),
			ServiceAccountId: d.Get("data_streams.0.service_account_id").(string),
		}

		if invokeType == "function" {
			yds.Action = &triggers.DataStream_InvokeFunction{
				InvokeFunction: getInvokeFunctionWithRetry(),
			}
		} else if invokeType == "container" {
			yds.Action = &triggers.DataStream_InvokeContainer{
				InvokeContainer: getInvokeContainerWithRetry(),
			}
		}
		batch, err := expandBatchSettings(d, "data_streams.0")
		if err != nil {
			return nil, err
		}
		if batch != nil {
			yds.BatchSettings = &triggers.DataStreamBatchSettings{
				Size:   batch.Size,
				Cutoff: batch.Cutoff,
			}
		}
		return &triggers.Trigger_Rule{
			Rule: &triggers.Trigger_Rule_DataStream{DataStream: yds},
		}, nil
	} else if _, ok := d.GetOk(triggerTypeMail); ok {
		mail := &triggers.Mail{}

		if invokeType == "function" {
			mail.Action = &triggers.Mail_InvokeFunction{
				InvokeFunction: getInvokeFunctionWithRetry(),
			}
		} else if invokeType == "container" {
			mail.Action = &triggers.Mail_InvokeContainer{
				InvokeContainer: getInvokeContainerWithRetry(),
			}
		}

		bucket, hasBucket := d.GetOk("mail.0.attachments_bucket_id")
		sa, hasSA := d.GetOk("mail.0.service_account_id")
		if hasSA && hasBucket {
			mail.AttachmentsBucket = &triggers.ObjectStorageBucketSettings{
				BucketId:         bucket.(string),
				ServiceAccountId: sa.(string),
			}
		}

		batch, err := expandBatchSettings(d, "mail.0")
		if err != nil {
			return nil, err
		}
		mail.BatchSettings = batch
		return &triggers.Trigger_Rule{
			Rule: &triggers.Trigger_Rule_Mail{Mail: mail},
		}, nil
	} else if _, ok := d.GetOk(triggerTypeLogging); ok {
		levels := []logging.LogLevel_Level{}

		for _, l := range convertStringSet(d.Get("logging.0.levels").(*schema.Set)) {
			if v, ok := levelNameToEnum[strings.ToLower(l)]; ok {
				levels = append(levels, v)
			}
		}

		logging := &triggers.Trigger_Logging{
			LogGroupId:   d.Get("logging.0.group_id").(string),
			ResourceId:   convertStringSet(d.Get("logging.0.resource_ids").(*schema.Set)),
			ResourceType: convertStringSet(d.Get("logging.0.resource_types").(*schema.Set)),
			StreamName:   convertStringSet(d.Get("logging.0.stream_names").(*schema.Set)),
			Levels:       levels,
		}

		if invokeType == "function" {
			logging.Action = &triggers.Trigger_Logging_InvokeFunction{
				InvokeFunction: getInvokeFunctionWithRetry(),
			}
		} else if invokeType == "container" {
			logging.Action = &triggers.Trigger_Logging_InvokeContainer{
				InvokeContainer: getInvokeContainerWithRetry(),
			}
		}

		batch, err := expandBatchSettings(d, "logging.0")
		if err != nil {
			return nil, err
		}
		if batch != nil {
			logging.BatchSettings = &triggers.LoggingBatchSettings{
				Size:   batch.Size,
				Cutoff: batch.Cutoff,
			}
		}
		return &triggers.Trigger_Rule{
			Rule: &triggers.Trigger_Rule_Logging{Logging: logging},
		}, nil
	}
	return nil, errors.New("Unknown rule type")
}

func resourceYandexFunctionTriggerCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Yandex Cloud Functions Trigger: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating Yandex Cloud Functions Trigger: %s", err)
	}

	rule, err := constructRule(d)
	if err != nil {
		return fmt.Errorf("Error constructing rule while creating Yandex Cloud Functions Trigger: %s", err)
	}

	req := triggers.CreateTriggerRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		Rule:        rule,
	}

	op, err := config.sdk.WrapOperation(config.sdk.Serverless().Triggers().Trigger().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Functions Trigger: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Function Trigger: %s", err)
	}

	md, ok := protoMetadata.(*triggers.CreateTriggerMetadata)
	if !ok {
		return fmt.Errorf("Could not get Yandex Cloud Functions Trigger ID from create operation metadata")
	}

	d.SetId(md.TriggerId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Functions Trigger: %s", err)
	}

	return resourceYandexFunctionTriggerRead(d, meta)
}

func resourceYandexFunctionTriggerUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Yandex Cloud Functions Trigger: %s", err)
	}

	d.Partial(true)
	rule, err := constructRule(d)
	if err != nil {
		return err
	}

	req := triggers.UpdateTriggerRequest{
		TriggerId:   d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		Rule:        rule,
	}

	op, err := config.sdk.Serverless().Triggers().Trigger().Update(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Yandex Cloud Functions Trigger: %s", err)
	}

	d.Partial(false)

	return resourceYandexFunctionTriggerRead(d, meta)
}

func flattenYandexFunctionTriggerInvokeOnce(d *schema.ResourceData, function *triggers.InvokeFunctionOnce) error {
	f := map[string]interface{}{
		"id":                 function.FunctionId,
		"tag":                function.FunctionTag,
		"service_account_id": function.ServiceAccountId,
	}
	return d.Set("function", []map[string]interface{}{f})
}

func flattenYandexFunctionTriggerInvokeWithRetry(d *schema.ResourceData, function *triggers.InvokeFunctionWithRetry) error {
	f := map[string]interface{}{
		"id":                 function.FunctionId,
		"tag":                function.FunctionTag,
		"service_account_id": function.ServiceAccountId,
	}

	if retrySettings := function.GetRetrySettings(); retrySettings != nil {
		f["retry_attempts"] = strconv.FormatInt(retrySettings.RetryAttempts, 10)
		if retrySettings.Interval != nil {
			f["retry_interval"] = strconv.FormatInt(retrySettings.Interval.Seconds, 10)
		}
	}

	err := d.Set("function", []map[string]interface{}{f})
	if err != nil {
		return err
	}

	if deadLetter := function.GetDeadLetterQueue(); deadLetter != nil {
		dlq := map[string]interface{}{
			"queue_id":           deadLetter.QueueId,
			"service_account_id": deadLetter.ServiceAccountId,
		}

		err = d.Set("dlq", []map[string]interface{}{dlq})
		if err != nil {
			return err
		}
	}
	return nil
}

func flattenYandexContainerTriggerInvokeOnce(d *schema.ResourceData, container *triggers.InvokeContainerOnce) error {
	f := map[string]interface{}{
		"id":                 container.ContainerId,
		"path":               container.Path,
		"service_account_id": container.ServiceAccountId,
	}
	return d.Set("container", []map[string]interface{}{f})
}

func flattenYandexContainerTriggerInvokeWithRetry(d *schema.ResourceData, container *triggers.InvokeContainerWithRetry) error {
	f := map[string]interface{}{
		"id":                 container.ContainerId,
		"path":               container.Path,
		"service_account_id": container.ServiceAccountId,
	}

	if retrySettings := container.GetRetrySettings(); retrySettings != nil {
		f["retry_attempts"] = strconv.FormatInt(retrySettings.RetryAttempts, 10)
		if retrySettings.Interval != nil {
			f["retry_interval"] = strconv.FormatInt(retrySettings.Interval.Seconds, 10)
		}
	}

	err := d.Set("container", []map[string]interface{}{f})
	if err != nil {
		return err
	}

	if deadLetter := container.GetDeadLetterQueue(); deadLetter != nil {
		dlq := map[string]interface{}{
			"queue_id":           deadLetter.QueueId,
			"service_account_id": deadLetter.ServiceAccountId,
		}

		err = d.Set("dlq", []map[string]interface{}{dlq})
		if err != nil {
			return err
		}
	}
	return nil
}

func flattenYandexFunctionTrigger(d *schema.ResourceData, trig *triggers.Trigger) error {
	d.Set("name", trig.Name)
	d.Set("folder_id", trig.FolderId)
	d.Set("description", trig.Description)
	d.Set("created_at", getTimestamp(trig.CreatedAt))
	if err := d.Set("labels", trig.Labels); err != nil {
		return err
	}

	if iot := trig.GetRule().GetIotMessage(); iot != nil {
		i := map[string]interface{}{
			"registry_id": iot.RegistryId,
			"device_id":   iot.DeviceId,
			"topic":       iot.MqttTopic,
		}

		if batch := iot.GetBatchSettings(); batch != nil {
			i["batch_size"] = strconv.FormatInt(batch.Size, 10)
			i["batch_cutoff"] = strconv.FormatInt(batch.Cutoff.Seconds, 10)
		}

		err := d.Set(triggerTypeIoT, []map[string]interface{}{i})
		if err != nil {
			return err
		}

		if function := iot.GetInvokeFunction(); function != nil {
			err = flattenYandexFunctionTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		} else if function := iot.GetInvokeContainer(); function != nil {
			err = flattenYandexContainerTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		}
	} else if yds := trig.GetRule().GetDataStream(); yds != nil {
		i := map[string]interface{}{
			"database":           yds.Database,
			"stream_name":        yds.Stream,
			"service_account_id": yds.ServiceAccountId,
		}

		if batch := yds.GetBatchSettings(); batch != nil {
			i["batch_size"] = strconv.FormatInt(batch.Size, 10)
			i["batch_cutoff"] = strconv.FormatInt(batch.Cutoff.Seconds, 10)
		}

		err := d.Set(triggerTypeYDS, []map[string]interface{}{i})
		if err != nil {
			return err
		}

		if function := yds.GetInvokeFunction(); function != nil {
			err = flattenYandexFunctionTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		} else if function := yds.GetInvokeContainer(); function != nil {
			err = flattenYandexContainerTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		}
	} else if mail := trig.GetRule().GetMail(); mail != nil {
		i := map[string]interface{}{}

		if bucket := mail.AttachmentsBucket; bucket != nil {
			i["attachments_bucket_id"] = bucket.BucketId
			i["service_account_id"] = bucket.ServiceAccountId
		}

		if batch := mail.GetBatchSettings(); batch != nil {
			i["batch_size"] = strconv.FormatInt(batch.Size, 10)
			i["batch_cutoff"] = strconv.FormatInt(batch.Cutoff.Seconds, 10)
		}

		err := d.Set(triggerTypeMail, []map[string]interface{}{i})
		if err != nil {
			return err
		}

		if function := mail.GetInvokeFunction(); function != nil {
			err = flattenYandexFunctionTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		} else if function := mail.GetInvokeContainer(); function != nil {
			err = flattenYandexContainerTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		}
	} else if messageQueue := trig.GetRule().GetMessageQueue(); messageQueue != nil {
		m := map[string]interface{}{
			"queue_id":           messageQueue.QueueId,
			"service_account_id": messageQueue.ServiceAccountId,
		}

		if messageQueue.VisibilityTimeout != nil {
			m["visibility_timeout"] = strconv.FormatInt(messageQueue.VisibilityTimeout.Seconds, 10)
		}

		if batch := messageQueue.GetBatchSettings(); batch != nil {
			m["batch_size"] = strconv.FormatInt(batch.Size, 10)
			m["batch_cutoff"] = strconv.FormatInt(batch.Cutoff.Seconds, 10)
		}

		err := d.Set(triggerTypeMessageQueue, []map[string]interface{}{m})
		if err != nil {
			return err
		}

		if function := messageQueue.GetInvokeFunction(); function != nil {
			err = flattenYandexFunctionTriggerInvokeOnce(d, function)
			if err != nil {
				return err
			}
		} else if function := messageQueue.GetInvokeContainer(); function != nil {
			err = flattenYandexContainerTriggerInvokeOnce(d, function)
			if err != nil {
				return err
			}
		}
	} else if storage := trig.GetRule().GetObjectStorage(); storage != nil {
		s := map[string]interface{}{
			"bucket_id": storage.BucketId,
			"prefix":    storage.Prefix,
			"suffix":    storage.Suffix,
		}

		events := map[triggers.Trigger_ObjectStorageEventType]string{
			triggers.Trigger_OBJECT_STORAGE_EVENT_TYPE_CREATE_OBJECT: "create",
			triggers.Trigger_OBJECT_STORAGE_EVENT_TYPE_UPDATE_OBJECT: "update",
			triggers.Trigger_OBJECT_STORAGE_EVENT_TYPE_DELETE_OBJECT: "delete",
		}

		if batch := storage.GetBatchSettings(); batch != nil {
			s["batch_size"] = strconv.FormatInt(batch.Size, 10)
			s["batch_cutoff"] = strconv.FormatInt(batch.Cutoff.Seconds, 10)
		}

		for _, t := range storage.EventType {
			if _, ok := events[t]; ok {
				s[events[t]] = true
			}
		}

		err := d.Set(triggerTypeObjectStorage, []map[string]interface{}{s})
		if err != nil {
			return err
		}

		if function := storage.GetInvokeFunction(); function != nil {
			err = flattenYandexFunctionTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		} else if function := storage.GetInvokeContainer(); function != nil {
			err = flattenYandexContainerTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		}
	} else if cr := trig.GetRule().GetContainerRegistry(); cr != nil {
		s := map[string]interface{}{
			"registry_id": cr.RegistryId,
			"image_name":  cr.ImageName,
			"tag":         cr.Tag,
		}

		if batch := cr.GetBatchSettings(); batch != nil {
			s["batch_size"] = strconv.FormatInt(batch.Size, 10)
			s["batch_cutoff"] = strconv.FormatInt(batch.Cutoff.Seconds, 10)
		}

		events := map[triggers.Trigger_ContainerRegistryEventType]string{
			triggers.Trigger_CONTAINER_REGISTRY_EVENT_TYPE_CREATE_IMAGE:     "create_image",
			triggers.Trigger_CONTAINER_REGISTRY_EVENT_TYPE_DELETE_IMAGE:     "delete_image",
			triggers.Trigger_CONTAINER_REGISTRY_EVENT_TYPE_CREATE_IMAGE_TAG: "create_image_tag",
			triggers.Trigger_CONTAINER_REGISTRY_EVENT_TYPE_DELETE_IMAGE_TAG: "delete_image_tag",
		}

		for _, t := range cr.EventType {
			if _, ok := events[t]; ok {
				s[events[t]] = true
			}
		}

		err := d.Set(triggerTypeContainerRegistry, []map[string]interface{}{s})
		if err != nil {
			return err
		}

		if function := cr.GetInvokeFunction(); function != nil {
			err = flattenYandexFunctionTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		} else if function := cr.GetInvokeContainer(); function != nil {
			err = flattenYandexContainerTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		}
	} else if timer := trig.GetRule().GetTimer(); timer != nil {
		t := map[string]interface{}{
			"cron_expression": timer.CronExpression,
			"payload":         timer.Payload,
		}

		err := d.Set(triggerTypeTimer, []map[string]interface{}{t})
		if err != nil {
			return err
		}

		if function := timer.GetInvokeFunctionWithRetry(); function != nil {
			err = flattenYandexFunctionTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		} else if function := timer.GetInvokeFunction(); function != nil {
			err = flattenYandexFunctionTriggerInvokeOnce(d, function)
			if err != nil {
				return err
			}
		} else if function := timer.GetInvokeContainerWithRetry(); function != nil {
			err = flattenYandexContainerTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		}
	} else if logGroup := trig.GetRule().GetCloudLogs(); logGroup != nil {

		groupIDs := &schema.Set{F: schema.HashString}
		for _, groupID := range logGroup.LogGroupId {
			groupIDs.Add(groupID)
		}
		lg := map[string]interface{}{
			"log_group_ids": groupIDs,
		}
		if batch := logGroup.GetBatchSettings(); batch != nil {
			lg["batch_size"] = strconv.FormatInt(batch.Size, 10)
			lg["batch_cutoff"] = strconv.FormatInt(batch.Cutoff.Seconds, 10)
		}
		if function := logGroup.GetInvokeFunction(); function != nil {
			err := flattenYandexFunctionTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		} else if function := logGroup.GetInvokeContainer(); function != nil {
			err := flattenYandexContainerTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		}
		err := d.Set(triggerTypeLogGroup, []map[string]interface{}{lg})
		if err != nil {
			return err
		}
	} else if logging := trig.GetRule().GetLogging(); logging != nil {

		resourceIDs := &schema.Set{F: schema.HashString}
		for _, id := range logging.ResourceId {
			resourceIDs.Add(id)
		}
		resourceTypes := &schema.Set{F: schema.HashString}
		for _, t := range logging.ResourceType {
			resourceTypes.Add(t)
		}
		levels := &schema.Set{F: schema.HashString}
		for _, level := range logging.Levels {
			if l, ok := levelEnumToName[level]; ok {
				levels.Add(l)
			}
		}
		streamNames := &schema.Set{F: schema.HashString}
		for _, name := range logging.StreamName {
			streamNames.Add(name)
		}

		lg := map[string]interface{}{
			"group_id":       logging.LogGroupId,
			"resource_ids":   resourceIDs,
			"resource_types": resourceTypes,
			"levels":         levels,
			"stream_names":   streamNames,
		}
		if batch := logging.GetBatchSettings(); batch != nil {
			lg["batch_size"] = strconv.FormatInt(batch.Size, 10)
			lg["batch_cutoff"] = strconv.FormatInt(batch.Cutoff.Seconds, 10)
		}
		if function := logging.GetInvokeFunction(); function != nil {
			err := flattenYandexFunctionTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		} else if function := logGroup.GetInvokeContainer(); function != nil {
			err := flattenYandexContainerTriggerInvokeWithRetry(d, function)
			if err != nil {
				return err
			}
		}
		err := d.Set(triggerTypeLogging, []map[string]interface{}{lg})
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceYandexFunctionTriggerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := triggers.GetTriggerRequest{
		TriggerId: d.Id(),
	}

	trig, err := config.sdk.Serverless().Triggers().Trigger().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Functions Trigger %q", d.Id()))
	}

	return flattenYandexFunctionTrigger(d, trig)
}

func resourceYandexFunctionTriggerDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := triggers.DeleteTriggerRequest{
		TriggerId: d.Id(),
	}

	op, err := config.sdk.Serverless().Triggers().Trigger().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Functions Trigger %q", d.Id()))
	}

	return nil
}

func expandRetrySettings(d *schema.ResourceData, prefix string) (*triggers.RetrySettings, error) {
	settings := &triggers.RetrySettings{}
	var err error
	present := false

	if _, ok := d.GetOk(prefix + ".0.retry_attempts"); ok {
		present = true
		settings.RetryAttempts, err = strconv.ParseInt(d.Get(prefix+".0.retry_attempts").(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Cannot define %s.retry_attempts for Yandex Cloud Functions Trigger: %s", prefix, err)
		}
	}

	if _, ok := d.GetOk(prefix + ".0.retry_interval"); ok {
		present = true
		retryInterval, err := strconv.ParseInt(d.Get(prefix+".0.retry_interval").(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Cannot define %s.retry_interval for Yandex Cloud Functions Trigger: %s", prefix, err)
		}
		settings.Interval = &duration.Duration{Seconds: retryInterval}
	}

	if !present {
		return nil, nil
	}

	return settings, nil
}

func expandDLQSettings(d *schema.ResourceData) (*triggers.PutQueueMessage, error) {
	if _, ok := d.GetOk("dlq"); !ok {
		return nil, nil
	}

	settings := &triggers.PutQueueMessage{
		QueueId:          d.Get("dlq.0.queue_id").(string),
		ServiceAccountId: d.Get("dlq.0.service_account_id").(string),
	}

	return settings, nil
}

func checkDisableRetrySettingsForMessageQueueTrigger(d *schema.ResourceData, prefix string) error {
	keys := []string{"dlq", prefix + ".0.retry_attempts", prefix + ".0.retry_interval"}
	forOutput := []string{"dlq", prefix + ".retry_attempts", prefix + ".retry_interval"}
	for i, name := range keys {
		if _, found := d.GetOk(name); found {
			return fmt.Errorf("Cannot define %s for Yandex Cloud Functions Trigger: not supported for message queue trigger", forOutput[i])
		}
	}
	return nil
}

func expandBatchSettings(d *schema.ResourceData, prefix string) (settings *triggers.BatchSettings, err error) {
	if _, ok := d.GetOk(prefix + ".batch_size"); !ok {
		return nil, nil
	}

	settings = &triggers.BatchSettings{}
	settings.Size, err = strconv.ParseInt(d.Get(prefix+".batch_size").(string), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Cannot define "+prefix+".batch_size for Yandex Cloud Functions Trigger: %s", err)
	}

	batchCutoff, err := strconv.ParseInt(d.Get(prefix+".batch_cutoff").(string), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Cannot define "+prefix+"batch_cutoff for Yandex Cloud Functions Trigger: %s", err)
	}
	settings.Cutoff = &duration.Duration{Seconds: batchCutoff}

	return settings, nil
}

func functionTriggerConflictingTypes(triggerType string) []string {
	res := make([]string, 0, len(functionTriggerTypesList)-1)
	for _, tType := range functionTriggerTypesList {
		if tType != triggerType {
			res = append(res, tType)
		}
	}
	return res
}
