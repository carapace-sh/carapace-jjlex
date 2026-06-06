package template

import "strings"

// methodReturnType returns the template type returned by calling method on the given type.
// Returns empty string if the type or method is unknown.
// Handles Option<T> by delegating to the inner type's methods.
func methodReturnType(typeName, methodName string) string {
	if m, ok := typeMethods[typeName]; ok {
		if rt, ok := m[methodName]; ok {
			return rt
		}
	}
	// Option<T> delegates methods to T
	if inner, ok := strings.CutPrefix(typeName, "Option<"); ok {
		inner = strings.TrimSuffix(inner, ">")
		return methodReturnType(inner, methodName)
	}
	return ""
}

// globalFunctionReturnType returns the template type returned by a global function.
// Returns empty string if the function is unknown.
func globalFunctionReturnType(funcName string) string {
	return globalFuncTypes[funcName]
}

// typeMethods maps type names to their method return types.
var typeMethods = map[string]map[string]string{
	"Commit": {
		"description":            "String",
		"trailers":               "List<Trailer>",
		"change_id":              "ChangeId",
		"commit_id":              "CommitId",
		"parents":                "List<Commit>",
		"author":                 "Signature",
		"committer":              "Signature",
		"signature":              "Option<CryptographicSignature>",
		"mine":                   "Boolean",
		"working_copies":         "List<WorkspaceRef>",
		"current_working_copy":   "Boolean",
		"bookmarks":              "List<CommitRef>",
		"local_bookmarks":        "List<CommitRef>",
		"remote_bookmarks":       "List<CommitRef>",
		"tags":                   "List<CommitRef>",
		"local_tags":             "List<CommitRef>",
		"remote_tags":            "List<CommitRef>",
		"divergent":              "Boolean",
		"hidden":                 "Boolean",
		"change_offset":          "Option<Integer>",
		"immutable":              "Boolean",
		"contained_in":           "Boolean",
		"conflict":               "Boolean",
		"empty":                  "Boolean",
		"diff":                   "TreeDiff",
		"files":                  "List<TreeEntry>",
		"conflicted_files":       "List<TreeEntry>",
		"root":                   "Boolean",
	},
	"ChangeId": {
		"normal_hex": "String",
		"short":     "String",
		"shortest":  "ShortestIdPrefix",
	},
	"CommitId": {
		"short":    "String",
		"shortest": "ShortestIdPrefix",
	},
	"String": {
		"len":           "Integer",
		"contains":      "Boolean",
		"match":         "String",
		"starts_with":   "Boolean",
		"ends_with":     "Boolean",
		"remove_prefix": "String",
		"remove_suffix": "String",
		"trim":          "String",
		"trim_start":    "String",
		"trim_end":      "String",
		"substr":        "String",
		"first_line":    "String",
		"lines":         "List<String>",
		"split":         "List<String>",
		"replace":       "String",
		"upper":         "String",
		"lower":         "String",
		"escape_json":   "String",
	},
	"ByteString": {
		"len":           "Integer",
		"contains":      "Boolean",
		"match":         "ByteString",
		"starts_with":   "Boolean",
		"ends_with":     "Boolean",
		"remove_prefix": "ByteString",
		"remove_suffix": "ByteString",
		"trim":          "ByteString",
		"trim_start":    "ByteString",
		"trim_end":      "ByteString",
		"substr":        "ByteString",
		"first_line":    "ByteString",
		"lines":         "List<ByteString>",
		"split":         "List<ByteString>",
		"replace":       "ByteString",
		"upper":         "ByteString",
		"lower":         "ByteString",
	},
	"Signature": {
		"name":      "String",
		"email":     "Email",
		"timestamp": "Timestamp",
	},
	"Email": {
		"local":  "String",
		"domain": "String",
	},
	"Timestamp": {
		"ago":    "String",
		"format": "String",
		"utc":    "Timestamp",
		"local":  "Timestamp",
		"after":  "Boolean",
		"before": "Boolean",
		"since":  "TimestampRange",
	},
	"TimestampRange": {
		"start":    "Timestamp",
		"end":      "Timestamp",
		"duration": "String",
	},
	"Operation": {
		"current_operation": "Boolean",
		"description":       "String",
		"id":                "OperationId",
		"attributes":        "String",
		"time":              "TimestampRange",
		"user":              "String",
		"snapshot":          "Boolean",
		"workspace_name":    "String",
		"root":              "Boolean",
		"parents":           "List<Operation>",
	},
	"OperationId": {
		"short": "String",
	},
	"CommitRef": {
		"name":                "RefSymbol",
		"remote":              "Option<RefSymbol>",
		"present":             "Boolean",
		"conflict":            "Boolean",
		"normal_target":       "Option<Commit>",
		"removed_targets":     "List<Commit>",
		"added_targets":      "List<Commit>",
		"tracked":             "Boolean",
		"tracking_present":    "Boolean",
		"tracking_ahead_count":  "SizeHint",
		"tracking_behind_count": "SizeHint",
		"synced":              "Boolean",
	},
	"RefSymbol": {},
	"ConfigValue": {
		"as_boolean":     "Boolean",
		"as_integer":     "Integer",
		"as_string":      "String",
		"as_string_list": "List<String>",
	},
	"CryptographicSignature": {
		"status":  "String",
		"key":     "String",
		"display": "String",
	},
	"ShortestIdPrefix": {
		"prefix": "String",
		"rest":   "String",
		"upper":  "ShortestIdPrefix",
		"lower":  "ShortestIdPrefix",
	},
	"TreeDiff": {
		"files":       "List<TreeDiffEntry>",
		"color_words": "Template",
		"git":         "Template",
		"stat":        "DiffStats",
		"summary":     "Template",
	},
	"TreeDiffEntry": {
		"path":             "RepoPath",
		"display_diff_path": "String",
		"status":           "String",
		"status_char":      "String",
		"source":           "TreeEntry",
		"target":           "TreeEntry",
	},
	"TreeEntry": {
		"path":               "RepoPath",
		"conflict":           "Boolean",
		"conflict_side_count": "Integer",
		"file_type":          "String",
		"executable":         "Boolean",
	},
	"RepoPath": {
		"absolute": "String",
		"display":  "String",
		"parent":   "Option<RepoPath>",
	},
	"DiffStats": {
		"files":         "List<DiffStatEntry>",
		"total_added":   "Integer",
		"total_removed": "Integer",
	},
	"DiffStatEntry": {
		"bytes_delta":      "Integer",
		"lines_added":      "Integer",
		"lines_removed":    "Integer",
		"path":             "RepoPath",
		"display_diff_path": "String",
		"status":           "String",
		"status_char":      "String",
	},
	"Trailer": {
		"key":   "String",
		"value": "String",
	},
	"WorkspaceRef": {
		"name":   "RefSymbol",
		"target": "Commit",
		"root":   "Template",
	},
	"SizeHint": {
		"lower": "Integer",
		"upper": "Option<Integer>",
		"exact": "Option<Integer>",
		"zero":  "Boolean",
	},
	"AnnotationLine": {
		"commit":               "Commit",
		"content":              "ByteString",
		"line_number":          "Integer",
		"original_line_number": "Integer",
		"first_line_in_hunk":   "Boolean",
	},
	"CommitEvolutionEntry": {
		"commit":       "Commit",
		"operation":    "Operation",
		"predecessors": "List<Commit>",
		"inter_diff":   "TreeDiff",
	},
	"RegexCaptures": {
		"len":  "Integer",
		"get":  "ByteString",
		"name": "ByteString",
	},
	"List": {
		"len":     "Integer",
		"join":    "Template",
		"filter":  "List",
		"map":     "AnyList",
		"any":     "Boolean",
		"all":     "Boolean",
		"first":   "",
		"last":    "",
		"get":     "",
		"reverse": "List",
		"skip":    "List",
		"take":    "List",
	},
	"List<Trailer>": {
		"contains_key": "Boolean",
	},
	"AnyList": {
		"join": "Template",
	},
}

