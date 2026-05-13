package jjlex

import "testing"

func TestSplit(t *testing.T) {
	for revset, expected := range map[string]CompletionContext{
		"":                         {Prefix: "", Type: CompletionTypeRevision},
		"rev":                      {Prefix: "rev", Type: CompletionTypeRevision},
		"rev-":                     {Prefix: "rev-", Type: CompletionTypeRevision},
		"rev&":                     {Prefix: "rev&", Type: CompletionTypeRevision},
		"rev &":                    {Prefix: "", Type: CompletionTypeRevision},
		"rev & ":                   {Prefix: "", Type: CompletionTypeRevision},
		"parents(":                 {Type: CompletionTypeFunctionArg},
		"parents(a":                {Prefix: "a", Type: CompletionTypeFunctionArg, ArgumentIndex: 0},
		"parents(a)":               {Prefix: "", Type: CompletionTypeOperator},
		"parents(a)|":              {Prefix: "", Type: CompletionTypeRevision},
		"parents(a)| ":             {Prefix: "", Type: CompletionTypeRevision},
		"parents(a) |":             {Prefix: "", Type: CompletionTypeRevision},
		"parents(a) | ":            {Prefix: "", Type: CompletionTypeRevision},
		"parents(a,":               {Prefix: "", Type: CompletionTypeFunctionArg, ArgumentIndex: 1},
		"parents(a,b":              {Prefix: "b", Type: CompletionTypeFunctionArg, ArgumentIndex: 1},
		"parents(a, b":             {Prefix: "b", Type: CompletionTypeFunctionArg, ArgumentIndex: 1},
		":":                        {Prefix: ":", Type: CompletionTypeOperator},
		"::":                       {Prefix: "", Type: CompletionTypeRevision},
		"::r":                      {Prefix: "r", Type: CompletionTypeRevision},
		"(":                        {Prefix: "", Type: CompletionTypeRevision},
		"(y ":                      {Prefix: "", Type: CompletionTypeOperator},
		"(y & z":                   {Prefix: "z", Type: CompletionTypeRevision},
		"(y & z)":                  {Prefix: "", Type: CompletionTypeOperator},
		"(y & z) ":                 {Prefix: "", Type: CompletionTypeOperator},
		"(y.":                      {Prefix: ".", Type: CompletionTypeOperator},
		"(y..":                     {Prefix: "", Type: CompletionTypeRevision},
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
		})
	}

}
