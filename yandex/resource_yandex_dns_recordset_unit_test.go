package yandex

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	terraform2 "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestCanonicalizeTXTRecordValue(t *testing.T) {
	fullChunk := strings.Repeat("a", 255)
	secondFullChunk := strings.Repeat("b", 255)
	escapedFullChunk := strings.Repeat("a", 254) + `\123`
	escapedQuoteFullChunk := strings.Repeat("a", 254) + `\"`
	oneDigitEscapeFullChunk := strings.Repeat("a", 254) + `\1`
	twoDigitEscapeFullChunk := strings.Repeat("a", 253) + `\12`
	outOfRangeEscapeFullChunk := strings.Repeat("a", 254) + `\256`
	utf8FullChunk := strings.Repeat("a", 253) + "é"
	overlongFinalChunk := strings.Repeat("b", 256)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "joins API split after 255 decoded bytes",
			input: `"` + fullChunk + `" "tail"`,
			want:  `"` + fullChunk + `tail"`,
		},
		{
			name:  "joins multiple full API chunks",
			input: `"` + fullChunk + `" "` + secondFullChunk + `" "tail"`,
			want:  `"` + fullChunk + secondFullChunk + `tail"`,
		},
		{
			name:  "numeric escape counts as one decoded byte",
			input: `"` + escapedFullChunk + `" "tail"`,
			want:  `"` + escapedFullChunk + `tail"`,
		},
		{
			name:  "escaped quote counts as one decoded byte",
			input: `"` + escapedQuoteFullChunk + `" "tail"`,
			want:  `"` + escapedQuoteFullChunk + `tail"`,
		},
		{
			name:  "join cannot complete one-digit numeric escape",
			input: `"` + oneDigitEscapeFullChunk + `" "23tail"`,
			want:  `"` + oneDigitEscapeFullChunk + `" "23tail"`,
		},
		{
			name:  "join cannot complete two-digit numeric escape",
			input: `"` + twoDigitEscapeFullChunk + `" "3tail"`,
			want:  `"` + twoDigitEscapeFullChunk + `" "3tail"`,
		},
		{
			name:  "out-of-range numeric escape remains unchanged",
			input: `"` + outOfRangeEscapeFullChunk + `" "tail"`,
			want:  `"` + outOfRangeEscapeFullChunk + `" "tail"`,
		},
		{
			name:  "UTF-8 payload is counted in bytes",
			input: `"` + utf8FullChunk + `" "tail"`,
			want:  `"` + utf8FullChunk + `tail"`,
		},
		{
			name:  "short intentional segments remain distinct",
			input: `"a" "b"`,
			want:  `"a" "b"`,
		},
		{
			name:  "adjacent short segments remain distinct",
			input: `"a""b"`,
			want:  `"a""b"`,
		},
		{
			name:  "empty first segment remains distinct",
			input: `"" "a"`,
			want:  `"" "a"`,
		},
		{
			name:  "empty final segment is not an API split",
			input: `"` + fullChunk + `" ""`,
			want:  `"` + fullChunk + `" ""`,
		},
		{
			name:  "overlong final segment is not an API split",
			input: `"` + fullChunk + `" "` + overlongFinalChunk + `"`,
			want:  `"` + fullChunk + `" "` + overlongFinalChunk + `"`,
		},
		{
			name:  "trailing whitespace is preserved",
			input: `"` + fullChunk + `" "tail" `,
			want:  `"` + fullChunk + `" "tail" `,
		},
		{
			name:  "single long segment remains unchanged",
			input: `"` + strings.Repeat("a", 600) + `"`,
			want:  `"` + strings.Repeat("a", 600) + `"`,
		},
		{
			name:  "unquoted value remains unchanged",
			input: "v=spf1 include:_spf.google.com ~all",
			want:  "v=spf1 include:_spf.google.com ~all",
		},
		{
			name:  "trailing junk remains unchanged",
			input: `"foo" bar`,
			want:  `"foo" bar`,
		},
		{
			name:  "unterminated value remains unchanged",
			input: `"foo`,
			want:  `"foo`,
		},
		{
			name:  "dangling escape remains unchanged",
			input: `"foo\`,
			want:  `"foo\`,
		},
		{
			name:  "empty string remains unchanged",
			input: "",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := canonicalizeTXTRecordValue(tt.input)
			if got != tt.want {
				t.Errorf("canonicalizeTXTRecordValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDNSRecordSetDataSchemaPreservesDistinctQuotedValues(t *testing.T) {
	dataSchema := resourceYandexDnsRecordSet().Schema["data"]
	values := schema.NewSet(dataSchema.Set, []interface{}{`"a" "b"`, `"ab"`})

	if values.Len() != 2 {
		t.Fatalf("DNS record data set contains %d values, want 2 distinct values", values.Len())
	}
}

func TestDNSRecordSetEquivalentTXTRepresentationsDoNotProduceDiffAfterRead(t *testing.T) {
	fullChunk := strings.Repeat("a", 255)
	joinedValue := `"` + fullChunk + `tail"`
	segmentedValue := `"` + fullChunk + `" "tail"`

	tests := []struct {
		name            string
		configuredValue string
	}{
		{
			name:            "joined configuration",
			configuredValue: joinedValue,
		},
		{
			name:            "explicitly segmented configuration",
			configuredValue: segmentedValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := resourceYandexDnsRecordSet()
			config := map[string]interface{}{
				"zone_id": "test-zone-id",
				"name":    "test.example.",
				"type":    "TXT",
				"ttl":     300,
				"data":    []interface{}{tt.configuredValue},
			}
			resourceData := schema.TestResourceDataRaw(t, resource.Schema, config)
			resourceData.SetId("test-zone-id/test.example./TXT")

			err := resourceData.Set("data", []interface{}{segmentedValue})
			require.NoError(t, err)
			require.Equal(t, []interface{}{segmentedValue}, resourceData.Get("data").(*schema.Set).List())

			diff, err := schema.InternalMap(resource.Schema).Diff(
				context.Background(),
				resourceData.State(),
				terraform2.NewResourceConfigRaw(config),
				nil,
				nil,
				false,
			)
			require.NoError(t, err)
			require.Nil(t, diff)
		})
	}
}

func TestDNSRecordSetDataSchemaPreservesOrdinaryValueHash(t *testing.T) {
	dataSchema := resourceYandexDnsRecordSet().Schema["data"]
	value := "192.0.2.1"

	require.Equal(t, schema.HashString(value), dataSchema.Set(value))
}
