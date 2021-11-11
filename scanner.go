package pogo

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Scanner to extract blocks by permitted starters
type Scanner struct {
	// Starters is a set of permitted starters
	Starters []Starter
	// Border of the last read block
	Border string
	// Prefix of the last read block
	Prefix string
	// Buffer of text of the last read block
	Buffer *bytes.Buffer
	// Line is current line number
	Line int

	input *bufio.Scanner
}

// ScannerOption customize scanner
type ScannerOption func(*Scanner)

// WithBuffer replace internal bufio.Scanner buffer
func WithBuffer(buf []byte, max int) ScannerOption {
	return func(s *Scanner) {
		s.input.Buffer(buf, max)
	}
}

// NewScanner to read from r
func NewScanner(r io.Reader, opts ...ScannerOption) *Scanner {
	s := &Scanner{
		Buffer: &bytes.Buffer{},
		input:  bufio.NewScanner(r),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// IsBlankLine return true if current line is blank
func (s *Scanner) IsBlankLine() bool {
	return s.input.Text() == ""
}

// Scan next block
func (s *Scanner) Scan() error {
	return recoverHandledError(s.mustScan)
}

func (s *Scanner) mustScan() {
	s.Border, s.Prefix = "", ""
	s.Buffer.Reset()
	s.skipBlankLines()
	s.start()
	s.mustReadLine(len(s.Border) + len(s.Prefix))
	border := s.Border
	if s.Prefix != "" {
		border += `"`
	}
	for s.input.Scan() {
		s.Line++
		if s.input.Text() == "" || !strings.HasPrefix(s.input.Text(), border) {
			return
		}
		s.mustReadLine(len(s.Border))
	}
	err := s.input.Err()
	if err == nil {
		err = io.EOF
	}
	panic(errors.WithStack(err))
}

func (s *Scanner) skipBlankLines() {
	for s.input.Text() == "" {
		s.Line++
		if !s.input.Scan() {
			if s.input.Err() == nil {
				panic(errors.WithStack(io.EOF))
			}
			panic(errors.WithStack(s.input.Err()))
		}
	}
}

func (s *Scanner) start() {
	for _, starter := range s.Starters {
		if border, prefix, ok := starter.Extract(s.input.Text()); ok {
			s.Border = border
			s.Prefix = prefix
			break
		}
	}
	if s.Border == "" && s.Prefix == "" {
		panic(errors.Errorf("no starter is matched line %d", s.Line))
	}
}

func (s *Scanner) mustReadLine(skip int) {
	line := s.input.Text()[skip:]
	if s.Prefix == "" {
		if s.Buffer.Len() > 0 {
			s.mustWrite("\n")
		}
		s.mustWrite(line)
		return
	}
	unquoted, err := strconv.Unquote(strings.TrimSpace(line))
	if err != nil {
		panic(errors.WithStack(err))
	}
	s.mustWrite(s.unescape(unquoted))
}

func (s *Scanner) mustWrite(text string) {
	if _, err := s.Buffer.WriteString(text); err != nil {
		panic(errors.WithStack(err))
	}
}

func (Scanner) unescape(text string) string {
	return strings.NewReplacer(
		"\\\\", "\\",
		"\\n", "\n",
		"\\r", "\r",
		"\\t", "\t",
	).Replace(text)
}
