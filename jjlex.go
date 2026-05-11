package jjlex

func Split(s string) CompletionContext {
	return NewCompletionAnalyzer(s).Analyze()
}
