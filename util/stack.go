package util

import (
	"regexp"
	"runtime"
	"strings"
)

var (
	stackReg = regexp.MustCompile(`\((0x[\s\S]+)\)`)
	appPath  = ""
)

func init() {
	_, file, _, _ := runtime.Caller(0)
	fileDivide := "/"
	arr := strings.Split(file, fileDivide)
	arr = arr[0 : len(arr)-2]
	appPath = strings.Join(arr, fileDivide) + fileDivide
}

// GetStack 获取调用信息
func GetStack(start, end int) []string {
	size := 2 << 10
	stack := make([]byte, size)
	runtime.Stack(stack, true)
	arr := strings.Split(string(stack), "\n")
	arr = arr[1:]
	max := len(arr) - 1
	result := []string{}
	for index := 0; index < max; index += 2 {
		if index+1 >= max {
			break
		}

		file := strings.Replace(arr[index+1], appPath, "", 1)
		tmpArr := strings.Split(arr[index], "/")
		fn := stackReg.ReplaceAllString(tmpArr[len(tmpArr)-1], "")
		str := fn + ": " + file
		result = append(result, str)
	}
	return result[start:end]
}
