package yandex

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/structs"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

const albDefaultSni = "tf-test-tls"
const albDefaultValidationContext = `-----BEGIN CERTIFICATE-----
MIIBpzCCAVGgAwIBAgIJAMttzZ34ksJIMA0GCSqGSIb3DQEBCwUAMC8xLTArBgNV
BAMMJGVkYmY4NzlhLWJmMDEtNGI5Yi05YjBmLTgyNDhiZWE3OTZiMTAeFw0yMDAy
MTgxMjAyMTFaFw0yMDAzMTkxMjAyMTFaMC8xLTArBgNVBAMMJGVkYmY4NzlhLWJm
MDEtNGI5Yi05YjBmLTgyNDhiZWE3OTZiMTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgC
QQDyxRijt3T5/HpPkFmo4DmrPEL3IHbqMedSwmcvYjEhex43qGLsAXC17e7tKpQE
VDYmdvJCE6T7AfezNWLc95JRAgMBAAGjUDBOMB0GA1UdDgQWBBRIq4vrr+4b//NF
PR2lXBPTWewVYDAfBgNVHSMEGDAWgBRIq4vrr+4b//NFPR2lXBPTWewVYDAMBgNV
HRMEBTADAQH/MA0GCSqGSIb3DQEBCwUAA0EARRiU9hEq7k9Sa2tbPF7lI9xxknjZ
D0M/nOBnNGaGBKG4hNAb5KMUSfrF6Jn6lp0yNIz+LNWNJQVOjZFiw2rM/g==
-----END CERTIFICATE-----`
const albDefaultBackendWeight = "1"
const albDefaultPanicThreshold = "50"
const albDefaultLoadBalancingMode = "LEAST_REQUEST"
const albDefaultLocalityPercent = "35"
const albDefaultTimeout = "3s"
const albDefaultInterval = "5s"
const albDefaultStrictLocality = "true"
const albDefaultServiceName = "true"
const albDefaultHTTP2 = "true"
const albDefaultHost = "tf-test-host"
const albDefaultPath = "/tf-test-path/"
const albDefaultPort = "3"
const albDefaultSendText = "tf-test-send"
const albDefaultReceiveText = "tf-test-receive"
const albDefaultDescription = "alb-bg-description"
const albDefaultDirectResponseBody = "Not Found"
const albDefaultDirectResponseStatus = "404"
const albDefaultModificationAppend = "header"
const albDefaultStatusResponse = "not_found"
const albDefaultRedirectResponseCode = "moved_permanently"
const albDefaultAutoHostRewrite = "true"
const albDefaultAllowHTTP10 = "true"
const albDefaultRewriteRequestID = "true"
const albDefaultMaxConcurrentStreams = "2"
const albDefaultHTTPToHTTPS = "true"
const albDefaultProxyProtocol = "false"
const albDefaultHeaderAffinity = "x-some-header"
const albDefaultAnyPrincipal = "true"
const albDefaultRemoteIP = "127.0.0.1/16"
const albDefaultHeaderName = "client-header"
const albDefaultHeaderValue = "client-value"
const albDefaultRBACAction = "allow"
const albDefaultIdleTimeout = "42s"
const albDefaultExpectedStatuses = "null"

type resourceALBLoadBalancerInfo struct {
	IsHTTPListener     bool
	IsStreamListener   bool
	IsTLSListener      bool
	IsRedirects        bool
	IsHTTPHandler      bool
	IsStreamHandler    bool
	IsDataSource       bool
	IsHTTP2Options     bool
	IsAllowHTTP10      bool
	IsRewriteRequestID bool
	IsLogOptions       bool
	IsIdleTimeout      bool

	BaseTemplate string

	BalancerName         string
	RouterName           string
	BackendGroupName     string
	TargetGroupName      string
	ListenerName         string
	BalancerDescription  string
	AllowHTTP10          string
	RewriteRequestID     string
	MaxConcurrentStreams string
	EndpointPort         string
	HTTPToHTTPS          string
	CertificateID        string
	IdleTimeout          string
	AllowZonalShift      bool
	AutoScalePolicy      bool
}

