package custommiddleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mitchellh/go-server-timing"

	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
)

func TestShouldCompress(t *testing.T) {
	compressTypes := []string{
		"text",
		"javascript",
		"json",
	}
	t.Run("should compress", func(t *testing.T) {
		if !shouldCompress(compressTypes, "json") {
			t.Fatalf("json should be compress")
		}
		if shouldCompress(compressTypes, "image/png") {
			t.Fatalf("image png should not be compress")
		}
	})
}

func TestSave(t *testing.T) {
	client := &cache.Client{
		Path: "/tmp/test.cache",
	}
	err := client.Init()
	if err != nil {
		t.Fatalf("cache init fail, %v", err)
	}
	defer client.Close()
	identity := []byte("save-test")
	t.Run("save no content", func(t *testing.T) {
		resp := &cache.Response{
			TTL:        30,
			StatusCode: http.StatusNoContent,
		}
		save(client, identity, resp, true)
		result, err := client.GetResponse(identity)
		if err != nil || result.TTL != resp.TTL || result.StatusCode != resp.StatusCode {
			t.Fatalf("save no content fail")
		}
	})

	t.Run("save gzip content", func(t *testing.T) {
		data := []byte("data")
		gzipData, _ := util.Gzip(data, 0)
		resp := &cache.Response{
			TTL:               30,
			StatusCode:        http.StatusOK,
			GzipBody:          gzipData,
			CompressMinLength: 1,
		}
		save(client, identity, resp, true)
		result, err := client.GetResponse(identity)
		if err != nil || result.TTL != resp.TTL || result.StatusCode != resp.StatusCode {
			t.Fatalf("save gzip content fail")
		}
		if len(result.Body) != 0 {
			t.Fatalf("raw content should be nil")
		}
		if !bytes.Equal(result.GzipBody, gzipData) {
			t.Fatalf("save gzip content is not equal original data")
		}
		if len(result.BrBody) == 0 {
			t.Fatalf("should save br data")
		}
	})
	t.Run("save br content", func(t *testing.T) {
		data := []byte("data")
		brData, _ := util.BrotliEncode(data, 0)
		resp := &cache.Response{
			TTL:               30,
			StatusCode:        http.StatusOK,
			BrBody:            brData,
			CompressMinLength: 1,
		}
		save(client, identity, resp, true)
		result, err := client.GetResponse(identity)
		if err != nil || result.TTL != resp.TTL || result.StatusCode != resp.StatusCode {
			t.Fatalf("save br content fail")
		}
		if len(result.Body) != 0 {
			t.Fatalf("raw content should be nil")
		}
		if !bytes.Equal(result.BrBody, brData) {
			t.Fatalf("save br content is not equal original data")
		}
		if len(result.GzipBody) == 0 {
			t.Fatalf("should save gzip data")
		}
	})

	t.Run("save content smaller than compress min length", func(t *testing.T) {
		data := []byte("需要一个很大的数据，如果没有，那就设置小的compressMinLength")
		resp := &cache.Response{
			TTL:        30,
			StatusCode: http.StatusOK,
			Body:       data,
		}
		save(client, identity, resp, true)
		result, err := client.GetResponse(identity)
		if err != nil {
			t.Fatalf("save samll content fail, %v", err)
		}
		if len(result.Body) == 0 {
			t.Fatalf("the body of small content response shoul not be nil")
		}
		if len(result.GzipBody) != 0 {
			t.Fatalf("the gzip body of small content response shoul be nil")
		}
	})

	t.Run("save content bigger than compress min length", func(t *testing.T) {
		data := []byte("需要一个很大的数据，如果没有，那就设置小的compressMinLength")
		resp := &cache.Response{
			TTL:               30,
			StatusCode:        http.StatusOK,
			Body:              data,
			CompressMinLength: 1,
		}
		save(client, identity, resp, true)
		result, err := client.GetResponse(identity)
		if err != nil || result.TTL != resp.TTL || result.StatusCode != resp.StatusCode {
			t.Fatalf("save big content fail")
		}
		gzipData := result.GzipBody
		if len(gzipData) == 0 {
			t.Fatalf("big cotent response should be gzip")
		}
		raw, _ := util.Gunzip(gzipData)
		if !bytes.Equal(raw, data) {
			t.Fatalf("big cotent response gzip fail")
		}

		brData := result.BrBody
		if len(brData) == 0 {
			t.Fatalf("big cotent response should be brotli")
		}
		raw, _ = util.BrotliDecode(brData)
		if !bytes.Equal(raw, data) {
			t.Fatalf("big cotent response brotli fail")
		}

	})
}

