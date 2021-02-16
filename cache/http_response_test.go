package cache

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/golang/snappy"
	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/pike/compress"
)

func TestCloneHeaderAndIgnore(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("Content-Encoding,Content-Length,Connection,Date", strings.Join(ignoreHeaders, ","))
	h := make(http.Header)
	for _, key := range ignoreHeaders {
		h.Add(key, "value")
	}
	newHeader := cloneHeaderAndIgnore(h)
	for _, key := range ignoreHeaders {
		assert.Empty(newHeader.Get(key))
	}
}

func TestHTTPResponseMarshal(t *testing.T) {
	assert := assert.New(t)
	header := make(http.Header)
	header.Add("a", "1")
	header.Add("a", "2")
	header.Add("b", "3")
	resp := &HTTPResponse{
		CompressSrv:               "compress",
		CompressMinLength:         1000,
		CompressContentTypeFilter: regexp.MustCompile(`a|b|c`),
		Header:                    header,
		StatusCode:                200,
		GzipBody:                  []byte("gzip"),
		BrBody:                    []byte("br"),
		RawBody:                   []byte("raw"),
	}
	data, err := resp.Bytes()
	assert.Nil(err)
	assert.Equal(`{"compressSrv":"compress","compressMinLength":1000,"contentTypeFilter":"a|b|c","-":{},"header":{"A":["1","2"],"B":["3"]},"statusCode":200,"gzipBody":"Z3ppcA==","brBody":"YnI=","rawBody":"cmF3"}`, string(data))

	newResp := &HTTPResponse{}
	err = newResp.FromBytes(data)
	assert.Nil(err)

	assert.Equal(resp.CompressSrv, newResp.CompressSrv)
	assert.Equal(resp.CompressMinLength, newResp.CompressMinLength)
	assert.Equal(resp.CompressContentTypeFilter, newResp.CompressContentTypeFilter)
	assert.Equal(resp.Header, newResp.Header)
	assert.Equal(resp.StatusCode, newResp.StatusCode)
	assert.Equal(resp.GzipBody, newResp.GzipBody)
	assert.Equal(resp.BrBody, newResp.BrBody)
	assert.Equal(resp.RawBody, newResp.RawBody)
}

func TestNewHTTPResponse(t *testing.T) {
	assert := assert.New(t)
	data := []byte("Hello world!")
	compressSrv := compress.Get("")
	tests := []struct {
		statusCode int
		header     http.Header
		encoding   string
		fn         func() ([]byte, error)
	}{
		{
			statusCode: 200,
			header:     http.Header{},
			encoding:   compress.EncodingGzip,
			fn: func() ([]byte, error) {
				return compressSrv.Gzip(data)
			},
		},
		{
			statusCode: 201,
			header:     http.Header{},
			encoding:   compress.EncodingBrotli,
			fn: func() ([]byte, error) {
				return compressSrv.Brotli(data)
			},
		},
		{
			statusCode: 200,
			header:     http.Header{},
			encoding:   compress.EncodingSnappy,
			fn: func() ([]byte, error) {
				dst := []byte{}
				dst = snappy.Encode(dst, data)
				return dst, nil
			},
		},
	}

	for _, tt := range tests {
		result, err := tt.fn()
		assert.Nil(err)
		resp, err := NewHTTPResponse(tt.statusCode, tt.header, tt.encoding, result)
		assert.Nil(err)
		assert.Equal(tt.statusCode, resp.StatusCode)
		switch tt.encoding {
		case compress.EncodingGzip:
			assert.NotNil(resp.GzipBody)
			assert.Nil(resp.RawBody)
			assert.Nil(resp.BrBody)
		case compress.EncodingBrotli:
			assert.NotNil(resp.BrBody)
			assert.Nil(resp.GzipBody)
			assert.Nil(resp.RawBody)
		default:
			assert.NotNil(resp.RawBody)
			assert.Nil(resp.GzipBody)
			assert.Nil(resp.BrBody)
		}
	}
}

