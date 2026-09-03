package claude

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil/picker"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/rde/localsession"
)

// errSelectionCancelled signals the user backed out of the stack / machine-type
// picker (empty-with-no-default impossible, "q", EOF, or Ctrl-C). The command
// treats it as a clean exit, like the resume picker's cancel.
var errSelectionCancelled = errors.New("selection cancelled")

// selectStackAndMachineType resolves the stack and machine type for a fresh
// session, mirroring the RDE web UI: pick a stack first, then a machine type
// compatible with that stack. When a still-valid combo is remembered for this
// project (and neither --stack nor --machine-type is set), it first offers a
// one-step "use your last setup" menu so returning users don't re-pick the same
// pair. For each pick the choice is, in order: an explicit flag, the only option
// when there's just one (so a single machine type starts the session without a
// prompt), or the interactive picker (default = the per-project saved choice,
// else the backend default, else the first). When stdin/stderr isn't a terminal
// the default is used without prompting.
//
// It returns the chosen stack ID and machine type NAME (the contracts, ready for
// CreateSession), plus human-friendly labels for each (for display).
func selectStackAndMachineType(ctx context.Context, cmd *cobra.Command, svc *internalrde.Service, log *stepLogger, workspaceID, repoPath, flagStack, flagMachineType string) (stack, stackLabel, machineType, machineLbl string, err error) {
	// Best-effort: a missing/corrupt prefs file yields the zero value, i.e.
	// "no prior choice", so we fall through to the backend default / first item.
	prefs, _ := localsession.LoadPrefs(repoPath)

	stacks, err := svc.ListStacks(ctx, workspaceID)
	if err != nil {
		return "", "", "", "", fmt.Errorf("list stacks: %w", err)
	}
	if len(stacks) == 0 {
		return "", "", "", "", fmt.Errorf("no stacks are available in this workspace")
	}
	stackIDs, stacksByID, backendDefaultStack := uniqueStacks(stacks)

	// Fast path: reuse the remembered combo for this project in one step.
	if flagStack == "" && flagMachineType == "" && interactivePicker(cmd) && prefs.Stack != "" && prefs.MachineType != "" {
		stack, machineType, machineLbl, done, err := offerReuse(ctx, cmd, svc, log, workspaceID, stackIDs, stacksByID, prefs)
		if err != nil || done {
			return stack, stackTitle(stacksByID, stack), machineType, machineLbl, err
		}
		// The remembered combo is stale, or the user chose "Change stack" — fall
		// through to the full pick below (saved values still seed the defaults).
	}

	stack, err = chooseStack(ctx, cmd, log, stackIDs, stacksByID, prefs.Stack, backendDefaultStack, flagStack)
	if err != nil {
		return "", "", "", "", err
	}

	machineType, machineLbl, err = chooseMachineForStack(ctx, cmd, svc, log, workspaceID, stack, prefs.MachineType, flagMachineType)
	if err != nil {
		return "", "", "", "", err
	}
	return stack, stackTitle(stacksByID, stack), machineType, machineLbl, nil
}

// chooseMachineForStack resolves the machine type for the chosen stack: it
// fetches the compatible types and defers to chooseOne, which auto-selects when
// only one is available — so a single compatible machine type starts the session
// without a machine-type prompt. It returns the chosen machine type's contract
// name and a human-friendly label for display.
func chooseMachineForStack(ctx context.Context, cmd *cobra.Command, svc *internalrde.Service, log *stepLogger, workspaceID, stack, prefMachineType, flagMachineType string) (name, label string, err error) {
	mts, err := svc.MachineTypesForStack(ctx, workspaceID, stack)
	if err != nil {
		return "", "", err
	}
	if len(mts) == 0 {
		return "", "", fmt.Errorf("no machine types are compatible with stack %q", stack)
	}
	mtByName := indexMachineTypes(mts)
	mtNames, backendDefaultMT := uniqueMachineTypeNames(mts)
	name, err = chooseOne(ctx, cmd, log, "machine type", "Select a machine type", mtNames, prefMachineType, backendDefaultMT, flagMachineType, machineItem(mtByName))
	if err != nil {
		return "", "", err
	}
	return name, machineLabel(mtByName[name]), nil
}

