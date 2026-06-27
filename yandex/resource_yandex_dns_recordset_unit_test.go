package yandex

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestCanonicalizeTXTRecordValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "split two segments",
			input: "\"chunk1\" \"chunk2\"",
			want:  "\"chunk1chunk2\"",
		},
		{
			name:  "already joined idempotent",
			input: "\"chunk1chunk2\"",
			want:  "\"chunk1chunk2\"",
		},
		{
			name:  "three segments",
			input: "\"a\" \"b\" \"c\"",
			want:  "\"abc\"",
		},
		{
			name:  "escaped quote single segment",
			input: "\"a\\\"b\"",
			want:  "\"a\\\"b\"",
		},
		{
			name:  "escaped quote at segment close",
			input: `"a\"" "b"`,
			want:  `"a\"b"`,
		},
		{
			name:  "escaped backslash",
			input: "\"a\\\\b\"",
			want:  "\"a\\\\b\"",
		},
		{
			name:  "escaped backslash across join",
			input: `"a\\" "b"`,
			want:  `"a\\b"`,
		},
		{
			name:  "unquoted unchanged",
			input: "v=spf1 include:_spf.google.com ~all",
			want:  "v=spf1 include:_spf.google.com ~all",
		},
		{
			name:  "malformed trailing junk",
			input: "\"foo\" bar",
			want:  "\"foo\" bar",
		},
		{
			name:  "unterminated",
			input: "\"foo",
			want:  "\"foo",
		},
		{
			name:  "dangling escape",
			input: `"foo\`,
			want:  `"foo\`,
		},
		{
			name:  "does not start with quote",
			input: "foo\" \"bar\"",
			want:  "foo\" \"bar\"",
		},
		{
			name:  "adjacent quotes",
			input: `"a""b"`,
			want:  `"ab"`,
		},
		{
			name:  "empty segments collapse",
			input: "\"\" \"a\"",
			want:  "\"a\"",
		},
		{
			name:  "all empty segments collapse",
			input: `"" ""`,
			want:  `""`,
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "multiple spaces/tabs between",
			input: "\"a\"\t \"b\"",
			want:  "\"ab\"",
		},
		{
			name:  "crlf between segments",
			input: "\"a\"\r\n\"b\"",
			want:  "\"ab\"",
		},
		{
			name:  "base64 case preservation",
			input: `"MIIBIjANBg" "kqhKiG9w0BAQ"`,
			want:  `"MIIBIjANBgkqhKiG9w0BAQ"`,
		},
		{
			name:  "very long split input",
			input: `"` + strings.Repeat("a", 600) + `" "` + strings.Repeat("b", 500) + `"`,
			want:  `"` + strings.Repeat("a", 600) + strings.Repeat("b", 500) + `"`,
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

func TestDNSRecordSetDataHashEquivalence(t *testing.T) {
	split := "\"chunk1\" \"chunk2\""
	joined := "\"chunk1chunk2\""
	plain := "1.2.3.4"

	hashSplit := dnsRecordSetDataHash(split)
	hashJoined := dnsRecordSetDataHash(joined)
	if hashSplit != hashJoined {
		t.Errorf("split hash %d != joined hash %d (split and joined should be equal)", hashSplit, hashJoined)
	}

	expectedJoined := schema.HashString(joined)
	if hashJoined != expectedJoined {
		t.Errorf("dnsRecordSetDataHash(%q) = %d, want schema.HashString(%q) = %d", joined, hashJoined, joined, expectedJoined)
	}

	hashPlain := dnsRecordSetDataHash(plain)
	expectedPlain := schema.HashString(plain)
	if hashPlain != expectedPlain {
		t.Errorf("dnsRecordSetDataHash(%q) = %d, want schema.HashString(%q) = %d", plain, hashPlain, plain, expectedPlain)
	}

	hashCaseA := dnsRecordSetDataHash(`"AbC" "dEF"`)
	hashCaseB := dnsRecordSetDataHash(`"abc" "def"`)
	if hashCaseA == hashCaseB {
		t.Errorf("case-distinct TXT values produced the same hash %d", hashCaseA)
	}

	for i := 0; i < 10; i++ {
		if got := dnsRecordSetDataHash(split); got != hashSplit {
			t.Errorf("dnsRecordSetDataHash(%q) iteration %d = %d, want deterministic hash %d", split, i, got, hashSplit)
		}
	}
}
