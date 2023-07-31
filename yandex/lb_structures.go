package yandex

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"

	"github.com/golang/protobuf/ptypes/duration"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
)

func resourceLBTargetGroupTargetHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["subnet_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["address"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	// TODO: SA1019: hashcode.String is deprecated: This will be removed in v2 without replacement. If you need its functionality, you can copy it, import crc32 directly, or reference the v1 package. (staticcheck)
	return hashcode.String(buf.String())
}

func resourceLBNetworkLoadBalancerListenerHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	for _, k := range []string{"name", "port"} {
		fmt.Fprintf(&buf, "\"%v\":", m[k])
	}
	targetPort := m["target_port"].(int)
	if targetPort == 0 {
		targetPort = m["port"].(int)
	}
	buf.WriteString(fmt.Sprintf("\"%v\":", targetPort))
	protocol := m["protocol"].(string)
	if len(protocol) == 0 {
		protocol = "tcp"
	}
	buf.WriteString(fmt.Sprintf("\"%v\":", protocol))
	return hashcode.String(buf.String())
}

func resourceLBNetworkLoadBalancerExternalAddressHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	for _, k := range []string{"address", "ip_version"} {
		fmt.Fprintf(&buf, "\"%v\":", m[k])
	}
	return hashcode.String(buf.String())
}

func resourceLBNetworkLoadBalancerInternalAddressHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	for _, k := range []string{"subnet_id", "address", "ip_version"} {
		fmt.Fprintf(&buf, "\"%v\":", m[k])
	}
	return hashcode.String(buf.String())
}

func resourceLBNetworkLoadBalancerAttachedTargetGroupHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["target_group_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	switch hcs := m["healthcheck"].(type) {
	case []interface{}:
		for _, hcc := range hcs {
			hc := hcc.(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%d-", healthCheckSettingsHash(hc)))
		}
	case []map[string]interface{}:
		for _, hc := range hcs {
			buf.WriteString(fmt.Sprintf("%d-", healthCheckSettingsHash(hc)))
		}
	default:
	}

	return hashcode.String(buf.String())
}

func healthCheckSettingsHash(hc map[string]interface{}) int {
	var buf bytes.Buffer

	for _, k := range []string{"name", "interval", "timeout", "unhealthy_threshold", "healthy_threshold"} {
		if v, ok := hc[k]; ok {
			buf.WriteString(fmt.Sprintf("\"%v\":", v))
		}
	}

	if httpOptions, ok := getFirstElement(hc, "http_options"); ok {
		buf.WriteString("http:")
		if v, ok := httpOptions["port"]; ok {
			buf.WriteString(fmt.Sprintf("%v-", v))
		}
		if v, ok := httpOptions["path"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}
	}

	if tcpOptions, ok := getFirstElement(hc, "tcp_options"); ok {
		buf.WriteString("tcp:")
		if v, ok := tcpOptions["port"]; ok {
			buf.WriteString(fmt.Sprintf("%v-", v))
		}
	}

	return hashcode.String(buf.String())
}

func expandLBListenerSpecs(d *schema.ResourceData) ([]*loadbalancer.ListenerSpec, error) {
	var result []*loadbalancer.ListenerSpec
	listenersSet := d.Get("listener").(*schema.Set)

	for _, v := range listenersSet.List() {
		config := v.(map[string]interface{})

		ls, err := expandLBListenerSpec(config)
		if err != nil {
			return nil, err
		}

		result = append(result, ls)
	}

	return result, nil
}

