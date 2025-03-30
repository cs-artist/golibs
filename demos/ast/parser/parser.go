// go/parser 使用 go/scanner 来扫描源代码，并构建抽象语法树（AST），同时提供更丰富的错误处理机制。

// go/ast 是 Go 语言标准库中的一个包，用于解析和处理 Go 源代码的抽象语法树（Abstract Syntax Tree, AST）。

// 抽象语法树是一种以树状结构表示源代码语法结构的方式，使得开发者可以在编译前或编译过程中对代码进行分析、转换或生成。

// go/ast 包中的主要类型:
// File：表示一个 Go 源文件，包含包声明、导入和一系列声明（如函数、变量、类型等）。
// Decl：表示各种声明，如 FuncDecl（函数声明）、GenDecl（通用声明，如变量、常量、类型）。
// Stmt：表示语句，如 BlockStmt（代码块）、IfStmt（条件语句）、ForStmt（循环语句）等。
// Expr：表示表达式，如 BinaryExpr（二元表达式）、CallExpr（函数调用）、Ident（标识符）等。
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// 以下是一个使用 go/parser 解析文件的示例
func test_parser_file() {
	src := `
		package main

		import "fmt"

		func main() {
			fmt.Println("Hello, World!")
		}
	`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "hello.go", src, parser.ParseComments)
	if err != nil {
		// 打印解析错误
		fmt.Printf("Parse error: %v", err)
		return
	}

	// 输出 AST
	ast.Print(fset, file)

	// 遍历 AST 并打印函数名
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			fmt.Println("Func name:", fn.Name)
		}
	}
}

// 以下是一个使用 go/parser 解析表达式的示例
func test_parser_expr() {
	// expr := `a == 1 && b == 2` // 简单二元表达式
	expr := `a == 1 && b == 2 && in_array(c, []int{1,2,3,4})` // in_array 函数调用, 通过函数来扩展表达式

	fset := token.NewFileSet()
	exprAst, err := parser.ParseExprFrom(fset, "", expr, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	ast.Print(fset, exprAst)
}

func main() {
	test_parser_file()
	test_parser_expr()
}
