package yandex

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
)

func resourceALBAllocationPolicyLocationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["zone_id"]; ok {
		fmt.Fprintf(&buf, "%s-", v.(string))
	}

	if v, ok := m["subnet_id"]; ok {
		fmt.Fprintf(&buf, "%s-", v.(string))
	}

	if v, ok := m["disable_traffic"]; ok {
		fmt.Fprintf(&buf, "%t-", v.(bool))
	}

	return hashcode.String(buf.String())
}

func resourceALBBackendGroupBackendHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["name"]; ok {
		fmt.Fprintf(&buf, "%s-", v.(string))
	}

	if v, ok := m["port"]; ok {
		fmt.Fprintf(&buf, "%d-", v.(int))
	}

	return hashcode.String(buf.String())
}

func resourceALBBackendGroupHealthcheckHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["timeout"]; ok {
		fmt.Fprintf(&buf, "%s-", v.(string))
	}

	if v, ok := m["interval"]; ok {
		fmt.Fprintf(&buf, "%s-", v.(string))
	}

	if v, ok := m["healthcheck_port"]; ok {
		fmt.Fprintf(&buf, "%d-", v.(int))
	}

	return hashcode.String(buf.String())
}

func expandALBStringListFromSchemaSet(v interface{}) ([]string, error) {
	var m []string
	if v == nil {
		return m, nil
	}
	for _, val := range v.(*schema.Set).List() {
		m = append(m, val.(string))
	}
	return m, nil
}

func expandALBInt64ListFromList(v interface{}) ([]int64, error) {
	var m []int64
	if v == nil {
		return m, nil
	}
	for _, val := range v.([]interface{}) {
		m = append(m, int64(val.(int)))
	}
	return m, nil
}

func expandALBHeaderModification(d *schema.ResourceData, key string) ([]*apploadbalancer.HeaderModification, error) {
	size := d.Get(key + ".#").(int)
	modifications := make([]*apploadbalancer.HeaderModification, size)

	for i := 0; i < size; i++ {
		currentKey := fmt.Sprintf(key+".%d.", i)
		modification := expandALBModification(d, currentKey)
		modifications[i] = modification
	}

	return modifications, nil
}

func expandALBModification(d *schema.ResourceData, key string) *apploadbalancer.HeaderModification {
	modification := &apploadbalancer.HeaderModification{}

	if v, ok := d.GetOk(key + "name"); ok {
		modification.SetName(v.(string))
	}

	if v, ok := d.GetOk(key + "replace"); ok {
		modification.SetReplace(v.(string))
	}

	if v, ok := d.GetOk(key + "append"); ok {
		modification.SetAppend(v.(string))
	}

	if v, ok := d.GetOk(key + "remove"); ok {
		modification.SetRemove(v.(bool))
	}

	return modification
}

func expandALBRoutes(d *schema.ResourceData) ([]*apploadbalancer.Route, error) {
	routeSetRaw, ok := d.GetOk("route")
	if !ok {
		return nil, nil
	}
	routeSet := routeSetRaw.([]interface{})

	var routes []*apploadbalancer.Route
	for i, b := range routeSet {
		routeConfig := b.(map[string]interface{})

		route, err := expandALBRoute(d, fmt.Sprintf("route.%d", i), routeConfig)
		if err != nil {
			return nil, err
		}

		routes = append(routes, route)
	}

	return routes, nil
}

func expandALBRoute(d *schema.ResourceData, path string, config map[string]interface{}) (*apploadbalancer.Route, error) {
	route := &apploadbalancer.Route{}

	if v, ok := config["name"]; ok {
		route.Name = v.(string)
	}

	if v, ok := config["http_route"]; ok {
		if len(v.([]interface{})) > 0 {
			route.SetHttp(expandALBHTTPRoute(d, path+".http_route.0", v.([]interface{})))
		}
	}

	if v, ok := config["grpc_route"]; ok && len(v.([]interface{})) > 0 {
		route.SetGrpc(expandALBGRPCRoute(d, path+".grpc_route.0", v.([]interface{})))
	}

	return route, nil
}

func expandALBHTTPRoute(d *schema.ResourceData, path string, v []interface{}) *apploadbalancer.HttpRoute {
	httpRoute := &apploadbalancer.HttpRoute{}
	config := v[0].(map[string]interface{})
	if val, ok := config["http_match"]; ok && len(val.([]interface{})) > 0 {
		httpRoute.Match = expandALBHTTPRouteMatch(d, path+".http_match.0", val)
	}
	if val, ok := config["http_route_action"]; ok && len(val.([]interface{})) > 0 {
		httpRoute.SetRoute(expandALBHTTPRouteAction(val))
	}

	if val, ok := config["redirect_action"]; ok && len(val.([]interface{})) > 0 {
		httpRoute.SetRedirect(expandALBRedirectAction(val))
	}

	if val, ok := config["direct_response_action"]; ok && len(val.([]interface{})) > 0 {
		httpRoute.SetDirectResponse(expandALBDirectResponseAction(val))
	}
	return httpRoute
}

func expandALBDirectResponseAction(v interface{}) *apploadbalancer.DirectResponseAction {
	directResponseAction := &apploadbalancer.DirectResponseAction{}

	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["status"]; ok {
		directResponseAction.Status = int64(val.(int))
	}

	if val, ok := config["body"]; ok {
		payload := &apploadbalancer.Payload{}
		payload.SetText(val.(string))
		directResponseAction.Body = payload
	}

	return directResponseAction
}

