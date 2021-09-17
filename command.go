package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xyproto/env"
	"github.com/xyproto/guessica"
	"github.com/xyproto/vt100"
)

var lastCommandFile = filepath.Join(userCacheDir, "o/last_command.sh")

// UserSave saves the file and the location history
func (e *Editor) UserSave(c *vt100.Canvas, tty *vt100.TTY, status *StatusBar) {
	// Save the file
	if err := e.Save(c, tty); err != nil {
		status.SetErrorMessage(err.Error())
		status.Show(c, e)
		return
	}

	// Save the current location in the location history and write it to file
	if absFilename, err := e.AbsFilename(); err == nil { // no error
		e.SaveLocation(absFilename, e.locationHistory)
	}

	// Status message
	status.Clear(c)
	status.SetMessage("Saved " + e.filename)
	status.Show(c, e)
}

// Actions is a list of action titles and a list of action functions.
// The key is an int that is the same for both.
type Actions struct {
	actionTitles    map[int]string
	actionFunctions map[int]func()
}

// NewActions will create a new Actions struct
func NewActions() *Actions {
	var a Actions
	a.actionTitles = make(map[int]string)
	a.actionFunctions = make(map[int]func())
	return &a
}

// NewActions2 will create a new Actions struct, while
// initializing it with the given slices of titles and functions
func NewActions2(actionTitles []string, actionFunctions []func()) (*Actions, error) {
	a := NewActions()
	if len(actionTitles) != len(actionFunctions) {
		return nil, errors.New("length of action titles and action functions differ")
	}
	for i, title := range actionTitles {
		a.actionTitles[i] = title
		a.actionFunctions[i] = actionFunctions[i]
	}
	return a, nil
}

// Add will add an action title and an action function
func (a *Actions) Add(title string, f func()) {
	i := len(a.actionTitles)
	a.actionTitles[i] = title
	a.actionFunctions[i] = f
}

// MenuChoices will return a string that lists the titles of
// the available actions.
func (a *Actions) MenuChoices() []string {
	// Create a list of strings that are menu choices,
	// while also creating a mapping from the menu index to a function.
	menuChoices := make([]string, len(a.actionTitles))
	for i, description := range a.actionTitles {
		menuChoices[i] = fmt.Sprintf("[%d] %s", i, description)
	}
	return menuChoices
}

// Perform will call the given function index
func (a *Actions) Perform(index int) {
	a.actionFunctions[index]()
}

