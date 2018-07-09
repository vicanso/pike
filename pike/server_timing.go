package pike

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// ServerTimingPike pike
	ServerTimingPike = iota
	// ServerTimingInitialization init中间件
	ServerTimingInitialization
	// ServerTimingIdentifier identifier中间件
	ServerTimingIdentifier
	// ServerTimingDirectorPicker director picker中间件
	ServerTimingDirectorPicker
	// ServerTimingCacheFetcher cache fetcher中间件
	ServerTimingCacheFetcher
	// ServerTimingProxy proxy中间件
	ServerTimingProxy
	// ServerTimingHeaderSetter header setter中间件
	ServerTimingHeaderSetter
	// ServerTimingFreshChecker fresh checker中间件
	ServerTimingFreshChecker
	// ServerTimingDispatcher dispatcher中间件
	ServerTimingDispatcher
	// ServerTimingEnd server timing end
	ServerTimingEnd
)

var (
	// serverTimingDesList server timing的描述
	serverTimingDesList = []string{
		"0;dur=%s;desc=\"pike\"",
		"1;dur=%s;desc=\"init\"",
		"2;dur=%s;desc=\"identifier\"",
		"3;dur=%s;desc=\"director picker\"",
		"4;dur=%s;desc=\"cache fetcher\"",
		"5;dur=%s;desc=\"proxy\"",
		"6;dur=%s;desc=\"header setter\"",
		"7;dur=%s;desc=\"fresh checker\"",
		"7;dur=%s;desc=\"dispatcher\"",
	}
)

type (
	// ServerTiming server timing
	ServerTiming struct {
		disabled      bool
		startedAt     int64
		startedAtList []int64
		useList       []int64
	}
)

// NewServerTiming create a server timing
func NewServerTiming() *ServerTiming {
	return &ServerTiming{
		startedAtList: make([]int64, ServerTimingEnd),
		useList:       make([]int64, ServerTimingEnd),
		startedAt:     time.Now().UnixNano(),
	}
}

// Reset 重置
func (st *ServerTiming) Reset() {
	useList := st.useList
	for i := range useList {
		useList[i] = 0
	}
	st.startedAt = time.Now().UnixNano()
}

// Start 开始server timing的记录
func (st *ServerTiming) Start(index int) func() {
	if st.disabled || index <= ServerTimingPike || index >= ServerTimingEnd {
		return noop
	}
	startedAt := time.Now().UnixNano()
	return func() {
		st.useList[index] = time.Now().UnixNano() - startedAt
		// if st.useList[index] > 10000000 {
		// 	log.Infof("%s use %d", serverTimingDesList[index], st.useList[index])
		// }
	}
}

// String 获取server timing的http header string
func (st *ServerTiming) String() string {
	if st.disabled {
		return ""
	}
	desList := []string{}
	ms := float64(time.Millisecond)
	// use := st.use
	appendDesc := func(v int64, str string) {
		desc := fmt.Sprintf(str, strconv.FormatFloat(float64(v)/ms, 'f', -1, 64))
		desList = append(desList, desc)
	}
	useList := st.useList
	useList[0] = time.Now().UnixNano() - st.startedAt

	for i, v := range st.useList {
		if v != 0 {
			appendDesc(v, serverTimingDesList[i])
		}
	}
	return strings.Join(desList, ",")
}
