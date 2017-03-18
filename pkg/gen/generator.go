package gen

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
)

type imports struct {
	imports []string
}

func (i *imports) Add(name string) {
	if i.imports == nil {
		i.Reset()
	}
	for _, n := range i.imports {
		if n == name {
			return
		}
	}
	i.imports = append(i.imports, name)
}

func (i *imports) Reset() {
	i.imports = make([]string, 0)
}

func (i *imports) ProtoString() string {
	if len(i.imports) > 0 {
		var buf bytes.Buffer
		for _, name := range i.imports {
			buf.WriteString("import \"" + name + "\";\n")
		}
		return buf.String() + "\n"
	}
	return ""
}

func (i *imports) GoString() string {
	if len(i.imports) > 0 {
		var buf bytes.Buffer
		for _, name := range i.imports {
			buf.WriteString("\t\"" + name + "\"\n")
		}
		return buf.String()
	}
	return ""
}

type enumfield struct {
	name  string
	alias string
}

func (e *enumfield) String() string {
	return strings.ToUpper(e.alias)
}

func (e *enumfield) SQLValue() string {
	return e.name
}

type enums struct {
	name  string
	enums []enumfield
}

func (e *enums) String() string {
	var buf bytes.Buffer
	buf.WriteString("enum " + e.name + " {\n")
	for i, n := range e.enums {
		buf.WriteString(fmt.Sprintf("\t\t%s = %d;\n", n.String(), i))
	}
	buf.WriteString("\t}\n")
	return buf.String()
}

type column struct {
	position   int64
	primarykey bool
	name       string
	datatype   string
	columntype string
	prototype  string
	enums      *enums
	table      *table
	maxlength  int64
	nullable   bool
	defvalue   string
}

func (c *column) IsChecksum() bool {
	return c.name == "checksum"
}

func (c *column) IsJSON() bool {
	return c.datatype == "json"
}

func (c *column) GenerateProtobuf() string {
	var buf bytes.Buffer
	if c.enums != nil {
		buf.WriteString(c.enums.String())
		buf.WriteString("\t")
	}
	buf.WriteString(fmt.Sprintf("%s %s = %d;", c.prototype, c.name, c.position))
	return buf.String()
}

func (c *column) GenerateSQLPlaceholder() string {
	if c.prototype == "orm.Geometry" {
		return "POINT(?,?)"
	}
	return "?"
}

func (c *column) GenerateSQLSelect() string {
	if c.prototype == "orm.Geometry" {
		return "astext(`" + c.name + "`)"
	}
	return "`" + c.name + "`"
}

func (c *column) GenerateVariableType() string {
	switch c.prototype {
	case "float":
		{
			return "float32"
		}
	}
	return c.prototype
}

func (c *column) GenerateNullValue() string {
	switch c.prototype {
	case "string":
		{
			return "\"\""
		}
	case "int32", "int64", "float":
		{
			return "0"
		}
	case "bool":
		{
			//should never have a bool primary key but who knows
			return "false"
		}
	}
	return "nil"
}

func (c *column) GetSQLType() string {
	switch c.prototype {
	case "string":
		{
			return "sql.NullString"
		}
	case "int32", "int64":
		{
			return "sql.NullInt64"
		}
	case "bool":
		{
			return "sql.NullBool"
		}
	case "float":
		{
			return "sql.NullFloat64"
		}
	case "google.protobuf.Timestamp":
		{
			return "mysql.NullTime"
		}
	case "orm.Geometry":
		{
			return "sql.NullString"
		}
	case "bytes":
		{
			switch c.datatype {
			case "blob", "mediumblob", "longblob", "varbinary", "binary":
				{
					return "sql.NullString"
				}
			}
		}
	}
	return "sql.NullString"
}

func (c *column) GenerateCast(value string) string {
	switch c.prototype {
	case "int32":
		{
			return "int32(" + value + ")"
		}
	case "float":
		{
			return "float32(" + value + ")"
		}
	case "google.protobuf.Timestamp":
		{
			return "orm.ToTimestamp(" + value + ")"
		}
	case "bytes":
		{
			return "[]byte(" + value + ")"
		}
	case "orm.Geometry":
		{
			return "orm.ToGeometry(" + value + ")"
		}
	}
	if c.enums != nil {
		enumprefix := CamelCase(c.table.name) + "_" + CamelCase(c.table.name) + CamelCase(c.name)
		return enumprefix + "(" + value + ")"
	}
	return value
}

