package git

import (
	"fmt"
	"strings"
)

type RemoteBranch struct {
	Remote string
	Branch string
	Ref    string
}

func ParseRemoteBranch(ref string) (RemoteBranch, error) {
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return RemoteBranch{}, fmt.Errorf("remote branch must be in remote/branch form: %s", ref)
	}
	return RemoteBranch{Remote: parts[0], Branch: parts[1], Ref: ref}, nil
}
