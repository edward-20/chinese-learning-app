package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/joho/godotenv"
)

var ReadWriteDb, ReadOnlyDb *sql.DB
var ReadWriteDbConnectionErr, ReadOnlyDbConnectionErr error

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbFilePath := os.Getenv("DB_FILE_PREFIX")

	switch os.Getenv("MODE") {
	case "production":
		dbFilePath += ".production.db"
	case "development":
		// make it so that everytime air refreshes we use a new database? this can be done by having a dbFilePath dynamically generated based on time?
		dbFilePath += ".development.db"
	case "testing":
		dbFilePath += ".testing.db"
	default:
		log.Fatal("Error, mode was not specified in environment file")
	}

	_, fileDoesntExistErr := os.Stat(dbFilePath)
	if fileDoesntExistErr != nil {
		// the database file doesn't exist so create it and initialise it
		_, err = os.Create(dbFilePath)
		if err != nil {
			log.Fatal("Error creating database file")
		}
	}

	ReadWriteDb, ReadWriteDbConnectionErr = sql.Open("sqlite3", dbFilePath+"?_journal=wal&busy_timeout=5000&_foreign_keys=on")
	if fileDoesntExistErr != nil {
		// if the database file didn't exist initialise it
		initScript0, err := os.ReadFile("./db/001_initial_schema.sql")
		if err != nil {
			log.Fatal("Unable to read sql initialiser script 0")
		}
		initScript1, err := os.ReadFile("./db/002_initialise_words.sql")
		if err != nil {
			log.Fatal("Unable to read sql initialiser script 1")
		}

		_, err = ReadWriteDb.Exec(string(initScript0))
		if err != nil {
			log.Fatal("Unable to execute initialiser script 0")
		}
		_, err = ReadWriteDb.Exec(string(initScript1))
		if err != nil {
			log.Fatal("Unable to execute initialiser script 1")
		}
	}
	ReadOnlyDb, ReadOnlyDbConnectionErr = sql.Open("sqlite3", dbFilePath+"?_journal=wal&busy_timeout=5000&mode=ro&_foreign_keys=on")
}
