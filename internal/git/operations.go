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
		if branch != "" {
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

func GetCommitsBetween(base, head string, limit int) ([]CommitInfo, error) {
	if base == "" {
		base = "origin/main"
	}
	if head == "" {
		head = "HEAD"
	}

	args := []string{"log", fmt.Sprintf("%s..%s", base, head), "--oneline"}
	if limit > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", limit))
	}
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		// Fallback to local git log of HEAD
		fallbackArgs := []string{"log"}
		if limit > 0 {
			fallbackArgs = append(fallbackArgs, "-n", fmt.Sprintf("%d", limit))
		}
		fallbackArgs = append(fallbackArgs, "--oneline")
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
	if base == "" {
		base = "origin/main"
	}
	if head == "" {
		head = "HEAD"
	}

	cmd := exec.Command("git", "diff", "--stat", fmt.Sprintf("%s...%s", base, head))
	output, err := cmd.Output()
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
