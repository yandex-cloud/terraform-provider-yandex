package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexServerlessContainerDefaultTimeout = 5 * time.Minute

func resourceYandexServerlessContainer() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of a Yandex Cloud Serverless Container.",

		CreateContext: resourceYandexServerlessContainerCreate,
		ReadContext:   resourceYandexServerlessContainerRead,
		UpdateContext: resourceYandexServerlessContainerUpdate,
		DeleteContext: resourceYandexServerlessContainerDelete,
		CustomizeDiff: resourceYandexServerlessContainerCustomizeDiff,
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
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"memory": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Memory in megabytes (**aligned to 128 MB**).",
			},

			"cores": {
				Type:        schema.TypeInt,
				Description: "Cores (**1+**) of the Yandex Cloud Serverless Container.",
				Default:     1,
				Optional:    true,
			},

			"core_fraction": {
				Type:        schema.TypeInt,
				Description: "Core fraction (**0...100**) of the Yandex Cloud Serverless Container.",
				Computed:    true,
				Optional:    true,
			},

			"execution_timeout": {
				Type:             schema.TypeString,
				Description:      "Execution timeout in seconds (**duration format**) for Yandex Cloud Serverless Container.",
				Computed:         true,
				Optional:         true,
				ValidateFunc:     validateParsableValue(parsePositiveDuration),
				DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
			},

			"concurrency": {
				Type:        schema.TypeInt,
				Description: "Concurrency of Yandex Cloud Serverless Container.",
				Optional:    true,
				Computed:    true,
			},

			"service_account_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["service_account_id"],
				Optional:    true,
			},

			"secrets": {
				Type:        schema.TypeList,
				Description: "Secrets for Yandex Cloud Serverless Container.",
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "Secret's ID.",
							Required:    true,
						},
						"version_id": {
							Type:        schema.TypeString,
							Description: "Secret's version ID.",
							Required:    true,
						},
						"key": {
							Type:        schema.TypeString,
							Description: "Secret's entries key which value will be stored in environment variable.",
							Required:    true,
						},
						"environment_variable": {
							Type:        schema.TypeString,
							Description: "Container's environment variable in which secret's value will be stored. Must begin with a letter (A-Z, a-z).",
							Required:    true,
						},
					},
				},
			},
			"storage_mounts": {
				Type:        schema.TypeList,
				Description: "(**DEPRECATED**, use `mounts.object_storage` instead) Storage mounts for Yandex Cloud Serverless Container.",
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mount_point_path": {
							Type:        schema.TypeString,
							Description: "Path inside the container to access the directory in which the bucket is mounted.",
							Required:    true,
						},
						"bucket": {
							Type:        schema.TypeString,
							Description: "Name of the mounting bucket.",
							Required:    true,
						},
						"prefix": {
							Type:        schema.TypeString,
							Description: "Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted.",
							Optional:    true,
						},
						"read_only": {
							Type:        schema.TypeBool,
							Description: "Mount the bucket in read-only mode.",
							Optional:    true,
						},
					},
				},
				Deprecated: fieldDeprecatedForAnother("storage_mounts", "mounts"),
			},

			"mounts": {
				Type:        schema.TypeList,
				Description: "Mounts for Yandex Cloud Serverless Container.",
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mount_point_path": {
							Type:        schema.TypeString,
							Description: "Path inside the container to access the directory in which the target is mounted.",
							Required:    true,
						},
						"mode": {
							Type:         schema.TypeString,
							Description:  "Mount’s accessibility mode. Valid values are `ro` and `rw`.",
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"rw", "ro"}, true),
						},
						"ephemeral_disk": {
							Type:        schema.TypeList,
							Description: "One of the available mount types. Disk available during the function execution time.",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size_gb": {
										Type:        schema.TypeInt,
										Description: "Size of the ephemeral disk in GB.",
										Required:    true,
									},
									"block_size_kb": {
										Type:        schema.TypeInt,
										Description: "Block size of the ephemeral disk in KB.",
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
						"object_storage": {
							Type:        schema.TypeList,
							Description: "Available mount types. Object storage as a mount.",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
										Type:        schema.TypeString,
										Description: "Name of the mounting bucket.",
										Required:    true,
									},
									"prefix": {
										Type:        schema.TypeString,
										Description: "Prefix within the bucket. If you leave this field empty, the entire bucket will be mounted.",
										Optional:    true,
									},
								},
							},
						},
					},
				},
			},

			"image": {
				Type:        schema.TypeList,
				Description: "Revision deployment image for Yandex Cloud Serverless Container.",
				MaxItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Description: "URL of image that will be deployed as Yandex Cloud Serverless Container.",
							Required:    true,
						},
						"work_dir": {
							Type:        schema.TypeString,
							Description: "Working directory for Yandex Cloud Serverless Container.",
							Optional:    true,
						},
						"digest": {
							Type:        schema.TypeString,
							Description: "Digest of image that will be deployed as Yandex Cloud Serverless Container. If presented, should be equal to digest that will be resolved at server side by URL. Container will be updated on digest change even if `image.0.url` stays the same. If field not specified then its value will be computed.",
							Optional:    true,
							Computed:    true,
						},
						"command": {
							Type:        schema.TypeList,
							Description: "List of commands for Yandex Cloud Serverless Container.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"args": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of arguments for Yandex Cloud Serverless Container.",
							Optional:    true,
						},
						"environment": {
							Type:        schema.TypeMap,
							Description: "A set of key/value environment variable pairs for Yandex Cloud Serverless Container. Each key must begin with a letter (A-Z, a-z).",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},
					},
				},
			},

			"url": {
				Type:        schema.TypeString,
				Description: "Invoke URL for the Yandex Cloud Serverless Container.",
				Computed:    true,
			},

			"revision_id": {
				Type:        schema.TypeString,
				Description: "Last revision ID of the Yandex Cloud Serverless Container.",
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"runtime": {
				Type:        schema.TypeList,
				Description: "Runtime for Yandex Cloud Serverless Container.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Description:  "Type of the runtime for Yandex Cloud Serverless Container. Valid values are `http` and `task`.",
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"http", "task"}, true),
						},
					},
				},
				Required: false,
				Optional: true,
				Computed: true,
			},

			"connectivity": {
				Type:        schema.TypeList,
				Description: "Network access. If specified the revision will be attached to specified network.",
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:        schema.TypeString,
							Description: "Network the revision will have access to.",
							Required:    true,
						},
					},
				},
			},

			"provision_policy": {
				Type:        schema.TypeList,
				Description: "Provision policy. If specified the revision will have prepared instances.",
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_instances": {
							Type:        schema.TypeInt,
							Description: "Minimum number of prepared instances that are always ready to serve requests.",
							Required:    true,
						},
					},
				},
			},

			"log_options": {
				Type:        schema.TypeList,
				Description: "Options for logging from Yandex Cloud Serverless Container.",
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disabled": {
							Type:        schema.TypeBool,
							Description: "Is logging from container disabled.",
							Optional:    true,
						},
						"log_group_id": {
							Type:          schema.TypeString,
							Description:   "Log entries are written to specified log group.",
							Optional:      true,
							ConflictsWith: []string{"log_options.0.folder_id"},
						},
						"folder_id": {
							Type:          schema.TypeString,
							Description:   "Log entries are written to default log group for specified folder.",
							Optional:      true,
							ConflictsWith: []string{"log_options.0.log_group_id"},
						},
						"min_level": {
							Type:        schema.TypeString,
							Description: "Minimum log entry level.",
							Optional:    true,
						},
					},
				},
			},

			"metadata_options": {
				Type:        schema.TypeList,
				Description: "Options set the access mode to revision's metadata endpoints.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gce_http_endpoint": {
							Type:         schema.TypeInt,
							Description:  "Enables access to GCE flavored metadata. Values: `0`- default, `1` - enabled, `2` - disabled.",
							ValidateFunc: validation.IntBetween(0, 2),
							Optional:     true,
							Computed:     true,
						},
						"aws_v1_http_endpoint": {
							Type:         schema.TypeInt,
							Description:  "Enables access to AWS flavored metadata (IMDSv1). Values: `0` - default, `1` - enabled, `2` - disabled.",
							ValidateFunc: validation.IntBetween(0, 2),
							Optional:     true,
							Computed:     true,
						},
					},
				},
			},

			"async_invocation": {
				Type:        schema.TypeList,
				Description: "Config for asynchronous invocations of Yandex Cloud Serverless Container.",
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_account_id": {
							Type:        schema.TypeString,
							Description: "Service account used for async invocation.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexServerlessContainerCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
	if diff.HasChange("mounts") || diff.HasChange("storage_mounts") {
		mounts := diff.Get("mounts").([]interface{})
		storageMounts := diff.Get("storage_mounts").([]interface{})

		mergedMounts := mergeContainerMountsAndStorageMounts(mounts, storageMounts)

		err := diff.SetNew("mounts", mergedMounts)
		if err != nil {
			return err
		}
	}
	return nil
}

