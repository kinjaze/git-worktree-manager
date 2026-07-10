package git

import "testing"

func TestParseRemoteBranch(t *testing.T) {
	ref, err := ParseRemoteBranch("origin/release/1.0")
	if err != nil {
		t.Fatalf("ParseRemoteBranch returned error: %v", err)
	}
	if ref.Remote != "origin" || ref.Branch != "release/1.0" || ref.Ref != "origin/release/1.0" {
		t.Fatalf("unexpected ref: %+v", ref)
	}
}

func TestParseRemoteBranchRejectsInvalidInput(t *testing.T) {
	if _, err := ParseRemoteBranch("main"); err == nil {
		t.Fatal("expected error for local-only branch")
	}
}
