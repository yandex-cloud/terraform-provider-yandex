package retry

import (
	"fmt"
	"strings"
	"time"
)

// Duration defines JSON marshal and unmarshal methods to conform to the
// protobuf JSON spec defined [here].
//
// [here]: https://protobuf.dev/reference/protobuf/google.protobuf/#duration
type Duration time.Duration

func (d Duration) String() string {
	return fmt.Sprint(time.Duration(d))
}

// MarshalJSON converts from d to a JSON string output.
func (d Duration) MarshalJSON() ([]byte, error) {
	ns := time.Duration(d).Nanoseconds()
	sec := ns / int64(time.Second)
	ns = ns % int64(time.Second)

	var sign string
	if sec < 0 || ns < 0 {
		sign, sec, ns = "-", -1*sec, -1*ns
	}

	// Generated output always contains 0, 3, 6, or 9 fractional digits,
	// depending on required precision.
	str := fmt.Sprintf("%s%d.%09d", sign, sec, ns)
	str = strings.TrimSuffix(str, "000")
	str = strings.TrimSuffix(str, "000")
	str = strings.TrimSuffix(str, ".000")

	return []byte(fmt.Sprintf("\"%ss\"", str)), nil
}
