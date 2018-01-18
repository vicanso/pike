package director

import (
	"strconv"
	"testing"
)

func testVar(t *testing.T, s1, s2 string) {
	if s1 != s2 {
		t.Fatalf("the value expect %q but %q", s1, s2)
	}
}

func testMatch(t *testing.T, d *Director, host, uri []byte, resultExpected bool) {
	result := d.Match(host, uri)
	if result != resultExpected {
		t.Fatalf("unexpected result in Pass %q %q: %v. Expecting %v", host, uri, result, resultExpected)
	}
}

func TestCreateDirector(t *testing.T) {
	name := "test"
	policy := "random"
	ping := "/ping"
	pass := "cache-control=no-cache"
	c := &Config{
		Name: name,
		Type: policy,
		Ping: ping,
		Pass: []string{
			pass,
		},
		Prefix: []string{
			"/tiny",
			"/albi",
		},
		Host: []string{
			"www.aslant.site",
			"aslant.site",
			"~npmtrend",
		},
		Backends: []string{
			"host:5001",
			"host:5002",
		},
	}
	dList := GetDirectors([]*Config{c})
	d := dList[0]
	testVar(t, d.Name, name)
	testVar(t, d.Policy, policy)
	testVar(t, d.Ping, ping)
	testVar(t, strconv.Itoa(d.Priority), "2")
	testVar(t, strconv.Itoa(len(d.Prefixs)), "2")
	testVar(t, string(d.Prefixs[0]), "/tiny")
	testVar(t, string(d.Prefixs[1]), "/albi")
	testVar(t, strconv.Itoa(len(d.Hosts)), "3")
	testVar(t, string(d.Hosts[0]), "www.aslant.site")
	testVar(t, string(d.Hosts[1]), "aslant.site")
	testVar(t, strconv.Itoa(len(d.Passes)), "1")
	testVar(t, string(d.Passes[0]), pass)

	testMatch(t, d, []byte("aslant.site"), []byte("/tiny"), true)
	testMatch(t, d, []byte("www.aslant.site"), []byte("/albi"), true)
	testMatch(t, d, []byte("www.npmtrend.com"), []byte("/albi"), true)
	testMatch(t, d, []byte("dcharts.com"), []byte("/albi"), false)
	testMatch(t, d, []byte("aslant.site"), []byte("/abc"), false)

	testDirector := CreateDirector(&Config{
		Name: "TEST",
	})
	testMatch(t, testDirector, []byte("aslant.site"), []byte("/tiny"), true)
}
