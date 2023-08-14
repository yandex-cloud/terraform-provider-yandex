package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

const yandexFunctionDefaultTimeout = 10 * time.Minute

func resourceYandexFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexFunctionCreate,
		Read:   resourceYandexFunctionRead,
		Update: resourceYandexFunctionUpdate,
		Delete: resourceYandexFunctionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexFunctionDefaultTimeout),
			Update: schema.DefaultTimeout(yandexFunctionDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexFunctionDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"user_hash": {
				Type:     schema.TypeString,
				Required: true,
			},

			"runtime": {
				Type:     schema.TypeString,
				Required: true,
			},

			"entrypoint": {
				Type:     schema.TypeString,
				Required: true,
			},

			"memory": {
				Type:     schema.TypeInt,
				Required: true,
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

			"execution_timeout": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"environment": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"package": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"content"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"object_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"sha_256": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"content": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"package"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zip_filename": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"image_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"loggroup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secrets": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"version_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"environment_variable": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"connectivity": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Yandex Cloud Function: %s", err)
	}

	versionReq, err := expandFunctionVersion(d)
	if err != nil {
		return err
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating Yandex Cloud Function: %s", err)
	}

	req := functions.CreateFunctionRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
	}

	op, err := config.sdk.WrapOperation(config.sdk.Serverless().Functions().Function().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Function: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Function: %s", err)
	}

	md, ok := protoMetadata.(*functions.CreateFunctionMetadata)
	if !ok {
		return fmt.Errorf("Could not get Yandex Cloud Function ID from create operation metadata")
	}

	d.SetId(md.FunctionId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Function: %s", err)
	}

	if versionReq != nil {
		versionReq.FunctionId = md.FunctionId
		op, err = config.sdk.WrapOperation(config.sdk.Serverless().Functions().Function().CreateVersion(ctx, versionReq))
		if err != nil {
			return fmt.Errorf("Error while requesting API to create version for Yandex Cloud Function: %s", err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error while requesting API to create version for Yandex Cloud Function: %s", err)
		}
	}

	return resourceYandexFunctionRead(d, meta)
}

func resourceYandexFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while updating Yandex Cloud Function: %s", err)
	}

	d.Partial(true)

	var updatePaths []string
	if d.HasChange("name") {
		updatePaths = append(updatePaths, "name")
	}

	if d.HasChange("description") {
		updatePaths = append(updatePaths, "description")
	}

	if d.HasChange("labels") {
		updatePaths = append(updatePaths, "labels")
	}

	lastVersionPaths := []string{
		"user_hash", "runtime", "entrypoint", "memory", "execution_timeout", "service_account_id",
		"environment", "tags", "package", "content", "secrets", "connectivity",
	}
	var versionPartialPaths []string
	for _, p := range lastVersionPaths {
		if d.HasChange(p) {
			versionPartialPaths = append(versionPartialPaths, p)
		}
	}

	var versionReq *functions.CreateFunctionVersionRequest
	if len(versionPartialPaths) != 0 {
		versionReq, err = expandFunctionVersion(d)
		if err != nil {
			return err
		}
	}

	if len(updatePaths) != 0 {
		req := functions.UpdateFunctionRequest{
			FunctionId:  d.Id(),
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Labels:      labels,
			UpdateMask:  &field_mask.FieldMask{Paths: updatePaths},
		}

		op, err := config.sdk.Serverless().Functions().Function().Update(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return fmt.Errorf("Error while requesting API to update Yandex Cloud Function: %s", err)
		}

	}

	if versionReq != nil {
		versionReq.FunctionId = d.Id()
		op, err := config.sdk.WrapOperation(config.sdk.Serverless().Functions().Function().CreateVersion(ctx, versionReq))
		if err != nil {
			return fmt.Errorf("Error while requesting API to create version for Yandex Cloud Function: %s", err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error while requesting API to create version for Yandex Cloud Function: %s", err)
		}

	}
	d.Partial(false)

	return resourceYandexFunctionRead(d, meta)
}

func resourceYandexFunctionRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := functions.GetFunctionRequest{
		FunctionId: d.Id(),
	}

	function, err := config.sdk.Serverless().Functions().Function().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Function %q", d.Id()))
	}

	versionReq := functions.GetFunctionVersionByTagRequest{
		FunctionId: d.Id(),
		Tag:        "$latest",
	}

	version, err := config.sdk.Serverless().Functions().Function().GetVersionByTag(ctx, &versionReq)
	if err != nil {
		return err
	}

	return flattenYandexFunction(d, function, version)
}

func resourceYandexFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := functions.DeleteFunctionRequest{
		FunctionId: d.Id(),
	}

	op, err := config.sdk.Serverless().Functions().Function().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Function %q", d.Id()))
	}

	return nil
}

func flattenYandexFunction(d *schema.ResourceData, function *functions.Function, version *functions.Version) error {
	d.Set("name", function.Name)
	d.Set("folder_id", function.FolderId)
	d.Set("description", function.Description)
	d.Set("created_at", getTimestamp(function.CreatedAt))
	if err := d.Set("labels", function.Labels); err != nil {
		return err
	}

	if version == nil {
		return nil
	}

	d.Set("version", version.Id)
	return flattenYandexFunctionVersion(d, version)
}
