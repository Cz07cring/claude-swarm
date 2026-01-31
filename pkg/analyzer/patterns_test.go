package analyzer

import (
	"strings"
	"testing"
)

// TestPatternWaitingConfirm tests the confirmation pattern matching
func TestPatternWaitingConfirm(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldMatch bool
	}{
		// 应该匹配的场景
		{
			name:        "waiting for confirmation",
			input:       "waiting for confirmation",
			shouldMatch: true,
		},
		{
			name:        "proceed with this plan?",
			input:       "Proceed with this plan?",
			shouldMatch: true,
		},
		{
			name:        "Are you sure",
			input:       "Are you sure you want to continue?",
			shouldMatch: true,
		},
		{
			name:        "Do you want to",
			input:       "Do you want to proceed?",
			shouldMatch: true,
		},
		{
			name:        "(yes/no) at end",
			input:       "Continue? (yes/no)",
			shouldMatch: true,
		},
		{
			name:        "[yes/no] format",
			input:       "Proceed? [yes/no]",
			shouldMatch: true,
		},
		{
			name:        "(Y/N) format",
			input:       "Continue? (Y/N)",
			shouldMatch: true,
		},
		{
			name:        "(y/n) format",
			input:       "Proceed? (y/n)",
			shouldMatch: true,
		},
		{
			name:        "[Y/n] format",
			input:       "Continue? [Y/n]",
			shouldMatch: true,
		},
		{
			name:        "[y/N] format",
			input:       "Proceed? [y/N]",
			shouldMatch: true,
		},
		{
			name:        "Press Enter",
			input:       "Press Enter to continue",
			shouldMatch: true,
		},
		{
			name:        "number range",
			input:       "Enter a number (1-5):",
			shouldMatch: true,
		},
		{
			name:        "option list with arrow",
			input:       "❯ 1. Yes\n  2. No",
			shouldMatch: true,
		},

		// 不应该匹配的场景（避免误判）
		{
			name:        "confirm in sentence",
			input:       "This will confirm the functionality works.",
			shouldMatch: false,
		},
		{
			name:        "confirm as verb",
			input:       "We need to confirm your identity.",
			shouldMatch: false,
		},
		{
			name:        "select in sentence",
			input:       "Please select an option from the menu.",
			shouldMatch: false,
		},
		{
			name:        "empty string",
			input:       "",
			shouldMatch: false,
		},
		{
			name:        "normal output",
			input:       "Processing your request...",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := PatternWaitingConfirm.MatchString(tt.input)
			if matched != tt.shouldMatch {
				t.Errorf("PatternWaitingConfirm.MatchString() = %v, want %v\nInput: %q",
					matched, tt.shouldMatch, tt.input)
			}
		})
	}
}

// TestPatternError tests error pattern matching
func TestPatternError(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldMatch bool
	}{
		{
			name:        "error colon",
			input:       "error: file not found",
			shouldMatch: true,
		},
		{
			name:        "failed to",
			input:       "failed to connect to server",
			shouldMatch: true,
		},
		{
			name:        "cannot",
			input:       "cannot read file",
			shouldMatch: true,
		},
		{
			name:        "exception",
			input:       "exception occurred during processing",
			shouldMatch: true,
		},
		{
			name:        "fatal",
			input:       "fatal: repository not found",
			shouldMatch: true,
		},
		{
			name:        "normal message",
			input:       "Processing completed successfully",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := PatternError.MatchString(tt.input)
			if matched != tt.shouldMatch {
				t.Errorf("PatternError.MatchString() = %v, want %v", matched, tt.shouldMatch)
			}
		})
	}
}

// TestPatternIdle tests idle pattern matching
func TestPatternIdle(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldMatch bool
	}{
		{
			name:        "prompt with Try",
			input:       "❯ Try asking me anything",
			shouldMatch: true,
		},
		{
			name:        "prompt with Welcome",
			input:       "❯ Welcome to Claude",
			shouldMatch: true,
		},
		{
			name:        "for shortcuts",
			input:       "Press Ctrl+C for shortcuts",
			shouldMatch: true,
		},
		{
			name:        "normal output",
			input:       "Working on your request...",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := PatternIdle.MatchString(tt.input)
			if matched != tt.shouldMatch {
				t.Errorf("PatternIdle.MatchString() = %v, want %v", matched, tt.shouldMatch)
			}
		})
	}
}

