# Refactor: Simplify Version Resolution in toolprovider/mise

## Problem

The `toolprovider/mise` package had accumulated complexity due to **semantic overloading of the version string parameter**. The `UnparsedVersion` field was used to represent:
- Concrete versions (`"20.19.3"`)
- Fuzzy/partial versions (`"20"`, `"3.3"`)
- Keywords (`"latest"`, `"installed"`)
- Empty strings (special semantics)

This caused:
- Scattered conditional logic checking for keywords throughout the codebase
- Version strings overriding `ResolutionStrategy` enum values
- Multi-stage resolution with hidden fallback chains
- Complex decision trees in functions like `versionExists()`, `isAlreadyInstalled()`, `miseVersionString()`
- Difficulty testing and reasoning about the code

## Solution

Introduced **early resolution with strategy normalization** to separate concerns into distinct phases:

```
Phase 1: Normalize Strategy
  └─ Convert unresolvable strategies (LatestInstalled + nothing installed) → LatestReleased

Phase 2: Resolve to Concrete Version
  └─ Always succeeds after normalization (or fails fast if version doesn't exist)

Phase 3: Check & Install
  └─ Simple logic, no fallbacks needed
```

**Key insight**: After strategy normalization, we can ALWAYS resolve to a concrete version before installation. This eliminates the circular dependency between checking what's installed and deciding what to install.

## Changes Made

### New Functions (resolve.go)

1. **`normalizeStrategy()`** - Handles the case where `LatestInstalled` strategy cannot be satisfied and normalizes it to `LatestReleased`. Returns the normalized strategy and whether it was changed.

2. **`resolveToConcreteVersion()`** - Unified resolution entry point that resolves any version input to a concrete version string. Both `Strict` and `LatestReleased` strategies use the same resolution logic (allowing fuzzy versions), while `LatestInstalled` checks installed versions.

3. **`versionExistsRemote()`** - Extracted remote version check without keyword handling.

### Refactored Main Flow (mise.go)

**Old flow:**
```go
1. Install plugin
2. Check if already installed (requires resolution internally)
3. Check if version exists (requires keyword handling)
4. Install
5. Resolve to concrete (post-install)
```

**New flow:**
```go
1. Install plugin
2. Handle "installed" keyword (convert to empty version + LatestInstalled strategy)
3. Normalize strategy (LatestInstalled with nothing installed → LatestReleased)
4. Resolve to concrete version (always before installation)
5. Check if version exists remotely (for better error messages)
6. Check if concrete version already installed
7. Install concrete version if needed
8. Verify installation succeeded
```

### Simplified Functions

- **`miseVersionString()`**: **30 lines → 3 lines** - Simply formats `toolName@concreteVersion`
- **`isAlreadyInstalled()`**: **45 lines → 3 lines** - Simple lookup with concrete version
- **`installToolVersion()`**: Updated signature to take `toolName` and `concreteVersion` instead of full `ToolRequest`

### Removed Code

- Deleted old `versionExists()` function with keyword handling (~30 lines)
- Removed `resolveToConcreteVersionAfterInstall()` wrapper
- Removed unused method wrappers (`resolveToLatestReleased()`, `resolveToLatestInstalled()`, `versionExists()` methods)
- Cleaned up redundant `TestVersionExists` test

## Key Discoveries

### Strict Strategy with Fuzzy Versions

During integration testing, we discovered that **`Strict` strategy is intended to work with fuzzy versions** (e.g., `java@21` → `21.0.2`). The strategy name is somewhat misleading:
- `Strict` doesn't mean "use exact version string without resolution"
- It means "don't fallback to LatestReleased if LatestInstalled fails"
- Both `Strict` and `LatestReleased` allow mise to resolve fuzzy versions

This is a valid feature that users rely on, so we preserved this behavior by having both strategies use `resolveToLatestReleased()`.

### "installed" Keyword Handling

The literal string `"installed"` as a version is a special keyword meaning "use latest installed, or install latest released if nothing is installed". This is now handled explicitly at the start of `InstallTool()` by converting it to an empty version with `LatestInstalled` strategy.

## Benefits

1. **Centralized complexity**: All keyword and fuzzy version handling in two functions (`normalizeStrategy` + `resolveToConcreteVersion`)

2. **No scattered conditionals**: Functions no longer check for "installed", "latest", or empty strings throughout the codebase

3. **Explicit decisions**: The fallback from "latest installed" to "latest released" is explicit and logged, not hidden in error handling

4. **Better separation of concerns**:
   - Strategy normalization: decides WHAT approach to use
   - Resolution: converts fuzzy/keywords to concrete
   - Installation: executes the decision

5. **Easier testing**: Each phase can be tested independently
   - Test normalization logic separately
   - Test resolution with mocked mise responses
   - Test installation with concrete versions only

6. **Better error messages**: Failures happen at clear phase boundaries with specific context. Resolution failures now properly return `ToolInstallError` with descriptive messages.


