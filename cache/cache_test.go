package cache

import (
	"bytes"
	"testing"

	"../vars"
)

func TestStatus(t *testing.T) {
	key := "aslant.site/users/me"
	status := GetStatus(key)
	none := vars.None
	if status != none {
		t.Fatalf("the status should be none")
	}
	status = GetStatus(key)
	if status == none {
		t.Fatalf("the status should not be none")
	}
	DeleteStatus(key)
	status = GetStatus(key)
	if status != none {
		t.Fatalf("the status should be none")
	}
	SetHitForPass(key)
	status = GetStatus(key)
	if status != vars.HitForPass {
		t.Fatalf("the status should be hit for pass")
	}
}

func TestDB(t *testing.T) {
	db, err := Init("/tmp/pike.db")
	if err != nil {
		t.Fatalf("open db fail, %v", err)
	}
	defer db.Close()
	key := "aslant.site/users/me"
	data := []byte("vicanso")
	err = Save(key, data)
	if err != nil {
		t.Fatalf("save data fail %v", err)
	}
	buf, err := Get(key)
	if err != nil {
		t.Fatalf("get data fail %v", err)
	}
	if bytes.Compare(data, buf) != 0 {
		t.Fatalf("get data fail")
	}
}
