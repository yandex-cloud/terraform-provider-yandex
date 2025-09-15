# CloudRouter Implementation Plan for Terraform Provider

## Overview
CloudRouter is a Yandex Cloud service for creating network routing topologies between VPC networks and private connections. The API exists in go-genproto but is not yet available in the vendor directory.

## Current Status
- ❌ CloudRouter API is NOT in vendor (go-genproto v0.25.0)
- ❌ No CloudRouter resources in Terraform provider
- ✅ API exists in upstream: https://github.com/yandex-cloud/go-genproto/tree/master/yandex/cloud/cloudrouter/v1

## Required Steps

### 1. Update Dependencies
First, need to check if a newer version of go-genproto includes CloudRouter:
```bash
# Check latest version
go list -m -versions github.com/yandex-cloud/go-genproto@latest

# If CloudRouter is in a newer version, update:
go get github.com/yandex-cloud/go-genproto@latest
go mod vendor
```

### 2. Verify CloudRouter API Structure
Based on GitHub repository, CloudRouter API includes:
- `RoutingInstance` - main resource for routing configuration
- Operations for managing routing instances
- Prefix management
- Private connection management

## Proposed Terraform Resources

### 1. `yandex_cloudrouter_routing_instance`
Main resource for creating and managing routing instances.

```hcl
resource "yandex_cloudrouter_routing_instance" "main" {
  name        = "my-routing-instance"
  description = "Routing between VPCs"
  folder_id   = var.folder_id
  
  labels = {
    environment = "production"
  }

  # VPC networks to include
  vpc_networks = [
    yandex_vpc_network.network1.id,
    yandex_vpc_network.network2.id,
  ]
  
  # IP prefixes configuration
  ip_prefixes {
    prefix      = "10.0.0.0/16"
    description = "Network 1 prefix"
  }
  
  ip_prefixes {
    prefix      = "10.1.0.0/16"
    description = "Network 2 prefix"
  }

  # Optional: Private connections (for hybrid cloud)
  private_connection {
    connection_id = var.cic_connection_id
    prefixes      = ["192.168.0.0/24"]
  }
}
```

### 2. `data.yandex_cloudrouter_routing_instance`
Data source for existing routing instances.

```hcl
# By ID
data "yandex_cloudrouter_routing_instance" "by_id" {
  routing_instance_id = "abc123"
}

# By VPC network
data "yandex_cloudrouter_routing_instance" "by_vpc" {
  vpc_network_id = yandex_vpc_network.main.id
}

# By private connection
data "yandex_cloudrouter_routing_instance" "by_connection" {
  cic_private_connection_id = var.connection_id
}
```

## Implementation Details

### Schema Definition
```go
func resourceYandexCloudRouterRoutingInstance() *schema.Resource {
    return &schema.Resource{
        Create: resourceYandexCloudRouterRoutingInstanceCreate,
        Read:   resourceYandexCloudRouterRoutingInstanceRead,
        Update: resourceYandexCloudRouterRoutingInstanceUpdate,
        Delete: resourceYandexCloudRouterRoutingInstanceDelete,
        
        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },
        
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
                Optional: true,
                Computed: true,
            },
            "labels": {
                Type:     schema.TypeMap,
                Optional: true,
                Elem:     &schema.Schema{Type: schema.TypeString},
            },
            "vpc_networks": {
                Type:     schema.TypeSet,
                Optional: true,
                Elem:     &schema.Schema{Type: schema.TypeString},
            },
            "ip_prefixes": {
                Type:     schema.TypeList,
                Optional: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "prefix": {
                            Type:         schema.TypeString,
                            Required:     true,
                            ValidateFunc: validation.IsCIDR,
                        },
                        "description": {
                            Type:     schema.TypeString,
                            Optional: true,
                        },
                    },
                },
            },
            "private_connection": {
                Type:     schema.TypeList,
                Optional: true,
                MaxItems: 1,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "connection_id": {
                            Type:     schema.TypeString,
                            Required: true,
                        },
                        "prefixes": {
                            Type:     schema.TypeSet,
                            Required: true,
                            Elem:     &schema.Schema{Type: schema.TypeString},
                        },
                    },
                },
            },
            // Computed fields
            "created_at": {
                Type:     schema.TypeString,
                Computed: true,
            },
            "status": {
                Type:     schema.TypeString,
                Computed: true,
            },
        },
    }
}
```