func expandALBRedirectAction(v interface{}) *apploadbalancer.RedirectAction {
	redirectAction := &apploadbalancer.RedirectAction{}

	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["replace_scheme"]; ok {
		redirectAction.ReplaceScheme = val.(string)
	}

	if val, ok := config["replace_host"]; ok {
		redirectAction.ReplaceHost = val.(string)
	}

	if val, ok := config["replace_port"]; ok {
		redirectAction.ReplacePort = int64(val.(int))
	}

	if val, ok := config["remove_query"]; ok {
		redirectAction.RemoveQuery = val.(bool)
	}

	if val, ok := config["replace_path"]; ok {
		redirectAction.SetReplacePath(val.(string))
	}

	if val, ok := config["replace_prefix"]; ok {
		redirectAction.SetReplacePrefix(val.(string))
	}

	if val, ok := config["response_code"]; ok {
		code := apploadbalancer.RedirectAction_RedirectResponseCode_value[strings.ToUpper(val.(string))]
		redirectAction.ResponseCode = apploadbalancer.RedirectAction_RedirectResponseCode(code)
	}

	return redirectAction
}

func expandALBHTTPRouteAction(v interface{}) *apploadbalancer.HttpRouteAction {
	routeAction := &apploadbalancer.HttpRouteAction{}

	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["backend_group_id"]; ok {
		routeAction.BackendGroupId = val.(string)
	}

	if val, ok := config["timeout"]; ok {
		d, err := parseDuration(val.(string))
		if err == nil {
			routeAction.Timeout = d
		}
	}

	if val, ok := config["idle_timeout"]; ok {
		d, err := parseDuration(val.(string))
		if err == nil {
			routeAction.IdleTimeout = d
		}
	}

	if val, ok := config["prefix_rewrite"]; ok {
		routeAction.PrefixRewrite = val.(string)
	}

	if val, ok := config["upgrade_types"]; ok {
		upgradeTypes, err := expandALBStringListFromSchemaSet(val)
		if err == nil {
			routeAction.UpgradeTypes = upgradeTypes
		}
	}

	if val, ok := config["host_rewrite"]; ok {
		routeAction.SetHostRewrite(val.(string))
	}

	if val, ok := config["auto_host_rewrite"]; ok {
		routeAction.SetAutoHostRewrite(val.(bool))
	}

	return routeAction
}

func expandALBGRPCRouteAction(v interface{}) *apploadbalancer.GrpcRouteAction {
	routeAction := &apploadbalancer.GrpcRouteAction{}

	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["backend_group_id"]; ok {
		routeAction.BackendGroupId = val.(string)
	}

	if val, ok := config["max_timeout"]; ok {
		d, err := parseDuration(val.(string))
		if err == nil {
			routeAction.MaxTimeout = d
		}
	}

	if val, ok := config["idle_timeout"]; ok {
		d, err := parseDuration(val.(string))
		if err == nil {
			routeAction.IdleTimeout = d
		}
	}

	if val, ok := config["host_rewrite"]; ok {
		routeAction.SetHostRewrite(val.(string))
	}

	if val, ok := config["auto_host_rewrite"]; ok {
		routeAction.SetAutoHostRewrite(val.(bool))
	}
	return routeAction
}

func expandALBHTTPRouteMatch(d *schema.ResourceData, path string, v interface{}) *apploadbalancer.HttpRouteMatch {
	httpRouteMatch := &apploadbalancer.HttpRouteMatch{}
	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["path"]; ok && len(val.([]interface{})) > 0 {
		httpRouteMatch.Path = expandALBStringMatch(d, path+".path.0", val)
	}
	if val, ok := config["http_method"]; ok {
		if res, err := expandALBStringListFromSchemaSet(val); err == nil {
			httpRouteMatch.HttpMethod = res
		}
	}
	return httpRouteMatch
}

func expandALBGRPCRoute(d *schema.ResourceData, path string, v []interface{}) *apploadbalancer.GrpcRoute {
	grpcRoute := &apploadbalancer.GrpcRoute{}
	config := v[0].(map[string]interface{})
	if val, ok := config["grpc_match"]; ok && len(val.([]interface{})) > 0 {
		grpcRoute.Match = expandALBGRPCRouteMatch(d, path+".grpc_match.0", val)
	}
	if val, ok := config["grpc_route_action"]; ok && len(val.([]interface{})) > 0 {
		grpcRoute.SetRoute(expandALBGRPCRouteAction(val))
	}
	if val, ok := config["grpc_status_response_action"]; ok && len(val.([]interface{})) > 0 {
		grpcRoute.SetStatusResponse(expandALBGRPCStatusResponseAction(val))
	}
	return grpcRoute
}

func expandALBGRPCStatusResponseAction(v interface{}) *apploadbalancer.GrpcStatusResponseAction {
	statusResponseAction := &apploadbalancer.GrpcStatusResponseAction{}

	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["status"]; ok {
		code := apploadbalancer.GrpcStatusResponseAction_Status_value[strings.ToUpper(val.(string))]
		statusResponseAction.Status = apploadbalancer.GrpcStatusResponseAction_Status(code)
	}

	return statusResponseAction
}

func expandALBGRPCRouteMatch(d *schema.ResourceData, path string, v interface{}) *apploadbalancer.GrpcRouteMatch {
	grpcRouteMatch := &apploadbalancer.GrpcRouteMatch{}
	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["fqmn"]; ok && len(val.([]interface{})) > 0 {
		grpcRouteMatch.Fqmn = expandALBStringMatch(d, path+".fqmn.0", val)
	}
	return grpcRouteMatch
}

func expandALBStringMatch(d *schema.ResourceData, path string, v interface{}) *apploadbalancer.StringMatch {
	stringMatch := &apploadbalancer.StringMatch{}
	if val, ok := d.GetOk(path + ".exact"); ok {
		stringMatch.SetExactMatch(val.(string))
	}

	if val, ok := d.GetOk(path + ".prefix"); ok {
		stringMatch.SetPrefixMatch(val.(string))
	}
	return stringMatch
}