// resolveWithoutPrompt handles the three cases where a choice needs no
// interactive picker: an explicit flag (validated against options), a lone
// option, or a non-terminal stdin/stderr (the resolved default). ok=false with
// a nil error means an interactive pick is required. It's shared by chooseOne
// and chooseStack so both honor flags and the non-TTY fallback identically —
// the OS prompt and status grouping are interactive-only refinements layered on
// top. options must be non-empty.
func resolveWithoutPrompt(cmd *cobra.Command, log *stepLogger, noun string, options []string, prefName, backendDefault, flag string) (value string, ok bool, err error) {
	if flag != "" {
		if indexOf(options, flag) < 0 {
			return "", true, fmt.Errorf("%s %q is not available; choose one of: %s", noun, flag, strings.Join(options, ", "))
		}
		log.step("Using %s %q", noun, flag)
		return flag, true, nil
	}
	if len(options) == 1 {
		log.step("Using the only %s available: %s", noun, options[0])
		return options[0], true, nil
	}
	if !interactivePicker(cmd) {
		idx := resolveDefault(options, prefName, backendDefault)
		log.step("Using default %s (stdin is not a terminal): %s", noun, options[idx])
		return options[idx], true, nil
	}
	return "", false, nil
}

// chooseOne resolves a single selection. An explicit flag is validated against
// the options and used as-is; a single option is auto-selected; a
// non-interactive stdin/stderr uses the resolved default without prompting;
// otherwise it shows the interactive picker. noun is used in messages
// ("stack"); label heads the picker. itemize, when non-nil, builds the picker
// row for each option (e.g. a human-friendly stack title, or a machine-type
// spec hint); the raw option string is always what's returned. options must be
// non-empty.
func chooseOne(ctx context.Context, cmd *cobra.Command, log *stepLogger, noun, label string, options []string, prefName, backendDefault, flag string, itemize func(string) picker.Item) (string, error) {
	if v, ok, err := resolveWithoutPrompt(cmd, log, noun, options, prefName, backendDefault, flag); ok || err != nil {
		return v, err
	}
	defaultIdx := resolveDefault(options, prefName, backendDefault)
	// Surface the default at the top of the list so it's the obvious
	// press-Enter choice, with the cursor and "(default)" badge on row 1.
	ordered := moveToFront(options, defaultIdx)
	items := make([]picker.Item, len(ordered))
	for i, opt := range ordered {
		if itemize != nil {
			items[i] = itemize(opt)
		} else {
			items[i] = picker.Item{Title: opt}
		}
	}
	idx, err := picker.Select(ctx, picker.Config{
		Prompt:     label,
		Items:      items,
		Cursor:     0,
		DefaultIdx: 0,
		In:         os.Stdin,
		Out:        cmd.ErrOrStderr(),
	})
	if errors.Is(err, picker.ErrCancelled) {
		return "", errSelectionCancelled
	}
	if err != nil {
		return "", err
	}
	return ordered[idx], nil
}

// osDisplayOrder is the order operating systems are offered in the OS prompt:
// macOS first, then Linux. Any other OS the backend reports is appended after
// these, in first-seen order.
var osDisplayOrder = []string{"macos", "linux"}

// stackStatusOrder groups the stack picker: recommended stable first, then
// bleeding-edge, most conservative last. A stack whose status isn't one of
// these is shown after them, under its own header, so no stack is ever hidden.
var stackStatusOrder = []string{"stable", "edge", "frozen"}

