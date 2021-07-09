package main

import (
    "fmt"
    "log"
    "os"
    "encoding/json"
    "github.com/rivo/tview"
    "github.com/gdamore/tcell/v2"
)

const (
    AppName = "Kaizenity"
    Version = "1.0.0"
    Hotkeys = "[i] Add  [D] Remove  [j] Select Next  [k] Select Prev  [L] Move Right  [H] Move Left  [q] Quit"
    ColorElem = tcell.ColorBlue
    FileName = "kaizenitydb.json"
)

var (
    pathInit string
    cards    Cards

    app = tview.NewApplication()
)

type Card struct {
    ID     int64  `json:"id"`
    Name   string `json:"name"`
    Desc   string `json:"desc"`
    Column int    `json:"column"`
    Pos    int64  `json:"pos"`
}

type Cards []Card

func (c *Cards) ReadCards(path string) error {
    path += FileName
    jsonBlob, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    err = json.Unmarshal(jsonBlob, c)
    return err
}

func (c *Cards) drawCards(column int, primitive tview.Primitive) {
    primitive.(*tview.List).Clear()
    for _, card := range *c {
        if card.Column == column {
            primitive.(*tview.List).AddItem(card.Name, card.Desc, 0, nil)
        }
    }
}

func mainDraw(columns []string, path string) (err error) {
    newPrimitive := func(text string) tview.Primitive {
        return tview.NewTextView().
            SetTextAlign(tview.AlignCenter).
            SetTextColor(ColorElem).
            SetText(text)
        }

    if err = cards.ReadCards(path); err != nil {
        return err
    }

    headerA := newPrimitive(columns[0])
    headerB := newPrimitive(columns[1])
    headerC := newPrimitive(columns[2])
    headerD := newPrimitive(columns[3])

    columnA := tview.NewList()
    columnB := tview.NewList()
    columnC := tview.NewList()
    columnD := tview.NewList()

    footer := newPrimitive("[ " + AppName + " " + Version + " ]  " + Hotkeys)

    grid := tview.NewGrid().
        SetRows(1, -1, 1).
        SetColumns(-1, -1, -1, -1).
        SetBorders(true)

	// Layout for screens narrower than 100 cells (only TODO is visible)
    grid.AddItem(headerA, 0, 0, 0, 0, 0, 0, false).
         AddItem(columnA, 0, 0, 0, 0, 0, 0, false).
         AddItem(headerB, 0, 0, 0, 0, 0, 0, false).
         AddItem(columnB, 0, 0, 0, 0, 0, 0, false).
         AddItem(headerC, 0, 0, 1, 4, 0, 0, false).
         AddItem(columnC, 1, 0, 1, 4, 0, 0, true).
         AddItem(headerD, 0, 0, 0, 0, 0, 0, false).
         AddItem(columnD, 0, 0, 0, 0, 0, 0, false).
         AddItem(footer, 2, 0, 1, 4, 0, 0, false)

	// Layout for screens wider than 100 cells
    grid.AddItem(headerA, 0, 0, 1, 1, 0, 100, false).
         AddItem(columnA, 1, 0, 1, 1, 0, 100, true).
         AddItem(headerB, 0, 1, 1, 1, 0, 100, false).
         AddItem(columnB, 1, 1, 1, 1, 0, 100, false).
         AddItem(headerC, 0, 2, 1, 1, 0, 100, false).
         AddItem(columnC, 1, 2, 1, 1, 0, 100, false).
         AddItem(headerD, 0, 3, 1, 1, 0, 100, false).
         AddItem(columnD, 1, 3, 1, 1, 0, 100, false).
         AddItem(footer, 2, 0, 1, 4, 0, 0, false)

    cards.drawCards(0, columnA)
    cards.drawCards(1, columnB)
    cards.drawCards(2, columnC)
    cards.drawCards(3, columnD)

    if err = app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
        log.Fatal(err)
    }
    return nil
}

func main() {
    columnsFlow := []string{"BACKLOG", "TODO", "DOING", "DONE"}
    pathInit = ""

    if err := mainDraw(columnsFlow, pathInit); err != nil {
        fmt.Println(err)
    }
}