// TestDangerKeywords tests dangerous keyword detection
func TestDangerKeywords(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		shouldBlock bool
		keyword  string
	}{
		// 应该阻止的命令
		{
			name:        "rm -rf",
			command:     "rm -rf /tmp/data",
			shouldBlock: true,
			keyword:     "rm -rf",
		},
		{
			name:        "sudo rm",
			command:     "sudo rm /etc/passwd",
			shouldBlock: true,
			keyword:     "sudo rm",
		},
		{
			name:        "chmod 777",
			command:     "chmod 777 /var/www",
			shouldBlock: true,
			keyword:     "chmod 777",
		},
		{
			name:        "git push --force",
			command:     "git push --force origin main",
			shouldBlock: true,
			keyword:     "git push --force",
		},
		{
			name:        "DROP TABLE",
			command:     "DROP TABLE users;",
			shouldBlock: true,
			keyword:     "drop table",
		},
		{
			name:        "DROP USER",
			command:     "DROP USER admin@localhost;",
			shouldBlock: true,
			keyword:     "drop user",
		},
		{
			name:        "dd if=",
			command:     "dd if=/dev/zero of=/dev/sda",
			shouldBlock: true,
			keyword:     "dd if=",
		},
		{
			name:        "chown -R",
			command:     "chown -R root:root /",
			shouldBlock: true,
			keyword:     "chown -r",
		},
		{
			name:        "DROP COLUMN",
			command:     "ALTER TABLE users DROP COLUMN password;",
			shouldBlock: true,
			keyword:     "drop column",
		},
		{
			name:        "REVOKE",
			command:     "REVOKE ALL PRIVILEGES ON *.* FROM 'user'@'localhost';",
			shouldBlock: true,
			keyword:     "revoke all",
		},

		// 安全的命令
		{
			name:        "safe cat",
			command:     "cat /etc/hosts",
			shouldBlock: false,
			keyword:     "",
		},
		{
			name:        "safe ls",
			command:     "ls -la /tmp",
			shouldBlock: false,
			keyword:     "",
		},
		{
			name:        "safe SELECT",
			command:     "SELECT * FROM users;",
			shouldBlock: false,
			keyword:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandLower := strings.ToLower(tt.command)
			blocked := false

			for _, keyword := range DangerKeywords {
				if strings.Contains(commandLower, keyword) {
					blocked = true
					break
				}
			}

			if blocked != tt.shouldBlock {
				t.Errorf("DangerKeywords check = %v, want %v\nCommand: %q",
					blocked, tt.shouldBlock, tt.command)
			}
		})
	}
}

// TestDangerKeywordsCoverage ensures we have good coverage
func TestDangerKeywordsCoverage(t *testing.T) {
	// 验证关键词列表包含重要的类别
	requiredCategories := map[string][]string{
		"File operations": {"delete", "remove", "rm -rf"},
		"Privilege escalation": {"sudo rm", "sudo dd"},
		"Permission changes": {"chmod 777", "chown -r"},
		"Git operations": {"git push --force", "git reset --hard"},
		"Database operations": {"drop table", "drop user", "drop column"},
		"Disk operations": {"dd if=", "> /etc/"},
	}

	for category, required := range requiredCategories {
		t.Run(category, func(t *testing.T) {
			for _, keyword := range required {
				found := false
				for _, dk := range DangerKeywords {
					if strings.ToLower(dk) == strings.ToLower(keyword) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Missing critical keyword in category %s: %q", category, keyword)
				}
			}
		})
	}
}

// BenchmarkPatternWaitingConfirm benchmarks the confirmation pattern
func BenchmarkPatternWaitingConfirm(b *testing.B) {
	input := "Do you want to proceed? (yes/no)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PatternWaitingConfirm.MatchString(input)
	}
}

// BenchmarkDangerKeywordCheck benchmarks keyword checking
func BenchmarkDangerKeywordCheck(b *testing.B) {
	command := "sudo rm -rf /etc/passwd"
	commandLower := strings.ToLower(command)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, keyword := range DangerKeywords {
			if strings.Contains(commandLower, keyword) {
				break
			}
		}
	}
}
