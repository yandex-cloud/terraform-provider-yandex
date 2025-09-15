# CDN Provider Improvements

## Missing Features

### 1. Add Stale Content Support
**Problem**: The `Stale` option is not implemented in the provider, but exists in the API.

**Solution**: Add to schema and implementation:
```go
// In schema
"stale": {
    Type:        schema.TypeList,
    Description: "List of errors to serve stale content",
    Optional:    true,
    Elem: &schema.Schema{
        Type: schema.TypeString,
        ValidateFunc: validation.StringInSlice([]string{
            "error", "http_403", "http_404", "http_429", 
            "http_500", "http_502", "http_503", "http_504",
            "invalid_header", "timeout", "updating",
        }, false),
    },
},

// In expandCDNResourceOptions
if rawOption, ok := d.GetOk("options.0.stale"); ok {
    optionsSet = true
    var values []string
    for _, v := range rawOption.([]interface{}) {
        values = append(values, v.(string))
    }
    if len(values) != 0 {
        result.Stale = &cdn.ResourceOptions_StringsListOption{
            Enabled: true,
            Value:   values,
        }
    }
}

// In flattenCDNResourceOptions
if options.Stale != nil {
    setIfEnabled("stale", options.Stale.Enabled, options.Stale.Value)
}
```

### 2. Add Provider Type to Resources
**Problem**: `provider_type` field is computed but never populated.

**Solution**: Read from API response:
```go
// In flattenYandexCDNResource
_ = d.Set("provider_type", resource.ProviderType)

// In flattenYandexCDNOriginGroup  
_ = d.Set("provider_type", originGroup.ProviderType)
```

### 3. Add New CDN Services

#### Cache Service
```go
// New resource: yandex_cdn_cache_purge
resource "yandex_cdn_cache_purge" "example" {
  resource_id = yandex_cdn_resource.example.id
  paths       = ["/path/to/purge/*"]
}
```

#### Raw Logs Service
```go
// New resource: yandex_cdn_raw_logs
resource "yandex_cdn_raw_logs" "example" {
  resource_id = yandex_cdn_resource.example.id
  bucket_name = "my-logs-bucket"
  prefix      = "cdn-logs/"
  status      = "Activated"
}
```

#### Rules Service - DETAILED DESIGN

**API Structure Analysis:**
- Rule ID: int64 (stored as string in Terraform to avoid precision loss)
- Resource ID: string (reference to CDN resource)
- Name: string (max 50 chars)
- Rule Pattern: string (regex, max 100 chars)
- Weight: int64 (0-9999, controls execution order)
- Options: ResourceOptions (same as CDN resource options)

**Terraform Schema Design:**
```go
// New resource: yandex_cdn_rule
resource "yandex_cdn_rule" "example" {
  resource_id   = yandex_cdn_resource.example.id  # Required
  name          = "redirect-rule"                 # Required, max 50 chars
  rule_pattern  = "/old-path/*"                  # Required, regex, max 100 chars
  weight        = 100                            # Optional, 0-9999, default 0
  
  # Options block - same structure as yandex_cdn_resource
  options {
    redirect_http_to_https = true
    custom_host_header     = "example.com"
    # All other CDN resource options are supported
  }
}
```

**Implementation Approach:**
1. Reuse `cdn_resource_validators.go` for options validation
2. Use string type for rule_id to avoid int64 precision loss
3. Add regex validation for rule_pattern
4. Implement weight validation (0-9999)
5. Share options schema with yandex_cdn_resource
```

#### Provider Service
```go
// New resource: yandex_cdn_provider
resource "yandex_cdn_provider" "default" {
  folder_id = "folder-id"
  type      = "gcore"  # or other provider types
  activated = true
}
```

### 4. Fix Data Type Issues

#### Origin Group ID as String
Already fixed in recent commits - using TypeString to avoid precision loss.

### 5. Add Resources Metadata to Origin Group
**Problem**: Can't see which CDN resources use an origin group.

**Solution**: Add computed field:
```go
"resources_metadata": {
    Type:        schema.TypeList,
    Description: "CDN resources using this origin group",
    Computed:    true,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "resource_id": {
                Type:     schema.TypeString,
                Computed: true,
            },
            "cname": {
                Type:     schema.TypeString,
                Computed: true,
            },
        },
    },
},
```

### 6. Add Shielding Support
```go
// New resource: yandex_cdn_shielding
resource "yandex_cdn_shielding" "example" {
  resource_id = yandex_cdn_resource.example.id
  enabled     = true
  location    = "eu-central"
}
```

## Code Quality Improvements

### 1. Extract Common Constants
```go
// cdn_constants.go
package yandex

