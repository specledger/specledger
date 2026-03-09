package spec

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const MaxBranchLength = 244

var StopWords = map[string]bool{
	"i": true, "a": true, "an": true, "the": true,
	"to": true, "for": true, "of": true, "in": true,
	"on": true, "at": true, "by": true, "with": true,
	"from": true, "is": true, "are": true, "was": true,
	"were": true, "be": true, "been": true, "being": true,
	"have": true, "has": true, "had": true, "do": true,
	"does": true, "did": true, "will": true, "would": true,
	"should": true, "could": true, "can": true, "may": true,
	"might": true, "must": true, "shall": true, "this": true,
	"that": true, "these": true, "those": true, "my": true,
	"your": true, "our": true, "their": true, "want": true,
	"need": true, "add": true, "get": true, "set": true,
}

var acronymPattern = regexp.MustCompile(`^[A-Z]{2,}[0-9]*$`)

func GenerateBranchName(description string, number int) string {
	words := tokenizeAndClean(description)

	words = FilterStopWords(words)

	for i, word := range words {
		original := getOriginalWord(description, word)
		words[i] = PreserveAcronyms(original, word)
	}

	if len(words) > 4 {
		words = words[:4]
	}

	shortName := strings.Join(words, "-")
	branchName := fmt.Sprintf("%03d-%s", number, shortName)

	return TruncateToLimit(branchName, MaxBranchLength)
}

func tokenizeAndClean(description string) []string {
	var words []string
	var currentWord strings.Builder

	for _, ch := range description {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			currentWord.WriteRune(unicode.ToLower(ch))
		} else if currentWord.Len() > 0 {
			word := currentWord.String()
			if len(word) > 0 {
				words = append(words, word)
			}
			currentWord.Reset()
		}
	}

	if currentWord.Len() > 0 {
		word := currentWord.String()
		if len(word) > 0 {
			words = append(words, word)
		}
	}

	return words
}

func FilterStopWords(words []string) []string {
	var filtered []string
	for _, word := range words {
		if !StopWords[word] {
			filtered = append(filtered, word)
		}
	}
	return filtered
}

func PreserveAcronyms(original, word string) string {
	if acronymPattern.MatchString(original) {
		return strings.ToLower(original)
	}

	if allCaps(original) && len(original) >= 2 {
		return strings.ToLower(original)
	}

	return word
}

func getOriginalWord(description, lowercaseWord string) string {
	words := regexp.MustCompile(`[A-Za-z0-9]+`).FindAllString(description, -1)
	for _, w := range words {
		if strings.ToLower(w) == lowercaseWord {
			return w
		}
	}
	return lowercaseWord
}

func allCaps(s string) bool {
	for _, ch := range s {
		if unicode.IsLetter(ch) && !unicode.IsUpper(ch) {
			return false
		}
	}
	return len(s) > 0
}

func TruncateToLimit(name string, maxBytes int) string {
	if len(name) <= maxBytes {
		return name
	}

	parts := strings.SplitN(name, "-", 2)
	if len(parts) != 2 {
		if len(name) > maxBytes {
			return name[:maxBytes]
		}
		return name
	}

	prefix := parts[0] + "-"
	maxSuffix := maxBytes - len(prefix)

	if maxSuffix <= 0 {
		return prefix[:maxBytes]
	}

	suffix := parts[1]
	if len(suffix) > maxSuffix {
		suffix = suffix[:maxSuffix]
	}

	suffix = strings.TrimRight(suffix, "-")

	result := prefix + suffix

	if len(result) > maxBytes {
		result = result[:maxBytes]
	}

	return result
}
