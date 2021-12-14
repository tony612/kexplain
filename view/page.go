package view

import (
	"container/list"
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

	pageData *pageData
	// Value is *pageData
	// stores old pageData like browser history
	// when going back to the parent field, pop the latest data
	pageDataHistory *list.List
}

// pageData is runtime calculated data related to position and fields
type pageData struct {
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

const highlight = "[black:green]"

func NewPage(doc *model.Doc) *Page {
	return &Page{
		Box:             tview.NewBox(),
		doc:             doc,
		pageData:        &pageData{},
		pageDataHistory: list.New(),
	}
}

func (p *Page) SetStopFn(fn func()) {
	p.stopFn = fn
}

func (p *Page) Draw(screen tcell.Screen) {
	p.Box.DrawForSubclass(screen, p)
	x, y, width, height := p.GetInnerRect()
	data := p.pageData
	data.windowHeight = height

	// Hit the bottom
	if data.height > 0 && data.currentY+height > data.height {
		data.currentY = data.height - height
	}
	// page is short, stick it at the top
	if height > data.height {
		data.currentY = 0
	}
	// Set selectedField to top position if Y changes
	if len(data.fieldsY) > 0 && len(data.fieldsY) > data.selectedField {
		if data.fieldsY[data.selectedField] < data.currentY {
			for i, y := range data.fieldsY {
				if y >= data.currentY {
					data.selectedField = i
					break
				}
			}
		}
		if data.fieldsY[data.selectedField] > data.currentY+data.windowHeight-1 {
			for i := len(data.fieldsY) - 1; i >= 0; i-- {
				if data.fieldsY[i] <= data.currentY+data.windowHeight-1 {
					data.selectedField = i
					break
				}
			}
		}
	}

	dc := drawCtx{
		screen: screen,
		x:      x,
		baseY:  data.currentY + y,
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
	data.height = dc.y - dc.baseY + data.currentY
}

func (p *Page) drawFields(dc *drawCtx) {
	kind := p.doc.GetDocKind()
	if kind == nil {
		return
	}
	data := p.pageData
	fieldsLen := len(kind.Keys())
	// selectedField selects the last one
	if data.selectedField >= fieldsLen {
		data.selectedField = fieldsLen - 1
	}
	if fieldsLen > 0 && len(data.fieldsY) == 0 {
		data.fieldsY = make([]int, fieldsLen)
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

		data.fieldsY[i] = dc.y
		if i == data.selectedField {
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
		data := p.pageData
		upFn := func(size int) {
			data.currentY -= size
			if data.currentY < 0 {
				data.currentY = 0
			}
		}
		downFn := func(size int) {
			data.currentY += size
			// Hitting the bottom is handled in Draw
		}
		switch event.Key() {
		case tcell.KeyUp, tcell.KeyCtrlP:
			upFn(1)
		case tcell.KeyDown, tcell.KeyCtrlN:
			downFn(1)
		case tcell.KeyPgUp, tcell.KeyCtrlB:
			upFn(data.windowHeight)
		case tcell.KeyPgDn, tcell.KeyCtrlF:
			downFn(data.windowHeight)
		case tcell.KeyTab:
			data.selectedField += 1
			if len(data.fieldsY) > data.selectedField {
				data.currentY = data.fieldsY[data.selectedField]
				if data.currentY < 0 {
					data.currentY = 0
				}
			}
		case tcell.KeyBacktab:
			data.selectedField -= 1
			if data.selectedField < 0 {
				data.selectedField = 0
			}
			if len(data.fieldsY) > data.selectedField {
				data.currentY = data.fieldsY[data.selectedField]
				if data.currentY < 0 {
					data.currentY = 0
				}
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'k':
				upFn(1)
			case 'j':
				downFn(1)
			case 'g':
				data.currentY = 0
			case 'G':
				data.currentY = data.height - data.windowHeight
				if data.currentY < 0 {
					data.currentY = 0
				}
			case 'q', 'Q':
				p.stopFn()
			case '[':
				if pressAlt(event) {
					newDoc := p.doc.FindParentDoc()
					if newDoc == nil {
						return
					}
					p.doc = newDoc
					if p.pageDataHistory.Len() == 0 {
						p.pageData = &pageData{}
					} else {
						p.pageData = p.pageDataHistory.Remove(p.pageDataHistory.Back()).(*pageData)
					}
				}
			}
		// Enter the sub field
		case tcell.KeyEnter:
			newDoc := p.doc.FindSubDoc(data.selectedField)
			if newDoc == nil {
				return
			}
			p.doc = newDoc
			p.pageDataHistory.PushBack(p.pageData)
			p.pageData = &pageData{}
		}
	})
}

func pressShift(e *tcell.EventKey) bool {
	return e.Modifiers()&tcell.ModShift != 0
}

func pressAlt(e *tcell.EventKey) bool {
	return e.Modifiers()&tcell.ModAlt != 0
}
