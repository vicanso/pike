package cache

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

// doGunzip gunzip
func doGunzip(buf []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

// Gunzip gunzip
func Gunzip(buf []byte) ([]byte, error) {
	return doGunzip(buf)
}

// Gzip gzip function
func Gzip(buf []byte) ([]byte, error) {
	return doGzip(buf)
}

// doGzip gzip
func doGzip(buf []byte) ([]byte, error) {
	level := compressLevel
	var b bytes.Buffer
	if level <= 0 {
		level = gzip.DefaultCompression
	}
	w, _ := gzip.NewWriterLevel(&b, level)
	_, err := w.Write(buf)
	if err != nil {
		return nil, err
	}
	w.Close()
	return b.Bytes(), nil
}