// chooseStack resolves the stack. For an explicit --stack flag, a single stack,
// or a non-terminal it behaves exactly like chooseOne (default over the whole
// catalog). Interactively it first asks for an operating system when more than
// one is available, then shows that OS's stacks grouped by status
// (stable → edge → frozen) with non-selectable dividers — turning a long, mixed
// catalog into a short, ordered list. It returns the chosen stack id.
func chooseStack(ctx context.Context, cmd *cobra.Command, log *stepLogger, ids []string, byID map[string]internalrde.Stack, prefStack, backendDefault, flagStack string) (string, error) {
	if v, ok, err := resolveWithoutPrompt(cmd, log, "stack", ids, prefStack, backendDefault, flagStack); ok || err != nil {
		return v, err
	}
	// Narrow to a single OS first (a no-op when the catalog is single-OS), then
	// pick from that OS's stacks grouped by status.
	osIDs, err := chooseStackOS(ctx, cmd, ids, byID, prefStack, backendDefault)
	if err != nil {
		return "", err
	}
	return pickStackGrouped(ctx, cmd, osIDs, byID, prefStack, backendDefault)
}

// chooseStackOS narrows the stack ids to a single operating system. It returns
// the ids unchanged (no prompt) when the catalog has at most one OS, or when
// any stack lacks an OS tag — in that case an OS filter could hide a stack, so
// we skip it and let the grouped picker show everything. Otherwise it shows a
// small "macOS / Linux" picker with the cursor on the OS of the would-be
// default stack, and returns the ids for the chosen OS.
func chooseStackOS(ctx context.Context, cmd *cobra.Command, ids []string, byID map[string]internalrde.Stack, prefStack, backendDefault string) ([]string, error) {
	oses := stackOSes(ids, byID)
	if len(oses) <= 1 || anyStackMissingOS(ids, byID) {
		return ids, nil
	}
	defOS := byID[defaultStackID(ids, prefStack, backendDefault)].OS
	cursor := indexOf(oses, defOS)
	if cursor < 0 {
		cursor = 0
	}
	items := make([]picker.Item, len(oses))
	for i, osName := range oses {
		items[i] = picker.Item{Title: cmdutil.OSDisplayName(osName)}
	}
	idx, err := picker.Select(ctx, picker.Config{
		Prompt:     "Select an operating system",
		Items:      items,
		Cursor:     cursor,
		DefaultIdx: -1,
		In:         os.Stdin,
		Out:        cmd.ErrOrStderr(),
	})
	if errors.Is(err, picker.ErrCancelled) {
		return nil, errSelectionCancelled
	}
	if err != nil {
		return nil, err
	}
	return filterStacksByOS(ids, byID, oses[idx]), nil
}

// pickStackGrouped shows the (already OS-filtered) stacks grouped by status with
// non-selectable dividers, and returns the chosen stack id. The cursor opens on
// the would-be default stack; the "(default)" badge is shown only when that
// default is a genuine saved-pref or backend default present in this list (not
// the first-item fallback that kicks in on the non-default OS). ids must be
// non-empty.
func pickStackGrouped(ctx context.Context, cmd *cobra.Command, ids []string, byID map[string]internalrde.Stack, prefStack, backendDefault string) (string, error) {
	items, stackAt := buildGroupedStackItems(ids, byID)
	cursor := indexOf(stackAt, defaultStackID(ids, prefStack, backendDefault))
	defaultIdx := -1
	if indexOf(ids, prefStack) >= 0 || indexOf(ids, backendDefault) >= 0 {
		defaultIdx = cursor // a real default is in this list — badge it
	}
	idx, err := picker.Select(ctx, picker.Config{
		Prompt:     "Select a stack",
		Items:      items,
		Cursor:     cursor,
		DefaultIdx: defaultIdx,
		In:         os.Stdin,
		Out:        cmd.ErrOrStderr(),
	})
	if errors.Is(err, picker.ErrCancelled) {
		return "", errSelectionCancelled
	}
	if err != nil {
		return "", err
	}
	return stackAt[idx], nil
}

