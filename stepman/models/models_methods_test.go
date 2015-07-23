package models

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	t.Log("Trivial compare")
	if res, err := CompareVersions("1.0.0", "1.0.1"); res != 1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}

	t.Log("Reverse compare")
	if res, err := CompareVersions("1.0.2", "1.0.1"); res != -1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}

	t.Log("Equal compare")
	if res, err := CompareVersions("1.0.2", "1.0.2"); res != 0 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}

	t.Log("Missing last num in first")
	if res, err := CompareVersions("7.0", "7.0.2"); res != 1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing last num in first - eql")
	if res, err := CompareVersions("7.0", "7.0.0"); res != 0 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing last num in second")
	if res, err := CompareVersions("7.0.2", "7.0"); res != -1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing last num in second - eql")
	if res, err := CompareVersions("7.0.0", "7.0"); res != 0 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing double-last num in first")
	if res, err := CompareVersions("7", "7.0.2"); res != 1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing double-last num in first - eql")
	if res, err := CompareVersions("7", "7.0.0"); res != 0 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing double-last num in second")
	if res, err := CompareVersions("7.0.2", "7"); res != -1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("Missing double-last num in second - eql")
	if res, err := CompareVersions("7.0.0", "7"); res != 0 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}

	// specials are not handled but should not cause any issue / panic
	t.Log("Special / non number component")
	if res, err := CompareVersions("7.x.1.2.3", "7.0.1.x"); err == nil {
		t.Fatal("Not supported compare should return an error!")
	} else {
		t.Log("[expected] Failed, res:", res, "| err:", err)
	}
}
