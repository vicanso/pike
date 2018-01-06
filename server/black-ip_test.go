package server

import "testing"

func TestBlackIP(t *testing.T) {
	b := &BlackIP{}
	ip := "4.4.4.4"
	b.Add(ip)
	if b.FindIndex(ip) == -1 {
		t.Fatalf("the ip should be in black list")
	}
	b.Remove(ip)
	if b.FindIndex(ip) != -1 {
		t.Fatalf("the ip shouldn't be in black list")
	}
}
