package pogo_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/vporoshok/pogo"
)

func TestScanner(t *testing.T) {
	t.Parallel()

	join := func(lines ...string) string {
		return strings.Join(lines, "\n")
	}

	cases := [...]struct {
		name   string
		source string
		border string
		prefix string
		text   string
		err    error
	}{
		{
			name: "short msgid",
			source: join(
				`msgid "Some short text"`,
			),
			border: "",
			prefix: "msgid ",
			text:   "Some short text",
			err:    io.EOF,
		},
		{
			name: "empty lines",
			source: join(
				``,
				``,
				`msgid "Some short text"`,
			),
			border: "",
			prefix: "msgid ",
			text:   "Some short text",
			err:    io.EOF,
		},
		{
			name: "multilines",
			source: join(
				`msgid ""`,
				`"Some text with\n"`,
				`"multilines"`,
			),
			border: "",
			prefix: "msgid ",
			text:   "Some text with\nmultilines",
			err:    io.EOF,
		},
		{
			name: "nothing",
			source: join(
				``,
				``,
				``,
			),
			border: "",
			prefix: "",
			text:   "",
			err:    io.EOF,
		},
		{
			name: "with next block",
			source: join(
				`# Some comment here`,
				`# very long`,
				`msgid "Some key"`,
				``,
			),
			border: "# ",
			prefix: "",
			text:   "Some comment here\nvery long",
			err:    nil,
		},
	}

	starters := []pogo.Starter{
		pogo.NewPlainStarter("# ", ""),
		pogo.NewPlainStarter("#| ", "msgid "),
		pogo.NewPlainStarter("", "msgid "),
		pogo.NewRegexpStarter("", `msgstr\[\d+\] `),
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			source := bytes.NewBufferString(c.source)
			s := pogo.NewScanner(source)
			s.Starters = starters
			assert.Equal(t, c.err, errors.Cause(s.Scan()))
			assert.Equal(t, c.border, s.Border)
			assert.Equal(t, c.prefix, s.Prefix)
			assert.Equal(t, c.text, s.Buffer.String())
		})
	}
}

func TestScannerError(t *testing.T) {
	t.Parallel()

	join := func(lines ...string) string {
		return strings.Join(lines, "\n")
	}

	cases := [...]struct {
		name   string
		source string
		err    string
	}{
		{
			name: "no starter found",
			source: join(
				`#) Some comment here`,
			),
			err: "no starter is matched line 1",
		},
	}

	starters := []pogo.Starter{
		pogo.NewPlainStarter("# ", ""),
		pogo.NewPlainStarter("#| ", "msgid "),
		pogo.NewPlainStarter("", "msgid "),
		pogo.NewRegexpStarter("", `msgstr\[\d+\] `),
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			source := bytes.NewBufferString(c.source)
			s := pogo.NewScanner(source)
			s.Starters = starters
			assert.EqualError(t, s.Scan(), c.err)
		})
	}
}
