package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	AppName   = "Kaizenity"
	Version   = "1.0.0"
	DBName    = "kaizenitydb.json"
	Hotkeys   = "[i] Add  [D] Remove  [h j k l] Select  [H J K L] Move  [q] Quit"
	ColorElem = tcell.ColorBlue
)

var (
	cards Cards
	app   = tview.NewApplication()
)

type Card struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Column int    `json:"column"`
	Pos    int64  `json:"pos"`
}

// Cards implements sort.Interface for []Card based on the Pos field
type Cards []Card

func (c Cards) Len() int           { return len(c) }
func (c Cards) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Cards) Less(i, j int) bool { return c[i].Pos < c[j].Pos }

func (c *Cards) ReadCards(path string) error {
	path += DBName
	jsonBlob, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonBlob, c)
	return err
}

func (c *Cards) DrawCards(column int, primitive tview.Primitive) {
	primitive.(*tview.List).Clear()
	for _, card := range *c {
		if card.Column == column {
			primitive.(*tview.List).AddItem(card.Name, card.Desc, 0, nil)
		}
	}
}

func eventInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'q':
		app.Stop()
	case 'j':
		idxCurrent := app.GetFocus().(*tview.List).GetCurrentItem()
		app.GetFocus().(*tview.List).SetCurrentItem(idxCurrent + 1)
	case 'k':
		idxCurrent := app.GetFocus().(*tview.List).GetCurrentItem()
		if idxCurrent > 0 {
			app.GetFocus().(*tview.List).SetCurrentItem(idxCurrent - 1)
		}
	}
	return event
}

func mainDraw(columnsStr []string, columnDefault int, path string) error {
	var numColumns = len(columnsStr)

	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetTextColor(ColorElem).
			SetText(text)
	}

	if err := cards.ReadCards(path); err != nil {
		return err
	}

	grid := tview.NewGrid().
		SetRows(1, -1, 1).
		SetBorders(true)

	headers := make([]tview.Primitive, 0)
	columns := make([]tview.Primitive, 0)
	for i := 0; i < numColumns; i++ {
		headers = append(headers, newPrimitive(columnsStr[i]))
		columns = append(columns, tview.NewList().SetSelectedFocusOnly(true))

		// Layout for screens narrower than 100 cells (only default column is visible)
		if i != columnDefault {
			grid.AddItem(headers[i], 0, 0, 0, 0, 0, 0, false).
				AddItem(columns[i], 0, 0, 0, 0, 0, 0, false)
		} else {
			grid.AddItem(headers[i], 0, 0, 1, numColumns, 0, 0, false).
				AddItem(columns[i], 1, 0, 1, numColumns, 0, 0, false)
		}

		// Layout for screens wider than 100 cells
		grid.AddItem(headers[i], 0, i, 1, 1, 0, 100, false).
			AddItem(columns[i], 1, i, 1, 1, 0, 100, false)
	}

	footer := newPrimitive("[ " + AppName + " " + Version + " ]  " + Hotkeys)
	grid.AddItem(footer, 2, 0, 1, numColumns, 0, 0, false)

	sort.Sort(Cards(cards))
	for i := 0; i < numColumns; i++ {
		cards.DrawCards(i, columns[i])
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return eventInput(event)
	})

	if err := app.SetRoot(grid, true).EnableMouse(true).SetFocus(columns[columnDefault]).Run(); err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {
	columnsStr := []string{"BACKLOG", "TODO", "DOING", "DONE"}
	columnDefault := 0
	pathInit := ""

	if err := mainDraw(columnsStr, columnDefault, pathInit); err != nil {
		fmt.Println(err)
	}
}
