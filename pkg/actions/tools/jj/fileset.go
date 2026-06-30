package jj

import (
	"slices"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-jjlex/pkg/fileset"
)

// ActionFilesetFunctions completes fileset function names.
//
//	all (Matches everything)
//	none (Matches nothing)
func ActionFilesetFunctions() carapace.Action {
	return carapace.ActionValuesDescribed(
		"all", "Matches everything",
		"none", "Matches nothing",
	).Suffix("()").Uid("jj", "fileset-function").Tag("fileset functions")
}

// ActionFilesetOperators completes fileset operators.
//
//	| (union)
//	& (intersection)
func ActionFilesetOperators() carapace.Action {
	return carapace.ActionValuesDescribed(
		"&", "x & y: Intersection (files in both x and y)",
		"|", "x | y: Union (files in x or y)",
		"~", "x ~ y: Difference (files in x but not in y); ~x: Negate (files not in x)",
	).Uid("jj", "fileset-operator").Tag("fileset operators")
}

// ActionFilesets completes fileset expressions with context-awareness
// using the fileset completion parser to determine what is expected at the cursor.
//
//	all()
//	glob:"*.rs" & ~glob:"test*"
func ActionFilesets() carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		ctx := fileset.ParseForCompletion(c.Value)

		// Compute the prefix: everything before the partial identifier being typed.
		// Sub-actions filter against c.Value, so we need to strip this prefix
		// before invoking them and re-attach it to the completion values.
		prefix := c.Value[:len(c.Value)-len(ctx.PartialIdent)]

		if ctx.InPattern {
			return actionForFilesetPatternValue(ctx).Prefix(prefix)
		}

		if ctx.Function != nil {
			return actionForFilesetFunctionArg(ctx).Prefix(prefix)
		}

		if expectsFilesetToken(ctx, fileset.ExpectedExpression) {
			return actionFilesetExpression(ctx).Prefix(prefix)
		}

		if expectsFilesetToken(ctx, fileset.ExpectedOperator) {
			// If ValidOperators is populated, filter to only those operators
			if len(ctx.ValidOperators) > 0 {
				batch := carapace.Batch()
				for _, op := range ctx.ValidOperators {
					batch = append(batch, carapace.ActionValuesDescribed(op.Op, op.Description))
				}
				return batch.ToA().NoSpace().Prefix(prefix)
			}
			return ActionFilesetOperators().NoSpace().Prefix(prefix)
		}

		if expectsFilesetToken(ctx, fileset.ExpectedClosingParen) {
			return carapace.ActionValues(")").Prefix(prefix)
		}

		if expectsFilesetToken(ctx, fileset.ExpectedComma) {
			return carapace.ActionValues(",").Prefix(prefix)
		}

		return actionFilesetExpression(ctx).Prefix(prefix)
	})
}

func expectsFilesetToken(ctx *fileset.CompletionContext, token fileset.ExpectedToken) bool {
	return slices.Contains(ctx.ExpectedTokens, token)
}

func actionFilesetExpression(_ *fileset.CompletionContext) carapace.Action {
	return carapace.Batch(
		ActionFilesetFunctions(),
		ActionFilesetPatterns().Suffix(":"),
		ActionRevFiles("@"),
	).ToA().NoSpace()
}

func actionForFilesetFunctionArg(ctx *fileset.CompletionContext) carapace.Action {
	if ctx.Function.IsZeroArg {
		return carapace.ActionValues(")")
	}
	return carapace.Batch(
		ActionFilesetFunctions(),
		ActionFilesetPatterns().Suffix(":"),
		ActionRevFiles("@"),
	).ToA().NoSpace()
}

func actionForFilesetPatternValue(ctx *fileset.CompletionContext) carapace.Action {
	switch ctx.PatternName {
	case "cwd", "file", "cwd-file", "glob", "cwd-glob", "prefix-glob", "cwd-prefix-glob",
		"root", "root-file", "root-glob", "root-prefix-glob",
		"glob-i", "cwd-glob-i", "prefix-glob-i", "cwd-prefix-glob-i",
		"root-glob-i", "root-prefix-glob-i":
		return ActionFilesetPatterns().Suffix(":").NoSpace()
	default:
		return carapace.ActionValues()
	}
}
