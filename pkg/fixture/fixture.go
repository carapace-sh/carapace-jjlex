package fixture

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Fixture struct {
	dir string
}

type Commit struct {
	CommitID    string    `json:"commit_id"`
	ChangeID    string    `json:"change_id"`
	Parents     []string  `json:"parents"`
	Description string    `json:"description"`
	Author      Signature `json:"author"`
	Committer   Signature `json:"committer"`
	Empty       bool      `json:"empty"`
}

type Signature struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Timestamp string `json:"timestamp"`
}

type Bookmark struct {
	Name   string   `json:"name"`
	Target []string `json:"target"`
}

type Tag struct {
	Name   string   `json:"name"`
	Target []string `json:"target"`
}

type State struct {
	Commits   []Commit   `json:"commits"`
	Bookmarks []Bookmark `json:"bookmarks"`
	Tags      []Tag      `json:"tags"`
	WorkingCopy CommitID `json:"working_copy"`
}

type CommitID string

func (s *State) BookmarkNames() []string {
	names := make([]string, len(s.Bookmarks))
	for i, b := range s.Bookmarks {
		names[i] = b.Name
	}
	return names
}

func (s *State) TagNames() []string {
	names := make([]string, len(s.Tags))
	for i, t := range s.Tags {
		names[i] = t.Name
	}
	return names
}

func (s *State) NonEmptyCommits() []Commit {
	var result []Commit
	for _, c := range s.Commits {
		if !c.Empty {
			result = append(result, c)
		}
	}
	return result
}

func (s *State) ChangeIDs() []string {
	ids := make([]string, len(s.Commits))
	for i, c := range s.Commits {
		ids[i] = c.ChangeID
	}
	return ids
}

func (s *State) CommitIDs() []string {
	ids := make([]string, len(s.Commits))
	for i, c := range s.Commits {
		ids[i] = c.CommitID
	}
	return ids
}

func Init(dir string) (*Fixture, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	f := &Fixture{dir: dir}

	if err := f.jj("git", "init"); err != nil {
		return nil, fmt.Errorf("jj git init: %w", err)
	}

	if err := f.jj("config", "set", "--repo", "user.email", "fixture@test.com"); err != nil {
		return nil, fmt.Errorf("set user.email: %w", err)
	}
	if err := f.jj("config", "set", "--repo", "user.name", "fixture"); err != nil {
		return nil, fmt.Errorf("set user.name: %w", err)
	}

	return f, nil
}

