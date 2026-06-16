package gh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) IsAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

func (c *Client) run(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func (c *Client) CreatePR(title, body, base, head string, draft bool, repo string) (*PRInfo, error) {
	args := []string{"pr", "create"}

	args = append(args, "--title", title)
	args = append(args, "--body", body)
	if base != "" {
		args = append(args, "--base", base)
	}
	if head != "" {
		args = append(args, "--head", head)
	}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	if draft {
		args = append(args, "--draft")
	}

	output, err := c.run(args...)
	if err != nil {
		return nil, err
	}

	prURL := strings.TrimSpace(output)
	return &PRInfo{
		URL:    prURL,
		Number: extractPRNumber(prURL),
	}, nil
}

func extractPRNumber(url string) int {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		var num int
		fmt.Sscanf(parts[len(parts)-1], "%d", &num)
		return num
	}
	return 0
}

type PRInfo struct {
	URL    string
	Number int
	Title  string
	State  string
}

func (c *Client) ListPRs(limit int) ([]PRInfo, error) {
	args := []string{"pr", "list", "--limit", fmt.Sprintf("%d", limit), "--json", "number,title,url,state"}
	output, err := c.run(args...)
	if err != nil {
		return nil, err
	}

	var prs []PRInfo
	if err := json.Unmarshal([]byte(output), &prs); err != nil {
		return nil, err
	}

	return prs, nil
}

type IssueInfo struct {
	Number int
	Title  string
	URL    string
	State  string
}

func (c *Client) ListIssues(limit int) ([]IssueInfo, error) {
	args := []string{"issue", "list", "--limit", fmt.Sprintf("%d", limit), "--json", "number,title,url,state"}
	output, err := c.run(args...)
	if err != nil {
		return nil, err
	}

	var issues []IssueInfo
	if err := json.Unmarshal([]byte(output), &issues); err != nil {
		return nil, err
	}

	return issues, nil
}

type RepoInfo struct {
	Name        string
	Description string
	URL         string
	IsPrivate   bool
}

func (c *Client) ListRepos(limit int) ([]RepoInfo, error) {
	args := []string{"repo", "list", "--limit", fmt.Sprintf("%d", limit), "--json", "name,description,url,isPrivate"}
	output, err := c.run(args...)
	if err != nil {
		return nil, err
	}

	var repos []RepoInfo
	if err := json.Unmarshal([]byte(output), &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

func (c *Client) OpenRepoInBrowser(repo string) error {
	cmd := exec.Command("gh", "repo", "view", repo, "--web")
	return cmd.Start()
}
