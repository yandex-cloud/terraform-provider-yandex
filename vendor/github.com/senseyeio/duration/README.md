Duration [![Build](https://travis-ci.org/senseyeio/duration.svg?branch=master)](https://travis-ci.org/senseyeio/duration) [![Coverage](https://coveralls.io/repos/github/senseyeio/duration/badge.svg?branch=master)](https://coveralls.io/github/senseyeio/duration?branch=master) [![Go Report Card](https://goreportcard.com/badge/senseyeio/duration)](https://goreportcard.com/report/senseyeio/duration) [![GoDoc](https://godoc.org/github.com/senseyeio/duration?status.svg)](https://godoc.org/github.com/senseyeio/duration)
=======
Parse ISO8601 duration strings, and use to shift dates/times.

Basic Example
-------------

```go
package main

import (
	"fmt"
	"time"

	"github.com/senseyeio/duration"
)

func main() {
	d, _ := iso8601.ParseISO8601("P1D")
	today := time.Now()
	tomorrow := d.Shift(today)
	fmt.Println(today.Format("Jan _2"))
	fmt.Println(tomorrow.Format("Jan _2"))
}
```

Why Does This Package Exist
---------------------------
> Why can't we just use a `time.Duration` and `time.Add`?

A very reasonable question.

The code below repeatedly adds 24 hours to a `time.Time`. You might expect the time on that date to stay the same, but [_there are not always 24 hours in a day_](http://infiniteundo.com/post/25326999628/falsehoods-programmers-believe-about-time). When the clocks change in New York, the time will skew by an hour. As you can see from the output, duration.Duration.Shift() can increment the date without shifting the time.

```go
package main

import (
	"fmt"
	"time"

	"github.com/senseyeio/duration"
)

func main() {
	loc, _ := time.LoadLocation("America/New_York")
	d, _ := iso8601.ParseISO8601("P1D")
	t1, _ := time.ParseInLocation("Jan 2, 2006 at 3:04pm", "Jan 1, 2006 at 3:04pm", loc)
	t2 := t1
	for i := 0; i < 365; i++ {
		t1 = t1.Add(24 * time.Hour)
		t2 = d.Shift(t2)
		fmt.Printf("time.Add:%d    Duration.Shift:%d\n", t1.Hour(), t2.Hour())
	}
}

// Outputs
// time.Add:15    Duration.Shift:15
// time.Add:15    Duration.Shift:15
// time.Add:15    Duration.Shift:15
// ...
// time.Add:16    Duration.Shift:15
// time.Add:16    Duration.Shift:15
// time.Add:16    Duration.Shift:15
// ...
```

-------
Months are tricky. Shifting by months uses `time.AddDate()`, which is great. However, be aware of how differing days in the month are accommodated. Dates will 'roll over' if the month you're shifting to has fewer days. e.g. if you start on Jan 30th and repeat every "P1M", you'll get this:

```
Jan 30, 2006
Mar 2, 2006
Apr 2, 2006
May 2, 2006
Jun 2, 2006
Jul 2, 2006
Aug 2, 2006
Sep 2, 2006
Oct 2, 2006
Nov 2, 2006
Dec 2, 2006
Jan 2, 2007
```
