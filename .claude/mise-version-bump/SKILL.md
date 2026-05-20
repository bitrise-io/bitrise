---
name: mise-version-bump
description: >
  Automates bumping the pinned Mise tool version in the bitrise-io/bitrise repo
  (toolprovider/mise/mise.go). Handles release discovery, changelog review,
  artifact download, sha256 verification, GCS mirror upload, file edits, and
  optional PR creation. Use when the user wants to "update Mise", "bump Mise",
  "upgrade mise version", "match stable with preview", or otherwise change the
  values of misePreviewVersion / miseStableVersion / their checksum maps.
---

# Mise version bump

Automates bumping `misePreviewVersion` / `miseStableVersion` and their checksum maps in `toolprovider/mise/mise.go`. Fail-fast on checksum mismatch or missing GCS objects. Ask the user twice: once for track/action, once to confirm the PR.

## Repo facts

- **Edit target:** `toolprovider/mise/mise.go`: `misePreviewVersion`, `misePreviewChecksums`, `miseStableVersion`, `miseStableChecksums`.
- **Platforms:** `linux-x64`, `linux-arm64`, `macos-x64`, `macos-arm64`.
- **Source repo:** `jdx/mise`. Valid releases: `vYYYY.M.D` (CalVer only with optional `: <description>`).
- **Artifact pattern:** `mise-v<VERSION>-<PLATFORM>.tar.gz` (e.g. tag `v2026.5.10` → `mise-v2026.5.10-linux-x64.tar.gz`).
- **GCS mirror:** `gs://mise-release-mirror/v<VERSION>/<ARTIFACT>`.
- **Commit/PR title:** `Update Mise version`. **Branch:** `bump/mise-<NEW_TAG>` (or `bump/mise-stable-<TAG>` for match-stable).

> If `toolprovider/mise/bootstrap.go` no longer contains `mise-release-mirror` or the four platform strings, warn and stop - these facts are stale.

## Step 0: Preflight (run in parallel)

- Confirm `go.mod` contains `module github.com/bitrise-io/bitrise/v2`. If not, ask the user to `cd` to the repo root and stop.
- `git status --porcelain` must be empty. If dirty, ask: continue or abort (recommend abort).
- Verify on PATH: `gh`, `gcloud`, `shasum`, `git`. List all missing and stop if any.

## Step 1: Read current state

Grep `toolprovider/mise/mise.go` for `misePreviewVersion` and `miseStableVersion`. Print:

```
Current preview: vYYYY.M.D  |  Current stable: vYYYY.M.D
```

## Step 2: Discover latest release

```
gh api repos/jdx/mise/releases/latest --jq '.tag_name'
```

Validate the tag matches `^v\d{4}\.\d+\.\d+$`. Print it. Stop on failure.

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

For each CalVer release in range: `gh release view <tag> --repo jdx/mise --json tagName,name,publishedAt,body,url`

Scan `body` (case-insensitive) for keywords related to declarative tool setup:

```
install   plugin     registry  ls-remote  use        settings
lockfile  asdf       env       shim       cache      mirror
tarball   tools      latest    core       idiomatic  mise.toml
direnv    backend    tool-versions        breaking   deprecat   remove
python    ruby       node      go         java       flutter
swift     kotlin     rust      dotnet     erlang     elixir
bun       deno       zig
```

Capture matching lines verbatim, attributed to their release tag and URL.

**Risk level** (always print with one-line justification):
- **low** - baseline.
- **medium** - ≥3 releases in range, OR any match outside breaking/deprecat/remove keywords.
- **high** - "breaking change" mentioned, OR a CLI subcommand we invoke is removed/renamed. Check with: `grep -rE 'mise (install|ls-remote|use|plugin|set|cache|exec|where|which|current|latest)' toolprovider/mise/`.
- **extreme** - asdf compatibility removed, registry format changed, or tarball layout changed.

**Related issues** (best effort): for matched lines that look like bugs, do one `gh search issues "in:title <keyword>" --repo bitrise-io/bitrise --limit 3`. Surface hits as `Possibly related: bitrise-io/bitrise#<num> - <title>`. Silent on no match.

## Step 5: Download artifacts

