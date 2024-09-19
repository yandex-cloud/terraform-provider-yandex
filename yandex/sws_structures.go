package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	smartcaptcha "github.com/yandex-cloud/go-genproto/yandex/cloud/smartcaptcha/v1"
	smartwebsecurity "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1"
	advanced_rate_limiter "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1/advanced_rate_limiter"
	waf "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1/waf"
	"google.golang.org/protobuf/proto"
)

func parseSmartcaptchaCaptchaChallengeType(str string) (smartcaptcha.CaptchaChallengeType, error) {
	val, ok := smartcaptcha.CaptchaChallengeType_value[str]
	if !ok {
		return smartcaptcha.CaptchaChallengeType(0), invalidKeyError("captcha_challenge_type", smartcaptcha.CaptchaChallengeType_value, str)
	}
	return smartcaptcha.CaptchaChallengeType(val), nil
}

func parseSmartcaptchaCaptchaComplexity(str string) (smartcaptcha.CaptchaComplexity, error) {
	val, ok := smartcaptcha.CaptchaComplexity_value[str]
	if !ok {
		return smartcaptcha.CaptchaComplexity(0), invalidKeyError("captcha_complexity", smartcaptcha.CaptchaComplexity_value, str)
	}
	return smartcaptcha.CaptchaComplexity(val), nil
}

func parseSmartcaptchaCaptchaPreCheckType(str string) (smartcaptcha.CaptchaPreCheckType, error) {
	val, ok := smartcaptcha.CaptchaPreCheckType_value[str]
	if !ok {
		return smartcaptcha.CaptchaPreCheckType(0), invalidKeyError("captcha_pre_check_type", smartcaptcha.CaptchaPreCheckType_value, str)
	}
	return smartcaptcha.CaptchaPreCheckType(val), nil
}

func expandStrings(v []interface{}) ([]string, error) {
	s := make([]string, len(v))
	if v == nil {
		return s, nil
	}

	for i, val := range v {
		s[i] = val.(string)
	}

	return s, nil
}

