# Streaming text replacement

[![GoDoc](https://godoc.org/github.com/icholy/replace?status.svg)](https://godoc.org/github.com/icholy/replace)

> This package provides a [x/text/transform.Transformer](https://godoc.org/golang.org/x/text/transform#Transformer)
> implementation that replaces text

## Example

``` go
package main

import (
	"io"
	"os"
	"regexp"

	"github.com/icholy/replace"
)

func main() {
	f, _ := os.Open("file")
	defer f.Close()

	r := replace.Chain(f,
		// simple replace
		replace.String("foo", "bar"),
		replace.Bytes([]byte("thing"), []byte("test")),

		// remove all words that start with baz
		replace.Regexp(regexp.MustCompile(`baz\w*`), nil),

		// surround all words with parentheses
		replace.RegexpString(regexp.MustCompile(`\w+`), "($0)"),

		// increment all numbers
		replace.RegexpStringFunc(regexp.MustCompile(`\d+`), func(match string) string {
			x, _ := strconv.Atoi(match)
			return strconv.Itoa(x+1)
		}),
	)

	_, _ = io.Copy(os.Stdout, r)
}
```

## Notes Regexp* functions

* `RegexpTransformer` is stateful and cannot be used concurrently.
* The `replace` functions should not save or modify any `[]byte` parameters they recieve.
* If a match is longer than `MaxMatchSize` it may be skipped (Default 2kb).
* For better performance, reduce the `MaxMatchSize` size to the largest possible match.
* Do not use with [transform.Chain](https://pkg.go.dev/golang.org/x/text/transform#Chain), see https://github.com/golang/go/issues/49117.