package gen

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"io/ioutil"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jhaynie/dbgen/pkg/orm"
	"github.com/jmoiron/sqlx"
)

var (
	database string
	username string
	password string
	hostname string
	port     int
	db       *sql.DB
	createdb = true
)

func init() {
	var defuser = "root"
	var defdb = fmt.Sprintf("testdbgen_%s", orm.UUID()[0:10])
	circle := os.Getenv("CIRCLECI")
	if circle == "true" {
		// when running on circle ci use the setup test db
		defuser = "ubuntu"
		defdb = "circle_test"
		createdb = false
	}
	flag.StringVar(&username, "username", defuser, "database username")
	flag.StringVar(&password, "password", "", "database password")
	flag.StringVar(&hostname, "hostname", "localhost", "database hostname")
	flag.IntVar(&port, "port", 3306, "database port")
	database = defdb
}

func GetDatabase() *sql.DB {
	return db
}

func GetDSN(name string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, hostname, port, name)
}

func openDB(name string) *sql.DB {
	dsn := GetDSN(name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return db
}

func dropDB() {
	if createdb {
		_, err := db.Exec(fmt.Sprintf("drop database %s", database))
		if err != nil {
			fmt.Printf("error dropping database named %s\n", database)
		}
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	if createdb {
		// open without a database so we can create a temp one
		d := openDB("")
		_, err := d.Exec(fmt.Sprintf("create database %s", database))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		d.Close()
	}
	// reopen now with the temp database
	db = openDB(database)
	x := m.Run()
	dropDB()
	db.Close()
	os.Exit(x)
}

func TestGenerator(t *testing.T) {
	dsn := GetDSN(database)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	args := make([]string, 0)
	args = append(args, "-u")
	args = append(args, username)
	args = append(args, "-h")
	args = append(args, hostname)
	args = append(args, "-P")
	args = append(args, fmt.Sprintf("%d", port))
	args = append(args, "--protocol=tcp")
	if password != "" {
		args = append(args, "-p")
		args = append(args, password)
	}
	args = append(args, database)
	cmd := exec.Command("mysql", args...)
	f, err := os.Open("./testdata/test.sql")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdin = bytes.NewReader(buf)
	t.Log("loading test schema into test database ", database)
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("generating test schema from test database ", database)
	err = Generate(db, database, "schema", "./gen")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("./gen")
	t.Log("compiling proto files")
	cwd, _ := os.Getwd()
	t.Log("cwd", cwd)
	args = make([]string, 0)
	args = append(args, "run")
	args = append(args, "--rm")
	args = append(args, "-v")
	args = append(args, cwd+":/app")
	args = append(args, "-v")
	args = append(args, os.Getenv("GOPATH")+":/go")
	args = append(args, "-w")
	args = append(args, "/app")
	args = append(args, "znly/protoc")
	args = append(args, "--go_out=gen/schema")
	args = append(args, "--proto_path=gen/schema")
	args = append(args, "--proto_path=/go/src")
	args = append(args, "gen/schema/activity_summary.proto")
	t.Log("running docker", strings.Join(args, " "))
	// run the docker container to compile the protoc file
	cmd = exec.Command("docker", args...)
	cmd.Dir = cwd
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		t.Log(stderr.String())
		t.Fatal(err)
	}
	// now run the generated unit tests
	cmd = exec.Command("go", "test", "-v", "./gen/...")
	cmd.Dir = cwd
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err = cmd.Run()
	if err != nil {
		t.Log(stderr.String())
		t.Fatal(err)
	}
	t.Log(stdout.String())
}
