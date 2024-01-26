//go:build !nopdf

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xyproto/files"
)

// SavePDF can save the text as a PDF document. It's pretty experimental.
func (e *Editor) SavePDF(title, filename string) error {
	// Check if the file exists
	if files.Exists(filename) {
		return fmt.Errorf("%s already exists", filename)
	}

	// Build a large string with the document contents, while expanding tabs.
	// Also count the maximum line length.
	var sb strings.Builder
	maxLineLength := 0
	for i := 0; i < e.Len(); i++ {
		line := e.Line(LineIndex(i))
		// Expand tabs for each line
		sb.WriteString(strings.ReplaceAll(line, "\t", strings.Repeat(" ", e.indentation.PerTab)) + "\n")
		// Count the maximum line length
		if len(line) > maxLineLength {
			maxLineLength = len(line)
		}
	}
	contents := sb.String()

	// If the lines are long, make the font smaller
	var smallFontSize, largeFontSize float64
	if maxLineLength > 100 {
		smallFontSize = 8 // the text should never be smaller than 8
		largeFontSize = 14
	} else if maxLineLength > 80 {
		smallFontSize = 9
		largeFontSize = 14
	} else {
		smallFontSize = 10.0
		largeFontSize = 14
	}

	// Create a timestamp for the current date, using the "2006-01-02" format
	timestamp := time.Now().Format("2006-01-02")

	// Use A4 and Unicode
	pdf := gofpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("") // "" defaults to "cp1252"

	pdf.SetTopMargin(30)

	// Top text
	pdf.SetHeaderFunc(func() {
		pdf.SetY(5)
		pdf.SetFont("Helvetica", "", 8)
		// Top right corner
		pdf.CellFormat(0, 0, timestamp, "", 0, "R", false, 0, "")
	})

	// Bottom text
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "", 8)
		// Bottom center
		pdf.CellFormat(0, 10, fmt.Sprintf("%d", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetY(20)
	ht := pdf.PointConvert(8.0)

	// Header
	pdf.SetFont("Courier", "B", largeFontSize)
	pdf.MultiCell(190, ht, tr(title+"\n\n"), "", "L", false)
	pdf.Ln(ht)

	// Body
	pdf.SetFont("Courier", "", smallFontSize)
	pdf.MultiCell(190, ht, tr(contents+"\n"), "", "L", false)
	pdf.Ln(ht)

	// Save to file
	return pdf.OutputFileAndClose(filename)
}
