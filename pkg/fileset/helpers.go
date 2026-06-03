package fileset

import "unicode"

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '\x0c'
}

// isFilesetIdentifierStart checks if a rune can start a fileset identifier.
// In fileset, identifier chars include XID_CONTINUE plus +, -, ., @, _, *, ?, [, ], /, \
func isFilesetIdentifierStart(r rune) bool {
	return isFilesetIdentifierPart(r)
}

// isFilesetIdentifierPart checks if a rune can be part of a fileset identifier.
// identifier = (XID_CONTINUE | "+" | "-" | "." | "@" | "_" | "*" | "?" | "[" | "]" | "/" | "\\")+
func isFilesetIdentifierPart(r rune) bool {
	switch r {
	case '+', '-', '.', '@', '_', '*', '?', '[', ']', '/', '\\':
		return true
	}
	return unicode.Is(unicode.Letter, r) || unicode.IsDigit(r) || unicode.IsMark(r)
}

// isStrictIdentifierPart checks if a rune is a valid part of a strict identifier.
// strict_identifier_part = (ASCII_ALPHANUMERIC | "_")+
func isStrictIdentifierPart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// isBareStringPart checks if a rune can be part of a bare_string.
// bare_string = ( ASCII_ALPHANUMERIC | " " | "+" | "-" | "." | "@" | "_" | "*" | "?" | "[" | "]" | "/" | "\\" | '\u{80}'..'\u{10ffff}' )+
func isBareStringPart(r rune) bool {
	if r >= '\u0080' {
		return true
	}
	switch r {
	case ' ', '+', '-', '.', '@', '_', '*', '?', '[', ']', '/', '\\':
		return true
	}
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func hexVal(r rune) int {
	switch {
	case r >= '0' && r <= '9':
		return int(r - '0')
	case r >= 'a' && r <= 'f':
		return int(r - 'a' + 10)
	case r >= 'A' && r <= 'F':
		return int(r - 'A' + 10)
	}
	return 0
}

func splitIdentifierParts(ident string, seps string) []string {
	var parts []string
	current := ""
	for _, ch := range ident {
		if containsRune(seps, ch) {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	parts = append(parts, current)
	return parts
}

func containsRune(s string, r rune) bool {
	for _, ch := range s {
		if ch == r {
			return true
		}
	}
	return false
}