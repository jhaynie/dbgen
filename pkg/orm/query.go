package orm

import (
	"bytes"
	"fmt"
	"strings"
)

type TableDef struct {
	Name  string
	Alias string
}

func (t TableDef) String() string {
	var n = "`" + t.Name + "`"
	if t.Alias != "" {
		return n + " `" + t.Alias + "`"
	}
	return n
}

func Table(name string) TableDef {
	return TableDef{
		Name: name,
	}
}

func TableAlias(name, alias string) TableDef {
	return TableDef{
		Name:  name,
		Alias: alias,
	}
}

type ColumnDef struct {
	Name  string
	Alias string
	Table string
	Expr  string
}

func (f ColumnDef) String() string {
	var n string
	if f.Name != "*" {
		n = "`" + f.Name + "`"
	} else {
		n = "*"
	}
	if f.Table != "" {
		n = "`" + f.Table + "`." + n
	}
	if f.Expr != "" {
		n = f.Expr + "(" + n + ")"
	}
	if f.Alias != "" {
		return n + " AS `" + f.Alias + "`"
	}
	return n
}

func Column(name string) ColumnDef {
	return ColumnDef{
		Name: name,
	}
}

func ColumnAlias(name, alias string) ColumnDef {
	return ColumnDef{
		Name:  name,
		Alias: alias,
	}
}

func TableColumn(table, name string) ColumnDef {
	return ColumnDef{
		Table: table,
		Name:  name,
	}
}

func TableColumnAlias(table, name, alias string) ColumnDef {
	return ColumnDef{
		Table: table,
		Name:  name,
		Alias: alias,
	}
}

func Sum(column string) ColumnDef {
	return ColumnDef{
		Expr: "SUM",
		Name: column,
	}
}

func SumAlias(column, alias string) ColumnDef {
	return ColumnDef{
		Expr:  "SUM",
		Name:  column,
		Alias: alias,
	}
}

func Count(column string) ColumnDef {
	return ColumnDef{
		Expr: "COUNT",
		Name: column,
	}
}

func CountAlias(column, alias string) ColumnDef {
	return ColumnDef{
		Expr:  "COUNT",
		Name:  column,
		Alias: alias,
	}
}

func Min(column string) ColumnDef {
	return ColumnDef{
		Expr: "MIN",
		Name: column,
	}
}

func MinAlias(column, alias string) ColumnDef {
	return ColumnDef{
		Expr:  "MIN",
		Name:  column,
		Alias: alias,
	}
}

func Max(column string) ColumnDef {
	return ColumnDef{
		Expr: "MAX",
		Name: column,
	}
}

func MaxAlias(column, alias string) ColumnDef {
	return ColumnDef{
		Expr:  "MAX",
		Name:  column,
		Alias: alias,
	}
}

type Operator string

const (
	OperatorEqual            Operator = "="
	OperatorNotEqual         Operator = "!="
	OperatorGreaterThan      Operator = ">"
	OperatorLessThan         Operator = "<"
	OperatorGreaterThanEqual Operator = ">="
	OperatorLessThanEqual    Operator = "<="
	OperatorNull             Operator = "IS NULL"
	OperatorNotNull          Operator = "IS NOT NULL"
	OperatorIn               Operator = "IN"
)

type ConditionDef struct {
	Name     string
	Func     string
	Operator Operator
	Value    interface{}
}

func (f ConditionDef) AddValue(array []interface{}) []interface{} {
	switch f.Operator {
	case OperatorNotNull, OperatorNull:
		{
			return array
		}
	}
	return append(array, f.Value)
}

func (f ConditionDef) String() string {
	var lhs string
	if f.Name != "" {
		lhs = "`" + f.Name + "`"
	} else {
		lhs = f.Func
	}
	switch f.Operator {
	case OperatorNotNull, OperatorNull:
		{
			return lhs + " " + string(f.Operator)
		}
	case OperatorIn:
		{
			return lhs + " " + string(f.Operator) + " (?)"
		}
	}
	return lhs + " " + string(f.Operator) + " ?"
}

type AndOr string

const (
	And AndOr = "AND"
	Or  AndOr = "OR"
)

type ConditionGroupDef struct {
	Conditions []ConditionDef
	AndOr      AndOr
}

