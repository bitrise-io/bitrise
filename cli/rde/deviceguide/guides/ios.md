<!-- Mirror of the RDE device-session guide shipped with the RDE backend; keep in sync with the backend release you target — do not edit here. -->
# iOS simulator sessions — driving the device

Read the main device sessions guide first (create, wait for READY, connect,
do-nots): `bitrise_devenv_device_guide` with `guide: "device-sessions"` (or the
resource `bitrise-devenv://guides/device-sessions`) or `bitrise rde
device-guide` with no platform argument. This page is the iOS specifics.
Everything here runs on the session VM (in-band via `execute`, or over SSH). Drive the simulator only through the recipes below — never through
the VM's desktop screenshot/click tools: a Simulator.app window may be visible
on the VM's desktop, it is not how you look at or drive the device; leave it
alone.

## What is on the VM

- One simulator named **`bitrise-preview`**, booted. Find it with
  `xcrun simctl list devices booted` (the UDID is the value in parentheses).
  `booted` also works as the UDID argument to most `simctl` commands while it
  is the only booted device.
- **serve-sim** (npm package in `~/serve-sim`, version pinned by its lockfile)
  streaming it on `127.0.0.1:3200` — this is the viewer's video source *and*
  your input/AX API. Its CLI: `cd ~/serve-sim && TMPDIR=/tmp node
  node_modules/.bin/serve-sim …`. **`TMPDIR=/tmp` is required**: the CLI finds
  the running server through `$TMPDIR/serve-sim/server-<UDID>.json`, the
  server writes it under `/tmp`, and a macOS login shell gets a per-user
  `TMPDIR` under `/var/folders/…` — without the prefix every command answers
  `No serve-sim server running. Run 'serve-sim' first.` while the stream is
  fine. That discovery file (`/tmp/serve-sim/server-<UDID>.json`: pid, port,
  url, streamUrl, wsUrl) is also the cheapest way to get the stream/WS URLs.
- Service port on the session: `device-web-view` (VM 3200 → local 3200, the
  same name and local port an Android session uses). From your machine:
  `ssh -N -L 3200:127.0.0.1:3200 …` then `http://127.0.0.1:3200`.
- Xcode with the stack's iOS runtimes; `xcrun simctl` for everything device
  lifecycle-ish (install, launch, screenshot, logs, appearance, location…).
  `xcrun simctl list runtimes` shows which iOS versions this stack can boot;
  a requested `os_version` it lacks is substituted with the newest, not
  rejected.
- **XcodeGen** (`xcodegen`) is installed — the way from loose Swift sources
  to a buildable project: write `project.yml`, `xcodegen generate`, then
  `xcodebuild -scheme App -destination "id=$UDID" build` (or `test`). A
  single-file SwiftUI app also builds with no project at all: `mkdir -p
  App.app && xcrun swiftc -sdk "$(xcrun --sdk iphonesimulator
  --show-sdk-path)" -target arm64-apple-ios17.0-simulator -parse-as-library
  -emit-executable -o App.app/App main.swift` (swiftc does not create the
  bundle directory), add `App.app/Info.plist` (CFBundleIdentifier,
  CFBundleExecutable, MinimumOSVersion), then `xcrun simctl install
  "$UDID" App.app`.
- Python 3.13 (asdf) — `pipx install fb-idb` works if you want the `idb`
  client; `idb_companion` is not pre-installed (Homebrew 6 requires
  `brew tap facebook/fb && brew trust facebook/fb && brew install
  idb-companion`). You do not need idb for anything below.

## Accessibility tree (what is on screen, where)

```bash
UDID=$(xcrun simctl list devices booted | sed -nE 's/.*\(([0-9A-F-]{36})\).*/\1/p' | head -1)
curl -s "http://127.0.0.1:3200/helper/$UDID/ax"
```

Returns a JSON array of element trees for the **frontmost app** — on the home
screen that is SpringBoard, so you get the home screen's icons, not an error.
For a second or two right after READY, and again right after a `launch`, it
can answer `noFrontmostApplication` instead while the foreground app is
still settling — retry after ~2 s; it is transient, not a broken device.
Each node has `type` (Button, Heading, StaticText, TextField, Cell, …),
`AXLabel`, `AXValue`, `AXUniqueId` (accessibility identifier when the app
sets one), `enabled`, `frame` (`x`,`y`,`width`,`height` in **points**,
portrait-up), and `children`. Filter it (jq / python) rather than reading it
raw; a Settings screen is ~20–30 nodes, a busy app can be hundreds.

Screen size in points = the **root node's `frame`** of that JSON (e.g.
393×852 for an iPhone 16). Point → normalized tap coordinate: `x/width`,
`y/height`. serve-sim's root `/ax` is a *different* endpoint: a server-sent
event stream with a flat `elements` array (`id`/`path`/`label`/`value`/
`role`/`role_description`, plus a `screen` object) in the same point space —
`curl -sN --max-time 5 http://127.0.0.1:3200/ax | head -c 2000` (the first
line is a bare `:` keep-alive; it can take a few seconds right after READY).
Prefer the per-device `/helper/<UDID>/ax` JSON: synchronous and complete.

