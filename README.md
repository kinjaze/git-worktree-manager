# Git Worktree Manager

`gwt` is a Git worktree lifecycle manager with a human-friendly TUI and agent-friendly CLI/JSON output.

## Install

### From source

```bash
go install github.com/kinjaze/git-worktree-manager/cmd/gwt@latest
```

### From a release archive

Download the archive for your platform, extract it, then move `gwt` into a directory on your `PATH`:

```bash
tar -xzf gwt_<version>_<os>_<arch>.tar.gz
chmod +x gwt
sudo mv gwt /usr/local/bin/
gwt --help
```

## Core commands

```bash
gwt create login-page --repo /path/to/repo --source origin/main --branch feat/login-page --path /path/to/worktree
gwt list
gwt update <id-or-path>
gwt merge-back <id-or-path>
gwt remove <id-or-path>
gwt tui
```

## Merge behavior

- `gwt update` uses ordinary merge from the recorded source branch.
- `gwt merge-back` uses `git merge --no-ff <worktreeBranch>`.
- Conflicts are left in Git's native conflict state and reported with next steps.

## Language

English is the default. Chinese can be enabled with:

```bash
gwt --lang zh <command>
gwt config set language zh
```

JSON field names, statuses, and error codes remain English.
