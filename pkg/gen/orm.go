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

// this is exteremly simplistic but good enough for now
func (t *table) pluralize(name string) string {
	if strings.HasSuffix(name, "s") {
		return name + "es"
	}
	if strings.HasSuffix(name, "y") {
		return name[0:len(name)-1] + "ies"
	}
	return name + "s"
}

func (t *table) GenerateORM(packageName string, writer io.Writer) error {
	buf := bufio.NewWriter(writer)

	var cbuf bytes.Buffer
	n := CamelCase(t.name)
	prefix := strings.ToLower(n)
	sqlprefix := prefix + "."
	pk := t.GetPrimaryKey()
	checksum := t.GetChecksum()
	colcount := len(t.columns)

	for _, column := range t.columns {
		if column.enums != nil {
			cbuf.WriteString("// write out a helper for serializing alias fields for enums which have special characters\n")
			cbuf.WriteString("func (x " + n + "_" + n + CamelCase(column.name) + ") SQLValue() string {\n")
			cbuf.WriteString("	switch int(x) {\n")
			for i, e := range column.enums.enums {
				cbuf.WriteString(fmt.Sprintf("\t\tcase %d: {\n", i))
				cbuf.WriteString("\t\t\treturn \"" + e.SQLValue() + "\"\n")
				cbuf.WriteString(fmt.Sprintf("\t\t}\n"))
			}
			cbuf.WriteString("	}\n")
			cbuf.WriteString("	return \"\"\n")
			cbuf.WriteString("}\n\n")

			enumprefix := n + "_" + n + CamelCase(column.name) + "_value"
			tn := n + "_" + n + CamelCase(column.name)
			t.goimports.Add("strings")

			cbuf.WriteString("// write out a helper for deserializing enums from SQL\n")
			cbuf.WriteString("func " + n + CamelCase(column.name) + "FromSQLValue(v sql.NullString) " + tn + " {\n")
			cbuf.WriteString("	return " + tn + "(" + enumprefix + "[strings.ToUpper(v.String)])\n")
			cbuf.WriteString("}\n\n")

			cbuf.WriteString("// write out a helper for deserializing enums from String\n")
			cbuf.WriteString("func " + n + CamelCase(column.name) + "FromStringValue(v string) " + tn + " {\n")
			cbuf.WriteString("	return " + tn + "(" + enumprefix + "[strings.ToUpper(v)])\n")
			cbuf.WriteString("}\n\n")
		}
	}

	buf.WriteString("package " + packageName + ";\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"context\"\n")
	buf.WriteString("\t\"database/sql\"\n")
	buf.WriteString("\t\"github.com/jhaynie/dbgen/pkg/orm\"\n")
	if len(t.goimports.imports) > 0 {
		buf.WriteString(t.goimports.GoString())
	}
	buf.WriteString(")\n\n")

	// write out any definitions
	buf.Write(cbuf.Bytes())

	// calculate the checksum
	buf.WriteString("// CalculateChecksum returns a checksum which is a SHA256 of all the values in the record excluding the primary key and checksum\n")
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
		buf.WriteString("// DBIsDirty returns true if changes have been made since the data was read based on the checksum\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBIsDirty", "", "(bool, string)"))
		buf.WriteString("\tchecksum := " + prefix + ".CalculateChecksum()\n")
		buf.WriteString("\treturn checksum != " + prefix + "." + CamelCase(checksum.name) + ", checksum\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")
	}

	// INSERT
	buf.WriteString("// DBCreate will create a new " + CamelCase(t.name) + " record in the database\n")
	buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBCreate", "ctx context.Context, db *sql.DB", "(sql.Result, error)"))
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
	buf.WriteString("\treturn db.ExecContext(ctx, q,\n")
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

	// INSERT with TX
	buf.WriteString("// DBCreateTx will create a new " + CamelCase(t.name) + " record in the database within an existing transaction\n")
	buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBCreateTx", "ctx context.Context, tx *sql.Tx", "(sql.Result, error)"))
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
	buf.WriteString("\treturn tx.ExecContext(ctx, q,\n")
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

	if pk != nil {
		// INSERT WITH IGNORING DUPLICATE KEY (acts like an upsert w/o a transaction)
		buf.WriteString("// DBCreateIgnoreDuplicate will create a new " + CamelCase(t.name) + " record in the database and will ignore duplicate key exception (acts like an upsert without a transaction)\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBCreateIgnoreDuplicate", "ctx context.Context, db *sql.DB", "(sql.Result, error)"))
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
		buf.WriteString(") ON DUPLICATE KEY UPDATE `" + pk.name + "` = `" + pk.name + "`\"\n")
		buf.WriteString("\treturn db.ExecContext(ctx, q,\n")
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

		// INSERT WITH IGNORING DUPLICATE KEY and Tx
		buf.WriteString("// DBCreateIgnoreDuplicateTx will create a new " + CamelCase(t.name) + " record in the database and will ignore duplicate key exception (acts like an upsert without a transaction)\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBCreateIgnoreDuplicateTx", "ctx context.Context, tx *sql.Tx", "(sql.Result, error)"))
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
		buf.WriteString(") ON DUPLICATE KEY UPDATE `" + pk.name + "` = `" + pk.name + "`\"\n")
		buf.WriteString("\treturn tx.ExecContext(ctx, q,\n")
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
	}

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
		buf.WriteString("// DBUpdate will update the " + n + " record in the database\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBUpdate", "ctx context.Context, db *sql.DB", "(sql.Result, error)"))
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
		buf.WriteString("\treturn db.ExecContext(ctx, q,\n")
		for _, column := range t.columns {
			if column.primarykey == false {
				buf.WriteString("\t\t" + column.GenerateSQL(sqlprefix) + ",\n")
			}
		}
		buf.WriteString("\t\t" + pk.GenerateSQL(sqlprefix) + ",\n")
		buf.WriteString("\t)\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

		// UPDATE with Tx
		buf.WriteString("// DBUpdateTx will update the " + n + " record in the database within an existing transaction\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBUpdateTx", "ctx context.Context, tx *sql.Tx", "(sql.Result, error)"))
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
		buf.WriteString("\treturn tx.ExecContext(ctx, q,\n")
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
		buf.WriteString("// DBDelete will delete the " + n + " record in the database\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBDelete", "ctx context.Context, db *sql.DB", "(bool, error)"))
		buf.WriteString("\tq := \"DELETE FROM `" + t.name + "` WHERE `" + pk.name + "` = ?\"\n")
		buf.WriteString("\tr, err := db.ExecContext(ctx, q, " + pk.GenerateSQL(sqlprefix) + ")\n")
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

		// DELETE Tx
		buf.WriteString("// DBDeleteTx will delete the " + n + " record in the database within an existing transaction\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBDeleteTx", "ctx context.Context, tx *sql.Tx", "(bool, error)"))
		buf.WriteString("\tq := \"DELETE FROM `" + t.name + "` WHERE `" + pk.name + "` = ?\"\n")
		buf.WriteString("\tr, err := tx.ExecContext(ctx, q, " + pk.GenerateSQL(sqlprefix) + ")\n")
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
		buf.WriteString("// DBFindOne finds a " + n + " for the primary key and populates the record with the results\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBFindOne", "ctx context.Context, db *sql.DB, "+pk.name+" "+pk.GenerateVariableType(), "(bool, error)"))
		buf.WriteString("\tq := \"SELECT ")
		for i, column := range t.columns {
			buf.WriteString(column.GenerateSQLSelect())
			if i+1 < colcount {
				buf.WriteString(",")
			}
		}
		buf.WriteString(" FROM `" + t.name + "` WHERE `" + pk.name + "` = ? LIMIT 1\"\n")
		buf.WriteString("\trow := db.QueryRowContext(ctx, q, " + pk.name + ")\n")
		buf.WriteString(generateScan("row", "", "false"))
		buf.WriteString("\treturn true, nil\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

		// FIND ONE with Tx
		buf.WriteString("// DBFindOneTx finds a " + n + " for the primary key and populates the record with the results within an existing transaction\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBFindOneTx", "ctx context.Context, tx *sql.Tx, "+pk.name+" "+pk.GenerateVariableType(), "(bool, error)"))
		buf.WriteString("\tq := \"SELECT ")
		for i, column := range t.columns {
			buf.WriteString(column.GenerateSQLSelect())
			if i+1 < colcount {
				buf.WriteString(",")
			}
		}
		buf.WriteString(" FROM `" + t.name + "` WHERE `" + pk.name + "` = ? LIMIT 1\"\n")
		buf.WriteString("\trow := tx.QueryRow(q, " + pk.name + ")\n")
		buf.WriteString(generateScan("row", "", "false"))
		buf.WriteString("\treturn true, nil\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

		// EXISTS
		buf.WriteString("// DBExists returns true if the " + n + " record exists in the database\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBExists", "ctx context.Context, db *sql.DB", "(bool, error)"))
		buf.WriteString("\tq := \"SELECT " + pk.GenerateSQLSelect() + " from `" + t.name + "` WHERE " + pk.GenerateSQLSelect() + " = ?\"\n")
		buf.WriteString("\tvar _" + pk.name + " " + pk.GetSQLType() + "\n")
		buf.WriteString("\terr := db.QueryRowContext(ctx, q, " + prefix + "." + CamelCase(pk.name) + ").Scan(&_" + pk.name + ")\n")
		buf.WriteString("\tif err != nil && err != sql.ErrNoRows {\n")
		buf.WriteString("\t\treturn false, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\treturn _" + pk.name + ".Valid, nil\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

		// EXISTS with Tx
		buf.WriteString("// DBExistsTx returns true if the " + n + " record exists in the database within an existing transaction\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBExistsTx", "ctx context.Context, tx *sql.Tx", "(bool, error)"))
		buf.WriteString("\tq := \"SELECT " + pk.GenerateSQLSelect() + " from `" + t.name + "` WHERE " + pk.GenerateSQLSelect() + " = ?\"\n")
		buf.WriteString("\tvar _" + pk.name + " " + pk.GetSQLType() + "\n")
		buf.WriteString("\terr := tx.QueryRow(q, " + prefix + "." + CamelCase(pk.name) + ").Scan(&_" + pk.name + ")\n")
		buf.WriteString("\tif err != nil && err != sql.ErrNoRows {\n")
		buf.WriteString("\t\treturn false, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\treturn _" + pk.name + ".Valid, nil\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

		// UPSERT
		buf.WriteString("// DBUpsert creates or updates a " + n + " record inside a safe transaction\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBUpsert", "ctx context.Context, db *sql.DB", "(bool, bool, error)"))
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
		buf.WriteString(") ON DUPLICATE KEY UPDATE ")
		for i, column := range t.columns {
			if column.primarykey == false {
				buf.WriteString("`" + column.name + "` = VALUES(`" + column.name + "`)")
				if i+1 < colcount {
					buf.WriteString(", ")
				}
			}
		}

		buf.WriteString("\"\n")
		buf.WriteString("\tr, err := db.ExecContext(ctx, q,\n")
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
		buf.WriteString("\tif err != nil {\n")
		buf.WriteString("\t\treturn false, false, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\tc, _ := r.RowsAffected()\n")
		buf.WriteString("\treturn c > 0, c == 0, nil\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

		// UPSERT with Tx
		buf.WriteString("// DBUpsertTx creates or updates a " + n + " record within an existing transaction\n")
		buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBUpsertTx", "ctx context.Context, tx *sql.Tx", "(bool, bool, error)"))
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
		buf.WriteString(") ON DUPLICATE KEY UPDATE ")
		for i, column := range t.columns {
			if column.primarykey == false {
				buf.WriteString("`" + column.name + "` = VALUES(`" + column.name + "`)")
				if i+1 < colcount {
					buf.WriteString(", ")
				}
			}
		}

		buf.WriteString("\"\n")
		buf.WriteString("\tr, err := tx.ExecContext(ctx, q,\n")
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
		buf.WriteString("\tif err != nil {\n")
		buf.WriteString("\t\treturn false, false, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\tc, _ := r.RowsAffected()\n")
		buf.WriteString("\treturn c > 0, c == 0, nil\n")
		buf.WriteString("}\n")
		buf.WriteString("\n")

	}

	var cstring bytes.Buffer
	for _, column := range t.columns {
		cstring.WriteString("\tparams = append(params, orm.Column(\"" + column.name + "\"))\n")
	}

	// FIND
	buf.WriteString("// DBFind will find a specific " + n + " with a filter\n")
	buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBFind", "ctx context.Context, db *sql.DB, _params ...interface{}", "(bool, error)"))
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
	buf.WriteString("\trow := db.QueryRowContext(ctx, q, p...)\n")
	buf.WriteString(generateScan("row", "", "false"))
	buf.WriteString("\treturn true, nil\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")

	// FIND with Tx
	buf.WriteString("// DBFindTx will find a specific " + n + " with a filter\n")
	buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBFindTx", "ctx context.Context, tx *sql.Tx, _params ...interface{}", "(bool, error)"))
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
	buf.WriteString("\trow := tx.QueryRowContext(ctx, q, p...)\n")
	buf.WriteString(generateScan("row", "", "false"))
	buf.WriteString("\treturn true, nil\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")

	// COUNT
	buf.WriteString("// DBCount will return the total number of " + n + " records with optional filters\n")
	buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBCount", "ctx context.Context, db *sql.DB, _params ...interface{}", "(int64, error)"))
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
	err := db.QueryRowContext(ctx, q, p...).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return count.Int64, nil