// stackGroup is a run of stack ids that share a status, kept in catalog order.
type stackGroup struct {
	status string
	ids    []string
}

// groupStacksByStatus buckets ids by status and returns the buckets in
// stackStatusOrder (stable → edge → frozen), with any other statuses appended
// in first-seen order. Empty buckets are omitted, and order within a bucket is
// the input (catalog) order. ids should already be filtered to a single OS.
func groupStacksByStatus(ids []string, byID map[string]internalrde.Stack) []stackGroup {
	byStatus := make(map[string][]string)
	var firstSeen []string
	for _, id := range ids {
		st := byID[id].Status
		if _, ok := byStatus[st]; !ok {
			firstSeen = append(firstSeen, st)
		}
		byStatus[st] = append(byStatus[st], id)
	}
	var groups []stackGroup
	emitted := make(map[string]bool)
	for _, s := range stackStatusOrder {
		if g := byStatus[s]; len(g) > 0 {
			groups = append(groups, stackGroup{status: s, ids: g})
			emitted[s] = true
		}
	}
	for _, s := range firstSeen {
		if !emitted[s] {
			groups = append(groups, stackGroup{status: s, ids: byStatus[s]})
		}
	}
	return groups
}

// buildGroupedStackItems renders the grouped stack rows for the picker: one
// non-selectable divider per status group followed by that group's stack rows.
// It returns the picker items alongside a parallel stackAt slice mapping each
// item index to its stack id ("" for divider rows), so the caller can map the
// chosen index back to a stack id.
func buildGroupedStackItems(ids []string, byID map[string]internalrde.Stack) (items []picker.Item, stackAt []string) {
	groups := groupStacksByStatus(ids, byID)
	items = make([]picker.Item, 0, len(ids)+len(groups))
	stackAt = make([]string, 0, len(ids)+len(groups))
	for _, g := range groups {
		items = append(items, picker.Item{Title: groupHeader(g.status), Divider: true})
		stackAt = append(stackAt, "")
		for _, id := range g.ids {
			items = append(items, picker.Item{Title: stackTitle(byID, id), Desc: stackGroupedSecondary(byID[id])})
			stackAt = append(stackAt, id)
		}
	}
	return items, stackAt
}

// groupHeader is the divider label for a status group: the status with its
// first letter capitalized ("edge" → "Edge"), or "Other" when the backend left
// the status blank.
func groupHeader(status string) string {
	if status == "" {
		return "Other"
	}
	return strings.ToUpper(status[:1]) + status[1:]
}

// stackGroupedSecondary is the dim secondary text for a stack row inside a
// status group: "<OS> <version>" (e.g. "macOS 26"). The status is dropped —
// the group's divider already conveys it — while the OS and version are kept
// because they differ row to row.
func stackGroupedSecondary(st internalrde.Stack) string {
	osPart := cmdutil.OSDisplayName(st.OS)
	if osPart == "" {
		return ""
	}
	if st.OSVersion > 0 {
		return osPart + " " + strconv.Itoa(int(st.OSVersion))
	}
	return osPart
}

// stackOSes returns the distinct operating systems present among the stacks,
// ordered with the well-known ones first (macOS, then Linux) and any others in
// first-seen order. Stacks with no OS tag are ignored here — anyStackMissingOS
// is what decides whether to skip the OS prompt for them.
func stackOSes(ids []string, byID map[string]internalrde.Stack) []string {
	present := make(map[string]bool)
	var firstSeen []string
	for _, id := range ids {
		osName := byID[id].OS
		if osName == "" || present[osName] {
			continue
		}
		present[osName] = true
		firstSeen = append(firstSeen, osName)
	}
	var ordered []string
	emitted := make(map[string]bool)
	for _, osName := range osDisplayOrder {
		if present[osName] {
			ordered = append(ordered, osName)
			emitted[osName] = true
		}
	}
	for _, osName := range firstSeen {
		if !emitted[osName] {
			ordered = append(ordered, osName)
		}
	}
	return ordered
}

