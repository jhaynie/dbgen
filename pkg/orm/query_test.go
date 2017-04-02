package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestField(t *testing.T) {
	assert := assert.New(t)
	f := Column("a")
	assert.Equal("`a`", f.String())

	f = ColumnAlias("a", "b")
	assert.Equal("`a` AS `b`", f.String())

	f = TableColumn("t", "a")
	assert.Equal("`t`.`a`", f.String())

	f = TableColumnAlias("t", "a", "b")
	assert.Equal("`t`.`a` AS `b`", f.String())

	f = Column("*")
	assert.Equal("*", f.String())

	f = Sum("foo")
	assert.Equal("SUM(`foo`)", f.String())

	f = SumAlias("foo", "sum")
	assert.Equal("SUM(`foo`) AS `sum`", f.String())

	f = Count("foo")
	assert.Equal("COUNT(`foo`)", f.String())

	f = CountAlias("foo", "count")
	assert.Equal("COUNT(`foo`) AS `count`", f.String())

	f = Min("foo")
	assert.Equal("MIN(`foo`)", f.String())

	f = MinAlias("foo", "min")
	assert.Equal("MIN(`foo`) AS `min`", f.String())

	f = Max("foo")
	assert.Equal("MAX(`foo`)", f.String())

	f = MaxAlias("foo", "max")
	assert.Equal("MAX(`foo`) AS `max`", f.String())
}

func TestTable(t *testing.T) {
	assert := assert.New(t)
	table := Table("a")
	assert.Equal("`a`", table.String())

	table = TableAlias("a", "b")
	assert.Equal("`a` `b`", table.String())
}

func TestQuery(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("`foo` = ?", IsEqual("foo", "bar").String())
	assert.Equal("`foo` != ?", IsNotEqual("foo", "bar").String())
	assert.Equal("`foo` > ?", IsGreaterThan("foo", 1).String())
	assert.Equal("`foo` >= ?", IsGreaterThanEqual("foo", 1).String())
	assert.Equal("`foo` < ?", IsLessThan("foo", 1).String())
	assert.Equal("`foo` <= ?", IsLessThanEqual("foo", 1).String())
	assert.Equal("`foo` IS NULL", IsNull("foo").String())
	assert.Equal("`foo` IS NOT NULL", IsNotNull("foo").String())
	assert.Equal("`foo` IN (?)", IsIn("foo", []string{"1", "2"}).String())
}

func TestQueryParams(t *testing.T) {
	assert := assert.New(t)
	q := IsEqual("foo", "bar")
	v := make([]interface{}, 0)
	v = q.AddValue(v)
	assert.Equal("bar", v[0])
}

func TestRange(t *testing.T) {
	assert := assert.New(t)

	r := Range(0, 10)
	assert.Equal("LIMIT 0,10", r.String())
}

func TestOrder(t *testing.T) {
	assert := assert.New(t)

	asc := Ascending("foo")
	assert.Equal("`foo` ASC", asc.String())

	desc := Descending("foo")
	assert.Equal("`foo` DESC", desc.String())
}

func TestLimit(t *testing.T) {
	assert := assert.New(t)
	limit := Limit(10)
	assert.Equal("LIMIT 10", limit.String())
}

