package yandex

import (
	"archive/zip"
	"bytes"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

const yandexFunctionDefaultTimeout = 10 * time.Minute
const versionCreateSourceContentMaxBytes = 3670016

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

	versionReq, err := expandLastVersion(d)
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

	lastVersionPaths := []string{"user_hash", "runtime", "entrypoint", "memory", "execution_timeout", "service_account_id", "environment", "tags", "package", "content"}
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

func expandLastVersion(d *schema.ResourceData) (*functions.CreateFunctionVersionRequest, error) {
	versionReq := &functions.CreateFunctionVersionRequest{}
	versionReq.Runtime = d.Get("runtime").(string)
	versionReq.Entrypoint = d.Get("entrypoint").(string)

	versionReq.Resources = &functions.Resources{Memory: int64(int(datasize.MB.Bytes()) * d.Get("memory").(int))}
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

	return versionReq, nil
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
	d.Set("image_size", version.ImageSize)
	d.Set("loggroup_id", version.LogGroupId)
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

	tags := &schema.Set{F: schema.HashString}
	for _, v := range version.Tags {
		if v != "$latest" {
			tags.Add(v)
		}
	}
	return d.Set("tags", tags)
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
