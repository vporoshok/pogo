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
	authorsRE   = regexp.MustCompile(fmt.Sprintf(`^(%s), ((\d+,\s+)+\d+)\.`, personRER))
	pluralRE    = regexp.MustCompile(`^nplurals=\d+; plural=(.+);$`)
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
	Title     string
	Copyright string
	Package   string
	Authors   []struct {
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
	PluralForms             []string
}

// FromEntry parse entry as header
func (header *Header) FromEntry(entry *Entry) {
	header.parseEntryComment(entry.TComment)
	header.Fuzzy = entry.Flags.Contain("fuzzy")
	header.parseEntryMsgStr(entry.MsgStr)
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
			header.Package = sub[1]
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
		key, val := split[0], strings.TrimSpace(split[1])
		switch key {
		default:
			header.Unknown = append(header.Unknown, [2]string{key, val})
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
			fmt.Println(val)
			sub := pluralRE.FindStringSubmatch(val)
			for _, rule := range strings.Split(sub[1], ";") {
				header.PluralForms = append(header.PluralForms, strings.TrimSpace(rule))
			}
		}
	}
}

// ToEntry format header to entry fields
func (header *Header) ToEntry() Entry {
	entry := Entry{}

	return entry
}
