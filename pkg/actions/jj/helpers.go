package jj

import (
	"fmt"
	"regexp"
	"strings"
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

func parseBookmarkValues(output []byte, remoteOnly bool) []string {
	lines := strings.Split(string(output), "\n")
	vals := make([]string, 0)
	bookmark := ""
	for _, line := range lines[:len(lines)-1] {
		switch {
		case strings.HasPrefix(line, "  @"):
			if matches := rTracking.FindStringSubmatch(line); matches != nil {
				if remoteOnly {
					vals = append(vals, fmt.Sprintf("%v@%v", bookmark, matches[1]), matches[4])
				}
			}
		default:
			if matches := rRemote.FindStringSubmatch(line); matches != nil {
				if remoteOnly {
					vals = append(vals, fmt.Sprintf("%v@%v", matches[1], matches[2]), matches[5])
				}
			} else if matches := rLocal.FindStringSubmatch(line); matches != nil {
				bookmark = matches[1]
				if !remoteOnly {
					vals = append(vals, matches[1], matches[4])
				}
			}
		}
	}
	return vals
}
