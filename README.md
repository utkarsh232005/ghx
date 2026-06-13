<div align="center">
  <h1>🛸 ghx</h1>
  <p><strong>AI-Powered GitHub Workflow TUI Assistant</strong></p>

  <p>
    <a href="https://github.com/KDM-cli/ghx/blob/main/LICENSE"><img src="https://img.shields.io/github/license/KDM-cli/ghx?style=for-the-badge&logo=mit" alt="license"/></a>
    <a href="https://go.dev"><img src="https://img.shields.io/badge/Made%20with-Go-00ADD8?style=for-the-badge&logo=go" alt="go"/></a>
  </p>
  
  <p>A fast, keyboard-driven Terminal User Interface (TUI) tool written in Go to automate your Git and GitHub workflows with integrated offline and cloud AI generation.</p>
</div>

<hr />

## 🌟 Key Features

<table width="100%">
  <tr>
    <td width="50%">
      <h3>⌨️ Keyboard-Driven UI</h3>
      <p>Seamless navigation with arrow keys and shortcut keys. No complex command syntax to memorize.</p>
    </td>
    <td width="50%">
      <h3>🤖 Deep AI Integration</h3>
      <p>Generate highly descriptive conventional commits, PR titles, and descriptions dynamically from actual diff statistics.</p>
    </td>
  </tr>
  <tr>
    <td width="50%">
      <h3>🔌 Multi-Provider AI Engine</h3>
      <p>Supports both local/offline providers (<b>Ollama</b>, <b>LM Studio</b>, <b>MLX</b>) and cloud providers (<b>OpenAI</b>, <b>Claude</b>).</p>
    </td>
    <td width="50%">
      <h3>📦 Remote Target Control</h3>
      <p>Resolve and toggle PR target repositories (e.g., origin fork vs upstream parent) and configure missing remotes right in the TUI.</p>
    </td>
  </tr>
</table>

<hr />

## 🚀 Installation & Build

### Prerequisites
- Go 1.21 or higher
- [GitHub CLI (gh)](https://cli.github.com/) installed and authenticated

### Build from Source
```bash
# Clone the repository
git clone https://github.com/KDM-cli/ghx.git
cd ghx

# Build the binary
go build -o ghx .
```

<hr />

## ⚙️ AI Configuration

### Local/Offline Models (Ollama)
We recommend running [Ollama](https://ollama.com) locally for a private, zero-latency offline workflow.
```bash
# Install and start Ollama, then pull llama3
ollama pull llama3
```

### Configure Settings
1. Launch `./ghx`.
2. Navigate to **Settings** and select your preferred active AI provider.
3. Configure API Keys (for cloud providers) or choose custom models inline by pressing <kbd>m</kbd>.

<hr />

## ⌨️ Keyboard Bindings

<details open>
<summary><b>Navigation & General controls</b></summary>

| Key | Description |
| :---: | :--- |
| <kbd>↑ / ↓</kbd> or <kbd>k / j</kbd> | Navigate options / scroll lists |
| <kbd>Tab</kbd> / <kbd>Shift+Tab</kbd> | Focus next / previous section or step |
| <kbd>Enter</kbd> | Select menu option or confirm step |
| <kbd>Esc</kbd> or <kbd>b</kbd> | Go back to previous screen or step |
| <kbd>q</kbd> | Quit application (from Home Screen) |
| <kbd>?</kbd> | Toggle Help view |

</details>

<details>
<summary><b>Contextual Actions</b></summary>

| Key | Context | Description |
| :---: | :--- | :--- |
| <kbd>Space</kbd> | Commit Stage list | Toggle staged state of selected file |
| <kbd>a</kbd> / <kbd>n</kbd> | Commit Stage list | Stage all files / unstage all files |
| <kbd>g</kbd> | Commit / PR screens | Generate suggestions using the configured AI model |
| <kbd>t</kbd> | PR review | Toggle target repository (e.g. `upstream` <-> `origin`) |
| <kbd>u</kbd> | PR review | Set/add a new `upstream` remote repository |
| <kbd>d</kbd> | PR review | Toggle Draft state of the pull request |
| <kbd>1 - 3</kbd> | Commit AI suggestions | Select the corresponding AI suggestion |
| <kbd>r</kbd> | Commit AI suggestions | Regenerate AI suggestions |

</details>

<hr />

## 📄 License

Distributed under the MIT License. See [LICENSE](https://github.com/KDM-cli/ghx/blob/main/LICENSE) for more information.
