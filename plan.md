# ghx - AI-Powered GitHub Workflow Assistant

A Go TUI tool for GitHub workflows with AI integration. **Zero commands - zero internet required for core features - AI-assisted workflows.**

## Core Concept

Run `ghx` and get an interactive menu-driven interface with AI assistance for commit messages, PR descriptions, code explanations, and more.

## Project Structure

```
ghx/
├── go.mod
├── go.sum
├── main.go                # Entry point - launches TUI directly
├── internal/
│   ├── app/
│   │   ├── app.go         # Main TUI application (bubbletea)
│   │   ├── state.go       # Application state
│   │   └── update.go      # Update logic
│   ├── screens/
│   │   ├── home.go        # Main menu screen
│   │   ├── status.go      # Git status screen
│   │   ├── commit.go      # Commit workflow screen
│   │   ├── push.go        # Push workflow screen
│   │   ├── pr.go          # PR creation screen
│   │   ├── issues.go      # Issues management screen
│   │   ├── repos.go       # Repo navigation screen
│   │   ├── ai.go          # AI chat/assistant screen
│   │   └── settings.go    # Settings screen
│   ├── components/
│   │   ├── menu.go        # Selectable menu component
│   │   ├── filelist.go    # Multi-select file list
│   │   ├── textinput.go   # Text input component
│   │   ├── textarea.go    # Multiline input component
│   │   ├── spinner.go      # Loading spinner
│   │   ├── statusbar.go   # Status bar component
│   │   ├── chat.go        # AI chat component
│   │   └── diffviewer.go  # Code diff viewer
│   ├── git/
│   │   ├── status.go      # File status detection
│   │   └── operations.go  # Commit, push operations
│   ├── gh/
│   │   └── client.go      # GitHub CLI wrapper
│   ├── ai/
│   │   ├── provider.go    # AI provider interface
│   │   ├── ollama.go      # Ollama provider (local)
│   │   ├── openai.go      # OpenAI provider
│   │   ├── claude.go      # Anthropic Claude provider
│   │   ├── mlx.go         # MLX provider (Apple Silicon)
│   │   ├── lmstudio.go    # LM Studio provider (local)
│   │   ├── prompts.go     # System prompts for different tasks
│   │   └── config.go      # AI provider configuration
│   └── db/
│       ├── db.go          # Local SQLite connection
│       └── history.go     # Command/chat history storage
├── styles/
│   └── theme.go           # Lipgloss styles
└── .ghx/
    ├── ghx.db             # Local SQLite database
    └── config.json        # AI provider settings
```

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - Pre-built components
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/glamour` - Markdown rendering for AI responses
- `modernc.org/sqlite` - Pure Go SQLite (no CGO)
- `github.com/go-git/go-git/v5` - Git operations
- `github.com/sashabaranov/go-openai` - OpenAI-compatible client (works with Ollama, OpenAI, LM Studio, etc.)
- `github.com/anthropics/anthropic-sdk-go` - Official Claude SDK

## Main Menu

```
$ ghx
┌─────────────────────────────────────────────────┐
│  ghx - AI-Powered GitHub Assistant             │
├─────────────────────────────────────────────────┤
│                                                 │
│  > Status      View git status                  │
│    Commit      Stage & commit files (AI assist) │
│    Diff        View staged changes              │
│    Push        Push to remote                   │
│    PR          Create pull request (AI assist)  │
│    Issues      Manage issues                    │
│    AI Chat     Ask AI about codebase            │
│    History     View recent commands             │
│    Settings    Configure providers & preferences│
│                                                 │
├─────────────────────────────────────────────────┤
│  AI: Ollama (llama3)  ↑/↓ Nav  Enter Go  q Quit │
└─────────────────────────────────────────────────┘
```

## AI Provider Support

### Supported Providers

| Provider | Type | Endpoint | Models |
|----------|------|----------|--------|
| **Ollama** | Local | `http://localhost:11434` | llama3, mistral, codellama, etc. |
| **OpenAI** | Cloud | `https://api.openai.com` | gpt-4o, gpt-4-turbo, gpt-3.5-turbo |
| **Claude** | Cloud | `https://api.anthropic.com` | claude-3-opus, claude-3-sonnet, claude-3-haiku |
| **MLX** | Local (macOS) | `http://localhost:8080` | Apple MLX models |
| **LM Studio** | Local | `http://localhost:1234` | Any GGUF model |
| **Custom** | Any | Configurable | OpenAI-compatible APIs |

### Configuration Screen

