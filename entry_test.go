package pogo_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vporoshok/pogo"
)

func TestReadEntry(t *testing.T) {
	t.Parallel()

	join := func(lines ...string) string {

		return strings.Join(lines, "\n")
	}

	cases := [...]struct {
		name   string
		source string
		plural int
		err    string
	}{
		{
			name: "full plural entry",
			source: join(
				`# TComment`,
				`#. EComment`,
				`#: Reference`,
				`#, Flags`,
				`#| msgctxt "PrevMsgCtxt"`,
				`#| msgid "PrevMsgID"`,
				`#| msgid_plural "PrevMsgIDP"`,
				`msgctxt "MsgCtxt"`,
				`msgid "MsgID"`,
				`msgid_plural "MsgIDP"`,
				`msgstr[0] "MsgStr"`,
				`msgstr[1] "MsgStrP"`,
			),
			plural: 2,
			err:    "",
		},
	}

	for _, c := range cases {
		// nolint:scopelint
		t.Run(c.name, func(t *testing.T) {
			s := pogo.NewScanner(bytes.NewBufferString(c.source))
			entry, err := pogo.ReadEntry(s, c.plural)
			if c.err != "" {
				assert.EqualError(t, err, c.err)
				return
			}
			require.Equal(t, io.EOF, errors.Cause(err), err)
			b := &bytes.Buffer{}
			f := pogo.NewFormatter(b)
			require.NoError(t, entry.Print(f))
			assert.Equal(t, c.source+"\n", b.String())
		})
	}
}
