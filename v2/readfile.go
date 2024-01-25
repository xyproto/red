package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/xyproto/binary"
)

// ReadFileAndProcessLines reads the named file concurrently, processes its lines, and updates the Editor.
func (e *Editor) ReadFileAndProcessLines(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if strings.HasSuffix(filename, ".gz") {
		data, err = gUnzipData(data)
		if err != nil {
			return err
		}
	}

	var isUTF16 bool
	e.binaryFile, isUTF16 = binary.DataAndUTF16(data)

	if isUTF16 {
		if newData, err := utf16ToUTF8(data); err == nil { // success
			data = newData
		}
	}

	var (
		reader           = bufio.NewReader(bytes.NewReader(data))
		lines            = make(map[int][]rune)
		index            int
		tabIndentCounter int64
		first            byte
	)

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		// Remove the newline character at the end of the line
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		if e.binaryFile {
			lines[index] = []rune(line)
		} else {
			line = opinionatedStringReplacer.Replace(line)
			if len(line) > 2 {
				first = line[0]
				if first == '\t' {
					tabIndentCounter++
				} else if first == ' ' && line[1] == ' ' {
					tabIndentCounter--
				}
			}
			lines[index] = []rune(line)
		}
		index++

		if err == io.EOF {
			break
		}
	}
	e.Clear()
	e.lines = lines
	if detectedTabs := tabIndentCounter > 0; !e.binaryFile && e.indentation.Spaces {
		e.detectedTabs = &detectedTabs
		e.indentation.Spaces = !detectedTabs
	}
	e.changed = true
	return nil
}

// LoadByteLine loads a single byte line
func (e *Editor) LoadByteLine(ib IndexByteLine, eMut, tcMut *sync.RWMutex, tabIndentCounter, numLines *int, wg *sync.WaitGroup) {
	// Require at least two bytes. Ignore lines with a single tab indentation or a single space
	if len(ib.byteLine) > 2 {
		first := ib.byteLine[0]
		if first == '\t' {
			tcMut.Lock()
			*tabIndentCounter++ // a tab indentation counts like a positive tab indentation
			tcMut.Unlock()
		} else if first == ' ' && ib.byteLine[1] == ' ' { // assume that two spaces is the smallest space indentation
			tcMut.Lock()
			*tabIndentCounter-- // a space indentation counts like a negative tab indentation
			tcMut.Unlock()
		}
	}
	eMut.Lock()
	e.lines[ib.index] = []rune(string(ib.byteLine))
	*numLines++
	eMut.Unlock()
	wg.Done()
}

// LoadBytes replaces the current editor contents with the given bytes
func (e *Editor) LoadBytes(data []byte) {
	lineCount := bytes.Count(data, []byte{'\n'}) + 1

	// Prepare an empty map to load the lines into
	e.Clear()
	e.lines = make(map[int][]rune, lineCount)

	var (
		isUTF16 bool
	)
	e.binaryFile, isUTF16 = binary.DataAndUTF16(data)

	if isUTF16 {
		if newData, err := utf16ToUTF8(data); err == nil { // success
			data = newData
		}
	}

	var (
		// Split the bytes into lines
		byteLines = bytes.Split(data, []byte{'\n'})

		// Place the lines into the editor, while counting tab indentations vs space indentations
		tabIndentCounter int

		// Count the number of lines as the lines are being processed
		numLines int

		// Mutex for the editor lines and the numLines counter
		eMut sync.RWMutex

		// Mutex for the tabIndentCounter
		tcMut sync.RWMutex
	)

	var wg sync.WaitGroup
	var byteLine []byte
	for i := 0; i < lineCount; i++ {
		byteLine = byteLines[i]
		wg.Add(1)
		go e.LoadByteLine(IndexByteLine{byteLine, i}, &eMut, &tcMut, &tabIndentCounter, &numLines, &wg)
	}
	wg.Wait()

	// If the last line is empty, delete it
	if numLines > 0 && len(e.lines[numLines-1]) == 0 {
		delete(e.lines, numLines-1)
	}

	if detectedTabs := tabIndentCounter > 0; detectedTabs && e.indentation.Spaces {
		// Check if there were more tab indentations than space indentations
		e.detectedTabs = &detectedTabs
		e.indentation.Spaces = !detectedTabs
	}

	// Mark the editor contents as "changed"
	e.changed = true
}
