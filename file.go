package pogo

import (
	"io"

	"github.com/pkg/errors"
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
	existed := make(map[[3]string]int, len(file.Entries))
	for i := range file.Entries {
		key := [3]string{
			file.Entries[i].MsgCtxt,
			file.Entries[i].MsgID,
			file.Entries[i].MsgIDP,
		}
		existed[key] = i
	}

	res := make([]Entry, len(next.Entries))
	for i := range next.Entries {
		key := [3]string{
			next.Entries[i].MsgCtxt,
			next.Entries[i].MsgID,
			next.Entries[i].MsgIDP,
		}
		if j, ok := existed[key]; ok {
			res[i] = file.Entries[j].Update(&next.Entries[i])
			recycle[j] = true
		} else {
			res[i] = next.Entries[i]
		}
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
