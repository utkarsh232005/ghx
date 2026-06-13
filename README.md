# ghx

[![Open in Bolt](https://bolt.new/static/open-in-bolt.svg)](https://bolt.new/~/sb1-znfc2vuh)

AI-Powered GitHub Workflow Assistant - A Go TUI tool for GitHub workflows with AI integration.

## Features

- **Interactive TUI** - Navigate with keyboard, no commands to memorize
- **AI Integration** - Generate commit messages, PR descriptions, and more
- **Multiple AI Providers**:
  - **Local/Offline**: Ollama, LM Studio, MLX
  - **Cloud**: OpenAI, Claude
- **Git Operations**: Status, commit, push, branch management
- **GitHub Integration**: PR creation, issues, repos
- **Fully Offline**: Works with local AI providers, no internet required

## Installation

```bash
git clone https://github.com/KDM-cli/ghx.git
cd ghx
go build -o ghx .
```

## Usage

```bash
./ghx
```

Navigate with arrow keys, press Enter to select.

## AI Providers

### Ollama (Recommended for offline)
```bash
curl -fsSL https://ollama.com/install.sh | sh
ollama pull llama3
```

### OpenAI / Claude
Set your API key in Settings screen.

## Key Bindings

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate |
| `Enter` | Select/Confirm |
| `Tab` | Next section |
| `Esc` | Back |
| `q` | Quit |
| `Space` | Toggle selection |
| `g` | Generate with AI |
| `?` | Help |

## License

MIT
