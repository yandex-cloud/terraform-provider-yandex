package yandex

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"strings"
)

func resourceALBVirtualHostHeaderModificationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["name"]; ok {
		fmt.Fprintf(&buf, "%s-", v.(string))
	}

	if v, ok := m["append"]; ok {
		fmt.Fprintf(&buf, "%s-", v.(string))
	}

	if v, ok := m["replace"]; ok {
		fmt.Fprintf(&buf, "%s-", v.(string))
	}

	if v, ok := m["remove"]; ok {
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

func resourceALBTargetGroupTargetHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["subnet_id"]; ok {
		fmt.Fprintf(&buf, "%s-", v.(string))
	}

	if v, ok := m["ip_address"]; ok {
		fmt.Fprintf(&buf, "%s-", v.(string))
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

func expandALBHeaderModification(d *schema.ResourceData, key string) ([]*apploadbalancer.HeaderModification, error) {
	var modifications []*apploadbalancer.HeaderModification
	modificationSet := d.Get(key).(*schema.Set)

	for _, b := range modificationSet.List() {
		modificationConfig := b.(map[string]interface{})

		backend, err := expandALBModification(modificationConfig)
		if err != nil {
			return nil, err
		}

		modifications = append(modifications, backend)
	}

	return modifications, nil
}

func expandALBModification(config map[string]interface{}) (*apploadbalancer.HeaderModification, error) {
	modification := &apploadbalancer.HeaderModification{}

	if v, ok := config["name"]; ok {
		modification.Name = v.(string)
	}

	if v, ok := config["append"]; ok {
		modification.SetAppend(v.(string))
	}

	if v, ok := config["replace"]; ok {
		modification.SetReplace(v.(string))
	}

	if v, ok := config["remove"]; ok {
		modification.SetRemove(v.(bool))
	}

	return modification, nil
}

func expandALBRoutes(d *schema.ResourceData) ([]*apploadbalancer.Route, error) {
	var routes []*apploadbalancer.Route
	routeSet := d.Get("route").([]interface{})

	for _, b := range routeSet {
		routeConfig := b.(map[string]interface{})

		route, err := expandALBRoute(routeConfig)
		if err != nil {
			return nil, err
		}

		routes = append(routes, route)
	}

	return routes, nil
}

func expandALBRoute(config map[string]interface{}) (*apploadbalancer.Route, error) {
	route := &apploadbalancer.Route{}

	if v, ok := config["name"]; ok {
		route.Name = v.(string)
	}

	if v, ok := config["http_route"]; ok {
		if len(v.([]interface{})) > 0 {
			route.SetHttp(expandALBHTTPRoute(v.([]interface{})))
		}
	}

	if v, ok := config["grpc_route"]; ok && len(v.([]interface{})) > 0 {
		route.SetGrpc(expandALBGRPCRoute(v.([]interface{})))
	}

	return route, nil
}

func expandALBHTTPRoute(v []interface{}) *apploadbalancer.HttpRoute {
	httpRoute := &apploadbalancer.HttpRoute{}
	config := v[0].(map[string]interface{})
	if val, ok := config["http_match"]; ok && len(val.([]interface{})) > 0 {
		httpRoute.Match = expandALBHTTPRouteMatch(val)
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

func expandALBHTTPRouteMatch(v interface{}) *apploadbalancer.HttpRouteMatch {
	httpRouteMatch := &apploadbalancer.HttpRouteMatch{}
	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["path"]; ok && len(val.([]interface{})) > 0 {
		httpRouteMatch.Path = expandALBStringMatch(val)
	}
	if val, ok := config["http_method"]; ok {
		if res, err := expandALBStringListFromSchemaSet(val); err == nil {
			httpRouteMatch.HttpMethod = res
		}
	}
	return httpRouteMatch
}

func expandALBGRPCRoute(v []interface{}) *apploadbalancer.GrpcRoute {
	grpcRoute := &apploadbalancer.GrpcRoute{}
	config := v[0].(map[string]interface{})
	if val, ok := config["grpc_match"]; ok && len(val.([]interface{})) > 0 {
		grpcRoute.Match = expandALBGRPCRouteMatch(val)
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

func expandALBGRPCRouteMatch(v interface{}) *apploadbalancer.GrpcRouteMatch {
	grpcRouteMatch := &apploadbalancer.GrpcRouteMatch{}
	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["fqmn"]; ok && len(val.([]interface{})) > 0 {
		grpcRouteMatch.Fqmn = expandALBStringMatch(val)
	}
	return grpcRouteMatch
}

func expandALBStringMatch(v interface{}) *apploadbalancer.StringMatch {
	stringMatch := &apploadbalancer.StringMatch{}
	config := v.([]interface{})[0].(map[string]interface{})

	if val, ok := config["exact"]; ok {
		stringMatch.SetExactMatch(val.(string))
	}

	if val, ok := config["prefix"]; ok {
		stringMatch.SetPrefixMatch(val.(string))
	}
	return stringMatch
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
		context := &apploadbalancer.ValidationContext{}
		if val, ok := ctx.([]interface{})[0].(map[string]interface{})["trusted_ca_bytes"]; ok {
			context.SetTrustedCaBytes(val.(string))
		}
		if val, ok := ctx.([]interface{})[0].(map[string]interface{})["trusted_ca_id"]; ok {
			context.SetTrustedCaId(val.(string))
		}
		tls.SetValidationContext(context)
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
	targetsSet := d.Get("target").(*schema.Set)

	for _, t := range targetsSet.List() {
		targetConfig := t.(map[string]interface{})

		target, err := expandALBTarget(targetConfig)
		if err != nil {
			return nil, err
		}

		targets = append(targets, target)
	}

	return targets, nil
}

func expandALBTarget(config map[string]interface{}) (*apploadbalancer.Target, error) {
	target := &apploadbalancer.Target{}

	if v, ok := config["subnet_id"]; ok {
		target.SubnetId = v.(string)
	}
	if v, ok := config["ip_address"]; ok {
		target.SetIpAddress(v.(string))
	}
	return target, nil
}

func flattenALBHeaderModification(modifications []*apploadbalancer.HeaderModification) (*schema.Set, error) {
	result := &schema.Set{F: resourceALBVirtualHostHeaderModificationHash}

	for _, modification := range modifications {
		flModification := map[string]interface{}{
			"name":    modification.Name,
			"append":  modification.GetAppend(),
			"replace": modification.GetReplace(),
			"remove":  modification.GetRemove(),
		}

		result.Add(flModification)
	}

	return result, nil
}

func flattenALBRoutes(routes []*apploadbalancer.Route) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	for _, route := range routes {
		flRoute := map[string]interface{}{
			"name": route.Name,
		}

		if route.GetHttp() != nil {
			flHttpRoute := flattenALBHTTPRoute(route.GetHttp())
			flRoute["http_route"] = flHttpRoute
		}

		if route.GetGrpc() != nil {
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

		flRoute["http_match"] = flMatch
	}

	if routeAction := route.GetRoute(); routeAction != nil {
		flRouteAction := []map[string]interface{}{
			{
				"backend_group_id":  routeAction.BackendGroupId,
				"max_timeout":       formatDuration(routeAction.MaxTimeout),
				"idle_timeout":      formatDuration(routeAction.IdleTimeout),
				"host_rewrite":      routeAction.GetHostRewrite(),
				"auto_host_rewrite": routeAction.GetAutoHostRewrite(),
			},
		}

		flRoute["grpc_route_action"] = flRouteAction
	}

	if statusResponseAction := route.GetStatusResponse(); statusResponseAction != nil {
		flRoute["grpc_status_response_action"] = []map[string]interface{}{
			{
				"status": strings.ToLower(statusResponseAction.Status.String()),
			},
		}
	}

	return []map[string]interface{}{flRoute}
}

func flattenALBStringMatch(match *apploadbalancer.StringMatch) []map[string]interface{} {
	flStringMatch := []map[string]interface{}{
		{
			"exact":  match.GetExactMatch(),
			"prefix": match.GetPrefixMatch(),
		},
	}

	return flStringMatch
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

	if routeAction := route.GetRoute(); routeAction != nil {
		flRouteAction := []map[string]interface{}{
			{
				"backend_group_id":  routeAction.BackendGroupId,
				"timeout":           formatDuration(routeAction.Timeout),
				"idle_timeout":      formatDuration(routeAction.IdleTimeout),
				"prefix_rewrite":    routeAction.PrefixRewrite,
				"upgrade_types":     routeAction.GetUpgradeTypes(),
				"host_rewrite":      routeAction.GetHostRewrite(),
				"auto_host_rewrite": routeAction.GetAutoHostRewrite(),
			},
		}

		flRoute["http_route_action"] = flRouteAction
	}

	if redirectAction := route.GetRedirect(); redirectAction != nil {
		flRedirectAction := []map[string]interface{}{
			{
				"replace_scheme": redirectAction.ReplaceScheme,
				"replace_host":   redirectAction.ReplaceHost,
				"replace_port":   int(redirectAction.ReplacePort),
				"remove_query":   redirectAction.RemoveQuery,
				"response_code":  strings.ToLower(redirectAction.ResponseCode.String()),
				"replace_path":   redirectAction.GetReplacePath(),
				"replace_prefix": redirectAction.GetReplacePrefix(),
			},
		}

		flRoute["redirect_action"] = flRedirectAction
	}

	if directAction := route.GetDirectResponse(); directAction != nil {
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

		flHealthchecks := &schema.Set{F: resourceALBBackendGroupHealthcheckHash}
		if healtchchecks := b.GetHealthchecks(); len(healtchchecks) > 0 {
			check := healtchchecks[0]

			flHealthcheck := map[string]interface{}{
				"timeout":                 formatDuration(check.Timeout),
				"interval":                formatDuration(check.Interval),
				"interval_jitter_percent": check.IntervalJitterPercent,
				"healthy_threshold":       check.HealthyThreshold,
				"unhealthy_threshold":     check.UnhealthyThreshold,
				"healthcheck_port":        int(check.HealthcheckPort),
			}

			if http := check.GetHttp(); http != nil {
				flHealthcheck["http_healthcheck"] = []map[string]interface{}{
					{
						"host":  http.Host,
						"path":  http.Path,
						"http2": http.UseHttp2,
					},
				}
			}

			if grpc := check.GetGrpc(); grpc != nil {
				flHealthcheck["grpc_healthcheck"] = []map[string]interface{}{
					{
						"service_name": grpc.ServiceName,
					},
				}
			}

			if stream := check.GetStream(); stream != nil {
				flHealthcheck["stream_healthcheck"] = []map[string]interface{}{
					{
						"receive": stream.Receive.String(),
						"send":    stream.Send.String(),
					},
				}
			}

			flHealthchecks.Add(flHealthcheck)
		}
		flBackend := map[string]interface{}{
			"name":                  b.Name,
			"port":                  int(b.Port),
			"http2":                 b.UseHttp2,
			"weight":                int(b.BackendWeight.Value),
			"tls":                   flTls,
			"load_balancing_config": flLoadBalancingConfig,
			"target_group_ids":      b.GetTargetGroups().TargetGroupIds,
			"healthcheck":           flHealthchecks,
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
		flHealthchecks := &schema.Set{F: resourceALBBackendGroupHealthcheckHash}
		if healthchecks := b.GetHealthchecks(); len(healthchecks) == 1 {
			check := healthchecks[0]

			flHealthcheck := map[string]interface{}{
				"timeout":                 formatDuration(check.Timeout),
				"interval":                formatDuration(check.Interval),
				"interval_jitter_percent": check.IntervalJitterPercent,
				"healthy_threshold":       check.HealthyThreshold,
				"unhealthy_threshold":     check.UnhealthyThreshold,
				"healthcheck_port":        int(check.HealthcheckPort),
			}

			if http := check.GetHttp(); http != nil {
				flHealthcheck["http_healthcheck"] = []map[string]interface{}{
					{
						"host":  http.Host,
						"path":  http.Path,
						"http2": http.UseHttp2,
					},
				}
			}

			if grpc := check.GetGrpc(); grpc != nil {
				flHealthcheck["grpc_healthcheck"] = []map[string]interface{}{
					{
						"service_name": grpc.ServiceName,
					},
				}
			}

			if stream := check.GetStream(); stream != nil {
				flHealthcheck["stream_healthcheck"] = []map[string]interface{}{
					{
						"receive": stream.Receive.String(),
						"send":    stream.Send.String(),
					},
				}
			}

			flHealthchecks.Add(flHealthcheck)
		}

		flBackend := map[string]interface{}{
			"name":                  b.Name,
			"port":                  int(b.Port),
			"weight":                int(b.BackendWeight.Value),
			"tls":                   flTls,
			"load_balancing_config": flLoadBalancingConfig,
			"target_group_ids":      b.GetTargetGroups().TargetGroupIds,
			"healthcheck":           flHealthchecks,
		}
		result.Add(flBackend)
	}

	return result, nil
}

func flattenALBTargets(tg *apploadbalancer.TargetGroup) (*schema.Set, error) {
	result := &schema.Set{F: resourceALBTargetGroupTargetHash}

	for _, t := range tg.Targets {
		flTarget := map[string]interface{}{
			"subnet_id":  t.SubnetId,
			"ip_address": t.GetIpAddress(),
		}
		result.Add(flTarget)
	}

	return result, nil
}
