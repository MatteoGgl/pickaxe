package frontmatter

const bom = "\xEF\xBB\xBF"

// Strip removes YAML frontmatter from markdown content.
// Frontmatter must start at byte 0 (after optional UTF-8 BOM) with "---\n" or "---\r\n".
// If no valid closing delimiter is found, content is returned unchanged.
func Strip(content string) string {
	// Strip optional BOM
	s := content
	if len(s) >= 3 && s[:3] == bom {
		s = s[3:]
	}

	// Must start with "---\n" or "---\r\n"
	if !hasPrefix(s, "---\n") && !hasPrefix(s, "---\r\n") {
		// Return with BOM stripped (even if no frontmatter)
		return s
	}

	// Find closing delimiter: \n---\n, \n---\r\n, or \n--- at EOF
	rest := s[3:] // skip opening "---"
	idx := findClose(rest)
	if idx < 0 {
		// No closing delimiter — return original (with BOM removed if applicable)
		// But spec says return unchanged on no-close, so restore original
		return content
	}

	return rest[idx:]
}

// findClose searches for the closing "---" line within s (which starts after the opening "---").
// s begins with "\n" or "\r\n". Returns the index of the character after the closing line,
// or -1 if not found.
func findClose(s string) int {
	for i := 0; i < len(s); {
		// Find next newline
		nl := indexByte(s, i, '\n')
		if nl < 0 {
			break
		}
		start := nl + 1
		// Check if "---" follows
		if start+3 > len(s) {
			break
		}
		if s[start:start+3] != "---" {
			i = start
			continue
		}
		after := start + 3
		if after == len(s) {
			// "---" at EOF
			return after
		}
		if s[after] == '\n' {
			return after + 1
		}
		if s[after] == '\r' && after+1 < len(s) && s[after+1] == '\n' {
			return after + 2
		}
		i = start
	}
	return -1
}

// Extract splits content into frontmatter (including delimiters) and body.
// When no valid frontmatter exists, fm is empty and body is the full content (BOM-stripped).
func Extract(content string) (fm string, body string) {
	s := content
	if len(s) >= 3 && s[:3] == bom {
		s = s[3:]
	}

	if !hasPrefix(s, "---\n") && !hasPrefix(s, "---\r\n") {
		return "", s
	}

	rest := s[3:] // skip opening "---"
	idx := findClose(rest)
	if idx < 0 {
		return "", content
	}

	return s[:3+idx], rest[idx:]
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func indexByte(s string, from int, b byte) int {
	for i := from; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
