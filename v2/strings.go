package main

import (
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// hasAnyPrefixWord checks if the given line is prefixed with any one of the given words
func hasAnyPrefixWord(line string, wordList []string) bool {
	for _, word := range wordList {
		if strings.HasPrefix(line, word+" ") {
			return true
		}
	}
	return false
}

// hasAnyPrefix checks if the given line is prefixed with any one of the given strings
func hasAnyPrefix(line string, stringList []string) bool {
	for _, s := range stringList {
		if strings.HasPrefix(line, s) {
			return true
		}
	}
	return false
}

// hasS checks if the given string slice contains the given string
func hasS(sl []string, s string) bool {
	for _, e := range sl {
		if e == s {
			return true
		}
	}
	return false
}

// firstWordContainsOneOf checks if the first word of the given string contains
// any one of the given strings
func firstWordContainsOneOf(s string, sl []string) bool {
	if s == "" {
		return false
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return false
	}
	firstWord := fields[0]
	for _, e := range sl {
		if strings.Contains(firstWord, e) {
			return true
		}
	}
	return false
}

// hasSuffix checks if the given string end with one of the given suffixes
func hasSuffix(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// hasKey checks if the given string map contains the given key
func hasKey(m map[string]string, key string) bool {
	_, found := m[key]
	return found
}

// filterS returns all strings that makes the function f return true
func filterS(sl []string, f func(string) bool) []string {
	var results []string
	for _, e := range sl {
		if f(e) {
			results = append(results, e)
		}
	}
	return results
}

// equalStringSlices checks if two given string slices are equal or not
// returns true if they are equal
func equalStringSlices(a, b []string) bool {
	lena := len(a)
	if lena != len(b) {
		return false
	}
	for i := 0; i < lena; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// hasWords checks if a range of more than one letter is found
func hasWords(s string) bool {
	letterCount := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			letterCount++
		} else {
			letterCount = 0
		}
		if letterCount > 1 {
			return true
		}
	}
	return false
}

// allUpper checks if all letters in a string are uppercase
func allUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// allLower checks if all letters in a string are lowercase
func allLower(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

// capitalizeWords can change "john bob" to "John Bob"
func capitalizeWords(s string) string {
	words := strings.Fields(s)
	var newWords []string
	for _, word := range words {
		if len(word) > 1 {
			capitalizedWord := strings.ToUpper(string(word[0])) + word[1:]
			newWords = append(newWords, capitalizedWord)
		} else {
			newWords = append(newWords, word)
		}
	}
	return strings.Join(newWords, " ")
}

// onlyAZaz checks if the given string only contains letters a-z and A-Z
func onlyAZaz(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

// smartSplit will split a string on spaces, but only spaces that are not within [], () or {}
func smartSplit(s string) []string {
	// Define constants for states.
	const (
		Outside = iota
		InParentheses
		InBrackets
		InBraces
	)

	state := Outside
	var result []string
	var word strings.Builder

	for _, ch := range s {
		switch ch {
		case '(':
			if state == Outside {
				state = InParentheses
			}
			word.WriteRune(ch)
		case ')':
			if state == InParentheses {
				state = Outside
			}
			word.WriteRune(ch)
		case '[':
			if state == Outside {
				state = InBrackets
			}
			word.WriteRune(ch)
		case ']':
			if state == InBrackets {
				state = Outside
			}
			word.WriteRune(ch)
		case '{':
			if state == Outside {
				state = InBraces
			}
			word.WriteRune(ch)
		case '}':
			if state == InBraces {
				state = Outside
			}
			word.WriteRune(ch)
		case ' ':
			if state == Outside {
				// Only split on space if outside of any brackets, braces, or parentheses.
				result = append(result, word.String())
				word.Reset()
			} else {
				word.WriteRune(ch)
			}
		default:
			word.WriteRune(ch)
		}
	}

	// Append the last word.
	if word.Len() > 0 {
		result = append(result, word.String())
	}

	return result
}

// isAllowedFilenameChar checks if the given rune is allowed in a typical cross-platform filename
func isAllowedFilenameChar(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) || isEmoji(r) {
		return true
	}
	switch r {
	case '.', ',', ':', '-', '+', '_', '/', '\\':
		return true
	default:
		return false
	}
}

// sanitizeFilename removes any character from the input string that is not part of a typical cross-platform filename
func sanitizeFilename(input string) string {
	var result []rune
	for _, r := range input {
		if isAllowedFilenameChar(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// getLeadingWhitespace returns the leading whitespace of the given string
func getLeadingWhitespace(line string) string {
	var whitespace []rune
	for _, char := range line {
		if unicode.IsSpace(char) {
			whitespace = append(whitespace, char)
		} else {
			break
		}
	}
	return string(whitespace)
}

func withinBackticks(line, what string) bool {
	first := []rune(what)[0]
	within := false
	lineRunes := []rune(line)
	whatRunes := []rune(what)

	for i, r := range lineRunes {
		if r == '`' { // `
			within = !within
			continue
		}
		if within && r == first {
			// check if the following runes also matches "what"
			// if they do, return true
			match := true
			for whatIndex, whatRune := range whatRunes {
				lineIndex := i + whatIndex
				if lineIndex >= len(lineRunes) {
					match = false
					break
				}
				if lineRunes[lineIndex] != whatRune {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}

// isEmoji checks if a rune is likely to be an emoji.
func isEmoji(r rune) bool {
	// Check if the rune falls within ranges that are likely to be used by emojis.
	return unicode.Is(unicode.S, r) || // Symbols
		unicode.Is(unicode.P, r) || // Punctuation
		r >= utf8.RuneSelf // Emojis are typically multi-byte characters in UTF-8
}

// trimRightSpace trims space but only at the right side of a string
func trimRightSpace(str string) string {
	return strings.TrimRightFunc(str, unicode.IsSpace)
}

// utf16ToUTF8 converts UTF-16 encoded data to UTF-8.
func utf16ToUTF8(data []byte) ([]byte, error) {
	// Check for byte order mark (BOM) to determine endianness.
	// This is a simplified approach; more robust handling may be required.
	bigEndian := false
	if len(data) >= 2 {
		bom := uint16(data[0])<<8 | uint16(data[1])
		if bom == 0xFEFF {
			bigEndian = true
			data = data[2:] // Remove BOM
		} else if bom == 0xFFFE {
			data = data[2:] // Remove BOM
		}
	}

	// Convert to uint16 slice.
	u16s := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		if bigEndian {
			u16s = append(u16s, uint16(data[i])<<8|uint16(data[i+1]))
		} else {
			u16s = append(u16s, uint16(data[i+1])<<8|uint16(data[i]))
		}
	}

	// Decode to UTF-8.
	runes := utf16.Decode(u16s)
	buf := make([]byte, 0, len(runes)*utf8.UTFMax)
	for _, r := range runes {
		buf = append(buf, string(r)...)
	}
	return buf, nil
}
