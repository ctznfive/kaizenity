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
	Col  int    `json:"column"`
	Pos  int    `json:"position"`
	Name string `json:"name"`
	Desc string `json:"description"`
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

func (c *Cards) AddCard(form *tview.Form, columns []tview.Primitive, idColumn int, grid *tview.Grid) error {
	inputName := form.GetFormItem(0).(*tview.InputField).GetText()
	inputDesc := form.GetFormItem(1).(*tview.InputField).GetText()

	if inputName != "" {
		if len(*c) == 0 {
			*c = append(*c, Card{
				Col:  0,
				Pos:  0,
				Name: inputName,
				Desc: inputDesc,
			})
		} else {
			lastPos := columns[idColumn].(*tview.List).GetItemCount()
			*c = append(*c, Card{
				Col:  idColumn,
				Pos:  lastPos,
				Name: inputName,
				Desc: inputDesc,
			})
		}
		if err := c.WriteCards(); err != nil {
			return err
		}
		c.DrawCards(idColumn, columns[idColumn])
	}
	flagInput = false
	app.SetRoot(grid, true).EnableMouse(true).SetFocus(columns[idColumn])
	app.GetFocus().(*tview.List).SetCurrentItem(app.GetFocus().(*tview.List).GetItemCount() - 1)
	return nil
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
		if card.Col == column {
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
	focusCol := app.GetFocus()
	idFocusCol := indexOf(focusCol, columns)
	lenFocusCol := focusCol.(*tview.List).GetItemCount()
	posFocusCard := focusCol.(*tview.List).GetCurrentItem()

	switch event.Rune() {
	case 'q':
		app.Stop()

	case 'j':
		focusCol.(*tview.List).SetCurrentItem(posFocusCard + 1)

	case 'k':
		if posFocusCard > 0 {
			focusCol.(*tview.List).SetCurrentItem(posFocusCard - 1)
		}

	case 'l':
		if idFocusCol != -1 {
			for i := idFocusCol; i < len(columns)-1; i++ {
				lenNextList := columns[i+1].(*tview.List).GetItemCount()
				if lenNextList != 0 {
					app.SetFocus(columns[i+1])
					break
				}
			}
		}

	case 'h':
		if idFocusCol != -1 {
			for i := idFocusCol; i > 0; i-- {
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
			cards.AddCard(formNewCard, columns, idFocusCol, grid)
		})

		flagInput = true
		app.SetRoot(createModal(formNewCard, 70, 10), true).
			EnableMouse(true).
			SetFocus(formNewCard)

	case 'D':
		if lenFocusCol > 0 {
			i := 0
			for _, card := range cards {
				if card.Col == idFocusCol && card.Pos == posFocusCard {
					continue
				}
				cards[i] = card
				if card.Col == idFocusCol && card.Pos > posFocusCard {
					cards[i].Pos -= 1
				}
				i++
			}
			cards = cards[:i]

			if err := cards.WriteCards(); err != nil {
				fmt.Println(err)
			}
			sort.Sort(Cards(cards))
			for i := 0; i < len(columns); i++ {
				cards.DrawCards(i, columns[i])
			}
			if lenFocusCol == 1 {
				for i, col := range columns {
					lenCol := col.(*tview.List).GetItemCount()
					if lenCol != 0 {
						app.SetFocus(columns[i])
						break
					}
				}
			}
		}

	case 'K':
		if lenFocusCol > 1 && posFocusCard > 0 {
			idxCur := -1
			idxPrev := -1
			for i, card := range cards {
				if card.Col == idFocusCol {
					if card.Pos == posFocusCard-1 {
						idxPrev = i
					}
					if card.Pos == posFocusCard {
						idxCur = i
					}
				}
			}
			if idxCur != -1 && idxPrev != -1 {
				cards[idxCur].Pos, cards[idxPrev].Pos = cards[idxPrev].Pos, cards[idxCur].Pos
			}

			if err := cards.WriteCards(); err != nil {
				fmt.Println(err)
			}
			sort.Sort(Cards(cards))
			for i := 0; i < len(columns); i++ {
				cards.DrawCards(i, columns[i])
			}
			app.GetFocus().(*tview.List).SetCurrentItem(posFocusCard - 1)
		}

	case 'J':
		if lenFocusCol > 1 && posFocusCard < lenFocusCol-1 {
			idxCur := 0
			idxNext := 0
			for i, card := range cards {
				if card.Col == idFocusCol {
					if card.Pos == posFocusCard+1 {
						idxNext = i
					}
					if card.Pos == posFocusCard {
						idxCur = i
					}
				}
			}
			if idxCur != -1 && idxNext != -1 {
				cards[idxCur].Pos, cards[idxNext].Pos = cards[idxNext].Pos, cards[idxCur].Pos
			}

			if err := cards.WriteCards(); err != nil {
				fmt.Println(err)
			}
			sort.Sort(Cards(cards))
			for i := 0; i < len(columns); i++ {
				cards.DrawCards(i, columns[i])
			}
			app.GetFocus().(*tview.List).SetCurrentItem(posFocusCard + 1)
		}

	case 'L':
		if lenFocusCol > 0 && idFocusCol < len(columns)-1 {
			lenNextCol := columns[idFocusCol+1].(*tview.List).GetItemCount()
			for i, card := range cards {
				if card.Col == idFocusCol {
					if card.Pos == posFocusCard {
						cards[i].Col += 1
						cards[i].Pos = lenNextCol
					}
					if card.Pos > posFocusCard {
						cards[i].Pos -= 1
					}
				}
			}

			if err := cards.WriteCards(); err != nil {
				fmt.Println(err)
			}
			sort.Sort(Cards(cards))
			for i := 0; i < len(columns); i++ {
				cards.DrawCards(i, columns[i])
			}
			app.SetFocus(columns[idFocusCol+1])
			app.GetFocus().(*tview.List).SetCurrentItem(lenNextCol)
		}

	case 'H':
		if lenFocusCol > 0 && idFocusCol > 0 {
			lenPrevCol := columns[idFocusCol-1].(*tview.List).GetItemCount()
			for i, card := range cards {
				if card.Col == idFocusCol {
					if card.Pos == posFocusCard {
						cards[i].Col -= 1
						cards[i].Pos = lenPrevCol
						app.SetFocus(columns[idFocusCol-1])
					}
					if card.Pos > posFocusCard {
						cards[i].Pos -= 1
					}
				}
			}

			if err := cards.WriteCards(); err != nil {
				fmt.Println(err)
			}
			sort.Sort(Cards(cards))
			for i := 0; i < len(columns); i++ {
				cards.DrawCards(i, columns[i])
			}
			app.SetFocus(columns[idFocusCol-1])
			app.GetFocus().(*tview.List).SetCurrentItem(lenPrevCol)
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

	app.SetRoot(grid, true).EnableMouse(true)
	if columns[columnDefault].(*tview.List).GetItemCount() != 0 {
		app.SetFocus(columns[columnDefault])
	} else {
		for i, col := range columns {
			lenCol := col.(*tview.List).GetItemCount()
			if lenCol != 0 {
				app.SetFocus(columns[i])
				break
			}
		}
	}

	if err := app.Run(); err != nil {
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