func TestBuildQuery(t *testing.T) {
	assert := assert.New(t)

	q, p := BuildQuery(IsEqual("foo", "bar"))
	assert.Equal("WHERE `foo` = ?", q)
	assert.NotNil(p)
	assert.Len(p, 1)
	assert.Equal("bar", p[0])

	q, p = BuildQuery(IsNotEqual("foo", "bar"))
	assert.Equal("WHERE `foo` != ?", q)
	assert.NotNil(p)
	assert.Len(p, 1)
	assert.Equal("bar", p[0])

	q, p = BuildQuery(IsEqual("foo", "bar"), IsNotNull("bar"))
	assert.Equal("WHERE `foo` = ? AND `bar` IS NOT NULL", q)
	assert.NotNil(p)
	assert.Len(p, 1)
	assert.Equal("bar", p[0])

	q, p = BuildQuery(Limit(10))
	assert.Equal("LIMIT 10", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(Range(0, 10))
	assert.Equal("LIMIT 0,10", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(Ascending("foo"))
	assert.Equal("ORDER BY `foo` ASC", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(Descending("foo"))
	assert.Equal("ORDER BY `foo` DESC", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(Descending("foo"), Limit(10))
	assert.Equal("ORDER BY `foo` DESC LIMIT 10", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(OrGrouping(IsEqual("foo", "bar"), IsEqual("foo", "foo")))
	assert.Equal("WHERE (`foo` = ? OR `foo` = ?)", q)
	assert.NotNil(p)
	assert.Len(p, 2)
	assert.Equal("bar", p[0])
	assert.Equal("foo", p[1])

	q, p = BuildQuery(OrGrouping(IsEqual("foo", "bar"), IsEqual("foo", "foo")), AndGrouping(IsNotNull("a"), IsNotNull("b")))
	assert.Equal("WHERE (`foo` = ? OR `foo` = ?) AND (`a` IS NOT NULL AND `b` IS NOT NULL)", q)
	assert.NotNil(p)
	assert.Len(p, 2)
	assert.Equal("bar", p[0])
	assert.Equal("foo", p[1])

	q, p = BuildQuery(Column("foo"))
	assert.Equal("SELECT `foo`", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(Column("foo"), ColumnAlias("bar", "b"))
	assert.Equal("SELECT `foo`, `bar` AS `b`", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(Table("foo"))
	assert.Equal("FROM `foo`", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(Table("foo"), Table("bar"))
	assert.Equal("FROM `foo`, `bar`", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(Table("foo"), TableAlias("bar", "b"))
	assert.Equal("FROM `foo`, `bar` `b`", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(Column("*"), Table("foo"), TableAlias("bar", "b"))
	assert.Equal("SELECT * FROM `foo`, `bar` `b`", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(
		TableColumn("foo", "a"),
		TableColumn("bar", "b"),
		Table("foo"),
		Table("bar"),
		IsEqual("a", "bar"),
	)
	assert.Equal("SELECT `foo`.`a`, `bar`.`b` FROM `foo`, `bar` WHERE `a` = ?", q)
	assert.NotNil(p)
	assert.Len(p, 1)
	assert.Equal("bar", p[0])

	q, p = BuildQuery(Sum("foo"))
	assert.Equal("SELECT SUM(`foo`)", q)
	assert.NotNil(p)
	assert.Len(p, 0)

	q, p = BuildQuery(IsEqual("repo_id", "123"), Limit(1))
	assert.Equal("WHERE `repo_id` = ? LIMIT 1", q)
	assert.NotNil(p)
	assert.Len(p, 1)

	q, p = BuildQuery(IsLessThanEqualExpr("DATE(CONVERT_TZ(date,'UTC','-07:00'))", "123"))
	assert.Equal("WHERE DATE(CONVERT_TZ(date,'UTC','-07:00')) <= ?", q)
	assert.NotNil(p)
	assert.Len(p, 1)

	q, p = BuildQuery(GroupBy("foo"))
	assert.Equal("GROUP BY `foo`", q)

	q, p = BuildQuery(GroupBy("foo"), GroupBy("bar"))
	assert.Equal("GROUP BY `foo`, `bar`", q)

	q, p = BuildQuery(GroupBy("foo"), GroupBy("bar"), Ascending("bar"))
	assert.Equal("GROUP BY `foo`, `bar` ORDER BY `bar` ASC", q)

	q, p = BuildQuery(Join("a", "b"))
	assert.Equal("WHERE a = b", q)

	q, p = BuildQuery(Join("a.id", "b.some_id"))
	assert.Equal("WHERE a.id = b.some_id", q)

	q, p = BuildQuery(Join("a.id", "b.some_id"), Join("x", "y"))
	assert.Equal("WHERE a.id = b.some_id AND x = y", q)

	q, p = BuildQuery(ColumnDef{Expr: "DATEDIFF(COALESCE(b.closed_at, NOW()), b.created_at)", Alias: "days_open"})
	assert.Equal("SELECT DATEDIFF(COALESCE(b.closed_at, NOW()), b.created_at) AS `days_open`", q)

	q, p = BuildQuery(ColumnExprAlias("DATEDIFF(COALESCE(b.closed_at, NOW()), b.created_at)", "days_open"))
	assert.Equal("SELECT DATEDIFF(COALESCE(b.closed_at, NOW()), b.created_at) AS `days_open`", q)

	q, p = BuildQuery(ColumnExpr("DATEDIFF(COALESCE(b.closed_at, NOW()), b.created_at)"))
	assert.Equal("SELECT DATEDIFF(COALESCE(b.closed_at, NOW()), b.created_at)", q)
}