func albLoadBalancerInfo() resourceALBLoadBalancerInfo {
	res := resourceALBLoadBalancerInfo{
		IsHTTPListener:       false,
		IsStreamListener:     false,
		IsTLSListener:        false,
		IsDataSource:         false,
		IsRedirects:          false,
		IsHTTPHandler:        false,
		IsStreamHandler:      false,
		IsHTTP2Options:       false,
		IsAllowHTTP10:        false,
		IsRewriteRequestID:   false,
		IsIdleTimeout:        false,
		BaseTemplate:         testAccALBBaseTemplate(acctest.RandomWithPrefix("tf-instance")),
		BalancerName:         acctest.RandomWithPrefix("tf-load-balancer"),
		RouterName:           acctest.RandomWithPrefix("tf-router"),
		BackendGroupName:     acctest.RandomWithPrefix("tf-bg"),
		TargetGroupName:      acctest.RandomWithPrefix("tf-tg"),
		ListenerName:         acctest.RandomWithPrefix("tf-listener"),
		BalancerDescription:  acctest.RandomWithPrefix("tf-load-balancer-description"),
		AllowHTTP10:          albDefaultAllowHTTP10,
		RewriteRequestID:     albDefaultRewriteRequestID,
		MaxConcurrentStreams: albDefaultMaxConcurrentStreams,
		EndpointPort:         albDefaultPort,
		HTTPToHTTPS:          albDefaultHTTPToHTTPS,
		CertificateID:        os.Getenv("ALB_TEST_CERTIFICATE_ID"),
		IdleTimeout:          albDefaultIdleTimeout,
		AllowZonalShift:      false,
		AutoScalePolicy:      false,
	}

	return res
}

type resourceALBVirtualHostInfo struct {
	IsModifyRequestHeaders            bool
	IsModifyResponseHeaders           bool
	IsHTTPRoute                       bool
	IsGRPCRoute                       bool
	IsHTTPRouteAction                 bool
	IsRedirectAction                  bool
	IsDirectResponseAction            bool
	IsGRPCRouteAction                 bool
	IsGRPCStatusResponseAction        bool
	IsRouteRBAC                       bool
	IsVirtualHostRBAC                 bool
	IsDataSource                      bool
	IsHTTPRouteActionHostRewrite      bool
	IsRateLimit                       bool
	IsRateLimitAllRequests            bool
	IsRateLimitRequestsPerIP          bool
	IsHTTPRouteRateLimit              bool
	IsHTTPRouteRateLimitAllRequests   bool
	IsHTTPRouteRateLimitRequestsPerIP bool
	IsGRPCRouteRateLimit              bool
	IsGRPCRouteRateLimitAllRequests   bool
	IsGRPCRouteRateLimitRequestsPerIP bool
	IsHTTPRoutePathRegexRewrite       bool

	BaseTemplate string

	VHName     string
	TGName     string
	RouterName string
	BGName     string

	RouterDescription               string
	ModificationAppend              string
	RouteName                       string
	DirectResponseStatus            string
	DirectResponseBody              string
	RedirectResponseCode            string
	HTTPRouteActionTimeout          string
	GRPCRouteActionTimeout          string
	GRPCStatusResponseActionStatus  string
	GRPCRouteActionAutoHostRewrite  string
	HTTPRouteActionHostRewrite      string
	HTTPRouteActionAutoHostRewrite  bool
	AnyPrincipals                   string
	RemoteIP                        string
	RBACAction                      string
	RateLimitRPS                    string
	RateLimitRPM                    string
	HTTPRouteRateLimitRPS           string
	HTTPRouteRateLimitRPM           string
	GRPCRouteRateLimitRPS           string
	GRPCRouteRateLimitRPM           string
	HTTPRouteRegexRewriteRegex      string
	HTTPRouteRegexRewriteSubstitute string
}

func albVirtualHostInfo() resourceALBVirtualHostInfo {
	res := resourceALBVirtualHostInfo{
		IsModifyRequestHeaders:         false,
		IsModifyResponseHeaders:        false,
		IsHTTPRoute:                    false,
		IsGRPCRoute:                    false,
		IsHTTPRouteAction:              false,
		IsRedirectAction:               false,
		IsDirectResponseAction:         false,
		IsGRPCRouteAction:              false,
		IsGRPCStatusResponseAction:     false,
		IsRouteRBAC:                    false,
		IsVirtualHostRBAC:              false,
		IsDataSource:                   false,
		IsHTTPRouteActionHostRewrite:   false,
		BaseTemplate:                   testAccALBBaseTemplate(acctest.RandomWithPrefix("tf-instance")),
		VHName:                         acctest.RandomWithPrefix("tf-virtual-host"),
		TGName:                         acctest.RandomWithPrefix("tf-tg"),
		RouterName:                     acctest.RandomWithPrefix("tf-router"),
		BGName:                         acctest.RandomWithPrefix("tf-bg"),
		RouterDescription:              albDefaultDescription,
		RouteName:                      acctest.RandomWithPrefix("tf-route"),
		ModificationAppend:             albDefaultModificationAppend,
		DirectResponseStatus:           albDefaultDirectResponseStatus,
		DirectResponseBody:             albDefaultDirectResponseBody,
		RedirectResponseCode:           albDefaultRedirectResponseCode,
		HTTPRouteActionTimeout:         albDefaultTimeout,
		GRPCRouteActionTimeout:         albDefaultTimeout,
		GRPCStatusResponseActionStatus: albDefaultStatusResponse,
		GRPCRouteActionAutoHostRewrite: albDefaultAutoHostRewrite,
		AnyPrincipals:                  albDefaultAnyPrincipal,
		RemoteIP:                       albDefaultRemoteIP,
		RBACAction:                     albDefaultRBACAction,
		HTTPRouteActionAutoHostRewrite: false,
	}

	return res
}

