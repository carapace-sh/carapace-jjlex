// Command grammardiff compares this library's grammar acceptance with the
// real jj binary: for every corpus expression it parses with the matching
// package (revset, fileset, or template) and runs jj, classifying jj's
// outcome as grammatical accept/reject vs. semantic rejection (unknown
// keywords, unexpected keyword arguments, type errors — these do not count
// as grammar rejections).
//
// Usage:
//
//	go run ./cmd/grammardiff [-jj <path-to-jj>] [-corpus <path.json>]
//
// Output is JSON on stdout; a human summary goes to stderr. Exit codes:
// 0 — no mismatches, 1 — mismatches found, 2 — harness error (cannot build
// a sandbox repo or jj is missing).
//
// The corpus is JSON:
//
//	{
//	  "template": ["expr1", ...],
//	  "revset":   [...],
//	  "fileset":  [...]
//	}
//
// When -corpus is not given, the bundled default corpus is used.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/carapace-sh/carapace-jjlex/pkg/fileset"
	"github.com/carapace-sh/carapace-jjlex/pkg/revset"
	"github.com/carapace-sh/carapace-jjlex/pkg/template"
)

type corpus map[string][]string

var defaultCorpus = corpus{
	"template": []string{
		`json({"commit_id" => change_id, "tags" => tags})`,
		"json({})",
		"{}",
		"{ a=>b }",
		`{"a" => 1,}`,
		"{a, b => c}",
		`{"a" => }`,
		`{"a"`,
		"{,}",
		"{a => |c| c}",
		"{a => f(x=1)}",
		`{"x" => !x}`,
		"{a => -1}",
		`json({a => none, "b" => glob:"*.rs"})`,
		"{x:y => none}",
		`{"a" => a ++ b, "c" => d ++ e}`,
		`{"a" => |c| c ++ "x"}`,
		`json({"a" => "b"}) ++ "c"`,
		`"c" ++ json({"a" => "b"})`,
		`json({"a" => |c| c.description() })`,
		`fill(80, "content", break_words=true)`,
		`pad_start(10, change_id, fill_char=".")`,
		`truncate_end(10, description, ellipsis="...")`,
		`git_web_url(remote="upstream")`,
		`hyperlink("https://x", "t", "f")`,
		`if(true, "a", "b")`,
		`config("ui.graph.style")`,
		`revset("mine()")`,
		`self.contained_in("visible_heads()")`,
		`revset("none()").map(|c| c.change_id())`,
		`self.files().map(|e| e.added_values().len())`,
		`self.working_copies().map(|w| w.current_workspace())`,
		`self.bookmarks().map(|b| b.added_targets().first().get(0))`,
		`self.contained_in(revset="mine()")`,
	},
	"revset": []string{
		"all()",
		"none()",
		"mine()",
		"root()",
		"x | y",
		"x | y & z",
		"~x",
		"::x",
		"x::y",
		"x..",
		"..x",
		"x-",
		"x+",
		"x - y",
		"x + y",
		"author:exact:\"a\"",
		"author(a)",
		"author(exact:\"a\")",
		"heads(x | y)",
		"present(x)",
		"latest(mine(), 3)",
		"coalesce(mine(), none())",
		"bookmarks(exact:\"main\")",
		"ancestors(x+, 2)",
		"descendants(x, 3)",
	},
	"fileset": []string{
		"all()",
		"none()",
		"glob:*.rs",
		"a & b",
		"a | b",
		"~all()",
		"a | !all()",
		"exact:a.txt",
		"a.html",
		"src",
		`"a b.txt"`,
		"{a, b}",
		"a&",
		"root:src",
		"cwd:. | root:.",
	},
}

type result struct {
	Language string `json:"language"`
	Expr     string `json:"expr"`
	Dev      string `json:"dev"`
	JJ       string `json:"jj"`
	JJError  string `json:"jjError,omitempty"`
	Verdict  string `json:"verdict"`
}

type summary struct {
	Results       []result `json:"results"`
	MismatchCount int      `json:"mismatchCount"`
}

func main() {
	jjBin := flag.String("jj", "jj", "path to the jj binary")
	corpusPath := flag.String("corpus", "", "path to a JSON corpus file (default: bundled)")
	flag.Parse()

	s, err := run(*jjBin, *corpusPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "checked %d expressions, %d mismatch(es)\n", len(s.Results), s.MismatchCount)
	if s.MismatchCount > 0 {
		os.Exit(1)
	}
}

func run(jjBin, corpusPath string) (*summary, error) {
	lines := defaultCorpus
	if corpusPath != "" {
		data, err := os.ReadFile(corpusPath)
		if err != nil {
			return nil, fmt.Errorf("reading corpus: %w", err)
		}
		if err := json.Unmarshal(data, &lines); err != nil {
			return nil, fmt.Errorf("parsing corpus: %w", err)
		}
	}

	repoDir, cleanup, err := newSandbox(jjBin)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	s := &summary{}
	for _, lang := range []string{"revset", "fileset", "template"} {
		exprs, ok := lines[lang]
		if !ok {
			continue
		}
		parser := parserFor(lang)
		if parser == nil {
			return nil, fmt.Errorf("unknown language %q", lang)
		}
		for _, expr := range exprs {
			r := check(jjBin, repoDir, lang, expr, parser)
			if r.Verdict == "mismatch" {
				s.MismatchCount++
			}
			s.Results = append(s.Results, r)
			fmt.Fprintf(os.Stderr, "%-8s %-7s jj=%-10s %s\n", lang, r.Dev, r.JJ, expr)
		}
	}
	return s, nil
}

