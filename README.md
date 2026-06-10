# EchoState

[![CI](https://github.com/<your-org>/echostate/actions/workflows/ci.yml/badge.svg)](https://github.com/<your-org>/echostate/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/<your-org>/echostate)](https://github.com/<your-org>/echostate/releases)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![Node](https://img.shields.io/badge/Node-18+-339933?logo=node.js)](https://nodejs.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

**Local-first, zero-cloud context bridge.** EchoState reads your local workspace state, maps the project tree, summarizes heavy context via a local LLM (Ollama), and securely serves the payload to a custom browser extension that injects it into ChatGPT and Claude web interfaces.

---

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌──────────────────┐
│  Go Daemon  │────▶│  Node Relay  │────▶│  Browser Ext.    │
│  :9081      │     │  :9080       │     │  (Manifest V3)   │
└─────────────┘     └──────────────┘     └──────────────────┘
       │                                        │
       ▼                                        ▼
  .echo-state.md                           chatgpt.com
  .git/                                    claude.ai
  project tree (+.gitignore)
  Ollama (:11434)
```

## Downloads

Pre-built binaries and packages are available on the [Releases](https://github.com/<your-org>/echostate/releases) page.

| Component | Platform | Format |
|-----------|----------|--------|
| **Daemon** | Linux amd64/arm64 | `.tar.gz` |
| **Daemon** | macOS amd64/arm64 | `.tar.gz` |
| **Daemon** | Windows amd64 | `.tar.gz` |
| **Relay API** | Any (Node.js) | `.tar.gz` |
| **Extension** | Chrome/Edge/Brave | `.zip` |
| **Source** | — | `.tar.gz` |

## Prerequisites

- **Go 1.21+** — to build the daemon from source
- **Node.js 18+** — for the relay API
- **Ollama** — optional, for context summarization (`curl -fsSL https://ollama.com/install.sh | sh`)

## Quick Start

### 1. Create your context file

Create `.echo-state.md` in your project root with current notes:

```markdown
# Current Sprint
- Refactoring auth module to use OAuth2
- Investigating memory leak in WebSocket handler
```

### 2. Start the Go daemon

```bash
cd daemon
go mod tidy
go run . &
```

Listens on `http://localhost:9081` — serves `GET /state` and `GET /health`.

### 3. Start the Node Relay API

```bash
cd api
npm install
node index.js &
```

Listens on `http://localhost:9080` — serves `GET /context`.

### 4. Install the browser extension

Load the extension from `extension/` into Chrome/Brave:
1. Open `chrome://extensions`
2. Enable **Developer mode**
3. Click **Load unpacked** and select the `extension/` directory

## Build from source

```bash
# Build everything for the host platform
make build-all

# Cross-compile the daemon for all platforms
make build-daemon-all

# Package the extension
make build-extension

# Full release build
make release
```

Artifacts land in `bin/`.

## How It Works

### The injected prompt

When you click **Inject Context** in the extension's floating badge:

```
=== Summarized Context ===
- Refactoring auth module to use OAuth2
- Investigating memory leak in WebSocket handler

=== File Tree ===
.
├── src
│   ├── main.go
│   └── server.go
├── .echo-state.md
└── go.mod

=== Last Commit ===
abc123def - Add OAuth2 middleware
```

### Summarization

If `.echo-state.md` exceeds 4000 characters:
1. The daemon sends the content to Ollama at `http://localhost:11434/api/generate`
2. Ollama returns a dense bullet-point summary
3. The summary is cached until the file changes
4. If Ollama is unreachable, the daemon falls back to the raw content

### DOM Injection

The extension uses two-stage injection:
1. **Primary**: Dispatches a native `ClipboardEvent` with `DataTransfer`
2. **Fallback**: Sets the input value directly via `HTMLTextAreaElement.value` setter (ChatGPT) or `textContent` (Claude)

## Configuration

| Variable | Default | Description |
|---|---|---|
| `summaryThreshold` | 4000 chars | Max length before summarization |
| `ollamaModel` | `llama3.2` | Ollama model for summarization |
| `POLL_INTERVAL` | 5000 ms | Extension polling interval |
| `RELAY_URL` | `http://localhost:9080/context` | Extension context source |

## MCP (Model Context Protocol)

EchoState exposes a **MCP server** via the `-mode mcp` flag, allowing local coding agents (Claude Desktop, OpenCode, Cursor, any MCP host) to call tools that return workspace context directly — no browser extension needed.

### Tools

| Tool | Description |
|------|-------------|
| `get_workspace_context` | Returns `.echo-state.md` notes, file tree, and latest git commit |

### Configuration (Claude Desktop)

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "echostate": {
      "command": "/path/to/echostate-daemon",
      "args": ["-mode", "mcp"]
    }
  }
}
```

For other MCP hosts, configure a **stdio** transport running the same command.

> **Note:** The daemon must run from a project root containing `.echo-state.md` and `.git/`. It reads these to build the workspace state.

## Agent Skill

A formal agent skill definition is available in [skills/echostate.md](skills/echostate.md). This file can be used to teach other agents how to use EchoState effectively, including the "First Turn Rule" and best practices for context grounding.

## Project Structure

```
.
├── daemon/               # Go daemon (source of truth)
│   ├── main.go           # Entry point, HTTP server
│   ├── state.go          # Thread-safe state management
│   ├── watcher.go        # fsnotify file watcher
│   ├── tree.go           # Project tree builder
│   ├── git.go            # Git commit reader
│   ├── ollama.go         # Ollama API client
│   └── gitignore.go      # .gitignore parser
├── api/                  # Node.js relay API
│   ├── index.js
│   └── package.json
├── extension/            # Vanilla JS Chrome extension (MV3)
│   ├── manifest.json
│   ├── content.js
│   └── style.css
├── scripts/              # Build & packaging scripts
│   ├── build-daemon.sh
│   └── package-extension.sh
├── .github/workflows/    # CI / Release pipelines
│   ├── ci.yml
│   └── release.yml
├── Makefile              # Central build entry point
└── README.md
```

## CI/CD

| Workflow | Trigger | What it does |
|---|---|---|
| **CI** | Push/PR to `main` | Builds daemon, runs `go vet`, validates API syntax, packages extension |
| **Release** | Tag `v*` | Cross-compiles daemon for 5 targets, packages extension & API, creates GitHub Release with all assets |

Tag a release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Contributing

1. Fork the repo
2. Create a feature branch
3. Run `make test` before committing
4. Open a pull request

## License

MIT
