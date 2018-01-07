package performance

import (
	"testing"
	"time"
)

func TestConcurrency(t *testing.T) {
	IncreaseConcurrency()
	IncreaseConcurrency()
	if GetConcurrency() != 2 {
		t.Fatalf("increase concurrency excpect 2 but %v", GetConcurrency())
	}
	DecreaseConcurrency()
	if GetConcurrency() != 1 {
		t.Fatalf("decrease concurrency excpect 1 but %v", GetConcurrency())
	}
}

func TestIncreaseRequestCount(t *testing.T) {
	IncreaseRequestCount()
	IncreaseRequestCount()
	now := time.Now()
	hour := now.Hour()
	minute := now.Minute()
	index := hour*60 + minute
	list := GetRequstCountList()
	if list[index] != 2 {
		t.Fatalf("the request count expect 2 but %v", list[index])
	}
}
