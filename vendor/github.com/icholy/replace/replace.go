package replace

import (
	"bytes"
	"io"
	"regexp"

	"golang.org/x/text/transform"
)

// Chain returns a reader which applies all provided transformers.
func Chain(r io.Reader, tt ...transform.Transformer) io.Reader {
	for _, t := range tt {
		r = transform.NewReader(r, t)
	}
	return r
}

// Transformer replaces text in a stream
// See: http://golang.org/x/text/transform
type Transformer struct {
	transform.NopResetter

	old, new []byte
	oldlen   int
}

var _ transform.Transformer = (*Transformer)(nil)

// Bytes returns a transformer that replaces all instances of old with new.
// Unlike bytes.Replace, empty old values don't match anything.
func Bytes(old, new []byte) Transformer {
	return Transformer{old: old, new: new, oldlen: len(old)}
}

// String returns a transformer that replaces all instances of old with new.
// Unlike strings.Replace, empty old values don't match anything.
func String(old, new string) Transformer {
	return Bytes([]byte(old), []byte(new))
}

// Transform implements golang.org/x/text/transform#Transformer
func (t Transformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	var n int
	// don't do anything for empty old string. We're forced to do this because an optimization in
	// transform.String prevents us from generating any output when the src is empty.
	// see: https://github.com/golang/text/blob/master/transform/transform.go#L570-L576
	if t.oldlen == 0 {
		n, err = fullcopy(dst, src)
		return n, n, err
	}
	// replace all instances of old with new
	for {
		i := bytes.Index(src[nSrc:], t.old)
		if i == -1 {
			break
		}
		// copy everything up to the match
		n, err = fullcopy(dst[nDst:], src[nSrc:nSrc+i])
		nSrc += n
		nDst += n
		if err != nil {
			return
		}
		// copy the new value
		n, err = fullcopy(dst[nDst:], t.new)
		if err != nil {
			return
		}
		nDst += n
		nSrc += t.oldlen
	}
	// if we're at the end, tack on any remaining bytes
	if atEOF {
		n, err = fullcopy(dst[nDst:], src[nSrc:])
		nDst += n
		nSrc += n
		return
	}
	// skip everything except the trailing len(r.old) - 1
	// we do this becasue there could be a match straddling
	// the boundary
	if skip := len(src[nSrc:]) - t.oldlen + 1; skip > 0 {
		n, err = fullcopy(dst[nDst:], src[nSrc:nSrc+skip])
		nSrc += n
		nDst += n
		if err != nil {
			return
		}
	}
	err = transform.ErrShortSrc
	return
}

// RegexpTransformer replaces regexp matches in a stream
// See: http://golang.org/x/text/transform
// Note: this Transformer is not safe for concurrent use.
type RegexpTransformer struct {
	// MaxMatchSize is the maximum size of a regexp match.
	// If a match exceeds this limit, it may be omitted.
	// (Default is 2kb).
	MaxMatchSize int

	re       *regexp.Regexp
	replace  func(src []byte, index []int) []byte
	overflow []byte
}

var _ transform.Transformer = (*RegexpTransformer)(nil)

// RegexpIndexFunc returns a transformer that replaces all matches of re with the return value of replace.
// The replace function recieves the underlying src buffer and indexes into that buffer.
// The []byte parameter passed to replace should not be modified and is not guaranteed to be valid after the function returns.
func RegexpIndexFunc(re *regexp.Regexp, replace func(src []byte, index []int) []byte) *RegexpTransformer {
	return &RegexpTransformer{
		re:           re,
		replace:      replace,
		MaxMatchSize: 2 << 10,
	}
}

// Regexp returns a transformer that replaces all matches of re with new
func Regexp(re *regexp.Regexp, new []byte) *RegexpTransformer {
	return RegexpIndexFunc(re, func(_ []byte, _ []int) []byte { return new })
}

