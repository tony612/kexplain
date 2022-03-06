package view

import (
	"regexp"
	"strings"
)

type line struct {
	wrap  int
	words []string
}

func (l *line) String() string {
	return strings.Join(l.words, " ")
}

func (l *line) Empty() bool {
	return len(l.words) == 0
}

func (l *line) Len() int {
	return len(l.String())
}

// Add adds the word to the line, returns true if we could, false if we
// didn't have enough room. It's always possible to add to an empty line.
func (l *line) Add(word string) bool {
	newLine := line{
		wrap:  l.wrap,
		words: append(l.words, word),
	}
	if newLine.Len() <= l.wrap || len(l.words) == 0 {
		l.words = newLine.words
		return true
	}
	return false
}

var bullet = regexp.MustCompile(`^(\d+\.?|-|\*)\s`)

const codePrefix = "    "

func wrapString(str string, wrap int) []string {
	wrapped := []string{}
	l := line{wrap: wrap}
	// track the last word added to the current line
	lastWord := ""
	flush := func() {
		if !l.Empty() {
			lastWord = ""
			wrapped = append(wrapped, l.String())
			l = line{wrap: wrap}
		}
	}

	// iterate over the lines in the original description
	for _, str := range strings.Split(str, "\n") {
		// preserve code blocks and blockquotes as-is
		if strings.HasPrefix(str, codePrefix) {
			flush()
			wrapped = append(wrapped, str)
			continue
		}

		// preserve empty lines after the first line, since they can separate logical sections
		if len(wrapped) > 0 && len(strings.TrimSpace(str)) == 0 {
			flush()
			wrapped = append(wrapped, "")
			continue
		}

		// flush if we should start a new line
		if shouldStartNewLine(lastWord, str) {
			flush()
		}
		words := strings.Fields(str)
		for _, word := range words {
			lastWord = word
			if !l.Add(word) {
				flush()
				if !l.Add(word) {
					panic("Couldn't add to empty line.")
				}
			}
		}
	}
	flush()
	return wrapped
}

func shouldStartNewLine(lastWord, str string) bool {
	// preserve line breaks ending in :
	if strings.HasSuffix(lastWord, ":") {
		return true
	}

	// preserve code blocks
	if strings.HasPrefix(str, codePrefix) {
		return true
	}
	str = strings.TrimSpace(str)
	// preserve empty lines
	if len(str) == 0 {
		return true
	}
	// preserve lines that look like they're starting lists
	if bullet.MatchString(str) == true {
		return true
	}
	// otherwise combine
	return false
}
