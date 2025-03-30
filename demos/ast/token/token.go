// go/scanner 是 Go 语言标准库中的一个包，用于将 Go 源代码分解成一系列的标记（tokens）。
// 提供了高效的方式来扫描源代码，并识别出其中的各种语法元素，如标识符、关键字、运算符、注释等。

// go/token 是 Go 语言标准库中的一个包，主要用于处理源代码的位置信息（如行号、列号等）。

// Token 表示源代码中的一个语法单元，如关键字、标识符、运算符等。每个 Token 都包含其类型、位置以及在源代码中的字面值。
package main

import (
	"fmt"
	"go/scanner"
	"go/token"
)

// 以下是一个使用 go/scanner 包扫描 Go 源代码并打印所有标记的基本示例
func test_scanner() {
	// 源代码字符串
	var src = `
		package main
		import "fmt"

		func main() {
			fmt.Println("Hello World!")
		}
	`

	// 创建一个新的 FileSet
	var fset = token.NewFileSet()

	// 假设源代码来自文件 "hello.go"
	var file = fset.AddFile("hello.go", fset.Base(), len(src))
	// 打印文件信息
	fmt.Println("File Name:", file.Name(), file.Size(), file.LineCount())

	// 创建一个新的扫描器
	var s scanner.Scanner
	s.Init(file, []byte(src), nil, scanner.ScanComments)

	// 扫描源代码并打印每个标记
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		// 每一行显示了标记的位置（文件:行:列）、标记类型以及字面值
		fmt.Printf("%s\t%v\t%q\n", fset.Position(pos), tok, lit)
	}
}

func main() {
	test_scanner()
}