func mergeContainerMountsAndStorageMounts(mounts []interface{}, storageMounts []interface{}) interface{} {
	var (
		uniqueMounts = make(map[string]struct{})
		mergedMounts = make([]interface{}, 0)
	)

	for _, m := range mounts {
		mount := m.(map[string]interface{})

		uniqueMounts[mount["mount_point_path"].(string)] = struct{}{}
		mergedMounts = append(mergedMounts, mount)
	}

	for _, m := range storageMounts {
		storageMount := m.(map[string]interface{})

		if _, ok := uniqueMounts[storageMount["mount_point_path"].(string)]; !ok {
			uniqueMounts[storageMount["mount_point_path"].(string)] = struct{}{}
			mergedMounts = append(mergedMounts, containerStorageMountToMount(storageMount))
		}
	}

	return mergedMounts
}

func containerStorageMountToMount(storageMount map[string]interface{}) interface{} {
	var (
		mount  = make(map[string]interface{})
		bucket string
		prefix string
	)

	for k, v := range storageMount {
		switch k {
		case "mount_point_path":
			mount["mount_point_path"] = v.(string)
		case "read_only":
			mount["mode"] = mapBoolModeToString(v.(bool))
		case "bucket":
			bucket = v.(string)
		case "prefix":
			prefix = v.(string)
		}
	}

	mount["object_storage"] = []map[string]interface{}{
		{
			"bucket": bucket,
			"prefix": prefix,
		},
	}

	return mount
}

func resourceYandexServerlessContainerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(ctx), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while creating Yandex Cloud Container: %s", err)
	}

	revisionReq, err := expandLastRevision(d)
	if err != nil {
		return diag.FromErr(err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return diag.Errorf("Error getting folder ID while creating Yandex Cloud Container: %s", err)
	}

	req := containers.CreateContainerRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
	}
	op, err := config.sdk.WrapOperation(config.sdk.Serverless().Containers().Container().Create(ctx, &req))
	if err != nil {
		return diag.Errorf("Error while requesting API to create Yandex Cloud Container: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while requesting API to create Yandex Cloud Container: %s", err)
	}
	md, ok := protoMetadata.(*containers.CreateContainerMetadata)
	if !ok {
		return diag.Errorf("Could not get Yandex Cloud Container ID from create operation metadata")
	}
	d.SetId(md.ContainerId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("Error while requesting API to create Yandex Cloud Container: %s", err)
	}

	var diags diag.Diagnostics
	if revisionReq != nil {
		revisionReq.ContainerId = md.ContainerId
		diags = resourceYandexServerlessContainerDiagsFromDeployRevisionError(
			resourceYandexServerlessContainerDeployRevision(ctx, config.sdk, revisionReq),
		)
	}

	return append(diags, resourceYandexServerlessContainerRead(ctx, d, meta)...)
}

func resourceYandexServerlessContainerDiagsFromDeployRevisionError(err error) diag.Diagnostics {
	if err == nil {
		return nil
	}
	return diag.Diagnostics{diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Failed to deploy revision for Yandex Cloud Container",
		Detail: "Error while requesting API to deploy revision for Yandex Cloud Container. " +
			"After resolving following issues apply resource again to deploy revision for Yandex Cloud Container:\n" +
			err.Error(),
	}}
}

