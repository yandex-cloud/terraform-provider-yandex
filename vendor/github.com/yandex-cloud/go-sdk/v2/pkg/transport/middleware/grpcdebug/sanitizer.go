package grpcdebug

import (
	"regexp"
	"strings"
)

// This file vendors a subset of api/pkg/antisecret into the sdk-v2 tree so the
// grpcdebug interceptor can mask credentials in logged payloads without taking
// an external dependency. The pattern set and redaction logic mirror
// antisecret.DefaultPatterns / antisecret.Sanitizer at the time of vendoring;
// keep in sync with /Users/.../cloud-go/pkg/antisecret/sanitizer.go when adding
// new patterns.

type redactMode int

const (
	redactModeReplace redactMode = iota
	redactModePrefix
	redactModeKeyValue
	redactModeBasicAuthURL
)

type secretPattern struct {
	name   string
	regexp *regexp.Regexp
	prefix string
	mode   redactMode
}

const (
	redactedPlaceholder = "[REDACTED]"
	redactChar          = '*'
)

var defaultSecretPatterns = []secretPattern{
	{
		name:   "yandex_cloud_iam_cookie_v1",
		regexp: regexp.MustCompile(`c1\.[A-Z0-9a-z_-]+[=]{0,2}\.[A-Z0-9a-z_-]{43,}[=]{0,2}`),
		prefix: "c1.",
		mode:   redactModePrefix,
	},
	{
		name:   "yandex_cloud_iam_token_v1",
		regexp: regexp.MustCompile(`t1\.[A-Z0-9a-z_-]+[=]{0,2}\.[A-Z0-9a-z_-]{43,}[=]{0,2}`),
		prefix: "t1.",
		mode:   redactModePrefix,
	},
	{
		name:   "yandex_cloud_iam_api_key_v1",
		regexp: regexp.MustCompile(`AQVN[A-Za-z0-9_-]{35,38}`),
		prefix: "AQVN",
		mode:   redactModePrefix,
	},
	{
		name:   "yandex_cloud_iam_access_secret",
		regexp: regexp.MustCompile(`YC[a-zA-Z0-9_-]{38}`),
		prefix: "YC",
		mode:   redactModePrefix,
	},
	{
		name:   "yandex_cloud_refresh_token",
		regexp: regexp.MustCompile(`rt1\.[A-Z0-9a-z_-]+[=]{0,2}\.[A-Z0-9a-z_-]{43,}[=]{0,2}`),
		prefix: "rt1.",
		mode:   redactModePrefix,
	},
	{
		name:   "yandex_cloud_session_token",
		regexp: regexp.MustCompile(`s1\.[A-Z0-9a-z_-]+[=]{0,2}\.[A-Z0-9a-z_-]{43,}[=]{0,2}`),
		prefix: "s1.",
		mode:   redactModePrefix,
	},
	{
		// Lower threshold than upstream antisecret ({50,}) — the y[0-6]_ prefix
		// is already a strong signal, and we have observed shorter/anonymized
		// tokens slipping past the 50-char gate in debug output.
		name:   "yandex_passport_oauth_token",
		regexp: regexp.MustCompile(`y[0-6]_[A-Za-z0-9_-]{20,}`),
		prefix: "y",
		mode:   redactModePrefix,
	},
	{
		name:   "jwt_token",
		regexp: regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`),
		prefix: "eyJ",
		mode:   redactModePrefix,
	},
	{
		name:   "aws_access_key_id",
		regexp: regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		prefix: "AKIA",
		mode:   redactModePrefix,
	},
	{
		name:   "aws_secret_access_key",
		regexp: regexp.MustCompile(`(?i)aws_secret_access_key\s*[=:]\s*[A-Za-z0-9/+=]{40}`),
		mode:   redactModeKeyValue,
	},
	{
		name:   "bearer_token",
		regexp: regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9_-]{20,}`),
		prefix: "Bearer ",
		mode:   redactModePrefix,
	},
	{
		name:   "basic_auth_url",
		regexp: regexp.MustCompile(`://[^:]+:[^@]+@`),
		mode:   redactModeBasicAuthURL,
	},
	{
		name:   "password_value",
		regexp: regexp.MustCompile(`(?i)password\s*[=:]\s*[^\s"']+`),
		mode:   redactModeKeyValue,
	},
}

// sanitizeSecrets runs every default pattern against the input string and
// returns a copy with each match redacted according to its mode.
func sanitizeSecrets(text string) string {
	for _, p := range defaultSecretPatterns {
		pattern := p // capture for closure
		text = p.regexp.ReplaceAllStringFunc(text, func(match string) string {
			return redactSecret(match, pattern)
		})
	}
	return text
}

func redactSecret(value string, p secretPattern) string {
	switch p.mode {
	case redactModePrefix:
		return redactWithPrefix(value, p.prefix)
	case redactModeKeyValue:
		return redactKeyValue(value)
	case redactModeBasicAuthURL:
		return "://" + redactedPlaceholder + "@"
	default:
		return redactedPlaceholder
	}
}

func redactWithPrefix(value, prefix string) string {
	if prefix == "" {
		return redactedPlaceholder
	}
	maskLen := len(value) - len(prefix)
	if maskLen <= 0 {
		return redactedPlaceholder
	}
	return prefix + strings.Repeat(string(redactChar), maskLen)
}

func redactKeyValue(value string) string {
	sepIdx := strings.IndexAny(value, "=:")
	if sepIdx == -1 {
		return redactedPlaceholder
	}
	endIdx := sepIdx + 1
	for endIdx < len(value) && (value[endIdx] == ' ' || value[endIdx] == '\t') {
		endIdx++
	}
	keyPart := value[:endIdx]
	return keyPart + redactedPlaceholder
}