type resourceALBHTTPRouterInfo struct {
	IsRBAC bool

	RouterName      string
	RBACHeaderName  string
	RBACHeaderValue string
	RBACAction      string
}

func albHTTPRouterInfo() resourceALBHTTPRouterInfo {
	res := resourceALBHTTPRouterInfo{
		IsRBAC:          false,
		RouterName:      acctest.RandomWithPrefix("tf-router"),
		RBACHeaderName:  albDefaultHeaderName,
		RBACHeaderValue: albDefaultHeaderValue,
		RBACAction:      albDefaultRBACAction,
	}

	return res
}

type resourceALBBackendGroupInfo struct {
	IsHTTPBackend                      bool
	IsGRPCBackend                      bool
	IsStreamBackend                    bool
	IsHTTPCheck                        bool
	IsGRPCCheck                        bool
	IsStreamCheck                      bool
	IsDataSource                       bool
	IsEmptyTLS                         bool
	IsStorageBackend                   bool
	UseHeaderAffinity                  bool
	KeepConnectionsOnHostHealthFailure bool

	BaseTemplate string

	TGName string
	BGName string

	BGDescription        string
	TlsSni               string
	TlsValidationContext string
	BackendWeight        string
	PanicThreshold       string
	LoadBalancingMode    string
	LocalityPercent      string
	StrictLocality       string
	Timeout              string
	Interval             string
	ServiceName          string
	HTTP2                string
	Host                 string
	Path                 string
	Port                 string
	ReceiveText          string
	SendText             string
	ProxyProtocol        string
	StorageBackendBucket string
	ExpectedStatuses     string
}

func albBackendGroupInfo() resourceALBBackendGroupInfo {
	res := resourceALBBackendGroupInfo{
		IsHTTPBackend:        false,
		IsStreamBackend:      false,
		IsGRPCBackend:        false,
		IsHTTPCheck:          false,
		IsGRPCCheck:          false,
		IsStreamCheck:        false,
		IsDataSource:         false,
		IsEmptyTLS:           false,
		IsStorageBackend:     false,
		UseHeaderAffinity:    false,
		BaseTemplate:         testAccALBBaseTemplate(acctest.RandomWithPrefix("tf-instance")),
		TGName:               acctest.RandomWithPrefix("tf-tg"),
		BGName:               acctest.RandomWithPrefix("tf-bg"),
		BGDescription:        albDefaultDescription,
		TlsSni:               albDefaultSni,
		TlsValidationContext: albDefaultValidationContext,
		BackendWeight:        albDefaultBackendWeight,
		PanicThreshold:       albDefaultPanicThreshold,
		LoadBalancingMode:    albDefaultLoadBalancingMode,
		LocalityPercent:      albDefaultLocalityPercent,
		StrictLocality:       albDefaultStrictLocality,
		Timeout:              albDefaultTimeout,
		Interval:             albDefaultInterval,
		ServiceName:          albDefaultServiceName,
		HTTP2:                albDefaultHTTP2,
		Host:                 albDefaultHost,
		Path:                 albDefaultPath,
		Port:                 albDefaultPort,
		ReceiveText:          albDefaultReceiveText,
		SendText:             albDefaultSendText,
		ProxyProtocol:        albDefaultProxyProtocol,
		ExpectedStatuses:     albDefaultExpectedStatuses,
	}

	return res
}

