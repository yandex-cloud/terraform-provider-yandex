package yandex

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
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
