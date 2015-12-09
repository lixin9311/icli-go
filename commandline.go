package icli

import (
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"strings"
	"unicode/utf8"
)

// CommandOption is an option like "ls" or "cd". Stdout is redirected to screen buffer.
// Function can call fmt.Print directly
type CommandOption struct {
	Command     string
	Description string
	Function    func(args ...string) error
}

// CommandLine handles the user inputs.
type CommandLine struct {
	cmd              []byte
	promt            []byte
	lineVoffset      int
	cursorBoffset    int
	cursorVoffset    int
	cursorCoffset    int
	posX, posY       int
	vlength          int
	historyOffset    int
	historyTmpOffset int
	cmds             map[string]CommandOption
	history          [NumOfHistory][]byte
}

// KeyCompletion based on existing Command Options.
func (cmdline *CommandLine) KeyCompletion() {
	cmd := strings.Split(string(cmdline.cmd), " ")
	if len(cmd) != 1 {
		return
	}
	for name := range cmdline.cmds {
		if strings.HasPrefix(name, cmd[0]) {
			cmdline.cmd = []byte(name + " ")
			cmdline.MoveCursorToEndOfTheLine()
			return
		}
	}
}

// SetPromt sets the prefix of user command input line.
func (cmdline *CommandLine) SetPromt(b []byte) {
	cmdline.promt = b
}

// AddCmd a Command Option.
func (cmdline *CommandLine) AddCmd(cmds []CommandOption) {
	if cmdline.cmds == nil {
		cmdline.cmds = map[string]CommandOption{}
	}
	for _, v := range cmds {
		cmdline.cmds[v.Command] = v
	}
}

// HasCmd checks whether the Command Option exists.
func (cmdline *CommandLine) HasCmd(cmd string) bool {
	if cmdline.cmds == nil {
		return false
	}
	if _, ok := cmdline.cmds[cmd]; !ok {
		return false
	}
	return true
}

// MoveUp handles key arrow up.
func (cmdline *CommandLine) MoveUp() {
	if cmdline.historyTmpOffset == 0 {
		cmdline.history[cmdline.historyOffset] = cmdline.cmd
	}
	history := cmdline.history[(NumOfHistory+cmdline.historyOffset-cmdline.historyTmpOffset-1)%NumOfHistory]
	if history == nil {
		return
	}
	cmdline.historyTmpOffset = (cmdline.historyTmpOffset + 1) % NumOfHistory
	cmdline.cmd = history
	cmdline.MoveCursorToEndOfTheLine()
}

// MoveDown handles key arrow down.
func (cmdline *CommandLine) MoveDown() {
	if cmdline.historyTmpOffset == 0 {
		return
	}
	cmdline.historyTmpOffset = (cmdline.historyTmpOffset - 1) % NumOfHistory
	cmdline.cmd = cmdline.history[(NumOfHistory+cmdline.historyOffset-cmdline.historyTmpOffset)%NumOfHistory]
	cmdline.MoveCursorToEndOfTheLine()
}

// Execute the current cmd.
func (cmdline *CommandLine) Execute() error {
	if len(cmdline.cmd) == 0 {
		return nil
	}
	cmdline.historyTmpOffset = 0
	cmdline.history[cmdline.historyOffset] = cmdline.cmd
	cmdline.historyOffset = (cmdline.historyOffset + 1) % NumOfHistory
	cmdArgs := strings.Split(string(cmdline.cmd), " ")
	if cmdline.HasCmd(cmdArgs[0]) {
		return cmdline.cmds[cmdArgs[0]].Function(cmdArgs...)
	}
	return CmdNotFound
}

// Draw itcmdline on point(x, y).
// \t should not exist, but here is a good excise for tab process.
func (cmdline *CommandLine) Draw(x, y int) {
	const coldef = termbox.ColorDefault
	cmdline.posX = x
	cmdline.posY = y
	vDx := x
	vDy := y
	tabstop := runewidth.StringWidth(string(cmdline.promt))
	bytePos := 0
	cmdNameEnd := len(cmdline.cmd)
	spacePos := strings.IndexRune(string(cmdline.cmd), ' ')
	if spacePos == -1 {
		spacePos = len(cmdline.cmd)
	}
	tabPos := strings.IndexRune(string(cmdline.cmd), '\t')
	if tabPos == -1 {
		tabPos = len(cmdline.cmd)
	}
	if spacePos > tabPos {
		cmdNameEnd = tabPos
	} else {
		cmdNameEnd = spacePos
	}
	cmdName := string(cmdline.cmd[:cmdNameEnd])
	text := append(cmdline.promt, cmdline.cmd...)
	for len(text) > 0 {
		r, size := utf8.DecodeRune(text)
		if vDx >= tabstop {
			tabstop += TabLength
		}
		if bytePos >= len(cmdline.promt) && bytePos < len(cmdline.promt)+cmdNameEnd {
			if cmdline.HasCmd(cmdName) {
				termbox.SetCell(vDx, vDy, r, termbox.ColorGreen, coldef)
			} else {
				termbox.SetCell(vDx, vDy, r, termbox.ColorRed, coldef)
			}
			vDx += runewidth.RuneWidth(r)
			goto next_rune
		}
		if r == '\t' {
			for vDx < tabstop {
				termbox.SetCell(vDx, vDy, ' ', coldef, coldef)
				vDx++
			}
		} else {
			termbox.SetCell(vDx, vDy, r, coldef, coldef)
			vDx += runewidth.RuneWidth(r)
		}
	next_rune:
		text = text[size:]
		bytePos += size
	}
	cmdline.vlength = vDx - x
}