func TestShouldCompressed(t *testing.T) {
	assert := assert.New(t)
	data := []byte("Hello world!")
	tests := []struct {
		header           http.Header
		rawBody          []byte
		gzipBody         []byte
		brBody           []byte
		shouldCompressed bool
	}{
		{
			shouldCompressed: false,
		},
		{
			header: http.Header{
				elton.HeaderContentType: []string{"image/png"},
			},
			rawBody:          data,
			shouldCompressed: false,
		},
		{
			header: http.Header{
				elton.HeaderContentType: []string{"application/json"},
			},
			rawBody:          data,
			shouldCompressed: true,
		},
		{
			header: http.Header{
				elton.HeaderContentType: []string{"application/json"},
			},
			gzipBody:         data,
			shouldCompressed: true,
		},
		{
			header: http.Header{
				elton.HeaderContentType: []string{"application/json"},
			},
			brBody:           data,
			shouldCompressed: true,
		},
	}
	for _, tt := range tests {
		resp := &HTTPResponse{
			Header:            tt.header,
			RawBody:           tt.rawBody,
			GzipBody:          tt.gzipBody,
			BrBody:            tt.brBody,
			CompressMinLength: 1,
		}
		result := resp.shouldCompressed()
		assert.Equal(tt.shouldCompressed, result)
	}
}

func TestGetRawBody(t *testing.T) {
	assert := assert.New(t)
	compressSrv := compress.Get("")
	data := []byte("Hello world!")
	gzipData, err := compressSrv.Gzip(data)
	assert.Nil(err)
	brData, err := compressSrv.Brotli(data)
	assert.Nil(err)

	tests := []struct {
		rawBody  []byte
		gzipBody []byte
		brBody   []byte
	}{
		{
			rawBody: data,
		},
		{
			gzipBody: gzipData,
		},
		{
			brBody: brData,
		},
	}
	for _, tt := range tests {
		resp := &HTTPResponse{
			RawBody:  tt.rawBody,
			BrBody:   tt.brBody,
			GzipBody: tt.gzipBody,
		}
		rawBody, err := resp.GetRawBody()
		assert.Nil(err)
		assert.Equal(data, rawBody)
	}
}

func TestCompress(t *testing.T) {
	assert := assert.New(t)
	data := []byte("Hello world!")
	compressSrv := compress.Get("")
	gzipData, err := compressSrv.Gzip(data)
	assert.Nil(err)
	brData, err := compressSrv.Brotli(data)
	assert.Nil(err)

	tests := []struct {
		rawBody  []byte
		gzipBody []byte
		brBody   []byte
	}{
		{
			rawBody: data,
		},
		{
			gzipBody: gzipData,
			brBody:   brData,
		},
	}
	for _, tt := range tests {
		resp := &HTTPResponse{
			Header: http.Header{
				elton.HeaderContentType: []string{"application/json"},
			},
			RawBody:           tt.rawBody,
			GzipBody:          tt.gzipBody,
			BrBody:            tt.brBody,
			CompressMinLength: 1,
		}
		err := resp.Compress()
		assert.Nil(err)
		assert.Nil(resp.RawBody)
		assert.Equal(gzipData, resp.GzipBody)
		assert.Equal(brData, resp.BrBody)
	}
}

