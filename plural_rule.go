package pogo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	pluralAllRE  = regexp.MustCompile(`^nplurals\s*=\s*(\d+)\s*;\s*plural\s*=(.+);$`)
	pluralRuleRE = regexp.MustCompile(`\s*([^?]+)\?\s*(\d+)\s*:(\s*(\d+)\s*)?`)
)

// PluralRules represent list of rules to evaluate which form should be used
type PluralRules []PluralRule

// Eval evaluate rules to find which form should be used
func (rules PluralRules) Eval(n int) int {
	if len(rules) == 1 {
		if rules[0].Check(n) {
			return 1
		}

		return 0
	}
	for i := range rules {
		if rules[i].Check(n) {
			return i
		}
	}

	return len(rules)
}

// Len is a number of rules
func (rules PluralRules) Len() int {
	return len(rules) + 1
}

// String implements fmt.Stringer
func (rules PluralRules) String() string {
	res := &strings.Builder{}
	_, _ = fmt.Fprintf(res, "nplurals=%d; plural=", rules.Len())
	if len(rules) == 1 {
		_, _ = fmt.Fprintf(res, "%s;", rules[0])
	} else {
		for i := range rules {
			_, _ = fmt.Fprintf(res, "%s ? %d : ", rules[i], i)
		}
		_, _ = fmt.Fprintf(res, "%d;", len(rules))
	}

	return res.String()
}

// ParsePluralRules from po format
func ParsePluralRules(source string) (PluralRules, error) {
	sub := pluralAllRE.FindStringSubmatch(source)
	if len(sub) != 3 {
		return nil, errors.New("invalid source format")
	}
	n, _ := strconv.Atoi(sub[1])
	switch n {
	case 0:

		return nil, errors.New("nplurals shouldn't be zero")

	case 1:
		k, err := strconv.Atoi(strings.TrimSpace(sub[2]))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if k != 0 {
			return nil, errors.Errorf("unexpected choice %d, expected 0", k)
		}

		return nil, nil

	case 2:
		rule, err := ParsePluralRule(sub[2])
		if err != nil {
			return nil, err
		}

		return PluralRules{rule}, nil

	default:

		return parsePluralRules(sub[2], n)
	}
}

func parsePluralRules(source string, n int) (PluralRules, error) {
	subs := pluralRuleRE.FindAllStringSubmatch(source, -1)
	if len(subs) != n-1 {
		return nil, errors.New("rules count missmatch")
	}
	res := make(PluralRules, n-1)
	for i, sub := range subs {
		k, _ := strconv.Atoi(sub[2])
		if k != i {
			return nil, errors.Errorf("unexpected choice %d, expected %d", k, i)
		}
		if i == n-2 {
			o, _ := strconv.Atoi(sub[4])
			if o != n-1 {
				return nil, errors.Errorf("unexpected choice %d, expected %d", o, n-1)
			}
		}
		rule, err := ParsePluralRule(strings.TrimSpace(sub[1]))
		if err != nil {
			return nil, err
		}
		res[i] = rule
	}
	return res, nil
}

// PluralRule is a condition to choose variant
type PluralRule interface {
	fmt.Stringer
	Check(int) bool
}

// ParsePluralRule from source
func ParsePluralRule(source string) (PluralRule, error) {
	return ruleBuilder{source}.Build()
}

type evaluer struct {
	source string
	Eval   func(int) int
}

func (e evaluer) String() string {
	return e.source
}

func number(n int) evaluer { return evaluer{strconv.Itoa(n), func(x int) int { return n }} }
func equiv() evaluer       { return evaluer{"n", func(x int) int { return x }} }
func mod(a, b evaluer) evaluer {
	return evaluer{
		fmt.Sprintf("%s%%%s", a, b),
		func(x int) int { return a.Eval(x) % b.Eval(x) },
	}
}

type parenthes struct {
	rule PluralRule
}

func (p parenthes) Check(n int) bool {
	return p.rule.Check(n)
}

func (p parenthes) String() string {
	return fmt.Sprintf("(%s)", p.rule)
}

type checker struct {
	op    string
	a, b  fmt.Stringer
	check func(int) bool
}

