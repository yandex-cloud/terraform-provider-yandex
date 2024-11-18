package yandex

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
)

const yandexFunctionDefaultTimeout = 10 * time.Minute
const versionCreateSourceContentMaxBytes = 3670016

func resourceYandexFunction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexFunctionCreate,
		ReadContext:   resourceYandexFunctionRead,
		UpdateContext: resourceYandexFunctionUpdate,
		DeleteContext: resourceYandexFunctionDelete,
		CustomizeDiff: resourceYandexFunctionCustomizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexFunctionDefaultTimeout),
			Update: schema.DefaultTimeout(yandexFunctionDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexFunctionDefaultTimeout),
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceYandexFunctionV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceYandexFunctionStateUpgradeV0,
				Version: 0,
			},
		},

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
				Computed: true,
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

			"storage_mounts": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mount_point_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"read_only": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				Deprecated: fieldDeprecatedForAnother("storage_mounts", "mounts"),
			},

			"mounts": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"rw", "ro"}, true),
						},
						"ephemeral_disk": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size_gb": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"block_size_kb": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"object_storage": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
										Type:     schema.TypeString,
										Required: true,
									},
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
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

			"async_invocation": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retries_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"service_account_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ymq_success_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"arn": {
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
						"ymq_failure_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"arn": {
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
					},
				},
			},

			"log_options": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"log_group_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"log_options.0.folder_id"},
						},
						"folder_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"log_options.0.log_group_id"},
						},
						"min_level": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"tmpfs_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"concurrency": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceYandexFunctionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(ctx), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while creating Yandex Cloud Function: %s", err)
	}

	versionReq, err := expandLastVersion(d)
	if err != nil {
		return diag.FromErr(err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return diag.Errorf("Error getting folder ID while creating Yandex Cloud Function: %s", err)
	}

	req := functions.CreateFunctionRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
	}

	op, err := config.sdk.WrapOperation(config.sdk.Serverless().Functions().Function().Create(ctx, &req))
	if err != nil {
		return diag.Errorf("Error while requesting API to create Yandex Cloud Function: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while requesting API to create Yandex Cloud Function: %s", err)
	}

	md, ok := protoMetadata.(*functions.CreateFunctionMetadata)
	if !ok {
		return diag.Errorf("Could not get Yandex Cloud Function ID from create operation metadata")
	}

	d.SetId(md.FunctionId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("Error while requesting API to create Yandex Cloud Function: %s", err)
	}

	var diags diag.Diagnostics
	if versionReq != nil {
		versionReq.FunctionId = md.FunctionId
		diags = resourceYandexFunctionDiagsFromCreateVersionError(
			resourceYandexFunctionCreateVersion(ctx, config.sdk, versionReq),
		)
	}

	return append(diags, resourceYandexFunctionRead(ctx, d, meta)...)
}

func resourceYandexFunctionDiagsFromCreateVersionError(err error) diag.Diagnostics {
	if err == nil {
		return nil
	}
	return diag.Diagnostics{diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Failed to create version for Yandex Cloud Function",
		Detail: "Error while requesting API to create version for Yandex Cloud Function. " +
			"After resolving following issues apply resource again to create version for Yandex Cloud Function:\n" +
			err.Error(),
	}}
}

func resourceYandexFunctionCreateVersion(
	ctx context.Context,
	sdk *ycsdk.SDK,
	req *functions.CreateFunctionVersionRequest,
) error {
	op, err := sdk.WrapOperation(sdk.Serverless().Functions().Function().CreateVersion(ctx, req))
	if err != nil {
		return err
	}
	return op.Wait(ctx)
}

func resourceYandexFunctionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(ctx), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while updating Yandex Cloud Function: %s", err)
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
		"environment", "tags", "package", "content", "secrets", "connectivity", "async_invocation",
		"storage_mounts", "mounts", "log_options", "tmpfs_size", "concurrency",
	}
	var versionPartialPaths []string
	for _, p := range lastVersionPaths {
		if d.HasChange(p) {
			versionPartialPaths = append(versionPartialPaths, p)
		}
	}

	var versionReq *functions.CreateFunctionVersionRequest
	if len(versionPartialPaths) != 0 {
		versionReq, err = expandLastVersion(d)
		if err != nil {
			return diag.FromErr(err)
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
			return diag.Errorf("Error while requesting API to update Yandex Cloud Function: %s", err)
		}

	}

	var diags diag.Diagnostics
	if versionReq != nil {
		versionReq.FunctionId = d.Id()
		diags = resourceYandexFunctionDiagsFromCreateVersionError(
			resourceYandexFunctionCreateVersion(ctx, config.sdk, versionReq),
		)
	}
	d.Partial(false)

	return append(diags, resourceYandexFunctionRead(ctx, d, meta)...)
}

func resourceYandexFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(ctx), d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := functions.GetFunctionRequest{
		FunctionId: d.Id(),
	}

	function, err := config.sdk.Serverless().Functions().Function().Get(ctx, &req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Function %q", d.Id())))
	}

	version, err := resolveFunctionLatestVersion(ctx, config, function.GetId())
	if err != nil {
		return diag.Errorf("Failed to get latest version of Yandex Function: %s", err)
	}

	return diag.FromErr(flattenYandexFunction(d, function, version, false))
}

func resolveFunctionLatestVersion(ctx context.Context, config *Config, functionID string) (*functions.Version, error) {
	versionReq := functions.GetFunctionVersionByTagRequest{
		FunctionId: functionID,
		Tag:        "$latest",
	}
	version, err := config.sdk.Serverless().Functions().Function().GetVersionByTag(ctx, &versionReq)
	if err != nil {
		if !isStatusWithCode(err, codes.NotFound) {
			return nil, err
		}
		return nil, nil
	}
	return version, nil
}

func resourceYandexFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(ctx), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := functions.DeleteFunctionRequest{
		FunctionId: d.Id(),
	}

	op, err := config.sdk.Serverless().Functions().Function().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Function %q", d.Id())))
	}

	return nil
}

func resourceYandexFunctionCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	if diff.HasChange("mounts") || diff.HasChange("storage_mounts") {
		mounts := diff.Get("mounts").([]interface{})
		storageMounts := diff.Get("storage_mounts").([]interface{})

		mergedMounts := mergeFunctionMountsAndStorageMounts(mounts, storageMounts)

		err := diff.SetNew("mounts", mergedMounts)
		if err != nil {
			return err
		}
	}
	return nil
}

func mergeFunctionMountsAndStorageMounts(mounts []interface{}, storageMounts []interface{}) interface{} {
	var (
		uniqueMounts = make(map[string]struct{})
		mergedMounts = make([]interface{}, 0)
	)

	for _, m := range mounts {
		mount := m.(map[string]interface{})

		uniqueMounts[mount["name"].(string)] = struct{}{}
		mergedMounts = append(mergedMounts, mount)
	}

	for _, m := range storageMounts {
		storageMount := m.(map[string]interface{})

		if _, ok := uniqueMounts[storageMount["mount_point_name"].(string)]; !ok {
			uniqueMounts[storageMount["mount_point_name"].(string)] = struct{}{}
			mergedMounts = append(mergedMounts, functionStorageMountToMount(storageMount))
		}
	}

	return mergedMounts
}

