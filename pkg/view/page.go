package view

import (
	"container/list"
	"fmt"
	"kexplain/pkg/model"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"k8s.io/kubectl/pkg/explain"
)

// Page is the view component of the page in kexplain,
// which is based on tview.
type Page struct {
	*tview.Box
	doc    *model.Doc
	stopFn func()

	staticData *pageStaticData
	pageData   *pageData
	// Value is *pageData
	// stores old pageData like browser history
	// when going back to the parent field, pop the latest data
	pageDataHistory *list.List

	// Command
	commandBar    *tview.InputField
	typingCommand bool
	command       string

	// searching
	searchText string
	searching  searchDirection
}

type searchDirection = int8

const (
	searchStop searchDirection = 0
	searchNext searchDirection = 1
	searchBack searchDirection = 2
)

const headerHeight = 1
const bottomHeight = 1

// pageStaticData is fixed data for a page, which is calculated only once for a page
type pageStaticData struct {
	windowHeight int
	// Y of fields, index is field index
	fieldsY []int
	// lines slices
	lines []string
}

func (d *pageStaticData) height() int {
	return len(d.lines)
}

// pageData is runtime calculated data related to position and fields
type pageData struct {
	// Y of first line
	currentY int
	// The index of the currently selected field
	selectedField int
}

const plainColor = tcell.ColorDefault

const kindPrefix = "KIND:     "
const versionPrefix = "VERSION:  "
const resourcePrefix = "RESOURCE: "
const descriptionLabel = "DESCRIPTION:"
const fieldsLabel = "FIELDS:"

const descIndent = 5
const fieldIndent = 3
const fieldDescIndent = 2

const maxFieldWidth = 15

var highlightStyle = tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack)
var fieldStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
var searchStyle = tcell.StyleDefault.Background(tcell.ColorLightYellow).Foreground(tcell.ColorBlack)

// NewPage returns a Page
func NewPage(doc *model.Doc) *Page {
	page := &Page{
		Box:             tview.NewBox().SetBackgroundColor(plainColor),
		doc:             doc,
		pageData:        &pageData{},
		pageDataHistory: list.New(),
		command:         ":",
		staticData:      &pageStaticData{},
	}
	commandBar := tview.NewInputField().
		SetLabel("").
		SetPlaceholder("").
		SetFieldWidth(0).
		SetFieldBackgroundColor(plainColor).
		SetDoneFunc(page.handleCommand)
	page.commandBar = commandBar
	page.calLines()
	return page
}

// SetStopFn sets the stop callback, which is called when pressing q/Q.
func (p *Page) SetStopFn(fn func()) {
	p.stopFn = fn
}

