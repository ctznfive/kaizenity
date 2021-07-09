package main

import (
    "fmt"
    "log"
	"github.com/rivo/tview"
)

const appName = "Kaizenity"
const version = "1.0.0"
const hotkeys = "[i] Add  [D] Remove  [j] Select Next  [k] Select Prev  [L] Move Right  [H] Move Left  [q] Quit"

var (
    app = tview.NewApplication()
    pathInit string
)

func mainLogic(columnsFlow []string, pathInit string) (err error) {
    /*
    if err := tasksDB.getTasks(pathInit); err != nil {
        return err
    }
    */

    newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	//finderFocus := app.GetFocus()

	headerA := newPrimitive(columnsFlow[0])
	headerB := newPrimitive(columnsFlow[1])
	headerC := newPrimitive(columnsFlow[2])
	headerD := newPrimitive(columnsFlow[3])

	columnA := tview.NewList()
	columnB := tview.NewList()
	columnC := tview.NewList()
	columnD := tview.NewList()

	footer := newPrimitive("[ " + appName + " " + version + " ]  " + hotkeys)

	//StagePopulation(allTasks, 0, stageA)
	//StagePopulation(allTasks, 1, stageB)
	//StagePopulation(allTasks, 2, stageC)

	grid := tview.NewGrid().
		SetRows(1, -1, 1).
		SetColumns(-1, -1, 30, -1).
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

	if err = app.SetRoot(grid, true).Run(); err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
    columnsFlow := []string{"BACKLOG", "TODO", "DOING", "DONE"}
	pathInit = ""

    if err := mainLogic(columnsFlow, pathInit); err != nil {
        fmt.Println(err)
    }
}
