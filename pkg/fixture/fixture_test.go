package fixture

import (
	"encoding/json"
	"strings"
	"testing"
)

// nonEmptyCommits returns only non-empty commits from the state
func nonEmptyCommits(s *State) []Commit {
	var result []Commit
	for _, c := range s.Commits {
		if !c.Empty {
			result = append(result, c)
		}
	}
	return result
}

func findCommitByDesc(commits []Commit, desc string) *Commit {
	for i := range commits {
		if strings.Contains(commits[i].Description, desc) {
			return &commits[i]
		}
	}
	return nil
}

func TestInitTemp(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if f.Dir() == "" {
		t.Fatal("expected non-empty dir")
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	ne := nonEmptyCommits(state)
	if len(ne) != 0 {
		t.Fatalf("expected 0 non-empty commits, got %d", len(ne))
	}
}

func TestCommitAdd(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if err := f.CommitAdd("hello.txt", "hello world", "add hello.txt"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	ne := nonEmptyCommits(state)
	if len(ne) != 1 {
		t.Fatalf("expected 1 non-empty commit, got %d", len(ne))
	}
	if ne[0].Description != "add hello.txt\n" {
		t.Fatalf("expected description 'add hello.txt\\n', got %q", ne[0].Description)
	}
	if ne[0].CommitID == "" {
		t.Fatal("expected non-empty commit ID")
	}
	if ne[0].ChangeID == "" {
		t.Fatal("expected non-empty change ID")
	}
}

func TestCommitChange(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if err := f.CommitAdd("file.txt", "initial", "add file.txt"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}
	if err := f.CommitChange("file.txt", "updated", "update file.txt"); err != nil {
		t.Fatalf("CommitChange: %v", err)
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	ne := nonEmptyCommits(state)
	if len(ne) != 2 {
		t.Fatalf("expected 2 non-empty commits, got %d", len(ne))
	}
}

func TestCommitRemove(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if err := f.CommitAdd("file.txt", "content", "add file.txt"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}
	if err := f.CommitRemove("file.txt", "remove file.txt"); err != nil {
		t.Fatalf("CommitRemove: %v", err)
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	ne := nonEmptyCommits(state)
	if len(ne) != 2 {
		t.Fatalf("expected 2 non-empty commits, got %d", len(ne))
	}
}

func TestBookmarks(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if err := f.CommitAdd("file.txt", "content", "initial commit"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}

	if err := f.CreateBookmark("feature"); err != nil {
		t.Fatalf("CreateBookmark: %v", err)
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if len(state.Bookmarks) != 1 {
		t.Fatalf("expected 1 bookmark, got %d", len(state.Bookmarks))
	}
	if state.Bookmarks[0].Name != "feature" {
		t.Fatalf("expected bookmark 'feature', got %q", state.Bookmarks[0].Name)
	}

	// Add a new commit and move the bookmark
	if err := f.CommitAdd("file2.txt", "more", "second commit"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}
	if err := f.MoveBookmark("feature"); err != nil {
		t.Fatalf("MoveBookmark: %v", err)
	}

	state, err = f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if len(state.Bookmarks) != 1 {
		t.Fatalf("expected 1 bookmark, got %d", len(state.Bookmarks))
	}

	// Delete bookmark
	if err := f.DeleteBookmark("feature"); err != nil {
		t.Fatalf("DeleteBookmark: %v", err)
	}
	state, err = f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if len(state.Bookmarks) != 0 {
		t.Fatalf("expected 0 bookmarks after delete, got %d", len(state.Bookmarks))
	}
}

func TestBookmarkAt(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if err := f.CommitAdd("file.txt", "content", "first"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}
	// Bookmark at previous revision using revset
	if err := f.CreateBookmarkAt("release", "@-"); err != nil {
		t.Fatalf("CreateBookmarkAt: %v", err)
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	found := false
	for _, b := range state.Bookmarks {
		if b.Name == "release" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'release' bookmark")
	}
}

func TestTags(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if err := f.CommitAdd("file.txt", "content", "tagged commit"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}

	if err := f.CreateTag("v1.0"); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if len(state.Tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(state.Tags))
	}
	if state.Tags[0].Name != "v1.0" {
		t.Fatalf("expected tag 'v1.0', got %q", state.Tags[0].Name)
	}
	if len(state.Tags[0].Target) != 1 {
		t.Fatalf("expected 1 target, got %d", len(state.Tags[0].Target))
	}

	// Delete tag
	if err := f.DeleteTag("v1.0"); err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}

	state, err = f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if len(state.Tags) != 0 {
		t.Fatalf("expected 0 tags after delete, got %d", len(state.Tags))
	}
}

func TestTagAt(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if err := f.CommitAdd("file.txt", "content", "first"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}
	if err := f.CreateTagAt("v0.1", "@-"); err != nil {
		t.Fatalf("CreateTagAt: %v", err)
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if len(state.Tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(state.Tags))
	}
}

func TestMerge(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	// Create a base commit
	if err := f.CommitAdd("base.txt", "base", "add base"); err != nil {
		t.Fatalf("CommitAdd base: %v", err)
	}

	// Create branch A
	if err := f.NewCommit("@-"); err != nil {
		t.Fatalf("NewCommit: %v", err)
	}
	if err := f.CommitAdd("a.txt", "feature A", "feature A"); err != nil {
		t.Fatalf("CommitAdd A: %v", err)
	}
	stateA, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState A: %v", err)
	}
	changeA := stateA.WorkingCopy

	// Create branch B
	if err := f.NewCommit("@---"); err != nil {
		t.Fatalf("NewCommit B: %v", err)
	}
	if err := f.CommitAdd("b.txt", "feature B", "feature B"); err != nil {
		t.Fatalf("CommitAdd B: %v", err)
	}
	stateB, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState B: %v", err)
	}
	changeB := stateB.WorkingCopy

	// Merge using commit IDs (revset)
	if err := f.Merge(string(changeA), string(changeB), "merge A and B"); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState after merge: %v", err)
	}

	// Find merge commit (should have 2 parents)
	var mergeCommit *Commit
	for i := range state.Commits {
		if len(state.Commits[i].Parents) == 2 {
			mergeCommit = &state.Commits[i]
			break
		}
	}
	if mergeCommit == nil {
		t.Fatal("expected merge commit with 2 parents")
	}
	if !strings.Contains(mergeCommit.Description, "merge A and B") {
		t.Fatalf("merge commit description should contain 'merge A and B', got %q", mergeCommit.Description)
	}
}

func TestWorkingCopy(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if err := f.CommitAdd("file.txt", "content", "first"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if state.WorkingCopy == "" {
		t.Fatal("expected non-empty working copy commit ID")
	}
}

func TestStateJSON(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if err := f.CommitAdd("file.txt", "content", "first commit"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}
	if err := f.CreateBookmark("main"); err != nil {
		t.Fatalf("CreateBookmark: %v", err)
	}
	if err := f.CreateTag("v1.0"); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}

	t.Logf("State JSON:\n%s", string(data))

	var state2 State
	if err := json.Unmarshal(data, &state2); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(state2.Commits) != len(state.Commits) {
		t.Fatalf("expected %d commits, got %d", len(state.Commits), len(state2.Commits))
	}
	if len(state2.Bookmarks) != len(state.Bookmarks) {
		t.Fatalf("expected %d bookmarks, got %d", len(state.Bookmarks), len(state2.Bookmarks))
	}
	if len(state2.Tags) != len(state.Tags) {
		t.Fatalf("expected %d tags, got %d", len(state.Tags), len(state2.Tags))
	}
}

func TestCommitParents(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if err := f.CommitAdd("a.txt", "a", "commit A"); err != nil {
		t.Fatalf("CommitAdd A: %v", err)
	}
	state1, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState 1: %v", err)
	}
	// The non-empty commit we just made
	ne1 := nonEmptyCommits(state1)
	if len(ne1) != 1 {
		t.Fatalf("expected 1 non-empty commit, got %d", len(ne1))
	}
	commitAID := ne1[0].CommitID

	if err := f.CommitAdd("b.txt", "b", "commit B"); err != nil {
		t.Fatalf("CommitAdd B: %v", err)
	}
	state2, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState 2: %v", err)
	}

	ne2 := nonEmptyCommits(state2)
	c := findCommitByDesc(ne2, "commit B")
	if c == nil {
		t.Fatal("could not find commit B")
	}
	if len(c.Parents) != 1 {
		t.Fatalf("expected 1 parent, got %d", len(c.Parents))
	}
	if c.Parents[0] != commitAID {
		t.Fatalf("expected parent %s, got %s", commitAID, c.Parents[0])
	}
}

