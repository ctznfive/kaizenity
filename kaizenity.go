/*****  K  A  I  Z  E  N  I  T  Y  *****/
/*** See LICENSE for license details ***/

package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"

	// Terminal UI libraries
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	AppName   = "Kaizenity"
	Version   = "1.0.0"
	DBName    = "kaizenitydb.json"
	Hotkeys   = "[a] Add  [i] Edit  [D] Remove  [h j k l] Select  [H J K L] Move  [Q] Quit"
	ColorElem = tcell.ColorBlue
)

var (
	// New terminal based application
	app = tview.NewApplication()

	// The path where the card database is stored (JSON file)
	// In the home directory of a user (= "home")
	// Or in the current program directory (!= "home")
	pathInit = ""

	// The number and names of the board columns
	columnsStr = []string{"BACKLOG", "TODO", "DOING", "DONE"}

	// Default column index
	columnDefault = 0

	// Activates hotkeys (= false) or keyboard text input (= true)
	flagInput = false

	// Data structure that stores all cards from JSON file
	cards Cards
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

// Returns the index of a column in a column slice
func indexOf(element tview.Primitive, columns []tview.Primitive) int {
	for i, c := range columns {
		if element == c {
			return i
		}
	}
	return -1
}

// Returns the modal (a centered message window)
func createModal(primitive tview.Primitive, x, y int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, x, 0).
		SetRows(0, y, 0).
		AddItem(primitive, 1, 1, 1, 1, 0, 0, true)
}

// Returns the path to the user's home directory, depending on the operating system
func getHomePath() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return dirname
}

// Creates an example card if the board is empty
func (c *Cards) CreateDefaultCard() {
	*c = append(*c, Card{
		Col:  0,
		Pos:  0,
		Name: "The default card",
		Desc: "Create the new one",
	})
}

// Adds a new card to the column at the specified position
func (c *Cards) AddCard(inputName, inputDesc string, idColumn, pos int) {
	if inputName != "" {
		if len(*c) == 0 {
			*c = append(*c, Card{
				Col:  0,
				Pos:  0,
				Name: inputName,
				Desc: inputDesc,
			})
		} else {
			*c = append(*c, Card{
				Col:  idColumn,
				Pos:  pos,
				Name: inputName,
				Desc: inputDesc,
			})
		}
	}
	flagInput = false
}

// Edits the current card
func (c *Cards) EditCard(inputName, inputDesc string, idColumn, pos int) {
	if inputName != "" {
		if len(*c) != 0 {
			for i, card := range *c {
				if card.Col == idColumn && card.Pos == pos {
					cards[i].Name = inputName
					cards[i].Desc = inputDesc
				}
			}
		}
	}
	flagInput = false
}

// Parses JSON file
func (c *Cards) ReadCards() error {
	var path string
	if pathInit == "home" {
		path = filepath.Join(getHomePath(), DBName)
	} else {
		path = filepath.Join(DBName)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.WriteCards()
	} else {
		jsonBlob, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		err = json.Unmarshal(jsonBlob, c)
		return err
	}
}

// Writes changes on the board to a JSON file
func (c *Cards) WriteCards() error {
	var path string
	if pathInit == "home" {
		path = filepath.Join(getHomePath(), DBName)
	} else {
		path = filepath.Join(DBName)
	}

	jsonBlob, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, jsonBlob, 0644)
	return err
}

// Draws cards on a column
func (c *Cards) DrawCards(idColumn int, column tview.Primitive) {
	column.(*tview.List).Clear()
	for _, card := range *c {
		if card.Col == idColumn {
			column.(*tview.List).AddItem(card.Name, card.Desc, 0, nil)
		}
	}
}

func (c *Cards) RefreshCards(columns []tview.Primitive) error {
	if err := cards.WriteCards(); err != nil {
		return err
	}
	sort.Sort(Cards(cards))
	for i := 0; i < len(columns); i++ {
		cards.DrawCards(i, columns[i])
	}
	return nil
}