func expandLBListenerSpec(config map[string]interface{}) (*loadbalancer.ListenerSpec, error) {
	ls := &loadbalancer.ListenerSpec{}

	if v, ok := config["name"]; ok {
		ls.Name = v.(string)
	}

	if v, ok := config["port"]; ok {
		ls.Port = int64(v.(int))
	}

	if v, ok := config["target_port"]; ok {
		ls.TargetPort = int64(v.(int))
	}

	if v, ok := config["protocol"]; ok {
		p, err := parseListenerProtocol(v.(string))
		if err != nil {
			return nil, err
		}
		ls.Protocol = p
	}

	if v, ok := config["external_address_spec"].(*schema.Set); ok && v.Len() > 0 {
		eas, err := expandLBExternalAddressSpec(v.List()[0].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		ls.Address = eas
	}

	if v, ok := config["internal_address_spec"].(*schema.Set); ok && v.Len() > 0 {
		if ls.Address != nil {
			return nil, fmt.Errorf("use one of 'external_address_spec' or 'internal_address_spec', not both")
		}
		ias, err := expandLBInternalAddressSpec(v.List()[0].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		ls.Address = ias
	}

	return ls, nil
}

func expandLBExternalAddressSpec(config map[string]interface{}) (*loadbalancer.ListenerSpec_ExternalAddressSpec, error) {
	as := &loadbalancer.ListenerSpec_ExternalAddressSpec{
		ExternalAddressSpec: &loadbalancer.ExternalAddressSpec{},
	}

	if v, ok := config["address"]; ok {
		as.ExternalAddressSpec.Address = v.(string)
	}

	if v, ok := config["ip_version"]; ok {
		v, err := parseIPVersion(v.(string))
		if err != nil {
			return nil, err
		}
		as.ExternalAddressSpec.IpVersion = v
	}

	return as, nil
}

func expandLBInternalAddressSpec(config map[string]interface{}) (*loadbalancer.ListenerSpec_InternalAddressSpec, error) {
	as := &loadbalancer.ListenerSpec_InternalAddressSpec{
		InternalAddressSpec: &loadbalancer.InternalAddressSpec{},
	}

	as.InternalAddressSpec.SubnetId = config["subnet_id"].(string)

	if v, ok := config["address"]; ok {
		as.InternalAddressSpec.Address = v.(string)
	}

	if v, ok := config["ip_version"]; ok {
		v, err := parseIPVersion(v.(string))
		if err != nil {
			return nil, err
		}
		as.InternalAddressSpec.IpVersion = v
	}

	return as, nil
}

func expandLBAttachedTargetGroups(d *schema.ResourceData) ([]*loadbalancer.AttachedTargetGroup, error) {
	var result []*loadbalancer.AttachedTargetGroup
	atgsSet := d.Get("attached_target_group").(*schema.Set)

	for _, v := range atgsSet.List() {
		config := v.(map[string]interface{})

		atg, err := expandLBAttachedTargetGroup(config)
		if err != nil {
			return nil, err
		}

		result = append(result, atg)
	}

	return result, nil
}

func expandLBAttachedTargetGroup(config map[string]interface{}) (*loadbalancer.AttachedTargetGroup, error) {
	atg := &loadbalancer.AttachedTargetGroup{}

	if v, ok := config["target_group_id"]; ok {
		atg.TargetGroupId = v.(string)
	}

	if v, ok := config["healthcheck"]; ok {
		hcConfigs := v.([]interface{})
		atg.HealthChecks = make([]*loadbalancer.HealthCheck, len(hcConfigs))

		for i := 0; i < len(hcConfigs); i++ {
			hcConfig := hcConfigs[i]
			hc, err := expandLBHealthcheck(hcConfig.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			atg.HealthChecks[i] = hc
		}
	}

	return atg, nil
}

func expandLBHealthcheck(config map[string]interface{}) (*loadbalancer.HealthCheck, error) {
	hc := &loadbalancer.HealthCheck{}

	if v, ok := config["name"]; ok {
		hc.Name = v.(string)
	}

	if v, ok := config["interval"]; ok {
		hc.Interval = &duration.Duration{Seconds: int64(v.(int))}
	}

	if v, ok := config["timeout"]; ok {
		hc.Timeout = &duration.Duration{Seconds: int64(v.(int))}
	}

	if v, ok := config["unhealthy_threshold"]; ok {
		hc.UnhealthyThreshold = int64(v.(int))
	}

	if v, ok := config["healthy_threshold"]; ok {
		hc.HealthyThreshold = int64(v.(int))
	}

	httpOptions, httpOptionsOk := getFirstElement(config, "http_options")
	tcpOptions, tcpOptionsOk := getFirstElement(config, "tcp_options")

	if httpOptionsOk && tcpOptionsOk {
		return nil, fmt.Errorf("Use one of 'http_options' or 'tcp_options', not both")
	}

	if httpOptionsOk {
		options, err := expandLBHealthcheckHTTPOptions(httpOptions)
		if err != nil {
			return nil, err
		}
		hc.Options = &loadbalancer.HealthCheck_HttpOptions_{
			HttpOptions: options,
		}
	}

	if tcpOptionsOk {
		options, err := expandLBHealthcheckTCPOptions(tcpOptions)
		if err != nil {
			return nil, err
		}
		hc.Options = &loadbalancer.HealthCheck_TcpOptions_{
			TcpOptions: options,
		}
	}

	return hc, nil
}

func expandLBHealthcheckHTTPOptions(config map[string]interface{}) (*loadbalancer.HealthCheck_HttpOptions, error) {
	options := &loadbalancer.HealthCheck_HttpOptions{}

	if v, ok := config["port"]; ok {
		options.Port = int64(v.(int))
	}

	if v, ok := config["path"]; ok {
		options.Path = v.(string)
	}

	return options, nil
}

func expandLBHealthcheckTCPOptions(config map[string]interface{}) (*loadbalancer.HealthCheck_TcpOptions, error) {
	options := &loadbalancer.HealthCheck_TcpOptions{}

	if v, ok := config["port"]; ok {
		options.Port = int64(v.(int))
	}

	return options, nil
}

func expandLBTargets(d *schema.ResourceData) ([]*loadbalancer.Target, error) {
	var targets []*loadbalancer.Target
	targetsSet := d.Get("target").(*schema.Set)

	for _, t := range targetsSet.List() {
		targetConfig := t.(map[string]interface{})

		target, err := expandLBTarget(targetConfig)
		if err != nil {
			return nil, err
		}

		targets = append(targets, target)
	}

	return targets, nil
}

func expandLBTarget(config map[string]interface{}) (*loadbalancer.Target, error) {
	target := &loadbalancer.Target{}

	if v, ok := config["subnet_id"]; ok {
		target.SubnetId = v.(string)
	}

	if v, ok := config["address"]; ok {
		target.Address = v.(string)
	}

	return target, nil
}

func flattenLBTargets(tg *loadbalancer.TargetGroup) (*schema.Set, error) {
	result := &schema.Set{F: resourceLBTargetGroupTargetHash}

	for _, t := range tg.Targets {
		flTarget := map[string]interface{}{
			"subnet_id": t.SubnetId,
			"address":   t.Address,
		}
		result.Add(flTarget)
	}

	return result, nil
}

func flattenLBListenerSpecs(nlb *loadbalancer.NetworkLoadBalancer) (*schema.Set, error) {
	result := &schema.Set{F: resourceLBNetworkLoadBalancerListenerHash}
	var (
		addressSpecKey     string
		flattenAddressSpec func(*loadbalancer.Listener) (*schema.Set, error)
	)
	switch nlb.Type {
	case loadbalancer.NetworkLoadBalancer_EXTERNAL:
		addressSpecKey = "external_address_spec"
		flattenAddressSpec = flattenLBExternalAddressSpec
	case loadbalancer.NetworkLoadBalancer_INTERNAL:
		addressSpecKey = "internal_address_spec"
		flattenAddressSpec = flattenLBInternalAddressSpec
	default:
		return nil, fmt.Errorf("Unknown network load balancer type: %v", nlb.Type)
	}
	for _, ls := range nlb.Listeners {
		as, err := flattenAddressSpec(ls)
		if err != nil {
			return nil, err
		}
		flListener := map[string]interface{}{
			"name":         ls.Name,
			"port":         int(ls.Port),
			"target_port":  int(ls.TargetPort),
			"protocol":     strings.ToLower(ls.Protocol.String()),
			addressSpecKey: as,
		}
		result.Add(flListener)
	}

	return result, nil
}

func flattenLBExternalAddressSpec(ls *loadbalancer.Listener) (*schema.Set, error) {
	result := map[string]interface{}{
		"address": ls.Address,
	}

	addr := net.ParseIP(ls.Address)
	isV4 := addr.To4() != nil
	if isV4 {
		result["ip_version"] = "ipv4"
	} else {
		result["ip_version"] = "ipv6"
	}

	return schema.NewSet(resourceLBNetworkLoadBalancerExternalAddressHash, []interface{}{result}), nil
}

func flattenLBInternalAddressSpec(ls *loadbalancer.Listener) (*schema.Set, error) {
	result := map[string]interface{}{
		"address": ls.Address,
	}

	addr := net.ParseIP(ls.Address)
	isV4 := addr.To4() != nil
	if isV4 {
		result["ip_version"] = "ipv4"
	} else {
		result["ip_version"] = "ipv6"
	}

	result["subnet_id"] = ls.SubnetId

	return schema.NewSet(resourceLBNetworkLoadBalancerExternalAddressHash, []interface{}{result}), nil
}

func flattenLBAttachedTargetGroups(nlb *loadbalancer.NetworkLoadBalancer) (*schema.Set, error) {
	result := &schema.Set{F: resourceLBNetworkLoadBalancerAttachedTargetGroupHash}

	for _, atg := range nlb.AttachedTargetGroups {
		hcs, err := flattenLBHealthchecks(atg)
		if err != nil {
			return nil, err
		}

		flATG := map[string]interface{}{
			"target_group_id": atg.TargetGroupId,
			"healthcheck":     hcs,
		}
		result.Add(flATG)
	}

	return result, nil
}

func flattenLBHealthchecks(atg *loadbalancer.AttachedTargetGroup) ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}

	for _, hc := range atg.HealthChecks {
		flHC := map[string]interface{}{
			"name":                hc.Name,
			"interval":            hc.Interval.Seconds,
			"timeout":             hc.Timeout.Seconds,
			"unhealthy_threshold": hc.UnhealthyThreshold,
			"healthy_threshold":   hc.HealthyThreshold,
		}
		switch hc.Options.(type) {
		case *loadbalancer.HealthCheck_HttpOptions_:
			flHC["http_options"] = []map[string]interface{}{
				{
					"port": hc.GetHttpOptions().Port,
					"path": hc.GetHttpOptions().Path,
				},
			}
		case *loadbalancer.HealthCheck_TcpOptions_:
			flHC["tcp_options"] = []map[string]interface{}{
				{
					"port": hc.GetTcpOptions().Port,
				},
			}
		default:
			return nil, fmt.Errorf("Unknown healthcheck options type: %T", hc.Options)
		}
		result = append(result, flHC)
	}

	return result, nil
}

func parseListenerProtocol(s string) (loadbalancer.Listener_Protocol, error) {
	switch strings.ToLower(s) {
	case "tcp":
		return loadbalancer.Listener_TCP, nil
	case "udp":
		return loadbalancer.Listener_UDP, nil
	case "":
		return loadbalancer.Listener_TCP, nil
	default:
		return loadbalancer.Listener_PROTOCOL_UNSPECIFIED,
			fmt.Errorf("value for 'protocol' must be 'tcp' or 'udp', not '%s'", s)
	}
}

func parseNetworkLoadBalancerType(s string) (loadbalancer.NetworkLoadBalancer_Type, error) {
	switch s {
	case "external":
		return loadbalancer.NetworkLoadBalancer_EXTERNAL, nil
	case "internal":
		return loadbalancer.NetworkLoadBalancer_INTERNAL, nil
	case "":
		return loadbalancer.NetworkLoadBalancer_EXTERNAL, nil
	default:
		return loadbalancer.NetworkLoadBalancer_TYPE_UNSPECIFIED,
			fmt.Errorf("value for 'type' must be 'external', not '%s'", s)
	}
}

func parseIPVersion(s string) (loadbalancer.IpVersion, error) {
	switch strings.ToLower(s) {
	case "ipv4":
		return loadbalancer.IpVersion_IPV4, nil
	case "ipv6":
		return loadbalancer.IpVersion_IPV6, nil
	case "":
		return loadbalancer.IpVersion_IPV4, nil
	default:
		return loadbalancer.IpVersion_IP_VERSION_UNSPECIFIED,
			fmt.Errorf("value for 'external ip version' must be 'ipv4' or 'ipv6', not '%s'", s)
	}
}

func getFirstElement(config map[string]interface{}, name string) (map[string]interface{}, bool) {
	if v, ok := config[name]; ok {
		switch v := v.(type) {
		case map[string]interface{}:
			return v, true
		case []interface{}:
			if len(v) > 0 {
				return v[0].(map[string]interface{}), true
			}
		case []map[string]interface{}:
			if len(v) > 0 {
				return v[0], true
			}
		default:
		}

	}
	return nil, false
}
