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
	// skip header
	return d.y - d.baseY + headerHeight
}

func (d *drawCtx) draw(text string, xOffset int, color tcell.Color) int {
	return d.drawWithEscape(text, xOffset, color, true)
}

func (d *drawCtx) drawWithEscape(text string, xOffset int, color tcell.Color, escape bool) int {
	if escape {
		text = tview.Escape(text)
	}
	_, actualWidth := d.print(text, d.x+d.indent+xOffset, color)
	return actualWidth
}

func (d *drawCtx) drawLine(text string, color tcell.Color) int {
	return d.drawLineWithEscape(text, color, true)
}

func (d *drawCtx) drawLineWithEscape(text string, color tcell.Color, escape bool) int {
	if escape {
		text = tview.Escape(text)
	}
	_, actualWidth := d.print(text, d.x+d.indent, color)
	d.y++
	return actualWidth
}

func (d *drawCtx) print(text string, x int, color tcell.Color) (int, int) {
	y := d.drawY()
	// skip header. tview.Print will check others
	if y < headerHeight {
		return 0, 0
	}
	return tview.Print(d.screen, text, x, y, d.width, tview.AlignLeft, color)
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

func (d *drawCtx) drawHorizontalLine(y int, color tcell.Color) {
	for x := d.x; x < d.x+d.width-1; x++ {
		d.screen.SetContent(x, y, tview.BoxDrawingsLightHorizontal, nil, tcell.StyleDefault.Foreground(color))
	}
}
