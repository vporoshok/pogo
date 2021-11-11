package pogo

import (
	"io"

	"github.com/pkg/errors"
	"github.com/vporoshok/muzzy"
)

// DefaultWidth of msgctxt, msgid and msgstr
const DefaultWidth = 80

// POFile represents an po file
type POFile struct {
	Header
	Entries []POEntry
}

// ReadPOFile from reader
func ReadPOFile(r io.Reader, opts ...ScannerOption) (*POFile, error) {
	po := &POFile{}

	s := NewScanner(r, opts...)
	first := true
	for {
		entry, err := ReadPOEntry(s, po.PluralForms.Len())
		if err != nil && errors.Cause(err) != io.EOF {
			return nil, err
		}
		if entry.MsgID != "" {
			po.Entries = append(po.Entries, entry)
		} else if first {
			po.Header.FromEntry(&entry)
		}
		first = false
		if err != nil {
			return po, nil
		}
	}
}

// Update file with next version
func (po *POFile) Update(next *POFile) *POFile {
	recycle := make([]bool, len(po.Entries))
	index := muzzy.NewSplitIndex(muzzy.NGramSplitter(3, true))
	entryID := func(entry POEntry) string {
		if entry.MsgIDP != "" {
			return entry.MsgID + "  \x00  " + entry.MsgIDP
		}
		return entry.MsgID
	}

	for i := range po.Entries {
		index.Add(entryID(po.Entries[i]))
	}

	res := make([]POEntry, len(next.Entries))
	for i := range next.Entries {
		j := index.Search(entryID(next.Entries[i]))
		if j >= 0 {
			d := index.Similarity(entryID(po.Entries[j]), entryID(next.Entries[i]))
			if d > 0.8 {
				res[i] = po.Entries[j].Update(&next.Entries[i])
				recycle[j] = true
				continue
			}
		}
		res[i] = next.Entries[i]
	}
	for i, ok := range recycle {
		if !ok {
			entry := po.Entries[i]
			if !entry.Obsolete {
				entry.Obsolete = true
				res = append(res, entry)
			}
		}
	}

	return &POFile{
		Header:  next.Header,
		Entries: res,
	}
}

// Print po-file to writer
func (po *POFile) Print(w io.Writer) error {
	f := NewFormatter(w)
	header := po.Header.ToEntry()
	if err := header.Print(f, DefaultWidth); err != nil {
		return err
	}
	for i := range po.Entries {
		if err := f.BreakLine(); err != nil {
			return err
		}
		if err := po.Entries[i].Print(f, DefaultWidth); err != nil {
			return err
		}
	}

	return nil
}

// MO convert to mo-file
func (po *POFile) MO() *MOFile {
	mo := &MOFile{
		Header:  po.Header,
		Entries: make(map[string][]string, len(po.Entries)),
	}
	for i := range po.Entries {
		id := po.Entries[i].MsgID
		if po.Entries[i].MsgCtxt != "" {
			id = po.Entries[i].MsgCtxt + ctxtSep + id
		}
		val := []string{po.Entries[i].MsgStr}
		if po.Entries[i].MsgStrP != nil {
			id += pluralSep + po.Entries[i].MsgIDP
			val = make([]string, len(po.Entries[i].MsgStrP))
			copy(val, po.Entries[i].MsgStrP)
		}
		mo.Entries[id] = val
	}

	return mo
}
