package pogo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04Z07:00"
	personRER  = `([^<]+) <([^>]+)>`
)

var (
	personRE    = regexp.MustCompile(personRER)
	copyrightRE = regexp.MustCompile(`^Copyright \(C\) (.+)`)
	packageRE   = regexp.MustCompile(`^This file is distributed under the same license as the (.+) package\.$`)
	authorsRE   = regexp.MustCompile(fmt.Sprintf(`^(%s), ((\d+,\s+)*\d+)\.$`, personRER))
)

// Person present an person
type Person struct {
	Name  string
	Email string
}

// Parse person string
func (person *Person) Parse(text string) {
	sub := personRE.FindStringSubmatch(text)
	if len(sub) == 3 {
		person.Name = sub[1]
		person.Email = sub[2]
	}
}

// String implements fmt.Stringer
func (person Person) String() string {
	return fmt.Sprintf("%s <%s>", person.Name, person.Email)
}

// Header of po file
type Header struct {
	Title          string
	Copyright      string
	PackageLicense string
	Authors        []struct {
		Person
		Years []int
	}
	Fuzzy                   bool
	ProjectIDVersion        string
	ReportMsgidBugsTo       string
	POTCreationDate         time.Time
	PORevisionDate          time.Time
	LastTranslator          Person
	LanguageTeam            string
	Language                string
	ContentType             string
	ContentTransferEncoding string
	Unknown                 [][2]string
	PluralForms             PluralRules
}

// FromEntry parse entry as header
func (header *Header) FromEntry(entry *POEntry) {
	header.parseEntryComment(entry.TComment)
	header.Fuzzy = entry.Flags.Contain("fuzzy")
	header.parseEntryMsgStr(entry.MsgStr)
	if header.ContentType == "" {
		header.ContentType = "text/plain; charset=UTF-8"
	}
	if header.ContentTransferEncoding == "" {
		header.ContentTransferEncoding = "8bit"
	}
}

func (header *Header) parseEntryComment(comment string) {
	for _, line := range strings.Split(comment, "\n") {
		switch {
		default:
			header.Title = strings.TrimRight(line, ".")
		case copyrightRE.MatchString(line):
			sub := copyrightRE.FindStringSubmatch(line)
			header.Copyright = sub[1]
		case packageRE.MatchString(line):
			sub := packageRE.FindStringSubmatch(line)
			header.PackageLicense = sub[1]
		case authorsRE.MatchString(line):
			sub := authorsRE.FindStringSubmatch(line)
			var person Person
			person.Parse(sub[1])
			yearsRaw := strings.Split(sub[len(sub)-2], ",")
			years := make([]int, len(yearsRaw))
			for i := range years {
				years[i], _ = strconv.Atoi(strings.TrimSpace(yearsRaw[i]))
			}
			header.Authors = append(header.Authors, struct {
				Person
				Years []int
			}{person, years})
		}
	}
}

func (header *Header) parseEntryMsgStr(text string) {
	for _, line := range strings.Split(text, "\n") {
		split := strings.SplitN(line, ":", 2)
		if len(split) != 2 {
			continue
		}
		header.parseKeyValue(split[0], strings.TrimSpace(split[1]))
	}
}

//nolint:gocyclo // it should be in one place
func (header *Header) parseKeyValue(key, val string) {
	switch key {
	case "Project-Id-Version":
		header.ProjectIDVersion = val
	case "Report-Msgid-Bugs-To":
		header.ReportMsgidBugsTo = val
	case "POT-Creation-Date":
		header.POTCreationDate, _ = time.Parse(timeFormat, val)
	case "PO-Revision-Date":
		header.PORevisionDate, _ = time.Parse(timeFormat, val)
	case "Last-Translator":
		header.LastTranslator.Parse(val)
	case "Language-Team":
		header.LanguageTeam = val
	case "Language":
		header.Language = val
	case "Content-Type":
		header.ContentType = val
	case "Content-Transfer-Encoding":
		header.ContentTransferEncoding = val
	case "Plural-Forms":
		header.PluralForms, _ = ParsePluralRules(val)
	default:
		header.Unknown = append(header.Unknown, [2]string{key, val})
	}
}

// ToEntry format header to entry fields
func (header *Header) ToEntry() POEntry {
	entry := POEntry{}
	entry.TComment = header.getEntryComment()
	if header.Fuzzy {
		entry.Flags.Add("fuzzy")
	}
	entry.MsgStr = header.getEntryMsgStr()

	return entry
}

func (header *Header) getEntryComment() string {
	res := &strings.Builder{}
	_, _ = fmt.Fprintf(res, "%s.\n", header.Title)
	if header.Copyright != "" {
		_, _ = fmt.Fprintf(res, "Copyright (C) %s\n", header.Copyright)
	}
	if header.PackageLicense != "" {
		_, _ = fmt.Fprintf(res, "This file is distributed under the same license as the %s package.\n", header.PackageLicense)
	}
	for i := range header.Authors {
		_, _ = fmt.Fprint(res, header.Authors[i].Person.String())
		for _, year := range header.Authors[i].Years {
			_, _ = fmt.Fprintf(res, ", %d", year)
		}
		_, _ = fmt.Fprint(res, ".\n")
	}

	return strings.TrimSpace(res.String())
}

func (header *Header) getEntryMsgStr() string {
	res := &strings.Builder{}
	_, _ = fmt.Fprintln(res, "Project-Id-Version:", header.ProjectIDVersion)
	_, _ = fmt.Fprintln(res, "Report-Msgid-Bugs-To:", header.ReportMsgidBugsTo)
	_, _ = fmt.Fprintln(res, "POT-Creation-Date:", header.POTCreationDate.Format(timeFormat))
	_, _ = fmt.Fprintln(res, "PO-Revision-Date:", header.PORevisionDate.Format(timeFormat))
	_, _ = fmt.Fprintln(res, "Last-Translator:", header.LastTranslator)
	_, _ = fmt.Fprintln(res, "Language-Team:", header.LanguageTeam)
	_, _ = fmt.Fprintln(res, "Language:", header.Language)
	_, _ = fmt.Fprintln(res, "Content-Type:", header.ContentType)
	_, _ = fmt.Fprintln(res, "Content-Transfer-Encoding:", header.ContentTransferEncoding)
	_, _ = fmt.Fprintln(res, "Plural-Forms:", header.PluralForms)
	for i := range header.Unknown {
		_, _ = fmt.Fprintf(res, "%s: %s\n", header.Unknown[i][0], header.Unknown[i][1])
	}

	return res.String()
}
