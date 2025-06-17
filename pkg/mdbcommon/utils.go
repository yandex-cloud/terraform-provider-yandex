package mdbcommon

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	protobuf_adapter "github.com/yandex-cloud/terraform-provider-yandex/pkg/adapters/protobuf"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetUserConfig extracts UserConfig from a cluster config struct by protobuf tag
//
// c must be a valid ClusterConfigStructure
func GetUserConfig(ctx context.Context, c interface{}, confAttrName string, diags *diag.Diagnostics) interface{} {

	topicErr := fmt.Sprintf("Failed to flatten %s", confAttrName)

	if c == nil {
		return nil
	}

	rc := reflect.ValueOf(c)

	if rc.Kind() == reflect.Ptr {
		if rc.IsNil() {
			diags.AddError(
				topicErr,
				fmt.Sprintf("Can't scan type %T for extract attributes. It's error in provider", c),
			)
			return nil
		}

		rc = rc.Elem()
	}

	if rc.Kind() != reflect.Struct {
		diags.AddError(
			topicErr,
			fmt.Sprintf("Can't scan type %T for extract attributes. It's error in provider", c),
		)
		return nil
	}

	rcType := rc.Type()
	var pgConf reflect.Value
	for i := 0; i < rcType.NumField(); i++ {
		field := rcType.Field(i)
		t, ok := protobuf_adapter.FindTag(field, "protobuf", "name")
		if !ok {
			continue
		}

		if !strings.Contains(t, confAttrName) {
			continue
		}

		pgConf = rc.Field(i)
	}
	if !pgConf.IsValid() {
		diags.AddError(
			topicErr,
			fmt.Sprintf(
				`
				Can't find %s in source struct type %T
				It's error in provider.
				`, confAttrName, c,
			),
		)
		return nil
	}

	if pgConf.Kind() == reflect.Ptr {
		pgConf = pgConf.Elem()
	}
	if pgConf.Kind() != reflect.Struct {
		diags.AddError(
			topicErr,
			fmt.Sprintf(
				`
				Can't scan type %T for extract attributes: %s must be a struct. 
				It's error in provider.
				`, c, confAttrName,
			),
		)
		return nil
	}

	pgConfType := pgConf.Type()
	var uConf interface{}
	for i := 0; i < pgConfType.NumField(); i++ {
		field := pgConfType.Field(i)
		t, ok := protobuf_adapter.FindTag(field, "protobuf", "name")
		if !ok {
			continue
		}

		if t != "user_config" {
			continue
		}

		uConf = pgConf.Field(i).Interface()
	}

	if uConf == nil {
		diags.AddError(
			topicErr,
			fmt.Sprintf(
				`
				Can't find user config in source struct type %T
				It's error in provider.
				`, c,
			),
		)
	}

	return uConf
}

func IsAttrZeroValue(val attr.Value, diags *diag.Diagnostics) bool {
	if val.IsNull() || val.IsUnknown() {
		return true
	}

	if valInt, ok := val.(types.Int64); ok {
		if valInt.ValueInt64() != 0 {
			return false
		}
		return true
	}

	if valStr, ok := val.(types.String); ok {
		if valStr.ValueString() != "" {
			return false
		}
		return true
	}

	if _, ok := val.(types.Bool); ok {
		return false
	}

	if _, ok := val.(types.List); ok {
		return false
	}

	if valFloat, ok := val.(types.Float64); ok {
		if valFloat.ValueFloat64() != 0 {
			return false
		}
		return true
	}

	if valNum, ok := val.(types.Number); ok {
		i, _ := valNum.ValueBigFloat().Int64()
		if !valNum.ValueBigFloat().IsInt() || i != 0 {
			return false
		}
		return true
	}

	if _, ok := val.(types.Tuple); ok {
		return false
	}

	diags.AddError("Zero value check error", fmt.Sprintf("Attribute has a unknown handling value %v", val.String()))
	return true
}

func GetAttrNamesSetFromMap(o types.Map, diags *diag.Diagnostics) map[string]struct{} {

	attrs := make(map[string]struct{})
	if o.IsNull() || o.IsUnknown() {
		return attrs
	}

	for attr := range o.Elements() {
		attrs[attr] = struct{}{}
	}

	return attrs
}

func FixDiskSizeOnAutoscalingChanges(ctx context.Context, plan, state types.Object, autoscalingEnabled bool, diags *diag.Diagnostics) types.Object {
	if state.IsNull() {
		return plan
	}
	planR := &Resource{}
	stateR := &Resource{}
	diags.Append(plan.As(ctx, planR, baseOptions)...)
	diags.Append(state.As(ctx, stateR, baseOptions)...)
	if diags.HasError() {
		return plan
	}

	if (stateR.DiskSize.ValueInt64() > planR.DiskSize.ValueInt64()) && autoscalingEnabled {
		planR.DiskSize = stateR.DiskSize
		obj, d := types.ObjectValueFrom(ctx, ResourceType.AttrTypes, planR)
		diags.Append(d...)
		return obj
	}
	return plan
}