// Specifies the behavior when a hotkey is pressed
func takeAction(event *tcell.EventKey, columns []tview.Primitive, grid *tview.Grid) *tcell.EventKey {
	if flagInput {
		return event
	}

	focusCol := app.GetFocus()
	idFocusCol := indexOf(focusCol, columns)
	lenFocusCol := focusCol.(*tview.List).GetItemCount()
	posFocusCard := focusCol.(*tview.List).GetCurrentItem()

	switch event.Rune() {
	// Press [Q] to exit the application
	case 'Q':
		app.Stop()

	// Press [j] to move the cursor down the column
	case 'j':
		focusCol.(*tview.List).SetCurrentItem(posFocusCard + 1)

	// Press [k] to move the cursor up the column
	case 'k':
		if posFocusCard > 0 {
			focusCol.(*tview.List).SetCurrentItem(posFocusCard - 1)
		}

	// Press [l] to move the cursor to the next column
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

	// Press [h] to move the cursor to the previous column
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

	// Press [a] to add a new card to the current column
	case 'a':
		formNewCard := tview.NewForm().
			AddInputField("Name: ", "", 70, nil, nil).
			AddInputField("Description: ", "", 70, nil, nil)

		formNewCard.SetButtonsAlign(tview.AlignCenter).
			SetBorder(true).
			SetTitle("Add new card").
			SetTitleAlign(tview.AlignCenter)

		if idFocusCol != -1 {
			formNewCard.AddButton("Create", func() {
				inputName := formNewCard.GetFormItem(0).(*tview.InputField).GetText()
				inputDesc := formNewCard.GetFormItem(1).(*tview.InputField).GetText()
				cards.AddCard(inputName, inputDesc, idFocusCol, lenFocusCol)
				cards.RefreshCards(columns)
				app.SetRoot(grid, true).EnableMouse(true).SetFocus(columns[idFocusCol])
				app.GetFocus().(*tview.List).SetCurrentItem(app.GetFocus().(*tview.List).GetItemCount() - 1)
			})
		}

		flagInput = true
		app.SetRoot(createModal(formNewCard, 70, 10), true).
			EnableMouse(true).
			SetFocus(formNewCard)

	// Press [i] to edit the current card
	case 'i':
		name, desc := focusCol.(*tview.List).GetItemText(posFocusCard)
		formNewCard := tview.NewForm().
			AddInputField("Name: ", name, 70, nil, nil).
			AddInputField("Description: ", desc, 70, nil, nil)

		formNewCard.SetButtonsAlign(tview.AlignCenter).
			SetBorder(true).
			SetTitle("Edit the card").
			SetTitleAlign(tview.AlignCenter)

		if idFocusCol != -1 {
			formNewCard.AddButton("Update", func() {
				inputName := formNewCard.GetFormItem(0).(*tview.InputField).GetText()
				inputDesc := formNewCard.GetFormItem(1).(*tview.InputField).GetText()
				cards.EditCard(inputName, inputDesc, idFocusCol, posFocusCard)
				cards.RefreshCards(columns)
				app.SetRoot(grid, true).EnableMouse(true).SetFocus(columns[idFocusCol])
				app.GetFocus().(*tview.List).SetCurrentItem(posFocusCard)
			})
		}

		flagInput = true
		app.SetRoot(createModal(formNewCard, 70, 10), true).
			EnableMouse(true).
			SetFocus(formNewCard)

	// Press [D] to delete the current card
	case 'D':
		if lenFocusCol > 0 && idFocusCol != -1 {
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

			cards.RefreshCards(columns)
			if len(cards) == 0 {
				cards.CreateDefaultCard()
				cards.RefreshCards(columns)
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

	// Press [K] to move the card up
	case 'K':
		if lenFocusCol > 1 && posFocusCard > 0 && idFocusCol != -1 {
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

			cards.RefreshCards(columns)
			app.GetFocus().(*tview.List).SetCurrentItem(posFocusCard - 1)
		}

	// Press [J] to move the card down
	case 'J':
		if lenFocusCol > 1 && posFocusCard < lenFocusCol-1 && idFocusCol != -1 {
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

			cards.RefreshCards(columns)
			app.GetFocus().(*tview.List).SetCurrentItem(posFocusCard + 1)
		}

	// Press [L] to move the card to the next column
	case 'L':
		if lenFocusCol > 0 && idFocusCol < len(columns)-1 && idFocusCol != -1 {
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

			cards.RefreshCards(columns)
			app.SetFocus(columns[idFocusCol+1])
			app.GetFocus().(*tview.List).SetCurrentItem(lenNextCol)
		}

	// Press [H] to move the card to the previous column
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

			cards.RefreshCards(columns)
			app.SetFocus(columns[idFocusCol-1])
			app.GetFocus().(*tview.List).SetCurrentItem(lenPrevCol)
		}
	}
	return event
}

// Runs the main logic
func mainLogic() error {
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetTextColor(ColorElem).
			SetText(text)
	}

	// Reading the JSON file
	if err := cards.ReadCards(); err != nil {
		return err
	}

	// Setting the grid of the table
	grid := tview.NewGrid().
		SetRows(1, -1, 1).
		SetBorders(true)

	// Setting columns of the table
	var numColumns = len(columnsStr)
	headers := make([]tview.Primitive, 0)
	columns := make([]tview.Primitive, 0)
	for i := 0; i < numColumns; i++ {
		headers = append(headers, newPrimitive(columnsStr[i]))
		columns = append(columns, tview.NewList().SetSelectedFocusOnly(true))
		grid.AddItem(headers[i], 0, i, 1, 1, 0, 0, false).
			AddItem(columns[i], 1, i, 1, 1, 0, 0, false)
	}

	// Setting the footer of the table
	footer := newPrimitive("[ " + AppName + " " + Version + " ]  " + Hotkeys)
	grid.AddItem(footer, 2, 0, 1, numColumns, 0, 0, false)

	// Drawing sorted cards
	sort.Sort(Cards(cards))
	for i := 0; i < numColumns; i++ {
		cards.DrawCards(i, columns[i])
	}

	// Setting a function which captures all key events
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return takeAction(event, columns, grid)
	})

	// Focusing the cursor on the card
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

	err := app.Run()
	return err
}

func main() {
	cards.CreateDefaultCard()

	if err := mainLogic(); err != nil {
		log.Fatal(err)
	}
}