// anyStackMissingOS reports whether some stack has no OS tag. When true the OS
// prompt is skipped, because filtering by OS would make the untagged stack
// unreachable.
func anyStackMissingOS(ids []string, byID map[string]internalrde.Stack) bool {
	for _, id := range ids {
		if byID[id].OS == "" {
			return true
		}
	}
	return false
}

// filterStacksByOS returns the ids whose stack runs the given OS, in input order.
func filterStacksByOS(ids []string, byID map[string]internalrde.Stack, osName string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if byID[id].OS == osName {
			out = append(out, id)
		}
	}
	return out
}

// defaultStackID returns the stack id that would be preselected — saved pref,
// else backend default, else the first — used to seed the OS prompt's cursor
// and the grouped picker's cursor. ids must be non-empty.
func defaultStackID(ids []string, prefStack, backendDefault string) string {
	return ids[resolveDefault(ids, prefStack, backendDefault)]
}

// reuseAction is which entry of the "use your last setup" menu was picked.
type reuseAction int

const (
	reuseUse reuseAction = iota
	reuseChangeStack
	reuseChangeMachine
)

// offerReuse shows the one-step "use your last setup" menu for a project that
// has a remembered stack+machine combo. done=true means the combo was resolved
// here (reused, or customized through the menu) and the returned values are
// final; done=false with a nil error means the remembered combo is stale or the
// user asked to change the stack, so the caller should run the full pick. A
// non-nil error (e.g. the user cancelled) aborts selection.
func offerReuse(ctx context.Context, cmd *cobra.Command, svc *internalrde.Service, log *stepLogger, workspaceID string, stackIDs []string, byID map[string]internalrde.Stack, prefs localsession.Prefs) (stack, machineType, machineLbl string, done bool, err error) {
	// The remembered stack must still be offered…
	if indexOf(stackIDs, prefs.Stack) < 0 {
		return "", "", "", false, nil
	}
	mts, err := svc.MachineTypesForStack(ctx, workspaceID, prefs.Stack)
	if err != nil {
		return "", "", "", false, err
	}
	mtNames, _ := uniqueMachineTypeNames(mts)
	// …and the remembered machine type must still be compatible with it.
	if indexOf(mtNames, prefs.MachineType) < 0 {
		return "", "", "", false, nil
	}
	mtByName := indexMachineTypes(mts)

	items, actions := buildReuseMenu(len(stackIDs) > 1, len(mtNames) > 1)
	// Nothing to customize (a single stack and a single machine type): reuse
	// without prompting.
	if len(actions) == 1 {
		return prefs.Stack, prefs.MachineType, machineLabel(mtByName[prefs.MachineType]), true, nil
	}

	idx, err := picker.Select(ctx, picker.Config{
		Prompt:     "Last used for this project",
		Note:       reuseDetail(stackTitle(byID, prefs.Stack), machineDisplayName(mtByName[prefs.MachineType]), machineSpec(mtByName[prefs.MachineType])),
		Items:      items,
		Cursor:     0,
		DefaultIdx: 0,
		In:         os.Stdin,
		Out:        cmd.ErrOrStderr(),
	})
	if errors.Is(err, picker.ErrCancelled) {
		return "", "", "", false, errSelectionCancelled
	}
	if err != nil {
		return "", "", "", false, err
	}

	switch actions[idx] {
	case reuseChangeStack:
		// Caller runs the full stack + machine pick.
		return "", "", "", false, nil
	case reuseChangeMachine:
		mt, err := chooseOne(ctx, cmd, log, "machine type", "Select a machine type", mtNames, prefs.MachineType, "", "", machineItem(mtByName))
		if err != nil {
			return "", "", "", false, err
		}
		return prefs.Stack, mt, machineLabel(mtByName[mt]), true, nil
	default: // reuseUse
		return prefs.Stack, prefs.MachineType, machineLabel(mtByName[prefs.MachineType]), true, nil
	}
}

