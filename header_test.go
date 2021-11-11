package pogo_test

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vporoshok/pogo"
)

func TestHeader(t *testing.T) {
	t.Parallel()

	join := func(lines ...string) string {
		return strings.Join(lines, "\n")
	}

	s := pogo.NewScanner(bytes.NewBufferString(join(
		`# Translation of kstars.po into Spanish.`,
		`# Copyright (C) 2008 None`,
		`# This file is distributed under the same license as the kdeedu package.`,
		`# Pablo de Vicente <pablo@foo.com>, 2005, 2006, 2007, 2008.`,
		`# Eloy Cuadra <eloy@bar.net>, 2007, 2008.`,
		`#, fuzzy`,
		`msgid ""`,
		`msgstr ""`,
		`"Project-Id-Version: kstars\n"`,
		`"Report-Msgid-Bugs-To: http://bugs.kde.org\n"`,
		`"POT-Creation-Date: 2008-09-01 09:37Z\n"`,
		`"PO-Revision-Date: 2008-07-22 18:13Z\n"`,
		`"Last-Translator: Eloy Cuadra <eloy@bar.net>\n"`,
		`"Language-Team: Spanish <kde-l10n-es@kde.org>\n"`,
		`"Language: es_ES\n"`,
		`"Content-Type: text/plain; charset=UTF-8\n"`,
		`"Content-Transfer-Encoding: 8bit\n"`,
		`"Plural-Forms: nplurals=3; plural=n%10 == 1 && n%100 != 11 ? 0 : n%10 >= 2 && n"`,
		`"%10 <= 4 && (n%100 < 10 || n%100 >= 20) ? 1 : 2;\n"`,
		`"MIME-Version: 1.0\n"`,
	)))

	entry, err := pogo.ReadPOEntry(s, 0)
	require.Equal(t, io.EOF, errors.Cause(err))

	header := pogo.Header{}
	header.FromEntry(&entry)
	assert.Equal(t, "Translation of kstars.po into Spanish", header.Title)
	assert.Equal(t, "2008 None", header.Copyright)
	assert.Equal(t, "kdeedu", header.PackageLicense)
	require.Len(t, header.Authors, 2)
	assert.Equal(t, []struct {
		pogo.Person
		Years []int
	}{
		{
			pogo.Person{"Pablo de Vicente", "pablo@foo.com"},
			[]int{2005, 2006, 2007, 2008},
		},
		{
			pogo.Person{"Eloy Cuadra", "eloy@bar.net"},
			[]int{2007, 2008},
		},
	},
		header.Authors,
	)
	assert.True(t, header.Fuzzy)
	assert.Equal(t, "kstars", header.ProjectIDVersion)
	assert.Equal(t, "http://bugs.kde.org", header.ReportMsgidBugsTo)
	assert.Equal(t, time.Date(2008, time.September, 1, 9, 37, 0, 0, time.UTC), header.POTCreationDate)
	assert.Equal(t, time.Date(2008, time.July, 22, 18, 13, 0, 0, time.UTC), header.PORevisionDate)
	assert.Equal(t, pogo.Person{"Eloy Cuadra", "eloy@bar.net"}, header.LastTranslator)
	assert.Equal(t, "Spanish <kde-l10n-es@kde.org>", header.LanguageTeam)
	assert.Equal(t, "es_ES", header.Language)
	assert.Equal(t, "text/plain; charset=UTF-8", header.ContentType)
	assert.Equal(t, "8bit", header.ContentTransferEncoding)
	assert.Equal(t, [][2]string{{"MIME-Version", "1.0"}}, header.Unknown)

	newEntry := header.ToEntry()
	assert.Equal(t, entry, newEntry)
}