const albVirtualHostConfigTemplate = `
{{ if .IsDataSource }}
data "yandex_alb_virtual_host" "test-virtual-host-ds" {
  virtual_host_id = yandex_alb_virtual_host.test-vh.id
}
{{ end }}
resource "yandex_alb_http_router" "test-router" {
  name        = "{{.RouterName}}"
  description = "{{.RouterDescription}}"
}
resource "yandex_alb_backend_group" "test-bg" {
  name        = "{{.BGName}}"
  {{if .IsHTTPRoute}}
  http_backend {
    name             = "test-http-backend"
    weight           = 1
    target_group_ids = ["${yandex_alb_target_group.test-target-group.id}"]
  }
  {{end}}
  {{if .IsGRPCRoute}}
  grpc_backend {
    name             = "test-grpc-backend"
    weight           = 1
    target_group_ids = ["${yandex_alb_target_group.test-target-group.id}"]
  }
  {{end}}
}
resource "yandex_alb_virtual_host" "test-vh" {
  name        = "{{.VHName}}"
  http_router_id = yandex_alb_http_router.test-router.id

  authority = ["*.foo.com", "*-bar.foo.com"]

  modify_request_headers {
    name   = "test-header"
    append = "{{.ModificationAppend}}"
  }
  {{if .IsVirtualHostRBAC}}
  route_options {
    rbac {
      action = "{{.RBACAction}}"
        principals {
          and_principals {
            remote_ip = "{{.RemoteIP}}"
          }
        }
      }
  }
 {{end}}

  {{ if or .IsHTTPRoute .IsGRPCRoute}}
  route {
    name = "{{.RouteName}}"
    route_options {
       {{if .IsRouteRBAC}}
       rbac {
         action = "{{.RBACAction}}"
         principals {
           and_principals {
             any = "{{.AnyPrincipals}}"
           }
         }
       }
       {{end}}
    }
    {{if .IsHTTPRoute}}
    http_route {
      http_match {
        path {
          prefix = "/http/match/"
        }
        http_method = ["GET", "PUT"]
      }
      {{if .IsHTTPRouteAction}}
      http_route_action {
        backend_group_id = yandex_alb_backend_group.test-bg.id
        timeout = "{{.HTTPRouteActionTimeout}}"
        auto_host_rewrite = "{{.HTTPRouteActionAutoHostRewrite}}"
        {{if .IsHTTPRouteActionHostRewrite}}
        host_rewrite = "{{.HTTPRouteActionHostRewrite}}"
        {{end}}

        {{ if .IsHTTPRoutePathRegexRewrite }}
        regex_rewrite {
          regex       = "{{ .HTTPRouteRegexRewriteRegex }}"
          substitute  = "{{ .HTTPRouteRegexRewriteSubstitute }}"
        }
        {{ end }}

        {{if .IsHTTPRouteRateLimit}}
          rate_limit {
              {{ if .IsHTTPRouteRateLimitAllRequests }}
                all_requests {
                  {{if .HTTPRouteRateLimitRPS}}
                    per_second = {{ .HTTPRouteRateLimitRPS }}
                  {{end}}

                  {{if .HTTPRouteRateLimitRPM}}
                    per_minute = {{ .HTTPRouteRateLimitRPM }}
                  {{end}}
                }
              {{end}}

              {{if .IsHTTPRouteRateLimitRequestsPerIP}}
                requests_per_ip {
                  {{if .HTTPRouteRateLimitRPS}}
                    per_second = {{ .HTTPRouteRateLimitRPS }}
                  {{end}}

                  {{if .HTTPRouteRateLimitRPM}}
                    per_minute = {{ .HTTPRouteRateLimitRPM }}
                  {{end}}
                }
              {{end}}
          }
        {{end}}
      }
      {{end}}
      {{if .IsDirectResponseAction}}
      direct_response_action {
        status = {{.DirectResponseStatus}}
        body = "{{.DirectResponseBody}}"
      }
      {{end}}
      {{if .IsRedirectAction}}
      redirect_action {
        response_code = "{{.RedirectResponseCode}}"
      }
      {{end}}
    }
    {{end}}
    {{if .IsGRPCRoute}}
    grpc_route {
      grpc_match {
        fqmn {
          exact = "some.service"
        }
      }
      {{if .IsGRPCRouteAction}}
      grpc_route_action {
        backend_group_id = yandex_alb_backend_group.test-bg.id
        max_timeout = "{{.GRPCRouteActionTimeout}}"
        auto_host_rewrite = {{.GRPCRouteActionAutoHostRewrite}}

        {{if .IsGRPCRouteRateLimit}}
          rate_limit {
              {{ if .IsGRPCRouteRateLimitAllRequests }}
                all_requests {
                  {{if .GRPCRouteRateLimitRPS}}
                    per_second = {{ .GRPCRouteRateLimitRPS }}
                  {{end}}

                  {{if .GRPCRouteRateLimitRPM}}
                    per_minute = {{ .GRPCRouteRateLimitRPM }}
                  {{end}}
                }
              {{end}}

              {{if .IsGRPCRouteRateLimitRequestsPerIP}}
                requests_per_ip {
                  {{if .GRPCRouteRateLimitRPS}}
                    per_second = {{ .GRPCRouteRateLimitRPS }}
                  {{end}}

                  {{if .GRPCRouteRateLimitRPM}}
                    per_minute = {{ .GRPCRouteRateLimitRPM }}
                  {{end}}
                }
              {{end}}
          }
        {{end}}
      }
      {{end}}
      {{if .IsGRPCStatusResponseAction}}
      grpc_status_response_action {
        status = "{{.GRPCStatusResponseActionStatus}}"
      }
      {{end}}
    }
    {{end}}
  }
  {{end}}

  {{if .IsRateLimit}}
    rate_limit {
        {{ if .IsRateLimitAllRequests }}
          all_requests {
            {{if .RateLimitRPS}}
              per_second = {{ .RateLimitRPS }}
            {{end}}

            {{if .RateLimitRPM}}
              per_minute = {{ .RateLimitRPM }}
            {{end}}
          }
        {{end}}

        {{if .IsRateLimitRequestsPerIP}}
          requests_per_ip {
            {{if .RateLimitRPS}}
              per_second = {{ .RateLimitRPS }}
            {{end}}

            {{if .RateLimitRPM}}
              per_minute = {{ .RateLimitRPM }}
            {{end}}
          }
        {{end}}
    }
  {{end}}
}
{{ if or .IsHTTPRoute .IsGRPCRoute }}
resource "yandex_alb_target_group" "test-target-group" {
  name		= "{{.TGName}}"

  target {
	subnet_id	= "${yandex_vpc_subnet.test-subnet.id}"
	ip_address	= "${yandex_compute_instance.test-instance-1.network_interface.0.ip_address}"
  }

  target {
	subnet_id	= "${yandex_vpc_subnet.test-subnet.id}"
	ip_address	= "${yandex_compute_instance.test-instance-2.network_interface.0.ip_address}"
  }
}
{{ end }}
{{.BaseTemplate}}
`

