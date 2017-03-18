package cmd

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jhaynie/dbgen/pkg/gen"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

var (
	Build   string
	Version string

	pkg      string
	username string
	password string
	hostname string
	database string
	dir      string
	port     int
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dbgen",
	Short: "Generate a set of go types and protobuf files from a set of database tables",
	Long: `
Generate a set of go types and protobuf files from a set of database tables

Example:

	dbgen --database foo --dir ./gen

`,
	Run: func(cmd *cobra.Command, args []string) {
		if database == "" {
			fmt.Println("database name is required")
			os.Exit(1)
		}
		if dir == "" {
			fmt.Println("directory name is required")
			os.Exit(1)
		}
		//username:password@protocol(address)/dbname?param=value
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, hostname, port, database)
		db, err := sqlx.Connect("mysql", dsn)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer db.Close()
		// run the generator
		if err := gen.Generate(db, database, pkg, dir); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().StringVar(&pkg, "package", "schema", "the golang package name")
	RootCmd.Flags().StringVarP(&database, "database", "d", "", "database name")
	RootCmd.Flags().StringVarP(&username, "username", "u", "root", "database username")
	RootCmd.Flags().StringVarP(&password, "password", "p", "", "database password")
	RootCmd.Flags().StringVar(&hostname, "hostname", "localhost", "database hostname")
	RootCmd.Flags().StringVar(&dir, "dir", "", "output directory to place generated files")
	RootCmd.Flags().IntVar(&port, "port", 3306, "database port")
}
