package jjlex

import (
	"math/rand"
	"strings"
	"testing"
)

func randomRevset(r *rand.Rand) string {
	operators := []string{"&", "|", "~", "+", "-", "..", "::", ":", ",", "(", ")"}
	functions := []string{"parents", "children", "ancestors", "descendants", "heads", "roots", "branches", "author", "committer", "description", "file", "mine", "tags", "git_refs", "all", "none", "empty", "merges", "conflict", "present", "latest", "connected"}
	words := []string{"main", "trunk", "dev", "feature", "fix", "release", "v1", "v2", "0", "1", "42", "@", "root", "tip"}

	n := r.Intn(10) + 1
	var result strings.Builder
	for range n {
		choice := r.Intn(6)
		switch choice {
		case 0:
			result.WriteString(words[r.Intn(len(words))])
		case 1:
			result.WriteString(operators[r.Intn(len(operators))])
		case 2:
			fn := functions[r.Intn(len(functions))]
			result.WriteString(fn + "(" + randomRevset(r) + ")")
		case 3:
			result.WriteString(" ")
		case 4:
			l := r.Intn(5)
			var s strings.Builder
			for range l {
				s.WriteString(string(rune('a' + r.Intn(26))))
			}
			result.WriteString(`"` + s.String() + `"`)
		case 5:
			l := r.Intn(5)
			var s strings.Builder
			for range l {
				s.WriteString(string(rune('a' + r.Intn(26))))
			}
			result.WriteString("'" + s.String() + "'")
		}
	}
	return result.String()
}

func FuzzSplit(f *testing.F) {
	r := rand.New(rand.NewSource(42))
	for range 100 {
		f.Add(randomRevset(r))
	}

	f.Fuzz(func(t *testing.T, s string) {
		ctx := Split(s)

		if ctx.FullInput != s {
			t.Errorf("FullInput mismatch: got %q, want %q", ctx.FullInput, s)
		}

		if ctx.Type == CompletionTypeUnknown && ctx.IsValid {
			t.Errorf("unknown type should not be valid")
		}

		if len(ctx.Tokens) == 0 {
			t.Errorf("expected at least one token (EOF)")
		}

		lastToken := ctx.Tokens[len(ctx.Tokens)-1]
		if lastToken.Type != TokenEOF {
			t.Errorf("last token should be EOF, got %v", lastToken.Type)
		}
	})
}
