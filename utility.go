package main

import "time"

type Timespan time.Duration

func (t Timespan) Format(format string) string {
	z := time.Unix(0, 0).UTC()
	return z.Add(time.Duration(t)).Format(format)
}
