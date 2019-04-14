package util

import "testing"

func TestGetStack(t *testing.T) {
	stack := GetStack(0, 5)
	if len(stack) == 0 {
		t.Fatalf("get stack fail")
	}
}
