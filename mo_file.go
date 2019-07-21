package pogo

import (
	"io"
	"strings"
)

const (
	ctxtSep   = "\x04"
	pluralSep = "\x00"
)

// MOFile represents an po file
type MOFile struct {
	Header
	Entries map[string][]string
}

// ReadMOFile from reader
func ReadMOFile(r io.Reader) (*MOFile, error) {
	mr := moReader{r: r}
	if err := mr.Read(); err != nil {
		return nil, err
	}

	file := &MOFile{
		Entries: make(map[string][]string, mr.N),
	}
	for i := 0; i < int(mr.N); i++ {
		id := mr.Originals[i]
		str := mr.Translations[i]
		if id == "" {
			file.Header.parseEntryMsgStr(str)
		} else {
			file.Entries[id] = strings.Split(str, pluralSep)
		}
	}

	return file, nil
}

// Print file to writer
func (file *MOFile) Write(w io.Writer) error {
	mw := moWriter{w: w, file: file}
	return mw.Write()
}

// Get translation by original
func (file *MOFile) Get(msg string) string {
	return file.Entries[msg][0]
}

// GetN plural translation by original
func (file *MOFile) GetN(msg, plural string, n int) string {
	i := file.PluralForms.Eval(n)
	id := msg + pluralSep + plural
	forms := file.Entries[id]
	if i >= len(forms) {
		return ""
	}
	return forms[i]
}

// GetCtxt translation by original and context
func (file *MOFile) GetCtxt(msg, ctxt string) string {
	id := ctxt + ctxtSep + msg
	return file.Entries[id][0]
}

// GetCtxtN plural translation by original and context
func (file *MOFile) GetCtxtN(msg, plural, ctxt string, n int) string {
	i := file.PluralForms.Eval(n)
	id := ctxt + ctxtSep + msg + pluralSep + plural
	forms := file.Entries[id]
	if i >= len(forms) {
		return ""
	}
	return forms[i]
}
