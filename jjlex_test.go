package jjlex

import "testing"

func TestSplit(t *testing.T) {
	for revset, expected := range map[string]CompletionContext{
		"":                          {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		" ":                         {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		" r":                        {IsValid: true, Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev":                       {IsValid: true, Prefix: "rev", Type: CompletionTypeRevision, AttachedRevset: "rev"},
		"123":                       {IsValid: true, Prefix: "123", Type: CompletionTypeRevision, AttachedRevset: "123"},
		"123a":                      {IsValid: true, Prefix: "123a", Type: CompletionTypeRevision, AttachedRevset: "123a"},
		"123a/":                     {IsValid: true, Prefix: "123a/", Type: CompletionTypeRevision, AttachedRevset: "123a/"},
		"123a/b":                    {IsValid: true, Prefix: "123a/b", Type: CompletionTypeRevision, AttachedRevset: "123a/b"},
		"v1":                        {IsValid: true, Prefix: "v1", Type: CompletionTypeRevision, AttachedRevset: "v1"},
		"v1.":                       {IsValid: true, Prefix: "v1.", Type: CompletionTypeRevision, AttachedRevset: "v1."},
		"v1.0.0":                    {IsValid: true, Prefix: "v1.0.0", Type: CompletionTypeRevision, AttachedRevset: "v1.0.0"},
		"v1.0.0-beta":               {IsValid: true, Prefix: "v1.0.0-beta", Type: CompletionTypeRevision, AttachedRevset: "v1.0.0-beta"},
		"rev@":                      {IsValid: true, Prefix: "rev@", Type: CompletionTypeRevision, AttachedRevset: "rev@"},
		"rev@origin":                {IsValid: true, Prefix: "rev@origin", Type: CompletionTypeRevision, AttachedRevset: "rev@origin"},
		"rev/push-hash":             {IsValid: true, Prefix: "rev/push-hash", Type: CompletionTypeRevision, AttachedRevset: "rev/push-hash"},
		"rev/push-hash@":            {IsValid: true, Prefix: "rev/push-hash@", Type: CompletionTypeRevision, AttachedRevset: "rev/push-hash@"},
		"rev/push-hash@origin":      {IsValid: true, Prefix: "rev/push-hash@origin", Type: CompletionTypeRevision, AttachedRevset: "rev/push-hash@origin"},
		"rev-":                      {IsValid: true, Prefix: "rev-", Type: CompletionTypeRevision, AttachedRevset: "rev-"},
		"rev&":                      {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"rev&r":                     {IsValid: true, Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev&r-":                    {IsValid: true, Prefix: "r-", Type: CompletionTypeRevision, AttachedRevset: "r-"},
		"rev &":                     {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"rev & ":                    {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"rev & r":                   {IsValid: true, Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev & r-":                  {IsValid: true, Prefix: "r-", Type: CompletionTypeRevision, AttachedRevset: "r-"},
		"rev|":                      {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"rev|r":                     {IsValid: true, Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev|r-":                    {IsValid: true, Prefix: "r-", Type: CompletionTypeRevision, AttachedRevset: "r-"},
		"rev |":                     {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"rev | ":                    {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"rev | r":                   {IsValid: true, Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "r"},
		"rev | r-":                  {IsValid: true, Prefix: "r-", Type: CompletionTypeRevision, AttachedRevset: "r-"},
		"parents(":                  {IsValid: true, Type: CompletionTypeFunctionArg, FunctionName: "parents", ExpectingRevset: true},
		"parents(a":                 {IsValid: true, Prefix: "a", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "a", ExpectingRevset: true},
		"parents(1":                 {IsValid: true, Prefix: "1", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "1", ExpectingRevset: true},
		"parents(1a":                {IsValid: true, Prefix: "1a", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "1a", ExpectingRevset: true},
		"parents(1a/":               {IsValid: true, Prefix: "1a/", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "1a/", ExpectingRevset: true},
		"parents(1a/b":              {IsValid: true, Prefix: "1a/b", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "1a/b", ExpectingRevset: true},
		"parents(v1":                {IsValid: true, Prefix: "v1", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "v1", ExpectingRevset: true},
		"parents(v1.":               {IsValid: true, Prefix: "v1.", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "v1.", ExpectingRevset: true},
		"parents(v1.0.0":            {IsValid: true, Prefix: "v1.0.0", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "v1.0.0", ExpectingRevset: true},
		"parents(v1.0.0-beta":       {IsValid: true, Prefix: "v1.0.0-beta", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "v1.0.0-beta", ExpectingRevset: true},
		"parents(a)":                {IsValid: true, Prefix: "", Type: CompletionTypeOperator, AttachedRevset: "parents(a)"},
		"parents(a)-":               {IsValid: true, Prefix: "", Type: CompletionTypeOperator, AttachedRevset: "parents(a)-"},
		"parents(a)|":               {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"parents(a)| ":              {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"parents(a) |":              {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"parents(a) | ":             {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"parents(a,":                {IsValid: true, Prefix: "", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 1, ExpectingRevset: false},
		"parents(a,1":               {IsValid: true, Prefix: "1", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 1, AttachedRevset: "", ExpectingRevset: false},
		"parents(a, 2":              {IsValid: true, Prefix: "2", Type: CompletionTypeFunctionArg, FunctionName: "parents", ArgumentIndex: 1, AttachedRevset: "", ExpectingRevset: false},
		"parents(children(a":        {IsValid: true, Prefix: "a", Type: CompletionTypeFunctionArg, FunctionName: "children", ArgumentIndex: 0, AttachedRevset: "a", ExpectingRevset: true},
		"parents(children(a,3":      {IsValid: true, Prefix: "3", Type: CompletionTypeFunctionArg, FunctionName: "children", ArgumentIndex: 1, AttachedRevset: "", ExpectingRevset: false},
		":":                         {IsValid: true, Prefix: ":", Type: CompletionTypeOperator},
		"::":                        {IsValid: true, Prefix: "", Type: CompletionTypeRevision, AttachedRevset: "::"},
		"::r":                       {IsValid: true, Prefix: "r", Type: CompletionTypeRevision, AttachedRevset: "::r"},
		"(":                         {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"(y ":                       {IsValid: true, Prefix: "", Type: CompletionTypeOperator},
		"(y & z":                    {IsValid: true, Prefix: "z", Type: CompletionTypeRevision},
		"(y & z)":                   {IsValid: true, Prefix: "", Type: CompletionTypeOperator},
		"(y & z) ":                  {IsValid: true, Prefix: "", Type: CompletionTypeOperator},
		"(y.":                       {IsValid: true, Prefix: ".", Type: CompletionTypeOperator},
		"(y..":                      {IsValid: true, Prefix: "", Type: CompletionTypeRevision},
		"bookmark()":                {IsValid: true, Prefix: "bookmark()", Type: CompletionTypeOperator, AttachedRevset: "bookmark()"},
		"heads(::@ & bookmark":      {IsValid: true, Prefix: "bookmark", Type: CompletionTypeRevision, AttachedRevset: "bookmark"},
		"heads(::@ & bookmark()":    {IsValid: true, Prefix: "bookmark()", Type: CompletionTypeOperator, AttachedRevset: "bookmark()"},
		"heads(::@ & bookmark())":   {IsValid: true, Prefix: "", Type: CompletionTypeOperator, AttachedRevset: "heads(::@ & bookmark())"},
		"description(":              {IsValid: true, Prefix: "", Type: CompletionTypeFunctionArg, FunctionName: "description", ArgumentIndex: 0},
		"description(a":             {IsValid: true, Prefix: "a", Type: CompletionTypeFunctionArg, FunctionName: "description", ArgumentIndex: 0},
		"description(a,err":         {IsValid: false},
		"\"parents(":                {IsValid: true, Prefix: "\"parents(", Type: CompletionTypeRevision, AttachedRevset: "\"parents(", ExpectingRevset: true},
		"\"parents(\"":              {IsValid: true, Prefix: "\"parents(\"", Type: CompletionTypeRevision, AttachedRevset: "\"parents(\"", ExpectingRevset: true},
		"\"parent(\"@git":           {IsValid: true, Prefix: "\"parent(\"@git", Type: CompletionTypeRevision, AttachedRevset: "\"parent(\"@git", ExpectingRevset: true},
		"parents(\"parents(\"":      {IsValid: true, Prefix: "\"parents(\"", Type: CompletionTypeRevision, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "\"parents(\"", ExpectingRevset: true},
		"parents(\"parents(\"@git":  {IsValid: true, Prefix: "\"parents(\"@git", Type: CompletionTypeRevision, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "\"parents(\"@git", ExpectingRevset: true},
		"parents(\"parents(\"-":     {IsValid: true, Prefix: "\"parents(\"-", Type: CompletionTypeRevision, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "\"parents(\"-", ExpectingRevset: true},
		"parents(\"parents(\"@git-": {IsValid: true, Prefix: "\"parents(\"@git-", Type: CompletionTypeRevision, FunctionName: "parents", ArgumentIndex: 0, AttachedRevset: "\"parents(\"@git-", ExpectingRevset: true},
	} {
		t.Run(revset, func(t *testing.T) {
			actual := Split(revset)
			if actual.IsValid != expected.IsValid {
				t.Fatalf("wrong isvalid (expected: %v, was: %v)", expected.IsValid, actual.IsValid)
			}

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
			if actual.ExpectingRevset != expected.ExpectingRevset {
				t.Fatalf("wrong expecting revset (expected: %v, was: %v)", expected.ExpectingRevset, actual.ExpectingRevset)
			}
		})
	}

}
