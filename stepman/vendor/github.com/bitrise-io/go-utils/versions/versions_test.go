package versions

import (
	"testing"

	"github.com/bitrise-io/go-utils/testutil"
)

func TestCompareVersions(t *testing.T) {
	res, err := CompareVersions("1.0.0", "1.0.1")
	testutil.EqualAndNoError(t, 1, res, err, "Trivial compare")

	res, err = CompareVersions("1.0.2", "1.0.1")
	testutil.EqualAndNoError(t, -1, res, err, "Reverse compare")

	res, err = CompareVersions("1.0.2", "1.0.2")
	testutil.EqualAndNoError(t, 0, res, err, "Equal compare")

	t.Log("1.0.0 <-> 0.9.8")
	if res, err := CompareVersions("1.0.0", "0.9.8"); res != -1 || err != nil {
		t.Fatal("Failed, res:", res, "| err:", err)
	}
	t.Log("0.9.8 <-> 1.0.0")
	if res, err := CompareVersions("0.9.8", "1.0.0"); res != 1 || err != nil {
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

	t.Log("No - 1.0.0<->0.9.8")
	isGreaterOrEql, err = IsVersionGreaterOrEqual("1.0.0", "0.9.8")
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isGreaterOrEql {
		t.Fatal("Invalid result")
	}

	t.Log("No - 0.9.8<->1.0.0")
	isGreaterOrEql, err = IsVersionGreaterOrEqual("0.9.8", "1.0.0")
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
