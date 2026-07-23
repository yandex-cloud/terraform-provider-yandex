package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"sort"
	"testing"
)

func Test_parseSetToStringArray(t *testing.T) {
	type args struct {
		set interface{}
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "set is nil -> return nil slice",
			args: args{
				set: nil,
			},
			want: nil,
		},
		{
			name: "set is empty -> return empty slice",
			args: args{
				set: schema.NewSet(schema.HashString, []interface{}{}),
			},
			want: []string{},
		},
		{
			name: "correct scenario",
			args: args{
				set: schema.NewSet(schema.HashString, []interface{}{
					"xyz",
					"abcabc",
					"babba",
				}),
			},
			want: []string{"xyz", "abcabc", "babba"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSetToStringArray(tt.args.set)
			sort.Strings(result)
			sort.Strings(tt.want)
			assert.Equalf(t, tt.want, result, "parseSetToStringArray(%v)", tt.args.set)
		})
	}
}

func Test_parseKafkaPermissionAllowHosts(t *testing.T) {
	type args struct {
		allowHosts interface{}
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "allowHosts is nil -> return nil slice",
			args: args{
				allowHosts: nil,
			},
			want: nil,
		},
		{
			name: "allowHosts is empty -> return nil slice",
			args: args{
				allowHosts: schema.NewSet(schema.HashString, []interface{}{}),
			},
			want: nil,
		},
		{
			name: "correct scenario",
			args: args{
				allowHosts: schema.NewSet(schema.HashString, []interface{}{
					"host2",
					"host1",
					"host3",
				}),
			},
			want: []string{"host1", "host2", "host3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseKafkaPermissionAllowHosts(tt.args.allowHosts)
			sort.Strings(result)
			sort.Strings(tt.want)
			assert.Equalf(t, tt.want, result, "parseKafkaPermissionAllowHosts(%v)", tt.args.allowHosts)
		})
	}
}

func TestParseKafkaMessageTimestampType(t *testing.T) {
	field := resourceYandexMDBKafkaClusterKafkaSettings().Schema["log_message_timestamp_type"]

	for _, value := range []string{
		"MESSAGE_TIMESTAMP_TYPE_CREATE_TIME",
		"MESSAGE_TIMESTAMP_TYPE_LOG_APPEND_TIME",
	} {
		t.Run(value, func(t *testing.T) {
			warnings, errors := field.ValidateFunc(value, "log_message_timestamp_type")
			assert.Empty(t, warnings)
			assert.Empty(t, errors)
		})
	}

	for _, value := range []string{
		"MESSAGE_TIMESTAMP_TYPE_UNSPECIFIED",
		"LOG_APPEND_TIME",
		"",
	} {
		t.Run("reject_"+value, func(t *testing.T) {
			warnings, errors := field.ValidateFunc(value, "log_message_timestamp_type")
			assert.Empty(t, warnings)
			require.Len(t, errors, 1)
			assert.Contains(t, errors[0].Error(), "log_message_timestamp_type")
		})
	}
}

func TestFlattenKafkaLogMessageTimestampType(t *testing.T) {
	tests := []struct {
		name     string
		config   KafkaConfigSettings
		expected string
		present  bool
	}{
		{
			name: "kafka_2_8_create_time",
			config: &kafka.KafkaConfig2_8{
				LogMessageTimestampType: kafka.MessageTimestampType_MESSAGE_TIMESTAMP_TYPE_CREATE_TIME,
			},
			expected: "MESSAGE_TIMESTAMP_TYPE_CREATE_TIME",
			present:  true,
		},
		{
			name: "kafka_3_log_append_time",
			config: &kafka.KafkaConfig3{
				LogMessageTimestampType: kafka.MessageTimestampType_MESSAGE_TIMESTAMP_TYPE_LOG_APPEND_TIME,
			},
			expected: "MESSAGE_TIMESTAMP_TYPE_LOG_APPEND_TIME",
			present:  true,
		},
		{
			name:    "kafka_4_unspecified",
			config:  &kafka.KafkaConfig4{},
			present: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flattened, err := flattenKafkaConfigSettings(test.config)
			require.NoError(t, err)

			value, present := flattened["log_message_timestamp_type"]
			assert.Equal(t, test.present, present)
			if test.present {
				assert.Equal(t, test.expected, value)
			}
		})
	}
}
