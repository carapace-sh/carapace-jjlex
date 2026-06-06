package jj

import (
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-jjlex/pkg/template"
)

// ActionTemplateFunctions completes template global function names.
//
//	if (conditional)
//	label (apply color label)
func ActionTemplateFunctions() carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		noArgs := carapace.ActionValuesDescribed(
			"true", "Boolean true literal",
			"false", "Boolean false literal",
		).Uid("jj", "template-function", "args", "false")

		withArgs := carapace.ActionValuesDescribed(
			"fill", "Fill lines at width",
			"indent", "Indent non-empty lines with prefix",
			"pad_start", "Left-justify with fill chars",
			"pad_end", "Right-justify with fill chars",
			"pad_centered", "Center with fill chars",
			"truncate_start", "Truncate from start",
			"truncate_end", "Truncate from end",
			"hash", "Hash and return hex digest",
			"label", "Apply color label",
			"hyperlink", "Render OSC 8 hyperlink",
			"raw_escape_sequence", "Preserve escape sequences",
			"stringify", "Format content to string",
			"json", "Serialize to JSON",
			"if", "Conditional evaluation",
			"coalesce", "First non-empty content",
			"concat", "Concatenate all",
			"join", "Insert separator between items",
			"separate", "Insert separator between non-empty items",
			"surround", "Wrap non-empty content",
			"config", "Look up config value",
			"git_web_url", "Convert git URL to HTTPS browse URL",
			"replace", "Replace matches using pattern",
		).Uid("jj", "template-function", "args", "true")

		return carapace.Batch(noArgs, withArgs.Suffix("(")).ToA()
	}).Tag("template functions")
}

// ActionTemplateOperators completes template operators.
//
//	++ (concatenation)
//	&& (logical and)
func ActionTemplateOperators() carapace.Action {
	return carapace.ActionValuesDescribed(
		"++", "Concatenation",
		"||", "Logical or",
		"&&", "Logical and",
		"==", "Equal",
		"!=", "Not equal",
		">=", "Greater than or equal",
		">", "Greater than",
		"<=", "Less than or equal",
		"<", "Less than",
		"+", "Addition",
		"-", "Subtraction",
		"*", "Multiplication",
		"/", "Division",
		"%", "Remainder",
		"!", "Logical not",
	).Uid("jj", "template-operator").Tag("template operators")
}

// ActionTemplates completes template expressions with context-awareness
// using the template completion parser to determine what is expected at the cursor.
//
//	if(true, "yes", "no")
//	change_id.short() ++ "\n"
func ActionTemplates() carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		ctx := template.ParseForCompletion(c.Value)

		// Compute the prefix: everything before the partial identifier being typed.
		// Sub-actions filter against c.Value, so we need to strip this prefix
		// before invoking them and re-attach it to the completion values.
		prefix := c.Value[:len(c.Value)-len(ctx.PartialIdent)]

		if ctx.InPattern {
			return actionForTemplatePatternValue(ctx).Prefix(prefix)
		}

		if ctx.MethodType != "" {
			return ActionTemplateTypeMethods(ctx.MethodType).Prefix(prefix)
		}

		if ctx.Function != nil {
			return actionForTemplateFunctionArg(ctx).Prefix(prefix)
		}

		if expectsTemplateToken(ctx, template.ExpectedExpression) && expectsTemplateToken(ctx, template.ExpectedOperator) {
			// Both expression and operator are valid - combine both actions
			batch := carapace.Batch(
				actionTemplateExpression(ctx),
				ActionTemplateOperators(),
			)
			return batch.ToA().NoSpace().Prefix(prefix)
		}

		if expectsTemplateToken(ctx, template.ExpectedExpression) {
			return actionTemplateExpression(ctx).Prefix(prefix)
		}

		if expectsTemplateToken(ctx, template.ExpectedOperator) {
			return ActionTemplateOperators().NoSpace().Prefix(prefix)
		}

		if expectsTemplateToken(ctx, template.ExpectedClosingParen) {
			return carapace.ActionValues(")").Prefix(prefix)
		}

		if expectsTemplateToken(ctx, template.ExpectedComma) {
			return carapace.ActionValues(",").Prefix(prefix)
		}

		if expectsTemplateToken(ctx, template.ExpectedEquals) {
			return carapace.ActionValues("=").Prefix(prefix)
		}

		if expectsTemplateToken(ctx, template.ExpectedLambdaClose) {
			return carapace.ActionValues("|").Prefix(prefix)
		}

		return actionTemplateExpression(ctx).Prefix(prefix)
	})
}

func expectsTemplateToken(ctx *template.CompletionContext, token template.ExpectedToken) bool {
	for _, t := range ctx.ExpectedTokens {
		if t == token {
			return true
		}
	}
	return false
}

