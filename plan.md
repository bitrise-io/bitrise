# Implementation Plan: Simplify Version Resolution in toolprovider/mise

## Problem Statement

The `toolprovider/mise` package has accumulated complexity due to **semantic overloading of the version string parameter**. The `UnparsedVersion` field is used to represent:

1. Concrete versions (`"20.19.3"`)
2. Fuzzy/partial versions (`"20"`, `"3.3"`)
3. Keywords (`"latest"`, `"installed"`)
4. Empty strings (special semantics)

This causes:
- Scattered conditional logic checking for keywords throughout the codebase
- Version strings overriding `ResolutionStrategy` enum values
- Multi-stage resolution with hidden fallback chains
- Complex decision trees in functions like `versionExists()`, `isAlreadyInstalled()`, `miseVersionString()`
- Difficulty testing and reasoning about the code

## Solution Approach

Introduce **early resolution with strategy normalization** to separate concerns into distinct phases:

```
Phase 1: Normalize Strategy
  └─ Convert unresolvable strategies (LatestInstalled + nothing installed) → LatestReleased

Phase 2: Resolve to Concrete Version
  └─ Always succeeds after normalization (or fails fast if version doesn't exist)

Phase 3: Check & Install
  └─ Simple logic, no fallbacks needed
```

**Key insight**: After strategy normalization, we can ALWAYS resolve to a concrete version before installation. This eliminates the circular dependency between checking what's installed and deciding what to install.

## Design Changes

### 1. New Types (in `resolve.go`)

```go
// VersionResolution represents the result of resolving a user-provided version string
type VersionResolution struct {
    // Concrete is the exact version number (e.g., "20.18.1")
    // This is what will actually be installed
    Concrete string

    // WasNormalized indicates if the strategy was changed during normalization
    // Useful for logging/debugging
    WasNormalized bool
}
```

**Reasoning**: Explicit type makes the resolution result clear and self-documenting. The `WasNormalized` flag helps with debugging and can be logged to explain to users why their "latest installed" request became "latest released".

### 2. New Function: Strategy Normalization (in `resolve.go`)

```go
// normalizeStrategy handles the case where LatestInstalled strategy cannot be satisfied
// Returns the normalized strategy and whether it was changed
func normalizeStrategy(
    execEnv execenv.ExecEnv,
    tool provider.ToolRequest,
) (provider.ResolutionStrategy, bool, error)
```

**Logic**:
1. If strategy is NOT `LatestInstalled`, return as-is
2. If strategy IS `LatestInstalled`:
   - Try `mise latest --installed tool@version`
   - If succeeds (something is installed): return `LatestInstalled`, false
   - If fails with `errNoMatchingVersion`: return `LatestReleased`, true
   - If other error: return error

**Reasoning**: This centralizes the "nothing installed → fall back to latest released" decision that's currently scattered across `miseVersionString()` and other functions. The boolean return indicates whether normalization occurred, which is useful for logging.

### 3. New Function: Unified Resolution (in `resolve.go`)

```go
// resolveToConcreteVersion resolves any version input to a concrete version string
// Assumes strategy has already been normalized via normalizeStrategy
func resolveToConcreteVersion(
    execEnv execenv.ExecEnv,
    toolName provider.ToolID,
    version string,
    strategy provider.ResolutionStrategy,
) (string, error)
```

**Logic**:
```
switch strategy:
  case Strict:
    return version  // Already concrete
  case LatestReleased:
    return resolveToLatestReleased(toolName, version)
  case LatestInstalled:
    return resolveToLatestInstalled(toolName, version)
    // Note: This should always succeed because we normalized
```

**Reasoning**: Single entry point for resolution eliminates duplication. The precondition (normalized strategy) is documented and should be enforced by the caller. This function is pure logic with no hidden fallbacks.

### 4. Refactor: Simplified `miseVersionString()` (in `install.go`)

**Current complexity** (lines 153-183):
- 30 lines
- Special handling for "installed" keyword overriding strategy
- Fallback chain when LatestInstalled fails
- Multiple error paths

