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
	conf := &Config{}
	testStr(t, conf.Name, "")
	testStr(t, conf.Listen, "")
	testStr(t, conf.DB, "")
	testStr(t, conf.AdminPath, "")
	testStr(t, conf.AdminToken, "")
	testInt(t, conf.HitForPass, 0)
	testInt(t, len(conf.Directors), 0)
	testInt(t, conf.Concurrency, 0)
}
