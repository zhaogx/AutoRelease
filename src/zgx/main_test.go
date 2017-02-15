package main

import (
	"testing"
	. "utils"
)

func TestF1(t *testing.T) {
	name := "Log.json"
	ret := Vlog_init(name)
	if ret < 0 {
		t.Errorf("Vlog init failed[%s]\n", name)
	}
	//t.Fail()
	t.Logf("[LOG]Vlog init succeed[%s]\n", name)
}

func TestF2(t *testing.T) {
	t.Skip("Skip TestF2")
	name := "Log2.json"
	ret := Vlog_init(name)
	if ret < 0 {
		t.Fatalf("Vlog init failed[%s]\n", name)
	}
	t.Logf("Vlog init succeed[%s]\n", name)
}
