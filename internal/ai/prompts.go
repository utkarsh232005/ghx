package ai

import (
	"fmt"
	"strings"
)

func GenerateCommitMessagePrompt(diff string) string {
	return fmt.Sprintf(`You are a git commit message generator. Analyze the following diff and generate exactly 3 concise, conventional commit message options.

Format: <type>(<scope>): <description>

Types: feat, fix, docs, style, refactor, test, chore

Rules:
1. POV: Write the message from the first-person perspective of the user committing the changes. Use imperative mood (e.g., "add features", "fix bug", "refactor styles"). Do NOT write third-person explanations or descriptions (e.g., do NOT say "This commit adds...", "A commit that...", or "This refactoring...").
2. Content ONLY: Return ONLY the raw commit message options. Do NOT include any introduction, explanations, conversational filler, markdown formatting (like backticks or code blocks), or JSON wrapper.
3. Formatting: Return exactly 3 options, each on a new line, prefixed with its number (e.g., "1. feat(scope): description"). Keep the first line of each option under 72 characters and do not end with a period.

Diff:
%s`, diff)
}

func GeneratePRTitlePrompt(commits, diffSummary string) string {
	return fmt.Sprintf(`You are a pull request title generator. Based on the following commits and changes, generate a concise, conventional PR title (e.g. "feat(ai): add model customization option").

Rules:
- Keep the title under 70 characters
- Do not end with a period
- Return ONLY the raw title. Do NOT include any explanations, introduction, quotes, or conversational filler.

Commits:
%s

Changes:
%s`, commits, diffSummary)
}

func GeneratePRDescriptionPrompt(commits, diffSummary string) string {
	return fmt.Sprintf(`You are a pull request description generator. Based on the following commits and changes, generate a comprehensive PR description.

Include:
- Summary (2-3 sentences)
- Key changes (bullet points)
- Testing notes
- Breaking changes (if any breaking changes (if any)

Format the response in markdown.

Commits:
%s

Changes:
%s`, commits, diffSummary)
}

func ExplainCodePrompt(code, language string) string {
	return fmt.Sprintf(`You are a code explainer. Explain the following %s code changes in simple terms.

Focus on:
- What the code does
- Why it was changed (if you can tell)
- Any potential impacts

Keep the explanation concise and clear.

Code:
%s`, language, code)
}

func ChatWithContextPrompt(context, question string) string {
	return fmt.Sprintf(`You are an AI assistant helping with a codebase. Here is some context about the current state:

%s

User question: %s

Provide a helpful, concise answer. If code examples would help, include them with proper syntax highlighting.`, context, question)
}

type FileContext struct {
	Path    string
	Status  string
	Content string
}

func BuildContext(files []FileContext) string {
	var sb strings.Builder
	for _, f := range files {
		sb.WriteString(fmt.Sprintf("\n--- %s (%s) ---\n", f.Path, f.Status))
		if f.Content != "" {
			sb.WriteString(f.Content)
		}
	}
	return sb.String()
}
