package commands

import (
	"strings"
	"testing"
)

func TestBuildArgs(t *testing.T) {
	catalog := GetCatalog()
	if len(catalog) == 0 {
		t.Fatal("Catalog is empty")
	}

	// Test git init with Bare parameter
	var gitInitCmd CommandDef
	found := false
	for _, group := range catalog {
		for _, cmd := range group.Commands {
			if cmd.Name == "git init" {
				gitInitCmd = cmd
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Fatal("git init not found in catalog")
	}

	// Case 1: Bare true
	vals1 := map[string]string{
		"Bare": "true",
	}
	args1 := BuildArgs(gitInitCmd, vals1)
	expected1 := "init --bare"
	if actual1 := strings.Join(args1, " "); actual1 != expected1 {
		t.Errorf("Expected args to be '%s', got '%s'", expected1, actual1)
	}

	// Case 2: Bare false
	vals2 := map[string]string{
		"Bare": "false",
	}
	args2 := BuildArgs(gitInitCmd, vals2)
	expected2 := "init"
	if actual2 := strings.Join(args2, " "); actual2 != expected2 {
		t.Errorf("Expected args to be '%s', got '%s'", expected2, actual2)
	}
}
