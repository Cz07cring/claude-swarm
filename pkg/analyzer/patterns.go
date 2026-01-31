package analyzer

import "regexp"

// Compiled regex patterns for detecting Claude states
var (
	// PatternWaitingConfirm matches confirmation prompts
	PatternWaitingConfirm = regexp.MustCompile(`(?i)(waiting for confirmation|proceed with this plan\?|Do you want to proceed|confirm|^\s*yes/no|\(yes/no\)|\[yes/no\]|[❯►>]\s*\d+\.\s*(Yes|No)|Select an option)`)

	// PatternError matches error messages
	PatternError = regexp.MustCompile(`(?i)(error:|failed to|cannot|exception|fatal:)`)

	// PatternToolCall matches tool calls (indicates working)
	PatternToolCall = regexp.MustCompile(`<function_calls>|<invoke>`)

	// PatternCompleted matches completion indicators
	PatternCompleted = regexp.MustCompile(`(?i)(completed|finished|done|success)`)

	// PatternIdle matches Claude's idle prompt
	PatternIdle = regexp.MustCompile(`(?m)^[❯►>]\s+(Try|Welcome|$)|for shortcuts\s*$`)
)

// DangerKeywords are keywords that indicate potentially dangerous operations
var DangerKeywords = []string{
	// File operations
	"delete",
	"remove",
	"rm -rf",
	"rm -r",
	"truncate",
	"unlink",
	"destroy",

	// Git dangerous operations
	"git reset --hard",
	"git push --force",
	"git push -f",
	"git clean -f",
	"git clean -fd",
	"git branch -D",
	"git rebase --hard",
	"--force-with-lease",

	// Database operations
	"drop table",
	"drop database",
	"truncate table",
	"delete from",

	// System operations
	"kill -9",
	"killall",
	"shutdown",
	"reboot",
	"format",
	"fdisk",
	"mkfs",

	// General danger indicators
	"force",
	"destructive",
	"purge",
	"wipe",
	"erase",
	"overwrite",
}

// SafeConfirmKeywords are keywords that indicate safe confirmation
var SafeConfirmKeywords = []string{
	"yes",
	"proceed",
	"confirm",
	"continue",
	"ok",
}
