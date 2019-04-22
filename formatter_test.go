package pogo_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vporoshok/pogo"
)

func TestFormatter(t *testing.T) {
	t.Parallel()

	join := func(lines ...string) string {

		return strings.Join(lines, "\n")
	}

	cases := [...]struct {
		name   string
		text   string
		border string
		prefix string
		width  int
		result string
	}{
		{
			name: "msgid",
			text: join(
				"Some long text with break lines and",
				"veryveryveryveryveryveryverylong word",
			),
			border: "",
			prefix: "msgid ",
			width:  30,
			result: join(
				`msgid ""`,
				`"Some long text with break "`,
				`"lines and\n"`,
				`"veryveryveryveryveryveryverylong "`,
				`"word"`,
				``,
			),
		},
		{
			name: "msgid",
			text: join(
				`Some long text "with break" lines and`,
				``,
				``,
				`veryveryveryveryveryveryverylong word`,
			),
			border: "",
			prefix: "msgid ",
			width:  30,
			result: join(
				`msgid ""`,
				`"Some long text \"with "`,
				`"break\" lines and\n"`,
				`"\n"`,
				`"\n"`,
				`"veryveryveryveryveryveryverylong "`,
				`"word"`,
				``,
			),
		},
		{
			name: "previous msgid",
			text: join(
				"Some long text with break lines and",
				"veryveryveryveryveryveryverylong word",
			),
			border: "#| ",
			prefix: "msgid ",
			width:  30,
			result: join(
				`#| msgid ""`,
				`#| "Some long text with "`,
				`#| "break lines and\n"`,
				`#| "veryveryveryveryveryveryverylong "`,
				`#| "word"`,
				``,
			),
		},
		{
			name: "translator comment",
			text: join(
				"Some long text with break lines and",
				"veryveryveryveryveryveryverylong word",
			),
			border: "# ",
			prefix: "",
			width:  0,
			result: join(
				`# Some long text with break lines and`,
				`# veryveryveryveryveryveryverylong word`,
				``,
			),
		},
		{
			name: "prefix oneline",
			text: join(
				"Some short text",
			),
			border: "",
			prefix: "msgid ",
			width:  0,
			result: join(
				`msgid "Some short text"`,
				``,
			),
		},
		{
			name: "prefix oneline",
			text: join(
				"Some short text",
			),
			border: "",
			prefix: "msgid ",
			width:  30,
			result: join(
				`msgid "Some short text"`,
				``,
			),
		},
	}

	result := &bytes.Buffer{}
	f := pogo.NewFormatter(result)
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			f.Border = c.border
			f.Prefix = c.prefix
			f.Width = c.width
			result.Reset()
			require.NoError(t, f.Format(c.text))
			assert.Equal(t, c.result, result.String())
		})
	}
}

type brokenWriter struct{}

func (brokenWriter) Write(_ []byte) (int, error) {

	return 0, errors.New("error")
}

func TestFormatterError(t *testing.T) {
	t.Parallel()

	f := pogo.NewFormatter(brokenWriter{})
	assert.EqualError(t, f.Format("test"), "error")
}

func TestFormatterPanic(t *testing.T) {
	t.Parallel()

	f := pogo.NewFormatter(nil)
	assert.Panics(t, func() {
		_ = f.Format("test")
	})
}
