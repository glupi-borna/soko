//go:build !debug
package debug

var DEBUG = false

func Log(args ...any) {}
func Assert(val bool, msg string) {}
