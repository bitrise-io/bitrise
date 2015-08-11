package pointers

import "testing"

func TestNewBoolPtr(t *testing.T) {
	t.Log("Create false ptr")
	if *NewBoolPtr(false) != false {
		t.Fatal("Invalid pointer")
	}

	t.Log("Create true ptr")
	if *NewBoolPtr(true) != true {
		t.Fatal("Invalid pointer")
	}

	t.Log("Try to change the original value - should not be affected!")
	mybool := true
	myboolPtr := NewBoolPtr(mybool)
	if *myboolPtr != true {
		t.Fatal("Invalid pointer - original value")
	}
	*myboolPtr = false
	if *myboolPtr != false {
		t.Fatal("Invalid pointer - changed value")
	}
	// the original var should remain intact!
	if mybool != true {
		t.Fatal("The original var was affected!!")
	}
}

func TestNewStringPtr(t *testing.T) {
	t.Log("Create a string")
	if *NewStringPtr("mystr") != "mystr" {
		t.Fatal("Invalid pointer")
	}

	t.Log("Try to change the original value - should not be affected!")
	myStr := "my-orig-str"
	myStrPtr := NewStringPtr(myStr)
	if *myStrPtr != "my-orig-str" {
		t.Fatal("Invalid pointer - original value")
	}
	*myStrPtr = "new-str-value"
	if *myStrPtr != "new-str-value" {
		t.Fatal("Invalid pointer - changed value")
	}
	// the original var should remain intact!
	if myStr != "my-orig-str" {
		t.Fatal("The original var was affected!!")
	}
}
