package config

import (
	"testing"
)

func testStr(t *testing.T, s1, s2 string) {
	if s1 != s2 {
		t.Fatalf("the value expect %q but %q", s1, s2)
	}
}

func testInt(t *testing.T, i1, i2 int) {
	if i1 != i2 {
		t.Fatalf("the value expect %q but %q", i1, i2)
	}
}

func TestConfig(t *testing.T) {
	testStr(t, Current.Name, "")
	testStr(t, Current.Listen, ":3015")
	testStr(t, Current.DB, "/tmp/pike")
	testStr(t, Current.AdminPath, "")
	testStr(t, Current.AdminToken, "")
	testInt(t, Current.HitForPass, 0)
	testInt(t, len(Current.Directors), 0)
	testInt(t, Current.Concurrency, 0)
}
