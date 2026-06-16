package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func Stage(files []string) error {
	if len(files) == 0 {
		return nil
	}

	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

func Unstage(files []string) error {
	if len(files) == 0 {
		return nil
	}

	args := append([]string{"reset", "HEAD", "--"}, files...)
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}

func CommitWithFiles(message string, files []string) error {
	if err := Stage(files); err != nil {
		return err
	}
	return Commit(message)
}

func Push(remote, branch string) error {
	if remote == "" {
		remote = "origin"
	}
	if branch == "" {
		var err error
		branch, err = GetCurrentBranch()
		if err != nil {
			return err
		}
	}

	cmd := exec.Command("git", "push", remote, branch)
	return cmd.Run()
}

func Pull(remote, branch string) error {
	if remote == "" {
		remote = "origin"
	}
	if branch == "" {
		var err error
		branch, err = GetCurrentBranch()
		if err != nil {
			return err
		}
	}

	cmd := exec.Command("git", "pull", remote, branch)
	return cmd.Run()
}

func RunCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

func CreateBranch(name string) error {
	cmd := exec.Command("git", "checkout", "-b", name)
	return cmd.Run()
}

func CheckoutBranch(name string) error {
	cmd := exec.Command("git", "checkout", name)
	return cmd.Run()
}

func GetBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var branches []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		branch := strings.TrimSpace(scanner.Text())
		if branch != "" {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

func GetRemoteBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-r", "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var branches []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		branch := strings.TrimSpace(scanner.Text())
		if branch != "" && strings.Contains(branch, "/") && !strings.Contains(branch, "HEAD") {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

type CommitInfo struct {
	Hash    string
	Message string
	Author  string
	Date    string
}

func GetLog(limit int) ([]CommitInfo, error) {
	cmd := exec.Command("git", "log", "-n", fmt.Sprintf("%d", limit), "--oneline")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []CommitInfo
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			commits = append(commits, CommitInfo{
				Hash:    parts[0],
				Message: parts[1],
			})
		}
	}

	return commits, nil
}

func GetDefaultBranch() string {
	// 1. Try refs/remotes/origin/HEAD
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	out, err := cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(out))
		parts := strings.Split(ref, "/")
		if len(parts) >= 3 {
			branch := strings.Join(parts[len(parts)-2:], "/")
			if showRefVerify("refs/remotes/" + branch) {
				return branch
			}
		}
	}

	// 2. Try common local branches
	for _, b := range []string{"main", "master"} {
		if showRefVerify("refs/heads/" + b) {
			return b
		}
	}

	// 3. Try common remote branches
	for _, b := range []string{"origin/main", "origin/master"} {
		if showRefVerify("refs/remotes/" + b) {
			return b
		}
	}

	// 4. Try getting current branch as fallback
	curr, err := GetCurrentBranch()
	if err == nil && curr != "" {
		return curr
	}

	return "main"
}

func showRefVerify(ref string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", ref)
	return cmd.Run() == nil
}

func GetCommitsBetween(base, head string, limit int) ([]CommitInfo, error) {
	if head == "" {
		head = "HEAD"
	}

	resolvedBase := base
	if resolvedBase == "" {
		resolvedBase = GetDefaultBranch()
	}

	currBranch, _ := GetCurrentBranch()
	if resolvedBase == head || resolvedBase == currBranch {
		defaultBase := GetDefaultBranch()
		if defaultBase != head && defaultBase != currBranch {
			resolvedBase = defaultBase
		} else {
			if currBranch == "main" {
				resolvedBase = "origin/main"
			} else {
				resolvedBase = "main"
			}
		}
	}

	// Resolve merge-base to avoid branch mismatch errors
	var logTarget string
	cmd := exec.Command("git", "merge-base", resolvedBase, head)
	mBaseBytes, err := cmd.Output()
	if err == nil {
		mBase := strings.TrimSpace(string(mBaseBytes))
		if mBase != "" {
			logTarget = fmt.Sprintf("%s..%s", mBase, head)
		}
	}
	if logTarget == "" {
		logTarget = fmt.Sprintf("%s..%s", resolvedBase, head)
	}

	args := []string{"log", logTarget, "--oneline"}
	if limit > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", limit))
	}

	cmd = exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		// Fallback to local git log of HEAD, limited to last 5 commits
		fallbackArgs := []string{"log", "-n", "5", "--oneline"}
		cmd = exec.Command("git", fallbackArgs...)
		output, err = cmd.Output()
		if err != nil {
			return nil, err
		}
	}

	var commits []CommitInfo
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			commits = append(commits, CommitInfo{
				Hash:    parts[0],
				Message: parts[1],
			})
		}
	}

	return commits, nil
}

func GetDiffStat(base, head string) (string, error) {
	if head == "" {
		head = "HEAD"
	}
	resolvedBase := base
	if resolvedBase == "" {
		resolvedBase = GetDefaultBranch()
	}

	currBranch, _ := GetCurrentBranch()
	if resolvedBase == head || resolvedBase == currBranch {
		defaultBase := GetDefaultBranch()
		if defaultBase != head && defaultBase != currBranch {
			resolvedBase = defaultBase
		} else {
			if currBranch == "main" {
				resolvedBase = "origin/main"
			} else {
				resolvedBase = "main"
			}
		}
	}

	// Try getting merge base diff first
	cmd := exec.Command("git", "merge-base", resolvedBase, head)
	mBaseBytes, err := cmd.Output()
	var diffCmd *exec.Cmd
	if err == nil {
		mBase := strings.TrimSpace(string(mBaseBytes))
		if mBase != "" {
			diffCmd = exec.Command("git", "diff", "--stat", fmt.Sprintf("%s..%s", mBase, head))
		}
	}

	if diffCmd == nil {
		diffCmd = exec.Command("git", "diff", "--stat", fmt.Sprintf("%s...%s", resolvedBase, head))
	}

	output, err := diffCmd.Output()
	if err != nil {
		// Fallback to simple git diff --stat HEAD
		cmd = exec.Command("git", "diff", "--stat", "HEAD")
		output, err = cmd.Output()
		if err != nil {
			return "", err
		}
	}

	return string(output), nil
}


func GetRemotes() ([]string, error) {
	cmd := exec.Command("git", "remote")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var remotes []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		remote := strings.TrimSpace(scanner.Text())
		if remote != "" {
			remotes = append(remotes, remote)
		}
	}

	return remotes, nil
}

func GetRemoteURL(remote string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remote)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func ParseNWOFromURL(urlStr string) string {
	urlStr = strings.TrimSuffix(urlStr, ".git")
	if strings.HasPrefix(urlStr, "git@") {
		parts := strings.SplitN(urlStr, ":", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	} else {
		parts := strings.Split(urlStr, "/")
		if len(parts) >= 2 {
			return parts[len(parts)-2] + "/" + parts[len(parts)-1]
		}
	}
	return ""
}

func AddRemote(name, urlStr string) error {
	urlStr = ExpandRemoteURL(urlStr)
	cmd := exec.Command("git", "remote", "add", name, urlStr)
	return cmd.Run()
}

func ExpandRemoteURL(urlStr string) string {
	if !strings.Contains(urlStr, "://") && !strings.Contains(urlStr, "@") && strings.Contains(urlStr, "/") {
		return fmt.Sprintf("https://github.com/%s.git", urlStr)
	}
	return urlStr
}
