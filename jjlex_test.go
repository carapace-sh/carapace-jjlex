package jjlex

import "testing"

func TestSplit(t *testing.T) {
	for revset, expected := range map[string]CompletionContext{
		"":                         {Prefix: "", Type: CompletionTypeRevision},
		" ":                        {Prefix: "", Type: CompletionTypeRevision},
		" r":                       {Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev":                      {Prefix: "rev", Type: CompletionTypeRevision, AttachedRevset: "rev"},
		"123":                      {Prefix: "123", Type: CompletionTypeRevision, AttachedRevset: "123"},
		"123a":                     {Prefix: "123a", Type: CompletionTypeRevision, AttachedRevset: "123a"},
		"123a/":                    {Prefix: "123a/", Type: CompletionTypeRevision, AttachedRevset: "123a/"},
		"123a/b":                   {Prefix: "123a/b", Type: CompletionTypeRevision, AttachedRevset: "123a/b"},
		"rev@":                     {Prefix: "rev@", Type: CompletionTypeRevision, AttachedRevset: "rev@"},
		"rev@origin":               {Prefix: "rev@origin", Type: CompletionTypeRevision, AttachedRevset: "rev@origin"},
		"rev/push-hash":            {Prefix: "rev/push-hash", Type: CompletionTypeRevision, AttachedRevset: "rev/push-hash"},
		"rev/push-hash@":           {Prefix: "rev/push-hash@", Type: CompletionTypeRevision, AttachedRevset: "rev/push-hash@"},
		"rev/push-hash@origin":     {Prefix: "rev/push-hash@origin", Type: CompletionTypeRevision, AttachedRevset: "rev/push-hash@origin"},
		"rev-":                     {Prefix: "rev-", Type: CompletionTypeRevision, AttachedRevset: "rev-"},
		"rev&":                     {Prefix: "", Type: CompletionTypeRevision},
		"rev&r":                    {Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev &":                    {Prefix: "", Type: CompletionTypeRevision},
		"rev & ":                   {Prefix: "", Type: CompletionTypeRevision},
		"parents(":                 {Type: CompletionTypeFunctionArg},
		"parents(a":                {Prefix: "a", Type: CompletionTypeFunctionArg, ArgumentIndex: 0, AttachedRevset: "a"},
		"parents(1":                {Prefix: "1", Type: CompletionTypeFunctionArg, ArgumentIndex: 1, AttachedRevset: "1"},
		"parents(1a":               {Prefix: "1a", Type: CompletionTypeFunctionArg, ArgumentIndex: 1, AttachedRevset: "1a"},
		"parents(1a/":              {Prefix: "1a/", Type: CompletionTypeFunctionArg, ArgumentIndex: 1, AttachedRevset: "1a/"},
		"parents(1a/b":             {Prefix: "1a/b", Type: CompletionTypeFunctionArg, ArgumentIndex: 1, AttachedRevset: "1a/b"},
		"parents(a)":               {Prefix: "", Type: CompletionTypeOperator, AttachedRevset: "parents(a)"},
		"parents(a)-":              {Prefix: "", Type: CompletionTypeOperator, AttachedRevset: "parents(a)-"},
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
			if actual.AttachedRevset != expected.AttachedRevset {
				t.Fatalf("wrong attached revset (expected: %s, was: %s)", expected.AttachedRevset, actual.AttachedRevset)
			}
		})
	}

}
