package utils

import "github.com/atotto/clipboard"

const (
	TypeText    = "text"
	TypeFile    = "file"
	TypeBitmap  = "bitmap"
	TypeUnknown = "unknown"
)

var xlipboard XlipboardService

// Xlipboard returns an object that provides access to the system clipboard.
func Xlipboard() *XlipboardService {
	return &xlipboard
}

// Text returns the current text data of the clipboard.
func (c *XlipboardService) Text() (text string, err error) {
	return clipboard.ReadAll()
}

// SetText sets the current text data of the clipboard.
func (c *XlipboardService) SetText(s string) error {
	return clipboard.WriteAll(s)
}
