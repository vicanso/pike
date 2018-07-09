package pike

import (
	"bytes"
	"net/http"
)

type (
	// Response http response
	Response struct {
		body      *bytes.Buffer
		headers   http.Header
		code      int
		Committed bool
	}
)

// NewResponse create a response
func NewResponse() *Response {
	return &Response{
		body:    new(bytes.Buffer),
		headers: make(http.Header),
		code:    http.StatusNotFound,
	}
}

// WriteHeader write header
func (w *Response) WriteHeader(code int) {
	w.code = code
}

// Header get header
func (w *Response) Header() http.Header {
	return w.headers
}

// Write write buffer
func (w *Response) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// Status get the response status
func (w *Response) Status() int {
	return w.code
}

// Size get the response size
func (w *Response) Size() int {
	return w.body.Len()
}

// Reset reset the response sturct
func (w *Response) Reset() {
	w.body.Reset()
	w.code = http.StatusNotFound
	w.Committed = false
	for k := range w.headers {
		delete(w.headers, k)
	}
}

// Bytes get the bytes of response
func (w *Response) Bytes() []byte {
	return w.body.Bytes()
}
