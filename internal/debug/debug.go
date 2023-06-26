//go:build debug
package debug

import "fmt"

var DEBUG = true

func Log(args ...any) {
	fmt.Println(args...)
}

func Assert(val bool, msg string) {
	if !val { panic("Assertion failed: " + msg) }
}
