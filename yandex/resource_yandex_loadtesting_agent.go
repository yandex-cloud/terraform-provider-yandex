package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	compute "github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	lt "github.com/yandex-cloud/go-genproto/yandex/cloud/loadtesting/api/v1"
	agent "github.com/yandex-cloud/go-genproto/yandex/cloud/loadtesting/api/v1/agent"
	"github.com/yandex-cloud/terraform-provider-yandex/common"

	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexLoadtestingDefaultTimeout = 10 * time.Minute

func resourceYandexLoadtestingAgent() *schema.Resource {
	return &schema.Resource{
		Description: "A Load Testing Agent resource. For more information, see [the official documentation](https://yandex.cloud/docs/load-testing/concepts/agent).",

		Create: resourceYandexLoadtestingAgentCreate,
		Read:   resourceYandexLoadtestingAgentRead,
		Delete: resourceYandexLoadtestingAgentDelete,
		Update: resourceYandexLoadtestingAgentUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexLoadtestingDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexLoadtestingDefaultTimeout),
			Update: schema.DefaultTimeout(yandexLoadtestingDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"compute_instance_id": {
				Type:        schema.TypeString,
				Description: "Compute Instance ID.",
				Computed:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"log_settings": {
				Type:        schema.TypeList,
				Description: "The logging settings of the load testing agent.",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_id": {
							Type:        schema.TypeString,
							Description: "The ID of cloud logging group to which the load testing agent sends logs.",
							Optional:    true,
							ForceNew:    true,
						},
					},
				},
			},

			"compute_instance": {
				Type:        schema.TypeList,
				Description: "The template for creating new compute instance running load testing agent.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_account_id": {
							Type:        schema.TypeString,
							Description: "The ID of the service account authorized for this load testing agent. Service account should have `loadtesting.generatorClient` or `loadtesting.externalAgent` role in the folder.",
							Required:    true,
						},

						"platform_id": {
							Type:        schema.TypeString,
							Description: "The Compute platform for virtual machine.",
							Optional:    true,
							Computed:    true,
							ForceNew:    true,
						},

						"resources": {
							Type:        schema.TypeList,
							Description: "Compute resource specifications for the instance.",
							ForceNew:    true,
							Required:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"memory": {
										Type:         schema.TypeFloat,
										Description:  "The memory size in GB. Defaults to 2 GB.",
										Optional:     true,
										ForceNew:     true,
										Default:      2,
										ValidateFunc: FloatAtLeast(1),
									},

									"cores": {
										Type:         schema.TypeInt,
										Description:  "The number of CPU cores for the instance. Defaults to 2 cores.",
										Optional:     true,
										ForceNew:     true,
										Default:      2,
										ValidateFunc: validation.IntAtLeast(1),
									},

									"core_fraction": {
										Type:         schema.TypeInt,
										Description:  "If provided, specifies baseline core performance as a percent.",
										Optional:     true,
										ForceNew:     true,
										Default:      100,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},

						"boot_disk": {
							Type:        schema.TypeList,
							Description: "Boot disk specifications for the instance.",
							Required:    true,
							ForceNew:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_delete": {
										Type:        schema.TypeBool,
										Description: "Whether the disk is auto-deleted when the instance is deleted. The default value is true.",
										Optional:    true,
										Default:     true,
										ForceNew:    true,
									},

									"device_name": {
										Type:        schema.TypeString,
										Description: "This value can be used to reference the device under `/dev/disk/by-id/`.",
										Optional:    true,
										Computed:    true,
										ForceNew:    true,
									},

									"disk_id": {
										Type:        schema.TypeString,
										Description: "The ID of created disk.",
										Computed:    true,
									},

									"initialize_params": {
										Type:        schema.TypeList,
										Description: "Parameters for creating a disk alongside the instance.",
										Required:    true,
										ForceNew:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Description: "A name of the boot disk.",
													Optional:    true,
													Computed:    true,
													ForceNew:    true,
												},

												"description": {
													Type:        schema.TypeString,
													Description: "A description of the boot disk.",
													Optional:    true,
													Computed:    true,
													ForceNew:    true,
												},

												"size": {
													Type:         schema.TypeInt,
													Description:  "The size of the disk in GB. Defaults to 15 GB.",
													Optional:     true,
													ForceNew:     true,
													Default:      15,
													ValidateFunc: validation.IntAtLeast(1),
												},

												"block_size": {
													Type:        schema.TypeInt,
													Description: "Block size of the disk, specified in bytes.",
													Optional:    true,
													Computed:    true,
													ForceNew:    true,
												},

												"type": {
													Type:        schema.TypeString,
													Description: "The disk type.",
													Optional:    true,
													ForceNew:    true,
													Default:     "network-ssd",
												},
											},
										},
									},
								},
							},
						},

						"network_interface": {
							Type:        schema.TypeList,
							Description: "Network specifications for the instance. This can be used multiple times for adding multiple interfaces.",
							Required:    true,
							ForceNew:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"subnet_id": {
										Type:        schema.TypeString,
										Description: "The ID of the subnet to attach this interface to. The subnet must reside in the same zone where this instance was created.",
										Required:    true,
									},

									"ipv4": {
										Type:        schema.TypeBool,
										Description: "Flag for allocating IPv4 address for the network interface.",
										Optional:    true,
										Default:     true,
										ForceNew:    true,
									},

									"ip_address": {
										Type:        schema.TypeString,
										Description: "Manual set static IP address.",
										Optional:    true,
										Computed:    true,
										ForceNew:    true,
									},

									"ipv6": {
										Type:        schema.TypeBool,
										Description: "Flag for allocating IPv6 address for the network interface.",
										Optional:    true,
										Computed:    true,
										ForceNew:    true,
									},

									"ipv6_address": {
										Type:        schema.TypeString,
										Description: "Manual set static IPv6 address.",
										Optional:    true,
										Computed:    true,
										ForceNew:    true,
									},

									"nat": {
										Type:        schema.TypeBool,
										Description: "Flag for using NAT.",
										Optional:    true,
										Default:     false,
										ForceNew:    true,
									},

									"index": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"mac_address": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"nat_ip_address": {
										Type:        schema.TypeString,
										Description: "A public address that can be used to access the internet over NAT.",
										Optional:    true,
										Computed:    true,
										ForceNew:    true,
									},

									"nat_ip_version": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"security_group_ids": {
										Type:        schema.TypeSet,
										Description: "Security group ids for network interface.",
										Computed:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Set:         schema.HashString,
										Optional:    true,
										ForceNew:    true,
									},
								},
							},
						},

						"zone_id": {
							Type:        schema.TypeString,
							Description: common.ResourceDescriptions["zone"],
							Optional:    true,
							Computed:    true,
							ForceNew:    true,
						},

						"labels": {
							Type:        schema.TypeMap,
							Description: "A set of key/value label pairs to assign to the instance.",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},

						"computed_labels": {
							Type:        schema.TypeMap,
							Description: "The set of labels `key:value` pairs assigned to this instance. This includes user custom `labels` and predefined items created by Yandex Cloud Load Testing.",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},

						"metadata": {
							Type:        schema.TypeMap,
							Description: "A set of metadata key/value pairs to make available from within the instance.",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},

						"computed_metadata": {
							Type:        schema.TypeMap,
							Description: "The set of metadata `key:value` pairs assigned to this instance. This includes user custom `metadata`, and predefined items created by Yandex Cloud Load Testing.",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},
					},
				},
			},
		},
	}
}

func resourceYandexLoadtestingAgentCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating Yandex Cloud Loadtesting Agent: %s", err)
	}

	computeParams, err := expandLoadtestingComputeInstanceTemplate(d, config)
	if err != nil {
		return err
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return err
	}

	logSettingsParams, err := expandLoadtestingAgentLogSettingsParams(d)
	if err != nil {
		return err
	}

	req := lt.CreateAgentRequest{
		FolderId:              folderID,
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		Labels:                labels,
		ComputeInstanceParams: computeParams,
		LogSettings:           logSettingsParams,
	}

	op, err := config.sdk.WrapOperation(config.sdk.Loadtesting().Agent().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Loadtesting Agent: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Loadtesting Agent: %s", err)
	}

	md, ok := protoMetadata.(*lt.CreateAgentMetadata)
	if !ok {
		return fmt.Errorf("Could not get Yandex Cloud Loadtesting Agent ID from create operation metadata")
	}

	d.SetId(md.AgentId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Yandex Cloud Loadtesting Agent: %s", err)
	}

	return resourceYandexLoadtestingAgentRead(d, meta)
}

func resourceYandexLoadtestingAgentRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := lt.GetAgentRequest{
		AgentId: d.Id(),
	}

	agent, err := config.sdk.Loadtesting().Agent().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Loadtesting Agent %q", d.Id()))
	}

	d.Set("name", agent.Name)
	d.Set("folder_id", agent.FolderId)
	d.Set("description", agent.Description)
	d.Set("compute_instance_id", agent.ComputeInstanceId)
	d.Set("labels", agent.Labels)

	logSettings, err := flattenLoadtestingAgentLogSettingsParams(agent)
	if err != nil {
		return err
	}
	if err := d.Set("log_settings", logSettings); err != nil {
		return err
	}

	instance, err := config.sdk.Compute().Instance().Get(ctx, &compute.GetInstanceRequest{
		InstanceId: agent.ComputeInstanceId,
		View:       compute.InstanceView_FULL,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Instance %q of Loadtesting Agent %q", agent.ComputeInstanceId, agent.Name))
	}

	origMetadata := d.Get("compute_instance.0.metadata")
	origLabels := d.Get("compute_instance.0.labels")
	compute_instance_template, err := flattenLoadtestingComputeInstanceTemplate(ctx, instance, config, origMetadata, origLabels)
	if err != nil {
		return err
	}
	if err := d.Set("compute_instance", compute_instance_template); err != nil {
		return err
	}

	return nil
}

func resourceYandexLoadtestingAgentDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := lt.DeleteAgentRequest{
		AgentId: d.Id(),
	}

	op, err := config.sdk.Loadtesting().Agent().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Loadtesting Agent %q", d.Id()))
	}

	return nil
}

func resourceYandexLoadtestingAgentUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req := &lt.UpdateAgentRequest{
		AgentId:               d.Id(),
		UpdateMask:            &field_mask.FieldMask{},
		ComputeInstanceParams: &agent.CreateComputeInstance{},
	}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	if d.HasChange("labels") {
		labels, err := expandLabels(d.Get("labels"))
		if err != nil {
			return err
		}
		req.Labels = labels
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	if d.HasChange("compute_instance") {
		computeParams, err := expandLoadtestingComputeInstanceTemplate(d, config)
		if err != nil {
			return err
		}

		req.ComputeInstanceParams = computeParams
		enrichUpdateMaskFromLoadtestingComputeInstanceTemplate(d, req)
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Loadtesting().Agent().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Agent %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while updating Agent %q: %s", d.Id(), err)
	}

	return resourceYandexLoadtestingAgentRead(d, meta)
}

func enrichUpdateMaskFromLoadtestingComputeInstanceTemplate(d *schema.ResourceData, req *lt.UpdateAgentRequest) {
	if d.HasChange("compute_instance.0.metadata") {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "compute_instance_params.metadata")
	}
	if d.HasChange("compute_instance.0.labels") {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "compute_instance_params.labels")
	}
	if d.HasChange("compute_instance.0.service_account_id") {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "compute_instance_params.service_account_id")
	}
}
