package test

import (
	"testing"
	"fmt"
)

func AssertEq[K comparable](v1, v2 K, t *testing.T) {
	eq := v1 == v2
	if !eq {
		t.Fatal("Expected", fmt.Sprint(v1), "to equal", fmt.Sprint(v2))
	}
}
