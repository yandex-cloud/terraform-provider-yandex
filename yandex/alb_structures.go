package yandex

import (
	"bytes"
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/protobuf/types/known/wrapperspb"

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

	// TODO: SA1019: hashcode.String is deprecated: This will be removed in v2 without replacement. If you need its functionality, you can copy it, import crc32 directly, or reference the v1 package. (staticcheck)
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
		modification, err := expandALBModification(d, currentKey)
		if err != nil {
			return nil, err
		}
		modifications = append(modifications, modification)
	}

	return modifications, nil
}

func expandALBModification(d *schema.ResourceData, path string) (*apploadbalancer.HeaderModification, error) {
	modification := &apploadbalancer.HeaderModification{}

	if v, ok := d.GetOk(path + "name"); ok {
		modification.SetName(v.(string))
	}

	replace, gotReplace := d.GetOk(path + "replace")
	remove, gotRemove := d.GetOk(path + "remove")
	appendValue, gotAppend := d.GetOk(path + "append")

	if isPlural(gotReplace, gotRemove, gotAppend) {
		return nil, fmt.Errorf("Cannot specify more than one of replace and remove and append operation for the header modification at the same time")
	}
	if gotReplace {
		modification.SetReplace(replace.(string))
	}

	if gotRemove {
		modification.SetRemove(remove.(bool))
	}

	if gotAppend {
		modification.SetAppend(appendValue.(string))
	}

	return modification, nil
}

func expandALBRateLimit(pathPrefix string, d *schema.ResourceData) (*apploadbalancer.RateLimit, error) {
	var result *apploadbalancer.RateLimit

	sizeValue, ok := d.GetOk(fmt.Sprintf("%v%v.#", pathPrefix, rateLimitSchemaKey))
	if !ok {
		return result, nil
	}

	size := sizeValue.(int)
	if size == 0 {
		return result, nil
	}

	if size > 1 {
		return nil, fmt.Errorf("too many rate limit objects, expected at most 1 got %v instead", size)
	}

	result = &apploadbalancer.RateLimit{}

	allRequests, err := expandALBLimit(
		fmt.Sprintf("%v%v.0.%v", pathPrefix, rateLimitSchemaKey, allRequestsSchemaKey), d,
	)
	if err != nil {
		return nil, err
	}

	result.AllRequests = allRequests

	requestsPerIP, err := expandALBLimit(
		fmt.Sprintf("%v%v.0.%v", pathPrefix, rateLimitSchemaKey, requestsPerIPSchemaKey), d,
	)
	if err != nil {
		return nil, err
	}

	result.RequestsPerIp = requestsPerIP

	return result, nil
}

func expandALBLimit(limitPath string, d *schema.ResourceData) (*apploadbalancer.RateLimit_Limit, error) {
	sizeValue, ok := d.GetOk(fmt.Sprintf("%v.#", limitPath))
	if !ok {
		return nil, nil
	}

	size := sizeValue.(int)
	if size == 0 {
		return nil, nil
	}

	if size > 1 {
		return nil, fmt.Errorf("too many limit objects, expected at most 1 got %v instead", size)
	}

	result := &apploadbalancer.RateLimit_Limit{}

	perSecondValue, ok := d.GetOk(fmt.Sprintf("%v.0.%v", limitPath, perSecondSchemaKey))
	if ok {
		result.Rate = &apploadbalancer.RateLimit_Limit_PerSecond{
			PerSecond: int64(perSecondValue.(int)),
		}
	}

	perMinuteValue, ok := d.GetOk(fmt.Sprintf("%v.0.%v", limitPath, perMinuteSchemaKey))
	if ok {
		result.Rate = &apploadbalancer.RateLimit_Limit_PerMinute{
			PerMinute: int64(perMinuteValue.(int)),
		}
	}

	return result, nil
}

func expandALBRoutes(d *schema.ResourceData) ([]*apploadbalancer.Route, error) {
	var routes []*apploadbalancer.Route

	for _, key := range IterateKeys(d, "route") {
		route, err := expandALBRoute(d, key)
		if err != nil {
			return nil, err
		}

		routes = append(routes, route)
	}

	return routes, nil
}

func expandALBRoute(d *schema.ResourceData, path string) (*apploadbalancer.Route, error) {
	route := &apploadbalancer.Route{}

	if v, ok := d.GetOk(path + "name"); ok {
		route.Name = v.(string)
	}

	if _, ok := d.GetOk(path + "route_options"); ok {
		ro, err := expandALBRouteOptions(d, path+"route_options.0.")
		if err != nil {
			return nil, err
		}
		route.RouteOptions = ro
	}

	_, gotHTTPRoute := d.GetOk(path + "http_route")
	_, gotGRPCRoute := d.GetOk(path + "grpc_route")

	if isPlural(gotHTTPRoute, gotGRPCRoute) {
		return nil, fmt.Errorf("Cannot specify both HTTP route and gRPC route for the route")
	}
	if !gotHTTPRoute && !gotGRPCRoute {
		return nil, fmt.Errorf("Either HTTP route or gRPC route should be specified for the route")
	}
	if gotHTTPRoute {
		r, err := expandALBHTTPRoute(d, path+"http_route.0.")
		if err != nil {
			return nil, err
		}
		route.SetHttp(r)
	}
	if gotGRPCRoute {
		r, err := expandALBGRPCRoute(d, path+"grpc_route.0.")
		if err != nil {
			return nil, err
		}
		route.SetGrpc(r)
	}

	return route, nil
}

func expandALBRouteOptions(d *schema.ResourceData, path string) (*apploadbalancer.RouteOptions, error) {
	ro := &apploadbalancer.RouteOptions{}
	if _, ok := d.GetOk(path + "rbac"); ok {
		rbac, err := expandALBRBAC(d, path+"rbac.0.")
		if err != nil {
			return nil, err
		}

		ro.Rbac = rbac
	}

	if v, ok := d.GetOk(path + "security_profile_id"); ok {
		ro.SecurityProfileId = v.(string)
	}

	return ro, nil
}

