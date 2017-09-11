package pointers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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

func TestNewTimePtr(t *testing.T) {
	t.Log("Create a time")
	if (*NewTimePtr(time.Date(2009, time.January, 1, 0, 0, 0, 0, time.UTC))).Equal(time.Date(2009, time.January, 1, 0, 0, 0, 0, time.UTC)) == false {
		t.Fatal("Invalid pointer")
	}

	t.Log("Try to change the original value - should not be affected!")
	myTime := time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC)
	myTimePtr := NewTimePtr(myTime)
	if (*myTimePtr).Equal(time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC)) == false {
		t.Fatal("Invalid pointer - original value")
	}
	*myTimePtr = time.Date(2015, time.January, 1, 0, 0, 0, 0, time.UTC)
	if *myTimePtr != time.Date(2015, time.January, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatal("Invalid pointer - changed value")
	}
	// the original var should remain intact!
	if myTime.Equal(time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC)) == false {
		t.Fatal("The original var was affected!!")
	}
}

func TestNewIntPtr(t *testing.T) {
	t.Log("Create 1 ptr")
	if *NewIntPtr(1) != 1 {
		t.Fatal("Invalid pointer")
	}

	t.Log("Create 0 ptr")
	if *NewIntPtr(0) != 0 {
		t.Fatal("Invalid pointer")
	}

	t.Log("Try to change the original value - should not be affected!")
	myint := 2
	myintPtr := NewIntPtr(myint)
	if *myintPtr != 2 {
		t.Fatal("Invalid pointer - original value")
	}

	*myintPtr = 3
	if *myintPtr != 3 {
		t.Fatal("Invalid pointer - changed value")
	}
	// the original var should remain intact!
	if myint != 2 {
		t.Fatal("The original var was affected!!")
	}
}

func TestNewInt64Ptr(t *testing.T) {
	t.Log("Create 1 ptr")
	if *NewInt64Ptr(1) != 1 {
		t.Fatal("Invalid pointer")
	}

	t.Log("Create 0 ptr")
	if *NewInt64Ptr(0) != 0 {
		t.Fatal("Invalid pointer")
	}

	t.Log("Try to change the original value - should not be affected!")
	myint := int64(2)
	myintPtr := NewInt64Ptr(myint)
	if *myintPtr != 2 {
		t.Fatal("Invalid pointer - original value")
	}

	*myintPtr = 3
	if *myintPtr != 3 {
		t.Fatal("Invalid pointer - changed value")
	}
	// the original var should remain intact!
	if myint != 2 {
		t.Fatal("The original var was affected!!")
	}
}

func TestNewMapStringInterfacePtr(t *testing.T) {
	t.Log("Create a map")
	{
		ptr := NewMapStringInterfacePtr(map[string]interface{}{"key": "value"})
		require.Equal(t, 1, len(*ptr))
		require.Equal(t, "value", (*ptr)["key"])
	}

	t.Log("Try to change the original value - should not be affected!")
	{
		myMap := map[string]interface{}{"key": "orig-value"}
		myMapPtr := NewMapStringInterfacePtr(myMap)
		require.Equal(t, 1, len(*myMapPtr))
		require.Equal(t, "orig-value", (*myMapPtr)["key"])

		(*myMapPtr)["key"] = "new-value"
		require.Equal(t, 1, len(*myMapPtr))
		require.Equal(t, "new-value", (*myMapPtr)["key"])

		// the original var should remain intact!
		require.Equal(t, 1, len(myMap))
		require.Equal(t, "orig-value", (myMap)["key"])
	}

	t.Log("Try to change the original value - should not be affected!")
	{
		type TestObject struct {
			Title   string
			Message string
		}

		// create ptr with an orig value
		testObj := TestObject{
			Title:   "title",
			Message: "message",
		}
		myMap := map[string]interface{}{"key": testObj}
		myMapPtr := NewMapStringInterfacePtr(myMap)

		require.Equal(t, 1, len(*myMapPtr))
		obj, ok := (*myMapPtr)["key"]
		require.Equal(t, true, ok)

		casted, ok := obj.(TestObject)
		require.Equal(t, true, ok)
		require.Equal(t, "title", casted.Title)
		require.Equal(t, "message", casted.Message)

		// modify the value of the ptr
		casted.Title = "new-title"
		casted.Message = "new-message"

		(*myMapPtr)["key"] = casted
		require.Equal(t, 1, len(*myMapPtr))
		newObj, ok := (*myMapPtr)["key"]
		require.Equal(t, true, ok)

		newCasted, ok := newObj.(TestObject)
		require.Equal(t, true, ok)
		require.Equal(t, "new-title", newCasted.Title)
		require.Equal(t, "new-message", newCasted.Message)

		// the original var should remain intact!
		require.Equal(t, 1, len(myMap))
		origObj, ok := myMap["key"]
		require.Equal(t, true, ok)

		origCasted, ok := origObj.(TestObject)
		require.Equal(t, true, ok)
		require.Equal(t, "title", origCasted.Title)
		require.Equal(t, "message", origCasted.Message)
	}
}

func TestBool(t *testing.T) {
	require.Equal(t, false, Bool(nil))

	sampleVal := true
	sampleValPtr := &sampleVal
	require.Equal(t, true, Bool(sampleValPtr))
}

func TestBoolWithDefault(t *testing.T) {
	require.Equal(t, false, BoolWithDefault(nil, false))
	require.Equal(t, true, BoolWithDefault(nil, true))

	sampleVal := true
	sampleValPtr := &sampleVal
	require.Equal(t, true, BoolWithDefault(sampleValPtr, false))
}

func TestString(t *testing.T) {
	require.Equal(t, "", String(nil))

	sampleStr := "sample string"
	sampleStrPtr := &sampleStr
	require.Equal(t, "sample string", String(sampleStrPtr))
}

func TestStringWithDefault(t *testing.T) {
	require.Equal(t, "", StringWithDefault(nil, ""))
	require.Equal(t, "default value", StringWithDefault(nil, "default value"))

	sampleStr := "sample string"
	sampleStrPtr := &sampleStr
	require.Equal(t, "sample string", StringWithDefault(sampleStrPtr, "default value"))
}

func TestTimeWithDefault(t *testing.T) {
	const longForm = "Jan 2, 2006 at 3:04pm (MST)"
	defaultTime, err := time.Parse(longForm, "Feb 3, 2013 at 7:54pm (PST)")
	require.NoError(t, err)

	require.Equal(t, defaultTime, TimeWithDefault(nil, defaultTime))

	anotherTime, err := time.Parse(longForm, "Feb 4, 2014 at 8:54pm (PST)")
	require.NoError(t, err)
	anotherTimePtr := &anotherTime

	require.Equal(t, anotherTime, TimeWithDefault(anotherTimePtr, defaultTime))
}

func TestInt(t *testing.T) {
	require.Equal(t, 0, Int(nil))

	sampleVal := 12
	sampleValPtr := &sampleVal
	require.Equal(t, 12, Int(sampleValPtr))
}

func TestIntWithDefault(t *testing.T) {
	require.Equal(t, 0, IntWithDefault(nil, 0))
	require.Equal(t, 12, IntWithDefault(nil, 12))
	require.Equal(t, -12, IntWithDefault(nil, -12))

	sampleVal := 23
	sampleValPtr := &sampleVal
	require.Equal(t, 23, IntWithDefault(sampleValPtr, 1))
}