func (c *column) GenerateSQLSetter(prefix string) string {
	switch c.prototype {
	case "string":
		{
			return c.GenerateCast(prefix + c.name + ".String")
		}
	case "int32":
		{
			return c.GenerateCast(prefix + c.name + ".Int64")
		}
	case "int64":
		{
			return c.GenerateCast(prefix + c.name + ".Int64")
		}
	case "bool":
		{
			return c.GenerateCast(prefix + c.name + ".Bool")
		}
	case "float":
		{
			return c.GenerateCast(prefix + c.name + ".Float64")
		}
	case "google.protobuf.Timestamp":
		{
			return c.GenerateCast(prefix + c.name)
		}
	case "bytes":
		{
			return c.GenerateCast(prefix + c.name + ".String")
		}
	case "orm.Geometry":
		{
			return c.GenerateCast(prefix + c.name + ".String")
		}
	}
	if c.enums != nil {
		enumprefix := CamelCase(c.table.name) + "_" + CamelCase(c.table.name) + CamelCase(c.name)
		typename := enumprefix + "_value"
		return c.GenerateCast(typename + "[" + prefix + c.name + ".String]")
	}
	fmt.Println("undefined type ", c)
	return "nil"
}

func (c *column) GenerateSQL(prefix string) string {
	name := CamelCase(c.name)
	switch c.prototype {
	case "string":
		{
			return "orm.ToSQLString(" + prefix + name + ")"
		}
	case "int32", "int64":
		{
			return "orm.ToSQLInt64(" + prefix + name + ")"
		}
	case "bool":
		{
			return "orm.ToSQLBool(" + prefix + name + ")"
		}
	case "float":
		{
			return "orm.ToSQLFloat64(" + prefix + name + ")"
		}
	case "google.protobuf.Timestamp":
		{
			return "orm.ToSQLDate(" + prefix + name + ")"
		}
	case "orm.Geometry":
		{
			return prefix + name + ".Longitude, " + prefix + name + ".Latitude"
		}
	case "bytes":
		{
			switch c.datatype {
			case "blob", "mediumblob", "longblob", "varbinary", "binary":
				{
					return "orm.ToSQLBlob(" + prefix + name + ")"
				}
			}
		}
	}
	if c.enums != nil {
		return prefix + name + ".SQLValue()"
	}
	fmt.Println("missing", c.name, c.prototype)
	return c.prototype
}

type table struct {
	name         string
	columns      []*column
	protoimports *imports
	goimports    *imports
}

func (t *table) GetChecksum() *column {
	for _, column := range t.columns {
		if column.IsChecksum() {
			return column
		}
	}
	return nil
}

func (t *table) AddColumn(position int64, name string, colkey string, datatype string, columntype string, columndef string, maxlength int64, nullable bool, prototype string, enums *enums) {
	t.columns = append(t.columns, &column{
		position:   position,
		primarykey: colkey == "PRI",
		name:       name,
		datatype:   datatype,
		columntype: columntype,
		prototype:  prototype,
		enums:      enums,
		table:      t,
		maxlength:  maxlength,
		nullable:   nullable,
		defvalue:   columndef,
	})
}

func (t *table) GetPrimaryKey() *column {
	for _, column := range t.columns {
		if column.primarykey {
			return column
		}
	}
	return nil
}

func (t *table) GenerateFuncPrefix(prefix string, table string, name string, params string, returnvalue string) string {
	return "func (" + prefix + " *" + table + ") " + name + "(" + params + ") " + returnvalue + " {\n"
}

func NewTable(name string) *table {
	return &table{
		name:         name,
		columns:      make([]*column, 0),
		goimports:    &imports{},
		protoimports: &imports{},
	}
}

