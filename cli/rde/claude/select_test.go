package claude

import (
	"strings"
	"testing"

	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
)

func TestResolveDefault(t *testing.T) {
	names := []string{"alpha", "beta", "gamma"}
	cases := map[string]struct {
		pref           string
		backendDefault string
		want           int
	}{
		"saved pref wins":                  {pref: "gamma", backendDefault: "beta", want: 2},
		"backend default when no pref":     {pref: "", backendDefault: "beta", want: 1},
		"stale pref falls to backend":      {pref: "deleted", backendDefault: "beta", want: 1},
		"stale pref and backend → first":   {pref: "deleted", backendDefault: "also-gone", want: 0},
		"nothing set → first":              {pref: "", backendDefault: "", want: 0},
		"pref present, no backend default": {pref: "beta", backendDefault: "", want: 1},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := resolveDefault(names, tc.pref, tc.backendDefault); got != tc.want {
				t.Errorf("resolveDefault(%q, %q) = %d, want %d", tc.pref, tc.backendDefault, got, tc.want)
			}
		})
	}
}

func TestUniqueStacks(t *testing.T) {
	stacks := []internalrde.Stack{
		{ID: "linux", Title: "Ubuntu 24.04"},
		{ID: "linux", Title: "Ubuntu 24.04"}, // duplicate id
		{ID: "mac", Title: "Xcode 16.0", IsDefault: true},
		{ID: "win", Title: "Windows"},
	}
	ids, byID, backendDefault := uniqueStacks(stacks)
	want := []string{"linux", "mac", "win"}
	if len(ids) != len(want) {
		t.Fatalf("ids = %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Errorf("ids[%d] = %q, want %q", i, ids[i], want[i])
		}
	}
	if backendDefault != "mac" {
		t.Errorf("backendDefault = %q, want mac", backendDefault)
	}
	if byID["mac"].Title != "Xcode 16.0" {
		t.Errorf("byID[mac].Title = %q, want Xcode 16.0", byID["mac"].Title)
	}
}

func TestUniqueMachineTypeNames_NoDefault(t *testing.T) {
	mts := []internalrde.MachineType{
		{Name: "small", ClusterName: "a"},
		{Name: "big", ClusterName: "b"},
	}
	names, backendDefault := uniqueMachineTypeNames(mts)
	if len(names) != 2 || names[0] != "small" || names[1] != "big" {
		t.Errorf("names = %v, want [small big]", names)
	}
	if backendDefault != "" {
		t.Errorf("backendDefault = %q, want empty", backendDefault)
	}
}

// TestResolveDefault_StackSwitchFallsBackToFirst documents the machine-type
// behavior when the user switches stacks: the saved machine type (valid for the
// old stack) and the backend default both may be absent from the new stack's
// compatible list, so selection falls back to the first available.
func TestResolveDefault_StackSwitchFallsBackToFirst(t *testing.T) {
	// Compatible machine types for the newly-chosen stack.
	compatible := []string{"arm-small", "arm-big"}
	// Saved pref + backend default are both x86 types, not in the list.
	got := resolveDefault(compatible, "x86-small", "x86-default")
	if got != 0 {
		t.Errorf("got index %d, want 0 (first available)", got)
	}
}

