package pogo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vporoshok/pogo"
)

func TestFlagsText(t *testing.T) {
	t.Parallel()

	cases := [...]struct {
		name   string
		source string
		result string
	}{
		{
			name:   "no flags",
			source: "",
			result: "",
		},
		{
			name:   "some tags",
			source: "foo, bar",
			result: "foo, bar",
		},
		{
			name:   "extra spaces",
			source: "  foo, \nbar  ",
			result: "foo, bar",
		},
		{
			name:   "duplicates",
			source: "foo, foo",
			result: "foo",
		},
	}

	var flags pogo.Flags
	for _, c := range cases {
		// nolint:scopelint
		t.Run(c.name, func(t *testing.T) {
			flags.Parse(c.source)
			assert.Equal(t, c.result, flags.String())
		})
	}
}

func TestFlags(t *testing.T) {
	var flags pogo.Flags

	assert.False(t, flags.Contain("foo"))
	assert.True(t, flags.Add("foo"))
	assert.True(t, flags.Add("bar"))
	assert.True(t, flags.Contain("foo"))
	assert.True(t, flags.Contain("bar"))
	assert.False(t, flags.Add("bar"))
	assert.True(t, flags.Contain("foo"))
	assert.True(t, flags.Contain("bar"))
	assert.True(t, flags.Remove("foo"))
	assert.False(t, flags.Contain("foo"))
	assert.True(t, flags.Contain("bar"))
	assert.False(t, flags.Remove("foo"))
	assert.False(t, flags.Contain("foo"))
	assert.True(t, flags.Contain("bar"))
}
