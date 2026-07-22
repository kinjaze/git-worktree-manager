# Changelog

## v0.1.2 - 2026-07-22

### Fixed

- Repair and refresh managed records when a worktree directory is moved outside `gwt`.
- Recover update/remove operations when both the source repository and worktree are moved under the same parent directory.

## v0.1.1 - 2026-07-20

### Fixed

- Allow removing manager records for stale Git worktrees whose directories were deleted outside `gwt`.
- Add a TUI force-delete confirmation for dirty worktrees so users can explicitly discard uncommitted changes.

## v0.1.0 - 2026-07-10

Initial release of `gwt`, a Git worktree lifecycle manager with both TUI and CLI workflows.

### Added

- Create, list, update, merge-back, and remove Git worktrees.
- Interactive TUI for managing recorded worktrees.
- Agent-friendly JSON output via CLI flags.
- English and Chinese UI language support.
- Merge conflict reporting with actionable next steps.
- Safety checks for dirty target worktrees before merge-back.

### Fixed

- Clear stale TUI operation messages after failed merge-back checks and subsequent refreshes or operations.
