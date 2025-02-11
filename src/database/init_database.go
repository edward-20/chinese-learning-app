package database

import (
	"database/sql"
	"fmt"
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
	dbFilePath := os.Getenv("DB_FILE")
	fmt.Println(dbFilePath)
	ReadWriteDb, ReadWriteDbConnectionErr = sql.Open("sqlite3", dbFilePath+"?_journal=wal&busy_timeout=5000&_foreign_keys=on")
	ReadOnlyDb, ReadOnlyDbConnectionErr = sql.Open("sqlite3", dbFilePath+"?_journal=wal&busy_timeout=5000&mode=ro&_foreign_keys=on")
}