// Cell return the CommandLine buffer on the screen.
func (cmdline *CommandLine) Cell() []termbox.Cell {
	cell := make([]termbox.Cell, cmdline.vlength)
	screen := termbox.CellBuffer()
	w, _ := termbox.Size()
	for i := 0; i < cmdline.vlength; i++ {
		cell[i] = screen[cmdline.posY*w+cmdline.posX+i]
	}
	return cell
}

// MoveCursorTo the postion of byte.
func (cmdline *CommandLine) MoveCursorTo(boffset int) {
	cmdline.cursorBoffset = boffset
	cmdline.cursorVoffset = 0
	cmdline.cursorCoffset = 0
	text := cmdline.cmd[:boffset]
	tabstop := 0
	for len(text) > 0 {
		r, size := utf8.DecodeRune(text)
		if cmdline.cursorVoffset >= tabstop {
			tabstop += TabLength
		}
		cmdline.cursorCoffset++
		if r == '\t' {
			cmdline.cursorVoffset = tabstop
		} else {
			cmdline.cursorVoffset += runewidth.RuneWidth(r)
		}
		text = text[size:]
	}
}

// RuneUnderCursor returns the current rune under cursor.
func (cmdline *CommandLine) RuneUnderCursor() (rune, int) {
	return utf8.DecodeRune(cmdline.cmd[cmdline.cursorBoffset:])
}

// RuneBeforeCursor returns the rune before the cursor.
func (cmdline *CommandLine) RuneBeforeCursor() (rune, int) {
	return utf8.DecodeLastRune(cmdline.cmd[:cmdline.cursorBoffset])
}

// MoveCursorOneRuneForward moves the cursor one rune forward.
func (cmdline *CommandLine) MoveCursorOneRuneForward() {
	if cmdline.cursorBoffset == len(cmdline.cmd) {
		return
	}
	_, size := cmdline.RuneUnderCursor()
	cmdline.MoveCursorTo(cmdline.cursorBoffset + size)
}

// MoveCursorOneRuneBackward moves the cursor one rune backward.
func (cmdline *CommandLine) MoveCursorOneRuneBackward() {
	if cmdline.cursorBoffset == 0 {
		return
	}
	_, size := cmdline.RuneBeforeCursor()
	cmdline.MoveCursorTo(cmdline.cursorBoffset - size)
}

// InsertRune to the postion of cursor.
func (cmdline *CommandLine) InsertRune(r rune) {
	var buf [utf8.UTFMax]byte
	// [4]byte to slice
	n := utf8.EncodeRune(buf[:], r)
	cmdline.cmd = byteSliceInsert(cmdline.cmd, cmdline.cursorBoffset, buf[:n])
	cmdline.MoveCursorOneRuneForward()
}

// DeleteRuneBackward handles backspace key.
func (cmdline *CommandLine) DeleteRuneBackward() {
	if cmdline.cursorBoffset == 0 {
		return
	}
	// move backword first
	cmdline.MoveCursorOneRuneBackward()
	_, size := cmdline.RuneUnderCursor()
	cmdline.cmd = byteSliceRemove(cmdline.cmd, cmdline.cursorBoffset, cmdline.cursorBoffset+size)
}

// DeleteRuneForward handles del key.
func (cmdline *CommandLine) DeleteRuneForward() {
	if cmdline.cursorBoffset == len(cmdline.cmd) {
		return
	}
	_, size := cmdline.RuneUnderCursor()
	cmdline.cmd = byteSliceRemove(cmdline.cmd, cmdline.cursorBoffset, cmdline.cursorBoffset+size)
}

// MoveCursorToEndOfTheLine handles end key.
func (cmdline *CommandLine) MoveCursorToEndOfTheLine() {
	cmdline.MoveCursorTo(len(cmdline.cmd))
}

// MoveCursorToBeginningOfTheLine handles home key.
func (cmdline *CommandLine) MoveCursorToBeginningOfTheLine() {
	cmdline.MoveCursorTo(0)
}

// DeleteTheRestOfTheLine like the name, handles CtrlK.
func (cmdline *CommandLine) DeleteTheRestOfTheLine() {
	cmdline.cmd = cmdline.cmd[:cmdline.cursorBoffset]
}

// Cursor reports the cursor postion.
func (cmdline *CommandLine) Cursor() (X, Y int) {
	X = cmdline.posX + cmdline.cursorVoffset + runewidth.StringWidth(string(cmdline.promt))
	Y = cmdline.posY
	return
}