```
┌─────────────────────────────────────────────────┐
│  AI Provider Settings                           │
├─────────────────────────────────────────────────┤
│                                                 │
│  Active Provider: [Ollama        ▼]             │
│                                                 │
│  ── Ollama Settings ──                          │
│  Host:     [http://localhost:11434  ]           │
│  Model:    [llama3                  ▼]         │
│                                                 │
│  ── OpenAI Settings ──                          │
│  API Key:  [sk-****                      ]     │
│  Model:    [gpt-4o                     ▼]       │
│                                                 │
│  ── Claude Settings ──                          │
│  API Key:  [sk-ant-****                ]       │
│  Model:    [claude-3-sonnet             ▼]      │
│                                                 │
│  ── LM Studio Settings ──                       │
│  Host:     [http://localhost:1234    ]          │
│                                                 │
├─────────────────────────────────────────────────┤
│  t Test Connection  s Save  b Back              │
└─────────────────────────────────────────────────┘
```

## AI-Assisted Workflows

### Smart Commit Messages

```
┌─────────────────────────────────────────────────┐
│  Commit - Select files                          │
├─────────────────────────────────────────────────┤
│  [x] internal/ai/provider.go                    │
│  [x] internal/ai/ollama.go                      │
│  [ ] main.go                                    │
│                                             │
├─────────────────────────────────────────────────┤
│  Commit message:                                │
│  Add Ollama integration for local AI inference │
│  _________________________________________      │
│                                             │
│  ┌─ AI Suggestions ────────────────────────┐   │
│  │ 1. Add Ollama provider with streaming    │   │
│  │    response support                     │   │
│  │ 2. Implement AI provider interface      │   │
│  │ 3. feat(ai): add local Ollama support   │   │
│  └─────────────────────────────────────────┘   │
│                                             │
├─────────────────────────────────────────────────┤
│  Tab AI Suggest  r Regenerate  Enter Commit     │
└─────────────────────────────────────────────────┘
```

### AI-Generated PR Descriptions

```
┌─────────────────────────────────────────────────┐
│  Create Pull Request - AI Assisted              │
├─────────────────────────────────────────────────┤
│  Base: main                                     │
│  Head: feature/ai-integration                   │
│                                             │
│  Title: Add multi-provider AI integration      │
│  _________________________________________      │
│                                             │
│  Description:                                   │
│  ┌─────────────────────────────────────────────┐│
│  │ ## Summary                           ││
│  │ This PR adds support for multiple AI providers│
│  │ (Ollama, OpenAI, Claude, MLX, LM Studio).    ││
│  │                                        ││
│  │ ## Changes                             ││
│  │ - Add provider interface for AI abstraction│  │
│  │ - Implement Ollama provider (local)    ││
│  │ - Implement OpenAI provider             ││
│  │ - Add configuration management          ││
│  │                                        ││
│  │ ## Testing                            ││
│  │ - Tested with Ollama llama3            ││
│  │ - Verified streaming responses         ││
│  └─────────────────────────────────────────┘   │
│                                             │
├─────────────────────────────────────────────────┤
│  Tab Fields  g AI Generate  Enter Create        │
└─────────────────────────────────────────────────┘
```

### AI Chat Screen

```
┌─────────────────────────────────────────────────┐
│  AI Assistant (Ollama - llama3)                 │
├─────────────────────────────────────────────────┤
│                                                 │
│  You: What does the provider.go file do?       │
│                                                 │
│  AI: The `provider.go` file defines the core   │
│  interface for all AI providers in ghx. It    │
│  includes:                                      │
│                                                 │
│  - `Provider` interface with `Chat()`,         │
│    `Stream()`, and `Configure()` methods       │
│  - `Message` and `Response` structs            │
│  - Common error types                          │
│                                                 │
│  You: Generate a commit message for my changes │
│                                                 │
│  AI: Based on your staged changes:             │
│  - Modified: internal/ai/ollama.go             │
│  - Added: internal/ai/prompts.go               │
│                                                 │
│  Suggested: "Add streaming support to Ollama   │
│  provider and extract system prompts"          │
│                                                 │
├─────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────┐   │
│  │ Ask AI anything...                [│    ]   │
│  └─────────────────────────────────────────┘   │
├─────────────────────────────────────────────────┤
│  Ctrl+L Clear  Ctrl+S Stop  Esc Back           │
└─────────────────────────────────────────────────┘
```

### Code Explainer