// buildReuseMenu assembles the rows and matching actions for offerReuse. The
// "Change stack" row appears only when more than one stack exists, and "Change
// machine type" only when the remembered stack has more than one compatible
// type — so we never offer a change with nothing to change.
func buildReuseMenu(multiStack, multiMachine bool) ([]picker.Item, []reuseAction) {
	items := []picker.Item{{Title: "Use this setup"}}
	actions := []reuseAction{reuseUse}
	if multiStack {
		items = append(items, picker.Item{Title: "Change stack"})
		actions = append(actions, reuseChangeStack)
	}
	if multiMachine {
		items = append(items, picker.Item{Title: "Change machine type"})
		actions = append(actions, reuseChangeMachine)
	}
	return items, actions
}

// reuseDetail is the two-line "Stack / Machine type" summary shown under the
// reuse-menu prompt, so the user can see exactly what "Use this setup" launches.
func reuseDetail(stackTitle, machineDisplay, machineSpec string) string {
	machine := machineDisplay
	if machineSpec != "" {
		machine += "  (" + machineSpec + ")"
	}
	return fmt.Sprintf("  %-13s %s\n  %-13s %s", "Stack", stackTitle, "Machine type", machine)
}

// machineItem builds the picker rows for the machine-type picker: machine types
// show their friendly title (or raw name) with the specs as dim secondary text.
// The picker still returns the raw option string (the machine type name).
func machineItem(byName map[string]internalrde.MachineType) func(string) picker.Item {
	return func(name string) picker.Item {
		mt, ok := byName[name]
		if !ok {
			return picker.Item{Title: name, Desc: machineSpecHint(name)}
		}
		return picker.Item{Title: machineDisplayName(mt), Desc: machineSpec(mt)}
	}
}

// indexMachineTypes maps machine-type name to its (first-seen) record, so the
// picker and reuse summary can look up the backend's friendly title/cpu/ram.
func indexMachineTypes(mts []internalrde.MachineType) map[string]internalrde.MachineType {
	out := make(map[string]internalrde.MachineType, len(mts))
	for _, mt := range mts {
		if _, ok := out[mt.Name]; !ok {
			out[mt.Name] = mt
		}
	}
	return out
}

// machineDisplayName returns the backend's friendly machine-type title, falling
// back to the raw name when none is provided.
func machineDisplayName(mt internalrde.MachineType) string {
	if mt.Title != "" {
		return mt.Title
	}
	return mt.Name
}

// machineLabel is the one-line, human-friendly machine-type label for
// confirmations — the friendly name plus its specs in parentheses, e.g.
// "M2 Pro Large (12 vCPU · 28 GB)".
func machineLabel(mt internalrde.MachineType) string {
	name := machineDisplayName(mt)
	if spec := machineSpec(mt); spec != "" {
		return name + " (" + spec + ")"
	}
	return name
}

// machineSpec returns the "<cpu> · <ram>" display, preferring the backend's
// structured fields and falling back to the spec parsed from the name.
func machineSpec(mt internalrde.MachineType) string {
	parts := make([]string, 0, 2)
	if mt.CPU != "" {
		parts = append(parts, mt.CPU)
	}
	if mt.RAM != "" {
		parts = append(parts, mt.RAM)
	}
	if len(parts) > 0 {
		return strings.Join(parts, " · ")
	}
	return machineSpecHint(mt.Name)
}

// stackTitle returns the human-friendly title for a stack id, falling back to
// the raw id when the backend supplied no title.
func stackTitle(byID map[string]internalrde.Stack, id string) string {
	if st, ok := byID[id]; ok && st.Title != "" {
		return st.Title
	}
	return id
}