const albHTTPRouterConfigTemplate = `
{{ if .IsDataSource }}
data "yandex_alb_http_router" "test-http-router-ds" {
  http_router_id = yandex_alb_http_router.test-vh.id
}
{{ end }}
resource "yandex_alb_http_router" "test-router" {
  name        = "{{.RouterName}}"
  description = "{{.RouterDescription}}"
  route_options {
    {{if .IsRBAC}}
    rbac {
      action = "{{.RBACAction}}"
      principals {
        and_principals {
          header {
            name = "{{.RBACHeaderName}}"
            value {
              exact = "{{.RBACHeaderValue}}"
            }
          }
        }
      }
    }
    {{end}}
  }
}
`

const albLoadBalancerConfigTemplate = `
{{ if .IsDataSource }}
data "yandex_alb_load_balancer" "test-alb-ds" {
  name = yandex_alb_load_balancer.test-balancer.name
}
{{ end }}
resource "yandex_alb_http_router" "test-router" {
  name        = "{{.RouterName}}"
}
{{ if .IsStreamHandler }}
resource "yandex_alb_backend_group" "test-bg" {
  name        = "{{.BackendGroupName}}"
  stream_backend {
    name             = "test-stream-backend"
    port             = 8080
    target_group_ids = ["${yandex_alb_target_group.test-target-group.id}"]
    load_balancing_config {
      panic_threshold                = "50"
      locality_aware_routing_percent = "35"
      strict_locality                = "true"
    }
    healthcheck {
      timeout  = "10s"
      interval = "10s"
      http_healthcheck {
        host  = "tf-test-host"
        path  = "/tf-test-path/"
        http2 = "true"
      }
    }
  }
}

resource "yandex_alb_target_group" "test-target-group" {
  name        = "{{.TargetGroupName}}"
}
{{ end }}
resource "yandex_alb_load_balancer" "test-balancer" {
  name        = "{{.BalancerName}}"
  description = "{{.BalancerDescription}}"

  allow_zonal_shift = "{{.AllowZonalShift}}"
  security_group_ids = [yandex_vpc_security_group.test-security-group.id]
  network_id  = yandex_vpc_network.test-network.id
  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
  allocation_policy {
    location {
      zone_id   = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.test-subnet.id
    }
  }
  {{ if .IsLogOptions }}
  log_options {
    disable = false
    discard_rule {
      http_codes = [300,301,307]
      http_code_intervals = ["HTTP_2XX"]
      grpc_codes = ["OK", "NOT_FOUND"]
      discard_percent = 90
    }
    discard_rule {
      http_code_intervals = ["HTTP_ALL"]
    }
  }
  {{ end }}

  {{ if .AutoScalePolicy }}
  auto_scale_policy {
  	min_zone_size = 2
    max_size = 2
  }
  {{ end }}

  {{ if or .IsHTTPListener .IsTLSListener .IsStreamListener}}
  listener {
    name = "{{.ListenerName}}"
    endpoint {
      address {
        internal_ipv4_address {
          subnet_id = yandex_vpc_subnet.test-subnet.id
        }
      }
      ports = [ {{.EndpointPort}} ]
    }
    {{if .IsHTTPListener}}
    http {
      {{if .IsHTTPHandler}}
      handler {
        http_router_id = yandex_alb_http_router.test-router.id
        {{if .IsAllowHTTP10}}
        allow_http10 = {{.AllowHTTP10}}
        {{end}}
		{{if .IsRewriteRequestID}}
		rewrite_request_id = {{.RewriteRequestID}}
		{{end}}
        {{if .IsHTTP2Options}}
        http2_options {
          max_concurrent_streams = {{.MaxConcurrentStreams}}
        }
        {{end}}
      }
      {{end}}
      {{if .IsRedirects}}
      redirects {
        http_to_https = {{.HTTPToHTTPS}}
      }
      {{end}}
    }
    {{end}}
    {{if .IsStreamListener}}
    stream {
      {{if .IsStreamHandler}}
      handler {
        backend_group_id = yandex_alb_backend_group.test-bg.id
        {{if .IsIdleTimeout}}
        idle_timeout = "{{.IdleTimeout}}"
        {{end}}
      }
      {{end}}
    }
    {{end}}
    {{if .IsTLSListener}}
    tls {
      default_handler {
        {{if .IsHTTPHandler}}
        http_handler {
          http_router_id = yandex_alb_http_router.test-router.id
          {{if .IsAllowHTTP10}}
		  allow_http10 = {{.AllowHTTP10}}
          {{end}}
		  {{if .IsRewriteRequestID}}
		  rewrite_request_id = {{.RewriteRequestID}}
		  {{end}}
          {{if .IsHTTP2Options}}
          http2_options {
            max_concurrent_streams = {{.MaxConcurrentStreams}}
          }
          {{end}}
        }
        {{end}}
        {{if .IsStreamHandler}}
        stream_handler {
          backend_group_id = yandex_alb_backend_group.test-bg.id
          {{if .IsIdleTimeout}}
          idle_timeout = "{{.IdleTimeout}}"
          {{end}}
        }
        {{end}}
        certificate_ids = ["{{.CertificateID}}"]
      }
      sni_handler {
        name = "host"
        server_names = ["host.url.com"]
        handler {
          http_handler {
            http_router_id = yandex_alb_http_router.test-router.id
			rewrite_request_id = true
            allow_http10 = true
          }
          certificate_ids = ["{{.CertificateID}}"]
        }
      }
    }
    {{end}}
  }
  {{end}}
}

resource "yandex_vpc_security_group" "test-security-group" {
  network_id = yandex_vpc_network.test-network.id

  ingress {
    protocol       = "TCP"
    description    = "healthchecks"
    port           = 30080
    v4_cidr_blocks = ["198.18.235.0/24", "198.18.248.0/24"]
  }
}

{{.BaseTemplate}}
`

