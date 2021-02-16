// MIT License

// Copyright (c) 2020 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// HTTP响应数据，只用于根据客户端支持编码以及最小压缩长度返回对应的数据

// 对于可缓存且大于最小缓存则可使用compress方法保存gzip与br两种缓存数据，
// 在客户端请求时根据客户端支持的编码返回，若不支持压缩，则从解压获取原始数据返回
// 对于不可缓存数据，根据客户端支持的编码以及数据长度返回对应数据

package cache

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/vicanso/elton"
	"github.com/vicanso/pike/compress"
)

var ignoreHeaders = []string{
	"Content-Encoding",
	"Content-Length",
	"Connection",
	"Date",
}

var ErrBodyIsNil = errors.New("body is nil")

var defaultCompressContentTypeFilter = regexp.MustCompile(`text|javascript|json|wasm|xml|font`)

type (
	// HTTPResponse http response's cache
	HTTPResponse struct {
		// 压缩服务名称
		CompressSrv string `json:"compressSrv,omitempty"`
		// 压缩最小尺寸
		CompressMinLength int    `json:"compressMinLength,omitempty"`
		ContentTypeFilter string `json:"contentTypeFilter,omitempty"`
		// 压缩数据类型
		CompressContentTypeFilter *regexp.Regexp `json:"-,omitempty"`
		// 响应头
		Header http.Header `json:"header,omitempty"`
		// 响应状态码
		StatusCode int    `json:"statusCode,omitempty"`
		GzipBody   []byte `json:"gzipBody,omitempty"`
		BrBody     []byte `json:"brBody,omitempty"`
		RawBody    []byte `json:"rawBody,omitempty"`
	}
)

func cloneHeaderAndIgnore(header http.Header) http.Header {
	h := header.Clone()
	for _, key := range ignoreHeaders {
		h.Del(key)
	}
	return h
}

// NewHTTPResponse new a http response
func NewHTTPResponse(statusCode int, header http.Header, encoding string, data []byte) (*HTTPResponse, error) {
	resp := &HTTPResponse{
		StatusCode: statusCode,
		Header:     cloneHeaderAndIgnore(header),
	}
	switch encoding {
	case compress.EncodingGzip:
		resp.GzipBody = data
	case compress.EncodingBrotli:
		resp.BrBody = data
	case "":
		resp.RawBody = data
	default:
		// 取默认的compress来解压
		compressSrv := compress.Get("")
		data, err := compressSrv.Decompress(encoding, data)
		if err != nil {
			return nil, err
		}
		header.Del(elton.HeaderContentEncoding)
		resp.RawBody = data
	}
	return resp, nil
}

// Bytes http response to bytes
func (resp *HTTPResponse) Bytes() ([]byte, error) {
	if resp.CompressContentTypeFilter != nil {
		resp.ContentTypeFilter = resp.CompressContentTypeFilter.String()
	}
	return json.Marshal(resp)
}

// FromBytes http response from bytes
func (resp *HTTPResponse) FromBytes(data []byte) (err error) {
	err = json.Unmarshal(data, resp)
	if err != nil {
		return
	}
	if resp.ContentTypeFilter != "" {
		resp.CompressContentTypeFilter, err = regexp.Compile(resp.ContentTypeFilter)
		if err != nil {
			return
		}
	}
	return
}

func (resp *HTTPResponse) shouldCompressed() bool {
	// 如果数据都小于最小压缩长度，则表示无需压缩
	if len(resp.RawBody) <= resp.CompressMinLength &&
		len(resp.GzipBody) <= resp.CompressMinLength &&
		len(resp.BrBody) <= resp.CompressMinLength {
		return false
	}
	filter := resp.CompressContentTypeFilter
	if filter == nil {
		filter = defaultCompressContentTypeFilter
	}
	// 数据类型匹配才可压缩
	return filter.MatchString(resp.Header.Get(elton.HeaderContentType))
}