func TestDispatcher(t *testing.T) {
	client := &cache.Client{
		Path: "/tmp/test.cache",
	}
	err := client.Init()
	if err != nil {
		t.Fatalf("cache init fail, %v", err)
	}
	defer client.Close()
	conf := DispatcherConfig{}
	t.Run("dispatch response", func(t *testing.T) {
		fn := Dispatcher(conf, client)(func(c echo.Context) error {
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.POST, "/users/me", nil)
		resp := &httptest.ResponseRecorder{
			Body: new(bytes.Buffer),
		}
		c := e.NewContext(req, resp)
		timing := &servertiming.Header{}
		timing.NewMetric(vars.PikeMetric)
		c.Set(vars.Timing, timing)
		c.Set(vars.Identity, []byte("abc"))
		c.Set(vars.Status, cache.Fetching)
		c.Set(vars.RID, "a")
		cr := &cache.Response{
			CreatedAt:  uint32(time.Now().Unix()),
			TTL:        300,
			StatusCode: 200,
			Body:       []byte("ABCD"),
		}
		c.Set(vars.Response, cr)
		fn(c)
		if resp.Code != 200 {
			t.Fatalf("the response code should be 200")
		}
		if string(resp.Body.Bytes()) != "ABCD" {
			t.Fatalf("the response body should be ABCD")
		}
		// 由于缓存的数据需要写数据库，因此需要延时关闭client
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("dispatch cacheable data", func(t *testing.T) {
		identity := []byte("abc")
		cr := &cache.Response{
			CreatedAt:  uint32(time.Now().Unix()),
			TTL:        300,
			StatusCode: 200,
			Body:       []byte("ABCD"),
		}
		fn := Dispatcher(conf, client)(func(c echo.Context) error {
			return nil
		})
		req := httptest.NewRequest(echo.POST, "/users/me", nil)
		resp := &httptest.ResponseRecorder{
			Body: new(bytes.Buffer),
		}
		e := echo.New()
		c := e.NewContext(req, resp)
		c.Set(vars.Identity, identity)
		c.Set(vars.Status, cache.Cacheable)
		c.Set(vars.Response, cr)
		c.Set(vars.RID, "a")
		fn(c)
		if !bytes.Equal(resp.Body.Bytes(), cr.Body) {
			t.Fatalf("dispatch cacheable data fail")
		}
	})

	t.Run("dispatch not modified", func(t *testing.T) {
		identity := []byte("abc")
		cr := &cache.Response{
			CreatedAt:  uint32(time.Now().Unix()),
			TTL:        300,
			StatusCode: 200,
			Body:       []byte("ABCD"),
		}
		fn := Dispatcher(conf, client)(func(c echo.Context) error {
			return nil
		})
		req := httptest.NewRequest(echo.GET, "/users/me", nil)
		resp := &httptest.ResponseRecorder{
			Body: new(bytes.Buffer),
		}
		e := echo.New()
		c := e.NewContext(req, resp)
		c.Set(vars.Fresh, true)
		c.Set(vars.Identity, identity)
		c.Set(vars.Status, cache.Cacheable)
		c.Set(vars.Response, cr)
		c.Set(vars.RID, "a")
		fn(c)
		if resp.Code != http.StatusNotModified {
			t.Fatalf("dispatch not modified fail")
		}
	})
}
