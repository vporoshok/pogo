package pogo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vporoshok/pogo"
)

func TestPlainScanner(t *testing.T) {
	t.Parallel()

	s := pogo.NewPlainStarter("#| ", "msgid ")
	border, prefix, ok := s.Extract(`#| msgid "Some short text"`)
	assert.Equal(t, "#| ", border)
	assert.Equal(t, "msgid ", prefix)
	assert.True(t, ok)
	_, _, ok = s.Extract(`# msgid "Some short text"`)
	assert.False(t, ok)
}

func TestRegexpScanner(t *testing.T) {
	t.Parallel()

	s := pogo.NewRegexpStarter(`#\| `, `msgstr\[\d+\] `)
	border, prefix, ok := s.Extract(`#| msgstr[12] "Some short text"`)
	assert.Equal(t, "#| ", border)
	assert.Equal(t, "msgstr[12] ", prefix)
	assert.True(t, ok)
	_, _, ok = s.Extract(`# msgstr[12] "Some short text"`)
	assert.False(t, ok)
}
