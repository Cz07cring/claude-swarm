# Archived V1 Documentation

This directory contains documentation for the old V1 architecture (tmux-based).

**These documents are outdated and kept for historical reference only.**

## V1 Architecture (Deprecated)

The V1 architecture used tmux sessions to manage multiple Claude Code instances. This approach had fundamental limitations:

- ❌ **Unreliable**: `tmux send-keys` couldn't control when Claude accepted input
- ❌ **Slow**: Tasks would often get stuck indefinitely
- ❌ **Complex**: Difficult to debug and maintain

## V2 Architecture (Current)

The current V2 architecture uses direct Claude CLI execution:

- ✅ **Reliable**: Direct command execution with full control
- ✅ **Fast**: 10-12 seconds per task
- ✅ **Free**: Uses local Claude CLI (no API costs)
- ✅ **Simple**: Clean architecture, easy to debug

## See Current Documentation

For current V2 documentation, see:
- `/docs/V2_INTEGRATION_COMPLETE.md` - V2 architecture overview
- `/docs/USAGE_GUIDE.md` - Usage guide (updated for V2)
- `/README.md` - Main project documentation

---

**Last Updated**: 2026-02-01
**Status**: Archived (V1 deprecated)
