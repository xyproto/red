package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/xyproto/vt100"
)

// SymbolMenu starts a loop where keypresses are handled. When a choice is made, a number is returned.
// x and y are returned. -1,-1 is "no choice", 0,0 is the top left index.
func (e *Editor) SymbolMenu(status *StatusBar, tty *vt100.TTY, title string, choices [][]string, titleColor, textColor, highlightColor vt100.AttributeColor) (int, int, bool) {
	// Clear the existing handler
	signal.Reset(syscall.SIGWINCH)

	var (
		c          = vt100.NewCanvas()
		symbolMenu = NewSymbolWidget(title, choices, titleColor, textColor, highlightColor, e.Background, c.W(), c.H())
		sigChan    = make(chan os.Signal, 1)
		running    = true
		changed    = true
		cancel     = false
	)

	// Set up a new resize handler
	signal.Notify(sigChan, syscall.SIGWINCH)

	go func() {
		for range sigChan {
			resizeMut.Lock()
			// Create a new canvas, with the new size
			nc := c.Resized()
			if nc != nil {
				vt100.Clear()
				c = nc
				symbolMenu.Draw(c)
				c.Redraw()
				changed = true
			}

			// Inform all elements that the terminal was resized
			resizeMut.Unlock()
		}
	}()

	vt100.Clear()
	vt100.Reset()
	c.Redraw()

	// Set the initial menu index
	symbolMenu.SelectIndex(0, 0)

	for running {

		// Draw elements in their new positions

		if changed {
			resizeMut.RLock()
			symbolMenu.Draw(c)
			resizeMut.RUnlock()
			// Update the canvas
			c.Draw()
		}

		// Handle events
		key := tty.String()
		switch key {
		case "↑", "c:16": // Up or ctrl-p
			resizeMut.Lock()
			symbolMenu.Up(c)
			changed = true
			resizeMut.Unlock()
		case "←": // Left
			resizeMut.Lock()
			symbolMenu.Left(c)
			changed = true
			resizeMut.Unlock()
		case "↓", "c:14": // Down or ctrl-n
			resizeMut.Lock()
			symbolMenu.Down(c)
			changed = true
			resizeMut.Unlock()
		case "→": // Right
			resizeMut.Lock()
			symbolMenu.Right(c)
			changed = true
			resizeMut.Unlock()
		case "c:1": // Top, ctrl-a
			resizeMut.Lock()
			symbolMenu.SelectFirst()
			changed = true
			resizeMut.Unlock()
		case "c:5": // Bottom, ctrl-e
			resizeMut.Lock()
			symbolMenu.SelectLast()
			changed = true
			resizeMut.Unlock()
		case "c:27", "q", "c:3", "c:17", "c:15": // ESC, q, ctrl-c, ctrl-q or ctrl-o
			running = false
			changed = true
			cancel = true
		case " ", "c:13": // Space or Return
			running = false
			changed = true
		case "n": // handy shortcut
			for y := 0; y < len(choices); y++ {
				for x := 0; x < len(choices[y]); x++ {
					if choices[y][x] == "ℕ" {
						return x, y, false
					}
				}
			}
		case "t": // handy shortcut
			for y := 0; y < len(choices); y++ {
				for x := 0; x < len(choices[y]); x++ {
					if choices[y][x] == "⊤" {
						return x, y, false
					}
				}
			}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9": // 0 .. 9
			number, err := strconv.Atoi(key)
			if err != nil {
				break
			}
			resizeMut.Lock()
			symbolMenu.SelectIndex(0, uint(number))
			changed = true
			resizeMut.Unlock()
		}

		// If the menu was changed, draw the canvas
		if changed {
			c.Draw()
		}

	}

	// Clear the existing handler
	signal.Reset(syscall.SIGWINCH)

	// Restore the signal handlers
	e.SetUpSignalHandlers(c, tty, status)

	x, y := symbolMenu.Selected()
	return x, y, cancel
}
