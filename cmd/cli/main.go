package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/vporoshok/pogo"
)

var (
	// Version of package
	Version = "undefined"

	app = kingpin.New("pogo", "PO-files manipulations").
		Author("Bastrykov Evgeniy <vporoshok@gmail.com>").
		Version(Version)

	wc      = app.Command("wc", "Count resources words and symbols")
	wcFiles = newFileList(wc.Arg("files", "List of PO-files"))
)

type fileList []string

func (fl *fileList) Set(value string) error {
	*fl = append(*fl, value)
	return nil
}

func (*fileList) String() string {
	return "test"
}

func (*fileList) IsCumulative() bool {
	return true
}

func newFileList(s kingpin.Settings) (target *[]string) {
	target = new([]string)
	s.SetValue((*fileList)(target))
	return
}

func main() {
	// nolint:gocritic
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case wc.FullCommand():
		actionWC(*wcFiles)
	}
}

func actionWC(files []string) {
	var lineCount, wordCount, charCount int

	processLine := func(line string) {
		lineCount++
		wordCount += len(strings.Fields(line))
		charCount += len([]rune(line))
	}

	processFile := func(file string) {
		r, err := os.Open(file) // nolint:gosec
		app.FatalIfError(err, "fail to open file %q", file)
		defer func() {
			_ = r.Close()
		}()
		po, err := pogo.ReadPOFile(r)
		app.FatalIfError(err, "fail to parse file %q", file)
		for i := range po.Entries {
			if len(po.Entries[i].MsgStrP) > 0 {
				for j := range po.Entries[i].MsgStrP {
					processLine(po.Entries[i].MsgStrP[j])
				}
			} else {
				processLine(po.Entries[i].MsgStr)
			}
		}
	}

	for _, file := range files {
		processFile(file)
	}

	fmt.Println("Lines:", lineCount)
	fmt.Println("Words:", wordCount)
	fmt.Println("Chars:", charCount)
}
