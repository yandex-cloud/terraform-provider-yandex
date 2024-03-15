package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	compute "github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	lt "github.com/yandex-cloud/go-genproto/yandex/cloud/loadtesting/api/v1"
)

const yandexLoadtestingDefaultTimeout = 10 * time.Minute

func resourceYandexLoadtestingAgent() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexLoadtestingAgentCreate,
		Read:   resourceYandexLoadtestingAgentRead,
		Delete: resourceYandexLoadtestingAgentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexLoadtestingDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexLoadtestingDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
				ForceNew: true,
			},

			"compute_instance": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_account_id": {
							Type:     schema.TypeString,
							Required: true,
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
									},

									"ip_address": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},

									"ipv6": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},

									"ipv6_address": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},

									"nat": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
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

	req := lt.CreateAgentRequest{
		FolderId:              folderID,
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		Labels:                labels,
		ComputeInstanceParams: computeParams,
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

	instance, err := config.sdk.Compute().Instance().Get(ctx, &compute.GetInstanceRequest{
		InstanceId: agent.ComputeInstanceId,
		View:       compute.InstanceView_FULL,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Instance %q of Loadtesting Agent %q", agent.ComputeInstanceId, agent.Name))
	}

	origMetadata := d.Get("compute_instance.0.metadata")
	compute_instance_template, err := flattenLoadtestingComputeInstanceTemplate(ctx, instance, config, origMetadata)
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
