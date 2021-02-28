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

package log

import (
	"net/url"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	err := zap.RegisterSink("lumberjack", newLumberJack)
	if err != nil {
		panic(err)
	}
}

var defaultLogger = newLoggerX("")

type LumberjackLogger struct {
	lumberjack.Logger
}

func (ll *LumberjackLogger) Sync() error {
	return nil
}

func newLumberJack(u *url.URL) (zap.Sink, error) {
	maxSize := 0
	v := u.Query().Get("maxSize")
	if v != "" {
		maxSize, _ = strconv.Atoi(v)
	}
	maxAge := 0
	v = u.Query().Get("maxAge")
	if v != "" {
		maxAge, _ = strconv.Atoi(v)
	}
	if maxAge == 0 {
		maxAge = 1
	}
	compress := false
	if u.Query().Get("compress") == "true" {
		compress = true
	}

	return &LumberjackLogger{
		Logger: lumberjack.Logger{
			MaxSize:  maxSize,
			MaxAge:   maxAge,
			Filename: u.Path,
			Compress: compress,
		},
	}, nil
}

// newLoggerX 初始化logger
func newLoggerX(outputPath string) *zap.Logger {

	c := zap.NewProductionConfig()
	if outputPath != "" {
		c.OutputPaths = []string{
			outputPath,
		}
		c.ErrorOutputPaths = []string{
			outputPath,
		}
	}

	// 在一秒钟内, 如果某个级别的日志输出量超过了 Initial, 那么在超过之后, 每 Thereafter 条日志才会输出一条, 其余的日志都将被删除
	// 如果需要输出所有日志，则设置为nil
	c.Sampling = nil
	// pike的日志比较简单，因此不添加caller
	c.DisableCaller = true

	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 只针对panic 以上的日志增加stack trace
	l, err := c.Build(zap.AddStacktrace(zap.DPanicLevel))
	if err != nil {
		panic(err)
	}
	return l
}

func SetOutputPath(outputPath string) {
	defaultLogger = newLoggerX(outputPath)
}

// Default get default logger
func Default() *zap.Logger {
	return defaultLogger
}
