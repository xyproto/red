//go:build nopdf

package main

import (
	"errors"
)

// SavePDF can save the text as a PDF document. It's pretty experimental.
func (e *Editor) SavePDF(title, filename string) error {
	return errors.New("PDF export support was disabled at build time")
}