func actionTemplateExpression(_ *template.CompletionContext) carapace.Action {
	return carapace.Batch(
		ActionTemplateFunctions(),
		ActionStringPatterns().Suffix(":"),
	).ToA().NoSpace()
}

func actionForTemplateFunctionArg(ctx *template.CompletionContext) carapace.Action {
	if ctx.Function.IsZeroArg {
		return carapace.ActionValues(")")
	}

	switch ctx.Function.Name {
	case "config":
		return ActionConfigs(true).NoSpace()
	default:
		return carapace.Batch(
			ActionTemplateFunctions(),
			ActionStringPatterns().Suffix(":"),
		).ToA().NoSpace()
	}
}

// ActionTemplateTypeMethods completes method names for a template type.
//
//	description (Commit method: description as String)
//	short (String method: shortened form)
func ActionTemplateTypeMethods(typeName string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		methods := templateTypeMethods(typeName)
		if len(methods) == 0 {
			return carapace.ActionValues()
		}
		noArgs := make([]string, 0, len(methods)*2)
		withArgs := make([]string, 0, len(methods)*2)
		for _, m := range methods {
			if m.noArgs {
				noArgs = append(noArgs, m.name, m.description)
			} else {
				withArgs = append(withArgs, m.name, m.description)
			}
		}
		batch := carapace.Batch()
		if len(noArgs) > 0 {
			batch = append(batch, carapace.ActionValuesDescribed(noArgs...).
				Suffix("()").
				Uid("jj", "template-method-"+strings.ToLower(typeName), "args", "false"))
		}
		if len(withArgs) > 0 {
			batch = append(batch, carapace.ActionValuesDescribed(withArgs...).
				Suffix("(").
				Uid("jj", "template-method-"+strings.ToLower(typeName), "args", "true"))
		}
		return batch.ToA()
	}).Tag("template methods")
}

type templateMethod struct {
	name        string
	description string
	noArgs      bool
}

