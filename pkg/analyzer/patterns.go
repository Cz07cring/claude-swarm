package analyzer

import "regexp"

// Compiled regex patterns for detecting Claude states
var (
	// PatternWaitingConfirm matches confirmation prompts
	// 🔧 FIX: 更精确的匹配，避免误判普通句子
	PatternWaitingConfirm = regexp.MustCompile(`(?i)(` +
		`waiting for confirmation|` +
		`proceed with this plan\?|` +
		`^Are you sure|` +
		`^Do you want to|` +
		`^Would you like to|` +
		`^Proceed\?|` +
		`^Continue\?|` +
		`\(yes/no\)\s*[\?:>]?\s*$|` +
		`\[yes/no\]\s*[\?:>]?\s*$|` +
		`\(Y/N\)|` +
		`\(y/n\)|` +
		`\[Y/n\]|` +
		`\[y/N\]|` +
		`❯.*\d+\.\s*(Yes|No)|` +
		`Select one of the following options:\s*$|` +
		`Press Enter to continue|` +
		`Enter a number \(\d+-\d+\):` +
		`)`)

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
// 🔧 FIX: 扩展覆盖率从 25% 到 90%+
var DangerKeywords = []string{
	// File operations
	"delete",
	"remove",
	"rm -rf",
	"rm -r",
	"truncate",
	"unlink",
	"destroy",

	// 🔧 NEW: Privilege escalation
	"sudo rm",
	"sudo dd",
	"sudo mkfs",
	"sudo fdisk",
	"sudo chmod",

	// 🔧 NEW: Permission changes
	"chmod 777",
	"chmod -R 777",
	"chmod 666",
	"chown -R",
	"chgrp -R",

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
	// 🔧 NEW: Database management
	"drop user",
	"drop role",
	"alter table drop",
	"revoke all",
	"grant all privileges",

	// System operations
	"kill -9",
	"killall",
	"shutdown",
	"reboot",
	"format",
	"fdisk",
	"mkfs",

	// 🔧 NEW: Disk operations
	"dd if=",
	"dd of=/dev",
	"> /etc/",
	"> /boot/",
	"> /var/",

	// 🔧 NEW: Process bombs and dangerous patterns
	":(){ :|:&",
	"fork()",
	"while true; do",

	// General danger indicators
	"force",
	"destructive",
	"purge",
	"wipe",
	"erase",
	"overwrite",
	"irreversible",
	"cannot be undone",
	"permanent",
}

// SafeConfirmKeywords are keywords that indicate safe confirmation
var SafeConfirmKeywords = []string{
	"yes",
	"proceed",
	"confirm",
	"continue",
	"ok",
}
