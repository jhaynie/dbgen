package gen

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

func (t *table) GenerateORM(packageName string, writer io.Writer) error {
	buf := bufio.NewWriter(writer)

	buf.WriteString("package " + packageName + ";\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"database/sql\"\n")
	buf.WriteString("\t\"github.com/jhaynie/dbgen/pkg/orm\"\n")
	if len(t.goimports.imports) > 0 {
		buf.WriteString(t.goimports.GoString())
	}
	buf.WriteString(")\n\n")

	n := CamelCase(t.name)
	prefix := strings.ToLower(n)
	sqlprefix := prefix + "."
	pk := t.GetPrimaryKey()
	checksum := t.GetChecksum()
	colcount := len(t.columns)

	for _, column := range t.columns {
		if column.enums != nil {
			buf.WriteString("// write out a helper for serializing alias fields for enums which have special characters\n")
			buf.WriteString("func (x " + n + "_" + n + CamelCase(column.name) + ") SQLValue() string {\n")
			buf.WriteString("	switch int(x) {\n")
			for i, e := range column.enums.enums {
				buf.WriteString(fmt.Sprintf("\t\tcase %d: {\n", i))
				buf.WriteString("\t\t\treturn \"" + e.SQLValue() + "\"\n")
				buf.WriteString(fmt.Sprintf("\t\t}\n"))
			}
			buf.WriteString("	}\n")
			buf.WriteString("	return \"\"\n")
			buf.WriteString("}\n\n")
		}
	}

	// calculate the checksum
	buf.WriteString("// Checksum returns a checksum which is a SHA256 of all the values in the record excluding the primary key and checksum\n")
	buf.WriteString(t.GenerateFuncPrefix(prefix, n, "CalculateChecksum", "", "string"))
	buf.WriteString("\treturn orm.HashStrings(\n")
	for _, column := range t.columns {
		if column.primarykey == false && column.IsChecksum() == false {
			s := prefix + "." + CamelCase(column.name)
			buf.WriteString("\t\torm.ToString(" + s + "),\n")
		}
	}
	buf.WriteString("\t)\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")

	if checksum != nil {
		buf.WriteString("// IsDirty returns true if changes have been made since the data was read based on the checksum\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBIsDirty", "", "(bool, string)"))
		buf.WriteString("\tchecksum := " + prefix + ".CalculateChecksum()\n")
		buf.WriteString("\treturn checksum != " + prefix + "." + CamelCase(checksum.name) + ", checksum\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")
	}

	// INSERT
	buf.WriteString("// Create will create a new " + CamelCase(t.name) + " record in the database\n")
	buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBCreate", "db *sql.DB", "(sql.Result, error)"))
	buf.WriteString("\tq := \"INSERT INTO `" + t.name + "` (")
	for i, column := range t.columns {
		buf.WriteString("`" + column.name + "`")
		if i+1 < colcount {
			buf.WriteString(",")
		}
	}
	buf.WriteString(") VALUES (")
	for i, column := range t.columns {
		buf.WriteString(column.GenerateSQLPlaceholder())
		if i+1 < colcount {
			buf.WriteString(",")
		}
	}
	buf.WriteString(")\"\n")
	buf.WriteString("\treturn db.Exec(q,\n")
	for _, column := range t.columns {
		if column.IsChecksum() {
			buf.WriteString("\t\t" + prefix + ".CalculateChecksum()")
		} else {
			buf.WriteString("\t\t" + column.GenerateSQL(sqlprefix))
		}
		buf.WriteString(",")
		buf.WriteString("\n")
	}
	buf.WriteString("\t)\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")

	generateScan := func(name string, indent string, returnstr string) string {
		var buf bytes.Buffer
		for _, column := range t.columns {
			buf.WriteString(indent + "\tvar _" + column.name + " " + column.GetSQLType() + "\n")
		}
		buf.WriteString(indent + "\terr := " + name + ".Scan(\n")
		for _, column := range t.columns {
			buf.WriteString(indent + "\t\t&_" + column.name + ",\n")
		}
		buf.WriteString(indent + "\t)\n")
		buf.WriteString(indent + "\tif err != nil && err != sql.ErrNoRows {\n")
		buf.WriteString(indent + "\t\treturn " + returnstr + ", err\n")
		buf.WriteString(indent + "\t}\n")
		if pk != nil {
			buf.WriteString(indent + "\tif _" + pk.name + ".Valid == false {\n")
			buf.WriteString(indent + "\t\treturn " + returnstr + ", nil\n")
			buf.WriteString(indent + "\t}\n")
		}
		for _, column := range t.columns {
			buf.WriteString(indent + "\t" + prefix + "." + CamelCase(column.name) + " = " + column.GenerateSQLSetter("_") + "\n")
		}
		return buf.String()
	}

	if pk != nil {
		// UPDATE
		buf.WriteString("// Update will update the " + n + " record in the database\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBUpdate", "db *sql.DB", "(sql.Result, error)"))
		if checksum != nil {
			buf.WriteString("\tdirty, checksum := " + prefix + ".DBIsDirty()\n")
			buf.WriteString("\tif dirty == false {\n")
			buf.WriteString("\t\treturn nil, nil\n")
			buf.WriteString("\t}\n")
			buf.WriteString("\t" + prefix + "." + CamelCase(checksum.name) + " = checksum\n")
		}
		buf.WriteString("\tq := \"UPDATE `" + t.name + "` SET ")
		for i, column := range t.columns {
			if column.primarykey == false {
				buf.WriteString("`" + column.name + "` = ")
				buf.WriteString(column.GenerateSQLPlaceholder())
				if i+1 < colcount {
					buf.WriteString(", ")
				}
			}
		}
		buf.WriteString(" WHERE `" + pk.name + "` = ?\"\n")
		buf.WriteString("\treturn db.Exec(q,\n")
		for _, column := range t.columns {
			if column.primarykey == false {
				buf.WriteString("\t\t" + column.GenerateSQL(sqlprefix) + ",\n")
			}
		}
		buf.WriteString("\t\t" + pk.GenerateSQL(sqlprefix) + ",\n")
		buf.WriteString("\t)\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

		// DELETE
		buf.WriteString("// Delete will delete the " + n + " record in the database\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBDelete", "db *sql.DB", "(bool, error)"))
		buf.WriteString("\tq := \"DELETE FROM `" + t.name + "` WHERE `" + pk.name + "` = ?\"\n")
		buf.WriteString("\tr, err := db.Exec(q, " + pk.GenerateSQL(sqlprefix) + ")\n")
		buf.WriteString("\tif err != nil && err != sql.ErrNoRows {\n")
		buf.WriteString("\t\treturn false, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\t" + prefix + "." + CamelCase(pk.name) + " = " + pk.GenerateNullValue() + "\n")
		buf.WriteString("\trows, err := r.RowsAffected()\n")
		buf.WriteString("\tif err != nil {\n")
		buf.WriteString("\t\treturn false, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\treturn rows > 0, nil\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

		// FIND ONE
		buf.WriteString("// FindOne finds a " + n + " for the primary key and populates the record with the results\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBFindOne", "db *sql.DB, "+pk.name+" "+pk.GenerateVariableType(), "(bool, error)"))
		buf.WriteString("\tq := \"SELECT ")
		for i, column := range t.columns {
			if column.primarykey == false {
				buf.WriteString(column.GenerateSQLSelect())
				if i+1 < colcount {
					buf.WriteString(",")
				}
			}
		}
		buf.WriteString(" FROM `" + t.name + "` WHERE `" + pk.name + "` = ? LIMIT 1\"\n")
		buf.WriteString("\trow := db.QueryRow(q, " + pk.name + ")\n")
		buf.WriteString(generateScan("row", "", "false"))
		buf.WriteString("\treturn true, nil\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

		// EXISTS
		buf.WriteString("// Exists returns true if the " + n + " record exists in the database\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBExists", "db *sql.DB", "(bool, error)"))
		buf.WriteString("\tq := \"SELECT " + pk.GenerateSQLSelect() + " from `" + t.name + "` WHERE " + pk.GenerateSQLSelect() + " = ?\"\n")
		buf.WriteString("\tvar _" + pk.name + " " + pk.GetSQLType() + "\n")
		buf.WriteString("\terr := db.QueryRow(q, " + prefix + "." + CamelCase(pk.name) + ").Scan(&_" + pk.name + ")\n")
		buf.WriteString("\tif err != nil && err != sql.ErrNoRows {\n")
		buf.WriteString("\t\treturn false, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\treturn _" + pk.name + ".Valid, nil\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

		// UPSERT
		buf.WriteString("// Upsert creates or updates a " + n + " record inside a safe transaction\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBUpsert", "db *sql.DB", "(bool, bool, error)"))
		buf.WriteString("\ttx, err := db.Begin()\n")
		buf.WriteString("\tif err != nil {\n")
		buf.WriteString("\t\treturn false, false, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\texists, err := " + prefix + ".DBExists(db)\n")
		buf.WriteString("\tif err != nil {\n")
		buf.WriteString("\t\ttx.Rollback()\n")
		buf.WriteString("\t\treturn false, false, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\tif exists {\n")
		buf.WriteString("\t\tr, err := " + prefix + ".DBUpdate(db)\n")
		buf.WriteString("\t\tif err != nil {\n")
		buf.WriteString("\t\t\ttx.Rollback()\n")
		buf.WriteString("\t\t\treturn false, false, err\n")
		buf.WriteString("\t\t}\n")
		buf.WriteString("\t\terr = tx.Commit()\n")
		buf.WriteString("\t\tif err != nil {\n")
		buf.WriteString("\t\t\ttx.Rollback()\n")
		buf.WriteString("\t\t\treturn false, false, err\n")
		buf.WriteString("\t\t}\n")
		buf.WriteString("\t\treturn false, r != nil, nil\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\tr, err := " + prefix + ".DBCreate(db)\n")
		buf.WriteString("\tif err != nil {\n")
		buf.WriteString("\t\ttx.Rollback()\n")
		buf.WriteString("\t\treturn false, false, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\terr = tx.Commit()\n")
		buf.WriteString("\tif err != nil {\n")
		buf.WriteString("\t\ttx.Rollback()\n")
		buf.WriteString("\t\treturn false, false, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\treturn r != nil, false, nil\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

	}

	var cstring bytes.Buffer
	for _, column := range t.columns {
		cstring.WriteString("\tparams = append(params, orm.Column(\"" + column.name + "\"))\n")
	}

	// FIND
	buf.WriteString("// Find a specific " + n + " with a filter\n")
	buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBFind", "db *sql.DB, _params ...interface{}", "(bool, error)"))
	buf.WriteString("\tparams := make([]interface{}, 0)\n")
	buf.WriteString(cstring.String())
	buf.WriteString(fmt.Sprintf(`	params = append(params, orm.Table("%s"))
	if len(_params) > 0 {
		for _, param := range _params {
			params = append(params, param)
		}
	}
`, t.name))
	buf.WriteString("\tq, p := orm.BuildQuery(params...)\n")
	buf.WriteString("\trow := db.QueryRow(q, p...)\n")
	buf.WriteString(generateScan("row", "", "false"))
	buf.WriteString("\treturn true, nil\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")

	// COUNT
	buf.WriteString("// Count the total number of " + n + " records with optional filters\n")
	buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBCount", "db *sql.DB, _params ...interface{}", "(int64, error)"))
	buf.WriteString(fmt.Sprintf(`	params := make([]interface{}, 0)
	params = append(params, orm.CountAlias("*", "count"))
	params = append(params, orm.Table("%s"))
	if len(_params) > 0 {
		for _, param := range _params {
			params = append(params, param)
		}
	}
`, t.name))
	buf.WriteString(`	q, p := orm.BuildQuery(params...)
	var count sql.NullInt64
	err := db.QueryRow(q, p...).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return count.Int64, nil
`)
	buf.WriteString("}\n")
	buf.WriteString("\n")

	// Find Many
	buf.WriteString("// Find returns " + n + " records with optional filters\n")
	buf.WriteString("func Find" + n + "s(db *sql.DB, _params ...interface{}) ([]*" + n + ", error) {\n")
	buf.WriteString("\tresults := make([]*" + n + ",0)\n")
	buf.WriteString("\tparams := make([]interface{}, 0)\n")
	buf.WriteString(cstring.String())
	buf.WriteString(fmt.Sprintf(`	params = append(params, orm.Table("%s"))
	if len(_params) > 0 {
		for _, param := range _params {
			params = append(params, param)
		}
	}
`, t.name))
	buf.WriteString("\tq, p := orm.BuildQuery(params...)\n")
	buf.WriteString("\trows, err := db.Query(q, p...)\n")
	buf.WriteString("\tif err != nil && err != sql.ErrNoRows {\n")
	buf.WriteString("\t\treturn nil, err\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tdefer rows.Close()\n")
	buf.WriteString("\tfor rows.Next() {\n")
	buf.WriteString("\t\t" + prefix + " := &" + n + "{}\n")
	buf.WriteString(generateScan("rows", "\t", "nil"))
	buf.WriteString("\t\tresults = append(results, " + prefix + ")\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn results, nil\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")

	buf.Flush()
	return nil
}

func (t *table) GenerateORMToDir(packageName, schemaDir string) error {
	f, err := os.OpenFile(path.Join(schemaDir, t.name+"_orm.go"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.GenerateORM(packageName, f)
}