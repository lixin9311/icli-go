package icli

import (
	"errors"
	"github.com/nsf/termbox-go"
	"os"
)

// Errors
var (
	ExitIcli    = errors.New("Exit the main icli loop.") // Function return this err to exit the mainloop
	CmdNotFound = errors.New("Command not found.")       // Command not found.
	PrintDesc   = errors.New("Print Description.")       // Print description.
)

const (
	// TabLength 4 or 8? That's a question.
	TabLength = 4
	// NumOfLines is the Maximum lines of the stdbuffer.
	NumOfLines = 1000
	// NumOfHistory is the Maximum lines of recored user input.
	NumOfHistory = 1000
	coldef       = termbox.ColorDefault
)

// AddCmd adds cmd to icli, options of the same name will overwrites by the latest one.
func AddCmd(cmds []CommandOption) {
	cl.AddCmd(cmds)
}

// SetPromt sets the promt of user input line.
func SetPromt(s string) {
	cl.SetPromt([]byte(s))
}

// Start the main icli loop.
// It will redirect the stdout.
func Start(errorHandler func(error) error) {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	// redirect the stdout
	old := os.Stdout
	os.Stdout = stdout.Fd()

	termbox.SetInputMode(termbox.InputEsc)
	cl.PrintDesc()
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				break mainloop
			case termbox.KeyArrowLeft, termbox.KeyCtrlB:
				cl.MoveCursorOneRuneBackward()
			case termbox.KeyArrowRight, termbox.KeyCtrlF:
				cl.MoveCursorOneRuneForward()
			case termbox.KeyArrowUp:
				cl.MoveUp()
			case termbox.KeyArrowDown:
				cl.MoveDown()
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				cl.DeleteRuneBackward()
			case termbox.KeyDelete, termbox.KeyCtrlD:
				cl.DeleteRuneForward()
			case termbox.KeyTab:
				cl.KeyCompletion()
			case termbox.KeySpace:
				cl.InsertRune(' ')
			case termbox.KeyCtrlK:
				cl.DeleteTheRestOfTheLine()
			case termbox.KeyHome, termbox.KeyCtrlA:
				cl.MoveCursorToBeginningOfTheLine()
			case termbox.KeyEnd, termbox.KeyCtrlE:
				cl.MoveCursorToEndOfTheLine()
			case termbox.KeyEnter:
				stdout.PutCell(cl.Cell(), true, nilFunc)
				err := cl.Execute()
				// if the error is within icli errors.
				if err == ExitIcli {
					break mainloop
				} else if err == CmdNotFound {
					stdout.PutString("Command not found.\n", nilFunc)
				} else if err == PrintDesc {
					cl.PrintDesc()
				} else {
					// if the error is not in icli errors or nil, call the error handler.
					err = errorHandler(err)
					if err == ExitIcli {
						break mainloop
					} else if err == PrintDesc {
						cl.PrintDesc()
					}
				}
				cl.cmd = []byte{}
				cl.MoveCursorToBeginningOfTheLine()
			default:
				if ev.Ch != 0 {
					cl.InsertRune(ev.Ch)
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
		redrawAll()
	}
	// recover the stdout
	os.Stdout = old
}
