package pogo

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"

	"github.com/pkg/errors"
)

// PluralRule is a condition to choose variant
type PluralRule func(int) bool

// ParsePluralRule from source
func ParsePluralRule(source string) (PluralRule, error) {

	return ruleBuilder{source}.Build()
}

type evaluer func(int) int

func number(n int) evaluer     { return func(x int) int { return n } }
func equiv() evaluer           { return func(x int) int { return x } }
func mod(a, b evaluer) evaluer { return func(x int) int { return a(x) % b(x) } }

func eql(a, b evaluer) PluralRule { return func(n int) bool { return a(n) == b(n) } }
func neq(a, b evaluer) PluralRule { return func(n int) bool { return a(n) != b(n) } }
func gtr(a, b evaluer) PluralRule { return func(n int) bool { return a(n) > b(n) } }
func geq(a, b evaluer) PluralRule { return func(n int) bool { return a(n) >= b(n) } }
func lss(a, b evaluer) PluralRule { return func(n int) bool { return a(n) < b(n) } }
func leq(a, b evaluer) PluralRule { return func(n int) bool { return a(n) <= b(n) } }

func and(a, b PluralRule) PluralRule { return func(n int) bool { return a(n) && b(n) } }
func or(a, b PluralRule) PluralRule  { return func(n int) bool { return a(n) || b(n) } }

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
		return cb.processExpression(te.X)

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

	return nil
}

func (cb ruleBuilder) invalidExpr(expr ast.Expr) {
	panic(errors.Errorf(
		"%d:%d: invalid expression '%s'",
		expr.Pos(), expr.End(),
		cb.source[int(expr.Pos())-1:int(expr.End())-1],
	))
}
