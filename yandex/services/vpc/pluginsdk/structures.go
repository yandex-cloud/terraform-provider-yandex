package pluginsdk

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func flattenStaticRoutes(routeTable *vpc.RouteTable) *schema.Set {
	staticRoutes := schema.NewSet(ResourceYandexVPCRouteTableHash, nil)

	for _, r := range routeTable.StaticRoutes {
		m := make(map[string]interface{})

		switch d := r.Destination.(type) {
		case *vpc.StaticRoute_DestinationPrefix:
			m["destination_prefix"] = d.DestinationPrefix
		}

		switch h := r.NextHop.(type) {
		case *vpc.StaticRoute_NextHopAddress:
			m["next_hop_address"] = h.NextHopAddress
		case *vpc.StaticRoute_GatewayId:
			m["gateway_id"] = h.GatewayId
		}

		staticRoutes.Add(m)
	}
	return staticRoutes
}

func ExpandStaticRoutes(v interface{}) ([]*vpc.StaticRoute, error) {
	staticRoutes := []*vpc.StaticRoute{}

	if v == nil {
		return staticRoutes, nil
	}

	routeList := v.(*schema.Set).List()
	for _, v := range routeList {
		sr, err := routeDescriptionToStaticRoute(v)
		if err != nil {
			return nil, fmt.Errorf("fail convert static route: %s", err)
		}
		staticRoutes = append(staticRoutes, sr)
	}
	return staticRoutes, nil
}

func routeDescriptionToStaticRoute(v interface{}) (*vpc.StaticRoute, error) {
	res, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("fail to cast %#v to map[string]interface{}", v)
	}

	var sr vpc.StaticRoute

	if v, ok := res["destination_prefix"].(string); ok {
		sr.Destination = &vpc.StaticRoute_DestinationPrefix{
			DestinationPrefix: v,
		}
	} else {
		return nil, errors.New("'static_route' should have a 'destination_prefix' field")
	}

	var nextHops = 0
	if v, ok := res["next_hop_address"].(string); ok && v != "" {
		sr.NextHop = &vpc.StaticRoute_NextHopAddress{
			NextHopAddress: v,
		}
		nextHops += 1
	}
	if v, ok := res["gateway_id"].(string); ok && v != "" {
		sr.NextHop = &vpc.StaticRoute_GatewayId{
			GatewayId: v,
		}
		nextHops += 1
	}

	if nextHops == 0 {
		return nil, errors.New("'static_route' should have a 'next_hop_address' or 'gateway_id' field")
	} else if nextHops > 1 {
		return nil, errors.New("'static_route' should have one of 'next_hop_address' or 'gateway_id' fields")
	}

	return &sr, nil
}

func flattenDhcpOptions(dhcpOptions *vpc.DhcpOptions) []interface{} {
	if dhcpOptions == nil {
		return nil
	}

	m := make(map[string]interface{})

	if dhcpOptions.DomainName != "" {
		m["domain_name"] = dhcpOptions.DomainName
	}

	if len(dhcpOptions.DomainNameServers) > 0 {
		m["domain_name_servers"] = dhcpOptions.DomainNameServers
	}

	if len(dhcpOptions.NtpServers) > 0 {
		m["ntp_servers"] = dhcpOptions.NtpServers
	}

	if len(m) > 0 {
		return []interface{}{m}
	}

	return nil
}

func expandDhcpOptions(d *schema.ResourceData) (*vpc.DhcpOptions, error) {
	var (
		v  interface{}
		ok bool
	)

	if v, ok = d.GetOk("dhcp_options"); !ok {
		return nil, nil
	}

	optsList := v.([]interface{})

	if len(optsList) == 0 {
		return nil, nil
	}

	optsDescriptor, ok := optsList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("fail to cast %#v to map[string]interface{}", optsList[0])
	}

	var dhcpOptions vpc.DhcpOptions

	if v, ok := optsDescriptor["domain_name"].(string); ok {
		dhcpOptions.DomainName = v
	}

	if v, ok := optsDescriptor["domain_name_servers"].([]interface{}); ok {
		vs := make([]string, 0, len(v))
		for _, s := range v {
			vs = append(vs, s.(string))
		}
		dhcpOptions.DomainNameServers = vs
	}

	if v, ok := optsDescriptor["ntp_servers"].([]interface{}); ok {
		vs := make([]string, 0, len(v))
		for _, s := range v {
			vs = append(vs, s.(string))
		}
		dhcpOptions.NtpServers = vs
	}

	return &dhcpOptions, nil
}