func TestInitInDir(t *testing.T) {
	dir := t.TempDir()
	f, err := Init(dir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer f.Cleanup()

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	ne := nonEmptyCommits(state)
	if len(ne) != 0 {
		t.Fatalf("expected 0 non-empty commits, got %d", len(ne))
	}
}

func TestDescribeAt(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	if err := f.CommitAdd("file.txt", "content", "initial"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}
	// Describe the non-empty parent commit
	if err := f.DescribeAt("@-", "updated description"); err != nil {
		t.Fatalf("DescribeAt: %v", err)
	}

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}

	ne := nonEmptyCommits(state)
	found := false
	for _, c := range ne {
		if strings.Contains(c.Description, "updated description") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected to find commit with 'updated description'")
	}
}

func TestEmptyField(t *testing.T) {
	f, err := InitTemp()
	if err != nil {
		t.Fatalf("InitTemp: %v", err)
	}
	defer f.Cleanup()

	state, err := f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}

	// Fresh repo: working copy commit should be empty
	if len(state.Commits) == 0 {
		t.Fatal("expected at least the working copy commit")
	}
	// The working copy should be marked empty
	var wc *Commit
	for i := range state.Commits {
		if state.Commits[i].CommitID == string(state.WorkingCopy) {
			wc = &state.Commits[i]
			break
		}
	}
	if wc == nil {
		t.Fatal("working copy commit not found in state")
	}
	if !wc.Empty {
		t.Fatal("working copy commit should be marked empty")
	}

	// After adding a file and committing, the new commit should not be empty
	if err := f.CommitAdd("file.txt", "content", "first commit"); err != nil {
		t.Fatalf("CommitAdd: %v", err)
	}
	state, err = f.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	ne := nonEmptyCommits(state)
	if len(ne) != 1 {
		t.Fatalf("expected 1 non-empty commit, got %d", len(ne))
	}
	if ne[0].Empty {
		t.Fatal("committed change should not be marked empty")
	}
}