func InitTemp() (*Fixture, error) {
	dir, err := os.MkdirTemp("", "jj-fixture-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	return Init(dir)
}

func (f *Fixture) Dir() string {
	return f.dir
}

func (f *Fixture) Cleanup() error {
	return os.RemoveAll(f.dir)
}

func (f *Fixture) Run(args ...string) error {
	return f.jj(args...)
}

func (f *Fixture) RunOutput(args ...string) (string, error) {
	return f.jjOutput(args...)
}

func (f *Fixture) jj(args ...string) error {
	cmd := exec.Command("jj", args...)
	cmd.Dir = f.dir
	cmd.Env = append(os.Environ(), "JJ_USER=fixture", "JJ_EMAIL=fixture@test.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("jj %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (f *Fixture) jjOutput(args ...string) (string, error) {
	cmd := exec.Command("jj", args...)
	cmd.Dir = f.dir
	cmd.Env = append(os.Environ(), "JJ_USER=fixture", "JJ_EMAIL=fixture@test.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("jj %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (f *Fixture) writeFile(path, content string) error {
	fullPath := filepath.Join(f.dir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0o644)
}

func (f *Fixture) removeFile(path string) error {
	fullPath := filepath.Join(f.dir, path)
	return os.Remove(fullPath)
}

func (f *Fixture) CommitAdd(path, content, message string) error {
	if err := f.writeFile(path, content); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	if err := f.jj("commit", "-m", message); err != nil {
		return fmt.Errorf("jj commit: %w", err)
	}
	return nil
}

func (f *Fixture) CommitChange(path, content, message string) error {
	if err := f.writeFile(path, content); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	if err := f.jj("commit", "-m", message); err != nil {
		return fmt.Errorf("jj commit: %w", err)
	}
	return nil
}

func (f *Fixture) CommitRemove(path, message string) error {
	if err := f.removeFile(path); err != nil {
		return fmt.Errorf("remove file: %w", err)
	}
	if err := f.jj("commit", "-m", message); err != nil {
		return fmt.Errorf("jj commit: %w", err)
	}
	return nil
}

func (f *Fixture) CreateBookmark(name string) error {
	return f.jj("bookmark", "create", name)
}

func (f *Fixture) CreateBookmarkAt(name, revset string) error {
	return f.jj("bookmark", "create", name, "-r", revset)
}

func (f *Fixture) MoveBookmark(name string) error {
	return f.jj("bookmark", "move", name)
}

func (f *Fixture) MoveBookmarkTo(name, revset string) error {
	return f.jj("bookmark", "move", name, "--to", revset)
}

func (f *Fixture) DeleteBookmark(name string) error {
	return f.jj("bookmark", "delete", name)
}

func (f *Fixture) RenameBookmark(oldName, newName string) error {
	return f.jj("bookmark", "rename", oldName, newName)
}

func (f *Fixture) ForgetBookmark(name string) error {
	return f.jj("bookmark", "forget", name)
}

func (f *Fixture) CreateTag(name string) error {
	return f.jj("tag", "set", name)
}

func (f *Fixture) CreateTagAt(name, revset string) error {
	return f.jj("tag", "set", name, "-r", revset)
}

func (f *Fixture) DeleteTag(name string) error {
	return f.jj("tag", "delete", name)
}

func (f *Fixture) ConfigSet(key, value string) error {
	return f.jj("config", "set", "--repo", key, value)
}

func (f *Fixture) NewCommit(parents ...string) error {
	args := []string{"new"}
	args = append(args, parents...)
	return f.jj(args...)
}

func (f *Fixture) Describe(message string) error {
	return f.jj("describe", "-m", message)
}

func (f *Fixture) DescribeAt(revset, message string) error {
	return f.jj("describe", "-r", revset, "-m", message)
}

func (f *Fixture) Merge(revset1, revset2, message string) error {
	if err := f.jj("new", revset1, revset2); err != nil {
		return fmt.Errorf("jj new: %w", err)
	}
	if err := f.jj("describe", "-m", message); err != nil {
		return fmt.Errorf("jj describe: %w", err)
	}
	return nil
}

func (f *Fixture) GetState() (*State, error) {
	state := &State{}

	commits, err := f.getLog()
	if err != nil {
		return nil, err
	}
	state.Commits = commits

	bookmarks, err := f.getBookmarks()
	if err != nil {
		return nil, err
	}
	state.Bookmarks = bookmarks

	tags, err := f.getTags()
	if err != nil {
		return nil, err
	}
	state.Tags = tags

	workingCopy, err := f.getWorkingCopy()
	if err != nil {
		return nil, err
	}
	state.WorkingCopy = workingCopy

	return state, nil
}

func (f *Fixture) getLog() ([]Commit, error) {
	out, err := f.jjOutput("log", "--no-pager", "--no-graph", "-T", `json(self) ++ "\t" ++ if(empty, "1", "0") ++ "\n"`, "-r", "all()")
	if err != nil {
		return nil, err
	}

	var commits []Commit
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		var c Commit
		if err := json.Unmarshal([]byte(parts[0]), &c); err != nil {
			return nil, fmt.Errorf("parse commit json: %w: %q", err, line)
		}
		c.Empty = len(parts) > 1 && parts[1] == "1"
		if c.CommitID == "0000000000000000000000000000000000000000" {
			continue
		}
		commits = append(commits, c)
	}
	return commits, nil
}

func (f *Fixture) getBookmarks() ([]Bookmark, error) {
	out, err := f.jjOutput("bookmark", "list", "--no-pager", "-T", "json(self) ++ \"\\n\"")
	if err != nil {
		if strings.Contains(err.Error(), "No bookmarks") {
			return nil, nil
		}
		return nil, err
	}

	var bookmarks []Bookmark
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var b Bookmark
		if err := json.Unmarshal([]byte(line), &b); err != nil {
			return nil, fmt.Errorf("parse bookmark json: %w: %q", err, line)
		}
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, nil
}

func (f *Fixture) getTags() ([]Tag, error) {
	out, err := f.jjOutput("tag", "list", "--no-pager", "-T", "json(self) ++ \"\\n\"")
	if err != nil {
		if strings.Contains(err.Error(), "No tags") {
			return nil, nil
		}
		return nil, err
	}

	var tags []Tag
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var t Tag
		if err := json.Unmarshal([]byte(line), &t); err != nil {
			return nil, fmt.Errorf("parse tag json: %w: %q", err, line)
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (f *Fixture) getWorkingCopy() (CommitID, error) {
	out, err := f.jjOutput("log", "--no-pager", "--no-graph", "-r", "@", "-T", "commit_id")
	if err != nil {
		return "", err
	}
	return CommitID(out), nil
}