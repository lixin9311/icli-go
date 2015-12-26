package icli

import (
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"os"
	"sync"
	"unicode/utf8"
)

// StdBuffer is a standard buffer to handle stdout or other input
// to redirect it onto the screen.
type StdBuffer struct {
	buf      [NumOfLines][]termbox.Cell
	offset   int
	writerFd *os.File
	readerFd *os.File
	byteLock sync.Mutex
	bufLock  sync.Mutex
}

// Draw the buffer.
func (stdbuf *StdBuffer) Draw(x, y, termW, termH int) {
	screen := termbox.CellBuffer()
	for i := 0; i < NumOfLines && i < y+1; i++ {
		pos := (NumOfLines + stdbuf.offset - i - 1) % NumOfLines
		line := stdbuf.buf[pos]
		if line == nil {
			continue
		}
		for k := 0; k < len(line) && k < termW-x; k++ {
			screen[(y-i)*termW+x+k] = line[k]
		}
	}
}

// PutString calls PutByte.
func (stdbuf *StdBuffer) PutString(s string, refresh func()) {
	stdbuf.PutByte([]byte(s), refresh)
}

// PutByte converts the byte to termbox.Cell.
// then calls PutCell
func (stdbuf *StdBuffer) PutByte(b []byte, refresh func()) {
	stdbuf.byteLock.Lock()
	defer stdbuf.byteLock.Unlock()
	length := runewidth.StringWidth(string(b))
	cell := make([]termbox.Cell, length)
	text := b
	i := 0
	start := 0
	for len(text) > 0 {
		r, size := utf8.DecodeRune(text)
		cell[i] = termbox.Cell{r, coldef, coldef}
		i += runewidth.RuneWidth(r)
		if r == '\n' {
			stdbuf.PutCell(cell[start:i], true, nilFunc)
			start = i
		}
		text = text[size:]
	}
	stdbuf.PutCell(cell[start:i], false, refresh)
}

// PutCell directly handles the termbox.Cell.
// Usually PutByte and PutString is enough.
func (stdbuf *StdBuffer) PutCell(c []termbox.Cell, endWithNewline bool, refresh func()) {
	stdbuf.bufLock.Lock()
	defer stdbuf.bufLock.Unlock()
	//termbox.Clear(coldef, coldef)
	w, _ := termbox.Size()
	cell := c
	for {
		currentLine := &stdbuf.buf[stdbuf.offset]
		if len(cell)+len(*currentLine) > w {
			length := w - len(*currentLine)
			*currentLine = append(*currentLine, cell[:length]...)
			cell = cell[length:]
			stdbuf.offset = (stdbuf.offset + 1) % NumOfLines
			stdbuf.buf[stdbuf.offset] = nil
		} else {
			*currentLine = append(*currentLine, cell...)
			if endWithNewline {
				stdbuf.offset = (stdbuf.offset + 1) % NumOfLines
				stdbuf.buf[stdbuf.offset] = nil
			}
			break
		}
	}
	refresh()
}

func (stdbuf *StdBuffer) pipe() {
	//reader := bufio.NewReader(stdbuf.readerFd)
	buf := make([]byte, 512)
	for {
		n, _ := stdbuf.readerFd.Read(buf)
		stdbuf.PutByte(buf[:n], redrawAll)
	}
}

// Fd init and return the file describer for redirct the stdout.
func (stdbuf *StdBuffer) Fd() *os.File {
	if stdbuf.writerFd == nil {
		stdbuf.readerFd, stdbuf.writerFd, _ = os.Pipe()
		go stdbuf.pipe()
	}
	return stdbuf.writerFd
}
