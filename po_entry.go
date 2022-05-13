package pogo

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const prevBorder = "#| "

// POEntry present one message
type POEntry struct {
	TComment    string
	EComment    string
	Reference   string
	Flags       Flags
	PrevMsgCtxt string
	PrevMsgID   string
	PrevMsgIDP  string
	MsgCtxt     string
	MsgID       string
	MsgIDP      string
	MsgStr      string
	MsgStrP     []string
	Obsolete    bool
}

var poStarters = []Starter{
	// EComment
	NewPlainStarter("#. ", ""),
	// Reference
	NewPlainStarter("#: ", ""),
	// Flags
	NewPlainStarter("#, ", ""),
	// PrevMsgCtxt
	NewPlainStarter(prevBorder, "msgctxt "),
	// PrevMsgID
	NewPlainStarter(prevBorder, "msgid "),
	// PrevMsgIDP
	NewPlainStarter(prevBorder, "msgid_plural "),
	// MsgCtxt
	NewPlainStarter("", "msgctxt "),
	NewPlainStarter("#~ ", "msgctxt "),
	// MsgID
	NewPlainStarter("", "msgid "),
	NewPlainStarter("#~ ", "msgid "),
	// MsgIDP
	NewPlainStarter("", "msgid_plural "),
	NewPlainStarter("#~ ", "msgid_plural "),
	// MsgStr
	NewPlainStarter("", "msgstr "),
	NewPlainStarter("#~ ", "msgstr "),
	// MsgStrP
	NewRegexpStarter("", `msgstr\[\d+\] `),
	NewRegexpStarter("#~ ", `msgstr\[\d+\] `),
	// TComment
	NewPlainStarter("# ", ""),
	NewPlainStarter("#", ""), // should be last in this list, because it match other starters
}

// ReadPOEntry from scanner
func ReadPOEntry(s *Scanner, pluralCount int) (entry POEntry, err error) {
	s.Starters = poStarters
	for {
		err = s.Scan()
		if err != nil && errors.Cause(err) != io.EOF {
			return
		}
		if applyErr := entry.applyBlock(s, pluralCount); applyErr != nil {
			return entry, applyErr
		}
		if err != nil || s.IsBlankLine() {
			return
		}
	}
}

func (entry *POEntry) applyBlock(s *Scanner, pluralCount int) (err error) {
	return recoverHandledError(func() {
		entry.mustApplyBlock(s, pluralCount)
	})
}

//nolint:gocyclo // it should be at one place
func (entry *POEntry) mustApplyBlock(s *Scanner, pluralCount int) {
	entry.checkObsolete(s)

	switch [2]string{s.Border, s.Prefix} {
	case [2]string{"#", ""}, [2]string{"# ", ""}:
		if entry.TComment != "" {
			entry.TComment += "\n"
		}
		entry.TComment += s.Buffer.String()
	case [2]string{"#. ", ""}:
		entry.mustBeEmpty(s, entry.EComment)
		entry.EComment = s.Buffer.String()
	case [2]string{"#: ", ""}:
		entry.mustBeEmpty(s, entry.Reference)
		entry.Reference = s.Buffer.String()
	case [2]string{"#, ", ""}:
		entry.mustBeEmpty(s, entry.Flags.String())
		entry.Flags.Parse(s.Buffer.String())
	case [2]string{prevBorder, "msgctxt "}:
		entry.mustBeEmpty(s, entry.PrevMsgCtxt)
		entry.PrevMsgCtxt = s.Buffer.String()
	case [2]string{prevBorder, "msgid "}:
		entry.mustBeEmpty(s, entry.PrevMsgID)
		entry.PrevMsgID = s.Buffer.String()
	case [2]string{prevBorder, "msgid_plural "}:
		entry.mustBeEmpty(s, entry.PrevMsgIDP)
		entry.PrevMsgIDP = s.Buffer.String()
	case [2]string{"", "msgctxt "},
		[2]string{"#~ ", "msgctxt "}:
		entry.mustBeEmpty(s, entry.MsgCtxt)
		entry.MsgCtxt = s.Buffer.String()
	case [2]string{"", "msgid "},
		[2]string{"#~ ", "msgid "}:
		entry.mustBeEmpty(s, entry.MsgID)
		entry.MsgID = s.Buffer.String()
	case [2]string{"", "msgid_plural "},
		[2]string{"#~ ", "msgid_plural "}:
		entry.mustBeEmpty(s, entry.MsgIDP)
		entry.MsgIDP = s.Buffer.String()
	case [2]string{"", "msgstr "},
		[2]string{"#~ ", "msgstr "}:
		entry.mustBeEmpty(s, entry.MsgStr)
		entry.MsgStr = s.Buffer.String()
	default:
		entry.updateMsgStrP(s, pluralCount)
	}
}