// Draw draws the view
func (p *Page) Draw(screen tcell.Screen) {
	p.Box.DrawForSubclass(screen, p)
	x, y, width, height := p.GetInnerRect()
	data := p.pageData
	p.staticData.windowHeight = height - headerHeight - bottomHeight

	pageHeight := p.staticData.height()
	fieldsY := p.staticData.fieldsY

	// Hit the bottom
	if pageHeight > 0 && data.currentY+p.staticData.windowHeight > pageHeight {
		data.currentY = pageHeight - p.staticData.windowHeight
	}
	// page is shorter than the screen, stick it at the top
	if height > pageHeight {
		data.currentY = 0
	}
	// Set selectedField to top position if Y changes
	if len(fieldsY) > 0 && len(fieldsY) > data.selectedField {
		// When selected field is above the whole page, set to the first within the page
		if fieldsY[data.selectedField] < data.currentY {
			for i, y := range fieldsY {
				if y >= data.currentY {
					data.selectedField = i
					break
				}
			}
		}
		// When selected field is below the whole page, set to the last within the page
		if fieldsY[data.selectedField] > data.currentY+p.staticData.windowHeight {
			for i := len(fieldsY) - 1; i >= 0; i-- {
				if fieldsY[i] <= data.currentY+p.staticData.windowHeight {
					data.selectedField = i
					break
				}
			}
		}
	}
	// selectedField selects the last one
	if data.selectedField >= len(p.staticData.fieldsY) {
		data.selectedField = len(p.staticData.fieldsY) - 1
	}

	var searchRe *regexp.Regexp
	if p.searchText != "" {
		searchRe, _ = regexp.Compile(p.searchText)
		var beginI, endI, delta int
		if p.searching == searchNext {
			beginI, endI, delta = data.currentY+1, len(p.staticData.lines), 1
		} else if p.searching == searchBack {
			beginI, endI, delta = data.currentY-1, -1, -1
		}
		for i := beginI; i != endI; i += delta {
			if searchRe != nil && searchRe.MatchString(p.staticData.lines[i]) {
				data.currentY = i
				break
			}
		}
	}
	// reset researching, so that up/down moving can work normally
	p.searching = searchStop

	dc := drawCtx{
		screen: screen,
		x:      x,
		baseY:  data.currentY + y,
		y:      0,
		width:  width,
	}

	//// Draw header
	dc.drawHorizontalLine(0, plainColor)
	tview.Print(screen, " "+p.doc.GetFullPath()+" ", 0, 0, dc.width, tview.AlignCenter, plainColor)

	fieldIdx := 0
	for i, l := range p.staticData.lines {
		drawY := dc.drawY()
		dc.drawLineWithEscape(l, plainColor, false)
		if i == fieldsY[data.selectedField] {
			field, begin := findFirstField(l)
			dc.overrideContent(field, begin, drawY, highlightStyle)
			fieldIdx++
		} else if fieldIdx < len(fieldsY) && i == fieldsY[fieldIdx] {
			field, begin := findFirstField(l)
			dc.overrideContent(field, begin, drawY, fieldStyle)
			fieldIdx++
		}
		if p.searchText != "" && searchRe != nil {
			found := searchRe.FindAllStringIndex(l, -1)
			for _, pair := range found {
				dc.overrideContent(l[pair[0]:pair[1]], pair[0], drawY, searchStyle)
			}
		}
	}

	//// Draw the command bar
	p.commandBar.SetRect(x+1, height-1, width, 1)
	p.commandBar.Draw(screen)
	tview.Print(dc.screen, p.command, x, height-1, 1, tview.AlignLeft, plainColor)
	if !p.typingCommand {
		screen.ShowCursor(x+1, height-1)
	}
}

// Focus is override of Box
func (p *Page) Focus(delegate func(p tview.Primitive)) {
	if p.typingCommand {
		delegate(p.commandBar)
	} else {
		p.Box.Focus(delegate)
	}
}

// HasFocus is override of Box
func (p *Page) HasFocus() bool {
	if p.typingCommand {
		return p.commandBar.HasFocus()
	}
	return p.Box.HasFocus()
}

func (p *Page) calLines() {
	c := newLinesCalculator()
	// KIND
	c.appendLine(kindPrefix + p.doc.GetKind())
	// VERSION
	c.appendLine(versionPrefix + p.doc.GetVersion())
	c.appendLine("")
	// RESOURCE
	resource := p.doc.GetFieldResource()
	if len(resource) > 0 {
		c.appendLine(resourcePrefix + resource)
		c.appendLine("")
	}
	// DESCRIPTION
	c.appendLine(descriptionLabel)
	c.indent += descIndent
	c.appendLines(p.doc.GetDescriptions())
	c.indent -= descIndent

	//// Draw fields
	c.appendLine("")
	c.appendLine(fieldsLabel)
	p.calFields(c)
	p.staticData.lines = c.lines
}

