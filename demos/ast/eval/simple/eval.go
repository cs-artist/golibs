// 下面是一个简单的表达式求值器示例，支持基本的算术运算（加、减、乘、除）和括号以及函数调用。
// 函数调用使用提前注册和动态函数映射来实现。

// govaluate：https://github.com/Knetic/govaluate.git
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strconv"
)

// Eval 表达式求值函数
func Eval(expr string) (float64, error) {
	// 解析表达式
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, err
	}

	// 创建一个求值器并计算表达式
	evaluator := NewEvaluator()
	return evaluator.eval(node)
}

// evaluator 结构体用于递归求值AST节点
type evaluator struct {
	funcs map[string]func(float64) float64 // 动态函数映射
}

func NewEvaluator() *evaluator {
	return &evaluator{
		funcs: map[string]func(float64) float64{
			"sin": math.Sin,
			"cos": math.Cos,
			"log": func(x float64) float64 {
				return math.Log(x)
			},
		},
	}
}

func (e *evaluator) eval(node ast.Expr) (float64, error) {
	switch n := node.(type) {
	case *ast.BinaryExpr: // 二元表达式
		left, err := e.eval(n.X)
		if err != nil {
			return 0, err
		}
		right, err := e.eval(n.Y)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			return left / right, nil
		default:
			return 0, fmt.Errorf("不支持的运算符: %s", n.Op)
		}
	case *ast.UnaryExpr: // 一元表达式
		operand, err := e.eval(n.X)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return operand, nil // +x 等于 x
		case token.SUB:
			return -operand, nil // -x
		default:
			return 0, fmt.Errorf("不支持的运算符: %s", n.Op)
		}
	case *ast.BasicLit: // 基本字面量
		switch n.Kind {
		case token.INT:
			val, err := strconv.ParseFloat(n.Value, 64)
			if err != nil {
				return 0, err
			}
			return val, nil
		case token.FLOAT:
			return strconv.ParseFloat(n.Value, 64)
		default:
			return 0, fmt.Errorf("不支持的文字类型: %s", n.Kind)
		}
	case *ast.ParenExpr: // 括号表达式
		return e.eval(n.X)
	case *ast.CallExpr: // 函数调用
		// 取出函数名
		fun, ok := n.Fun.(*ast.Ident)
		if !ok {
			return 0, fmt.Errorf("不支持的函数调用类型: %T", n.Fun)
		}

		fn, ok := e.funcs[fun.Name]
		if !ok {
			return 0, fmt.Errorf("不支持的函数: %s", fun.Name)
		}
		// 取出参数
		args := []float64{}
		for _, arg := range n.Args {
			argVal, err := e.eval(arg)
			if err != nil {
				return 0, err
			}
			args = append(args, argVal)
		}
		return fn(args[0]), nil
	case *ast.Ident: // 标识符
		return 0, fmt.Errorf("不支持的标识符: %s", n.Name)
	default:
		return 0, fmt.Errorf("不支持的表达式类型: %T", node)
	}
}

func main() {
	expressions := []string{
		"1 + 2 * 3",
		"(1 + 2) * 3",
		"10 / 2 + 5",
		"-5 + 3",
		"3.14 * 2",
		"sin(0)",
		"cos(0)",
		"log(2)",
		"exp(2)",
		"x + 2",
	}

	for _, expr := range expressions {
		result, err := Eval(expr)
		if err != nil {
			fmt.Printf("表达式: %s 错误: %v\n", expr, err)
		} else {
			fmt.Printf("表达式: %s = %v\n", expr, result)
		}
	}
}