func resourceYandexServerlessContainerDeployRevision(
	ctx context.Context,
	sdk *ycsdk.SDK,
	req *containers.DeployContainerRevisionRequest,
) error {
	op, err := sdk.WrapOperation(sdk.Serverless().Containers().Container().DeployRevision(ctx, req))
	if err != nil {
		return err
	}
	return op.Wait(ctx)
}

func resourceYandexServerlessContainerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(ctx), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while updating Yandex Cloud Container: %s", err)
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

	lastRevisionPaths := []string{
		"memory", "cores", "core_fraction", "execution_timeout", "service_account_id",
		"secrets", "image", "concurrency", "connectivity", "storage_mounts", "mounts", "log_options", "provision_policy",
		"runtime", "metadata_options", "async_invocation",
	}
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
			return diag.FromErr(err)
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
			return diag.Errorf("Error while requesting API to update Yandex Cloud Container: %s", err)
		}
	}

	var diags diag.Diagnostics
	if revisionReq != nil {
		revisionReq.ContainerId = d.Id()
		diags = resourceYandexServerlessContainerDiagsFromDeployRevisionError(
			resourceYandexServerlessContainerDeployRevision(ctx, config.sdk, revisionReq),
		)
	}
	d.Partial(false)

	return append(diags, resourceYandexServerlessContainerRead(ctx, d, meta)...)
}

func resourceYandexServerlessContainerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(ctx), d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := containers.GetContainerRequest{ContainerId: d.Id()}
	container, err := config.sdk.Serverless().Containers().Container().Get(ctx, &req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Container %q", d.Id())))
	}

	revision, err := resolveContainerLastRevision(ctx, config, d.Id())
	if err != nil {
		return diag.Errorf("Failed to resolve last revision of Yandex Cloud Container: %s", err)
	}

	return diag.FromErr(flattenYandexServerlessContainer(d, container, revision, false))
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

func resourceYandexServerlessContainerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(ctx), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := containers.DeleteContainerRequest{
		ContainerId: d.Id(),
	}

	op, err := config.sdk.Serverless().Containers().Container().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Container %q", d.Id())))
	}

	return nil
}

