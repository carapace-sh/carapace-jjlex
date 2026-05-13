package jjlex

import "testing"

func TestSplit(t *testing.T) {
	for revset, expected := range map[string]CompletionContext{
		"rev": {
			Prefix: "rev",
			Type:   CompletionTypeRevision,
		},
		"rev-": {
			Prefix: "rev-",
			Type:   CompletionTypeRevision,
		},
		"rev&": {
			Prefix: "rev&",
			Type:   CompletionTypeRevision,
		},
		"rev &": {
			Prefix: "",
			Type:   CompletionTypeRevision,
		},
		"rev & ": {
			Prefix: "",
			Type:   CompletionTypeRevision,
		},
		"parent(": {
			Type: CompletionTypeFunctionArg,
		},
		"parent(a": {
			Prefix: "a",
			Type:   CompletionTypeFunctionArg,
		},
	} {
		t.Run(revset, func(t *testing.T) {
			actual := Split(revset)
			if actual.Prefix != expected.Prefix {
				t.Fatalf("wrong prefix (expected: %s, was: %s)", expected.Prefix, actual.Prefix)
			}
			if actual.Type != expected.Type {
				t.Fatalf("wrong type (expected: %s, was: %s)", expected.Type, actual.Type)
			}
		})
	}

}