func parserFor(lang string) func(string) error {
	switch lang {
	case "revset":
		return func(e string) error { _, err := revset.Parse(e); return err }
	case "fileset":
		return func(e string) error { _, err := fileset.Parse(e); return err }
	case "template":
		return func(e string) error { _, err := template.Parse(e); return err }
	}
	return nil
}

func check(jjBin, repoDir, lang, expr string, parser func(string) error) result {
	r := result{Language: lang, Expr: expr}
	r.Dev = verdictFor(parser(expr))
	stderr, runErr := runJJ(jjBin, repoDir, lang, expr)
	r.JJ, r.JJError = classifyJJ(lang, stderr, runErr)
	r.Verdict = verdict(r.Dev, r.JJ)
	return r
}

func verdictFor(err error) string {
	if err == nil {
		return "accept"
	}
	return "reject"
}

func verdict(dev, jj string) string {
	// "semantic" means jj parsed the grammar fine and only semantic checks
	// (unknown keywords, unexpected keyword arguments, type errors) failed.
	switch jj {
	case "accept", "semantic":
		if dev == "accept" {
			return "match"
		}
		return "mismatch"
	default: // "syntax"
		if dev == "reject" {
			return "match"
		}
		return "mismatch"
	}
}

func runJJ(jjBin, dir, lang, expr string) (string, error) {
	var args []string
	switch lang {
	case "template":
		args = []string{"log", "--no-graph", "-r", "none()", "-T", expr}
	case "revset":
		args = []string{"log", "--no-graph", "-r", expr, "-T", `""`}
	case "fileset":
		args = []string{"file", "list", expr}
	default:
		return "", fmt.Errorf("unknown language %q", lang)
	}
	cmd := exec.Command(jjBin, args...)
	cmd.Dir = dir
	cmd.Env = sandboxEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	return stderr.String(), runErr
}

// classifyJJ maps a jj outcome to a grammar-relevant verdict:
//
//	accept   — jj ran successfully
//	syntax   — jj rejected the expression grammatically
//	semantic — jj parsed the grammar but semantic checks failed at build
//	           time (unknown keywords, unexpected kwargs, type errors)
//	harness  — jj failed outside the parse phase (missing binary, repo
//	           unavailable, etc.); not usable as a grammar verdict
func classifyJJ(lang, stderr string, runErr error) (string, string) {
	if runErr == nil {
		return "accept", ""
	}
	if isSemanticFailure(stderr) {
		return "semantic", stderr
	}
	if isGrammarParsePhase(lang, stderr) {
		return "syntax", stderr
	}
	// Eval-phase rejections (e.g. unknown revision symbols) are semantic:
	// the revset grammar accepted the expression structure.
	if strings.Contains(stderr, "doesn't exist") {
		return "semantic", stderr
	}
	return "harness", stderr
}

func isGrammarParsePhase(lang, stderr string) bool {
	switch lang {
	case "template":
		return strings.Contains(stderr, "Failed to parse template")
	case "revset":
		return strings.Contains(stderr, "Failed to parse revset")
	case "fileset":
		return strings.Contains(stderr, "Failed to parse fileset")
	}
	return false
}

// semanticFailureMarkers lists jj build-time rejections that still prove the
// expression was grammatically accepted.
var semanticFailureMarkers = []string{
	"Keyword `",
	"Unexpected keyword arguments",
	"Map cannot be defined here",
	"Lambda cannot be defined here",
	"doesn't exist",
	"Expected expression of type",
	"Expected integer",
	"Method `",
	"Function `",
	"`:` is not an infix operator",
	"Expected argument",
	"expects arguments",
	"Invalid ",
	"Failed to parse [a-z]+ pattern",
	"conflicting types",
	"could not be resolved",
	"wrong type",
}

func isSemanticFailure(stderr string) bool {
	for _, m := range semanticFailureMarkers {
		if strings.Contains(stderr, m) {
			return true
		}
	}
	return false
}

// sandboxEnv returns a neutral environment for jj: pristine user config, no
// pager, no editor, no colors.
func sandboxEnv() []string {
	env := os.Environ()
	env = append(env,
		"JJ_CONFIG=/dev/null",
		"JJ_EDITOR=false",
		"JJ_PAGER=cat",
		"NO_COLOR=1",
		"TERM=dumb",
	)
	return env
}

// newSandbox creates a throwaway jj repository in a temp dir with a pristine
// HOME so user config cannot influence the outcome.
func newSandbox(jjBin string) (dir string, cleanup func(), err error) {
	base, err := os.MkdirTemp("", "jjlex-grammardiff-")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp dir: %w", err)
	}
	fail := func(e error) (string, func(), error) {
		os.RemoveAll(base)
		return "", nil, e
	}
	repo := filepath.Join(base, "repo")
	home := filepath.Join(base, "home")
	if err := os.MkdirAll(home, 0o700); err != nil {
		return fail(fmt.Errorf("creating home dir: %w", err))
	}

	cmd := exec.Command(jjBin, "git", "init", repo)
	cmd.Dir = base
	cmd.Env = append(sandboxEnv(), "HOME="+home)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if initErr := cmd.Run(); initErr != nil {
		return fail(fmt.Errorf("jj git init: %w: %s", initErr, stderr.String()))
	}
	return repo, func() { os.RemoveAll(base) }, nil
}