func genField(tableName string, columnName string, dataType string, columnType string, table *table) (string, *enums) {
	switch dataType {
	case "char", "varchar", "text", "longtext", "mediumtext", "tinytext", "json":
		{
			return "string", nil
		}
	case "blob", "mediumblob", "longblob", "varbinary", "binary":
		{
			return "bytes", nil
		}
	case "date", "time", "datetime", "timestamp":
		{
			table.goimports.Add("github.com/go-sql-driver/mysql")
			table.protoimports.Add("google/protobuf/timestamp.proto")
			return "google.protobuf.Timestamp", nil
		}
	case "tinyint", "bool":
		{
			return "bool", nil
		}
	case "smallint", "int", "mediumint", "bigint":
		{
			return "int32", nil
		}
	case "float", "decimal", "double":
		{
			return "float", nil
		}
	case "enum", "set":
		{
			enumList := regexp.MustCompile(`[enum|set]\((.+?)\)`).FindStringSubmatch(columnType)
			e := &enums{}
			e.enums = make([]enumfield, 0)
			for _, fn := range strings.FieldsFunc(enumList[1], func(c rune) bool {
				cs := string(c)
				return "," == cs || "'" == cs
			}) {
				e.enums = append(e.enums, enumfield{fn, fn})
			}
			for i, v := range e.enums {
				r := v.name[0:1]
				switch r {
				case "+":
					{
						e.enums[i].alias = "Plus_" + v.name[1:]
						break
					}
				case "-":
					{
						e.enums[i].alias = "Minus_" + v.name[1:]
						break
					}
				}
			}
			e.name = CamelCase(tableName) + CamelCase(columnName)
			return e.name, e
		}
	case "geometry":
		{
			table.protoimports.Add("github.com/jhaynie/dbgen/pkg/orm/geometry.proto")
			return "orm.Geometry", nil
		}
	}
	return "unknown_" + dataType, nil
}

func DiscoverTables(db *sqlx.DB, schema string) ([]*table, error) {
	q := `SELECT 
		TABLE_NAME, 
		COLUMN_NAME, 
		COLUMN_KEY,
		IS_NULLABLE, 
		DATA_TYPE,
		COLUMN_DEFAULT,
		CHARACTER_MAXIMUM_LENGTH, 
		NUMERIC_PRECISION, 
		NUMERIC_SCALE, 
		COLUMN_TYPE,
		ORDINAL_POSITION
	FROM INFORMATION_SCHEMA.COLUMNS 
	WHERE TABLE_SCHEMA = ? 
	ORDER BY TABLE_NAME, ORDINAL_POSITION`

	rows, err := db.Query(q, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*table, 0)

	var currentTable *table

	for rows.Next() {
		var tableName, columnName, columnKey, isNullable, dataType, columnType, columnDef sql.NullString
		var maxLength, precision, scale, position sql.NullInt64
		if err := rows.Scan(&tableName, &columnName, &columnKey, &isNullable, &dataType, &columnDef, &maxLength, &precision, &scale, &columnType, &position); err != nil {
			return nil, err
		}
		if currentTable == nil || (currentTable != nil && currentTable.name != tableName.String) {
			currentTable = NewTable(tableName.String)
			tables = append(tables, currentTable)
		}
		field, e := genField(tableName.String, columnName.String, dataType.String, columnType.String, currentTable)
		currentTable.AddColumn(position.Int64, columnName.String, columnKey.String, dataType.String, columnType.String, columnDef.String, maxLength.Int64, isNullable.String == "YES", field, e)
	}

	return tables, nil
}

func Generate(db *sqlx.DB, schema string, packageName string, dirname string) error {
	schemaDir := path.Join(dirname, packageName)
	if err := os.MkdirAll(schemaDir, 0777); err != nil {
		return err
	}
	tables, err := DiscoverTables(db, schema)
	if err != nil {
		return err
	}
	if err := GenerateTestMain(packageName, schemaDir); err != nil {
		return err
	}
	for _, table := range tables {
		if err := table.GenerateProtobufToDir(packageName, schemaDir); err != nil {
			return err
		}
		if err := table.GenerateORMToDir(packageName, schemaDir); err != nil {
			return err
		}
		if err := table.GenerateORMTestCaseToDir(packageName, schemaDir); err != nil {
			return err
		}
	}
	return nil
}