func expandCaptchaSecurityRulesSlice(d *schema.ResourceData) ([]*smartcaptcha.SecurityRule, error) {
	count := d.Get("security_rule.#").(int)
	slice := make([]*smartcaptcha.SecurityRule, count)

	for i := 0; i < count; i++ {
		securityRules, err := expandCaptchaSecurityRules(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = securityRules
	}

	return slice, nil
}

func expandCaptchaSecurityRules(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.SecurityRule, error) {
	val := new(smartcaptcha.SecurityRule)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.priority", indexes...)); ok {
		val.SetPriority(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition", indexes...)); ok {
		condition, err := expandCaptchaSecurityRulesCondition(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.override_variant_uuid", indexes...)); ok {
		val.SetOverrideVariantUuid(v.(string))
	}

	return val, nil
}

func expandCaptchaSecurityRulesCondition(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition, error) {
	val := new(smartcaptcha.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host", indexes...)); ok {
		host, err := expandCaptchaSecurityRulesConditionHost(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHost(host)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri", indexes...)); ok {
		uri, err := expandCaptchaSecurityRulesConditionUri(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetUri(uri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers", indexes...)); ok {
		headers, err := expandCaptchaSecurityRulesConditionHeadersSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandCaptchaSecurityRulesConditionSourceIp(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionHost(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_HostMatcher, error) {
	val := new(smartcaptcha.Condition_HostMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts", indexes...)); ok {
		hosts, err := expandCaptchaSecurityRulesConditionHostHostsSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHosts(hosts)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionHostHostsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartcaptcha.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.#", indexes...)).(int)
	slice := make([]*smartcaptcha.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		hosts, err := expandCaptchaSecurityRulesConditionHostHosts(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = hosts
	}

	return slice, nil
}

func expandCaptchaSecurityRulesConditionHostHosts(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_StringMatcher, error) {
	val := new(smartcaptcha.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionUri(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_UriMatcher, error) {
	val := new(smartcaptcha.Condition_UriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path", indexes...)); ok {
		path, err := expandCaptchaSecurityRulesConditionUriPath(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries", indexes...)); ok {
		queries, err := expandCaptchaSecurityRulesConditionUriQueriesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionUriPath(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_StringMatcher, error) {
	val := new(smartcaptcha.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionUriQueriesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartcaptcha.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartcaptcha.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandCaptchaSecurityRulesConditionUriQueries(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandCaptchaSecurityRulesConditionUriQueries(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_QueryMatcher, error) {
	val := new(smartcaptcha.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandCaptchaSecurityRulesConditionUriQueriesValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionUriQueriesValue(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_StringMatcher, error) {
	val := new(smartcaptcha.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionHeadersSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartcaptcha.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartcaptcha.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandCaptchaSecurityRulesConditionHeaders(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandCaptchaSecurityRulesConditionHeaders(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_HeaderMatcher, error) {
	val := new(smartcaptcha.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandCaptchaSecurityRulesConditionHeadersValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionHeadersValue(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_StringMatcher, error) {
	val := new(smartcaptcha.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionSourceIp(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_IpMatcher, error) {
	val := new(smartcaptcha.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandCaptchaSecurityRulesConditionSourceIpIpRangesMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandCaptchaSecurityRulesConditionSourceIpIpRangesNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandCaptchaSecurityRulesConditionSourceIpGeoIpMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandCaptchaSecurityRulesConditionSourceIpGeoIpNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionSourceIpIpRangesMatch(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_IpRangesMatcher, error) {
	val := new(smartcaptcha.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionSourceIpIpRangesNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_IpRangesMatcher, error) {
	val := new(smartcaptcha.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionSourceIpGeoIpMatch(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_GeoIpMatcher, error) {
	val := new(smartcaptcha.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionSourceIpGeoIpNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_GeoIpMatcher, error) {
	val := new(smartcaptcha.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandCaptchaOverrideVariantsSlice(d *schema.ResourceData) ([]*smartcaptcha.OverrideVariant, error) {
	count := d.Get("override_variant.#").(int)
	slice := make([]*smartcaptcha.OverrideVariant, count)

	for i := 0; i < count; i++ {
		overrideVariants, err := expandCaptchaOverrideVariants(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = overrideVariants
	}

	return slice, nil
}

func expandCaptchaOverrideVariants(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.OverrideVariant, error) {
	val := new(smartcaptcha.OverrideVariant)

	if v, ok := d.GetOk(fmt.Sprintf("override_variant.%d.uuid", indexes...)); ok {
		val.SetUuid(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("override_variant.%d.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("override_variant.%d.complexity", indexes...)); ok {
		captchaComplexity, err := parseSmartcaptchaCaptchaComplexity(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetComplexity(captchaComplexity)
	}

	if v, ok := d.GetOk(fmt.Sprintf("override_variant.%d.pre_check_type", indexes...)); ok {
		captchaPreCheckType, err := parseSmartcaptchaCaptchaPreCheckType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetPreCheckType(captchaPreCheckType)
	}

	if v, ok := d.GetOk(fmt.Sprintf("override_variant.%d.challenge_type", indexes...)); ok {
		captchaChallengeType, err := parseSmartcaptchaCaptchaChallengeType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetChallengeType(captchaChallengeType)
	}

	return val, nil
}

func flattenSmartcaptchaOverrideVariantSlice(vs []*smartcaptcha.OverrideVariant) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		overrideVariant, err := flatten_yandex_cloud_smartcaptcha_v1_OverrideVariant(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(overrideVariant) != 0 {
			s = append(s, overrideVariant[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartcaptcha_v1_OverrideVariant(v *smartcaptcha.OverrideVariant) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["challenge_type"] = v.ChallengeType.String()
	m["complexity"] = v.Complexity.String()
	m["description"] = v.Description
	m["pre_check_type"] = v.PreCheckType.String()
	m["uuid"] = v.Uuid

	return []map[string]interface{}{m}, nil
}

func flattenSmartcaptchaSecurityRuleSlice(vs []*smartcaptcha.SecurityRule) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		securityRule, err := flatten_yandex_cloud_smartcaptcha_v1_SecurityRule(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(securityRule) != 0 {
			s = append(s, securityRule[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartcaptcha_v1_SecurityRule(v *smartcaptcha.SecurityRule) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	condition, err := flatten_yandex_cloud_smartcaptcha_v1_Condition(v.Condition)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["condition"] = condition
	m["description"] = v.Description
	m["name"] = v.Name
	m["override_variant_uuid"] = v.OverrideVariantUuid
	m["priority"] = v.Priority

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartcaptcha_v1_Condition(v *smartcaptcha.Condition) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	headers, err := flattenSmartcaptchaSecurityRuleconditionheadersSlice(v.Headers)
	if err != nil { // isElem: false, ret: 2
		return nil, err
	}
	m["headers"] = headers
	host, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_HostMatcher(v.Host)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["host"] = host
	sourceIp, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_IpMatcher(v.SourceIp)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["source_ip"] = sourceIp
	uri, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_UriMatcher(v.Uri)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["uri"] = uri

	return []map[string]interface{}{m}, nil
}

func flattenSmartcaptchaSecurityRuleconditionheadersSlice(vs []*smartcaptcha.Condition_HeaderMatcher) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		headers, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_HeaderMatcher(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(headers) != 0 {
			s = append(s, headers[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartcaptcha_v1_Condition_HeaderMatcher(v *smartcaptcha.Condition_HeaderMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["name"] = v.Name
	value, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_StringMatcher(v.Value)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["value"] = value

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartcaptcha_v1_Condition_StringMatcher(v *smartcaptcha.Condition_StringMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exact_match"] = v.GetExactMatch()
	m["exact_not_match"] = v.GetExactNotMatch()
	m["pire_regex_match"] = v.GetPireRegexMatch()
	m["pire_regex_not_match"] = v.GetPireRegexNotMatch()
	m["prefix_match"] = v.GetPrefixMatch()
	m["prefix_not_match"] = v.GetPrefixNotMatch()

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartcaptcha_v1_Condition_HostMatcher(v *smartcaptcha.Condition_HostMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	hosts, err := flattenSmartcaptchaSecurityRuleconditionhosthostsSlice(v.Hosts)
	if err != nil { // isElem: false, ret: 2
		return nil, err
	}
	m["hosts"] = hosts

	return []map[string]interface{}{m}, nil
}

func flattenSmartcaptchaSecurityRuleconditionhosthostsSlice(vs []*smartcaptcha.Condition_StringMatcher) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		hosts, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_StringMatcher(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(hosts) != 0 {
			s = append(s, hosts[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartcaptcha_v1_Condition_IpMatcher(v *smartcaptcha.Condition_IpMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	geoIpMatch, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_GeoIpMatcher(v.GeoIpMatch)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["geo_ip_match"] = geoIpMatch
	geoIpNotMatch, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_GeoIpMatcher(v.GeoIpNotMatch)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["geo_ip_not_match"] = geoIpNotMatch
	ipRangesMatch, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_IpRangesMatcher(v.IpRangesMatch)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["ip_ranges_match"] = ipRangesMatch
	ipRangesNotMatch, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_IpRangesMatcher(v.IpRangesNotMatch)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["ip_ranges_not_match"] = ipRangesNotMatch

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartcaptcha_v1_Condition_GeoIpMatcher(v *smartcaptcha.Condition_GeoIpMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["locations"] = v.Locations

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartcaptcha_v1_Condition_IpRangesMatcher(v *smartcaptcha.Condition_IpRangesMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ip_ranges"] = v.IpRanges

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartcaptcha_v1_Condition_UriMatcher(v *smartcaptcha.Condition_UriMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	path, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_StringMatcher(v.Path)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["path"] = path
	queries, err := flattenSmartcaptchaSecurityRuleconditionuriqueriesSlice(v.Queries)
	if err != nil { // isElem: false, ret: 2
		return nil, err
	}
	m["queries"] = queries

	return []map[string]interface{}{m}, nil
}

func flattenSmartcaptchaSecurityRuleconditionuriqueriesSlice(vs []*smartcaptcha.Condition_QueryMatcher) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		queries, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_QueryMatcher(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(queries) != 0 {
			s = append(s, queries[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartcaptcha_v1_Condition_QueryMatcher(v *smartcaptcha.Condition_QueryMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["key"] = v.Key
	value, err := flatten_yandex_cloud_smartcaptcha_v1_Condition_StringMatcher(v.Value)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["value"] = value

	return []map[string]interface{}{m}, nil
}

func expandCaptchaSecurityRulesSlice_(d *schema.ResourceData) ([]*smartcaptcha.SecurityRule, error) {
	count := d.Get("security_rule.#").(int)
	slice := make([]*smartcaptcha.SecurityRule, count)

	for i := 0; i < count; i++ {
		securityRules, err := expandCaptchaSecurityRules_(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = securityRules
	}

	return slice, nil
}

func expandCaptchaSecurityRules_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.SecurityRule, error) {
	val := new(smartcaptcha.SecurityRule)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.priority", indexes...)); ok {
		val.SetPriority(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition", indexes...)); ok {
		condition, err := expandCaptchaSecurityRulesCondition_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.override_variant_uuid", indexes...)); ok {
		val.SetOverrideVariantUuid(v.(string))
	}

	return val, nil
}

func expandCaptchaSecurityRulesCondition_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition, error) {
	val := new(smartcaptcha.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host", indexes...)); ok {
		host, err := expandCaptchaSecurityRulesConditionHost_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHost(host)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri", indexes...)); ok {
		uri, err := expandCaptchaSecurityRulesConditionUri_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetUri(uri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers", indexes...)); ok {
		headers, err := expandCaptchaSecurityRulesConditionHeadersSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandCaptchaSecurityRulesConditionSourceIp_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionHost_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_HostMatcher, error) {
	val := new(smartcaptcha.Condition_HostMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts", indexes...)); ok {
		hosts, err := expandCaptchaSecurityRulesConditionHostHostsSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHosts(hosts)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionHostHostsSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartcaptcha.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.#", indexes...)).(int)
	slice := make([]*smartcaptcha.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		hosts, err := expandCaptchaSecurityRulesConditionHostHosts_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = hosts
	}

	return slice, nil
}

func expandCaptchaSecurityRulesConditionHostHosts_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_StringMatcher, error) {
	val := new(smartcaptcha.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.host.0.hosts.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionUri_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_UriMatcher, error) {
	val := new(smartcaptcha.Condition_UriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path", indexes...)); ok {
		path, err := expandCaptchaSecurityRulesConditionUriPath_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries", indexes...)); ok {
		queries, err := expandCaptchaSecurityRulesConditionUriQueriesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionUriPath_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_StringMatcher, error) {
	val := new(smartcaptcha.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionUriQueriesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartcaptcha.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartcaptcha.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandCaptchaSecurityRulesConditionUriQueries_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandCaptchaSecurityRulesConditionUriQueries_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_QueryMatcher, error) {
	val := new(smartcaptcha.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandCaptchaSecurityRulesConditionUriQueriesValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionUriQueriesValue_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_StringMatcher, error) {
	val := new(smartcaptcha.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionHeadersSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartcaptcha.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartcaptcha.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandCaptchaSecurityRulesConditionHeaders_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandCaptchaSecurityRulesConditionHeaders_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_HeaderMatcher, error) {
	val := new(smartcaptcha.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandCaptchaSecurityRulesConditionHeadersValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionHeadersValue_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_StringMatcher, error) {
	val := new(smartcaptcha.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionSourceIp_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_IpMatcher, error) {
	val := new(smartcaptcha.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandCaptchaSecurityRulesConditionSourceIpIpRangesMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandCaptchaSecurityRulesConditionSourceIpIpRangesNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandCaptchaSecurityRulesConditionSourceIpGeoIpMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandCaptchaSecurityRulesConditionSourceIpGeoIpNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionSourceIpIpRangesMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_IpRangesMatcher, error) {
	val := new(smartcaptcha.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionSourceIpIpRangesNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_IpRangesMatcher, error) {
	val := new(smartcaptcha.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionSourceIpGeoIpMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_GeoIpMatcher, error) {
	val := new(smartcaptcha.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandCaptchaSecurityRulesConditionSourceIpGeoIpNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.Condition_GeoIpMatcher, error) {
	val := new(smartcaptcha.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandCaptchaOverrideVariantsSlice_(d *schema.ResourceData) ([]*smartcaptcha.OverrideVariant, error) {
	count := d.Get("override_variant.#").(int)
	slice := make([]*smartcaptcha.OverrideVariant, count)

	for i := 0; i < count; i++ {
		overrideVariants, err := expandCaptchaOverrideVariants_(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = overrideVariants
	}

	return slice, nil
}

func expandCaptchaOverrideVariants_(d *schema.ResourceData, indexes ...interface{}) (*smartcaptcha.OverrideVariant, error) {
	val := new(smartcaptcha.OverrideVariant)

	if v, ok := d.GetOk(fmt.Sprintf("override_variant.%d.uuid", indexes...)); ok {
		val.SetUuid(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("override_variant.%d.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("override_variant.%d.complexity", indexes...)); ok {
		captchaComplexity, err := parseSmartcaptchaCaptchaComplexity(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetComplexity(captchaComplexity)
	}

	if v, ok := d.GetOk(fmt.Sprintf("override_variant.%d.pre_check_type", indexes...)); ok {
		captchaPreCheckType, err := parseSmartcaptchaCaptchaPreCheckType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetPreCheckType(captchaPreCheckType)
	}

	if v, ok := d.GetOk(fmt.Sprintf("override_variant.%d.challenge_type", indexes...)); ok {
		captchaChallengeType, err := parseSmartcaptchaCaptchaChallengeType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetChallengeType(captchaChallengeType)
	}

	return val, nil
}

func parseAdvancedXrateXlimiterAdvancedRateLimiterRuleXAction(str string) (advanced_rate_limiter.AdvancedRateLimiterRule_Action, error) {
	val, ok := advanced_rate_limiter.AdvancedRateLimiterRule_Action_value[str]
	if !ok {
		return advanced_rate_limiter.AdvancedRateLimiterRule_Action(0), invalidKeyError("action", advanced_rate_limiter.AdvancedRateLimiterRule_Action_value, str)
	}
	return advanced_rate_limiter.AdvancedRateLimiterRule_Action(val), nil
}

func parseAdvancedXrateXlimiterAdvancedRateLimiterRuleXDynamicQuotaXCharacteristicXKeyCharacteristicXType(str string) (advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic_Type, error) {
	val, ok := advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic_Type_value[str]
	if !ok {
		return advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic_Type(0), invalidKeyError("type", advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic_Type_value, str)
	}
	return advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic_Type(val), nil
}

func parseAdvancedXrateXlimiterAdvancedRateLimiterRuleXDynamicQuotaXCharacteristicXSimpleCharacteristicXType(str string) (advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic_Type, error) {
	val, ok := advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic_Type_value[str]
	if !ok {
		return advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic_Type(0), invalidKeyError("type", advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic_Type_value, str)
	}
	return advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic_Type(val), nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesSlice(d *schema.ResourceData) ([]*advanced_rate_limiter.AdvancedRateLimiterRule, error) {
	count := d.Get("advanced_rate_limiter_rule.#").(int)
	slice := make([]*advanced_rate_limiter.AdvancedRateLimiterRule, count)

	for i := 0; i < count; i++ {
		advancedRateLimiterRules, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRules(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = advancedRateLimiterRules
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRules(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.priority", indexes...)); ok {
		val.SetPriority(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dry_run", indexes...)); ok {
		val.SetDryRun(v.(bool))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota", indexes...)); ok {
		staticQuota, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuota(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetStaticQuota(staticQuota)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota", indexes...)); ok {
		dynamicQuota, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuota(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetDynamicQuota(dynamicQuota)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuota(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule_StaticQuota, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule_StaticQuota)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.action", indexes...)); ok {
		action, err := parseAdvancedXrateXlimiterAdvancedRateLimiterRuleXAction(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetAction(action)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition", indexes...)); ok {
		condition, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaCondition(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.limit", indexes...)); ok {
		val.SetLimit(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.period", indexes...)); ok {
		val.SetPeriod(int64(v.(int)))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaCondition(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority", indexes...)); ok {
		authority, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthority(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethod(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUri(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers", indexes...)); ok {
		headers, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeadersSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIp(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthority(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthorityAuthoritiesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthorityAuthoritiesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthorityAuthorities(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthorityAuthorities(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethod(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethodHttpMethodsSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethodHttpMethodsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethodHttpMethods(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethodHttpMethods(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUri(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriPath(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueriesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriPath(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueriesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueries(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueries(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueriesValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueriesValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeadersSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeaders(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeaders(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeadersValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeadersValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIp(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpIpRangesMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpIpRangesNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpGeoIpMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpGeoIpNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpIpRangesMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpIpRangesNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpGeoIpMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpGeoIpNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuota(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.action", indexes...)); ok {
		action, err := parseAdvancedXrateXlimiterAdvancedRateLimiterRuleXAction(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetAction(action)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition", indexes...)); ok {
		condition, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCondition(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.limit", indexes...)); ok {
		val.SetLimit(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.period", indexes...)); ok {
		val.SetPeriod(int64(v.(int)))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic", indexes...)); ok {
		characteristics, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCharacteristics(characteristics)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCondition(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority", indexes...)); ok {
		authority, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthority(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethod(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUri(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers", indexes...)); ok {
		headers, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeadersSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIp(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthority(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthorityAuthoritiesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthorityAuthoritiesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthorityAuthorities(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthorityAuthorities(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethod(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethodHttpMethodsSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethodHttpMethodsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethodHttpMethods(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethodHttpMethods(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUri(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriPath(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueriesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriPath(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueriesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueries(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueries(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueriesValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueriesValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeadersSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeaders(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeaders(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeadersValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeadersValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIp(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpIpRangesMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpIpRangesNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpGeoIpMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpGeoIpNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpIpRangesMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpIpRangesNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpGeoIpMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpGeoIpNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.#", indexes...)).(int)
	slice := make([]*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic, count)

	for i := 0; i < count; i++ {
		characteristics, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristics(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = characteristics
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristics(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.simple_characteristic", indexes...)); ok {
		simpleCharacteristic, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsSimpleCharacteristic(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSimpleCharacteristic(simpleCharacteristic)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.key_characteristic", indexes...)); ok {
		keyCharacteristic, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsKeyCharacteristic(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetKeyCharacteristic(keyCharacteristic)
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.case_insensitive", indexes...)); ok {
		val.SetCaseInsensitive(v.(bool))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsSimpleCharacteristic(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.simple_characteristic.0.type", indexes...)); ok {
		simpleCharacteristicType, err := parseAdvancedXrateXlimiterAdvancedRateLimiterRuleXDynamicQuotaXCharacteristicXSimpleCharacteristicXType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(simpleCharacteristicType)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsKeyCharacteristic(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.key_characteristic.0.type", indexes...)); ok {
		keyCharacteristicType, err := parseAdvancedXrateXlimiterAdvancedRateLimiterRuleXDynamicQuotaXCharacteristicXKeyCharacteristicXType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(keyCharacteristicType)
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.key_characteristic.0.value", indexes...)); ok {
		val.SetValue(v.(string))
	}

	return val, nil
}

func flattenAdvancedXrateXlimiterAdvancedRateLimiterRuleSlice(vs []*advanced_rate_limiter.AdvancedRateLimiterRule) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		advancedRateLimiterRule, err := flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(advancedRateLimiterRule) != 0 {
			s = append(s, advancedRateLimiterRule[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule(v *advanced_rate_limiter.AdvancedRateLimiterRule) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["description"] = v.Description
	m["dry_run"] = v.DryRun
	dynamicQuota, err := flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule_DynamicQuota(v.GetDynamicQuota())
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["dynamic_quota"] = dynamicQuota
	m["name"] = v.Name
	m["priority"] = v.Priority
	staticQuota, err := flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule_StaticQuota(v.GetStaticQuota())
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["static_quota"] = staticQuota

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule_DynamicQuota(v *advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["action"] = v.Action.String()
	characteristic, err := flattenAdvancedXrateXlimiterAdvancedRateLimiterRuledynamicQuotacharacteristicSlice(v.Characteristics)
	if err != nil { // isElem: false, ret: 2
		return nil, err
	}
	m["characteristic"] = characteristic
	condition, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition(v.Condition)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["condition"] = condition
	m["limit"] = v.Limit
	m["period"] = v.Period

	return []map[string]interface{}{m}, nil
}

func flattenAdvancedXrateXlimiterAdvancedRateLimiterRuledynamicQuotacharacteristicSlice(vs []*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		characteristic, err := flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule_DynamicQuota_Characteristic(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(characteristic) != 0 {
			s = append(s, characteristic[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule_DynamicQuota_Characteristic(v *advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["case_insensitive"] = v.CaseInsensitive
	keyCharacteristic, err := flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic(v.GetKeyCharacteristic())
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["key_characteristic"] = keyCharacteristic
	simpleCharacteristic, err := flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic(v.GetSimpleCharacteristic())
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["simple_characteristic"] = simpleCharacteristic

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic(v *advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["type"] = v.Type.String()
	m["value"] = v.Value

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic(v *advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["type"] = v.Type.String()

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_Condition(v *smartwebsecurity.Condition) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	authority, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_AuthorityMatcher(v.Authority)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["authority"] = authority
	headers, err := flattenSmartwebsecurityAdvancedRateLimiterRuledynamicQuotaconditionheadersSlice(v.Headers)
	if err != nil { // isElem: false, ret: 2
		return nil, err
	}
	m["headers"] = headers
	httpMethod, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_HttpMethodMatcher(v.HttpMethod)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["http_method"] = httpMethod
	requestUri, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_RequestUriMatcher(v.RequestUri)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["request_uri"] = requestUri
	sourceIp, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_IpMatcher(v.SourceIp)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["source_ip"] = sourceIp

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_Condition_AuthorityMatcher(v *smartwebsecurity.Condition_AuthorityMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	authorities, err := flattenSmartwebsecurityAdvancedRateLimiterRuledynamicQuotaconditionauthorityauthoritiesSlice(v.Authorities)
	if err != nil { // isElem: false, ret: 2
		return nil, err
	}
	m["authorities"] = authorities

	return []map[string]interface{}{m}, nil
}

func flattenSmartwebsecurityAdvancedRateLimiterRuledynamicQuotaconditionauthorityauthoritiesSlice(vs []*smartwebsecurity.Condition_StringMatcher) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		authorities, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_StringMatcher(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(authorities) != 0 {
			s = append(s, authorities[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_Condition_StringMatcher(v *smartwebsecurity.Condition_StringMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exact_match"] = v.GetExactMatch()
	m["exact_not_match"] = v.GetExactNotMatch()
	m["pire_regex_match"] = v.GetPireRegexMatch()
	m["pire_regex_not_match"] = v.GetPireRegexNotMatch()
	m["prefix_match"] = v.GetPrefixMatch()
	m["prefix_not_match"] = v.GetPrefixNotMatch()

	return []map[string]interface{}{m}, nil
}

func flattenSmartwebsecurityAdvancedRateLimiterRuledynamicQuotaconditionheadersSlice(vs []*smartwebsecurity.Condition_HeaderMatcher) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		headers, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_HeaderMatcher(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(headers) != 0 {
			s = append(s, headers[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_Condition_HeaderMatcher(v *smartwebsecurity.Condition_HeaderMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["name"] = v.Name
	value, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_StringMatcher(v.Value)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["value"] = value

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_Condition_HttpMethodMatcher(v *smartwebsecurity.Condition_HttpMethodMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	httpMethods, err := flattenSmartwebsecurityAdvancedRateLimiterRuledynamicQuotaconditionhttpMethodhttpMethodsSlice(v.HttpMethods)
	if err != nil { // isElem: false, ret: 2
		return nil, err
	}
	m["http_methods"] = httpMethods

	return []map[string]interface{}{m}, nil
}

func flattenSmartwebsecurityAdvancedRateLimiterRuledynamicQuotaconditionhttpMethodhttpMethodsSlice(vs []*smartwebsecurity.Condition_StringMatcher) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		httpMethods, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_StringMatcher(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(httpMethods) != 0 {
			s = append(s, httpMethods[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_Condition_RequestUriMatcher(v *smartwebsecurity.Condition_RequestUriMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	path, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_StringMatcher(v.Path)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["path"] = path
	queries, err := flattenSmartwebsecurityAdvancedRateLimiterRuledynamicQuotaconditionrequestUriqueriesSlice(v.Queries)
	if err != nil { // isElem: false, ret: 2
		return nil, err
	}
	m["queries"] = queries

	return []map[string]interface{}{m}, nil
}

func flattenSmartwebsecurityAdvancedRateLimiterRuledynamicQuotaconditionrequestUriqueriesSlice(vs []*smartwebsecurity.Condition_QueryMatcher) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		queries, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_QueryMatcher(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(queries) != 0 {
			s = append(s, queries[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_Condition_QueryMatcher(v *smartwebsecurity.Condition_QueryMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["key"] = v.Key
	value, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_StringMatcher(v.Value)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["value"] = value

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_Condition_IpMatcher(v *smartwebsecurity.Condition_IpMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	geoIpMatch, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_GeoIpMatcher(v.GeoIpMatch)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["geo_ip_match"] = geoIpMatch
	geoIpNotMatch, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_GeoIpMatcher(v.GeoIpNotMatch)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["geo_ip_not_match"] = geoIpNotMatch
	ipRangesMatch, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_IpRangesMatcher(v.IpRangesMatch)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["ip_ranges_match"] = ipRangesMatch
	ipRangesNotMatch, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition_IpRangesMatcher(v.IpRangesNotMatch)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["ip_ranges_not_match"] = ipRangesNotMatch

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_Condition_GeoIpMatcher(v *smartwebsecurity.Condition_GeoIpMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["locations"] = v.Locations

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_Condition_IpRangesMatcher(v *smartwebsecurity.Condition_IpRangesMatcher) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["ip_ranges"] = v.IpRanges

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_advanced_rate_limiter_AdvancedRateLimiterRule_StaticQuota(v *advanced_rate_limiter.AdvancedRateLimiterRule_StaticQuota) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["action"] = v.Action.String()
	condition, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition(v.Condition)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["condition"] = condition
	m["limit"] = v.Limit
	m["period"] = v.Period

	return []map[string]interface{}{m}, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesSlice_(d *schema.ResourceData) ([]*advanced_rate_limiter.AdvancedRateLimiterRule, error) {
	count := d.Get("advanced_rate_limiter_rule.#").(int)
	slice := make([]*advanced_rate_limiter.AdvancedRateLimiterRule, count)

	for i := 0; i < count; i++ {
		advancedRateLimiterRules, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRules_(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = advancedRateLimiterRules
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRules_(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.priority", indexes...)); ok {
		val.SetPriority(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dry_run", indexes...)); ok {
		val.SetDryRun(v.(bool))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota", indexes...)); ok {
		staticQuota, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuota_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetStaticQuota(staticQuota)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota", indexes...)); ok {
		dynamicQuota, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuota_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetDynamicQuota(dynamicQuota)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuota_(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule_StaticQuota, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule_StaticQuota)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.action", indexes...)); ok {
		action, err := parseAdvancedXrateXlimiterAdvancedRateLimiterRuleXAction(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetAction(action)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition", indexes...)); ok {
		condition, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaCondition_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.limit", indexes...)); ok {
		val.SetLimit(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.period", indexes...)); ok {
		val.SetPeriod(int64(v.(int)))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaCondition_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority", indexes...)); ok {
		authority, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthority_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethod_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUri_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers", indexes...)); ok {
		headers, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeadersSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIp_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthority_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthorityAuthoritiesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthorityAuthoritiesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthorityAuthorities_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionAuthorityAuthorities_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethod_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethodHttpMethodsSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethodHttpMethodsSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethodHttpMethods_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHttpMethodHttpMethods_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUri_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriPath_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueriesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriPath_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueriesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueries_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueries_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueriesValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionRequestUriQueriesValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeadersSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeaders_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeaders_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeadersValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionHeadersValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIp_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpIpRangesMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpIpRangesNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpGeoIpMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpGeoIpNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpIpRangesMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpIpRangesNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpGeoIpMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesStaticQuotaConditionSourceIpGeoIpNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.static_quota.0.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuota_(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.action", indexes...)); ok {
		action, err := parseAdvancedXrateXlimiterAdvancedRateLimiterRuleXAction(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetAction(action)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition", indexes...)); ok {
		condition, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCondition_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.limit", indexes...)); ok {
		val.SetLimit(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.period", indexes...)); ok {
		val.SetPeriod(int64(v.(int)))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic", indexes...)); ok {
		characteristics, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCharacteristics(characteristics)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCondition_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority", indexes...)); ok {
		authority, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthority_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethod_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUri_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers", indexes...)); ok {
		headers, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeadersSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIp_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthority_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthorityAuthoritiesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthorityAuthoritiesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthorityAuthorities_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionAuthorityAuthorities_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethod_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethodHttpMethodsSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethodHttpMethodsSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethodHttpMethods_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHttpMethodHttpMethods_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUri_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriPath_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueriesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriPath_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueriesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueries_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueries_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueriesValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionRequestUriQueriesValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeadersSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeaders_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeaders_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeadersValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionHeadersValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIp_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpIpRangesMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpIpRangesNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpGeoIpMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpGeoIpNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpIpRangesMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpIpRangesNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpGeoIpMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaConditionSourceIpGeoIpNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic, error) {
	count := d.Get(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.#", indexes...)).(int)
	slice := make([]*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic, count)

	for i := 0; i < count; i++ {
		characteristics, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristics_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = characteristics
	}

	return slice, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristics_(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic)

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.simple_characteristic", indexes...)); ok {
		simpleCharacteristic, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsSimpleCharacteristic_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSimpleCharacteristic(simpleCharacteristic)
	}

	if _, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.key_characteristic", indexes...)); ok {
		keyCharacteristic, err := expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsKeyCharacteristic_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetKeyCharacteristic(keyCharacteristic)
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.case_insensitive", indexes...)); ok {
		val.SetCaseInsensitive(v.(bool))
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsSimpleCharacteristic_(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_SimpleCharacteristic)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.simple_characteristic.0.type", indexes...)); ok {
		simpleCharacteristicType, err := parseAdvancedXrateXlimiterAdvancedRateLimiterRuleXDynamicQuotaXCharacteristicXSimpleCharacteristicXType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(simpleCharacteristicType)
	}

	return val, nil
}

func expandAdvancedRateLimiterProfileAdvancedRateLimiterRulesDynamicQuotaCharacteristicsKeyCharacteristic_(d *schema.ResourceData, indexes ...interface{}) (*advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic, error) {
	val := new(advanced_rate_limiter.AdvancedRateLimiterRule_DynamicQuota_Characteristic_KeyCharacteristic)

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.key_characteristic.0.type", indexes...)); ok {
		keyCharacteristicType, err := parseAdvancedXrateXlimiterAdvancedRateLimiterRuleXDynamicQuotaXCharacteristicXKeyCharacteristicXType(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetType(keyCharacteristicType)
	}

	if v, ok := d.GetOk(fmt.Sprintf("advanced_rate_limiter_rule.%d.dynamic_quota.0.characteristic.%d.key_characteristic.0.value", indexes...)); ok {
		val.SetValue(v.(string))
	}

	return val, nil
}

func flattenWafRulesSlice(vs []*waf.RuleSetDescriptor_RuleDescription) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		rules, err := flatten_yandex_cloud_smartwebsecurity_v1_waf_RuleSetDescriptor_RuleDescription(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(rules) != 0 {
			s = append(s, rules[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_waf_RuleSetDescriptor_RuleDescription(v *waf.RuleSetDescriptor_RuleDescription) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["anomaly_score"] = v.AnomalyScore
	m["id"] = v.Id
	m["paranoia_level"] = v.ParanoiaLevel

	return []map[string]interface{}{m}, nil
}

func parseWafWafProfileXAnalyzeRequestBodyXAction(str string) (waf.WafProfile_AnalyzeRequestBody_Action, error) {
	val, ok := waf.WafProfile_AnalyzeRequestBody_Action_value[str]
	if !ok {
		return waf.WafProfile_AnalyzeRequestBody_Action(0), invalidKeyError("action", waf.WafProfile_AnalyzeRequestBody_Action_value, str)
	}
	return waf.WafProfile_AnalyzeRequestBody_Action(val), nil
}

func expandWafProfileRulesSlice(d *schema.ResourceData) ([]*waf.WafProfileRule, error) {
	count := d.Get("rule.#").(int)
	slice := make([]*waf.WafProfileRule, count)

	for i := 0; i < count; i++ {
		rules, err := expandWafProfileRules(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = rules
	}

	return slice, nil
}

func expandWafProfileRules(d *schema.ResourceData, indexes ...interface{}) (*waf.WafProfileRule, error) {
	val := new(waf.WafProfileRule)

	if v, ok := d.GetOk(fmt.Sprintf("rule.%d.rule_id", indexes...)); ok {
		val.SetRuleId(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("rule.%d.is_enabled", indexes...)); ok {
		val.SetIsEnabled(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("rule.%d.is_blocking", indexes...)); ok {
		val.SetIsBlocking(v.(bool))
	}

	return val, nil
}

func expandWafProfileExclusionRulesSlice(d *schema.ResourceData) ([]*waf.WafProfileExclusionRule, error) {
	count := d.Get("exclusion_rule.#").(int)
	slice := make([]*waf.WafProfileExclusionRule, count)

	for i := 0; i < count; i++ {
		exclusionRules, err := expandWafProfileExclusionRules(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = exclusionRules
	}

	return slice, nil
}

func expandWafProfileExclusionRules(d *schema.ResourceData, indexes ...interface{}) (*waf.WafProfileExclusionRule, error) {
	val := new(waf.WafProfileExclusionRule)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition", indexes...)); ok {
		condition, err := expandWafProfileExclusionRulesCondition(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.exclude_rules", indexes...)); ok {
		excludeRules, err := expandWafProfileExclusionRulesExcludeRules(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetExcludeRules(excludeRules)
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.log_excluded", indexes...)); ok {
		val.SetLogExcluded(v.(bool))
	}

	return val, nil
}

func expandWafProfileExclusionRulesCondition(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority", indexes...)); ok {
		authority, err := expandWafProfileExclusionRulesConditionAuthority(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandWafProfileExclusionRulesConditionHttpMethod(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandWafProfileExclusionRulesConditionRequestUri(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers", indexes...)); ok {
		headers, err := expandWafProfileExclusionRulesConditionHeadersSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandWafProfileExclusionRulesConditionSourceIp(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionAuthority(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandWafProfileExclusionRulesConditionAuthorityAuthoritiesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionAuthorityAuthoritiesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandWafProfileExclusionRulesConditionAuthorityAuthorities(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandWafProfileExclusionRulesConditionAuthorityAuthorities(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionHttpMethod(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandWafProfileExclusionRulesConditionHttpMethodHttpMethodsSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionHttpMethodHttpMethodsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandWafProfileExclusionRulesConditionHttpMethodHttpMethods(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandWafProfileExclusionRulesConditionHttpMethodHttpMethods(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionRequestUri(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandWafProfileExclusionRulesConditionRequestUriPath(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandWafProfileExclusionRulesConditionRequestUriQueriesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionRequestUriPath(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionRequestUriQueriesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandWafProfileExclusionRulesConditionRequestUriQueries(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandWafProfileExclusionRulesConditionRequestUriQueries(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandWafProfileExclusionRulesConditionRequestUriQueriesValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionRequestUriQueriesValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionHeadersSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandWafProfileExclusionRulesConditionHeaders(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandWafProfileExclusionRulesConditionHeaders(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandWafProfileExclusionRulesConditionHeadersValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionHeadersValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionSourceIp(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandWafProfileExclusionRulesConditionSourceIpIpRangesMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandWafProfileExclusionRulesConditionSourceIpIpRangesNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandWafProfileExclusionRulesConditionSourceIpGeoIpMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandWafProfileExclusionRulesConditionSourceIpGeoIpNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionSourceIpIpRangesMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionSourceIpIpRangesNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionSourceIpGeoIpMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionSourceIpGeoIpNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandWafProfileExclusionRulesExcludeRules(d *schema.ResourceData, indexes ...interface{}) (*waf.WafProfileExclusionRule_ExcludeRules, error) {
	val := new(waf.WafProfileExclusionRule_ExcludeRules)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.exclude_rules.0.exclude_all", indexes...)); ok {
		val.SetExcludeAll(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.exclude_rules.0.rule_ids", indexes...)); ok {
		ruleIds, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetRuleIds(ruleIds)
	}

	return val, nil
}

func expandWafProfileAnalyzeRequestBody(d *schema.ResourceData) (*waf.WafProfile_AnalyzeRequestBody, error) {
	val := new(waf.WafProfile_AnalyzeRequestBody)

	if v, ok := d.GetOk("analyze_request_body.0.is_enabled"); ok {
		val.SetIsEnabled(v.(bool))
	}

	if v, ok := d.GetOk("analyze_request_body.0.size_limit"); ok {
		val.SetSizeLimit(int64(v.(int)))
	}

	if v, ok := d.GetOk("analyze_request_body.0.size_limit_action"); ok {
		action, err := parseWafWafProfileXAnalyzeRequestBodyXAction(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetSizeLimitAction(action)
	}

	empty := new(waf.WafProfile_AnalyzeRequestBody)
	if proto.Equal(val, empty) {
		return nil, nil
	}

	return val, nil
}

func expandWafProfileCoreRuleSet(d *schema.ResourceData) (*waf.WafProfile_CoreRuleSet, error) {
	val := new(waf.WafProfile_CoreRuleSet)

	if v, ok := d.GetOk("core_rule_set.0.inbound_anomaly_score"); ok {
		val.SetInboundAnomalyScore(int64(v.(int)))
	}

	if v, ok := d.GetOk("core_rule_set.0.paranoia_level"); ok {
		val.SetParanoiaLevel(int64(v.(int)))
	}

	if _, ok := d.GetOk("core_rule_set.0.rule_set"); ok {
		ruleSet, err := expandWafProfileCoreRuleSetRuleSet(d)
		if err != nil {
			return nil, err
		}

		val.SetRuleSet(ruleSet)
	}

	empty := new(waf.WafProfile_CoreRuleSet)
	if proto.Equal(val, empty) {
		return nil, nil
	}

	return val, nil
}

func expandWafProfileCoreRuleSetRuleSet(d *schema.ResourceData) (*waf.RuleSet, error) {
	val := new(waf.RuleSet)

	if v, ok := d.GetOk("core_rule_set.0.rule_set.0.name"); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk("core_rule_set.0.rule_set.0.version"); ok {
		val.SetVersion(v.(string))
	}

	return val, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfile_AnalyzeRequestBody(v *waf.WafProfile_AnalyzeRequestBody) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["is_enabled"] = v.IsEnabled
	m["size_limit"] = v.SizeLimit
	m["size_limit_action"] = v.SizeLimitAction.String()

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfile_CoreRuleSet(v *waf.WafProfile_CoreRuleSet) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["inbound_anomaly_score"] = v.InboundAnomalyScore
	m["paranoia_level"] = v.ParanoiaLevel
	ruleSet, err := flatten_yandex_cloud_smartwebsecurity_v1_waf_RuleSet(v.RuleSet)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["rule_set"] = ruleSet

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_waf_RuleSet(v *waf.RuleSet) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["name"] = v.Name
	m["version"] = v.Version

	return []map[string]interface{}{m}, nil
}

func flattenWafExclusionRuleSlice(vs []*waf.WafProfileExclusionRule) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		exclusionRule, err := flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfileExclusionRule(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(exclusionRule) != 0 {
			s = append(s, exclusionRule[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfileExclusionRule(v *waf.WafProfileExclusionRule) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	condition, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition(v.Condition)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["condition"] = condition
	m["description"] = v.Description
	excludeRules, err := flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfileExclusionRule_ExcludeRules(v.ExcludeRules)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["exclude_rules"] = excludeRules
	m["log_excluded"] = v.LogExcluded
	m["name"] = v.Name

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfileExclusionRule_ExcludeRules(v *waf.WafProfileExclusionRule_ExcludeRules) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["exclude_all"] = v.ExcludeAll
	m["rule_ids"] = v.RuleIds

	return []map[string]interface{}{m}, nil
}

func flattenWafRuleSlice(vs []*waf.WafProfileRule) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		rule, err := flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfileRule(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(rule) != 0 {
			s = append(s, rule[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_waf_WafProfileRule(v *waf.WafProfileRule) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["is_blocking"] = v.IsBlocking
	m["is_enabled"] = v.IsEnabled
	m["rule_id"] = v.RuleId

	return []map[string]interface{}{m}, nil
}

func expandWafProfileRulesSlice_(d *schema.ResourceData) ([]*waf.WafProfileRule, error) {
	count := d.Get("rule.#").(int)
	slice := make([]*waf.WafProfileRule, count)

	for i := 0; i < count; i++ {
		rules, err := expandWafProfileRules_(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = rules
	}

	return slice, nil
}

func expandWafProfileRules_(d *schema.ResourceData, indexes ...interface{}) (*waf.WafProfileRule, error) {
	val := new(waf.WafProfileRule)

	if v, ok := d.GetOk(fmt.Sprintf("rule.%d.rule_id", indexes...)); ok {
		val.SetRuleId(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("rule.%d.is_enabled", indexes...)); ok {
		val.SetIsEnabled(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("rule.%d.is_blocking", indexes...)); ok {
		val.SetIsBlocking(v.(bool))
	}

	return val, nil
}

func expandWafProfileExclusionRulesSlice_(d *schema.ResourceData) ([]*waf.WafProfileExclusionRule, error) {
	count := d.Get("exclusion_rule.#").(int)
	slice := make([]*waf.WafProfileExclusionRule, count)

	for i := 0; i < count; i++ {
		exclusionRules, err := expandWafProfileExclusionRules_(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = exclusionRules
	}

	return slice, nil
}

func expandWafProfileExclusionRules_(d *schema.ResourceData, indexes ...interface{}) (*waf.WafProfileExclusionRule, error) {
	val := new(waf.WafProfileExclusionRule)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition", indexes...)); ok {
		condition, err := expandWafProfileExclusionRulesCondition_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.exclude_rules", indexes...)); ok {
		excludeRules, err := expandWafProfileExclusionRulesExcludeRules_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetExcludeRules(excludeRules)
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.log_excluded", indexes...)); ok {
		val.SetLogExcluded(v.(bool))
	}

	return val, nil
}

func expandWafProfileExclusionRulesCondition_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority", indexes...)); ok {
		authority, err := expandWafProfileExclusionRulesConditionAuthority_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandWafProfileExclusionRulesConditionHttpMethod_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandWafProfileExclusionRulesConditionRequestUri_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers", indexes...)); ok {
		headers, err := expandWafProfileExclusionRulesConditionHeadersSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandWafProfileExclusionRulesConditionSourceIp_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionAuthority_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandWafProfileExclusionRulesConditionAuthorityAuthoritiesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionAuthorityAuthoritiesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandWafProfileExclusionRulesConditionAuthorityAuthorities_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandWafProfileExclusionRulesConditionAuthorityAuthorities_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionHttpMethod_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandWafProfileExclusionRulesConditionHttpMethodHttpMethodsSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionHttpMethodHttpMethodsSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandWafProfileExclusionRulesConditionHttpMethodHttpMethods_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandWafProfileExclusionRulesConditionHttpMethodHttpMethods_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionRequestUri_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandWafProfileExclusionRulesConditionRequestUriPath_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandWafProfileExclusionRulesConditionRequestUriQueriesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionRequestUriPath_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionRequestUriQueriesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandWafProfileExclusionRulesConditionRequestUriQueries_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandWafProfileExclusionRulesConditionRequestUriQueries_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandWafProfileExclusionRulesConditionRequestUriQueriesValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionRequestUriQueriesValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionHeadersSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandWafProfileExclusionRulesConditionHeaders_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandWafProfileExclusionRulesConditionHeaders_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandWafProfileExclusionRulesConditionHeadersValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionHeadersValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionSourceIp_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandWafProfileExclusionRulesConditionSourceIpIpRangesMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandWafProfileExclusionRulesConditionSourceIpIpRangesNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandWafProfileExclusionRulesConditionSourceIpGeoIpMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandWafProfileExclusionRulesConditionSourceIpGeoIpNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionSourceIpIpRangesMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionSourceIpIpRangesNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionSourceIpGeoIpMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandWafProfileExclusionRulesConditionSourceIpGeoIpNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandWafProfileExclusionRulesExcludeRules_(d *schema.ResourceData, indexes ...interface{}) (*waf.WafProfileExclusionRule_ExcludeRules, error) {
	val := new(waf.WafProfileExclusionRule_ExcludeRules)

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.exclude_rules.0.exclude_all", indexes...)); ok {
		val.SetExcludeAll(v.(bool))
	}

	if v, ok := d.GetOk(fmt.Sprintf("exclusion_rule.%d.exclude_rules.0.rule_ids", indexes...)); ok {
		ruleIds, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetRuleIds(ruleIds)
	}

	return val, nil
}

func expandWafProfileAnalyzeRequestBody_(d *schema.ResourceData) (*waf.WafProfile_AnalyzeRequestBody, error) {
	val := new(waf.WafProfile_AnalyzeRequestBody)

	if v, ok := d.GetOk("analyze_request_body.0.is_enabled"); ok {
		val.SetIsEnabled(v.(bool))
	}

	if v, ok := d.GetOk("analyze_request_body.0.size_limit"); ok {
		val.SetSizeLimit(int64(v.(int)))
	}

	if v, ok := d.GetOk("analyze_request_body.0.size_limit_action"); ok {
		action, err := parseWafWafProfileXAnalyzeRequestBodyXAction(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetSizeLimitAction(action)
	}

	empty := new(waf.WafProfile_AnalyzeRequestBody)
	if proto.Equal(val, empty) {
		return nil, nil
	}

	return val, nil
}

func expandWafProfileCoreRuleSet_(d *schema.ResourceData) (*waf.WafProfile_CoreRuleSet, error) {
	val := new(waf.WafProfile_CoreRuleSet)

	if v, ok := d.GetOk("core_rule_set.0.inbound_anomaly_score"); ok {
		val.SetInboundAnomalyScore(int64(v.(int)))
	}

	if v, ok := d.GetOk("core_rule_set.0.paranoia_level"); ok {
		val.SetParanoiaLevel(int64(v.(int)))
	}

	if _, ok := d.GetOk("core_rule_set.0.rule_set"); ok {
		ruleSet, err := expandWafProfileCoreRuleSetRuleSet_(d)
		if err != nil {
			return nil, err
		}

		val.SetRuleSet(ruleSet)
	}

	empty := new(waf.WafProfile_CoreRuleSet)
	if proto.Equal(val, empty) {
		return nil, nil
	}

	return val, nil
}

func expandWafProfileCoreRuleSetRuleSet_(d *schema.ResourceData) (*waf.RuleSet, error) {
	val := new(waf.RuleSet)

	if v, ok := d.GetOk("core_rule_set.0.rule_set.0.name"); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk("core_rule_set.0.rule_set.0.version"); ok {
		val.SetVersion(v.(string))
	}

	return val, nil
}

func parseSmartwebsecuritySecurityProfileXDefaultAction(str string) (smartwebsecurity.SecurityProfile_DefaultAction, error) {
	val, ok := smartwebsecurity.SecurityProfile_DefaultAction_value[str]
	if !ok {
		return smartwebsecurity.SecurityProfile_DefaultAction(0), invalidKeyError("default_action", smartwebsecurity.SecurityProfile_DefaultAction_value, str)
	}
	return smartwebsecurity.SecurityProfile_DefaultAction(val), nil
}

func parseSmartwebsecuritySecurityRuleXRuleConditionXAction(str string) (smartwebsecurity.SecurityRule_RuleCondition_Action, error) {
	val, ok := smartwebsecurity.SecurityRule_RuleCondition_Action_value[str]
	if !ok {
		return smartwebsecurity.SecurityRule_RuleCondition_Action(0), invalidKeyError("action", smartwebsecurity.SecurityRule_RuleCondition_Action_value, str)
	}
	return smartwebsecurity.SecurityRule_RuleCondition_Action(val), nil
}

func parseSmartwebsecuritySecurityRuleXSmartProtectionXMode(str string) (smartwebsecurity.SecurityRule_SmartProtection_Mode, error) {
	val, ok := smartwebsecurity.SecurityRule_SmartProtection_Mode_value[str]
	if !ok {
		return smartwebsecurity.SecurityRule_SmartProtection_Mode(0), invalidKeyError("mode", smartwebsecurity.SecurityRule_SmartProtection_Mode_value, str)
	}
	return smartwebsecurity.SecurityRule_SmartProtection_Mode(val), nil
}

func parseSmartwebsecuritySecurityRuleXWafXMode(str string) (smartwebsecurity.SecurityRule_Waf_Mode, error) {
	val, ok := smartwebsecurity.SecurityRule_Waf_Mode_value[str]
	if !ok {
		return smartwebsecurity.SecurityRule_Waf_Mode(0), invalidKeyError("mode", smartwebsecurity.SecurityRule_Waf_Mode_value, str)
	}
	return smartwebsecurity.SecurityRule_Waf_Mode(val), nil
}

func expandSecurityProfileSecurityRulesSlice(d *schema.ResourceData) ([]*smartwebsecurity.SecurityRule, error) {
	count := d.Get("security_rule.#").(int)
	slice := make([]*smartwebsecurity.SecurityRule, count)

	for i := 0; i < count; i++ {
		securityRules, err := expandSecurityProfileSecurityRules(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = securityRules
	}

	return slice, nil
}

func expandSecurityProfileSecurityRules(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.SecurityRule, error) {
	val := new(smartwebsecurity.SecurityRule)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.priority", indexes...)); ok {
		val.SetPriority(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.dry_run", indexes...)); ok {
		val.SetDryRun(v.(bool))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition", indexes...)); ok {
		ruleCondition, err := expandSecurityProfileSecurityRulesRuleCondition(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRuleCondition(ruleCondition)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection", indexes...)); ok {
		smartProtection, err := expandSecurityProfileSecurityRulesSmartProtection(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSmartProtection(smartProtection)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf", indexes...)); ok {
		securityRuleWaf, err := expandSecurityProfileSecurityRulesWaf(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetWaf(securityRuleWaf)
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleCondition(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.SecurityRule_RuleCondition, error) {
	val := new(smartwebsecurity.SecurityRule_RuleCondition)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.action", indexes...)); ok {
		action, err := parseSmartwebsecuritySecurityRuleXRuleConditionXAction(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetAction(action)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition", indexes...)); ok {
		condition, err := expandSecurityProfileSecurityRulesRuleConditionCondition(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionCondition(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority", indexes...)); ok {
		authority, err := expandSecurityProfileSecurityRulesRuleConditionConditionAuthority(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethod(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandSecurityProfileSecurityRulesRuleConditionConditionRequestUri(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers", indexes...)); ok {
		headers, err := expandSecurityProfileSecurityRulesRuleConditionConditionHeadersSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandSecurityProfileSecurityRulesRuleConditionConditionSourceIp(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionAuthority(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandSecurityProfileSecurityRulesRuleConditionConditionAuthorityAuthoritiesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionAuthorityAuthoritiesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandSecurityProfileSecurityRulesRuleConditionConditionAuthorityAuthorities(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionAuthorityAuthorities(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethod(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethodHttpMethodsSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethodHttpMethodsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethodHttpMethods(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethodHttpMethods(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionRequestUri(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriPath(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueriesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriPath(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueriesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueries(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueries(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueriesValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueriesValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHeadersSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandSecurityProfileSecurityRulesRuleConditionConditionHeaders(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHeaders(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesRuleConditionConditionHeadersValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHeadersValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionSourceIp(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpIpRangesMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpIpRangesNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpGeoIpMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpGeoIpNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpIpRangesMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpIpRangesNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpGeoIpMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpGeoIpNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtection(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.SecurityRule_SmartProtection, error) {
	val := new(smartwebsecurity.SecurityRule_SmartProtection)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.mode", indexes...)); ok {
		mode, err := parseSmartwebsecuritySecurityRuleXSmartProtectionXMode(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetMode(mode)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition", indexes...)); ok {
		condition, err := expandSecurityProfileSecurityRulesSmartProtectionCondition(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionCondition(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority", indexes...)); ok {
		authority, err := expandSecurityProfileSecurityRulesSmartProtectionConditionAuthority(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethod(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUri(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers", indexes...)); ok {
		headers, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHeadersSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIp(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionAuthority(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandSecurityProfileSecurityRulesSmartProtectionConditionAuthorityAuthoritiesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionAuthorityAuthoritiesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandSecurityProfileSecurityRulesSmartProtectionConditionAuthorityAuthorities(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionAuthorityAuthorities(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethod(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethodHttpMethodsSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethodHttpMethodsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethodHttpMethods(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethodHttpMethods(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUri(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriPath(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueriesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriPath(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueriesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueries(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueries(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueriesValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueriesValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHeadersSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHeaders(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHeaders(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHeadersValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHeadersValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIp(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpIpRangesMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpIpRangesNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpGeoIpMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpGeoIpNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpIpRangesMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpIpRangesNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpGeoIpMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpGeoIpNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWaf(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.SecurityRule_Waf, error) {
	val := new(smartwebsecurity.SecurityRule_Waf)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.mode", indexes...)); ok {
		mode, err := parseSmartwebsecuritySecurityRuleXWafXMode(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetMode(mode)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition", indexes...)); ok {
		condition, err := expandSecurityProfileSecurityRulesWafCondition(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.waf_profile_id", indexes...)); ok {
		val.SetWafProfileId(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafCondition(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority", indexes...)); ok {
		authority, err := expandSecurityProfileSecurityRulesWafConditionAuthority(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandSecurityProfileSecurityRulesWafConditionHttpMethod(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandSecurityProfileSecurityRulesWafConditionRequestUri(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers", indexes...)); ok {
		headers, err := expandSecurityProfileSecurityRulesWafConditionHeadersSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandSecurityProfileSecurityRulesWafConditionSourceIp(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionAuthority(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandSecurityProfileSecurityRulesWafConditionAuthorityAuthoritiesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionAuthorityAuthoritiesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandSecurityProfileSecurityRulesWafConditionAuthorityAuthorities(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesWafConditionAuthorityAuthorities(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionHttpMethod(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandSecurityProfileSecurityRulesWafConditionHttpMethodHttpMethodsSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionHttpMethodHttpMethodsSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandSecurityProfileSecurityRulesWafConditionHttpMethodHttpMethods(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesWafConditionHttpMethodHttpMethods(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionRequestUri(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandSecurityProfileSecurityRulesWafConditionRequestUriPath(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandSecurityProfileSecurityRulesWafConditionRequestUriQueriesSlice(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionRequestUriPath(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionRequestUriQueriesSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandSecurityProfileSecurityRulesWafConditionRequestUriQueries(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesWafConditionRequestUriQueries(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesWafConditionRequestUriQueriesValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionRequestUriQueriesValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionHeadersSlice(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandSecurityProfileSecurityRulesWafConditionHeaders(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesWafConditionHeaders(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesWafConditionHeadersValue(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionHeadersValue(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionSourceIp(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandSecurityProfileSecurityRulesWafConditionSourceIpIpRangesMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandSecurityProfileSecurityRulesWafConditionSourceIpIpRangesNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandSecurityProfileSecurityRulesWafConditionSourceIpGeoIpMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandSecurityProfileSecurityRulesWafConditionSourceIpGeoIpNotMatch(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionSourceIpIpRangesMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionSourceIpIpRangesNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionSourceIpGeoIpMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionSourceIpGeoIpNotMatch(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func flattenSmartwebsecuritySecurityRuleSlice(vs []*smartwebsecurity.SecurityRule) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))

	for _, v := range vs {
		securityRule, err := flatten_yandex_cloud_smartwebsecurity_v1_SecurityRule(v)
		if err != nil {
			// B // isElem: true, ret: 2
			return nil, err
		}

		if len(securityRule) != 0 {
			s = append(s, securityRule[0])
		}
	}

	return s, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_SecurityRule(v *smartwebsecurity.SecurityRule) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["description"] = v.Description
	m["dry_run"] = v.DryRun
	m["name"] = v.Name
	m["priority"] = v.Priority
	ruleCondition, err := flatten_yandex_cloud_smartwebsecurity_v1_SecurityRule_RuleCondition(v.GetRuleCondition())
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["rule_condition"] = ruleCondition
	smartProtection, err := flatten_yandex_cloud_smartwebsecurity_v1_SecurityRule_SmartProtection(v.GetSmartProtection())
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["smart_protection"] = smartProtection
	waf_, err := flatten_yandex_cloud_smartwebsecurity_v1_SecurityRule_Waf(v.GetWaf())
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["waf"] = waf_

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_SecurityRule_RuleCondition(v *smartwebsecurity.SecurityRule_RuleCondition) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["action"] = v.Action.String()
	condition, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition(v.Condition)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["condition"] = condition

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_SecurityRule_SmartProtection(v *smartwebsecurity.SecurityRule_SmartProtection) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	condition, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition(v.Condition)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["condition"] = condition
	m["mode"] = v.Mode.String()

	return []map[string]interface{}{m}, nil
}

func flatten_yandex_cloud_smartwebsecurity_v1_SecurityRule_Waf(v *smartwebsecurity.SecurityRule_Waf) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	condition, err := flatten_yandex_cloud_smartwebsecurity_v1_Condition(v.Condition)
	if err != nil {
		// B // isElem: false, ret: 2
		return nil, err
	}
	m["condition"] = condition
	m["mode"] = v.Mode.String()
	m["waf_profile_id"] = v.WafProfileId

	return []map[string]interface{}{m}, nil
}

func expandSecurityProfileSecurityRulesSlice_(d *schema.ResourceData) ([]*smartwebsecurity.SecurityRule, error) {
	count := d.Get("security_rule.#").(int)
	slice := make([]*smartwebsecurity.SecurityRule, count)

	for i := 0; i < count; i++ {
		securityRules, err := expandSecurityProfileSecurityRules_(d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = securityRules
	}

	return slice, nil
}

func expandSecurityProfileSecurityRules_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.SecurityRule, error) {
	val := new(smartwebsecurity.SecurityRule)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.priority", indexes...)); ok {
		val.SetPriority(int64(v.(int)))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.dry_run", indexes...)); ok {
		val.SetDryRun(v.(bool))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition", indexes...)); ok {
		ruleCondition, err := expandSecurityProfileSecurityRulesRuleCondition_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRuleCondition(ruleCondition)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection", indexes...)); ok {
		smartProtection, err := expandSecurityProfileSecurityRulesSmartProtection_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSmartProtection(smartProtection)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf", indexes...)); ok {
		securityRuleWaf, err := expandSecurityProfileSecurityRulesWaf_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetWaf(securityRuleWaf)
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.description", indexes...)); ok {
		val.SetDescription(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleCondition_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.SecurityRule_RuleCondition, error) {
	val := new(smartwebsecurity.SecurityRule_RuleCondition)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.action", indexes...)); ok {
		action, err := parseSmartwebsecuritySecurityRuleXRuleConditionXAction(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetAction(action)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition", indexes...)); ok {
		condition, err := expandSecurityProfileSecurityRulesRuleConditionCondition_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionCondition_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority", indexes...)); ok {
		authority, err := expandSecurityProfileSecurityRulesRuleConditionConditionAuthority_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethod_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandSecurityProfileSecurityRulesRuleConditionConditionRequestUri_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers", indexes...)); ok {
		headers, err := expandSecurityProfileSecurityRulesRuleConditionConditionHeadersSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandSecurityProfileSecurityRulesRuleConditionConditionSourceIp_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionAuthority_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandSecurityProfileSecurityRulesRuleConditionConditionAuthorityAuthoritiesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionAuthorityAuthoritiesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandSecurityProfileSecurityRulesRuleConditionConditionAuthorityAuthorities_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionAuthorityAuthorities_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethod_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethodHttpMethodsSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethodHttpMethodsSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethodHttpMethods_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHttpMethodHttpMethods_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionRequestUri_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriPath_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueriesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriPath_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueriesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueries_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueries_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueriesValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionRequestUriQueriesValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHeadersSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandSecurityProfileSecurityRulesRuleConditionConditionHeaders_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHeaders_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesRuleConditionConditionHeadersValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionHeadersValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionSourceIp_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpIpRangesMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpIpRangesNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpGeoIpMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpGeoIpNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpIpRangesMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpIpRangesNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpGeoIpMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesRuleConditionConditionSourceIpGeoIpNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.rule_condition.0.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtection_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.SecurityRule_SmartProtection, error) {
	val := new(smartwebsecurity.SecurityRule_SmartProtection)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.mode", indexes...)); ok {
		mode, err := parseSmartwebsecuritySecurityRuleXSmartProtectionXMode(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetMode(mode)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition", indexes...)); ok {
		condition, err := expandSecurityProfileSecurityRulesSmartProtectionCondition_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionCondition_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority", indexes...)); ok {
		authority, err := expandSecurityProfileSecurityRulesSmartProtectionConditionAuthority_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethod_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUri_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers", indexes...)); ok {
		headers, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHeadersSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIp_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionAuthority_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandSecurityProfileSecurityRulesSmartProtectionConditionAuthorityAuthoritiesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionAuthorityAuthoritiesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandSecurityProfileSecurityRulesSmartProtectionConditionAuthorityAuthorities_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionAuthorityAuthorities_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethod_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethodHttpMethodsSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethodHttpMethodsSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethodHttpMethods_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHttpMethodHttpMethods_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUri_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriPath_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueriesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriPath_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueriesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueries_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueries_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueriesValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionRequestUriQueriesValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHeadersSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHeaders_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHeaders_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesSmartProtectionConditionHeadersValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionHeadersValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIp_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpIpRangesMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpIpRangesNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpGeoIpMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpGeoIpNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpIpRangesMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpIpRangesNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpGeoIpMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesSmartProtectionConditionSourceIpGeoIpNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.smart_protection.0.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWaf_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.SecurityRule_Waf, error) {
	val := new(smartwebsecurity.SecurityRule_Waf)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.mode", indexes...)); ok {
		mode, err := parseSmartwebsecuritySecurityRuleXWafXMode(v.(string))
		if err != nil {
			return nil, err
		}

		val.SetMode(mode)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition", indexes...)); ok {
		condition, err := expandSecurityProfileSecurityRulesWafCondition_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetCondition(condition)
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.waf_profile_id", indexes...)); ok {
		val.SetWafProfileId(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafCondition_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition, error) {
	val := new(smartwebsecurity.Condition)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority", indexes...)); ok {
		authority, err := expandSecurityProfileSecurityRulesWafConditionAuthority_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthority(authority)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method", indexes...)); ok {
		httpMethod, err := expandSecurityProfileSecurityRulesWafConditionHttpMethod_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethod(httpMethod)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri", indexes...)); ok {
		requestUri, err := expandSecurityProfileSecurityRulesWafConditionRequestUri_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetRequestUri(requestUri)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers", indexes...)); ok {
		headers, err := expandSecurityProfileSecurityRulesWafConditionHeadersSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHeaders(headers)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip", indexes...)); ok {
		sourceIp, err := expandSecurityProfileSecurityRulesWafConditionSourceIp_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetSourceIp(sourceIp)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionAuthority_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_AuthorityMatcher, error) {
	val := new(smartwebsecurity.Condition_AuthorityMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities", indexes...)); ok {
		authorities, err := expandSecurityProfileSecurityRulesWafConditionAuthorityAuthoritiesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetAuthorities(authorities)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionAuthorityAuthoritiesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		authorities, err := expandSecurityProfileSecurityRulesWafConditionAuthorityAuthorities_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = authorities
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesWafConditionAuthorityAuthorities_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.authority.0.authorities.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionHttpMethod_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HttpMethodMatcher, error) {
	val := new(smartwebsecurity.Condition_HttpMethodMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods", indexes...)); ok {
		httpMethods, err := expandSecurityProfileSecurityRulesWafConditionHttpMethodHttpMethodsSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetHttpMethods(httpMethods)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionHttpMethodHttpMethodsSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_StringMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_StringMatcher, count)

	for i := 0; i < count; i++ {
		httpMethods, err := expandSecurityProfileSecurityRulesWafConditionHttpMethodHttpMethods_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = httpMethods
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesWafConditionHttpMethodHttpMethods_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.http_method.0.http_methods.%d.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionRequestUri_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_RequestUriMatcher, error) {
	val := new(smartwebsecurity.Condition_RequestUriMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path", indexes...)); ok {
		path, err := expandSecurityProfileSecurityRulesWafConditionRequestUriPath_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetPath(path)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries", indexes...)); ok {
		queries, err := expandSecurityProfileSecurityRulesWafConditionRequestUriQueriesSlice_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetQueries(queries)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionRequestUriPath_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.path.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionRequestUriQueriesSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_QueryMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_QueryMatcher, count)

	for i := 0; i < count; i++ {
		queries, err := expandSecurityProfileSecurityRulesWafConditionRequestUriQueries_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = queries
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesWafConditionRequestUriQueries_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_QueryMatcher, error) {
	val := new(smartwebsecurity.Condition_QueryMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesWafConditionRequestUriQueriesValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionRequestUriQueriesValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.request_uri.0.queries.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionHeadersSlice_(d *schema.ResourceData, indexes ...interface{}) ([]*smartwebsecurity.Condition_HeaderMatcher, error) {
	count := d.Get(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.#", indexes...)).(int)
	slice := make([]*smartwebsecurity.Condition_HeaderMatcher, count)

	for i := 0; i < count; i++ {
		headers, err := expandSecurityProfileSecurityRulesWafConditionHeaders_(d, append(indexes, i)...)
		if err != nil {
			return nil, err
		}

		slice[i] = headers
	}

	return slice, nil
}

func expandSecurityProfileSecurityRulesWafConditionHeaders_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_HeaderMatcher, error) {
	val := new(smartwebsecurity.Condition_HeaderMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.name", indexes...)); ok {
		val.SetName(v.(string))
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value", indexes...)); ok {
		value, err := expandSecurityProfileSecurityRulesWafConditionHeadersValue_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetValue(value)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionHeadersValue_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_StringMatcher, error) {
	val := new(smartwebsecurity.Condition_StringMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.exact_match", indexes...)); ok {
		val.SetExactMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.exact_not_match", indexes...)); ok {
		val.SetExactNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.prefix_match", indexes...)); ok {
		val.SetPrefixMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.prefix_not_match", indexes...)); ok {
		val.SetPrefixNotMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.pire_regex_match", indexes...)); ok {
		val.SetPireRegexMatch(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.headers.%d.value.0.pire_regex_not_match", indexes...)); ok {
		val.SetPireRegexNotMatch(v.(string))
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionSourceIp_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpMatcher, error) {
	val := new(smartwebsecurity.Condition_IpMatcher)

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.ip_ranges_match", indexes...)); ok {
		ipRangesMatch, err := expandSecurityProfileSecurityRulesWafConditionSourceIpIpRangesMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesMatch(ipRangesMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.ip_ranges_not_match", indexes...)); ok {
		ipRangesNotMatch, err := expandSecurityProfileSecurityRulesWafConditionSourceIpIpRangesNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetIpRangesNotMatch(ipRangesNotMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.geo_ip_match", indexes...)); ok {
		geoIpMatch, err := expandSecurityProfileSecurityRulesWafConditionSourceIpGeoIpMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpMatch(geoIpMatch)
	}

	if _, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.geo_ip_not_match", indexes...)); ok {
		geoIpNotMatch, err := expandSecurityProfileSecurityRulesWafConditionSourceIpGeoIpNotMatch_(d, indexes...)
		if err != nil {
			return nil, err
		}

		val.SetGeoIpNotMatch(geoIpNotMatch)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionSourceIpIpRangesMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.ip_ranges_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionSourceIpIpRangesNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_IpRangesMatcher, error) {
	val := new(smartwebsecurity.Condition_IpRangesMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.ip_ranges_not_match.0.ip_ranges", indexes...)); ok {
		ipRanges, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetIpRanges(ipRanges)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionSourceIpGeoIpMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.geo_ip_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}

func expandSecurityProfileSecurityRulesWafConditionSourceIpGeoIpNotMatch_(d *schema.ResourceData, indexes ...interface{}) (*smartwebsecurity.Condition_GeoIpMatcher, error) {
	val := new(smartwebsecurity.Condition_GeoIpMatcher)

	if v, ok := d.GetOk(fmt.Sprintf("security_rule.%d.waf.0.condition.0.source_ip.0.geo_ip_not_match.0.locations", indexes...)); ok {
		locations, err := expandStrings(v.([]interface{}))
		if err != nil {
			return nil, err
		}

		val.SetLocations(locations)
	}

	return val, nil
}
