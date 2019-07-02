package pogo

import (
	"io"

	"github.com/pkg/errors"
	"github.com/vporoshok/muzzy"
)

// DefaultWidth of msgctxt, msgid and msgstr
const DefaultWidth = 80

// File represents an po file
type File struct {
	Header  Header
	Entries []Entry
}

// ReadFile from reader
func ReadFile(r io.Reader) (*File, error) {
	file := &File{}

	s := NewScanner(r)
	first := true
	for {
		entry, err := ReadEntry(s, file.Header.PluralForms.Len())
		if err != nil && errors.Cause(err) != io.EOF {

			return nil, err
		}
		if entry.MsgID != "" {
			file.Entries = append(file.Entries, entry)
		} else if first {
			file.Header.FromEntry(&entry)
		}
		first = false
		if err != nil {

			return file, nil
		}
	}
}

// Update file with next version of file
func (file *File) Update(next *File) *File {
	recycle := make([]bool, len(file.Entries))
	index := muzzy.NewSplitIndex(muzzy.NGramSplitter(3, true))
	entryID := func(entry Entry) string {
		if entry.MsgIDP != "" {
			return entry.MsgID + "  \x00  " + entry.MsgIDP
		}
		return entry.MsgID
	}

	for i := range file.Entries {
		index.Add(entryID(file.Entries[i]))
	}

	res := make([]Entry, len(next.Entries))
	for i := range next.Entries {
		j := index.Search(entryID(next.Entries[i]))
		if j >= 0 {
			d := index.Similarity(entryID(file.Entries[j]), entryID(next.Entries[i]))
			if d > 0.8 {
				res[i] = file.Entries[j].Update(&next.Entries[i])
				recycle[j] = true
				continue
			}
		}
		res[i] = next.Entries[i]
	}
	for i, ok := range recycle {
		if !ok {
			entry := file.Entries[i]
			entry.Obsolete = true
			res = append(res, entry)
		}
	}

	return &File{
		Header:  file.Header,
		Entries: res,
	}
}

// Print file to writer
func (file *File) Print(w io.Writer) error {
	f := NewFormatter(w)
	header := file.Header.ToEntry()
	if err := header.Print(f, DefaultWidth); err != nil {

		return err
	}
	for i := range file.Entries {
		if err := f.BreakLine(); err != nil {

			return err
		}
		if err := file.Entries[i].Print(f, DefaultWidth); err != nil {

			return err
		}
	}

	return nil
}