func expandALBAllocationPolicy(d *schema.ResourceData) (*apploadbalancer.AllocationPolicy, error) {
	var locations []*apploadbalancer.Location
	config := d.Get("allocation_policy").([]interface{})[0].(map[string]interface{})

	if v, ok := config["location"]; ok {
		locationsList := v.(*schema.Set)
		for _, b := range locationsList.List() {
			locationConfig := b.(map[string]interface{})

			location, err := expandALBLocation(locationConfig)
			if err != nil {
				return nil, err
			}

			locations = append(locations, location)
		}
	}

	return &apploadbalancer.AllocationPolicy{Locations: locations}, nil
}

func expandALBLocation(config map[string]interface{}) (*apploadbalancer.Location, error) {
	location := &apploadbalancer.Location{}

	if v, ok := config["zone_id"]; ok {
		location.ZoneId = v.(string)
	}

	if v, ok := config["subnet_id"]; ok {
		location.SubnetId = v.(string)
	}

	if v, ok := config["disable_traffic"]; ok {
		location.DisableTraffic = v.(bool)
	}

	return location, nil
}

func expandALBListeners(d *schema.ResourceData) ([]*apploadbalancer.ListenerSpec, error) {
	var listeners []*apploadbalancer.ListenerSpec
	backendSet := d.Get("listener").([]interface{})

	for _, l := range backendSet {
		config := l.(map[string]interface{})

		listener, err := expandALBListener(config)
		if err != nil {
			return nil, err
		}

		listeners = append(listeners, listener)
	}

	return listeners, nil
}

func expandALBListener(config map[string]interface{}) (*apploadbalancer.ListenerSpec, error) {
	listener := &apploadbalancer.ListenerSpec{}

	if v, ok := config["name"]; ok {
		listener.Name = v.(string)
	}

	if v, ok := config["endpoint"]; ok {
		listener.EndpointSpecs = expandALBEndpoints(v)
	}

	if conf, ok := getFirstElementConfig(config, "http"); ok {
		listener.SetHttp(expandALBHTTPListener(conf))
	}

	if conf, ok := getFirstElementConfig(config, "tls"); ok {
		listener.SetTls(expandALBTLSListener(conf))
	}

	return listener, nil
}

func getFirstElementConfig(config map[string]interface{}, key string) (map[string]interface{}, bool) {
	if v, ok := config[key]; ok {
		switch v := v.(type) {
		case []interface{}:
			if len(v) > 0 {
				var resultConfig map[string]interface{}
				if result := v[0]; result != nil {
					resultConfig = result.(map[string]interface{})
				} else {
					resultConfig = map[string]interface{}{}
				}
				return resultConfig, true
			}
		case []map[string]interface{}:
			if len(v) > 0 {
				if result := v[0]; result != nil {
					return result, true
				}
			}
		}
	}
	return nil, false
}

func expandALBTLSListener(config map[string]interface{}) *apploadbalancer.TlsListener {
	tlsListener := &apploadbalancer.TlsListener{}
	if conf, ok := getFirstElementConfig(config, "default_handler"); ok {
		tlsListener.SetDefaultHandler(expandALBTLSHandler(conf))
	}
	if v, ok := config["sni_handler"]; ok {
		tlsListener.SniHandlers = expandALBSNIMatches(v)
	}

	return tlsListener
}

func expandALBSNIMatches(v interface{}) []*apploadbalancer.SniMatch {
	var matches []*apploadbalancer.SniMatch

	if v != nil {
		matchSet := v.([]interface{})

		for _, h := range matchSet {
			match := &apploadbalancer.SniMatch{}
			config := h.(map[string]interface{})

			if val, ok := config["name"]; ok {
				match.Name = val.(string)
			}

			if val, ok := config["server_names"]; ok {
				if serverNames, err := expandALBStringListFromSchemaSet(val); err == nil {
					match.ServerNames = serverNames
				}
			}

			if val, ok := config["handler"]; ok {
				handlerConfig := val.([]interface{})
				if len(handlerConfig) == 1 {
					match.Handler = expandALBTLSHandler(handlerConfig[0].(map[string]interface{}))
				}
			}

			matches = append(matches, match)
		}
	}
	return matches
}

func expandALBHTTPListener(config map[string]interface{}) *apploadbalancer.HttpListener {
	httpListener := &apploadbalancer.HttpListener{}

	if conf, ok := getFirstElementConfig(config, "handler"); ok {
		httpListener.Handler = expandALBHTTPHandler(conf)
	}

	if conf, ok := getFirstElementConfig(config, "redirects"); ok {
		if v, ok := conf["http_to_https"]; ok {
			httpListener.Redirects = &apploadbalancer.Redirects{HttpToHttps: v.(bool)}
		}
	}

	return httpListener
}

func expandALBHTTPHandler(config map[string]interface{}) *apploadbalancer.HttpHandler {
	httpHandler := &apploadbalancer.HttpHandler{}

	if v, ok := config["allow_http10"]; ok {
		httpHandler.SetAllowHttp10(v.(bool))
	}

	if v, ok := config["http_router_id"]; ok {
		httpHandler.HttpRouterId = v.(string)
	}

	if conf, ok := getFirstElementConfig(config, "http2_options"); ok {
		http2Options := &apploadbalancer.Http2Options{}
		if val, ok := conf["max_concurrent_streams"]; ok {
			http2Options.MaxConcurrentStreams = int64(val.(int))
		}
		httpHandler.SetHttp2Options(http2Options)
	}

	return httpHandler
}

