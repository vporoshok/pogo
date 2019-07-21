package pogo

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

const (
	magic uint32 = 0x950412de
)

var errMOFile = errors.New("invalid mo file")

type position struct {
	Length, Offset uint32
}

type moReader struct {
	r      io.Reader
	offset uint32

	N            uint32     // number of strings
	O            uint32     // offset of table with original strings
	T            uint32     // offset of table with translation strings
	S            uint32     // size of hashing table
	H            uint32     // offset of hashing table
	OTable       []position // length and offset of original strings
	TTable       []position // length and offset of translated strings
	Originals    []string   // original strings
	Translations []string   // translated strings
}

func (mr *moReader) Read() error {
	err := recoverHandledError(mr.mustRead)

	return errors.Wrapf(err, "at %d", mr.offset)
}

func (mr *moReader) mustRead() {
	var (
		first, revision uint32
	)

	mr.mustReadUint32(&first)
	if first != magic {
		panic(errors.Wrap(errMOFile, "magic number mistmatch"))
	}
	mr.mustReadUint32(&revision)
	if revision != 0 {
		panic(errors.Wrapf(errMOFile, "unsupported format revision %d", revision))
	}

	mr.mustReadTables()

	mr.Originals = mr.mustReadStrings(mr.OTable)
	mr.Translations = mr.mustReadStrings(mr.TTable)
}

func (mr *moReader) mustReadTables() {
	mr.mustReadUint32(&mr.N)
	if mr.N == 0 {
		return
	}
	mr.mustReadUint32(&mr.O)
	if mr.O < 28 {
		panic(errors.Wrap(errMOFile, "bad original table offset"))
	}
	mr.mustReadUint32(&mr.T)
	if mr.T < mr.O+mr.N*8 {
		panic(errors.Wrap(errMOFile, "bad translation table offset"))
	}
	mr.mustReadUint32(&mr.S)
	mr.mustReadUint32(&mr.H)
	if mr.S > 0 && mr.H < mr.T+mr.N*8 {
		panic(errors.Wrap(errMOFile, "bad hashing table offset"))
	}
	mr.mustSeek(mr.O)
	mr.OTable = mr.mustReadPositionTable(mr.H + mr.S*4)
	last := mr.OTable[mr.N-1]
	mr.TTable = mr.mustReadPositionTable(last.Offset + last.Length + 1)
}

func (mr *moReader) mustReadPositionTable(offset uint32) []position {
	table := make([]position, mr.N)
	for i := 0; i < int(mr.N); i++ {
		mr.mustReadPosition(&table[i])
		if table[i].Offset < offset {
			panic(errors.Wrap(errMOFile, "bad offset in table"))
		}
		offset = table[i].Offset + table[i].Length + 1
	}

	return table
}

func (mr *moReader) mustReadPosition(pos *position) {
	if err := binary.Read(mr.r, binary.LittleEndian, pos); err != nil {
		panic(errors.WithStack(err))
	}
	mr.offset += 8
}

func (mr *moReader) mustReadStrings(table []position) []string {
	res := make([]string, len(table))
	for i, pos := range table {
		mr.mustSeek(pos.Offset)
		s := make([]byte, pos.Length+1)
		n, err := mr.r.Read(s)
		mr.offset += uint32(n)
		if err != nil {
			panic(errors.WithStack(err))
		}
		if s[pos.Length] != 0 {
			panic(errors.Wrap(errMOFile, "expected null byte"))
		}
		res[i] = string(s[:pos.Length])
	}

	return res
}

func (mr *moReader) mustReadUint32(x *uint32) {
	if err := binary.Read(mr.r, binary.LittleEndian, x); err != nil {
		panic(errors.WithStack(err))
	}
	mr.offset += 4
}

func (mr *moReader) mustSeek(offset uint32) {
	buf := make([]byte, offset-mr.offset)
	n, err := mr.r.Read(buf)
	mr.offset += uint32(n)
	if err != nil {
		panic(errors.WithStack(err))
	}
}
