package compress

import (
	"compress/gzip"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
	"github.com/vicanso/pike/config"
)

var compressTestData = []byte(`Brotli is a data format specification[2] for data streams compressed with a specific combination of the general-purpose LZ77 lossless compression algorithm, Huffman coding and 2nd order context modelling. Brotli is a compression algorithm developed by Google and works best for text compression.

Google employees Jyrki Alakuijala and Zoltán Szabadka initially developed Brotli to decrease the size of transmissions of WOFF2 web fonts, and in that context Brotli was a continuation of the development of zopfli, which is a zlib-compatible implementation of the standard gzip and deflate specifications. Brotli allows a denser packing than gzip and deflate because of several algorithmic and format level improvements: the use of context models for literals and copy distances, describing copy distances through past distances, use of move-to-front queue in entropy code selection, joint-entropy coding of literal and copy lengths, the use of graph algorithms in block splitting, and a larger backward reference window are example improvements. The Brotli specification was generalized in September 2015 for HTTP stream compression (content-encoding type 'br'). This generalized iteration also improved the compression ratio by using a pre-defined dictionary of frequently used words and phrases.`)

func TestConvertConfig(t *testing.T) {
	assert := assert.New(t)

	name := "compress-test"
	levels := map[string]uint{
		"gzip": 9,
		"br":   8,
	}
	configs := []config.CompressConfig{
		{
			Name:   name,
			Levels: levels,
		},
	}
	opts := convertConfigs(configs)
	assert.Equal(1, len(opts))
	assert.Equal(name, opts[0].Name)
	assert.Equal(9, opts[0].Levels["gzip"])
	assert.Equal(8, opts[0].Levels["br"])
}

func TestCompressLevel(t *testing.T) {
	assert := assert.New(t)
	srv := NewService()
	assert.Equal(gzip.DefaultCompression, srv.GetLevel(EncodingGzip))
	assert.Equal(brotli.DefaultCompression, srv.GetLevel(EncodingBrotli))

	srv.SetLevels(map[string]int{
		EncodingGzip:   1,
		EncodingBrotli: 2,
	})
	assert.Equal(1, srv.GetLevel(EncodingGzip))
	assert.Equal(2, srv.GetLevel(EncodingBrotli))
}

func TestCompressList(t *testing.T) {
	assert := assert.New(t)
	srvList := NewServices([]CompressOption{
		{
			Name: "test",
			Levels: map[string]int{
				"gzip": 1,
				"br":   2,
			},
		},
	})
	srv := srvList.Get("test")
	assert.Equal(1, srv.GetLevel("gzip"))
	assert.Equal(2, srv.GetLevel("br"))

	srvList.Reset([]CompressOption{
		{
			Name: "test1",
			Levels: map[string]int{
				"gzip": 3,
				"br":   4,
			},
		},
	})
	// compress并不删除原有的srv
	assert.Equal(srv, srvList.Get("test"))

	srv = srvList.Get("test1")
	assert.Equal(3, srv.GetLevel("gzip"))
	assert.Equal(4, srv.GetLevel("br"))

	// 从默认获取
	assert.Equal(defaultCompressSrv, Get("test"))
	Reset([]config.CompressConfig{
		{
			Name: "test",
			Levels: map[string]uint{
				"gzip": 3,
				"br":   4,
			},
		},
	})
	assert.Equal(3, Get("test").GetLevel("gzip"))
	assert.Equal(4, Get("test").GetLevel("br"))
}

func TestDecompress(t *testing.T) {
	assert := assert.New(t)
	data := compressTestData
	// 不同的压缩解压
	tests := []struct {
		fn       func() ([]byte, error)
		encoding string
	}{
		{
			fn: func() ([]byte, error) {
				return Get("").Gzip(data)
			},
			encoding: EncodingGzip,
		},
		{
			fn: func() ([]byte, error) {
				return Get("").Brotli(data)
			},
			encoding: EncodingBrotli,
		},
		{
			fn: func() ([]byte, error) {
				return doLZ4Encode(data, 0)
			},
			encoding: EncodingLZ4,
		},
		{
			fn: func() ([]byte, error) {
				dst := doSnappyEncode(data)
				return dst, nil
			},
			encoding: EncodingSnappy,
		},
		{
			fn: func() ([]byte, error) {
				return data, nil
			},
			encoding: "",
		},
	}
	for _, tt := range tests {
		result, err := tt.fn()
		assert.Nil(err)
		assert.NotEmpty(result)
		result, err = Get("").Decompress(tt.encoding, result)
		assert.Nil(err)
		assert.Equal(data, result)
	}
	_, err := Get("").Decompress("a", nil)
	assert.Equal(notSupportedEncoding, err)
}
