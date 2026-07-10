package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kinjaze/git-worktree-manager/internal/core"
	"github.com/kinjaze/git-worktree-manager/internal/i18n"
	"github.com/kinjaze/git-worktree-manager/internal/jsonapi"
)

func printJSON(response jsonapi.Response) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

func handleError(err error, tr i18n.Translator, jsonOutput bool) error {
	if err == nil {
		return nil
	}
	status := jsonapi.StatusFailed
	code := jsonapi.ErrGitCommandFailed
	message := err.Error()
	localized := tr.T("error.gitCommandFailed")
	data := any(nil)
	if coreErr, ok := err.(core.Error); ok {
		code = coreErr.Code
		message = coreErr.Message
		data = coreErr.Data
		localized = localizedError(code, tr)
		if code == jsonapi.ErrMergeConflict {
			status = jsonapi.StatusConflict
		}
	}
	if jsonOutput {
		_ = printJSON(jsonapi.Failure(status, code, message, localized, data))
	} else {
		fmt.Fprintf(os.Stderr, "%s: %s\n", localized, message)
		if data != nil {
			fmt.Fprintf(os.Stderr, "%v\n", data)
		}
	}
	return err
}

func localizedError(code string, tr i18n.Translator) string {
	switch code {
	case jsonapi.ErrInvalidArgument:
		return tr.T("error.invalidArgument")
	case jsonapi.ErrMergeConflict:
		return tr.T("error.mergeConflict")
	case jsonapi.ErrWorktreeDirty:
		return tr.T("error.worktreeDirty")
	case jsonapi.ErrTargetDirty:
		return tr.T("error.targetWorktreeDirty")
	case jsonapi.ErrWorktreeNotFound:
		return tr.T("error.worktreeNotFound")
	case jsonapi.ErrNotGitRepository:
		return tr.T("error.notGitRepository")
	case jsonapi.ErrSourceRefNotFound:
		return tr.T("error.sourceRefNotFound")
	case jsonapi.ErrMetadataCorrupt:
		return tr.T("error.metadataCorrupt")
	case jsonapi.ErrMetadataLocked:
		return tr.T("error.metadataLocked")
	default:
		return tr.T("error.gitCommandFailed")
	}
}
