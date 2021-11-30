package view

import (
	"explainx/model"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"k8s.io/kubectl/pkg/explain"
)

type Page struct {
	*tview.Box
	doc    *model.Doc
	stopFn func()

	// runtime calculated

	// Y of first line
	currentY int
	// height of the whole page
	height int
	// height of the winddow
	windowHeight int
	// The index of the currently selected field
	selectedField int
	// Y of fields, index is field index
	fieldsY []int
}

const plainColor = tcell.ColorDefault
const fieldColor = tcell.ColorGreen

const kindPrefix = "KIND:     "
const versionPrefix = "VERSION:  "
const resourcePrefix = "RESOURCE: "
const descriptionLabel = "DESCRIPTION:"
const fieldsLabel = "FIELDS:"

const descIndent = 5
const fieldIndent = 3
const fieldDescIndent = 2

const defaultWrap = 80

const maxFieldWidth = 15

const highlight = "[green::u]"

func NewPage(doc *model.Doc) *Page {
	return &Page{
		Box: tview.NewBox(),
		doc: doc,
	}
}

func (p *Page) SetStopFn(fn func()) {
	p.stopFn = fn
}

func (p *Page) Draw(screen tcell.Screen) {
	p.Box.DrawForSubclass(screen, p)
	x, y, width, height := p.GetInnerRect()
	p.windowHeight = height

	// Hit the bottom
	if p.height > 0 && p.currentY+height > p.height {
		p.currentY = p.height - height
	}
	// Set selectedField to top position if Y changes
	if len(p.fieldsY) > 0 && len(p.fieldsY) > p.selectedField {
		if p.fieldsY[p.selectedField] < p.currentY {
			for i, y := range p.fieldsY {
				if y >= p.currentY {
					p.selectedField = i
					break
				}
			}
		}
		if p.fieldsY[p.selectedField] > p.currentY+p.windowHeight-1 {
			for i := len(p.fieldsY) - 1; i >= 0; i-- {
				if p.fieldsY[i] <= p.currentY+p.windowHeight-1 {
					p.selectedField = i
					break
				}
			}
		}
	}

	dc := drawCtx{
		screen: screen,
		x:      x,
		baseY:  p.currentY + y,
		y:      0,
		width:  width,
		wrap:   defaultWrap,
	}

	bottomLimit := y + height
	_, totalHeight := screen.Size()
	if bottomLimit > totalHeight {
		bottomLimit = totalHeight
	}

	//// Draw header

	// KIND
	dc.drawLine(kindPrefix+p.doc.GetKind(), plainColor)
	// VERSION
	dc.drawLine(versionPrefix+p.doc.GetVersion(), plainColor)
	dc.newLine()
	// RESOURCE
	resource := p.doc.GetFieldResource()
	if len(resource) > 0 {
		dc.drawLine(resourcePrefix+resource, plainColor)
	}
	dc.newLine()
	// DESCRIPTION
	dc.drawLine(descriptionLabel, plainColor)
	dc.indent += descIndent
	dc.drawLines(p.doc.GetDescriptions(), plainColor)
	dc.indent -= descIndent

	//// Draw fields

	// FIELDS
	dc.newLine()
	dc.drawLine(fieldsLabel, plainColor)
	p.drawFields(&dc)
	p.height = dc.y - dc.baseY + p.currentY
}

func (p *Page) drawFields(dc *drawCtx) {
	kind := p.doc.GetSubFieldKind()
	if kind == nil {
		return
	}
	fieldsLen := len(kind.Keys())
	// selectedField selects the last one
	if p.selectedField >= fieldsLen {
		p.selectedField = fieldsLen - 1
	}
	if fieldsLen > 0 && len(p.fieldsY) == 0 {
		p.fieldsY = make([]int, fieldsLen)
	}
	dc.indent += fieldIndent
	defer func() {
		dc.indent -= fieldIndent
	}()
	for i, key := range kind.Keys() {
		v := kind.Fields[key]
		required := ""
		if kind.IsRequired(key) {
			required = " -required-"
		}

		spaceLen := maxFieldWidth - len(key)
		if spaceLen <= 0 {
			spaceLen = 3
		}

		p.fieldsY[i] = dc.y
		if i == p.selectedField {
			dc.draw(highlight+key, 0, fieldColor)
		} else {
			dc.draw(key, 0, fieldColor)
		}
		dc.draw(fmt.Sprintf("%s<%s>%s", strings.Repeat(" ", spaceLen),
			explain.GetTypeName(v), required), len(key), plainColor)
		dc.y++

		dc.indent += fieldDescIndent
		dc.drawWrapped(v.GetDescription(), plainColor)
		dc.indent -= fieldDescIndent
		dc.newLine()
	}
}

func (p *Page) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		upFn := func(size int) {
			p.currentY -= size
			if p.currentY < 0 {
				p.currentY = 0
			}
		}
		downFn := func(size int) {
			p.currentY += size
			// Hitting the bottom is handled in Draw
		}
		switch event.Key() {
		case tcell.KeyUp, tcell.KeyCtrlP:
			upFn(1)
		case tcell.KeyDown, tcell.KeyCtrlN:
			downFn(1)
		case tcell.KeyPgUp, tcell.KeyCtrlB:
			upFn(p.windowHeight)
		case tcell.KeyPgDn, tcell.KeyCtrlF:
			downFn(p.windowHeight)
		case tcell.KeyTab:
			p.selectedField += 1
			if len(p.fieldsY) > p.selectedField {
				p.currentY = p.fieldsY[p.selectedField]
				if p.currentY < 0 {
					p.currentY = 0
				}
			}
		case tcell.KeyBacktab:
			p.selectedField -= 1
			if p.selectedField < 0 {
				p.selectedField = 0
			}
			if len(p.fieldsY) > p.selectedField {
				p.currentY = p.fieldsY[p.selectedField]
				if p.currentY < 0 {
					p.currentY = 0
				}
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'k':
				upFn(1)
			case 'j':
				downFn(1)
			case 'g':
				p.currentY = 0
			case 'G':
				p.currentY = p.height - p.windowHeight
				if p.currentY < 0 {
					p.currentY = 0
				}
			case 'q', 'Q':
				p.stopFn()
			}
		}
	})

}

func pressShift(e *tcell.EventKey) bool {
	return e.Modifiers()&tcell.ModShift != 0
}