// interactivePicker reports whether an interactive picker can run: it reads keys
// from stdin and draws to stderr, so both must be a terminal.
func interactivePicker(cmd *cobra.Command) bool {
	return cmdutil.IsTerminal(os.Stdin) && cmdutil.IsTerminalWriter(cmd.ErrOrStderr())
}

// machineSpecRe matches the "<vCPU>c-<RAM>g" tail of a machine-type name.
var machineSpecRe = regexp.MustCompile(`^(\d+)c-(\d+)g$`)

// machineSpecHint derives a "<n> vCPU · <m> GB" hint from a machine-type name
// by parsing its last '.'-separated segment (e.g. "g2.mac.m2pro.4c-6g" →
// "4c-6g" → "4 vCPU · 6 GB"). Returns "" when the segment doesn't match the
// "<n>c-<m>g" shape (e.g. "g2.mac"), so an unrecognized name simply shows no
// hint rather than a wrong one.
func machineSpecHint(name string) string {
	seg := name
	if i := strings.LastIndex(name, "."); i >= 0 {
		seg = name[i+1:]
	}
	m := machineSpecRe.FindStringSubmatch(seg)
	if m == nil {
		return ""
	}
	return fmt.Sprintf("%s vCPU · %s GB", m[1], m[2])
}

// moveToFront returns options with the element at idx moved to the front,
// preserving the relative order of the rest. idx must be a valid index; idx 0
// (or out of range) returns options unchanged.
func moveToFront(options []string, idx int) []string {
	if idx <= 0 || idx >= len(options) {
		return options
	}
	out := make([]string, 0, len(options))
	out = append(out, options[idx])
	out = append(out, options[:idx]...)
	out = append(out, options[idx+1:]...)
	return out
}

// resolveDefault returns the index in names to preselect, applying the
// precedence saved-pref → backend default → first. A pref or backend default
// absent from names is skipped, so the result is always a valid index (names
// must be non-empty). This is also what makes a saved/backend-default machine
// type that isn't compatible with the chosen stack fall back to first-available.
func resolveDefault(names []string, prefName, backendDefaultName string) int {
	if i := indexOf(names, prefName); i >= 0 {
		return i
	}
	if i := indexOf(names, backendDefaultName); i >= 0 {
		return i
	}
	return 0
}

// indexOf returns the index of target in names, or -1. An empty target never
// matches (it's the "unset" sentinel for prefs / backend defaults).
func indexOf(names []string, target string) int {
	if target == "" {
		return -1
	}
	for i, n := range names {
		if n == target {
			return i
		}
	}
	return -1
}

// uniqueStacks returns the stack ids in catalog order with duplicates removed,
// a lookup from stack id to its record, and the first id the backend flagged as
// default ("" if none).
func uniqueStacks(stacks []internalrde.Stack) (ids []string, byID map[string]internalrde.Stack, backendDefault string) {
	byID = make(map[string]internalrde.Stack, len(stacks))
	seen := make(map[string]bool, len(stacks))
	for _, st := range stacks {
		if !seen[st.ID] {
			seen[st.ID] = true
			ids = append(ids, st.ID)
			byID[st.ID] = st
		}
		if st.IsDefault && backendDefault == "" {
			backendDefault = st.ID
		}
	}
	return ids, byID, backendDefault
}

// uniqueMachineTypeNames returns the machine type names in catalog order with
// duplicates removed (a name can be offered by several clusters), plus the first
// name the backend flagged as default ("" if none).
func uniqueMachineTypeNames(mts []internalrde.MachineType) (names []string, backendDefault string) {
	seen := make(map[string]bool, len(mts))
	for _, mt := range mts {
		if !seen[mt.Name] {
			seen[mt.Name] = true
			names = append(names, mt.Name)
		}
		if mt.IsDefault && backendDefault == "" {
			backendDefault = mt.Name
		}
	}
	return names, backendDefault
}
