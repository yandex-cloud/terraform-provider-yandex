package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	lt "github.com/yandex-cloud/go-genproto/yandex/cloud/loadtesting/api/v1"
)

func dataSourceYandexLoadtestingAgent() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexLoadtestingAgentRead,
		Schema: map[string]*schema.Schema{
			"agent_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"compute_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
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
						},
					},
				},
			},

			"compute_instance": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"labels": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},

						"computed_labels": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},

						"zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"platform_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"resources": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"memory": {
										Type:     schema.TypeFloat,
										Computed: true,
									},

									"cores": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"core_fraction": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},

						"metadata": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},

						"computed_metadata": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},

						"boot_disk": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_delete": {
										Type:     schema.TypeBool,
										Computed: true,
									},

									"device_name": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"disk_id": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"initialize_params": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"description": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"size": {
													Type:     schema.TypeInt,
													Computed: true,
												},

												"block_size": {
													Type:     schema.TypeInt,
													Computed: true,
												},

												"type": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},

						"network_interface": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"subnet_id": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"ipv4": {
										Type:     schema.TypeBool,
										Computed: true,
									},

									"ip_address": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"ipv6": {
										Type:     schema.TypeBool,
										Computed: true,
									},

									"ipv6_address": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"nat": {
										Type:     schema.TypeBool,
										Computed: true,
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
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexLoadtestingAgentRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	agentID := d.Get("agent_id").(string)

	agent, err := config.sdk.Loadtesting().Agent().Get(ctx, &lt.GetAgentRequest{
		AgentId: agentID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Loadtesting Agent with ID %q", agentID))
	}

	d.Set("name", agent.Name)
	d.Set("folder_id", agent.FolderId)
	d.Set("description", agent.Description)
	d.Set("compute_instance_id", agent.ComputeInstanceId)
	d.Set("labels", agent.Labels)

	d.SetId(agent.Id)

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
		return handleNotFoundError(err, d, fmt.Sprintf("Instance %q of Loadtesting Agent with ID %q", agent.ComputeInstanceId, agent.Name))
	}

	compute_instance_template, err := flattenLoadtestingComputeInstanceTemplate(ctx, instance, config, map[string]string{}, map[string]string{})
	if err != nil {
		return err
	}
	if err := d.Set("compute_instance", compute_instance_template); err != nil {
		return err
	}

	return nil
}
