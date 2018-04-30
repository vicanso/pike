package httplog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo"
)

func TestParse(t *testing.T) {
	tags := Parse([]byte("Pike {host}{method} {path} {proto} {query} {remote} {client-ip} {scheme} {uri} {~jt} {>X-Request-Id} {<X-Response-Id} {when} {when-iso} {when-iso-ms} {when-unix} {status} {size} {referer} {userAgent} {latency} {latency-ms}ms"))
	count := 44
	if len(tags) != count {
		t.Fatalf("the tags length expect %v but %v", count, len(tags))
	}
	startedAt := time.Now()

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "http://aslant.site:5000/users/login?cache-control=no-cache", nil)
	resp := httptest.NewRecorder()
	c := e.NewContext(req, resp)
	c.Request().Header.Set("Referer", "http://pike.aslant.site/")
	c.Request().Header.Set("User-Agent", "pike/client")
	cookie := &http.Cookie{
		Name:  "jt",
		Value: "cookieValue",
	}
	c.SetCookie(cookie)
	c.Request().Header.Set("X-Request-Id", "requestId")
	c.Response().Header().Set("X-Response-Id", "responseId")
	c.Response().Write([]byte("hello world"))

	str := Format(c, tags, startedAt)
	fmt.Println(str)
	if strings.Index(str, "{") != -1 {
		t.Fatalf("the log of request fail")
	}
	tags = Parse([]byte(""))
	if len(tags) != 0 {
		t.Fatalf("the empty log format should be null")
	}
}

func TestFileWrite(t *testing.T) {
	now := time.Now()
	date := now.Format("2006-01-02")
	path := "/tmp"
	file := path + "/" + date
	os.Remove(file)
	fileWriter := &FileWriter{
		Path:     path,
		Category: Date,
	}
	message := []byte("ABCD")
	fileWriter.Write(message)
	fileWriter.Close()
	buf, _ := ioutil.ReadFile(file)
	if string(buf) != string(message)+"\n" {
		t.Fatalf("file write data fail")
	}

	os.Remove(file)
	fileWriter = &FileWriter{
		Path: file,
	}
	fileWriter.Write(message)
	fileWriter.Close()
	buf, _ = ioutil.ReadFile(file)
	if string(buf) != string(message)+"\n" {
		t.Fatalf("file write data fail")
	}
}

func TestUDPWrite(t *testing.T) {
	message := []byte("ABCD")
	udpWriter := &UDPWriter{
		URI: "127.0.0.1:7000",
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		udpWriter.Write(message)
		udpWriter.Close()
	}()
	addr, _ := net.ResolveUDPAddr("udp", ":7000")
	conn, _ := net.ListenUDP("udp", addr)
	buf := make([]byte, 1024)
	n, _, _ := conn.ReadFromUDP(buf)
	conn.Close()
	if bytes.Equal(buf[0:n], message) == false {
		t.Fatalf("udp write data fail")
	}
}