`curl: (7) Failed to connect` on 3200 means serve-sim is not running — read
`tail -50 ~/serve-sim.log`. A `Fatal error: Incorrect actor executor
assumption … FrameCapture` there means serve-sim crashed on this stack's
macOS/Xcode; the simulator itself keeps working over `xcrun simctl`, but the
stream, `/ax` and the input CLI are gone for this session (the main guide §6
says what to do).

## Input (serve-sim CLI)

```bash
cd ~/serve-sim && export TMPDIR=/tmp && S="node node_modules/.bin/serve-sim"
$S -l                                    # {"running":true,…,"port":3200,…} — sanity check (see below)
$S tap    -d "$UDID" 0.5 0.47            # normalized 0..1 coordinates
$S type   -d "$UDID" "hello world"       # US keyboard; focus a text field first
$S button -d "$UDID" home                # hardware button
$S gesture -d "$UDID" '{"type":"begin","x":0.5,"y":0.8}'   # then move/end events for swipes
$S rotate -d "$UDID" landscape_left      # portrait | portrait_upside_down | landscape_left | landscape_right
```

After `rotate`, wait ~3 s and re-read the AX tree before you tap or shoot:
the screen's width and height swap, so every normalized coordinate you
computed in the other orientation is stale. A control whose frame lies past
the new screen width is genuinely unreachable — that is a layout bug, not a
tap failure (verify by rotating back and tapping it there).

`-l` obeys the same `TMPDIR` rule as every other subcommand but degrades
*silently*: with a wrong `TMPDIR` it prints `{"running":false}` — the same
output as a dead server — where `tap` would have printed the `No serve-sim
server running` error. Treat a false `-l` as the `TMPDIR` mistake first
(`lsof -iTCP:3200` shows the listener alive) before concluding serve-sim is
down.

Each invocation is a Node process (~0.2–0.8 s); batch several in one
`execute` call. `$S --help` lists more (permissions, camera, …). Never
pass `--host 0.0.0.0` to serve-sim: it exposes a shell-exec route.

Language: test an app in another locale by launching it with Apple's launch
arguments — `xcrun simctl launch "$UDID" com.example.app -AppleLanguages
"(ar)" -AppleLocale ar_AE` — no system-wide change needed. There is no
locale option on `device_spec`.

Other tooling is welcome on this simulator — `idb` (`pipx install fb-idb` +
`idb_companion`), XCUITest / `xcodebuild test -destination "id=$UDID"`,
Appium's XCUITest driver, Maestro — as long as it targets **this UDID**.
`xcrun simctl` itself has **no** tap/type. Pass the UDID explicitly
(`--udid`, `appium:udid`, `-destination id=`); tools that pick or create a
device by name/type will otherwise boot a second simulator that nobody
streams.

Troubleshooting: `No serve-sim server running` while `lsof -iTCP:3200` shows a
listener means your shell's `TMPDIR` is wrong (see above), not a dead server.

## Install, launch, screenshot, logs

```bash
xcrun simctl install "$UDID" /path/App.app          # simulator build (arm64 on Apple silicon hosts)
xcrun simctl launch  "$UDID" com.example.app        # prints the pid
xcrun simctl terminate "$UDID" com.example.app
xcrun simctl io "$UDID" screenshot --type=png /tmp/shot.png
xcrun simctl spawn "$UDID" log stream --style compact --predicate 'process == "MyApp"'
xcrun simctl ui "$UDID" appearance dark            # light|dark
xcrun simctl openurl "$UDID" "myapp://deep/link"
```

Screenshots are ~2–3 MB PNG at native resolution (content-dependent; the home
screen is the worst case); `sips -Z 800 in.png --out small.png` gets it to
~350 KB. Do not base64 them into `execute` output — that is hundreds of
thousands of characters and has corrupted files. Point the human at the
session page's device view, or copy the file out with `download` / `scp` as in
the main guide §3.

**Let the screen settle before you shoot.** A screenshot taken within ~2 s of
`launch`, `openurl` or a rotation can be a solid black frame while
`/helper/<UDID>/ax` already shows the new screen; at ~3 s it is correct. Wait
~3 s and re-shoot before reporting a blank screen or a broken layout — the
tree is the ground truth, the pixels lag it.

The simulator has no Date & Time settings pane (it takes the host's clock) —
automation of that screen has nothing to find.

## Do not

- `xcrun simctl shutdown|erase|delete|create` on `bitrise-preview`, or boot a
  second device and drive that one — serve-sim streams the UDID it was started
  with; the viewer and `device.state` follow *that* device.
- kill `serve-sim` or anything on port 3200; `open -a Simulator`; drive the
  Simulator window through desktop click/screenshot tools.
- run `~/bin/simulator-up.sh` while `device.state` is `BOOTING`.

Recovery — only when `device.state` is `FAILED` **and** `pgrep -f
simulator-up.sh` prints nothing: `~/bin/simulator-up.sh` (boots what is not
running, restarts serve-sim, re-reports readiness; exits at once if a run is
already in progress). Then re-check `device.state`. A `FAILED` whose notes say
serve-sim crashed on this stack will fail the same way again — see the main
guide §6.
