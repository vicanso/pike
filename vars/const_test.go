package vars

import (
	"strconv"
	"testing"
)

func testVar(t *testing.T, s1, s2 string) {
	if s1 != s2 {
		t.Fatalf("the value expect %q but %q", s1, s2)
	}
}

func TestVars(t *testing.T) {
	testVar(t, string(AcceptEncoding), "Accept-Encoding")
	testVar(t, string(ContentEncoding), "Content-Encoding")
	testVar(t, string(ContentLength), "Content-Length")
	testVar(t, string(Server), "Server")
	testVar(t, string(XForwardedFor), "X-Forwarded-For")
	testVar(t, string(Age), "Age")
	testVar(t, string(CacheControl), "Cache-Control")
	testVar(t, string(LineBreak), "\n")
	testVar(t, string(Colon), ":")
	testVar(t, string(Space), " ")

	testVar(t, string(ServerName), "Pike")
	testVar(t, string(Gzip), "gzip")
	testVar(t, string(Br), "br")
	testVar(t, string(Get), "GET")
	testVar(t, string(Head), "HEAD")

	testVar(t, ErrDirectorUnavailable.Error(), "director unavailable")
	testVar(t, ErrServiceUnavailable.Error(), "service unavailable")
	testVar(t, ErrDbNotInit.Error(), "db isn't init")

	testVar(t, strconv.Itoa(Pass), "0")
	testVar(t, strconv.Itoa(Fetching), "1")
	testVar(t, strconv.Itoa(Waiting), "2")
	testVar(t, strconv.Itoa(HitForPass), "3")
	testVar(t, strconv.Itoa(Cacheable), "4")

	testVar(t, strconv.Itoa(CompressMinLength), "1024")
	testVar(t, Random, "random")
	testVar(t, RoundRobin, "roundRobin")
	testVar(t, LeastConn, "leastConn")
	testVar(t, IPHash, "ipHash")
	testVar(t, URIHash, "uriHash")
	testVar(t, First, "first")
	testVar(t, Header, "header")
}
