package compiler

import (
	"crypto/sha1"
	"fmt"
	"reflect"
	"strings"

	"github.com/expr-lang/expr/ast"
)

func Identifier(node ast.Node) string {
	switch n := node.(type) {
	case *ast.UnaryNode:
		buf := strings.Builder{}
		switch n.Operator {
		case "+":
			buf.WriteString("")
		case "-":
			buf.WriteString("-(")
		case "not", "!":
			buf.WriteString("not (")
		}
		buf.WriteString(Identifier(n.Node))
		buf.WriteString(")")
		return buf.String()
	case *ast.BinaryNode:
		ls := Identifier(n.Left)
		rs := Identifier(n.Right)
		_, lw := n.Left.(*ast.BinaryNode)
		_, rw := n.Right.(*ast.BinaryNode)
		op := n.Operator
		switch op {
		case "==", "!=", "and", "or", "+", "*", "||", "&&", ">=", ">": // right / left can be swap
			if op == ">=" || op == ">" || rs <= ls {
				ls, rs = rs, ls
				lw, rw = rw, lw
			}
			if op == "&&" {
				op = "and"
			} else if op == "||" {
				op = "or"
			} else if op == ">=" { // a >= b is equal to b <= a
				op = "<="
			} else if op == ">" { // a > b is equal to b < a
				op = "<"
			}
		case "**", "^":
			op = "**"
		default:
			// do nothing
		}

		buf := strings.Builder{}
		if lw {
			buf.WriteString("(")
			buf.WriteString(ls)
			buf.WriteString(")")
		} else {
			buf.WriteString(ls)
		}
		buf.WriteString(" " + op + " ")
		if rw {
			buf.WriteString("(")
			buf.WriteString(rs)
			buf.WriteString(")")
		} else {
			buf.WriteString(rs)
		}
		return buf.String()
	default:
		return node.String()
	}
}

func (c *compiler) Visit(node *ast.Node) {
	switch n := (*node).(type) {
	case *ast.BinaryNode:
		switch n.Operator {
		case "??", "and", "or", "||", "&&":
			// do nothing
		default:
			c.countCommonExpr(*node)
		}
	case *ast.CallNode,
		 *ast.BuiltinNode:
		c.countCommonExpr(*node)
	default:
		// do nothing
	}
}

func (c *compiler) countCommonExpr(n ast.Node) {
	subExpr := Identifier(n)
	if c.exprRecords == nil || subExpr == "" {
		return
	}
	hash := fmt.Sprintf("%x", sha1.Sum([]byte(subExpr)))
	if cs, ok := c.exprRecords[hash]; !ok {
		loc := n.Location()
		c.exprRecords[hash] = &exprRecord{cnt: 1, id: -1, loc: loc}
	} else {
		cs.cnt = cs.cnt + 1
	}
}

func (c *compiler) needReuseCommon(n ast.Node) (bool, bool, int) {
	needReuseCommon, isFirstOccur, exprUniqId := false, false, -1
	if c.exprRecords != nil {
		expr := Identifier(n)
		hash := fmt.Sprintf("%x", sha1.Sum([]byte(expr)))
		cs, ok := c.exprRecords[hash]
		if ok && cs.cnt > 1 {
			if cs.id == -1 {
				cs.id = c.commonExprInc
				cs.loc = n.Location()
				c.commonExpr[cs.id] = expr
				c.commonExprInc += 1
			}
			needReuseCommon = true
			isFirstOccur = reflect.DeepEqual(n.Location(), cs.loc)
			exprUniqId = cs.id
		}
	}
	return needReuseCommon, isFirstOccur, exprUniqId
}