```
┌─────────────────────────────────────────────────┐
│  Diff Viewer - internal/ai/ollama.go           │
├─────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────┐│
│  │  1  +func (o *Ollama) Stream(ctx context.Context,│
│  │  2  +  messages []Message) (<-chan string, error) {│
│  │  3  +  // ... streaming implementation      ││
│  │  4  +}                                   ││
│  └─────────────────────────────────────────┘   │
│                                                 │
│  ┌─ AI Explanation ───────────────────────────┐ │
│  │ The `Stream` method implements real-time   │ │
│  │ response streaming from Ollama. It returns │ │
│  │ a read-only channel that emits tokens as   │ │
│  │ they're generated, enabling a chat-like   │ │
│  │ experience in the TUI.                     │ │
│  └─────────────────────────────────────────┘   │
│                                             │
├─────────────────────────────────────────────────┤
│  e Explain  r Review  s Suggest  b Back         │
└─────────────────────────────────────────────────┘
```

## Key Bindings

### Global
- `q` / `Ctrl+C` - Quit / Go back
- `?` - Show help overlay
- `Tab` - Next field / section
- `Shift+Tab` - Previous field
- `/` - Search

### Navigation
- `↑` / `k` - Move up
- `↓` / `j` - Move down
- `Enter` - Select / Confirm
- `Esc` - Cancel / Back

### AI Specific
- `g` - Generate with AI
- `r` - Regenerate AI response
- `e` - Explain code with AI
- `Ctrl+L` - Clear AI chat
- `Ctrl+S` - Stop AI generation

## Implementation Phases

### Phase 1: TUI Foundation
- [ ] Initialize Go module
- [ ] Set up bubbletea application structure
- [ ] Implement screen routing
- [ ] Create base components

### Phase 2: Git Integration
- [ ] File status detection with go-git
- [ ] Status screen with file list
- [ ] Diff viewing

### Phase 3: AI Provider System
- [ ] Define provider interface
- [ ] Implement Ollama provider (local, no API key)
- [ ] Implement OpenAI provider
- [ ] Implement Claude provider
- [ ] Implement LM Studio provider (OpenAI-compatible)
- [ ] Add MLX support for Apple Silicon
- [ ] Configuration persistence

### Phase 4: AI-Assisted Workflows
- [ ] Generate commit messages from diffs
- [ ] Generate PR descriptions from commits
- [ ] Code explanation on demand
- [ ] Streaming responses in TUI

### Phase 5: AI Chat Interface
- [ ] Chat screen with history
- [ ] Context-aware questions (current files, diffs)
- [ ] Markdown rendering for responses
- [ ] Chat history persistence

### Phase 6: Polish
- [ ] Error handling
- [ ] Provider fallback/switching
- [ ] Rate limiting awareness
- [ ] Theme support
- [ ] Build release binaries

## Configuration File: `.ghx/config.json`

```json
{
  "ai": {
    "active_provider": "ollama",
    "providers": {
      "ollama": {
        "host": "http://localhost:11434",
        "model": "llama3",
        "options": {
          "temperature": 0.7,
          "num_ctx": 4096
        }
      },
      "openai": {
        "api_key": "sk-...",
        "model": "gpt-4o",
        "base_url": ""
      },
      "claude": {
        "api_key": "sk-ant-...",
        "model": "claude-3-sonnet-20240229"
      },
      "lmstudio": {
        "host": "http://localhost:1234",
        "model": "local-model"
      },
      "mlx": {
        "host": "http://localhost:8080",
        "model": ""
      }
    }
  },
  "ui": {
    "theme": "dark",
    "show_ai_suggestions": true
  }
}
```

## AI Prompts

### Commit Message Generation
```
You are a git commit message generator. Analyze the following diff and generate a concise, conventional commit message. Format: <type>(<scope>): <description>

Types: feat, fix, docs, style, refactor, test, chore

Diff:
{diff_content}

Generate 3 commit message options.
```

### PR Description Generation
```
You are a pull request description generator. Based on the following commits and changes, generate a comprehensive PR description including:
- Summary
- Key changes (bullet points)
- Testing notes
- Breaking changes (if any)

Commits:
{commit_messages}

Changes:
{diff_summary}
```

### Code Explanation
```
You are a code explainer. Explain the following code changes in simple terms. Focus on:
- What the code does
- Why it was changed
- Any potential impacts

Code:
{code_snippet}
```

## Usage

```bash
# Just run it
ghx

# Everything is interactive - no commands to memorize

# Local-only mode (Ollama) - works completely offline
# Set active provider to Ollama in settings
```

## Offline Support

With **Ollama** or **LM Studio** as the active provider:
- Fully offline AI assistance
- No API keys required
- No internet needed
- Privacy-first (code never leaves machine)
