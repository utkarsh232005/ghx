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
	return fmt.Sprintf(`You are a pull request title generator. Based strictly on the actual commits and file changes provided below, generate a single, concise, conventional pull request title.

Rules:
1. Do NOT use placeholder or template titles. The title must reflect the SPECIFIC changes shown in the Commits and Changes section.
2. Format: <type>(<scope>): <short description in imperative mood>
   - Types: feat, fix, docs, style, refactor, test, chore
   - Scope should represent the specific package or component being modified (e.g., "repos", "issues", "navigation", "ai", etc.).
3. Keep the title under 70 characters.
4. Do NOT end the title with a period.
5. Return ONLY the raw title string. Do NOT include any quotes, markdown formatting, explanations, introduction, or conversational filler.

Commits:
%s

Changes:
%s`, commits, diffSummary)
}

func GeneratePRDescriptionPrompt(commits, diffSummary string) string {
	return fmt.Sprintf(`You are a pull request description generator. Based on the following commits and changes, generate a highly detailed, professional PR description in Markdown format.

Structure the description exactly as follows:

# Description
[A detailed 2-3 sentence overview of what this PR introduces and why these changes were made.]

## Type of Change
[Select/keep only the tags that apply to this PR from: [Feature] [Bug Fix] [Refactor] [Chore] [Documentation] [Test] [Breaking Change]]

## Key Changes
[List the detailed changes using markdown checklist format:
- [ ] Added X...
- [ ] Fixed Y...]

## Testing Details
[Detail the test coverage and verification steps performed.]

## Breaking Changes / Impact
[Specify if there are any breaking changes or migration requirements. If none, write "None".]

Formatting Rules:
- Output the response in clean, raw Markdown.
- Do NOT wrap the entire response in triple backticks or markdown code blocks. Output the raw headers and text directly.

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
