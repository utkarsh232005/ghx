package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

type RepoInfo struct {
	Branch  string
	Remotes []string
}

type CommandResult struct {
	Output string
}

func Info(path string) (RepoInfo, error) {
	repo, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return RepoInfo{}, err
	}

	info := RepoInfo{Branch: "detached"}
	head, err := repo.Head()
	if err == nil {
		if head.Name().IsBranch() {
			info.Branch = head.Name().Short()
		} else {
			info.Branch = head.Hash().String()[:7]
		}
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return info, err
	}
	for _, remote := range remotes {
		urls := remote.Config().URLs
		if len(urls) == 0 {
			info.Remotes = append(info.Remotes, remote.Config().Name)
			continue
		}
		info.Remotes = append(info.Remotes, remote.Config().Name+" -> "+strings.Join(urls, ", "))
	}

	return info, nil
}

func Diff(paths []string, staged bool) (string, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--staged")
	}
	if len(paths) > 0 {
		args = append(args, "--")
		args = append(args, paths...)
	}
	out, err := runGit(args...)
	if err != nil || strings.TrimSpace(out) != "" || len(paths) != 1 || staged {
		return out, err
	}
	data, readErr := os.ReadFile(paths[0])
	if readErr != nil {
		return out, nil
	}
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		lines[i] = "+" + line
	}
	return "diff --git a/" + paths[0] + " b/" + paths[0] + "\nnew file mode 100644\n--- /dev/null\n+++ b/" + paths[0] + "\n" + strings.Join(lines, "\n"), nil
}

func Commit(paths []string, message string) (CommandResult, error) {
	if len(paths) == 0 {
		return CommandResult{}, fmt.Errorf("select at least one file to commit")
	}
	if strings.TrimSpace(message) == "" {
		return CommandResult{}, fmt.Errorf("commit message cannot be empty")
	}

	addArgs := append([]string{"add", "--"}, paths...)
	addOut, err := runGit(addArgs...)
	if err != nil {
		return CommandResult{Output: addOut}, err
	}

	commitOut, err := runGit("commit", "-m", strings.TrimSpace(message))
	return CommandResult{Output: strings.TrimSpace(addOut + "\n" + commitOut)}, err
}

func Push() (CommandResult, error) {
	out, err := runGit("push")
	return CommandResult{Output: out}, err
}

func CheckoutBranch(name string) (CommandResult, error) {
	out, err := runGit("checkout", "-b", name)
	return CommandResult{Output: out}, err
}

func SuggestedCommitMessage(files []FileStatus, paths []string) string {
	selected := map[string]bool{}
	for _, path := range paths {
		selected[path] = true
	}

	added := 0
	modified := 0
	deleted := 0
	for _, file := range files {
		if len(selected) > 0 && !selected[file.Path] {
			continue
		}
		status := file.ShortStatus()
		switch {
		case strings.Contains(status, "A") || strings.Contains(status, "?"):
			added++
		case strings.Contains(status, "D"):
			deleted++
		default:
			modified++
		}
	}

	total := added + modified + deleted
	if total == 0 {
		total = len(paths)
	}
	if total == 1 && len(paths) == 1 {
		return "chore: update " + paths[0]
	}
	if added > modified && added > deleted {
		return fmt.Sprintf("feat: add %d file updates", total)
	}
	if deleted > 0 && deleted >= modified {
		return fmt.Sprintf("chore: remove %d file updates", total)
	}
	return fmt.Sprintf("chore: update %d files", total)
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())
	errOut := strings.TrimSpace(stderr.String())
	if err != nil {
		if errOut != "" {
			return strings.TrimSpace(out + "\n" + errOut), fmt.Errorf(errOut)
		}
		return out, err
	}
	if errOut != "" {
		return strings.TrimSpace(out + "\n" + errOut), nil
	}
	return out, nil
}
