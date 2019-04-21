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

// NewScanner to read from r
func NewScanner(r io.Reader) *Scanner {

	return &Scanner{
		Buffer: &bytes.Buffer{},
		input:  bufio.NewScanner(r),
	}
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
		panic(errors.Errorf("no starter is matched line: %s", s.input.Text()))
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

// // ReadEntry from stream
// func ReadEntry(r *bufio.Reader) (Entry, error) {
// 	er := &entryReader{}
// 	err := er.Read(r)

// 	return er.entry, err
// }

// type entryReader struct {
// 	kind    string
// 	builder *strings.Builder
// 	entry   Entry
// }

// func (er *entryReader) Read(r *bufio.Reader) error {
// 	er.builder = &strings.Builder{}
// 	for {
// 		line, err := r.ReadBytes('\n')
// 		if err == io.EOF {
// 			er.Flush()

// 			return io.EOF
// 		}
// 		if err != nil {

// 			return errors.WithStack(err)
// 		}
// 		line = bytes.TrimSpace(line)
// 		if len(line) == 0 {
// 			if er.kind == "" {
// 				continue
// 			}
// 			er.Flush()

// 			return nil
// 		}
// 		if err = er.ParseLine(string(line)); err != nil {

// 			return err
// 		}
// 	}
// }

// var prefixes = [...]struct {
// 	prefix    string
// 	kind      string
// 	unquote   bool
// 	breakLine bool
// }{
// 	{prefix: "# ", kind: "TComment", unquote: false, breakLine: true},
// 	{prefix: "#. ", kind: "EComment", unquote: false, breakLine: true},
// 	{prefix: "#: ", kind: "Reference", unquote: false, breakLine: false},
// 	{prefix: "#, ", kind: "Flags", unquote: false, breakLine: false},
// 	{prefix: "#| msgctxt ", kind: "PrevMsgCtxt", unquote: true, breakLine: false},
// 	{prefix: "#| msgid ", kind: "PrevMsgID", unquote: true, breakLine: false},
// 	{prefix: "#| msgid_plural ", kind: "PrevMsgIDP", unquote: true, breakLine: false},
// 	{prefix: "msgctxt ", kind: "MsgCtxt", unquote: true, breakLine: false},
// 	{prefix: "msgid ", kind: "MsgID", unquote: true, breakLine: false},
// 	{prefix: "msgid_plural ", kind: "MsgIDP", unquote: true, breakLine: false},
// 	{prefix: "msgstr ", kind: "MsgStr", unquote: true, breakLine: false},
// }

// var pluralStr = regexp.MustCompile(`^msgstr\[(%d+)\]`)

// func (er *entryReader) ParseLine(line string) error {
// 	for _, opt := range prefixes {
// 		if strings.HasPrefix(line, opt.prefix) {

// 			return er.AddLine(opt.kind, strings.TrimPrefix(line, opt.prefix), opt.unquote, opt.breakLine)
// 		}
// 	}
// 	if pluralStr.MatchString(line) {
// 		matches := pluralStr.FindAllStringSubmatch(line, 1)
// 		num := matches[0][1]

// 		return er.AddLine(fmt.Sprintf("MsgStrP.%s", num), line[strings.Index(line, " "):], true, false)
// 	}
// 	switch true {
// 	case strings.HasPrefix(line, `#| "`):
// 		line = strings.TrimPrefix(line, "#| ")
// 		fallthrough
// 	case strings.HasPrefix(line, `"`):

// 		return er.AddLine(er.kind, line, true, false)
// 	}

// 	return errors.Errorf("unexpected line %q", line)
// }

// func (er *entryReader) AddLine(kind, line string, unquote, withNewLine bool) error {
// 	if kind == "" {

// 		return errors.New("attempt to add line with no kind")
// 	}
// 	if kind != er.kind {
// 		er.Flush()
// 		er.kind = kind
// 	}
// 	if withNewLine && er.builder.Len() > 0 {
// 		if err := er.builder.WriteByte('\n'); err != nil {

// 			return errors.WithStack(err)
// 		}
// 	}
// 	if unquote {
// 		var err error

// 		if line, err = strconv.Unquote(strings.TrimSpace(line)); err != nil {

// 			return errors.WithStack(err)
// 		}
// 	}

// 	_, err := er.builder.WriteString(line)

// 	return errors.WithStack(err)
// }

// // nolint:gocyclo
// func (er *entryReader) Flush() {
// 	kinds := strings.Split(er.kind, "#")
// 	switch kinds[0] {
// 	case "TComment":
// 		er.entry.TComment = er.builder.String()
// 	case "EComment":
// 		er.entry.EComment = er.builder.String()
// 	case "Reference":
// 		er.entry.Reference = er.builder.String()
// 	case "Flags":
// 		flags := strings.Split(er.builder.String(), ",")
// 		for i := range flags {
// 			flags[i] = strings.TrimSpace(flags[i])
// 		}
// 		er.entry.Flags = flags
// 	case "PrevMsgCtxt":
// 		er.entry.PrevMsgCtxt = er.builder.String()
// 	case "PrevMsgID":
// 		er.entry.PrevMsgID = er.builder.String()
// 	case "PrevMsgIDP":
// 		er.entry.PrevMsgIDP = er.builder.String()
// 	case "MsgCtxt":
// 		er.entry.MsgCtxt = er.builder.String()
// 	case "MsgID":
// 		er.entry.MsgID = er.builder.String()
// 	case "MsgIDP":
// 		er.entry.MsgIDP = er.builder.String()
// 	case "MsgStr":
// 		er.entry.MsgStr = er.builder.String()
// 	case "MsgStrP":
// 		quantity, _ := strconv.Atoi(kinds[1])
// 		er.entry.MsgStrP[quantity] = er.builder.String()
// 	}
// 	er.builder.Reset()
// }