func expandALBTLSHandler(config map[string]interface{}) *apploadbalancer.TlsHandler {
	tlsHandler := &apploadbalancer.TlsHandler{}

	if conf, ok := getFirstElementConfig(config, "http_handler"); ok {
		tlsHandler.SetHttpHandler(expandALBHTTPHandler(conf))
	}

	if v, ok := config["certificate_ids"]; ok {
		if certificateIDs, err := expandALBStringListFromSchemaSet(v); err == nil {
			tlsHandler.CertificateIds = certificateIDs
		}
	}

	return tlsHandler
}

func expandALBEndpoints(v interface{}) []*apploadbalancer.EndpointSpec {
	var endpoints []*apploadbalancer.EndpointSpec
	if v != nil {

		for _, h := range v.([]interface{}) {
			endpoint := &apploadbalancer.EndpointSpec{}
			config := h.(map[string]interface{})

			if val, ok := config["address"]; ok {
				endpoint.AddressSpecs = expandALBEndpointAddresses(val)
			}

			if val, ok := config["ports"]; ok {
				if ports, err := expandALBInt64ListFromList(val); err == nil {
					endpoint.Ports = ports
				}
			}

			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

func expandALBEndpointAddresses(v interface{}) []*apploadbalancer.AddressSpec {
	var addresses []*apploadbalancer.AddressSpec
	if v != nil {

		for _, h := range v.([]interface{}) {
			elem := &apploadbalancer.AddressSpec{}
			elemConfig := h.(map[string]interface{})

			if config, ok := getFirstElementConfig(elemConfig, "external_ipv4_address"); ok {
				address := &apploadbalancer.ExternalIpv4AddressSpec{}
				if value, ok := config["address"]; ok {
					address.Address = value.(string)
				}
				elem.SetExternalIpv4AddressSpec(address)
			}

			if config, ok := getFirstElementConfig(elemConfig, "internal_ipv4_address"); ok {
				address := &apploadbalancer.InternalIpv4AddressSpec{}
				if value, ok := config["address"]; ok {
					address.Address = value.(string)
				}
				if value, ok := config["subnet_id"]; ok {
					address.SubnetId = value.(string)
				}
				elem.SetInternalIpv4AddressSpec(address)
			}

			if config, ok := getFirstElementConfig(elemConfig, "external_ipv6_address"); ok {
				address := &apploadbalancer.ExternalIpv6AddressSpec{}
				if value, ok := config["address"]; ok {
					address.Address = value.(string)
				}
				elem.SetExternalIpv6AddressSpec(address)
			}

			addresses = append(addresses, elem)
		}
	}
	return addresses
}

func expandALBHTTPBackends(d *schema.ResourceData) (*apploadbalancer.HttpBackendGroup, error) {
	var backends []*apploadbalancer.HttpBackend
	backendSet := d.Get("http_backend").(*schema.Set)

	for _, b := range backendSet.List() {
		backendConfig := b.(map[string]interface{})

		backend, err := expandALBHTTPBackend(backendConfig)
		if err != nil {
			return nil, err
		}

		backends = append(backends, backend)
	}

	return &apploadbalancer.HttpBackendGroup{Backends: backends}, nil
}

func expandALBHTTPBackend(config map[string]interface{}) (*apploadbalancer.HttpBackend, error) {
	backend := &apploadbalancer.HttpBackend{}

	if v, ok := config["name"]; ok {
		backend.Name = v.(string)
	}

	if v, ok := config["port"]; ok {
		backend.Port = int64(v.(int))
	}

	if v, ok := config["http2"]; ok {
		backend.UseHttp2 = v.(bool)
	}

	if v, ok := config["weight"]; ok {
		backend.BackendWeight = &wrappers.Int64Value{
			Value: int64(v.(int)),
		}
	}

	if v, ok := config["healthcheck"]; ok {
		backend.Healthchecks = expandHealthChecks(v)
	}

	if v, ok := config["tls"]; ok && len(v.([]interface{})) > 0 {
		backend.Tls = expandALBTls(v)
	}

	if v, ok := config["load_balancing_config"]; ok && len(v.([]interface{})) > 0 {
		backend.LoadBalancingConfig = expandALBLoadBalancingConfig(v)
	}

	if v, ok := config["target_group_ids"]; ok {
		backend.SetTargetGroups(expandALBTargetGroupIds(v))
	}
	return backend, nil
}

func expandALBTargetGroupIds(v interface{}) *apploadbalancer.TargetGroupsBackend {
	var l []string
	if v != nil {
		for _, val := range v.([]interface{}) {
			l = append(l, val.(string))
		}
	}

	return &apploadbalancer.TargetGroupsBackend{TargetGroupIds: l}
}

func expandALBLoadBalancingConfig(v interface{}) *apploadbalancer.LoadBalancingConfig {
	albConfig := &apploadbalancer.LoadBalancingConfig{}
	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["strict_locality"]; ok {
		albConfig.StrictLocality = val.(bool)
	}

	if val, ok := config["locality_aware_routing_percent"]; ok {
		albConfig.LocalityAwareRoutingPercent = int64(val.(int))
	}

	if val, ok := config["panic_threshold"]; ok {
		albConfig.PanicThreshold = int64(val.(int))
	}
	return albConfig
}

func expandHealthChecks(v interface{}) []*apploadbalancer.HealthCheck {
	var healthchecks []*apploadbalancer.HealthCheck

	if v != nil {
		healthchecksSet := v.(*schema.Set)

		for _, h := range healthchecksSet.List() {
			healthcheck := &apploadbalancer.HealthCheck{}
			config := h.(map[string]interface{})

			if val, ok := config["timeout"]; ok {
				d, err := parseDuration(val.(string))
				if err == nil {
					healthcheck.Timeout = d
				}
			}

			if val, ok := config["interval"]; ok {
				d, err := parseDuration(val.(string))
				if err == nil {
					healthcheck.Interval = d
				}
			}

			if val, ok := config["stream_healthcheck"]; ok {
				stream := val.([]interface{})
				if len(stream) > 0 {
					healthcheck.SetStream(expandALBStreamHealthcheck(stream[0]))
				}
			}

			if val, ok := config["http_healthcheck"]; ok {
				http := val.([]interface{})
				if len(http) > 0 {
					healthcheck.SetHttp(expandALBHTTPHealthcheck(http[0]))
				}
			}

			if val, ok := config["grpc_healthcheck"]; ok {
				grpc := val.([]interface{})
				if len(grpc) > 0 {
					healthcheck.SetGrpc(expandALBGRPCHealthcheck(grpc[0]))
				}
			}

			if val, ok := config["healthy_threshold"]; ok {
				healthcheck.HealthyThreshold = int64(val.(int))
			}

			if val, ok := config["unhealthy_threshold"]; ok {
				healthcheck.UnhealthyThreshold = int64(val.(int))
			}

			if val, ok := config["healthcheck_port"]; ok {
				healthcheck.HealthcheckPort = int64(val.(int))
			}

			if val, ok := config["interval_jitter_percent"]; ok {
				healthcheck.IntervalJitterPercent = val.(float64)
			}

			healthchecks = append(healthchecks, healthcheck)
		}
	}
	return healthchecks
}

func expandALBHTTPHealthcheck(v interface{}) *apploadbalancer.HealthCheck_HttpHealthCheck {
	healthcheck := &apploadbalancer.HealthCheck_HttpHealthCheck{}
	config := v.(map[string]interface{})

	if val, ok := config["host"]; ok {
		healthcheck.Host = val.(string)
	}

	if val, ok := config["path"]; ok {
		healthcheck.Path = val.(string)
	}

	if val, ok := config["http2"]; ok {
		healthcheck.UseHttp2 = val.(bool)
	}

	return healthcheck
}

func expandALBGRPCHealthcheck(v interface{}) *apploadbalancer.HealthCheck_GrpcHealthCheck {
	healthcheck := &apploadbalancer.HealthCheck_GrpcHealthCheck{}
	config := v.(map[string]interface{})

	if val, ok := config["service_name"]; ok {
		healthcheck.ServiceName = val.(string)
	}

	return healthcheck
}

func expandALBStreamHealthcheck(v interface{}) *apploadbalancer.HealthCheck_StreamHealthCheck {
	healthcheck := &apploadbalancer.HealthCheck_StreamHealthCheck{}
	config := v.(map[string]interface{})

	if val, ok := config["receive"]; ok {
		payload := &apploadbalancer.Payload{}
		payload.SetText(val.(string))
		healthcheck.Receive = payload
	}

	if val, ok := config["send"]; ok {
		payload := &apploadbalancer.Payload{}
		payload.SetText(val.(string))
		healthcheck.Send = payload
	}
	return healthcheck
}

func expandALBTls(v interface{}) *apploadbalancer.BackendTls {
	tls := &apploadbalancer.BackendTls{}
	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["sni"]; ok {
		tls.Sni = val.(string)
	}
	if ctx, ok := config["validation_context"]; ok {
		list := ctx.([]interface{})
		if len(list) > 0 {
			context := &apploadbalancer.ValidationContext{}
			if val, ok := ctx.([]interface{})[0].(map[string]interface{})["trusted_ca_bytes"]; ok {
				context.SetTrustedCaBytes(val.(string))
			}
			if val, ok := ctx.([]interface{})[0].(map[string]interface{})["trusted_ca_id"]; ok {
				context.SetTrustedCaId(val.(string))
			}
			tls.SetValidationContext(context)
		}
	}

	return tls
}

func expandALBGRPCBackends(d *schema.ResourceData) (*apploadbalancer.GrpcBackendGroup, error) {
	var backends []*apploadbalancer.GrpcBackend
	backendSet := d.Get("grpc_backend").(*schema.Set)

	for _, b := range backendSet.List() {
		backendConfig := b.(map[string]interface{})

		backend, err := expandALBGRPCBackend(backendConfig)
		if err != nil {
			return nil, err
		}

		backends = append(backends, backend)
	}

	return &apploadbalancer.GrpcBackendGroup{Backends: backends}, nil
}

func expandALBGRPCBackend(config map[string]interface{}) (*apploadbalancer.GrpcBackend, error) {
	backend := &apploadbalancer.GrpcBackend{}

	if v, ok := config["name"]; ok {
		backend.Name = v.(string)
	}
	if v, ok := config["port"]; ok {
		backend.Port = int64(v.(int))
	}

	if v, ok := config["tls"]; ok && len(v.([]interface{})) > 0 {
		backend.Tls = expandALBTls(v)
	}

	if v, ok := config["load_balancing_config"]; ok && len(v.([]interface{})) > 0 {
		backend.LoadBalancingConfig = expandALBLoadBalancingConfig(v)
	}

	if v, ok := config["healthcheck"]; ok {
		backend.Healthchecks = expandHealthChecks(v)
	}

	if v, ok := config["weight"]; ok {
		backend.BackendWeight = &wrappers.Int64Value{
			Value: int64(v.(int)),
		}
	}

	if v, ok := config["target_group_ids"]; ok {
		backend.SetTargetGroups(expandALBTargetGroupIds(v))
	}
	return backend, nil
}

func expandALBTargets(d *schema.ResourceData) ([]*apploadbalancer.Target, error) {
	var targets []*apploadbalancer.Target

	key := "target"
	size := d.Get(key + ".#").(int)

	for i := 0; i < size; i++ {
		currentKey := fmt.Sprintf(key+".%d.", i)
		target := expandALBTarget(d, currentKey)
		targets = append(targets, target)
	}

	return targets, nil
}

func expandALBTarget(d *schema.ResourceData, key string) *apploadbalancer.Target {
	target := &apploadbalancer.Target{}

	if v, ok := d.GetOk(key + "subnet_id"); ok {
		target.SubnetId = v.(string)
	}
	if v, ok := d.GetOk(key + "ip_address"); ok {
		target.SetIpAddress(v.(string))
	}
	return target
}

func flattenALBHeaderModification(modifications []*apploadbalancer.HeaderModification) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	for _, modification := range modifications {
		flModification := map[string]interface{}{
			"name": modification.Name,
		}
		switch modification.Operation.(type) {
		case *apploadbalancer.HeaderModification_Append:
			flModification["append"] = modification.GetAppend()
		case *apploadbalancer.HeaderModification_Replace:
			flModification["replace"] = modification.GetReplace()
		case *apploadbalancer.HeaderModification_Remove:
			flModification["remove"] = modification.GetRemove()
		}

		result = append(result, flModification)
	}

	return result, nil
}

func flattenALBRoutes(routes []*apploadbalancer.Route) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	for _, route := range routes {
		flRoute := map[string]interface{}{
			"name": route.Name,
		}

		switch route.GetRoute().(type) {
		case *apploadbalancer.Route_Http:
			flHttpRoute := flattenALBHTTPRoute(route.GetHttp())
			flRoute["http_route"] = flHttpRoute
		case *apploadbalancer.Route_Grpc:
			flGrpcRoute := flattenALBGRPCRoute(route.GetGrpc())
			flRoute["grpc_route"] = flGrpcRoute
		}

		result = append(result, flRoute)
	}

	return result, nil
}

func flattenALBGRPCRoute(route *apploadbalancer.GrpcRoute) []map[string]interface{} {
	flRoute := make(map[string]interface{})

	if route.GetMatch() != nil {
		flMatch := []map[string]interface{}{
			{
				"fqmn": flattenALBStringMatch(route.Match.Fqmn),
			},
		}

		flRoute["grpc_match"] = flMatch
	}

	switch route.GetAction().(type) {
	case *apploadbalancer.GrpcRoute_Route:
		routeAction := route.GetRoute()
		flRouteAction := []map[string]interface{}{
			{
				"backend_group_id": routeAction.BackendGroupId,
				"max_timeout":      formatDuration(routeAction.MaxTimeout),
				"idle_timeout":     formatDuration(routeAction.IdleTimeout),
			},
		}
		switch routeAction.GetHostRewriteSpecifier().(type) {
		case *apploadbalancer.GrpcRouteAction_HostRewrite:
			flRouteAction[0]["host_rewrite"] = routeAction.GetHostRewrite()
		case *apploadbalancer.GrpcRouteAction_AutoHostRewrite:
			flRouteAction[0]["auto_host_rewrite"] = routeAction.GetAutoHostRewrite()
		}

		flRoute["grpc_route_action"] = flRouteAction

	case *apploadbalancer.GrpcRoute_StatusResponse:
		flRoute["grpc_status_response_action"] = []map[string]interface{}{
			{
				"status": strings.ToLower(route.GetStatusResponse().Status.String()),
			},
		}
	}

	return []map[string]interface{}{flRoute}
}

func flattenALBStringMatch(match *apploadbalancer.StringMatch) []map[string]interface{} {
	switch match.GetMatch().(type) {
	case *apploadbalancer.StringMatch_ExactMatch:
		return []map[string]interface{}{
			{
				"exact": match.GetExactMatch(),
			},
		}
	case *apploadbalancer.StringMatch_PrefixMatch:
		return []map[string]interface{}{
			{
				"prefix": match.GetPrefixMatch(),
			},
		}
	}

	return []map[string]interface{}{}
}

func flattenALBHTTPRoute(route *apploadbalancer.HttpRoute) []map[string]interface{} {
	flRoute := make(map[string]interface{})

	if route.GetMatch() != nil {
		flMatch := []map[string]interface{}{
			{
				"http_method": route.Match.HttpMethod,
				"path":        flattenALBStringMatch(route.Match.Path),
			},
		}

		flRoute["http_match"] = flMatch
	}

	switch route.GetAction().(type) {
	case *apploadbalancer.HttpRoute_Route:
		routeAction := route.GetRoute()
		flRouteAction := []map[string]interface{}{
			{
				"backend_group_id": routeAction.BackendGroupId,
				"timeout":          formatDuration(routeAction.Timeout),
				"idle_timeout":     formatDuration(routeAction.IdleTimeout),
				"prefix_rewrite":   routeAction.PrefixRewrite,
				"upgrade_types":    routeAction.GetUpgradeTypes(),
			},
		}

		switch routeAction.GetHostRewriteSpecifier().(type) {
		case *apploadbalancer.HttpRouteAction_HostRewrite:
			flRouteAction[0]["host_rewrite"] = routeAction.GetHostRewrite()
		case *apploadbalancer.HttpRouteAction_AutoHostRewrite:
			flRouteAction[0]["auto_host_rewrite"] = routeAction.GetAutoHostRewrite()
		}

		flRoute["http_route_action"] = flRouteAction
	case *apploadbalancer.HttpRoute_Redirect:
		redirectAction := route.GetRedirect()
		flRedirectAction := []map[string]interface{}{
			{
				"replace_scheme": redirectAction.ReplaceScheme,
				"replace_host":   redirectAction.ReplaceHost,
				"replace_port":   int(redirectAction.ReplacePort),
				"remove_query":   redirectAction.RemoveQuery,
				"response_code":  strings.ToLower(redirectAction.ResponseCode.String()),
			},
		}

		switch redirectAction.GetPath().(type) {
		case *apploadbalancer.RedirectAction_ReplacePath:
			flRedirectAction[0]["replace_path"] = redirectAction.GetReplacePath()
		case *apploadbalancer.RedirectAction_ReplacePrefix:
			flRedirectAction[0]["replace_prefix"] = redirectAction.GetReplacePrefix()
		}

		flRoute["redirect_action"] = flRedirectAction
	case *apploadbalancer.HttpRoute_DirectResponse:
		directAction := route.GetDirectResponse()
		flDirectAction := []map[string]interface{}{
			{
				"status": int(directAction.Status),
				"body":   directAction.Body.GetText(),
			},
		}

		flRoute["direct_response_action"] = flDirectAction
	}

	return []map[string]interface{}{flRoute}
}

func flattenALBListeners(alb *apploadbalancer.LoadBalancer) ([]interface{}, error) {
	var result []interface{}

	for _, listener := range alb.GetListeners() {

		flListener := map[string]interface{}{
			"name":     listener.Name,
			"endpoint": flattenALBEndpoints(listener.Endpoints),
		}

		switch listener.GetListener().(type) {
		case *apploadbalancer.Listener_Http:
			if http := listener.GetHttp(); http != nil {
				flListener["http"] = flattenALBHTTPListener(http)
			}
		case *apploadbalancer.Listener_Tls:
			if tls := listener.GetTls(); tls != nil {
				flListener["tls"] = flattenALBTLSListener(tls)
			}
		}

		result = append(result, flListener)
	}

	return result, nil
}

func flattenALBEndpoints(endpoints []*apploadbalancer.Endpoint) []interface{} {
	var result []interface{}

	for _, endpoint := range endpoints {
		flEndpoint := map[string]interface{}{
			"address": flattenALBAddresses(endpoint.GetAddresses()),
		}
		var ports []int
		for _, p := range endpoint.GetPorts() {
			ports = append(ports, int(p))
		}
		flEndpoint["ports"] = ports
		result = append(result, flEndpoint)
	}

	return result
}

func flattenALBAddresses(addresses []*apploadbalancer.Address) []interface{} {
	var result []interface{}

	for _, address := range addresses {
		flAddress := map[string]interface{}{}
		if exIPv4 := address.GetExternalIpv4Address(); exIPv4 != nil {
			flAddress["external_ipv4_address"] = []map[string]interface{}{
				{
					"address": exIPv4.GetAddress(),
				},
			}
		}
		if exIPv6 := address.GetExternalIpv6Address(); exIPv6 != nil {
			flAddress["external_ipv6_address"] = []map[string]interface{}{
				{
					"address": exIPv6.GetAddress(),
				},
			}
		}
		if inIPv4 := address.GetInternalIpv4Address(); inIPv4 != nil {
			flAddress["internal_ipv4_address"] = []map[string]interface{}{
				{
					"address":   inIPv4.GetAddress(),
					"subnet_id": inIPv4.GetSubnetId(),
				},
			}
		}
		result = append(result, flAddress)
	}

	return result
}

func flattenALBHTTPListener(httpListener *apploadbalancer.HttpListener) []interface{} {
	flHTTPListener := map[string]interface{}{
		"handler": flattenALBHTTPHandler(httpListener.GetHandler()),
	}
	if redirects := httpListener.GetRedirects(); redirects != nil {
		flHTTPListener["redirects"] = []map[string]interface{}{{
			"http_to_https": redirects.GetHttpToHttps(),
		},
		}
	}
	return []interface{}{flHTTPListener}
}

func flattenALBTLSListener(tlsListener *apploadbalancer.TlsListener) []interface{} {
	flTLSListener := map[string]interface{}{
		"default_handler": flattenALBTLSHandler(tlsListener.GetDefaultHandler()),
		"sni_handler":     flattenALBSniHandlers(tlsListener.GetSniHandlers()),
	}
	return []interface{}{flTLSListener}
}

func flattenALBSniHandlers(matches []*apploadbalancer.SniMatch) []interface{} {
	var result []interface{}
	for _, m := range matches {
		flMatch := map[string]interface{}{
			"name":         m.GetName(),
			"server_names": m.GetServerNames(),
			"handler":      flattenALBTLSHandler(m.GetHandler()),
		}
		result = append(result, flMatch)
	}
	return result
}

func flattenALBHTTPHandler(httpHandler *apploadbalancer.HttpHandler) []interface{} {
	if httpHandler != nil {
		flHTTPHandler := map[string]interface{}{
			"http_router_id": httpHandler.GetHttpRouterId(),
		}

		switch httpHandler.ProtocolSettings.(type) {
		case *apploadbalancer.HttpHandler_Http2Options:
			flHTTPHandler["http2_options"] = []map[string]interface{}{
				{
					"max_concurrent_streams": httpHandler.GetHttp2Options().GetMaxConcurrentStreams(),
				},
			}
		case *apploadbalancer.HttpHandler_AllowHttp10:
			flHTTPHandler["allow_http10"] = httpHandler.GetAllowHttp10()
		}

		return []interface{}{flHTTPHandler}
	}
	return []interface{}{}
}

func flattenALBTLSHandler(tlsHandler *apploadbalancer.TlsHandler) []interface{} {
	if tlsHandler != nil {
		flTLSHandler := map[string]interface{}{
			"certificate_ids": tlsHandler.GetCertificateIds(),
			"http_handler":    flattenALBHTTPHandler(tlsHandler.GetHttpHandler()),
		}
		return []interface{}{flTLSHandler}
	}
	return []interface{}{}
}

func flattenALBAllocationPolicy(alb *apploadbalancer.LoadBalancer) ([]map[string]interface{}, error) {
	result := &schema.Set{F: resourceALBAllocationPolicyLocationHash}

	for _, location := range alb.GetAllocationPolicy().Locations {

		flLocation := map[string]interface{}{
			"zone_id":         location.ZoneId,
			"subnet_id":       location.SubnetId,
			"disable_traffic": location.DisableTraffic,
		}
		result.Add(flLocation)
	}

	return []map[string]interface{}{
		{"location": result},
	}, nil
}

func flattenALBHTTPBackends(bg *apploadbalancer.BackendGroup) (*schema.Set, error) {
	result := &schema.Set{F: resourceALBBackendGroupBackendHash}

	for _, b := range bg.GetHttp().Backends {
		var flTls []map[string]interface{}
		if tls := b.GetTls(); tls != nil {
			flTls = []map[string]interface{}{
				{
					"sni":                tls.Sni,
					"validation_context": tls.ValidationContext,
				},
			}
		}
		var flLoadBalancingConfig []map[string]interface{}
		if lbConfig := b.GetLoadBalancingConfig(); lbConfig != nil {
			flLoadBalancingConfig = []map[string]interface{}{
				{
					"panic_threshold":                lbConfig.PanicThreshold,
					"locality_aware_routing_percent": lbConfig.LocalityAwareRoutingPercent,
					"strict_locality":                lbConfig.StrictLocality,
				},
			}
		}

		flHealthchecks := flattenALBHealthchecks(b.GetHealthchecks())

		flBackend := map[string]interface{}{
			"name":                  b.Name,
			"port":                  int(b.Port),
			"http2":                 b.UseHttp2,
			"weight":                int(b.BackendWeight.Value),
			"tls":                   flTls,
			"load_balancing_config": flLoadBalancingConfig,
			"healthcheck":           flHealthchecks,
		}
		switch b.GetBackendType().(type) {
		case *apploadbalancer.HttpBackend_TargetGroups:
			flBackend["target_group_ids"] = b.GetTargetGroups().TargetGroupIds
		}
		result.Add(flBackend)
	}

	return result, nil
}

func flattenALBGRPCBackends(bg *apploadbalancer.BackendGroup) (*schema.Set, error) {
	result := &schema.Set{F: resourceALBBackendGroupBackendHash}

	for _, b := range bg.GetGrpc().Backends {
		var flTls []map[string]interface{}
		if tls := b.GetTls(); tls != nil {
			flTls = []map[string]interface{}{
				{
					"sni":                tls.Sni,
					"validation_context": tls.ValidationContext,
				},
			}
		}
		var flLoadBalancingConfig []map[string]interface{}
		if lbConfig := b.GetLoadBalancingConfig(); lbConfig != nil {
			flLoadBalancingConfig = []map[string]interface{}{
				{
					"panic_threshold":                lbConfig.PanicThreshold,
					"locality_aware_routing_percent": lbConfig.LocalityAwareRoutingPercent,
					"strict_locality":                lbConfig.StrictLocality,
				},
			}
		}
		flHealthchecks := flattenALBHealthchecks(b.GetHealthchecks())

		flBackend := map[string]interface{}{
			"name":                  b.Name,
			"port":                  int(b.Port),
			"weight":                int(b.BackendWeight.Value),
			"tls":                   flTls,
			"load_balancing_config": flLoadBalancingConfig,
			"healthcheck":           flHealthchecks,
		}
		switch b.GetBackendType().(type) {
		case *apploadbalancer.GrpcBackend_TargetGroups:
			flBackend["target_group_ids"] = b.GetTargetGroups().TargetGroupIds
		}
		result.Add(flBackend)
	}

	return result, nil
}

func flattenALBHealthchecks(healthchecks []*apploadbalancer.HealthCheck) interface{} {
	flHealthchecks := &schema.Set{F: resourceALBBackendGroupHealthcheckHash}
	if len(healthchecks) > 0 {
		check := healthchecks[0]

		flHealthcheck := map[string]interface{}{
			"timeout":                 formatDuration(check.Timeout),
			"interval":                formatDuration(check.Interval),
			"interval_jitter_percent": check.IntervalJitterPercent,
			"healthy_threshold":       check.HealthyThreshold,
			"unhealthy_threshold":     check.UnhealthyThreshold,
			"healthcheck_port":        int(check.HealthcheckPort),
		}
		switch check.GetHealthcheck().(type) {
		case *apploadbalancer.HealthCheck_Http:
			http := check.GetHttp()
			flHealthcheck["http_healthcheck"] = []map[string]interface{}{
				{
					"host":  http.Host,
					"path":  http.Path,
					"http2": http.UseHttp2,
				},
			}
		case *apploadbalancer.HealthCheck_Grpc:
			flHealthcheck["grpc_healthcheck"] = []map[string]interface{}{
				{
					"service_name": check.GetGrpc().ServiceName,
				},
			}
		case *apploadbalancer.HealthCheck_Stream:
			stream := check.GetStream()
			flHealthcheck["stream_healthcheck"] = []map[string]interface{}{
				{
					"receive": stream.Receive.String(),
					"send":    stream.Send.String(),
				},
			}
		}

		flHealthchecks.Add(flHealthcheck)
	}

	return flHealthchecks
}

func flattenALBTargets(tg *apploadbalancer.TargetGroup) []interface{} {
	var result []interface{}

	for _, t := range tg.Targets {
		flTarget := map[string]interface{}{
			"subnet_id": t.SubnetId,
		}

		switch t.GetAddressType().(type) {
		case *apploadbalancer.Target_IpAddress:
			flTarget["ip_address"] = t.GetIpAddress()
		}

		result = append(result, flTarget)
	}

	return result
}
