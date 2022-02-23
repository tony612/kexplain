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
}

func (d *drawCtx) drawY() int {
	// skip header
	return d.y - d.baseY + headerHeight
}

func (d *drawCtx) drawLineWithEscape(text string, color tcell.Color, escape bool) int {
	if escape {
		text = tview.Escape(text)
	}
	_, actualWidth := d.print(text, d.x, color)
	d.y++
	return actualWidth
}

func (d *drawCtx) print(text string, x int, color tcell.Color) (int, int) {
	if text == "" {
		return 0, 0
	}
	y := d.drawY()
	// skip header. tview.Print will check others
	if y < headerHeight {
		return 0, 0
	}
	return tview.Print(d.screen, text, x, y, d.width, tview.AlignLeft, color)
}

func (d *drawCtx) drawHorizontalLine(y int, color tcell.Color) {
	for x := d.x; x < d.x+d.width-1; x++ {
		d.screen.SetContent(x, y, tview.BoxDrawingsLightHorizontal, nil, tcell.StyleDefault.Foreground(color))
	}
}

func (d *drawCtx) overrideContent(s string, begin int, y int, style tcell.Style) {
	// skip header. tview.Print will check others
	if y < headerHeight {
		return
	}
	for i, r := range s {
		d.screen.SetContent(d.x+begin+i, y, r, nil, style)
	}

}
