package jj

import "strings"

func parseLines(output []byte) []string {
	raw := strings.Split(string(output), "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func parseDescribedLines(output []byte) []string {
	lines := parseLines(output)
	vals := make([]string, 0, len(lines)*2)
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		name := parts[0]
		var desc string
		if len(parts) > 1 {
			desc = parts[1]
		}
		vals = append(vals, name, desc)
	}
	return vals
}

func parseBookmarkValues(output []byte, remoteOnly bool) []string {
	lines := parseLines(output)
	vals := make([]string, 0, len(lines)*2)
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		name := parts[0]
		hasAt := strings.Contains(name, "@")
		if remoteOnly && !hasAt {
			continue
		}
		if !remoteOnly && hasAt {
			continue
		}
		var desc string
		if len(parts) > 1 {
			desc = strings.TrimSpace(parts[1])
		}
		vals = append(vals, name, desc)
	}
	return vals
}

func stringsContain(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
