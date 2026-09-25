package ydb

import (
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

// String returns string representation of Driver
func (d *Driver) String() string {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteString("Driver{")
	fmt.Fprintf(buffer, "Endpoint:%q,", d.config.Endpoint())
	fmt.Fprintf(buffer, "Database:%q,", d.config.Database())
	fmt.Fprintf(buffer, "Secure:%v", d.config.Secure())
	if c, has := d.config.Credentials().(fmt.Stringer); has {
		fmt.Fprintf(buffer, ",Credentials:%v", c.String())
	}
	buffer.WriteByte('}')

	return buffer.String()
}
