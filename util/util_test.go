package util

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAndGetValueFromEnv(t *testing.T) {
	assert := assert.New(t)
	key := "_pike"
	value := "1"
	result := CheckAndGetValueFromEnv("$" + key)
	assert.Equal(result, "", "get not exists key from env should return nil")
	os.Setenv(key, value)
	result = CheckAndGetValueFromEnv("$" + key)
	assert.Equal(result, value)
}

func TestGetIdentity(t *testing.T) {
	assert := assert.New(t)
	req := &http.Request{
		Method:     "GET",
		Host:       "aslant.site",
		RequestURI: "/users/me",
	}
	id := GetIdentity(req)
	assert.Equal(len(id), 25)
	assert.Equal(string(id), "GET aslant.site /users/me")

	req = &http.Request{
		Method:     "GET",
		Host:       "aslant.site",
		RequestURI: "/中文",
	}
	id = GetIdentity(req)
	assert.Equal(len(id), 23)
	assert.Equal(string(id), "GET aslant.site /中文")
}

func TestGenerateGetIdentity(t *testing.T) {
	assert := assert.New(t)
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
	assert.Equal(string(buf), expectID)
}

func TestContainString(t *testing.T) {
	assert := assert.New(t)
	assert.True(ContainString([]string{
		"1",
		"2",
	}, "2"))
	assert.False(ContainString([]string{
		"1",
		"2",
	}, "3"))
}

func TestConvertToHTTPHeader(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(ConvertToHTTPHeader(nil))
	assert.Nil(ConvertToHTTPHeader([]string{}))
	h := ConvertToHTTPHeader([]string{
		"X-Token:1",
		"X-Request-ID:1234",
	})
	assert.Equal(len(h), 2)
	assert.Equal(h.Get("X-Token"), "1")
	assert.Equal(h.Get("X-Request-ID"), "1234")
}

func BenchmarkGetIdentity(b *testing.B) {
	b.ReportAllocs()
	req := httptest.NewRequest(http.MethodGet, "/users/me?cache=no-cache&id=1", nil)
	for i := 0; i < b.N; i++ {
		_ = GetIdentity(req)
	}
}
