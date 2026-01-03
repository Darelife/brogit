package replay

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"brogit/pkg/api"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// Replayer handles the application of diffs to the git repo
type Replayer struct {
	RepoPath string
	dmp      *diffmatchpatch.DiffMatchPatch
}

func NewReplayer(path string) *Replayer {
	return &Replayer{
		RepoPath: path,
		dmp:      diffmatchpatch.New(),
	}
}

// ProcessCommands takes a list of entries and applies them sequentially
func (r *Replayer) ProcessCommands(entries []api.DiffEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Track baselines for each user+file: map[UserID]map[FilePath]Content
	userBaselines := make(map[string]map[string]string)

	// Track the state of files at the start of the batch (for users who haven't sent an update yet)
	originalFileStates := make(map[string]string)

	// 1. Group by consecutive user to maintain Git history granularity
	groups := groupByUser(entries)

	// 2. Process each group
	for _, group := range groups {
		if err := r.processGroup(group, userBaselines, originalFileStates); err != nil {
			return fmt.Errorf("failed to process group for user %s: %w", group[0].UserID, err)
		}
	}

	return nil
}

func groupByUser(entries []api.DiffEntry) [][]api.DiffEntry {
	var groups [][]api.DiffEntry
	if len(entries) == 0 {
		return groups
	}

	currentGroup := []api.DiffEntry{entries[0]}
	currentUser := entries[0].UserID

	for i := 1; i < len(entries); i++ {
		// Start a new group if user changes
		if entries[i].UserID == currentUser {
			currentGroup = append(currentGroup, entries[i])
		} else {
			groups = append(groups, currentGroup)
			currentGroup = []api.DiffEntry{entries[i]}
			currentUser = entries[i].UserID
		}
	}
	groups = append(groups, currentGroup)
	return groups
}

func (r *Replayer) processGroup(group []api.DiffEntry, userBaselines map[string]map[string]string, originalFileStates map[string]string) error {
	user := group[0].UserID
	log.Printf("Processing batch for user: %s (%d changes)", user, len(group))

	// Ensure we are working on the latest main/dev state (for now assuming linear history on main)
	// In a real scenario, we might want to checkout a specific branch, but for "Live Editing"
	// we are patching ON TOP of the current state.

	// Apply Changes with Patching
	for _, entry := range group {
		if entry.Type == api.ChangeTypeFileEdit {
			if err := r.applyChange(user, entry, userBaselines, originalFileStates); err != nil {
				return err
			}
		}
	}

	// Commit
	if err := r.runGit("add", "."); err != nil {
		return err
	}
	// Check if there are changes to commit
	if err := r.runGit("commit", "--allow-empty", "-m", fmt.Sprintf("Brogit update by %s", user)); err != nil {
		return err
	}

	return nil
}

func (r *Replayer) applyChange(user string, entry api.DiffEntry, userBaselines map[string]map[string]string, originalFileStates map[string]string) error {
	fullPath := filepath.Join(r.RepoPath, entry.FilePath)

	// 1. Get Current Master Text (from disk)
	// We read this every time because previous steps in the loop might have changed it
	masterBytes, err := os.ReadFile(fullPath)
	if os.IsNotExist(err) {
		masterBytes = []byte{}
	} else if err != nil {
		return fmt.Errorf("failed to read master file %s: %w", entry.FilePath, err)
	}
	masterText := string(masterBytes)

	// 2. Capture Original State if not present
	// This represents the state of the file BEFORE any edits in this entire batch were applied.
	if _, exists := originalFileStates[entry.FilePath]; !exists {
		originalFileStates[entry.FilePath] = masterText
	}

	// 3. Get User's Baseline
	// If this is the first time we see this user editing this file in this batch,
	// their baseline is assumed to be the INITIAL MasterText (or they are out of sync).
	// For this prototype, we'll initialize it to MasterText if missing.
	// Ideally, we should initialize it to the state BEFORE the batch started, but
	// we can lazily init to current master. If they are truly concurrent with others in this batch,
	// this might be slightly off if we don't track "GlobalBatchStart" state, but it handles sequential well.
	if userBaselines[user] == nil {
		userBaselines[user] = make(map[string]string)
	}
	baseline, ok := userBaselines[user][entry.FilePath]
	if !ok {
		// New user in this batch: assume they started from the Original State
		// This fixes the concurrency issue where later checks overwrote earlier ones because they assumed a newer base.
		baseline = originalFileStates[entry.FilePath]
	}

	// 4. Compute Patch: Diff(Baseline, UserContent)
	patches := r.dmp.PatchMake(baseline, entry.Content)

	// 5. Apply Patch to MasterText
	newMasterText, _ := r.dmp.PatchApply(patches, masterText)
	// Note: PatchApply returns appliedText and a list of booleans for success/fail.
	// We are ignoring failures for now (best effort).

	// 6. Write Result back to disk
	// Ensure dir exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(fullPath, []byte(newMasterText), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", entry.FilePath, err)
	}

	// 7. Update User's Baseline to be what they just sent
	userBaselines[user][entry.FilePath] = entry.Content

	return nil
}

func (r *Replayer) runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.RepoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed: %s: %w", strings.Join(args, " "), string(out), err)
	}
	return nil
}
