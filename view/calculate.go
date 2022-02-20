package view

import (
	"strings"

	"github.com/rivo/tview"
)

type linesCalculator struct {
	y      int
	indent int
	wrap   int
	lines  []string
}

const defaultWrap = 80

func newLinesCalculator() *linesCalculator {
	return &linesCalculator{
		y:      0,
		indent: 0,
		wrap:   defaultWrap,
		lines:  []string{},
	}
}

func (c *linesCalculator) appendLine(line string) {
	c.appendLineWithEscape(line, true)
}

func (c *linesCalculator) appendLineWithEscape(line string, escape bool) {
	if c.indent > 0 {
		line = strings.Repeat(" ", c.indent) + line
	}
	if escape {
		line = tview.Escape(line)
	}
	c.lines = append(c.lines, line)
	c.y++
}

func (c *linesCalculator) appendWrapped(text string) {
	if c.wrap == 0 {
		c.appendLine(text)
		return
	}
	lines := wrapString(text, c.wrap-c.indent)
	for _, line := range lines {
		c.appendLine(line)
	}
}

func (c *linesCalculator) appendLines(lines []string) {
	empty := true
	for i, text := range lines {
		if text == "" {
			continue
		}
		empty = false
		if i != 0 {
			c.appendLine("")
		}
		c.appendWrapped(text)
	}
	if empty {
		c.appendLine("<empty>")
	}
}