## Use Cases

### 1. VPC Stitching
Connect multiple VPC networks within same organization:
```hcl
resource "yandex_cloudrouter_routing_instance" "vpc_stitching" {
  name = "multi-vpc-routing"
  
  vpc_networks = [
    yandex_vpc_network.prod.id,
    yandex_vpc_network.dev.id,
    yandex_vpc_network.staging.id,
  ]
  
  # Automatically handles routing between all VPCs
}
```

### 2. Hybrid Cloud Connection
Connect on-premises network to cloud VPCs:
```hcl
resource "yandex_cloudrouter_routing_instance" "hybrid" {
  name = "hybrid-cloud-routing"
  
  vpc_networks = [yandex_vpc_network.main.id]
  
  private_connection {
    connection_id = var.interconnect_connection_id
    prefixes      = ["10.0.0.0/8"]  # On-prem network
  }
  
  ip_prefixes {
    prefix = "172.16.0.0/12"  # Cloud network
  }
}
```

### 3. Multi-Region Routing
Route between regions (when supported):
```hcl
resource "yandex_cloudrouter_routing_instance" "multi_region" {
  name = "cross-region-routing"
  
  vpc_networks = [
    var.moscow_vpc_id,
    var.petersburg_vpc_id,
  ]
  
  # Routing rules for cross-region traffic
}
```

## Testing Requirements

1. **Unit Tests:**
   - Schema validation
   - CRUD operations mock
   - Error handling

2. **Integration Tests:**
   - Create routing instance with single VPC
   - Create routing instance with multiple VPCs
   - Update prefixes
   - Add/remove private connections
   - Import existing routing instance

3. **Acceptance Tests:**
   ```go
   func TestAccCloudRouterRoutingInstance_basic(t *testing.T) {
       // Test basic routing instance creation
   }
   
   func TestAccCloudRouterRoutingInstance_vpcStitching(t *testing.T) {
       // Test multiple VPC connection
   }
   
   func TestAccCloudRouterRoutingInstance_privateConnection(t *testing.T) {
       // Test with CIC private connection
   }
   ```

## Documentation Requirements

1. **Resource Documentation:**
   - Complete schema reference
   - Multiple usage examples
   - Import instructions
   - Known limitations

2. **Data Source Documentation:**
   - Query methods (by ID, by VPC, by connection)
   - Output attributes
   - Usage examples

3. **Guides:**
   - VPC Stitching guide
   - Hybrid cloud setup guide
   - Migration from manual routing

## Dependencies

1. **Required Services:**
   - VPC service for network management
   - Interconnect service for private connections (optional)
   - IAM for access control

2. **Permissions:**
   - `cloudrouter.editor` role for resource management
   - `vpc.viewer` for VPC network access
   - `interconnect.viewer` for private connection access

## Timeline Estimate

1. **Week 1:** Update dependencies, study API
2. **Week 2:** Implement basic CRUD operations
3. **Week 3:** Add advanced features (prefixes, connections)
4. **Week 4:** Testing and documentation

## Blockers

1. **API Availability:** CloudRouter API must be included in vendor
2. **Service Activation:** CloudRouter must be GA or available for testing
3. **Documentation:** Need official API documentation for CloudRouter

## Next Steps

1. Contact Yandex Cloud team about CloudRouter API availability
2. Check if newer go-genproto version includes CloudRouter
3. Request access to CloudRouter service for testing
4. Start implementation once API is available