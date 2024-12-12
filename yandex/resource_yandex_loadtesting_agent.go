package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	compute "github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	lt "github.com/yandex-cloud/go-genproto/yandex/cloud/loadtesting/api/v1"
	agent "github.com/yandex-cloud/go-genproto/yandex/cloud/loadtesting/api/v1/agent"

	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexLoadtestingDefaultTimeout = 10 * time.Minute

func resourceYandexLoadtestingAgent() *schema.Resource {
	return &schema.Resource{
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
				Type:     schema.TypeString,
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

			"compute_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"log_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"compute_instance": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_account_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"platform_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"resources": {
							Type:     schema.TypeList,
							ForceNew: true,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"memory": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ForceNew:     true,
										Default:      2,
										ValidateFunc: FloatAtLeast(1),
									},

									"cores": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										Default:      2,
										ValidateFunc: validation.IntAtLeast(1),
									},

									"core_fraction": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										Default:      100,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},

						"boot_disk": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_delete": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},

									"device_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},

									"disk_id": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"initialize_params": {
										Type:     schema.TypeList,
										Required: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
													ForceNew: true,
												},

												"description": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
													ForceNew: true,
												},

												"size": {
													Type:         schema.TypeInt,
													Optional:     true,
													ForceNew:     true,
													Default:      15,
													ValidateFunc: validation.IntAtLeast(1),
												},

												"block_size": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
													ForceNew: true,
												},

												"type": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
													Default:  "network-ssd",
												},
											},
										},
									},
								},
							},
						},

						"network_interface": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"subnet_id": {
										Type:     schema.TypeString,
										Required: true,
									},

									"ipv4": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},

									"ip_address": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},

									"ipv6": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},

									"ipv6_address": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},

									"nat": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
										ForceNew: true,
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
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},

									"nat_ip_version": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"security_group_ids": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},

						"zone_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},

						"computed_labels": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},

						"metadata": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},

						"computed_metadata": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
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
