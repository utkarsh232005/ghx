package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type FileStatus struct {
	Path   string
	Status string
	Staged bool
	Branch string
}

type GitInfo struct {
	Branch     string
	Remote     string
	Ahead      int
	Behind     int
	HasChanges bool
}

type Status struct {
	Staged    []FileStatus
	Modified  []FileStatus
	Untracked []FileStatus
	Info      GitInfo
}

func GetStatus() (*Status, error) {
	status := &Status{}

	info, err := getGitInfo()
	if err == nil {
		status.Info = *info
	}

	cmd := exec.Command("git", "status", "--porcelain=v1", "--branch")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "## ") {
			continue
		}

		if len(line) < 4 {
			continue
		}

		x := line[0]
		y := line[1]
		path := line[3:]

		file := FileStatus{
			Path:   path,
			Branch: status.Info.Branch,
		}

		if x != ' ' && x != '?' {
			file.Status = string(x)
			file.Staged = true
			status.Staged = append(status.Staged, file)
		}

		if y != ' ' && y != '?' {
			file.Status = string(y)
			file.Staged = false
			status.Modified = append(status.Modified, file)
		}

		if x == '?' && y == '?' {
			file.Status = "?"
			file.Staged = false
			status.Untracked = append(status.Untracked, file)
		}
	}

	status.HasChanges = len(status.Staged) > 0 || len(status.Modified) > 0 || len(status.Untracked) > 0

	return status, nil
}

func getGitInfo() (*GitInfo, error) {
	info := &GitInfo{}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	info.Branch = strings.TrimSpace(string(output))

	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	output, err = cmd.Output()
	if err == nil {
		info.Remote = strings.TrimSpace(string(output))
	}

	if info.Remote != "" {
		cmd = exec.Command("git", "rev-list", "--left-right", "--count", info.Branch+"..."+info.Remote)
		output, err = cmd.Output()
		if err == nil {
			parts := strings.Fields(string(output))
			if len(parts) == 2 {
				fmt.Sscanf(parts[0], "%d", &info.Ahead)
				fmt.Sscanf(parts[1], "%d", &info.Behind)
			}
		}
	}

	return info, nil
}

func GetDiff(files []string) (string, error) {
	args := []string{"diff", "--color=never"}
	if len(files) > 0 {
		args = append(args, "--")
		args = append(args, files...)
	} else {
		args = []string{"diff", "--cached", "--color=never"}
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func ReadFileContent(path string) (string, error) {
	repoRoot, err := GetRepoRoot()
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(repoRoot, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
