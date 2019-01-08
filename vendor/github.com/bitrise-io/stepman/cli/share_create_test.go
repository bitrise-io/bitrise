package cli

import "testing"

func TestValidateTag(t *testing.T) {
	cases := []struct {
		tag   string
		valid bool
	}{
		{tag: "1.0.0", valid: true},
		{tag: "1.0", valid: false},
		{tag: "v1.0.0", valid: false},
	}

	for _, tc := range cases {
		got := validateTag(tc.tag)
		valid := got == nil

		if valid != tc.valid {
			t.Errorf("validateTag(%s) == nil should be %t but got %s", tc.tag, tc.valid, got)
		}
	}
}