func (g ConditionGroupDef) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	l := len(g.Conditions)
	for i, condition := range g.Conditions {
		buf.WriteString(condition.String())
		if i+1 < l {
			buf.WriteString(" " + string(g.AndOr) + " ")
		}
	}
	buf.WriteString(")")
	return buf.String()
}

func (g *ConditionGroupDef) Add(conditions ...ConditionDef) {
	g.Conditions = make([]ConditionDef, 0)
	for _, condition := range conditions {
		g.Conditions = append(g.Conditions, condition)
	}
}

func (g ConditionGroupDef) AddValue(params []interface{}) []interface{} {
	for _, c := range g.Conditions {
		params = c.AddValue(params)
	}
	return params
}

func OrGrouping(conditions ...ConditionDef) ConditionGroupDef {
	g := &ConditionGroupDef{
		AndOr: Or,
	}
	g.Add(conditions...)
	return *g
}

func AndGrouping(conditions ...ConditionDef) ConditionGroupDef {
	g := &ConditionGroupDef{
		AndOr: And,
	}
	g.Add(conditions...)
	return *g
}

func IsEqual(name string, value interface{}) ConditionDef {
	return ConditionDef{
		Name:     name,
		Operator: OperatorEqual,
		Value:    value,
	}
}

func IsEqualExpr(expr string, value interface{}) ConditionDef {
	return ConditionDef{
		Func:     expr,
		Operator: OperatorEqual,
		Value:    value,
	}
}

func IsNotEqual(name string, value interface{}) ConditionDef {
	return ConditionDef{
		Name:     name,
		Operator: OperatorNotEqual,
		Value:    value,
	}
}

func IsNotEqualExpr(expr string, value interface{}) ConditionDef {
	return ConditionDef{
		Func:     expr,
		Operator: OperatorNotEqual,
		Value:    value,
	}
}

func IsGreaterThan(name string, value interface{}) ConditionDef {
	return ConditionDef{
		Name:     name,
		Operator: OperatorGreaterThan,
		Value:    value,
	}
}

func IsGreaterThanExpr(expr string, value interface{}) ConditionDef {
	return ConditionDef{
		Func:     expr,
		Operator: OperatorGreaterThan,
		Value:    value,
	}
}

func IsGreaterThanEqual(name string, value interface{}) ConditionDef {
	return ConditionDef{
		Name:     name,
		Operator: OperatorGreaterThanEqual,
		Value:    value,
	}
}

func IsGreaterThanEqualExpr(expr string, value interface{}) ConditionDef {
	return ConditionDef{
		Func:     expr,
		Operator: OperatorGreaterThanEqual,
		Value:    value,
	}
}

func IsLessThan(name string, value interface{}) ConditionDef {
	return ConditionDef{
		Name:     name,
		Operator: OperatorLessThan,
		Value:    value,
	}
}

func IsLessThanExpr(expr string, value interface{}) ConditionDef {
	return ConditionDef{
		Func:     expr,
		Operator: OperatorLessThan,
		Value:    value,
	}
}

func IsLessThanEqual(name string, value interface{}) ConditionDef {
	return ConditionDef{
		Name:     name,
		Operator: OperatorLessThanEqual,
		Value:    value,
	}
}

func IsLessThanEqualExpr(expr string, value interface{}) ConditionDef {
	return ConditionDef{
		Func:     expr,
		Operator: OperatorLessThanEqual,
		Value:    value,
	}
}

func IsNull(name string) ConditionDef {
	return ConditionDef{
		Name:     name,
		Operator: OperatorNull,
	}
}

func IsNullExpr(expr string) ConditionDef {
	return ConditionDef{
		Func:     expr,
		Operator: OperatorNull,
	}
}

func IsNotNull(name string) ConditionDef {
	return ConditionDef{
		Name:     name,
		Operator: OperatorNotNull,
	}
}

func IsNotExpr(expr string) ConditionDef {
	return ConditionDef{
		Func:     expr,
		Operator: OperatorNotNull,
	}
}

func IsIn(name string, value interface{}) ConditionDef {
	return ConditionDef{
		Name:     name,
		Operator: OperatorIn,
		Value:    value,
	}
}

