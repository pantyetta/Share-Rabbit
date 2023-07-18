package main

import "testing"

func TestAdd(t *testing.T) {
	err := add("key", "value")
	if err != nil {
		t.Errorf("Catch err: [%v]", err)
	}
}

func TestGet(t *testing.T) {

	value, err := get("key")
	if err != nil {
		t.Errorf("Catch err: [%v]", err)
	}
	if value != "value" {
		t.Errorf("Catch err: don't match \"value\"")
	}
}
