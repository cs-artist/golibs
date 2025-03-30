// 下面是一个简单的表达式求值器示例，支持基本的算术运算（加、减、乘、除）和括号以及函数调用。
// 修改函数注册机制，使用reflect，使其能够处理任意类型的参数和返回值。

// govaluate：https://github.com/Knetic/govaluate.git
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"reflect"
	"strconv"
)

// functionRegistry 存储已注册的函数
type functionRegistry struct {
	funcs map[string]reflect.Value
}

func NewFunctionRegistry() *functionRegistry {
	return &functionRegistry{
		funcs: make(map[string]reflect.Value),
	}
}

// Register 注册一个函数
func (r *functionRegistry) Register(name string, fn interface{}) error {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return fmt.Errorf("参数必须是函数类型")
	}
	r.funcs[name] = v
	return nil
}

// evaluator 结构体用于递归求值AST节点
type evaluator struct {
	registry *functionRegistry // 动态函数注册
}

func NewEvaluator(fr *functionRegistry) *evaluator {
	return &evaluator{
		registry: fr,
	}
}

// Eval 表达式求值函数
func (e *evaluator) Eval(expr string) (interface{}, error) {
	// 解析表达式
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, err
	}
	// 计算表达式
	return e.eval(node)
}

func (e *evaluator) eval(node ast.Expr) (interface{}, error) {
	switch n := node.(type) {
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
		var funcName string
		switch fun := n.Fun.(type) {
		case *ast.Ident:
			funcName = fun.Name
		case *ast.SelectorExpr: // 处理 pkg.Func 的形式
			pkg, ok := fun.X.(*ast.Ident)
			if !ok {
				return 0, fmt.Errorf("不支持的函数调用类型: %T", fun.X)
			}
			funcName = pkg.Name + "." + fun.Sel.Name
		default:
			return 0, fmt.Errorf("不支持的函数调用类型: %T", fun)
		}

		// 查找注册的函数
		fn, ok := e.registry.funcs[funcName]
		if !ok {
			return 0, fmt.Errorf("不支持的函数: %s", funcName)
		}

		// 取出参数
		args, err := e.evalArgs(n.Args)
		if err != nil {
			return 0, err
		}
		// 调用函数
		return e.callFunc(fn, args)
	case *ast.Ident: // 标识符
		return 0, fmt.Errorf("不支持的标识符: %s", n.Name)
	default:
		return 0, fmt.Errorf("不支持的表达式类型: %T", node)
	}
}

// evalArgs 参数求值
func (e *evaluator) evalArgs(args []ast.Expr) ([]interface{}, error) {
	evaluatedArgs := make([]interface{}, 0, len(args))
	for _, arg := range args {
		val, err := e.eval(arg)
		if err != nil {
			return nil, err
		}
		evaluatedArgs = append(evaluatedArgs, val)
	}
	return evaluatedArgs, nil
}

// callFunc 使用反射调用函数
func (e *evaluator) callFunc(fn reflect.Value, args []interface{}) (interface{}, error) {
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}
	out := fn.Call(in)
	result := make([]interface{}, len(out))
	for i, val := range out {
		result[i] = val.Interface()
	}
	if len(result) != 1 {
		return 0, fmt.Errorf("函数返回值数量错误: %d", len(result))
	}
	return result[0], nil
}

func main() {
	registry := NewFunctionRegistry()
	// 注册常用数学函数
	registry.Register("sin", math.Sin)
	registry.Register("cos", math.Cos)
	// 注册自定义函数
	registry.Register("add", func(a, b interface{}) interface{} {
		switch aTyped := a.(type) {
		case float64:
			bTyped, ok := b.(float64)
			if !ok {
				panic("类型不匹配")
			}
			return aTyped + bTyped
		case int:
			bTyped, ok := b.(int)
			if !ok {
				panic("类型不匹配")
			}
			return aTyped + bTyped
		default:
			panic("不支持的类型")
		}
	})

	evaluator := NewEvaluator(registry)

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
		"add(5, 3)",
		"add(5.5, 3.3)",
	}

	for _, expr := range expressions {
		result, err := evaluator.Eval(expr)
		if err != nil {
			fmt.Printf("表达式: %s 错误: %v\n", expr, err)
		} else {
			fmt.Printf("表达式: %s = %v\n", expr, result)
		}
	}
}
