package server

import "testing"

func TestBlockIP(t *testing.T) {
	b := &BlockIP{}
	ip := "4.4.4.4"
	b.Add(ip)
	if b.FindIndex(ip) == -1 {
		t.Fatalf("the ip should be in block list")
	}
	b.Remove(ip)
	if b.FindIndex(ip) != -1 {
		t.Fatalf("the ip shouldn't be in block list")
	}
}
