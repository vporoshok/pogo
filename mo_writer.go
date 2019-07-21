package pogo

import (
	"encoding/binary"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type moWriter struct {
	w io.Writer

	file         *MOFile
	Originals    []string
	Translations []string
}

func (mw *moWriter) Write() error {
	return recoverHandledError(mw.mustWrite)
}

func (mw *moWriter) mustWrite() {
	mw.prepareResources()

	mw.mustWriteUint32(magic)
	mw.mustWriteUint32(0)          // format revision
	N := uint32(len(mw.Originals)) // number of strings
	mw.mustWriteUint32(N)
	O := uint32(28) // offset of original table (length of fixed headers)
	mw.mustWriteUint32(O)
	T := O + N*8 // offset of translations table
	mw.mustWriteUint32(T)
	S := uint32(0) // size of hashing table (not implemented)
	mw.mustWriteUint32(S)
	H := T + N*8 // offset of hashing table
	mw.mustWriteUint32(H)
	offset := H
	offset = mw.mustWritePositions(offset, mw.Originals)
	mw.mustWritePositions(offset, mw.Translations)
	mw.mustWriteStrings(mw.Originals)
	mw.mustWriteStrings(mw.Translations)
}

func (mw *moWriter) prepareResources() {
	mw.Originals = make([]string, 1, len(mw.file.Entries)+1)
	mw.Originals[0] = "" // header entry
	for id := range mw.file.Entries {
		if id != "" {
			mw.Originals = append(mw.Originals, id)
		}
	}
	sort.Strings(mw.Originals)
	header := mw.file.Header.ToEntry()
	mw.Translations = make([]string, len(mw.Originals))
	for i, id := range mw.Originals {
		if id == "" {
			mw.Translations[i] = header.MsgStr
		} else {
			mw.Translations[i] = strings.Join(mw.file.Entries[id], pluralSep)
		}
	}
}

func (mw *moWriter) mustWritePositions(offset uint32, data []string) uint32 {
	for _, s := range data {
		n := uint32(len(s))
		mw.mustWriteUint32(n)
		mw.mustWriteUint32(offset)
		offset += n + 1
	}

	return offset
}

func (mw *moWriter) mustWriteUint32(x uint32) {
	if err := binary.Write(mw.w, binary.LittleEndian, x); err != nil {
		panic(errors.WithStack(err))
	}
}

func (mw *moWriter) mustWriteStrings(data []string) {
	for _, x := range data {
		mw.mustWriteString(x)
	}
}

func (mw *moWriter) mustWriteString(x string) {
	if _, err := mw.w.Write([]byte(x)); err != nil {
		panic(errors.WithStack(err))
	}
	if _, err := mw.w.Write([]byte{'\x00'}); err != nil {
		panic(errors.WithStack(err))
	}
}