const (
    // Stale error types
    CDNStaleError           = "error"
    CDNStaleHTTP403         = "http_403"
    CDNStaleHTTP404         = "http_404"
    CDNStaleHTTP429         = "http_429"
    CDNStaleHTTP500         = "http_500"
    CDNStaleHTTP502         = "http_502"
    CDNStaleHTTP503         = "http_503"
    CDNStaleHTTP504         = "http_504"
    CDNStaleInvalidHeader   = "invalid_header"
    CDNStaleTimeout         = "timeout"
    CDNStaleUpdating        = "updating"
)

var CDNStaleErrors = []string{
    CDNStaleError,
    CDNStaleHTTP403,
    CDNStaleHTTP404,
    CDNStaleHTTP429,
    CDNStaleHTTP500,
    CDNStaleHTTP502,
    CDNStaleHTTP503,
    CDNStaleHTTP504,
    CDNStaleInvalidHeader,
    CDNStaleTimeout,
    CDNStaleUpdating,
}
```

### 2. Better Error Handling
```go
func resourceYandexCDNResourceCreate(d *schema.ResourceData, meta interface{}) error {
    // ...
    op, err := sdk.WrapOperation(op, err)
    if err != nil {
        return fmt.Errorf("error creating CDN resource: %w", err)
    }
    
    err = op.Wait(ctx)
    if err != nil {
        return fmt.Errorf("error waiting for CDN resource creation: %w", err)
    }
    
    response, err := op.Response()
    if err != nil {
        return fmt.Errorf("error getting CDN resource creation response: %w", err)
    }
    // ...
}
```

### 3. Add Import Tests
```go
func TestAccCDNResource_import(t *testing.T) {
    resourceName := "yandex_cdn_resource.test"
    
    resource.Test(t, resource.TestCase{
        PreCheck:     func() { testAccPreCheck(t) },
        Providers:    testAccProviders,
        CheckDestroy: testAccCheckCDNResourceDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccCDNResource_basic(),
            },
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
                ImportStateVerifyIgnore: []string{
                    "origin_group_id",  // Can be derived from origin_group_name
                },
            },
        },
    })
}
```

## Testing Improvements

### 1. Add Validation Tests
```go
func TestCDNResourceValidation(t *testing.T) {
    cases := []struct {
        name      string
        options   map[string]interface{}
        expectErr bool
    }{
        {
            name: "valid stale options",
            options: map[string]interface{}{
                "stale": []string{"http_404", "http_500"},
            },
            expectErr: false,
        },
        {
            name: "invalid stale option",
            options: map[string]interface{}{
                "stale": []string{"invalid_error"},
            },
            expectErr: true,
        },
        // ... more test cases
    }
    
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            // Test validation logic
        })
    }
}
```

### 2. Add Integration Tests for New Features
```go
func TestAccCDNResource_staleContent(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:  func() { testAccPreCheck(t) },
        Providers: testAccProviders,
        Steps: []resource.TestStep{
            {
                Config: testAccCDNResource_withStale(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr(
                        "yandex_cdn_resource.test",
                        "options.0.stale.#", "3",
                    ),
                    resource.TestCheckResourceAttr(
                        "yandex_cdn_resource.test",
                        "options.0.stale.0", "http_404",
                    ),
                ),
            },
        },
    })
}
```