**New implementation**:
```go
func miseVersionString(toolName provider.ToolID, concreteVersion string) string {
    return fmt.Sprintf("%s@%s", toolName, concreteVersion)
}
```

**Reasoning**: Once we have a concrete version, generating the mise command string is trivial. All complexity moves to the resolution phase where it belongs. This also makes it obvious what we're installing.

### 5. Refactor: Simplified `isAlreadyInstalled()` (in `install.go`)

**Current complexity** (lines 107-151):
- Needs to resolve fuzzy versions before checking
- Has its own resolver functions as parameters
- 45 lines of logic

**New implementation**:
```go
func (m *MiseToolProvider) isAlreadyInstalled(
    toolName provider.ToolID,
    concreteVersion string,
) (bool, error) {
    // Simply check if the concrete version exists locally
    return versionExistsLocal(m.ExecEnv, toolName, concreteVersion)
}
```

**Reasoning**: When working with concrete versions only, checking installation status becomes a simple lookup. No resolution needed, no fallback logic.

### 6. Refactor: Simplified `versionExists()` (in `resolve.go`)

**Current complexity** (lines 123-154):
- Special cases for "installed" keyword
- Special cases for "latest" keyword
- Fallback from local to remote for "installed"

**New approach**:
The keyword handling moves to `normalizeStrategy()` and `resolveToConcreteVersion()`. This function becomes:

```go
// versionExistsRemote checks if a version exists in the remote registry
// version can be fuzzy (e.g., "20") or concrete (e.g., "20.18.1")
func versionExistsRemote(
    execEnv execenv.ExecEnv,
    toolName provider.ToolID,
    version string,
) (bool, error)
```

**Reasoning**: Separate local and remote checks explicitly. No keyword handling needed—callers use the right function for their needs. The "installed" keyword fallback logic is eliminated because strategy normalization handles it.

### 7. Refactor: Main Flow in `InstallTool()` (in `mise.go`)

**Current flow** (lines 105-154):
```
1. Install plugin
2. Check if already installed (requires resolution internally)
3. Check if version exists (requires keyword handling)
4. Install
5. Resolve to concrete (post-install)
```

**New flow**:
```go
func (m *MiseToolProvider) InstallTool(tool provider.ToolRequest) (provider.ToolInstallResult, error) {
    // 1. Install plugin (unchanged)
    useNix := canBeInstalledWithNix(...)
    if !useNix {
        err := m.InstallPlugin(tool)
        // ...
    }
    installRequest := installRequest(tool, useNix)

    // 2. Normalize strategy
    normalizedStrategy, wasNormalized, err := normalizeStrategy(m.ExecEnv, installRequest)
    if err != nil {
        return provider.ToolInstallResult{}, err
    }
    if wasNormalized {
        log.Debugf("No installed versions found, falling back to latest released")
    }

    // 3. Resolve to concrete version (always succeeds after normalization)
    concreteVersion, err := resolveToConcreteVersion(
        m.ExecEnv,
        installRequest.ToolName,
        installRequest.UnparsedVersion,
        normalizedStrategy,
    )
    if err != nil {
        return provider.ToolInstallResult{}, fmt.Errorf("resolve version: %w", err)
    }
    log.Debugf("Resolved %s@%s to concrete version %s",
        installRequest.ToolName, installRequest.UnparsedVersion, concreteVersion)

    // 4. Check if concrete version already installed
    isAlreadyInstalled, err := m.isAlreadyInstalled(installRequest.ToolName, concreteVersion)
    if err != nil {
        return provider.ToolInstallResult{}, err
    }

    // 5. Install concrete version
    if !isAlreadyInstalled {
        err = m.installToolVersion(installRequest.ToolName, concreteVersion)
        if err != nil {
            return provider.ToolInstallResult{}, err
        }
    }

    // 6. Verify installation (sanity check)
    installedVersion, err := resolveToLatestInstalled(m.ExecEnv, installRequest.ToolName, concreteVersion)
    if err != nil || installedVersion != concreteVersion {
        return provider.ToolInstallResult{}, fmt.Errorf(
            "verification failed: expected %s, got %s", concreteVersion, installedVersion)
    }

    return provider.ToolInstallResult{
        ToolName:           installRequest.ToolName,
        IsAlreadyInstalled: isAlreadyInstalled,
        ConcreteVersion:    concreteVersion,
    }, nil
}
```

