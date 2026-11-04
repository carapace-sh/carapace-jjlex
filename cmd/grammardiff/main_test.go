package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestClassifyJJ(t *testing.T) {
	cases := []struct {
		name   string
		lang   string
		stderr string
		runErr error
		want   string
	}{
		{
			name:   "success",
			lang:   "template",
			runErr: nil,
			want:   "accept",
		},
		{
			name:   "template syntax error",
			lang:   "template",
			stderr: "error: Failed to parse template: Syntax error\nCaused by:  --> 1:2\n  = expected <expression>\n",
			runErr: errors.New("exit status 1"),
			want:   "syntax",
		},
		{
			name:   "unknown keyword is semantic",
			lang:   "template",
			stderr: "error: Failed to parse template: Keyword `b` doesn't exist\nCaused by:  --> 1:4\n  = Keyword `b` doesn't exist\n",
			runErr: errors.New("exit status 1"),
			want:   "semantic",
		},
		{
			name:   "unexpected keyword arguments is semantic",
			lang:   "template",
			stderr: "error: Failed to parse template: Function `if`: Unexpected keyword arguments\n",
			runErr: errors.New("exit status 1"),
			want:   "semantic",
		},
		{
			name:   "map cannot be defined here is semantic",
			lang:   "template",
			stderr: "error: Failed to parse template: Map cannot be defined here\n",
			runErr: errors.New("exit status 1"),
			want:   "semantic",
		},
		{
			name:   "lambda cannot be defined here is semantic",
			lang:   "template",
			stderr: "error: Failed to parse template: Lambda cannot be defined here\n",
			runErr: errors.New("exit status 1"),
			want:   "semantic",
		},
		{
			name:   "expected expression of type is semantic",
			lang:   "template",
			stderr: "error: Failed to parse template: Expected expression of type `Template`, but actual type is `List<Commit>`\n",
			runErr: errors.New("exit status 1"),
			want:   "semantic",
		},
		{
			name:   "revset infix message for : pattern is semantic",
			lang:   "revset",
			stderr: "error: Failed to parse revset: `:` is not an infix operator\nCaused by:  --> 1:1\n",
			runErr: errors.New("exit status 1"),
			want:   "semantic",
		},
		{
			name:   "revset infix minus is syntax",
			lang:   "revset",
			stderr: "error: Failed to parse revset: `-` is not an infix operator\nCaused by:  --> 1:3\n",
			runErr: errors.New("exit status 1"),
			want:   "syntax",
		},
		{
			name:   "revset unknown revision at eval time is semantic",
			lang:   "revset",
			stderr: "Error: Revision `y` doesn't exist\n",
			runErr: errors.New("exit status 1"),
			want:   "semantic",
		},
		{
			name:   "no repo is harness",
			lang:   "template",
			stderr: "Error: There is no jj repo in \".\"\n",
			runErr: errors.New("exit status 1"),
			want:   "harness",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kind, _ := classifyJJ(tc.lang, tc.stderr, tc.runErr)
			if kind != tc.want {
				t.Errorf("classifyJJ(%q, %q) = %q, want %q", tc.lang, tc.stderr, kind, tc.want)
			}
		})
	}
}

func TestVerdict(t *testing.T) {
	cases := []struct {
		dev, jj, want string
	}{
		{"accept", "accept", "match"},
		{"accept", "semantic", "match"},
		{"accept", "syntax", "mismatch"},
		{"reject", "syntax", "match"},
		{"reject", "accept", "mismatch"},
		{"reject", "semantic", "mismatch"},
		{"accept", "harness", "harness"},
		{"reject", "harness", "harness"},
	}
	for _, tc := range cases {
		if got := verdict(tc.dev, tc.jj); got != tc.want {
			t.Errorf("verdict(%q, %q) = %q, want %q", tc.dev, tc.jj, got, tc.want)
		}
	}
}

// languages is a package-level var in main.go; the test just verifies the
// corpus covers every supported language in that order.
func TestDefaultCorpusNonEmpty(t *testing.T) {
	for _, lang := range languages {
		if len(defaultCorpus[lang]) == 0 {
			t.Errorf("default corpus has no entries for %q", lang)
		}
	}
	want := []string{"revset", "fileset", "template"}
	if !reflect.DeepEqual(languages, want) {
		t.Errorf("language iteration order = %v, want %v", languages, want)
	}
}
