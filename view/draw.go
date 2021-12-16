package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type drawCtx struct {
	screen tcell.Screen
	x      int
	y      int
	baseY  int
	width  int
	wrap   int
	indent int
}

func (d *drawCtx) drawY() int {
	return d.y - d.baseY
}

func (d *drawCtx) draw(text string, xOffset int, color tcell.Color) int {
	return d.drawWithEscape(text, xOffset, color, true)
}

func (d *drawCtx) drawWithEscape(text string, xOffset int, color tcell.Color, escape bool) int {
	if escape {
		text = tview.Escape(text)
	}
	_, actualWidth := tview.Print(d.screen, text, d.x+d.indent+xOffset, d.drawY(), d.width, tview.AlignLeft, color)
	return actualWidth
}

func (d *drawCtx) drawLine(text string, color tcell.Color) int {
	return d.drawLineWithEscape(text, color, true)
}

func (d *drawCtx) drawLineWithEscape(text string, color tcell.Color, escape bool) int {
	if escape {
		text = tview.Escape(text)
	}
	_, actualWidth := tview.Print(d.screen, text, d.x+d.indent, d.drawY(), d.width, tview.AlignLeft, color)
	d.y++
	return actualWidth
}

func (d *drawCtx) newLine() {
	d.y++
}

func (d *drawCtx) drawLines(texts []string, color tcell.Color) int {
	empty := true
	for i, text := range texts {
		if text == "" {
			continue
		}
		empty = false
		if i != 0 {
			d.y++
		}
		d.drawWrapped(text, color)
	}
	if empty {
		d.drawLine("<empty>", color)
		return 0
	}
	return 0
}

func (d *drawCtx) drawWrapped(text string, color tcell.Color) {
	if d.wrap == 0 {
		d.drawLine(text, color)
	}
	lines := wrapString(text, d.wrap-d.indent)
	for _, line := range lines {
		d.drawLine(line, color)
	}
}