// GetRawBody get raw body of http response(not compress)
func (resp *HTTPResponse) GetRawBody() (rawBody []byte, err error) {
	rawBody = resp.RawBody
	if len(rawBody) != 0 {
		return
	}
	compressSrv := compress.Get("")
	// 原始数据为空，需要从gzip或br中解压
	if len(resp.GzipBody) != 0 {
		return compressSrv.Gunzip(resp.GzipBody)
	}
	if len(resp.BrBody) != 0 {
		return compressSrv.BrotliDecode(resp.BrBody)
	}
	return
}

// Compress compress http response's data
func (resp *HTTPResponse) Compress() (err error) {
	// 如果数据不需要压缩，则直接返回
	if !resp.shouldCompressed() {
		return
	}
	// 如果gzip与br均已压缩
	if len(resp.GzipBody) != 0 && len(resp.BrBody) != 0 {
		return
	}
	rawBody, err := resp.GetRawBody()
	if err != nil {
		return
	}
	// 如果原始数据为空，则直接报错，因为如果数据为空，则在前置判断是否可压缩已返回
	if len(rawBody) == 0 {
		err = ErrBodyIsNil
		return
	}
	compressSrv := compress.Get(resp.CompressSrv)
	if len(resp.GzipBody) == 0 {
		resp.GzipBody, err = compressSrv.Gzip(rawBody)
		if err != nil {
			return
		}
	}
	if len(resp.BrBody) == 0 {
		resp.BrBody, err = compressSrv.Brotli(rawBody)
		if err != nil {
			return
		}
	}
	// 压缩后清空原始数据，因为基本所有的客户端都支持gzip，
	// 没必要再保存原始数据，如果有需要，可以从gzip中解压
	resp.RawBody = nil
	return
}

func (resp *HTTPResponse) getBodyByAcceptEncoding(acceptEncoding string) (encoding string, body []byte, err error) {
	compressSrv := compress.Get(resp.CompressSrv)

	// 如果支持br，而且br有数据
	acceptBr := strings.Contains(acceptEncoding, compress.EncodingBrotli)
	if acceptBr && len(resp.BrBody) != 0 {
		return compress.EncodingBrotli, resp.BrBody, nil
	}

	// 如果支持gzip，而且gzip有数据
	acceptGzip := strings.Contains(acceptEncoding, compress.EncodingGzip)
	if acceptGzip && len(resp.GzipBody) != 0 {
		return compress.EncodingGzip, resp.GzipBody, nil
	}

	// 获取原始数据压缩
	rawBody, err := resp.GetRawBody()
	if err != nil {
		return "", nil, err
	}
	shouldCompressed := resp.shouldCompressed()
	// 数据不应该压缩，直接返回
	if !shouldCompressed {
		return "", rawBody, nil
	}

	// 支持br，数据从原始数据压缩
	if acceptBr {
		brBody, err := compressSrv.Brotli(rawBody)
		if err != nil {
			return "", nil, err
		}
		return compress.EncodingBrotli, brBody, nil
	}

	// 支持gzip，数据从原始数据压缩
	if acceptGzip {
		gzipBody, err := compressSrv.Gzip(rawBody)
		if err != nil {
			return "", nil, err
		}
		return compress.EncodingGzip, gzipBody, nil
	}

	// 都不支持，返回原始数据
	return "", rawBody, nil
}

// Fill fill response to context
func (resp *HTTPResponse) Fill(c *elton.Context) (err error) {
	encoding, body, err := resp.getBodyByAcceptEncoding(c.GetRequestHeader(elton.HeaderAcceptEncoding))
	if err != nil {
		return
	}
	c.MergeHeader(resp.Header)
	c.SetHeader(elton.HeaderContentEncoding, encoding)
	c.StatusCode = resp.StatusCode

	c.BodyBuffer = bytes.NewBuffer(body)
	return
}
