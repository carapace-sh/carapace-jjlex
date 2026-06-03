package template

func dedupTokens(tokens []ExpectedToken) []ExpectedToken {
	seen := make(map[ExpectedToken]bool)
	var result []ExpectedToken
	for _, t := range tokens {
		if !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}
	return result
}

func dedupOperators(ops []ValidOperator) []ValidOperator {
	seen := make(map[string]bool)
	var result []ValidOperator
	for _, op := range ops {
		if !seen[op.Op] {
			seen[op.Op] = true
			result = append(result, op)
		}
	}
	return result
}