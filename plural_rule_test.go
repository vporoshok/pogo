package pogo_test

import (
	"strings"
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
			name:   "unknown comparison",
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
			assert.Equal(t, c.source, rule.String())
			for n, r := range c.checks {
				assert.Equal(t, r, rule.Check(n))
			}
		})
	}
}

func TestParsePluralRules(t *testing.T) {
	t.Parallel()

	join := func(chunks ...string) string {
		return strings.Join(chunks, "")
	}

	cases := [...]struct {
		name   string
		source string
		err    string
		checks map[int]int
	}{
		{
			name:   "Japanese",
			source: "nplurals=1; plural=0;",
			checks: map[int]int{
				0:   0,
				1:   0,
				21:  0,
				116: 0,
			},
		},
		{
			name:   "English",
			source: "nplurals=2; plural=n != 1;",
			checks: map[int]int{
				0:   1,
				1:   0,
				21:  1,
				116: 1,
			},
		},
		{
			name:   "French",
			source: "nplurals=2; plural=n > 1;",
			checks: map[int]int{
				0:   0,
				1:   0,
				21:  1,
				116: 1,
			},
		},
		{
			name:   "Latvian",
			source: "nplurals=3; plural=n%10 == 1 && n%100 != 11 ? 0 : n != 0 ? 1 : 2;",
			checks: map[int]int{
				0:   2,
				1:   0,
				11:  1,
				116: 1,
			},
		},
		{
			name:   "Gaeilge",
			source: "nplurals=3; plural=n == 1 ? 0 : n == 2 ? 1 : 2;",
			checks: map[int]int{
				0:   2,
				1:   0,
				2:   1,
				116: 2,
			},
		},
		{
			name:   "Romanian",
			source: "nplurals=3; plural=n == 1 ? 0 : (n == 0 || (n%100 > 0 && n%100 < 20)) ? 1 : 2;",
			checks: map[int]int{
				0:   1,
				1:   0,
				31:  2,
				116: 1,
			},
		},
		{
			name: "Russian",
			source: join(
				`nplurals=3; plural=n%10 == 1 && n%100 != 11 ? 0 `,
				`: n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) ? 1 : 2;`,
			),
			checks: map[int]int{
				1:   0,
				21:  0,
				2:   1,
				3:   1,
				5:   2,
				12:  2,
				116: 2,
			},
		},
		{
			name:   "invalid source format",
			source: "something wrong",
			err:    "invalid source format",
		},
		{
			name:   "zero plurals",
			source: "nplurals=0; plural= ;",
			err:    "nplurals shouldn't be zero",
		},
		{
			name:   "one bad plural",
			source: "nplurals=1; plural=1;",
			err:    "unexpected choice 1, expected 0",
		},
		{
			name:   "two bad plurals",
			source: "nplurals=2; plural=1;",
			err:    "1:2: invalid expression '1'",
		},
		{
			name:   "three bad plurals",
			source: "nplurals=3; plural=n%10 == 1 && n%100 != 11 ? 0 : true ? 1 : 2;",
			err:    "1:5: invalid expression 'true'",
		},
		{
			name:   "bad number of rules",
			source: "nplurals=3; plural=n%10 == 1 && n%100 != 11 ? 0 : 1;",
			err:    "rules count missmatch",
		},
		{
			name: "bad choice",
			source: join(
				`nplurals=3; plural=n%10 == 1 && n%100 != 11 ? 0 `,
				`: n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) ? 2 : 1;`,
			),
			err: "unexpected choice 2, expected 1",
		},
		{
			name: "bad last choice",
			source: join(
				`nplurals=3; plural=n%10 == 1 && n%100 != 11 ? 0 `,
				`: n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) ? 1 : 3;`,
			),
			err: "unexpected choice 3, expected 2",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			rules, err := pogo.ParsePluralRules(c.source)
			if c.err != "" {
				assert.EqualError(t, err, c.err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, c.source, rules.String())
			for n, r := range c.checks {
				assert.Equal(t, r, rules.Eval(n), "n = %d", n)
			}
		})
	}
}
