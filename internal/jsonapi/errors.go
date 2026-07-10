package jsonapi

const (
	StatusCreated    = "created"
	StatusListed     = "listed"
	StatusUpdated    = "updated"
	StatusMergedBack = "merged_back"
	StatusRemoved    = "removed"
	StatusConflict   = "conflict"
	StatusFailed     = "failed"
)

const (
	ErrInvalidArgument     = "INVALID_ARGUMENT"
	ErrNotGitRepository    = "NOT_A_GIT_REPOSITORY"
	ErrSourceRefNotFound   = "SOURCE_REF_NOT_FOUND"
	ErrBranchAlreadyExists = "BRANCH_ALREADY_EXISTS"
	ErrWorktreeNotFound    = "WORKTREE_NOT_FOUND"
	ErrMetadataNotFound    = "METADATA_NOT_FOUND"
	ErrMetadataCorrupt     = "METADATA_CORRUPT"
	ErrMetadataLocked      = "METADATA_LOCKED"
	ErrTargetDirty         = "TARGET_WORKTREE_DIRTY"
	ErrWorktreeDirty       = "WORKTREE_DIRTY"
	ErrMergeConflict       = "GIT_MERGE_CONFLICT"
	ErrGitCommandFailed    = "GIT_COMMAND_FAILED"
)
