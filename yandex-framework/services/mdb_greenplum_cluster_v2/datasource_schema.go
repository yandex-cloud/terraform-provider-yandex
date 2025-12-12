package mdb_greenplum_cluster_v2

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func YandexMdbGreenplumClusterV2DatasourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "A Greenplum® cluster resource.",
		MarkdownDescription: "A Greenplum® cluster resource.",
		Attributes: map[string]schema.Attribute{

			"cloud_storage": schema.SingleNestedAttribute{

				Attributes: map[string]schema.Attribute{

					"enable": schema.BoolAttribute{
						MarkdownDescription: "enable Cloud Storage for cluster",
						Description: "enable Cloud Storage for cluster" +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.cloud_storageyandex.cloud.mdb.greenplum.v1.CloudStorage.enable
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},
				},
				MarkdownDescription: "Cloud storage settings",
				Description: "Cloud storage settings" +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.cloud_storage
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"cluster_config": schema.SingleNestedAttribute{

				Attributes: map[string]schema.Attribute{

					"background_activities": schema.SingleNestedAttribute{

						Attributes: map[string]schema.Attribute{

							"analyze_and_vacuum": schema.SingleNestedAttribute{

								Attributes: map[string]schema.Attribute{

									"analyze_timeout": schema.Int64Attribute{
										MarkdownDescription: "Maximum duration of the `ANALYZE` operation, in seconds. The default value is `36000`. As soon as this period expires, the `ANALYZE` operation will be forced to terminate.",
										Description: "Maximum duration of the `ANALYZE` operation, in seconds. The default value is `36000`. As soon as this period expires, the `ANALYZE` operation will be forced to terminate." +
											// proto paths: +
											// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.analyze_and_vacuumyandex.cloud.mdb.greenplum.v1.AnalyzeAndVacuum.analyze_timeout
											"package: yandex.cloud.mdb.greenplum.v1\n" +
											"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
										Computed: true,
									},

									"start": schema.SingleNestedAttribute{

										Attributes: map[string]schema.Attribute{

											"hours": schema.Int64Attribute{
												MarkdownDescription: "hours",
												Description: "hours" +
													// proto paths: +
													// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.analyze_and_vacuumyandex.cloud.mdb.greenplum.v1.AnalyzeAndVacuum.startyandex.cloud.mdb.greenplum.v1.BackgroundActivityStartAt.hours
													"package: yandex.cloud.mdb.greenplum.v1\n" +
													"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
												Computed: true,
											},

											"minutes": schema.Int64Attribute{
												MarkdownDescription: "minutes",
												Description: "minutes" +
													// proto paths: +
													// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.analyze_and_vacuumyandex.cloud.mdb.greenplum.v1.AnalyzeAndVacuum.startyandex.cloud.mdb.greenplum.v1.BackgroundActivityStartAt.minutes
													"package: yandex.cloud.mdb.greenplum.v1\n" +
													"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
												Computed: true,
											},
										},
										MarkdownDescription: "Time when analyze will start",
										Description: "Time when analyze will start" +
											// proto paths: +
											// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.analyze_and_vacuumyandex.cloud.mdb.greenplum.v1.AnalyzeAndVacuum.start
											"package: yandex.cloud.mdb.greenplum.v1\n" +
											"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
										Computed: true,
									},

									"vacuum_timeout": schema.Int64Attribute{
										MarkdownDescription: "Maximum duration of the `VACUUM` operation, in seconds. The default value is `36000`. As soon as this period expires, the `VACUUM` operation will be forced to terminate.",
										Description: "Maximum duration of the `VACUUM` operation, in seconds. The default value is `36000`. As soon as this period expires, the `VACUUM` operation will be forced to terminate." +
											// proto paths: +
											// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.analyze_and_vacuumyandex.cloud.mdb.greenplum.v1.AnalyzeAndVacuum.vacuum_timeout
											"package: yandex.cloud.mdb.greenplum.v1\n" +
											"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
										Computed: true,
									},
								},
								MarkdownDescription: "Configuration for `ANALYZE` and `VACUUM` operations.",
								Description: "Configuration for `ANALYZE` and `VACUUM` operations." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.analyze_and_vacuum
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"query_killer_scripts": schema.SingleNestedAttribute{

								Attributes: map[string]schema.Attribute{

									"idle": schema.SingleNestedAttribute{

										Attributes: map[string]schema.Attribute{

											"enable": schema.BoolAttribute{
												MarkdownDescription: "Use query killer or not",
												Description: "Use query killer or not" +
													// proto paths: +
													// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.idleyandex.cloud.mdb.greenplum.v1.QueryKiller.enable
													"package: yandex.cloud.mdb.greenplum.v1\n" +
													"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
												Computed: true,
											},

											"ignore_users": schema.SetAttribute{
												ElementType:         types.StringType,
												MarkdownDescription: "Ignore these users when considering queries to terminate",
												Description: "Ignore these users when considering queries to terminate" +
													// proto paths: +
													// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.idleyandex.cloud.mdb.greenplum.v1.QueryKiller.ignore_users
													"package: yandex.cloud.mdb.greenplum.v1\n" +
													"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
												Computed: true,
											},

											"max_age": schema.Int64Attribute{
												MarkdownDescription: "Maximum duration for this type of queries (in seconds).",
												Description: "Maximum duration for this type of queries (in seconds)." +
													// proto paths: +
													// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.idleyandex.cloud.mdb.greenplum.v1.QueryKiller.max_age
													"package: yandex.cloud.mdb.greenplum.v1\n" +
													"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
												Computed: true,
											},
										},
										MarkdownDescription: "Configuration of script that kills long running queries that are in `idle` state.",
										Description: "Configuration of script that kills long running queries that are in `idle` state." +
											// proto paths: +
											// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.idle
											"package: yandex.cloud.mdb.greenplum.v1\n" +
											"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
										Computed: true,
									},

									"idle_in_transaction": schema.SingleNestedAttribute{

										Attributes: map[string]schema.Attribute{

											"enable": schema.BoolAttribute{
												MarkdownDescription: "Use query killer or not",
												Description: "Use query killer or not" +
													// proto paths: +
													// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.idle_in_transactionyandex.cloud.mdb.greenplum.v1.QueryKiller.enable
													"package: yandex.cloud.mdb.greenplum.v1\n" +
													"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
												Computed: true,
											},

											"ignore_users": schema.SetAttribute{
												ElementType:         types.StringType,
												MarkdownDescription: "Ignore these users when considering queries to terminate",
												Description: "Ignore these users when considering queries to terminate" +
													// proto paths: +
													// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.idle_in_transactionyandex.cloud.mdb.greenplum.v1.QueryKiller.ignore_users
													"package: yandex.cloud.mdb.greenplum.v1\n" +
													"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
												Computed: true,
											},

											"max_age": schema.Int64Attribute{
												MarkdownDescription: "Maximum duration for this type of queries (in seconds).",
												Description: "Maximum duration for this type of queries (in seconds)." +
													// proto paths: +
													// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.idle_in_transactionyandex.cloud.mdb.greenplum.v1.QueryKiller.max_age
													"package: yandex.cloud.mdb.greenplum.v1\n" +
													"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
												Computed: true,
											},
										},
										MarkdownDescription: "Configuration of script that kills long running queries that are in `idle in transaction` state.",
										Description: "Configuration of script that kills long running queries that are in `idle in transaction` state." +
											// proto paths: +
											// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.idle_in_transaction
											"package: yandex.cloud.mdb.greenplum.v1\n" +
											"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
										Computed: true,
									},

									"long_running": schema.SingleNestedAttribute{

										Attributes: map[string]schema.Attribute{

											"enable": schema.BoolAttribute{
												MarkdownDescription: "Use query killer or not",
												Description: "Use query killer or not" +
													// proto paths: +
													// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.long_runningyandex.cloud.mdb.greenplum.v1.QueryKiller.enable
													"package: yandex.cloud.mdb.greenplum.v1\n" +
													"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
												Computed: true,
											},

											"ignore_users": schema.SetAttribute{
												ElementType:         types.StringType,
												MarkdownDescription: "Ignore these users when considering queries to terminate",
												Description: "Ignore these users when considering queries to terminate" +
													// proto paths: +
													// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.long_runningyandex.cloud.mdb.greenplum.v1.QueryKiller.ignore_users
													"package: yandex.cloud.mdb.greenplum.v1\n" +
													"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
												Computed: true,
											},

											"max_age": schema.Int64Attribute{
												MarkdownDescription: "Maximum duration for this type of queries (in seconds).",
												Description: "Maximum duration for this type of queries (in seconds)." +
													// proto paths: +
													// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.long_runningyandex.cloud.mdb.greenplum.v1.QueryKiller.max_age
													"package: yandex.cloud.mdb.greenplum.v1\n" +
													"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
												Computed: true,
											},
										},
										MarkdownDescription: "Configuration of script that kills long running queries (in any state).",
										Description: "Configuration of script that kills long running queries (in any state)." +
											// proto paths: +
											// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scriptsyandex.cloud.mdb.greenplum.v1.QueryKillerScripts.long_running
											"package: yandex.cloud.mdb.greenplum.v1\n" +
											"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
										Computed: true,
									},
								},
								MarkdownDescription: "Configuration for long running queries killer.",
								Description: "Configuration for long running queries killer." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.query_killer_scripts
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"table_sizes": schema.SingleNestedAttribute{

								Attributes: map[string]schema.Attribute{

									"starts": schema.SetNestedAttribute{
										NestedObject: schema.NestedAttributeObject{

											Attributes: map[string]schema.Attribute{

												"hours": schema.Int64Attribute{
													MarkdownDescription: "hours",
													Description: "hours" +
														// proto paths: +
														// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.table_sizesyandex.cloud.mdb.greenplum.v1.TableSizes.startsyandex.cloud.mdb.greenplum.v1.BackgroundActivityStartAt.hours
														"package: yandex.cloud.mdb.greenplum.v1\n" +
														"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
													Computed: true,
												},

												"minutes": schema.Int64Attribute{
													MarkdownDescription: "minutes",
													Description: "minutes" +
														// proto paths: +
														// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.table_sizesyandex.cloud.mdb.greenplum.v1.TableSizes.startsyandex.cloud.mdb.greenplum.v1.BackgroundActivityStartAt.minutes
														"package: yandex.cloud.mdb.greenplum.v1\n" +
														"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
													Computed: true,
												},
											},
										},
										MarkdownDescription: "Time when start \"table_sizes\" script",
										Description: "Time when start \"table_sizes\" script" +
											// proto paths: +
											// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.table_sizesyandex.cloud.mdb.greenplum.v1.TableSizes.starts
											"package: yandex.cloud.mdb.greenplum.v1\n" +
											"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
										Computed: true,
									},
								},
								MarkdownDescription: "Enables scripts that collects tables sizes to `*_sizes` tables in `mdb_toolkit` schema.",
								Description: "Enables scripts that collects tables sizes to `*_sizes` tables in `mdb_toolkit` schema." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activitiesyandex.cloud.mdb.greenplum.v1.BackgroundActivitiesConfig.table_sizes
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},
						},
						MarkdownDescription: "Managed Greenplum® background tasks configuration.",
						Description: "Managed Greenplum® background tasks configuration." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.background_activities
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"greenplum_config_set_6": schema.SingleNestedAttribute{

						Attributes: map[string]schema.Attribute{

							"gp_add_column_inherits_table_setting": schema.BoolAttribute{
								MarkdownDescription: "https://docs.vmware.com/en/VMware-Tanzu-Greenplum/6/greenplum-database/GUID-ref_guide-config_params-guc-list.html#gp_add_column_inherits_table_setting",
								Description: "https://docs.vmware.com/en/VMware-Tanzu-Greenplum/6/greenplum-database/GUID-ref_guide-config_params-guc-list.html#gp_add_column_inherits_table_setting" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_add_column_inherits_table_setting
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_autostats_mode": schema.StringAttribute{
								MarkdownDescription: "Specifies the mode for triggering automatic statistics collection after DML.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_autostats_mode",
								Description: "Specifies the mode for triggering automatic statistics collection after DML.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_autostats_mode" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_autostats_mode
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_autostats_on_change_threshold": schema.Int64Attribute{
								MarkdownDescription: "Specifies the threshold for automatic statistics collection when gp_autostats_mode is set to on_change.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_autostats_on_change_threshold",
								Description: "Specifies the threshold for automatic statistics collection when gp_autostats_mode is set to on_change.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_autostats_on_change_threshold" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_autostats_on_change_threshold
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_cached_segworkers_threshold": schema.Int64Attribute{
								MarkdownDescription: "Define amount of working processes in segment, that keeping in warm cash after end of query for usage again in next queries.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_cached_segworkers_threshold",
								Description: "Define amount of working processes in segment, that keeping in warm cash after end of query for usage again in next queries.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_cached_segworkers_threshold" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_cached_segworkers_threshold
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_enable_global_deadlock_detector": schema.BoolAttribute{
								MarkdownDescription: "Controls whether the Greenplum Database Global Deadlock Detector is enabled to manage concurrent UPDATE and DELETE operations on heap tables to improve performance. See Inserting, Updating, and Deleting Datain the Greenplum Database Administrator Guide. The default is off, the Global Deadlock Detector is deactivated.\n If the Global Deadlock Detector is deactivated (the default), Greenplum Database runs concurrent update and delete operations on a heap table serially.\n If the Global Deadlock Detector is enabled, concurrent updates are permitted and the Global Deadlock Detector determines when a deadlock exists, and breaks the deadlock by cancelling one or more backend processes associated with the youngest transaction(s) involved.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_enable_global_deadlock_detector",
								Description: "Controls whether the Greenplum Database Global Deadlock Detector is enabled to manage concurrent UPDATE and DELETE operations on heap tables to improve performance. See Inserting, Updating, and Deleting Datain the Greenplum Database Administrator Guide. The default is off, the Global Deadlock Detector is deactivated.\n If the Global Deadlock Detector is deactivated (the default), Greenplum Database runs concurrent update and delete operations on a heap table serially.\n If the Global Deadlock Detector is enabled, concurrent updates are permitted and the Global Deadlock Detector determines when a deadlock exists, and breaks the deadlock by cancelling one or more backend processes associated with the youngest transaction(s) involved.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_enable_global_deadlock_detector" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_enable_global_deadlock_detector
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_enable_zstd_memory_accounting": schema.BoolAttribute{
								MarkdownDescription: "Forces ZSTD lib use Greenplum memory allocation system.",
								Description: "Forces ZSTD lib use Greenplum memory allocation system." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_enable_zstd_memory_accounting
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_global_deadlock_detector_period": schema.Int64Attribute{
								MarkdownDescription: "Specifies the executing interval (in seconds) of the global deadlock detector background worker process.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_global_deadlock_detector_period",
								Description: "Specifies the executing interval (in seconds) of the global deadlock detector background worker process.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_global_deadlock_detector_period" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_global_deadlock_detector_period
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_max_plan_size": schema.Int64Attribute{
								MarkdownDescription: "Specifies the total maximum uncompressed size of a query execution plan multiplied by the number of Motion operators (slices) in the plan.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_max_plan_size",
								Description: "Specifies the total maximum uncompressed size of a query execution plan multiplied by the number of Motion operators (slices) in the plan.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_max_plan_size" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_max_plan_size
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_max_slices": schema.Int64Attribute{
								MarkdownDescription: "Max amount of slice-processes for one query in one segment.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_max_slices",
								Description: "Max amount of slice-processes for one query in one segment.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_max_slices" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_max_slices
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_resource_group_memory_limit": schema.Float64Attribute{
								MarkdownDescription: "Identifies the maximum percentage of system memory resources to allocate to resource groups on each Greenplum Database segment node.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_resource_group_memory_limit",
								Description: "Identifies the maximum percentage of system memory resources to allocate to resource groups on each Greenplum Database segment node.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_resource_group_memory_limit" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_resource_group_memory_limit
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_vmem_protect_segworker_cache_limit": schema.Int64Attribute{
								MarkdownDescription: "Set memory limit (in MB) for working process. If a query executor process consumes more than this configured amount, then the process will not be cached for use in subsequent queries after the process completes.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_vmem_protect_segworker_cache_limit",
								Description: "Set memory limit (in MB) for working process. If a query executor process consumes more than this configured amount, then the process will not be cached for use in subsequent queries after the process completes.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#gp_vmem_protect_segworker_cache_limit" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_vmem_protect_segworker_cache_limit
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_workfile_compression": schema.BoolAttribute{
								MarkdownDescription: "Specifies whether the temporary files created, when a hash aggregation or hash join operation spills to disk, are compressed.\n https://docs.greenplum.org/6-5/ref_guide/config_params/guc-list.html#gp_workfile_compression",
								Description: "Specifies whether the temporary files created, when a hash aggregation or hash join operation spills to disk, are compressed.\n https://docs.greenplum.org/6-5/ref_guide/config_params/guc-list.html#gp_workfile_compression" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_workfile_compression
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_workfile_limit_files_per_query": schema.Int64Attribute{
								MarkdownDescription: "Sets the maximum number of temporary spill files (also known as workfiles) allowed per query per segment.\n Spill files are created when executing a query that requires more memory than it is allocated.\n The current query is terminated when the limit is exceeded.\n Set the value to 0 (zero) to allow an unlimited number of spill files. master session reload\n https://docs.greenplum.org/6-5/ref_guide/config_params/guc-list.html#gp_workfile_limit_files_per_query\n Default value is 10000",
								Description: "Sets the maximum number of temporary spill files (also known as workfiles) allowed per query per segment.\n Spill files are created when executing a query that requires more memory than it is allocated.\n The current query is terminated when the limit is exceeded.\n Set the value to 0 (zero) to allow an unlimited number of spill files. master session reload\n https://docs.greenplum.org/6-5/ref_guide/config_params/guc-list.html#gp_workfile_limit_files_per_query\n Default value is 10000" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_workfile_limit_files_per_query
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_workfile_limit_per_query": schema.Int64Attribute{
								MarkdownDescription: "Sets the maximum disk size an individual query is allowed to use for creating temporary spill files at each segment.\n The default value is 0, which means a limit is not enforced.\n https://docs.greenplum.org/6-5/ref_guide/config_params/guc-list.html#gp_workfile_limit_per_query",
								Description: "Sets the maximum disk size an individual query is allowed to use for creating temporary spill files at each segment.\n The default value is 0, which means a limit is not enforced.\n https://docs.greenplum.org/6-5/ref_guide/config_params/guc-list.html#gp_workfile_limit_per_query" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_workfile_limit_per_query
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"gp_workfile_limit_per_segment": schema.Int64Attribute{
								MarkdownDescription: "Sets the maximum total disk size that all running queries are allowed to use for creating temporary spill files at each segment.\n The default value is 0, which means a limit is not enforced.\n https://docs.greenplum.org/6-5/ref_guide/config_params/guc-list.html#gp_workfile_limit_per_segment",
								Description: "Sets the maximum total disk size that all running queries are allowed to use for creating temporary spill files at each segment.\n The default value is 0, which means a limit is not enforced.\n https://docs.greenplum.org/6-5/ref_guide/config_params/guc-list.html#gp_workfile_limit_per_segment" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.gp_workfile_limit_per_segment
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"idle_in_transaction_session_timeout": schema.Int64Attribute{
								MarkdownDescription: "Max time (in ms) which session can idle in open transaction\n https://postgrespro.ru/docs/postgrespro/current/runtime-config-client#GUC-IDLE-IN-TRANSACTION-SESSION-TIMEOUT",
								Description: "Max time (in ms) which session can idle in open transaction\n https://postgrespro.ru/docs/postgrespro/current/runtime-config-client#GUC-IDLE-IN-TRANSACTION-SESSION-TIMEOUT" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.idle_in_transaction_session_timeout
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"lock_timeout": schema.Int64Attribute{
								MarkdownDescription: "Max time (in ms) which query will wait lock free on object\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#lock_timeout",
								Description: "Max time (in ms) which query will wait lock free on object\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#lock_timeout" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.lock_timeout
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"log_statement": schema.StringAttribute{
								MarkdownDescription: "Controls which SQL statements are logged. DDL logs all data definition commands like CREATE, ALTER, and DROP commands.\n MOD logs all DDL statements, plus INSERT, UPDATE, DELETE, TRUNCATE, and COPY FROM.\n PREPARE and EXPLAIN ANALYZE statements are also logged if their contained command is of an appropriate type.\n https://docs.greenplum.org/6-5/ref_guide/config_params/guc-list.html#log_statement\n Default value is ddl",
								Description: "Controls which SQL statements are logged. DDL logs all data definition commands like CREATE, ALTER, and DROP commands.\n MOD logs all DDL statements, plus INSERT, UPDATE, DELETE, TRUNCATE, and COPY FROM.\n PREPARE and EXPLAIN ANALYZE statements are also logged if their contained command is of an appropriate type.\n https://docs.greenplum.org/6-5/ref_guide/config_params/guc-list.html#log_statement\n Default value is ddl" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.log_statement
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"max_connections": schema.Int64Attribute{
								MarkdownDescription: "Maximum number of inbound connections on master segment",
								Description: "Maximum number of inbound connections on master segment" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.max_connections
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"max_prepared_transactions": schema.Int64Attribute{
								MarkdownDescription: "Sets the maximum number of transactions that can be in the \"prepared\" state simultaneously\n https://www.postgresql.org/docs/9.6/runtime-config-resource.html",
								Description: "Sets the maximum number of transactions that can be in the \"prepared\" state simultaneously\n https://www.postgresql.org/docs/9.6/runtime-config-resource.html" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.max_prepared_transactions
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"max_slot_wal_keep_size": schema.Int64Attribute{
								MarkdownDescription: "Specify the maximum size of WAL files that replication slots are allowed to retain in the pg_wal directory at checkpoint time.\n https://www.postgresql.org/docs/current/runtime-config-replication.html",
								Description: "Specify the maximum size of WAL files that replication slots are allowed to retain in the pg_wal directory at checkpoint time.\n https://www.postgresql.org/docs/current/runtime-config-replication.html" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.max_slot_wal_keep_size
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"max_statement_mem": schema.Int64Attribute{
								MarkdownDescription: "Sets the maximum memory limit for a query. Helps avoid out-of-memory errors on a segment host during query processing as a result of setting statement_mem too high.\n Taking into account the configuration of a single segment host, calculate max_statement_mem as follows:\n (seghost_physical_memory) / (average_number_concurrent_queries)\n When changing both max_statement_mem and statement_mem, max_statement_mem must be changed first, or listed first in the postgresql.conf file.\n https://greenplum.docs.pivotal.io/6-19/ref_guide/config_params/guc-list.html#max_statement_mem\n Default value is 2097152000 (2000MB)",
								Description: "Sets the maximum memory limit for a query. Helps avoid out-of-memory errors on a segment host during query processing as a result of setting statement_mem too high.\n Taking into account the configuration of a single segment host, calculate max_statement_mem as follows:\n (seghost_physical_memory) / (average_number_concurrent_queries)\n When changing both max_statement_mem and statement_mem, max_statement_mem must be changed first, or listed first in the postgresql.conf file.\n https://greenplum.docs.pivotal.io/6-19/ref_guide/config_params/guc-list.html#max_statement_mem\n Default value is 2097152000 (2000MB)" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.max_statement_mem
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"runaway_detector_activation_percent": schema.Int64Attribute{
								MarkdownDescription: "Percent of utilized Greenplum Database vmem that triggers the termination of queries.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#runaway_detector_activation_percent",
								Description: "Percent of utilized Greenplum Database vmem that triggers the termination of queries.\n https://techdocs.broadcom.com/us/en/vmware-tanzu/data-solutions/tanzu-greenplum/6/greenplum-database/ref_guide-config_params-guc-list.html#runaway_detector_activation_percent" +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6yandex.cloud.mdb.greenplum.v1.GreenplumConfigSet6.user_configyandex.cloud.mdb.greenplum.v1.GreenplumConfig6.runaway_detector_activation_percent
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},
						},
						MarkdownDescription: "",
						Description: "" +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.greenplum_config_set_6
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"pool": schema.SingleNestedAttribute{

						Attributes: map[string]schema.Attribute{

							"client_idle_timeout": schema.Int64Attribute{
								MarkdownDescription: "Client pool idle timeout, in seconds.\n\n Drop stale client connection after this much seconds of idleness, which is not in transaction.\n\n Set to zero to disable.",
								Description: "Client pool idle timeout, in seconds.\n\n Drop stale client connection after this much seconds of idleness, which is not in transaction.\n\n Set to zero to disable." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.poolyandex.cloud.mdb.greenplum.v1.ConnectionPoolerConfigSet.user_configyandex.cloud.mdb.greenplum.v1.ConnectionPoolerConfig.client_idle_timeout
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"idle_in_transaction_timeout": schema.Int64Attribute{
								MarkdownDescription: "Client pool idle in transaction timeout, in seconds.\n\n Drop client connection in transaction after this much seconds of idleness.\n\n Set to zero to disable.",
								Description: "Client pool idle in transaction timeout, in seconds.\n\n Drop client connection in transaction after this much seconds of idleness.\n\n Set to zero to disable." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.poolyandex.cloud.mdb.greenplum.v1.ConnectionPoolerConfigSet.user_configyandex.cloud.mdb.greenplum.v1.ConnectionPoolerConfig.idle_in_transaction_timeout
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"mode": schema.StringAttribute{
								MarkdownDescription: "Route server pool mode.",
								Description: "Route server pool mode." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.poolyandex.cloud.mdb.greenplum.v1.ConnectionPoolerConfigSet.user_configyandex.cloud.mdb.greenplum.v1.ConnectionPoolerConfig.mode
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"size": schema.Int64Attribute{
								MarkdownDescription: "The number of servers in the server pool. Clients are placed in a wait queue when all servers are busy.\n\n Set to zero to disable the limit.",
								Description: "The number of servers in the server pool. Clients are placed in a wait queue when all servers are busy.\n\n Set to zero to disable the limit." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.poolyandex.cloud.mdb.greenplum.v1.ConnectionPoolerConfigSet.user_configyandex.cloud.mdb.greenplum.v1.ConnectionPoolerConfig.size
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},
						},
						MarkdownDescription: "Odyssey® pool settings.",
						Description: "Odyssey® pool settings." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.pool
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"pxf_config": schema.SingleNestedAttribute{

						Attributes: map[string]schema.Attribute{

							"connection_timeout": schema.Int64Attribute{
								MarkdownDescription: "Timeout for connection to the Apache Tomcat® server when making read requests.\n\n Specify values in seconds.",
								Description: "Timeout for connection to the Apache Tomcat® server when making read requests.\n\n Specify values in seconds." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.pxf_configyandex.cloud.mdb.greenplum.v1.PXFConfigSet.user_configyandex.cloud.mdb.greenplum.v1.PXFConfig.connection_timeout
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/pxf.proto\n",
								Computed: true,
							},

							"max_threads": schema.Int64Attribute{
								MarkdownDescription: "Maximum number of the Apache Tomcat® threads.\n\n To prevent situations when requests get stuck or fail due to running out of memory or malfunctioning of the Java garbage collector, specify the number of the Apache Tomcat® threads. Learn more about adjusting the number of threads in the [VMware Greenplum® Platform Extension Framework](https://docs.vmware.com/en/VMware-Greenplum-Platform-Extension-Framework/6.9/greenplum-platform-extension-framework/cfg_mem.html) documentation.",
								Description: "Maximum number of the Apache Tomcat® threads.\n\n To prevent situations when requests get stuck or fail due to running out of memory or malfunctioning of the Java garbage collector, specify the number of the Apache Tomcat® threads. Learn more about adjusting the number of threads in the [VMware Greenplum® Platform Extension Framework](https://docs.vmware.com/en/VMware-Greenplum-Platform-Extension-Framework/6.9/greenplum-platform-extension-framework/cfg_mem.html) documentation." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.pxf_configyandex.cloud.mdb.greenplum.v1.PXFConfigSet.user_configyandex.cloud.mdb.greenplum.v1.PXFConfig.max_threads
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/pxf.proto\n",
								Computed: true,
							},

							"pool_allow_core_thread_timeout": schema.BoolAttribute{
								MarkdownDescription: "Determines whether the timeout for core streaming threads is permitted.",
								Description: "Determines whether the timeout for core streaming threads is permitted." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.pxf_configyandex.cloud.mdb.greenplum.v1.PXFConfigSet.user_configyandex.cloud.mdb.greenplum.v1.PXFConfig.pool_allow_core_thread_timeout
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/pxf.proto\n",
								Computed: true,
							},

							"pool_core_size": schema.Int64Attribute{
								MarkdownDescription: "Number of core streaming threads per pool.",
								Description: "Number of core streaming threads per pool." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.pxf_configyandex.cloud.mdb.greenplum.v1.PXFConfigSet.user_configyandex.cloud.mdb.greenplum.v1.PXFConfig.pool_core_size
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/pxf.proto\n",
								Computed: true,
							},

							"pool_max_size": schema.Int64Attribute{
								MarkdownDescription: "Maximum allowed number of core streaming threads.",
								Description: "Maximum allowed number of core streaming threads." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.pxf_configyandex.cloud.mdb.greenplum.v1.PXFConfigSet.user_configyandex.cloud.mdb.greenplum.v1.PXFConfig.pool_max_size
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/pxf.proto\n",
								Computed: true,
							},

							"pool_queue_capacity": schema.Int64Attribute{
								MarkdownDescription: "Maximum number of requests you can add to a pool queue for core streaming threads.\n\n If `0`, no pool queue is generated.",
								Description: "Maximum number of requests you can add to a pool queue for core streaming threads.\n\n If `0`, no pool queue is generated." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.pxf_configyandex.cloud.mdb.greenplum.v1.PXFConfigSet.user_configyandex.cloud.mdb.greenplum.v1.PXFConfig.pool_queue_capacity
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/pxf.proto\n",
								Computed: true,
							},

							"upload_timeout": schema.Int64Attribute{
								MarkdownDescription: "Timeout for connection to the Apache Tomcat® server when making write requests.\n\n Specify the values in seconds.",
								Description: "Timeout for connection to the Apache Tomcat® server when making write requests.\n\n Specify the values in seconds." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.pxf_configyandex.cloud.mdb.greenplum.v1.PXFConfigSet.user_configyandex.cloud.mdb.greenplum.v1.PXFConfig.upload_timeout
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/pxf.proto\n",
								Computed: true,
							},

							"xms": schema.Int64Attribute{
								MarkdownDescription: "Maximum size, in megabytes, of the JVM heap for the PXF daemon.",
								Description: "Maximum size, in megabytes, of the JVM heap for the PXF daemon." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.pxf_configyandex.cloud.mdb.greenplum.v1.PXFConfigSet.user_configyandex.cloud.mdb.greenplum.v1.PXFConfig.xms
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/pxf.proto\n",
								Computed: true,
							},

							"xmx": schema.Int64Attribute{
								MarkdownDescription: "Initial size, in megabytes, of the JVM heap for the PXF daemon.",
								Description: "Initial size, in megabytes, of the JVM heap for the PXF daemon." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.pxf_configyandex.cloud.mdb.greenplum.v1.PXFConfigSet.user_configyandex.cloud.mdb.greenplum.v1.PXFConfig.xmx
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/pxf.proto\n",
								Computed: true,
							},
						},
						MarkdownDescription: "",
						Description: "" +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_configyandex.cloud.mdb.greenplum.v1.ClusterConfigSet.pxf_config
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},
				},
				MarkdownDescription: "Greenplum® and Odyssey® configuration.",
				Description: "Greenplum® and Odyssey® configuration." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.cluster_config
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Greenplum® cluster resource to return.\n\n To get the cluster ID, use a [ClusterService.List] request.",
				Description: "ID of the Greenplum® cluster resource to return.\n\n To get the cluster ID, use a [ClusterService.List] request." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.id
					// -> yandex.cloud.mdb.greenplum.v1.GetClusterRequest.cluster_id
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster_service.proto\n",
				Optional: true,
				Computed: true,

				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 50),
				},
			},

			"config": schema.SingleNestedAttribute{

				Attributes: map[string]schema.Attribute{

					"access": schema.SingleNestedAttribute{

						Attributes: map[string]schema.Attribute{

							"data_lens": schema.BoolAttribute{
								MarkdownDescription: "Allows data export from the cluster to DataLens.",
								Description: "Allows data export from the cluster to DataLens." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.configyandex.cloud.mdb.greenplum.v1.GreenplumConfig.accessyandex.cloud.mdb.greenplum.v1.Access.data_lens
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
								Computed: true,
							},

							"data_transfer": schema.BoolAttribute{
								MarkdownDescription: "Allows access for DataTransfer.",
								Description: "Allows access for DataTransfer." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.configyandex.cloud.mdb.greenplum.v1.GreenplumConfig.accessyandex.cloud.mdb.greenplum.v1.Access.data_transfer
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
								Computed: true,
							},

							"web_sql": schema.BoolAttribute{
								MarkdownDescription: "Allows SQL queries to the cluster databases from the management console.",
								Description: "Allows SQL queries to the cluster databases from the management console." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.configyandex.cloud.mdb.greenplum.v1.GreenplumConfig.accessyandex.cloud.mdb.greenplum.v1.Access.web_sql
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
								Computed: true,
							},

							"yandex_query": schema.BoolAttribute{
								MarkdownDescription: "Allow access for YandexQuery.",
								Description: "Allow access for YandexQuery." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.configyandex.cloud.mdb.greenplum.v1.GreenplumConfig.accessyandex.cloud.mdb.greenplum.v1.Access.yandex_query
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
								Computed: true,
							},
						},
						MarkdownDescription: "Access policy for external services.",
						Description: "Access policy for external services." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.configyandex.cloud.mdb.greenplum.v1.GreenplumConfig.access
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"assign_public_ip": schema.BoolAttribute{
						MarkdownDescription: "Determines whether the cluster has a public IP address.\n\n After the cluster has been created, this setting cannot be changed.",
						Description: "Determines whether the cluster has a public IP address.\n\n After the cluster has been created, this setting cannot be changed." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.configyandex.cloud.mdb.greenplum.v1.GreenplumConfig.assign_public_ip
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"backup_retain_period_days": schema.Int64Attribute{
						MarkdownDescription: "Retention policy of automated backups.",
						Description: "Retention policy of automated backups." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.configyandex.cloud.mdb.greenplum.v1.GreenplumConfig.backup_retain_period_days
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"backup_window_start": schema.StringAttribute{
						MarkdownDescription: "Time to start the daily backup, in the UTC timezone.",
						Description: "Time to start the daily backup, in the UTC timezone." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.configyandex.cloud.mdb.greenplum.v1.GreenplumConfig.backup_window_start
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"subnet_id": schema.StringAttribute{
						MarkdownDescription: "ID of the subnet the cluster belongs to. This subnet should be a part of the cloud network the cluster belongs to (see [Cluster.network_id]).",
						Description: "ID of the subnet the cluster belongs to. This subnet should be a part of the cloud network the cluster belongs to (see [Cluster.network_id])." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.configyandex.cloud.mdb.greenplum.v1.GreenplumConfig.subnet_id
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"version": schema.StringAttribute{
						MarkdownDescription: "Version of the Greenplum® server software.",
						Description: "Version of the Greenplum® server software." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.configyandex.cloud.mdb.greenplum.v1.GreenplumConfig.version
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"zone_id": schema.StringAttribute{
						MarkdownDescription: "ID of the availability zone the cluster belongs to.\n To get a list of available zones, use the [yandex.cloud.compute.v1.ZoneService.List] request.",
						Description: "ID of the availability zone the cluster belongs to.\n To get a list of available zones, use the [yandex.cloud.compute.v1.ZoneService.List] request." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.configyandex.cloud.mdb.greenplum.v1.GreenplumConfig.zone_id
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},
				},
				MarkdownDescription: "Greenplum® cluster configuration.",
				Description: "Greenplum® cluster configuration." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.config
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"created_at": schema.StringAttribute{
				MarkdownDescription: "Time when the cluster was created.",
				Description: "Time when the cluster was created." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.created_at
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: "Determines whether the cluster is protected from being deleted.",
				Description: "Determines whether the cluster is protected from being deleted." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.deletion_protection
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the Greenplum® cluster.",
				Description: "Description of the Greenplum® cluster." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.description
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"environment": schema.StringAttribute{
				MarkdownDescription: "Deployment environment of the Greenplum® cluster.",
				Description: "Deployment environment of the Greenplum® cluster." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.environment
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"folder_id": schema.StringAttribute{
				MarkdownDescription: "ID of the folder that the Greenplum® cluster belongs to.",
				Description: "ID of the folder that the Greenplum® cluster belongs to." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.folder_id
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Optional: true,
				Computed: true,
			},

			"host_group_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Host groups hosting VMs of the cluster.",
				Description: "Host groups hosting VMs of the cluster." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.host_group_ids
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Custom labels for the Greenplum® cluster as `key:value` pairs. Maximum 64 labels per resource.",
				Description: "Custom labels for the Greenplum® cluster as `key:value` pairs. Maximum 64 labels per resource." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.labels
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"logging": schema.SingleNestedAttribute{

				Attributes: map[string]schema.Attribute{

					"command_center_enabled": schema.BoolAttribute{
						MarkdownDescription: "send Yandex Command Center logs",
						Description: "send Yandex Command Center logs" +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.loggingyandex.cloud.mdb.greenplum.v1.LoggingConfig.command_center_enabled
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"enabled": schema.BoolAttribute{
						MarkdownDescription: "",
						Description: "" +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.loggingyandex.cloud.mdb.greenplum.v1.LoggingConfig.enabled
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"folder_id": schema.StringAttribute{
						MarkdownDescription: "",
						Description: "" +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.loggingyandex.cloud.mdb.greenplum.v1.LoggingConfig.folder_id
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Optional: true,
						Computed: true,

						Validators: []validator.String{
							stringvalidator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("log_group_id"),
							), stringvalidator.RegexMatches(regexp.MustCompile("^(([a-zA-Z][-a-zA-Z0-9_.]{0,63})?)$"), "error validating regexp"),
						},
					},

					"greenplum_enabled": schema.BoolAttribute{
						MarkdownDescription: "send Greenplum logs",
						Description: "send Greenplum logs" +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.loggingyandex.cloud.mdb.greenplum.v1.LoggingConfig.greenplum_enabled
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"log_group_id": schema.StringAttribute{
						MarkdownDescription: "",
						Description: "" +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.loggingyandex.cloud.mdb.greenplum.v1.LoggingConfig.log_group_id
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},

					"pooler_enabled": schema.BoolAttribute{
						MarkdownDescription: "send Pooler logs",
						Description: "send Pooler logs" +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.loggingyandex.cloud.mdb.greenplum.v1.LoggingConfig.pooler_enabled
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
						Computed: true,
					},
				},
				MarkdownDescription: "Cloud logging configuration",
				Description: "Cloud logging configuration" +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.logging
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"maintenance_window": schema.SingleNestedAttribute{

				Attributes: map[string]schema.Attribute{

					"anytime": schema.SingleNestedAttribute{

						MarkdownDescription: "An any-time maintenance window.",
						Description: "An any-time maintenance window." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.maintenance_windowyandex.cloud.mdb.greenplum.v1.MaintenanceWindow.anytime
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/maintenance.proto\n",
						Computed: true,
					},

					"weekly_maintenance_window": schema.SingleNestedAttribute{

						Attributes: map[string]schema.Attribute{

							"day": schema.StringAttribute{
								MarkdownDescription: "Day of the week.",
								Description: "Day of the week." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.maintenance_windowyandex.cloud.mdb.greenplum.v1.MaintenanceWindow.weekly_maintenance_windowyandex.cloud.mdb.greenplum.v1.WeeklyMaintenanceWindow.day
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/maintenance.proto\n",
								Computed: true,
							},

							"hour": schema.Int64Attribute{
								MarkdownDescription: "Hour of the day in the UTC timezone.",
								Description: "Hour of the day in the UTC timezone." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.maintenance_windowyandex.cloud.mdb.greenplum.v1.MaintenanceWindow.weekly_maintenance_windowyandex.cloud.mdb.greenplum.v1.WeeklyMaintenanceWindow.hour
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/maintenance.proto\n",
								Computed: true,
							},
						},
						MarkdownDescription: "A weekly maintenance window.",
						Description: "A weekly maintenance window." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.maintenance_windowyandex.cloud.mdb.greenplum.v1.MaintenanceWindow.weekly_maintenance_window
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/maintenance.proto\n",
						Computed: true,
					},
				},
				MarkdownDescription: "A Greenplum® cluster maintenance window. Should be defined by either one of the two options.",
				Description: "A Greenplum® cluster maintenance window. Should be defined by either one of the two options." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.maintenance_window
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"master_config": schema.SingleNestedAttribute{

				Attributes: map[string]schema.Attribute{

					"resources": schema.SingleNestedAttribute{

						Attributes: map[string]schema.Attribute{

							"disk_size": schema.Int64Attribute{
								MarkdownDescription: "Volume of the storage used by the host, in bytes.",
								Description: "Volume of the storage used by the host, in bytes." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.master_configyandex.cloud.mdb.greenplum.v1.MasterSubclusterConfig.resourcesyandex.cloud.mdb.greenplum.v1.Resources.disk_size
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"disk_type_id": schema.StringAttribute{
								MarkdownDescription: "Type of the storage used by the host: `network-hdd`, `network-ssd` or `local-ssd`.",
								Description: "Type of the storage used by the host: `network-hdd`, `network-ssd` or `local-ssd`." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.master_configyandex.cloud.mdb.greenplum.v1.MasterSubclusterConfig.resourcesyandex.cloud.mdb.greenplum.v1.Resources.disk_type_id
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"resource_preset_id": schema.StringAttribute{
								MarkdownDescription: "ID of the preset for computational resources allocated to a host.\n\n Available presets are listed in the [documentation](/docs/managed-greenplum/concepts/instance-types).",
								Description: "ID of the preset for computational resources allocated to a host.\n\n Available presets are listed in the [documentation](/docs/managed-greenplum/concepts/instance-types)." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.master_configyandex.cloud.mdb.greenplum.v1.MasterSubclusterConfig.resourcesyandex.cloud.mdb.greenplum.v1.Resources.resource_preset_id
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},
						},
						MarkdownDescription: "Computational resources allocated to Greenplum® master subcluster hosts.",
						Description: "Computational resources allocated to Greenplum® master subcluster hosts." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.master_configyandex.cloud.mdb.greenplum.v1.MasterSubclusterConfig.resources
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
						Computed: true,
					},
				},
				MarkdownDescription: "Configuration of the Greenplum® master subcluster.",
				Description: "Configuration of the Greenplum® master subcluster." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.master_config
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"master_host_count": schema.Int64Attribute{
				MarkdownDescription: "Number of hosts in the master subcluster.",
				Description: "Number of hosts in the master subcluster." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.master_host_count
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"master_host_group_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Host groups hosting VMs of the master subcluster.",
				Description: "Host groups hosting VMs of the master subcluster." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.master_host_group_ids
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"monitoring": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{

					Attributes: map[string]schema.Attribute{

						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the monitoring system.",
							Description: "Description of the monitoring system." +
								// proto paths: +
								// -> yandex.cloud.mdb.greenplum.v1.Cluster.monitoringyandex.cloud.mdb.greenplum.v1.Monitoring.description
								"package: yandex.cloud.mdb.greenplum.v1\n" +
								"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
							Computed: true,
						},

						"link": schema.StringAttribute{
							MarkdownDescription: "Link to the monitoring system charts for the Greenplum® cluster.",
							Description: "Link to the monitoring system charts for the Greenplum® cluster." +
								// proto paths: +
								// -> yandex.cloud.mdb.greenplum.v1.Cluster.monitoringyandex.cloud.mdb.greenplum.v1.Monitoring.link
								"package: yandex.cloud.mdb.greenplum.v1\n" +
								"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
							Computed: true,
						},

						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the monitoring system.",
							Description: "Name of the monitoring system." +
								// proto paths: +
								// -> yandex.cloud.mdb.greenplum.v1.Cluster.monitoringyandex.cloud.mdb.greenplum.v1.Monitoring.name
								"package: yandex.cloud.mdb.greenplum.v1\n" +
								"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
							Computed: true,
						},
					},
				},
				MarkdownDescription: "Description of monitoring systems relevant to the Greenplum® cluster.",
				Description: "Description of monitoring systems relevant to the Greenplum® cluster." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.monitoring
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Greenplum® cluster.\n The name is unique within the folder.",
				Description: "Name of the Greenplum® cluster.\n The name is unique within the folder." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.name
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"network_id": schema.StringAttribute{
				MarkdownDescription: "ID of the cloud network that the cluster belongs to.",
				Description: "ID of the cloud network that the cluster belongs to." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.network_id
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"planned_operation": schema.SingleNestedAttribute{

				Attributes: map[string]schema.Attribute{

					"delayed_until": schema.StringAttribute{
						MarkdownDescription: "Delay time for the maintenance operation.",
						Description: "Delay time for the maintenance operation." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.planned_operationyandex.cloud.mdb.greenplum.v1.MaintenanceOperation.delayed_until
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/maintenance.proto\n",
						Computed: true,
					},

					"info": schema.StringAttribute{
						MarkdownDescription: "The description of the operation.",
						Description: "The description of the operation." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.planned_operationyandex.cloud.mdb.greenplum.v1.MaintenanceOperation.info
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/maintenance.proto\n",
						Computed: true,
					},
				},
				MarkdownDescription: "Maintenance operation planned at nearest [maintenance_window].",
				Description: "Maintenance operation planned at nearest [maintenance_window]." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.planned_operation
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"security_group_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "User security groups.",
				Description: "User security groups." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.security_group_ids
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"segment_config": schema.SingleNestedAttribute{

				Attributes: map[string]schema.Attribute{

					"resources": schema.SingleNestedAttribute{

						Attributes: map[string]schema.Attribute{

							"disk_size": schema.Int64Attribute{
								MarkdownDescription: "Volume of the storage used by the host, in bytes.",
								Description: "Volume of the storage used by the host, in bytes." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.segment_configyandex.cloud.mdb.greenplum.v1.SegmentSubclusterConfig.resourcesyandex.cloud.mdb.greenplum.v1.Resources.disk_size
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"disk_type_id": schema.StringAttribute{
								MarkdownDescription: "Type of the storage used by the host: `network-hdd`, `network-ssd` or `local-ssd`.",
								Description: "Type of the storage used by the host: `network-hdd`, `network-ssd` or `local-ssd`." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.segment_configyandex.cloud.mdb.greenplum.v1.SegmentSubclusterConfig.resourcesyandex.cloud.mdb.greenplum.v1.Resources.disk_type_id
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},

							"resource_preset_id": schema.StringAttribute{
								MarkdownDescription: "ID of the preset for computational resources allocated to a host.\n\n Available presets are listed in the [documentation](/docs/managed-greenplum/concepts/instance-types).",
								Description: "ID of the preset for computational resources allocated to a host.\n\n Available presets are listed in the [documentation](/docs/managed-greenplum/concepts/instance-types)." +
									// proto paths: +
									// -> yandex.cloud.mdb.greenplum.v1.Cluster.segment_configyandex.cloud.mdb.greenplum.v1.SegmentSubclusterConfig.resourcesyandex.cloud.mdb.greenplum.v1.Resources.resource_preset_id
									"package: yandex.cloud.mdb.greenplum.v1\n" +
									"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
								Computed: true,
							},
						},
						MarkdownDescription: "Computational resources allocated to Greenplum® segment subcluster hosts.",
						Description: "Computational resources allocated to Greenplum® segment subcluster hosts." +
							// proto paths: +
							// -> yandex.cloud.mdb.greenplum.v1.Cluster.segment_configyandex.cloud.mdb.greenplum.v1.SegmentSubclusterConfig.resources
							"package: yandex.cloud.mdb.greenplum.v1\n" +
							"filename: yandex/cloud/mdb/greenplum/v1/config.proto\n",
						Computed: true,
					},
				},
				MarkdownDescription: "Configuration of the Greenplum® segment subcluster.",
				Description: "Configuration of the Greenplum® segment subcluster." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.segment_config
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"segment_host_count": schema.Int64Attribute{
				MarkdownDescription: "Number of hosts in the segment subcluster.",
				Description: "Number of hosts in the segment subcluster." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.segment_host_count
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"segment_host_group_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Host groups hosting VMs of the segment subcluster.",
				Description: "Host groups hosting VMs of the segment subcluster." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.segment_host_group_ids
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"segment_in_host": schema.Int64Attribute{
				MarkdownDescription: "Number of segments per host.",
				Description: "Number of segments per host." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.segment_in_host
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},

			"service_account_id": schema.StringAttribute{
				MarkdownDescription: "Service account that will be used to access a Yandex Cloud resources",
				Description: "Service account that will be used to access a Yandex Cloud resources" +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.service_account_id
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Optional: true,
				Computed: true,
			},

			"user_name": schema.StringAttribute{
				MarkdownDescription: "Owner user name.",
				Description: "Owner user name." +
					// proto paths: +
					// -> yandex.cloud.mdb.greenplum.v1.Cluster.user_name
					"package: yandex.cloud.mdb.greenplum.v1\n" +
					"filename: yandex/cloud/mdb/greenplum/v1/cluster.proto\n",
				Computed: true,
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},

		Blocks: map[string]schema.Block{},
	}
}
