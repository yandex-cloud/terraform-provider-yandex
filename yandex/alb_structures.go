package yandex

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

func resourceALBTargetGroupTargetHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["subnet_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["ip_address"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
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

func flattenALBTargets(tg *apploadbalancer.TargetGroup) (*schema.Set, error) {
	result := &schema.Set{F: resourceALBTargetGroupTargetHash}

	for _, t := range tg.Targets {
		flTarget := map[string]interface{}{
			"subnet_id":    t.SubnetId,
			"address_type": t.AddressType,
		}
		result.Add(flTarget)
	}

	return result, nil
}
