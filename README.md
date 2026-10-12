# hue-tui

[![CI](https://github.com/angristan/hue-tui/actions/workflows/ci.yml/badge.svg)](https://github.com/angristan/hue-tui/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/angristan/hue-tui)](https://go.dev/)
[![License](https://img.shields.io/github/license/angristan/hue-tui)](LICENSE)

A terminal-based Philips Hue control application written in Go.

![demo](demo/demo.gif)

## Features

- **Bridge Discovery**: Automatic discovery via mDNS and Philips Hue cloud
- **Bridge Pairing**: Easy link button pairing flow
- **Light Control**: Toggle, brightness, color temperature
- **Room Grouping**: Lights organized by room with group controls
- **Scene Activation**: Browse and activate scenes
- **Real-time Updates**: Server-sent events for live state updates
- **Search**: Filter lights by name
- **Keyboard-driven**: Full vim-style navigation
- **Mouse input**: Click to select, click a light status dot to toggle, scroll with the wheel, click/drag brightness bars (including the details panel); drag requests are throttled

## Installation

### Homebrew

```bash
brew tap angristan/tap
brew install hue-tui
```

### From releases

Download the latest release from the [releases page](https://github.com/angristan/hue-tui/releases).

### Using Go

```bash
go install github.com/angristan/hue-tui/cmd/hue@latest
```

### From source

```bash
git clone https://github.com/angristan/hue-tui
cd hue-tui
make build
./hue
```

## Usage

```bash
hue
```

On first run, the app will:

1. Search for Hue bridges on your network
2. Prompt you to press the link button on your bridge
3. Save the connection credentials for future use

### Limit the UI to a room

Use a Hue API v1 group ID or a comma-separated list of room/zone IDs:

```bash
hue --group 3      # Arbeitszimmer (room)
hue --group 81     # Play Bars (zone)
hue --group 3,81   # Show both
```

Hue v2 room/zone UUIDs also work. Without `--group`, only rooms are shown (original behavior). An unknown ID, even in a list, causes an error rather than exposing unrelated rooms. This scopes the UI, not the API key's permissions.

## Keybindings

### Navigation

| Key       | Action              |
| --------- | ------------------- |
| `j` / `↓` | Move down           |
| `k` / `↑` | Move up             |
| `h` / `←` | Decrease brightness |
| `l` / `→` | Increase brightness |
| `H` / `L` | Fine-adjust brightness; `Shift/Ctrl/Alt+←/→` when sent by the terminal |

### Mouse control

- Click a room, zone or light to select it.
- Click a light's status dot to toggle it.
- Click or drag a brightness bar in the list or right-hand details panel to set the brightness (1–100%). Dragging a room/zone slider adjusts all its lights.
- Scroll with the mouse wheel. Requests during dragging are rate-limited; the final value is sent on release.

### Light Control

| Key     | Action                   |
| ------- | ------------------------ |
| `Space` | Toggle light on/off      |
| `0`     | Set brightness to 100%   |
| `1-9`   | Set brightness to 10-90% |
| `w` / `c` | Warmer / cooler color temperature |
| `W` / `C` | Fine-adjust color temperature; `Alt+w/c` also works |
| `[` / `]` | Decrease / increase hue |
| `{` / `}` | Fine-adjust hue; `Alt+[/]` also works |
| `-` / `=` | Decrease / increase saturation |
| `_` / `+` | Fine-adjust saturation; `Alt+-/=` also works |

### Room Control

| Key | Action                      |
| --- | --------------------------- |
| `a` | Turn all lights in room on  |
| `x` | Turn all lights in room off |

### Other

| Key   | Action            |
| ----- | ----------------- |
| `s`   | Open scenes modal |
| `/`   | Search lights     |
| `Tab` | Toggle side panel |
| `r`   | Refresh           |
| `q`   | Quit              |

## Tuning increments

Normal and fine increments are configurable in `~/.config/hue-cli/config.json` via an optional `tuning` object, or with environment variables. The defaults are:

| Variable | Default | Unit |
| --- | ---: | --- |
| `HUE_TUI_BRIGHTNESS_STEP` | 3 | percentage points |
| `HUE_TUI_FINE_BRIGHTNESS_STEP` | 1 | percentage points |
| `HUE_TUI_TEMPERATURE_STEP` | 25 | mirek |
| `HUE_TUI_FINE_TEMPERATURE_STEP` | 5 | mirek |
| `HUE_TUI_HUE_STEP` | 20 | degrees |
| `HUE_TUI_FINE_HUE_STEP` | 2 | degrees |
| `HUE_TUI_SATURATION_STEP` | 10 | percentage points |
| `HUE_TUI_FINE_SATURATION_STEP` | 1 | percentage points |

Example JSON fields:

```json
"tuning": {
  "brightness_step": 3,
  "fine_brightness_step": 1,
  "temperature_step": 25,
  "fine_temperature_step": 5
}
```

Other optional fields: `hue_step`, `fine_hue_step`, `saturation_step`, and `fine_saturation_step`.

Priority: environment variable > JSON setting > built-in default. All step settings must be positive integers within their supported range; invalid values fall back to defaults.

Example environment override: `HUE_TUI_BRIGHTNESS_STEP=2 hue --group 3,81`.

Alt key combinations require terminal support and may be intercepted (for example, Option+arrows used for word navigation in iTerm2). The printable Shift variants above are the portable alternative and require no terminal configuration.

## Configuration

Configuration is stored in `~/.config/hue-cli/config.json`:

```json
{
  "bridges": [
    {
      "host": "192.168.1.100",
      "username": "<app-key>",
      "bridge_id": "001788FFFE123456"
    }
  ],
  "last_bridge_id": "001788FFFE123456",
  "tuning": {
    "brightness_step": 3,
    "fine_brightness_step": 1
  }
}
```

## Requirements

- Philips Hue Bridge (v2 API)
- Network access to the bridge

## Tech Stack

| Component      | Library                                                   |
| -------------- | --------------------------------------------------------- |
| TUI Framework  | [Bubble Tea](https://github.com/charmbracelet/bubbletea)  |
| Styling        | [Lip Gloss](https://github.com/charmbracelet/lipgloss)    |
| mDNS Discovery | [hashicorp/mdns](https://github.com/hashicorp/mdns)       |
| WebSocket      | [gorilla/websocket](https://github.com/gorilla/websocket) |

## Project Structure

```
hue-tui/
├── cmd/hue/              Entry point
└── internal/
    ├── api/              Hue Bridge V2 API client
    │   ├── client.go     HTTP client
    │   ├── discovery.go  mDNS + cloud discovery
    │   ├── events.go     Server-sent events
    │   └── pairing.go    Link button pairing
    ├── config/           Configuration management
    ├── models/           Data models (Light, Room, Scene, Color)
    └── tui/              Terminal UI
        ├── screens/      Setup, Main, Scenes screens
        ├── components/   Reusable UI components
        ├── styles/       Color theme
        └── messages/     Cross-screen messages
```

## Development

```bash
# Build
make build

# Run
make run

# Test
make test

# Lint (requires golangci-lint)
make lint

# Format
make fmt
```

## License

MIT