```bash
tmp=$(mktemp -d -t mise-bump-XXXX)
gh release download <targetVersion> --repo jdx/mise --dir "$tmp" \
  --pattern 'mise-v*-linux-x64.tar.gz' \
  --pattern 'mise-v*-linux-arm64.tar.gz' \
  --pattern 'mise-v*-macos-x64.tar.gz' \
  --pattern 'mise-v*-macos-arm64.tar.gz' \
  --pattern 'SHASUMS256.txt'
```

Verify `$tmp` contains exactly 4 `.tar.gz` files + `SHASUMS256.txt`. If anything is missing, `rm -rf "$tmp"` and stop.

## Step 6: Verify checksums (hard fail on mismatch)

Compute `shasum -a 256` for each artifact. Compare against `SHASUMS256.txt` (fallback: `gh release view --json assets`).

**On any mismatch:** print a `| platform | expected | actual | match |` table, `rm -rf "$tmp"`, and STOP: do not edit files or touch GCS.

**On full match:** save checksums as `computedChecksums[platform]` and continue.

## Step 7: GCS mirror

For each of the 4 artifacts:
1. Check: `gcloud storage objects describe gs://mise-release-mirror/v<VERSION>/<ARTIFACT> --format='value(name)' 2>/dev/null`
2. If absent → upload. If present → log "already mirrored", skip.

Then verify: `gcloud storage ls gs://mise-release-mirror/v<VERSION>/` must list all 4. If any missing → `rm -rf "$tmp"` and STOP, do not edit files.

After successful verification, `rm -rf "$tmp"`.

## Step 8: Edit `toolprovider/mise/mise.go`

**For `{stable-match-preview}`: this is where execution resumes after skipping steps 4–7.**

Update the relevant block(s) using the Edit tool with enough context to disambiguate preview vs stable (include the `misePreviewVersion =` or `miseStableVersion =` const line). For `{stable-match-preview}`, copy current preview values into the stable block.

Preserve `gofmt` column alignment. After editing, re-read to verify the new version + 4 hashes appear exactly once each. On any issue: `git checkout -- toolprovider/mise/mise.go` and stop.

## Step 9: PR

Build the PR body below. Write "unchanged" for tracks that were not bumped; write "No changes detected." for empty changelog sections.

```markdown
## Summary

Updates pinned Mise version(s):
- Preview: <oldPreview> → <newPreview>
- Stable:  <oldStable>  → <newStable>

Risk: **<level>** - <one-line justification>

## Changes affecting our work (declarative tool setup)

- [vYYYY.M.D](<url>)
  - <verbatim changelog line>

## Releases in range

- [vYYYY.M.D](<url>) - <release name>

## Full diff

[Compare <oldMin>…<newMax>](https://github.com/jdx/mise/compare/<oldMin>...<newMax>)

## Verification

- [x] SHA256 verified against `SHASUMS256.txt`. All 4 platforms match.
- [x] Artifacts present at `gs://mise-release-mirror/v<VERSION>/`.
```

Ask: "Create the PR now?"

**Yes:** branch (`bump/mise-<targetVersion>` or `bump/mise-stable-<currentPreview>`; append `-2` if exists), `git add`, `git commit -m "Update Mise version"`, `git push -u origin`, `gh pr create`. Print PR URL.

**No:** save body to `/tmp/mise-bump-<VERSION>-pr-description.md`. Leave edits in working tree. Print file path.

## Step 10: Final report

Print as markdown. Same sections in the same order for every update type. Empty sections get "nothing to report", never skipped.

### Mise <track-label> update: <oldA> → <newA>[ + stable <oldB> → <newB>]

Type: `<preview / stable / preview + stable / match-stable-with-preview>`
Risk: **<level>** - <one-line justification>

**Changes affecting our work** (declarative tool setup)
- vYYYY.M.D - <summary> ([link](<url>))

**Releases in range**
- [vYYYY.M.D](<url>)

**Compare**
[<oldMin>…<newMax>](https://github.com/jdx/mise/compare/<oldMin>...<newMax>)

**Checksums applied**

| platform | sha256 |
|---|---|
| linux-x64 | `<hash>` |
| linux-arm64 | `<hash>` |
| macos-x64 | `<hash>` |
| macos-arm64 | `<hash>` |

*(For match-stable: "Stable now matches preview values — no new checksums.")*

**Next**

PR: <URL> OR Suggested PR description saved to `/tmp/mise-bump-<VERSION>-pr-description.md`