func functionStorageMountToMount(storageMount map[string]interface{}) interface{} {
	var (
		mount  = make(map[string]interface{})
		bucket string
		prefix string
	)

	for k, v := range storageMount {
		switch k {
		case "mount_point_name":
			mount["name"] = v.(string)
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

func expandLastVersion(d *schema.ResourceData) (*functions.CreateFunctionVersionRequest, error) {
	versionReq := &functions.CreateFunctionVersionRequest{}
	versionReq.Runtime = d.Get("runtime").(string)
	versionReq.Entrypoint = d.Get("entrypoint").(string)

	versionReq.Resources = &functions.Resources{Memory: int64(datasize.MB.Bytes()) * int64(d.Get("memory").(int))}
	if v, ok := d.GetOk("execution_timeout"); ok {
		i, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Cannot define execution_timeout for Yandex Cloud Function: %s", err)
		}
		versionReq.ExecutionTimeout = &duration.Duration{Seconds: i}
	}
	if v, ok := d.GetOk("service_account_id"); ok {
		versionReq.ServiceAccountId = v.(string)
	}
	if v, ok := d.GetOk("environment"); ok {
		env, err := expandLabels(v)
		if err != nil {
			return nil, fmt.Errorf("Cannot define environment variables for Yandex Cloud Function: %s", err)
		}
		if len(env) != 0 {
			versionReq.Environment = env
		}
	}
	if v, ok := d.GetOk("tags"); ok {
		set := v.(*schema.Set)
		for _, t := range set.List() {
			v := t.(string)
			versionReq.Tag = append(versionReq.Tag, v)
		}
	}
	if _, ok := d.GetOk("package"); ok {
		pkg := &functions.Package{
			BucketName: d.Get("package.0.bucket_name").(string),
			ObjectName: d.Get("package.0.object_name").(string),
		}
		if v, ok := d.GetOk("package.0.sha_256"); ok {
			pkg.Sha256 = v.(string)
		}
		versionReq.PackageSource = &functions.CreateFunctionVersionRequest_Package{Package: pkg}
	} else if _, ok := d.GetOk("content"); ok {
		content, err := ZipPathToBytes(d.Get("content.0.zip_filename").(string))
		if err != nil {
			return nil, fmt.Errorf("Cannot define content for Yandex Cloud Function: %s", err)
		}
		if size := len(content); size > versionCreateSourceContentMaxBytes {
			return nil, fmt.Errorf("Zip archive content size %v exceeds the maximum size %v, use object storage to upload the content", size, versionCreateSourceContentMaxBytes)
		}
		versionReq.PackageSource = &functions.CreateFunctionVersionRequest_Content{Content: content}
	} else {
		return nil, fmt.Errorf("Package or content option must be present for Yandex Cloud Function")
	}
	if v, ok := d.GetOk("secrets"); ok {
		secretsList := v.([]interface{})

		versionReq.Secrets = make([]*functions.Secret, len(secretsList))
		for i, s := range secretsList {
			secret := s.(map[string]interface{})

			fs := &functions.Secret{}
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
				fs.Reference = &functions.Secret_EnvironmentVariable{EnvironmentVariable: environmentVariable.(string)}
			}

			versionReq.Secrets[i] = fs
		}
	}

	if v, ok := d.GetOk("mounts"); ok {
		mountsList := v.([]interface{})
		versionReq.Mounts = make([]*functions.Mount, len(mountsList))

		for i, m := range mountsList {
			mount := m.(map[string]interface{})

			fm := &functions.Mount{}
			if name, ok := mount["name"].(string); ok {
				fm.Name = name
			}
			if mode, ok := mount["mode"].(string); ok {
				fm.Mode = mapFunctionModeFromTF(mode)
			} else {
				fm.Mode = functions.Mount_MODE_UNSPECIFIED
			}

			if ephemeralDiskList, ok := mount["ephemeral_disk"].([]interface{}); ok && len(ephemeralDiskList) > 0 {
				var (
					ephemeralDisk = ephemeralDiskList[0].(map[string]interface{})
					diskSpec      functions.Mount_DiskSpec
				)

				if gbValue, ok := ephemeralDisk["size_gb"].(int); ok {
					diskSpec.Size = toBytes(gbValue)
				}
				if gbValue, ok := ephemeralDisk["block_size_kb"].(int); ok {
					diskSpec.BlockSize = kilobytesToBytes(gbValue)
				}

				fm.Target = &functions.Mount_EphemeralDiskSpec{EphemeralDiskSpec: &diskSpec}
			}

			if objectStorageList, ok := mount["object_storage"].([]interface{}); ok && len(objectStorageList) > 0 {
				var (
					objectStorage     = objectStorageList[0].(map[string]interface{})
					objectStorageSpec functions.Mount_ObjectStorage
				)

				if bucket, ok := objectStorage["bucket"].(string); ok {
					objectStorageSpec.BucketId = bucket
				}
				if prefix, ok := objectStorage["prefix"].(string); ok {
					objectStorageSpec.Prefix = prefix
				}

				fm.Target = &functions.Mount_ObjectStorage_{ObjectStorage: &objectStorageSpec}
			}

			versionReq.Mounts[i] = fm
		}
	}

	if connectivity := expandFunctionConnectivity(d); connectivity != nil {
		versionReq.Connectivity = connectivity
	}
	if v, ok := d.GetOk("async_invocation.0"); ok {
		asyncConfig := v.(map[string]interface{})
		config := &functions.AsyncInvocationConfig{}

		if maxRetries, ok := asyncConfig["retries_count"]; ok {
			config.RetriesCount = int64(maxRetries.(int))
		}
		if saID, ok := asyncConfig["service_account_id"]; ok {
			config.ServiceAccountId = saID.(string)
		}
		config.SuccessTarget = expandFunctionAsyncYMQTarget(d, "ymq_success_target")
		config.FailureTarget = expandFunctionAsyncYMQTarget(d, "ymq_failure_target")
		versionReq.AsyncInvocationConfig = config
	}

	{
		logOptions, err := expandFunctionLogOptions(d)
		if err != nil {
			return nil, err
		}
		versionReq.LogOptions = logOptions
	}

	versionReq.TmpfsSize = 0
	if v, ok := d.GetOk("tmpfs_size"); ok {
		versionReq.TmpfsSize = int64(int(datasize.MB.Bytes()) * v.(int))
	}

	if v, ok := d.GetOk("concurrency"); ok {
		versionReq.Concurrency = int64(v.(int))
	}

	return versionReq, nil
}

func mapFunctionModeFromTF(mode string) functions.Mount_Mode {
	if mode == "rw" {
		return functions.Mount_READ_WRITE
	} else if mode == "ro" {
		return functions.Mount_READ_ONLY
	} else {
		// Shouldn't have happened due to validation
		panic("unknown mode: " + mode)
	}
}

func mapFunctionModeFromPB(mode functions.Mount_Mode) string {
	switch mode {
	case functions.Mount_READ_ONLY:
		return "ro"
	case functions.Mount_READ_WRITE:
		return "rw"
	default:
		panic("unknown mode: " + mode.String())
	}
}

func mapBoolModeToString(isReadOnly bool) string {
	if isReadOnly {
		return "ro"
	}
	return "rw"
}

func flattenYandexFunction(
	d *schema.ResourceData,
	function *functions.Function,
	version *functions.Version,
	allFields bool,
) error {
	d.Set("name", function.Name)
	d.Set("folder_id", function.FolderId)
	d.Set("description", function.Description)
	d.Set("created_at", getTimestamp(function.CreatedAt))
	d.Set("labels", function.Labels)

	if version == nil {
		return nil
	}

	d.Set("version", version.Id)
	d.Set("image_size", version.ImageSize)
	d.Set("runtime", version.Runtime)
	d.Set("entrypoint", version.Entrypoint)
	d.Set("service_account_id", version.ServiceAccountId)
	d.Set("environment", version.Environment)

	if version.Resources != nil {
		d.Set("memory", int(version.Resources.Memory/int64(datasize.MB.Bytes())))
	}
	if version.ExecutionTimeout != nil && version.ExecutionTimeout.Seconds != 0 {
		d.Set("execution_timeout", strconv.FormatInt(version.ExecutionTimeout.Seconds, 10))
	}
	if connectivity := flattenFunctionConnectivity(version.Connectivity); connectivity != nil {
		d.Set("connectivity", connectivity)
	}
	if asyncConfig := flattenFunctionAsyncConfig(version.AsyncInvocationConfig); asyncConfig != nil {
		d.Set("async_invocation", asyncConfig)
	}
	d.Set("log_options", flattenFunctionLogOptions(d, version.LogOptions, function.FolderId, allFields))

	tags := &schema.Set{F: schema.HashString}
	for _, v := range version.Tags {
		if v != "$latest" {
			tags.Add(v)
		}
	}
	d.Set("tags", tags)

	d.Set("secrets", flattenFunctionSecrets(version.Secrets))
	d.Set("mounts", flattenVersionMounts(version.Mounts))
	d.Set("tmpfs_size", int(version.TmpfsSize/int64(datasize.MB.Bytes())))
	d.Set("concurrency", version.Concurrency)

	return nil
}

func zipPathToWriter(root string, buffer io.Writer) error {
	rootDir := filepath.Dir(root)
	zipWriter := zip.NewWriter(buffer)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel := strings.TrimPrefix(path, rootDir)
		entry, err := zipWriter.Create(rel)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(entry, file); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = zipWriter.Close()
	if err != nil {
		return err
	}
	return nil
}

func ZipPathToBytes(root string) ([]byte, error) {

	// first, check if the path corresponds to already zipped file
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if info.Mode().IsRegular() {
		bytes, err := ioutil.ReadFile(root)
		if err != nil {
			return nil, err
		}
		if isZipContent(bytes) {
			// file has already zipped, return its content
			return bytes, nil
		}
	}

	// correct path (make directory looks like a directory)
	if info.Mode().IsDir() && !strings.HasSuffix(root, string(os.PathSeparator)) {
		root = root + "/"
	}

	// do real zipping of the given path
	var buffer bytes.Buffer
	err = zipPathToWriter(root, &buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func isZipContent(buf []byte) bool {
	return len(buf) > 3 &&
		buf[0] == 0x50 && buf[1] == 0x4B &&
		(buf[2] == 0x3 || buf[2] == 0x5 || buf[2] == 0x7) &&
		(buf[3] == 0x4 || buf[3] == 0x6 || buf[3] == 0x8)
}

func flattenFunctionSecrets(secrets []*functions.Secret) []map[string]interface{} {
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

func flattenVersionStorageMounts(storageMounts []*functions.StorageMount) []map[string]interface{} {
	s := make([]map[string]interface{}, len(storageMounts))

	for i, storageMount := range storageMounts {
		s[i] = map[string]interface{}{
			"mount_point_name": storageMount.MountPointName,
			"bucket":           storageMount.BucketId,
			"prefix":           storageMount.Prefix,
			"read_only":        storageMount.ReadOnly,
		}
	}
	return s
}

func flattenVersionMounts(mounts []*functions.Mount) []map[string]interface{} {
	s := make([]map[string]interface{}, len(mounts))

	for i, mount := range mounts {
		s[i] = map[string]interface{}{
			"name": mount.Name,
		}

		if mount.Mode != functions.Mount_MODE_UNSPECIFIED {
			s[i]["mode"] = mapFunctionModeFromPB(mount.Mode)
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

func expandFunctionConnectivity(d *schema.ResourceData) *functions.Connectivity {
	if id, ok := d.GetOk("connectivity.0.network_id"); ok {
		return &functions.Connectivity{NetworkId: id.(string)}
	}
	return nil
}

func flattenFunctionConnectivity(connectivity *functions.Connectivity) []interface{} {
	if connectivity == nil || connectivity.NetworkId == "" {
		return nil
	}
	return []interface{}{map[string]interface{}{"network_id": connectivity.NetworkId}}
}

func expandFunctionAsyncYMQTarget(d *schema.ResourceData, targetType string) *functions.AsyncInvocationConfig_ResponseTarget {
	if v, ok := d.GetOk("async_invocation.0." + targetType + ".0"); ok {
		ymqSuccess := v.(map[string]interface{})
		saID := ymqSuccess["service_account_id"].(string)
		arn := ymqSuccess["arn"].(string)

		return &functions.AsyncInvocationConfig_ResponseTarget{
			Target: &functions.AsyncInvocationConfig_ResponseTarget_YmqTarget{
				YmqTarget: &functions.YMQTarget{
					ServiceAccountId: saID,
					QueueArn:         arn,
				},
			},
		}
	}
	return &functions.AsyncInvocationConfig_ResponseTarget{
		Target: &functions.AsyncInvocationConfig_ResponseTarget_EmptyTarget{},
	}
}

func flattenFunctionAsyncConfig(config *functions.AsyncInvocationConfig) []interface{} {
	if config == nil {
		return nil
	}
	res := map[string]interface{}{"retries_count": config.RetriesCount}
	if config.ServiceAccountId != "" {
		res["service_account_id"] = config.ServiceAccountId
	}
	if successTarget := flattenFunctionAsyncResponseTarget(config.SuccessTarget); successTarget != nil {
		res["ymq_success_target"] = successTarget
	}
	if failureTarget := flattenFunctionAsyncResponseTarget(config.SuccessTarget); failureTarget != nil {
		res["ymq_failure_target"] = failureTarget
	}
	return []interface{}{res}
}

func flattenFunctionAsyncResponseTarget(target *functions.AsyncInvocationConfig_ResponseTarget) []interface{} {
	switch s := target.Target.(type) {
	case *functions.AsyncInvocationConfig_ResponseTarget_YmqTarget:
		return []interface{}{
			map[string]interface{}{
				"service_account_id": s.YmqTarget.ServiceAccountId,
				"arn":                s.YmqTarget.QueueArn,
			},
		}
	default:
		return nil
	}
}

func expandFunctionLogOptions(d *schema.ResourceData) (*functions.LogOptions, error) {
	v, ok := d.GetOk("log_options.0")
	if !ok {
		return nil, nil
	}
	logOptionsMap := v.(map[string]interface{})
	if logOptionsMap["disabled"].(bool) {
		return &functions.LogOptions{
			Disabled: true,
		}, nil
	}
	logOptions := &functions.LogOptions{}
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

func flattenFunctionLogOptions(
	d *schema.ResourceData,
	logOptions *functions.LogOptions,
	functionFolderID string,
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
		case *functions.LogOptions_LogGroupId:
			res["log_group_id"] = destination.LogGroupId
		case *functions.LogOptions_FolderId:
			if allFields ||
				len(d.Get("log_options.0.folder_id").(string)) > 0 ||
				destination.FolderId != functionFolderID {

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
