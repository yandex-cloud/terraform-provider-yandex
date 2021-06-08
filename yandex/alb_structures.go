package yandex

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

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