// CommandMenu will display a menu with various commands that can be browsed with arrow up and arrow down
// Also returns the selected menu index (can be -1).
func (e *Editor) CommandMenu(c *vt100.Canvas, tty *vt100.TTY, status *StatusBar, undo *Undo, lastMenuIndex int, forced bool, lk *LockKeeper) int {

	const insertFilename = "include.txt"

	wrapWidth := e.wrapWidth
	if wrapWidth == 0 {
		wrapWidth = 80
	}

	// Let the menu item for wrapping words suggest the minimum of e.wrapWidth and the terminal width
	if c != nil {
		w := int(c.Width())
		if w < wrapWidth {
			wrapWidth = w - int(0.05*float64(w))
		}
	}

	wrapWhenTypingToggleText := "Enable word wrap when typing"
	if e.wrapWhenTyping {
		wrapWhenTypingToggleText = "Disable word wrap when typing"
	}

	var extraDashes = false

	// Add initial menu titles and actions
	// Remember to add "undo.Snapshot(e)" in front of function calls that may modify the current file!
	actions, err := NewActions2(
		[]string{
			"Save and quit",
			wrapWhenTypingToggleText,
			"Word wrap at " + strconv.Itoa(wrapWidth),
			"Sort the list of strings on the current line",
			"Insert \"" + insertFilename + "\" at the current line",
		},
		[]func(){
			func() { // save and quit
				e.clearOnQuit = true
				e.UserSave(c, tty, status)
				e.quit = true        // indicate that the user wishes to quit
				e.clearOnQuit = true // clear the terminal after quitting
			},
			func() { // toggle word wrap when typing
				e.wrapWhenTyping = !e.wrapWhenTyping
				if e.wrapWidth == 0 {
					e.wrapWidth = 79
				}
			},
			func() { // word wrap
				// word wrap at the current width - 5, with an allowed overshoot of 5 runes
				tmpWrapAt := e.wrapWidth
				e.wrapWidth = wrapWidth
				if e.WrapAllLinesAt(wrapWidth-5, 5) {
					e.redraw = true
					e.redrawCursor = true
				}
				e.wrapWidth = tmpWrapAt
			},
			func() { // sort strings on the current line
				undo.Snapshot(e)
				if err := e.SortStrings(c, status); err != nil {
					status.Clear(c)
					status.SetErrorMessage(err.Error())
					status.Show(c, e)
				}
			},
			func() { // insert file
				editedFileDir := filepath.Dir(e.filename)
				if err := e.InsertFile(c, filepath.Join(editedFileDir, insertFilename)); err != nil {
					status.Clear(c)
					status.SetErrorMessage(err.Error())
					status.Show(c, e)
				}
			},
		},
	)
	if err != nil {
		// If this happens, menu actions and menu functions are not added properly
		// and it should fail hard, so that this can be fixed.
		panic(err)
	}

	if strings.HasSuffix(e.filename, "PKGBUILD") {
		actions.Add("Call Guessica", func() {
			status.Clear(c)
			status.SetMessage("Calling Guessica")
			status.Show(c, e)

			// Use the temporary directory defined in TMPDIR, with fallback to /tmp
			tempdir := env.Str("TMPDIR", "/tmp")

			tempFilename := ""

			var (
				f   *os.File
				err error
			)
			if f, err = ioutil.TempFile(tempdir, "__o*"+"guessica"); err == nil {
				// no error, everything is fine
				tempFilename = f.Name()
				// TODO: Implement e.SaveAs
				oldFilename := e.filename
				e.filename = tempFilename
				err = e.Save(c, tty)
				e.filename = oldFilename
			}
			if err != nil {
				status.SetErrorMessage(err.Error())
				status.Show(c, e)
				return
			}

			if tempFilename == "" {
				status.SetErrorMessage("Could not create a temporary file")
				status.Show(c, e)
				return
			}

			// Show the status message to the user right now
			status.Draw(c, e.pos.offsetY)

			// Call Guessica, which may take a little while
			err = guessica.UpdateFile(tempFilename)

			if err != nil {
				status.SetErrorMessage("Failed to update PKGBUILD: " + err.Error())
				status.Show(c, e)
			} else {
				if _, err := e.Load(c, tty, tempFilename); err != nil {
					status.ClearAll(c)
					status.SetMessage(err.Error())
					status.Show(c, e)
				}
				// Mark the data as changed, despite just having loaded a file
				e.changed = true
				e.redrawCursor = true

			}
		})
	}

	// Add the syntax highlighting toggle menu item
	if !envNoColor {
		syntaxToggleText := "Disable syntax highlighting"
		if !e.syntaxHighlight {
			syntaxToggleText = "Enable syntax highlighting"
		}
		actions.Add(syntaxToggleText, func() {
			e.ToggleSyntaxHighlight()
		})
	}

	// Add the unlock menu item
	// TODO: Detect if the current file is locked first
	if forced {
		actions.Add("Unlock if locked", func() {
			if absFilename, err := e.AbsFilename(); err == nil { // no issues
				lk.Load()
				lk.Unlock(absFilename)
				lk.Save()
			}
		})
	}

	// Add the portal menu item
	if portal, err := LoadPortal(); err == nil { // no problems
		actions.Add("Close portal at "+portal.String(), func() {
			ClosePortal(e)
		})
	} else {
		// Could not close portal, try opening a new one
		if portal, err := e.NewPortal(); err == nil { // no problems
			actions.Add("Open portal at "+portal.String(), func() {
				portal.Save()
			})
		}
	}

	// Add the "Default theme" menu item text and menu function
	actions.Add("Default theme", func() {
		e.setDefaultTheme()
		e.syntaxHighlight = true
		e.FullResetRedraw(c, status, true)
	})

	// Add the option to change the colors, for non-light themes (fg != black)
	if !e.Light && !envNoColor { // Not a light theme and NO_COLOR is not set

		// Add the "Red/Black theme" menu item text and menu function
		actions.Add("Red/black theme", func() {
			e.setRedBlackTheme()
			e.syntaxHighlight = true
			e.FullResetRedraw(c, status, true)
		})

		// Add the "Light Theme" menu item text and menu function
		actions.Add("Light theme", func() {
			e.setLightTheme()
			e.syntaxHighlight = true
			e.FullResetRedraw(c, status, true)
		})

		// Add the Amber, Green and Blue theme options
		colors := []vt100.AttributeColor{
			vt100.Yellow,
			vt100.LightGreen,
			vt100.LightBlue,
		}
		colorText := []string{
			"Amber",
			"Green",
			"Blue",
		}

		// Add menu items and menu functions for changing the text color
		// while also turning off syntax highlighting.
		for i, color := range colors {
			color := color // per-loop copy of the color variable, since it's closed over
			actions.Add(colorText[i]+" theme", func() {
				e.Foreground = color
				e.Background = vt100.BackgroundDefault // black background
				e.syntaxHighlight = false
				e.FullResetRedraw(c, status, true)
			})
		}
	}

	menuChoices := actions.MenuChoices()

	// Launch a generic menu
	useMenuIndex := 0
	if lastMenuIndex > 0 {
		useMenuIndex = lastMenuIndex
	}

	selected := e.Menu(status, tty, "Select an action", menuChoices, menuTitleColor, menuArrowColor, menuTextColor, menuHighlightColor, menuSelectedColor, useMenuIndex, extraDashes)

	// Redraw the editor contents
	//e.DrawLines(c, true, false)

	if selected < 0 {
		// Output the selected item text
		status.SetMessage("No action taken")
		status.Show(c, e)

		// Do not immediately redraw the editor
		e.redraw = false
		return selected
	}

	// Perform the selected action by passing the function index
	actions.Perform(selected)

	// Redraw editor
	e.redraw = true
	e.redrawCursor = true
	return selected
}

// getCommand takes an *exec.Cmd and returns the command
// it represents, but with "/usr/bin/sh -c " trimmed away.
func getCommand(cmd *exec.Cmd) string {
	s := cmd.Path + " " + strings.Join(cmd.Args[1:], " ")
	return strings.TrimPrefix(s, "/usr/bin/sh -c ")
}

// Save the command to a temporary file, given an exec.Cmd struct
func saveCommand(cmd *exec.Cmd) error {

	p := lastCommandFile

	// First create the folder for the lock file overview, if needed
	folderPath := filepath.Dir(p)
	os.MkdirAll(folderPath, os.ModePerm)

	// Prepare the file
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer f.Close()

	// Strip the leading /usr/bin/sh -c command, if present
	commandString := getCommand(cmd)

	// Write the contents, ignore the number of written bytes
	_, err = f.WriteString(fmt.Sprintf("#!/bin/sh\n%s\n", commandString))
	return err
}
