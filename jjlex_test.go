package jjlex

import "testing"

func TestSplit(t *testing.T) {
	for revset, expected := range map[string]CompletionContext{
		"":                        {Prefix: "", Type: CompletionTypeRevision},
		" ":                       {Prefix: "", Type: CompletionTypeRevision},
		" r":                      {Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev":                     {Prefix: "rev", Type: CompletionTypeRevision, AttachedRevset: "rev"},
		"123":                     {Prefix: "123", Type: CompletionTypeRevision, AttachedRevset: "123"},
		"123a":                    {Prefix: "123a", Type: CompletionTypeRevision, AttachedRevset: "123a"},
		"123a/":                   {Prefix: "123a/", Type: CompletionTypeRevision, AttachedRevset: "123a/"},
		"123a/b":                  {Prefix: "123a/b", Type: CompletionTypeRevision, AttachedRevset: "123a/b"},
		"v1":                      {Prefix: "v1", Type: CompletionTypeRevision, AttachedRevset: "v1"},
		"v1.":                     {Prefix: "v1.", Type: CompletionTypeRevision, AttachedRevset: "v1."},
		"v1.0.0":                  {Prefix: "v1.0.0", Type: CompletionTypeRevision, AttachedRevset: "v1.0.0"},
		"v1.0.0-beta":             {Prefix: "v1.0.0-beta", Type: CompletionTypeRevision, AttachedRevset: "v1.0.0-beta"},
		"rev@":                    {Prefix: "rev@", Type: CompletionTypeRevision, AttachedRevset: "rev@"},
		"rev@origin":              {Prefix: "rev@origin", Type: CompletionTypeRevision, AttachedRevset: "rev@origin"},
		"rev/push-hash":           {Prefix: "rev/push-hash", Type: CompletionTypeRevision, AttachedRevset: "rev/push-hash"},
		"rev/push-hash@":          {Prefix: "rev/push-hash@", Type: CompletionTypeRevision, AttachedRevset: "rev/push-hash@"},
		"rev/push-hash@origin":    {Prefix: "rev/push-hash@origin", Type: CompletionTypeRevision, AttachedRevset: "rev/push-hash@origin"},
		"rev-":                    {Prefix: "rev-", Type: CompletionTypeRevision, AttachedRevset: "rev-"},
		"rev&":                    {Prefix: "", Type: CompletionTypeRevision},
		"rev&r":                   {Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev&r-":                  {Prefix: "r-", Type: CompletionTypeRevision, AttachedRevset: "r-"},
		"rev &":                   {Prefix: "", Type: CompletionTypeRevision},
		"rev & ":                  {Prefix: "", Type: CompletionTypeRevision},
		"rev & r":                 {Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev & r-":                {Prefix: "r-", Type: CompletionTypeRevision, AttachedRevset: "r-"},
		"rev|":                    {Prefix: "", Type: CompletionTypeRevision},
		"rev|r":                   {Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev|r-":                  {Prefix: "r-", Type: CompletionTypeRevision, AttachedRevset: "r-"},
		"rev |":                   {Prefix: "", Type: CompletionTypeRevision},
		"rev | ":                  {Prefix: "", Type: CompletionTypeRevision},
		"rev | r":                 {Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev | r-":                {Prefix: "r-", Type: CompletionTypeRevision, AttachedRevset: "r-"},
		"parents(":                {Type: CompletionTypeFunctionArg, FunctionName: "parents"},
		"parents(a":               {Prefix: "a", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "a"},
		"parents(1":               {Prefix: "1", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "1"},
		"parents(1a":              {Prefix: "1a", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "1a"},
		"parents(1a/":             {Prefix: "1a/", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "1a/"},
		"parents(1a/b":            {Prefix: "1a/b", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "1a/b"},
		"parents(v1":              {Prefix: "v1", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "v1"},
		"parents(v1.":             {Prefix: "v1.", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "v1."},
		"parents(v1.0.0":          {Prefix: "v1.0.0", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "v1.0.0"},
		"parents(v1.0.0-beta":     {Prefix: "v1.0.0-beta", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "v1.0.0-beta"},
		"parents(a)":              {Prefix: "", Type: CompletionTypeOperator, AttachedRevset: "parents(a)"},
		"parents(a)-":             {Prefix: "", Type: CompletionTypeOperator, AttachedRevset: "parents(a)-"},
		"parents(a)|":             {Prefix: "", Type: CompletionTypeRevision},
		"parents(a)| ":            {Prefix: "", Type: CompletionTypeRevision},
		"parents(a) |":            {Prefix: "", Type: CompletionTypeRevision},
		"parents(a) | ":           {Prefix: "", Type: CompletionTypeRevision},
		"parents(a,":              {Prefix: "", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 1},
		"parents(a,b":             {Prefix: "b", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 1, AttachedRevset: "b"},
		"parents(a, b":            {Prefix: "b", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 1, AttachedRevset: "b"},
		"parents(children(a":      {Prefix: "a", Type: CompletionTypeFunctionArg, FunctionName: "children", ArgumentIndex: 0, AttachedRevset: "a"},
		":":                       {Prefix: ":", Type: CompletionTypeOperator},
		"::":                      {Prefix: "", Type: CompletionTypeRevision, AttachedRevset: "::"},
		"::r":                     {Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "::r"},
		"(":                       {Prefix: "", Type: CompletionTypeRevision},
		"(y ":                     {Prefix: "", Type: CompletionTypeOperator},
		"(y & z":                  {Prefix: "z", Type: CompletionTypeRevision},
		"(y & z)":                 {Prefix: "", Type: CompletionTypeOperator},
		"(y & z) ":                {Prefix: "", Type: CompletionTypeOperator},
		"(y.":                     {Prefix: ".", Type: CompletionTypeOperator},
		"(y..":                    {Prefix: "", Type: CompletionTypeRevision},
		"bookmark()":              {Prefix: "bookmark()", Type: CompletionTypeOperator, AttachedRevset: "bookmark()"},
		"heads(::@ & bookmark":    {Prefix: "bookmark", Type: CompletionTypeRevision, AttachedRevset: "bookmark"},
		"heads(::@ & bookmark()":  {Prefix: "bookmark()", Type: CompletionTypeOperator, AttachedRevset: "bookmark()"},
		"heads(::@ & bookmark())": {Prefix: "", Type: CompletionTypeOperator, AttachedRevset: "heads(::@ & bookmark())"},
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
			if actual.FunctionName != expected.FunctionName {
				t.Fatalf("wrong function name (expected: %s, was: %s)", expected.FunctionName, actual.FunctionName)
			}
			if actual.ArgumentIndex != expected.ArgumentIndex {
				t.Fatalf("wrong argument index (expected: %v, was: %v)", expected.ArgumentIndex, actual.ArgumentIndex)
			}
		})
	}

}
