package pogo

import (
	"bytes"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// Formatter to write text with an options
type Formatter struct {
	// Border is an prefix to each line
	Border string
	// Prefix is an extra prefix to first line between border and text
	Prefix string
	// Width of line to wrap (no wrap if less or equal zero)
	Width int

	buffer *bytes.Buffer
	output io.Writer
}

// NewFormatter to write in w
func NewFormatter(w io.Writer) *Formatter {
	return &Formatter{
		buffer: &bytes.Buffer{},
		output: w,
	}
}

// Format text
func (f *Formatter) Format(text string) error {
	return recoverHandledError(func() {
		f.mustFormat(text)
	})
}

// BreakLine add \n
func (f *Formatter) BreakLine() error {
	_, err := io.WriteString(f.output, "\n")
	return errors.WithStack(err)
}

func (f *Formatter) mustFormat(text string) {
	lines := f.splitLines(text)
	if f.Prefix != "" {
		f.mustWrite(f.output, f.Border)
		f.mustWrite(f.output, f.Prefix)
		if f.isOneQuotedLine(lines) {
			f.mustWrite(f.output, "\"")
			f.mustWrite(f.output, lines[0])
			f.mustWrite(f.output, "\"\n")
			return
		}
		f.mustWrite(f.output, "\"\"\n")
	}
	for _, line := range lines {
		f.writeLine(line)
	}
}

func (f *Formatter) isOneQuotedLine(lines []string) bool {
	if len(lines) != 1 {
		return false
	}
	if f.Width < 1 {
		return true
	}
	return len(f.Border)+len(f.Prefix)+2+len(lines[0]) <= f.Width
}

func (f *Formatter) splitLines(text string) []string {
	if f.Prefix == "" {
		return strings.Split(text, "\n")
	}
	return strings.SplitAfter(f.escape(text), "\\n")
}

func (f *Formatter) escape(text string) string {
	return strings.NewReplacer(
		"\"", "\\\"",
		"\\", "\\\\",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	).Replace(text)
}

func (f *Formatter) len() int {
	n := f.buffer.Len() + len(f.Border)
	if f.Prefix != "" {
		n += 2
	}
	return n
}

func (f *Formatter) writeLine(line string) {
	if f.Width > 0 {
		for _, word := range strings.SplitAfter(line, " ") {
			if f.len()+len(word) > f.Width {
				f.flush()
			}
			f.mustWrite(f.buffer, word)
		}
	} else {
		f.mustWrite(f.buffer, line)
	}
	f.flush()
}

func (f *Formatter) flush() {
	if f.buffer.Len() == 0 && f.Width > 0 {
		return
	}
	if f.buffer.Len() == 0 && f.Prefix == "" {
		f.mustWrite(f.output, strings.TrimRight(f.Border, " "))
		f.mustWrite(f.output, "\n")
		return
	}
	f.mustWrite(f.output, f.Border)
	if f.Prefix != "" {
		f.mustWrite(f.output, `"`)
	}
	f.mustWrite(f.output, f.buffer.String())
	if f.Prefix != "" {
		f.mustWrite(f.output, `"`)
	}
	f.mustWrite(f.output, "\n")
	f.buffer.Reset()
}

func (Formatter) mustWrite(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		panic(errors.WithStack(err))
	}
}
