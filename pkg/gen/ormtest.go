package gen

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path"
	"strings"
)

func GenerateTestMain(packageName string, schemaDir string) error {
	buf, err := os.OpenFile(path.Join(schemaDir, "testmain_test.go"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer buf.Close()
	buf.WriteString("package " + packageName + "\n")
	buf.WriteString(`
import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"testing"
	"github.com/jhaynie/dbgen/pkg/orm"
	_ "github.com/go-sql-driver/mysql"
)

var (
	database string
	username string
	password string
	hostname string
	port int
	db *sql.DB
)

func init() {
	flag.StringVar(&username, "username", "root", "database username")
	flag.StringVar(&password, "password", "", "database password")
	flag.StringVar(&hostname, "hostname", "localhost", "database hostname")
	flag.IntVar(&port, "port", 3306, "database port")
	database = fmt.Sprintf("testdb_%s", orm.UUID()[0:10])
}

func GetDatabase() *sql.DB {
	return db
}

func openDB(name string) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, hostname, port, name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return db
}

func dropDB() {
	_, err := db.Exec(fmt.Sprintf("drop database %s", database))
	if err != nil {
		fmt.Printf("error dropping database named %s\n", database)
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	// open without a database so we can create a temp one
	d := openDB("")
	defer d.Close()
	_, err := d.Exec(fmt.Sprintf("create database %s", database))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	d.Close()
	// reopen now with the temp database
	db = openDB(database)
	x := m.Run()
	dropDB()
	db.Close()
	os.Exit(x)
}

`)
	return nil
}

func (t *table) GenerateORMTestCase(packageName string, writer io.Writer) error {
	buf := bufio.NewWriter(writer)
	var codebuf bytes.Buffer
	imports := &imports{}
	imports.Add("context")
	imports.Add("fmt")
	imports.Add("testing")
	imports.Add("os")

	codebuf.WriteString("func Create" + CamelCase(t.name) + "Table(ctx context.Context) {\n")
	codebuf.WriteString("\tdb := GetDatabase()\n")
	codebuf.WriteString("\tq := \"CREATE TABLE `" + t.name + "` (\" + \n")
	for i, column := range t.columns {
		codebuf.WriteString("\t\t\"`" + column.name + "` " + column.columntype)
		if column.defvalue != "" {
			codebuf.WriteString(" DEFAULT \\\"" + column.defvalue + "\\\"")
		}
		if column.nullable == false {
			codebuf.WriteString(" NOT NULL")
		}
		if column.primarykey {
			codebuf.WriteString(" PRIMARY KEY")
		}
		if i+1 < len(t.columns) {
			codebuf.WriteString(",")
		}
		codebuf.WriteString("\" + \n")
	}
	codebuf.WriteString("\t\t\") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;\" \n")
	codebuf.WriteString(`	_, err := db.ExecContext(ctx, q)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
`)
	codebuf.WriteString("}\n\n")

	codebuf.WriteString("func Delete" + CamelCase(t.name) + "Table(ctx context.Context) {\n")
	codebuf.WriteString("\tdb := GetDatabase()\n")
	codebuf.WriteString("\tq := \"DELETE FROM `" + t.name + "`\"\n")
	codebuf.WriteString(`	_, err := db.ExecContext(ctx, q)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
`)
	codebuf.WriteString("}\n\n")

	codebuf.WriteString("func Test" + CamelCase(t.name) + "(t *testing.T) {\n")
	codebuf.WriteString("\tctx := context.Background()\n")
	codebuf.WriteString("\tdb := GetDatabase()\n")
	codebuf.WriteString("\tCreate" + CamelCase(t.name) + "Table(ctx)\n")
	codebuf.WriteString("\tDelete" + CamelCase(t.name) + "Table(ctx)\n")
	codebuf.WriteString("\t" + t.name + " := " + CamelCase(t.name) + "{}\n")
	for _, column := range t.columns {
		codebuf.WriteString("\t" + t.name + "." + CamelCase(column.name) + " = ")
		switch column.prototype {
		case "string":
			{
				if column.primarykey {
					codebuf.WriteString("orm.UUID()")
					imports.Add("github.com/jhaynie/dbgen/pkg/orm")
				} else {
					if column.IsJSON() {
						codebuf.WriteString("\"{\\\"value\\\":\\\"" + column.name + "\\\"}\"")
					} else {
						value := column.name
						if column.maxlength > 0 && len(value) > int(column.maxlength) {
							value = value[0 : column.maxlength-1]
						}
						codebuf.WriteString("\"" + value + "\"")
					}
				}
			}
		case "int32", "int64":
			{
				imports.Add("github.com/jhaynie/dbgen/pkg/orm")
				codebuf.WriteString(column.GenerateCast("orm.RandUID()"))
			}
		case "bool":
			{
				codebuf.WriteString("true")
			}
		case "float":
			{
				codebuf.WriteString(column.GenerateCast("1.104"))
			}
		case "google.protobuf.Timestamp":
			{
				imports.Add("time")
				imports.Add("github.com/go-sql-driver/mysql")
				codebuf.WriteString(column.GenerateCast("mysql.NullTime{time.Now(),true}"))
			}
		case "bytes":
			{
				codebuf.WriteString("[]byte{0x1,0x2}")
			}
		default:
			{
				if column.enums != nil {
					n := CamelCase(t.name) + "_" + strings.ToUpper(column.enums.enums[0].String())
					codebuf.WriteString(column.GenerateCast(n))
				} else {
					codebuf.WriteString("nil")
				}
			}
		}
		codebuf.WriteString("\n")
	}
	codebuf.WriteString("\tr, err := " + t.name + ".DBCreate(ctx, db)\n")
	codebuf.WriteString(`	if err != nil {
		t.Fatal(err)
	}
	rowCount, err := r.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}
	if rowCount != 1 {
		t.Fatalf("rowCount should have been 1 but was %d", rowCount)
	}
`)
	pk := t.GetPrimaryKey()
	if pk != nil {

		codebuf.WriteString("\texists, err := " + t.name + ".DBExists(ctx, db)\n")
		codebuf.WriteString(`	if err != nil {
		t.Fatal(err)
	}
	if exists == false {
		t.Fatal("exists was false and should have been true")
	}
`)
		codebuf.WriteString("\tcount, err := " + t.name + ".DBCount(ctx, db)\n")
		codebuf.WriteString(`	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("could should have been 1 but was %d", count)
	}
`)
		codebuf.WriteString("\toldpk := " + t.name + "." + CamelCase(pk.name) + "\n")
		codebuf.WriteString("\tdeleted, err := " + t.name + ".DBDelete(ctx, db)\n")
		codebuf.WriteString(`	if err != nil {
		t.Fatal(err)
	}
	if deleted == false {
		t.Fatal("record was not deleted")
	}
`)
		codebuf.WriteString("\tfound, err := " + t.name + ".DBFindOne(ctx, db, " + t.name + "." + CamelCase(pk.name) + ")\n")
		codebuf.WriteString(`	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("record was found and it should have been deleted")
	}
`)
		codebuf.WriteString("\texists, err = " + t.name + ".DBExists(ctx, db)\n")
		codebuf.WriteString(`	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("exists was true and should have been false")
	}
`)
		codebuf.WriteString("\tcount, err = " + t.name + ".DBCount(ctx, db)\n")
		codebuf.WriteString(`	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("could should have been 0 but was %d", count)
	}
`)
		codebuf.WriteString("\t// reset since delete will nullify it\n")
		codebuf.WriteString("\t" + t.name + "." + CamelCase(pk.name) + " = oldpk\n")
		codebuf.WriteString("\tinserted, updated, err := " + t.name + ".DBUpsert(ctx, db)\n")
		codebuf.WriteString(`	if err != nil {
		t.Fatal(err)
	}
	if inserted == false {
		t.Fatal("upsert should return inserted = true but was false")
	}
	if updated {
		t.Fatal("upsert should return updated = false but was true")
	}
`)
		codebuf.WriteString("\tinserted, updated, err = " + t.name + ".DBUpsert(ctx, db)\n")
		codebuf.WriteString(`	if err != nil {
		t.Fatal(err)
	}
	if inserted {
		t.Fatal("upsert should return inserted = false but was true")
	}
	if updated == false {
		t.Fatal("upsert should return updated = false but was true")
	}

`)
		codebuf.WriteString("\tfound, err = " + t.name + ".DBFindOne(ctx, db, oldpk)\n")
		codebuf.WriteString(`	if err != nil {
		t.Fatal(err)
	}
	if found == false {
		t.Fatal("findOne should return found = true but was false")
	}

`)
	}

	codebuf.WriteString("}\n")

	buf.WriteString("package " + packageName + "\n\n")
	buf.WriteString("import (\n")
	buf.WriteString(imports.GoString())
	buf.WriteString(")\n\n")
	buf.Write(codebuf.Bytes())
	buf.WriteString("\n")

	buf.Flush()
	return nil
}

func (t *table) GenerateORMTestCaseToDir(packageName, schemaDir string) error {
	f, err := os.OpenFile(path.Join(schemaDir, t.name+"_orm_test.go"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.GenerateORMTestCase(packageName, f)
}
