package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/smartcaptcha/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1"
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
	headers, err := flattenSmartwebsecuritySecurityRuleruleConditionconditionheadersSlice(v.Headers)
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

	authorities, err := flattenSmartwebsecuritySecurityRuleruleConditionconditionauthorityauthoritiesSlice(v.Authorities)
	if err != nil { // isElem: false, ret: 2
		return nil, err
	}
	m["authorities"] = authorities

	return []map[string]interface{}{m}, nil
}

func flattenSmartwebsecuritySecurityRuleruleConditionconditionauthorityauthoritiesSlice(vs []*smartwebsecurity.Condition_StringMatcher) ([]interface{}, error) {
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

func flattenSmartwebsecuritySecurityRuleruleConditionconditionheadersSlice(vs []*smartwebsecurity.Condition_HeaderMatcher) ([]interface{}, error) {
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

	httpMethods, err := flattenSmartwebsecuritySecurityRuleruleConditionconditionhttpMethodhttpMethodsSlice(v.HttpMethods)
	if err != nil { // isElem: false, ret: 2
		return nil, err
	}
	m["http_methods"] = httpMethods

	return []map[string]interface{}{m}, nil
}

func flattenSmartwebsecuritySecurityRuleruleConditionconditionhttpMethodhttpMethodsSlice(vs []*smartwebsecurity.Condition_StringMatcher) ([]interface{}, error) {
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
	queries, err := flattenSmartwebsecuritySecurityRuleruleConditionconditionrequestUriqueriesSlice(v.Queries)
	if err != nil { // isElem: false, ret: 2
		return nil, err
	}
	m["queries"] = queries

	return []map[string]interface{}{m}, nil
}

func flattenSmartwebsecuritySecurityRuleruleConditionconditionrequestUriqueriesSlice(vs []*smartwebsecurity.Condition_QueryMatcher) ([]interface{}, error) {
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
