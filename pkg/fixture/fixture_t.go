package fixture

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type T struct {
	t *testing.T
	f *Fixture
}

type Sandbox interface {
	Dir() string
	Env(string, string) // TODO should be setenv in sandbox
}

func InitT(t *testing.T, s Sandbox) *T {
	t.Helper()
	f, err := Init(s.Dir())
	for _, e := range f.Env() {
		k, v, _ := strings.Cut(e, "=")
		s.Env(k, v)
	}
	if err != nil {
		t.Fatalf("init fixture: %v", err)
	}
	return &T{t: t, f: f}
}

func (t *T) Cleanup() {
	t.t.Helper()
	if err := t.f.Cleanup(); err != nil {
		t.t.Fatalf("cleanup: %v", err)
	}
}

func (t *T) Run(args ...string) {
	t.t.Helper()
	if err := t.f.Run(args...); err != nil {
		t.t.Fatalf("jj %s: %v", strings.Join(args, " "), err)
	}
}

func (t *T) RunOutput(args ...string) string {
	t.t.Helper()
	out, err := t.f.RunOutput(args...)
	if err != nil {
		t.t.Fatalf("jj %s: %v", strings.Join(args, " "), err)
	}
	return out
}

func (t *T) CommitAdd(path, content, message string) {
	t.t.Helper()
	if err := t.f.CommitAdd(path, content, message); err != nil {
		t.t.Fatalf("commit add: %v", err)
	}
}

func (t *T) CommitChange(path, content, message string) {
	t.t.Helper()
	if err := t.f.CommitChange(path, content, message); err != nil {
		t.t.Fatalf("commit change: %v", err)
	}
}

func (t *T) CommitRemove(path, message string) {
	t.t.Helper()
	if err := t.f.CommitRemove(path, message); err != nil {
		t.t.Fatalf("commit remove: %v", err)
	}
}

func (t *T) CreateBookmark(name string) {
	t.t.Helper()
	if err := t.f.CreateBookmark(name); err != nil {
		t.t.Fatalf("create bookmark %q: %v", name, err)
	}
}

func (t *T) CreateBookmarkAt(name, revset string) {
	t.t.Helper()
	if err := t.f.CreateBookmarkAt(name, revset); err != nil {
		t.t.Fatalf("create bookmark %q at %q: %v", name, revset, err)
	}
}

func (t *T) MoveBookmark(name string) {
	t.t.Helper()
	if err := t.f.MoveBookmark(name); err != nil {
		t.t.Fatalf("move bookmark %q: %v", name, err)
	}
}

func (t *T) MoveBookmarkTo(name, revset string) {
	t.t.Helper()
	if err := t.f.MoveBookmarkTo(name, revset); err != nil {
		t.t.Fatalf("move bookmark %q to %q: %v", name, revset, err)
	}
}

func (t *T) DeleteBookmark(name string) {
	t.t.Helper()
	if err := t.f.DeleteBookmark(name); err != nil {
		t.t.Fatalf("delete bookmark %q: %v", name, err)
	}
}

func (t *T) RenameBookmark(oldName, newName string) {
	t.t.Helper()
	if err := t.f.RenameBookmark(oldName, newName); err != nil {
		t.t.Fatalf("rename bookmark %q to %q: %v", oldName, newName, err)
	}
}

func (t *T) ForgetBookmark(name string) {
	t.t.Helper()
	if err := t.f.ForgetBookmark(name); err != nil {
		t.t.Fatalf("forget bookmark %q: %v", name, err)
	}
}

func (t *T) CreateTag(name string) {
	t.t.Helper()
	if err := t.f.CreateTag(name); err != nil {
		t.t.Fatalf("create tag %q: %v", name, err)
	}
}

func (t *T) CreateTagAt(name, revset string) {
	t.t.Helper()
	if err := t.f.CreateTagAt(name, revset); err != nil {
		t.t.Fatalf("create tag %q at %q: %v", name, revset, err)
	}
}

func (t *T) DeleteTag(name string) {
	t.t.Helper()
	if err := t.f.DeleteTag(name); err != nil {
		t.t.Fatalf("delete tag %q: %v", name, err)
	}
}

func (t *T) ConfigSet(key, value string) {
	t.t.Helper()
	if err := t.f.ConfigSet(key, value); err != nil {
		t.t.Fatalf("config set %s=%s: %v", key, value, err)
	}
}

func (t *T) Env() []string {
	return t.f.Env()
}

func (t *T) NewCommit(parents ...string) {
	t.t.Helper()
	if err := t.f.NewCommit(parents...); err != nil {
		t.t.Fatalf("new commit: %v", err)
	}
}

func (t *T) Describe(message string) {
	t.t.Helper()
	if err := t.f.Describe(message); err != nil {
		t.t.Fatalf("describe: %v", err)
	}
}

func (t *T) DescribeAt(revset, message string) {
	t.t.Helper()
	if err := t.f.DescribeAt(revset, message); err != nil {
		t.t.Fatalf("describe at %q: %v", revset, err)
	}
}

func (t *T) Merge(revset1, revset2, message string) {
	t.t.Helper()
	if err := t.f.Merge(revset1, revset2, message); err != nil {
		t.t.Fatalf("merge %q and %q: %v", revset1, revset2, err)
	}
}

func (t *T) GetState() *State {
	t.t.Helper()
	state, err := t.f.GetState()
	if err != nil {
		t.t.Fatalf("get state: %v", err)
	}
	return state
}

func (t *T) AddRemote(name string) error {
	t.t.Helper()
	return t.f.AddRemote(name)
}

func (t *T) RemoveRemote(name string) error {
	t.t.Helper()
	return t.f.RemoveRemote(name)
}

func (t *T) DumpState() error {
	m, err := json.MarshalIndent(t.GetState(), "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(m))
	return nil
}
