package ai

import (
	"fmt"
	"strings"
)

func GenerateCommitMessagePrompt(diff string) string {
	return fmt.Sprintf(`You are a git commit message generator. Analyze the following diff and generate a concise, conventional commit message.

Format: <type>(<scope>): <description>

Types: feat, fix, docs, style, refactor, test, chore

Generate 3 commit message options. Each on a new line with a number.

Rules:
- Keep the first line under 72 characters
- Use imperative mood ("add" not "added")
- No period at the end
- If there are breaking changes, start with "BREAKING CHANGE:"

Diff:
%s`, diff)
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