func expandALBRBAC(d *schema.ResourceData, path string) (*apploadbalancer.RBAC, error) {
	rbac := &apploadbalancer.RBAC{}

	if v, ok := d.GetOk(path + "action"); ok {
		action := v.(string)
		code, getCode := apploadbalancer.RBAC_Action_value[strings.ToUpper(action)]
		if !getCode {
			return nil, fmt.Errorf("failed to resolve ALB rbac action: found %s", action)
		}
		rbac.SetAction(apploadbalancer.RBAC_Action(code))
	}

	for _, key := range IterateKeys(d, path+"principals") {
		principals, err := expandALBPrincipals(d, key)
		if err != nil {
			return nil, err
		}

		rbac.Principals = append(rbac.Principals, principals)
	}

	return rbac, nil
}

func expandALBPrincipals(d *schema.ResourceData, path string) (*apploadbalancer.Principals, error) {
	var principals []*apploadbalancer.Principal

	for _, key := range IterateKeys(d, path+"and_principals") {
		principal, err := expandALBPrincipal(d, key)
		if err != nil {
			return nil, err
		}

		principals = append(principals, principal)
	}

	return &apploadbalancer.Principals{AndPrincipals: principals}, nil
}

func expandALBPrincipal(d *schema.ResourceData, path string) (*apploadbalancer.Principal, error) {
	principal := &apploadbalancer.Principal{}

	_, gotHeader := d.GetOk(path + "header")
	remoteIP, gotRemoteIP := d.GetOk(path + "remote_ip")
	anyValue, gotAny := d.GetOk(path + "any")

	if isPlural(gotHeader, gotRemoteIP, gotAny) {
		return nil, fmt.Errorf("Cannot specify more than one of header pricnipal and remote ip pricnipal and any principal for the RBAC principal at the same time")
	}

	if !gotHeader && !gotRemoteIP && !gotAny {
		return nil, fmt.Errorf("Either header pricnipal or remote ip pricnipal or any principal should be specified for the RBAC principal")
	}

	if gotHeader {
		headerMatcher, err := expandALBHeaderMatcher(d, path+"header.0.")
		if err != nil {
			return nil, err
		}
		principal.SetHeader(headerMatcher)
	}

	if gotRemoteIP {
		principal.SetRemoteIp(remoteIP.(string))
	}

	if gotAny {
		principal.SetAny(anyValue.(bool))
	}

	return principal, nil
}

func expandALBHeaderMatcher(d *schema.ResourceData, path string) (*apploadbalancer.Principal_HeaderMatcher, error) {
	headerMatcher := &apploadbalancer.Principal_HeaderMatcher{}

	if v, ok := d.GetOk(path + "name"); ok {
		headerMatcher.Name = v.(string)
	}

	if _, ok := d.GetOk(path + "value"); ok {
		value, err := expandALBStringMatch(d, path+"value.0.")
		if err != nil {
			return nil, err
		}
		headerMatcher.Value = value
	}

	return headerMatcher, nil
}

func expandALBHTTPRoute(d *schema.ResourceData, path string) (*apploadbalancer.HttpRoute, error) {
	httpRoute := &apploadbalancer.HttpRoute{}

	if _, ok := d.GetOk(path + "http_match"); ok {
		m, err := expandALBHTTPRouteMatch(d, path+"http_match.0.")
		if err != nil {
			return nil, err
		}
		httpRoute.Match = m
	}

	_, gotHTTPRouteAction := d.GetOk(path + "http_route_action")
	_, gotRedirectAction := d.GetOk(path + "redirect_action")
	_, gotDirectResponseAction := d.GetOk(path + "direct_response_action")

	if isPlural(gotHTTPRouteAction, gotRedirectAction, gotDirectResponseAction) {
		return nil, fmt.Errorf("Cannot specify more than one of HTTP route action and redirect action and direct response action for the HTTP route at the same time")
	}
	if !gotHTTPRouteAction && !gotRedirectAction && !gotDirectResponseAction {
		return nil, fmt.Errorf("Either HTTP route action or redirect action or direct response action should be specified for the HTTP route")
	}
	if gotHTTPRouteAction {
		action, err := expandALBHTTPRouteAction(d, path+"http_route_action.0.")
		if err != nil {
			return nil, err
		}
		httpRoute.SetRoute(action)
	}
	if gotRedirectAction {
		action, err := expandALBRedirectAction(d, path+"redirect_action.0.")
		if err != nil {
			return nil, err
		}
		httpRoute.SetRedirect(action)
	}
	if gotDirectResponseAction {
		httpRoute.SetDirectResponse(expandALBDirectResponseAction(d, path+"direct_response_action.0."))
	}

	return httpRoute, nil
}

func expandALBDirectResponseAction(d *schema.ResourceData, path string) *apploadbalancer.DirectResponseAction {
	status := d.Get(path + "status")
	directResponseAction := &apploadbalancer.DirectResponseAction{
		Status: int64(status.(int)),
	}

	if body, ok := d.GetOk(path + "body"); ok {
		payload := &apploadbalancer.Payload{}
		payload.SetText(body.(string))
		directResponseAction.Body = payload
	}

	return directResponseAction
}

