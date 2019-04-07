package util

import (
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	// spaceByte 空格
	spaceByte = byte(' ')

	host       = "host"
	method     = "method"
	path       = "path"
	proto      = "proto"
	scheme     = "scheme"
	uri        = "uri"
	userAgent  = "userAgent"
	query      = "query"
	httpProto  = "HTTP"
	httpsProto = "HTTPS"
)

// CheckAndGetValueFromEnv 检查并从env中获取值
func CheckAndGetValueFromEnv(value string) (result string) {
	// key必须为${key}的形式
	reg := regexp.MustCompile(`\$(.+)`)
	groups := reg.FindAllStringSubmatch(value, -1)
	if len(groups) != 0 {
		v := os.Getenv(groups[0][1])
		if len(v) != 0 {
			result = v
		}
	}
	return
}

// GenerateGetIdentity 生成get identity的函数
func GenerateGetIdentity(format string) func(*http.Request) []byte {
	keys := strings.Split(format, " ")
	return func(req *http.Request) []byte {
		values := make([]string, len(keys))
		size := 0
		for i, key := range keys {
			switch key {
			case host:
				values[i] = req.Host
			case method:
				values[i] = req.Method
			case path:
				values[i] = req.URL.Path
			case proto:
				values[i] = req.Proto
			case scheme:
				if req.TLS != nil {
					values[i] = httpsProto
				} else {
					values[i] = httpProto
				}
			case uri:
				values[i] = req.RequestURI
			case userAgent:
				values[i] = req.UserAgent()
			case query:
				values[i] = req.URL.RawQuery
			default:
				first := key[0]
				newKey := key[1:]
				switch first {
				case byte('~'):
					// cookie
					cookie, _ := req.Cookie(newKey)
					if cookie != nil {
						values[i] = cookie.Value
					}
				case byte('>'):
					// request header
					values[i] = req.Header.Get(newKey)
				case byte('?'):
					// request query fields
					values[i] = req.URL.Query().Get(newKey)
					// the invalid field will be ignore
				}
			}
			size += len(values[i])
		}
		spaceCount := len(values) - 1
		buffer := make([]byte, size+spaceCount)
		index := 0
		for i, v := range values {
			copy(buffer[index:], v)
			index += len(v)
			if i < spaceCount {
				buffer[index] = spaceByte
				index++
			}
		}
		return buffer
	}
}

// GetIdentity 获取该请求对应的标识
func GetIdentity(req *http.Request) []byte {
	methodLen := len(req.Method)
	hostLen := len(req.Host)
	uriLen := len(req.RequestURI)
	buffer := make([]byte, methodLen+hostLen+uriLen+2)
	len := 0

	copy(buffer[len:], req.Method)
	len += methodLen

	buffer[len] = spaceByte
	len++

	copy(buffer[len:], req.Host)
	len += hostLen

	buffer[len] = spaceByte
	len++

	copy(buffer[len:], req.RequestURI)
	return buffer
}

// ContainString contain string
func ContainString(arr []string, str string) bool {
	found := false
	for _, v := range arr {
		if v == str {
			found = true
			break
		}
	}
	return found
}

// ConvertToHTTPHeader convert to http header
func ConvertToHTTPHeader(v []string) http.Header {
	if len(v) == 0 {
		return nil
	}
	header := make(http.Header)
	for _, item := range v {
		tmpArr := strings.Split(item, ":")
		if len(tmpArr) != 2 {
			continue
		}
		value := CheckAndGetValueFromEnv(tmpArr[1])
		if value == "" {
			value = tmpArr[1]
		}
		header.Add(tmpArr[0], value)
	}
	return header
}
