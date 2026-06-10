# EchoState

Local-first, zero-cloud context bridge. Reads a local workspace state, maps the project tree, summarizes heavy context via a local LLM (Ollama), and securely serves the payload to a custom browser extension. The extension injects the context into ChatGPT and Claude web interfaces.

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌──────────────────┐
│  Go Daemon  │────▶│  Node Relay  │────▶│  Browser Ext.    │
│  :9081      │     │  :9080       │     │  (Plasmo/React)  │
└─────────────┘     └──────────────┘     └──────────────────┘
       │                                        │
       ▼                                        ▼
  .echo-state.md                           chatgpt.com
  .git/                                    claude.ai
  project tree (+.gitignore)
  Ollama (:11434)
```

## Prerequisites

- **Go 1.21+** — for the daemon
- **Node.js 18+** — for the relay API
- **Ollama** — optional, for context summarization (install: `curl -fsSL https://ollama.com/install.sh | sh`)

## Quick Start

### 1. Create your context file

Create `.echo-state.md` in your project root with your current architectural notes, tasks, or any context you want available to the chat AI:

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

The daemon runs on `http://localhost:9081` and serves `GET /state`. It:
- Watches `.echo-state.md` for changes
- Builds the project tree (respecting `.gitignore`)
- Reads the latest git commit from `.git/logs/HEAD`
- If `.echo-state.md` exceeds 4000 characters, summarizes it via Ollama (`llama3.2`)

### 3. Start the Node Relay API

```bash
cd api
npm install
node index.js &
```

The relay runs on `http://localhost:9080` and serves `GET /context`. It:
- Fetches state from the Go daemon
- Enforces CORS — only `https://chatgpt.com`, `https://claude.ai`, and `http://localhost:3000` are allowed

### 4. Build and install the browser extension

```bash
cd extension
npm install
npm run build
```

Load the extension from `extension/build/chrome-mv3/` into Chrome/Brave:
1. Open `chrome://extensions`
2. Enable **Developer mode**
3. Click **Load unpacked** and select `extension/build/chrome-mv3/`

### Startup (all at once)

Install tools like `concurrently` or use a simple script:

```bash
# Terminal 1 — Daemon
cd daemon && go run .

# Terminal 2 — Relay API
cd api && node index.js

# Terminal 3 — Extension dev mode
cd extension && npm run dev
```

Or use a shell script (`start.sh`):

```bash
#!/usr/bin/env bash
set -e

echo "Starting EchoState daemon..."
cd daemon && go run . &
DAEMON_PID=$!

echo "Starting EchoState relay API..."
cd ../api && node index.js &
API_PID=$!

echo "EchoState running: daemon=$DAEMON_PID api=$API_PID"
echo "Press Ctrl+C to stop."

trap "kill $DAEMON_PID $API_PID 2>/dev/null; exit" INT TERM
wait
```

## How It Works

### The injected prompt

When you click **Inject Context** in the extension's floating badge, the following text is pasted into the chat textarea:

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

The extension uses a two-stage injection strategy:
1. **Primary**: Dispatches a native `ClipboardEvent` with `DataTransfer` containing the formatted context
2. **Fallback**: Directly sets the input value and dispatches an `InputEvent` (for ChatGPT: native `HTMLTextAreaElement.value` setter; for Claude: `textContent` assignment)

## Configuration

| Variable | Default | Description |
|---|---|---|
| `summaryThreshold` | 4000 chars | Max `.echo-state.md` length before summarization |
| `ollamaModel` | `llama3.2` | Ollama model for summarization |
| `POLL_INTERVAL` | 5000 ms | Extension polling interval |
| `RELAY_URL` | `http://localhost:9080/context` | Extension context source |

## Project Structure

```
.
├── daemon/          # Go daemon (source of truth)
│   ├── go.mod
│   ├── main.go      # Entry point, HTTP server
│   ├── state.go     # Thread-safe state management
│   ├── watcher.go   # fsnotify file watcher
│   ├── tree.go      # Project tree builder
│   ├── git.go       # Git commit reader
│   ├── ollama.go    # Ollama API client
│   └── gitignore.go # .gitignore parser
├── api/             # Node.js relay API
│   ├── package.json
│   └── index.js
├── extension/       # Plasmo browser extension
│   ├── package.json
│   ├── tsconfig.json
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   └── src/
│       ├── style.css
│       └── content.tsx    # Content script (UI + injection)
└── README.md
```
