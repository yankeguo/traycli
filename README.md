# traycli

A cross-platform system tray app that runs a configurable command and keeps it alive. Stdout and stderr are captured to files; if the process exits, it restarts after 5 seconds.

## Features

- **Cross-platform**: Windows, macOS, Linux
- **Simple config**: JSON config at `~/.traycli/config.json`
- **Auto-restart**: Waits 5 seconds after exit, then restarts
- **Output capture**: stdout → `~/.traycli/stdout.txt`, stderr → `~/.traycli/stderr.txt`
- **Tray menu**: Right-click to see uptime and restart count; quit from the menu

## Installation

### Build from source

Requires Go 1.25+ and CGO (for systray and dialog).

```bash
git clone https://github.com/yankeguo/traycli.git
cd traycli
go build -o traycli .
```

**Windows** (hide console window):

```bash
go build -ldflags "-H=windowsgui" -o traycli.exe .
```

### Linux dependencies

For systray and native dialogs:

```bash
# Debian/Ubuntu
sudo apt install gcc libgtk-3-dev libayatana-appindicator3-dev

# Fedora
sudo dnf install gtk3-devel libappindicator-gtk3-devel
```

### macOS

Pack as an app bundle to run without a visible Dock icon. Minimal structure:

```
traycli.app/
  Contents/
    Info.plist
    MacOS/
      traycli
```

Example `Info.plist` snippet:

```xml
<key>LSUIElement</key>
<true/>
<key>NSHighResolutionCapable</key>
<true/>
```

## Configuration

Create `~/.traycli/config.json`:

```json
{
  "cmd": ["python3", "-m", "http.server", "8080"],
  "env": {
    "CUSTOM_VAR": "value"
  }
}
```

- **cmd**: Array of strings; first element is the program, rest are arguments. Executed directly (no shell).
- **env**: Optional key-value object. Merged into the process environment; overrides existing variables.

If `config.json` is missing, traycli shows an error dialog, creates an empty template, opens it for editing, and exits.

## Usage

1. Put your command in `~/.traycli/config.json` under the `cmd` key
2. Run `traycli` (or `traycli.exe` on Windows)
3. Check the system tray for the traycli icon
4. Right-click the icon:
   - See uptime and restart count
   - Choose **Quit** to exit

Output files:

- `~/.traycli/stdout.txt` — stdout
- `~/.traycli/stderr.txt` — stderr

Logs are appended across restarts.

## Platform notes

| Platform  | Notes |
|-----------|-------|
| **Windows** | Build with `-ldflags "-H=windowsgui"` to avoid a console window. Child process also runs without a visible console. |
| **macOS**   | Use an app bundle with `LSUIElement` to hide from the Dock. |
| **Linux**   | Requires GTK3 and AppIndicator; may need a system tray implementation. |
