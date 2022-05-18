package yandex

import (
	"bytes"
	"fmt"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
	var modifications []*apploadbalancer.HeaderModification

	for _, currentKey := range IterateKeys(d, key) {
		modification := expandALBModification(d, currentKey)
		modifications = append(modifications, modification)
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

	if _, ok := d.GetOk(path + ".http_route"); ok {
		r, err := expandALBHTTPRoute(d, path+".http_route.0")
		if err != nil {
			return nil, err
		}
		route.SetHttp(r)
	}

	if v, ok := config["grpc_route"]; ok && len(v.([]interface{})) > 0 {
		route.SetGrpc(expandALBGRPCRoute(d, path+".grpc_route.0", v.([]interface{})))
	}

	return route, nil
}

func expandALBHTTPRoute(d *schema.ResourceData, path string) (*apploadbalancer.HttpRoute, error) {
	httpRoute := &apploadbalancer.HttpRoute{}

	if _, ok := d.GetOk(path + ".http_match"); ok {
		m, err := expandALBHTTPRouteMatch(d, path+".http_match.0")
		if err != nil {
			return nil, err
		}
		httpRoute.Match = m
	}
	if _, ok := d.GetOk(path + ".http_route_action"); ok {
		action, err := expandALBHTTPRouteAction(d, path+".http_route_action.0")
		if err != nil {
			return nil, err
		}
		httpRoute.SetRoute(action)
	}
	if _, ok := d.GetOk(path + ".redirect_action"); ok {
		httpRoute.SetRedirect(expandALBRedirectAction(d, path+".redirect_action.0"))
	}
	if _, ok := d.GetOk(path + ".direct_response_action"); ok {
		httpRoute.SetDirectResponse(expandALBDirectResponseAction(d, path+".direct_response_action.0"))
	}

	return httpRoute, nil
}

func expandALBDirectResponseAction(d *schema.ResourceData, path string) *apploadbalancer.DirectResponseAction {
	status := d.Get(path + ".status")
	directResponseAction := &apploadbalancer.DirectResponseAction{
		Status: int64(status.(int)),
	}

	if body, ok := d.GetOk(path + ".body"); ok {
		payload := &apploadbalancer.Payload{}
		payload.SetText(body.(string))
		directResponseAction.Body = payload
	}

	return directResponseAction
}

func expandALBRedirectAction(d *schema.ResourceData, path string) *apploadbalancer.RedirectAction {
	readStr := func(field string) (string, bool) {
		s, ok := d.GetOk(path + "." + field)
		if ok {
			return s.(string), true
		}

		return "", false
	}

	redirectAction := &apploadbalancer.RedirectAction{}

	if val, ok := readStr("replace_scheme"); ok {
		redirectAction.ReplaceScheme = val
	}

	if val, ok := readStr("replace_host"); ok {
		redirectAction.ReplaceHost = val
	}

	if val, ok := d.GetOk(path + ".replace_port"); ok {
		redirectAction.ReplacePort = int64(val.(int))
	}

	if val, ok := d.GetOk(path + ".remove_query"); ok {
		redirectAction.RemoveQuery = val.(bool)
	}

	if val, ok := readStr("replace_path"); ok {
		redirectAction.SetReplacePath(val)
	}

	if val, ok := readStr("replace_prefix"); ok {
		redirectAction.SetReplacePrefix(val)
	}

	if val, ok := readStr("response_code"); ok {
		code := apploadbalancer.RedirectAction_RedirectResponseCode_value[strings.ToUpper(val)]
		redirectAction.ResponseCode = apploadbalancer.RedirectAction_RedirectResponseCode(code)
	}

	return redirectAction
}

func expandALBHTTPRouteAction(d *schema.ResourceData, path string) (*apploadbalancer.HttpRouteAction, error) {
	readStr := func(field string) (string, bool) {
		s, ok := d.GetOk(path + "." + field)
		if ok {
			return s.(string), true
		}

		return "", false
	}

	routeAction := &apploadbalancer.HttpRouteAction{
		BackendGroupId: d.Get(path + ".backend_group_id").(string),
	}

	if val, ok := readStr("timeout"); ok {
		d, err := parseDuration(val)
		if err != nil {
			return nil, err
		}
		routeAction.Timeout = d
	}

	if val, ok := readStr("idle_timeout"); ok {
		d, err := parseDuration(val)
		if err != nil {
			return nil, err
		}
		routeAction.IdleTimeout = d
	}

	if val, ok := readStr("prefix_rewrite"); ok {
		routeAction.PrefixRewrite = val
	}

	if val, ok := d.GetOk(path + ".upgrade_types"); ok {
		upgradeTypes, err := expandALBStringListFromSchemaSet(val)
		if err != nil {
			return nil, err
		}
		routeAction.UpgradeTypes = upgradeTypes
	}

	if val, ok := readStr("host_rewrite"); ok {
		routeAction.SetHostRewrite(val)
	}

	if val, ok := d.GetOk(path + ".auto_host_rewrite"); ok {
		routeAction.SetAutoHostRewrite(val.(bool))
	}

	return routeAction, nil
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

func expandALBHTTPRouteMatch(d *schema.ResourceData, path string) (*apploadbalancer.HttpRouteMatch, error) {
	httpRouteMatch := &apploadbalancer.HttpRouteMatch{}

	if _, ok := d.GetOk(path + ".path"); ok {
		httpRouteMatch.Path = expandALBStringMatch(d, path+".path.0")
	}

	if val, ok := d.GetOk(path + ".http_method"); ok {
		res, err := expandALBStringListFromSchemaSet(val)
		if err != nil {
			return nil, err
		}

		httpRouteMatch.HttpMethod = res
	}
	return httpRouteMatch, nil
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
		grpcRouteMatch.Fqmn = expandALBStringMatch(d, path+".fqmn.0")
	}
	return grpcRouteMatch
}

func expandALBStringMatch(d *schema.ResourceData, path string) *apploadbalancer.StringMatch {
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

	for _, key := range IterateKeys(d, "listener") {
		lis, err := expandALBListener(d, key)
		if err != nil {
			return nil, err
		}

		listeners = append(listeners, lis)
	}

	return listeners, nil
}

func expandALBListener(d *schema.ResourceData, key string) (*apploadbalancer.ListenerSpec, error) {
	listener := &apploadbalancer.ListenerSpec{}

	if v, ok := d.GetOk(key + "name"); ok {
		listener.Name = v.(string)
	}

	if v, ok := d.GetOk(key + "endpoint"); ok {
		listener.EndpointSpecs = expandALBEndpoints(v)
	}

	if conf, ok := getFirstElementConfigIfExists(d, key+"http"); ok {
		listener.SetHttp(expandALBHTTPListener(conf))
	}

	if conf, ok := getFirstElementConfigIfExists(d, key+"tls"); ok {
		listener.SetTls(expandALBTLSListener(conf))
	}

	if conf, ok := getFirstElementConfigIfExists(d, key+"stream"); ok {
		listener.SetStream(expandALBStreamListener(conf))
	}

	return listener, nil
}

func getFirstElementConfigIfExists(d *schema.ResourceData, key string) (map[string]interface{}, bool) {
	if v, ok := d.GetOk(key); ok {
		arr := v.([]interface{})
		if len(arr) > 0 {
			var resultConfig map[string]interface{}
			if result := arr[0]; result != nil {
				resultConfig = result.(map[string]interface{})
			} else {
				resultConfig = map[string]interface{}{}
			}
			return resultConfig, true
		}
	}
	return nil, false
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

func expandALBStreamListener(config map[string]interface{}) *apploadbalancer.StreamListener {
	streamListener := &apploadbalancer.StreamListener{}

	if conf, ok := getFirstElementConfig(config, "handler"); ok {
		streamListener.Handler = expandALBStreamHandler(conf)
	}

	return streamListener
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

func expandALBStreamHandler(config map[string]interface{}) *apploadbalancer.StreamHandler {
	streamHandler := &apploadbalancer.StreamHandler{}

	if v, ok := config["backend_group_id"]; ok {
		streamHandler.BackendGroupId = v.(string)
	}

	return streamHandler
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

	if conf, ok := getFirstElementConfig(config, "stream_handler"); ok {
		tlsHandler.SetStreamHandler(expandALBStreamHandler(conf))
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

	for _, key := range IterateKeys(d, "http_backend") {
		backend, err := expandALBHTTPBackend(d, key)
		if err != nil {
			return nil, err
		}
		backends = append(backends, backend)
	}

	affinity, err := expandALBHTTPSessionAffinity(d)
	if err != nil {
		return nil, fmt.Errorf("Error expanding session affinity while creating ALB Backend Group: %w", err)
	}

	return &apploadbalancer.HttpBackendGroup{
		Backends:        backends,
		SessionAffinity: affinity,
	}, nil
}

func expandALBHTTPSessionAffinity(d *schema.ResourceData) (apploadbalancer.HttpBackendGroup_SessionAffinity, error) {
	if _, ok := d.GetOk("session_affinity.0.connection"); ok {
		conn := &apploadbalancer.ConnectionSessionAffinity{}

		if v, ok := d.GetOk("session_affinity.0.connection.0.source_ip"); ok {
			conn.SourceIp = v.(bool)
		}
		return &apploadbalancer.HttpBackendGroup_Connection{
			Connection: conn,
		}, nil
	}

	if _, ok := d.GetOk("session_affinity.0.header"); ok {
		header := &apploadbalancer.HeaderSessionAffinity{}

		if v, ok := d.GetOk("session_affinity.0.header.0.header_name"); ok {
			header.HeaderName = v.(string)
		}
		return &apploadbalancer.HttpBackendGroup_Header{
			Header: header,
		}, nil
	}

	if _, ok := d.GetOk("session_affinity.0.cookie"); ok {
		cookie := &apploadbalancer.CookieSessionAffinity{}

		if v, ok := d.GetOk("session_affinity.0.cookie.0.name"); ok {
			cookie.Name = v.(string)
		}

		if v, ok := d.GetOk("session_affinity.0.cookie.0.ttl"); ok {
			ttl, err := parseDuration(v.(string))
			if err != nil {
				return nil, fmt.Errorf("failed to read cookie ttl value: %v", err)
			}
			cookie.Ttl = ttl
		}
		return &apploadbalancer.HttpBackendGroup_Cookie{
			Cookie: cookie,
		}, nil
	}

	return nil, nil
}

func expandALBGRPCSessionAffinity(d *schema.ResourceData) (apploadbalancer.GrpcBackendGroup_SessionAffinity, error) {
	if _, ok := d.GetOk("session_affinity.0.connection"); ok {
		conn := &apploadbalancer.ConnectionSessionAffinity{}

		if v, ok := d.GetOk("session_affinity.0.connection.0.source_ip"); ok {
			conn.SourceIp = v.(bool)
		}
		return &apploadbalancer.GrpcBackendGroup_Connection{
			Connection: conn,
		}, nil
	}

	if _, ok := d.GetOk("session_affinity.0.header"); ok {
		header := &apploadbalancer.HeaderSessionAffinity{}

		if v, ok := d.GetOk("session_affinity.0.header.0.header_name"); ok {
			header.HeaderName = v.(string)
		}
		return &apploadbalancer.GrpcBackendGroup_Header{
			Header: header,
		}, nil
	}

	if _, ok := d.GetOk("session_affinity.0.cookie"); ok {
		cookie := &apploadbalancer.CookieSessionAffinity{}

		if v, ok := d.GetOk("session_affinity.0.cookie.0.name"); ok {
			cookie.Name = v.(string)
		}
		return &apploadbalancer.GrpcBackendGroup_Cookie{
			Cookie: cookie,
		}, nil
	}

	return nil, nil
}

func expandALBStreamSessionAffinity(d *schema.ResourceData) (apploadbalancer.StreamBackendGroup_SessionAffinity, error) {
	if _, ok := d.GetOk("session_affinity.0.header"); ok {
		return nil, fmt.Errorf("Header affinity is not supported for stream backend group")
	}

	if _, ok := d.GetOk("session_affinity.0.cookie"); ok {
		return nil, fmt.Errorf("Cookie affinity is not supported for stream backend group")
	}

	if _, ok := d.GetOk("session_affinity.0.connection"); ok {
		conn := &apploadbalancer.ConnectionSessionAffinity{}

		if v, ok := d.GetOk("session_affinity.0.connection.0.source_ip"); ok {
			conn.SourceIp = v.(bool)
		}
		return &apploadbalancer.StreamBackendGroup_Connection{
			Connection: conn,
		}, nil
	}

	return nil, nil
}

func expandALBStreamBackends(d *schema.ResourceData) (*apploadbalancer.StreamBackendGroup, error) {
	var backends []*apploadbalancer.StreamBackend

	for _, key := range IterateKeys(d, "stream_backend") {
		backend, err := expandALBStreamBackend(d, key)
		if err != nil {
			return nil, err
		}
		backends = append(backends, backend)
	}

	affinity, err := expandALBStreamSessionAffinity(d)
	if err != nil {
		return nil, fmt.Errorf("Error expanding session affinity while creating ALB Backend Group: %w", err)
	}

	return &apploadbalancer.StreamBackendGroup{
		Backends:        backends,
		SessionAffinity: affinity,
	}, nil
}

func expandALBStreamBackend(d *schema.ResourceData, key string) (*apploadbalancer.StreamBackend, error) {
	backend := &apploadbalancer.StreamBackend{}

	if v, ok := d.GetOk(key + "name"); ok {
		backend.SetName(v.(string))
	}

	if v, ok := d.GetOk(key + "port"); ok {
		backend.SetPort(int64(v.(int)))
	}

	if v, ok := d.GetOk(key + "weight"); ok {
		backend.SetBackendWeight(&wrappers.Int64Value{
			Value: int64(v.(int)),
		})
	}

	if _, ok := d.GetOk(key + "healthcheck"); ok {
		backend.SetHealthchecks(expandHealthChecks(d, key))
	}

	if v, ok := d.GetOk(key + "tls"); ok && len(v.([]interface{})) == 1 {
		backend.SetTls(expandALBTls(d, key))
	}

	if v, ok := d.GetOk(key + "load_balancing_config"); ok && len(v.([]interface{})) > 0 {
		backend.SetLoadBalancingConfig(expandALBLoadBalancingConfig(v))
	}

	if v, ok := d.GetOk(key + "target_group_ids"); ok {
		backend.SetTargetGroups(expandALBTargetGroupIds(v))
	}

	if v, ok := d.GetOk(key + "enable_proxy_protocol"); ok {
		backend.SetEnableProxyProtocol(v.(bool))
	}

	return backend, nil
}

func expandALBHTTPBackend(d *schema.ResourceData, key string) (*apploadbalancer.HttpBackend, error) {
	backend := &apploadbalancer.HttpBackend{}

	if v, ok := d.GetOk(key + "name"); ok {
		backend.SetName(v.(string))
	}

	if v, ok := d.GetOk(key + "port"); ok {
		backend.SetPort(int64(v.(int)))
	}

	if v, ok := d.GetOk(key + "http2"); ok {
		backend.SetUseHttp2(v.(bool))
	}

	if v, ok := d.GetOk(key + "weight"); ok {
		backend.SetBackendWeight(&wrappers.Int64Value{
			Value: int64(v.(int)),
		})
	}

	if _, ok := d.GetOk(key + "healthcheck"); ok {
		backend.SetHealthchecks(expandHealthChecks(d, key))
	}

	if v, ok := d.GetOk(key + "tls"); ok && len(v.([]interface{})) == 1 {
		backend.SetTls(expandALBTls(d, key))
	}

	if v, ok := d.GetOk(key + "load_balancing_config"); ok && len(v.([]interface{})) > 0 {
		backend.SetLoadBalancingConfig(expandALBLoadBalancingConfig(v))
	}

	var (
		haveTargetGroups  = false
		haveStorageBucket = false
	)
	if v, ok := d.GetOk(key + "target_group_ids"); ok && len(v.([]interface{})) > 0 {
		backend.SetTargetGroups(expandALBTargetGroupIds(v))
		haveTargetGroups = true
	}
	if v, ok := d.GetOk(key + "storage_bucket"); ok {
		backend.SetStorageBucket(expandALBStorageBucket(v))
		haveStorageBucket = backend.GetStorageBucket() != nil
	}

	switch {
	case !haveTargetGroups && !haveStorageBucket:
		return nil, fmt.Errorf("Either target_group_ids or storage_bucket should be specified for http backend")
	case haveTargetGroups && haveStorageBucket:
		return nil, fmt.Errorf("Cannot specify both target_group_ids and storage_bucket for http backend")
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

func expandALBStorageBucket(v interface{}) *apploadbalancer.StorageBucketBackend {
	bucket := v.(string)
	if len(bucket) == 0 {
		return nil
	}
	return &apploadbalancer.StorageBucketBackend{
		Bucket: bucket,
	}
}

func expandALBLoadBalancingConfig(v interface{}) *apploadbalancer.LoadBalancingConfig {
	albConfig := &apploadbalancer.LoadBalancingConfig{}
	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["strict_locality"]; ok {
		albConfig.SetStrictLocality(val.(bool))
	}

	if val, ok := config["locality_aware_routing_percent"]; ok {
		albConfig.SetLocalityAwareRoutingPercent(int64(val.(int)))
	}

	if val, ok := config["panic_threshold"]; ok {
		albConfig.SetPanicThreshold(int64(val.(int)))
	}
	return albConfig
}

func expandHealthChecks(d *schema.ResourceData, key string) []*apploadbalancer.HealthCheck {
	var healthChecks []*apploadbalancer.HealthCheck

	for _, currentKey := range IterateKeys(d, key+"healthcheck") {
		healthCheck := expandHealthCheck(d, currentKey)
		healthChecks = append(healthChecks, healthCheck)
	}
	return healthChecks
}

func expandHealthCheck(d *schema.ResourceData, key string) *apploadbalancer.HealthCheck {
	healthCheck := &apploadbalancer.HealthCheck{}

	if val, ok := d.GetOk(key + "timeout"); ok {
		duration, err := parseDuration(val.(string))
		if err == nil {
			healthCheck.SetTimeout(duration)
		}
	}

	if val, ok := d.GetOk(key + "interval"); ok {
		duration, err := parseDuration(val.(string))
		if err == nil {
			healthCheck.SetInterval(duration)
		}
	}

	if val, ok := d.GetOk(key + "stream_healthcheck"); ok {
		stream := val.([]interface{})
		if len(stream) > 0 {
			healthCheck.SetStream(expandALBStreamHealthCheck(stream[0]))
		}
	}

	if val, ok := d.GetOk(key + "http_healthcheck"); ok {
		http := val.([]interface{})
		if len(http) > 0 {
			healthCheck.SetHttp(expandALBHTTPHealthCheck(http[0]))
		}
	}

	if val, ok := d.GetOk(key + "grpc_healthcheck"); ok {
		grpc := val.([]interface{})
		if len(grpc) > 0 {
			healthCheck.SetGrpc(expandALBGRPCHealthCheck(grpc[0]))
		}
	}

	if val, ok := d.GetOk(key + "healthy_threshold"); ok {
		healthCheck.SetHealthyThreshold(int64(val.(int)))
	}

	if val, ok := d.GetOk(key + "unhealthy_threshold"); ok {
		healthCheck.SetUnhealthyThreshold(int64(val.(int)))
	}

	if val, ok := d.GetOk(key + "healthcheck_port"); ok {
		healthCheck.SetHealthcheckPort(int64(val.(int)))
	}

	if val, ok := d.GetOk(key + "interval_jitter_percent"); ok {
		healthCheck.SetIntervalJitterPercent(val.(float64))
	}

	return healthCheck
}

func expandALBHTTPHealthCheck(v interface{}) *apploadbalancer.HealthCheck_HttpHealthCheck {
	healthCheck := &apploadbalancer.HealthCheck_HttpHealthCheck{}
	config := v.(map[string]interface{})

	if val, ok := config["host"]; ok {
		healthCheck.SetHost(val.(string))
	}

	if val, ok := config["path"]; ok {
		healthCheck.SetPath(val.(string))
	}

	if val, ok := config["http2"]; ok {
		healthCheck.SetUseHttp2(val.(bool))
	}

	return healthCheck
}

func expandALBGRPCHealthCheck(v interface{}) *apploadbalancer.HealthCheck_GrpcHealthCheck {
	healthCheck := &apploadbalancer.HealthCheck_GrpcHealthCheck{}
	config := v.(map[string]interface{})

	if val, ok := config["service_name"]; ok {
		healthCheck.SetServiceName(val.(string))
	}

	return healthCheck
}

func expandALBStreamHealthCheck(v interface{}) *apploadbalancer.HealthCheck_StreamHealthCheck {
	healthCheck := &apploadbalancer.HealthCheck_StreamHealthCheck{}
	config := v.(map[string]interface{})

	if val, ok := config["receive"]; ok {
		payload := &apploadbalancer.Payload{}
		payload.SetText(val.(string))
		healthCheck.SetReceive(payload)
	}

	if val, ok := config["send"]; ok {
		payload := &apploadbalancer.Payload{}
		payload.SetText(val.(string))
		healthCheck.SetSend(payload)
	}
	return healthCheck
}

func expandALBTls(d *schema.ResourceData, key string) *apploadbalancer.BackendTls {
	tls := &apploadbalancer.BackendTls{}
	// there will be only one tls
	for _, tlsKey := range IterateKeys(d, key+"tls") {
		if val, ok := d.GetOk(tlsKey + "sni"); ok {
			tls.SetSni(val.(string))
		}
		if _, ok := d.GetOk(tlsKey + "validation_context"); ok {
			context := &apploadbalancer.ValidationContext{}
			// there will be only one validation context
			for _, contextKey := range IterateKeys(d, tlsKey+"validation_context") {
				if val, ok := d.GetOk(contextKey + "trusted_ca_bytes"); ok {
					context.SetTrustedCaBytes(val.(string))
				}
				if val, ok := d.GetOk(contextKey + "trusted_ca_id"); ok {
					context.SetTrustedCaId(val.(string))
				}
			}
			tls.SetValidationContext(context)
		}
	}
	return tls
}

func expandALBGRPCBackends(d *schema.ResourceData) (*apploadbalancer.GrpcBackendGroup, error) {
	var backends []*apploadbalancer.GrpcBackend

	for _, key := range IterateKeys(d, "grpc_backend") {
		backend, err := expandALBGRPCBackend(d, key)
		if err != nil {
			return nil, err
		}

		backends = append(backends, backend)
	}

	affinity, err := expandALBGRPCSessionAffinity(d)
	if err != nil {
		return nil, fmt.Errorf("Error expanding session affinity while creating ALB Backend Group: %w", err)
	}

	return &apploadbalancer.GrpcBackendGroup{
		Backends:        backends,
		SessionAffinity: affinity,
	}, nil
}

func expandALBGRPCBackend(d *schema.ResourceData, key string) (*apploadbalancer.GrpcBackend, error) {
	backend := &apploadbalancer.GrpcBackend{}

	if v, ok := d.GetOk(key + "name"); ok {
		backend.SetName(v.(string))
	}
	if v, ok := d.GetOk(key + "port"); ok {
		backend.SetPort(int64(v.(int)))
	}

	if v, ok := d.GetOk(key + "tls"); ok && len(v.([]interface{})) == 1 {
		backend.SetTls(expandALBTls(d, key))
	}

	if v, ok := d.GetOk(key + "load_balancing_config"); ok && len(v.([]interface{})) > 0 {
		backend.SetLoadBalancingConfig(expandALBLoadBalancingConfig(v))
	}

	if _, ok := d.GetOk(key + "healthcheck"); ok {
		backend.SetHealthchecks(expandHealthChecks(d, key))
	}

	if v, ok := d.GetOk(key + "weight"); ok {
		backend.SetBackendWeight(&wrappers.Int64Value{
			Value: int64(v.(int)),
		})
	}

	if v, ok := d.GetOk(key + "target_group_ids"); ok {
		backend.SetTargetGroups(expandALBTargetGroupIds(v))
	}
	return backend, nil
}

func IterateKeys(d *schema.ResourceData, key string) []string {
	size := d.Get(key + ".#").(int)
	var keys []string
	for i := 0; i < size; i++ {
		currentKey := fmt.Sprintf(key+".%d.", i)
		keys = append(keys, currentKey)
	}
	return keys
}

func expandALBTargets(d *schema.ResourceData) ([]*apploadbalancer.Target, error) {
	var targets []*apploadbalancer.Target

	for _, key := range IterateKeys(d, "target") {
		target, err := expandALBTarget(d, key)
		if err != nil {
			return nil, err
		}

		targets = append(targets, target)
	}

	return targets, nil
}

func expandALBTarget(d *schema.ResourceData, key string) (*apploadbalancer.Target, error) {
	target := &apploadbalancer.Target{}

	subnet, gotSubnet := d.GetOk(key + "subnet_id")
	privateAddr, gotPrivateAddr := d.GetOk(key + "private_ipv4_address")
	if gotSubnet && gotPrivateAddr {
		return nil, fmt.Errorf("Cannot specify both subnet_id and private_ipv4_address for a target")
	}

	if gotSubnet {
		target.SetSubnetId(subnet.(string))
	}
	if v, ok := d.GetOk(key + "ip_address"); ok {
		target.SetIpAddress(v.(string))
	}
	if gotPrivateAddr {
		target.SetPrivateIpv4Address(privateAddr.(bool))
	}
	return target, nil
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
		case *apploadbalancer.Listener_Stream:
			if stream := listener.GetStream(); stream != nil {
				flListener["stream"] = flattenALBStreamListener(stream)
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

func flattenALBStreamListener(streamListener *apploadbalancer.StreamListener) []interface{} {
	flHTTPListener := map[string]interface{}{
		"handler": flattenALBStreamHandler(streamListener.GetHandler()),
	}

	return []interface{}{flHTTPListener}
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

func flattenALBStreamHandler(streamHandler *apploadbalancer.StreamHandler) []interface{} {
	if streamHandler != nil {
		flHTTPHandler := map[string]interface{}{
			"backend_group_id": streamHandler.GetBackendGroupId(),
		}

		return []interface{}{flHTTPHandler}
	}
	return []interface{}{}
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
			"stream_handler":  flattenALBStreamHandler(tlsHandler.GetStreamHandler()),
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

func flattenALBHTTPSessionAffinity(bg *apploadbalancer.HttpBackendGroup) ([]interface{}, error) {
	var affinityMap map[string]interface{}

	switch {
	case bg.GetHeader() != nil:
		affinityMap = map[string]interface{}{
			"header": []interface{}{
				map[string]interface{}{
					"header_name": bg.GetHeader().GetHeaderName(),
				},
			},
		}
	case bg.GetConnection() != nil:
		affinityMap = map[string]interface{}{
			"connection": []interface{}{
				map[string]interface{}{
					"source_ip": bg.GetConnection().GetSourceIp(),
				},
			},
		}
	case bg.GetCookie() != nil:
		affinityMap = map[string]interface{}{
			"cookie": []interface{}{
				map[string]interface{}{
					"name": bg.GetCookie().GetName(),
					"ttl":  formatDuration(bg.GetCookie().GetTtl()),
				},
			},
		}
	default:
		return nil, nil
	}
	return []interface{}{affinityMap}, nil
}

func flattenALBGRPCSessionAffinity(bg *apploadbalancer.GrpcBackendGroup) ([]interface{}, error) {
	var affinityMap map[string]interface{}

	switch {
	case bg.GetHeader() != nil:
		affinityMap = map[string]interface{}{
			"header": []interface{}{
				map[string]interface{}{
					"header_name": bg.GetHeader().GetHeaderName(),
				},
			},
		}
	case bg.GetConnection() != nil:
		affinityMap = map[string]interface{}{
			"connection": []interface{}{
				map[string]interface{}{
					"source_ip": bg.GetConnection().GetSourceIp(),
				},
			},
		}
	case bg.GetCookie() != nil:
		affinityMap = map[string]interface{}{
			"cookie": []interface{}{
				map[string]interface{}{
					"name": bg.GetCookie().GetName(),
					"ttl":  formatDuration(bg.GetCookie().GetTtl()),
				},
			},
		}
	default:
		return nil, nil
	}
	return []interface{}{affinityMap}, nil
}

func flattenALBStreamSessionAffinity(bg *apploadbalancer.StreamBackendGroup) ([]interface{}, error) {
	if conn := bg.GetConnection(); conn != nil {
		affinityMap := map[string]interface{}{
			"connection": []interface{}{
				map[string]interface{}{
					"source_ip": bg.GetConnection().GetSourceIp(),
				},
			},
		}
		return []interface{}{affinityMap}, nil
	}

	return nil, nil
}

func flattenALBHTTPBackends(bg *apploadbalancer.BackendGroup) ([]interface{}, error) {
	var result []interface{}

	for _, b := range bg.GetHttp().Backends {
		flBackend := map[string]interface{}{
			"name":                  b.Name,
			"port":                  int(b.Port),
			"http2":                 b.UseHttp2,
			"weight":                getWeight(b.GetBackendWeight()),
			"tls":                   flattenALBBackendTLS(b.GetTls()),
			"load_balancing_config": flattenALBLoadBalancingConfig(b.GetLoadBalancingConfig()),
			"healthcheck":           flattenALBHealthChecks(b.GetHealthchecks()),
		}
		switch b.GetBackendType().(type) {
		case *apploadbalancer.HttpBackend_TargetGroups:
			flBackend["target_group_ids"] = b.GetTargetGroups().TargetGroupIds
		case *apploadbalancer.HttpBackend_StorageBucket:
			flBackend["storage_bucket"] = b.GetStorageBucket().GetBucket()
		}
		result = append(result, flBackend)
	}

	return result, nil
}

func flattenALBBackendTLS(tls *apploadbalancer.BackendTls) []map[string]interface{} {
	if tls == nil {
		return []map[string]interface{}{}
	}
	return []map[string]interface{}{{
		"sni":                tls.Sni,
		"validation_context": flattenALBValidationContext(tls.ValidationContext),
	}}
}

func flattenALBLoadBalancingConfig(lbConfig *apploadbalancer.LoadBalancingConfig) []map[string]interface{} {
	if lbConfig == nil {
		return []map[string]interface{}{}
	}
	return []map[string]interface{}{{
		"panic_threshold":                lbConfig.PanicThreshold,
		"locality_aware_routing_percent": lbConfig.LocalityAwareRoutingPercent,
		"strict_locality":                lbConfig.StrictLocality,
	}}
}

func flattenALBValidationContext(context *apploadbalancer.ValidationContext) []interface{} {
	if context == nil {
		return []interface{}{}
	}
	flContext := map[string]interface{}{}
	switch context.GetTrustedCa().(type) {
	case *apploadbalancer.ValidationContext_TrustedCaBytes:
		flContext["trusted_ca_bytes"] = context.GetTrustedCaBytes()
	case *apploadbalancer.ValidationContext_TrustedCaId:
		flContext["trusted_ca_id"] = context.GetTrustedCaId()
	}
	return []interface{}{flContext}
}

func flattenALBStreamBackends(bg *apploadbalancer.BackendGroup) ([]interface{}, error) {
	var result []interface{}

	for _, b := range bg.GetStream().Backends {
		flBackend := map[string]interface{}{
			"name":                  b.Name,
			"port":                  int(b.Port),
			"weight":                getWeight(b.GetBackendWeight()),
			"tls":                   flattenALBBackendTLS(b.GetTls()),
			"load_balancing_config": flattenALBLoadBalancingConfig(b.GetLoadBalancingConfig()),
			"healthcheck":           flattenALBHealthChecks(b.GetHealthchecks()),
			"enable_proxy_protocol": b.GetEnableProxyProtocol(),
		}
		switch b.GetBackendType().(type) {
		case *apploadbalancer.StreamBackend_TargetGroups:
			flBackend["target_group_ids"] = b.GetTargetGroups().TargetGroupIds
		}
		result = append(result, flBackend)
	}

	return result, nil
}

func getWeight(weight *wrapperspb.Int64Value) int {
	if weight == nil {
		return 1
	}
	return int(weight.Value)
}

func flattenALBGRPCBackends(bg *apploadbalancer.BackendGroup) ([]interface{}, error) {
	var result []interface{}

	for _, b := range bg.GetGrpc().Backends {
		flBackend := map[string]interface{}{
			"name":                  b.Name,
			"port":                  int(b.Port),
			"weight":                getWeight(b.GetBackendWeight()),
			"tls":                   flattenALBBackendTLS(b.GetTls()),
			"load_balancing_config": flattenALBLoadBalancingConfig(b.GetLoadBalancingConfig()),
			"healthcheck":           flattenALBHealthChecks(b.GetHealthchecks()),
		}
		switch b.GetBackendType().(type) {
		case *apploadbalancer.GrpcBackend_TargetGroups:
			flBackend["target_group_ids"] = b.GetTargetGroups().TargetGroupIds
		}
		result = append(result, flBackend)
	}

	return result, nil
}

func flattenALBHealthChecks(healthChecks []*apploadbalancer.HealthCheck) interface{} {
	var result []interface{}
	if len(healthChecks) > 0 {
		check := healthChecks[0]

		flHealthCheck := map[string]interface{}{
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
			flHealthCheck["http_healthcheck"] = []map[string]interface{}{
				{
					"host":  http.Host,
					"path":  http.Path,
					"http2": http.UseHttp2,
				},
			}
		case *apploadbalancer.HealthCheck_Grpc:
			flHealthCheck["grpc_healthcheck"] = []map[string]interface{}{
				{
					"service_name": check.GetGrpc().ServiceName,
				},
			}
		case *apploadbalancer.HealthCheck_Stream:
			stream := check.GetStream()
			flHealthCheck["stream_healthcheck"] = []map[string]interface{}{
				{
					"receive": stream.Receive.GetText(),
					"send":    stream.Send.GetText(),
				},
			}
		}

		result = append(result, flHealthCheck)
	}

	return result
}

func flattenALBTargets(tg *apploadbalancer.TargetGroup) []interface{} {
	var result []interface{}

	for _, t := range tg.Targets {
		flTarget := map[string]interface{}{}

		if len(t.SubnetId) > 0 {
			flTarget["subnet_id"] = t.GetSubnetId()
		} else {
			flTarget["private_ipv4_address"] = t.GetPrivateIpv4Address()
		}

		switch t.GetAddressType().(type) {
		case *apploadbalancer.Target_IpAddress:
			flTarget["ip_address"] = t.GetIpAddress()
		}

		result = append(result, flTarget)
	}

	return result
}