func templateTypeMethods(typeName string) []templateMethod {
	switch typeName {
	case "Commit":
		return []templateMethod{
			{"description", "Commit description as String", true},
			{"trailers", "List of Trailer objects", true},
			{"change_id", "ChangeId of the commit", true},
			{"commit_id", "CommitId of the commit", true},
			{"parents", "List of parent Commits", true},
			{"author", "Signature of the author", true},
			{"committer", "Signature of the committer", true},
			{"signature", "Cryptographic signature", true},
			{"mine", "Whether authored by current user", true},
			{"working_copies", "List of WorkspaceRef objects", true},
			{"current_working_copy", "Whether this is the current working copy", true},
			{"bookmarks", "List of all CommitRef bookmarks", true},
			{"local_bookmarks", "List of local CommitRef bookmarks", true},
			{"remote_bookmarks", "List of remote CommitRef bookmarks", true},
			{"tags", "List of all CommitRef tags", true},
			{"local_tags", "List of local CommitRef tags", true},
			{"remote_tags", "List of remote CommitRef tags", true},
			{"divergent", "Whether this commit is divergent", true},
			{"hidden", "Whether this commit is hidden", true},
			{"change_offset", "Change offset for divergent commits", true},
			{"immutable", "Whether this commit is immutable", true},
			{"contained_in", "Whether contained in revset", false},
			{"conflict", "Whether this commit has conflicts", true},
			{"empty", "Whether this commit modifies no files", true},
			{"diff", "TreeDiff of changes", false},
			{"files", "List of TreeEntry files", false},
			{"conflicted_files", "List of conflicted TreeEntry files", true},
			{"root", "Whether this is the root commit", true},
		}
	case "ChangeId":
		return []templateMethod{
			{"normal_hex", "Normal hex representation", true},
			{"short", "Shortened hex string", false},
			{"shortest", "Shortest unique prefix", false},
		}
	case "CommitId":
		return []templateMethod{
			{"short", "Shortened hex string", false},
			{"shortest", "Shortest unique prefix", false},
		}
	case "String":
		return []templateMethod{
			{"len", "Length of the string", true},
			{"contains", "Whether string contains needle", false},
			{"match", "Match against string pattern", false},
			{"starts_with", "Whether string starts with needle", false},
			{"ends_with", "Whether string ends with needle", false},
			{"remove_prefix", "Remove prefix if present", false},
			{"remove_suffix", "Remove suffix if present", false},
			{"trim", "Trim whitespace from both ends", true},
			{"trim_start", "Trim whitespace from start", true},
			{"trim_end", "Trim whitespace from end", true},
			{"substr", "Extract substring by index", false},
			{"first_line", "First line of the string", true},
			{"lines", "Split into List of lines", true},
			{"split", "Split by separator pattern", false},
			{"replace", "Replace matches with replacement", false},
			{"upper", "Convert to uppercase", true},
			{"lower", "Convert to lowercase", true},
			{"escape_json", "Escape for JSON string", true},
		}
	case "ByteString":
		return []templateMethod{
			{"len", "Length of the byte string", true},
			{"contains", "Whether byte string contains needle", false},
			{"match", "Match against string pattern", false},
			{"starts_with", "Whether byte string starts with needle", false},
			{"ends_with", "Whether byte string ends with needle", false},
			{"remove_prefix", "Remove prefix if present", false},
			{"remove_suffix", "Remove suffix if present", false},
			{"trim", "Trim whitespace from both ends", true},
			{"trim_start", "Trim whitespace from start", true},
			{"trim_end", "Trim whitespace from end", true},
			{"substr", "Extract substring by index", false},
			{"first_line", "First line of the byte string", true},
			{"lines", "Split into List of lines", true},
			{"split", "Split by separator pattern", false},
			{"replace", "Replace matches with replacement", false},
			{"upper", "Convert to uppercase", true},
			{"lower", "Convert to lowercase", true},
		}
	case "Signature":
		return []templateMethod{
			{"name", "Name of the person", true},
			{"email", "Email address as Email type", true},
			{"timestamp", "Timestamp of the signature", true},
		}
	case "Email":
		return []templateMethod{
			{"local", "Local part of the email", true},
			{"domain", "Domain part of the email", true},
		}
	case "Timestamp":
		return []templateMethod{
			{"ago", "Relative time string (e.g. '2 hours ago')", true},
			{"format", "Format with strftime pattern", false},
			{"utc", "Convert to UTC timezone", true},
			{"local", "Convert to local timezone", true},
			{"after", "Whether timestamp is after date", false},
			{"before", "Whether timestamp is before date", false},
			{"since", "TimestampRange from this timestamp", false},
		}
	case "TimestampRange":
		return []templateMethod{
			{"start", "Start timestamp", true},
			{"end", "End timestamp", true},
			{"duration", "Duration as String", true},
		}
	case "Operation":
		return []templateMethod{
			{"current_operation", "Whether this is the current operation", true},
			{"description", "Operation description as String", true},
			{"id", "OperationId of the operation", true},
			{"attributes", "Operation attributes as String", true},
			{"time", "TimestampRange of the operation", true},
			{"user", "User who performed the operation", true},
			{"snapshot", "Whether this is a snapshot operation", true},
			{"workspace_name", "Workspace name for the operation", true},
			{"root", "Whether this is the root operation", true},
			{"parents", "List of parent Operations", true},
		}
	case "OperationId":
		return []templateMethod{
			{"short", "Shortened hex string", false},
		}
	case "CommitRef":
		return []templateMethod{
			{"name", "RefSymbol name of the ref", true},
			{"remote", "Optional remote RefSymbol", true},
			{"present", "Whether the ref is present", true},
			{"conflict", "Whether the ref is in conflict", true},
			{"normal_target", "Normal target Commit if present", true},
			{"removed_targets", "List of removed target Commits", true},
			{"added_targets", "List of added target Commits", true},
			{"tracked", "Whether the ref is tracked", true},
			{"tracking_present", "Whether a tracking ref exists", true},
			{"tracking_ahead_count", "SizeHint of ahead count", true},
			{"tracking_behind_count", "SizeHint of behind count", true},
			{"synced", "Whether tracking is synced", true},
		}
	case "ConfigValue":
		return []templateMethod{
			{"as_boolean", "Convert to Boolean", true},
			{"as_integer", "Convert to Integer", true},
			{"as_string", "Convert to String", true},
			{"as_string_list", "Convert to List of Strings", true},
		}
	case "CryptographicSignature":
		return []templateMethod{
			{"status", "Signature verification status", true},
			{"key", "Key that signed the commit", true},
			{"display", "Display string for the signature", true},
		}
	case "ShortestIdPrefix":
		return []templateMethod{
			{"prefix", "Shortest unique prefix string", true},
			{"rest", "Remaining part of the ID", true},
			{"upper", "Uppercase version of prefix", true},
			{"lower", "Lowercase version of prefix", true},
		}
	case "TreeDiff":
		return []templateMethod{
			{"files", "List of TreeDiffEntry files", true},
			{"color_words", "Color words diff as Template", false},
			{"git", "Git-style diff as Template", false},
			{"stat", "DiffStats for the diff", false},
			{"summary", "Summary of the diff as Template", true},
		}
	case "TreeDiffEntry":
		return []templateMethod{
			{"path", "RepoPath of the file", true},
			{"display_diff_path", "Display path as String", true},
			{"status", "Status as String", true},
			{"status_char", "Status character as String", true},
			{"source", "Source TreeEntry", true},
			{"target", "Target TreeEntry", true},
		}
	case "TreeEntry":
		return []templateMethod{
			{"path", "RepoPath of the file", true},
			{"conflict", "Whether the file is conflicted", true},
			{"conflict_side_count", "Number of conflict sides", true},
			{"file_type", "File type as String", true},
			{"executable", "Whether the file is executable", true},
		}
	case "RepoPath":
		return []templateMethod{
			{"absolute", "Absolute path as String", true},
			{"display", "Display path as String", true},
			{"parent", "Parent directory as Option<RepoPath>", true},
		}
	case "DiffStats":
		return []templateMethod{
			{"files", "List of DiffStatEntry files", true},
			{"total_added", "Total lines added", true},
			{"total_removed", "Total lines removed", true},
		}
	case "DiffStatEntry":
		return []templateMethod{
			{"bytes_delta", "Byte size change", true},
			{"lines_added", "Lines added", true},
			{"lines_removed", "Lines removed", true},
			{"path", "RepoPath of the file", true},
			{"display_diff_path", "Display path as String", true},
			{"status", "Status as String", true},
			{"status_char", "Status character as String", true},
		}
	case "Trailer":
		return []templateMethod{
			{"key", "Trailer key as String", true},
			{"value", "Trailer value as String", true},
		}
	case "WorkspaceRef":
		return []templateMethod{
			{"name", "Workspace name as RefSymbol", true},
			{"target", "Target Commit", true},
			{"root", "Root as Template", true},
		}
	case "SizeHint":
		return []templateMethod{
			{"lower", "Lower bound as Integer", true},
			{"upper", "Upper bound as Option<Integer>", true},
			{"exact", "Exact value as Option<Integer>", true},
			{"zero", "Whether the count is zero", true},
		}
	case "AnnotationLine":
		return []templateMethod{
			{"commit", "Commit this line belongs to", true},
			{"content", "Line content as ByteString", true},
			{"line_number", "Line number as Integer", true},
			{"original_line_number", "Original line number as Integer", true},
			{"first_line_in_hunk", "Whether this is the first line in a hunk", true},
		}
	case "CommitEvolutionEntry":
		return []templateMethod{
			{"commit", "The Commit", true},
			{"operation", "The Operation", true},
			{"predecessors", "List of predecessor Commits", true},
			{"inter_diff", "Inter-diff as TreeDiff", false},
		}
	case "RegexCaptures":
		return []templateMethod{
			{"len", "Number of capture groups", true},
			{"get", "Capture group by index as ByteString", false},
			{"name", "Capture group by name as ByteString", false},
		}
	case "List":
		return []templateMethod{
			{"len", "Number of items", true},
			{"join", "Join with separator as Template", false},
			{"filter", "Filter items with lambda", false},
			{"map", "Map items with lambda", false},
			{"any", "Whether any item matches lambda", false},
			{"all", "Whether all items match lambda", false},
			{"first", "First item", true},
			{"last", "Last item", true},
			{"get", "Item by index", false},
			{"reverse", "Reversed list", true},
			{"skip", "Skip first N items", false},
			{"take", "Take first N items", false},
		}
	case "List<Trailer>":
		return []templateMethod{
			{"contains_key", "Whether list contains key", false},
		}
	case "AnyList":
		return []templateMethod{
			{"join", "Join with separator as Template", false},
		}
	default:
		return templateTypeMethodsGeneric(typeName)
	}
}

