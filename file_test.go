package pogo_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vporoshok/pogo"
)

func TestFileMerge(t *testing.T) {
	t.Parallel()

	join := func(lines ...string) string { return strings.Join(lines, "\n") }

	header := join(
		`# Translation of pogo.po into Russian.`,
		`# Bastrykov Evgeniy <bastrykov@foo.com>, 2019.`,
		`msgid ""`,
		`msgstr ""`,
		`"Project-Id-Version: pogo\n"`,
		`"Report-Msgid-Bugs-To: http://github.com/vporoshok/pogo\n"`,
		`"POT-Creation-Date: 2019-04-24 09:37Z\n"`,
		`"PO-Revision-Date: 2019-04-24 18:13Z\n"`,
		`"Last-Translator: Bastrykov Evgeniy <bastrykov@foo.com>\n"`,
		`"Language-Team: \n"`,
		`"Language: ru_RU\n"`,
		`"Content-Type: text/plain; charset=UTF-8\n"`,
		`"Content-Transfer-Encoding: 8bit\n"`,
		`"Plural-Forms: nplurals=3; plural=n%10 == 1 && n%100 != 11 ? 0 : n%10 >= 2 && "`,
		`"n%10 <= 4 && (n%100 < 10 || n%100 >= 20) ? 1 : 2;\n"`,
		``,
	)

	cases := [...]struct {
		name, curr, next, result string
	}{
		{
			"updated",
			join(header,
				`# Comment`, `msgid "One"`, `msgstr "Один"`, ``,
			),
			join(header,
				`#. EComment`, `msgid "One"`, `msgstr ""`, ``,
			),
			join(header,
				`# Comment`, `#. EComment`, `msgid "One"`, `msgstr "Один"`, ``,
			),
		},
		{
			"added",
			join(header,
				`msgid "One"`, `msgstr "Один"`, ``,
			),
			join(header,
				`msgid "One"`, `msgstr "Один"`, ``,
				`msgid "Two"`, `msgstr ""`, ``,
			),
			join(header,
				`msgid "One"`, `msgstr "Один"`, ``,
				`msgid "Two"`, `msgstr ""`, ``,
			),
		},
		{
			"removed",
			join(header,
				`msgid "One"`, `msgstr "Один"`, ``,
				`msgid "Two"`, `msgstr "Два"`, ``,
			),
			join(header,
				`msgid "One"`, `msgstr ""`, ``,
			),
			join(header,
				`msgid "One"`, `msgstr "Один"`, ``,
				`#~ msgid "Two"`, `#~ msgstr "Два"`, ``,
			),
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			curr, err := pogo.ReadFile(bytes.NewBufferString(c.curr))
			require.NoError(t, err)
			next, err := pogo.ReadFile(bytes.NewBufferString(c.next))
			require.NoError(t, err)
			result := curr.Update(next)
			b := &bytes.Buffer{}
			require.NoError(t, result.Print(b))
			assert.Equal(t, c.result, b.String())
		})
	}
}