func securityRuleDescriptionToRuleSpec(dir string, v interface{}) (*vpc.SecurityGroupRuleSpec, error) {
	res, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("fail to cast %#v to map[string]interface{}", v)
	}

	sr := new(vpc.SecurityGroupRuleSpec)

	directionId := vpc.SecurityGroupRule_Direction_value[strings.ToUpper(dir)]
	sr.SetDirection(vpc.SecurityGroupRule_Direction(directionId))

	if v, ok := res["description"].(string); ok {
		sr.SetDescription(v)
	}

	if v, ok := res["security_group_id"].(string); ok && v != "" {
		sr.SetSecurityGroupId(v)
	}

	if v, ok := res["predefined_target"].(string); ok && v != "" {
		sr.SetPredefinedTarget(v)
	}

	if p, ok := res["protocol"].(string); ok {
		sr.SetProtocolName(strings.ToUpper(p))
	}

	if v, ok := res["labels"]; ok {
		labels, err := sdkcommon.ExpandLabels(v)
		if err != nil {
			return sr, err
		}
		sr.SetLabels(labels)
	}

	if cidr, ok := securityRuleCidrsFromMap(res); ok {
		sr.SetCidrBlocks(cidr)
	}

	ports, err := securityRulePortsFromMap(res)
	if err != nil {
		return sr, err
	}
	sr.SetPorts(ports)

	return sr, nil
}

func expandSecurityGroupRulesSpec(d *schema.ResourceData) ([]*vpc.SecurityGroupRuleSpec, error) {

	securityRules := make([]*vpc.SecurityGroupRuleSpec, 0)

	for _, dir := range []string{"egress", "ingress"} {
		if v, ok := d.GetOk(dir); ok {
			for _, rule := range v.(*schema.Set).List() {
				if r, err := securityRuleDescriptionToRuleSpec(dir, rule); err == nil {
					securityRules = append(securityRules, r)
				} else {
					return securityRules, err
				}
			}
		}
	}

	return securityRules, nil
}

func securityRuleCidrsFromMap(res map[string]interface{}) (*vpc.CidrBlocks, bool) {
	var v4Blocks interface{} = nil
	var v6Blocks interface{} = nil

	if v, ok := res["v4_cidr_blocks"]; ok {
		v4Blocks = v
	}

	if v, ok := res["v6_cidr_blocks"]; ok {
		v6Blocks = v
	}

	return sdkcommon.SecurityRuleCirds(v4Blocks, v6Blocks)
}

// return nil on ANY-typed port range
func securityRulePortsFromMap(res map[string]interface{}) (*vpc.PortRange, error) {
	port := int64(res["port"].(int))
	fromPort := int64(res["from_port"].(int))
	toPort := int64(res["to_port"].(int))

	return sdkcommon.SecurityRulePorts(port, fromPort, toPort)
}

func flattenSecurityGroupRulesProto(g *vpc.SecurityGroupRule) (port, fromPort, toPort int64) {
	port = -1
	fromPort = -1
	toPort = -1

	if ports := g.GetPorts(); ports != nil {
		if ports.FromPort == ports.ToPort {
			port = ports.FromPort
		} else {
			fromPort = ports.FromPort
			toPort = ports.ToPort
		}
	}

	return
}

func FlattenSecurityGroupRulesSpec(sg []*vpc.SecurityGroupRule) (*schema.Set, *schema.Set) {
	ingress := schema.NewSet(ResourceYandexVPCSecurityGroupRuleHash, nil)
	egress := schema.NewSet(ResourceYandexVPCSecurityGroupRuleHash, nil)

	for _, g := range sg {
		r := make(map[string]interface{})

		r["id"] = g.GetId()
		r["description"] = g.GetDescription()
		r["labels"] = g.GetLabels()
		r["protocol"] = g.GetProtocolName()
		r["security_group_id"] = g.GetSecurityGroupId()
		r["predefined_target"] = g.GetPredefinedTarget()

		r["port"], r["from_port"], r["to_port"] = flattenSecurityGroupRulesProto(g)

		if cidr := g.GetCidrBlocks(); cidr != nil {
			if cidr.V4CidrBlocks != nil {
				r["v4_cidr_blocks"] = sdkcommon.ConvertStringArrToInterface(cidr.V4CidrBlocks)
			}

			if cidr.V6CidrBlocks != nil {
				r["v6_cidr_blocks"] = sdkcommon.ConvertStringArrToInterface(cidr.V6CidrBlocks)
			}
		}

		switch g.GetDirection().String() {
		case "INGRESS":
			ingress.Add(r)
		case "EGRESS":
			egress.Add(r)
		}
	}
	return ingress, egress
}

func flattenVpcAddressDnsRecords(specs []*vpc.DnsRecord) []map[string]interface{} {
	res := make([]map[string]interface{}, len(specs))

	for i, spec := range specs {
		res[i] = map[string]interface{}{
			"fqdn":        spec.Fqdn,
			"dns_zone_id": spec.DnsZoneId,
			"ttl":         int(spec.Ttl),
			"ptr":         spec.Ptr,
		}
	}

	return res
}

