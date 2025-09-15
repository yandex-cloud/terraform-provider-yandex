# Test configuration for CDN rewrite functionality
terraform {
  required_providers {
    yandex = {
      source = "yandex-cloud/yandex"
    }
  }
}

provider "yandex" {
  # Provider configuration
}

# Test resource with rewrite
resource "yandex_cdn_resource" "test_rewrite" {
  cname = "test-rewrite.example.com"
  
  origin_group_id = "123456" # Replace with actual origin group ID
  
  options {
    # Test rewrite rule
    rewrite {
      enabled = true
      body    = "^/(.*)/$ /$1/index.html"
      flag    = "break"
    }
    
    edge_cache_settings = 3600
  }
}

output "rewrite_config" {
  value = yandex_cdn_resource.test_rewrite.options[0].rewrite
  description = "Current rewrite configuration from state"
}