func IsInExpr(expr string, value interface{}) ConditionDef {
	return ConditionDef{
		Func:     expr,
		Operator: OperatorIn,
		Value:    value,
	}
}

type LimitDef struct {
	Total int32
}

func (l LimitDef) String() string {
	return fmt.Sprintf("LIMIT %d", l.Total)
}

func Limit(max int32) LimitDef {
	return LimitDef{max}
}

type RangeDef struct {
	Offset int32
	Max    int32
}

func (r RangeDef) String() string {
	return fmt.Sprintf("LIMIT %d,%d", r.Offset, r.Max)
}

func Range(offset, max int32) RangeDef {
	return RangeDef{offset, max}
}

type Direction string

const (
	DirectionAscending  Direction = "ASC"
	DirectionDescending Direction = "DESC"
)

type OrderDef struct {
	Name      string
	Direction Direction
}

func (o OrderDef) String() string {
	return "`" + o.Name + "` " + string(o.Direction)
}

func Ascending(name string) OrderDef {
	return OrderDef{name, DirectionAscending}
}

func Descending(name string) OrderDef {
	return OrderDef{name, DirectionDescending}
}

type GroupDef struct {
	Name string
}

func (g GroupDef) String() string {
	return "`" + g.Name + "`"
}

func GroupBy(name string) GroupDef {
	return GroupDef{name}
}

type JoinDef struct {
	A string
	B string
}

func (j JoinDef) String() string {
	return j.A + " = " + j.B
}

func Join(a, b string) JoinDef {
	return JoinDef{a, b}
}

func BuildQuery(components ...interface{}) (string, []interface{}) {
	var buf bytes.Buffer
	var hasField, hasTable, hasWhere, hasGroup, hasOrder, hasLimit bool
	params := make([]interface{}, 0)
	for _, component := range components {
		if f, ok := component.(ColumnDef); ok {
			if hasField == false {
				hasField = true
				buf.WriteString("SELECT ")
			} else {
				buf.WriteString(", ")
			}
			buf.WriteString(f.String())
			continue
		}
		if t, ok := component.(TableDef); ok {
			if hasTable == false {
				hasTable = true
				buf.WriteString(" FROM ")
			} else {
				buf.WriteString(", ")
			}
			buf.WriteString(t.String())
			continue
		}
		if g, ok := component.(ConditionGroupDef); ok {
			if hasWhere == false {
				hasWhere = true
				buf.WriteString(" WHERE ")
			} else {
				buf.WriteString(" AND ")
			}
			buf.WriteString(g.String())
			params = g.AddValue(params)
			continue
		}
		if c, ok := component.(ConditionDef); ok {
			if hasWhere == false {
				hasWhere = true
				buf.WriteString(" WHERE ")
			} else {
				buf.WriteString(" AND ")
			}
			buf.WriteString(c.String())
			params = c.AddValue(params)
			continue
		}
		if j, ok := component.(JoinDef); ok {
			if hasWhere == false {
				hasWhere = true
				buf.WriteString(" WHERE ")
			} else {
				buf.WriteString(" AND ")
			}
			buf.WriteString(j.String())
			continue
		}
		if g, ok := component.(GroupDef); ok {
			if hasGroup == false {
				hasGroup = true
				buf.WriteString(" GROUP BY ")
			} else {
				buf.WriteString(", ")
			}
			buf.WriteString(g.String())
			continue
		}
		if o, ok := component.(OrderDef); ok {
			if hasOrder == false {
				hasOrder = true
				buf.WriteString(" ORDER BY ")
			} else {
				buf.WriteString(",")
			}
			buf.WriteString(o.String())
			buf.WriteString(" ")
			continue
		}
		if l, ok := component.(LimitDef); ok {
			if hasLimit == false {
				hasLimit = true
				if strings.HasSuffix(buf.String(), " ") == false {
					buf.WriteString(" ")
				}
				buf.WriteString(l.String())
				continue
			}
		}
		if r, ok := component.(RangeDef); ok {
			if hasLimit == false {
				hasLimit = true
				if strings.HasSuffix(buf.String(), " ") == false {
					buf.WriteString(" ")
				}
				buf.WriteString(r.String())
				continue
			}
		}
	}
	return strings.TrimSpace(buf.String()), params
}