func expandLastRevision(d *schema.ResourceData) (*containers.DeployContainerRevisionRequest, error) {
	revisionReq := &containers.DeployContainerRevisionRequest{}

	revisionReq.Resources = &containers.Resources{Memory: int64(datasize.MB.Bytes()) * int64(d.Get("memory").(int))}
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

	if v, ok := d.GetOk("provision_policy.0.min_instances"); ok {
		revisionReq.ProvisionPolicy = &containers.ProvisionPolicy{
			MinInstances: int64(v.(int)),
		}
	}

	if v, ok := d.GetOk("service_account_id"); ok {
		revisionReq.ServiceAccountId = v.(string)
	}

	if v, ok := d.GetOk("secrets"); ok {
		secretsList := v.([]interface{})

		revisionReq.Secrets = make([]*containers.Secret, len(secretsList))
		for i, s := range secretsList {
			secret := s.(map[string]interface{})

			fs := &containers.Secret{}
			if ID, ok := secret["id"]; ok {
				fs.Id = ID.(string)
			}
			if versionID, ok := secret["version_id"]; ok {
				fs.VersionId = versionID.(string)
			}
			if key, ok := secret["key"]; ok {
				fs.Key = key.(string)
			}
			if environmentVariable, ok := secret["environment_variable"]; ok {
				fs.Reference = &containers.Secret_EnvironmentVariable{EnvironmentVariable: environmentVariable.(string)}
			}

			revisionReq.Secrets[i] = fs
		}
	}

	if v, ok := d.GetOk("mounts"); ok {
		mountsList := v.([]interface{})
		revisionReq.Mounts = make([]*containers.Mount, len(mountsList))

		for i, m := range mountsList {
			mount := m.(map[string]interface{})

			cm := &containers.Mount{}
			if name, ok := mount["mount_point_path"].(string); ok {
				cm.MountPointPath = name
			}
			if mode, ok := mount["mode"].(string); ok {
				cm.Mode = mapContainerModeFromTF(mode)
			} else {
				cm.Mode = containers.Mount_MODE_UNSPECIFIED
			}

			if ephemeralDiskList, ok := mount["ephemeral_disk"].([]interface{}); ok && len(ephemeralDiskList) > 0 {
				var (
					ephemeralDisk = ephemeralDiskList[0].(map[string]interface{})
					diskSpec      containers.Mount_DiskSpec
				)

				if gbValue, ok := ephemeralDisk["size_gb"].(int); ok {
					diskSpec.Size = toBytes(gbValue)
				}
				if gbValue, ok := ephemeralDisk["block_size_kb"].(int); ok {
					diskSpec.BlockSize = kilobytesToBytes(gbValue)
				}

				cm.Target = &containers.Mount_EphemeralDiskSpec{EphemeralDiskSpec: &diskSpec}
			}

			if objectStorageList, ok := mount["object_storage"].([]interface{}); ok && len(objectStorageList) > 0 {
				var (
					objectStorage     = objectStorageList[0].(map[string]interface{})
					objectStorageSpec containers.Mount_ObjectStorage
				)

				if bucket, ok := objectStorage["bucket"].(string); ok {
					objectStorageSpec.BucketId = bucket
				}
				if prefix, ok := objectStorage["prefix"].(string); ok {
					objectStorageSpec.Prefix = prefix
				}

				cm.Target = &containers.Mount_ObjectStorage_{ObjectStorage: &objectStorageSpec}
			}

			revisionReq.Mounts[i] = cm
		}
	}

	paths := make([]string, len(revisionReq.Mounts))
	for i, v := range revisionReq.Mounts {
		paths[i] = v.MountPointPath
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
	if connectivity := expandServerlessContainerConnectivity(d); connectivity != nil {
		revisionReq.Connectivity = connectivity
	}

	{
		logOptions, err := expandServerlessContainerLogOptions(d)
		if err != nil {
			return nil, err
		}
		revisionReq.LogOptions = logOptions
	}

	if v, ok := d.GetOk("runtime.0"); ok {
		revisionReq.Runtime = expandServerlessContainerRuntime(v)
	}

	revisionReq.MetadataOptions = expandServerlessContainerMetadataOptions(d)

	if v, ok := d.GetOk("async_invocation.0"); ok {
		revisionReq.AsyncInvocationConfig = expandServerlessContainerAsyncInvocationConfig(v)
	}

	return revisionReq, nil
}

func expandServerlessContainerMetadataOptions(d *schema.ResourceData) *containers.MetadataOptions {
	metadataOptions := containers.MetadataOptions{}
	if v, ok := d.GetOk("metadata_options.0.gce_http_endpoint"); ok {
		metadataOptions.GceHttpEndpoint = containers.MetadataOption(v.(int))
	}
	if v, ok := d.GetOk("metadata_options.0.aws_v1_http_endpoint"); ok {
		metadataOptions.AwsV1HttpEndpoint = containers.MetadataOption(v.(int))
	}

	return &metadataOptions
}

func expandServerlessContainerRuntime(v interface{}) *containers.Runtime {
	var (
		runtimeMap = v.(map[string]interface{})
		t          = runtimeMap["type"].(string)
	)

	switch t {
	case "http":
		return &containers.Runtime{Type: &containers.Runtime_Http_{Http: &containers.Runtime_Http{}}}
	case "task":
		return &containers.Runtime{Type: &containers.Runtime_Task_{Task: &containers.Runtime_Task{}}}
	default:
		// should never happen
		panic("unknown runtime type: " + t)
	}
}

func expandServerlessContainerAsyncInvocationConfig(v interface{}) *containers.AsyncInvocationConfig {
	asyncConfig := v.(map[string]interface{})
	config := &containers.AsyncInvocationConfig{}

	if saID, ok := asyncConfig["service_account_id"]; ok {
		config.ServiceAccountId = saID.(string)
	}

	return config
}

func mapContainerModeFromTF(mode string) containers.Mount_Mode {
	if mode == "rw" {
		return containers.Mount_READ_WRITE
	} else if mode == "ro" {
		return containers.Mount_READ_ONLY
	} else {
		// Shouldn't have happened due to validation
		panic("unknown mode: " + mode)
	}
}

func mapContainerModeFromPB(mode containers.Mount_Mode) string {
	switch mode {
	case containers.Mount_READ_ONLY:
		return "ro"
	case containers.Mount_READ_WRITE:
		return "rw"
	default:
		panic("unknown mode: " + mode.String())
	}
}

func flattenYandexServerlessContainer(
	d *schema.ResourceData,
	container *containers.Container,
	revision *containers.Revision,
	allFields bool,
) error {
	d.Set("name", container.Name)
	d.Set("folder_id", container.FolderId)
	d.Set("description", container.Description)
	d.Set("created_at", getTimestamp(container.CreatedAt))
	d.Set("url", container.Url)
	d.Set("labels", container.Labels)

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
	d.Set("secrets", flattenRevisionSecrets(revision.Secrets))
	d.Set("mounts", flattenRevisionMounts(revision.Mounts))

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
	if connectivity := flattenServerlessContainerConnectivity(revision.Connectivity); connectivity != nil {
		d.Set("connectivity", connectivity)
	}
	d.Set("log_options", flattenServerlessContainerLogOptions(d, revision.LogOptions, container.FolderId, allFields))

	if revision.ProvisionPolicy != nil {
		d.Set("provision_policy", []map[string]interface{}{
			{
				"min_instances": revision.ProvisionPolicy.MinInstances,
			},
		})
	}
	if revision.GetRuntime() != nil {
		d.Set("runtime", flattenServerlessContainerRuntime(revision.GetRuntime()))
	}

	d.Set("metadata_options", flattenServerlessContainerMetadataOptions(revision))

	if asyncConfig := flattenServerlessContainerRevisionAsyncInvocationConfig(revision.AsyncInvocationConfig); asyncConfig != nil {
		d.Set("async_invocation", asyncConfig)
	}

	return nil
}

func flattenServerlessContainerMetadataOptions(revision *containers.Revision) []map[string]interface{} {
	var gceHttpEndpoint, awsV1HttpEndpoint int
	if revision.MetadataOptions != nil {
		gceHttpEndpoint = int(revision.MetadataOptions.GceHttpEndpoint)
		awsV1HttpEndpoint = int(revision.MetadataOptions.AwsV1HttpEndpoint)
	}

	metadataOptions := map[string]interface{}{
		"gce_http_endpoint":    gceHttpEndpoint,
		"aws_v1_http_endpoint": awsV1HttpEndpoint,
	}

	return []map[string]interface{}{metadataOptions}
}

func flattenServerlessContainerRuntime(runtime *containers.Runtime) interface{} {
	runtimeMap := make(map[string]interface{})

	switch t := runtime.Type.(type) {
	case *containers.Runtime_Http_:
		runtimeMap["type"] = "http"
	case *containers.Runtime_Task_:
		runtimeMap["type"] = "task"
	default:
		panic(fmt.Sprintf("unknown runtime type: %T", t))
	}

	return []interface{}{runtimeMap}
}

func flattenRevisionMounts(mounts []*containers.Mount) interface{} {
	s := make([]map[string]interface{}, len(mounts))

	for i, mount := range mounts {
		s[i] = map[string]interface{}{
			"mount_point_path": mount.MountPointPath,
		}

		if mount.Mode != containers.Mount_MODE_UNSPECIFIED {
			s[i]["mode"] = mapContainerModeFromPB(mount.Mode)
		}

		if mount.GetEphemeralDiskSpec() != nil {
			s[i]["ephemeral_disk"] = []map[string]interface{}{
				{
					"size_gb":       toGigabytes(mount.GetEphemeralDiskSpec().Size),
					"block_size_kb": bytesToKilobytes(mount.GetEphemeralDiskSpec().BlockSize),
				},
			}
		}

		if mount.GetObjectStorage() != nil {
			s[i]["object_storage"] = []map[string]interface{}{
				{
					"bucket": mount.GetObjectStorage().BucketId,
					"prefix": mount.GetObjectStorage().Prefix,
				},
			}
		}
	}

	return s
}

func flattenRevisionSecrets(secrets []*containers.Secret) []map[string]interface{} {
	s := make([]map[string]interface{}, len(secrets))

	for i, secret := range secrets {
		s[i] = map[string]interface{}{
			"id":                   secret.Id,
			"version_id":           secret.VersionId,
			"key":                  secret.Key,
			"environment_variable": secret.GetEnvironmentVariable(),
		}
	}
	return s
}

func flattenRevisionStorageMounts(storageMounts []*containers.StorageMount) []map[string]interface{} {
	s := make([]map[string]interface{}, len(storageMounts))

	for i, storageMount := range storageMounts {
		s[i] = map[string]interface{}{
			"mount_point_path": storageMount.MountPointPath,
			"bucket":           storageMount.BucketId,
			"prefix":           storageMount.Prefix,
			"read_only":        storageMount.ReadOnly,
		}
	}
	return s
}

func expandServerlessContainerConnectivity(d *schema.ResourceData) *containers.Connectivity {
	if id, ok := d.GetOk("connectivity.0.network_id"); ok {
		return &containers.Connectivity{NetworkId: id.(string)}
	}
	return nil
}

func flattenServerlessContainerConnectivity(connectivity *containers.Connectivity) []interface{} {
	if connectivity == nil || connectivity.NetworkId == "" {
		return nil
	}
	return []interface{}{map[string]interface{}{"network_id": connectivity.NetworkId}}
}

func expandServerlessContainerLogOptions(d *schema.ResourceData) (*containers.LogOptions, error) {
	v, ok := d.GetOk("log_options.0")
	if !ok {
		return nil, nil
	}
	logOptionsMap := v.(map[string]interface{})
	if logOptionsMap["disabled"].(bool) {
		return &containers.LogOptions{
			Disabled: true,
		}, nil
	}
	logOptions := &containers.LogOptions{}
	if folderID, ok := logOptionsMap["folder_id"]; ok {
		logOptions.SetFolderId(folderID.(string))
	}
	if logGroupID, ok := logOptionsMap["log_group_id"]; ok {
		logOptions.SetLogGroupId(logGroupID.(string))
	}
	if level := logOptionsMap["min_level"]; len(level.(string)) > 0 {
		logLevel, ok := logging.LogLevel_Level_value[level.(string)]
		if !ok {
			return nil, fmt.Errorf("unknown log level: %s", level)
		}
		logOptions.MinLevel = logging.LogLevel_Level(logLevel)
	}
	return logOptions, nil
}

func flattenServerlessContainerLogOptions(
	d *schema.ResourceData,
	logOptions *containers.LogOptions,
	containerFolderID string,
	allFields bool,
) []interface{} {
	if logOptions == nil {
		return nil
	}
	res := make(map[string]interface{})
	if !allFields && logOptions.Disabled {
		res["disabled"] = true
		return []interface{}{res}
	}
	if allFields || len(d.Get("log_options.0.min_level").(string)) > 0 || logOptions.MinLevel != 0 {
		res["min_level"] = logging.LogLevel_Level_name[int32(logOptions.MinLevel)]
	}
	if logOptions.Destination != nil {
		switch destination := logOptions.Destination.(type) {
		case *containers.LogOptions_LogGroupId:
			res["log_group_id"] = destination.LogGroupId
		case *containers.LogOptions_FolderId:
			if allFields ||
				len(d.Get("log_options.0.folder_id").(string)) > 0 ||
				destination.FolderId != containerFolderID {

				res["folder_id"] = destination.FolderId
			}
		}
	}
	if !allFields && len(d.Get("log_options").([]interface{})) <= 0 && len(res) <= 0 {
		return nil
	}
	res["disabled"] = logOptions.Disabled
	return []interface{}{res}
}

func flattenServerlessContainerRevisionAsyncInvocationConfig(config *containers.AsyncInvocationConfig) []interface{} {
	if config == nil {
		return nil
	}
	res := map[string]interface{}{}
	if config.ServiceAccountId != "" {
		res["service_account_id"] = config.ServiceAccountId
	}
	return []interface{}{res}
}