func TestMoveToFront(t *testing.T) {
	cases := map[string]struct {
		in   []string
		idx  int
		want []string
	}{
		"middle moves to front, rest order kept": {in: []string{"a", "b", "c", "d"}, idx: 2, want: []string{"c", "a", "b", "d"}},
		"last moves to front":                    {in: []string{"a", "b", "c"}, idx: 2, want: []string{"c", "a", "b"}},
		"already first is unchanged":             {in: []string{"a", "b", "c"}, idx: 0, want: []string{"a", "b", "c"}},
		"out of range is unchanged":              {in: []string{"a", "b"}, idx: 5, want: []string{"a", "b"}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := moveToFront(tc.in, tc.idx)
			if len(got) != len(tc.want) {
				t.Fatalf("moveToFront(%v, %d) = %v, want %v", tc.in, tc.idx, got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("moveToFront(%v, %d) = %v, want %v", tc.in, tc.idx, got, tc.want)
				}
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	names := []string{"a", "b", "c"}
	if indexOf(names, "b") != 1 {
		t.Error("indexOf should find b at 1")
	}
	if indexOf(names, "") != -1 {
		t.Error("empty target should never match")
	}
	if indexOf(names, "z") != -1 {
		t.Error("absent target should be -1")
	}
}

func TestStackTitle(t *testing.T) {
	byID := map[string]internalrde.Stack{
		"osx-xcode-16.0.x-edge": {ID: "osx-xcode-16.0.x-edge", Title: "Xcode 16.0"},
		"linux-ubuntu-24.04":    {ID: "linux-ubuntu-24.04", Title: "Ubuntu 24.04"},
		"osx-no-title":          {ID: "osx-no-title"}, // backend supplied no title
	}
	for in, want := range map[string]string{
		"osx-xcode-16.0.x-edge": "Xcode 16.0",
		"linux-ubuntu-24.04":    "Ubuntu 24.04",
		"osx-no-title":          "osx-no-title",  // empty title → raw id
		"unknown-stack":         "unknown-stack", // absent → raw id
	} {
		if got := stackTitle(byID, in); got != want {
			t.Errorf("stackTitle(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStackGroupedSecondary(t *testing.T) {
	// Inside a status group the status is dropped (the divider conveys it); the
	// OS and version are kept because they differ row to row.
	cases := map[string]struct {
		st   internalrde.Stack
		want string
	}{
		"macos with version drops status": {internalrde.Stack{OS: "macos", OSVersion: 26, Status: "edge"}, "macOS 26"},
		"linux with version drops status": {internalrde.Stack{OS: "linux", OSVersion: 24, Status: "stable"}, "Linux 24"},
		"no version still shows os":       {internalrde.Stack{OS: "macos", Status: "frozen"}, "macOS"},
		"no os yields empty":              {internalrde.Stack{Status: "edge"}, ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := stackGroupedSecondary(tc.st); got != tc.want {
				t.Errorf("stackGroupedSecondary = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGroupHeader(t *testing.T) {
	for in, want := range map[string]string{
		"edge":   "Edge",
		"stable": "Stable",
		"frozen": "Frozen",
		"":       "Other",
		"beta":   "Beta",
	} {
		if got := groupHeader(in); got != want {
			t.Errorf("groupHeader(%q) = %q, want %q", in, got, want)
		}
	}
}

// stackByID is a small helper to build the id→stack lookup the grouping/OS
// helpers take, in the same order the ids slice is returned.
func stackByID(stacks ...internalrde.Stack) (ids []string, byID map[string]internalrde.Stack) {
	byID = make(map[string]internalrde.Stack, len(stacks))
	for _, st := range stacks {
		ids = append(ids, st.ID)
		byID[st.ID] = st
	}
	return ids, byID
}

func TestGroupStacksByStatus(t *testing.T) {
	// Deliberately mixed input order and a status outside the known set.
	ids, byID := stackByID(
		internalrde.Stack{ID: "s1", Status: "stable"},
		internalrde.Stack{ID: "e1", Status: "edge"},
		internalrde.Stack{ID: "f1", Status: "frozen"},
		internalrde.Stack{ID: "s2", Status: "stable"},
		internalrde.Stack{ID: "b1", Status: "beta"}, // unknown status → after the known ones
		internalrde.Stack{ID: "e2", Status: "edge"},
	)
	groups := groupStacksByStatus(ids, byID)

	wantStatuses := []string{"stable", "edge", "frozen", "beta"}
	if len(groups) != len(wantStatuses) {
		t.Fatalf("groups = %d, want %d: %+v", len(groups), len(wantStatuses), groups)
	}
	for i, want := range wantStatuses {
		if groups[i].status != want {
			t.Errorf("groups[%d].status = %q, want %q", i, groups[i].status, want)
		}
	}
	// Catalog order preserved within a group.
	if got := groups[0].ids; len(got) != 2 || got[0] != "s1" || got[1] != "s2" {
		t.Errorf("stable ids = %v, want [s1 s2]", got)
	}
	if got := groups[1].ids; len(got) != 2 || got[0] != "e1" || got[1] != "e2" {
		t.Errorf("edge ids = %v, want [e1 e2]", got)
	}
}

func TestBuildGroupedStackItems(t *testing.T) {
	ids, byID := stackByID(
		internalrde.Stack{ID: "e1", Title: "Xcode 27", OS: "macos", OSVersion: 27, Status: "edge"},
		internalrde.Stack{ID: "s1", Title: "Xcode 26.5", OS: "macos", OSVersion: 26, Status: "stable"},
	)
	items, stackAt := buildGroupedStackItems(ids, byID)

	// Layout: [divider Stable, s1, divider Edge, e1].
	if len(items) != 4 || len(stackAt) != 4 {
		t.Fatalf("items=%d stackAt=%d, want 4 each", len(items), len(stackAt))
	}
	if !items[0].Divider || items[0].Title != "Stable" || stackAt[0] != "" {
		t.Errorf("items[0] = %+v (stackAt %q), want Stable divider", items[0], stackAt[0])
	}
	if items[1].Divider || items[1].Title != "Xcode 26.5" || items[1].Desc != "macOS 26" || stackAt[1] != "s1" {
		t.Errorf("items[1] = %+v (stackAt %q), want Xcode 26.5 row for s1", items[1], stackAt[1])
	}
	if !items[2].Divider || items[2].Title != "Edge" || stackAt[2] != "" {
		t.Errorf("items[2] = %+v (stackAt %q), want Edge divider", items[2], stackAt[2])
	}
	if items[3].Divider || stackAt[3] != "e1" {
		t.Errorf("items[3] = %+v (stackAt %q), want row for e1", items[3], stackAt[3])
	}
}

func TestStackOSes(t *testing.T) {
	// Linux appears first in the catalog but macOS must still be offered first;
	// an unknown OS is appended; empty-OS stacks are ignored; duplicates collapse.
	ids, byID := stackByID(
		internalrde.Stack{ID: "l1", OS: "linux"},
		internalrde.Stack{ID: "m1", OS: "macos"},
		internalrde.Stack{ID: "w1", OS: "windows"},
		internalrde.Stack{ID: "m2", OS: "macos"},
		internalrde.Stack{ID: "x1", OS: ""},
	)
	got := stackOSes(ids, byID)
	want := []string{"macos", "linux", "windows"}
	if len(got) != len(want) {
		t.Fatalf("stackOSes = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("stackOSes[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestFilterStacksByOS(t *testing.T) {
	ids, byID := stackByID(
		internalrde.Stack{ID: "m1", OS: "macos"},
		internalrde.Stack{ID: "l1", OS: "linux"},
		internalrde.Stack{ID: "m2", OS: "macos"},
	)
	got := filterStacksByOS(ids, byID, "macos")
	if len(got) != 2 || got[0] != "m1" || got[1] != "m2" {
		t.Errorf("filterStacksByOS(macos) = %v, want [m1 m2]", got)
	}
	if got := filterStacksByOS(ids, byID, "windows"); len(got) != 0 {
		t.Errorf("filterStacksByOS(windows) = %v, want empty", got)
	}
}

func TestAnyStackMissingOS(t *testing.T) {
	tagged, byIDTagged := stackByID(
		internalrde.Stack{ID: "m1", OS: "macos"},
		internalrde.Stack{ID: "l1", OS: "linux"},
	)
	if anyStackMissingOS(tagged, byIDTagged) {
		t.Error("all stacks tagged, want false")
	}
	untagged, byIDUntagged := stackByID(
		internalrde.Stack{ID: "m1", OS: "macos"},
		internalrde.Stack{ID: "x1", OS: ""},
	)
	if !anyStackMissingOS(untagged, byIDUntagged) {
		t.Error("one stack missing OS, want true")
	}
}

func TestDefaultStackID(t *testing.T) {
	ids := []string{"a", "b", "c"}
	if got := defaultStackID(ids, "b", "c"); got != "b" {
		t.Errorf("saved pref: got %q, want b", got)
	}
	if got := defaultStackID(ids, "", "c"); got != "c" {
		t.Errorf("backend default: got %q, want c", got)
	}
	if got := defaultStackID(ids, "gone", "also-gone"); got != "a" {
		t.Errorf("fallback: got %q, want a (first)", got)
	}
}

func TestReuseDetail(t *testing.T) {
	// Two lines: the stack title, then the machine display name plus its spec in
	// parentheses.
	got := reuseDetail("Xcode 26", "M2 Pro Large", "6 vCPU · 14 GB")
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("reuseDetail lines = %d, want 2: %q", len(lines), got)
	}
	if !strings.Contains(lines[0], "Stack") || !strings.Contains(lines[0], "Xcode 26") {
		t.Errorf("stack line = %q", lines[0])
	}
	if !strings.Contains(lines[1], "Machine type") || !strings.Contains(lines[1], "M2 Pro Large") || !strings.Contains(lines[1], "6 vCPU · 14 GB") {
		t.Errorf("machine line = %q", lines[1])
	}
	// No spec → no parenthetical.
	if got := reuseDetail("Ubuntu 24.04", "g2.mac", ""); strings.Contains(got, "(") {
		t.Errorf("reuseDetail without specs should have no parens: %q", got)
	}
}

func TestBuildReuseMenu(t *testing.T) {
	for _, tc := range []struct {
		name         string
		multiStack   bool
		multiMachine bool
		wantTitles   []string
		wantActions  []reuseAction
	}{
		{"both changeable", true, true,
			[]string{"Use this setup", "Change stack", "Change machine type"},
			[]reuseAction{reuseUse, reuseChangeStack, reuseChangeMachine}},
		{"single machine type hides machine row", true, false,
			[]string{"Use this setup", "Change stack"},
			[]reuseAction{reuseUse, reuseChangeStack}},
		{"single stack hides stack row", false, true,
			[]string{"Use this setup", "Change machine type"},
			[]reuseAction{reuseUse, reuseChangeMachine}},
		{"single of both → reuse only", false, false,
			[]string{"Use this setup"},
			[]reuseAction{reuseUse}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			items, actions := buildReuseMenu(tc.multiStack, tc.multiMachine)
			if len(items) != len(tc.wantTitles) {
				t.Fatalf("items = %d, want %d", len(items), len(tc.wantTitles))
			}
			for i, want := range tc.wantTitles {
				if items[i].Title != want {
					t.Errorf("items[%d].Title = %q, want %q", i, items[i].Title, want)
				}
			}
			for i, want := range tc.wantActions {
				if actions[i] != want {
					t.Errorf("actions[%d] = %d, want %d", i, actions[i], want)
				}
			}
		})
	}
}

func TestMachineSpecHint(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{"g2.mac.m2pro.4c-6g", "4 vCPU · 6 GB"},
		{"g2.linux.amd-zen5.8c-32g", "8 vCPU · 32 GB"},
		{"8c-16g", "8 vCPU · 16 GB"}, // bare segment, no dots
		{"g2.mac", ""},               // last segment "mac" doesn't match
		{"g2.linux.bad", ""},
		{"", ""},
	} {
		if got := machineSpecHint(tc.name); got != tc.want {
			t.Errorf("machineSpecHint(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestMachineItem_UsesBackendTitleAndSpecs(t *testing.T) {
	byName := indexMachineTypes([]internalrde.MachineType{
		{Name: "g2.mac.m2pro.4c-6g", Title: "M2 Pro Large", CPU: "4 vCPU", RAM: "6 GB"},
		{Name: "g2.bare"}, // no friendly metadata
	})
	item := machineItem(byName)

	// Friendly title becomes the row title; the specs are the dim secondary
	// text. The raw contract name is intentionally NOT shown — it's just noise.
	got := item("g2.mac.m2pro.4c-6g")
	if got.Title != "M2 Pro Large" {
		t.Errorf("title = %q, want M2 Pro Large", got.Title)
	}
	if got.Desc != "4 vCPU · 6 GB" {
		t.Errorf("desc = %q, want %q", got.Desc, "4 vCPU · 6 GB")
	}
	if strings.Contains(got.Desc, "g2.mac.m2pro.4c-6g") {
		t.Errorf("desc %q should not include the raw machine-type name", got.Desc)
	}

	// No backend metadata → row title is the raw name (no name duplicated in desc).
	bare := item("g2.bare")
	if bare.Title != "g2.bare" {
		t.Errorf("title = %q, want raw name g2.bare", bare.Title)
	}

	// Name absent from the catalog map → graceful fallback to the raw name.
	if unknown := item("nope"); unknown.Title != "nope" {
		t.Errorf("title = %q, want nope", unknown.Title)
	}
}

func TestMachineLabel(t *testing.T) {
	cases := map[string]struct {
		mt   internalrde.MachineType
		want string
	}{
		"backend title + specs": {
			internalrde.MachineType{Name: "g2.mac.m2pro.12c-28g", Title: "M2 Pro Large", CPU: "12 vCPU", RAM: "28 GB"},
			"M2 Pro Large (12 vCPU · 28 GB)",
		},
		"no title, specs parsed from name": {
			internalrde.MachineType{Name: "g2.mac.m2pro.12c-28g"},
			"g2.mac.m2pro.12c-28g (12 vCPU · 28 GB)",
		},
		"no title, no parseable specs": {
			internalrde.MachineType{Name: "g2.bare"},
			"g2.bare",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := machineLabel(tc.mt); got != tc.want {
				t.Errorf("machineLabel = %q, want %q", got, tc.want)
			}
		})
	}
}
