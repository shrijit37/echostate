---
name: echostate
description: Use EchoState to bridge local workspace context (notes, file tree, git status) into the agent's turn. It reads .echo-state.md, performs LLM-powered summarization (via local Ollama), and maps the project structure. Essential for zero-cloud, local-first workspace grounding.
---

# EchoState Context Bridge

**EchoState** is a local-first, zero-cloud context bridge. It allows you to quickly ground yourself in a workspace by reading three primary sources of truth:
1.  **`.echo-state.md`**: A local markdown file where the human (or you) keeps current project notes, goals, and status. Large files are automatically summarized using a local Ollama instance.
2.  **File Tree**: A complete, depth-limited map of the project files, respecting `.gitignore` rules.
3.  **Git Status**: The latest commit hash and message.

Use this skill whenever you join a new project, start a session, or feel "lost" in the codebase.

---

## Your Tool: `get_workspace_context`

This is the primary way to interact with EchoState. It returns a structured JSON payload containing the entire workspace state.

### Usage

```javascript
get_workspace_context()
```

### Result Schema

```json
{
  "raw_context": "The full content of .echo-state.md (if under the threshold)",
  "summarized_context": "An AI-generated summary of .echo-state.md (if over the threshold)",
  "file_tree": "A formatted directory tree string",
  "last_commit": {
    "hash": "abc123def...",
    "message": "feat: add auth module"
  }
}
```

---

## When to use EchoState

1.  **The "First Turn" Rule**: Call `get_workspace_context` at the very beginning of a project to understand the human's current goals and the project structure.
2.  **Task Switching**: When moving from one large task to another, check if `.echo-state.md` has been updated with new instructions or context.
3.  **Debugging**: Use the file tree to identify where relevant files might live before doing expensive `find` or `grep` operations.
4.  **Context Recovery**: If you are interrupted or your context window is cleared, EchoState is your "long-term memory" in the filesystem.

---

## Rules and Best Practices

- **Respect `.echo-state.md`**: This file is the human's "Project Directive". If it contradicts your general knowledge, follow the file.
- **Update the state**: You are encouraged to update `.echo-state.md` yourself when you complete a major milestone (use `write` or `edit` on the file). This keeps the context fresh for the next agent turn or for the human.
- **Don't over-poll**: The state typically only changes when files change. One call per major task is usually sufficient.
- **Ollama Fallback**: If the local Ollama instance is offline, EchoState will return the `raw_context` even if it's large. Be prepared to handle large blocks of text if that happens.

---

## Setup & Troubleshooting

The EchoState daemon must be running in the project root. If `get_workspace_context` is not available as a tool, ensure the MCP server is configured:

**Claude Desktop / OpenCode Configuration:**
```json
{
  "mcpServers": {
    "echostate": {
      "command": "echostate-daemon",
      "args": ["-mode", "mcp"]
    }
  }
}
```

If you receive an error about "failed to refresh workspace state", check that you are in a valid Git repository or that the daemon has permission to read the current directory.
