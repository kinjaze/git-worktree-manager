package metadata

const SchemaVersion = 1

type File struct {
	SchemaVersion int      `json:"schemaVersion"`
	Worktrees     []Record `json:"worktrees"`
}

type Record struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	SourceRepoPath     string `json:"sourceRepoPath"`
	SourceRemoteBranch string `json:"sourceRemoteBranch"`
	SourceRemote       string `json:"sourceRemote"`
	SourceBranchName   string `json:"sourceBranchName"`
	TargetLocalBranch  string `json:"targetLocalBranch"`
	WorktreeBranch     string `json:"worktreeBranch"`
	Path               string `json:"path"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	LastKnownHead      string `json:"lastKnownHead"`
	Status             string `json:"status"`
}

func EmptyFile() File {
	return File{SchemaVersion: SchemaVersion, Worktrees: []Record{}}
}