// globalFuncTypes maps global function names to their return types.
var globalFuncTypes = map[string]string{
	"fill":                "Template",
	"indent":              "Template",
	"pad_start":           "Template",
	"pad_end":             "Template",
	"pad_centered":        "Template",
	"truncate_start":      "Template",
	"truncate_end":        "Template",
	"hash":                "String",
	"label":               "Template",
	"hyperlink":           "Template",
	"raw_escape_sequence": "Template",
	"stringify":           "String",
	"json":                "String",
	"if":                  "",
	"coalesce":            "Template",
	"concat":              "Template",
	"join":                "Template",
	"separate":            "Template",
	"surround":            "Template",
	"config":              "Option<ConfigValue>",
	"git_web_url":         "String",
	"replace":             "Template",
}

// commitKeywords maps Commit zero-arg method names to their return types.
// These are available as top-level keywords in commit templates (e.g. `description`
// is equivalent to `self.description()`).
var commitKeywords = map[string]string{
	"description":          "String",
	"trailers":            "List<Trailer>",
	"change_id":           "ChangeId",
	"commit_id":           "CommitId",
	"parents":             "List<Commit>",
	"author":              "Signature",
	"committer":           "Signature",
	"signature":           "Option<CryptographicSignature>",
	"mine":                "Boolean",
	"working_copies":      "List<WorkspaceRef>",
	"current_working_copy": "Boolean",
	"bookmarks":           "List<CommitRef>",
	"local_bookmarks":     "List<CommitRef>",
	"remote_bookmarks":    "List<CommitRef>",
	"tags":                "List<CommitRef>",
	"local_tags":          "List<CommitRef>",
	"remote_tags":         "List<CommitRef>",
	"divergent":           "Boolean",
	"hidden":              "Boolean",
	"change_offset":       "Option<Integer>",
	"immutable":           "Boolean",
	"conflict":            "Boolean",
	"empty":               "Boolean",
	"root":                "Boolean",
}
