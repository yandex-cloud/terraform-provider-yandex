package sdkcommon

import (
	"fmt"

	"github.com/golang/protobuf/ptypes/duration"
)

func ParsePositiveDuration(s string) (*duration.Duration, error) {
	d, err := ParseDuration(s)
	if err != nil {
		return nil, err
	}

	if d.GetSeconds() == 0 && d.GetNanos() == 0 {
		return nil, fmt.Errorf("can not use zero duration")
	}

	return d, nil
}
