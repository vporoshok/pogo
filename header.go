package pogo

import "time"

// Person present an person
type Person struct {
	Name  string
	Email string
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
	PluralForms             []string
}
