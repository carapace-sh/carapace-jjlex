package jjlex

func Split(s string) CompletionContext {
	return newCompletionAnalyzer(s).Analyze()
}
