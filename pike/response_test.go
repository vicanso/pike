package pike

import (
	"net/http"
	"testing"
)

func TestResponse(t *testing.T) {
	resp := NewResponse()
	resp.WriteHeader(http.StatusOK)
	resp.Header().Set("A", "B")
	resp.Write([]byte("C"))

	if resp.Status() != http.StatusOK {
		t.Fatalf("get response status fail")
	}

	if resp.Header().Get("A") != "B" {
		t.Fatalf("get response header fail")
	}

	if resp.Size() != 1 {
		t.Fatalf("get response size fail")
	}

	resp.Reset()
	if resp.Status() != http.StatusNotFound {
		t.Fatalf("resset response status fail")
	}

	if len(resp.Header()) != 0 {
		t.Fatalf("reset response header fail")
	}

	if resp.Size() != 0 {
		t.Fatalf("reset response body fail")
	}

}