func TestGetBodyByAcceptEncoding(t *testing.T) {
	assert := assert.New(t)
	data := []byte("Hello world!")
	compressSrv := compress.Get("")
	gzipData, err := compressSrv.Gzip(data)
	assert.Nil(err)
	brData, err := compressSrv.Brotli(data)
	assert.Nil(err)

	tests := []struct {
		rawBody        []byte
		gzipBody       []byte
		brBody         []byte
		acceptEncoding string
		minLength      int
		resultEncoding string
		result         []byte
	}{
		// 支持br且已存在br
		{
			brBody:         brData,
			acceptEncoding: compress.EncodingBrotli,
			resultEncoding: compress.EncodingBrotli,
			result:         brData,
		},
		// 支持gzip且已存在gzip
		{
			gzipBody:       gzipData,
			acceptEncoding: compress.EncodingGzip,
			resultEncoding: compress.EncodingGzip,
			result:         gzipData,
		},
		// 数据不应该被压缩，没有gzip，而且原始数据小于最小压缩长度
		{
			rawBody:        data,
			minLength:      1000,
			acceptEncoding: compress.EncodingGzip,
			resultEncoding: "",
			result:         data,
		},
		// 支持br但没有br，且数据不应该压缩
		{
			rawBody:        data,
			minLength:      1000,
			acceptEncoding: compress.EncodingBrotli,
			resultEncoding: "",
			result:         data,
		},
		// 支持br，而且原始数据大于最小压缩长度，压缩后返回
		{
			rawBody:        data,
			acceptEncoding: compress.EncodingBrotli,
			resultEncoding: compress.EncodingBrotli,
			result:         brData,
		},
		// 支持gzip，而且压缩数据大于最小压缩长度，压缩后返回
		{
			rawBody:        data,
			acceptEncoding: compress.EncodingGzip,
			resultEncoding: compress.EncodingGzip,
			result:         gzipData,
		},
		// 不支持压缩，从br中返回
		{
			brBody:         brData,
			acceptEncoding: "",
			resultEncoding: "",
			result:         data,
		},
		// 不支持压缩，从gzip中返回
		{
			gzipBody:       gzipData,
			acceptEncoding: "",
			resultEncoding: "",
			result:         data,
		},
	}
	for _, tt := range tests {
		resp := &HTTPResponse{
			Header: http.Header{
				elton.HeaderContentType: []string{"application/json"},
			},
			RawBody:           tt.rawBody,
			GzipBody:          tt.gzipBody,
			BrBody:            tt.brBody,
			CompressMinLength: tt.minLength,
		}
		encoding, body, err := resp.getBodyByAcceptEncoding(tt.acceptEncoding)
		assert.Nil(err)
		assert.Equal(tt.resultEncoding, encoding)
		assert.Equal(tt.result, body)
	}
}

func TestFill(t *testing.T) {
	assert := assert.New(t)
	data := []byte("Hello world!")
	compressSrv := compress.Get("")
	gzipData, err := compressSrv.Gzip(data)
	assert.Nil(err)
	brData, err := compressSrv.Brotli(data)
	assert.Nil(err)

	tests := []struct {
		rawBody        []byte
		acceptEncoding string
		resultEncoding string
		result         []byte
	}{
		{
			rawBody:        data,
			acceptEncoding: compress.EncodingBrotli,
			resultEncoding: compress.EncodingBrotli,
			result:         brData,
		},
		{
			rawBody:        data,
			acceptEncoding: compress.EncodingGzip,
			resultEncoding: compress.EncodingGzip,
			result:         gzipData,
		},
		{
			rawBody:        data,
			acceptEncoding: "",
			resultEncoding: "",
			result:         data,
		},
	}
	for _, tt := range tests {
		resp := &HTTPResponse{
			Header: http.Header{
				elton.HeaderContentType: []string{"application/json"},
			},
			RawBody:    tt.rawBody,
			StatusCode: 200,
		}
		req := httptest.NewRequest("GET", "/", nil)
		c := elton.NewContext(httptest.NewRecorder(), req)
		c.SetRequestHeader(elton.HeaderAcceptEncoding, tt.acceptEncoding)
		err := resp.Fill(c)
		assert.Nil(err)
		assert.Equal(200, c.StatusCode)
		assert.Equal(tt.resultEncoding, c.GetHeader(elton.HeaderContentEncoding))
		assert.Equal(tt.result, c.BodyBuffer.Bytes())
	}
}
