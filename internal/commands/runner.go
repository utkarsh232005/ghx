package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/KDM-cli/ghx/internal/db"
)

// BuildArgs constructs the command-line arguments slice from a CommandDef and values map.
func BuildArgs(cmd CommandDef, paramValues map[string]string) []string {
	var args []string

	// Subcommands are split by whitespace
	// e.g. "worktree add" -> ["worktree", "add"]
	if cmd.SubCommand != "" {
		args = append(args, strings.Fields(cmd.SubCommand)...)
	}

	for _, param := range cmd.Parameters {
		val, ok := paramValues[param.Name]
		if !ok {
			val = param.DefaultValue
		}

		switch param.Type {
		case ParamBool:
			if val == "true" {
				if param.Flag != "" {
					args = append(args, param.Flag)
				} else {
					args = append(args, "true")
				}
			}
		case ParamString, ParamChoice:
			if val != "" {
				if param.Flag != "" {
					args = append(args, param.Flag, val)
				} else {
					args = append(args, val)
				}
			}
		}
	}

	return args
}

type RunResult struct {
	Output    string
	Success   bool
	Error     error
	FullCmd   string
}

// RunCommand executes the command synchronously, captures output, and saves it to history.
func RunCommand(ctx context.Context, database *db.DB, cmdDef CommandDef, paramValues map[string]string) RunResult {
	base := cmdDef.CommandBase
	args := BuildArgs(cmdDef, paramValues)
	fullCmdStr := fmt.Sprintf("%s %s", base, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, base, args...)
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outputStr := stdout.String()
	errStr := stderr.String()

	var finalOutput string
	success := err == nil

	if success {
		finalOutput = outputStr
	} else {
		if errStr != "" {
			finalOutput = fmt.Sprintf("Error: %s\n%s", err.Error(), errStr)
		} else {
			finalOutput = fmt.Sprintf("Error: %s\n", err.Error())
		}
	}

	// Add to database command history
	if database != nil {
		_ = database.AddHistory(base, strings.Join(args, " "), finalOutput, success)
	}

	return RunResult{
		Output:  finalOutput,
		Success: success,
		Error:   err,
		FullCmd: fullCmdStr,
	}
}

// SuspendAndRun creates a bubbletea tea.Cmd that suspends the TUI and executes the command directly in the terminal.
func SuspendAndRun(database *db.DB, cmdDef CommandDef, paramValues map[string]string, onComplete func(err error) tea.Msg) tea.Cmd {
	base := cmdDef.CommandBase
	args := BuildArgs(cmdDef, paramValues)

	c := exec.Command(base, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return tea.ExecProcess(c, func(err error) tea.Msg {
		// Log to database on completion
		var finalOutput string
		success := err == nil
		if success {
			finalOutput = "Command completed successfully in interactive session."
		} else {
			finalOutput = fmt.Sprintf("Command failed in interactive session: %v", err)
		}

		if database != nil {
			_ = database.AddHistory(base, strings.Join(args, " "), finalOutput, success)
		}

		return onComplete(err)
	})
}
