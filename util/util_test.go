package util

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCheckAndGetValueFromEnv(t *testing.T) {
	key := "_pike"
	value := "1"
	result := CheckAndGetValueFromEnv("$" + key)
	if result != "" {
		t.Fatalf("get not exists key from env should return nil")
	}
	os.Setenv(key, value)
	result = CheckAndGetValueFromEnv("$" + key)
	if result != value {
		t.Fatalf("get value from env fail")
	}
}

func TestGetIdentity(t *testing.T) {
	req := &http.Request{
		Method:     "GET",
		Host:       "aslant.site",
		RequestURI: "/users/me",
	}
	id := GetIdentity(req)
	if len(id) != 25 || string(id) != "GET aslant.site /users/me" {
		t.Fatalf("get identity fail")
	}

	req = &http.Request{
		Method:     "GET",
		Host:       "aslant.site",
		RequestURI: "/中文",
	}
	id = GetIdentity(req)
	if len(id) != 23 || string(id) != "GET aslant.site /中文" {
		t.Fatalf("get identity(include chinese) fail")
	}
}

func TestGenerateGetIdentity(t *testing.T) {
	c := &http.Cookie{
		Name:  "jt",
		Value: "HJxX4OOoX7",
	}
	fn := GenerateGetIdentity("host method path proto scheme uri userAgent query ~jt >X-Token ?id")
	req := httptest.NewRequest(http.MethodGet, "/users/me?cache=no-cache&id=1", nil)
	req.Header.Set("User-Agent", "golang-http")
	req.Header.Set("X-Token", "ABCD")
	req.AddCookie(c)
	buf := fn(req)
	expectID := "example.com GET /users/me HTTP/1.1 HTTP /users/me?cache=no-cache&id=1 golang-http cache=no-cache&id=1 HJxX4OOoX7 ABCD 1"
	if string(buf) != expectID {
		t.Fatalf("get identity fail")
	}
}

func TestContainString(t *testing.T) {
	if !ContainString([]string{
		"1",
		"2",
	}, "2") {
		t.Fatalf("contain string fail")
	}
	if ContainString([]string{
		"1",
		"2",
	}, "3") {
		t.Fatalf("contain string fail")
	}
}

func TestConvertToHTTPHeader(t *testing.T) {
	if ConvertToHTTPHeader(nil) != nil ||
		ConvertToHTTPHeader([]string{}) != nil {
		t.Fatalf("convert nil string array should return nil")
	}
	h := ConvertToHTTPHeader([]string{
		"X-Token:1",
		"X-Request-ID:1234",
	})
	if len(h) != 2 ||
		h.Get("X-Token") != "1" ||
		h.Get("X-Request-ID") != "1234" {
		t.Fatalf("convert to http header fail")
	}
}

func BenchmarkGetIdentity(b *testing.B) {
	b.ReportAllocs()
	req := httptest.NewRequest(http.MethodGet, "/users/me?cache=no-cache&id=1", nil)
	for i := 0; i < b.N; i++ {
		_ = GetIdentity(req)
	}
}
