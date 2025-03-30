// SQL parser
// Go package for parsing MySQL SQL queries
//
// https://github.com/cch123/elasticsql.git
// https://elasticsearch.cn/article/114
package main

import (
	"fmt"

	"github.com/xwb1989/sqlparser"
)

func main() {
	// sql := `SELECT id, name FROM users WHERE age > 30 ORDER BY name DESC LIMIT 10`
	// sql := `SELECT id, name FROM users WHERE age != 30 ORDER BY name DESC LIMIT 10`
	// sql := `SELECT id, name FROM users WHERE age > 30 and (name < 10 or id > 10) ORDER BY name DESC LIMIT 10`
	// sql := `SELECT id, name FROM users WHERE age not in (1, 2, 3) and name < 10 ORDER BY name DESC LIMIT 10`
	sql := `select * from aaa where a=1 and x = '三个男人' and create_time between '2015-01-01T00:00:00+0800' and '2016-01-01T00:00:00+0800' and process_id > 1 order by id desc limit 100,10`
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Otherwise do something with stmt
	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		output(stmt)
		parse(stmt)
	default:
		fmt.Println("不支持的 SQL 类型")
	}
}

func output(stmt *sqlparser.Select) {
	fmt.Printf("Select statement: \n"+
		"  Cache: %s\n"+
		"  Comments: %#v\n"+
		"  Distinct: %s\n"+
		"  Hints: %s\n"+
		"  SelectExprs: %#v\n"+
		"  From: %#v\n"+
		"  Where: %#v\n"+
		"  GroupBy: %#v\n"+
		"  Having: %#v\n"+
		"  OrderBy: %#v\n"+
		"  Limit: %#v\n"+
		"  Lock: %#v\n",
		stmt.Cache, stmt.Comments, stmt.Distinct, stmt.Hints,
		stmt.SelectExprs, stmt.From, stmt.Where, stmt.GroupBy, stmt.Having,
		stmt.OrderBy, stmt.Limit, stmt.Lock)
}

// parse 解析 select 语句
func parse(stmt *sqlparser.Select) {
	fmt.Println("========= parse ==========")

	// 输出列
	var cols []string
	for _, expr := range stmt.SelectExprs {
		switch selExpr := expr.(type) {
		case *sqlparser.AliasedExpr:
			cols = append(cols, sqlparser.String(selExpr.Expr))
		}
	}
	fmt.Printf("cols: %v\n", cols)

	// 输出表名
	fmt.Printf("table: %s\n", sqlparser.String(stmt.From))

	// 输出条件
	fmt.Printf("where: %s\n", handleSelectWhere(stmt.Where.Expr))

	// 输出排序
	for _, order := range stmt.OrderBy {
		fmt.Printf("order by: %s %s\n", sqlparser.String(order.Expr), order.Direction)
	}

	// 输出分页
	fmt.Printf("limit: %s %s\n", sqlparser.String(stmt.Limit.Offset), sqlparser.String(stmt.Limit.Rowcount))
}

// handleSelectWhere 处理 select 语句的 where 条件
func handleSelectWhere(expr sqlparser.Expr) string {
	switch e := expr.(type) {
	case *sqlparser.AndExpr:
		return handleSelectWhere(e.Left) + " && " + handleSelectWhere(e.Right)
	case *sqlparser.OrExpr:
		return handleSelectWhere(e.Left) + " || " + handleSelectWhere(e.Right)
	case *sqlparser.ParenExpr:
		return "(" + handleSelectWhere(e.Expr) + ")"
	case *sqlparser.NotExpr:
		return "! " + handleSelectWhere(e.Expr)
	case *sqlparser.ComparisonExpr:
		return handleSelectWhere(e.Left) + " " + e.Operator + " " + handleSelectWhere(e.Right)
	case *sqlparser.RangeCond:
		return handleSelectWhere(e.Left) + " " + e.Operator + " [" + handleSelectWhere(e.From) + ", " + handleSelectWhere(e.To) + "]"
	case *sqlparser.ColName:
		return e.Name.String()
	case *sqlparser.SQLVal:
		return string(e.Val)
	case sqlparser.ValTuple:
		return sqlparser.String(e)
	default:
		fmt.Printf("unknown expr type: %T\n", expr)
		return ""
	}
}
