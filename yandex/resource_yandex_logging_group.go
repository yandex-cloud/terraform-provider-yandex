package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexLoggingGroupDefaultTimeout = 30 * time.Second

func resourceYandexLoggingGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexLoggingGroupCreate,
		Read:   resourceYandexLoggingGroupRead,
		Update: resourceYandexLoggingGroupUpdate,
		Delete: performYandexLoggingGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(yandexLoggingGroupDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"folder_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"retention_period": {
				Type:             schema.TypeString,
				Computed:         true,
				Optional:         true,
				ValidateFunc:     validateParsableValue(parseDuration),
				DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
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

			"cloud_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexLoggingGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("error getting folder ID while creating log group: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("error expanding labels while creating log group: %s", err)
	}

	retentionPeriod, err := parseDuration(d.Get("retention_period").(string))
	if err != nil {
		return fmt.Errorf("error parsing retention_period while creating log group: %s", err)
	}

	req := logging.CreateLogGroupRequest{
		FolderId:        folderID,
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		Labels:          labels,
		RetentionPeriod: retentionPeriod,
	}

	if err := performYandexLoggingGroupCreate(d, config, &req); err != nil {
		return err
	}

	return resourceYandexLoggingGroupRead(d, meta)
}

func performYandexLoggingGroupCreate(d *schema.ResourceData, config *Config, req *logging.CreateLogGroupRequest) error {
	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Logging().LogGroup().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create log group: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while get log group create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*logging.CreateLogGroupMetadata)
	if !ok {
		return fmt.Errorf("could not get log group ID from create operation metadata")
	}

	d.SetId(md.LogGroupId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting operation to create log group: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("log group creation failed: %s", err)
	}
	return nil
}

func resourceYandexLoggingGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req := logging.UpdateLogGroupRequest{
		LogGroupId: d.Id(),
		UpdateMask: &field_mask.FieldMask{},
	}

	if err := performYandexLoggingGroupUpdate(d, config, &req); err != nil {
		return err
	}

	return resourceYandexLoggingGroupRead(d, meta)
}

func performYandexLoggingGroupUpdate(d *schema.ResourceData, config *Config, req *logging.UpdateLogGroupRequest) error {
	d.Partial(true)

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	if d.HasChange("retention_period") {
		retentionPeriod, err := parseDuration(d.Get("retention_period").(string))
		if err != nil {
			return fmt.Errorf("error parsing retention_period while updating log group: %s", err)
		}
		req.RetentionPeriod = retentionPeriod
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "retention_period")
	}

	if d.HasChange("labels") {
		labelsProp, err := expandLabels(d.Get("labels"))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Logging().LogGroup().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update log group: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error updating log group %q: %s", d.Id(), err)
	}

	d.Partial(false)

	return nil
}

func performYandexLoggingGroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.Logging().LogGroup().Delete(ctx, &logging.DeleteLogGroupRequest{LogGroupId: d.Id()})
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Logging group %q", d.Id()))
	}

	return nil
}

func resourceYandexLoggingGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	group, err := performYandexLoggingGroupRead(d, config)
	if err != nil {
		return err
	}

	return flattenYandexLoggingGroup(d, group)
}

func performYandexLoggingGroupRead(d *schema.ResourceData, config *Config) (*logging.LogGroup, error) {
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	group, err := config.sdk.Logging().LogGroup().Get(ctx, &logging.GetLogGroupRequest{LogGroupId: d.Id()})
	if err != nil {
		return nil, handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Logging group %q", d.Get("name").(string)))
	}

	return group, nil
}

func flattenYandexLoggingGroup(d *schema.ResourceData, group *logging.LogGroup) error {
	d.Set("name", group.Name)
	d.Set("folder_id", group.FolderId)
	d.Set("retention_period", formatDuration(group.RetentionPeriod))
	d.Set("description", group.Description)
	d.Set("status", group.Status.String())
	d.Set("cloud_id", group.CloudId)
	d.Set("created_at", getTimestamp(group.CreatedAt))
	return d.Set("labels", group.Labels)
}
