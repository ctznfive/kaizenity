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
	app       = tview.NewApplication()
	cards     Cards
	pathInit  string
	flagInput bool
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

func indexOf(element tview.Primitive, columns []tview.Primitive) int {
	for k, v := range columns {
		if element == v {
			return k
		}
	}
	return -1
}

func createModal(primitive tview.Primitive, x, y int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, x, 0).
		SetRows(0, y, 0).
		AddItem(primitive, 1, 1, 1, 1, 0, 0, true)
}

func addCard(form *tview.Form, columns []tview.Primitive, idColumn int, grid *tview.Grid) {
	inputName := form.GetFormItem(0).(*tview.InputField).GetText()
	inputDesc := form.GetFormItem(1).(*tview.InputField).GetText()

	if inputName != "" {
		if len(cards) == 0 {
			cards = append(cards, Card{
				ID:     0,
				Name:   inputName,
				Desc:   inputDesc,
				Column: idColumn,
				Pos:    0,
			})
		} else {
			cards = append(cards, Card{
				ID:     cards[len(cards)-1].ID + 1,
				Name:   inputName,
				Desc:   inputDesc,
				Column: idColumn,
				Pos:    cards[len(cards)-1].ID + 1,
			})
		}
		if err := cards.WriteCards(); err != nil {
			fmt.Println(err)
		}
		cards.DrawCards(idColumn, columns[idColumn])
	}
	flagInput = false
	app.SetRoot(grid, true).EnableMouse(true).SetFocus(columns[idColumn])
}

func (c *Cards) ReadCards() error {
	path := pathInit + DBName
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.WriteCards()
	} else {
		jsonBlob, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		err = json.Unmarshal(jsonBlob, c)
		return nil
	}
}

func (c *Cards) DrawCards(column int, primitive tview.Primitive) {
	primitive.(*tview.List).Clear()
	for _, card := range *c {
		if card.Column == column {
			primitive.(*tview.List).AddItem(card.Name, card.Desc, 0, nil)
		}
	}
}

func (c *Cards) WriteCards() error {
	path := pathInit + DBName
	jsonBlob, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, jsonBlob, 0644)
	return nil
}

func eventInput(event *tcell.EventKey, columns []tview.Primitive, grid *tview.Grid) *tcell.EventKey {
	if flagInput {
		return event
	}
	focus := app.GetFocus()
	idFocus := indexOf(focus, columns)
	idCurrent := focus.(*tview.List).GetCurrentItem()

	switch event.Rune() {
	case 'q':
		app.Stop()

	case 'j':
		focus.(*tview.List).SetCurrentItem(idCurrent + 1)

	case 'k':
		if idCurrent > 0 {
			focus.(*tview.List).SetCurrentItem(idCurrent - 1)
		}

	case 'l':
		if idFocus != -1 {
			for i := idFocus; i < len(columns)-1; i++ {
				lenNextList := columns[i+1].(*tview.List).GetItemCount()
				if lenNextList != 0 {
					app.SetFocus(columns[i+1])
					break
				}
			}
		}

	case 'h':
		if idFocus != -1 {
			for i := idFocus; i > 0; i-- {
				lenPrevList := columns[i-1].(*tview.List).GetItemCount()
				if lenPrevList != 0 {
					app.SetFocus(columns[i-1])
					break
				}
			}
		}

	case 'i':
		formNewCard := tview.NewForm().
			AddInputField("Name: ", "", 70, nil, nil).
			AddInputField("Description: ", "", 70, nil, nil)

		formNewCard.SetButtonsAlign(tview.AlignCenter).
			SetBorder(true).
			SetTitle("Add new card").
			SetTitleAlign(tview.AlignCenter)

		formNewCard.AddButton("Create", func() {
			addCard(formNewCard, columns, idFocus, grid)
		})

		flagInput = true
		app.SetRoot(createModal(formNewCard, 70, 10), true).
			EnableMouse(true).
			SetFocus(formNewCard)

	case 'D':
		i := 0
		idx := cards[idCurrent].ID
		for _, card := range cards {
			if card.ID != idx {
				cards[i] = card
				i++
			}
		}
		cards = cards[:i]
		focus.(*tview.List).RemoveItem(idCurrent)

		if err := cards.WriteCards(); err != nil {
			fmt.Println(err)
		}
		sort.Sort(Cards(cards))
		for i := 0; i < len(columns); i++ {
			cards.DrawCards(i, columns[i])
		}
	}
	return event
}

func mainDraw(columnsStr []string, columnDefault int) error {
	var numColumns = len(columnsStr)

	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetTextColor(ColorElem).
			SetText(text)
	}

	if err := cards.ReadCards(); err != nil {
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
		grid.AddItem(headers[i], 0, i, 1, 1, 0, 0, false).
			AddItem(columns[i], 1, i, 1, 1, 0, 0, false)
	}

	footer := newPrimitive("[ " + AppName + " " + Version + " ]  " + Hotkeys)
	grid.AddItem(footer, 2, 0, 1, numColumns, 0, 0, false)

	sort.Sort(Cards(cards))
	for i := 0; i < numColumns; i++ {
		cards.DrawCards(i, columns[i])
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return eventInput(event, columns, grid)
	})

	if err := app.SetRoot(grid, true).EnableMouse(true).SetFocus(columns[columnDefault]).Run(); err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {
	columnsStr := []string{"BACKLOG", "TODO", "DOING", "DONE"}
	columnDefault := 0
	pathInit = ""
	flagInput = false

	if err := mainDraw(columnsStr, columnDefault); err != nil {
		fmt.Println(err)
	}
}
