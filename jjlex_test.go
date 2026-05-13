package jjlex

import "testing"

func TestSplit(t *testing.T) {
	for revset, expected := range map[string]CompletionContext{
		"":                         {Prefix: "", Type: CompletionTypeRevision},
		" ":                        {Prefix: "", Type: CompletionTypeRevision},
		" r":                       {Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev":                      {Prefix: "rev", Type: CompletionTypeRevision, AttachedRevset: "rev"},
		"rev-":                     {Prefix: "rev-", Type: CompletionTypeRevision, AttachedRevset: "rev-"},
		"rev&":                     {Prefix: "", Type: CompletionTypeRevision},
		"rev&r":                    {Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev &":                    {Prefix: "", Type: CompletionTypeRevision},
		"rev & ":                   {Prefix: "", Type: CompletionTypeRevision},
		"parents(":                 {Type: CompletionTypeFunctionArg},
		"parents(a":                {Prefix: "a", Type: CompletionTypeFunctionArg, ArgumentIndex: 0, AttachedRevset: "a"},
		"parents(a)":               {Prefix: "", Type: CompletionTypeOperator, AttachedRevset: "parents(a)"},
		"parents(a)|":              {Prefix: "", Type: CompletionTypeRevision},
		"parents(a)| ":             {Prefix: "", Type: CompletionTypeRevision},
		"parents(a) |":             {Prefix: "", Type: CompletionTypeRevision},
		"parents(a) | ":            {Prefix: "", Type: CompletionTypeRevision},
		"parents(a,":               {Prefix: "", Type: CompletionTypeFunctionArg, ArgumentIndex: 1},
		"parents(a,b":              {Prefix: "b", Type: CompletionTypeFunctionArg, ArgumentIndex: 1, AttachedRevset: "b"},
		"parents(a, b":             {Prefix: "b", Type: CompletionTypeFunctionArg, ArgumentIndex: 1, AttachedRevset: "b"},
		":":                        {Prefix: ":", Type: CompletionTypeOperator},
		"::":                       {Prefix: "", Type: CompletionTypeRevision, AttachedRevset: "::"},
		"::r":                      {Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "::r"},
		"(":                        {Prefix: "", Type: CompletionTypeRevision},
		"(y ":                      {Prefix: "", Type: CompletionTypeOperator},
		"(y & z":                   {Prefix: "z", Type: CompletionTypeRevision, AttachedRevset: "z"},
		"(y & z)":                  {Prefix: "", Type: CompletionTypeOperator, AttachedRevset: "(y & z)"},
		"(y & z) ":                 {Prefix: "", Type: CompletionTypeOperator},
		"(y.":                      {Prefix: ".", Type: CompletionTypeOperator, AttachedRevset: "y"},
		"(y..":                     {Prefix: "", Type: CompletionTypeRevision, AttachedRevset: "y.."},
		"heads(::@ & bookmark":     {Prefix: "bookmark", Type: CompletionTypeRevision},
		"heads(::@ & bookmarks())": {Prefix: "", Type: CompletionTypeOperator},
	} {
		t.Run(revset, func(t *testing.T) {
			actual := Split(revset)
			if actual.Prefix != expected.Prefix {
				t.Fatalf("wrong prefix (expected: %s, was: %s)", expected.Prefix, actual.Prefix)
			}
			if actual.Type != expected.Type {
				t.Fatalf("wrong type (expected: %s, was: %s)", expected.Type, actual.Type)
			}
			if actual.AttachedRevset != expected.AttachedRevset {
				t.Fatalf("wrong attached revset (expected: %s, was: %s)", expected.AttachedRevset, actual.AttachedRevset)
			}
		})
	}

}