func (p *Page) calFields(c *linesCalculator) {
	kind := p.doc.GetDocKind()
	if kind == nil {
		return
	}
	data := p.staticData
	fieldsLen := len(kind.Keys())
	data.fieldsY = make([]int, fieldsLen)
	c.indent += fieldIndent
	defer func() {
		c.indent -= fieldIndent
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

		data.fieldsY[i] = c.y
		fieldLine := key + fmt.Sprintf("%s<%s>%s", strings.Repeat(" ", spaceLen), explain.GetTypeName(v), required)
		c.appendLine(fieldLine)

		c.indent += fieldDescIndent
		c.appendWrapped(v.GetDescription())
		c.indent -= fieldDescIndent
		c.appendLine("")
	}
}

// InputHandler is override of Box, which handles keyboard inputs.
func (p *Page) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		defer func() {
			if !p.typingCommand {
				setFocus(p)
			}
		}()
		if p.typingCommand {
			// Pass event on to child primitive.
			if p.commandBar != nil && p.commandBar.HasFocus() {
				currText := p.commandBar.GetText()
				if handler := p.commandBar.InputHandler(); handler != nil {
					handler(event, setFocus)
				}
				// Exit inputting when backspace and current text is empty like what `less` does
				if currText == "" && (event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2) {
					p.doneCommandTyping()
				}
				return
			}
		}
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
			upFn(p.staticData.windowHeight)
		case tcell.KeyPgDn, tcell.KeyCtrlF:
			downFn(p.staticData.windowHeight)
		case tcell.KeyTab:
			data.selectedField++
			if len(p.staticData.fieldsY) > data.selectedField {
				data.currentY = p.staticData.fieldsY[data.selectedField]
				if data.currentY < 0 {
					data.currentY = 0
				}
			}
		case tcell.KeyBacktab:
			data.selectedField--
			if data.selectedField < 0 {
				data.selectedField = 0
			}
			if len(p.staticData.fieldsY) > data.selectedField {
				data.currentY = p.staticData.fieldsY[data.selectedField]
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
				data.currentY = p.staticData.height() - p.staticData.windowHeight
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
						p.resetData()
					} else {
						p.pageData = p.pageDataHistory.Remove(p.pageDataHistory.Back()).(*pageData)
						p.calLines()
					}
				}
			case '/':
				p.typingCommand = true
				p.command = "/"
				setFocus(p.commandBar)
			case 'n':
				if p.searchText == "" {
					return
				}
				p.searching = searchNext
			case 'N':
				if p.searchText == "" {
					return
				}
				p.searching = searchBack
			}
		// Enter the sub field
		case tcell.KeyEnter:
			newDoc := p.doc.FindSubDoc(data.selectedField)
			if newDoc == nil {
				return
			}
			p.doc = newDoc
			p.pageDataHistory.PushBack(p.pageData)
			p.resetData()
		}
	})
}

func (p *Page) resetData() {
	p.pageData = &pageData{}
	p.calLines()
}

func (p *Page) handleCommand(key tcell.Key) {
	switch key {
	// inputfield component also use KeyTab and KeyBacktab for done
	// we only handle Enter and Escape
	case tcell.KeyEnter, tcell.KeyEscape:
		// Only Enter means confirm the input
		if key == tcell.KeyEnter {
			p.searchText = p.commandBar.GetText()
			if p.searchText != "" {
				p.searching = searchNext
			} else {
				p.searching = searchStop
			}
		}

		p.doneCommandTyping()
	}
}

func (p *Page) doneCommandTyping() {
	p.typingCommand = false
	p.command = ":"
	p.commandBar.SetText("")
}

func pressShift(e *tcell.EventKey) bool {
	return e.Modifiers()&tcell.ModShift != 0
}

func pressAlt(e *tcell.EventKey) bool {
	return e.Modifiers()&tcell.ModAlt != 0
}

func findFirstField(s string) (string, int) {
	left := -1
	right := -1
	for i := 0; i < len(s); i++ {
		if left > 0 {
			if s[i] == ' ' {
				right = i
				break
			}
		} else {
			if s[i] != ' ' {
				left = i
			}
		}
	}
	if left > 0 && right > 0 {
		return s[left:right], left
	}
	return "", left
}
