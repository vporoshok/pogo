package pogo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vporoshok/pogo"
)

func TestParsePluralRule(t *testing.T) {
	t.Parallel()

	cases := [...]struct {
		name   string
		source string
		err    string
		checks map[int]bool
	}{
		{
			name:   "equal",
			source: "n == 1",
			checks: map[int]bool{
				0: false,
				1: true,
				8: false,
			},
		},
		{
			name:   "and",
			source: "n != 1 && n%10 == 1 && n < 100",
			checks: map[int]bool{
				0:   false,
				1:   false,
				8:   false,
				11:  true,
				111: false,
			},
		},
		{
			name:   "parenthes",
			source: "(n != 1 && n%10 == 1)",
			checks: map[int]bool{
				0:  false,
				1:  false,
				8:  false,
				11: true,
			},
		},
		{
			name:   "range",
			source: "(n >= 1 && n <= 12)",
			checks: map[int]bool{
				0:  false,
				1:  true,
				8:  true,
				11: true,
				23: false,
			},
		},
		{
			name:   "range",
			source: "n > 13 || n < 12",
			checks: map[int]bool{
				0:  true,
				1:  true,
				8:  true,
				11: true,
				12: false,
				13: false,
				23: true,
			},
		},
		{
			name:   "parsing error",
			source: "n <> k",
			err:    `1:4: expected operand, found '>'`,
		},
		{
			name:   "invalid expression",
			source: "n",
			err:    `1:2: invalid expression 'n'`,
		},
		{
			name:   "unknown comparsion",
			source: "n + 2",
			err:    `1:6: invalid expression 'n + 2'`,
		},
		{
			name:   "unknown symbol",
			source: "n > k",
			err:    `5:6: invalid expression 'k'`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			rule, err := pogo.ParsePluralRule(c.source)
			if c.err != "" {
				assert.EqualError(t, err, c.err)
				return
			}
			require.NoError(t, err)
			for n, r := range c.checks {
				assert.Equal(t, r, rule(n))
			}
		})
	}
}