func (entry *POEntry) checkObsolete(s *Scanner) {
	if s.Border == "#~ " {
		if entry.Obsolete {
			return
		}
		if entry.MsgCtxt != "" ||
			entry.MsgID != "" ||
			entry.MsgIDP != "" ||
			entry.MsgStr != "" ||
			entry.MsgStrP != nil {
			panic(errors.Errorf("mixed obsolete and not obsolete blocks at %d", s.Line-1))
		}
		entry.Obsolete = true
	}
	if s.Border == "" {
		if entry.Obsolete {
			panic(errors.Errorf("mixed obsolete and not obsolete blocks at %d", s.Line-1))
		}
	}
}

func (entry *POEntry) updateMsgStrP(s *Scanner, pluralCount int) {
	if !strings.HasPrefix(s.Prefix, "msgstr[") {
		fmt.Println(s.Prefix, s.Border, s.Buffer.String())
		return
	}
	if entry.MsgStrP == nil {
		entry.MsgStrP = make([]string, pluralCount)
	}
	n, err := strconv.Atoi(s.Prefix[7 : len(s.Prefix)-2])
	if err != nil {
		panic(errors.WithStack(err))
	}
	if n >= pluralCount {
		panic(errors.Errorf("unknown plural form %d at %d", n, s.Line-1))
	}
	entry.mustBeEmpty(s, entry.MsgStrP[n])
	entry.MsgStrP[n] = s.Buffer.String()
}

func (POEntry) mustBeEmpty(s *Scanner, text string) {
	if text != "" {
		panic(errors.Errorf("duplicate block %q at %d", s.Border+s.Prefix, s.Line))
	}
}

// Print entry in PO format
func (entry *POEntry) Print(f *Formatter, width int) error {
	return recoverHandledError(func() {
		entry.mustPrint(f, width)
	})
}

//nolint:gocyclo // it ahould be at one place
func (entry *POEntry) mustPrint(f *Formatter, width int) {
	mustFormat := func(text string) {
		if err := f.Format(text); err != nil {
			panic(err)
		}
	}

	f.Width = 0
	if entry.TComment != "" {
		f.Border, f.Prefix = "# ", ""
		mustFormat(entry.TComment)
	}
	if entry.EComment != "" {
		f.Border, f.Prefix = "#. ", ""
		mustFormat(entry.EComment)
	}
	if entry.Reference != "" {
		f.Border, f.Prefix = "#: ", ""
		mustFormat(entry.Reference)
	}
	if len(entry.Flags) > 0 {
		f.Border, f.Prefix = "#, ", ""
		mustFormat(entry.Flags.String())
	}
	if entry.PrevMsgCtxt != "" {
		f.Border, f.Prefix = prevBorder, "msgctxt "
		mustFormat(entry.PrevMsgCtxt)
	}
	if entry.PrevMsgID != "" {
		f.Border, f.Prefix = prevBorder, "msgid "
		mustFormat(entry.PrevMsgID)
	}
	if entry.PrevMsgIDP != "" {
		f.Border, f.Prefix = prevBorder, "msgid_plural "
		mustFormat(entry.PrevMsgIDP)
	}

	f.Border = ""
	f.Width = width
	if entry.Obsolete {
		f.Border = "#~ "
	}
	if entry.MsgCtxt != "" {
		f.Prefix = "msgctxt "
		mustFormat(entry.MsgCtxt)
	}
	f.Prefix = "msgid "
	mustFormat(entry.MsgID)
	if entry.MsgIDP != "" {
		f.Prefix = "msgid_plural "
		mustFormat(entry.MsgIDP)
	}
	if len(entry.MsgStrP) == 0 {
		f.Prefix = "msgstr "
		mustFormat(entry.MsgStr)
	}
	for i := range entry.MsgStrP {
		f.Prefix = fmt.Sprintf("msgstr[%d] ", i)
		mustFormat(entry.MsgStrP[i])
	}
}

// Update return merge result of entry with next version
func (entry *POEntry) Update(next *POEntry) POEntry {
	res := *entry
	res.EComment = next.EComment
	if res.MsgCtxt != next.MsgCtxt {
		res.PrevMsgCtxt, res.MsgCtxt = res.MsgCtxt, next.MsgCtxt
		res.Flags.Add("fuzzy")
	}
	if res.MsgID != next.MsgID {
		res.PrevMsgID, res.MsgID = res.MsgID, next.MsgID
		res.Flags.Add("fuzzy")
	}
	if res.MsgIDP != next.MsgIDP {
		res.PrevMsgIDP, res.MsgIDP = res.MsgIDP, next.MsgIDP
		res.Flags.Add("fuzzy")
	}

	return res
}
