package yandex

import "strings"

// canonicalizeTXTRecordValue joins a TXT value only when its segmentation
// matches the DNS API's 255-byte chunking of a longer character-string.
// Intentional shorter segments, escapes, and malformed input are preserved.
func canonicalizeTXTRecordValue(s string) string {
	if len(s) == 0 || s[0] != '"' {
		return s
	}

	type segment struct {
		content    string
		decodedLen int
	}

	segments := make([]segment, 0, 2)
	i := 0
	for {
		if i >= len(s) || s[i] != '"' {
			return s
		}
		i++
		contentStart := i
		decodedLen := 0

		for i < len(s) {
			switch s[i] {
			case '\\':
				if i+1 >= len(s) {
					return s
				}
				decodedLen++
				if s[i+1] >= '0' && s[i+1] <= '9' {
					if i+3 >= len(s) ||
						s[i+2] < '0' || s[i+2] > '9' ||
						s[i+3] < '0' || s[i+3] > '9' {
						return s
					}
					escapedByte := int(s[i+1]-'0')*100 + int(s[i+2]-'0')*10 + int(s[i+3]-'0')
					if escapedByte > 255 {
						return s
					}
					i += 4
					continue
				}
				i += 2
			case '"':
				segments = append(segments, segment{content: s[contentStart:i], decodedLen: decodedLen})
				i++
				goto segmentClosed
			default:
				decodedLen++
				i++
			}
		}
		return s

	segmentClosed:
		separatorStart := i
		for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r' || s[i] == '\n') {
			i++
		}
		if i == len(s) {
			if i != separatorStart {
				return s
			}
			break
		}
	}

	if len(segments) < 2 {
		return s
	}
	for i, segment := range segments {
		if segment.decodedLen == 0 || segment.decodedLen > 255 {
			return s
		}
		if i < len(segments)-1 && segment.decodedLen != 255 {
			return s
		}
	}

	var joined strings.Builder
	joined.Grow(len(s))
	joined.WriteByte('"')
	for _, segment := range segments {
		joined.WriteString(segment.content)
	}
	joined.WriteByte('"')
	return joined.String()
}
