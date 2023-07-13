package format

import (
	"strconv"
	"math"
	"time"
)

var size_units = []string{"b", "K", "M", "G", "T", "P", "E", "Z", "Y"}
var second_units = []string{"ns", "us", "ms", "s"}
var log1024 = math.Log(1024)

func Bytes(bytes float64) string {
	if bytes == 0 { return "0" + size_units[0] }
	order := math.Log(bytes) / log1024
	return strconv.FormatFloat(bytes / math.Pow(1024, math.Floor(order)), 'f', 0, 64) + size_units[int(order)]
}

func Duration(d time.Duration) string {
	if d < 1000 {
		return strconv.Itoa(int(d)) + "ns"
	} else if d < 1000 * 1000 {
		return strconv.Itoa(int(d / 1000)) + "us"
	} else if d < 1000 * 1000 * 1000 {
		return strconv.Itoa(int(d / (1000*1000))) + "ms"
	} else {
		seconds := d / (1000 * 1000 * 1000)
		if seconds < 60 {
			return strconv.Itoa(int(seconds)) + "s"
		} else if seconds < 60 * 60 {
			return strconv.Itoa(int(seconds / 60)) + "m"
		} else {
			return strconv.Itoa(int(seconds / (60*60))) + "h"
		}
	}
	return d.String()
}

var WidgetVars = map[string]any {
	"Bytes": Bytes,
	"Duration": Duration,
}