func (cr checker) String() string {
	return fmt.Sprintf("%s %s %s", cr.a, cr.op, cr.b)
}

func (cr checker) Check(n int) bool {
	return cr.check(n)
}

func eql(a, b evaluer) PluralRule {
	return checker{
		"==", a, b,
		func(n int) bool { return a.Eval(n) == b.Eval(n) },
	}
}

func neq(a, b evaluer) PluralRule {
	return checker{
		"!=", a, b,
		func(n int) bool { return a.Eval(n) != b.Eval(n) },
	}
}

func gtr(a, b evaluer) PluralRule {
	return checker{
		">", a, b,
		func(n int) bool { return a.Eval(n) > b.Eval(n) },
	}
}

func geq(a, b evaluer) PluralRule {
	return checker{
		">=", a, b,
		func(n int) bool { return a.Eval(n) >= b.Eval(n) },
	}
}

func lss(a, b evaluer) PluralRule {
	return checker{
		"<", a, b,
		func(n int) bool { return a.Eval(n) < b.Eval(n) },
	}
}

func leq(a, b evaluer) PluralRule {
	return checker{
		"<=", a, b,
		func(n int) bool { return a.Eval(n) <= b.Eval(n) },
	}
}

func and(a, b PluralRule) PluralRule {
	return checker{
		"&&", a, b,
		func(n int) bool { return a.Check(n) && b.Check(n) },
	}
}

func or(a, b PluralRule) PluralRule {
	return checker{
		"||", a, b,
		func(n int) bool { return a.Check(n) || b.Check(n) },
	}
}

type ruleBuilder struct {
	source string
}

func (cb ruleBuilder) Build() (PluralRule, error) {
	expr, err := parser.ParseExpr(cb.source)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var ch PluralRule
	err = recoverHandledError(func() {
		ch = cb.processExpression(expr)
	})

	return ch, err
}

func (cb ruleBuilder) processExpression(expr ast.Expr) PluralRule {
	switch te := expr.(type) {
	case *ast.ParenExpr:
		return parenthes{cb.processExpression(te.X)}

	case *ast.BinaryExpr:
		return cb.processComparsion(te)
	}

	cb.invalidExpr(expr)
	return nil
}

func (cb ruleBuilder) processComparsion(expr *ast.BinaryExpr) PluralRule {
	switch expr.Op {
	case token.LAND:
		return and(cb.processExpression(expr.X), cb.processExpression(expr.Y))

	case token.LOR:
		return or(cb.processExpression(expr.X), cb.processExpression(expr.Y))

	case token.EQL:
		return eql(cb.processArithmetic(expr.X), cb.processArithmetic(expr.Y))

	case token.NEQ:
		return neq(cb.processArithmetic(expr.X), cb.processArithmetic(expr.Y))

	case token.GTR:
		return gtr(cb.processArithmetic(expr.X), cb.processArithmetic(expr.Y))

	case token.GEQ:
		return geq(cb.processArithmetic(expr.X), cb.processArithmetic(expr.Y))

	case token.LSS:
		return lss(cb.processArithmetic(expr.X), cb.processArithmetic(expr.Y))

	case token.LEQ:
		return leq(cb.processArithmetic(expr.X), cb.processArithmetic(expr.Y))
	}

	cb.invalidExpr(expr)
	return nil
}

func (cb ruleBuilder) processArithmetic(expr ast.Expr) evaluer {
	switch te := expr.(type) {
	case *ast.BasicLit:
		x, err := strconv.Atoi(te.Value)
		if err == nil {
			return number(x)
		}

	case *ast.Ident:
		if te.Name == "n" {
			return equiv()
		}

	case *ast.BinaryExpr:
		if te.Op == token.REM {
			return mod(cb.processArithmetic(te.X), cb.processArithmetic(te.Y))
		}
	}

	cb.invalidExpr(expr)
	return evaluer{}
}

func (cb ruleBuilder) invalidExpr(expr ast.Expr) {
	panic(errors.Errorf(
		"%d:%d: invalid expression '%s'",
		expr.Pos(), expr.End(),
		cb.source[int(expr.Pos())-1:int(expr.End())-1],
	))
}
