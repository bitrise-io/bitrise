package bitrise

import (
	"testing"
)

func TestIsVersionGreaterOrEqual(t *testing.T) {
	t.Log("Yes - Trivial")
	isGreaterOrEql, err := IsVersionGreaterOrEqual("1.1", "1.0")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isGreaterOrEql {
		t.Fatal("Invalid result")
	}

	t.Log("Yes - Trivial - eq")
	isGreaterOrEql, err = IsVersionGreaterOrEqual("1.0", "1.0")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isGreaterOrEql {
		t.Fatal("Invalid result")
	}

	t.Log("No - Trivial")
	isGreaterOrEql, err = IsVersionGreaterOrEqual("1.0", "1.1")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isGreaterOrEql {
		t.Fatal("Invalid result")
	}

	t.Log("Yes - bit more complex - eq")
	isGreaterOrEql, err = IsVersionGreaterOrEqual("1.0.0", "1.0")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isGreaterOrEql {
		t.Fatal("Invalid result")
	}

	t.Log("Yes - bit more complex")
	isGreaterOrEql, err = IsVersionGreaterOrEqual("1.0.1", "1.0")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isGreaterOrEql {
		t.Fatal("Invalid result")
	}

	t.Log("No - bit more complex")
	isGreaterOrEql, err = IsVersionGreaterOrEqual("0.9.1", "1.0")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isGreaterOrEql {
		t.Fatal("Invalid result")
	}
}

func TestIsVersionBetween(t *testing.T) {
	t.Log("Yes - Trivial")
	isBetween, err := IsVersionBetween("1.1", "1.0", "1.2")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isBetween {
		t.Fatal("Invalid result")
	}

	t.Log("No - Trivial")
	isBetween, err = IsVersionBetween("1.3", "1.0", "1.2")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isBetween {
		t.Fatal("Invalid result")
	}

	t.Log("Yes - eq lower")
	isBetween, err = IsVersionBetween("1.0", "1.0", "1.2")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isBetween {
		t.Fatal("Invalid result")
	}

	t.Log("Yes - eq upper")
	isBetween, err = IsVersionBetween("1.2", "1.0", "1.2")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isBetween {
		t.Fatal("Invalid result")
	}

	t.Log("Yes - Bit more complex")
	isBetween, err = IsVersionBetween("1.0.1", "1.0", "1.2")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isBetween {
		t.Fatal("Invalid result")
	}

	t.Log("Yes - Bit more complex - eq")
	isBetween, err = IsVersionBetween("1.2.0", "1.0", "1.2")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isBetween {
		t.Fatal("Invalid result")
	}

	t.Log("No - Bit more complex")
	isBetween, err = IsVersionBetween("1.2.1", "1.0", "1.2")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isBetween {
		t.Fatal("Invalid result")
	}
}
