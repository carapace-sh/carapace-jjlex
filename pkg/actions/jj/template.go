package jj

import (
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
	return carapace.Batch(
		ActionTemplateFunctions(),
		ActionStringPatterns().Suffix(":"),
	).ToA().NoSpace()
}

func actionForTemplatePatternValue(ctx *template.CompletionContext) carapace.Action {
	switch ctx.PatternName {
	case "exact", "exact-i", "substring", "substring-i", "glob", "glob-i", "regex", "regex-i":
		return ActionStringPatterns().Suffix(":").NoSpace()
	default:
		return carapace.ActionValues()
	}
}