func expandALBRedirectAction(d *schema.ResourceData, path string) (*apploadbalancer.RedirectAction, error) {
	readStr := func(field string) (string, bool) {
		s, ok := d.GetOk(path + field)
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

	if val, ok := d.GetOk(path + "replace_port"); ok {
		redirectAction.ReplacePort = int64(val.(int))
	}

	if val, ok := d.GetOk(path + "remove_query"); ok {
		redirectAction.RemoveQuery = val.(bool)
	}

	replacePath, gotReplacePath := readStr("replace_path")
	replacePrefix, gotReplacePrefix := readStr("replace_prefix")

	if isPlural(gotReplacePrefix, gotReplacePath) {
		return nil, fmt.Errorf("Cannot specify both replace path and replace prefix for the redirect action")
	}
	if gotReplacePath {
		redirectAction.SetReplacePath(replacePath)
	}

	if gotReplacePrefix {
		redirectAction.SetReplacePrefix(replacePrefix)
	}

	if val, ok := readStr("response_code"); ok {
		code, getCode := apploadbalancer.RedirectAction_RedirectResponseCode_value[strings.ToUpper(val)]
		if !getCode {
			return nil, fmt.Errorf("failed to resolve ALB response code: found %s", val)
		}
		redirectAction.ResponseCode = apploadbalancer.RedirectAction_RedirectResponseCode(code)
	}

	return redirectAction, nil
}

func expandALBHTTPRouteAction(d *schema.ResourceData, path string) (*apploadbalancer.HttpRouteAction, error) {
	readStr := func(field string) (string, bool) {
		s, ok := d.GetOk(path + field)
		if ok {
			return s.(string), true
		}

		return "", false
	}

	routeAction := &apploadbalancer.HttpRouteAction{
		BackendGroupId: d.Get(path + "backend_group_id").(string),
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

	if val, ok := d.GetOk(path + "upgrade_types"); ok {
		upgradeTypes, err := expandALBStringListFromSchemaSet(val)
		if err != nil {
			return nil, err
		}
		routeAction.UpgradeTypes = upgradeTypes
	}
	hostRewrite, gotHostRewrite := readStr("host_rewrite")
	autoHostRewrite, gotAutoHostRewrite := d.GetOk(path + "auto_host_rewrite")

	if isPlural(gotHostRewrite, gotAutoHostRewrite) {
		return nil, fmt.Errorf("Cannot specify both host rewrite and auto host rewrite for the HTTP route action")
	}

	if gotHostRewrite {
		routeAction.SetHostRewrite(hostRewrite)
	}

	if gotAutoHostRewrite {
		routeAction.SetAutoHostRewrite(autoHostRewrite.(bool))
	}

	rateLimit, err := expandALBRateLimit(path, d)
	if err != nil {
		return nil, err
	}

	routeAction.RateLimit = rateLimit

	return routeAction, nil
}

func expandALBGRPCRouteAction(d *schema.ResourceData, path string) (*apploadbalancer.GrpcRouteAction, error) {
	routeAction := &apploadbalancer.GrpcRouteAction{}

	if val, ok := d.GetOk(path + "backend_group_id"); ok {
		routeAction.BackendGroupId = val.(string)
	}

	if val, ok := d.GetOk(path + "max_timeout"); ok {
		d, err := parseDuration(val.(string))
		if err == nil {
			routeAction.MaxTimeout = d
		}
	}

	if val, ok := d.GetOk(path + "idle_timeout"); ok {
		d, err := parseDuration(val.(string))
		if err == nil {
			routeAction.IdleTimeout = d
		}
	}

	if val, ok := d.GetOk(path + "host_rewrite"); ok {
		routeAction.SetHostRewrite(val.(string))
	}

	if val, ok := d.GetOk(path + "auto_host_rewrite"); ok {
		routeAction.SetAutoHostRewrite(val.(bool))
	}

	rateLimit, err := expandALBRateLimit(path, d)
	if err != nil {
		return nil, err
	}

	routeAction.RateLimit = rateLimit

	return routeAction, nil
}

func expandALBHTTPRouteMatch(d *schema.ResourceData, path string) (*apploadbalancer.HttpRouteMatch, error) {
	httpRouteMatch := &apploadbalancer.HttpRouteMatch{}

	if _, ok := d.GetOk(path + "path"); ok {
		p, err := expandALBStringMatch(d, path+"path.0.")
		if err != nil {
			return nil, err
		}
		httpRouteMatch.SetPath(p)
	}

	if val, ok := d.GetOk(path + "http_method"); ok {
		res, err := expandALBStringListFromSchemaSet(val)
		if err != nil {
			return nil, err
		}

		httpRouteMatch.HttpMethod = res
	}
	return httpRouteMatch, nil
}

func expandALBGRPCRoute(d *schema.ResourceData, path string) (*apploadbalancer.GrpcRoute, error) {
	grpcRoute := &apploadbalancer.GrpcRoute{}
	if _, ok := d.GetOk(path + "grpc_match"); ok {
		match, err := expandALBGRPCRouteMatch(d, path+"grpc_match.0.")
		if err != nil {
			return nil, err
		}
		grpcRoute.SetMatch(match)
	}

	_, gotGRPCRouteAction := d.GetOk(path + "grpc_route_action")
	gRPCStatusResponseAction, gotGRPCStatusResponseAction := d.GetOk(path + "grpc_status_response_action")

	if isPlural(gotGRPCRouteAction, gotGRPCStatusResponseAction) {
		return nil, fmt.Errorf("Cannot specify both gRPC route action and gRPC status response action for the gRPC route")
	}
	if !gotGRPCRouteAction && !gotGRPCStatusResponseAction {
		return nil, fmt.Errorf("Either gRPC route action or gRPC status response action should be specified for the gRPC route")
	}
	if gotGRPCRouteAction {
		routeAction, err := expandALBGRPCRouteAction(d, path+"grpc_route_action.0.")
		if err != nil {
			return nil, err
		}

		grpcRoute.SetRoute(routeAction)
	}
	if gotGRPCStatusResponseAction {
		status, err := expandALBGRPCStatusResponseAction(gRPCStatusResponseAction)
		if err != nil {
			return nil, err
		}
		grpcRoute.SetStatusResponse(status)
	}

	return grpcRoute, nil
}

func expandALBGRPCStatusResponseAction(v interface{}) (*apploadbalancer.GrpcStatusResponseAction, error) {
	statusResponseAction := &apploadbalancer.GrpcStatusResponseAction{}

	config := v.([]interface{})[0].(map[string]interface{})
	if val, ok := config["status"]; ok {
		status, getStatus := apploadbalancer.GrpcStatusResponseAction_Status_value[strings.ToUpper(val.(string))]
		if !getStatus {
			return nil, fmt.Errorf("failed to resolve ALB grpc status response action: found %s", val)
		}
		statusResponseAction.Status = apploadbalancer.GrpcStatusResponseAction_Status(status)
	}

	return statusResponseAction, nil
}

func expandALBGRPCRouteMatch(d *schema.ResourceData, path string) (*apploadbalancer.GrpcRouteMatch, error) {
	grpcRouteMatch := &apploadbalancer.GrpcRouteMatch{}
	if _, ok := d.GetOk(path + "fqmn"); ok {
		fqmn, err := expandALBStringMatch(d, path+"fqmn.0.")
		if err != nil {
			return nil, err
		}
		grpcRouteMatch.SetFqmn(fqmn)
	}
	return grpcRouteMatch, nil
}

func expandALBStringMatch(d *schema.ResourceData, path string) (*apploadbalancer.StringMatch, error) {
	stringMatch := &apploadbalancer.StringMatch{}
	exactMatch, gotExactMatch := d.GetOk(path + "exact")
	prefixMatch, gotPrefixMatch := d.GetOk(path + "prefix")
	regexMatch, gotRegexMatch := d.GetOk(path + "regex")

	if isPlural(gotExactMatch, gotPrefixMatch, gotRegexMatch) {
		return nil, fmt.Errorf("Cannot specify more than one of exact, prefix and regex match for the string match")
	}
	if !gotExactMatch && !gotPrefixMatch && !gotRegexMatch {
		return nil, fmt.Errorf("At least on of exact, prefix or regex match should be specified for the string match")
	}
	if gotExactMatch {
		stringMatch.SetExactMatch(exactMatch.(string))
	}
	if gotPrefixMatch {
		stringMatch.SetPrefixMatch(prefixMatch.(string))
	}
	if gotRegexMatch {
		stringMatch.SetRegexMatch(regexMatch.(string))
	}

	return stringMatch, nil
}

func expandALBAllocationPolicy(d *schema.ResourceData) (*apploadbalancer.AllocationPolicy, error) {
	if v, ok := d.GetOk("allocation_policy"); !ok || len(v.([]interface{})) == 0 {
		return nil, fmt.Errorf("empty allocation_policy is not supported")
	}

	var locations []*apploadbalancer.Location
	if v, ok := d.GetOk("allocation_policy.0.location"); ok {
		locationsList, ok := v.(*schema.Set)
		if !ok {
			return nil, fmt.Errorf("type error for location set")
		}

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

func expandALBLogOptions(d *schema.ResourceData) (*apploadbalancer.LogOptions, error) {
	v, ok := d.GetOk("log_options")
	if !ok || len(v.([]interface{})) == 0 {
		return nil, nil
	}

	var l apploadbalancer.LogOptions
	if v, ok := d.GetOk("log_options.0.disable"); ok {
		l.Disable = v.(bool)
	}

	for _, key := range IterateKeys(d, "log_options.0.discard_rule") {
		rule, err := expandDiscardRule(d, key)
		if err != nil {
			return nil, err
		}
		l.DiscardRules = append(l.DiscardRules, rule)
	}

	if v, ok := d.GetOk("log_options.0.log_group_id"); ok {
		l.LogGroupId = v.(string)
	}

	return &l, nil
}

func expandDiscardRule(d *schema.ResourceData, key string) (*apploadbalancer.LogDiscardRule, error) {
	var ret apploadbalancer.LogDiscardRule
	var err error
	if v, ok := d.GetOk(key + "http_codes"); ok {
		ret.HttpCodes, err = expandALBInt64ListFromList(v)
		if err != nil {
			return nil, err
		}
	}
	if v, ok := d.GetOk(key + "http_code_intervals"); ok {
		ret.HttpCodeIntervals, err = expandALBCodeIntervalSlice(v.([]interface{}))
		if err != nil {
			return nil, err
		}
	}
	if v, ok := d.GetOk(key + "grpc_codes"); ok {
		ret.GrpcCodes, err = expandGRPCCodeSlice(v.([]interface{}))
		if err != nil {
			return nil, err
		}
	}
	if v, ok := d.GetOk(key + "discard_percent"); ok {
		ret.DiscardPercent = &wrapperspb.Int64Value{Value: int64(v.(int))}
	}
	return &ret, nil
}

func expandALBCodeIntervalSlice(v []interface{}) ([]apploadbalancer.HttpCodeInterval, error) {
	s := make([]apploadbalancer.HttpCodeInterval, len(v))
	if v == nil {
		return s, nil
	}

	for i, val := range v {
		httpCodeInterval, err := parseAlbHttpCodeInterval(val.(string))
		if err != nil {
			return nil, err
		}

		s[i] = httpCodeInterval
	}

	return s, nil
}

func expandGRPCCodeSlice(v []interface{}) ([]code.Code, error) {
	s := make([]code.Code, len(v))
	if v == nil {
		return s, nil
	}

	for i, val := range v {
		code_, err := parseCodeCode(val.(string))
		if err != nil {
			return nil, err
		}

		s[i] = code_
	}

	return s, nil
}

func parseAlbHttpCodeInterval(str string) (apploadbalancer.HttpCodeInterval, error) {
	val, ok := apploadbalancer.HttpCodeInterval_value[str]
	if !ok {
		return apploadbalancer.HttpCodeInterval(0), invalidKeyError("http_code_interval", apploadbalancer.HttpCodeInterval_value, str)
	}
	return apploadbalancer.HttpCodeInterval(val), nil
}

func parseCodeCode(str string) (code.Code, error) {
	val, ok := code.Code_value[str]
	if !ok {
		return code.Code(0), invalidKeyError("code", code.Code_value, str)
	}
	return code.Code(val), nil
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

func isPlural(values ...bool) bool {
	n := 0
	for _, value := range values {
		if value {
			n++
		}
	}
	return n > 1
}

func expandALBListener(d *schema.ResourceData, path string) (*apploadbalancer.ListenerSpec, error) {
	listener := &apploadbalancer.ListenerSpec{}

	if v, ok := d.GetOk(path + "name"); ok {
		listener.Name = v.(string)
	}

	if _, ok := d.GetOk(path + "endpoint"); ok {
		endpoints, err := expandALBEndpoints(d, path+"endpoint")
		if err != nil {
			return nil, err
		}
		listener.SetEndpointSpecs(endpoints)
	}

	nonNilCount := 0
	setListener := func(listenerType string) error {
		if _, got := d.GetOk(path + listenerType + ".0"); got {
			pathToListener := path + listenerType + ".0."
			switch listenerType {
			case "http":
				if httpListener, err := expandALBHTTPListener(d, pathToListener); err != nil {
					return err
				} else if httpListener != nil {
					nonNilCount++
					listener.SetHttp(httpListener)
				}
			case "tls":
				if tlsListener, err := expandALBTLSListener(d, pathToListener); err != nil {
					return err
				} else if tlsListener != nil {
					nonNilCount++
					listener.SetTls(tlsListener)
				}
			case "stream":
				if streamListener, err := expandALBStreamListener(d, pathToListener); err != nil {
					return err
				} else if streamListener != nil {
					nonNilCount++
					listener.SetStream(streamListener)
				}
			}
		}
		return nil
	}

	listenerTypes := []string{"http", "tls", "stream"}

	for _, listenerType := range listenerTypes {
		if err := setListener(listenerType); err != nil {
			return nil, err
		}
	}

	if nonNilCount == 0 {
		return nil, fmt.Errorf("Either HTTP listener or Stream listener or TLS listener should be specified for the ALB listener")
	} else if nonNilCount > 1 {
		return nil, fmt.Errorf("Cannot specify more than one of HTTP listener and Stream listener and TLS listener for the ALB listener at the same time")
	}

	return listener, nil
}

type maybeUsedObject[T any] struct {
	object *T
}

func (mObj *maybeUsedObject[T]) maybeCreate() *T {
	if mObj.object == nil {
		mObj.object = new(T)
	}
	return mObj.object
}

func (mObj *maybeUsedObject[T]) get() *T {
	return mObj.object
}

func expandALBTLSListener(d *schema.ResourceData, path string) (*apploadbalancer.TlsListener, error) {
	var mTlsListener maybeUsedObject[apploadbalancer.TlsListener]
	if _, ok := d.GetOk(path + "default_handler.0"); ok {
		if handler, err := expandALBTLSHandler(d, path+"default_handler.0."); err != nil {
			return nil, err
		} else if handler != nil {
			mTlsListener.maybeCreate().SetDefaultHandler(handler)
		}
	}
	if _, ok := d.GetOk(path + "sni_handler"); ok {
		if sniHandlers, err := expandALBSNIMatches(d, path+"sni_handler"); err != nil {
			return nil, err
		} else if sniHandlers != nil {
			mTlsListener.maybeCreate().SetSniHandlers(sniHandlers)
		}
	}

	return mTlsListener.get(), nil
}

func expandALBSNIMatch(d *schema.ResourceData, path string) (*apploadbalancer.SniMatch, error) {
	var mMatch maybeUsedObject[apploadbalancer.SniMatch]

	if val, ok := d.GetOk(path + "name"); ok {
		mMatch.maybeCreate().SetName(val.(string))
	}

	if val, ok := d.GetOk(path + "server_names"); ok {
		if serverNames, err := expandALBStringListFromSchemaSet(val); err == nil {
			mMatch.maybeCreate().SetServerNames(serverNames)
		}
	}

	if _, ok := d.GetOk(path + "handler.0"); ok {
		if handler, err := expandALBTLSHandler(d, path+"handler.0."); err != nil {
			return nil, err
		} else if handler != nil {
			mMatch.maybeCreate().SetHandler(handler)
		}
	}

	return mMatch.get(), nil
}

func expandALBSNIMatches(d *schema.ResourceData, path string) ([]*apploadbalancer.SniMatch, error) {
	var matches []*apploadbalancer.SniMatch

	for _, key := range IterateKeys(d, path) {
		match, err := expandALBSNIMatch(d, key)
		if err != nil {
			return nil, err
		}
		if match != nil {
			matches = append(matches, match)
		}
	}
	if len(matches) == 0 {
		return nil, nil
	}
	return matches, nil
}

func expandALBStreamListener(d *schema.ResourceData, path string) (*apploadbalancer.StreamListener, error) {
	var mStreamListener maybeUsedObject[apploadbalancer.StreamListener]

	if _, ok := d.GetOk(path + "handler.0"); ok {
		handler, err := expandALBStreamHandler(d, path+"handler.0.")
		if err != nil {
			return nil, err
		}
		if handler != nil {
			mStreamListener.maybeCreate().SetHandler(handler)
		}
	}

	return mStreamListener.get(), nil
}

func expandALBHTTPListener(d *schema.ResourceData, path string) (*apploadbalancer.HttpListener, error) {
	var mHttpListener maybeUsedObject[apploadbalancer.HttpListener]

	if _, ok := d.GetOk(path + "handler.0"); ok {
		if handler, err := expandALBHTTPHandler(d, path+"handler.0."); err != nil {
			return nil, err
		} else if handler != nil {
			mHttpListener.maybeCreate().SetHandler(handler)
		}
	}

	if _, ok := d.GetOk(path + "redirects.0"); ok {
		currentKey := path + "redirects.0." + "http_to_https"
		if v, ok := d.GetOk(currentKey); ok {
			mHttpListener.maybeCreate().SetRedirects(&apploadbalancer.Redirects{HttpToHttps: v.(bool)})
		}
	}

	return mHttpListener.get(), nil
}

func expandALBStreamHandler(d *schema.ResourceData, path string) (*apploadbalancer.StreamHandler, error) {
	var mStreamHandler maybeUsedObject[apploadbalancer.StreamHandler]

	if v, ok := d.GetOk(path + "backend_group_id"); ok {
		mStreamHandler.maybeCreate().SetBackendGroupId(v.(string))
	}
	if v, ok := d.GetOk(path + "idle_timeout"); ok {
		d, err := parseDuration(v.(string))
		if err != nil {
			return nil, err
		}
		mStreamHandler.maybeCreate().SetIdleTimeout(d)
	}
	return mStreamHandler.get(), nil
}

func expandALBHTTPHandler(d *schema.ResourceData, path string) (*apploadbalancer.HttpHandler, error) {
	var mHttpHandler maybeUsedObject[apploadbalancer.HttpHandler]

	if v, ok := d.GetOk(path + "http_router_id"); ok {
		mHttpHandler.maybeCreate().SetHttpRouterId(v.(string))
	}

	if v, ok := d.GetOk(path + "rewrite_request_id"); ok {
		mHttpHandler.maybeCreate().SetRewriteRequestId(v.(bool))
	}

	allowHTTP10, gotAllowHTTP10 := d.GetOk(path + "allow_http10")
	_, gotHTTP2Options := d.GetOk(path + "http2_options.0")

	if isPlural(gotAllowHTTP10, gotHTTP2Options) {
		return nil, fmt.Errorf("Cannot specify both allow HTTP 1.0 and HTTP 2 options for the HTTP Handler")
	}

	if gotAllowHTTP10 {
		mHttpHandler.maybeCreate().SetAllowHttp10(allowHTTP10.(bool))
	}

	if gotHTTP2Options {
		currentKey := path + "http2_options.0." + "max_concurrent_streams"
		if val, ok := d.GetOk(currentKey); ok {
			mHttpHandler.maybeCreate().SetHttp2Options(
				&apploadbalancer.Http2Options{
					MaxConcurrentStreams: int64(val.(int)),
				})
		}
	}

	return mHttpHandler.get(), nil
}

func expandALBTLSHandler(d *schema.ResourceData, path string) (*apploadbalancer.TlsHandler, error) {
	var mTlsHandler maybeUsedObject[apploadbalancer.TlsHandler]

	_, gotHTTPHandler := d.GetOk(path + "http_handler.0")
	_, gotStreamHandler := d.GetOk(path + "stream_handler.0")

	// todo: there will be an error with validation: user can send no handlers
	if !gotHTTPHandler && !gotStreamHandler {
		return nil, fmt.Errorf("Either HTTP handler or Stream handler should be specified for the TLS Handler")
	}
	assignCount := 0
	if gotHTTPHandler {
		if handler, err := expandALBHTTPHandler(d, path+"http_handler.0."); err != nil {
			return nil, err
		} else if handler != nil {
			mTlsHandler.maybeCreate().SetHttpHandler(handler)
			assignCount++
		}
	}

	if gotStreamHandler {
		if handler, err := expandALBStreamHandler(d, path+"stream_handler.0."); err != nil {
			return nil, err
		} else if handler != nil {
			mTlsHandler.maybeCreate().SetStreamHandler(handler)
			assignCount++
		}
	}

	if assignCount > 1 {
		return nil, fmt.Errorf("Cannot specify both HTTP handler and Stream handler for the TLS Handler")
	}

	if v, ok := d.GetOk(path + "certificate_ids"); ok {
		if certificateIDs, err := expandALBStringListFromSchemaSet(v); err == nil {
			mTlsHandler.maybeCreate().CertificateIds = certificateIDs
		}
	}

	return mTlsHandler.get(), nil
}

func expandALBEndpoint(d *schema.ResourceData, path string) (*apploadbalancer.EndpointSpec, error) {
	endpoint := &apploadbalancer.EndpointSpec{}

	if _, ok := d.GetOk(path + "address"); ok {
		address, err := expandALBEndpointAddresses(d, path+"address")
		if err != nil {
			return nil, err
		}
		endpoint.SetAddressSpecs(address)
	}

	if val, ok := d.GetOk(path + "ports"); ok {
		if ports, err := expandALBInt64ListFromList(val); err == nil {
			endpoint.Ports = ports
		}
	}

	return endpoint, nil
}

func expandALBEndpoints(d *schema.ResourceData, path string) ([]*apploadbalancer.EndpointSpec, error) {
	var endpoints []*apploadbalancer.EndpointSpec

	for _, key := range IterateKeys(d, path) {
		endpoint, err := expandALBEndpoint(d, key)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

func expandALBEndpointAddress(d *schema.ResourceData, path string) (*apploadbalancer.AddressSpec, error) {
	endpointAddress := &apploadbalancer.AddressSpec{}

	_, gotExternalIPV4Address := d.GetOk(path + "external_ipv4_address.0")
	_, gotInternalIPV4Address := d.GetOk(path + "internal_ipv4_address.0")
	_, gotExternalIPV6Address := d.GetOk(path + "external_ipv6_address.0")

	if isPlural(gotExternalIPV4Address, gotInternalIPV4Address, gotExternalIPV6Address) {
		return nil, fmt.Errorf("Cannot specify more than one of external ipv4 address and internal ipv4 address and external ipv6 address for the endpoint address at the same time")
	}
	if !gotExternalIPV4Address && !gotInternalIPV4Address && !gotExternalIPV6Address {
		return nil, fmt.Errorf("Either external ipv4 address or internal ipv4 address or external ipv6 address should be specified for the HTTP route")
	}

	if gotExternalIPV4Address {
		currentKey := path + "external_ipv4_address.0." + "address"
		address := &apploadbalancer.ExternalIpv4AddressSpec{}
		if value, ok := d.GetOk(currentKey); ok {
			address.Address = value.(string)
		}
		endpointAddress.SetExternalIpv4AddressSpec(address)
	}

	if gotInternalIPV4Address {
		currentPath := path + "internal_ipv4_address.0."
		address := &apploadbalancer.InternalIpv4AddressSpec{}
		if value, ok := d.GetOk(currentPath + "address"); ok {
			address.Address = value.(string)
		}
		if value, ok := d.GetOk(currentPath + "subnet_id"); ok {
			address.SubnetId = value.(string)
		}
		endpointAddress.SetInternalIpv4AddressSpec(address)
	}

	if gotExternalIPV6Address {
		currentKey := path + "external_ipv6_address.0." + "address"
		address := &apploadbalancer.ExternalIpv6AddressSpec{}
		if value, ok := d.GetOk(currentKey); ok {
			address.Address = value.(string)
		}
		endpointAddress.SetExternalIpv6AddressSpec(address)
	}

	return endpointAddress, nil
}

func expandALBEndpointAddresses(d *schema.ResourceData, path string) ([]*apploadbalancer.AddressSpec, error) {
	var addresses []*apploadbalancer.AddressSpec

	for _, key := range IterateKeys(d, path) {
		address, err := expandALBEndpointAddress(d, key)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
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
		config, err := expandALBLoadBalancingConfig(v)
		if err != nil {
			return nil, err
		}
		backend.SetLoadBalancingConfig(config)
	}

	if v, ok := d.GetOk(key + "target_group_ids"); ok {
		backend.SetTargetGroups(expandALBTargetGroupIds(v))
	}

	if v, ok := d.GetOk(key + "enable_proxy_protocol"); ok {
		backend.SetEnableProxyProtocol(v.(bool))
	}

	if v, ok := d.GetOk(key + keepConnectionsOnHostHealthFailureSchemaKey); ok {
		backend.SetKeepConnectionsOnHostHealthFailure(v.(bool))
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
		config, err := expandALBLoadBalancingConfig(v)
		if err != nil {
			return nil, err
		}
		backend.SetLoadBalancingConfig(config)
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
	case isPlural(haveTargetGroups, haveStorageBucket):
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

func expandALBLoadBalancingConfig(v interface{}) (*apploadbalancer.LoadBalancingConfig, error) {
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

	if val, ok := config["mode"]; ok {
		mode, getMode := apploadbalancer.LoadBalancingMode_value[strings.ToUpper(val.(string))]
		if !getMode {
			return nil, fmt.Errorf("failed to resolve ALB load balamcing config mode: found %s", val)
		}
		albConfig.SetMode(apploadbalancer.LoadBalancingMode(mode))
	}
	return albConfig, nil
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

	if _, ok := d.GetOk(key + "stream_healthcheck"); ok {
		healthCheck.SetStream(expandALBStreamHealthCheck(d, key+"stream_healthcheck.0."))
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

	if val, ok := config[expectedStatusesSchemaKey]; ok {
		if statuses, err := expandALBInt64ListFromList(val); err == nil {
			healthCheck.SetExpectedStatuses(statuses)
		}
	}

	return healthCheck
}

func expandALBGRPCHealthCheck(v interface{}) *apploadbalancer.HealthCheck_GrpcHealthCheck {
	healthCheck := &apploadbalancer.HealthCheck_GrpcHealthCheck{}

	if config, ok := v.(map[string]interface{}); ok {
		if val, ok := config["service_name"]; ok {
			if serviceName, ok := val.(string); ok {
				healthCheck.SetServiceName(serviceName)
			}
		}
	}

	return healthCheck
}

func expandALBStreamHealthCheck(d *schema.ResourceData, key string) *apploadbalancer.HealthCheck_StreamHealthCheck {
	healthCheck := &apploadbalancer.HealthCheck_StreamHealthCheck{}

	if val, ok := d.GetOk(key + "receive"); ok {
		receive := val.(string)
		if receive != "" {
			payload := &apploadbalancer.Payload{}
			payload.SetText(receive)
			healthCheck.SetReceive(payload)
		}
	}

	if val, ok := d.GetOk(key + "send"); ok {
		send := val.(string)
		if send != "" {
			payload := &apploadbalancer.Payload{}
			payload.SetText(send)
			healthCheck.SetSend(payload)
		}
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
		config, err := expandALBLoadBalancingConfig(v)
		if err != nil {
			return nil, err
		}
		backend.SetLoadBalancingConfig(config)
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
	if isPlural(gotSubnet, gotPrivateAddr) {
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

func flattenALBRateLimit(rateLimit *apploadbalancer.RateLimit) []map[string]interface{} {
	if rateLimit == nil {
		return nil
	}

	result := make(map[string]interface{})

	if allRequests := rateLimit.GetAllRequests(); allRequests != nil {
		result[allRequestsSchemaKey] = []map[string]interface{}{flattenALBLimit(allRequests)}
	}

	if requestsPerIP := rateLimit.GetRequestsPerIp(); requestsPerIP != nil {
		result[requestsPerIPSchemaKey] = []map[string]interface{}{flattenALBLimit(requestsPerIP)}
	}

	return []map[string]interface{}{result}
}

func flattenALBLimit(limit *apploadbalancer.RateLimit_Limit) map[string]interface{} {
	if limit == nil {
		return nil
	}

	result := make(map[string]interface{})

	if limit.GetPerSecond() != 0 {
		result[perSecondSchemaKey] = int(limit.GetPerSecond())
	}

	if limit.GetPerMinute() != 0 {
		result[perMinuteSchemaKey] = int(limit.GetPerMinute())
	}

	return result
}

func flattenALBRoutes(routes []*apploadbalancer.Route) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	for _, route := range routes {
		flRoute := map[string]interface{}{
			"name": route.Name,
		}

		ro, err := flattenALBRouteOptions(route.GetRouteOptions())
		if err != nil {
			return nil, err
		}
		flRoute["route_options"] = ro

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

func flattenALBRouteOptions(ro *apploadbalancer.RouteOptions) ([]map[string]interface{}, error) {
	if ro == nil {
		return nil, nil
	}
	flOptions := map[string]interface{}{}
	if ro.GetRbac() != nil {
		rbac, err := flattenALBRBAC(ro.GetRbac())
		if err != nil {
			return nil, err
		}
		flOptions["rbac"] = rbac
	}

	if ro.SecurityProfileId != "" {
		flOptions["security_profile_id"] = ro.SecurityProfileId
	}

	return []map[string]interface{}{flOptions}, nil
}

func flattenALBRBAC(rbac *apploadbalancer.RBAC) ([]map[string]interface{}, error) {
	var principals []map[string]interface{}

	for _, principal := range rbac.GetPrincipals() {
		pr, err := flattenALBPrincipals(principal)
		if err != nil {
			return nil, err
		}

		principals = append(principals, pr)
	}

	return []map[string]interface{}{{
		"action":     strings.ToLower(rbac.GetAction().String()),
		"principals": principals,
	}}, nil
}

func flattenALBPrincipals(principals *apploadbalancer.Principals) (map[string]interface{}, error) {
	var andPrincipals []map[string]interface{}

	for _, principal := range principals.GetAndPrincipals() {
		pr, err := flattenALBPrincipal(principal)
		if err != nil {
			return nil, err
		}

		andPrincipals = append(andPrincipals, pr)
	}

	return map[string]interface{}{
		"and_principals": andPrincipals,
	}, nil
}

func flattenALBPrincipal(principal *apploadbalancer.Principal) (map[string]interface{}, error) {
	flPrincipal := map[string]interface{}{}

	switch identifier := principal.GetIdentifier().(type) {
	case *apploadbalancer.Principal_Header:
		header, err := flattenALBHeaderMatcher(principal.GetHeader())
		if err != nil {
			return nil, err
		}
		flPrincipal["header"] = header
	case *apploadbalancer.Principal_RemoteIp:
		flPrincipal["remote_ip"] = principal.GetRemoteIp()
	case *apploadbalancer.Principal_Any:
		flPrincipal["any"] = principal.GetAny()
	default:
		return nil, fmt.Errorf("[ERROR] Unexpected Principal Identifier type %T!\n", identifier)
	}

	return flPrincipal, nil
}

func flattenALBHeaderMatcher(headerMatcher *apploadbalancer.Principal_HeaderMatcher) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{
		"name":  headerMatcher.GetName(),
		"value": flattenALBStringMatch(headerMatcher.GetValue()),
	}}, nil
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
				rateLimitSchemaKey: flattenALBRateLimit(routeAction.RateLimit),
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
	case *apploadbalancer.StringMatch_RegexMatch:
		return []map[string]interface{}{
			{
				"regex": match.GetRegexMatch(),
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
				rateLimitSchemaKey: flattenALBRateLimit(routeAction.RateLimit),
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
			"idle_timeout":     streamHandler.IdleTimeout,
		}

		return []interface{}{flHTTPHandler}
	}
	return []interface{}{}
}

func flattenALBHTTPHandler(httpHandler *apploadbalancer.HttpHandler) []interface{} {
	if httpHandler != nil {
		flHTTPHandler := map[string]interface{}{
			"http_router_id":     httpHandler.GetHttpRouterId(),
			"rewrite_request_id": httpHandler.GetRewriteRequestId(),
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
		"mode":                           lbConfig.GetMode().String(),
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
			keepConnectionsOnHostHealthFailureSchemaKey: b.GetKeepConnectionsOnHostHealthFailure(),
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

func flattenALBHealthChecks(healthChecks []*apploadbalancer.HealthCheck) []interface{} {
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
					"host":                    http.Host,
					"path":                    http.Path,
					"http2":                   http.UseHttp2,
					expectedStatusesSchemaKey: http.ExpectedStatuses,
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
			flStreamHealthcheck := map[string]interface{}{
				"send":    stream.GetSend().GetText(),
				"receive": stream.GetReceive().GetText(),
			}

			flHealthCheck["stream_healthcheck"] = []map[string]interface{}{flStreamHealthcheck}
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

func flattenALBLogOptions(alb *apploadbalancer.LoadBalancer) ([]map[string]interface{}, error) {
	v := alb.GetLogOptions()
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["disable"] = v.Disable
	discardRule, err := flattenAlbLogOptionsDiscardRuleSlice(v.DiscardRules)
	if err != nil {
		return nil, err
	}
	m["discard_rule"] = discardRule
	m["log_group_id"] = v.LogGroupId

	return []map[string]interface{}{m}, nil
}

func flattenAlbLogOptionsDiscardRuleSlice(vs []*apploadbalancer.LogDiscardRule) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		discardRule, err := flattenALBLogDiscardRule(v)
		if err != nil {
			return nil, err
		}

		if len(discardRule) != 0 {
			s = append(s, discardRule[0])
		}
	}

	return s, nil
}

func flattenALBLogDiscardRule(v *apploadbalancer.LogDiscardRule) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	var discardPercent interface{}
	if v.DiscardPercent != nil {
		discardPercent = v.DiscardPercent.GetValue()
	}
	m["discard_percent"] = discardPercent

	var grpcCodes []string
	for _, v := range v.GrpcCodes {
		grpcCodes = append(grpcCodes, v.String())
	}
	m["grpc_codes"] = grpcCodes

	var httpCodeIntervals []string
	for _, v := range v.HttpCodeIntervals {
		httpCodeIntervals = append(httpCodeIntervals, v.String())
	}
	m["http_code_intervals"] = httpCodeIntervals
	m["http_codes"] = v.HttpCodes

	return []map[string]interface{}{m}, nil
}

func invalidKeyError(name string, valid map[string]int32, got string) error {
	keys := make([]string, 0, len(valid))
	for k := range valid {
		keys = append(keys, k)
	}
	return fmt.Errorf("expected %q to be one of %v, got %q", name, keys, got)
}
