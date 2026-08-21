package processor

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	// Spacing alignment regex engines
	puncBeforeRegex = regexp.MustCompile(`\s+([.,!?:;])`)
	puncAfterRegex  = regexp.MustCompile(`([.,!?:;])([^\s.,!?:;])`)

	// Smart grammatical indefinite article adjustments
	articleRegex = regexp.MustCompile(`(?i)\b(a)\s+([aeiouh][a-zA-Z]*)`)
)

// ProcessText automatically cleans grammar, casing, numbers, and structures natively.
func ProcessText(input string) string {
	if len(strings.TrimSpace(input)) == 0 {
		return ""
	}

	// Step 1: Intelligently parse and convert numbers (Hex/Bin -> Decimal)
	processedText := convertNumericData(input)

	// Step 2: Separate clean string tokens
	words := strings.Fields(processedText)
	if len(words) == 0 {
		return ""
	}

	n := len(words)
	skip := make([]bool, n)

	// Step 3: Fast Lookahead Smart Quote Pair Adjustment
	quoteOpen := false
	for i := 0; i < n; i++ {
		if words[i] == "'" {
			if !quoteOpen {
				if next := findNextValid(skip, i, n); next != -1 {
					words[next] = "'" + words[next]
					skip[i] = true
					quoteOpen = true
				}
			} else {
				if prev := findPreviousValid(skip, i); prev != -1 {
					words[prev] = words[prev] + "'"
					skip[i] = true
					quoteOpen = false
				}
			}
		}
	}

	// Step 4: Reassemble words using string builder
	var builder strings.Builder
	first := true
	for i := 0; i < n; i++ {
		if skip[i] {
			continue
		}
		if !first {
			builder.WriteByte(' ')
		}
		builder.WriteString(words[i])
		first = false
	}
	result := builder.String()

	// Step 5: Clean spacing configurations before and after punctuation
	result = puncBeforeRegex.ReplaceAllString(result, "$1")
	result = puncAfterRegex.ReplaceAllString(result, "$1 $2")

	// Step 6: Apply smart contextual capitalization rules
	result = normalizeCasingAndSentences(result)

	// Step 7: Resolve structural indefinite vowels (a -> an)
	result = articleRegex.ReplaceAllStringFunc(result, func(match string) string {
		parts := strings.Fields(match)
		if len(parts) == 2 {
			article := parts[0]
			nextWord := parts[1]

			if strings.ToLower(article) == "a" {
				if unicode.IsUpper(rune(article[0])) {
					return "An " + nextWord
				}
				return "an " + nextWord
			}
		}
		return match
	})

	return result
}

// convertNumericData translates true numeric systems without damaging basic dictionary language strings.
func convertNumericData(input string) string {
	words := strings.Fields(input)
	for i, word := range words {
		// Strip attached formatting characters to safely isolate literal sequences
		cleanWord := strings.Trim(word, ".,!?:;\"'")
		if len(cleanWord) == 0 {
			continue
		}

		// 1. Isolate Binary strings containing strictly 0s and 1s (Skip lone 0 and 1)
		if isPureBinary(cleanWord) && len(cleanWord) > 1 {
			if val, err := strconv.ParseInt(cleanWord, 2, 64); err == nil {
				words[i] = strings.Replace(word, cleanWord, strconv.FormatInt(val, 10), 1)
				continue
			}
		}

		// 2. Isolate Hex strings (Explicit prefixes like 0x or data contains numbers mixed with valid hex alpha digits)
		if isHexadecimal(cleanWord) {
			hexVal := cleanWord
			if strings.HasPrefix(strings.ToLower(cleanWord), "0x") {
				hexVal = cleanWord[2:]
			}
			if val, err := strconv.ParseInt(hexVal, 16, 64); err == nil {
				words[i] = strings.Replace(word, cleanWord, strconv.FormatInt(val, 10), 1)
			}
		}
	}
	return strings.Join(words, " ")
}

// normalizeCasingAndSentences cleans random capital noise while fixing sentence starting capitals.
func normalizeCasingAndSentences(input string) string {
	if len(input) == 0 {
		return input
	}

	runes := []rune(input)
	capitalizeNext := true

	for i := 0; i < len(runes); i++ {
		if unicode.IsLetter(runes[i]) {
			if capitalizeNext {
				runes[i] = unicode.ToUpper(runes[i])
				capitalizeNext = false
			} else {
				// FIX: Forcefully downcases rogue inner capital letters here
				runes[i] = unicode.ToLower(runes[i])
			}
			continue
		}

		// Sentence boundaries detection layer
		if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' {
			// Skip ellipsis (...) sequences gracefully
			if i < len(runes)-1 && runes[i+1] == runes[i] {
				continue
			}
			capitalizeNext = true
		}
	}

	return string(runes)
}

func isPureBinary(s string) bool {
	for _, c := range s {
		if c != '0' && c != '1' {
			return false
		}
	}
	return true
}

func isHexadecimal(s string) bool {
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "0x") && len(s) > 2 {
		return true
	}

	hasNum := false
	hasAlpha := false

	for _, c := range lower {
		if c >= '0' && c <= '9' {
			hasNum = true
		} else if c >= 'a' && c <= 'f' {
			hasAlpha = true
		} else {
			return false // Contains invalid characters outside hexadecimal bounds
		}
	}

	// Avoid parsing normal words like "a" or "be" as hex unless explicitly mixed with digits
	return (hasNum && hasAlpha) || (hasNum && !hasAlpha && len(s) > 1 && s[0] != '0') || (hasAlpha && !hasNum && len(s) >= 2 && s != "be")
}

func findPreviousValid(skip []bool, currentIdx int) int {
	for i := currentIdx - 1; i >= 0; i-- {
		if !skip[i] {
			return i
		}
	}
	return -1
}

func findNextValid(skip []bool, currentIdx int, total int) int {
	for i := currentIdx + 1; i < total; i++ {
		if !skip[i] {
			return i
		}
	}
	return -1
}
