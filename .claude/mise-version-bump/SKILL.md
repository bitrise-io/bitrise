---
name: mise-version-bump
description: >
  Bumps the pinned Mise version in toolprovider/mise/mise.go. Use when the user
  wants to update, bump, or upgrade Mise, or change misePreviewVersion /
  miseStableVersion / their checksum maps.
---

# Mise version bump

Automates bumping `misePreviewVersion` / `miseStableVersion` and their checksum maps in `toolprovider/mise/mise.go`. Fail-fast on checksum mismatch or missing GCS objects. Ask the user twice: once for track/action, once to confirm the PR.

## Repo facts

Before starting, read `toolprovider/mise/mise.go` and `toolprovider/mise/bootstrap.go` to discover:
- The version constant names and checksum map names (preview and stable).
- The set of platform keys used in the checksum maps.
- The artifact naming pattern, GCS bucket name, and GCS path structure.

Use these values throughout — do not assume them from memory.

- **Source repo:** `jdx/mise`. Valid releases: `vYYYY.M.D` (CalVer only with optional `: <description>`).
- **Commit/PR title:** `Update Mise version`. **Branch:** `bump/mise-<NEW_TAG>` (or `bump/mise-stable-<TAG>` for match-stable).

## Step 0: Preflight (run in parallel)

- Confirm `go.mod` contains `module github.com/bitrise-io/bitrise/v2`. If not, ask the user to `cd` to the repo root and stop.
- Working tree must be clean. If dirty, ask: continue or abort (recommend abort).
- Verify `gh`, `gcloud`, `shasum`, and `git` are on PATH. List all missing and stop if any.
- Verify an active gcloud account exists. If not, instruct the user to run `gcloud auth login` and stop.
- Verify access to the GCS bucket discovered from `bootstrap.go`. If access fails, surface the error and stop — do not proceed to download or upload.

## Step 1: Read current state

Read `toolprovider/mise/mise.go` and extract the current preview and stable version strings. Print both.

## Step 2: Discover latest release

Fetch the latest release tag from `jdx/mise`. Validate it is a CalVer tag (e.g. `v2025.1.0`). Print it. Stop on failure.

## Step 3: Decide track(s) to bump

**Q1 - Which track?**
- **(a)** Preview
- **(b)** Stable
- **(c)** Both

**Q2 - What action?**
- **(a)** Update to `<latest>`
- **(b)** Repair current version - re-downloads, re-verifies, re-uploads missing GCS objects, overwrites checksums in `mise.go`; no version string change
- **(c)** Match stable with current preview - no download *(only if Q1 = b or c)*
- **(d)** Update preview to `<latest>`, stable to current preview *(only if Q1 = c)*

| State | Q1 | Q2 |
|---|---|---|
| `preview == stable == latest` | - (skip) | b |
| `preview == stable < latest` | a, b, c | a, b |
| `preview > stable` AND `preview == latest` | a, b, c | a, b, c, d |
| `preview > stable` AND `latest > preview` | a, b, c | a, b, c, d |
| `preview < stable` | Unexpected - surface values. Ask: update preview to match stable (Q1=a, Q2=a), repair (Q2=b), or abort. |

Record `targetVersion` and `tracks` (`{preview}`, `{stable}`, `{preview,stable}`, `{stable-match-preview}`).

**For `{stable-match-preview}` (Q1=b or c, Q2=c): skip steps 4–7 and go directly to step 8.**

## Step 4: Changelog review

Range: `(fromTag, toTag]` where `fromTag` = older of the tracks being moved, `toTag` = `targetVersion`.

For each CalVer release in range, fetch its details from `jdx/mise`.

Scan `body` (case-insensitive) for keywords related to declarative tool setup:

```
registry  lockfile  asdf       env        shim       mirror
tarball   tools     core       idiomatic  mise.toml  direnv
backend   tool-versions        breaking   deprecat   remove
```

Also scan for mise core tools and popular mobile dev tools.

Capture matching lines verbatim, attributed to their release tag and URL.

**Risk level** (always print with one-line justification):
- **low** - baseline.
- **medium** - ≥3 releases in range, OR any match outside breaking/deprecat/remove keywords.
- **high** - "breaking change" mentioned, OR a CLI subcommand we invoke is removed/renamed. Check `toolprovider/mise/` for all mise subcommands invoked and verify none are removed or renamed.
- **extreme** - asdf compatibility removed, registry format changed, or tarball layout changed.

**Related issues** (best effort): for matched lines that look like bugs, search for related issues in `bitrise-io/bitrise`. Surface hits as `Possibly related: bitrise-io/bitrise#<num> - <title>`. Silent on no match.

## Step 5: Download artifacts

Create a temp dir and use `gh api` to fetch one artifact per platform key (using the artifact naming pattern from the repo facts step) plus the checksums file.

Verify the temp dir contains exactly one artifact per platform + the checksums file. If anything is missing, delete the temp dir and stop.

## Step 6: Verify checksums (hard fail on mismatch)

Compute checksum for each artifact, infer algorithm from filename or content. Compare against the checksums file.

**On any mismatch:** print the mismatching platforms with expected vs actual checksums, delete the temp dir, and STOP: do not edit files or touch GCS.

**On full match:** save checksums as `computedChecksums[platform]` and continue.

## Step 7: GCS mirror

Using the GCS bucket and path structure from the repo facts step, for each artifact:
1. Check if the object already exists in GCS.
2. If absent → upload. If present → log "already mirrored", skip.

Then list the version prefix to confirm all platform artifacts are present. If any are missing → delete the temp dir and STOP, do not edit files.

After successful verification, delete the temp dir.

## Step 8: Edit `toolprovider/mise/mise.go`

**For `{stable-match-preview}`: this is where execution resumes after skipping steps 4–7.**

Update the relevant version constant(s) and checksum map(s) using the Edit tool, including enough surrounding context to unambiguously target the preview or stable block. For `{stable-match-preview}`, copy current preview values into the stable block.

## Step 9: PR

Ask: "Create the PR now?"

**Yes:** commit, push, and open a PR. Print the PR URL.

**No:** leave edits in working tree.

## Step 10: Final report

Print a brief markdown summary covering: what changed (tracks and versions), risk level with justification, relevant changelog highlights, and PR URL or status.
