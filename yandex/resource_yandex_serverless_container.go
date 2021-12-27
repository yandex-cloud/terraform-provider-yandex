package yandex

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexServerlessContainerDefaultTimeout = 5 * time.Minute

func resourceYandexServerlessContainer() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexServerlessContainerCreate,
		Read:   resourceYandexServerlessContainerRead,
		Update: resourceYandexServerlessContainerUpdate,
		Delete: resourceYandexServerlessContainerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexServerlessContainerDefaultTimeout),
			Update: schema.DefaultTimeout(yandexServerlessContainerDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexServerlessContainerDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
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

			"memory": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Container memory in megabytes, should be aligned to 128",
			},

			"cores": {
				Type:     schema.TypeInt,
				Default:  1,
				Optional: true,
			},

			"core_fraction": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},

			"execution_timeout": {
				Type:             schema.TypeString,
				Computed:         true,
				Optional:         true,
				ValidateFunc:     validateParsableValue(parsePositiveDuration),
				DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
			},

			"concurrency": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"image": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"work_dir": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"digest": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"command": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"args": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"environment": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},

			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"revision_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexServerlessContainerCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Yandex Cloud Container: %s", err)
	}

	revisionReq, err := expandLastRevision(d)
	if err != nil {
		return err
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating Yandex Cloud Container: %s", err)
	}

	req := containers.CreateContainerRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
	}
	op, err := config.sdk.WrapOperation(config.sdk.Serverless().Containers().Container().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Container: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Container: %s", err)
	}
	md, ok := protoMetadata.(*containers.CreateContainerMetadata)
	if !ok {
		return fmt.Errorf("Could not get Yandex Cloud Container ID from create operation metadata")
	}
	d.SetId(md.ContainerId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Container: %s", err)
	}

	if revisionReq != nil {
		revisionReq.ContainerId = md.ContainerId
		op, err := config.sdk.Serverless().Containers().Container().DeployRevision(ctx, revisionReq)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return fmt.Errorf("Error while requesting API to deploy revision for Yandex Cloud Container: %s", err)
		}
	}

	return resourceYandexServerlessContainerRead(d, meta)
}

func resourceYandexServerlessContainerUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while updating Yandex Cloud Container: %s", err)
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

	lastRevisionPaths := []string{"memory", "cores", "core_fraction", "execution_timeout", "service_account_id", "image", "concurrency"}
	var revisionUpdatePaths []string
	for _, p := range lastRevisionPaths {
		if d.HasChange(p) {
			revisionUpdatePaths = append(revisionUpdatePaths, p)
		}
	}

	var revisionReq *containers.DeployContainerRevisionRequest
	if len(revisionUpdatePaths) != 0 {
		revisionReq, err = expandLastRevision(d)
		if err != nil {
			return err
		}
	}

	if len(updatePaths) != 0 {
		req := containers.UpdateContainerRequest{
			ContainerId: d.Id(),
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Labels:      labels,
			UpdateMask:  &field_mask.FieldMask{Paths: updatePaths},
		}

		op, err := config.sdk.Serverless().Containers().Container().Update(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return fmt.Errorf("Error while requesting API to update Yandex Cloud Container: %s", err)
		}
	}

	if revisionReq != nil {
		revisionReq.ContainerId = d.Id()

		op, err := config.sdk.Serverless().Containers().Container().DeployRevision(ctx, revisionReq)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return fmt.Errorf("Error while requesting API to deploy revision for Yandex Cloud Container: %s", err)
		}
	}
	d.Partial(false)

	return resourceYandexServerlessContainerRead(d, meta)
}

func resourceYandexServerlessContainerRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := containers.GetContainerRequest{ContainerId: d.Id()}
	container, err := config.sdk.Serverless().Containers().Container().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Container %q", d.Id()))
	}

	revision, err := resolveContainerLastRevision(ctx, config, d.Id())
	if err != nil {
		return fmt.Errorf("Failed to resolve last revision of Yandex Cloud Container: %s", err)
	}

	return flattenYandexServerlessContainer(d, container, revision)
}

