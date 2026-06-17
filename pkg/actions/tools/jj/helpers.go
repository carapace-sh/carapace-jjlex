package jj

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/carapace-sh/carapace-jjlex/pkg/revset"
)

var (
	rLocal    = regexp.MustCompile(`^(?P<bookmark>[^@: ]+): (?P<changeid>[^ ]+) (?P<commitid>[^ ]+) (?P<description>.*)$`)
	rRemote   = regexp.MustCompile(`^(?P<bookmark>[^@: ]+)@(?P<remote>[^: ]+): (?P<changeid>[^ ]+) (?P<commitid>[^ ]+) (?P<description>.*)$`)
	rTracking = regexp.MustCompile(`^  @(?P<remote>[^: ]+): (?P<changeid>[^ ]+) (?P<commitid>[^ ]+) (?P<description>.*)$`)
)

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

func parseTabSeparatedLines(output []byte) []string {
	lines := parseLines(output)
	vals := make([]string, 0, len(lines)*2)
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 2)
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
	simple, nonSimple := parseBookmarkValuesSplit(output, remoteOnly)
	return append(simple, nonSimple...)
}

// parseBookmarkValuesSplit splits bookmark values into simple identifiers
// (valid bare revset identifiers) and non-simple ones (needing quotes).
// Both lists are returned as [value, description, value, description, ...] pairs.
func parseBookmarkValuesSplit(output []byte, remoteOnly bool) (simple []string, nonSimple []string) {
	lines := strings.Split(string(output), "\n")
	simple = make([]string, 0)
	nonSimple = make([]string, 0)
	bookmark := ""
	for _, line := range lines[:len(lines)-1] {
		switch {
		case strings.HasPrefix(line, "  @"):
			if matches := rTracking.FindStringSubmatch(line); matches != nil {
				if remoteOnly {
					remote := stripDisplayQuotes(matches[1])
					val := fmt.Sprintf("%v@%v", bookmark, remote)
					classifyBookmark(val, matches[4], &simple, &nonSimple)
				}
			}
		default:
			if matches := rRemote.FindStringSubmatch(line); matches != nil {
				if remoteOnly {
					val := fmt.Sprintf("%v@%v", stripDisplayQuotes(matches[1]), stripDisplayQuotes(matches[2]))
					classifyBookmark(val, matches[5], &simple, &nonSimple)
				}
			} else if matches := rLocal.FindStringSubmatch(line); matches != nil {
				bookmark = stripDisplayQuotes(matches[1])
				if !remoteOnly {
					classifyBookmark(bookmark, matches[4], &simple, &nonSimple)
				}
			}
		}
	}
	return simple, nonSimple
}

// classifyBookmark appends the bookmark value to the appropriate list
// based on whether it's a simple revset identifier.
func classifyBookmark(name, description string, simple, nonSimple *[]string) {
	if revset.IsSimpleIdentifier(name) {
		*simple = append(*simple, name, description)
	} else {
		*nonSimple = append(*nonSimple, name, description)
	}
}

// stripDisplayQuotes removes surrounding double quotes that jj adds to
// bookmark names containing special characters (e.g. "parents(" → parents().
// jj's bookmark list output quotes identifiers that aren't simple revset
// identifiers, but the raw name is what we need for completion values.
func stripDisplayQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
