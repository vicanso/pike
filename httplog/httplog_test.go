package httplog

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestParse(t *testing.T) {
	tags := Parse([]byte("Pike {host}{method} {path} {proto} {query} {remote} {client-ip} {scheme} {uri} {~jt} {>X-Request-Id} {<X-Response-Id} {when} {when-iso} {when-iso-ms} {when-unix} {status} {payload-size} {size} {referer} {userAgent} {latency} {latency-ms}ms"))
	count := 46
	if len(tags) != count {
		t.Fatalf("the tags length expect %v but %v", count, len(tags))
	}
	startedAt := time.Now()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://aslant.site:5000/users/login?cache-control=no-cache")
	ctx.Request.Header.Set("Referer", "http://pike.aslant.site/")
	ctx.Request.Header.Set("User-Agent", "fasthttp/client")
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetBody([]byte("{\"name\": \"vicanso\"}"))
	ctx.Request.Header.SetCookie("jt", "cookieValue")
	ctx.Request.Header.SetCanonical([]byte("X-Request-Id"), []byte("requestId"))
	ctx.Response.Header.SetCanonical([]byte("X-Response-Id"), []byte("responseId"))
	ctx.SetBody([]byte("hello world"))
	buf := Format(ctx, tags, startedAt)
	log.Print(string(buf))
	if bytes.Index(buf, []byte("{")) != -1 {
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