`)
	buf.WriteString("}\n")
	buf.WriteString("\n")

	// COUNT with Tx
	buf.WriteString("// DBCountTx will return the total number of " + n + " records with optional filters\n")
	buf.WriteString(t.GenerateFuncPrefix(prefix, n, "DBCountTx", "ctx context.Context, tx *sql.Tx, _params ...interface{}", "(int64, error)"))
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
	err := tx.QueryRowContext(ctx, q, p...).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return count.Int64, nil
`)
	buf.WriteString("}\n")
	buf.WriteString("\n")

	deleteAll := t.pluralize("DeleteAll" + n)

	// Delete all
	buf.WriteString("// " + deleteAll + " deletes all " + n + " records in the database with optional filters\n")
	buf.WriteString("func " + deleteAll + "(ctx context.Context, db *sql.DB, _params ...interface{}) (error) {\n")
	buf.WriteString("\tparams := make([]interface{}, 0)\n")
	buf.WriteString(fmt.Sprintf(`	params = append(params, orm.Table("%s"))
	if len(_params) > 0 {
		for _, param := range _params {
			params = append(params, param)
		}
	}
`, t.name))
	buf.WriteString("\tq, p := orm.BuildQuery(params...)\n")
	buf.WriteString("\t_, err := db.ExecContext(ctx, \"DELETE \"+ q, p...)\n")
	buf.WriteString("\treturn err\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")

	// Delete all Tx
	buf.WriteString("// " + deleteAll + "Tx deletes all " + n + " records in the database with optional filters\n")
	buf.WriteString("func " + deleteAll + "Tx(ctx context.Context, tx *sql.Tx, _params ...interface{}) (error) {\n")
	buf.WriteString("\tparams := make([]interface{}, 0)\n")
	buf.WriteString(fmt.Sprintf(`	params = append(params, orm.Table("%s"))
	if len(_params) > 0 {
		for _, param := range _params {
			params = append(params, param)
		}
	}
`, t.name))
	buf.WriteString("\tq, p := orm.BuildQuery(params...)\n")
	buf.WriteString("\t_, err := tx.ExecContext(ctx, \"DELETE \"+ q, p...)\n")
	buf.WriteString("\treturn err\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")

	find := t.pluralize("Find" + n)

	// Find Many
	buf.WriteString("// " + find + " returns " + n + " records with optional filters\n")
	buf.WriteString("func " + find + "(ctx context.Context, db *sql.DB, _params ...interface{}) ([]*" + n + ", error) {\n")
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
	buf.WriteString("\trows, err := db.QueryContext(ctx, q, p...)\n")
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

	// Find Many Tx
	buf.WriteString("// " + find + "Tx returns " + n + " records with optional filters\n")
	buf.WriteString("func " + find + "Tx(ctx context.Context, tx *sql.Tx, _params ...interface{}) ([]*" + n + ", error) {\n")
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
	buf.WriteString("\trows, err := tx.QueryContext(ctx, q, p...)\n")
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

	count := t.pluralize("Count" + n)

	// Count
	buf.WriteString("// " + count + " returns the number of " + t.pluralize(n) + " with optional filters\n")
	buf.WriteString("func " + count + "(ctx context.Context, tx *sql.Tx, _params ...interface{}) (int, error) {\n")
	buf.WriteString("\tparams := make([]interface{}, 0)\n")
	buf.WriteString(fmt.Sprintf(`	params = append(params, orm.Count("*"), orm.Table("%s"))
	if len(_params) > 0 {
		for _, param := range _params {
			params = append(params, param)
		}
	}
`, t.name))
	buf.WriteString("\tq, p := orm.BuildQuery(params...)\n")
	buf.WriteString("\tvar c int\n")
	buf.WriteString("\terr := tx.QueryRowContext(ctx, q, p...).Scan(&c)\n")
	buf.WriteString("\tif err != nil && err != sql.ErrNoRows {\n")
	buf.WriteString("\t\treturn 0, err\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn c, nil\n")
	buf.WriteString("}\n")
	buf.WriteString("\n")

	// Count Tx
	buf.WriteString("// " + count + "Tx returns the number of " + t.pluralize(n) + " with optional filters\n")
	buf.WriteString("func " + count + "Tx(ctx context.Context, tx *sql.Tx, _params ...interface{}) (int, error) {\n")
	buf.WriteString("\tparams := make([]interface{}, 0)\n")
	buf.WriteString(fmt.Sprintf(`	params = append(params, orm.Count("*"), orm.Table("%s"))
	if len(_params) > 0 {
		for _, param := range _params {
			params = append(params, param)
		}
	}
`, t.name))
	buf.WriteString("\tq, p := orm.BuildQuery(params...)\n")
	buf.WriteString("\tvar c int\n")
	buf.WriteString("\terr := tx.QueryRowContext(ctx, q, p...).Scan(&c)\n")
	buf.WriteString("\tif err != nil && err != sql.ErrNoRows {\n")
	buf.WriteString("\t\treturn 0, err\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn c, nil\n")
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
