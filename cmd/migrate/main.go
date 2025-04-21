package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	defaultDBURL = "mysql://root:${DB_PASSWORD}@tcp(db:3306)/goforms"
)

func main() {
	var dbURL string
	flag.StringVar(&dbURL, "db-url", os.ExpandEnv(defaultDBURL), "Database URL")
	flag.Parse()

	if len(os.Args) < 2 {
		log.Fatal("Please provide a migration command (up/down)")
	}

	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	if err := runMigration(m, os.Args[1]); err != nil {
		log.Fatal(err)
	}
}

func runMigration(m *migrate.Migrate, command string) error {
	switch command {
	case "up":
		return m.Up()
	case "down":
		return m.Down()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}
