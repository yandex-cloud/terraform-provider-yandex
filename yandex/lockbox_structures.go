package yandex

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type lockboxEntryCheck struct {
	Key string
	// set Val or Regexp
	Val    string
	Regexp *regexp.Regexp
}

func flattenLockboxVersion(v *lockbox.Version) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["created_at"] = getTimestamp(v.CreatedAt)
	m["description"] = v.Description
	m["destroy_at"] = getTimestamp(v.DestroyAt)
	m["id"] = v.Id
	m["payload_entry_keys"] = v.PayloadEntryKeys
	m["secret_id"] = v.SecretId
	m["status"] = v.Status.String()

	return []map[string]interface{}{m}, nil
}

func expandLockboxSecretVersionEntriesSlice(ctx context.Context, d *schema.ResourceData) ([]*lockbox.PayloadEntryChange, error) {
	count := d.Get("entries.#").(int)
	slice := make([]*lockbox.PayloadEntryChange, count)

	for i := 0; i < count; i++ {
		versionPayloadEntries, err := expandLockboxSecretVersionEntries(ctx, d, i)
		if err != nil {
			return nil, err
		}

		slice[i] = versionPayloadEntries
	}

	return slice, nil
}

func expandLockboxSecretVersionEntries(ctx context.Context, d *schema.ResourceData, indexes ...interface{}) (*lockbox.PayloadEntryChange, error) {
	val := new(lockbox.PayloadEntryChange)

	if v, ok := d.GetOk(fmt.Sprintf("entries.%d.key", indexes...)); ok {
		val.SetKey(v.(string))
	}

	if v, ok := d.GetOk(fmt.Sprintf("entries.%d.text_value", indexes...)); ok {
		val.SetTextValue(v.(string))
	}

	if execRaw, ok := d.GetOk(fmt.Sprintf("entries.%d.command.0", indexes...)); ok {
		if val.GetTextValue() != "" {
			// We must validate manually - https://github.com/hashicorp/terraform-plugin-sdk/issues/470
			return nil, fmt.Errorf("key %v has both text_value and command, but only one of those must be set", val.GetKey())
		}
		execMap := execRaw.(map[string]interface{})
		result, err := resolveCommand(ctx, execMap)
		if err != nil {
			return nil, err
		}
		val.SetTextValue(result)
	}

	if val.GetTextValue() == "" {
		return nil, fmt.Errorf("no value for key %v", val.GetKey())
	}

	return val, nil
}

func resolveCommand(ctx context.Context, command map[string]interface{}) (string, error) {
	path := command["path"].(string)
	envMap := expandStringStringMap(command["env"].(map[string]interface{}))
	args := expandStringSlice(command["args"].([]interface{}))
	result, err := runCommand(ctx, path, envMap, args...)
	return result, err
}

func runCommand(ctx context.Context, path string, envMap map[string]string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, path, args...)
	for k, v := range envMap {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	command.Stdout = &outBuf
	command.Stderr = &errBuf

	err := command.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %s %s", err, outBuf.String(), errBuf.String())
	}

	return outBuf.String(), nil
}

func flattenLockboxSecretVersionEntriesSlice(vs []*lockbox.Payload_Entry) ([]interface{}, error) {
	s := make([]interface{}, 0, len(vs))
	for _, v := range vs {
		s = append(s, flattenLockboxSecretVersionEntry(v))
	}
	return s, nil
}

func flattenLockboxSecretVersionEntry(v *lockbox.Payload_Entry) map[string]interface{} {
	return map[string]interface{}{
		"key":        v.Key,
		"text_value": v.GetTextValue(),
	}
}

func flattenPasswordPayloadSpecification(passwordPayloadSpecification *lockbox.PasswordPayloadSpecification) []map[string]interface{} {
	if passwordPayloadSpecification == nil {
		return nil
	}

	m := make(map[string]interface{})

	m["password_key"] = passwordPayloadSpecification.PasswordKey
	if passwordPayloadSpecification.Length != 0 {
		m["length"] = passwordPayloadSpecification.Length
	}
	if passwordPayloadSpecification.IncludeUppercase != nil {
		m["include_uppercase"] = passwordPayloadSpecification.IncludeUppercase.Value
	}
	if passwordPayloadSpecification.IncludeLowercase != nil {
		m["include_lowercase"] = passwordPayloadSpecification.IncludeLowercase.Value
	}
	if passwordPayloadSpecification.IncludeDigits != nil {
		m["include_digits"] = passwordPayloadSpecification.IncludeDigits.Value
	}
	if passwordPayloadSpecification.IncludePunctuation != nil {
		m["include_punctuation"] = passwordPayloadSpecification.IncludePunctuation.Value
	}
	if passwordPayloadSpecification.IncludedPunctuation != "" {
		m["included_punctuation"] = passwordPayloadSpecification.IncludedPunctuation
	}
	if passwordPayloadSpecification.ExcludedPunctuation != "" {
		m["excluded_punctuation"] = passwordPayloadSpecification.ExcludedPunctuation
	}
	return []map[string]interface{}{m}
}

func expandPasswordPayloadSpecification(d *schema.ResourceData) (*lockbox.PasswordPayloadSpecification, error) {
	if v, ok := d.GetOk("password_payload_specification"); !ok || len(v.([]interface{})) == 0 {
		return nil, nil
	}
	var pps lockbox.PasswordPayloadSpecification
	pps.PasswordKey = d.Get("password_payload_specification.0.password_key").(string)
	pps.Length = int64(d.Get("password_payload_specification.0.length").(int))
	pps.IncludeUppercase = &wrapperspb.BoolValue{Value: d.Get("password_payload_specification.0.include_uppercase").(bool)}
	pps.IncludeLowercase = &wrapperspb.BoolValue{Value: d.Get("password_payload_specification.0.include_lowercase").(bool)}
	pps.IncludeDigits = &wrapperspb.BoolValue{Value: d.Get("password_payload_specification.0.include_digits").(bool)}
	pps.IncludePunctuation = &wrapperspb.BoolValue{Value: d.Get("password_payload_specification.0.include_punctuation").(bool)}
	pps.IncludedPunctuation = d.Get("password_payload_specification.0.included_punctuation").(string)
	pps.ExcludedPunctuation = d.Get("password_payload_specification.0.excluded_punctuation").(string)
	return &pps, nil
}

func testAccOutputToLockbox(secretId, sensitiveAttr, entryKey string) string {
	return testAccOutputToLockboxAsMap(secretId, map[string]string{sensitiveAttr: entryKey})
}

func testAccOutputToLockboxAsMap(secretId string, sensitiveAttrToEntryKey map[string]string) string {
	var entries string
	for k, v := range sensitiveAttrToEntryKey {
		entries += fmt.Sprintf("  entry_for_%s = \"%s\"\n", k, v)
	}
	return fmt.Sprintf(`
output_to_lockbox {
  secret_id = %s
%s
}
`, secretId, entries)
}
