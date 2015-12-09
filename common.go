package icli

import (
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"sync"
)

var (
	cl     CommandLine
	stdout StdBuffer
)

func nilFunc() {}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}
func fill(x, y, w, h int, cell termbox.Cell) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			termbox.SetCell(x+dx, y+dy, cell.Ch, cell.Fg, cell.Bg)
		}
	}
}

func byteSliceGrow(s []byte, desiredCap int) []byte {
	if cap(s) < desiredCap {
		ns := make([]byte, len(s), desiredCap)
		copy(ns, s)
		return ns
	}
	return s
}

func byteSliceRemove(text []byte, from, to int) []byte {
	size := to - from
	copy(text[from:], text[to:])
	text = text[:len(text)-size]
	return text
}

func byteSliceInsert(text []byte, offset int, what []byte) []byte {
	n := len(text) + len(what)
	text = byteSliceGrow(text, n)
	text = text[:n]
	copy(text[offset+len(what):], text[offset:])
	copy(text[offset:], what)
	return text
}

var lock sync.Mutex

// Not thread-safe.
func redrawAll() {
	lock.Lock()
	defer lock.Unlock()
	termbox.Clear(coldef, coldef)
	w, h := termbox.Size()
	cl.Draw(0, h-3)
	stdout.Draw(0, h-4, w, h)
	tbprint(0, h-1, termbox.ColorMagenta, coldef, "Press ESC to quit")
	termbox.SetCursor(cl.Cursor())
	termbox.Flush()
}