**Reasoning**:
- Linear flow: normalize → resolve → check → install → verify
- No hidden fallbacks or circular dependencies
- Each phase has a single responsibility
- Easy to add logging at each step
- The verification step at the end is a sanity check to catch any mise weirdness

### 8. Update: `installToolVersion()` Signature (in `install.go`)

**Current**:
```go
func (m *MiseToolProvider) installToolVersion(tool provider.ToolRequest) error
```

**New**:
```go
func (m *MiseToolProvider) installToolVersion(toolName provider.ToolID, concreteVersion string) error
```

**Reasoning**: This function should only install what it's told. It no longer needs the full ToolRequest with strategy because resolution has already happened. Simpler signature, clearer intent.

## Benefits

1. **Centralized complexity**: All keyword and fuzzy version handling in one place (`normalizeStrategy` + `resolveToConcreteVersion`)

2. **No scattered conditionals**: Functions like `versionExists`, `isAlreadyInstalled`, `miseVersionString` no longer check for "installed", "latest", empty strings

3. **Explicit decisions**: The fallback from "latest installed" to "latest released" is explicit and logged, not hidden in error handling

4. **Better separation of concerns**:
   - Strategy normalization: decides WHAT approach to use
   - Resolution: converts fuzzy/keywords to concrete
   - Installation: executes the decision

5. **Easier testing**: Each phase can be tested independently
   - Test normalization logic separately
   - Test resolution with mocked mise responses
   - Test installation with concrete versions only

6. **Better error messages**: Failures happen at clear phase boundaries with specific context

7. **Future-proof**: Adding new keywords or resolution strategies only requires changes in 1-2 functions

## Migration Path

### Phase 1: Add New Functions (Non-Breaking)
1. Add `normalizeStrategy()` in `resolve.go`
2. Add `resolveToConcreteVersion()` in `resolve.go`
3. Add `versionExistsRemote()` in `resolve.go` (rename current `versionExists`)
4. Add tests for new functions

### Phase 2: Refactor Main Flow
1. Update `InstallTool()` to use new flow
2. Simplify `miseVersionString()`
3. Simplify `isAlreadyInstalled()`
4. Update `installToolVersion()` signature

### Phase 3: Remove Old Code
1. Delete fallback logic in old `miseVersionString()`
2. Delete keyword handling in `versionExists()`
3. Update tests to use new approach
4. Remove unused helper functions if any

### Phase 4: Validate
1. Run full integration test suite
2. Test edge cases: "installed" with nothing installed, fuzzy versions, empty strings
3. Verify error messages are clear
4. Check performance (should be similar or better due to fewer redundant calls)

## Risks & Mitigations

**Risk**: Extra `mise latest --installed` call in normalization adds latency
- **Mitigation**: This call only happens for LatestInstalled strategy. It replaces the equivalent check that currently happens later in `miseVersionString()`, so it's not actually an extra call.

**Risk**: Breaking changes if other code depends on current behavior
- **Mitigation**: The provider interface (`InstallTool()`) signature doesn't change. Internal refactoring only. Behavior should be identical from the outside.

**Risk**: Edge cases we haven't considered
- **Mitigation**: Comprehensive test coverage including all keyword combinations. The test file `resolve_test.go` already has good coverage of inputs—adapt these tests to the new functions.

## Success Criteria

1. **Complexity reduction**:
   - `miseVersionString()`: 30 lines → ~5 lines
   - `isAlreadyInstalled()`: 45 lines → ~10 lines
   - `versionExists()`: 32 lines → ~15 lines (split into focused functions)

2. **No keyword checks outside resolution**:
   - Grep for `version == "installed"` or `version == "latest"` should only find them in `normalizeStrategy()` and `resolveToConcreteVersion()`

3. **Linear flow**: No fallback chains or circular dependencies in main install flow

4. **All tests pass**: Both unit tests and integration tests

5. **Clear logging**: Debug logs explain strategy normalization and resolution decisions