// RegexpString returns a transformer that replaces all matches of re with template
// Inside template, $ signs are interpreted as in Expand, so for instance $1 represents the text of the first submatch.
func RegexpString(re *regexp.Regexp, template string) *RegexpTransformer {
	return RegexpIndexFunc(re, func(src []byte, index []int) []byte {
		return re.Expand(nil, []byte(template), src, index)
	})
}

// RegexpFunc returns a transformer that replaces all matches of re with the result of calling replace with the match.
// The []byte parameter passed to replace should not be modified and is not guaranteed to be valid after the function returns.
func RegexpFunc(re *regexp.Regexp, replace func([]byte) []byte) *RegexpTransformer {
	return RegexpIndexFunc(re, func(src []byte, index []int) []byte {
		return replace(src[index[0]:index[1]])
	})
}

// RegexpStringFunc returns a transformer that replaces all matches of re with the result of calling replace with the match.
func RegexpStringFunc(re *regexp.Regexp, replace func(string) string) *RegexpTransformer {
	return RegexpIndexFunc(re, func(src []byte, index []int) []byte {
		return []byte(replace(string(src[index[0]:index[1]])))
	})
}

// RegexpSubmatchFunc returns a transformer that replaces all matches of re with the result of calling replace with the submatch.
// The [][]byte parameter passed to replace should not be modified and is not guaranteed to be valid after the function returns.
func RegexpSubmatchFunc(re *regexp.Regexp, replace func([][]byte) []byte) *RegexpTransformer {
	return RegexpIndexFunc(re, func(src []byte, index []int) []byte {
		match := make([][]byte, 1+re.NumSubexp())
		for i := range match {
			start, end := index[i*2], index[i*2+1]
			if start >= 0 {
				match[i] = src[start:end]
			}
		}
		return replace(match)
	})
}

// RegexpStringSubmatchFunc returns a transformer that replaces all matches of re with the result of calling replace with the submatch.
func RegexpStringSubmatchFunc(re *regexp.Regexp, replace func([]string) string) *RegexpTransformer {
	return RegexpIndexFunc(re, func(src []byte, index []int) []byte {
		match := make([]string, 1+re.NumSubexp())
		for i := range match {
			start, end := index[i*2], index[i*2+1]
			if start >= 0 {
				match[i] = string(src[start:end])
			}
		}
		return []byte(replace(match))
	})
}

// Transform implements golang.org/x/text/transform#Transformer
func (t *RegexpTransformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	var n int
	// copy any overflow from the last call
	if len(t.overflow) > 0 {
		n, err = fullcopy(dst, t.overflow)
		nDst += n
		if err != nil {
			t.overflow = t.overflow[n:]
			return
		}
		t.overflow = nil
	}
	for _, index := range t.re.FindAllSubmatchIndex(src, -1) {
		// copy everything up to the match
		n, err = fullcopy(dst[nDst:], src[nSrc:index[0]])
		nSrc += n
		nDst += n
		if err != nil {
			return
		}
		// skip the match if it ends at the end the src buffer.
		// it could potentially match more
		if index[1] == len(src) && !atEOF {
			break
		}
		// copy the replacement
		rep := t.replace(src, index)
		n, err = fullcopy(dst[nDst:], rep)
		nDst += n
		nSrc = index[1]
		if err != nil {
			t.overflow = rep[n:]
			return
		}
	}
	// if we're at the end, tack on any remaining bytes
	if atEOF {
		n, err = fullcopy(dst[nDst:], src[nSrc:])
		nDst += n
		nSrc += n
		return
	}
	// skip any bytes which exceede the max match size
	if skip := len(src[nSrc:]) - t.MaxMatchSize; skip > 0 {
		n, err = fullcopy(dst[nDst:], src[nSrc:nSrc+skip])
		nSrc += n
		nDst += n
		if err != nil {
			return
		}
	}
	err = transform.ErrShortSrc
	return
}

// Reset resets the state and allows a Transformer to be reused.
func (t *RegexpTransformer) Reset() {
	t.overflow = nil
}

func fullcopy(dst, src []byte) (n int, err error) {
	n = copy(dst, src)
	if n < len(src) {
		err = transform.ErrShortDst
	}
	return
}
