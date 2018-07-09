package pike

import (
	"strings"
	"testing"
	"time"
)

func TestServerTiming(t *testing.T) {
	st := NewServerTiming()
	done := st.Start(ServerTimingInitialization)
	time.Sleep(10 * time.Millisecond)
	done()
	arr := strings.Split(st.String(), ",")
	if len(arr) != 2 {
		t.Fatalf("get server timing fail")
	}
	time.Sleep(10 * time.Millisecond)
	st.Reset()
	for _, v := range st.useList {
		if v != 0 {
			t.Fatalf("reset use list fail")
		}
	}
	if time.Now().UnixNano()-st.startedAt > int64(time.Millisecond) {
		t.Fatalf("reset start time fail")
	}
}
