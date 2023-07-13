package test

import (
	"fmt"
	"time"
	"testing"
	"github.com/glupi-borna/soko/internal/format"
)

func TestFormat(t *testing.T) {
	var test_sizes = []float64{1, 512, 1023}
	for _, val := range test_sizes {
		AssertEq(format.Bytes(val), fmt.Sprint(int(val)) + "b", t)
		AssertEq(format.Bytes(val * 1024), fmt.Sprint(int(val)) + "K", t)
		AssertEq(format.Bytes(val * 1024 * 1024), fmt.Sprint(int(val)) + "M", t)
		AssertEq(format.Bytes(val * 1024 * 1024 * 1024), fmt.Sprint(int(val)) + "G", t)
	}

	var test_durations = []time.Duration{1, 5, 10, 20, 30}
	for _, val := range test_durations {
		AssertEq(format.Duration(val), fmt.Sprint(int(val)) + "ns", t)
		AssertEq(format.Duration(val * 1000), fmt.Sprint(int(val)) + "us", t)
		AssertEq(format.Duration(val * 1000 * 1000), fmt.Sprint(int(val)) + "ms", t)
		AssertEq(format.Duration(val * 1000 * 1000 * 1000), fmt.Sprint(int(val)) + "s", t)
		AssertEq(format.Duration(val * 1000 * 1000 * 1000 * 60), fmt.Sprint(int(val)) + "m", t)
		AssertEq(format.Duration(val * 1000 * 1000 * 1000 * 60 * 60), fmt.Sprint(int(val)) + "h", t)
	}
}
