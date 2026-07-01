package commands

type ParameterType string

const (
	ParamString ParameterType = "string"
	ParamBool   ParameterType = "bool"
	ParamChoice ParameterType = "choice"
)

type Parameter struct {
	Name         string        `json:"name"`
	Flag         string        `json:"flag"` // e.g. "-m" or "--global" or "" (for positional)
	Type         ParameterType `json:"type"`
	Description  string        `json:"description"`
	Required     bool          `json:"required"`
	DefaultValue string        `json:"default_value"`
	Choices      []string      `json:"choices,omitempty"`
}

type CommandDef struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	CommandBase     string      `json:"command_base"` // "git" or "gh"
	SubCommand      string      `json:"sub_command"`  // e.g. "checkout", "pr create"
	ArgsTemplate    []string    `json:"args_template"`
	Parameters      []Parameter `json:"parameters"`
	Destructive     bool        `json:"destructive"`
	RequiresSuspend bool        `json:"requires_suspend"`
}

type CommandGroup struct {
	Name     string       `json:"name"`
	IsGit    bool         `json:"is_git"`
	Commands []CommandDef `json:"commands"`
}

func GetCatalog() []CommandGroup {
	return []CommandGroup{
		{
			Name:  "Repository (Git)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git init",
					Description: "Create an empty Git repository or reinitialize an existing one",
					CommandBase: "git",
					SubCommand:  "init",
					Parameters: []Parameter{
						{Name: "Bare", Flag: "--bare", Type: ParamBool, Description: "Create a bare repository", Required: false},
					},
				},
				{
					Name:        "git clone",
					Description: "Clone a repository into a new directory",
					CommandBase: "git",
					SubCommand:  "clone",
					Parameters: []Parameter{
						{Name: "Repository URL/Path", Flag: "", Type: ParamString, Description: "URL or local path of repository to clone", Required: true},
						{Name: "Bare", Flag: "--bare", Type: ParamBool, Description: "Make a bare Git repository", Required: false},
					},
				},
				{
					Name:        "git worktree add",
					Description: "Manage multiple working trees",
					CommandBase: "git",
					SubCommand:  "worktree add",
					Parameters: []Parameter{
						{Name: "Path", Flag: "", Type: ParamString, Description: "Directory path for the new worktree", Required: true},
						{Name: "Commit/Branch", Flag: "", Type: ParamString, Description: "Branch or commit to checkout in the new worktree", Required: false},
					},
				},
				{
					Name:        "git archive",
					Description: "Create an archive of files from a named tree",
					CommandBase: "git",
					SubCommand:  "archive",
					Parameters: []Parameter{
						{Name: "Format", Flag: "--format", Type: ParamChoice, Description: "Format of the resulting archive", Required: false, Choices: []string{"tar", "zip", "tar.gz"}},
						{Name: "Output File", Flag: "--output", Type: ParamString, Description: "Write the archive to this file", Required: false},
						{Name: "Tree-ish", Flag: "", Type: ParamString, Description: "The tree or commit to archive (defaults to HEAD)", Required: true, DefaultValue: "HEAD"},
					},
				},
				{
					Name:        "git bundle create",
					Description: "Move objects and refs by archive",
					CommandBase: "git",
					SubCommand:  "bundle create",
					Parameters: []Parameter{
						{Name: "File Path", Flag: "", Type: ParamString, Description: "Path to output bundle file", Required: true},
						{Name: "Branch/Range", Flag: "", Type: ParamString, Description: "Branch or revision range to bundle (e.g. main)", Required: true, DefaultValue: "main"},
					},
				},
			},
		},
		{
			Name:  "Configuration (Git)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git config",
					Description: "Get and set repository or global options",
					CommandBase: "git",
					SubCommand:  "config",
					Parameters: []Parameter{
						{Name: "Scope", Flag: "", Type: ParamChoice, Description: "Configuration scope", Required: true, DefaultValue: "--global", Choices: []string{"--global", "--system", "--local"}},
						{Name: "Action", Flag: "", Type: ParamChoice, Description: "Action to perform", Required: true, DefaultValue: "--list", Choices: []string{"--list", "--get", "--unset"}},
						{Name: "Key", Flag: "", Type: ParamString, Description: "Setting name (e.g. user.name)", Required: false},
						{Name: "Value", Flag: "", Type: ParamString, Description: "Setting value (required when setting a key)", Required: false},
					},
				},
			},
		},
		{
			Name:  "Status & Add (Git)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git status",
					Description: "Show the working tree status",
					CommandBase: "git",
					SubCommand:  "status",
					Parameters: []Parameter{
						{Name: "Format", Flag: "", Type: ParamChoice, Description: "Output format", Required: false, DefaultValue: "--short", Choices: []string{"--short", "-s", "--porcelain", "Default"}},
					},
				},
				{
					Name:        "git add",
					Description: "Add file contents to the index",
					CommandBase: "git",
					SubCommand:  "add",
					Parameters: []Parameter{
						{Name: "Target", Flag: "", Type: ParamChoice, Description: "Files to add", Required: true, DefaultValue: ".", Choices: []string{".", "-A", "*", "--all", "--patch", "-p", "-i", "-u"}},
						{Name: "Specific File", Flag: "", Type: ParamString, Description: "Specific file path (overrides target choice if provided)", Required: false},
					},
					RequiresSuspend: true, // For interactive add (-i, -p)
				},
			},
		},
		{
			Name:  "Commit (Git)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git commit",
					Description: "Record changes to the repository",
					CommandBase: "git",
					SubCommand:  "commit",
					Parameters: []Parameter{
						{Name: "Message", Flag: "-m", Type: ParamString, Description: "Commit message", Required: false},
						{Name: "Stage All First", Flag: "-a", Type: ParamBool, Description: "Stage modified/deleted files automatically", Required: false},
						{Name: "Amend", Flag: "--amend", Type: ParamBool, Description: "Amend the last commit", Required: false},
						{Name: "Allow Empty", Flag: "--allow-empty", Type: ParamBool, Description: "Allow empty commit", Required: false},
						{Name: "Signoff", Flag: "--signoff", Type: ParamBool, Description: "Add Signed-off-by line", Required: false},
						{Name: "No Edit", Flag: "--no-edit", Type: ParamBool, Description: "Reuse last commit message without editing", Required: false},
						{Name: "Verbose", Flag: "--verbose", Type: ParamBool, Description: "Show diff in commit message template", Required: false},
					},
				},
			},
		},
		{
			Name:  "Branch & Switch (Git)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git branch",
					Description: "List, create, or delete branches",
					CommandBase: "git",
					SubCommand:  "branch",
					Parameters: []Parameter{
						{Name: "List Options", Flag: "", Type: ParamChoice, Description: "List filters", Required: false, DefaultValue: "Local Only", Choices: []string{"Local Only", "-a", "-r", "--show-current"}},
						{Name: "Delete (Safe)", Flag: "-d", Type: ParamString, Description: "Delete branch name", Required: false},
						{Name: "Delete (Force)", Flag: "-D", Type: ParamString, Description: "Force delete branch name", Required: false},
						{Name: "Rename Current", Flag: "-m", Type: ParamString, Description: "Rename current branch to...", Required: false},
					},
				},
				{
					Name:        "git switch",
					Description: "Switch branches",
					CommandBase: "git",
					SubCommand:  "switch",
					Parameters: []Parameter{
						{Name: "Branch Name", Flag: "", Type: ParamString, Description: "Branch to switch to", Required: true},
						{Name: "Create Branch", Flag: "-c", Type: ParamBool, Description: "Create branch if it does not exist", Required: false},
					},
				},
				{
					Name:        "git checkout",
					Description: "Switch branches or restore working tree files",
					CommandBase: "git",
					SubCommand:  "checkout",
					Parameters: []Parameter{
						{Name: "Branch/File", Flag: "", Type: ParamString, Description: "Branch name or file path", Required: true},
						{Name: "Create Branch", Flag: "-b", Type: ParamBool, Description: "Create and checkout new branch", Required: false},
					},
				},
			},
		},
		{
			Name:  "Merge & Rebase (Git)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git merge",
					Description: "Join two or more development histories together",
					CommandBase: "git",
					SubCommand:  "merge",
					Parameters: []Parameter{
						{Name: "Branch/Commit", Flag: "", Type: ParamString, Description: "Branch to merge into current branch", Required: false},
						{Name: "Squash", Flag: "--squash", Type: ParamBool, Description: "Squash merge", Required: false},
						{Name: "No FF", Flag: "--no-ff", Type: ParamBool, Description: "Create a merge commit even on fast-forward", Required: false},
						{Name: "FF Only", Flag: "--ff-only", Type: ParamBool, Description: "Refuse to merge unless fast-forward possible", Required: false},
						{Name: "Abort", Flag: "--abort", Type: ParamBool, Description: "Abort current merge conflict resolution", Required: false},
						{Name: "Continue", Flag: "--continue", Type: ParamBool, Description: "Continue merge after conflict resolution", Required: false},
					},
				},
				{
					Name:        "git rebase",
					Description: "Reapply commits on top of another base tip",
					CommandBase: "git",
					SubCommand:  "rebase",
					Parameters: []Parameter{
						{Name: "Upstream Branch", Flag: "", Type: ParamString, Description: "Upstream branch (e.g. main)", Required: false},
						{Name: "Interactive", Flag: "-i", Type: ParamBool, Description: "Interactive rebase (requires suspending TUI)", Required: false},
						{Name: "Continue", Flag: "--continue", Type: ParamBool, Description: "Continue rebase after conflicts", Required: false},
						{Name: "Abort", Flag: "--abort", Type: ParamBool, Description: "Abort rebase", Required: false},
						{Name: "Skip", Flag: "--skip", Type: ParamBool, Description: "Skip current patch", Required: false},
					},
					RequiresSuspend: true,
				},
			},
		},
		{
			Name:  "Remote, Fetch & Pull (Git)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git remote",
					Description: "Manage set of tracked repositories",
					CommandBase: "git",
					SubCommand:  "remote",
					Parameters: []Parameter{
						{Name: "Action", Flag: "", Type: ParamChoice, Description: "Remote action", Required: true, DefaultValue: "show", Choices: []string{"show", "add", "remove", "rename", "set-url", "prune", "update"}},
						{Name: "Remote Name", Flag: "", Type: ParamString, Description: "Remote name (e.g. origin)", Required: false, DefaultValue: "origin"},
						{Name: "URL/New Name", Flag: "", Type: ParamString, Description: "URL for add/set-url, new name for rename", Required: false},
					},
				},
				{
					Name:        "git fetch",
					Description: "Download objects and refs from another repository",
					CommandBase: "git",
					SubCommand:  "fetch",
					Parameters: []Parameter{
						{Name: "Remote", Flag: "", Type: ParamString, Description: "Remote to fetch (e.g. origin)", Required: false, DefaultValue: "origin"},
						{Name: "All Remotes", Flag: "--all", Type: ParamBool, Description: "Fetch all remotes", Required: false},
						{Name: "Prune", Flag: "--prune", Type: ParamBool, Description: "Remove remote-tracking references that no longer exist", Required: false},
						{Name: "Fetch Tags", Flag: "--tags", Type: ParamBool, Description: "Fetch all tags from remote", Required: false},
					},
				},
				{
					Name:        "git pull",
					Description: "Fetch from and integrate with another repository or a local branch",
					CommandBase: "git",
					SubCommand:  "pull",
					Parameters: []Parameter{
						{Name: "Remote", Flag: "", Type: ParamString, Description: "Remote repository", Required: false, DefaultValue: "origin"},
						{Name: "Branch", Flag: "", Type: ParamString, Description: "Branch to pull", Required: false, DefaultValue: "main"},
						{Name: "Rebase", Flag: "--rebase", Type: ParamBool, Description: "Rebase current branch on top of upstream after fetching", Required: false},
						{Name: "FF Only", Flag: "--ff-only", Type: ParamBool, Description: "Only fast-forward pull", Required: false},
					},
				},
			},
		},
		{
			Name:  "Push (Git)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git push",
					Description: "Update remote refs along with associated objects",
					CommandBase: "git",
					SubCommand:  "push",
					Parameters: []Parameter{
						{Name: "Remote", Flag: "", Type: ParamString, Description: "Remote repository name", Required: false, DefaultValue: "origin"},
						{Name: "Branch", Flag: "", Type: ParamString, Description: "Branch name to push", Required: false, DefaultValue: "main"},
						{Name: "Set Upstream", Flag: "-u", Type: ParamBool, Description: "Set tracking branch", Required: false},
						{Name: "Force", Flag: "--force", Type: ParamBool, Description: "Force push (overwrite remote history!)", Required: false},
						{Name: "Force With Lease", Flag: "--force-with-lease", Type: ParamBool, Description: "Force push safely (only if no remote updates)", Required: false},
						{Name: "Push Tags", Flag: "--tags", Type: ParamBool, Description: "Push all local tags", Required: false},
						{Name: "Delete Remote Branch", Flag: "--delete", Type: ParamString, Description: "Delete specified branch on remote", Required: false},
					},
				},
			},
		},
		{
			Name:  "History, Log & Diff (Git)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git log",
					Description: "Show commit logs",
					CommandBase: "git",
					SubCommand:  "log",
					Parameters: []Parameter{
						{Name: "Style", Flag: "", Type: ParamChoice, Description: "Visualization style", Required: false, DefaultValue: "--oneline", Choices: []string{"--oneline", "--graph", "--stat", "-p", "Default"}},
						{Name: "Decorate", Flag: "--decorate", Type: ParamBool, Description: "Display branch/tag names", Required: false},
						{Name: "Limit", Flag: "-n", Type: ParamString, Description: "Limit number of commits", Required: false, DefaultValue: "10"},
					},
				},
				{
					Name:        "git shortlog",
					Description: "Summarize git log output",
					CommandBase: "git",
					SubCommand:  "shortlog",
					Parameters: []Parameter{
						{Name: "Summary", Flag: "-s", Type: ParamBool, Description: "Suppress commit description, show counts only", Required: false},
						{Name: "Numbered", Flag: "-n", Type: ParamBool, Description: "Sort output by commit counts", Required: false},
					},
				},
				{
					Name:        "git diff",
					Description: "Show changes between commits, commit and working tree, etc.",
					CommandBase: "git",
					SubCommand:  "diff",
					Parameters: []Parameter{
						{Name: "Target", Flag: "", Type: ParamChoice, Description: "What to diff", Required: false, DefaultValue: "Working Tree", Choices: []string{"Working Tree", "HEAD", "--cached"}},
						{Name: "Compare Branches", Flag: "", Type: ParamString, Description: "Diff branch1 against branch2 (e.g. main..feature)", Required: false},
					},
				},
			},
		},
		{
			Name:  "Reset & Restore (Git)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git reset",
					Description: "Reset current HEAD to the specified state",
					CommandBase: "git",
					SubCommand:  "reset",
					Parameters: []Parameter{
						{Name: "Mode", Flag: "", Type: ParamChoice, Description: "Reset mode", Required: true, DefaultValue: "--mixed", Choices: []string{"--soft", "--mixed", "--hard"}},
						{Name: "Commit / Target", Flag: "", Type: ParamString, Description: "Commit hash or HEAD (defaults to HEAD)", Required: false, DefaultValue: "HEAD"},
					},
					Destructive: true,
				},
				{
					Name:        "git restore",
					Description: "Restore working tree files",
					CommandBase: "git",
					SubCommand:  "restore",
					Parameters: []Parameter{
						{Name: "Staged Only", Flag: "--staged", Type: ParamBool, Description: "Restore index instead of working tree", Required: false},
						{Name: "Target Path", Flag: "", Type: ParamString, Description: "File path or folder to restore (e.g. .)", Required: true, DefaultValue: "."},
					},
				},
			},
		},
		{
			Name:  "Stash & Tag (Git)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git stash",
					Description: "Stash the changes in a dirty working directory away",
					CommandBase: "git",
					SubCommand:  "stash",
					Parameters: []Parameter{
						{Name: "Action", Flag: "", Type: ParamChoice, Description: "Stash action", Required: true, DefaultValue: "push", Choices: []string{"push", "pop", "apply", "list", "show", "drop", "clear"}},
						{Name: "Message (for push)", Flag: "-m", Type: ParamString, Description: "Description message", Required: false},
						{Name: "Stash ID (for pop/apply/drop)", Flag: "", Type: ParamString, Description: "e.g. stash@{0}", Required: false},
					},
				},
				{
					Name:        "git tag",
					Description: "Create, list, delete or verify a tag object signed with GPG",
					CommandBase: "git",
					SubCommand:  "tag",
					Parameters: []Parameter{
						{Name: "Tag Name", Flag: "", Type: ParamString, Description: "Name of the tag", Required: false},
						{Name: "Annotated", Flag: "-a", Type: ParamBool, Description: "Create annotated tag", Required: false},
						{Name: "Message (for -a)", Flag: "-m", Type: ParamString, Description: "Tag description message", Required: false},
						{Name: "Delete Tag", Flag: "-d", Type: ParamString, Description: "Delete tag by name", Required: false},
					},
				},
			},
		},
		{
			Name:  "Advanced Git Operations",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git cherry-pick",
					Description: "Apply the changes introduced by some existing commits",
					CommandBase: "git",
					SubCommand:  "cherry-pick",
					Parameters: []Parameter{
						{Name: "Commit Hash", Flag: "", Type: ParamString, Description: "Commit to cherry-pick", Required: false},
						{Name: "Continue", Flag: "--continue", Type: ParamBool, Description: "Continue cherry-pick after conflict resolution", Required: false},
						{Name: "Abort", Flag: "--abort", Type: ParamBool, Description: "Abort current cherry-pick", Required: false},
						{Name: "Skip", Flag: "--skip", Type: ParamBool, Description: "Skip current commit and continue", Required: false},
					},
				},
				{
					Name:        "git reflog",
					Description: "Manage reflog information",
					CommandBase: "git",
					SubCommand:  "reflog",
				},
				{
					Name:        "git clean",
					Description: "Remove untracked files from the working tree",
					CommandBase: "git",
					SubCommand:  "clean",
					Parameters: []Parameter{
						{Name: "Force", Flag: "-f", Type: ParamBool, Description: "Required by default to clean files", Required: true, DefaultValue: "true"},
						{Name: "Directories", Flag: "-d", Type: ParamBool, Description: "Remove untracked directories as well", Required: false},
						{Name: "Ignored Files", Flag: "-x", Type: ParamBool, Description: "Remove ignored files too", Required: false},
					},
					Destructive: true,
				},
				{
					Name:        "git bisect",
					Description: "Use binary search to find the commit that introduced a bug",
					CommandBase: "git",
					SubCommand:  "bisect",
					Parameters: []Parameter{
						{Name: "Action", Flag: "", Type: ParamChoice, Description: "Bisect step", Required: true, DefaultValue: "start", Choices: []string{"start", "good", "bad", "reset"}},
					},
				},
				{
					Name:        "git revert",
					Description: "Revert some existing commits",
					CommandBase: "git",
					SubCommand:  "revert",
					Parameters: []Parameter{
						{Name: "Commit Hash", Flag: "", Type: ParamString, Description: "Commit to revert", Required: true},
						{Name: "No Edit", Flag: "--no-edit", Type: ParamBool, Description: "Do not edit commit message", Required: false},
					},
				},
				{
					Name:        "git show",
					Description: "Show various types of objects",
					CommandBase: "git",
					SubCommand:  "show",
					Parameters: []Parameter{
						{Name: "Object", Flag: "", Type: ParamString, Description: "Commit hash, tag name, or HEAD", Required: false, DefaultValue: "HEAD"},
					},
				},
				{
					Name:        "git blame",
					Description: "Show what revision and author last modified each line of a file",
					CommandBase: "git",
					SubCommand:  "blame",
					Parameters: []Parameter{
						{Name: "File Path", Flag: "", Type: ParamString, Description: "Path to file", Required: true},
					},
				},
				{
					Name:        "git grep",
					Description: "Print lines matching a pattern",
					CommandBase: "git",
					SubCommand:  "grep",
					Parameters: []Parameter{
						{Name: "Pattern", Flag: "", Type: ParamString, Description: "Regular expression or text pattern to search for", Required: true},
					},
				},
			},
		},
		{
			Name:  "Git Plumbing (Advanced)",
			IsGit: true,
			Commands: []CommandDef{
				{
					Name:        "git cat-file",
					Description: "Provide content or type and size information for repository objects",
					CommandBase: "git",
					SubCommand:  "cat-file",
					Parameters: []Parameter{
						{Name: "Show Type", Flag: "-t", Type: ParamBool, Description: "Show object type", Required: false},
						{Name: "Show Size", Flag: "-s", Type: ParamBool, Description: "Show object size", Required: false},
						{Name: "Pretty Print", Flag: "-p", Type: ParamBool, Description: "Pretty print content", Required: false},
						{Name: "Object ID", Flag: "", Type: ParamString, Description: "SHA1/Object ID of target", Required: true},
					},
				},
				{
					Name:        "git hash-object",
					Description: "Compute object ID and optionally creates a blob from a file",
					CommandBase: "git",
					SubCommand:  "hash-object",
					Parameters: []Parameter{
						{Name: "Write", Flag: "-w", Type: ParamBool, Description: "Write the object into the database", Required: false},
						{Name: "File Path", Flag: "", Type: ParamString, Description: "File to hash", Required: true},
					},
				},
				{
					Name:        "git rev-parse",
					Description: "Pick out and massage parameters",
					CommandBase: "git",
					SubCommand:  "rev-parse",
					Parameters: []Parameter{
						{Name: "Abbrev Ref", Flag: "--abbrev-ref", Type: ParamString, Description: "Get short branch name (e.g. HEAD)", Required: false},
						{Name: "Object", Flag: "", Type: ParamString, Description: "Branch/Tag to parse", Required: false},
					},
				},
				{
					Name:        "git ls-files",
					Description: "Show information about files in the index and the working tree",
					CommandBase: "git",
					SubCommand:  "ls-files",
					Parameters: []Parameter{
						{Name: "Staged", Flag: "--stage", Type: ParamBool, Description: "Show staged contents' mode bits, object name and stage number", Required: false},
						{Name: "Modified", Flag: "-m", Type: ParamBool, Description: "Show modified files", Required: false},
						{Name: "Deleted", Flag: "-d", Type: ParamBool, Description: "Show deleted files", Required: false},
					},
				},
				{
					Name:        "git gc",
					Description: "Cleanup unnecessary files and optimize the local repository",
					CommandBase: "git",
					SubCommand:  "gc",
					Parameters: []Parameter{
						{Name: "Aggressive", Flag: "--aggressive", Type: ParamBool, Description: "Optimize more aggressively (takes longer)", Required: false},
					},
				},
				{
					Name:        "git fsck",
					Description: "Verifies the connectivity and validity of the objects in the database",
					CommandBase: "git",
					SubCommand:  "fsck",
				},
			},
		},
		{
			Name:  "GitHub CLI: Auth & Config",
			IsGit: false,
			Commands: []CommandDef{
				{
					Name:            "gh auth status",
					Description:     "View authentication status",
					CommandBase:     "gh",
					SubCommand:      "auth status",
					RequiresSuspend: true,
				},
				{
					Name:            "gh auth login",
					Description:     "Log in to a GitHub host",
					CommandBase:     "gh",
					SubCommand:      "auth login",
					RequiresSuspend: true,
				},
				{
					Name:            "gh auth logout",
					Description:     "Log out of a GitHub host",
					CommandBase:     "gh",
					SubCommand:      "auth logout",
					RequiresSuspend: true,
				},
				{
					Name:        "gh config list",
					Description: "Print configuration",
					CommandBase: "gh",
					SubCommand:  "config list",
				},
			},
		},
		{
			Name:  "GitHub CLI: Repos & PRs",
			IsGit: false,
			Commands: []CommandDef{
				{
					Name:        "gh repo list",
					Description: "List repositories owned by user or organization",
					CommandBase: "gh",
					SubCommand:  "repo list",
					Parameters: []Parameter{
						{Name: "Owner/Org", Flag: "", Type: ParamString, Description: "Owner username or org name (defaults to self)", Required: false},
						{Name: "Limit", Flag: "--limit", Type: ParamString, Description: "Max number of repositories to list", Required: false, DefaultValue: "30"},
					},
				},
				{
					Name:        "gh repo view",
					Description: "View a repository",
					CommandBase: "gh",
					SubCommand:  "repo view",
					Parameters: []Parameter{
						{Name: "Repo", Flag: "", Type: ParamString, Description: "repository (e.g. owner/repo)", Required: false},
						{Name: "Web", Flag: "--web", Type: ParamBool, Description: "Open repository in the browser", Required: false},
					},
				},
				{
					Name:        "gh pr list",
					Description: "List pull requests in a repository",
					CommandBase: "gh",
					SubCommand:  "pr list",
					Parameters: []Parameter{
						{Name: "State", Flag: "--state", Type: ParamChoice, Description: "Filter by state", Required: false, DefaultValue: "open", Choices: []string{"open", "closed", "merged", "all"}},
						{Name: "Limit", Flag: "--limit", Type: ParamString, Description: "Maximum number of pull requests to fetch", Required: false, DefaultValue: "30"},
					},
				},
				{
					Name:        "gh pr view",
					Description: "View a pull request",
					CommandBase: "gh",
					SubCommand:  "pr view",
					Parameters: []Parameter{
						{Name: "Number/Branch", Flag: "", Type: ParamString, Description: "PR number or branch name", Required: false},
						{Name: "Web", Flag: "--web", Type: ParamBool, Description: "Open PR in web browser", Required: false},
					},
				},
				{
					Name:        "gh pr create",
					Description: "Create a pull request",
					CommandBase: "gh",
					SubCommand:  "pr create",
					Parameters: []Parameter{
						{Name: "Title", Flag: "--title", Type: ParamString, Description: "Title of PR", Required: false},
						{Name: "Body", Flag: "--body", Type: ParamString, Description: "Description body", Required: false},
						{Name: "Draft", Flag: "--draft", Type: ParamBool, Description: "Create PR as draft", Required: false},
					},
					RequiresSuspend: true,
				},
				{
					Name:        "gh pr merge",
					Description: "Merge a pull request",
					CommandBase: "gh",
					SubCommand:  "pr merge",
					Parameters: []Parameter{
						{Name: "Number/Branch", Flag: "", Type: ParamString, Description: "PR number or branch name", Required: false},
						{Name: "Merge Method", Flag: "", Type: ParamChoice, Description: "Method of merging", Required: false, DefaultValue: "--merge", Choices: []string{"--merge", "--rebase", "--squash"}},
					},
					RequiresSuspend: true,
				},
			},
		},
		{
			Name:  "GitHub CLI: Issues & Runs",
			IsGit: false,
			Commands: []CommandDef{
				{
					Name:        "gh issue list",
					Description: "List issues in a repository",
					CommandBase: "gh",
					SubCommand:  "issue list",
					Parameters: []Parameter{
						{Name: "State", Flag: "--state", Type: ParamChoice, Description: "Filter issues by state", Required: false, DefaultValue: "open", Choices: []string{"open", "closed", "all"}},
						{Name: "Assignee", Flag: "--assignee", Type: ParamString, Description: "Filter by assignee", Required: false},
					},
				},
				{
					Name:        "gh issue view",
					Description: "View an issue",
					CommandBase: "gh",
					SubCommand:  "issue view",
					Parameters: []Parameter{
						{Name: "Number", Flag: "", Type: ParamString, Description: "Issue number", Required: true},
						{Name: "Web", Flag: "--web", Type: ParamBool, Description: "Open in browser", Required: false},
					},
				},
				{
					Name:        "gh run list",
					Description: "List recent workflow runs",
					CommandBase: "gh",
					SubCommand:  "run list",
					Parameters: []Parameter{
						{Name: "Limit", Flag: "--limit", Type: ParamString, Description: "Limit counts", Required: false, DefaultValue: "20"},
					},
				},
				{
					Name:        "gh run view",
					Description: "View a workflow run",
					CommandBase: "gh",
					SubCommand:  "run view",
					Parameters: []Parameter{
						{Name: "Run ID", Flag: "", Type: ParamString, Description: "Workflow run ID", Required: false},
					},
				},
			},
		},
	}
}