const albBackendGroupConfigTemplate = `
{{ if .IsDataSource }}
data "yandex_alb_backend_group" "test-bg-ds" {
  name = yandex_alb_backend_group.test-bg.name
}
{{ end }}
resource "yandex_alb_backend_group" "test-bg" {
  name        = "{{.BGName}}"
  description = "{{.BGDescription}}"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
  {{ if .IsHTTPBackend }}

  {{ if .UseHeaderAffinity }}
  session_affinity {
    header {
      header_name = "x-some-header"
    }
  }
  {{ end }}

  http_backend {
    name             = "test-http-backend"
    weight           = {{.BackendWeight}}
    port             = {{.Port}}
    {{ if .IsStorageBackend }}
    storage_bucket = "{{.StorageBackendBucket}}"
    {{ else }}
    target_group_ids = ["${yandex_alb_target_group.test-target-group.id}"]
    {{ end }}
    tls {
      {{ if not .IsEmptyTLS }}
      sni = "{{.TlsSni}}"
      validation_context {
        trusted_ca_bytes = <<EOF
{{.TlsValidationContext}}
EOF
      }
      {{end}}
    }
    load_balancing_config {
      panic_threshold                = {{.PanicThreshold}}
      locality_aware_routing_percent = {{.LocalityPercent}}
      strict_locality                = {{.StrictLocality}}
      mode                           = "{{.LoadBalancingMode}}"
    }
    {{ if .IsGRPCCheck }}
    healthcheck {
      timeout  = "{{.Timeout}}"
      interval = "{{.Interval}}"
      grpc_healthcheck {
        service_name = "{{.ServiceName}}"
      }
    }
    {{end}}
    {{ if .IsStreamCheck }}
    healthcheck {
      timeout  = "{{.Timeout}}"
      interval = "{{.Interval}}"
      stream_healthcheck {
        receive = "{{.ReceiveText}}"
        send    = "{{.SendText}}"
      }
    }
    {{end}}
    {{ if .IsHTTPCheck }}
    healthcheck {
      timeout = "{{.Timeout}}"
      interval = "{{.Interval}}"
      http_healthcheck {
        host  = "{{.Host}}"
        path  = "{{.Path}}"
        http2 = "{{.HTTP2}}"
        expected_statuses = {{ .ExpectedStatuses }}
      }
    }
    {{end}}
    http2 = "{{.HTTP2}}"
  }
  {{end}}
  {{ if .IsStreamBackend }}
  stream_backend {
    name             = "test-stream-backend"
    weight           = {{.BackendWeight}}
    port             = {{.Port}}
    enable_proxy_protocol = {{.ProxyProtocol}}
    keep_connections_on_host_health_failure = {{.KeepConnectionsOnHostHealthFailure}}
    target_group_ids = ["${yandex_alb_target_group.test-target-group.id}"]
    tls {
      {{ if not .IsEmptyTLS }}
      sni = "{{.TlsSni}}"
      validation_context {
        trusted_ca_bytes = <<EOF
{{.TlsValidationContext}}
EOF
      }
      {{end}}
    }
    load_balancing_config {
      panic_threshold                = {{.PanicThreshold}}
      locality_aware_routing_percent = {{.LocalityPercent}}
      strict_locality                = {{.StrictLocality}}
      mode                           = "{{.LoadBalancingMode}}"
    }
    {{ if .IsGRPCCheck }}
    healthcheck {
      timeout  = "{{.Timeout}}"
      interval = "{{.Interval}}"
      grpc_healthcheck {
        service_name = "{{.ServiceName}}"
      }
    }
    {{end}}
    {{ if .IsStreamCheck }}
    healthcheck {
      timeout  = "{{.Timeout}}"
      interval = "{{.Interval}}"
      stream_healthcheck {
        receive = "{{.ReceiveText}}"
        send    = "{{.SendText}}"
      }
    }
    {{end}}
    {{ if .IsHTTPCheck }}
    healthcheck {
      timeout = "{{.Timeout}}"
      interval = "{{.Interval}}"
      http_healthcheck {
        host  = "{{.Host}}"
        path  = "{{.Path}}"
        http2 = "{{.HTTP2}}"
      }
    }
    {{end}}
  }
  {{end}}
  {{ if .IsGRPCBackend }}
  grpc_backend {
    name             = "test-grpc-backend"
    weight           = {{.BackendWeight}}
    port             = {{.Port}}
    target_group_ids = ["${yandex_alb_target_group.test-target-group.id}"]
    tls {
      sni = "{{.TlsSni}}"
      validation_context {
        trusted_ca_bytes = <<EOF
{{.TlsValidationContext}}
EOF
      }
    }
    load_balancing_config {
      panic_threshold                = {{.PanicThreshold}}
      locality_aware_routing_percent = {{.LocalityPercent}}
      strict_locality                = {{.StrictLocality}}
      mode                           = "{{.LoadBalancingMode}}"
    }
    {{ if .IsGRPCCheck }}
    healthcheck {
      timeout  = "{{.Timeout}}"
      interval = "{{.Interval}}"
      grpc_healthcheck {
        service_name = "{{.ServiceName}}"
      }
    }
    {{end}}
    {{ if .IsStreamCheck }}
    healthcheck {
      timeout  = "{{.Timeout}}"
      interval = "{{.Interval}}"
      stream_healthcheck {
        receive = "{{.ReceiveText}}"
        send    = "{{.SendText}}"
      }
    }
    {{end}}
    {{ if .IsHTTPCheck }}
    healthcheck {
      timeout  = "{{.Timeout}}"
      interval = "{{.Interval}}"
      http_healthcheck {
        host  = "{{.Host}}"
        path  = "{{.Path}}"
        http2 = "{{.HTTP2}}"
      }
    }
    {{end}}
  }
  {{end}}
}
{{ if or .IsHTTPBackend .IsGRPCBackend .IsStreamBackend}}
resource "yandex_alb_target_group" "test-target-group" {
  name		= "{{.TGName}}"

  target {
	subnet_id	= "${yandex_vpc_subnet.test-subnet.id}"
	ip_address	= "${yandex_compute_instance.test-instance-1.network_interface.0.ip_address}"
  }

  target {
	subnet_id	= "${yandex_vpc_subnet.test-subnet.id}"
	ip_address	= "${yandex_compute_instance.test-instance-2.network_interface.0.ip_address}"
  }
}
{{ end }}

{{.BaseTemplate}}
`