func resolveContainerLastRevision(ctx context.Context, config *Config, containerID string) (*containers.Revision, error) {
	listRevisionsReq := &containers.ListContainersRevisionsRequest{
		Id:     &containers.ListContainersRevisionsRequest_ContainerId{ContainerId: containerID},
		Filter: fmt.Sprintf("status='%s'", containers.Revision_ACTIVE.String()),
	}
	resp, err := config.sdk.Serverless().Containers().Container().ListRevisions(ctx, listRevisionsReq)
	if err != nil {
		return nil, err
	}
	if len(resp.Revisions) == 0 {
		return nil, nil
	}
	return resp.Revisions[0], nil
}

func resourceYandexServerlessContainerDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := containers.DeleteContainerRequest{
		ContainerId: d.Id(),
	}

	op, err := config.sdk.Serverless().Containers().Container().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Container %q", d.Id()))
	}

	return nil
}

func expandLastRevision(d *schema.ResourceData) (*containers.DeployContainerRevisionRequest, error) {
	revisionReq := &containers.DeployContainerRevisionRequest{}

	revisionReq.Resources = &containers.Resources{Memory: int64(int(datasize.MB.Bytes()) * d.Get("memory").(int))}
	if v, ok := d.GetOk("cores"); ok {
		revisionReq.Resources.Cores = int64(v.(int))
	}
	if v, ok := d.GetOk("core_fraction"); ok {
		revisionReq.Resources.CoreFraction = int64(v.(int))
	}

	if v, ok := d.GetOk("execution_timeout"); ok {
		timeout, err := parseDuration(v.(string))
		if err != nil {
			return nil, fmt.Errorf("Cannot define execution_timeout for Yandex Cloud Container: %s", err)
		}
		revisionReq.ExecutionTimeout = timeout
	}

	if v, ok := d.GetOk("concurrency"); ok {
		revisionReq.Concurrency = int64(v.(int))
	}

	if v, ok := d.GetOk("service_account_id"); ok {
		revisionReq.ServiceAccountId = v.(string)
	}

	revisionReq.ImageSpec = &containers.ImageSpec{
		ImageUrl:   d.Get("image.0.url").(string),
		WorkingDir: d.Get("image.0.work_dir").(string),
	}
	if v, ok := d.GetOk("image.0.command"); ok {
		revisionReq.ImageSpec.Command = &containers.Command{
			Command: expandStringSlice(v.([]interface{})),
		}
	}
	if v, ok := d.GetOk("image.0.args"); ok {
		revisionReq.ImageSpec.Args = &containers.Args{
			Args: expandStringSlice(v.([]interface{})),
		}
	}
	if v, ok := d.GetOk("image.0.environment"); ok {
		env, err := expandLabels(v)
		if err != nil {
			return nil, fmt.Errorf("Cannot define image environment variables for Yandex Cloud Container: %s", err)
		}
		if len(env) != 0 {
			revisionReq.ImageSpec.Environment = env
		}
	}

	return revisionReq, nil
}

func flattenYandexServerlessContainer(d *schema.ResourceData, container *containers.Container, revision *containers.Revision) error {
	d.Set("name", container.Name)
	d.Set("folder_id", container.FolderId)
	d.Set("description", container.Description)
	d.Set("created_at", getTimestamp(container.CreatedAt))
	d.Set("url", container.Url)
	if err := d.Set("labels", container.Labels); err != nil {
		return err
	}

	if revision == nil {
		return nil
	}

	d.Set("revision_id", revision.Id)

	if revision.Resources != nil {
		d.Set("memory", int(revision.Resources.Memory/int64(datasize.MB.Bytes())))
		d.Set("cores", int(revision.Resources.Cores))
		d.Set("core_fraction", int(revision.Resources.CoreFraction))
	}
	if revision.ExecutionTimeout != nil {
		d.Set("execution_timeout", formatDuration(revision.ExecutionTimeout))
	}
	d.Set("concurrency", int(revision.Concurrency))
	d.Set("service_account_id", revision.ServiceAccountId)

	if revision.Image != nil {
		m := make(map[string]interface{})
		m["url"] = revision.Image.ImageUrl
		m["digest"] = revision.Image.ImageDigest
		m["work_dir"] = revision.Image.WorkingDir
		if revision.Image.Command != nil {
			m["command"] = revision.Image.Command.Command
		}
		if revision.Image.Args != nil {
			m["args"] = revision.Image.Args.Args
		}
		m["environment"] = revision.Image.Environment

		d.Set("image", []map[string]interface{}{m})
	}

	return nil
}
