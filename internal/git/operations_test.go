package git

import "testing"

func TestParseNWOFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com/owner/repo.git", "owner/repo"},
		{"https://github.com/owner/repo", "owner/repo"},
		{"git@github.com:owner/repo.git", "owner/repo"},
		{"git@github.com:owner/repo", "owner/repo"},
		{"https://github.com/owner/repo-name.git", "owner/repo-name"},
		{"", ""},
	}

	for _, test := range tests {
		actual := ParseNWOFromURL(test.url)
		if actual != test.expected {
			t.Errorf("ParseNWOFromURL(%q) = %q; expected %q", test.url, actual, test.expected)
		}
	}
}

func TestExpandRemoteURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"owner/repo", "https://github.com/owner/repo.git"},
		{"https://github.com/owner/repo.git", "https://github.com/owner/repo.git"},
		{"git@github.com:owner/repo.git", "git@github.com:owner/repo.git"},
		{"invalidurl", "invalidurl"},
	}

	for _, test := range tests {
		actual := ExpandRemoteURL(test.url)
		if actual != test.expected {
			t.Errorf("ExpandRemoteURL(%q) = %q; expected %q", test.url, actual, test.expected)
		}
	}
}
