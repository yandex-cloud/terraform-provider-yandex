package descriptions

const (
	// Main schema descriptions
	Datasource = "Use this data source to get information about a Yandex Managed OpenSearch cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-opensearch/concepts)."
	Resource   = "Manages a OpenSearch cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-opensearch/concepts)."

	// Configuration blocks
	Config     = "Configuration of the OpenSearch cluster."
	Opensearch = "Configuration for OpenSearch node groups."
	Dashboards = "Configuration for Dashboards node groups."

	// Configuration attributes
	Version       = "Version of OpenSearch."
	AdminPassword = "Password for admin user of OpenSearch."
	Plugins       = "A set of requested OpenSearch plugins."

	// Node group attributes
	NodeGroups          = "A set of named OpenSearch node group configurations."
	DashboardNodeGroups = "A set of named Dashboard node group configurations."
	NodeGroupName       = "Name of OpenSearch node group."
	HostsCount          = "Number of hosts in this node group."
	ZoneIDs             = "A set of availability zones where hosts of node group may be allocated."
	SubnetIDs           = "A set of the subnets, to which the hosts belongs. The subnets must be a part of the network to which the cluster belongs."
	AssignPublicIP      = "Sets whether the hosts should get a public IP address."
	Roles               = "A set of OpenSearch roles assigned to hosts. Available roles are: `DATA`, `MANAGER`. Default: [`DATA`, `MANAGER`]."

	// Access attributes
	Access       = "Enable access to the Yandex Cloud services."
	DataTransfer = "Enable access to the [Data Transfer](https://yandex.cloud/docs/data-transfer) service."
	Serverless   = "Enable access to the [Cloud Functions](https://yandex.cloud/docs/functions) service."

	// Maintenance window attributes
	MaintenanceWindow = "Maintenance window for the cluster."
	MaintenanceType   = "Type of maintenance window. Can be either `ANYTIME` or `WEEKLY`."
	MaintenanceDay    = "Day of the week for `WEEKLY` maintenance. Can be one of `MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`."
	MaintenanceHour   = "Hour of the day in UTC for maintenance (1-24)."

	// Main cluster attributes
	ClusterID        = "The ID of the OpenSearch cluster that the resource belongs to."
	Name             = "Name of the OpenSearch cluster. The name must be unique within the folder."
	Environment      = "Deployment environment of the OpenSearch cluster. Can be either `PRESTABLE` or `PRODUCTION`. Default: `PRODUCTION`. **It is not possible to change this value after cluster creation**."
	NetworkID        = "ID of the network, to which the OpenSearch cluster belongs. It is not possible to change this value after cluster creation."
	Health           = "Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`. For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-opensearch/api-ref/Cluster/)."
	Status           = " Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`. For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-opensearch/api-ref/Cluster/)."
	SecurityGroupIDs = "A set of security groups IDs which assigned to hosts of the cluster."
	ServiceAccountID = "ID of the service account authorized for this cluster."

	// SAML settings
	AuthSettings               = "Authentication settings for Dashboards."
	SAML                       = "SAML authentication options."
	SAMLEnabled                = "Enables SAML authentication."
	SAMLIdpEntityID            = "ID of the SAML Identity Provider."
	SAMLIdpMetadataFileContent = "Metadata file content of the SAML Identity Provider. You can either put file content manually or use [`file` function](https://developer.hashicorp.com/terraform/language/functions/file)"
	SAMLSpEntityID             = "Service provider entity ID."
	SAMLDashboardsURL          = "Dashboards URL."
	SAMLRolesKey               = "Roles key."
	SAMLSubjectKey             = "Subject key."
)