func testALBBackendGroupConfig_basic(in resourceALBBackendGroupInfo) string {
	m := structs.Map(in)
	config := templateConfig(albBackendGroupConfigTemplate, m)
	return config
}

func testALBLoadBalancerConfig_basic(in resourceALBLoadBalancerInfo) string {
	m := structs.Map(in)
	config := templateConfig(albLoadBalancerConfigTemplate, m)
	return config
}

func testALBVirtualHostConfig_basic(in resourceALBVirtualHostInfo) string {
	m := structs.Map(in)
	config := templateConfig(albVirtualHostConfigTemplate, m)
	return config
}

func testALBHTTPRouterConfig_basic(in resourceALBHTTPRouterInfo) string {
	m := structs.Map(in)
	config := templateConfig(albHTTPRouterConfigTemplate, m)
	return config
}

func testAccCheckALBBackendGroupValues(bg *apploadbalancer.BackendGroup, expectedHTTPBackends, expectedGRPCBackends, expectedStreamBackends bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if (bg.GetHttp() != nil) != expectedHTTPBackends {
			return fmt.Errorf("invalid presence or absence of HTTP backend Application Backend Group %s", bg.Name)
		}

		if (bg.GetGrpc() != nil) != expectedGRPCBackends {
			return fmt.Errorf("invalid presence or absence of gRPC backend Application Backend Group %s", bg.Name)
		}

		if (bg.GetStream() != nil) != expectedStreamBackends {
			return fmt.Errorf("invalid presence or absence of Stream backend Application Backend Group %s", bg.Name)
		}

		return nil
	}
}

