package cache

import "testing"

func TestLRU(t *testing.T) {
	lru := NewLRU(10)
	key1 := "abcd"
	value1 := &HTTPCache{}
	key2 := "defg"
	value2 := &HTTPCache{}

	lru.Add(key1, value1)
	lru.Add(key1, value1)
	lru.Add(key2, value2)
	v, ok := lru.Get(key1)
	if !ok || v != value1 {
		t.Fatalf("add lru cache fail")
	}

	lru.RemoveOldest()
	v, _ = lru.Get(key2)
	if v != nil {
		t.Fatalf("remove oldest lru cache fail")
	}

	lru.Remove(key1)
	v, ok = lru.Get(key1)
	if v != nil {
		t.Fatalf("remove lru cache fail")
	}

	lru.Add(key1, value1)

	if lru.Len() != 1 {
		t.Fatalf("get lru len fail")
	}

	count := 0
	lru.ForEach(func(key string, value *HTTPCache) {
		count++
	})
	if count != lru.Len() {
		t.Fatalf("lru for each fail")
	}

	lru.Clear()
	if lru.Len() != 0 {
		t.Fatalf("lru clear fail")
	}
}