func templateTypeMethodsGeneric(typeName string) []templateMethod {
	if inner, ok := strings.CutPrefix(typeName, "Option<"); ok {
		inner = strings.TrimSuffix(inner, ">")
		return templateTypeMethods(inner)
	}
	if inner, ok := strings.CutPrefix(typeName, "List<"); ok {
		inner = strings.TrimSuffix(inner, ">")
		listMethods := []templateMethod{
			{"len", "Number of items", true},
			{"join", "Join with separator as Template", false},
			{"filter", "Filter items with lambda", false},
			{"map", "Map items with lambda", false},
			{"any", "Whether any item matches lambda", false},
			{"all", "Whether all items match lambda", false},
			{"first", "First item", true},
			{"last", "Last item", true},
			{"get", "Item by index", false},
			{"reverse", "Reversed list", true},
			{"skip", "Skip first N items", false},
			{"take", "Take first N items", false},
		}
		switch inner {
		case "Trailer":
			listMethods = append(listMethods, templateMethod{"contains_key", "Whether list contains key", false})
		}
		return listMethods
	}
	return nil
}

func actionForTemplatePatternValue(ctx *template.CompletionContext) carapace.Action {
	switch ctx.PatternName {
	case "exact", "exact-i", "substring", "substring-i", "glob", "glob-i", "regex", "regex-i":
		return ActionStringPatterns().Suffix(":").NoSpace()
	default:
		return carapace.ActionValues()
	}
}