func testAccCheckALBBackendGroupGRPCBackend(bg *apploadbalancer.BackendGroup, expectedTrustedCaBytes string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		backends := bg.GetGrpc().GetBackends()
		if len(backends) == 0 {
			return fmt.Errorf("invalid absence of grpc backend in Application Backend Group %s", bg.GetName())
		}
		return checkALBBackendGroupTrustedCaBytes(backends[0].GetTls(), expectedTrustedCaBytes)
	}
}

func testAccCheckALBBackendGroupHTTPBackend(bg *apploadbalancer.BackendGroup, expectedTrustedCaBytes string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		backends := bg.GetHttp().GetBackends()
		if len(backends) == 0 {
			return fmt.Errorf("invalid absence of http backend in Application Backend Group %s", bg.GetName())
		}
		return checkALBBackendGroupTrustedCaBytes(backends[0].GetTls(), expectedTrustedCaBytes)
	}
}

func checkALBBackendGroupTrustedCaBytes(tls *apploadbalancer.BackendTls, expectedTrustedCaBytes string) error {
	if tls == nil {
		return fmt.Errorf("invalid absence of backend TLS in Application Backend Group")
	}
	if bytes := strings.TrimSpace(tls.GetValidationContext().GetTrustedCaBytes()); bytes != expectedTrustedCaBytes {
		return fmt.Errorf("expected %s but %s was found in trusted ca bytes in Application Backend Group", expectedTrustedCaBytes, bytes)
	}

	return nil
}

func testAccCheckALBVirtualHostValues(vh *apploadbalancer.VirtualHost, expectedHttpRoute, expectedGrpcRoute bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, route := range vh.GetRoutes() {
			if (route.GetHttp() != nil) != expectedHttpRoute {
				return fmt.Errorf("invalid presence or absence of http backend Application Backend Group %s", vh.Name)
			}

			if (route.GetGrpc() != nil) != expectedGrpcRoute {
				return fmt.Errorf("invalid presence or absence of grpc backend Application Backend Group %s", vh.Name)
			}
		}

		return nil
	}
}

func testAccALBGeneralTGTemplate(tgName, tgDesc, baseTemplate string, targetsCount int, isDataSource bool) string {
	targets := make([]string, targetsCount)
	for i := 1; i <= targetsCount; i++ {
		targets[i-1] = fmt.Sprintf("test-instance-%d", i)
	}
	return templateConfig(`
{{ if .IsDataSource }}
data "yandex_alb_target_group" "test-tg-ds" {
  name = yandex_alb_target_group.test-tg.name
}
{{ end }}
resource "yandex_alb_target_group" "test-tg" {
  name        = "{{.TGName}}"
  description = "{{.TGDescription}}"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
  {{range .Targets}}
  target {
    subnet_id  = yandex_vpc_subnet.test-subnet.id
    ip_address = yandex_compute_instance.{{.}}.network_interface.0.ip_address
  }
  {{end}}
}

{{.BaseTemplate}}
		`,
		map[string]interface{}{
			"TGName":        tgName,
			"TGDescription": tgDesc,
			"BaseTemplate":  baseTemplate,
			"Targets":       targets,
			"IsDataSource":  isDataSource,
		},
	)
}

func testAccCheckALBTargetGroupValues(tg *apploadbalancer.TargetGroup, expectedInstanceNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		subnetIPMap, err := getSubnetIPMap(expectedInstanceNames)
		if err != nil {
			return err
		}

		if len(tg.GetTargets()) != len(expectedInstanceNames) {
			return fmt.Errorf("invalid count of targets in Application Target Group %s", tg.Name)
		}

		for _, t := range tg.GetTargets() {
			if addresses, ok := subnetIPMap[t.GetSubnetId()]; ok {
				addressExists := false
				for _, a := range addresses {
					if a == t.GetIpAddress() {
						addressExists = true
						break
					}
				}
				if !addressExists {
					return fmt.Errorf("invalid Target's Address %s in Application Target Group %s", t.GetIpAddress(), tg.Name)
				}
			} else {
				return fmt.Errorf("invalid Target's SubnetID %s in Application Target Group %s", t.GetSubnetId(), tg.Name)
			}
		}

		return nil
	}
}

func testAccALBBaseTemplate(instanceName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "test-image" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "test-instance-1" {
  name        = "%[1]s-1"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  resources {
    cores         = 2
    core_fraction = 20
    memory        = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = data.yandex_compute_image.test-image.id
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.test-subnet.id
  }

  scheduling_policy {
    preemptible = true
  }
}

resource "yandex_compute_instance" "test-instance-2" {
  name        = "%[1]s-2"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  resources {
    cores         = 2
    core_fraction = 20
    memory        = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = data.yandex_compute_image.test-image.id
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.test-subnet.id
  }

  scheduling_policy {
    preemptible = true
  }
}

resource "yandex_vpc_network" "test-network" {}

resource "yandex_vpc_subnet" "test-subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.test-network.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instanceName,
	)
}
