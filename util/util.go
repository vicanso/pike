package util

import (
	"bytes"
	"compress/gzip"
	"image/jpeg"
	"image/png"
	"io/ioutil"

	"github.com/google/brotli/go/cbrotli"
)

// Gzip 对数据压缩
func Gzip(buf []byte, level int) ([]byte, error) {
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

// Brotli brotli压缩
func Brotli(buf []byte, quality int) ([]byte, error) {
	if quality == 0 {
		quality = 9
	}
	return cbrotli.Encode(buf, cbrotli.WriterOptions{
		Quality: quality,
		LGWin:   0,
	})
}

// Gunzip 解压数据
func Gunzip(buf []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

// CompressJPEG 压缩jpeg图片
func CompressJPEG(buf []byte, quality int) ([]byte, error) {
	if quality <= 0 {
		quality = 70
	}
	img, err := jpeg.Decode(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	newBuf := bytes.NewBuffer(nil) //开辟一个新的空buff
	err = jpeg.Encode(newBuf, img, &jpeg.Options{
		Quality: quality,
	})
	return newBuf.Bytes(), err
}

// CompressPNG 压缩png图片
func CompressPNG(buf []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	newBuf := bytes.NewBuffer(nil) //开辟一个新的空buff
	err = png.Encode(newBuf, img)
	if err != nil {
		return nil, err
	}
	return newBuf.Bytes(), nil
}