func flattenExternalIpV4AddressSpec(address *vpc.ExternalIpv4Address) []interface{} {
	if address == nil {
		return nil
	}

	m := make(map[string]interface{})

	if address.Address != "" {
		m["address"] = address.GetAddress()
	}
	if address.ZoneId != "" {
		m["zone_id"] = address.GetZoneId()
	}

	if r := address.GetRequirements(); r != nil {
		if r.DdosProtectionProvider != "" {
			m["ddos_protection_provider"] = r.DdosProtectionProvider
		}
		if r.OutgoingSmtpCapability != "" {
			m["outgoing_smtp_capability"] = r.OutgoingSmtpCapability
		}
	}

	if len(m) > 0 {
		return []interface{}{m}
	}
	return nil
}

func expandAddressRequirements(addrDesc map[string]interface{}) (*vpc.AddressRequirements, bool) {
	var (
		set          bool
		requirements vpc.AddressRequirements
	)

	if v, ok := addrDesc["ddos_protection_provider"].(string); ok {
		set = true
		requirements.SetDdosProtectionProvider(v)
	}

	if v, ok := addrDesc["outgoing_smtp_capability"].(string); ok {
		set = true
		requirements.SetOutgoingSmtpCapability(v)
	}

	return &requirements, set
}

func expandVpcAddressDnsRecords(d *schema.ResourceData) ([]*vpc.DnsRecordSpec, error) {
	var (
		v  interface{}
		ok bool
	)

	if v, ok = d.GetOk("dns_record"); !ok {
		return nil, nil
	}

	specs := v.([]interface{})
	recs := make([]*vpc.DnsRecordSpec, len(specs))

	for i, raw := range specs {
		d := raw.(map[string]interface{})
		r := &vpc.DnsRecordSpec{Fqdn: d["fqdn"].(string)}
		if s, ok := d["dns_zone_id"]; ok {
			r.DnsZoneId = s.(string)
		}
		if s, ok := d["ttl"]; ok {
			r.Ttl = int64(s.(int))
		}
		if s, ok := d["ptr"]; ok {
			r.Ptr = s.(bool)
		}
		recs[i] = r
	}
	return recs, nil
}

func expandExternalIpv4Address(d *schema.ResourceData) (*vpc.ExternalIpv4AddressSpec, error) {
	var (
		v  interface{}
		ok bool
	)

	if v, ok = d.GetOk("external_ipv4_address"); !ok {
		return nil, nil
	}

	addresses := v.([]interface{})

	if len(addresses) == 0 {
		return nil, nil
	}

	addrDesc, ok := addresses[0].(map[string]interface{})
	if !ok {
		return nil, addressError("fail to cast %#v to map[string]interface{}", addresses[0])
	}

	var addrSpec vpc.ExternalIpv4AddressSpec

	if v, ok := addrDesc["address"].(string); ok {
		addrSpec.SetAddress(v)
	}

	if v, ok := addrDesc["zone_id"].(string); ok {
		addrSpec.SetZoneId(v)
	}

	if requirements, ok := expandAddressRequirements(addrDesc); ok {
		addrSpec.SetRequirements(requirements)
	}

	return &addrSpec, nil
}

func expandInternalIpv4Address(d *schema.ResourceData) (*vpc.InternalIpv4AddressSpec, error) {
	var (
		v  interface{}
		ok bool
	)

	if v, ok = d.GetOk("internal_ipv4_address"); !ok {
		return nil, nil
	}

	addresses := v.([]interface{})

	if len(addresses) == 0 {
		return nil, nil
	}

	addrDesc, ok := addresses[0].(map[string]interface{})
	if !ok {
		return nil, addressError("fail to cast %#v to map[string]interface{}", addresses[0])
	}

	var addrSpec vpc.InternalIpv4AddressSpec

	if v, ok := addrDesc["address"].(string); ok {
		addrSpec.SetAddress(v)
	}

	if v, ok := addrDesc["subnet_id"].(string); ok {
		addrSpec.SetSubnetId(v)
	}

	return &addrSpec, nil
}

func flattenInternalIpV4AddressSpec(address *vpc.InternalIpv4Address) []interface{} {
	if address == nil {
		return nil
	}

	m := make(map[string]interface{})

	if address.Address != "" {
		m["address"] = address.GetAddress()
	}
	if address.GetSubnetId() != "" {
		m["subnet_id"] = address.GetSubnetId()
	}

	if len(m) > 0 {
		return []interface{}{m}
	}
	return nil
}

func flattenSharedEgressGateway(sharedEgressGateway *vpc.SharedEgressGateway) []interface{} {
	if sharedEgressGateway == nil {
		return nil
	}

	m := make(map[string]interface{})

	return []interface{}{m}
}
